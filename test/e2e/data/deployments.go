package data

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/mongodb/mongodb-atlas-kubernetes/pkg/api/v1/provider"

	"github.com/mongodb/mongodb-atlas-kubernetes/pkg/util/toptr"

	v1 "github.com/mongodb/mongodb-atlas-kubernetes/pkg/api/v1"
	"github.com/mongodb/mongodb-atlas-kubernetes/pkg/api/v1/common"
	"github.com/mongodb/mongodb-atlas-kubernetes/pkg/controller/customresource"
)

func CreateDeploymentWithKeepPolicy(namespace, name string) *v1.AtlasDeployment {
	deployment := CreateBasicDeployment(namespace, name)
	deployment.SetAnnotations(map[string]string{
		customresource.ResourcePolicyAnnotation: customresource.ResourcePolicyKeep,
	})
	return deployment
}

func CreateBasicDeployment(namespace, name string) *v1.AtlasDeployment {
	return &v1.AtlasDeployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: v1.AtlasDeploymentSpec{
			Project: common.ResourceRefNamespaced{
				Name:      ProjectName,
				Namespace: namespace,
			},
			DeploymentSpec: &v1.DeploymentSpec{
				Name: "cluster-basics",
				ProviderSettings: &v1.ProviderSettingsSpec{
					InstanceSizeName:    "M2", // TODO: add some constant
					ProviderName:        "TENANT",
					RegionName:          "US_EAST_1",
					BackingProviderName: "AWS",
				},
			},
		},
	}
}

func CreateDeploymentWithBackup(namespace, name string) *v1.AtlasDeployment {
	deployment := &v1.AtlasDeployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: v1.AtlasDeploymentSpec{
			Project: common.ResourceRefNamespaced{
				Name:      ProjectName,
				Namespace: namespace,
			},
			DeploymentSpec: &v1.DeploymentSpec{
				Name:                  "deployment-backup",
				ProviderBackupEnabled: toptr.MakePtr(true),
				ProviderSettings: &v1.ProviderSettingsSpec{
					InstanceSizeName: "M10", // TODO: add some constant
					ProviderName:     "AWS",
					RegionName:       "US_EAST_1",
				},
			},
		},
	}
	return deployment
}

func NewDeploymentWithBackupSpec(namespace string) v1.AtlasDeploymentSpec {
	return v1.AtlasDeploymentSpec{
		Project: common.ResourceRefNamespaced{
			Name:      ProjectName,
			Namespace: namespace,
		},
		DeploymentSpec: &v1.DeploymentSpec{
			Name:                  "deployment-backup",
			ProviderBackupEnabled: toptr.MakePtr(false),
			ProviderSettings: &v1.ProviderSettingsSpec{
				InstanceSizeName: "M20", // TODO: add some constant
				ProviderName:     "AWS",
				RegionName:       "US_EAST_1",
			},
		},
	}
}

func CreateDeploymentWithMultiregionAWS(namespace, name string) *v1.AtlasDeployment {
	return CreateDeploymentWithMultiregion(namespace, name, provider.ProviderAWS)
}

func CreateDeploymentWithMultiregionAzure(namespace, name string) *v1.AtlasDeployment {
	return CreateDeploymentWithMultiregion(namespace, name, provider.ProviderAzure)
}

func CreateDeploymentWithMultiregionGCP(namespace, name string) *v1.AtlasDeployment {
	return CreateDeploymentWithMultiregion(namespace, name, provider.ProviderGCP)
}

func CreateDeploymentWithMultiregion(namespace, name string, providerName provider.ProviderName) *v1.AtlasDeployment {
	var regions []string
	switch providerName {
	case provider.ProviderAWS:
		regions = []string{"US_EAST_1", "US_WEST_2"}
	case provider.ProviderAzure:
		regions = []string{"NORWAY_EAST", "GERMANY_NORTH"}
	case provider.ProviderGCP:
		regions = []string{"CENTRAL_US", "EASTERN_US"}
	}

	if len(regions) == 0 {
		panic("unknown provider")
	}

	return &v1.AtlasDeployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: v1.AtlasDeploymentSpec{
			Project: common.ResourceRefNamespaced{
				Name:      ProjectName,
				Namespace: namespace,
			},
			DeploymentSpec: &v1.DeploymentSpec{
				Name:                  "deployment-multiregion",
				ProviderBackupEnabled: toptr.MakePtr(true),
				ClusterType:           "REPLICASET",
				ProviderSettings: &v1.ProviderSettingsSpec{
					InstanceSizeName: "M10", // TODO: add some constant
					ProviderName:     providerName,
				},
				ReplicationSpecs: []v1.ReplicationSpec{
					{
						NumShards: toptr.MakePtr(int64(1)),
						ZoneName:  "US-Zone",
						RegionsConfig: map[string]v1.RegionsConfig{
							regions[0]: {
								AnalyticsNodes: toptr.MakePtr(int64(0)),
								ElectableNodes: toptr.MakePtr(int64(1)),
								Priority:       toptr.MakePtr(int64(6)),
								ReadOnlyNodes:  toptr.MakePtr(int64(0)),
							},
							regions[1]: {
								AnalyticsNodes: toptr.MakePtr(int64(0)),
								ElectableNodes: toptr.MakePtr(int64(2)),
								Priority:       toptr.MakePtr(int64(7)),
								ReadOnlyNodes:  toptr.MakePtr(int64(0)),
							},
						},
					},
				},
			},
		}}
}

func CreateBasicFreeDeployment(name, namespace string) *v1.AtlasDeployment {
	return &v1.AtlasDeployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: v1.AtlasDeploymentSpec{
			Project: common.ResourceRefNamespaced{
				Name:      ProjectName,
				Namespace: namespace,
			},
			DeploymentSpec: &v1.DeploymentSpec{
				Name: "cluster-basics-free",
				ProviderSettings: &v1.ProviderSettingsSpec{
					InstanceSizeName:    "M0", // TODO: add some constant
					ProviderName:        "TENANT",
					RegionName:          "US_EAST_1",
					BackingProviderName: "AWS",
				},
			},
		},
	}
}
