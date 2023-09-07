package v1

import (
	"context"
	"net/http"
	"testing"

	"github.com/go-test/deep"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/atlas/mongodbatlas"

	"github.com/mongodb/mongodb-atlas-kubernetes/pkg/api/v1/common"
	"github.com/mongodb/mongodb-atlas-kubernetes/pkg/controller/atlas"
	"github.com/mongodb/mongodb-atlas-kubernetes/pkg/util/toptr"
)

type projectClient struct {
	GetProject func() (*mongodbatlas.Project, *mongodbatlas.Response, error)
}

func (projectClient) GetAllProjects(_ context.Context, _ *mongodbatlas.ListOptions) (*mongodbatlas.Projects, *mongodbatlas.Response, error) {
	panic("not implemented")
}
func (projectClient) GetOneProject(_ context.Context, _ string) (*mongodbatlas.Project, *mongodbatlas.Response, error) {
	panic("not implemented")
}
func (pc *projectClient) GetOneProjectByName(_ context.Context, _ string) (*mongodbatlas.Project, *mongodbatlas.Response, error) {
	return pc.GetProject()
}
func (projectClient) Create(_ context.Context, _ *mongodbatlas.Project, _ *mongodbatlas.CreateProjectOptions) (*mongodbatlas.Project, *mongodbatlas.Response, error) {
	panic("not implemented")
}
func (projectClient) Update(_ context.Context, _ string, _ *mongodbatlas.ProjectUpdateRequest) (*mongodbatlas.Project, *mongodbatlas.Response, error) {
	panic("not implemented")
}
func (projectClient) Delete(_ context.Context, _ string) (*mongodbatlas.Response, error) {
	panic("not implemented")
}
func (projectClient) GetProjectTeamsAssigned(_ context.Context, _ string) (*mongodbatlas.TeamsAssigned, *mongodbatlas.Response, error) {
	panic("not implemented")
}
func (projectClient) AddTeamsToProject(_ context.Context, _ string, _ []*mongodbatlas.ProjectTeam) (*mongodbatlas.TeamsAssigned, *mongodbatlas.Response, error) {
	panic("not implemented")
}
func (projectClient) RemoveUserFromProject(_ context.Context, _ string, _ string) (*mongodbatlas.Response, error) {
	panic("not implemented")
}
func (projectClient) Invitations(_ context.Context, _ string, _ *mongodbatlas.InvitationOptions) ([]*mongodbatlas.Invitation, *mongodbatlas.Response, error) {
	panic("not implemented")
}
func (projectClient) Invitation(_ context.Context, _ string, _ string) (*mongodbatlas.Invitation, *mongodbatlas.Response, error) {
	panic("not implemented")
}
func (projectClient) InviteUser(_ context.Context, _ string, _ *mongodbatlas.Invitation) (*mongodbatlas.Invitation, *mongodbatlas.Response, error) {
	panic("not implemented")
}
func (projectClient) UpdateInvitation(_ context.Context, _ string, _ *mongodbatlas.Invitation) (*mongodbatlas.Invitation, *mongodbatlas.Response, error) {
	panic("not implemented")
}
func (projectClient) UpdateInvitationByID(_ context.Context, _ string, _ string, _ *mongodbatlas.Invitation) (*mongodbatlas.Invitation, *mongodbatlas.Response, error) {
	panic("not implemented")
}
func (projectClient) DeleteInvitation(_ context.Context, _ string, _ string) (*mongodbatlas.Response, error) {
	panic("not implemented")
}
func (projectClient) GetProjectSettings(_ context.Context, _ string) (*mongodbatlas.ProjectSettings, *mongodbatlas.Response, error) {
	panic("not implemented")
}
func (projectClient) UpdateProjectSettings(_ context.Context, _ string, _ *mongodbatlas.ProjectSettings) (*mongodbatlas.ProjectSettings, *mongodbatlas.Response, error) {
	panic("not implemented")
}

func Test_FederatedAuthSpec_ToAtlas(t *testing.T) {
	t.Run("Can convert valid spec to Atlas", func(t *testing.T) {
		orgID := "test-org"
		idpID := "test-idp"

		projectID := "test-project-id"

		pc := &projectClient{
			GetProject: func() (*mongodbatlas.Project, *mongodbatlas.Response, error) {
				return &mongodbatlas.Project{
						ID:    projectID,
						OrgID: orgID,
					}, &mongodbatlas.Response{
						Response: &http.Response{
							Status:     "",
							StatusCode: http.StatusOK,
						},
					}, nil
			},
		}

		spec := &AtlasFederatedAuthSpec{
			Enabled:                  toptr.MakePtr(true),
			ConnectionSecretRef:      common.ResourceRefNamespaced{},
			DomainAllowList:          []string{"test.com"},
			DomainRestrictionEnabled: toptr.MakePtr(true),
			SSODebugEnabled:          toptr.MakePtr(true),
			PostAuthRoleGrants:       []string{"role-3", "role-4"},
			RoleMappings: []RoleMapping{
				{
					ExternalGroupName: "test-group",
					RoleAssignments: []RoleAssignment{
						{
							ProjectName: "test-project",
							Role:        "test-role",
						},
					},
				},
			},
		}

		result, err := spec.ToAtlas(orgID, idpID, &mongodbatlas.Client{
			Projects: pc,
		})

		assert.NoError(t, err, "ToAtlas() failed")
		assert.NotNil(t, result, "ToAtlas() result is nil")

		expected := &mongodbatlas.FederatedSettingsConnectedOrganization{
			DomainAllowList:          spec.DomainAllowList,
			DomainRestrictionEnabled: spec.DomainRestrictionEnabled,
			IdentityProviderID:       idpID,
			OrgID:                    orgID,
			PostAuthRoleGrants:       spec.PostAuthRoleGrants,
			RoleMappings: []*mongodbatlas.RoleMappings{
				{
					ExternalGroupName: spec.RoleMappings[0].ExternalGroupName,
					ID:                idpID,
					RoleAssignments: []*mongodbatlas.RoleAssignments{
						{
							GroupID: projectID,
							OrgID:   "",
							Role:    spec.RoleMappings[0].RoleAssignments[0].Role,
						},
					},
				},
			},
			UserConflicts: nil,
		}

		diff := deep.Equal(expected, result)
		assert.Nil(t, diff, diff)
	})

	t.Run("Should return an error when project is not available", func(t *testing.T) {
		orgID := "test-org"
		idpID := "test-idp"

		pc := &projectClient{
			GetProject: func() (*mongodbatlas.Project, *mongodbatlas.Response, error) {
				return nil, &mongodbatlas.Response{
					Response: &http.Response{
						Status:     atlas.NotInGroup,
						StatusCode: http.StatusNotFound,
					},
				}, nil
			},
		}

		spec := &AtlasFederatedAuthSpec{
			Enabled:                  toptr.MakePtr(true),
			ConnectionSecretRef:      common.ResourceRefNamespaced{},
			DomainAllowList:          []string{"test.com"},
			DomainRestrictionEnabled: toptr.MakePtr(true),
			SSODebugEnabled:          toptr.MakePtr(true),
			PostAuthRoleGrants:       []string{"role-3", "role-4"},
			RoleMappings: []RoleMapping{
				{
					ExternalGroupName: "test-group",
					RoleAssignments: []RoleAssignment{
						{
							ProjectName: "test-project",
							Role:        "test-role",
						},
					},
				},
			},
		}

		result, err := spec.ToAtlas(orgID, idpID, &mongodbatlas.Client{
			Projects: pc,
		})

		assert.Error(t, err, "ToAtlas() should fail")
		assert.NotNil(t, result, "ToAtlas() result should not be nil")
	})
}
