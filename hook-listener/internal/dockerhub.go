package internal

import (
	"encoding/json"
	"net/http"
)

type DockerHubEvent struct {
	CallbackUrl string     `json:"callback_url"`
	Repository  Repository `json:"repository"`
	PushedData  PushedData `json:"push_data"`
}

type Repository struct {
	RepoName  string `json:"repo_name"`
	RepoUrl   string `json:"repo_url"`
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

type PushedData struct {
	Tag      string `json:"tag"`
	PushedAt uint64 `json:"pushed_at"`
}

type DockerHubController struct {
	proxy *KubernetesProxy
}

func NewDockerHubController(proxy *KubernetesProxy) *DockerHubController {
	return &DockerHubController{proxy: proxy}
}

func (handler DockerHubController) HandleDockerHubHookRequest(w http.ResponseWriter, req *http.Request) {
	var event DockerHubEvent
	if err := json.NewDecoder(req.Body).Decode(&event); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	deployments, err := handler.proxy.listDeployment(req.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	for _, deployment := range deployments.Items {
		if isTurbineApp(deployment) && containsContainer(deployment, event.Repository.RepoName, event.PushedData.Tag) {
			err := handler.proxy.restartDeployment(req.Context(), deployment.Name)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
		}
	}
}
