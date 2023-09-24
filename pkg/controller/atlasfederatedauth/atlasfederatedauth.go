package atlasfederatedauth

import (
	"context"
	"fmt"
	"reflect"

	"go.mongodb.org/atlas/mongodbatlas"

	mdbv1 "github.com/mongodb/mongodb-atlas-kubernetes/pkg/api/v1"
	"github.com/mongodb/mongodb-atlas-kubernetes/pkg/controller/workflow"
)

const (
	ErrInvalidProvider    = "INVALID_PROVIDER"
	ErrNotOrgGroupCreator = "NOT_ORG_GROUP_CREATOR"
)

func (r *AtlasFederatedAuthReconciler) ensureFederatedAuth(service *workflow.Context, fedauth *mdbv1.AtlasFederatedAuth) workflow.Result {
	// If disabled, skip with no error
	if fedauth.Spec.Enabled == nil || (fedauth.Spec.Enabled != nil && !*fedauth.Spec.Enabled) {
		return workflow.OK().WithMessage(string(workflow.FederatedAuthIsNotEnabledInCR))
	}

	orgID := service.Connection.OrgID

	// Get current IDP for the ORG
	atlasFedSettings, _, err := service.Client.FederatedSettings.Get(context.Background(), orgID)
	if err != nil {
		return workflow.Terminate(workflow.FederatedAuthNotAvailable, err.Error())
	}

	atlasFedSettingsID := atlasFedSettings.ID

	// Get current Org config
	orgConfig, _, err := service.Client.FederatedSettings.GetConnectedOrg(context.Background(), atlasFedSettingsID, orgID)
	if err != nil {
		return workflow.Terminate(workflow.FederatedAuthOrgNotConnected, err.Error())
	}

	idpID := orgConfig.IdentityProviderID

	operatorConf, err := fedauth.Spec.ToAtlas(orgID, idpID, &service.Client)
	if err != nil {
		return workflow.Terminate(workflow.Internal, fmt.Sprintln("Can not convert Federated Auth spec to Atlas", err.Error()))
	}

	if result := r.ensureIDPSettings(atlasFedSettingsID, idpID, fedauth, &service.Client); !result.IsOk() {
		return result
	}

	if federatedSettingsAreEqual(operatorConf, orgConfig) {
		return workflow.OK()
	}

	updatedSettings, _, err := service.Client.FederatedSettings.UpdateConnectedOrg(context.Background(), atlasFedSettingsID, orgID, operatorConf)
	if err != nil {
		return workflow.Terminate(workflow.Internal, fmt.Sprintln("Can not update federation settings", err.Error()))
	}

	if updatedSettings.UserConflicts != nil && len(*updatedSettings.UserConflicts) != 0 {
		users := make([]string, 0, len(*updatedSettings.UserConflicts))
		for i := range *updatedSettings.UserConflicts {
			users = append(users, (*updatedSettings.UserConflicts)[i].EmailAddress)
		}

		return workflow.Terminate(workflow.FederatedAuthUsersConflict,
			fmt.Sprintln("The following users are in conflict", users))
	}

	return workflow.OK()
}

func (r *AtlasFederatedAuthReconciler) ensureIDPSettings(federationSettingsID, idpID string, fedauth *mdbv1.AtlasFederatedAuth, client *mongodbatlas.Client) workflow.Result {
	idpSettings, _, err := client.FederatedSettings.GetIdentityProvider(context.Background(), federationSettingsID, idpID)
	if err != nil {
		return workflow.Terminate(workflow.Internal, err.Error())
	}

	if fedauth.Spec.SSODebugEnabled != nil {
		if idpSettings.SsoDebugEnabled != nil && *idpSettings.SsoDebugEnabled == *fedauth.Spec.SSODebugEnabled {
			return workflow.OK()
		}

		*idpSettings.SsoDebugEnabled = *fedauth.Spec.SSODebugEnabled
		_, _, err := client.FederatedSettings.UpdateIdentityProvider(context.Background(), federationSettingsID, idpID, idpSettings)
		if err != nil {
			return workflow.Terminate(workflow.Internal, err.Error())
		}
	}

	// TODO: Add more IDP settings
	return workflow.OK()
}

func federatedSettingsAreEqual(operator, atlas *mongodbatlas.FederatedSettingsConnectedOrganization) bool {
	operator.UserConflicts = nil
	atlas.UserConflicts = nil
	return reflect.DeepEqual(operator, atlas)
}
