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
	"common/bchcls/asset_mgmt/asset_key_func"
	"common/bchcls/cached_stub"
	"common/bchcls/consent_mgmt"
	"common/bchcls/crypto"
	"common/bchcls/custom_errors"
	"common/bchcls/data_model"
	"common/bchcls/datatype"
	"common/bchcls/key_mgmt"
	"common/bchcls/user_mgmt"
	"common/bchcls/user_mgmt/user_groups"
	"common/bchcls/utils"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/pkg/errors"
)

// GetStringSliceFromInterface converts an interface to a string slice
func GetStringSliceFromInterface(data []interface{}) []string {
	s := make([]string, len(data))
	for i, v := range data {
		s[i] = fmt.Sprint(v)
	}
	return s
}

// CallerIsAdminOfService checks if caller is an admin (service or org) of a given service
func CallerIsAdminOfService(caller data_model.User, service string, serviceOrg string) bool {
	callerOMR := convertToSolutionUser(caller)
	// Check if caller is direct service admin
	if utils.InList(callerOMR.SolutionInfo.Services, service) {
		return true
	}

	// Check if caller is org admin
	if callerOMR.SolutionInfo.IsOrgAdmin && callerOMR.Org == serviceOrg {
		return true
	}

	return false
}

// GetDataInternal is the internal function for downloading data using index
// Assume caller is consent target
func GetDataInternal(stub cached_stub.CachedStubInterface, caller data_model.User, fieldNames []string, startValues []string, endValues []string, maxNum int) ([]OwnerDataResult, error) {
	defer utils.ExitFnLog(utils.EnterFnLog())

	datas := []OwnerDataResult{}

	// Use index to find all consents
	iter, err := asset_mgmt.GetAssetManager(stub, caller).GetAssetIter(OwnerDataNamespace, IndexData, fieldNames, startValues, endValues, true, false, KeyPathFunc, "", maxNum, nil)
	if err != nil {
		logger.Errorf("GetAssets failed: %v", err)
		return nil, errors.Wrap(err, "GetAssets failed")
	}

	// Iterate over all data and filter
	defer iter.Close()
	for iter.HasNext() {
		dataAsset, err := iter.Next()
		if err != nil {
			customErr := &custom_errors.IterError{}
			logger.Errorf("%v: %v", customErr, err)
			continue
		}

		if utils.IsStringEmpty(dataAsset.AssetId) {
			continue
		}

		if data_model.IsEncryptedData(dataAsset.PrivateData) {
			logger.Error("Failed to decrypt data asset")
			continue
		}

		data := convertOwnerDataFromAsset(dataAsset)
		datas = append(datas, data)
	}

	return datas, nil
}

// GetPatientDataID composes patient data ID
func GetPatientDataID(owner string, datatype string, timestamp int64) string {
	defer utils.ExitFnLog(utils.EnterFnLog())
	// return owner + datatype + strconv.FormatInt(timestamp, 10)
	time, _ := utils.ConvertToString(timestamp)
	return owner + datatype + time
}

// GetOwnerDataID composes owner data ID
func GetOwnerDataID(owner string, datatype string, timestamp int64) string {
	defer utils.ExitFnLog(utils.EnterFnLog())
	return owner + datatype + strconv.FormatInt(timestamp, 10)
}

// GetLatestOwnerDataAssetID is a helper function that gets the asset ID of the latest owner data uploaded
func GetLatestOwnerDataAssetID(stub cached_stub.CachedStubInterface, owner string, datatype string) string {
	defer utils.ExitFnLog(utils.EnterFnLog())
	// return assetID, nil
	latestDataID := GetOwnerDataID(owner, datatype, -1)
	return asset_mgmt.GetAssetId(OwnerDataNamespace, latestDataID)
}

// GetLatestPatientDataAssetID is a helper function that gets the asset ID of the latest patient data uploaded
func GetLatestPatientDataAssetID(stub cached_stub.CachedStubInterface, owner string, datatype string) string {
	defer utils.ExitFnLog(utils.EnterFnLog())
	// return assetID, nil
	latestDataID := GetPatientDataID(owner, datatype, -1)
	return asset_mgmt.GetAssetId(OwnerDataNamespace, latestDataID)
}

// GetDataWithAssetID returns OwnerData object given an assetID
func GetDataWithAssetID(stub cached_stub.CachedStubInterface, caller data_model.User, assetID string, owner string, datatypeID string) (OwnerDataResult, error) {
	defer utils.ExitFnLog(utils.EnterFnLog())

	if utils.IsStringEmpty(assetID) {
		logger.Errorf("Failed to get assetID of latest data; empty assetID")
		return OwnerDataResult{}, errors.New("Failed to get assetID of latest data; empty assetID")
	}

	keyPath, err := GetKeyPath(stub, caller, assetID)
	if err != nil {
		customErr := &GetKeyPathError{Caller: caller.ID, AssetID: assetID}
		logger.Errorf(customErr.Error())
		return OwnerDataResult{}, errors.New(customErr.Error())
	}

	if len(keyPath) <= 0 {
		customErr := &GetKeyPathError{Caller: caller.ID, AssetID: assetID}
		logger.Errorf(customErr.Error())
		return OwnerDataResult{}, errors.New(customErr.Error())
	}

	assetManager := asset_mgmt.GetAssetManager(stub, caller)
	dataKey, err := assetManager.GetAssetKey(assetID, keyPath)
	if err != nil {
		logger.Errorf("Failed to get data key with latest owner data asset ID: %v", err)
		return OwnerDataResult{}, errors.Wrap(err, "Failed to get data key with latest owner data asset ID")
	}

	dataAsset, err := assetManager.GetAsset(assetID, dataKey)
	if err != nil {
		customErr := &custom_errors.GetAssetDataError{AssetId: assetID}
		logger.Errorf("%v: %v", customErr, err)
		return OwnerDataResult{}, errors.Wrap(err, customErr.Error())
	}

	if utils.IsStringEmpty(dataAsset.AssetId) {
		customErr := &custom_errors.GetAssetDataError{AssetId: assetID}
		logger.Errorf(customErr.Error())
		return OwnerDataResult{}, errors.New(customErr.Error())
	}

	data := convertOwnerDataFromAsset(dataAsset)
	return data, nil
}

// CheckConsentOptionIsInvalid checks if the passed in string slice is a valid combination of write and read, or deny
func CheckConsentOptionIsInvalid(option []string) bool {
	invalidOption := utils.FilterOutFromSet(option, []string{consentOptionWrite, consentOptionRead, consentOptionDeny})
	if len(invalidOption) > 0 {
		logger.Errorf("invalid consent option: %v", invalidOption)
		return true
	}

	// consent option cannot be write and deny, or read and deny
	if (utils.InList(option, consentOptionWrite) && utils.InList(option, consentOptionDeny)) || (utils.InList(option, consentOptionRead) && utils.InList(option, consentOptionDeny)) {
		logger.Errorf("invalid consent option, cannot pair deny with another option")
		return true
	}

	return false
}

// OMRServiceAssetKeyPathFunc retrieves the key path from caller's pub/priv key to a service asset key, given service asset.
var OMRServiceAssetKeyPathFunc asset_key_func.AssetKeyPathFunc = func(stub cached_stub.CachedStubInterface, caller data_model.User, asset data_model.Asset) ([]string, error) {
	return GetKeyPathFromCallerToServiceAsset(stub, caller, asset.AssetKeyId), nil
}

// GetKeyPathFromCallerToServiceAsset returns the key path from caller's pub/prv key to a service asset
func GetKeyPathFromCallerToServiceAsset(stub cached_stub.CachedStubInterface, caller data_model.User, assetKeyID string) []string {
	// add caller's pub/priv key id
	keyPath := []string{caller.GetPubPrivKeyId()}

	// ===================================================================
	// admin paths:
	// if caller is org admin
	solutionCaller := convertToSolutionUser(caller)
	if solutionCaller.SolutionInfo.IsOrgAdmin {
		org, _ := user_mgmt.GetUserData(stub, caller, solutionCaller.Org, true, false)
		if org.PrivateKey != nil {
			keyPath = append(keyPath, org.GetPrivateKeyHashSymKeyId(), org.GetPubPrivKeyId(), org.GetSymKeyId())
		}
	}
	// ===================================================================

	// add asset key id
	keyPath = append(keyPath, assetKeyID)
	return keyPath
}

// GetKeyPathFromOrgAdminToOrgUser returns key path from one org user to another
// SymKey path from orgUser2 (admin) to orgUser1 should be:
// pub-priv-orgUser2 private-hash-org pub-priv-org private-hash-orgUser1 pub-priv-orgUser1 sym-orgUser1
// PrvKey path from orgUser2 (admin) to orgUser1 should be:
// pub-priv-orgUser2 private-hash-org pub-priv-org private-hash-orgUser1 pub-priv-orgUser1
func GetKeyPathFromOrgAdminToOrgUser(caller data_model.User, orgID string, orgUserID string, getSymKey bool) []string {
	keyPath := []string{caller.GetPubPrivKeyId(), key_mgmt.GetPrivateKeyHashSymKeyId(orgID), key_mgmt.GetPubPrivKeyId(orgID), key_mgmt.GetPrivateKeyHashSymKeyId(orgUserID), key_mgmt.GetPubPrivKeyId(orgUserID)}
	if getSymKey {
		keyPath = append(keyPath, key_mgmt.GetSymKeyId(orgUserID))
	}
	return keyPath
}

// GetLogSymKeyFromKey deterministically generates and returns a log sym key from hash of key.
func GetLogSymKeyFromKey(key data_model.Key) data_model.Key {
	logSymKeyBytes := append(key.KeyBytes, "logSymKey"...)
	return data_model.Key{
		ID:       key.GetLogSymKeyId(),
		KeyBytes: crypto.GetSymKeyFromHash(logSymKeyBytes),
		Type:     key_mgmt.KEY_TYPE_SYM,
	}
}

// OMRAssetKeyPathFuncForLogging retrieves the key path from caller's pub/priv key to asset key, given asset.
var OMRAssetKeyPathFuncForLogging asset_key_func.AssetKeyPathFunc = func(stub cached_stub.CachedStubInterface, caller data_model.User, asset data_model.Asset) ([]string, error) {
	return GetKeyPathFromCallerToLog(stub, caller, asset.AssetKeyId), nil
}

// GetKeyPathFromCallerToLog returns a key path from caller to log asset
func GetKeyPathFromCallerToLog(stub cached_stub.CachedStubInterface, caller data_model.User, assetKeyID string) []string {
	keyPath := []string{caller.GetPubPrivKeyId()}

	solutionCaller := convertToSolutionUser(caller)
	if solutionCaller.SolutionInfo.IsOrgAdmin {
		org := solutionCaller.Org
		orgKeyPath := append(keyPath, []string{key_mgmt.GetPrivateKeyHashSymKeyId(org), key_mgmt.GetPubPrivKeyId(org)}...)

		// get org's services
		groupIDs, err := user_groups.SlowGetSubgroups(stub, org)
		if err != nil {
			logger.Warning("Failed to SlowGetSubgroups for caller's org")
		}

		for _, groupID := range groupIDs {
			groupPrivateKeyHashID := key_mgmt.GetPrivateKeyHashSymKeyId(groupID)
			groupPrivateKeyID := key_mgmt.GetPubPrivKeyId(groupID)
			groupSymKeyID := key_mgmt.GetSymKeyId(groupID)
			groupLogSymKeyID := key_mgmt.GetLogSymKeyId(groupID)

			// construct group admin key path
			groupAdminKeyPath := append(orgKeyPath, []string{groupPrivateKeyHashID, groupPrivateKeyID, groupSymKeyID, groupLogSymKeyID, assetKeyID}...)

			// check if group admin key path exists
			pathExists, err := key_mgmt.VerifyAccessPath(stub, groupAdminKeyPath)
			if err != nil {
				logger.Warning("Failed to VerifyAccessPath")
			}
			if pathExists {
				fmt.Println()
				fmt.Println("Org user to log path exists: ", groupAdminKeyPath)
				fmt.Println()
				return groupAdminKeyPath
			}
		}

	} else if len(solutionCaller.SolutionInfo.Services) > 0 {
		for _, serviceID := range solutionCaller.SolutionInfo.Services {
			if caller.ID == serviceID {
				// if caller is default service admin, no need to loop throught the rest of services
				break
			}

			// check if key path exists for each service
			servicePrivateKeyHashID := key_mgmt.GetPrivateKeyHashSymKeyId(serviceID)
			servicePrivateKeyID := key_mgmt.GetPubPrivKeyId(serviceID)
			serviceSymKeyID := key_mgmt.GetSymKeyId(serviceID)
			serviceLogSymKeyID := key_mgmt.GetLogSymKeyId(serviceID)

			// construct group admin key path
			groupAdminKeyPath := append(keyPath, []string{servicePrivateKeyHashID, servicePrivateKeyID, serviceSymKeyID, serviceLogSymKeyID, assetKeyID}...)

			// check if group admin key path exists
			pathExists, err := key_mgmt.VerifyAccessPath(stub, groupAdminKeyPath)
			if err != nil {
				logger.Warning("Failed to VerifyAccessPath")
			}
			if pathExists {
				fmt.Println()
				fmt.Println("Service user to log path exists: ", groupAdminKeyPath)
				fmt.Println()
				return groupAdminKeyPath
			}
		}
	} else if solutionCaller.Role == SOLUTION_ROLE_AUDIT {
		auditPermissions, err := GetAuditPermissions(stub, caller, solutionCaller.ID)
		if err != nil {
			logger.Error(err)
		}

		for _, auditPermission := range auditPermissions {
			auditPermissionSymKeyID := key_mgmt.GetSymKeyId(auditPermission.AuditPermissionID)

			serviceID := auditPermission.ServiceID
			serviceLogSymKeyID := key_mgmt.GetLogSymKeyId(serviceID)

			auditKeyPath := append(keyPath, []string{auditPermissionSymKeyID, serviceLogSymKeyID, assetKeyID}...)
			pathExists, err := key_mgmt.VerifyAccessPath(stub, auditKeyPath)
			if err != nil {
				logger.Warning("Failed to VerifyAccessPath")
			}
			if pathExists {
				fmt.Println()
				fmt.Println("Audit user to log path exists: ", auditKeyPath)
				fmt.Println()
				return auditKeyPath
			}
		}
	}

	// key path: caller prv key, caller sym key, caller log sym key, asset key
	keyPath = append(keyPath, []string{caller.GetSymKeyId(), caller.GetLogSymKeyId(), assetKeyID}...)
	return keyPath
}

// KeyPathFunc retrieves the key path from caller's pub/priv key to asset key, given asset
var KeyPathFunc asset_key_func.AssetKeyPathFunc = func(stub cached_stub.CachedStubInterface, caller data_model.User, asset data_model.Asset) ([]string, error) {

	// if caller is org admin, pass org as optional param
	solutionCaller := convertToSolutionUser(caller)
	if solutionCaller.SolutionInfo.IsOrgAdmin {
		return GetKeyPath(stub, caller, asset.AssetId, solutionCaller.Org)
	}

	return GetKeyPath(stub, caller, asset.AssetId)
}

// GetKeyPath returns keyPath to asset sym key given caller and assetID
// optional parameters can be passed in; the first option must be orgID, the second option can be consentTargetID
//
// all solution level assets:
// org, service, user, enrollment, data, consent, contract, log
func GetKeyPath(stub cached_stub.CachedStubInterface, caller data_model.User, assetID string, options ...string) ([]string, error) {
	assetData, err := asset_mgmt.GetEncryptedAssetData(stub, assetID)
	if err != nil {
		customErr := &custom_errors.GetAssetDataError{AssetId: assetID}
		logger.Errorf("%v", customErr)
		return []string{}, errors.WithStack(customErr)
	}

	// retrieve assetKeyID
	assetKeyID := assetData.AssetKeyId
	assetType := assetData.Metadata["namespace"]

	// assume asset does not exist
	if utils.IsStringEmpty(assetType) || len(assetData.OwnerIds) <= 0 {
		return []string{}, nil
	}

	fmt.Println()
	fmt.Println("assetType: ", assetType)
	fmt.Println("Owners: ", assetData.OwnerIds)
	fmt.Println()

	// Use key path to get key
	// ========================================================================================
	// Option 1: caller is owner of asset
	// Key path: [caller private key ID, asset key ID]
	keyPath := []string{caller.GetPubPrivKeyId(), assetKeyID}

	// Verify key path exists in graph
	pathExists, err := key_mgmt.VerifyAccessPath(stub, keyPath)
	if err != nil {
		logger.Errorf("KeyPath verification failed")
		return nil, errors.Wrap(err, "KeyPath verification failed")
	}

	if pathExists {
		fmt.Println()
		fmt.Println("Option 1 path exists: ", keyPath)
		fmt.Println()
		return keyPath, nil
	}

	// ========================================================================================
	// Option 2a: caller is direct admin of owner of asset
	// Key path: [caller private key ID, owner private key hash, owner private key ID, asset key ID]
	// first get owner(group) private key hash
	owner := data_model.User{ID: assetData.OwnerIds[0]}
	keyPath = []string{caller.GetPubPrivKeyId(), owner.GetPrivateKeyHashSymKeyId(), owner.GetPubPrivKeyId(), assetKeyID}

	// Verify key path exists in graph
	pathExists, err = key_mgmt.VerifyAccessPath(stub, keyPath)
	if err != nil {
		logger.Errorf("KeyPath verification failed")
		return nil, errors.Wrap(err, "KeyPath verification failed")
	}

	if pathExists {
		fmt.Println()
		fmt.Println("Option 2a path exists: ", keyPath)
		fmt.Println()
		return keyPath, nil
	}

	// ========================================================================================
	// Option 2b: caller is indirect admin of owner of asset
	// Key path: [caller private key ID, org private key hash, org private key ID, service private key hash, service private key ID, asset key ID]
	// special case if assetType is serviceAsset
	if assetType == ServiceAssetNamespace {
		// convert assetData to service to get org ID
		servicePublicData := ServicePublicData{}
		json.Unmarshal(assetData.PublicData, &servicePublicData)
		org := data_model.User{ID: servicePublicData.OrgID}
		service := data_model.User{ID: servicePublicData.ServiceID}

		keyPath = []string{caller.GetPubPrivKeyId(), org.GetPrivateKeyHashSymKeyId(), org.GetPubPrivKeyId(), service.GetPrivateKeyHashSymKeyId(), service.GetPubPrivKeyId(), assetKeyID}
		pathExists, err = key_mgmt.VerifyAccessPath(stub, keyPath)
		if err != nil {
			logger.Errorf("KeyPath verification failed")
			return nil, errors.Wrap(err, "KeyPath verification failed")
		}

		if pathExists {
			fmt.Println()
			fmt.Println("Option 2b path exists: ", keyPath)
			fmt.Println()
			return keyPath, nil
		}
	}

	if len(options) >= 1 {
		orgID := options[0]
		org := data_model.User{ID: orgID}

		keyPath = []string{caller.GetPubPrivKeyId(), org.GetPrivateKeyHashSymKeyId(), org.GetPubPrivKeyId(), owner.GetPrivateKeyHashSymKeyId(), owner.GetPubPrivKeyId(), assetKeyID}
		pathExists, err = key_mgmt.VerifyAccessPath(stub, keyPath)
		if err != nil {
			logger.Errorf("KeyPath verification failed")
			return nil, errors.Wrap(err, "KeyPath verification failed")
		}

		if pathExists {
			fmt.Println()
			fmt.Println("Option 2b path exists: ", keyPath)
			fmt.Println()
			return keyPath, nil
		}
	}

	// ========================================================================================
	if assetType == "data_model.Consent" {
		// Option 2c: caller is direct admin of consent target
		// Key path: [caller private key ID, service private key hash, service private key ID, consent asset key ID]
		if len(options) >= 2 {
			orgID := options[0]
			org := data_model.User{ID: orgID}

			consentTargetID := options[1]

			target := data_model.User{ID: consentTargetID}
			keyPath = []string{caller.GetPubPrivKeyId(), target.GetPrivateKeyHashSymKeyId(), target.GetPubPrivKeyId(), assetKeyID}

			pathExists, err = key_mgmt.VerifyAccessPath(stub, keyPath)
			if err != nil {
				logger.Errorf("KeyPath verification failed")
				return nil, errors.Wrap(err, "KeyPath verification failed")
			}

			if pathExists {
				fmt.Println()
				fmt.Println("Option 2c path exists: ", keyPath)
				fmt.Println()
				return keyPath, nil
			}

			// ========================================================================================
			// Option 2d: caller is org admin of consent target
			// Key path: [caller private key ID, org private key hash, org private key ID, service private key hash, service private key ID, consent asset key ID]
			keyPath = []string{caller.GetPubPrivKeyId(), org.GetPrivateKeyHashSymKeyId(), org.GetPubPrivKeyId(), target.GetPrivateKeyHashSymKeyId(), target.GetPubPrivKeyId(), assetKeyID}

			pathExists, err = key_mgmt.VerifyAccessPath(stub, keyPath)
			if err != nil {
				logger.Errorf("KeyPath verification failed")
				return nil, errors.Wrap(err, "KeyPath verification failed")
			}

			if pathExists {
				fmt.Println()
				fmt.Println("Option 2d path exists: ", keyPath)
				fmt.Println()
				return keyPath, nil
			}
		}
	}

	// ========================================================================================
	if assetType == ContractAssetNamespace {
		contractPublicData := ContractPublicData{}
		err = json.Unmarshal(assetData.PublicData, &contractPublicData)
		ownerService := data_model.User{ID: contractPublicData.OwnerServiceID}
		requesterService := data_model.User{ID: contractPublicData.RequesterServiceID}

		// Option 2e: caller is direct admin (service admin) of contract data requester or data owner
		// Key path: [caller private key ID, service owner private key hash, service owner private key ID, asset key ID]
		keyPath = []string{caller.GetPubPrivKeyId(), ownerService.GetPrivateKeyHashSymKeyId(), ownerService.GetPubPrivKeyId(), assetKeyID}
		pathExists, err = key_mgmt.VerifyAccessPath(stub, keyPath)
		if err != nil {
			logger.Errorf("KeyPath verification failed")
			return nil, errors.Wrap(err, "KeyPath verification failed")
		}

		if pathExists {
			fmt.Println()
			fmt.Println("Option 2e path exists: ", keyPath)
			fmt.Println()
			return keyPath, nil
		}

		keyPath = []string{caller.GetPubPrivKeyId(), requesterService.GetPrivateKeyHashSymKeyId(), requesterService.GetPubPrivKeyId(), assetKeyID}
		pathExists, err = key_mgmt.VerifyAccessPath(stub, keyPath)
		if err != nil {
			logger.Errorf("KeyPath verification failed")
			return nil, errors.Wrap(err, "KeyPath verification failed")
		}

		if pathExists {
			fmt.Println()
			fmt.Println("Option 2e path exists: ", keyPath)
			fmt.Println()
			return keyPath, nil
		}

		// Option 2f: caller is indirect admin (org admin) of contract data requester or data owner
		// Key path: [caller private key ID, org private key hash, org private key ID, service owner private key hash, service owner private key ID, asset key ID]
		if len(options) >= 1 {
			orgID := options[0]
			org := data_model.User{ID: orgID}

			keyPath = []string{caller.GetPubPrivKeyId(), org.GetPrivateKeyHashSymKeyId(), org.GetPubPrivKeyId(), ownerService.GetPrivateKeyHashSymKeyId(), ownerService.GetPubPrivKeyId(), assetKeyID}
			pathExists, err = key_mgmt.VerifyAccessPath(stub, keyPath)
			if err != nil {
				logger.Errorf("KeyPath verification failed")
				return nil, errors.Wrap(err, "KeyPath verification failed")
			}

			if pathExists {
				fmt.Println()
				fmt.Println("Option 2f path exists: ", keyPath)
				fmt.Println()
				return keyPath, nil
			}

			keyPath = []string{caller.GetPubPrivKeyId(), org.GetPrivateKeyHashSymKeyId(), org.GetPubPrivKeyId(), requesterService.GetPrivateKeyHashSymKeyId(), requesterService.GetPubPrivKeyId(), assetKeyID}
			pathExists, err = key_mgmt.VerifyAccessPath(stub, keyPath)
			if err != nil {
				logger.Errorf("KeyPath verification failed")
				return nil, errors.Wrap(err, "KeyPath verification failed")
			}

			if pathExists {
				fmt.Println()
				fmt.Println("Option 2f path exists: ", keyPath)
				fmt.Println()
				return keyPath, nil
			}
		}
	}

	// ========================================================================================
	if assetType == OwnerDataNamespace { // data
		// Option 3a: caller has access via contract, caller is contract requester
		// Key path: [caller private key ID, datatype sym key ID, asset key ID]
		datatypeKeyID := datatype.GetDatatypeKeyID(assetData.Datatypes[0], assetData.OwnerIds[0])
		keyPath = []string{caller.GetPubPrivKeyId(), datatypeKeyID, assetKeyID}

		pathExists, err = key_mgmt.VerifyAccessPath(stub, keyPath)
		if err != nil {
			logger.Errorf("KeyPath verification failed")
			return nil, errors.Wrap(err, "KeyPath verification failed")
		}

		if pathExists {
			fmt.Println()
			fmt.Println("Option 3a path exists: ", keyPath)
			fmt.Println()
			return keyPath, nil
		}

		// Option 3b: caller has access via consent, caller is target
		// Key path: [caller private key ID, consent asset key ID, datatype sym key ID, asset key ID]
		// if assetType == "data_model.Consent", Key path: [caller private key ID, consent asset key ID], same as Option 1
		// Consent AssetKeyID is consentID
		// using assetData.Datatypes[0] as consent datatype ID, assetData.OwnerIds[0] as consent owner ID, and caller ID as consent target ID
		consentID := consent_mgmt.GetConsentID(assetData.Datatypes[0], caller.ID, assetData.OwnerIds[0])
		keyPath = []string{caller.GetPubPrivKeyId(), consentID, datatype.GetDatatypeKeyID(assetData.Datatypes[0], assetData.OwnerIds[0]), assetKeyID}
		// Verify key path exists in graph
		pathExists, err = key_mgmt.VerifyAccessPath(stub, keyPath)
		if err != nil {
			logger.Errorf("KeyPath verification failed")
			return nil, errors.Wrap(err, "KeyPath verification failed")
		}
		if pathExists {
			fmt.Println()
			fmt.Println("Option 3b path exists: ", keyPath)
			fmt.Println()
			return keyPath, nil
		}

		// ========================================================================================
		// Option 3c: caller is service admin of service, service has access via consent
		// Key path: [caller private key ID, service private key hash, service private key ID, consent asset key ID, datatype sym key ID, asset key ID]
		if len(options) >= 2 {
			orgID := options[0]
			consentTargetID := options[1]

			org := data_model.User{ID: orgID}
			target := data_model.User{ID: consentTargetID}
			consentID := consent_mgmt.GetConsentID(assetData.Datatypes[0], consentTargetID, assetData.OwnerIds[0])
			keyPath = []string{caller.GetPubPrivKeyId(), target.GetPrivateKeyHashSymKeyId(), target.GetPubPrivKeyId(), consentID, datatype.GetDatatypeKeyID(assetData.Datatypes[0], orgID), assetKeyID}

			// Verify key path exists in graph
			pathExists, err = key_mgmt.VerifyAccessPath(stub, keyPath)
			if err != nil {
				logger.Errorf("KeyPath verification failed")
				return nil, errors.Wrap(err, "KeyPath verification failed")
			}

			if pathExists {
				fmt.Println()
				fmt.Println("Option 3c path exists: ", keyPath)
				fmt.Println()
				return keyPath, nil
			}

			// ========================================================================================
			// Option 3d: caller is org admin of service, service has access via consent
			// Key path: [caller private key ID, org private key hash, org private key ID, service private key hash, service private key ID, consent asset key ID, datatype sym key ID, asset key ID]
			keyPath = []string{caller.GetPubPrivKeyId(), org.GetPrivateKeyHashSymKeyId(), org.GetPubPrivKeyId(), target.GetPrivateKeyHashSymKeyId(), target.GetPubPrivKeyId(), consentID, datatype.GetDatatypeKeyID(assetData.Datatypes[0], orgID), assetKeyID}

			// Verify key path exists in graph
			pathExists, err = key_mgmt.VerifyAccessPath(stub, keyPath)
			if err != nil {
				logger.Errorf("KeyPath verification failed")
				return nil, errors.Wrap(err, "KeyPath verification failed")
			}

			if pathExists {
				fmt.Println()
				fmt.Println("Option 3d path exists: ", keyPath)
				fmt.Println()
				return keyPath, nil
			}
		}
	}
	return []string{}, nil
}

// GetPrivateKeyPath returns a key path to the private key of a user asset given a valid sym key path
func GetPrivateKeyPath(symKeyPath []string) []string {
	return symKeyPath[:len(symKeyPath)-1]
}

// GetUserAssetSymAndPrivateKeyPaths returns symKey path and privateKey path from caller to a given user assets
func GetUserAssetSymAndPrivateKeyPaths(stub cached_stub.CachedStubInterface, caller data_model.User, userID string) ([]string, []string, error) {
	consentTargetUserAssetID := user_mgmt.GetUserAssetID(userID)
	solutionCaller := convertToSolutionUser(caller)
	symKeyPath := []string{}
	var err error
	if solutionCaller.SolutionInfo.IsOrgAdmin {
		symKeyPath, err = GetKeyPath(stub, caller, consentTargetUserAssetID, solutionCaller.Org)
		if err != nil {
			customErr := &GetKeyPathError{Caller: caller.ID, AssetID: consentTargetUserAssetID}
			logger.Errorf(customErr.Error())
			return []string{}, []string{}, errors.New(customErr.Error())
		}
	} else {
		symKeyPath, err = GetKeyPath(stub, caller, consentTargetUserAssetID)
		if err != nil {
			customErr := &GetKeyPathError{Caller: caller.ID, AssetID: consentTargetUserAssetID}
			logger.Errorf(customErr.Error())
			return []string{}, []string{}, errors.New(customErr.Error())
		}
	}

	if len(symKeyPath) <= 0 {
		errMsg := "Caller does not have access to consent target keys"
		logger.Errorf(errMsg)
		return []string{}, []string{}, errors.New(errMsg)
	}
	prvKeyPath := GetPrivateKeyPath(symKeyPath)

	return symKeyPath, prvKeyPath, nil
}

// GetDatatypeKeyPath returns keyPath from a caller to a datatype
// Optionally pass consent target ID if needed
func GetDatatypeKeyPath(stub cached_stub.CachedStubInterface, caller data_model.User, datatypeID string, ownerID string, options ...string) ([]string, error) {
	datatypeKeyID := datatype.GetDatatypeKeyID(datatypeID, ownerID)

	// Use key path to get key
	// ========================================================================================
	// Option 0: caller has direct access to datatype
	// Key path: [caller private key ID, datatype key ID]
	keyPath := []string{caller.GetPubPrivKeyId(), datatypeKeyID}

	// Verify key path exists in graph
	pathExists, err := key_mgmt.VerifyAccessPath(stub, keyPath)
	if err != nil {
		logger.Errorf("KeyPath verification failed")
		return nil, errors.Wrap(err, "KeyPath verification failed")
	}

	if pathExists {
		fmt.Println()
		fmt.Println("Option 0 path exists: ", keyPath)
		fmt.Println()
		return keyPath, nil
	}

	// ========================================================================================
	// Option 1: caller has direct access to datatype / caller is owner
	// Key path: [caller private key ID, caller sym key ID, datatype key ID]
	keyPath = []string{caller.GetPubPrivKeyId(), caller.GetSymKeyId(), datatypeKeyID}

	// Verify key path exists in graph
	pathExists, err = key_mgmt.VerifyAccessPath(stub, keyPath)
	if err != nil {
		logger.Errorf("KeyPath verification failed")
		return nil, errors.Wrap(err, "KeyPath verification failed")
	}

	if pathExists {
		fmt.Println()
		fmt.Println("Option 1 path exists: ", keyPath)
		fmt.Println()
		return keyPath, nil
	}

	// ========================================================================================
	// Option 2: caller is direct admin of owner
	// Key path: [caller private key ID, owner private key hash, owner private key ID, asset key ID]
	owner := data_model.User{ID: ownerID}
	keyPath = []string{caller.GetPubPrivKeyId(), owner.GetPrivateKeyHashSymKeyId(), owner.GetPubPrivKeyId(), owner.GetSymKeyId(), datatypeKeyID}

	// Verify key path exists in graph
	pathExists, err = key_mgmt.VerifyAccessPath(stub, keyPath)
	if err != nil {
		logger.Errorf("KeyPath verification failed")
		return nil, errors.Wrap(err, "KeyPath verification failed")
	}

	if pathExists {
		fmt.Println()
		fmt.Println("Option 2a path exists: ", keyPath)
		fmt.Println()
		return keyPath, nil
	}

	// ========================================================================================
	// Option 3: caller is org admin of owner
	// Key path: [caller private key ID, org private key hash, org private key ID, owner private key hash, owner private key ID, owner sym key ID, datatype key ID]
	solutionCaller := convertToSolutionUser(caller)
	org := data_model.User{ID: solutionCaller.Org}

	if solutionCaller.SolutionInfo.IsOrgAdmin {
		keyPath = []string{caller.GetPubPrivKeyId(), org.GetPrivateKeyHashSymKeyId(), org.GetPubPrivKeyId(), owner.GetPrivateKeyHashSymKeyId(), owner.GetPubPrivKeyId(), owner.GetSymKeyId(), datatypeKeyID}
		pathExists, err = key_mgmt.VerifyAccessPath(stub, keyPath)
		if err != nil {
			logger.Errorf("KeyPath verification failed")
			return nil, errors.Wrap(err, "KeyPath verification failed")
		}

		if pathExists {
			fmt.Println()
			fmt.Println("Option 3 path exists: ", keyPath)
			fmt.Println()
			return keyPath, nil
		}
	}

	// ========================================================================================
	// Option 4: caller is target, has access through consent
	// Key path: [caller private key ID, consentKeyID, datatypeKeyID]
	// consentKeyID = consentID
	consentID := consent_mgmt.GetConsentID(datatypeID, caller.ID, ownerID)
	keyPath = []string{caller.GetPubPrivKeyId(), consentID, datatypeKeyID}

	// Verify key path exists in graph
	pathExists, err = key_mgmt.VerifyAccessPath(stub, keyPath)
	if err != nil {
		logger.Errorf("KeyPath verification failed")
		return nil, errors.Wrap(err, "KeyPath verification failed")
	}

	if pathExists {
		fmt.Println()
		fmt.Println("Option 4 path exists: ", keyPath)
		fmt.Println()
		return keyPath, nil
	}

	// ========================================================================================
	// Option 5: caller is service admin of service, service has access via consent
	// Key path: [caller private key ID, service private key hash, service private key ID, consent asset key ID, datatype sym key ID, asset key ID]
	if len(options) > 0 {
		consentTargetID := options[0]

		target := data_model.User{ID: consentTargetID}
		consentID = consent_mgmt.GetConsentID(datatypeID, consentTargetID, ownerID)
		keyPath = []string{caller.GetPubPrivKeyId(), target.GetPrivateKeyHashSymKeyId(), target.GetPubPrivKeyId(), consentID, datatypeKeyID}

		// Verify key path exists in graph
		pathExists, err = key_mgmt.VerifyAccessPath(stub, keyPath)
		if err != nil {
			logger.Errorf("KeyPath verification failed")
			return nil, errors.Wrap(err, "KeyPath verification failed")
		}

		if pathExists {
			fmt.Println()
			fmt.Println("Option 5 path exists: ", keyPath)
			fmt.Println()
			return keyPath, nil
		}

		// ========================================================================================
		// Option 6: caller is org admin of service, service has access via consent
		// Key path: [caller private key ID, org private key hash, org private key ID, service private key hash, service private key ID, consent asset key ID, datatype sym key ID, asset key ID]
		keyPath = []string{caller.GetPubPrivKeyId(), org.GetPrivateKeyHashSymKeyId(), org.GetPubPrivKeyId(), target.GetPrivateKeyHashSymKeyId(), target.GetPubPrivKeyId(), consentID, datatypeKeyID}

		// Verify key path exists in graph
		pathExists, err = key_mgmt.VerifyAccessPath(stub, keyPath)
		if err != nil {
			logger.Errorf("KeyPath verification failed")
			return nil, errors.Wrap(err, "KeyPath verification failed")
		}

		if pathExists {
			fmt.Println()
			fmt.Println("Option 6 path exists: ", keyPath)
			fmt.Println()
			return keyPath, nil
		}
	}

	return []string{}, nil
}

// Admin of a target service can get caller object of the service.
// Admin of a service can perform actions on behalf of this service.
func GetServiceCaller(stub cached_stub.CachedStubInterface, caller data_model.User, target string) (data_model.User, error) {
	solutionCaller := convertToSolutionUser(caller)

	callerObj := data_model.User{}

	if solutionCaller.SolutionInfo.IsOrgAdmin {
		orgCaller, err := user_mgmt.GetUserData(stub, caller, solutionCaller.Org, true, false)
		if err != nil {
			customErr := &GetOrgError{Org: solutionCaller.Org}
			logger.Errorf("%v: %v", customErr, err)
			return data_model.User{}, errors.Wrap(err, customErr.Error())
		}

		if orgCaller.PrivateKey == nil {
			errMsg := "Caller does not have access to org private key"
			logger.Errorf(errMsg)
			return data_model.User{}, errors.New(errMsg)
		} else {
			callerObj = orgCaller
		}
	} else if utils.InList(solutionCaller.SolutionInfo.Services, target) {
		// if caller is service admin, use service as caller
		// construct new caller object representing service itself
		serviceCaller, err := user_mgmt.GetUserData(stub, caller, target, true, false)
		if err != nil {
			customErr := &GetOrgError{Org: target}
			logger.Errorf("%v: %v", customErr, err)
			return data_model.User{}, errors.Wrap(err, customErr.Error())
		}

		if serviceCaller.PrivateKey == nil {
			logger.Errorf("Caller does not have access to service private key")
			return data_model.User{}, errors.New("Caller does not have access to service private key")
		} else {
			callerObj = serviceCaller
		}
	} else {
		callerObj = caller
	}

	return callerObj, nil
}
