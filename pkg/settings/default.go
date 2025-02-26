package settings

import (
	"encoding/json"

	"github.com/sirupsen/logrus"
)

type Image struct {
	ContainerImage string `json:"containerImage,omitempty"`
	Description    string `json:"description,omitempty"`
	Default        bool   `json:"default,omitempty"`
}

const defaultImgVersion = "v1.8.0"

// SetDefaultNotebookImages set default notebook images
// resources please refer to: https://www.kubeflow.org/docs/components/notebooks/container-images/
func setDefaultNotebookImages() string {
	defaultImgs := map[string][]Image{
		"jupyter": {
			{
				ContainerImage: "kubeflownotebookswg/jupyter-scipy:" + defaultImgVersion,
				Description:    "JupyterLab + PyTorch",
				Default:        true,
			},
			{
				ContainerImage: "kubeflownotebookswg/jupyter-pytorch:" + defaultImgVersion,
				Description:    "JupyterLab + PyTorch",
			},
			{
				ContainerImage: "kubeflownotebookswg/jupyter-pytorch-full:" + defaultImgVersion,
				Description:    "JupyterLab + PyTorch + Common Packages",
			},
			{
				ContainerImage: "kubeflownotebookswg/jupyter-pytorch-cuda:" + defaultImgVersion,
				Description:    "JupyterLab + PyTorch + CUDA",
			},
			{
				ContainerImage: "kubeflownotebookswg/jupyter-pytorch-cuda-full:" + defaultImgVersion,
				Description:    "JupyterLab + PyTorch + CUDA + Common Packages",
			},
			{
				ContainerImage: "kubeflownotebookswg/jupyter-tensorflow:" + defaultImgVersion,
				Description:    "JupyterLab + PyTorch",
			},
			{
				ContainerImage: "kubeflownotebookswg/jupyter-tensorflow-full:" + defaultImgVersion,
				Description:    "JupyterLab + PyTorch + Common Packages",
			},
			{
				ContainerImage: "kubeflownotebookswg/jupyter-tensorflow-cuda:" + defaultImgVersion,
				Description:    "JupyterLab + PyTorch + CUDA",
			},
			{
				ContainerImage: "kubeflownotebookswg/jupyter-tensorflow-cuda-full:" + defaultImgVersion,
				Description:    "JupyterLab + PyTorch + CUDA + Common Packages",
			},
		},
		"code-server": {
			{
				ContainerImage: "kubeflownotebookswg/codeserver-python:" + defaultImgVersion,
				Description:    "Visual Studio Code + Conda Python",
				Default:        true,
			},
		},
		"rstudio": {
			{
				ContainerImage: "kubeflownotebookswg/rstudio-tidyverse:" + defaultImgVersion,
				Description:    "RStudio + Tidyverse",
				Default:        true,
			},
		},
	}
	stringImg, err := json.Marshal(defaultImgs)
	if err != nil {
		logrus.Fatal(err)
	}
	return string(stringImg)
}
