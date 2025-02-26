package v1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1 "github.com/oneblock-ai/oneblock/pkg/apis/management.oneblock.ai/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=nb,scope=Namespaced
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=`.metadata.creationTimestamp`

// Notebook is the Schema for the notebooks API
type Notebook struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NotebookSpec   `json:"spec,omitempty"`
	Status NotebookStatus `json:"status,omitempty"`
}

// NotebookSpec defines the desired state of Dataset
type NotebookSpec struct {
	// Template describes the notebooks that will be created.
	Template    NotebookTemplateSpec `json:"template,omitempty"`
	ServiceType corev1.ServiceType   `json:"serviceType,omitempty"`
}

type NotebookTemplateSpec struct {
	Spec corev1.PodSpec `json:"spec,omitempty"`
}

// NotebookStatus defines the observed state of Dataset
type NotebookStatus struct {
	// Conditions is an array of current conditions
	Conditions []v1.Condition `json:"conditions"`
	// ReadyReplicas is the number of Pods created by the StatefulSet controller that have a Ready Condition.
	ReadyReplicas int32 `json:"readyReplicas"`
	// ContainerState is the state of underlying container.
	State corev1.ContainerState `json:"state"`
}
