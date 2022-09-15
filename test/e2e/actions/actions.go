// `actions` additional functions which accept testDataProvider struct and could be used as additional acctions in the tests since they all typical

package actions

import (
	"fmt"
	"time"

	"github.com/mongodb/mongodb-atlas-kubernetes/test/e2e/actions/kube"

	v1 "github.com/mongodb/mongodb-atlas-kubernetes/pkg/api/v1"

	"k8s.io/apimachinery/pkg/types"

	"github.com/mongodb/mongodb-atlas-kubernetes/test/e2e/api/atlas"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mongodb/mongodb-atlas-kubernetes/test/e2e/model"
)

func UpdateFirstDeploymentSpec(spec v1.AtlasDeploymentSpec) func(data *model.TestDataProvider) {
	return func(data *model.TestDataProvider) {
		if len(data.InitialDeployments) == 0 {
			Fail("No Deployments found")
		}
		By(fmt.Sprintf("Update Deployment %s", data.InitialDeployments[0].GetName()), func() {
			Expect(data.K8SClient.Get(data.Context, types.NamespacedName{Name: data.InitialDeployments[0].ObjectMeta.GetName(),
				Namespace: data.Resources.Namespace}, data.InitialDeployments[0])).To(Succeed())
			data.InitialDeployments[0].Spec = spec
			Expect(data.K8SClient.Update(data.Context, data.InitialDeployments[0])).To(Succeed())
			Eventually(kube.DeploymentReady(data)).WithTimeout(15*time.Minute).WithPolling(20*time.Second).Should(Equal("True"),
				fmt.Sprintf("Deployment is not ready. Status: %v", kube.GetDeploymentStatus(data)))
		})
	}
}

func activateFirstDeployment(data *model.TestDataProvider, paused bool) {
	By("Activate Deployment", func() {
		Expect(data.K8SClient.Get(data.Context,
			types.NamespacedName{Name: data.InitialDeployments[0].GetName(),
				Namespace: data.Resources.Namespace},
			data.InitialDeployments[0])).Should(Succeed())
		updateSpec := data.InitialDeployments[0].Spec
		updateSpec.DeploymentSpec.Paused = &paused // TODO: for advanced too
		UpdateFirstDeploymentSpec(updateSpec)(data)
	})
	By("Check additional Deployment field `paused`\n", func() {
		aClient := atlas.GetClientOrFail()
		Eventually(func(g Gomega) {
			uDeployment := aClient.GetDeployment(data.Project.ID(), data.InitialDeployments[0].Spec.DeploymentSpec.Name) // TODO for advanced too
			g.Expect(uDeployment.Paused).Should(Equal(data.InitialDeployments[0].Spec.DeploymentSpec.Paused))
		}).WithTimeout(5 * time.Minute).WithPolling(20 * time.Second).Should(Succeed())
	})
}

func SuspendDeployment(data *model.TestDataProvider) {
	activateFirstDeployment(data, true)
}

func ReactivateDeployment(data *model.TestDataProvider) {
	activateFirstDeployment(data, false)
}

func DeleteFirstUser(data *model.TestDataProvider) {
	By("User can delete Database User", func() {
		// data.Resources.ProjectID = kube.GetProjectResource(data.Resources.Namespace, data.Resources.GetAtlasProjectFullKubeName()).Status.ID
		// since it is could be several users, we should
		// - delete k8s resource
		// - delete one user from the list,
		// - check Atlas doesn't have the initial user and have the rest
		By("Delete k8s resources")
		if len(data.Users) == 0 {
			Fail("No users to delete")
		}
		Expect(data.K8SClient.Get(data.Context, types.NamespacedName{Name: data.Users[0].Name, Namespace: data.Users[0].Namespace}, data.Users[0])).Should(Succeed())
		Expect(data.K8SClient.Delete(data.Context, data.Users[0])).Should(Succeed())
		Eventually(func(g Gomega) {
			aClient := atlas.GetClientOrFail()
			user, err := aClient.GetDBUser(data.Users[0].Spec.DatabaseName, data.Users[0].Spec.Username, data.Project.ID())
			g.Expect(err).To(BeNil())
			g.Expect(user).To(BeNil())
		}).WithTimeout(2 * time.Minute).WithPolling(20 * time.Second).Should(Succeed())

		// the rest users should be still there
		data.Users = data.Users[1:]
	})
}
