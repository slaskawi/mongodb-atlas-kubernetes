// different ways to deploy operator
package deploy

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mongodb/mongodb-atlas-kubernetes/test/e2e/actions/kube"

	"k8s.io/apimachinery/pkg/types"

	mdbv1 "github.com/mongodb/mongodb-atlas-kubernetes/pkg/api/v1"
	"github.com/mongodb/mongodb-atlas-kubernetes/test/e2e/api/atlas"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	kubecli "github.com/mongodb/mongodb-atlas-kubernetes/test/e2e/cli/kubecli"
	"github.com/mongodb/mongodb-atlas-kubernetes/test/e2e/cli/kustomize"
	"github.com/mongodb/mongodb-atlas-kubernetes/test/e2e/config"
	"github.com/mongodb/mongodb-atlas-kubernetes/test/e2e/model"
	"github.com/mongodb/mongodb-atlas-kubernetes/test/e2e/utils"
)

// prepareNamespaceOperatorResources create copy of `/deploy/namespaced` folder with kustomization file for overriding namespace
func prepareNamespaceOperatorResources(input model.UserInputs) {
	fullPath := input.GetOperatorFolder()
	os.Mkdir(fullPath, os.ModePerm)
	utils.CopyFile(config.DefaultNamespacedCRDConfig, filepath.Join(fullPath, "crds.yaml"))
	utils.CopyFile(config.DefaultNamespacedOperatorConfig, filepath.Join(fullPath, "namespaced-config.yaml"))
	data := []byte(
		"namespace: " + input.Namespace + "\n" +
			"resources:" + "\n" +
			"- crds.yaml" + "\n" +
			"- namespaced-config.yaml",
	)
	utils.SaveToFile(filepath.Join(fullPath, "kustomization.yaml"), data)
}

// CopyKustomizeNamespaceOperator create copy of `/deploy/namespaced` folder with kustomization file for overriding namespace
func prepareWideOperatorResources(input model.UserInputs) {
	fullPath := input.GetOperatorFolder()
	os.Mkdir(fullPath, os.ModePerm)
	utils.CopyFile(config.DefaultClusterWideCRDConfig, filepath.Join(fullPath, "crds.yaml"))
	utils.CopyFile(config.DefaultClusterWideOperatorConfig, filepath.Join(fullPath, "clusterwide-config.yaml"))
}

// CopyKustomizeNamespaceOperator create copy of `/deploy/namespaced` folder with kustomization file for overriding namespace
func prepareMultiNamespaceOperatorResources(input model.UserInputs, watchedNamespaces []string) {
	fullPath := input.GetOperatorFolder()
	err := os.Mkdir(fullPath, os.ModePerm)
	Expect(err).ShouldNot(HaveOccurred())
	utils.CopyFile(config.DefaultClusterWideCRDConfig, filepath.Join(fullPath, "crds.yaml"))
	utils.CopyFile(config.DefaultClusterWideOperatorConfig, filepath.Join(fullPath, "multinamespace-config.yaml"))
	namespaces := strings.Join(watchedNamespaces, ",")
	patchWatch := []byte(
		"apiVersion: apps/v1\n" +
			"kind: Deployment\n" +
			"metadata:\n" +
			"  name: mongodb-atlas-operator\n" +
			"spec:\n" +
			"  template:\n" +
			"    spec:\n" +
			"      containers:\n" +
			"      - name: manager\n" +
			"        env:\n" +
			"        - name: WATCH_NAMESPACE\n" +
			"          value: \"" + namespaces + "\"",
	)
	err = utils.SaveToFile(filepath.Join(fullPath, "patch.yaml"), patchWatch)
	Expect(err).ShouldNot(HaveOccurred())
	kustomization := []byte(
		"resources:\n" +
			"- multinamespace-config.yaml\n" +
			"patches:\n" +
			"- path: patch.yaml\n" +
			"  target:\n" +
			"    group: apps\n" +
			"    version: v1\n" +
			"    kind: Deployment\n" +
			"    name: mongodb-atlas-operator",
	)
	err = utils.SaveToFile(filepath.Join(fullPath, "kustomization.yaml"), kustomization)
	Expect(err).ShouldNot(HaveOccurred())
}

func NamespacedOperator(data *model.TestDataProvider) {
	prepareNamespaceOperatorResources(data.Resources)
	By("Deploy namespaced Operator\n", func() {
		kubecli.Apply("-k", data.Resources.GetOperatorFolder())
		Eventually(
			func(g Gomega) string {
				status, err := kubecli.GetPodStatus(data.Context, data.K8SClient, data.Resources.Namespace)
				g.Expect(err).ShouldNot(HaveOccurred())
				return status
			},
			"5m", "3s",
		).Should(Equal("Running"), "The operator should successfully run")
	})
}

func ClusterWideOperator(data *model.TestDataProvider) {
	prepareWideOperatorResources(data.Resources)
	By("Deploy clusterwide Operator \n", func() {
		kubecli.Apply("-k", data.Resources.GetOperatorFolder())
		Eventually(
			func() string {
				status, err := kubecli.GetPodStatus(data.Context, data.K8SClient, config.DefaultOperatorNS)
				if err != nil {
					return ""
				}
				return status
			},
			"5m", "3s",
		).Should(Equal("Running"), "The operator should successfully run")
	})
}

func MultiNamespaceOperator(data *model.TestDataProvider, watchNamespace []string) {
	prepareMultiNamespaceOperatorResources(data.Resources, watchNamespace)
	By("Deploy multinamespaced Operator \n", func() {
		kustomOperatorPath := data.Resources.GetOperatorFolder() + "/final.yaml"
		utils.SaveToFile(kustomOperatorPath, kustomize.Build(data.Resources.GetOperatorFolder()))
		kubecli.Apply(kustomOperatorPath)
		Eventually(
			func() string {
				status, err := kubecli.GetPodStatus(data.Context, data.K8SClient, config.DefaultOperatorNS)
				if err != nil {
					return ""
				}
				return status
			},
			"5m", "3s",
		).Should(Equal("Running"), "The operator should successfully run")
	})
}

func DeleteProject(testData *model.TestDataProvider) {
	By("Delete Project", func() {
		projectId := testData.Project.Status.ID
		Expect(testData.K8SClient.Get(testData.Context, types.NamespacedName{Name: testData.Project.Name, Namespace: testData.Project.Namespace}, testData.Project)).Should(Succeed(), "Get project failed")
		Expect(testData.K8SClient.Delete(testData.Context, testData.Project)).Should(Succeed(), "Delete project failed")
		aClient := atlas.GetClientOrFail()
		Eventually(func() bool {
			return aClient.IsProjectExists(projectId)
		}).WithTimeout(5*time.Minute).WithPolling(20*time.Second).Should(BeTrue(), "Project was not deleted in Atlas")
	})
}

func DeleteUsers(testData *model.TestDataProvider) {
	By("Delete Users", func() {
		for _, user := range testData.Users {
			Expect(testData.K8SClient.Get(testData.Context, types.NamespacedName{Name: user.Name, Namespace: user.Namespace}, user)).Should(Succeed(), "Get user failed")
			Expect(testData.K8SClient.Delete(testData.Context, user)).Should(Succeed(), "Delete user failed")
		}
	})
}

func DeleteInitialDeployments(testData *model.TestDataProvider) {
	By("Delete initial deployments", func() {
		for _, deployment := range testData.InitialDeployments {
			projectId := testData.Project.Status.ID
			deploymentName := deployment.Spec.DeploymentSpec.Name
			Expect(testData.K8SClient.Get(testData.Context, types.NamespacedName{Name: deployment.Name,
				Namespace: testData.Resources.Namespace}, deployment)).Should(Succeed(), "Get deployment failed")
			Expect(testData.K8SClient.Delete(testData.Context, deployment)).Should(Succeed(), "Deployment %s was not deleted", deployment.Name)
			aClient := atlas.GetClientOrFail()
			Eventually(func() bool {
				return aClient.IsDeploymentExist(projectId, deploymentName)
			}).WithTimeout(5*time.Minute).WithPolling(20*time.Second).Should(BeFalse(), "Deployment should be deleted in Atlas")
		}
	})
}

func Project(testData *model.TestDataProvider) {
	if testData.Project.GetNamespace() == "" {
		testData.Project.Namespace = testData.Resources.Namespace
	}
	By(fmt.Sprintf("Deploy Project %s", testData.Project.GetName()), func() {
		err := testData.K8SClient.Create(testData.Context, testData.Project)
		Expect(err).ShouldNot(HaveOccurred(), "Project %s was not created", testData.Project.GetName())
		Eventually(kube.GetReadyProjectStatus(testData)).WithTimeout(5*time.Minute).WithPolling(20*time.Second).
			Should(Not(Equal("False")), "Project %s should be ready", testData.Project.GetName())
	})
	By(fmt.Sprintf("Wait for Project %s", testData.Project.GetName()), func() {
		Eventually(func() bool {
			statuses := kube.GetProjectStatus(testData)
			return statuses.ID != ""
		}, 5*time.Minute, 5*time.Second).Should(BeTrue(), "Project %s is not ready", kube.GetProjectStatus(testData))
	})
}

func InitialDeployments(testData *model.TestDataProvider) {
	By("Deploy Initial Deployments", func() {
		for _, deployment := range testData.InitialDeployments {
			if deployment.Namespace == "" {
				deployment.Namespace = testData.Resources.Namespace
				deployment.Spec.Project.Namespace = testData.Resources.Namespace // TODO: remove after fix
			}
			err := testData.K8SClient.Create(testData.Context, deployment)
			Expect(err).ShouldNot(HaveOccurred(), fmt.Sprintf("Deployment was not created: %v", deployment))

			deploymentForCheck := &mdbv1.AtlasDeployment{}
			Eventually(func(g Gomega) {
				err = testData.K8SClient.Get(testData.Context, types.NamespacedName{Name: deployment.GetName(), Namespace: deployment.GetNamespace()}, deploymentForCheck)
				Expect(err).Should(BeNil(), fmt.Sprintf("Deployment not found: %v", deploymentForCheck))
				By(fmt.Sprintf("Deployment %s status: %v", deploymentForCheck.ObjectMeta.Name, deploymentForCheck.Status))
				g.Eventually(kube.DeploymentReady(testData)).WithTimeout(60*time.Minute).WithPolling(20*time.Second).Should(Equal("True"), "Deployment should be ready")
			}, time.Minute*60, time.Second*5).Should(Succeed(), "Deployment was not created")
		}
	})
}

func Users(testData *model.TestDataProvider) {
	By("Deploy Users", func() {
		for _, user := range testData.Users {
			if user.Namespace == "" { // TODO: remove after fix
				user.Namespace = testData.Resources.Namespace
			}
			if user.Spec.PasswordSecret != nil {
				secret := utils.UserSecretPassword()
				Expect(kubecli.CreateUserSecret(testData.Context, testData.K8SClient, secret,
					user.Spec.PasswordSecret.Name, testData.Resources.Namespace)).Should(Succeed(),
					"Create user secret failed")
				// TODO: remake namespaces after fix
			}
			err := testData.K8SClient.Create(testData.Context, user)
			Expect(err).ShouldNot(HaveOccurred(), fmt.Sprintf("User was not created: %v", user))
		}
	})
}
