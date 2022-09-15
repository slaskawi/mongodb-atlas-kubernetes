package kube

import (
	"k8s.io/apimachinery/pkg/types"

	"github.com/mongodb/mongodb-atlas-kubernetes/pkg/api/v1/status"
	kubecli "github.com/mongodb/mongodb-atlas-kubernetes/test/e2e/cli/kubecli"
	"github.com/mongodb/mongodb-atlas-kubernetes/test/e2e/model"
)

func GetReadyProjectStatus(data *model.TestDataProvider) func() string {
	return func() string {
		condition, _ := kubecli.GetProjectStatusCondition(data.Context, data.K8SClient, status.ReadyType, data.Resources.Namespace, data.Resources.Project.ObjectMeta.GetName())
		return condition
	}
}

func DeploymentReady(data *model.TestDataProvider) func() string {
	return func() string {
		condition, _ := kubecli.GetDeploymentStatusCondition(data.Context, data.K8SClient, status.ReadyType, data.Resources.Namespace, data.InitialDeployments[0].ObjectMeta.GetName())
		return condition
	}
}

func GetDeploymentStatus(data *model.TestDataProvider) status.AtlasDeploymentStatus {
	err := data.K8SClient.Get(data.Context, types.NamespacedName{Name: data.InitialDeployments[0].ObjectMeta.GetName(),
		Namespace: data.Resources.Namespace}, data.InitialDeployments[0])
	if err != nil {
		return status.AtlasDeploymentStatus{}
	}
	return data.InitialDeployments[0].Status
}

func GetProjectStatus(data *model.TestDataProvider) status.AtlasProjectStatus {
	err := data.K8SClient.Get(data.Context, types.NamespacedName{Name: data.Project.ObjectMeta.GetName(),
		Namespace: data.Resources.Namespace}, data.Project)
	if err != nil {
		return status.AtlasProjectStatus{}
	}
	return data.Project.Status
}
