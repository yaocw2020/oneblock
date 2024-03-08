package modeltemplate

import (
	"context"
	"fmt"
	"sync"

	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"

	mlv1 "github.com/oneblock-ai/oneblock/pkg/apis/ml.oneblock.ai/v1"
	ctlmlv1 "github.com/oneblock-ai/oneblock/pkg/generated/controllers/ml.oneblock.ai/v1"
	"github.com/oneblock-ai/oneblock/pkg/server/config"
	"github.com/oneblock-ai/oneblock/pkg/utils"
	"github.com/oneblock-ai/oneblock/pkg/utils/constant"
)

const (
	templateControllerSetDefautVersion  = "template-controller-set-default-version"
	templateControllerSyncLatestVersion = "template-controller-sync-latest-version"
	templateControllerAssignVersion     = "template-controller-assign-version"
)

type Handler struct {
	templateController    ctlmlv1.ModelTemplateController
	templateClient        ctlmlv1.ModelTemplateClient
	templateCache         ctlmlv1.ModelTemplateCache
	templateVersionClient ctlmlv1.ModelTemplateVersionClient
	templateVersionCache  ctlmlv1.ModelTemplateVersionCache

	latestVersionMap map[string]int
	mutex            *sync.Mutex
}

func Register(ctx context.Context, mgmt *config.Management) error {
	templates := mgmt.OneBlockMLFactory.Ml().V1().ModelTemplate()
	templateVersions := mgmt.OneBlockMLFactory.Ml().V1().ModelTemplateVersion()
	h := &Handler{
		templateController:    templates,
		templateClient:        templates,
		templateCache:         templates.Cache(),
		templateVersionClient: templateVersions,
		templateVersionCache:  templateVersions.Cache(),

		latestVersionMap: make(map[string]int),
		mutex:            &sync.Mutex{},
	}

	if err := h.initLatestVersionMap(); err != nil {
		return fmt.Errorf("failed to init latest version map: %w", err)
	}

	templates.OnChange(ctx, templateControllerSetDefautVersion, h.SetDefaultVersion)
	templates.OnChange(ctx, templateControllerSyncLatestVersion, h.SyncLatestVersion)
	templates.OnRemove(ctx, templateControllerSyncLatestVersion, h.DeleteLatestVersion)
	templateVersions.OnChange(ctx, templateControllerAssignVersion, h.AssignVersion)

	return nil
}

// initLatestVersionMap initializes the latest version map by querying all the template versions
func (h *Handler) initLatestVersionMap() error {
	tvs, err := h.templateVersionClient.List("", metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, tv := range tvs.Items {
		ref := utils.NewRef(tv.Namespace, tv.Spec.TemplateName)
		// avoid gosec G601: Implicit memory aliasing from the loop variable
		tempTv := tv
		if mlv1.VersionAssigned.IsTrue(&tempTv) && tv.Status.Version > h.latestVersionMap[ref] {
			h.mutex.Lock()
			h.latestVersionMap[ref] = tv.Status.Version
			h.mutex.Unlock()
		}
	}

	return nil
}

// SetDefaultVersion sets the default version for the template
func (h *Handler) SetDefaultVersion(_ string, tp *mlv1.ModelTemplate) (*mlv1.ModelTemplate, error) {
	if tp == nil || tp.DeletionTimestamp != nil {
		return tp, nil
	}

	// set the first version as the default version if the default version id is empty.
	if tp.Spec.DefaultVersionID == "" {
		firstTemplateVersion, err := h.getFirstTemplateVersion(tp)
		if err != nil {
			return nil, err
		}
		if firstTemplateVersion == nil {
			return tp, nil
		}

		if mlv1.VersionAssigned.IsFalse(firstTemplateVersion) {
			return nil, fmt.Errorf("the first template version %s/%s of template %s/%s haven7t been assigned a version number",
				firstTemplateVersion.Namespace, firstTemplateVersion.Name, tp.Namespace, tp.Name)
		}
		tpCopy := tp.DeepCopy()
		tpCopy.Spec.DefaultVersionID = utils.NewRef(firstTemplateVersion.Namespace, firstTemplateVersion.Name)
		tpCopy.Status.DefaultVersion = firstTemplateVersion.Status.Version
		return h.templateClient.Update(tpCopy)
	}

	tpv, err := h.templateVersionCache.Get(utils.Ref(tp.Spec.DefaultVersionID).Parse())
	if err != nil {
		return nil, err
	}
	if mlv1.VersionAssigned.IsFalse(tpv) {
		return nil, fmt.Errorf("the template version %s haven't been assigned a version number", tp.Spec.DefaultVersionID)
	}
	if tpv.Status.Version == tp.Status.DefaultVersion {
		return tp, nil
	}
	tpCopy := tp.DeepCopy()
	tpCopy.Status.DefaultVersion = tpv.Status.Version
	return h.templateClient.Update(tpCopy)
}

// getFirstTemplateVersion gets the first template version of the template
func (h *Handler) getFirstTemplateVersion(tp *mlv1.ModelTemplate) (*mlv1.ModelTemplateVersion, error) {
	selector := labels.Set(map[string]string{constant.LabelModelTemplateName: tp.Name}).AsSelector()
	tpvs, err := h.templateVersionCache.List(tp.Namespace, selector)
	if err != nil {
		return nil, err
	}
	if len(tpvs) == 0 {
		return nil, nil
	}

	return tpvs[0], nil
}

func (h *Handler) DeleteLatestVersion(_ string, tp *mlv1.ModelTemplate) (*mlv1.ModelTemplate, error) {
	if tp == nil {
		return tp, nil
	}

	key := utils.NewRef(tp.Namespace, tp.Name)
	h.mutex.Lock()
	defer h.mutex.Unlock()
	delete(h.latestVersionMap, key)

	return tp, nil
}

// SyncLatestVersion syncs the latest version from memory to the template CR
func (h *Handler) SyncLatestVersion(_ string, tp *mlv1.ModelTemplate) (*mlv1.ModelTemplate, error) {
	if tp == nil || tp.DeletionTimestamp != nil {
		return tp, nil
	}

	key := utils.NewRef(tp.Namespace, tp.Name)
	if tp.Status.LatestVersion == h.latestVersionMap[key] {
		return tp, nil
	}

	tpCopy := tp.DeepCopy()
	tpCopy.Status.LatestVersion = h.latestVersionMap[key]
	logrus.Debugf("delete the latest version record of %s/%s, %+v", tp.Namespace, tp.Name, h.latestVersionMap)

	return h.templateClient.Update(tpCopy)
}

// AssignVersion assigns a version number to the template version
func (h *Handler) AssignVersion(_ string, tpv *mlv1.ModelTemplateVersion) (*mlv1.ModelTemplateVersion, error) {
	if tpv == nil || tpv.DeletionTimestamp != nil || mlv1.VersionAssigned.IsTrue(tpv) {
		return tpv, nil
	}

	tp, err := h.templateCache.Get(tpv.Namespace, tpv.Spec.TemplateName)
	if err != nil {
		return nil, err
	}

	tpvCopy := tpv.DeepCopy()
	// add template label
	if tpvCopy.Labels == nil {
		tpvCopy.Labels = make(map[string]string)
	}
	tpvCopy.Labels[constant.LabelModelTemplateName] = tp.Name

	// add owner reference
	flagTrue := true
	tpvCopy.OwnerReferences = append(tpv.OwnerReferences, metav1.OwnerReference{
		Name:               tp.Name,
		APIVersion:         tp.APIVersion,
		UID:                tp.UID,
		Kind:               tp.Kind,
		BlockOwnerDeletion: &flagTrue,
		Controller:         &flagTrue,
	})

	// assign version
	tpvCopy.Status.Version = h.latestVersionMap[utils.NewRef(tp.Namespace, tp.Name)] + 1
	mlv1.VersionAssigned.True(&tpvCopy.Status)

	tpv, err = h.templateVersionClient.Update(tpvCopy)
	if err != nil {
		return nil, err
	}

	h.mutex.Lock()
	h.latestVersionMap[utils.NewRef(tp.Namespace, tp.Name)]++
	h.mutex.Unlock()
	logrus.Debug("Assign new version: ", h.latestVersionMap)

	// trigger the template controller to sync the latest version
	h.templateController.Enqueue(tp.Namespace, tp.Name)

	return tpv, nil
}
