/*******************************************************************************
 *
 *
 * (c) Copyright Merative US L.P. and others 2020-2022 
 *
 * SPDX-Licence-Identifier: Apache 2.0
 *
 *******************************************************************************/

package main

import (
	"common/bchcls/asset_mgmt"
	"common/bchcls/cached_stub"
	"common/bchcls/crypto"
	"common/bchcls/custom_errors"
	"common/bchcls/data_model"
	"common/bchcls/index"
	"common/bchcls/key_mgmt"
	"common/bchcls/user_access_ctrl"
	"common/bchcls/user_mgmt"
	"common/bchcls/utils"

	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
)

// AuditPermissionIndex is used to specify the table name for AuditPermission assets
const AuditPermissionIndex = "AuditPermissionTable"

// AuditPermissionPrefix is the prefix for all AuditPermission IDs
const AuditPermissionPrefix = "AuditPermission"

// AuditPermissionNamespace is used to create the assetIDs for AuditPermission Assets
const AuditPermissionNamespace = "AuditPermissionAsset"

// AuditPermission is used to track which services an audit user has permissions for
// and to create a keypath so that auditors have access to a service's logs
// The keypath from an auditor to a service's logs will be as follows:
// AuditPermission.Auditor.PublicKey --> AuditPermission.SymKey --> AuditPermission.Service.LogSymKey
type AuditPermission struct {
	AuditPermissionID string `json:"audit_permission_id"`
	AuditorID         string `json:"auditor_id"`
	ServiceID         string `json:"service_id"`
	IsActive          bool   `json:"is_active"`
}

// AddAuditorPermission gives auditor permission to a user; the user role type must be audit
// Caller must be org or service admin
// args = [ userID, serviceID, auditPermissionAssetKeyB64 ]
func AddAuditorPermission(stub cached_stub.CachedStubInterface, caller data_model.User, args []string) ([]byte, error) {
	defer utils.ExitFnLog(utils.EnterFnLog())
	logger.Debugf("args: %v", args)

	if len(args) != 3 {
		customErr := &custom_errors.LengthCheckingError{Type: "AddAuditorPermission arguments length"}
		logger.Errorf(customErr.Error())
		return nil, errors.New(customErr.Error())
	}

	auditorID := args[0]
	serviceID := args[1]
	keyBytes, err := crypto.ParseSymKeyB64(args[2])
	if err != nil {
		errMsg := "Invalid auditPermissionSymKey"
		logger.Error(errMsg)
		return nil, errors.New(errMsg)
	}

	if utils.IsStringEmpty(auditorID) {
		customErr := &custom_errors.LengthCheckingError{Type: "auditorID"}
		logger.Errorf(customErr.Error())
		return nil, errors.WithStack(customErr)
	}

	if utils.IsStringEmpty(serviceID) {
		customErr := &custom_errors.LengthCheckingError{Type: "serviceID"}
		logger.Errorf(customErr.Error())
		return nil, errors.WithStack(customErr)
	}

	if keyBytes == nil {
		errMsg := "Invalid auditPermissionSymKey"
		logger.Error(errMsg)
		return nil, errors.New(errMsg)
	}

	callerObj := caller

	solutionCaller := convertToSolutionUser(caller)
	// caller is either org admin
	if solutionCaller.SolutionInfo.IsOrgAdmin {
		orgCaller, err := user_mgmt.GetUserData(stub, caller, solutionCaller.Org, true, false)
		if err != nil {
			customErr := &GetOrgError{Org: solutionCaller.Org}
			logger.Errorf("%v: %v", customErr, err)
			return nil, errors.Wrap(err, customErr.Error())
		}

		if orgCaller.PrivateKey == nil {
			errMsg := "Caller does not have access to org private key"
			logger.Error(errMsg)
			return nil, errors.New(errMsg)
		}

		callerObj = orgCaller
	} else {
		// or caller is service admin
		serviceCaller, err := user_mgmt.GetUserData(stub, caller, serviceID, true, false)
		if err != nil {
			customErr := &GetOrgError{Org: serviceID}
			logger.Errorf("%v: %v", customErr, err)
			return nil, errors.Wrap(err, customErr.Error())
		}

		if serviceCaller.PrivateKey == nil {
			errMsg := "Caller does not have access to service private key"
			logger.Errorf(errMsg)
			return nil, errors.New(errMsg)
		}

		callerObj = serviceCaller
	}

	solutionUser, err := GetSolutionUserWithParams(stub, callerObj, auditorID, false, false)
	if solutionUser.Role != SOLUTION_ROLE_AUDIT {
		logger.Errorf("User role must be audit")
		return nil, errors.New("User role must be audit")
	}

	return nil, putAuditPermission(stub, callerObj, auditorID, serviceID, keyBytes)
}

// RemoveAuditorPermission removed auditor permission from a user
// Caller must be org or service admin
// args = [ userID, serviceID ]
func RemoveAuditorPermission(stub cached_stub.CachedStubInterface, caller data_model.User, args []string) ([]byte, error) {
	defer utils.ExitFnLog(utils.EnterFnLog())
	logger.Debugf("args: %v", args)

	if len(args) != 2 {
		customErr := &custom_errors.LengthCheckingError{Type: "RemoveAuditorPermission arguments length"}
		logger.Errorf(customErr.Error())
		return nil, errors.New(customErr.Error())
	}

	auditorID := args[0]
	serviceID := args[1]

	if utils.IsStringEmpty(auditorID) {
		customErr := &custom_errors.LengthCheckingError{Type: "auditorID"}
		logger.Errorf(customErr.Error())
		return nil, errors.WithStack(customErr)
	}

	if utils.IsStringEmpty(serviceID) {
		customErr := &custom_errors.LengthCheckingError{Type: "serviceID"}
		logger.Errorf(customErr.Error())
		return nil, errors.WithStack(customErr)
	}

	callerObj := caller

	solutionCaller := convertToSolutionUser(caller)
	// caller is either org admin
	if solutionCaller.SolutionInfo.IsOrgAdmin {
		orgCaller, err := user_mgmt.GetUserData(stub, caller, solutionCaller.Org, true, false)
		if err != nil {
			customErr := &GetOrgError{Org: solutionCaller.Org}
			logger.Errorf("%v: %v", customErr, err)
			return nil, errors.Wrap(err, customErr.Error())
		}

		if orgCaller.PrivateKey == nil {
			errMsg := "Caller does not have access to org private key"
			logger.Error(errMsg)
			return nil, errors.New(errMsg)
		}

		callerObj = orgCaller
	} else {
		// or caller is service admin
		serviceCaller, err := user_mgmt.GetUserData(stub, caller, serviceID, true, false)
		if err != nil {
			customErr := &GetOrgError{Org: serviceID}
			logger.Errorf("%v: %v", customErr, err)
			return nil, errors.Wrap(err, customErr.Error())
		}

		if serviceCaller.PrivateKey == nil {
			errMsg := "Caller does not have access to service private key"
			logger.Errorf(errMsg)
			return nil, errors.New(errMsg)
		}

		callerObj = serviceCaller
	}

	return nil, deactivateAuditPermission(stub, callerObj, auditorID, serviceID)
}

// getAuditPermission gets the AuditPermission for the given auditor and service
func getAuditPermission(stub cached_stub.CachedStubInterface, caller data_model.User, auditorID, serviceID string) (AuditPermission, error) {
	defer utils.ExitFnLog(utils.EnterFnLog())

	auditPermissionID := GetAuditPermissionID(auditorID, serviceID)
	auditPermissionAssetID := asset_mgmt.GetAssetId(AuditPermissionNamespace, auditPermissionID)
	keyPath, err := GetKeyPath(stub, caller, auditPermissionAssetID)
	if err != nil {
		customErr := &GetKeyPathError{Caller: caller.ID, AssetID: auditPermissionAssetID}
		logger.Errorf(customErr.Error())
		return AuditPermission{}, errors.New(customErr.Error())
	}

	if len(keyPath) <= 0 {
		customErr := &GetKeyPathError{Caller: caller.ID, AssetID: auditPermissionAssetID}
		logger.Errorf(customErr.Error())
		return AuditPermission{}, errors.New(customErr.Error())
	}

	assetManager := asset_mgmt.GetAssetManager(stub, caller)

	auditPermissionKey, err := assetManager.GetAssetKey(auditPermissionAssetID, keyPath)
	if err != nil {
		logger.Errorf("Failed to GetAssetKey for auditPermissionKey: %v", err)
		return AuditPermission{}, errors.Wrap(err, "Failed to GetAssetKey for auditPermissionKey")
	}

	auditPermissionAsset, err := assetManager.GetAsset(auditPermissionAssetID, auditPermissionKey)
	if err != nil {
		customErr := &custom_errors.GetAssetDataError{AssetId: auditPermissionAssetID}
		logger.Errorf("%v: %v", customErr, err)
		return AuditPermission{}, errors.Wrap(err, customErr.Error())
	}

	if utils.IsStringEmpty(auditPermissionAsset.AssetId) {
		customErr := &custom_errors.GetAssetDataError{AssetId: auditPermissionAssetID}
		logger.Errorf(customErr.Error())
		return AuditPermission{}, errors.New(customErr.Error())
	}

	auditPermission := convertAuditPermissionFromAsset(auditPermissionAsset)

	return auditPermission, nil
}

// GetAuditPermissions gets all active AuditPermissions for the given auditorID
func GetAuditPermissions(stub cached_stub.CachedStubInterface, caller data_model.User, auditorID string) ([]AuditPermission, error) {
	defer utils.ExitFnLog(utils.EnterFnLog())

	if utils.IsStringEmpty(auditorID) {
		customErr := &custom_errors.LengthCheckingError{Type: "auditorID"}
		logger.Errorf(customErr.Error())
		return nil, errors.WithStack(customErr)
	}

	auditPermissions := []AuditPermission{}
	iter, err := asset_mgmt.GetAssetManager(stub, caller).GetAssetIter(AuditPermissionNamespace, AuditPermissionIndex, []string{"auditor_id", "is_active"}, []string{auditorID, "true"}, []string{auditorID, "true"}, true, false, KeyPathFunc, "", -1, nil)
	if err != nil {
		errMsg := "GetAuditPermission assets failed"
		logger.Error(errMsg)
		return []AuditPermission{}, errors.Wrap(err, errMsg)
	}

	defer iter.Close()
	for iter.HasNext() {
		auditPermissionAsset, err := iter.Next()
		if err != nil {
			customErr := &custom_errors.IterError{}
			logger.Errorf("%v: %v", customErr, err)
			return nil, errors.Wrap(err, customErr.Error())
		}

		auditPermission := convertAuditPermissionFromAsset(auditPermissionAsset)
		if !utils.IsStringEmpty(auditPermission.AuditorID) {
			auditPermissions = append(auditPermissions, auditPermission)
		}
	}

	return auditPermissions, nil
}

// activateAuditPermission activates an AuditPermission object
// caller should always be org or service admin
func activateAuditPermission(stub cached_stub.CachedStubInterface, caller data_model.User, auditPermission AuditPermission) error {
	auditPermission.IsActive = true

	auditPermissionAsset, err := convertAuditPermissionToAsset(auditPermission)
	if err != nil {
		customErr := &ConvertToAssetError{Asset: "auditPermissionAsset"}
		logger.Errorf("%v: %v", customErr, err)
		return errors.Wrap(err, customErr.Error())
	}

	auditPermissionAssetID := asset_mgmt.GetAssetId(AuditPermissionNamespace, auditPermission.AuditPermissionID)
	keyPath, err := GetKeyPath(stub, caller, auditPermissionAssetID)
	if err != nil {
		customErr := &GetKeyPathError{Caller: caller.ID, AssetID: auditPermissionAssetID}
		logger.Errorf(customErr.Error())
		return errors.New(customErr.Error())
	}

	if len(keyPath) <= 0 {
		customErr := &GetKeyPathError{Caller: caller.ID, AssetID: auditPermissionAssetID}
		logger.Errorf(customErr.Error())
		return errors.New(customErr.Error())
	}

	assetManager := asset_mgmt.GetAssetManager(stub, caller)

	auditPermissionKey, err := assetManager.GetAssetKey(auditPermissionAssetID, keyPath)
	if err != nil {
		logger.Errorf("Failed to GetAssetKey for auditPermissionKey: %v", err)
		return errors.Wrap(err, "Failed to GetAssetKey for auditPermissionKey")
	}

	auditPermissionAsset.AssetKeyId = auditPermissionKey.ID
	err = assetManager.UpdateAsset(auditPermissionAsset, auditPermissionKey, true)
	if err != nil {
		customErr := &PutAssetError{Asset: auditPermission.AuditPermissionID}
		logger.Errorf("%v: %v", customErr, err)
		return errors.Wrap(err, customErr.Error())
	}

	err = giveUsersAccessToAuditPermission(stub, caller, auditPermission.AuditorID, auditPermission.ServiceID, auditPermissionKey)
	if err != nil {
		errMsg := "Failed to give users access to AuditPermission"
		logger.Error(errMsg)
		return errors.Wrap(err, errMsg)
	}

	return nil
}

func giveUsersAccessToAuditPermission(stub cached_stub.CachedStubInterface, caller data_model.User, auditorID, serviceID string, auditPermissionSymKey data_model.Key) error {

	// establish key relationships so auditor can access service logs
	userAccessManager := user_access_ctrl.GetUserAccessManager(stub, caller)

	// give auditor access to the auditPermission
	auditorUser, err := user_mgmt.GetUserData(stub, caller, auditorID, false, false)
	if err != nil {
		customErr := &GetUserError{User: "auditor user"}
		logger.Errorf("%v: %v", customErr, err)
		return errors.Wrap(err, customErr.Error())
	}

	auditorPublicKey := auditorUser.GetPublicKey()

	err = userAccessManager.AddAccessByKey(auditorPublicKey, auditPermissionSymKey)
	if err != nil {
		customErr := &custom_errors.AddAccessError{Key: "auditorPublicKey to auditPermissionSymKey"}
		logger.Errorf("%v: %v", customErr, err)
		return errors.Wrap(err, customErr.Error())
	}

	// give auditPermission access to service logs
	service, err := user_mgmt.GetUserData(stub, caller, serviceID, true, false)
	if err != nil {
		customErr := &GetServiceError{Service: serviceID}
		logger.Errorf("%v: %v", customErr, err)
		return errors.Wrap(err, customErr.Error())
	}

	servicePublicKey := service.GetPublicKey()

	err = userAccessManager.AddAccessByKey(servicePublicKey, auditPermissionSymKey)
	if err != nil {
		customErr := &custom_errors.AddAccessError{Key: "serviceSymKey to auditPermissionSymKey"}
		logger.Errorf("%v: %v", customErr, err)
		return errors.Wrap(err, customErr.Error())
	}

	serviceLogSymKey := service.GetLogSymKey()

	err = userAccessManager.AddAccessByKey(auditPermissionSymKey, serviceLogSymKey)
	if err != nil {
		customErr := &custom_errors.AddAccessError{Key: "auditPermissionSymKey to serviceLogSymKey"}
		logger.Errorf("%v: %v", customErr, err)
		return errors.Wrap(err, customErr.Error())
	}

	return nil
}

// putAuditPermission activates an AuditPermission object if it already exists or creates an AuditPermission asset
// caller should always be org or service admin
func putAuditPermission(stub cached_stub.CachedStubInterface, caller data_model.User, auditorID, serviceID string, symKeyBytes []byte) error {
	defer utils.ExitFnLog(utils.EnterFnLog())

	// check if auditPermission exists
	auditPermission, err := getAuditPermission(stub, caller, auditorID, serviceID)
	if (auditPermission != AuditPermission{}) {
		if err != nil {
			errMsg := "Failed to get AuditPermission"
			logger.Error(errMsg)
			return errors.Wrap(err, errMsg)
		}

		err := activateAuditPermission(stub, caller, auditPermission)
		if err != nil {
			errMsg := "Failed to activate AuditPermission"
			logger.Error(errMsg)
			return errors.Wrap(err, errMsg)
		}

		return nil
	}

	// create asset
	auditPermissionID := GetAuditPermissionID(auditorID, serviceID)
	auditPermission = AuditPermission{AuditPermissionID: auditPermissionID, AuditorID: auditorID, ServiceID: serviceID, IsActive: true}
	auditPermissionAsset, err := convertAuditPermissionToAsset(auditPermission)
	if err != nil {
		errMsg := "Failed to convert AuditPermission to asset"
		logger.Error(errMsg)
		return errors.Wrap(err, errMsg)
	}

	// create asset's symkey
	auditPermissionSymKey := data_model.Key{
		ID:       key_mgmt.GetSymKeyId(auditPermissionID),
		KeyBytes: symKeyBytes,
		Type:     key_mgmt.KEY_TYPE_SYM,
	}

	// save asset
	assetManager := asset_mgmt.GetAssetManager(stub, caller)
	err = assetManager.AddAsset(auditPermissionAsset, auditPermissionSymKey, false)
	if err != nil {
		errMsg := "Failed to AddAsset AuditPermission"
		logger.Error(errMsg)
		return errors.Wrap(err, errMsg)
	}

	err = giveUsersAccessToAuditPermission(stub, caller, auditorID, serviceID, auditPermissionSymKey)
	if err != nil {
		errMsg := "Failed to give users access to AuditPermission"
		logger.Error(errMsg)
		return errors.Wrap(err, errMsg)
	}

	return nil
}

// deactivateAuditPermission deactivates an AuditPermission object
// caller should always be org or service admin
// The IsActive field of the AuditPermission will be set to false and the edge from the AuditPermission to the Service's LogSymKey will be removed
func deactivateAuditPermission(stub cached_stub.CachedStubInterface, caller data_model.User, auditorID, serviceID string) error {
	defer utils.ExitFnLog(utils.EnterFnLog())

	auditPermission, err := getAuditPermission(stub, caller, auditorID, serviceID)
	if err != nil {
		errMsg := "Failed to get AuditPermission"
		logger.Error(errMsg)
		return errors.Wrap(err, errMsg)
	}

	auditPermission.IsActive = false

	auditPermissionAssetID := asset_mgmt.GetAssetId(AuditPermissionNamespace, auditPermission.AuditPermissionID)
	keyPath, err := GetKeyPath(stub, caller, auditPermissionAssetID)
	if err != nil {
		customErr := &GetKeyPathError{Caller: caller.ID, AssetID: auditPermissionAssetID}
		logger.Errorf(customErr.Error())
		return errors.New(customErr.Error())
	}

	if len(keyPath) <= 0 {
		customErr := &GetKeyPathError{Caller: caller.ID, AssetID: auditPermissionAssetID}
		logger.Errorf(customErr.Error())
		return errors.New(customErr.Error())
	}

	assetManager := asset_mgmt.GetAssetManager(stub, caller)

	auditPermissionKey, err := assetManager.GetAssetKey(auditPermissionAssetID, keyPath)
	if err != nil {
		logger.Errorf("Failed to GetAssetKey for auditPermissionKey: %v", err)
		return errors.Wrap(err, "Failed to GetAssetKey for auditPermissionKey")
	}

	auditPermissionAsset, err := convertAuditPermissionToAsset(auditPermission)
	if err != nil {
		customErr := &ConvertToAssetError{Asset: "auditPermissionAsset"}
		logger.Errorf("%v: %v", customErr, err)
		return errors.Wrap(err, customErr.Error())
	}

	auditPermissionAsset.AssetKeyId = auditPermissionKey.ID
	err = assetManager.UpdateAsset(auditPermissionAsset, auditPermissionKey, true)
	if err != nil {
		customErr := &PutAssetError{Asset: auditPermission.AuditPermissionID}
		logger.Errorf("%v: %v", customErr, err)
		return errors.Wrap(err, customErr.Error())
	}

	userAccessManager := user_access_ctrl.GetUserAccessManager(stub, caller)

	auditPermissionSymKeyID := key_mgmt.GetSymKeyId(auditPermission.AuditPermissionID)
	serviceLogSymKeyID := key_mgmt.GetLogSymKeyId(auditPermission.ServiceID)

	err = userAccessManager.RemoveAccessByKey(auditPermissionSymKeyID, serviceLogSymKeyID)
	if err != nil {
		customErr := &custom_errors.AddAccessError{Key: "auditPermissionSymKey to serviceLogSymKey"}
		logger.Errorf("%v: %v", customErr, err)
		return errors.Wrap(err, customErr.Error())
	}

	return nil
}

// GetAuditPermissionID creates the ID for an AuditPermission
func GetAuditPermissionID(userID, serviceID string) string {
	defer utils.ExitFnLog(utils.EnterFnLog())
	return fmt.Sprintf("%s-%s-%s", AuditPermissionPrefix, userID, serviceID)
}

func convertAuditPermissionFromAsset(asset *data_model.Asset) AuditPermission {
	defer utils.ExitFnLog(utils.EnterFnLog())

	auditPermission := AuditPermission{}

	// all info is currentl stored as private data
	json.Unmarshal(asset.PrivateData, &auditPermission)
	return auditPermission
}

// convertAuditPermissionToAsset converts an AuditPermission into an asset
func convertAuditPermissionToAsset(auditPermission AuditPermission) (data_model.Asset, error) {
	defer utils.ExitFnLog(utils.EnterFnLog())

	auditPermissionID := GetAuditPermissionID(auditPermission.AuditorID, auditPermission.ServiceID)

	asset := data_model.Asset{}
	asset.AssetId = asset_mgmt.GetAssetId(AuditPermissionNamespace, auditPermissionID)
	asset.Datatypes = []string{}
	asset.OwnerIds = []string{auditPermission.ServiceID}
	asset.IndexTableName = AuditPermissionIndex

	metadata := make(map[string]string)
	metadata["namespace"] = AuditPermissionNamespace
	asset.Metadata = metadata

	// currently, nothing in the permission should be public
	var publicData interface{}
	asset.PublicData, _ = json.Marshal(&publicData)

	// currently, whole permission should be private
	privateData, err := json.Marshal(&auditPermission)
	if err != nil {
		customErr := &custom_errors.MarshalError{Type: "AuditPermission PrivateData"}
		logger.Errorf("%v: %v", customErr, err)
		return data_model.Asset{}, errors.Wrap(err, customErr.Error())
	}
	asset.PrivateData = privateData

	return asset, nil
}

// SetupAuditPermissionIndex sets up indices for AuditPermissions
func SetupAuditPermissionIndex(stub cached_stub.CachedStubInterface) error {
	defer utils.ExitFnLog(utils.EnterFnLog())

	auditPermissionTable := index.GetTable(stub, AuditPermissionIndex, "audit_permission_id")
	auditPermissionTable.AddIndex([]string{"auditor_id", "is_active", "audit_permission_id"}, false)
	err := auditPermissionTable.SaveToLedger()
	if err != nil {
		return err
	}

	return nil
}
