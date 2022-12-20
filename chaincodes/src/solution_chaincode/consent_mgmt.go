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
	"common/bchcls/key_mgmt"
	"encoding/json"
	"net/url"
	"strconv"
	"strings"
	"time"

	"common/bchcls/asset_mgmt"
	"common/bchcls/cached_stub"
	"common/bchcls/consent_mgmt"
	"common/bchcls/crypto"
	"common/bchcls/custom_errors"
	"common/bchcls/data_model"
	"common/bchcls/datatype"
	"common/bchcls/history"
	"common/bchcls/simple_rule"
	"common/bchcls/user_access_ctrl"
	"common/bchcls/user_mgmt"
	"common/bchcls/utils"

	"github.com/pkg/errors"
)

// There are 3 use cases
// 1) Patient (user) gives consent to a target service under a service/datatype pair
// 2) Service (owner) gives consent to a target service for the owner's datatype
// 3) Patient (owner) gives consent to a target service or user for the owner's datatype

const consentOptionWrite = "write"
const consentOptionRead = "read"
const consentOptionDeny = "deny"

// Consent object
// Target can be user or service
type Consent struct {
	Owner      string   `json:"owner"`
	Service    string   `json:"service"`
	Datatype   string   `json:"datatype"`
	Target     string   `json:"target"`
	Option     []string `json:"option"`
	Timestamp  int64    `json:"timestamp"`
	Expiration int64    `json:"expiration"`
}

// ConsentValidation object
type ConsentValidation struct {
	Owner             string           `json:"owner"`
	Target            string           `json:"target"`
	Datatype          string           `json:"datatype"`
	Requester         string           `json:"requester"`
	RequestedAccess   string           `json:"requested_access"`
	PermissionGranted bool             `json:"permission_granted"`
	Token             string           `json:"token"`
	Timestamp         int64            `json:"timestamp"`
	Message           string           `json:"message"`
	FilterRule        simple_rule.Rule `json:"filterRule"`
}

// ConsentValidationToken object
// Token should be encrypted with consentSymKey before getting attached to consent validation object
type ConsentValidationToken struct {
	Owner      string `json:"owner"`
	Target     string `json:"target"`
	Datatype   string `json:"datatype"`
	Access     string `json:"access"`
	Timestamp  int64  `json:"timestamp"`
	ConsentKey []byte `json:"consent_key"`
}

type ValidationResultWithLog struct {
	ConsentValidation ConsentValidation                   `json:"validation"`
	TransactionLog    data_model.ExportableTransactionLog `json:"transaction_log"`
}

// Consent requests is used for all present and future consents
type ConsentRequest struct {
	Owner       string            `json:"user"`
	Org         string            `json:"org"`
	Service     string            `json:"service"`
	ServiceName string            `json:"service_name"`
	Status      string            `json:"status"`
	EnrollDate  int64             `json:"enroll_date"`
	Datatypes   []ServiceDatatype `json:"datatypes"`
}

// consent log object
type ConsentLog struct {
	Owner    string      `json:"owner"`
	Target   string      `json:"target"`
	Datatype string      `json:"datatype"`
	Service  string      `json:"service"`
	Data     interface{} `json:"data"`
}

// PutConsentPatientData adds or updates consent, must be called by patient
// Consent can also be given to a reference service, a reference service uses a datatype from
// a different organization
//
// Steps:
// 1) Validate consentSymKeyB64 and consent object
// 2) Check patient enrollment status
// 3) Call consent package PutConsent function
// args = [ consentBytes, consentKeyB64 ]
//
// Note that write consent implies read
func PutConsentPatientData(stub cached_stub.CachedStubInterface, caller data_model.User, args []string) ([]byte, error) {
	defer utils.ExitFnLog(utils.EnterFnLog())
	logger.Debugf("args: %v", args)

	if len(args) != 1 && len(args) != 2 {
		customErr := &custom_errors.LengthCheckingError{Type: "PutConsentPatientData arguments length"}
		logger.Errorf(customErr.Error())
		return nil, errors.New(customErr.Error())
	}

	// ==============================================================
	// Validate incoming consent object and consent sym key
	// ==============================================================

	var consentOMR = Consent{}
	consentBytes := []byte(args[0])
	err := json.Unmarshal(consentBytes, &consentOMR)
	if err != nil {
		customErr := &custom_errors.UnmarshalError{Type: "consentOMR"}
		logger.Errorf("%v: %v", customErr, err)
		return nil, errors.Wrap(err, customErr.Error())
	}

	// validate consent owner
	if caller.ID != consentOMR.Owner {
		logger.Errorf("Caller can only give consent for himself. Caller: %v,  Consent Owner: %v", caller.ID, consentOMR.Owner)
		return nil, errors.New("Caller can only give consent for himself")
	}

	if consentOMR.Target != consentOMR.Service {
		if utils.InList(consentOMR.Option, consentOptionWrite) {
			logger.Errorf("Cannot give WRITE consent if target service is not datatype's service")
			return nil, errors.New("Cannot give WRITE consent if target service is not datatype's service")
		}
	}

	if consentOMR.Owner == consentOMR.Target {
		logger.Errorf("Owner and target cannot be the same")
		return nil, errors.New("Owner and target cannot be the same")
	}

	// Validate service
	service, err := GetServiceInternal(stub, caller, consentOMR.Service, false)
	if err != nil {
		customErr := &GetServiceError{Service: consentOMR.Service}
		logger.Errorf("%v: %v", customErr, err)
		return nil, errors.Wrap(err, customErr.Error())
	}

	if utils.IsStringEmpty(service.ServiceID) {
		customErr := &custom_errors.LengthCheckingError{Type: "service.ServiceID"}
		logger.Errorf(customErr.Error())
		return nil, errors.WithStack(customErr)
	}

	// check that consent datatype belongs to service
	valid := false
	for _, serviceDatatype := range service.Datatypes {
		if serviceDatatype.DatatypeID == consentOMR.Datatype {
			valid = true
			break
		}
	}

	if !valid {
		logger.Errorf("This service does not contain the specified datatype")
		return nil, errors.New("This service does not contain the specified datatype")
	}

	// validate consent option
	invalidOption := CheckConsentOptionIsInvalid(consentOMR.Option)
	if invalidOption {
		logger.Errorf("invalid consent option: %v", consentOMR.Option)
		return nil, errors.New("invalid consent option")
	}

	// Check consentDate is within 10 mins of current time
	currTime := time.Now().Unix()
	if currTime-consentOMR.Timestamp > 10*60 || currTime-consentOMR.Timestamp < -10*60 {
		logger.Errorf("Invalid Timestamp (current time: %v)  %v", currTime, consentOMR.Timestamp)
		return nil, errors.New("Invalid Timestamp, not within possible time range")
	}

	// Check expiration
	if consentOMR.Expiration > 0 && consentOMR.Expiration < consentOMR.Timestamp {
		logger.Errorf("Expiration date has passed")
		return nil, errors.New("Expiration date has passed")
	}

	// Get consent first
	isNewConsent := false
	existingConsent, err := GetConsentInternal(stub, caller, consentOMR.Target, consentOMR.Datatype, consentOMR.Owner)
	if utils.IsStringEmpty(existingConsent.Owner) {
		isNewConsent = true
	}

	// ==============================================================
	// Check patient enrollment
	// ==============================================================
	enrollmentID := GetEnrollmentID(consentOMR.Owner, consentOMR.Service)
	// Since caller can only be patient, we do not need to pass any options here
	enrollment, err := GetEnrollmentInternal(stub, caller, consentOMR.Owner, consentOMR.Service)
	if err != nil {
		customErr := &GetEnrollmentError{Enrollment: enrollmentID}
		logger.Errorf(customErr.Error())
		return nil, errors.New(customErr.Error())
	}

	//if patient is not currently enrolled, he can only deny consent
	if enrollment.Status != "active" {
		if !utils.InList(consentOMR.Option, consentOptionDeny) {
			logger.Errorf("Patient not currently enrolled, can only deny")
			return nil, errors.New("Patient not currently enrolled, can only deny")
		}
	}

	// ==============================================================
	// Call consent package PutConsent
	// ==============================================================
	consentCommon, err := convertToConsentCommon(stub, consentOMR)
	if err != nil {
		errMsg := "Failed to convertToConsentCommon"
		logger.Errorf("%v: %v", errMsg, err)
		return nil, errors.Wrap(err, errMsg)
	}

	consentCommonBytes, err := json.Marshal(&consentCommon)
	if err != nil {
		customErr := &custom_errors.MarshalError{Type: "Consent [Common]"}
		logger.Errorf("%v: %v", customErr, err)
		return nil, errors.Wrap(err, customErr.Error())
	}

	// ==============================================================
	// Logging
	// ==============================================================
	assetManager := asset_mgmt.GetAssetManager(stub, caller)
	enrollmentAssetID := asset_mgmt.GetAssetId(EnrollmentAssetNamespace, enrollment.EnrollmentID)
	// construct key path to get enrollment key for update
	keyPath, err := GetKeyPath(stub, caller, enrollmentAssetID)
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

	data := make(map[string]interface{})
	data["option"] = consentOMR.Option
	consentLog := ConsentLog{Owner: consentOMR.Owner, Target: consentOMR.Target, Datatype: consentOMR.Datatype, Service: consentOMR.Service, Data: data}
	solutionLog := SolutionLog{
		TransactionID: stub.GetTxID(),
		Namespace:     "OMR",
		FunctionName:  "PutConsentPatientData",
		CallerID:      caller.ID,
		Timestamp:     consentOMR.Timestamp,
		Data:          consentLog}
	err = AddLogWithParams(stub, caller, solutionLog, enrollmentLogSymKey)
	if err != nil {
		customErr := &AddSolutionLogError{FunctionName: solutionLog.FunctionName}
		logger.Errorf("%v: %v", customErr, err)
		return nil, errors.Wrap(err, customErr.Error())
	}

	// ================================================================================
	// If new consent
	// Add access from patient log sym key to enrollmentLogSymKey and consentLogSymKey
	// ================================================================================
	if isNewConsent {
		// Create datatype owner sym key
		_, err = datatype.AddDatatypeSymKey(stub, caller, consentOMR.Datatype, consentOMR.Owner)
		if err != nil {
			errMsg := "Failed to add datatype sym key in SDK "
			logger.Errorf("%v: %v", errMsg, err)
			return nil, errors.Wrap(err, errMsg)
		}

		// get consent owner sym key
		// since caller is consent owner, just get caller sym jey
		consentOwnerLogSymKey := caller.GetLogSymKey()
		if consentOwnerLogSymKey.KeyBytes == nil {
			logger.Errorf("Invalid consentOwnerLogSymKey")
			return nil, errors.New("Invalid consentOwnerLogSymKey")
		}

		// Edge from service log sym key to enrollment log sym key already added in EnrollPatient function
		userAccessManager := user_access_ctrl.GetUserAccessManager(stub, caller)
		if key_mgmt.KeyExists(stub, enrollmentLogSymKey.ID) {
			err = userAccessManager.AddAccessByKey(consentOwnerLogSymKey, enrollmentLogSymKey)
			if err != nil {
				customErr := &custom_errors.AddAccessError{Key: "consent key to consent log sym key"}
				logger.Errorf("%v: %v", customErr, err)
				return nil, errors.Wrap(err, customErr.Error())
			}
		}

		// Check consentKeyB64 if it is passed in
		consentKey := data_model.Key{ID: consent_mgmt.GetConsentID(consentOMR.Datatype, consentOMR.Target, consentOMR.Owner), Type: key_mgmt.KEY_TYPE_SYM}
		consentKeyB64 := ""
		if len(args) != 2 {
			customErr := &custom_errors.LengthCheckingError{Type: "Missing consent key"}
			logger.Errorf(customErr.Error())
			return nil, errors.New(customErr.Error())
		}

		consentKeyB64 = args[1]
		consentKeyBytes, err := crypto.ParseSymKeyB64(consentKeyB64)
		if err != nil {
			logger.Errorf("Invalid consentKeyB64")
			return nil, errors.Wrap(err, "Invalid consentKeyB64")
		}

		if consentKeyBytes == nil {
			logger.Errorf("Invalid consentKeyB64")
			return nil, errors.New("Invalid consentKeyB64")
		}

		consentKey.KeyBytes = consentKeyBytes

		// Add access from owner log sym key to consentLogSymKey
		// Needed for validate consent
		consentLogSymKey := GetLogSymKeyFromKey(consentKey)
		err = userAccessManager.AddAccessByKey(consentOwnerLogSymKey, consentLogSymKey)
		if err != nil {
			customErr := &custom_errors.AddAccessError{Key: "owner log sym key to consent log sym key"}
			logger.Errorf("%v: %v", customErr, err)
			return nil, errors.Wrap(err, customErr.Error())
		}

		return consent_mgmt.PutConsent(stub, caller, []string{string(consentCommonBytes), consentKeyB64})
	}

	return consent_mgmt.PutConsent(stub, caller, []string{string(consentCommonBytes)})
}

// PutConsentOwnerData adds or updates consent for owner data
// Consent can also be given to a reference service, a reference service uses a datatype from
// a different organization
//
// Steps:
// 1) Validate consentSymKeyB64 and consent object
// 2) Call consent package PutConsent function
// args = [ consentBytes, consentKeyB64 ]
//
// Note that write consent implies read
func PutConsentOwnerData(stub cached_stub.CachedStubInterface, caller data_model.User, args []string) ([]byte, error) {
	defer utils.ExitFnLog(utils.EnterFnLog())
	logger.Debugf("args: %v", args)

	if len(args) != 1 && len(args) != 2 {
		customErr := &custom_errors.LengthCheckingError{Type: "PutConsentOwnerData arguments length"}
		logger.Errorf(customErr.Error())
		return nil, errors.New(customErr.Error())
	}

	// ==============================================================
	// Validate incoming consent object and consent sym key
	// ==============================================================

	var consentOMR = Consent{}
	consentBytes := []byte(args[0])
	err := json.Unmarshal(consentBytes, &consentOMR)
	if err != nil {
		customErr := &custom_errors.UnmarshalError{Type: "consentOMR"}
		logger.Errorf("%v: %v", customErr, err)
		return nil, errors.Wrap(err, customErr.Error())
	}

	// Validate target
	if utils.IsStringEmpty(consentOMR.Target) {
		customErr := &custom_errors.LengthCheckingError{Type: "Target.ID"}
		logger.Errorf(customErr.Error())
		return nil, errors.WithStack(customErr)
	}

	// Validate owner
	if utils.IsStringEmpty(consentOMR.Owner) {
		customErr := &custom_errors.LengthCheckingError{Type: "consentOMR owner"}
		logger.Errorf(customErr.Error())
		return nil, errors.WithStack(customErr)
	}

	if consentOMR.Owner == consentOMR.Target {
		logger.Errorf("Owner and target cannot be the same")
		return nil, errors.New("Owner and target cannot be the same")
	}

	// Since owner could be a patient, service, or org, make sure caller has access to owner
	// also need owner sym key for logging
	callerObj := caller
	owner, err := user_mgmt.GetUserData(stub, callerObj, consentOMR.Owner, true, false)
	if err != nil {
		customErr := &GetUserError{User: consentOMR.Owner}
		logger.Errorf("%v: %v", customErr, err)
		return nil, errors.Wrap(err, customErr.Error())
	}

	if owner.PrivateKey == nil {
		logger.Errorf("Caller does not have access to owner")
		return nil, errors.New("Caller does not have access to owner")
	}

	// Validate datatype
	dtype, err := GetDatatypeWithParams(stub, caller, consentOMR.Datatype)
	if err != nil {
		logger.Errorf("Failed to GetDatatypeWithParams: %v, %v", consentOMR.Datatype, err)
		return nil, errors.Wrap(err, "Failed to GetDatatypeWithParams: "+consentOMR.Datatype)
	}

	if utils.IsStringEmpty(dtype.DatatypeID) {
		customErr := &custom_errors.LengthCheckingError{Type: "dtype.DatatypeID"}
		logger.Errorf(customErr.Error())
		return nil, errors.WithStack(customErr)
	}

	// boolean flag indicating whether the owner is a service, otherwise it must be a patient
	serviceOwner := true
	// try to get service on owner to check if it is a service
	service, err := GetServiceInternal(stub, callerObj, consentOMR.Owner, false)
	if err != nil || utils.IsStringEmpty(service.ServiceID) {
		serviceOwner = false
	}

	// If owner is a service, datatype must belong to service
	if serviceOwner {
		valid := false
		for _, serviceDatatype := range service.Datatypes {
			if serviceDatatype.DatatypeID == consentOMR.Datatype {
				valid = true
				break
			}
		}

		if !valid {
			logger.Errorf("This service does not contain the specified datatype")
			return nil, errors.New("This service does not contain the specified datatype")
		}

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
			// service ID is always owner ID for service owner consents
			// construct new caller object representing service itself
			serviceCaller, err := user_mgmt.GetUserData(stub, caller, consentOMR.Owner, true, false)
			if err != nil {
				customErr := &GetOrgError{Org: consentOMR.Owner}
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
	}

	invalidOption := CheckConsentOptionIsInvalid(consentOMR.Option)
	if invalidOption {
		logger.Errorf("invalid consent option: %v", consentOMR.Option)
		return nil, errors.New("invalid consent option")
	}

	// Check consentDate is within 10 mins of current time
	currTime := time.Now().Unix()
	if currTime-consentOMR.Timestamp > 10*60 || currTime-consentOMR.Timestamp < -10*60 {
		logger.Errorf("Invalid Timestamp (current time: %v)  %v", currTime, consentOMR.Timestamp)
		return nil, errors.New("Invalid Timestamp, not within possible time range")
	}

	// Check expiration
	if consentOMR.Expiration > 0 && consentOMR.Expiration < consentOMR.Timestamp {
		logger.Errorf("Expiration date has passed")
		return nil, errors.New("Expiration date has passed")
	}

	// Get consent first
	isNewConsent := false
	existingConsent, err := GetConsentInternal(stub, callerObj, consentOMR.Target, consentOMR.Datatype, consentOMR.Owner)
	if utils.IsStringEmpty(existingConsent.Owner) {
		isNewConsent = true
	}

	// ==============================================================
	// Convert to Common consent object and get consent key
	// ==============================================================
	consentOMR.Service = consentOMR.Owner
	consentCommon, err := convertToConsentCommon(stub, consentOMR)
	if err != nil {
		errMsg := "Failed to convertToConsentCommon"
		logger.Errorf("%v: %v", errMsg, err)
		return nil, errors.Wrap(err, errMsg)
	}

	consentCommonBytes, err := json.Marshal(&consentCommon)
	if err != nil {
		customErr := &custom_errors.MarshalError{Type: "Consent [Common]"}
		logger.Errorf("%v: %v", customErr, err)
		return nil, errors.Wrap(err, customErr.Error())
	}

	consentKey := data_model.Key{}
	consentKeyB64 := ""

	if isNewConsent {
		// Check consentKeyB64 if it is passed in
		consentKey = data_model.Key{ID: key_mgmt.GetSymKeyId(consent_mgmt.GetConsentID(consentOMR.Datatype, consentOMR.Target, consentOMR.Owner)), Type: key_mgmt.KEY_TYPE_SYM}
		if len(args) != 2 {
			customErr := &custom_errors.LengthCheckingError{Type: "Missing consent key"}
			logger.Errorf(customErr.Error())
			return nil, errors.New(customErr.Error())
		}

		consentKeyB64 = args[1]
		consentKey.KeyBytes, err = crypto.ParseSymKeyB64(consentKeyB64)
		if err != nil {
			logger.Errorf("Invalid consentKeyB64")
			return nil, errors.Wrap(err, "Invalid consentKeyB64")
		}

		if consentKey.KeyBytes == nil {
			logger.Errorf("Invalid consentKeyB64")
			return nil, errors.New("Invalid consentKeyB64")
		}
	} else {
		// Get consent key for logging
		consentKey, err = GetConsentKeyInternal(stub, callerObj, consentOMR.Target, consentOMR.Datatype, consentOMR.Owner)
		if err != nil {
			logger.Errorf("Failed getting consent key")
			return nil, errors.Wrap(err, "Failed getting consent key")
		}
	}

	// ==============================================================
	// Logging
	// ==============================================================
	consentLogSymKey := GetLogSymKeyFromKey(consentKey)

	data := make(map[string]interface{})
	data["option"] = consentOMR.Option
	consentLog := ConsentLog{Owner: consentOMR.Owner, Target: consentOMR.Target, Datatype: consentOMR.Datatype, Service: consentOMR.Service, Data: data}
	solutionLog := SolutionLog{
		TransactionID: stub.GetTxID(),
		Namespace:     "OMR",
		FunctionName:  "PutConsentOwnerData",
		CallerID:      caller.ID,
		Timestamp:     consentOMR.Timestamp,
		Data:          consentLog}

	err = AddLogWithParams(stub, callerObj, solutionLog, consentLogSymKey)
	if err != nil {
		customErr := &AddSolutionLogError{FunctionName: solutionLog.FunctionName}
		logger.Errorf("%v: %v", customErr, err)
		return nil, errors.Wrap(err, customErr.Error())
	}

	if isNewConsent {
		// Create datatype owner sym key
		_, err = datatype.AddDatatypeSymKey(stub, caller, consentOMR.Datatype, consentOMR.Owner)
		if err != nil {
			errMsg := "Failed to add datatype sym key in SDK "
			logger.Errorf("%v: %v", errMsg, err)
			return nil, errors.Wrap(err, errMsg)
		}

		ownerLogSymKey := owner.GetLogSymKey()
		userAccessManager := user_access_ctrl.GetUserAccessManager(stub, callerObj)
		err = userAccessManager.AddAccessByKey(ownerLogSymKey, consentLogSymKey)
		if err != nil {
			customErr := &custom_errors.AddAccessError{Key: "owner log sym key to consent log sym key"}
			logger.Errorf("%v: %v", customErr, err)
			return nil, errors.Wrap(err, customErr.Error())
		}
		return consent_mgmt.PutConsent(stub, callerObj, []string{string(consentCommonBytes), consentKeyB64})
	}

	return consent_mgmt.PutConsent(stub, callerObj, []string{string(consentCommonBytes)})
}

// GetConsent returns a single consent object for an owner/target/datatype pair
// If consent does not exist, return error
// args = [ ownerID, targetID, datatypeID ]
func GetConsent(stub cached_stub.CachedStubInterface, caller data_model.User, args []string) ([]byte, error) {
	defer utils.ExitFnLog(utils.EnterFnLog())
	logger.Debugf("args: %v", args)

	if len(args) != 3 {
		customErr := &custom_errors.LengthCheckingError{Type: "GetConsent arguments length"}
		logger.Errorf(customErr.Error())
		return nil, errors.New(customErr.Error())
	}

	ownerID := args[0]
	targetID := args[1]
	datatypeID := args[2]

	consentOMR, err := GetConsentInternal(stub, caller, targetID, datatypeID, ownerID)
	if err != nil {
		customErr := &GetConsentError{Consent: "Consent for " + targetID + ", " + datatypeID}
		logger.Errorf("%v: %v", customErr, err)
		return nil, errors.Wrap(err, customErr.Error())
	}

	if utils.IsStringEmpty(consentOMR.Owner) {
		customErr := &GetConsentError{Consent: "Empty consent for " + targetID + ", " + datatypeID}
		logger.Errorf("%v", customErr)
		return nil, customErr
	}

	return json.Marshal(&consentOMR)
}

// ValidateConsent validates access for an owner/target/datatype pair based on consent
// Can be called by consent owner, target or anyone with access to owner or target
// Currently having write consent also implies read consent
// args = [ ownerID, targetID, datatypeID, access, timestamp]
func ValidateConsent(stub cached_stub.CachedStubInterface, caller data_model.User, args []string) ([]byte, error) {
	defer utils.ExitFnLog(utils.EnterFnLog())
	logger.Debugf("args: %v", args)

	if len(args) != 5 {
		customErr := &custom_errors.LengthCheckingError{Type: "ValidateConsent arguments length"}
		logger.Errorf(customErr.Error())
		return nil, errors.New(customErr.Error())
	}

	ownerID := args[0]
	targetID := args[1]
	datatypeID := args[2]
	access := args[3]

	if access != consentOptionWrite && access != consentOptionRead && access != consentOptionDeny {
		logger.Errorf("Access must be write, read, or deny")
		return nil, errors.New("Access must be write, read, or deny")
	}

	timestamp, err := strconv.ParseInt(args[4], 10, 64)
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

	callerObj, err := GetServiceCaller(stub, caller, targetID)
	if err != nil {
		getUserErr := &GetUserError{User: targetID}
		logger.Errorf("Failed to get Target Service as a caller: %v", err)
		return nil, errors.Wrap(err, getUserErr.Error())
	}

	// ==============================================================
	// Construct consent validation object
	// ==============================================================
	validation := ConsentValidation{}
	validation.Datatype = datatypeID
	validation.Owner = ownerID
	validation.Target = targetID
	validation.Requester = caller.ID
	validation.PermissionGranted = false
	validation.RequestedAccess = access
	validation.Token = ""
	validation.Message = ""
	validation.Timestamp = timestamp

	accessGranted := true
	filterRule, consentKey, err := consent_mgmt.ValidateConsent(stub, callerObj, []string{datatypeID, ownerID, targetID, strings.ToUpper(access), args[4]})
	// If error type is validate consent error, then keep going, this means consent option did not match, we still want to return validation object
	if err != nil {
		accessGranted = false
	}

	if consentKey.KeyBytes == nil {
		accessGranted = false
	}

	validation.FilterRule = filterRule

	// ==============================================================
	// Construct token
	// ==============================================================
	if accessGranted {
		//encrypt token with caller's public key
		token := ConsentValidationToken{}
		token.Access = validation.RequestedAccess
		token.Timestamp = validation.Timestamp
		token.Datatype = validation.Datatype
		token.Target = validation.Target
		token.Owner = validation.Owner
		// have to return only keyBytes
		// otherwise it will fail to encrypt token bytes with pub key; message would be too long for RSA public key size
		token.ConsentKey = consentKey.KeyBytes
		tokenBytes, err := json.Marshal(&token)
		if err != nil {
			customErr := &custom_errors.MarshalError{Type: "Token"}
			logger.Errorf("%v: %v", customErr, err)
			return nil, errors.Wrap(err, customErr.Error())
		}

		encTokenBytes, err := crypto.EncryptWithPublicKey(caller.PublicKey, tokenBytes)
		if err != nil {
			customErr := &custom_errors.EncryptionError{ToEncrypt: "token bytes", EncryptionKey: "caller pub key"}
			logger.Errorf("%v: %v", customErr, err)
			return nil, errors.Wrap(err, customErr.Error())
		}

		validation.Token = url.QueryEscape(crypto.EncodeToB64String(encTokenBytes))
		validation.PermissionGranted = true
		validation.Message = "permission granted"
	} else {
		validation.PermissionGranted = false
		validation.Message = "permission denied"
	}

	// ==============================================================
	// Logging
	// ==============================================================
	consentLogSymKey := GetLogSymKeyFromKey(consentKey)

	data := make(map[string]interface{})
	data["access"] = access
	validateConsentLog := ConsentLog{Owner: ownerID, Target: targetID, Datatype: datatypeID, Data: data}
	solutionLog := SolutionLog{
		TransactionID: stub.GetTxID(),
		Namespace:     "OMR",
		FunctionName:  "ValidateConsent",
		CallerID:      caller.ID,
		Timestamp:     timestamp,
		Data:          validateConsentLog,
	}
	exportableLog, err := GenerateExportableSolutionLog(stub, callerObj, solutionLog, consentLogSymKey)
	if err != nil {
		customErr := &GenerateExportableTransactionLogError{Function: solutionLog.FunctionName}
		logger.Errorf("%v: %v", customErr, err)
		return nil, errors.Wrap(err, customErr.Error())
	}

	returnData := ValidationResultWithLog{}
	returnData.ConsentValidation = validation
	returnData.TransactionLog = exportableLog

	return json.Marshal(&returnData)
}

// AddValidateConsentQueryLog is a helper function that adds query log for validate consent function
// query log will not be added if caller does not have access to target and consent key
func AddValidateConsentQueryLog(stub cached_stub.CachedStubInterface, caller data_model.User, args []string) error {
	defer utils.ExitFnLog(utils.EnterFnLog())
	logger.Debugf("args: %v", args)

	if len(args) != 2 {
		customErr := &custom_errors.LengthCheckingError{Type: "AddValidateConsentQueryLog arguments length"}
		logger.Errorf(customErr.Error())
		return customErr
	}

	// parse consent validation object
	cv := ConsentValidation{}
	cvBytes := []byte(args[0])
	err := json.Unmarshal(cvBytes, &cv)
	if err != nil {
		customErr := &custom_errors.UnmarshalError{Type: "consent validation"}
		logger.Errorf("%v: %v", customErr.Error(), err)
		return errors.Wrap(err, customErr.Error())
	}

	callerObj := caller
	// if caller is org admin, use org as caller
	solutionCaller := convertToSolutionUser(caller)

	callerObj, err = GetServiceCaller(stub, caller, cv.Target)
	if err != nil {
		getUserErr := &GetUserError{User: cv.Target}
		logger.Errorf("Failed to get Target Service as a caller: %v", err)
		return errors.Wrap(err, getUserErr.Error())
	}

	// Check if caller has access to target, only add access if he does
	// Access from owner log sym key to consent log sym key is added in PutConsent functions
	target, err := user_mgmt.GetUserData(stub, callerObj, cv.Target, true, false)
	if err != nil {
		customErr := &GetUserError{User: cv.Target}
		logger.Errorf("%v: %v", customErr, err)
		return errors.Wrap(err, customErr.Error())
	}

	if target.PrivateKey != nil {
		targetLogSymKey := target.GetLogSymKey()
		consentKey, err := GetConsentKeyInternal(stub, callerObj, cv.Target, cv.Datatype, cv.Owner, solutionCaller.Org, cv.Target)
		if err != nil {
			logger.Errorf("Failed getting consent key")
			return errors.Wrap(err, "Failed getting consent key")
		}

		consentLogSymKey := GetLogSymKeyFromKey(consentKey)
		userAccessManager := user_access_ctrl.GetUserAccessManager(stub, callerObj)
		err = userAccessManager.AddAccessByKey(targetLogSymKey, consentLogSymKey)
		if err != nil {
			customErr := &custom_errors.AddAccessError{Key: "target log sym key to consent log sym key"}
			logger.Errorf("%v: %v", customErr, err)
			return errors.Wrap(err, customErr.Error())
		}

		return history.PutQueryTransactionLog(stub, callerObj, []string{args[1]})
	}
	return nil
}

// GetConsentInternal returns consent object, to be used internally
// If consent does not exist in Common, return empty consent object
func GetConsentInternal(stub cached_stub.CachedStubInterface, caller data_model.User, targetID string, datatypeID string, ownerID string) (Consent, error) {
	defer utils.ExitFnLog(utils.EnterFnLog())

	if utils.IsStringEmpty(targetID) {
		customErr := &custom_errors.LengthCheckingError{Type: "targetID"}
		logger.Errorf(customErr.Error())
		return Consent{}, customErr
	}

	if utils.IsStringEmpty(datatypeID) {
		customErr := &custom_errors.LengthCheckingError{Type: "datatypeID"}
		logger.Errorf(customErr.Error())
		return Consent{}, customErr
	}

	if utils.IsStringEmpty(ownerID) {
		customErr := &custom_errors.LengthCheckingError{Type: "ownerID"}
		logger.Errorf(customErr.Error())
		return Consent{}, customErr
	}

	consentCommonBytes, err := consent_mgmt.GetConsent(stub, caller, []string{datatypeID, targetID, ownerID})
	if err != nil {
		customErr := &GetConsentError{Consent: "Consent for " + targetID + ", " + datatypeID}
		logger.Errorf("%v: %v", customErr, err)
		return Consent{}, errors.Wrap(err, customErr.Error())
	}

	if consentCommonBytes == nil {
		customErr := &GetConsentError{Consent: "Consent for " + targetID + ", " + datatypeID}
		logger.Errorf(customErr.Error())
		return Consent{}, customErr
	}

	consentCommon := data_model.Consent{}
	err = json.Unmarshal(consentCommonBytes, &consentCommon)
	if err != nil {
		customErr := &custom_errors.UnmarshalError{Type: "consentCommon"}
		logger.Errorf("%v: %v", customErr, err)
		return Consent{}, errors.Wrap(err, customErr.Error())
	}

	consentOMR := convertFromConsentCommon(consentCommon)

	return consentOMR, nil
}

// GetConsentKeyInternal returns consent key object
// Option must be in the order of: orgId, consentTargetId
// If consent does not exist in Common, return empty key object
func GetConsentKeyInternal(stub cached_stub.CachedStubInterface, caller data_model.User, targetID string, datatypeID string, ownerID string, options ...string) (data_model.Key, error) {
	defer utils.ExitFnLog(utils.EnterFnLog())

	if utils.IsStringEmpty(targetID) {
		customErr := &custom_errors.LengthCheckingError{Type: "targetID"}
		logger.Errorf(customErr.Error())
		return data_model.Key{}, errors.WithStack(customErr)
	}

	if utils.IsStringEmpty(datatypeID) {
		customErr := &custom_errors.LengthCheckingError{Type: "datatypeID"}
		logger.Errorf(customErr.Error())
		return data_model.Key{}, errors.WithStack(customErr)
	}

	if utils.IsStringEmpty(ownerID) {
		customErr := &custom_errors.LengthCheckingError{Type: "ownerID"}
		logger.Errorf(customErr.Error())
		return data_model.Key{}, errors.WithStack(customErr)
	}

	consentCommonBytes, err := consent_mgmt.GetConsent(stub, caller, []string{datatypeID, targetID, ownerID})
	if err != nil {
		customErr := &GetConsentError{Consent: "Consent for " + targetID + ", " + datatypeID}
		logger.Errorf("%v: %v", customErr, err)
		return data_model.Key{}, errors.Wrap(err, customErr.Error())
	}

	if consentCommonBytes == nil {
		customErr := &GetConsentError{Consent: "Consent for " + targetID + ", " + datatypeID}
		logger.Errorf(customErr.Error())
		return data_model.Key{}, errors.WithStack(customErr)
	}

	consentCommon := data_model.Consent{}
	err = json.Unmarshal(consentCommonBytes, &consentCommon)
	if err != nil {
		customErr := &custom_errors.UnmarshalError{Type: "consentCommon"}
		logger.Errorf("%v: %v", customErr, err)
		return data_model.Key{}, errors.Wrap(err, customErr.Error())
	}

	assetManager := asset_mgmt.GetAssetManager(stub, caller)
	consentID := consent_mgmt.GetConsentID(datatypeID, targetID, ownerID)
	consentAssetID, err := consent_mgmt.GetConsentAssetID(stub, consentID)
	if err != nil {
		logger.Errorf("Failed to get consentAssetID: %v", err)
		return data_model.Key{}, errors.Wrap(err, "Failed to get consentAssetID")
	}

	// TODO:
	// if caller is org admin, pass org ID and consent target ID
	consentKeyPath, err := GetKeyPath(stub, caller, consentAssetID, options...)
	if err != nil {
		customErr := &GetKeyPathError{Caller: caller.ID, AssetID: consentAssetID}
		logger.Errorf(customErr.Error())
		return data_model.Key{}, errors.New(customErr.Error())
	}

	if len(consentKeyPath) <= 0 {
		customErr := &GetKeyPathError{Caller: caller.ID, AssetID: consentAssetID}
		logger.Errorf(customErr.Error())
		return data_model.Key{}, errors.New(customErr.Error())
	}

	consentKey, err := assetManager.GetAssetKey(consentAssetID, consentKeyPath)
	if err != nil {
		logger.Errorf("Failed to get consentKey: %v", err)
		return data_model.Key{}, errors.Wrap(err, "Failed to get consentKey")
	}

	return consentKey, nil
}

// GetConsents returns all consents for a service user (target owner) pair
// args = [ serviceID, userID ]
func GetConsents(stub cached_stub.CachedStubInterface, caller data_model.User, args []string) ([]byte, error) {
	defer utils.ExitFnLog(utils.EnterFnLog())
	logger.Debugf("args: %v", args)

	if len(args) != 2 {
		customErr := &custom_errors.LengthCheckingError{Type: "GetConsents arguments length"}
		logger.Errorf(customErr.Error())
		return nil, errors.New(customErr.Error())
	}

	serviceID := args[0]
	userID := args[1]

	consentOMRs, err := GetConsentsInternal(stub, caller, userID, serviceID)
	if err != nil {
		errMsg := "Failed to get consents"
		logger.Errorf(errMsg)
		return nil, errors.Wrap(err, errMsg)
	}

	return json.Marshal(consentOMRs)
}

// GetConsentsInternal is an internal helper function for returning all consents for a service user (target owner) pair
func GetConsentsInternal(stub cached_stub.CachedStubInterface, caller data_model.User, ownerID string, targetID string) ([]Consent, error) {
	defer utils.ExitFnLog(utils.EnterFnLog())

	if utils.IsStringEmpty(ownerID) {
		customErr := &custom_errors.LengthCheckingError{Type: "ownerID"}
		logger.Errorf(customErr.Error())
		return nil, errors.WithStack(customErr)
	}

	if utils.IsStringEmpty(targetID) {
		customErr := &custom_errors.LengthCheckingError{Type: "targetID"}
		logger.Errorf(customErr.Error())
		return nil, errors.WithStack(customErr)
	}

	consentCommons := []data_model.Consent{}
	consentOMRs := []Consent{}

	consentCommonsBytes, err := consent_mgmt.GetConsentsWithTargetIDAndOwnerID(stub, caller, []string{targetID, ownerID})
	if err != nil {
		errMsg := "Failed to get consents"
		logger.Errorf(errMsg)
		return nil, errors.Wrap(err, errMsg)
	}

	err = json.Unmarshal(consentCommonsBytes, &consentCommons)
	if err != nil {
		customErr := &custom_errors.UnmarshalError{Type: "consentCommons"}
		logger.Errorf("%v: %v", customErr, err)
		return nil, errors.Wrap(err, customErr.Error())
	}

	for _, consentCommon := range consentCommons {
		consentOMR := convertFromConsentCommon(consentCommon)
		if utils.IsStringEmpty(consentOMR.Owner) {
			customErr := &custom_errors.LengthCheckingError{Type: "consentOMR.Owner"}
			logger.Debugf(customErr.Error())
			continue
		}
		consentOMRs = append(consentOMRs, consentOMR)
	}

	return consentOMRs, nil
}

// Get consents with owner ID
func GetConsentsWithOwnerID(stub cached_stub.CachedStubInterface, caller data_model.User, args []string) ([]byte, error) {
	defer utils.ExitFnLog(utils.EnterFnLog())
	logger.Debugf("args: %v", args)

	consentCommons := []data_model.Consent{}
	consentOMRs := []Consent{}

	consentCommonsBytes, err := consent_mgmt.GetConsentsWithOwnerID(stub, caller, args)
	if err != nil {
		errMsg := "Failed to get consents"
		logger.Errorf(errMsg)
		return nil, errors.Wrap(err, errMsg)
	}
	err = json.Unmarshal(consentCommonsBytes, &consentCommons)
	if err != nil {
		customErr := &custom_errors.UnmarshalError{Type: "consentCommons"}
		logger.Errorf("%v: %v", customErr, err)
		return nil, errors.Wrap(err, customErr.Error())
	}

	for _, consentCommon := range consentCommons {
		consentOMR := convertFromConsentCommon(consentCommon)
		if utils.IsStringEmpty(consentOMR.Owner) {
			customErr := &custom_errors.LengthCheckingError{Type: "consentOMR.Owner"}
			logger.Debugf(customErr.Error())
			continue
		}
		consentOMRs = append(consentOMRs, consentOMR)
	}

	return json.Marshal(consentOMRs)
}

// Get consents with target ID
func GetConsentsWithTargetID(stub cached_stub.CachedStubInterface, caller data_model.User, args []string) ([]byte, error) {
	defer utils.ExitFnLog(utils.EnterFnLog())
	logger.Debugf("args: %v", args)

	consentCommons := []data_model.Consent{}
	consentOMRs := []Consent{}

	consentCommonsBytes, err := consent_mgmt.GetConsentsWithTargetID(stub, caller, args)
	if err != nil {
		errMsg := "Failed to get consents"
		logger.Errorf(errMsg)
		return nil, errors.Wrap(err, errMsg)
	}
	err = json.Unmarshal(consentCommonsBytes, &consentCommons)
	if err != nil {
		customErr := &custom_errors.UnmarshalError{Type: "consentCommons"}
		logger.Errorf("%v: %v", customErr, err)
		return nil, errors.Wrap(err, customErr.Error())
	}

	for _, consentCommon := range consentCommons {
		consentOMR := convertFromConsentCommon(consentCommon)
		if utils.IsStringEmpty(consentOMR.Owner) {
			customErr := &custom_errors.LengthCheckingError{Type: "consentOMR.Owner"}
			logger.Debugf(customErr.Error())
			continue
		}
		consentOMRs = append(consentOMRs, consentOMR)
	}

	return json.Marshal(consentOMRs)
}

func DecryptConsentValidationToken(stub cached_stub.CachedStubInterface, caller data_model.User, encTokenBytesB64 string) (ConsentValidationToken, error) {
	defer utils.ExitFnLog(utils.EnterFnLog())
	logger.Debugf("tokenB64: %v", encTokenBytesB64)

	// NOTE: We don't need to call url.QueryUnescape() function because Node.js automatically
	// replace escape characters with original values when we pass token in URL.
	encTokenBytes, err := crypto.DecodeStringB64(encTokenBytesB64)
	if err != nil {
		logger.Errorf("Invalid validation token (B64): decoding error: %v", err)
		return ConsentValidationToken{}, errors.Wrap(err, "Invalid validation token (B64): decoding error")
	}

	tokenBytes, err := crypto.DecryptWithPrivateKey(caller.PrivateKey, encTokenBytes)
	if err != nil {
		customErr := &custom_errors.DecryptionError{ToDecrypt: "encTokenBytes", DecryptionKey: "caller private key"}
		logger.Errorf("%v: %v", customErr, err)
		return ConsentValidationToken{}, errors.Wrap(err, customErr.Error())
	}

	if tokenBytes == nil {
		customErr := &custom_errors.DecryptionError{ToDecrypt: "encTokenBytes", DecryptionKey: "caller private key"}
		logger.Errorf("%v", customErr)
		return ConsentValidationToken{}, customErr
	}

	token := ConsentValidationToken{}
	err = json.Unmarshal(tokenBytes, &token)
	if err != nil {
		customErr := &custom_errors.UnmarshalError{Type: "tokenBytes"}
		logger.Errorf("%v: %v", customErr, err)
		return ConsentValidationToken{}, errors.Wrap(err, customErr.Error())
	}

	if token.Timestamp <= 0 {
		logger.Error("Invalid validation token: no timestamp")
		return ConsentValidationToken{}, errors.New("Invalid validation token: no timestamp")
	}

	currTime := time.Now().Unix()
	if currTime-token.Timestamp > 15*60 {
		logger.Errorf("Expired token, past 15 mins allowance (current time: %v)", currTime)
		return ConsentValidationToken{}, errors.New("Expired token, past 15 mins allowance")
	}

	if currTime < token.Timestamp {
		logger.Errorf("Invalid validation token: invalid future timestamp %v", token.Timestamp)
		return ConsentValidationToken{}, errors.New("Invalid validation token: invalid future timestamp")
	}

	return token, nil
}

// GetAllConsentRequests gets all requests (current and pending) for a patient
// args = [patient ID, serviceID]
// If serviceID is not empty, then we only get consent requests for one enrollment
// If serviceID is empty, we get consent requests for all enrollments of the patient
// Steps
// check patient enrollment, get service ID from enrollment
//  for each service ID, get datatypes under the service
//  for each service datatype pair, check consent given
func GetAllConsentRequests(stub cached_stub.CachedStubInterface, caller data_model.User, args []string) ([]byte, error) {
	defer utils.ExitFnLog(utils.EnterFnLog())
	logger.Debugf("args: %v", args)

	if len(args) != 2 {
		customErr := &custom_errors.LengthCheckingError{Type: "GetAllConsentRequests arguments length"}
		logger.Errorf(customErr.Error())
		return nil, errors.New(customErr.Error())
	}

	patientID := args[0]
	if utils.IsStringEmpty(patientID) {
		customErr := &custom_errors.LengthCheckingError{Type: "patientID"}
		logger.Errorf(customErr.Error())
		return nil, errors.WithStack(customErr)
	}

	enrollments := []Enrollment{}
	enrollment := Enrollment{}

	serviceID := args[1]
	if utils.IsStringEmpty(serviceID) {
		// get all enrollments for this patient
		enrollmentsBytes, err := GetPatientEnrollments(stub, caller, []string{patientID})
		err = json.Unmarshal(enrollmentsBytes, &enrollments)
		if err != nil {
			customErr := &custom_errors.UnmarshalError{Type: "enrollments"}
			logger.Errorf("%v: %v", customErr, err)
			return nil, errors.Wrap(err, customErr.Error())
		}
	} else {
		// if caller is org admin pass org ID here
		solutionCaller := convertToSolutionUser(caller)
		var err error
		if solutionCaller.SolutionInfo.IsOrgAdmin {
			enrollment, err = GetEnrollmentInternal(stub, caller, patientID, serviceID, solutionCaller.Org)
			if err != nil {
				customErr := &GetEnrollmentError{Enrollment: GetEnrollmentID(patientID, serviceID)}
				logger.Errorf(customErr.Error())
				return nil, errors.New(customErr.Error())
			}
		} else {
			enrollment, err = GetEnrollmentInternal(stub, caller, patientID, serviceID)
			if err != nil {
				customErr := &GetEnrollmentError{Enrollment: GetEnrollmentID(patientID, serviceID)}
				logger.Errorf(customErr.Error())
				return nil, errors.New(customErr.Error())
			}
		}

		enrollments = append(enrollments, enrollment)
	}

	consentRequests := []ConsentRequest{}

	// get consent for all service datatype pair
	for _, enrollment := range enrollments {
		serviceDatatypes := []ServiceDatatype{}

		service, err := GetServiceInternal(stub, caller, enrollment.ServiceID, false)
		if err != nil {
			customErr := &GetServiceError{Service: enrollment.ServiceID}
			logger.Errorf("%v: %v", customErr, err)
			return nil, errors.Wrap(err, customErr.Error())
		}

		// check for consent to service datatypes
		for _, datatype := range service.Datatypes {
			consent, err := GetConsentInternal(stub, caller, service.ServiceID, datatype.DatatypeID, patientID)

			if err != nil {
				// TODO:
				// Check for error type, so if consent does not exist, append no consent option
				datatype.Access = []string{}
			} else {
				datatype.Access = consent.Option
			}

			//if enrollment is inactive and consent is denied, skip
			if enrollment.Status == "inactive" && utils.InList(consent.Option, consentOptionDeny) {
				continue
			} else {
				serviceDatatypes = append(serviceDatatypes, datatype)
			}
		}

		//check for consents for datatypes that are not defined in the service (reference datatypes)
		consents, err := GetConsentsInternal(stub, caller, patientID, service.ServiceID)
		if err != nil {
			logger.Errorf("Failed to get consents: %v", err)
		}

		for _, consent := range consents {
			if !service.hasDatatype(consent.Datatype) {
				serviceDatatype := ServiceDatatype{}
				serviceDatatype.DatatypeID = consent.Datatype
				serviceDatatype.ServiceID = consent.Service
				serviceDatatype.Access = consent.Option
				serviceDatatypes = append(serviceDatatypes, serviceDatatype)
			}
		}

		consentRequest := ConsentRequest{}
		consentRequest.Owner = patientID
		consentRequest.Org = service.OrgID
		consentRequest.Service = service.ServiceID
		consentRequest.ServiceName = service.ServiceName
		consentRequest.Status = enrollment.Status
		consentRequest.EnrollDate = enrollment.EnrollDate
		consentRequest.Datatypes = serviceDatatypes

		consentRequests = append(consentRequests, consentRequest)
	}

	return json.Marshal(consentRequests)
}

func convertToConsentCommon(stub cached_stub.CachedStubInterface, consentOMR Consent) (data_model.Consent, error) {
	defer utils.ExitFnLog(utils.EnterFnLog())

	consentCommon := data_model.Consent{}
	consentCommon.TargetID = consentOMR.Target
	consentCommon.OwnerID = consentOMR.Owner
	consentCommon.DatatypeID = consentOMR.Datatype

	if utils.InList(consentOMR.Option, consentOptionDeny) {
		consentCommon.Access = consent_mgmt.ACCESS_DENY
	} else if utils.InList(consentOMR.Option, consentOptionWrite) {
		consentCommon.Access = consent_mgmt.ACCESS_WRITE
	} else {
		consentCommon.Access = consent_mgmt.ACCESS_READ
	}

	consentCommon.ConsentDate = consentOMR.Timestamp
	consentCommon.ExpirationDate = consentOMR.Expiration
	data := make(map[string]interface{})
	data["consent"] = consentOMR.Option
	data["service"] = consentOMR.Service
	consentCommon.Data = data

	// get off-chain datastore connection id, if one is setup
	dsConnectionID, err := GetActiveConnectionID(stub)
	if err != nil {
		errMsg := "Failed to GetActiveConnectionID"
		logger.Errorf("%v: %v", errMsg, err)
		return data_model.Consent{}, errors.Wrap(err, errMsg)
	}
	consentCommon.ConnectionID = dsConnectionID

	return consentCommon, nil
}

func convertFromConsentCommon(consentCommon data_model.Consent) Consent {
	defer utils.ExitFnLog(utils.EnterFnLog())

	consentOMR := Consent{}
	consentOMR.Owner = consentCommon.OwnerID
	consentOMR.Target = consentCommon.TargetID
	consentOMR.Datatype = consentCommon.DatatypeID
	consentOMR.Timestamp = consentCommon.ConsentDate
	consentOMR.Expiration = consentCommon.ExpirationDate
	if consentCommon.Data != nil {
		if consent, ok := consentCommon.Data.(map[string]interface{})["consent"].([]interface{}); ok {
			consentOMR.Option = GetStringSliceFromInterface(consent)
		}

		if service, ok := consentCommon.Data.(map[string]interface{})["service"].(string); ok {
			consentOMR.Service = service
		}
	}

	return consentOMR
}
