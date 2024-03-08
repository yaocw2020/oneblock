package v1

import (
	"github.com/rancher/wrangler/v2/pkg/condition"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:resource:shortName=mt;mts
// +kubebuilder:scope=Namespaced
// +kubebuilder:printcolumn:name="DESCRIPTION",type=string,JSONPath=`.spec.description`
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=`.metadata.creationTimestamp`

// ModelTemplate is the Schema for the LLM template
type ModelTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ModelTemplateSpec   `json:"spec,omitempty"`
	Status ModelTemplateStatus `json:"status,omitempty"`
}

type ModelTemplateSpec struct {
	// +optional
	Description string `json:"description,omitempty"`
	// +optional
	DefaultVersionID string `json:"defaultVersionId"`
}

type ModelTemplateStatus struct {
	// +optional
	DefaultVersion int `json:"defaultVersion,omitempty"`
	// +optional
	LatestVersion int `json:"latestVersion,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:resource:shortName=mtv;mtvs
// +kubebuilder:scope=Namespaced
// +kubebuilder:printcolumn:name="TEMPLATE_ID",type=string,JSONPath=`.spec.templatedId`
// +kubebuilder:printcolumn:name="DESCRIPTION",type=string,JSONPath=`.spec.description`
// +kubebuilder:printcolumn:name="VERSION",type=integer,JSONPath=`.status.version`
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=`.metadata.creationTimestamp`

// ModelTemplateVersion is the Schema for the LLM template version
type ModelTemplateVersion struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ModelTemplateVersionSpec   `json:"spec,omitempty"`
	Status ModelTemplateVersionStatus `json:"status,omitempty"`
}

type ModelTemplateVersionSpec struct {
	// +kubebuilder:validation:Required
	TemplateName string `json:"template_name"`
	// +optional
	Description string `json:"description"`

	DeploymentConfig DeploymentConfig `json:"deployment_config"`
	EngineConfig     EngineConfig     `json:"engine_config"`
	ScalingConfig    ScalingConfig    `json:"scaling_config"`
}

type DeploymentConfig struct {
	// +optional
	AutoScalingConfig AutoScalingConfig `json:"auto_scaling_config"`
	// +optional
	MaxConcurrentQueries int `json:"max_concurrent_queries"`
	// +optional
	RayActorOptions RayActorOptions `json:"ray_actor_options"`
}

type AutoScalingConfig struct {
	// +kuberbuilder:default=1
	MinReplicas int `json:"min_replicas"`
	// +kuberbuilder:default=1
	InitialReplicas int `json:"initial_replicas"`
	// +kuberbuilder:default=8
	MaxReplicas int `json:"max_replicas"`

	TargetNumOngoingRequestsPerReplica float64 `json:"target_num_ongoing_requests_per_replica"`
	// +kuberbuilder:default=10.0
	MetricsIntervalSecond float64 `json:"metrics_interval_s"`
	// +kuberbuilder:default=30.0
	LookBackPeriodSecond float64 `json:"look_back_period_s"`
	// +kuberbuilder:default=1.0
	SmoothingFactor float64 `json:"smoothing_factor"`
	// +kuberbuilder:default=300.0
	DownscaleDelaySecond float64 `json:"downscale_delay_s"`
	// +kuberbuilder:default=90.0
	UpscaleDelaySecond float64 `json:"upscale_delay_s"`
}

type RayActorOptions struct {
	Resources Resources `json:"resources"`
}

type Resources struct {
	// TODO define resource
	// AcceleratorTypeL4 float64 `json:"accelerator_type_l4"`
	// AcceleratorType string `json:"accelerator_type"` // L4, V100, T4, P100, A100
}

type EngineConfig struct {
	ModelID string `json:"model_id"`
	// +optional
	HfModelID string `json:"hf_model_id,omitempty"`
	// +optional
	// +kuberbuilder:default:="VLLMEngine"
	EngineType InferenceEngineType `json:"type,omitempty"`
	// +optional
	EngineKwargs EngineKwargs `json:"engine_kwargs"`
	// +optional
	MaxTotalTokens int `json:"max_total_tokens"`
	// +optional
	Generation Generation `json:"generation"`
}

type InferenceEngineType string

var (
	VLLM      InferenceEngineType = "VLLMEngine"
	TRTLLM    InferenceEngineType = "TRTLLMEngine"
	Embedding InferenceEngineType = "EmbeddingEngine"
)

type EngineKwargs struct {
	// +kuberbuilder:default=true
	TrustRemoteCode bool `json:"trust_remote_code"`
	// +optional
	MaxNumBatchedTokens int `json:"max_num_batched_tokens"`
	// +optional
	MaxNumSeqs int `json:"max_num_seqs"`
	// +optional
	GpuMemoryUtilization float64 `json:"gpu_memory_utilization"`
}

type Generation struct {
	// +optional
	PromptFormat      PromptFormat `json:"prompt_format"`
	StoppingSequences []string     `json:"stopping_sequences,omitempty"`
}

type PromptFormat struct {
	System               string `json:"system,omitempty"`
	Assistant            string `json:"assistant,omitempty"`
	TrailingAssistant    string `json:"trailing_assistant,omitempty"`
	User                 string `json:"user,omitempty"`
	SystemInUser         bool   `json:"system_in_user,omitempty"`
	DefaultSystemMessage string `json:"default_system_message,omitempty"`
}

// ScalingConfig defines the resources assigned to each model replica. This corresponds to Ray AIR ScalingConfig.
type ScalingConfig struct {
	// +optional
	NumWorkers int `json:"num_workers,omitempty"`
	// +optional
	NumGpusPerWorker int `json:"num_gpus_per_worker,omitempty"`
	// +optional
	NumCpusPerWorker int `json:"num_cpus_per_worker,omitempty"`
	// +optional
	// +kuberbuilder:default="STRICT_PACK"
	PlacementStrategy string `json:"placement_strategy,omitempty"`
	// +optional
	ResourcesPerWorker Resources `json:"resources_per_worker"`
}

type PlacementStrategy string

var (
	StrictPack   PlacementStrategy = "STRICT_PACK"
	Pack         PlacementStrategy = "PACK"
	StrictSpread PlacementStrategy = "STRICT_SPREAD"
	Spread       PlacementStrategy = "SPREAD"
)

type ModelTemplateVersionStatus struct {
	// +optional
	Version int `json:"version,omitempty"`

	// +optional
	Conditions []Condition `json:"conditions,omitempty"`
}

// version number was assigned to templateVersion object's status.Version
var VersionAssigned condition.Cond = "assigned"
