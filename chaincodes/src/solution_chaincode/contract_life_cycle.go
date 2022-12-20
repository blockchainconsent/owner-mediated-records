// /*******************************************************************************
//  *
//  *
 * (c) Copyright Merative US L.P. and others 2020-2022 
 *
 * SPDX-Licence-Identifier: Apache 2.0
//  *
//  *******************************************************************************/

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
	"strconv"
	"time"

	"github.com/pkg/errors"
)

const IndexContract = "ContractTable"
const ContractAssetNamespace = "ContractAsset"

// ContractDetail object
type ContractDetail struct {
	ContractID          string      `json:"contract_id"`
	ContractDetailType  string      `json:"contract_detail_type"`
	ContractDetailTerms interface{} `json:"contract_detail_terms"`
	CreateDate          int64       `json:"create_date"`
	CreatedBy           string      `json:"created_by"`
}

// Contract object
type Contract struct {
	ContractID         string           `json:"contract_id"`
	OwnerOrgID         string           `json:"owner_org_id"`
	OwnerServiceID     string           `json:"owner_service_id"`
	RequesterOrgID     string           `json:"requester_org_id"`
	RequesterServiceID string           `json:"requester_service_id"`
	ContractTerms      interface{}      `json:"contract_terms"`
	State              string           `json:"state"`
	CreateDate         int64            `json:"create_date"`
	UpdateDate         int64            `json:"update_date"`
	ContractDetails    []ContractDetail `json:"contract_details"`
	PaymentRequired    string           `json:"payment_required"`
	PaymentVerified    string           `json:"payment_verified"`
	MaxNumDownload     int              `json:"max_num_download"`
	NumDownload        int              `json:"num_download"`
}

// ContractLog object
type ContractLog struct {
	Contract         string      `json:"contract"`
	OwnerService     string      `json:"owner_service"`
	RequesterService string      `json:"requester_service"`
	OwnerOrg         string      `json:"owner_org"`
	RequesterOrg     string      `json:"requester_org"`
	Datatype         string      `json:"datatype"`
	Data             interface{} `json:"data"`
}

// ContractPublicData consists of a contract's public fields
type ContractPublicData struct {
	OwnerServiceID     string `json:"owner_service_id"`
	RequesterServiceID string `json:"requester_service_id"`
}

// CreateContract creates a new contract asset
// Can be called by both contract requester and contract owner
// Returns error if contract already exists
//
// 1) Validate contract fields and sym key
// 2) Verify caller has permission
// 3) Store contract as asset and use contract sym key as asset key
// 4) Encrypt contract sym key with owner pub key and requester pub key
//
// args = [ contract, contractKey ]
func CreateContract(stub cached_stub.CachedStubInterface, caller data_model.User, args []string) ([]byte, error) {
	defer utils.ExitFnLog(utils.EnterFnLog())
	logger.Debugf("args: %v", args)

	if len(args) != 2 {
		customErr := &custom_errors.LengthCheckingError{Type: "CreateContract arguments length"}
		logger.Errorf(customErr.Error())
		return nil, errors.New(customErr.Error())
	}

	// ==============================================================
	// Validation
	// ==============================================================
	var contract = Contract{}
	contractBytes := []byte(args[0])
	err := json.Unmarshal(contractBytes, &contract)
	if err != nil {
		customErr := &custom_errors.UnmarshalError{Type: "contract"}
		logger.Errorf("%v: %v", customErr, err)
		return nil, errors.Wrap(err, customErr.Error())
	}

	// Validate contract ID
	if utils.IsStringEmpty(contract.ContractID) {
		customErr := &custom_errors.LengthCheckingError{Type: "contract.ContractID"}
		logger.Errorf(customErr.Error())
		return nil, errors.WithStack(customErr)
	}

	// Validate owner org ID
	if utils.IsStringEmpty(contract.OwnerOrgID) {
		customErr := &custom_errors.LengthCheckingError{Type: "contract.OwnerOrgID"}
		logger.Errorf(customErr.Error())
		return nil, errors.WithStack(customErr)
	}

	ownerOrg, err := user_mgmt.GetUserData(stub, caller, contract.OwnerOrgID, false, false)
	if err != nil {
		customErr := &GetUserError{User: contract.OwnerOrgID}
		logger.Errorf("%v: %v", customErr, err)
		return nil, errors.Wrap(err, customErr.Error())
	}

	if utils.IsStringEmpty(ownerOrg.ID) {
		customErr := &custom_errors.LengthCheckingError{Type: "ownerOrg.ID"}
		logger.Errorf(customErr.Error())
		return nil, errors.WithStack(customErr)
	}

	// Validate owner service ID
	if utils.IsStringEmpty(contract.OwnerServiceID) {
		customErr := &custom_errors.LengthCheckingError{Type: "contract.OwnerServiceID"}
		logger.Errorf(customErr.Error())
		return nil, errors.WithStack(customErr)
	}

	ownerService, err := GetServiceInternal(stub, caller, contract.OwnerServiceID, false)
	if err != nil {
		customErr := &GetServiceError{Service: contract.OwnerServiceID}
		logger.Errorf("%v: %v", customErr, err)
		return nil, errors.Wrap(err, customErr.Error())
	}

	if utils.IsStringEmpty(ownerService.ServiceID) {
		customErr := &custom_errors.LengthCheckingError{Type: "ownerService.ServiceID"}
		logger.Errorf(customErr.Error())
		return nil, errors.WithStack(customErr)
	}

	// Return error if there is already a contract with this ID
	solutionCaller := convertToSolutionUser(caller)
	existingContract := Contract{}
	if solutionCaller.SolutionInfo.IsOrgAdmin {
		existingContract, err = GetContractInternal(stub, caller, contract.ContractID, solutionCaller.Org)
		if err != nil {
			customErr := &GetContractError{ContractID: contract.ContractID}
			logger.Errorf(customErr.Error())
			return nil, errors.New(customErr.Error())
		}
	} else {
		existingContract, err = GetContractInternal(stub, caller, contract.ContractID)
		if err != nil {
			customErr := &GetContractError{ContractID: contract.ContractID}
			logger.Errorf(customErr.Error())
			return nil, errors.New(customErr.Error())
		}
	}

	if !utils.IsStringEmpty(existingContract.ContractID) {
		logger.Errorf("A contract with this ID already exists")
		return nil, errors.New("A contract with this ID already exists")
	}

	// Validate contract terms
	contractTermsBytes, err := json.Marshal(&contract.ContractTerms)
	if err != nil {
		logger.Errorf("Invalid contract terms")
		return nil, errors.New("Invalid contract terms")
	}

	if string(contractTermsBytes) == "" || string(contractTermsBytes) == "{}" {
		contract.ContractTerms = ownerService.Terms
		logger.Debug("Contract terms copied from owner service terms")
	}

	// Validate requester org ID
	if utils.IsStringEmpty(contract.RequesterOrgID) {
		customErr := &custom_errors.LengthCheckingError{Type: "contract.RequesterOrgID"}
		logger.Errorf(customErr.Error())
		return nil, errors.WithStack(customErr)
	}

	requesterOrg, err := user_mgmt.GetUserData(stub, caller, contract.RequesterOrgID, false, false)
	if err != nil {
		customErr := &GetUserError{User: contract.RequesterOrgID}
		logger.Errorf("%v: %v", customErr, err)
		return nil, errors.Wrap(err, customErr.Error())
	}

	if utils.IsStringEmpty(requesterOrg.ID) {
		customErr := &custom_errors.LengthCheckingError{Type: "requesterOrg.ID"}
		logger.Errorf(customErr.Error())
		return nil, errors.WithStack(customErr)
	}

	// if owner org is same as requester org, throw error as contract cannot be created for same org
	if contract.OwnerOrgID == contract.RequesterOrgID {
		logger.Errorf("Contract cannot be created for the same org")
		return nil, errors.New("Contract cannot be created for the same org")
	}

	// Validate requester service ID
	if utils.IsStringEmpty(contract.RequesterServiceID) {
		customErr := &custom_errors.LengthCheckingError{Type: "contract.RequesterServiceID"}
		logger.Errorf(customErr.Error())
		return nil, errors.WithStack(customErr)
	}

	requesterService, err := GetServiceInternal(stub, caller, contract.RequesterServiceID, false)
	if err != nil {
		customErr := &GetServiceError{Service: contract.RequesterServiceID}
		logger.Errorf("%v: %v", customErr, err)
		return nil, errors.Wrap(err, customErr.Error())
	}

	if utils.IsStringEmpty(requesterService.ServiceID) {
		customErr := &custom_errors.LengthCheckingError{Type: "requesterService.ServiceID"}
		logger.Errorf(customErr.Error())
		return nil, errors.WithStack(customErr)
	}

	// Check create date is within 10 mins of current time
	currTime := time.Now().Unix()
	if currTime-contract.CreateDate > 10*60 || currTime-contract.CreateDate < -10*60 {
		logger.Errorf("Invalid create date (current time: %v)  %v", currTime, contract.CreateDate)
		return nil, errors.New("Invalid create date, not within possible time range")
	}

	// Validate contract payment required
	if contract.PaymentRequired != "yes" && contract.PaymentRequired != "no" {
		logger.Errorf("Invalid contract payment required field (must be yes or no): %v", contract.PaymentRequired)
		return nil, errors.New("Invalid contract payment required field, must be yes or no")
	}

	// Validate contract sym key
	contractKey := data_model.Key{ID: key_mgmt.GetSymKeyId(contract.ContractID), Type: key_mgmt.KEY_TYPE_SYM}
	contractKey.KeyBytes, err = crypto.ParseSymKeyB64(args[1])
	if err != nil {
		logger.Errorf("Invalid contractKey")
		return nil, errors.Wrap(err, "Invalid contractKey")
	}

	if contractKey.KeyBytes == nil {
		logger.Errorf("Invalid contractKey")
		return nil, errors.New("Invalid contractKey")
	}

	// Set MaxNumDownload, NumDownload, PaymentVerified, UpdateDate
	contract.MaxNumDownload = 0
	contract.NumDownload = 0
	contract.PaymentVerified = "no"
	contract.PaymentRequired = ownerService.PaymentRequired
	contract.UpdateDate = contract.CreateDate

	//construct contract detail
	contractDetail := ContractDetail{}
	contractDetail.ContractID = contract.ContractID
	contractDetail.ContractDetailType = "request"
	contractDetail.ContractDetailTerms = contract.ContractTerms
	contractDetail.CreateDate = contract.CreateDate
	contractDetail.CreatedBy = caller.ID
	contract.ContractDetails = append(contract.ContractDetails, contractDetail)

	// ==============================================================
	// Change caller
	// ==============================================================
	callerObj := caller
	isAdminRequesterService := false
	// caller is org admin of contract owner org || caller is a service admin of contract owner service
	if (solutionCaller.SolutionInfo.IsOrgAdmin && solutionCaller.Org == contract.OwnerOrgID) || utils.InList(solutionCaller.SolutionInfo.Services, contract.OwnerServiceID) {
		// If caller is org admin, then have to use key paths
		if solutionCaller.SolutionInfo.IsOrgAdmin {
			symKeyPath, prvKeyPath, err := GetUserAssetSymAndPrivateKeyPaths(stub, caller, contract.OwnerServiceID)
			if err != nil {
				logger.Errorf("Failed to get symKeyPath and prvKeyPath for user asset")
				return nil, errors.Wrap(err, "Failed to get symKeyPath and prvKeyPath for user asset")
			}
			// act as contract owner service
			callerObj, err = user_mgmt.GetUserData(stub, caller, contract.OwnerServiceID, true, false, symKeyPath, prvKeyPath)
			if err != nil {
				customErr := &GetUserError{User: contract.OwnerServiceID}
				logger.Errorf("%v: %v", customErr, err)
				return nil, errors.Wrap(err, customErr.Error())
			}
		} else {
			callerObj, err = user_mgmt.GetUserData(stub, caller, contract.OwnerServiceID, true, false)
			if err != nil {
				customErr := &GetUserError{User: contract.OwnerServiceID}
				logger.Errorf("%v: %v", customErr, err)
				return nil, errors.Wrap(err, customErr.Error())
			}
		}
	} else {
		// caller must be either org admin of contract requester org || caller is a service admin of contract requester service
		// If caller is org admin, then have to use key paths
		if solutionCaller.SolutionInfo.IsOrgAdmin {
			symKeyPath, prvKeyPath, err := GetUserAssetSymAndPrivateKeyPaths(stub, caller, contract.RequesterServiceID)
			if err != nil {
				logger.Errorf("Failed to get symKeyPath and prvKeyPath for user asset")
				return nil, errors.Wrap(err, "Failed to get symKeyPath and prvKeyPath for user asset")
			}
			callerObj, err = user_mgmt.GetUserData(stub, caller, contract.RequesterServiceID, true, false, symKeyPath, prvKeyPath)
			if err != nil {
				customErr := &GetUserError{User: contract.RequesterServiceID}
				logger.Errorf("%v: %v", customErr, err)
				return nil, errors.Wrap(err, customErr.Error())
			}
		} else {
			callerObj, err = user_mgmt.GetUserData(stub, caller, contract.RequesterServiceID, true, false)
			if err != nil {
				customErr := &GetUserError{User: contract.RequesterServiceID}
				logger.Errorf("%v: %v", customErr, err)
				return nil, errors.Wrap(err, customErr.Error())
			}
		}

		isAdminRequesterService = true
	}

	if callerObj.PrivateKey == nil {
		logger.Errorf("Caller does not have access to owner private key")
		return nil, errors.New("Caller does not have access to owner private key")
	}

	// Update contract state
	if isAdminRequesterService {
		contract.State = "requested"
	} else { // isAdminOwnerService
		contract.State = "contractReady"
	}

	// ==============================================================
	// Save contract as asset
	// ==============================================================

	// 2nd parameter passed to convertContractToAsset function is asset ownerlist, which contains
	// either contract requester service or contract owner service
	contractAsset := data_model.Asset{}
	ownersList := []string{callerObj.ID}
	contractAsset, err = convertContractToAsset(contract, ownersList)
	if err != nil {
		customErr := &ConvertToAssetError{Asset: "contractAsset"}
		logger.Errorf("%v: %v", customErr, err)
		return nil, errors.Wrap(err, customErr.Error())
	}
	// contractAsset.AssetKeyId = contractKey.ID

	assetManager := asset_mgmt.GetAssetManager(stub, callerObj)
	err = assetManager.AddAsset(contractAsset, contractKey, true)
	if err != nil {
		customErr := &PutAssetError{Asset: contract.ContractID}
		logger.Errorf("%v: %v", customErr, err)
		return nil, errors.Wrap(err, customErr.Error())
	}

	// ==============================================================
	// Give write access to contract owner/requester
	// ==============================================================
	accessControl := data_model.AccessControl{}
	if isAdminRequesterService {
		accessControl.UserId = contract.OwnerServiceID
	} else {
		accessControl.UserId = contract.RequesterServiceID
	}
	accessControl.AssetId = contractAsset.AssetId
	accessControl.Access = user_access_ctrl.ACCESS_WRITE
	accessControl.AssetKey = &contractKey

	err = assetManager.AddAccessToAsset(accessControl, true)
	if err != nil {
		custom_err := &custom_errors.AddAccessError{Key: contractKey.ID}
		logger.Errorf("%v: %v", custom_err, err)
		return nil, errors.Wrap(err, custom_err.Error())
	}

	// ==============================================================
	// Logging
	// ==============================================================
	contractLogSymKey := GetLogSymKeyFromKey(contractKey)

	contractLog := ContractLog{Contract: contract.ContractID, OwnerService: contract.OwnerServiceID, RequesterService: contract.RequesterServiceID, OwnerOrg: contract.OwnerOrgID, RequesterOrg: contract.RequesterOrgID}
	solutionLog := SolutionLog{
		TransactionID: stub.GetTxID(),
		Namespace:     "OMR",
		FunctionName:  "CreateContract",
		CallerID:      caller.ID,
		Timestamp:     contract.CreateDate,
		Data:          contractLog}

	err = AddLogWithParams(stub, callerObj, solutionLog, contractLogSymKey)
	if err != nil {
		customErr := &AddSolutionLogError{FunctionName: solutionLog.FunctionName}
		logger.Errorf("%v: %v", customErr, err)
		return nil, errors.Wrap(err, customErr.Error())
	}

	// =============================================================================
	// Add access from owner/requester service log sym key to contract log sym key
	// =============================================================================

	// callerObj is either requester service or owner service
	serviceLogSymKey := callerObj.GetLogSymKey()
	userAccessManager := user_access_ctrl.GetUserAccessManager(stub, callerObj)
	err = userAccessManager.AddAccessByKey(serviceLogSymKey, contractLogSymKey)
	if err != nil {
		customErr := &custom_errors.AddAccessError{Key: "service log sym key to contract log sym key"}
		logger.Errorf("%v: %v", customErr, err)
		return nil, errors.Wrap(err, customErr.Error())
	}

	return nil, nil
}

// GetContractInternal is the internal function for getting a single contract
// Returns empty contract if not found, does not return error if no contract found
// Optional parameter: caller org ID, required if caller is an org admin
func GetContractInternal(stub cached_stub.CachedStubInterface, caller data_model.User, contractID string, options ...string) (Contract, error) {
	defer utils.ExitFnLog(utils.EnterFnLog())

	assetManager := asset_mgmt.GetAssetManager(stub, caller)
	contractAssetID := asset_mgmt.GetAssetId(ContractAssetNamespace, contractID)

	keyPath, err := GetKeyPath(stub, caller, contractAssetID, options...)
	if err != nil {
		customErr := &GetKeyPathError{Caller: caller.ID, AssetID: contractAssetID}
		logger.Errorf(customErr.Error())
		return Contract{}, errors.New(customErr.Error())
	}

	if len(keyPath) <= 0 {
		return Contract{}, nil
	}

	contractKey, err := assetManager.GetAssetKey(contractAssetID, keyPath)
	if err != nil {
		logger.Errorf("Failed to get contractKey: %v", err)
		return Contract{}, errors.Wrap(err, "Failed to get contractKey")
	}

	contractAsset, err := assetManager.GetAsset(contractAssetID, contractKey)
	if err != nil {
		customErr := &custom_errors.GetAssetDataError{AssetId: contractAssetID}
		logger.Errorf("%v: %v", customErr, err)
		return Contract{}, errors.Wrap(err, customErr.Error())
	}

	if data_model.IsEncryptedData(contractAsset.PrivateData) {
		logger.Error("Failed to verify to contract")
		return Contract{}, errors.New("Failed to verify access to contract")
	}

	return convertContractFromAsset(contractAsset), nil
}

// GetContract returns a contract object given contract ID
// args = [ contractID ]
// returns empty contract if no contract is found
func GetContract(stub cached_stub.CachedStubInterface, caller data_model.User, args []string) ([]byte, error) {
	defer utils.ExitFnLog(utils.EnterFnLog())
	logger.Debugf("args: %v", args)

	if len(args) != 1 {
		customErr := &custom_errors.LengthCheckingError{Type: "GetContract arguments length"}
		logger.Errorf(customErr.Error())
		return nil, errors.New(customErr.Error())
	}

	contractID := args[0]
	if utils.IsStringEmpty(contractID) {
		customErr := &custom_errors.LengthCheckingError{Type: "contractID"}
		logger.Errorf(customErr.Error())
		return nil, errors.WithStack(customErr)
	}

	solutionCaller := convertToSolutionUser(caller)
	contract := Contract{}
	var err error
	if solutionCaller.SolutionInfo.IsOrgAdmin {
		contract, err = GetContractInternal(stub, caller, contractID, solutionCaller.Org)
		if err != nil {
			customErr := &GetContractError{ContractID: contractID}
			logger.Errorf(customErr.Error())
			return nil, errors.New(customErr.Error())
		}
	} else {
		contract, err = GetContractInternal(stub, caller, contractID)
		if err != nil {
			customErr := &GetContractError{ContractID: contractID}
			logger.Errorf(customErr.Error())
			return nil, errors.New(customErr.Error())
		}
	}

	return json.Marshal(contract)
}

// AddContractDetailDownload updates contract after downloading owner data as requester
// This function is needed because this is an invoke operation, and download is query
// so we have to call this function from JS side after query completes
// args = [contactId, encContractStr, datatype]
func AddContractDetailDownload(stub cached_stub.CachedStubInterface, caller data_model.User, args []string) ([]byte, error) {
	defer utils.ExitFnLog(utils.EnterFnLog())
	logger.Debugf("args: %v", args)

	if len(args) != 3 {
		customErr := &custom_errors.LengthCheckingError{Type: "AddContractDetailDownload arguments length"}
		logger.Errorf(customErr.Error())
		return nil, errors.New(customErr.Error())
	}

	// ==============================================================
	// Validation
	// ==============================================================

	contractID := args[0]
	if utils.IsStringEmpty(contractID) {
		customErr := &custom_errors.LengthCheckingError{Type: "contractID"}
		logger.Errorf(customErr.Error())
		return nil, errors.WithStack(customErr)
	}

	encContractStr := args[1]
	if utils.IsStringEmpty(encContractStr) {
		customErr := &custom_errors.LengthCheckingError{Type: "encContractStr"}
		logger.Errorf(customErr.Error())
		return nil, errors.WithStack(customErr)
	}

	datatypeID := args[2]
	if utils.IsStringEmpty(datatypeID) {
		customErr := &custom_errors.LengthCheckingError{Type: "datatypeID"}
		logger.Errorf(customErr.Error())
		return nil, errors.WithStack(customErr)
	}

	contractAssetID := asset_mgmt.GetAssetId(ContractAssetNamespace, contractID)
	contractAssetDataTmp, err := asset_mgmt.GetEncryptedAssetData(stub, contractAssetID)
	if err != nil {
		customErr := &custom_errors.GetAssetDataError{AssetId: contractAssetID}
		logger.Errorf("%v", customErr)
		return nil, errors.WithStack(customErr)
	}

	contractPublicData := ContractPublicData{}
	err = json.Unmarshal(contractAssetDataTmp.PublicData, &contractPublicData)
	if err != nil {
		customErr := &custom_errors.UnmarshalError{Type: "contract public data"}
		logger.Errorf("%v: %v", customErr.Error(), err)
		return nil, errors.Wrap(err, customErr.Error())
	}

	encContractBytes, err := crypto.DecodeStringB64(encContractStr)
	if err != nil {
		logger.Errorf("Failed to DecodeStringB64: %v", err)
		return nil, errors.New("Failed to DecodeStringB64: " + encContractStr)
	}

	// ==============================================================
	// Change caller
	// ==============================================================
	solutionCaller := convertToSolutionUser(caller)
	callerObj := caller
	// caller must be either org admin of contract requester org || caller is a service admin of contract requester service
	// If caller is org admin, then have to use key paths
	if solutionCaller.SolutionInfo.IsOrgAdmin {
		symKeyPath, prvKeyPath, err := GetUserAssetSymAndPrivateKeyPaths(stub, caller, contractPublicData.RequesterServiceID)
		if err != nil {
			logger.Errorf("Failed to get symKeyPath and prvKeyPath for user asset")
			return nil, errors.Wrap(err, "Failed to get symKeyPath and prvKeyPath for user asset")
		}
		callerObj, err = user_mgmt.GetUserData(stub, caller, contractPublicData.RequesterServiceID, true, false, symKeyPath, prvKeyPath)
		if err != nil {
			customErr := &GetUserError{User: contractPublicData.RequesterServiceID}
			logger.Errorf("%v: %v", customErr, err)
			return nil, errors.Wrap(err, customErr.Error())
		}
	} else {
		if caller.ID != contractPublicData.RequesterServiceID { // No need to get user object if caller is contract requester service itself
			callerObj, err = user_mgmt.GetUserData(stub, caller, contractPublicData.RequesterServiceID, true, false)
			if err != nil {
				customErr := &GetUserError{User: contractPublicData.RequesterServiceID}
				logger.Errorf("%v: %v", customErr, err)
				return nil, errors.Wrap(err, customErr.Error())
			}
		}
	}

	if callerObj.PrivateKey == nil {
		logger.Errorf("Caller does not have access to owner private key")
		return nil, errors.New("Caller does not have access to owner private key")
	}

	// ==============================================================
	// More validation
	// ==============================================================
	// Get contract key
	assetManager := asset_mgmt.GetAssetManager(stub, callerObj)
	keyPath, err := GetKeyPath(stub, callerObj, contractAssetID)
	if err != nil {
		customErr := &GetKeyPathError{Caller: callerObj.ID, AssetID: contractAssetID}
		logger.Errorf(customErr.Error())
		return nil, errors.New(customErr.Error())
	}

	if len(keyPath) <= 0 {
		customErr := &GetKeyPathError{Caller: callerObj.ID, AssetID: contractAssetID}
		logger.Errorf(customErr.Error())
		return nil, errors.New(customErr.Error())
	}

	contractKey, err := assetManager.GetAssetKey(contractAssetID, keyPath)
	if err != nil {
		logger.Errorf("Failed to get contractKey: %v", err)
		return nil, errors.Wrap(err, "Failed to get contractKey")
	}

	contractBytes, err := crypto.DecryptWithSymKey(contractKey.KeyBytes, encContractBytes)
	if err != nil {
		customErr := &custom_errors.DecryptionError{ToDecrypt: "encContractBytes", DecryptionKey: "contract key"}
		logger.Infof("%v: %v", customErr, err)
		return nil, errors.Wrap(err, customErr.Error())
	}

	contract := Contract{}
	err = json.Unmarshal(contractBytes, &contract)
	if err != nil {
		customErr := &custom_errors.UnmarshalError{Type: "contractBytes"}
		logger.Errorf("%v: %v", customErr, err)
		return nil, errors.Wrap(err, customErr.Error())
	}

	//check that contract id of the encrypted contract coming from outside matches the id
	if contract.ContractID != contractID {
		logger.Error("Contract Id does not match the encrypted contract passed into the chaincode")
		return nil, errors.New("Contract Id does not match the encrypted contract passed into the chaincode")
	}

	// ==============================================================
	// Manage key relationships
	// ==============================================================
	if contract.MaxNumDownload <= contract.NumDownload {
		contract.State = "downloadDone"

		userAccessManager := user_access_ctrl.GetUserAccessManager(stub, callerObj)

		datatypeSymKeyPath, err := GetDatatypeKeyPath(stub, callerObj, datatypeID, contract.OwnerServiceID)
		if err != nil {
			customErr := &GetDatatypeKeyPathError{Caller: callerObj.ID, DatatypeID: datatypeID}
			logger.Errorf(customErr.Error())
			return nil, errors.New(customErr.Error())
		}

		// Remove access to datatype sym key
		datatypeSymKey, err := GetDatatypeSymKey(stub, callerObj, datatypeID, contract.OwnerServiceID, datatypeSymKeyPath)
		if err != nil {
			logger.Errorf("Failed to GetDatatypeSymKey: %v", err)
			return nil, errors.Wrap(err, "Failed to GetDatatypeSymKey")
		}

		requester := data_model.User{ID: contract.RequesterServiceID}
		err = userAccessManager.RemoveAccessByKey(requester.GetPubPrivKeyId(), datatypeSymKey.ID)
		if err != nil {
			customErr := &custom_errors.AddAccessError{Key: "requester service pub key to owner data key"}
			logger.Errorf("%v: %v", customErr, err)
			return nil, errors.Wrap(err, customErr.Error())
		}
	}

	// ==============================================================
	// Update contract
	// ==============================================================
	// get contract asset ownersList, because ownerList cannot change
	contractAssetData, err := asset_mgmt.GetEncryptedAssetData(stub, contractAssetID)
	if err != nil {
		customErr := &custom_errors.GetAssetDataError{AssetId: contractAssetID}
		logger.Errorf("%v", customErr)
		return nil, errors.WithStack(customErr)
	}

	// Convert contract object to Asset
	contractAsset, err := convertContractToAsset(contract, contractAssetData.OwnerIds)
	if err != nil {
		customErr := &ConvertToAssetError{Asset: "contractAsset"}
		logger.Errorf("%v: %v", customErr, err)
		return nil, errors.Wrap(err, customErr.Error())
	}

	err = assetManager.UpdateAsset(contractAsset, contractKey, true)
	if err != nil {
		customErr := &PutAssetError{Asset: contract.ContractID}
		logger.Errorf("%v: %v", customErr, err)
		return nil, errors.Wrap(err, customErr.Error())
	}

	// ==============================================================
	// Logging
	// ==============================================================
	contractLogSymKey := GetLogSymKeyFromKey(contractKey)

	contractLog := ContractLog{Contract: contract.ContractID, Datatype: datatypeID, OwnerService: contract.OwnerServiceID, RequesterService: contract.RequesterServiceID, OwnerOrg: contract.OwnerOrgID, RequesterOrg: contract.RequesterOrgID}
	solutionLog := SolutionLog{
		TransactionID: stub.GetTxID(),
		Namespace:     "OMR",
		FunctionName:  "DownloadOwnerDataAsRequester",
		CallerID:      caller.ID,
		Timestamp:     contract.UpdateDate,
		Data:          contractLog}

	err = AddLogWithParams(stub, callerObj, solutionLog, contractLogSymKey)
	if err != nil {
		customErr := &AddSolutionLogError{FunctionName: solutionLog.FunctionName}
		logger.Errorf("%v: %v", customErr, err)
		return nil, errors.Wrap(err, customErr.Error())
	}

	return nil, nil
}

// AddContractDetail adds contract detail for an existing contract
// Appends a ContractDetail object to the contract
// args = [contactId, contractStatus, terms, timestamp]
func AddContractDetail(stub cached_stub.CachedStubInterface, caller data_model.User, args []string) ([]byte, error) {
	defer utils.ExitFnLog(utils.EnterFnLog())
	logger.Debugf("args: %v", args)

	if len(args) != 4 {
		customErr := &custom_errors.LengthCheckingError{Type: "AddContractDetail arguments length"}
		logger.Errorf(customErr.Error())
		return nil, errors.New(customErr.Error())
	}

	// ==============================================================
	// Validation
	// ==============================================================
	contractID := args[0]
	if utils.IsStringEmpty(contractID) {
		customErr := &custom_errors.LengthCheckingError{Type: "contractID"}
		logger.Errorf(customErr.Error())
		return nil, errors.WithStack(customErr)
	}

	// Return error if a contract with this ID does not exist
	solutionCaller := convertToSolutionUser(caller)
	contract := Contract{}
	var err error
	if solutionCaller.SolutionInfo.IsOrgAdmin {
		contract, err = GetContractInternal(stub, caller, contractID, solutionCaller.Org)
		if err != nil {
			customErr := &GetContractError{ContractID: contractID}
			logger.Errorf(customErr.Error())
			return nil, errors.New(customErr.Error())
		}
	} else {
		contract, err = GetContractInternal(stub, caller, contractID)
		if err != nil {
			customErr := &GetContractError{ContractID: contractID}
			logger.Errorf(customErr.Error())
			return nil, errors.New(customErr.Error())
		}
	}

	if utils.IsStringEmpty(contract.ContractID) {
		customErr := &GetContractError{ContractID: contractID}
		logger.Errorf(customErr.Error())
		return nil, errors.New(customErr.Error())
	}

	// Validate contract status
	contractStatus := args[1]
	if utils.IsStringEmpty(contractStatus) {
		customErr := &custom_errors.LengthCheckingError{Type: "contractStatus"}
		logger.Errorf(customErr.Error())
		return nil, errors.WithStack(customErr)
	}

	// Validate terms
	var terms interface{}
	var addTerms = false
	termsBytes := []byte("{}")
	if args[2] != "" && args[2] != "{}" && args[2] != "[]" {
		addTerms = true
		termsBytes = []byte(args[2])
	}

	err = json.Unmarshal(termsBytes, &terms)
	if err != nil {
		customErr := &custom_errors.UnmarshalError{Type: "termsBytes"}
		logger.Errorf("%v: %v", customErr, err)
		return nil, errors.Wrap(err, customErr.Error())
	}

	// Validate timestamp
	timestamp, err := strconv.ParseInt(args[3], 10, 64)
	if err != nil {
		logger.Errorf("Error converting timestamp to type int64")
		return nil, errors.Wrap(err, "Error converting timestamp to type int64")
	}

	// ==============================================================
	// Change caller
	// ==============================================================
	callerObj := caller
	isAdminRequesterService := false
	// caller is org admin of contract owner org || caller is a service admin of contract owner service
	if (solutionCaller.SolutionInfo.IsOrgAdmin && solutionCaller.Org == contract.OwnerOrgID) || utils.InList(solutionCaller.SolutionInfo.Services, contract.OwnerServiceID) {
		// If caller is org admin, then have to use key paths
		if solutionCaller.SolutionInfo.IsOrgAdmin {
			symKeyPath, prvKeyPath, err := GetUserAssetSymAndPrivateKeyPaths(stub, caller, contract.OwnerServiceID)
			if err != nil {
				logger.Errorf("Failed to get symKeyPath and prvKeyPath for user asset")
				return nil, errors.Wrap(err, "Failed to get symKeyPath and prvKeyPath for user asset")
			}
			// act as contract owner service
			callerObj, err = user_mgmt.GetUserData(stub, caller, contract.OwnerServiceID, true, false, symKeyPath, prvKeyPath)
			if err != nil {
				customErr := &GetUserError{User: contract.OwnerServiceID}
				logger.Errorf("%v: %v", customErr, err)
				return nil, errors.Wrap(err, customErr.Error())
			}
		} else {
			callerObj, err = user_mgmt.GetUserData(stub, caller, contract.OwnerServiceID, true, false)
			if err != nil {
				customErr := &GetUserError{User: contract.OwnerServiceID}
				logger.Errorf("%v: %v", customErr, err)
				return nil, errors.Wrap(err, customErr.Error())
			}
		}
	} else {
		// caller must be either org admin of contract requester org || caller is a service admin of contract requester service
		// If caller is org admin, then have to use key paths
		if solutionCaller.SolutionInfo.IsOrgAdmin {
			symKeyPath, prvKeyPath, err := GetUserAssetSymAndPrivateKeyPaths(stub, caller, contract.RequesterServiceID)
			if err != nil {
				logger.Errorf("Failed to get symKeyPath and prvKeyPath for user asset")
				return nil, errors.Wrap(err, "Failed to get symKeyPath and prvKeyPath for user asset")
			}
			callerObj, err = user_mgmt.GetUserData(stub, caller, contract.RequesterServiceID, true, false, symKeyPath, prvKeyPath)
			if err != nil {
				customErr := &GetUserError{User: contract.RequesterServiceID}
				logger.Errorf("%v: %v", customErr, err)
				return nil, errors.Wrap(err, customErr.Error())
			}
		} else {
			callerObj, err = user_mgmt.GetUserData(stub, caller, contract.RequesterServiceID, true, false)
			if err != nil {
				customErr := &GetUserError{User: contract.RequesterServiceID}
				logger.Errorf("%v: %v", customErr, err)
				return nil, errors.Wrap(err, customErr.Error())
			}
		}

		isAdminRequesterService = true
	}

	if callerObj.PrivateKey == nil {
		logger.Errorf("Caller does not have access to owner private key")
		return nil, errors.New("Caller does not have access to owner private key")
	}

	// ==============================================================
	// Update contract status
	// ==============================================================

	var okAdd = false
	if contractStatus == "request" {
		if (contract.State == "new" || contract.State == "requested") && isAdminRequesterService {
			// If called by requester set status to "requested"
			okAdd = true
			contract.State = "requested"
		} else if (contract.State == "new" || contract.State == "requested") && !isAdminRequesterService {
			// If called by owner set status to "contractReady"
			okAdd = true
			contract.State = "contractReady"
		}
		// Handle "Add Terms" by both Requester and Owner
	} else if contractStatus == "terms" {
		if (contract.State == "requested" || contract.State == "contractReady") && isAdminRequesterService {
			// If called by requester set status to "requested"
			okAdd = true
			contract.State = "requested"
		} else if (contract.State == "requested" || contract.State == "contractReady") && !isAdminRequesterService {
			// If called by owner set status to "contractReady"
			okAdd = true
			contract.State = "contractReady"
		}
	} else if contractStatus == "sign" {
		addTerms = false
		if contract.State == "contractReady" && isAdminRequesterService {
			okAdd = true
			contract.State = "contractSigned"
		}
	} else if contractStatus == "payment" {
		addTerms = false
		if (contract.PaymentRequired == "yes") && (contract.State == "contractSigned" || contract.State == "paymentDone") && isAdminRequesterService {
			okAdd = true
			contract.State = "paymentDone"
		}
	} else if contractStatus == "verify" {
		addTerms = false
		if (contract.PaymentRequired == "yes") && (contract.State == "contractSigned" || contract.State == "paymentDone") && !isAdminRequesterService {
			okAdd = true
			contract.State = "paymentVerified"
			contract.PaymentVerified = "yes"
		}
	} else if contractStatus == "terminate" {
		addTerms = false
		if (contract.State == "new" || contract.State == "requested" || contract.State == "contractReady" || contract.State == "contractSigned" || contract.State == "paymentDone" || contract.State == "paymentVerified" || contract.State == "downloadReady" || contract.State == "downloadDone") && isAdminRequesterService {
			okAdd = true
			contract.State = "terminated"
		} else if (contract.State == "new" || contract.State == "requested" || contract.State == "contractReady" || contract.State == "contractSigned" || contract.State == "paymentDone" || contract.State == "paymentVerified" || contract.State == "downloadReady" || contract.State == "downloadDone") && !isAdminRequesterService {
			okAdd = true
			contract.State = "terminated"
		}
	} else {
		logger.Error("Invalid contract status: " + contractStatus + ". Contract state: " + contract.State)
		return nil, errors.New("Invalid contract status: " + contractStatus + ". Contract state: " + contract.State)
	}

	if !okAdd {
		logger.Error(caller.ID + " user is not allowed to change contract " + contractID + " when contract state is " + contract.State)
		return nil, errors.New(caller.ID + " user is not allowed to change contract " + contractID + " when contract state is " + contract.State)
	}

	// ==============================================================
	// Update contract asset
	// ==============================================================
	//construct contract detail
	contractDetail := ContractDetail{}
	contractDetail.ContractID = contractID
	contractDetail.ContractDetailType = contractStatus
	contractDetail.ContractDetailTerms = terms
	contractDetail.CreateDate = timestamp
	contractDetail.CreatedBy = caller.ID
	contract.ContractDetails = append(contract.ContractDetails, contractDetail)
	contract.UpdateDate = timestamp

	if addTerms {
		contract.ContractTerms = terms
	}

	// get contract asset ownersList, because ownerList cannot change
	contractAssetData, err := asset_mgmt.GetEncryptedAssetData(stub, asset_mgmt.GetAssetId(ContractAssetNamespace, contractID))
	if err != nil {
		customErr := &custom_errors.GetAssetDataError{AssetId: asset_mgmt.GetAssetId(ContractAssetNamespace, contractID)}
		logger.Errorf("%v", customErr)
		return nil, errors.WithStack(customErr)
	}

	// Convert contract object to Asset
	contractAsset, err := convertContractToAsset(contract, contractAssetData.OwnerIds)
	if err != nil {
		customErr := &ConvertToAssetError{Asset: "contractAsset"}
		logger.Errorf("%v: %v", customErr, err)
		return nil, errors.Wrap(err, customErr.Error())
	}
	// contractAsset.AssetKeyId = contract.ContractID

	assetManager := asset_mgmt.GetAssetManager(stub, callerObj)
	contractAssetID := asset_mgmt.GetAssetId(ContractAssetNamespace, contractID)
	keyPath, err := GetKeyPath(stub, callerObj, contractAssetID)
	if err != nil {
		customErr := &GetKeyPathError{Caller: caller.ID, AssetID: contractAssetID}
		logger.Errorf(customErr.Error())
		return nil, errors.New(customErr.Error())
	}

	if len(keyPath) <= 0 {
		customErr := &GetKeyPathError{Caller: caller.ID, AssetID: contractAssetID}
		logger.Errorf(customErr.Error())
		return nil, errors.New(customErr.Error())
	}

	contractKey, err := assetManager.GetAssetKey(contractAssetID, keyPath)
	if err != nil {
		logger.Errorf("Failed to get contractKey: %v", err)
		return nil, errors.Wrap(err, "Failed to get contractKey")
	}

	err = assetManager.UpdateAsset(contractAsset, contractKey, true)
	if err != nil {
		customErr := &PutAssetError{Asset: contract.ContractID}
		logger.Errorf("%v: %v", customErr, err)
		return nil, errors.Wrap(err, customErr.Error())
	}

	// ==============================================================
	// Logging
	// ==============================================================
	contractLogSymKey := GetLogSymKeyFromKey(contractKey)

	contractLog := ContractLog{Contract: contract.ContractID, OwnerService: contract.OwnerServiceID, RequesterService: contract.RequesterServiceID, OwnerOrg: contract.OwnerOrgID, RequesterOrg: contract.RequesterOrgID}
	solutionLog := SolutionLog{
		TransactionID: stub.GetTxID(),
		Namespace:     "OMR",
		FunctionName:  "AddContractDetail" + contract.State,
		CallerID:      caller.ID,
		Timestamp:     contract.UpdateDate,
		Data:          contractLog}

	err = AddLogWithParams(stub, callerObj, solutionLog, contractLogSymKey)
	if err != nil {
		customErr := &AddSolutionLogError{FunctionName: solutionLog.FunctionName}
		logger.Errorf("%v: %v", customErr, err)
		return nil, errors.Wrap(err, customErr.Error())
	}

	// =============================================================================
	// Add access from owner/requester service log sym key to contract log sym key
	// =============================================================================

	userAccessManager := user_access_ctrl.GetUserAccessManager(stub, callerObj)
	err = userAccessManager.AddAccessByKey(callerObj.GetLogSymKey(), contractLogSymKey)
	if err != nil {
		customErr := &custom_errors.AddAccessError{Key: "service log sym key to contract log sym key"}
		logger.Errorf("%v: %v", customErr, err)
		return nil, errors.Wrap(err, customErr.Error())
	}

	return nil, nil
}

// GivePermissionByContract gives a requester permission to download for a contract
// Can only be used by contract data owner
// args = [contractID, maxNumDownloadAllowed, timestamp, datatypeID]
// datatypeID = the datatype that requester wants data for
func GivePermissionByContract(stub cached_stub.CachedStubInterface, caller data_model.User, args []string) ([]byte, error) {
	defer utils.ExitFnLog(utils.EnterFnLog())
	logger.Debugf("args: %v", args)

	if len(args) != 4 {
		customErr := &custom_errors.LengthCheckingError{Type: "GivePermissionByContract arguments length"}
		logger.Errorf(customErr.Error())
		return nil, errors.New(customErr.Error())
	}

	// ==============================================================
	// Validation
	// ==============================================================

	// Validate contractID
	contractID := args[0]
	if utils.IsStringEmpty(contractID) {
		customErr := &custom_errors.LengthCheckingError{Type: "contractID"}
		logger.Errorf(customErr.Error())
		return nil, errors.WithStack(customErr)
	}

	// Return error if a contract with this ID does not exist
	solutionCaller := convertToSolutionUser(caller)
	contract := Contract{}
	var err error
	if solutionCaller.SolutionInfo.IsOrgAdmin {
		contract, err = GetContractInternal(stub, caller, contractID, solutionCaller.Org)
		if err != nil {
			customErr := &GetContractError{ContractID: contractID}
			logger.Errorf(customErr.Error())
			return nil, errors.New(customErr.Error())
		}
	} else {
		contract, err = GetContractInternal(stub, caller, contractID)
		if err != nil {
			customErr := &GetContractError{ContractID: contractID}
			logger.Errorf(customErr.Error())
			return nil, errors.New(customErr.Error())
		}
	}

	if utils.IsStringEmpty(contract.ContractID) {
		customErr := &GetContractError{ContractID: contractID}
		logger.Errorf(customErr.Error())
		return nil, errors.New(customErr.Error())
	}

	// Validate maxNumDownloadAllowed
	maxNumDownloadAllowed, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		logger.Errorf("Error converting maxNumDownloadAllowedB64 to type int64")
		return nil, errors.Wrap(err, "Error converting maxNumDownloadAllowedB64 to type int64")
	}

	if int(maxNumDownloadAllowed) <= contract.NumDownload {
		logger.Errorf("Number of downloads by requester is already %v", contract.NumDownload)
		return nil, errors.New("Number of downloads by requester is already " + string(contract.NumDownload))
	}

	if !((contract.State == "contractSigned" && contract.PaymentRequired == "no") ||
		(contract.State == "paymentVerified" && contract.PaymentRequired == "yes") ||
		contract.State == "downloadReady" || contract.State == "downloadDone") {
		logger.Errorf("The contract state of " + contract.State + " does not allow granting permission")
		return nil, errors.New("The contract state of " + contract.State + " does not allow granting permission")
	}

	// Validate timestamp
	timestamp, err := strconv.ParseInt(args[2], 10, 64)
	if err != nil {
		logger.Errorf("Error converting timestamp to type int64")
		return nil, errors.Wrap(err, "Error converting timestamp to type int64")
	}

	// Validate datatype
	datatypeID := args[3]
	if utils.IsStringEmpty(datatypeID) {
		customErr := &custom_errors.LengthCheckingError{Type: "datatypeID"}
		logger.Errorf(customErr.Error())
		return nil, errors.WithStack(customErr)
	}

	datatype, err := GetDatatypeWithParams(stub, caller, datatypeID)
	if err != nil {
		logger.Errorf("Failed to GetDatatypeWithParams: %v, %v", datatypeID, err)
		return nil, errors.Wrap(err, "Failed to GetDatatypeWithParams: "+datatypeID)
	}

	if utils.IsStringEmpty(datatype.DatatypeID) {
		customErr := &custom_errors.LengthCheckingError{Type: "datatype.DatatypeID"}
		logger.Errorf(customErr.Error())
		return nil, errors.WithStack(customErr)
	}

	// ==============================================================
	// Change caller
	// ==============================================================
	callerObj := caller
	if solutionCaller.Org != contract.OwnerOrgID {
		logger.Errorf("Caller org is not the same as owner org")
		return nil, errors.New("Caller org is not the same as owner org")
	}

	// If caller is org admin, then have to use key paths
	if solutionCaller.SolutionInfo.IsOrgAdmin {
		symKeyPath, prvKeyPath, err := GetUserAssetSymAndPrivateKeyPaths(stub, caller, contract.OwnerServiceID)
		if err != nil {
			logger.Errorf("Failed to get symKeyPath and prvKeyPath for user asset")
			return nil, errors.Wrap(err, "Failed to get symKeyPath and prvKeyPath for user asset")
		}
		// act as contract owner service
		callerObj, err = user_mgmt.GetUserData(stub, caller, contract.OwnerServiceID, true, false, symKeyPath, prvKeyPath)
		if err != nil {
			customErr := &GetUserError{User: contract.OwnerServiceID}
			logger.Errorf("%v: %v", customErr, err)
			return nil, errors.Wrap(err, customErr.Error())
		}
	} else {
		callerObj, err = user_mgmt.GetUserData(stub, caller, contract.OwnerServiceID, true, false)
		if err != nil {
			customErr := &GetUserError{User: contract.OwnerServiceID}
			logger.Errorf("%v: %v", customErr, err)
			return nil, errors.Wrap(err, customErr.Error())
		}
	}

	if callerObj.PrivateKey == nil {
		logger.Errorf("Caller does not have access to owner private key")
		return nil, errors.New("Caller does not have access to owner private key")
	}

	// ==============================================================
	// Update contract with contract detail
	// ==============================================================
	contractDetail := ContractDetail{}
	contractDetail.ContractID = contractID
	contractDetail.ContractDetailType = "permission"
	contractDetail.ContractDetailTerms = map[string]int{"previous_max_num_download": contract.MaxNumDownload, "max_num_download": int(maxNumDownloadAllowed)}
	contractDetail.CreateDate = timestamp
	contractDetail.CreatedBy = caller.ID

	contract.UpdateDate = timestamp
	contract.ContractDetails = append(contract.ContractDetails, contractDetail)
	contract.MaxNumDownload = int(maxNumDownloadAllowed)

	if contract.NumDownload >= contract.MaxNumDownload {
		contract.State = "downloadDone"
	} else {
		contract.State = "downloadReady"
	}

	// get contract asset ownersList, because ownerList cannot change
	contractAssetData, err := asset_mgmt.GetEncryptedAssetData(stub, asset_mgmt.GetAssetId(ContractAssetNamespace, contractID))
	if err != nil {
		customErr := &custom_errors.GetAssetDataError{AssetId: asset_mgmt.GetAssetId(ContractAssetNamespace, contractID)}
		logger.Errorf("%v", customErr)
		return nil, errors.WithStack(customErr)
	}

	// Convert contract object to Asset
	contractAsset, err := convertContractToAsset(contract, contractAssetData.OwnerIds)
	if err != nil {
		customErr := &ConvertToAssetError{Asset: "contractAsset"}
		logger.Errorf("%v: %v", customErr, err)
		return nil, errors.Wrap(err, customErr.Error())
	}

	assetManager := asset_mgmt.GetAssetManager(stub, callerObj)
	contractAssetID := asset_mgmt.GetAssetId(ContractAssetNamespace, contractID)
	keyPath, err := GetKeyPath(stub, callerObj, contractAssetID)
	if err != nil {
		customErr := &GetKeyPathError{Caller: caller.ID, AssetID: contractAssetID}
		logger.Errorf(customErr.Error())
		return nil, errors.New(customErr.Error())
	}

	if len(keyPath) <= 0 {
		customErr := &GetKeyPathError{Caller: caller.ID, AssetID: contractAssetID}
		logger.Errorf(customErr.Error())
		return nil, errors.New(customErr.Error())
	}

	contractKey, err := assetManager.GetAssetKey(contractAssetID, keyPath)
	if err != nil {
		logger.Errorf("Failed to get contractKey: %v", err)
		return nil, errors.Wrap(err, "Failed to get contractKey")
	}

	err = assetManager.UpdateAsset(contractAsset, contractKey, true)
	if err != nil {
		customErr := &PutAssetError{Asset: contract.ContractID}
		logger.Errorf("%v: %v", customErr, err)
		return nil, errors.Wrap(err, customErr.Error())
	}

	// ==============================================================
	// Add access to contract requester to download owner data
	// ==============================================================

	datatypeSymKeyPath, err := GetDatatypeKeyPath(stub, callerObj, datatypeID, contract.OwnerServiceID)
	if err != nil {
		customErr := &GetDatatypeKeyPathError{Caller: callerObj.ID, DatatypeID: datatypeID}
		logger.Errorf(customErr.Error())
		return nil, errors.New(customErr.Error())
	}

	datatypeSymKey, err := GetDatatypeSymKey(stub, callerObj, datatypeID, contract.OwnerServiceID, datatypeSymKeyPath)
	if err != nil {
		logger.Errorf("Failed to GetDatatypeSymKey: %v", err)
		return nil, errors.Wrap(err, "Failed to GetDatatypeSymKey")
	}

	requester, err := user_mgmt.GetUserData(stub, callerObj, contract.RequesterServiceID, false, false)
	if err != nil {
		customErr := &GetUserError{User: contract.RequesterServiceID}
		logger.Errorf("%v: %v", customErr, err)
		return nil, errors.Wrap(err, customErr.Error())
	}
	userAccessManager := user_access_ctrl.GetUserAccessManager(stub, callerObj)
	err = userAccessManager.AddAccessByKey(requester.GetPublicKey(), datatypeSymKey)
	if err != nil {
		customErr := &custom_errors.AddAccessError{Key: "requester service pub key to owner data key"}
		logger.Errorf("%v: %v", customErr, err)
		return nil, errors.Wrap(err, customErr.Error())
	}

	// ==============================================================
	// Logging
	// ==============================================================
	contractLogSymKey := GetLogSymKeyFromKey(contractKey)

	contractLog := ContractLog{Contract: contract.ContractID, Datatype: datatypeID, OwnerService: contract.OwnerServiceID, RequesterService: contract.RequesterServiceID, OwnerOrg: contract.OwnerOrgID, RequesterOrg: contract.RequesterOrgID}
	solutionLog := SolutionLog{
		TransactionID: stub.GetTxID(),
		Namespace:     "OMR",
		FunctionName:  "GivePermissionByContract",
		CallerID:      caller.ID,
		Timestamp:     contract.UpdateDate,
		Data:          contractLog}

	err = AddLogWithParams(stub, callerObj, solutionLog, contractLogSymKey)
	if err != nil {
		customErr := &AddSolutionLogError{FunctionName: solutionLog.FunctionName}
		logger.Errorf("%v: %v", customErr, err)
		return nil, errors.Wrap(err, customErr.Error())
	}

	return nil, nil
}

// GetContractsInternal is the internal private function for getting contracts
func GetContractsInternal(stub cached_stub.CachedStubInterface, caller data_model.User, fieldNames []string, values []string) ([]Contract, error) {
	defer utils.ExitFnLog(utils.EnterFnLog())

	contracts := []Contract{}

	// Use index to find all consents
	iter, err := asset_mgmt.GetAssetManager(stub, caller).GetAssetIter(ContractAssetNamespace, IndexContract, fieldNames, values, values, true, false, KeyPathFunc, "", -1, nil)
	if err != nil {
		logger.Errorf("GetAssets failed: %v", err)
		return nil, errors.Wrap(err, "GetAssets failed")
	}

	// Iterate over all data and filter
	defer iter.Close()
	for iter.HasNext() {
		contractAsset, err := iter.Next()
		if err != nil {
			customErr := &custom_errors.IterError{}
			logger.Errorf("%v: %v", customErr, err)
			continue
		}

		// if data_model.IsEncryptedData(contractAsset.PrivateData) {
		// 	logger.Error("Failed to decrypt contract asset")
		// 	continue
		// }

		contract := convertContractFromAsset(contractAsset)
		contracts = append(contracts, contract)
	}

	return contracts, nil
}

// GetContractsAsOwner returns contracts as owner, filtered by state
// args = [ownerID, state]
// state is an optional parameter
func GetContractsAsOwner(stub cached_stub.CachedStubInterface, caller data_model.User, args []string) ([]byte, error) {
	defer utils.ExitFnLog(utils.EnterFnLog())
	logger.Debugf("args: %v", args)

	if len(args) != 2 {
		customErr := &custom_errors.LengthCheckingError{Type: "GetContractsAsOwner arguments length"}
		logger.Errorf(customErr.Error())
		return nil, errors.WithStack(customErr)
	}

	// ==============================================================
	// Validation
	// ==============================================================

	ownerID := args[0]
	if utils.IsStringEmpty(ownerID) {
		customErr := &custom_errors.LengthCheckingError{Type: "ownerID"}
		logger.Errorf(customErr.Error())
		return nil, errors.WithStack(customErr)
	}

	state := args[1]

	// ==============================================================
	// Get contracts
	// ==============================================================

	contracts := []Contract{}
	if utils.IsStringEmpty(state) {
		var err error
		contracts, err = GetContractsInternal(stub, caller, []string{"owner_service_id"}, []string{ownerID})
		if err != nil {
			customErr := &GetDatasError{FieldNames: []string{"owner_service_id", "state"}, Values: []string{ownerID, state}}
			logger.Errorf("%v: %v", customErr, err)
			return nil, errors.Wrap(err, customErr.Error())
		}
	} else {
		var err error
		contracts, err = GetContractsInternal(stub, caller, []string{"owner_service_id", "state"}, []string{ownerID, state})
		if err != nil {
			customErr := &GetDatasError{FieldNames: []string{"owner_service_id", "state"}, Values: []string{ownerID, state}}
			logger.Errorf("%v: %v", customErr, err)
			return nil, errors.Wrap(err, customErr.Error())
		}
	}

	// return contracts
	return json.Marshal(&contracts)
}

// GetContractsAsRequester returns contracts as requester, filtered by state
// args = [requesterID, state]
// state is an optional parameter
func GetContractsAsRequester(stub cached_stub.CachedStubInterface, caller data_model.User, args []string) ([]byte, error) {
	defer utils.ExitFnLog(utils.EnterFnLog())
	logger.Debugf("args: %v", args)

	if len(args) != 2 {
		customErr := &custom_errors.LengthCheckingError{Type: "GetContractsAsRequester arguments length"}
		logger.Errorf(customErr.Error())
		return nil, errors.WithStack(customErr)
	}

	// ==============================================================
	// Validation
	// ==============================================================

	requesterID := args[0]
	if utils.IsStringEmpty(requesterID) {
		customErr := &custom_errors.LengthCheckingError{Type: "requesterID"}
		logger.Errorf(customErr.Error())
		return nil, errors.WithStack(customErr)
	}

	state := args[1]

	// ==============================================================
	// Get contracts
	// ==============================================================

	contracts := []Contract{}
	if utils.IsStringEmpty(state) {
		var err error
		contracts, err = GetContractsInternal(stub, caller, []string{"requester_service_id"}, []string{requesterID})
		if err != nil {
			customErr := &GetDatasError{FieldNames: []string{"requester_service_id", "state"}, Values: []string{requesterID, state}}
			logger.Errorf("%v: %v", customErr, err)
			return nil, errors.Wrap(err, customErr.Error())
		}
	} else {
		var err error
		contracts, err = GetContractsInternal(stub, caller, []string{"requester_service_id", "state"}, []string{requesterID, state})
		if err != nil {
			customErr := &GetDatasError{FieldNames: []string{"requester_service_id", "state"}, Values: []string{requesterID, state}}
			logger.Errorf("%v: %v", customErr, err)
			return nil, errors.Wrap(err, customErr.Error())
		}
	}

	// return contracts
	return json.Marshal(&contracts)
}

// private function that converts asset to contract
func convertContractFromAsset(asset *data_model.Asset) Contract {
	defer utils.ExitFnLog(utils.EnterFnLog())

	contract := Contract{}
	json.Unmarshal(asset.PrivateData, &contract)
	return contract
}

// convertContractToAsset is a private function that converts contract to asset
// contract asset owners list always contains both the contract creater, and
// the data owner/requester service as the 2nd owner in the list
//
// Parameters:
// contract is Contract object to be converted
// ownersList contains the creator service of the contract, either requester service or owner service
func convertContractToAsset(contract Contract, ownersList []string) (data_model.Asset, error) {
	defer utils.ExitFnLog(utils.EnterFnLog())

	asset := data_model.Asset{}
	asset.AssetId = asset_mgmt.GetAssetId(ContractAssetNamespace, contract.ContractID)
	asset.Datatypes = []string{}
	metaData := make(map[string]string)
	metaData["namespace"] = ContractAssetNamespace
	asset.Metadata = metaData

	publicData := ContractPublicData{}
	publicData.RequesterServiceID = contract.RequesterServiceID
	publicData.OwnerServiceID = contract.OwnerServiceID
	asset.PublicData, _ = json.Marshal(publicData)

	contractPrivateData, err := json.Marshal(&contract)
	if err != nil {
		customErr := &custom_errors.MarshalError{Type: "Contract PrivateData"}
		logger.Errorf("%v: %v", customErr, err)
		return data_model.Asset{}, errors.Wrap(err, customErr.Error())
	}
	asset.PrivateData = contractPrivateData
	asset.OwnerIds = ownersList
	asset.IndexTableName = IndexContract

	// save asset to offchain-datastore, if one is setup
	// dsConnectionID, err := GetActiveConnectionID(stub)
	// if err != nil {
	// 	errMsg := "Failed to GetActiveConnectionID"
	// 	logger.Errorf("%v: %v", errMsg, err)
	// 	return data_model.Asset{}, errors.Wrap(err, errMsg)
	// }
	// if !utils.IsStringEmpty(dsConnectionID) {
	// 	asset.SetDatastoreConnectionID(dsConnectionID)
	// }

	return asset, nil
}

// SetupContractIndex sets up contract index table
func SetupContractIndex(stub cached_stub.CachedStubInterface) error {
	defer utils.ExitFnLog(utils.EnterFnLog())

	// Contract Indices
	contractTable := index.GetTable(stub, IndexContract, "contract_id")
	contractTable.AddIndex([]string{"owner_service_id", "state", "contract_id"}, false)
	contractTable.AddIndex([]string{"requester_service_id", "state", "contract_id"}, false)
	contractTable.AddIndex([]string{"owner_service_id", "requester_service_id", "state", "contract_id"}, false)
	contractTable.AddIndex([]string{"requester_service_id", "owner_service_id", "state", "contract_id"}, false)
	err := contractTable.SaveToLedger()

	if err != nil {
		return err
	}

	return nil
}
