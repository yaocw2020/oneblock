# Model Management and Serving

## Summary
Deploy a custom LLM on the 1Block.AI platform is a key component for in-house machine learning works. This enhancement is to provide a way to allow users create, track and mange version of models, and to deploy the model and business logic as a service.

### Related Issues

https://github.com/oneblock-ai/oneblock/issues/25

## Motivation
- ModelTemplate provides a way to allow users to create, define, and manage LLMs with different configurations and versions.
- MLServe provides a way to allow users to deploy machine learning models and business logic as a service to a ML cluster(RayCluster). 
  - It is a scalable model serving tool for building online inference APIs.
  - It is a single toolkit to serve everything from deep learning models built with frameworks like PyTorch, TensorFlow, and Keras, to Scikit-Learn models, to arbitrary Python business logic.
  - It has several features and performance optimizations for serving Large Language Models such as response streaming, dynamic request batching, multi-node/multi-GPU serving, etc.

### Goals
- Provide a list of built-in OSS models, e.g., mistral-7b, llama-7b, falcon-7b, gemma-7b, etc.
- Allow users to create, track and manage versions of models.
- Allow one-click deployment using a model template.
- Provide a way to allow users to deploy ML models and business logic as a service.

### Non-goals [optional]

N/A

## Proposal

### User Stories

#### Story 1
As a user, I want to deploy a pre-configured LLM template as an inference service with one-click deployment.

#### Story 2
As a user, I want to create a new large language model by providing a hugging face model ID(e.g., google/gemma-7b).

### User Experience In Detail

#### Story 1
As a user, I want to deploy the built-in LLM template as an inference service API with one-click .
1. Select a built-in LLM template from the model template listing page.
2. Click the "Deploy" button to deploy the model as a service.
3. Fill in the serve form, e.g., name, namespace, MLCluster, etc.
4. Click the "Save" button to deploy the model as an inference service API.
```YAML
apiVersion: ml.oneblock.ai/v1alpha1
kind: MLServe
metadata:
  name: mistral-7b-llm
  namespace: default
spec:
  applications:
  - name: mistral-7b-llm
    route_prefix: /
    import_path: rayllm.backend:router_application
  modelRef:
    name: mistral-7b-instruct
    namespace: default
  mlClusterRef: # use existing or create a new ray cluster
    name: public-ml-cluster
    namespace: oneblock-public
```

#### Story 2
As a user, I want to create a new large language model by providing a hugging face model ID(e.g., google/gemma-7b).
1. Go to the model template page.
2. Click the "Create" button to create a new model template.
3. Fill in the model template form, e.g., name, namespace, hf_model_id, max_total_tokens, engine_kwargs, etc.
4. Click the "Save" button to create a new model template.

Example of ModelTemplate YAML:
```YAML
apiVersion: ml.oneblock.ai/v1alpha1
kind: ModelTemplate
metadata:
  name: mistral-7b-instruct
  namespace: default
spec:
  description: An open-source mistral-7b-instruct model.
---
apiVersion: ml.oneblock.ai/v1alpha1
kind: ModelTemplateVersion
metadata:
  name: mistral-7b-instruct-v1
  namespace: default
spec:
  deployment_config:
    # This corresponds to Ray Serve settings, as generated with
    # `serve build`.
    autoscaling_config:
      min_replicas: 1
      initial_replicas: 1
      max_replicas: 1
      target_num_ongoing_requests_per_replica: 24
      metrics_interval_s: 10.0
      look_back_period_s: 30.0
      smoothing_factor: 0.5
      downscale_delay_s: 300.0
      upscale_delay_s: 90.0
    max_concurrent_queries: 64
    ray_actor_options:
      # Resources assigned to each model deployment. The deployment will be
      # initialized first, and then start prediction workers which actually hold the model.
      resources:
        accelerator_type:<GPU_type>: 0.01 # auto configured by MLServe's mlClusterRef
        # accelerator_type_cpu: 0.01
  engine_config:
    # Model id - this is a RayLLM id
    model_id: "mistralai/Mistral-7B-Instruct-v0.2"
    # Id of the model on Hugging Face Hub. Can also be a disk path. Defaults to model_id if not specified.
    hf_model_id: "mistralai/Mistral-7B-Instruct-v0.2"
    type: VLLMEngine # TRTLLMEngine VLLMEngine
    # LLM engine keyword arguments passed when constructing the model.
    engine_kwargs:
      trust_remote_code: true
      max_num_batched_tokens: 2048
      max_num_seqs: 64
      gpu_memory_utilization: 0.9
      # dtype: half
    max_total_tokens: 2048
    # Optional Ray Runtime Environment configuration. See Ray documentation for more details.
    # Add dependent libraries, environment variables, etc.
    runtime_env:
    # env_vars:
    # YOUR_ENV_VAR: "your_value"
    # Optional configuration for loading the model from S3 instead of Hugging Face Hub. You can use this to speed up downloads or load models not on Hugging Face Hub.
    # s3_mirror_config:
    #   bucket_uri: s3://large-dl-models-mirror/models/
    generation:
      # Format to convert user API input into prompts to feed into the LLM engine. {instruction} refers to user-supplied input.
      prompt_format:
        system: "{instruction}\n"  # System message. Will default to default_system_message
        assistant: "### Response:\n{instruction}\n"  # Past assistant message. Used in chat completions API.
        trailing_assistant: "### Response:\n"  # New assistant message. After this point, model will generate tokens.
        user: "### Instruction:\n{instruction}\n"  # User message.
        default_system_message: "Below is an instruction that describes a task. Write a response that appropriately completes the request."  # Default system message.
        system_in_user: false  # Whether the system prompt is inside the user prompt. If true, the user field should include '{system}'
        add_system_tags_even_if_message_is_empty: false  # Whether to include the system tags even if the user message is empty.
        strip_whitespace: false  # Whether to automaticall strip whitespace from left and right of user supplied messages for chat completions
      # Stopping sequences. The generation will stop when it encounters any of the sequences, or the tokenizer EOS token.
      # Those can be strings, integers (token ids) or lists of integers.
      # Stopping sequences supplied by the user in a request will be appended to this.
      stopping_sequences: ["### Response:", "### End"]
  # Resources assigned to each model replica. This corresponds to Ray AIR ScalingConfig.
  scaling_config:
    # If using multiple GPUs set num_gpus_per_worker to be 1 and then set num_workers to be the number of GPUs you want to use.
    num_workers: 1
    num_gpus_per_worker: 1
    num_cpus_per_worker: 2
    placement_strategy: "STRICT_PACK"
    resources_per_worker:
      accelerator_type:<GPU_type>: 0.01 # auto configured by MLServe's mlClusterRef

```

### API changes
Three new CRDs will be added to the API:
- `ModelTemplate`: A model template is a blueprint for creating a new model.
	- `ModelTemplateVersion`: Specify the details of the model template with an auto-increment version number.
- `MLServe`: A MLServe is a model serving configuration for deploying a model as a service.

## Design

### Pre-requisite
- Worker node with GPU accelerator is required for LLM serving.

### Implementation Overview

- Users can define the ModelTemplate by following the [RayServe](https://docs.ray.io/en/latest/serve/index.html) and [vLLM](https://docs.vllm.ai/en/latest/index.html) config spec.
- 

Overview on how the enhancement will be implemented.

### Test plan
User can test the model serving by using the following steps:
1. Create or select an existing model template.
2. Deploy the model as a service.
3. Wait for the serve to be ready.
4. Test the model inference API using the OpenAI python SDK. e.g.,
```python
import openai

query = "Once upon a time,"

client = openai.OpenAI(
    base_url = "https://public-ml-cluster.oneblock-public:8000/v1",
    api_key = "not_a_real_api_key"
)
# Note: not all arguments are currently supported and will be ignored by the backend.
chat_completion = client.chat.completions.create(
    model="mistralai/Mistral-7B-Instruct-v0.2",
    messages=[{"role": "system", "content": "You are a helpful assistant."}, 
              {"role": "user", "content": query}],
    temperature=0.1,
)
print(chat_completion.choices[0].message.content)
```

### Upgrade strategy

N/A

## Note [optional]

Additional nodes.
