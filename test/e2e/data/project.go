package data

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/mongodb/mongodb-atlas-kubernetes/test/e2e/utils"

	v1 "github.com/mongodb/mongodb-atlas-kubernetes/pkg/api/v1"
	"github.com/mongodb/mongodb-atlas-kubernetes/pkg/api/v1/project"
)

const ProjectName = "my-project"

func DefaultProject(namespace string) *v1.AtlasProject {
	return &v1.AtlasProject{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ProjectName,
			Namespace: namespace,
		},
		Spec: v1.AtlasProjectSpec{
			Name: utils.RandomName("Test Atlas Operator Project"),
			ProjectIPAccessList: []project.IPAccessList{
				{
					IPAddress: "0.0.0.0/1",
					Comment:   "Everyone has access. For the test purpose only.",
				},
				{
					IPAddress: "128.0.0.0/1",
					Comment:   "Everyone has access. For the test purpose only.",
				},
			},
		},
	}
}
