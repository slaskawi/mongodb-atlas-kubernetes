package int

import (
	"context"
	"fmt"
	"os"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.mongodb.org/atlas/mongodbatlas"
	corev1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	mdbv1 "github.com/mongodb/mongodb-atlas-kubernetes/pkg/api/v1"
	"github.com/mongodb/mongodb-atlas-kubernetes/pkg/api/v1/common"
	"github.com/mongodb/mongodb-atlas-kubernetes/pkg/api/v1/status"
	"github.com/mongodb/mongodb-atlas-kubernetes/pkg/util/testutil"
	"github.com/mongodb/mongodb-atlas-kubernetes/pkg/util/toptr"
)

var _ = Describe("AtlasFederatedAuth test", Label("AtlasFederatedAuth", "federated-auth"), func() {
	var orgID string
	var testNamespace *corev1.Namespace
	var stopManager context.CancelFunc
	var connectionSecret corev1.Secret

	var originalConnectedOrgConfig *mongodbatlas.FederatedSettingsConnectedOrganization
	var originalFederationSettings *mongodbatlas.FederatedSettings
	var originalIdp *mongodbatlas.FederatedSettingsIdentityProvider

	var fedauth *mdbv1.AtlasFederatedAuth

	BeforeAll(func() {
		It("Should verify if an IDP is configured for Test organization", func() {

			By("Checking if ATLAS_ORG_ID is set", func() {
				orgID = os.Getenv("ATLAS_ORG_ID")
				Expect(orgID).NotTo(BeEmpty())
			})

			By("Getting original IDP", func() {
				originalIdp, _, err := atlasClient.FederatedSettings.Get(context.Background(), orgID)
				Expect(err).NotTo(HaveOccurred())
				Expect(originalIdp).NotTo(BeNil())
			})

			By("Checking if Federation Settings enabled for the org", func() {
				originalFederationSettings, _, err := atlasClient.FederatedSettings.Get(context.Background(), orgID)
				Expect(err).NotTo(HaveOccurred())
				Expect(originalFederationSettings).NotTo(BeNil())
			})

			var err error
			By("Getting existing org config", func() {
				originalConnectedOrgConfig, _, err = atlasClient.FederatedSettings.GetConnectedOrg(context.Background(), originalFederationSettings.ID, orgID)
				Expect(err).NotTo(HaveOccurred())
				Expect(err).NotTo(BeNil())
			})
		})

		It("Starting the operator with protection OFF", func() {
			testNamespace, stopManager = prepareControllers(false)
			Expect(testNamespace).ToNot(BeNil())
			Expect(stopManager).ToNot(BeNil())
		})

		It("Creating project connection secret", func() {
			connectionSecret = buildConnectionSecret(fmt.Sprintf("%s-atlas-key", testNamespace.Name))
			Expect(k8sClient.Create(context.Background(), &connectionSecret)).To(Succeed())
		})
	})

	AfterAll(func() {
		It("Should delete connection secret", func() {
			Expect(k8sClient.Delete(context.Background(), &connectionSecret)).To(Succeed())
		})

		It("Should stop the operator", func() {
			stopManager()
			err := k8sClient.Delete(context.Background(), testNamespace)
			Expect(err).ToNot(HaveOccurred())
		})
	})

	It("Should be able to update existing Organization's federations settings", func() {
		By("Creating a FederatedAuthConfig resource", func() {
			roles := []mdbv1.RoleMapping{}

			for i := range originalConnectedOrgConfig.RoleMappings {
				atlasRole := *(originalConnectedOrgConfig.RoleMappings[i])
				newRole := mdbv1.RoleMapping{
					ExternalGroupName: atlasRole.ExternalGroupName,
					RoleAssignments:   []mdbv1.RoleAssignment{},
				}

				for j := range atlasRole.RoleAssignments {
					atlasRS := atlasRole.RoleAssignments[j]
					project, _, err := atlasClient.Projects.GetOneProject(context.Background(), atlasRS.GroupID)
					Expect(err).NotTo(HaveOccurred())
					Expect(project).NotTo(BeNil())

					newRS := mdbv1.RoleAssignment{
						ProjectName: project.Name,
						Role:        atlasRS.Role,
					}
					newRole.RoleAssignments = append(newRole.RoleAssignments, newRS)
				}
				roles = append(roles, newRole)
			}

			fedauth = &mdbv1.AtlasFederatedAuth{
				TypeMeta: metav1.TypeMeta{
					Kind:       "AtlasFederatedAuth",
					APIVersion: "atlas.mongodb.com/v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "testAuth",
					Namespace: testNamespace.Name,
				},
				Spec: mdbv1.AtlasFederatedAuthSpec{
					Enabled: toptr.MakePtr(true),
					ConnectionSecretRef: common.ResourceRefNamespaced{
						Name:      connectionSecret.Name,
						Namespace: connectionSecret.Namespace,
					},
					DomainAllowList:          originalConnectedOrgConfig.DomainAllowList,
					DomainRestrictionEnabled: originalConnectedOrgConfig.DomainRestrictionEnabled,
					SSODebugEnabled:          originalIdp.SsoDebugEnabled,
					PostAuthRoleGrants:       originalConnectedOrgConfig.PostAuthRoleGrants,
					RoleMappings:             roles,
				},
				Status: status.AtlasFederatedAuthStatus{},
			}

			Expect(k8sClient.Create(context.Background(), fedauth)).NotTo(HaveOccurred())
		})

		By("Making sure Federated Auth is ready", func() {
			Eventually(func(g Gomega) bool {
				return testutil.CheckCondition(k8sClient, fedauth, status.TrueCondition(status.ReadyType), validateDeploymentUpdatingFunc(g))
			}).WithTimeout(30 * time.Minute).WithPolling(PollingInterval).Should(BeTrue())
		})
	})
})
