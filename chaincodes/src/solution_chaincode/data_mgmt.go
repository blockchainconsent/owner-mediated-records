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
	"encoding/base64"
	"encoding/json"
	"strconv"
	"time"

	"common/bchcls/asset_mgmt"
	"common/bchcls/cached_stub"
	"common/bchcls/consent_mgmt"
	"common/bchcls/crypto"
	"common/bchcls/custom_errors"
	"common/bchcls/data_model"
	"common/bchcls/datatype"
	"common/bchcls/index"
	"common/bchcls/key_mgmt"
	"common/bchcls/user_access_ctrl"
	"common/bchcls/user_mgmt"
	"common/bchcls/user_mgmt/user_keys"
	"common/bchcls/utils"

	"github.com/pkg/errors"
)

const IndexData = "dataTable"
const OwnerDataNamespace = "OwnerDataAsset"

// De-identified fields:
//   - Owner
//   - Service
type OwnerData struct {
	DataID    string      `json:"data_id"`
	Owner     string      `json:"owner"`
	Service   string      `json:"service"`
	Datatype  string      `json:"datatype"`
	Timestamp int64       `json:"timestamp"`
	Data      interface{} `json:"data"`
}

type OwnerDataResult struct {
	Owner     string      `json:"owner"`
	Service   string      `json:"service"`
	Datatype  string      `json:"datatype"`
	Timestamp int64       `json:"timestamp"`
	Data      interface{} `json:"data"`
}

type OwnerDataDownloadResult struct {
	OwnerDatas        []OwnerDataResult `json:"owner_datas"`
	EncryptedContract string            `json:"encrypted_contract"`
	Datatype          string            `json:"datatype"`
}

type OwnerDataResultWithLog struct {
	OwnerDatas     []OwnerDataResult                   `json:"owner_datas"`
	TransactionLog data_model.ExportableTransactionLog `json:"transaction_log"`
}

// log object for data upload and download
type DataLog struct {
	Owner    string      `json:"owner"`
	Target   string      `json:"target"`
	Datatype string      `json:"datatype"`
	Service  string      `json:"service"`
	Data     interface{} `json:"data"`
}

// UploadUserData uploads patient data
// 1) Validate patient data
// 2) Store patient data as asset, same asset key is used per datatype owner pair
// args = [ patientData, dataKeyB64 ]
func UploadUserData(stub cached_stub.CachedStubInterface, caller data_model.User, args []string) ([]byte, error) {
	defer utils.ExitFnLog(utils.EnterFnLog())
	logger.Debugf("args: %v", args)

	if len(args) != 1 && len(args) != 2 {
		customErr := &custom_errors.LengthCheckingError{Type: "UploadUserData arguments length"}
		logger.Errorf(customErr.Error())
		return nil, errors.New(customErr.Error())
	}

	// ==============================================================
	// Validation
	// ==============================================================
	var patientData = OwnerData{}
	patientDataBytes := []byte(args[0])
	err := json.Unmarshal(patientDataBytes, &patientData)
	if err != nil {
		customErr := &custom_errors.UnmarshalError{Type: "patientData"}
		logger.Errorf("%v: %v", customErr, err)
		return nil, errors.Wrap(err, customErr.Error())
	}

	// Validate owner
	if utils.IsStringEmpty(patientData.Owner) {
		customErr := &custom_errors.LengthCheckingError{Type: "patientData.Owner"}
		logger.Errorf(customErr.Error())
		return nil, errors.WithStack(customErr)
	}

	// Validate datatype
	datatype, err := GetDatatypeWithParams(stub, caller, patientData.Datatype)
	if err != nil {
		logger.Errorf("Failed to GetDatatypeWithParams: %v, %v", patientData.Datatype, err)
		return nil, errors.Wrap(err, "Failed to GetDatatypeWithParams: "+patientData.Datatype)
	}

	if utils.IsStringEmpty(datatype.DatatypeID) {
		customErr := &custom_errors.LengthCheckingError{Type: "datatype.DatatypeID"}
		logger.Errorf(customErr.Error())
		return nil, errors.WithStack(customErr)
	}

	// Check timestamp is within 10 mins of current time
	currTime := time.Now().Unix()
	if currTime-patientData.Timestamp > 10*60 || currTime-patientData.Timestamp < -10*60 {
		logger.Errorf("Invalid Timestamp (current time: %v)  %v", currTime, patientData.Timestamp)
		return nil, errors.New("Invalid Timestamp, not within possible time range")
	}

	// set data ID
	patientData.DataID = GetPatientDataID(patientData.Owner, patientData.Datatype, patientData.Timestamp)

	// ==============================================================
	// Get service and act as service
	// because org admin cannot update data asset
	// ==============================================================
	callerObj := caller

	// If caller is org admin of consent target, get consent target user and act as consent target user
	solutionCaller := convertToSolutionUser(caller)
	consentTargetCaller := data_model.User{}
	if !utils.InList(solutionCaller.SolutionInfo.Services, patientData.Service) {
		// If caller is org admin, get sym key path and prv key path first
		symKeyPath, prvKeyPath, err := GetUserAssetSymAndPrivateKeyPaths(stub, caller, patientData.Service)
		if err != nil {
			logger.Errorf("Failed to get symKeyPath and prvKeyPath for user asset")
			return nil, errors.Wrap(err, "Failed to get symKeyPath and prvKeyPath for user asset")
		}

		consentTargetCaller, err = user_mgmt.GetUserData(stub, caller, patientData.Service, true, false, symKeyPath, prvKeyPath)
		if err != nil {
			customErr := &GetUserError{User: patientData.Service}
			logger.Errorf("%v: %v", customErr, err)
			return nil, errors.Wrap(err, customErr.Error())
		}
	} else {
		// If caller is service admin, use default path of GetUserData function to get consent target user
		consentTargetCaller, err = user_mgmt.GetUserData(stub, caller, patientData.Service, true, false)
		if err != nil {
			customErr := &GetUserError{User: patientData.Service}
			logger.Errorf("%v: %v", customErr, err)
			return nil, errors.Wrap(err, customErr.Error())
		}
	}

	if consentTargetCaller.PrivateKey == nil {
		logger.Errorf("Caller does not have access to consent target private key")
		return nil, errors.New("Caller does not have access to consent target private key")
	}

	callerObj = consentTargetCaller

	// ==============================================================
	// Check patient enrollment and consent status
	// ==============================================================
	enrollment, err := GetEnrollmentInternal(stub, callerObj, patientData.Owner, patientData.Service)
	if err != nil {
		customErr := &GetEnrollmentError{Enrollment: GetEnrollmentID(patientData.Owner, patientData.Service)}
		logger.Errorf(customErr.Error())
		return nil, errors.New(customErr.Error())
	}

	// If not currently enrolled, return error
	if enrollment.Status != "active" {
		logger.Errorf("Data owner not currently enrolled")
		return nil, errors.New("Data owner not currently enrolled")
	}

	// Check consent option
	consent, err := GetConsentInternal(stub, caller, patientData.Service, patientData.Datatype, patientData.Owner)
	if err != nil {
		customErr := &GetConsentError{Consent: "Consent for " + patientData.Service + ", " + patientData.Datatype + ", " + patientData.Owner}
		logger.Errorf("%v: %v", customErr, err)
		return nil, errors.Wrap(err, customErr.Error())
	}

	if !utils.InList(consent.Option, consentOptionWrite) {
		logger.Errorf("Do not have permission to upload data, no write permission")
		return nil, errors.New("Do not have permission to upload data, no write permission")
	}

	// ==============================================================
	// Get datatype sym key
	// ==============================================================
	datatypeSymKeyPath, err := GetDatatypeKeyPath(stub, callerObj, patientData.Datatype, patientData.Owner)
	if err != nil {
		customErr := &GetDatatypeKeyPathError{Caller: callerObj.ID, DatatypeID: patientData.Datatype}
		logger.Errorf(customErr.Error())
		return nil, errors.New(customErr.Error())
	}

	datatypeSymKey, err := GetDatatypeSymKey(stub, callerObj, patientData.Datatype, consent.Owner, datatypeSymKeyPath)
	if err != nil {
		logger.Errorf("Failed to GetDatatypeSymKey: %v", err)
		return nil, errors.Wrap(err, "Failed to GetDatatypeSymKey")
	}

	if datatypeSymKey.KeyBytes == nil {
		logger.Errorf("Failed to get datatypeSymKey")
		return nil, errors.New("Failed to get datatypeSymKey")
	}

	// ==============================================================
	// Save patientData as asset
	// ==============================================================
	assetManager := asset_mgmt.GetAssetManager(stub, callerObj)
	patientDataAsset, err := convertOwnerDataToAsset(stub, patientData)
	if err != nil {
		customErr := &ConvertToAssetError{Asset: "ownerDataAsset"}
		logger.Errorf("%v: %v", customErr, err)
		return nil, errors.Wrap(err, customErr.Error())
	}

	// we use timestamp of -1 to represent "latest"
	// first try to get asset for asset ID ownerId + datatypeId + latest
	latestPatientDataID := GetPatientDataID(patientData.Owner, patientData.Datatype, -1)
	latestPatientDataAssetID := asset_mgmt.GetAssetId(OwnerDataNamespace, latestPatientDataID)

	keyPath, err := GetKeyPath(stub, callerObj, latestPatientDataAssetID)
	if err != nil {
		customErr := &GetKeyPathError{Caller: callerObj.ID, AssetID: latestPatientDataAssetID}
		logger.Errorf(customErr.Error())
		return nil, errors.New(customErr.Error())
	}

	if len(keyPath) == 0 {
		// key path length is 0, assume asset does not exist
		// if asset does not exist
		// use key passed in from JS, add new asset with asset ID ownerId + datatypeId + timestamp
		//  add new asset with asset ID ownerId + datatypeId + latest
		if len(args) != 2 {
			customErr := &custom_errors.LengthCheckingError{Type: "UploadUserData arguments length"}
			logger.Errorf(customErr.Error())
			return nil, errors.New(customErr.Error())
		}

		// Validate data key
		dataKey := data_model.Key{ID: key_mgmt.GetSymKeyId(patientData.Owner + patientData.Datatype), Type: key_mgmt.KEY_TYPE_SYM}
		dataKey.KeyBytes, err = crypto.ParseSymKeyB64(args[1])
		if err != nil {
			logger.Errorf("Invalid dataKey")
			return nil, errors.Wrap(err, "Invalid dataKey")
		}

		if dataKey.KeyBytes == nil {
			logger.Errorf("Invalid dataKey")
			return nil, errors.New("Invalid dataKey")
		}

		// add access from ownerPubKey to dataKey
		ownerPubKey, err := user_keys.GetUserPublicKey(stub, caller, patientData.Owner)
		if err != nil {
			errMsg := "Failed to get public key of " + patientData.Owner
			logger.Errorf("%v: %v", errMsg, err)
			return nil, errors.Wrap(err, errMsg)
		}

		userAccessManager := user_access_ctrl.GetUserAccessManager(stub, caller)
		err = userAccessManager.AddAccessByKey(ownerPubKey, dataKey)
		if err != nil {
			customErr := &custom_errors.AddAccessError{Key: "owner pub key to data key"}
			logger.Errorf("%v: %v", customErr, err)
			return nil, errors.Wrap(err, customErr.Error())
		}

		// add new asset
		err = assetManager.AddAsset(patientDataAsset, dataKey, false)
		if err != nil {
			customErr := &PutAssetError{Asset: patientDataAsset.AssetId}
			logger.Errorf("%v: %v", customErr, err)
			return nil, errors.Wrap(err, customErr.Error())
		}

		// add latest asset
		latestPatientData := patientData
		latestPatientData.Timestamp = -1
		latestPatientData.DataID = GetPatientDataID(patientData.Owner, patientData.Datatype, -1)
		latestPatientDataAsset, err := convertOwnerDataToAsset(stub, latestPatientData)
		if err != nil {
			customErr := &ConvertToAssetError{Asset: "ownerDataAsset"}
			logger.Errorf("%v: %v", customErr, err)
			return nil, errors.Wrap(err, customErr.Error())
		}

		err = assetManager.AddAsset(latestPatientDataAsset, dataKey, false)
		if err != nil {
			customErr := &PutAssetError{Asset: latestPatientDataID}
			logger.Errorf("%v: %v", customErr, err)
			return nil, errors.Wrap(err, customErr.Error())
		}
	} else {
		dataKey, err := assetManager.GetAssetKey(latestPatientDataAssetID, keyPath)
		if err != nil {
			logger.Errorf("Failed to get data AssetKey for existing data: %v", err)
			return nil, errors.Wrap(err, "Failed to data AssetKey for existing data")
		}
		// asset exists
		// first add new asset with ID of ownerId + datatypeId + timestamp under existing key
		err = assetManager.AddAsset(patientDataAsset, dataKey, false)
		if err != nil {
			customErr := &PutAssetError{Asset: patientDataAsset.AssetId}
			logger.Errorf("%v: %v", customErr, err)
			return nil, errors.Wrap(err, customErr.Error())
		}

		// then update latest asset
		latestPatientData := patientData
		latestPatientData.Timestamp = -1
		latestPatientData.DataID = GetPatientDataID(patientData.Owner, patientData.Datatype, -1)
		latestPatientDataAsset, err := convertOwnerDataToAsset(stub, latestPatientData)
		if err != nil {
			customErr := &ConvertToAssetError{Asset: "ownerDataAsset"}
			logger.Errorf("%v: %v", customErr, err)
			return nil, errors.Wrap(err, customErr.Error())
		}

		err = assetManager.UpdateAsset(latestPatientDataAsset, dataKey, true)
		if err != nil {
			customErr := &PutAssetError{Asset: latestPatientDataAssetID}
			logger.Errorf("%v: %v", customErr, err)
			return nil, errors.Wrap(err, customErr.Error())
		}
	}

	// ==============================================================
	// Logging
	// ==============================================================
	enrollmentAssetID := asset_mgmt.GetAssetId(EnrollmentAssetNamespace, enrollment.EnrollmentID)
	keyPath, err = GetKeyPath(stub, callerObj, enrollmentAssetID)
	if err != nil {
		customErr := &GetKeyPathError{Caller: caller.ID, AssetID: enrollmentAssetID}
		logger.Errorf(customErr.Error())
		return nil, errors.New(customErr.Error())
	}

	if len(keyPath) <= 0 {
		customErr := &GetKeyPathError{Caller: caller.ID, AssetID: enrollmentAssetID}
		logger.Errorf(customErr.Error())
		return nil, errors.New(customErr.Error())
	}

	enrollmentKey, err := assetManager.GetAssetKey(enrollmentAssetID, keyPath)
	if err != nil {
		logger.Errorf("Failed to GetAssetKey for enrollmentKey: %v", err)
		return nil, errors.Wrap(err, "Failed to GetAssetKey for enrollmentKey")
	}

	// access from ownerLogSymKey and targetLogSymKey to enrollmentLogSymKey already added in enroll mgmt
	enrollmentLogSymKey := GetLogSymKeyFromKey(enrollmentKey)

	dataLog := DataLog{Owner: patientData.Owner, Datatype: patientData.Datatype, Service: patientData.Service}
	solutionLog := SolutionLog{
		TransactionID: stub.GetTxID(),
		Namespace:     "OMR",
		FunctionName:  "UploadUserData",
		CallerID:      caller.ID,
		Timestamp:     patientData.Timestamp,
		Data:          dataLog}

	err = AddLogWithParams(stub, callerObj, solutionLog, enrollmentLogSymKey)
	if err != nil {
		customErr := &AddSolutionLogError{FunctionName: solutionLog.FunctionName}
		logger.Errorf("%v: %v", customErr, err)
		return nil, errors.Wrap(err, customErr.Error())
	}

	return nil, nil
}

// UploadOwnerData uploads owner data
// 1) Validate owner data
// 2) Store owner data as asset
// args = [ ownerData, dataKeyB64 ]
func UploadOwnerData(stub cached_stub.CachedStubInterface, caller data_model.User, args []string) ([]byte, error) {
	defer utils.ExitFnLog(utils.EnterFnLog())
	logger.Debugf("args: %v", args)

	if len(args) != 1 && len(args) != 2 {
		customErr := &custom_errors.LengthCheckingError{Type: "UploadOwnerData arguments length"}
		logger.Errorf(customErr.Error())
		return nil, errors.New(customErr.Error())
	}

	// ==============================================================
	// Validation
	// ==============================================================
	var ownerData = OwnerData{}
	ownerDataBytes := []byte(args[0])
	err := json.Unmarshal(ownerDataBytes, &ownerData)
	if err != nil {
		customErr := &custom_errors.UnmarshalError{Type: "ownerData"}
		logger.Errorf("%v: %v", customErr, err)
		return nil, errors.Wrap(err, customErr.Error())
	}

	// Validate owner
	if utils.IsStringEmpty(ownerData.Owner) {
		customErr := &custom_errors.LengthCheckingError{Type: "ownerData.Owner"}
		logger.Errorf(customErr.Error())
		return nil, errors.WithStack(customErr)
	}

	if ownerData.Service != ownerData.Owner {
		logger.Errorf("Service must be same as owner: %v, %v", ownerData.Service, ownerData.Owner)
		return nil, errors.New("Service must be same as owner")
	}

	// Check timestamp is within 10 mins of current time
	currTime := time.Now().Unix()
	if currTime-ownerData.Timestamp > 10*60 || currTime-ownerData.Timestamp < -10*60 {
		logger.Errorf("Invalid Timestamp (current time: %v)  %v", currTime, ownerData.Timestamp)
		return nil, errors.New("Invalid Timestamp, not within possible time range")
	}

	// set data ID
	// owner is same as service
	ownerData.DataID = GetOwnerDataID(ownerData.Owner, ownerData.Datatype, ownerData.Timestamp)

	// ==============================================================
	// Get owner and act as owner
	// because org admin cannot update data asset
	// ==============================================================
	callerObj := caller

	// If caller is org admin of owner, get owner user and act as owner
	if caller.ID != ownerData.Owner {
		solutionCaller := convertToSolutionUser(caller)
		// Owner is org, caller is org admin || owner is service, caller is service admin
		if solutionCaller.Org == ownerData.Owner || utils.InList(solutionCaller.SolutionInfo.Services, ownerData.Owner) {
			callerObj, err = user_mgmt.GetUserData(stub, caller, ownerData.Owner, true, false)
			if err != nil {
				customErr := &GetUserError{User: ownerData.Owner}
				logger.Errorf("%v: %v", customErr, err)
				return nil, errors.Wrap(err, customErr.Error())
			}
		} else {
			// owner is service, caller is org admin
			// If caller is org admin, get sym key path and prv key path first
			symKeyPath, prvKeyPath, err := GetUserAssetSymAndPrivateKeyPaths(stub, caller, ownerData.Owner)
			if err != nil {
				logger.Errorf("Failed to get symKeyPath and prvKeyPath for user asset")
				return nil, errors.Wrap(err, "Failed to get symKeyPath and prvKeyPath for user asset")
			}

			callerObj, err = user_mgmt.GetUserData(stub, caller, ownerData.Owner, true, false, symKeyPath, prvKeyPath)
			if err != nil {
				customErr := &GetUserError{User: ownerData.Owner}
				logger.Errorf("%v: %v", customErr, err)
				return nil, errors.Wrap(err, customErr.Error())
			}
		}

		if callerObj.PrivateKey == nil {
			logger.Errorf("Caller does not have access to owner private key")
			return nil, errors.New("Caller does not have access to owner private key")
		}
	}

	// ==============================================================
	// Add datatype sym key
	// ==============================================================
	// Create datatype owner sym key
	_, err = datatype.AddDatatypeSymKey(stub, callerObj, ownerData.Datatype, ownerData.Owner)
	if err != nil {
		errMsg := "Failed to add datatype sym key in SDK "
		logger.Errorf("%v: %v", errMsg, err)
		return nil, errors.Wrap(err, errMsg)
	}

	datatypeSymKeyPath, err := GetDatatypeKeyPath(stub, callerObj, ownerData.Datatype, ownerData.Owner)
	if err != nil {
		customErr := &GetDatatypeKeyPathError{Caller: callerObj.ID, DatatypeID: ownerData.Datatype}
		logger.Errorf(customErr.Error())
		return nil, errors.New(customErr.Error())
	}

	// gets from cache, call GetDatatypeSymKey so PutAsset doesn't have to get it again
	datatypeSymKey, err := GetDatatypeSymKey(stub, callerObj, ownerData.Datatype, ownerData.Owner, datatypeSymKeyPath)
	if err != nil {
		logger.Errorf("Failed to GetDatatypeSymKey: %v", err)
		return nil, errors.Wrap(err, "Failed to GetDatatypeSymKey")
	}

	if datatypeSymKey.KeyBytes == nil {
		logger.Errorf("Failed to get datatypeSymKey")
		return nil, errors.New("Failed to get datatypeSymKey")
	}

	// ==============================================================
	// Save ownerData as asset
	// ==============================================================
	userAccessManager := user_access_ctrl.GetUserAccessManager(stub, caller)

	// Convert ownerData object to Asset
	assetManager := asset_mgmt.GetAssetManager(stub, callerObj)
	ownerDataAsset, err := convertOwnerDataToAsset(stub, ownerData)
	if err != nil {
		customErr := &ConvertToAssetError{Asset: "ownerDataAsset"}
		logger.Errorf("%v: %v", customErr, err)
		return nil, errors.Wrap(err, customErr.Error())
	}

	// we use timestamp of -1 to represent "latest"
	// first get asset for asset ID ownerId + datatypeId + latest
	latestOwnerDataID := GetOwnerDataID(ownerData.Owner, ownerData.Datatype, -1)
	latestOwnerDataAssetID := asset_mgmt.GetAssetId(OwnerDataNamespace, latestOwnerDataID)
	keyPath, err := GetKeyPath(stub, callerObj, latestOwnerDataAssetID)
	if err != nil {
		customErr := &GetKeyPathError{Caller: caller.ID, AssetID: latestOwnerDataAssetID}
		logger.Errorf(customErr.Error())
		return nil, errors.New(customErr.Error())
	}

	dataKey := data_model.Key{}
	if len(keyPath) == 0 {
		// key path length is 0, assume asset does not exist
		// use key passed in from JS, add new asset with asset ID ownerId + datatypeId + timestamp
		//  add new asset with asset ID ownerId + datatypeId + latest
		if len(args) != 2 {
			customErr := &custom_errors.LengthCheckingError{Type: "UploadOwnerData arguments length"}
			logger.Errorf(customErr.Error())
			return nil, errors.New(customErr.Error())
		}

		// Validate data key
		dataKey = data_model.Key{ID: key_mgmt.GetSymKeyId(ownerData.Owner + ownerData.Datatype), Type: key_mgmt.KEY_TYPE_SYM}
		dataKey.KeyBytes, err = crypto.ParseSymKeyB64(args[1])
		if err != nil {
			logger.Errorf("Invalid dataKey")
			return nil, errors.Wrap(err, "Invalid dataKey")
		}

		if dataKey.KeyBytes == nil {
			logger.Errorf("Invalid dataKey")
			return nil, errors.New("Invalid dataKey")
		}

		// add access from ownerPubKey to dataKey
		ownerPubKey, err := user_keys.GetUserPublicKey(stub, caller, ownerData.Owner)
		if err != nil {
			errMsg := "Failed to get public key of " + ownerData.Owner
			logger.Errorf("%v: %v", errMsg, err)
			return nil, errors.Wrap(err, errMsg)
		}

		err = userAccessManager.AddAccessByKey(ownerPubKey, dataKey)
		if err != nil {
			customErr := &custom_errors.AddAccessError{Key: "owner pub key to data key"}
			logger.Errorf("%v: %v", customErr, err)
			return nil, errors.Wrap(err, customErr.Error())
		}

		// add new asset
		err = assetManager.AddAsset(ownerDataAsset, dataKey, false)
		if err != nil {
			customErr := &PutAssetError{Asset: ownerDataAsset.AssetId}
			logger.Errorf("%v: %v", customErr, err)
			return nil, errors.Wrap(err, customErr.Error())
		}

		// add latest asset
		latestOwnerData := ownerData
		latestOwnerData.Timestamp = -1
		latestOwnerData.DataID = GetOwnerDataID(ownerData.Owner, ownerData.Datatype, -1)
		latestOwnerDataAsset, err := convertOwnerDataToAsset(stub, latestOwnerData)
		if err != nil {
			customErr := &ConvertToAssetError{Asset: "ownerDataAsset"}
			logger.Errorf("%v: %v", customErr, err)
			return nil, errors.Wrap(err, customErr.Error())
		}

		err = assetManager.AddAsset(latestOwnerDataAsset, dataKey, false)
		if err != nil {
			customErr := &PutAssetError{Asset: ownerDataAsset.AssetId}
			logger.Errorf("%v: %v", customErr, err)
			return nil, errors.Wrap(err, customErr.Error())
		}
	} else {
		dataKey, err = assetManager.GetAssetKey(latestOwnerDataAssetID, keyPath)
		if err != nil {
			logger.Errorf("Failed to get data AssetKey for existing data: %v", err)
			return nil, errors.Wrap(err, "Failed to data AssetKey for existing data")
		}

		// asset exists
		//  add new asset with ID of ownerId + datatypeId + timestamp under existing key
		err = assetManager.AddAsset(ownerDataAsset, dataKey, false)
		if err != nil {
			customErr := &PutAssetError{Asset: latestOwnerDataAssetID}
			logger.Errorf("%v: %v", customErr, err)
			return nil, errors.Wrap(err, customErr.Error())
		}

		// update latest asset
		latestOwnerData := ownerData
		latestOwnerData.Timestamp = -1
		latestOwnerData.DataID = GetOwnerDataID(ownerData.Owner, ownerData.Datatype, -1)
		latestOwnerDataAsset, err := convertOwnerDataToAsset(stub, latestOwnerData)
		if err != nil {
			customErr := &ConvertToAssetError{Asset: "ownerDataAsset"}
			logger.Errorf("%v: %v", customErr, err)
			return nil, errors.Wrap(err, customErr.Error())
		}

		err = assetManager.UpdateAsset(latestOwnerDataAsset, dataKey, true)
		if err != nil {
			customErr := &PutAssetError{Asset: latestOwnerDataAssetID}
			logger.Errorf("%v: %v", customErr, err)
			return nil, errors.Wrap(err, customErr.Error())
		}
	}

	// ==============================================================
	// Logging
	// ==============================================================
	logSymKey := GetLogSymKeyFromKey(dataKey)
	dataLog := DataLog{Owner: ownerData.Owner, Datatype: ownerData.Datatype, Service: ownerData.Service}
	solutionLog := SolutionLog{
		TransactionID: stub.GetTxID(),
		Namespace:     "OMR",
		FunctionName:  "UploadOwnerData",
		CallerID:      caller.ID,
		Timestamp:     ownerData.Timestamp,
		Data:          dataLog}

	err = AddLogWithParams(stub, callerObj, solutionLog, logSymKey)
	if err != nil {
		customErr := &AddSolutionLogError{FunctionName: solutionLog.FunctionName}
		logger.Errorf("%v: %v", customErr, err)
		return nil, errors.Wrap(err, customErr.Error())
	}

	// ==============================================================
	// Add access from ownerLogSymKey to dataLogSymKey
	// ==============================================================
	ownerLogSymKey := callerObj.GetLogSymKey()
	err = userAccessManager.AddAccessByKey(ownerLogSymKey, logSymKey)
	if err != nil {
		customErr := &custom_errors.AddAccessError{Key: "ownerLogSymKey to logSymKey"}
		logger.Errorf("%v: %v", customErr, err)
		return nil, errors.Wrap(err, customErr.Error())
	}

	return nil, nil
}

// DownloadOwnerDataAsOwner downloads owner data
// Should only be used by owner or callers with access to owner
// args = [service, datatype, latestOnly, startTimestamp, endTimestamp, maxNum, timestamp]
// If latest only is true, then other filters are ignored
func DownloadOwnerDataAsOwner(stub cached_stub.CachedStubInterface, caller data_model.User, args []string) ([]byte, error) {
	defer utils.ExitFnLog(utils.EnterFnLog())
	logger.Debugf("args: %v", args)

	if len(args) != 7 {
		customErr := &custom_errors.LengthCheckingError{Type: "DownloadOwnerDataAsOwner arguments length"}
		logger.Errorf(customErr.Error())
		return nil, errors.WithStack(customErr)
	}

	// ==============================================================
	// Validation
	// ==============================================================
	owner := args[0]
	if utils.IsStringEmpty(owner) {
		customErr := &custom_errors.LengthCheckingError{Type: "owner"}
		logger.Errorf(customErr.Error())
		return nil, errors.WithStack(customErr)
	}

	datatype := args[1]
	if utils.IsStringEmpty(datatype) {
		customErr := &custom_errors.LengthCheckingError{Type: "datatype"}
		logger.Errorf(customErr.Error())
		return nil, errors.WithStack(customErr)
	}

	latestOnlyFlag := args[2]
	if utils.IsStringEmpty(latestOnlyFlag) {
		customErr := &custom_errors.LengthCheckingError{Type: "latestOnlyFlag"}
		logger.Errorf(customErr.Error())
		return nil, errors.WithStack(customErr)
	}

	if latestOnlyFlag != "true" && latestOnlyFlag != "false" {
		logger.Errorf("Error: Latest only flag must be true or false")
		return nil, errors.New("Error: Latest only flag must be true or false")
	}

	startTimestamp, err := strconv.ParseInt(args[3], 10, 64)
	if err != nil {
		logger.Errorf("Error converting startTimestamp to type int64")
		return nil, errors.Wrap(err, "Error converting startTimestamp to type int64")
	}

	endTimestamp, err := strconv.ParseInt(args[4], 10, 64)
	if err != nil {
		logger.Errorf("Error converting endTimestamp to type int64")
		return nil, errors.Wrap(err, "Error converting endTimestamp to type int64")
	}

	maxNum, err := strconv.ParseInt(args[5], 10, 64)
	if err != nil {
		logger.Errorf("Error converting maxNum to type int")
		return nil, errors.Wrap(err, "Error converting maxNum to type int")
	}

	if maxNum < 0 {
		logger.Errorf("Max num must be greater than 0")
		return nil, errors.New("Max num must be greater than 0")
	}

	timestamp, err := strconv.ParseInt(args[6], 10, 64)
	if err != nil {
		logger.Errorf("Error converting timestamp to type int64")
		return nil, errors.Wrap(err, "Error converting maxNum to type int64")
	}

	// Check timestamp is within 10 mins of current time
	currTime := time.Now().Unix()
	if currTime-timestamp > 10*60 || currTime-timestamp < -10*60 {
		logger.Errorf("Invalid Timestamp (current time: %v)  %v", currTime, timestamp)
		return nil, errors.New("Invalid Timestamp, not within possible time range")
	}

	// ==============================================================
	// Get owner and act as owner
	// ==============================================================
	callerObj := caller

	// owner is not patient, owner must be service
	if caller.ID != owner {
		solutionCaller := convertToSolutionUser(caller)
		// Owner is org, caller is org admin || owner is service, caller is service admin
		if solutionCaller.Org == owner || utils.InList(solutionCaller.SolutionInfo.Services, owner) {
			callerObj, err = user_mgmt.GetUserData(stub, caller, owner, true, false)
			if err != nil {
				customErr := &GetUserError{User: owner}
				logger.Errorf("%v: %v", customErr, err)
				return nil, errors.Wrap(err, customErr.Error())
			}
		} else {
			// owner is service, caller is org admin
			// If caller is org admin, get sym key path and prv key path first
			symKeyPath, prvKeyPath, err := GetUserAssetSymAndPrivateKeyPaths(stub, caller, owner)
			if err != nil {
				logger.Errorf("Failed to get symKeyPath and prvKeyPath for user asset")
				return nil, errors.Wrap(err, "Failed to get symKeyPath and prvKeyPath for user asset")
			}

			callerObj, err = user_mgmt.GetUserData(stub, caller, owner, true, false, symKeyPath, prvKeyPath)
			if err != nil {
				customErr := &GetUserError{User: owner}
				logger.Errorf("%v: %v", customErr, err)
				return nil, errors.Wrap(err, customErr.Error())
			}
		}

		if callerObj.PrivateKey == nil {
			logger.Errorf("Caller does not have access to owner private key")
			return nil, errors.New("Caller does not have access to owner private key")
		}
	}

	// ==============================================================
	// Download data using index
	// ==============================================================
	ownerDatas := []OwnerDataResult{}
	assetID := GetLatestOwnerDataAssetID(stub, owner, datatype)
	if latestOnlyFlag == "true" {
		data, err := GetDataWithAssetID(stub, callerObj, assetID, owner, datatype)
		if err != nil {
			logger.Errorf("Failed to get data with assetID")
			return nil, errors.WithStack(err)
		}

		ownerDatas = append(ownerDatas, data)

	} else {
		startValues := []string{owner, datatype}
		if startTimestamp >= 0 {
			startTimestampStr, err := utils.ConvertToString(startTimestamp)
			if err != nil {
				errMsg := "Failed to ConvertToString for startTimestamp"
				logger.Errorf("%v: %v", errMsg, err)
				return nil, errors.Wrap(err, errMsg)
			}
			startValues = append(startValues, startTimestampStr)
		}

		endValues := []string{owner, datatype}
		if endTimestamp > 0 {
			endTimestampStr, err := utils.ConvertToString(endTimestamp)
			if err != nil {
				errMsg := "Failed to ConvertToString for endTimestamp"
				logger.Errorf("%v: %v", errMsg, err)
				return nil, errors.Wrap(err, errMsg)
			}
			endValues = append(endValues, endTimestampStr)
		}

		ownerDatas, err = GetDataInternal(stub, callerObj, []string{"owner", "datatype", "timestamp"}, startValues, endValues, int(maxNum))
		if err != nil {
			customErr := &GetDatasError{FieldNames: []string{"owner", "datatype", "timestamp"}, Values: startValues}
			logger.Errorf("%v: %v", customErr, err)
			return nil, errors.Wrap(err, customErr.Error())
		}
	}

	// ==============================================================
	// Logging
	// ==============================================================

	keyPath, err := GetKeyPath(stub, callerObj, assetID)
	if err != nil {
		customErr := &GetKeyPathError{Caller: caller.ID, AssetID: assetID}
		logger.Errorf(customErr.Error())
		return nil, errors.New(customErr.Error())
	}

	if len(keyPath) <= 0 {
		customErr := &GetKeyPathError{Caller: caller.ID, AssetID: assetID}
		logger.Errorf(customErr.Error())
		return nil, errors.New(customErr.Error())
	}

	assetManager := asset_mgmt.GetAssetManager(stub, callerObj)
	dataKey, err := assetManager.GetAssetKey(assetID, keyPath)
	if err != nil {
		logger.Errorf("Failed to get data key with latest owner data asset ID: %v", err)
		return nil, errors.Wrap(err, "Failed to get data key with latest owner data asset ID")
	}

	dataLogSymKey := GetLogSymKeyFromKey(dataKey)

	dataLog := DataLog{Owner: owner, Datatype: datatype}
	solutionLog := SolutionLog{
		TransactionID: stub.GetTxID(),
		Namespace:     "OMR",
		FunctionName:  "DownloadOwnerDataAsOwner",
		CallerID:      caller.ID,
		Timestamp:     timestamp,
		Data:          dataLog}
	exportableLog, err := GenerateExportableSolutionLog(stub, caller, solutionLog, dataLogSymKey)
	if err != nil {
		customErr := &GenerateExportableTransactionLogError{Function: solutionLog.FunctionName}
		logger.Errorf("%v: %v", customErr, err)
		return nil, errors.Wrap(err, customErr.Error())
	}

	returnData := OwnerDataResultWithLog{}
	returnData.OwnerDatas = ownerDatas
	returnData.TransactionLog = exportableLog
	logger.Infof("got owner data: %v", len(ownerDatas))

	return json.Marshal(&returnData)
}

// DownloadOwnerDataAsRequester downloads owner data
// args = [contractID, datatype, latestOnly, startTimestamp, endTimestamp, maxNum, timestamp]
// If latest only is true, then other filters are ignored
func DownloadOwnerDataAsRequester(stub cached_stub.CachedStubInterface, caller data_model.User, args []string) ([]byte, error) {
	defer utils.ExitFnLog(utils.EnterFnLog())
	logger.Debugf("args: %v", args)

	if len(args) != 7 {
		customErr := &custom_errors.LengthCheckingError{Type: "DownloadOwnerDataAsRequester arguments length"}
		logger.Errorf(customErr.Error())
		return nil, errors.WithStack(customErr)
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

	// Return error if a contract with this ID does not exist
	if utils.IsStringEmpty(contract.ContractID) {
		customErr := &GetContractError{ContractID: contractID}
		logger.Errorf(customErr.Error())
		return nil, errors.New(customErr.Error())
	}

	// Return error if a contract with this ID does not exist
	if contract.State != "downloadReady" {
		logger.Errorf("Download is not ready")
		return nil, errors.New("Download is not ready")
	}

	// ==============================================================
	// Change caller
	// ==============================================================
	callerObj := caller
	if solutionCaller.Org != contract.RequesterOrgID {
		logger.Errorf("Caller org is not the same as requester org")
		return nil, errors.New("Caller org is not the same as requester org")
	}

	// If caller is org admin, then have to use key paths
	if solutionCaller.SolutionInfo.IsOrgAdmin {
		symKeyPath, prvKeyPath, err := GetUserAssetSymAndPrivateKeyPaths(stub, caller, contract.RequesterServiceID)
		if err != nil {
			logger.Errorf("Failed to get symKeyPath and prvKeyPath for user asset")
			return nil, errors.Wrap(err, "Failed to get symKeyPath and prvKeyPath for user asset")
		}
		// act as contract owner service
		callerObj, err = user_mgmt.GetUserData(stub, caller, contract.RequesterServiceID, true, false, symKeyPath, prvKeyPath)
		if err != nil {
			customErr := &GetUserError{User: contract.OwnerServiceID}
			logger.Errorf("%v: %v", customErr, err)
			return nil, errors.Wrap(err, customErr.Error())
		}
	} else {
		callerObj, err = user_mgmt.GetUserData(stub, caller, contract.RequesterServiceID, true, false)
		if err != nil {
			customErr := &GetUserError{User: contract.OwnerServiceID}
			logger.Errorf("%v: %v", customErr, err)
			return nil, errors.Wrap(err, customErr.Error())
		}
	}

	datatype := args[1]
	if utils.IsStringEmpty(datatype) {
		customErr := &custom_errors.LengthCheckingError{Type: "datatype"}
		logger.Errorf(customErr.Error())
		return nil, errors.WithStack(customErr)
	}

	latestOnlyFlag := args[2]
	if utils.IsStringEmpty(latestOnlyFlag) {
		customErr := &custom_errors.LengthCheckingError{Type: "latestOnlyFlag"}
		logger.Errorf(customErr.Error())
		return nil, errors.WithStack(customErr)
	}

	if latestOnlyFlag != "true" && latestOnlyFlag != "false" {
		logger.Errorf("Error: Latest only flag must be true or false")
		return nil, errors.New("Error: Latest only flag must be true or false")
	}

	startTimestamp, err := strconv.ParseInt(args[3], 10, 64)
	if err != nil {
		logger.Errorf("Error converting startTimestamp to type int64")
		return nil, errors.Wrap(err, "Error converting startTimestamp to type int64")
	}

	endTimestamp, err := strconv.ParseInt(args[4], 10, 64)
	if err != nil {
		logger.Errorf("Error converting endTimestamp to type int64")
		return nil, errors.Wrap(err, "Error converting endTimestamp to type int64")
	}

	maxNum, err := strconv.ParseInt(args[5], 10, 64)
	if err != nil {
		logger.Errorf("Error converting maxNum to type int64")
		return nil, errors.Wrap(err, "Error converting maxNum to type int64")
	}

	if maxNum < 0 {
		logger.Errorf("Max num must be greater than 0")
		return nil, errors.New("Max num must be greater than 0")
	}

	timestamp, err := strconv.ParseInt(args[6], 10, 64)
	if err != nil {
		logger.Errorf("Error converting timestamp to type int64")
		return nil, errors.Wrap(err, "Error converting maxNum to type int64")
	}

	// Check timestamp is within 10 mins of current time
	currTime := time.Now().Unix()
	if currTime-timestamp > 10*60 || currTime-timestamp < -10*60 {
		logger.Errorf("Invalid Timestamp (current time: %v)  %v", currTime, timestamp)
		return nil, errors.New("Invalid Timestamp, not within possible time range")
	}

	// ==============================================================
	// Update contract and manage relationship
	// ==============================================================
	contractDetail := ContractDetail{}
	contractDetail.ContractID = contractID
	contractDetail.ContractDetailType = "download"
	contractDetail.ContractDetailTerms = "{}"
	contractDetail.CreateDate = timestamp
	contractDetail.CreatedBy = caller.ID

	contract.UpdateDate = timestamp
	contract.ContractDetails = append(contract.ContractDetails, contractDetail)

	// ==============================================================
	// Download data using index
	// ==============================================================
	ownerDatas := []OwnerDataResult{}
	if contract.MaxNumDownload > contract.NumDownload {
		if latestOnlyFlag == "true" {
			assetID := GetLatestOwnerDataAssetID(stub, contract.OwnerServiceID, datatype)
			data, err := GetDataWithAssetID(stub, callerObj, assetID, contract.OwnerServiceID, datatype)
			if err != nil {
				logger.Errorf("Failed to get data with assetID")
				return nil, errors.WithStack(err)
			}

			ownerDatas = append(ownerDatas, data)

		} else {
			startValues := []string{contract.OwnerServiceID, datatype}
			if startTimestamp >= 0 {
				startTimestampStr, err := utils.ConvertToString(startTimestamp)
				if err != nil {
					errMsg := "Failed to ConvertToString for startTimestamp"
					logger.Errorf("%v: %v", errMsg, err)
					return nil, errors.Wrap(err, errMsg)
				}
				startValues = append(startValues, startTimestampStr)
			}

			endValues := []string{contract.OwnerServiceID, datatype}
			if endTimestamp > 0 {
				endTimestampStr, err := utils.ConvertToString(endTimestamp)
				if err != nil {
					errMsg := "Failed to ConvertToString for endTimestamp"
					logger.Errorf("%v: %v", errMsg, err)
					return nil, errors.Wrap(err, errMsg)
				}
				endValues = append(endValues, endTimestampStr)
			}

			ownerDatas, err = GetDataInternal(stub, callerObj, []string{"owner", "datatype", "timestamp"}, startValues, endValues, int(maxNum))

			if err != nil {
				customErr := &GetDatasError{FieldNames: []string{"owner", "datatype", "timestamp"}, Values: startValues}
				logger.Errorf("%v: %v", customErr, err)
				return nil, errors.Wrap(err, customErr.Error())
			}

		}
		contract.NumDownload = contract.NumDownload + 1
	}

	// ==============================================================
	// Pass contract and owner data back to JS
	// ==============================================================
	// Get contract key
	assetManager := asset_mgmt.GetAssetManager(stub, callerObj)
	contractAssetID := asset_mgmt.GetAssetId(ContractAssetNamespace, contractID)
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

	//ecrypt contract
	contractBytes, err := json.Marshal(&contract)
	if err != nil {
		customErr := &custom_errors.MarshalError{Type: "contract"}
		logger.Errorf("%v: %v", customErr, err)
		return nil, errors.Wrap(err, customErr.Error())
	}

	encContractBytes, err := crypto.EncryptWithSymKey(contractKey.KeyBytes, contractBytes)
	if err != nil {
		customErr := &custom_errors.EncryptionError{ToEncrypt: "ContractBytes", EncryptionKey: contractKey.ID}
		logger.Errorf("%v: %v", customErr, err)
		return nil, errors.Wrap(err, customErr.Error())
	}

	if encContractBytes == nil {
		customErr := &custom_errors.EncryptionError{ToEncrypt: "ContractBytes", EncryptionKey: contractKey.ID}
		logger.Errorf("%v", customErr)
		return nil, errors.WithStack(customErr)
	}

	encContractStr := base64.StdEncoding.EncodeToString(encContractBytes)

	returnData := OwnerDataDownloadResult{}
	returnData.OwnerDatas = ownerDatas
	returnData.EncryptedContract = encContractStr
	returnData.Datatype = datatype
	logger.Infof("got owner data: %v", len(ownerDatas))

	return json.Marshal(&returnData)
}

// DownloadOwnerDataWithConsent downloads owner data with consent
// Should only be used by consent target or callers with access to consent target
// Owner should use DownloadOwnerDataAsOwner function
//
// args = [target, owner, datatype, latestOnly, startTimestamp, endTimestamp, maxNum, timestamp]
// If latest only is true, then other filters are ignored
// Only used by consent target, owner should use DownloadOwnerDataAsOwner function
func DownloadOwnerDataWithConsent(stub cached_stub.CachedStubInterface, caller data_model.User, args []string) ([]byte, error) {
	defer utils.ExitFnLog(utils.EnterFnLog())
	logger.Debugf("args: %v", args)

	if len(args) != 8 {
		customErr := &custom_errors.LengthCheckingError{Type: "DownloadOwnerDataWithConsent arguments length"}
		logger.Errorf(customErr.Error())
		return nil, errors.WithStack(customErr)
	}

	// ==============================================================
	// Validation
	// ==============================================================
	target := args[0]
	if utils.IsStringEmpty(target) {
		customErr := &custom_errors.LengthCheckingError{Type: "target"}
		logger.Errorf(customErr.Error())
		return nil, errors.WithStack(customErr)
	}

	owner := args[1]
	if utils.IsStringEmpty(owner) {
		customErr := &custom_errors.LengthCheckingError{Type: "owner"}
		logger.Errorf(customErr.Error())
		return nil, errors.WithStack(customErr)
	}

	datatype := args[2]
	if utils.IsStringEmpty(datatype) {
		customErr := &custom_errors.LengthCheckingError{Type: "datatype"}
		logger.Errorf(customErr.Error())
		return nil, errors.WithStack(customErr)
	}

	latestOnlyFlag := args[3]
	if utils.IsStringEmpty(latestOnlyFlag) {
		customErr := &custom_errors.LengthCheckingError{Type: "latestOnlyFlag"}
		logger.Errorf(customErr.Error())
		return nil, errors.WithStack(customErr)
	}

	if latestOnlyFlag != "true" && latestOnlyFlag != "false" {
		logger.Errorf("Error: Latest only flag must be true or false")
		return nil, errors.New("Error: Latest only flag must be true or false")
	}

	startTimestamp, err := strconv.ParseInt(args[4], 10, 64)
	if err != nil {
		logger.Errorf("Error converting startTimestamp to type int64")
		return nil, errors.Wrap(err, "Error converting startTimestamp to type int64")
	}

	endTimestamp, err := strconv.ParseInt(args[5], 10, 64)
	if err != nil {
		logger.Errorf("Error converting endTimestamp to type int64")
		return nil, errors.Wrap(err, "Error converting endTimestamp to type int64")
	}
	maxNum, err := strconv.ParseInt(args[6], 10, 64)
	if err != nil {
		logger.Errorf("Error converting maxNum to type int")
		return nil, errors.Wrap(err, "Error converting maxNum to type int")
	}

	if maxNum < 0 {
		logger.Errorf("Max num must be greater than 0")
		return nil, errors.New("Max num must be greater than 0")
	}

	timestamp, err := strconv.ParseInt(args[7], 10, 64)
	if err != nil {
		logger.Errorf("Error converting timestamp to type int64")
		return nil, errors.Wrap(err, "Error converting timestamp to type int64")
	}

	// Check timestamp is within 10 mins of current time
	currTime := time.Now().Unix()
	if currTime-timestamp > 10*60 || currTime-timestamp < -10*60 {
		logger.Errorf("Invalid Timestamp (current time: %v)  %v", currTime, timestamp)
		return nil, errors.New("Invalid Timestamp, not within possible time range")
	}

	// ==============================================================
	// Check access and consent
	// ==============================================================
	// If caller is owner, skip checking consent
	callerObj := caller
	if caller.ID != owner {
		// check consent, make sure it's valid
		consent, err := GetConsentInternal(stub, caller, target, datatype, owner)
		if err != nil {
			customErr := &GetConsentError{Consent: "Consent for " + target + ", " + datatype}
			logger.Errorf("%v: %v", customErr, err)
			return nil, errors.Wrap(err, customErr.Error())
		}

		if !utils.InList(consent.Option, consentOptionRead) && !utils.InList(consent.Option, consentOptionWrite) {
			logger.Errorf("Caller does not have read consent to access owner data")
			return nil, errors.New("Caller does not have read consent to access owner data")
		}

		solutionCaller := convertToSolutionUser(caller)
		// Consent target is org, caller is org admin || consent target is service, caller is service admin
		if solutionCaller.Org == target || utils.InList(solutionCaller.SolutionInfo.Services, target) {
			callerObj, err = user_mgmt.GetUserData(stub, caller, target, true, false)
			if err != nil {
				customErr := &GetUserError{User: target}
				logger.Errorf("%v: %v", customErr, err)
				return nil, errors.Wrap(err, customErr.Error())
			}
		} else {
			// owner is service, caller is org admin
			// If caller is org admin, get sym key path and prv key path first
			symKeyPath, prvKeyPath, err := GetUserAssetSymAndPrivateKeyPaths(stub, caller, target)
			if err != nil {
				logger.Errorf("Failed to get symKeyPath and prvKeyPath for user asset")
				return nil, errors.Wrap(err, "Failed to get symKeyPath and prvKeyPath for user asset")
			}

			callerObj, err = user_mgmt.GetUserData(stub, caller, target, true, false, symKeyPath, prvKeyPath)
			if err != nil {
				customErr := &GetUserError{User: target}
				logger.Errorf("%v: %v", customErr, err)
				return nil, errors.Wrap(err, customErr.Error())
			}
		}

		if callerObj.PrivateKey == nil {
			logger.Errorf("Caller does not have access to owner private key")
			return nil, errors.New("Caller does not have access to owner private key")
		}
	}

	// Need to get consent key for logging
	consentKey, err := GetConsentKeyInternal(stub, callerObj, target, datatype, owner)
	if err != nil {
		logger.Errorf("Failed getting consent key")
		return nil, errors.Wrap(err, "Failed getting consent key")
	}

	// ==============================================================
	// Download data using index
	// ==============================================================
	ownerDatas := []OwnerDataResult{}
	if latestOnlyFlag == "true" {
		assetID := GetLatestOwnerDataAssetID(stub, owner, datatype)
		data, err := GetDataWithAssetID(stub, callerObj, assetID, owner, datatype)
		if err != nil {
			logger.Errorf("Failed to get data with assetID")
			return nil, errors.WithStack(err)
		}

		ownerDatas = append(ownerDatas, data)
	} else {
		startValues := []string{owner, datatype}
		if startTimestamp >= 0 {
			startTimestampStr, err := utils.ConvertToString(startTimestamp)
			if err != nil {
				errMsg := "Failed to ConvertToString for startTimestamp"
				logger.Errorf("%v: %v", errMsg, err)
				return nil, errors.Wrap(err, errMsg)
			}
			startValues = append(startValues, startTimestampStr)
		}

		endValues := []string{owner, datatype}
		if endTimestamp > 0 {
			endTimestampStr, err := utils.ConvertToString(endTimestamp)
			if err != nil {
				errMsg := "Failed to ConvertToString for endTimestamp"
				logger.Errorf("%v: %v", errMsg, err)
				return nil, errors.Wrap(err, errMsg)
			}
			endValues = append(endValues, endTimestampStr)
		}

		ownerDatas, err = GetDataInternal(stub, callerObj, []string{"owner", "datatype", "timestamp"}, startValues, endValues, int(maxNum))
		if err != nil {
			customErr := &GetDatasError{FieldNames: []string{"owner", "datatype", "timestamp"}, Values: startValues}
			logger.Errorf("%v: %v", customErr, err)
			return nil, errors.Wrap(err, customErr.Error())
		}
	}

	// ==============================================================
	// Logging
	// ==============================================================
	consentLogSymKey := GetLogSymKeyFromKey(consentKey)

	dataLog := DataLog{Owner: owner, Datatype: datatype, Target: target}
	solutionLog := SolutionLog{
		TransactionID: stub.GetTxID(),
		Namespace:     "OMR",
		FunctionName:  "DownloadOwnerDataWithConsent",
		CallerID:      caller.ID,
		Timestamp:     timestamp,
		Data:          dataLog}
	exportableLog, err := GenerateExportableSolutionLog(stub, caller, solutionLog, consentLogSymKey)
	if err != nil {
		customErr := &GenerateExportableTransactionLogError{Function: solutionLog.FunctionName}
		logger.Errorf("%v: %v", customErr, err)
		return nil, errors.Wrap(err, customErr.Error())
	}

	returnData := OwnerDataResultWithLog{}
	returnData.OwnerDatas = ownerDatas
	returnData.TransactionLog = exportableLog
	logger.Infof("got owner data: %v", len(ownerDatas))
	return json.Marshal(&returnData)
}

// DownloadUserData downloads patient data
// args = [service, patient, datatype, latestOnly, startTimestamp, endTimestamp, maxNum, timestamp]
// If latest only is true, then other filters are ignored
func DownloadUserData(stub cached_stub.CachedStubInterface, caller data_model.User, args []string) ([]byte, error) {
	defer utils.ExitFnLog(utils.EnterFnLog())
	logger.Debugf("args: %v", args)

	if len(args) != 8 {
		customErr := &custom_errors.LengthCheckingError{Type: "DownloadUserData arguments length"}
		logger.Errorf(customErr.Error())
		return nil, errors.WithStack(customErr)
	}

	// ==============================================================
	// Validation
	// ==============================================================
	service := args[0]
	if utils.IsStringEmpty(service) {
		customErr := &custom_errors.LengthCheckingError{Type: "service"}
		logger.Errorf(customErr.Error())
		return nil, errors.WithStack(customErr)
	}

	patient := args[1]
	if utils.IsStringEmpty(patient) {
		customErr := &custom_errors.LengthCheckingError{Type: "patient"}
		logger.Errorf(customErr.Error())
		return nil, errors.WithStack(customErr)
	}

	datatypeID := args[2]
	if utils.IsStringEmpty(datatypeID) {
		customErr := &custom_errors.LengthCheckingError{Type: "datatypeID"}
		logger.Errorf(customErr.Error())
		return nil, errors.WithStack(customErr)
	}

	latestOnlyFlag := args[3]
	if utils.IsStringEmpty(latestOnlyFlag) {
		customErr := &custom_errors.LengthCheckingError{Type: "latestOnlyFlag"}
		logger.Errorf(customErr.Error())
		return nil, errors.WithStack(customErr)
	}

	if latestOnlyFlag != "true" && latestOnlyFlag != "false" {
		logger.Errorf("Error: Latest only flag must be true or false")
		return nil, errors.New("Error: Latest only flag must be true or false")
	}

	startTimestamp, err := strconv.ParseInt(args[4], 10, 64)
	if err != nil {
		logger.Errorf("Error converting startTimestamp to type int64")
		return nil, errors.Wrap(err, "Error converting startTimestamp to type int64")
	}

	endTimestamp, err := strconv.ParseInt(args[5], 10, 64)
	if err != nil {
		logger.Errorf("Error converting endTimestamp to type int64")
		return nil, errors.Wrap(err, "Error converting endTimestamp to type int64")
	}

	maxNum, err := strconv.ParseInt(args[6], 10, 64)
	if err != nil {
		logger.Errorf("Error converting maxNum to type int")
		return nil, errors.Wrap(err, "Error converting maxNum to type int")
	}

	if maxNum < 0 {
		logger.Errorf("Max num must be greater than 0")
		return nil, errors.New("Max num must be greater than 0")
	}

	timestamp, err := strconv.ParseInt(args[7], 10, 64)
	if err != nil {
		logger.Errorf("Error converting timestamp to type int64")
		return nil, errors.Wrap(err, "Error converting maxNum to type int64")
	}

	// Check timestamp is within 10 mins of current time
	currTime := time.Now().Unix()
	if currTime-timestamp > 10*60 || currTime-timestamp < -10*60 {
		logger.Errorf("Invalid Timestamp (current time: %v)  %v", currTime, timestamp)
		return nil, errors.New("Invalid Timestamp, not within possible time range")
	}

	// ==============================================================
	// Check access and consent
	// ==============================================================
	// If caller is consent owner, skip checking consent
	callerObj := caller
	if caller.ID != patient {
		// check consent, make sure it's valid
		consent, err := GetConsentInternal(stub, caller, service, datatypeID, patient)
		if err != nil {
			customErr := &GetConsentError{Consent: "Consent for " + service + ", " + datatypeID}
			logger.Errorf("%v: %v", customErr, err)
			return nil, errors.Wrap(err, customErr.Error())
		}

		if !utils.InList(consent.Option, consentOptionRead) && !utils.InList(consent.Option, consentOptionWrite) {
			logger.Errorf("Caller does not have read consent to access patient data")
			return nil, errors.New("Caller does not have read consent to access patient data")
		}

		// If caller is org admin of consent target, get consent target user and act as consent target user
		solutionCaller := convertToSolutionUser(caller)
		if !utils.InList(solutionCaller.SolutionInfo.Services, service) {
			// If caller is org admin, get sym key path and prv key path first
			symKeyPath, prvKeyPath, err := GetUserAssetSymAndPrivateKeyPaths(stub, caller, consent.Target)
			if err != nil {
				logger.Errorf("Failed to get symKeyPath and prvKeyPath for user asset")
				return nil, errors.Wrap(err, "Failed to get symKeyPath and prvKeyPath for user asset")
			}

			callerObj, err = user_mgmt.GetUserData(stub, caller, consent.Target, true, false, symKeyPath, prvKeyPath)
			if err != nil {
				customErr := &GetUserError{User: consent.Target}
				logger.Errorf("%v: %v", customErr, err)
				return nil, errors.Wrap(err, customErr.Error())
			}
		} else {
			// If caller is service admin, use default path of GetUserData function
			callerObj, err = user_mgmt.GetUserData(stub, caller, consent.Target, true, false)
			if err != nil {
				customErr := &GetUserError{User: consent.Target}
				logger.Errorf("%v: %v", customErr, err)
				return nil, errors.Wrap(err, customErr.Error())
			}
		}

		if callerObj.PrivateKey == nil {
			logger.Errorf("Caller does not have access to consent target private key")
			return nil, errors.New("Caller does not have access to consent target private key")
		}
	}

	// ==============================================================
	// Download
	// ==============================================================
	patientDatas := []OwnerDataResult{}
	if latestOnlyFlag == "true" {
		assetID := GetLatestPatientDataAssetID(stub, patient, datatypeID)
		data, err := GetDataWithAssetID(stub, callerObj, assetID, patient, datatypeID)
		if err != nil {
			logger.Errorf("Failed to get data with assetID")
			return nil, errors.WithStack(err)
		}

		patientDatas = append(patientDatas, data)

	} else {
		startValues := []string{patient, datatypeID}
		if startTimestamp >= 0 {
			startTimestampStr, err := utils.ConvertToString(startTimestamp)
			if err != nil {
				errMsg := "Failed to ConvertToString for startTimestamp"
				logger.Errorf("%v: %v", errMsg, err)
				return nil, errors.Wrap(err, errMsg)
			}
			startValues = append(startValues, startTimestampStr)
		}

		endValues := []string{patient, datatypeID}
		if endTimestamp > 0 {
			endTimestampStr, err := utils.ConvertToString(endTimestamp)
			if err != nil {
				errMsg := "Failed to ConvertToString for endTimestamp"
				logger.Errorf("%v: %v", errMsg, err)
				return nil, errors.Wrap(err, errMsg)
			}
			endValues = append(endValues, endTimestampStr)
		}

		patientDatas, err = GetDataInternal(stub, callerObj, []string{"owner", "datatype", "timestamp"}, startValues, endValues, int(maxNum))
		if err != nil {
			customErr := &GetDatasError{FieldNames: []string{"owner", "datatype", "timestamp"}, Values: startValues}
			logger.Errorf("%v: %v", customErr, err)
			return nil, errors.Wrap(err, customErr.Error())
		}
	}

	if len(patientDatas) == 0 {
		logger.Errorf("Download failed, got 0 data")
		return nil, errors.New("Download failed, got 0 data")
	}
	// ==============================================================
	// Logging
	// ==============================================================
	assetManager := asset_mgmt.GetAssetManager(stub, callerObj)

	enrollmentID := GetEnrollmentID(patient, service)
	enrollmentAssetID := asset_mgmt.GetAssetId(EnrollmentAssetNamespace, enrollmentID)

	keyPath, err := GetKeyPath(stub, callerObj, enrollmentAssetID)
	if err != nil {
		customErr := &GetKeyPathError{Caller: callerObj.ID, AssetID: enrollmentAssetID}
		logger.Errorf(customErr.Error())
		return nil, errors.New(customErr.Error())
	}

	if len(keyPath) <= 0 {
		customErr := &GetKeyPathError{Caller: caller.ID, AssetID: enrollmentAssetID}
		logger.Errorf(customErr.Error())
		return nil, errors.New(customErr.Error())
	}

	enrollmentKey, err := assetManager.GetAssetKey(enrollmentAssetID, keyPath)
	if err != nil {
		logger.Errorf("Failed to GetAssetKey for enrollmentKey: %v", err)
		return nil, errors.Wrap(err, "Failed to GetAssetKey for enrollmentKey")
	}

	enrollmentLogSymKey := GetLogSymKeyFromKey(enrollmentKey)

	dataLog := DataLog{Owner: patient, Datatype: datatypeID, Target: service}
	solutionLog := SolutionLog{
		TransactionID: stub.GetTxID(),
		Namespace:     "OMR",
		FunctionName:  "DownloadUserData",
		CallerID:      caller.ID,
		Timestamp:     timestamp,
		Data:          dataLog}
	exportableLog, err := GenerateExportableSolutionLog(stub, caller, solutionLog, enrollmentLogSymKey)
	if err != nil {
		customErr := &GenerateExportableTransactionLogError{Function: solutionLog.FunctionName}
		logger.Errorf("%v: %v", customErr, err)
		return nil, errors.Wrap(err, customErr.Error())
	}

	returnData := OwnerDataResultWithLog{}
	returnData.OwnerDatas = patientDatas
	returnData.TransactionLog = exportableLog
	logger.Infof("got patient data: %v", len(patientDatas))
	return json.Marshal(&returnData)
}

// DownloadUserDataConsentToken checks incoming consent validation token, if passes, calls GetDataInternal
// args = [latest_only, start_timestamp, end_timestamp, maxNum, timestamp, token]
func DownloadUserDataConsentToken(stub cached_stub.CachedStubInterface, caller data_model.User, args []string) ([]byte, error) {
	defer utils.ExitFnLog(utils.EnterFnLog())
	logger.Debugf("args: %v", args)

	if len(args) != 6 {
		customErr := &custom_errors.LengthCheckingError{Type: "DownloadUserDataConsentToken arguments length"}
		logger.Errorf(customErr.Error())
		return nil, errors.New(customErr.Error())
	}

	// ==============================================================
	// Validation
	// ==============================================================
	latestOnlyFlag := args[0]
	if utils.IsStringEmpty(latestOnlyFlag) {
		customErr := &custom_errors.LengthCheckingError{Type: "latestOnlyFlag"}
		logger.Errorf(customErr.Error())
		return nil, errors.WithStack(customErr)
	}

	if latestOnlyFlag != "true" && latestOnlyFlag != "false" {
		logger.Errorf("Error: Latest only flag must be true or false")
		return nil, errors.New("Error: Latest only flag must be true or false")
	}

	startTimestamp, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		logger.Errorf("Error converting startTimestamp to type int64")
		return nil, errors.Wrap(err, "Error converting startTimestamp to type int64")
	}

	endTimestamp, err := strconv.ParseInt(args[2], 10, 64)
	if err != nil {
		logger.Errorf("Error converting endTimestamp to type int64")
		return nil, errors.Wrap(err, "Error converting endTimestamp to type int64")
	}

	maxNum, err := strconv.ParseInt(args[3], 10, 64)
	if err != nil {
		logger.Errorf("Error converting maxNum to type int")
		return nil, errors.Wrap(err, "Error converting maxNum to type int")
	}

	if maxNum < 0 {
		logger.Errorf("Max num must be greater than 0")
		return nil, errors.New("Max num must be greater than 0")
	}

	timestamp, err := strconv.ParseInt(args[4], 10, 64)
	if err != nil {
		logger.Errorf("Error converting timestamp to type int64")
		return nil, errors.Wrap(err, "Error converting maxNum to type int64")
	}

	// Check timestamp is within 10 mins of current time
	currTime := time.Now().Unix()
	if currTime-timestamp > 10*60 || currTime-timestamp < -10*60 {
		logger.Errorf("Invalid Timestamp (current time: %v)  %v", currTime, timestamp)
		return nil, errors.New("Invalid Timestamp, not within possible time range")
	}

	tokenB64 := args[5]
	if utils.IsStringEmpty(tokenB64) {
		customErr := &custom_errors.LengthCheckingError{Type: "tokenB64"}
		logger.Errorf(customErr.Error())
		return nil, errors.WithStack(customErr)
	}

	token, err := DecryptConsentValidationToken(stub, caller, tokenB64)
	if err != nil {
		customErr := &DecryptTokenError{}
		return nil, errors.Wrap(err, customErr.Error())
	}

	// ==============================================================
	// Construct caller object
	// ==============================================================

	callerObj := caller
	// If caller is consent owner, skip changing caller
	if caller.ID != token.Owner {
		// If caller is org admin of consent target, get consent target user and act as consent target user
		solutionCaller := convertToSolutionUser(caller)
		if !utils.InList(solutionCaller.SolutionInfo.Services, token.Target) {
			// If caller is org admin, get sym key path and prv key path first
			symKeyPath, prvKeyPath, err := GetUserAssetSymAndPrivateKeyPaths(stub, caller, token.Target)
			if err != nil {
				logger.Errorf("Failed to get symKeyPath and prvKeyPath for user asset")
				return nil, errors.Wrap(err, "Failed to get symKeyPath and prvKeyPath for user asset")
			}

			callerObj, err = user_mgmt.GetUserData(stub, caller, token.Target, true, false, symKeyPath, prvKeyPath)
			if err != nil {
				customErr := &GetUserError{User: token.Target}
				logger.Errorf("%v: %v", customErr, err)
				return nil, errors.Wrap(err, customErr.Error())
			}
		} else {
			// If caller is service admin, use default path of GetUserData function
			callerObj, err = user_mgmt.GetUserData(stub, caller, token.Target, true, false)
			if err != nil {
				customErr := &GetUserError{User: token.Target}
				logger.Errorf("%v: %v", customErr, err)
				return nil, errors.Wrap(err, customErr.Error())
			}
		}

		if callerObj.PrivateKey == nil {
			logger.Errorf("Caller does not have access to consent target private key")
			return nil, errors.New("Caller does not have access to consent target private key")
		}
	}

	// ==============================================================
	// Download data
	// ==============================================================
	patientDatas := []OwnerDataResult{}

	if latestOnlyFlag == "true" {
		assetID := GetLatestPatientDataAssetID(stub, token.Owner, token.Datatype)
		data, err := GetDataWithAssetID(stub, callerObj, assetID, token.Owner, token.Datatype)
		if err != nil {
			logger.Errorf("Failed to get data with assetID")
			return nil, errors.WithStack(err)
		}

		patientDatas = append(patientDatas, data)

	} else {
		startValues := []string{token.Owner, token.Datatype}
		if startTimestamp >= 0 {
			startTimestampStr, err := utils.ConvertToString(startTimestamp)
			if err != nil {
				errMsg := "Failed to ConvertToString for startTimestamp"
				logger.Errorf("%v: %v", errMsg, err)
				return nil, errors.Wrap(err, errMsg)
			}
			startValues = append(startValues, startTimestampStr)
		}

		endValues := []string{token.Owner, token.Datatype}
		if endTimestamp > 0 {
			endTimestampStr, err := utils.ConvertToString(endTimestamp)
			if err != nil {
				errMsg := "Failed to ConvertToString for endTimestamp"
				logger.Errorf("%v: %v", errMsg, err)
				return nil, errors.Wrap(err, errMsg)
			}
			endValues = append(endValues, endTimestampStr)
		}

		patientDatas, err = GetDataInternal(stub, callerObj, []string{"owner", "datatype"}, startValues, endValues, int(maxNum))
		if err != nil {
			customErr := &GetDatasError{FieldNames: []string{"owner", "datatype", "timestamp"}, Values: startValues}
			logger.Errorf("%v: %v", customErr, err)
			return nil, errors.Wrap(err, customErr.Error())
		}
	}

	// ==============================================================
	// Logging
	// ==============================================================
	assetManager := asset_mgmt.GetAssetManager(stub, callerObj)

	enrollmentID := GetEnrollmentID(token.Owner, token.Target)
	enrollmentAssetID := asset_mgmt.GetAssetId(EnrollmentAssetNamespace, enrollmentID)

	keyPath, err := GetKeyPath(stub, callerObj, enrollmentAssetID)
	if err != nil {
		customErr := &GetKeyPathError{Caller: caller.ID, AssetID: enrollmentAssetID}
		logger.Errorf(customErr.Error())
		return nil, errors.New(customErr.Error())
	}

	if len(keyPath) <= 0 {
		customErr := &GetKeyPathError{Caller: caller.ID, AssetID: enrollmentAssetID}
		logger.Errorf(customErr.Error())
		return nil, errors.New(customErr.Error())
	}

	enrollmentKey, err := assetManager.GetAssetKey(enrollmentAssetID, keyPath)
	if err != nil {
		logger.Errorf("Failed to GetAssetKey for enrollmentKey: %v", err)
		return nil, errors.Wrap(err, "Failed to GetAssetKey for enrollmentKey")
	}

	enrollmentLogSymKey := GetLogSymKeyFromKey(enrollmentKey)

	dataLog := DataLog{Owner: token.Owner, Datatype: token.Datatype, Target: token.Target}
	solutionLog := SolutionLog{
		TransactionID: stub.GetTxID(),
		Namespace:     "OMR",
		FunctionName:  "DownloadUserDataConsentToken",
		CallerID:      caller.ID,
		Timestamp:     timestamp,
		Data:          dataLog}
	exportableLog, err := GenerateExportableSolutionLog(stub, caller, solutionLog, enrollmentLogSymKey)
	if err != nil {
		customErr := &GenerateExportableTransactionLogError{Function: solutionLog.FunctionName}
		logger.Errorf("%v: %v", customErr, err)
		return nil, errors.Wrap(err, customErr.Error())
	}

	returnData := OwnerDataResultWithLog{}
	returnData.OwnerDatas = patientDatas
	returnData.TransactionLog = exportableLog
	logger.Infof("got patient data: %v", len(patientDatas))
	return json.Marshal(&returnData)
}

// DownloadOwnerDataConsentToken checks incoming consent validation token, if passes, calls GetDataInternal
// args = [latest_only, start_timestamp, end_timestamp, maxNum, timestamp, token]
func DownloadOwnerDataConsentToken(stub cached_stub.CachedStubInterface, caller data_model.User, args []string) ([]byte, error) {
	defer utils.ExitFnLog(utils.EnterFnLog())
	logger.Debugf("args: %v", args)

	if len(args) != 6 {
		customErr := &custom_errors.LengthCheckingError{Type: "DownloadOwnerDataConsentToken arguments length"}
		logger.Errorf(customErr.Error())
		return nil, errors.New(customErr.Error())
	}

	// ==============================================================
	// Validation
	// ==============================================================
	latestOnlyFlag := args[0]
	if utils.IsStringEmpty(latestOnlyFlag) {
		customErr := &custom_errors.LengthCheckingError{Type: "latestOnlyFlag"}
		logger.Errorf(customErr.Error())
		return nil, errors.WithStack(customErr)
	}

	if latestOnlyFlag != "true" && latestOnlyFlag != "false" {
		logger.Errorf("Error: Latest only flag must be true or false")
		return nil, errors.New("Error: Latest only flag must be true or false")
	}

	startTimestamp, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		logger.Errorf("Error converting startTimestamp to type int64")
		return nil, errors.Wrap(err, "Error converting startTimestamp to type int64")
	}

	endTimestamp, err := strconv.ParseInt(args[2], 10, 64)
	if err != nil {
		logger.Errorf("Error converting endTimestamp to type int64")
		return nil, errors.Wrap(err, "Error converting endTimestamp to type int64")
	}

	maxNum, err := strconv.ParseInt(args[3], 10, 64)
	if err != nil {
		logger.Errorf("Error converting maxNum to type int")
		return nil, errors.Wrap(err, "Error converting maxNum to type int")
	}

	if maxNum < 0 {
		logger.Errorf("Max num must be greater than 0")
		return nil, errors.New("Max num must be greater than 0")
	}

	timestamp, err := strconv.ParseInt(args[4], 10, 64)
	if err != nil {
		logger.Errorf("Error converting timestamp to type int64")
		return nil, errors.Wrap(err, "Error converting maxNum to type int64")
	}

	// Check timestamp is within 10 mins of current time
	currTime := time.Now().Unix()
	if currTime-timestamp > 10*60 || currTime-timestamp < -10*60 {
		logger.Errorf("Invalid Timestamp (current time: %v)  %v", currTime, timestamp)
		return nil, errors.New("Invalid Timestamp, not within possible time range")
	}

	tokenB64 := args[5]
	if utils.IsStringEmpty(tokenB64) {
		customErr := &custom_errors.LengthCheckingError{Type: "tokenB64"}
		logger.Errorf(customErr.Error())
		return nil, errors.WithStack(customErr)
	}

	token, err := DecryptConsentValidationToken(stub, caller, tokenB64)
	if err != nil {
		customErr := &DecryptTokenError{}
		return nil, errors.Wrap(err, customErr.Error())
	}

	// ==============================================================
	// Construct caller object
	// ==============================================================
	callerObj := caller
	// If caller is consent owner, skip changing caller
	if caller.ID != token.Owner {
		// If caller is org admin of consent target, get consent target user and act as consent target user
		solutionCaller := convertToSolutionUser(caller)
		if !utils.InList(solutionCaller.SolutionInfo.Services, token.Target) {
			// If caller is org admin, get sym key path and prv key path first
			symKeyPath, prvKeyPath, err := GetUserAssetSymAndPrivateKeyPaths(stub, caller, token.Target)
			if err != nil {
				logger.Errorf("Failed to get symKeyPath and prvKeyPath for user asset")
				return nil, errors.Wrap(err, "Failed to get symKeyPath and prvKeyPath for user asset")
			}

			callerObj, err = user_mgmt.GetUserData(stub, caller, token.Target, true, false, symKeyPath, prvKeyPath)
			if err != nil {
				customErr := &GetUserError{User: token.Target}
				logger.Errorf("%v: %v", customErr, err)
				return nil, errors.Wrap(err, customErr.Error())
			}
		} else {
			// If caller is service admin, use default path of GetUserData function
			callerObj, err = user_mgmt.GetUserData(stub, caller, token.Target, true, false)
			if err != nil {
				customErr := &GetUserError{User: token.Target}
				logger.Errorf("%v: %v", customErr, err)
				return nil, errors.Wrap(err, customErr.Error())
			}
		}

		if callerObj.PrivateKey == nil {
			logger.Errorf("Caller does not have access to consent target private key")
			return nil, errors.New("Caller does not have access to consent target private key")
		}
	}

	// ==============================================================
	// Download data
	// ==============================================================
	ownerDatas := []OwnerDataResult{}

	if latestOnlyFlag == "true" {
		assetID := GetLatestOwnerDataAssetID(stub, token.Owner, token.Datatype)
		data, err := GetDataWithAssetID(stub, callerObj, assetID, token.Owner, token.Datatype)
		if err != nil {
			logger.Errorf("Failed to get data with assetID")
			return nil, errors.WithStack(err)
		}

		ownerDatas = append(ownerDatas, data)

	} else {
		startValues := []string{token.Owner, token.Datatype}
		if startTimestamp >= 0 {
			startTimestampStr, err := utils.ConvertToString(startTimestamp)
			if err != nil {
				errMsg := "Failed to ConvertToString for startTimestamp"
				logger.Errorf("%v: %v", errMsg, err)
				return nil, errors.Wrap(err, errMsg)
			}
			startValues = append(startValues, startTimestampStr)
		}

		endValues := []string{token.Owner, token.Datatype}
		if endTimestamp > 0 {
			endTimestampStr, err := utils.ConvertToString(endTimestamp)
			if err != nil {
				errMsg := "Failed to ConvertToString for endTimestamp"
				logger.Errorf("%v: %v", errMsg, err)
				return nil, errors.Wrap(err, errMsg)
			}
			endValues = append(endValues, endTimestampStr)
		}

		ownerDatas, err = GetDataInternal(stub, callerObj, []string{"owner", "datatype"}, startValues, endValues, int(maxNum))
		if err != nil {
			customErr := &GetDatasError{FieldNames: []string{"owner", "datatype", "timestamp"}, Values: startValues}
			logger.Errorf("%v: %v", customErr, err)
			return nil, errors.Wrap(err, customErr.Error())
		}
	}

	// ==============================================================
	// Logging
	// ==============================================================
	consentKey := data_model.Key{ID: consent_mgmt.GetConsentID(token.Datatype, token.Target, token.Owner), Type: key_mgmt.KEY_TYPE_SYM}
	consentKey.KeyBytes = token.ConsentKey
	consentLogSymKey := GetLogSymKeyFromKey(consentKey)

	dataLog := DataLog{Owner: token.Owner, Datatype: token.Datatype, Target: token.Target}
	solutionLog := SolutionLog{
		TransactionID: stub.GetTxID(),
		Namespace:     "OMR",
		FunctionName:  "DownloadOwnerDataConsentToken",
		CallerID:      caller.ID,
		Timestamp:     timestamp,
		Data:          dataLog}
	exportableLog, err := GenerateExportableSolutionLog(stub, caller, solutionLog, consentLogSymKey)
	if err != nil {
		customErr := &GenerateExportableTransactionLogError{Function: solutionLog.FunctionName}
		logger.Errorf("%v: %v", customErr, err)
		return nil, errors.Wrap(err, customErr.Error())
	}

	returnData := OwnerDataResultWithLog{}
	returnData.OwnerDatas = ownerDatas
	returnData.TransactionLog = exportableLog
	logger.Infof("got owner data: %v", len(ownerDatas))
	return json.Marshal(&returnData)
}

// DeleteUserData deletes patient data
// Only data owner can delete data
func DeleteUserData(stub cached_stub.CachedStubInterface, caller data_model.User, args []string) ([]byte, error) {
	defer utils.ExitFnLog(utils.EnterFnLog())
	logger.Debugf("args: %v", args)
	return nil, nil
}

func convertOwnerDataToAsset(stub cached_stub.CachedStubInterface, data OwnerData) (data_model.Asset, error) {
	defer utils.ExitFnLog(utils.EnterFnLog())
	asset := data_model.Asset{}
	asset.AssetId = asset_mgmt.GetAssetId(OwnerDataNamespace, data.DataID)
	asset.Datatypes = []string{data.Datatype}
	metaData := make(map[string]string)
	metaData["namespace"] = OwnerDataNamespace
	asset.Metadata = metaData
	var publicData interface{}
	asset.PublicData, _ = json.Marshal(&publicData)
	asset.PrivateData, _ = json.Marshal(&data)
	asset.OwnerIds = []string{data.Owner}
	asset.IndexTableName = IndexData

	// save asset to offchain-datastore, if one is setup
	dsConnectionID, err := GetActiveConnectionID(stub)
	if err != nil {
		errMsg := "Failed to GetActiveConnectionID"
		logger.Errorf("%v: %v", errMsg, err)
		return data_model.Asset{}, errors.Wrap(err, errMsg)
	}
	if !utils.IsStringEmpty(dsConnectionID) {
		asset.SetDatastoreConnectionID(dsConnectionID)
	}

	return asset, nil
}

// private function that converts asset to data
func convertOwnerDataFromAsset(asset *data_model.Asset) OwnerDataResult {
	defer utils.ExitFnLog(utils.EnterFnLog())
	ownerData := OwnerDataResult{}
	json.Unmarshal(asset.PrivateData, &ownerData)
	return ownerData
}

// SetupDataIndex sets up data index table for data operations
func SetupDataIndex(stub cached_stub.CachedStubInterface) error {
	defer utils.ExitFnLog(utils.EnterFnLog())

	dataTable := index.GetTable(stub, IndexData, "data_id")
	dataTable.AddIndex([]string{"owner", "datatype", "timestamp", "data_id"}, false)
	err := dataTable.SaveToLedger()
	if err != nil {
		return err
	}
	return nil
}
