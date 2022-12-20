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
	"common/bchcls/asset_mgmt/asset_manager"
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
	"time"

	"github.com/pkg/errors"
)

const IndexEnrollment = "EnrollmentTable"
const EnrollmentPrefix = "Enrollment"
const EnrollmentAssetNamespace = "EnrollmentAsset"

// De-identified fields:
//   - UserID
//   - UserName
//   - ServiceID
//   - ServiceName
type Enrollment struct {
	EnrollmentID string `json:"enrollment_id"`
	UserID       string `json:"user_id"`
	UserName     string `json:"user_name"`
	ServiceID    string `json:"service_id"`
	ServiceName  string `json:"service_name"`
	EnrollDate   int64  `json:"enroll_date"`
	Status       string `json:"status"`
}

type EnrollmentResult struct {
	UserID      string `json:"user_id"`
	UserName    string `json:"user_name"`
	ServiceID   string `json:"service_id"`
	ServiceName string `json:"service_name"`
	EnrollDate  int64  `json:"enroll_date"`
	Status      string `json:"status"`
}

type enrollmentPublicData struct {
	EnrollmentID string `json:"enrollment_id"`
	UserID       string `json:"user_id"`
	UserName     string `json:"user_name"`
	ServiceID    string `json:"service_id"`
	ServiceName  string `json:"service_name"`
}

type enrollmentPrivateData struct {
	EnrollDate int64  `json:"enroll_date"`
	Status     string `json:"status"`
}

// Enroll patient
// 1) Validate enrollmentSymKey and enrollment object
// 2) Encrypt enrollmentKey with serviceSymKey
// 3) Encrypt enrollmentKey with userSymKey
// 4) Save enrollment object as asset with enrollmentKey as assetKey
// args = [ enrollmentBytes, enrollmentSymKeyB64 ]
func EnrollPatient(stub cached_stub.CachedStubInterface, caller data_model.User, args []string) ([]byte, error) {
	defer utils.ExitFnLog(utils.EnterFnLog())
	logger.Debugf("args: %v", args)

	if len(args) != 2 {
		customErr := &custom_errors.LengthCheckingError{Type: "EnrollPatient arguments length"}
		logger.Errorf(customErr.Error())
		return nil, errors.New(customErr.Error())
	}

	// ==============================================================
	// Validate incoming enrollement object and sym key
	// ==============================================================
	enrollmentBytes := []byte(args[0])
	enrollment := Enrollment{}
	err := json.Unmarshal(enrollmentBytes, &enrollment)
	if err != nil {
		customErr := &custom_errors.UnmarshalError{Type: "Enrollment"}
		logger.Errorf("%v: %v", customErr, err)
		return nil, errors.Wrap(err, customErr.Error())
	}

	if utils.IsStringEmpty(enrollment.UserID) {
		customErr := &custom_errors.LengthCheckingError{Type: "UserID"}
		logger.Errorf(customErr.Error())
		return nil, errors.WithStack(customErr)
	}

	if enrollment.UserID == caller.ID {
		logger.Errorf("Caller and enrolled patient must be different")
		return nil, errors.New("Caller and enrolled patient must be different")
	}

	// check that EnrollDate is within 10 mins of current time
	currTime := time.Now().Unix()
	if currTime-enrollment.EnrollDate > 10*60 || currTime-enrollment.EnrollDate < -10*60 {
		logger.Errorf("Invalid EnrollDate (current time: %v)  %v", currTime, enrollment.EnrollDate)
		return nil, errors.New("Invalid EnrollDate, not within possible time range")
	}

	enrollmentID := GetEnrollmentID(enrollment.UserID, enrollment.ServiceID)
	enrollment.EnrollmentID = enrollmentID
	enrollmentKey := data_model.Key{ID: key_mgmt.GetSymKeyId(enrollmentID), Type: key_mgmt.KEY_TYPE_SYM}
	enrollmentKey.KeyBytes, err = crypto.ParseSymKeyB64(args[1])
	if err != nil {
		logger.Errorf("Invalid enrollmentSymKey")
		return nil, errors.Wrap(err, "Invalid enrollmentSymKey")
	}

	if enrollmentKey.KeyBytes == nil {
		logger.Errorf("Invalid enrollmentSymKey")
		return nil, errors.New("Invalid enrollmentSymKey")
	}

	// make sure user exists
	// also get enrollment username from existing user
	existingUser, err := user_mgmt.GetUserData(stub, caller, enrollment.UserID, false, false)
	if err != nil {
		customErr := &GetUserError{User: "enrollment user"}
		logger.Errorf("%v: %v", customErr, err)
		return nil, errors.Wrap(err, customErr.Error())
	}

	if utils.IsStringEmpty(existingUser.ID) {
		customErr := &custom_errors.LengthCheckingError{Type: "existingUser.UserID"}
		logger.Errorf(customErr.Error())
		return nil, errors.WithStack(customErr)
	}

	enrollment.UserName = existingUser.Name

	if utils.IsStringEmpty(enrollment.ServiceID) {
		customErr := &custom_errors.LengthCheckingError{Type: "ServiceID"}
		logger.Errorf(customErr.Error())
		return nil, errors.WithStack(customErr)
	}

	existingService, err := user_mgmt.GetUserData(stub, caller, enrollment.ServiceID, true, false)
	if err != nil {
		customErr := &GetServiceError{Service: enrollment.ServiceID}
		logger.Errorf("%v: %v", customErr, err)
		return nil, errors.Wrap(err, customErr.Error())
	}

	serviceOrg := GetOrgIDFromServiceSubgroup(existingService)
	if utils.IsStringEmpty(serviceOrg) {
		customErr := &custom_errors.LengthCheckingError{Type: "serviceOrg"}
		logger.Errorf(customErr.Error())
		return nil, errors.WithStack(customErr)
	}

	// Check that caller is an service admin or org admin of this service
	if !CallerIsAdminOfService(caller, enrollment.ServiceID, serviceOrg) {
		logger.Error("Caller is not admin of the service")
		return nil, errors.New("Caller is not admin of the service")
	}

	enrollment.ServiceName = existingService.Name

	if enrollment.Status != "active" {
		logger.Error("Enrollment status must be active")
		return nil, errors.New("Enrollment status must be active")
	}

	// ==============================================================
	// Check existing enrollment
	// ==============================================================
	enrollmentExists, err := CheckEnrollmentExists(stub, caller, enrollmentID)
	if err != nil {
		customErr := &GetEnrollmentError{Enrollment: enrollmentID}
		logger.Errorf("%v: %v", customErr, err)
		return nil, errors.Wrap(err, customErr.Error())
	}

	// Already enrolled, just return
	if enrollmentExists {
		return UpdateEnrollmentInternal(stub, caller, enrollment)
	}

	// ==============================================================
	// Save enrollment as asset
	// ==============================================================
	// Convert enrollment object to Asset
	enrollmentAsset, err := convertEnrollmentToAsset(stub, enrollment)
	if err != nil {
		customErr := &ConvertToAssetError{Asset: "enrollmentAsset"}
		logger.Errorf("%v: %v", customErr, err)
		return nil, errors.Wrap(err, customErr.Error())
	}

	assetManager := asset_mgmt.GetAssetManager(stub, caller)
	err = assetManager.AddAsset(enrollmentAsset, enrollmentKey, false)
	if err != nil {
		customErr := &PutAssetError{Asset: enrollmentID}
		logger.Errorf("%v: %v", customErr, err)
		return nil, errors.Wrap(err, customErr.Error())
	}

	// ==============================================================
	// Establish key relationships
	// ==============================================================
	servicePubKey := existingService.GetPublicKey()
	userPubKey := existingUser.GetPublicKey()

	// add access from servicePubKey to enrollmentSymKey
	userAccessManager := user_access_ctrl.GetUserAccessManager(stub, caller)
	err = userAccessManager.AddAccessByKey(servicePubKey, enrollmentKey)
	if err != nil {
		customErr := &custom_errors.AddAccessError{Key: "service pub key to enrollment key"}
		logger.Errorf("%v: %v", customErr, err)
		return nil, errors.Wrap(err, customErr.Error())
	}

	// add access from userPubKey to enrollmentSymKey
	err = userAccessManager.AddAccessByKey(userPubKey, enrollmentKey)
	if err != nil {
		customErr := &custom_errors.AddAccessError{Key: "user pub key to enrollment key"}
		logger.Errorf("%v: %v", customErr, err)
		return nil, errors.Wrap(err, customErr.Error())
	}

	// ==============================================================
	// Add access from service log sym key to enrollmentLogSymKey
	// ==============================================================
	// EnrollmentLogSymKey is used as log sym key for PutConsent
	serviceLogSymKey := existingService.GetLogSymKey()
	enrollmentLogSymKey := GetLogSymKeyFromKey(enrollmentKey)
	err = userAccessManager.AddAccessByKey(serviceLogSymKey, enrollmentLogSymKey)
	if err != nil {
		customErr := &custom_errors.AddAccessError{Key: "service log sym key to enrollment log sym key"}
		logger.Errorf("%v: %v", customErr, err)
		return nil, errors.Wrap(err, customErr.Error())
	}

	return nil, nil
}

// Unenroll patient
// args = [ serviceID, userID ]
func UnenrollPatient(stub cached_stub.CachedStubInterface, caller data_model.User, args []string) ([]byte, error) {
	defer utils.ExitFnLog(utils.EnterFnLog())
	logger.Debugf("args: %v", args)

	if len(args) != 2 {
		customErr := &custom_errors.LengthCheckingError{Type: "UnenrollPatient arguments length"}
		logger.Errorf(customErr.Error())
		return nil, errors.New(customErr.Error())
	}

	// ==============================================================
	// Validation
	// ==============================================================

	serviceID := args[0]
	if utils.IsStringEmpty(serviceID) {
		customErr := &custom_errors.LengthCheckingError{Type: "ServiceID"}
		logger.Errorf(customErr.Error())
		return nil, errors.WithStack(customErr)
	}

	userID := args[1]
	if utils.IsStringEmpty(userID) {
		customErr := &custom_errors.LengthCheckingError{Type: "userID"}
		logger.Errorf(customErr.Error())
		return nil, errors.WithStack(customErr)
	}

	// make sure service exists
	existingService, err := GetServiceInternal(stub, caller, serviceID, false)
	if err != nil {
		customErr := &GetServiceError{Service: serviceID}
		logger.Errorf("%v: %v", customErr, err)
		return nil, errors.Wrap(err, customErr.Error())
	}

	if utils.IsStringEmpty(existingService.ServiceID) {
		customErr := &custom_errors.LengthCheckingError{Type: "existingService.ServiceID"}
		logger.Errorf(customErr.Error())
		return nil, errors.WithStack(customErr)
	}

	if !CallerIsAdminOfService(caller, serviceID, existingService.OrgID) {
		logger.Error("Caller is not admin of the service")
		return nil, errors.New("Caller is not admin of the service")
	}

	enrollment, err := GetEnrollmentInternal(stub, caller, userID, serviceID)
	if err != nil {
		customErr := &GetEnrollmentError{Enrollment: GetEnrollmentID(userID, serviceID)}
		logger.Errorf(customErr.Error())
		return nil, errors.New(customErr.Error())
	}

	enrollment.Status = "inactive"
	return UpdateEnrollmentInternal(stub, caller, enrollment)
}

// Internal function for updating enrollment
func UpdateEnrollmentInternal(stub cached_stub.CachedStubInterface, caller data_model.User, enrollment Enrollment) ([]byte, error) {
	defer utils.ExitFnLog(utils.EnterFnLog())

	// have to update as either default org admin or default service admin
	// because hasWriteAccessToAsset function in UpdateAsset does not check indirect org admin relationship
	// so we have to substitute callers
	callerObj := caller
	// if caller is org admin, use org as caller
	solutionCaller := convertToSolutionUser(caller)
	if solutionCaller.SolutionInfo.IsOrgAdmin {
		orgCaller, err := user_mgmt.GetUserData(stub, caller, solutionCaller.Org, true, false)
		if err != nil {
			customErr := &GetOrgError{Org: solutionCaller.Org}
			logger.Errorf("%v: %v", customErr, err)
			return nil, errors.Wrap(err, customErr.Error())
		}

		if orgCaller.PrivateKey == nil {
			errMsg := "Caller does not have access to org private key"
			logger.Errorf(errMsg)
			return nil, errors.New(errMsg)
		} else {
			callerObj = orgCaller
		}
	} else {
		// if caller is service admin, use service as caller
		// construct new caller object representing service itself
		serviceCaller, err := user_mgmt.GetUserData(stub, caller, enrollment.ServiceID, true, false)
		if err != nil {
			customErr := &GetOrgError{Org: enrollment.ServiceID}
			logger.Errorf("%v: %v", customErr, err)
			return nil, errors.Wrap(err, customErr.Error())
		}

		// if not org admin then must be service admin to be able to update service
		if serviceCaller.PrivateKey == nil {
			logger.Errorf("Caller does not have access to service private key")
			return nil, errors.New("Caller does not have access to service private key")
		} else {
			callerObj = serviceCaller
		}
	}

	assetManager := asset_mgmt.GetAssetManager(stub, callerObj)

	enrollmentAssetID := asset_mgmt.GetAssetId(EnrollmentAssetNamespace, enrollment.EnrollmentID)
	// construct key path to get enrollment key for update
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

	// Convert enrollment object to Asset
	enrollmentAsset, err := convertEnrollmentToAsset(stub, enrollment)
	if err != nil {
		customErr := &ConvertToAssetError{Asset: "enrollmentAsset"}
		logger.Errorf("%v: %v", customErr, err)
		return nil, errors.Wrap(err, customErr.Error())
	}
	enrollmentAsset.AssetKeyId = enrollmentKey.ID

	err = assetManager.UpdateAsset(enrollmentAsset, enrollmentKey, true)
	if err != nil {
		customErr := &PutAssetError{Asset: enrollment.EnrollmentID}
		logger.Errorf("%v: %v", customErr, err)
		return nil, errors.Wrap(err, customErr.Error())
	}

	return nil, nil
}

// GetPatientEnrollments gets all enrollments of an user
// args = [ userID, statusFilter ]
// statusFilter is an optional parameter for filtering enrollment status
func GetPatientEnrollments(stub cached_stub.CachedStubInterface, caller data_model.User, args []string) ([]byte, error) {
	defer utils.ExitFnLog(utils.EnterFnLog())
	logger.Debugf("args: %v", args)

	if len(args) != 1 && len(args) != 2 {
		customErr := &custom_errors.LengthCheckingError{Type: "GetPatientEnrollments arguments length"}
		logger.Errorf(customErr.Error())
		return nil, errors.New(customErr.Error())
	}

	// ==============================================================
	// Validation
	// ==============================================================
	userID := args[0]
	if utils.IsStringEmpty(userID) {
		customErr := &custom_errors.LengthCheckingError{Type: "userID"}
		logger.Errorf(customErr.Error())
		return nil, errors.WithStack(customErr)
	}

	statusFilter := ""
	if len(args) == 2 {
		statusFilter = args[1]
	}

	// Validate status filter
	if !utils.IsStringEmpty(statusFilter) {
		if statusFilter != "active" && statusFilter != "inactive" {
			logger.Error("Invalid status filter, must be active or inactive")
			return nil, errors.New("Invalid status filter, must be active or inactive")
		}
	}

	// ==============================================================
	// Get all enrollments of this user using index table
	// ==============================================================
	enrollments := []EnrollmentResult{}
	var iter asset_manager.AssetIteratorInterface
	var err error
	if !utils.IsStringEmpty(statusFilter) {
		iter, err = asset_mgmt.GetAssetManager(stub, caller).GetAssetIter(EnrollmentAssetNamespace, IndexEnrollment, []string{"user_id", "status"}, []string{userID, statusFilter}, []string{userID, statusFilter}, true, false, KeyPathFunc, "", -1, nil)
		if err != nil {
			logger.Errorf("GetServiceAssets failed: %v", err)
			return nil, errors.Wrap(err, "GetServiceAssets failed")
		}
	} else {
		iter, err = asset_mgmt.GetAssetManager(stub, caller).GetAssetIter(EnrollmentAssetNamespace, IndexEnrollment, []string{"user_id"}, []string{userID}, []string{userID}, true, false, KeyPathFunc, "", -1, nil)
		if err != nil {
			logger.Errorf("GetServiceAssets failed: %v", err)
			return nil, errors.Wrap(err, "GetServiceAssets failed")
		}
	}

	// Iterate over all enrollments
	defer iter.Close()
	for iter.HasNext() {
		enrollmentAsset, err := iter.Next()
		if err != nil {
			customErr := &custom_errors.IterError{}
			logger.Errorf("%v: %v", customErr, err)
			return nil, errors.Wrap(err, customErr.Error())
		}
		if !data_model.IsEncryptedData(enrollmentAsset.PrivateData) {
			enrollment := convertEnrollmentResultFromAsset(enrollmentAsset)
			enrollments = append(enrollments, enrollment)
		}
	}

	return json.Marshal(&enrollments)
}

// GetServiceEnrollments gets all enrollments of a service
// args = [ serviceID, statusFilter ]
// statusFilter is an optional parameter for filtering enrollment status
func GetServiceEnrollments(stub cached_stub.CachedStubInterface, caller data_model.User, args []string) ([]byte, error) {
	defer utils.ExitFnLog(utils.EnterFnLog())
	logger.Debugf("args: %v", args)

	if len(args) != 1 && len(args) != 2 {
		customErr := &custom_errors.LengthCheckingError{Type: "GetServiceEnrollments arguments length"}
		logger.Errorf(customErr.Error())
		return nil, errors.New(customErr.Error())
	}

	// ==============================================================
	// Validation
	// ==============================================================
	serviceID := args[0]
	if utils.IsStringEmpty(serviceID) {
		customErr := &custom_errors.LengthCheckingError{Type: "serviceID"}
		logger.Errorf(customErr.Error())
		return nil, errors.WithStack(customErr)
	}

	statusFilter := ""
	if len(args) == 2 {
		statusFilter = args[1]
	}

	// Validate status filter
	if !utils.IsStringEmpty(statusFilter) {
		if statusFilter != "active" && statusFilter != "inactive" {
			logger.Error("Invalid status filter, must be active or inactive")
			return nil, errors.New("Invalid status filter, must be active or inactive")
		}
	}

	// ==============================================================
	// Get all enrollments of this service using index table
	// ==============================================================
	enrollments := []EnrollmentResult{}
	var iter asset_manager.AssetIteratorInterface
	var err error
	if !utils.IsStringEmpty(statusFilter) {
		iter, err = asset_mgmt.GetAssetManager(stub, caller).GetAssetIter(EnrollmentAssetNamespace, IndexEnrollment, []string{"service_id", "status"}, []string{serviceID, statusFilter}, []string{serviceID, statusFilter}, true, false, KeyPathFunc, "", -1, nil)
		if err != nil {
			logger.Errorf("GetServiceAssets failed: %v", err)
			return nil, errors.Wrap(err, "GetServiceAssets failed")
		}
	} else {
		iter, err = asset_mgmt.GetAssetManager(stub, caller).GetAssetIter(EnrollmentAssetNamespace, IndexEnrollment, []string{"service_id"}, []string{serviceID}, []string{serviceID}, true, false, KeyPathFunc, "", -1, nil)
		if err != nil {
			logger.Errorf("GetServiceAssets failed: %v", err)
			return nil, errors.Wrap(err, "GetServiceAssets failed")
		}
	}

	// Iterate over all enrollments
	defer iter.Close()
	for iter.HasNext() {
		enrollmentAsset, err := iter.Next()
		if err != nil {
			customErr := &custom_errors.IterError{}
			logger.Errorf("%v: %v", customErr, err)
			return nil, errors.Wrap(err, customErr.Error())
		}
		if !data_model.IsEncryptedData(enrollmentAsset.PrivateData) {
			enrollment := convertEnrollmentResultFromAsset(enrollmentAsset)
			enrollments = append(enrollments, enrollment)
		}
	}

	return json.Marshal(&enrollments)
}

// GetEnrollmentInternal is the internal function for getting a single enrollment
func GetEnrollmentInternal(stub cached_stub.CachedStubInterface, caller data_model.User, userID string, serviceID string, options ...string) (Enrollment, error) {
	defer utils.ExitFnLog(utils.EnterFnLog())

	assetManager := asset_mgmt.GetAssetManager(stub, caller)

	enrollmentID := GetEnrollmentID(userID, serviceID)
	enrollmentAssetID := asset_mgmt.GetAssetId(EnrollmentAssetNamespace, enrollmentID)
	keyPath, err := GetKeyPath(stub, caller, enrollmentAssetID, options...)
	if err != nil {
		customErr := &GetKeyPathError{Caller: caller.ID, AssetID: enrollmentAssetID}
		logger.Errorf(customErr.Error())
		return Enrollment{}, errors.New(customErr.Error())
	}

	if len(keyPath) <= 0 {
		customErr := &GetKeyPathError{Caller: caller.ID, AssetID: enrollmentAssetID}
		logger.Errorf(customErr.Error())
		return Enrollment{}, errors.New(customErr.Error())
	}

	enrollmentKey, err := assetManager.GetAssetKey(enrollmentAssetID, keyPath)
	if err != nil {
		logger.Errorf("Failed to GetAssetKey for enrollmentKey: %v", err)
		return Enrollment{}, errors.Wrap(err, "Failed to GetAssetKey for enrollmentKey")
	}

	enrollmentAsset, err := assetManager.GetAsset(enrollmentAssetID, enrollmentKey)
	if err != nil {
		customErr := &custom_errors.GetAssetDataError{AssetId: enrollmentAssetID}
		logger.Errorf("%v: %v", customErr, err)
		return Enrollment{}, errors.Wrap(err, customErr.Error())
	}

	if utils.IsStringEmpty(enrollmentAsset.AssetId) {
		customErr := &custom_errors.GetAssetDataError{AssetId: enrollmentAssetID}
		logger.Errorf(customErr.Error())
		return Enrollment{}, errors.New(customErr.Error())
	}

	enrollment := convertEnrollmentFromAsset(enrollmentAsset)
	if utils.IsStringEmpty(enrollment.UserID) {
		customErr := &GetEnrollmentError{Enrollment: enrollmentID}
		logger.Errorf(customErr.Error())
		return Enrollment{}, errors.New(customErr.Error())
	}

	return enrollment, nil
}

func GetEnrollmentID(UserID string, ServiceID string) string {
	defer utils.ExitFnLog(utils.EnterFnLog())
	return EnrollmentPrefix + "-" + UserID + "-" + ServiceID
}

func convertEnrollmentToAsset(stub cached_stub.CachedStubInterface, enrollment Enrollment) (data_model.Asset, error) {
	defer utils.ExitFnLog(utils.EnterFnLog())

	enrollmentID := GetEnrollmentID(enrollment.UserID, enrollment.ServiceID)

	asset := data_model.Asset{}
	asset.AssetId = asset_mgmt.GetAssetId(EnrollmentAssetNamespace, enrollmentID)
	asset.Datatypes = []string{}
	metaData := make(map[string]string)
	metaData["namespace"] = EnrollmentAssetNamespace
	asset.Metadata = metaData
	enrollmentAssetPublicData, err := getEnrollmentAssetPublicData(enrollment)
	if err != nil {
		errMsg := "getEnrollmentAssetPublicData failed"
		logger.Errorf("%v: %v", errMsg, err)
		return data_model.Asset{}, errors.Wrap(err, errMsg)
	}
	asset.PublicData = enrollmentAssetPublicData
	enrollmentAssetPrivateData, err := getEnrollmentAssetPrivateData(enrollment)
	if err != nil {
		errMsg := "getEnrollmentAssetPrivateData failed"
		logger.Errorf("%v: %v", errMsg, err)
		return data_model.Asset{}, errors.Wrap(err, errMsg)
	}
	asset.PrivateData = enrollmentAssetPrivateData
	asset.OwnerIds = []string{enrollment.ServiceID}
	asset.IndexTableName = IndexEnrollment

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

// private function that converts asset to Enrollment
func convertEnrollmentFromAsset(asset *data_model.Asset) Enrollment {
	defer utils.ExitFnLog(utils.EnterFnLog())
	enrollment := Enrollment{}
	var publicData enrollmentPublicData
	var privateData enrollmentPrivateData
	json.Unmarshal(asset.PublicData, &publicData)
	json.Unmarshal(asset.PrivateData, &privateData)

	enrollment.EnrollmentID = publicData.EnrollmentID
	enrollment.UserID = publicData.UserID
	enrollment.UserName = publicData.UserName
	enrollment.ServiceID = publicData.ServiceID
	enrollment.ServiceName = publicData.ServiceName
	enrollment.EnrollDate = privateData.EnrollDate
	enrollment.Status = privateData.Status
	return enrollment
}

// private function that converts asset to EnrollmentResult
func convertEnrollmentResultFromAsset(asset *data_model.Asset) EnrollmentResult {
	defer utils.ExitFnLog(utils.EnterFnLog())
	enrollment := EnrollmentResult{}
	var publicData enrollmentPublicData
	var privateData enrollmentPrivateData
	json.Unmarshal(asset.PublicData, &publicData)
	json.Unmarshal(asset.PrivateData, &privateData)

	enrollment.UserID = publicData.UserID
	enrollment.UserName = publicData.UserName
	enrollment.ServiceID = publicData.ServiceID
	enrollment.ServiceName = publicData.ServiceName
	enrollment.EnrollDate = privateData.EnrollDate
	enrollment.Status = privateData.Status
	return enrollment
}

func getEnrollmentAssetPublicData(enrollment Enrollment) ([]byte, error) {
	defer utils.ExitFnLog(utils.EnterFnLog())

	publicData := enrollmentPublicData{}
	publicData.EnrollmentID = enrollment.EnrollmentID
	publicData.ServiceID = enrollment.ServiceID
	publicData.ServiceName = enrollment.ServiceName
	publicData.UserID = enrollment.UserID
	publicData.UserName = enrollment.UserName
	publicBytes, err := json.Marshal(&publicData)
	if err != nil {
		customErr := &custom_errors.MarshalError{Type: "publicData"}
		logger.Errorf("%v: %v", customErr, err)
		return nil, errors.Wrap(err, customErr.Error())
	}
	return publicBytes, nil
}

func getEnrollmentAssetPrivateData(enrollment Enrollment) ([]byte, error) {
	defer utils.ExitFnLog(utils.EnterFnLog())

	privateData := enrollmentPrivateData{}
	privateData.EnrollDate = enrollment.EnrollDate
	privateData.Status = enrollment.Status
	privateBytes, err := json.Marshal(&privateData)
	if err != nil {
		customErr := &custom_errors.MarshalError{Type: "privateData"}
		logger.Errorf("%v: %v", customErr, err)
		return nil, errors.Wrap(err, customErr.Error())
	}
	return privateBytes, nil
}

func CheckEnrollmentExists(stub cached_stub.CachedStubInterface, caller data_model.User, enrollmentID string) (bool, error) {
	enrollmentAssetID := asset_mgmt.GetAssetId(EnrollmentAssetNamespace, enrollmentID)
	assetManager := asset_mgmt.GetAssetManager(stub, caller)
	asset, err := assetManager.GetAsset(enrollmentAssetID, data_model.Key{})
	if err != nil {
		customErr := &custom_errors.GetAssetDataError{AssetId: enrollmentAssetID}
		logger.Errorf("%v: %v", customErr, err)
		return false, errors.Wrap(err, customErr.Error())
	}

	if !utils.IsStringEmpty(asset.AssetId) {
		return true, nil
	}

	return false, nil
}

func SetupEnrollmentIndex(stub cached_stub.CachedStubInterface) error {
	defer utils.ExitFnLog(utils.EnterFnLog())

	// Enrollment Indices
	enrollmentTable := index.GetTable(stub, IndexEnrollment, "enrollment_id")
	enrollmentTable.AddIndex([]string{"user_id", "status", "service_id", "enrollment_id"}, false)
	enrollmentTable.AddIndex([]string{"service_id", "status", "user_id", "enrollment_id"}, false)
	err := enrollmentTable.SaveToLedger()
	if err != nil {
		return err
	}

	return nil
}
