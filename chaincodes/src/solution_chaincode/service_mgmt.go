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
	"encoding/json"
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
	"common/bchcls/user_mgmt/user_groups"
	"common/bchcls/utils"

	"github.com/pkg/errors"
)

const IndexService = "ServiceTable"
const ServiceAssetNamespace = "ServiceAsset"

// De-identified fields:
//   - ServiceID
//   - ServiceName
//   - OrgID
type Service struct {
	ServiceID           string            `json:"service_id"`
	ServiceName         string            `json:"service_name"`
	Datatypes           []ServiceDatatype `json:"datatypes"`
	OrgID               string            `json:"org_id"`
	Email               string            `json:"email"`
	Summary             string            `json:"summary"`
	Terms               interface{}       `json:"terms"`
	PaymentRequired     string            `json:"payment_required"`
	Status              string            `json:"status"`
	SolutionPrivateData interface{}       `json:"solution_private_data"`
	CreateDate          int64             `json:"create_date"`
	UpdateDate          int64             `json:"update_date"`
}

// Private structs used for conversion between service and asset
type servicePrivateData struct {
	Email               string      `json:"email"`
	SolutionPrivateData interface{} `json:"solution_private_data"`
}

type ServicePublicData struct {
	ServiceID       string            `json:"service_id"`
	ServiceName     string            `json:"service_name"`
	Datatypes       []ServiceDatatype `json:"datatypes"`
	OrgID           string            `json:"org_id"`
	Summary         string            `json:"summary"`
	Terms           interface{}       `json:"terms"`
	PaymentRequired string            `json:"payment_required"`
	Status          string            `json:"status"`
	CreateDate      int64             `json:"create_date"`
	UpdateDate      int64             `json:"update_date"`
}

// datatypes attached to a service
// TODO: merge this with datatype struct in datatype_mgmt
type ServiceDatatype struct {
	DatatypeID string   `json:"datatype_id"`
	ServiceID  string   `json:"service_id"`
	Access     []string `json:"access"`
}

// RegisterService
// Can only be called by org admins of the org the service belongs to
// Internally, the function does the following:
// 1) register service as subgroup in user_mgmt
// 2) map old CM’s service struct to Asset, AssetSymKey = ServiceSymKey
// args = [ serviceBytes ]
func RegisterService(stub cached_stub.CachedStubInterface, caller data_model.User, args []string) ([]byte, error) {
	defer utils.ExitFnLog(utils.EnterFnLog())
	logger.Debugf("args: %v", args)

	if len(args) != 1 {
		customErr := &custom_errors.LengthCheckingError{Type: "RegisterService arguments length"}
		logger.Errorf(customErr.Error())
		return nil, errors.New(customErr.Error())
	}

	// ==============================================================
	// Validate incoming service object
	// ==============================================================
	service := Service{}
	serviceBytes := []byte(args[0])
	err := json.Unmarshal(serviceBytes, &service)
	if err != nil {
		customErr := &custom_errors.UnmarshalError{Type: "Service"}
		logger.Errorf("%v: %v", customErr, err)
		return nil, errors.Wrap(err, customErr.Error())
	}

	if utils.IsStringEmpty(service.ServiceID) {
		customErr := &custom_errors.LengthCheckingError{Type: "ServiceID"}
		logger.Errorf(customErr.Error())
		return nil, errors.WithStack(customErr)
	}

	callerObj := caller
	// if caller is org admin, use org as caller
	solutionCaller := convertToSolutionUser(caller)
	if solutionCaller.SolutionInfo.IsOrgAdmin {
		// Make sure caller's org matches service orgID
		if solutionCaller.Org != service.OrgID {
			logger.Error("Caller does not belong in the same org as service org")
			return nil, errors.New("Caller does not belong in the same org as service org")
		}

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
		// if caller is not org admin, cannot register service
		errMsg := "Caller must be org admin"
		logger.Errorf(errMsg)
		return nil, errors.New(errMsg)
	}

	// Check for existing service ID, return error if ID already exists
	// Since service is stored as asset, get asset
	assetManager := asset_mgmt.GetAssetManager(stub, callerObj)
	existingService, err := assetManager.GetAsset(asset_mgmt.GetAssetId(ServiceAssetNamespace, service.ServiceID), data_model.Key{})
	if err != nil {
		customErr := &custom_errors.GetAssetDataError{AssetId: "ServiceID"}
		logger.Errorf("%v: %v", customErr, err)
		return nil, errors.Wrap(err, customErr.Error())
	}

	if !utils.IsStringEmpty(existingService.AssetId) {
		logger.Errorf("Failed to RegisterService because this id already exists")
		return nil, errors.New("Failed to RegisterService because this id already exists")
	}

	// Validate service name
	if utils.IsStringEmpty(service.ServiceName) {
		customErr := &custom_errors.LengthCheckingError{Type: "ServiceName"}
		logger.Errorf(customErr.Error())
		return nil, errors.WithStack(customErr)
	}

	// Validate OrgID
	if utils.IsStringEmpty(service.OrgID) {
		customErr := &custom_errors.LengthCheckingError{Type: "service.OrgID"}
		logger.Errorf(customErr.Error())
		return nil, errors.WithStack(customErr)
	}

	// Validate datatypes
	if len(service.Datatypes) == 0 {
		logger.Error("Must have at least one datatype")
		return nil, errors.New("Must have at least one datatype")
	}

	// Get datatypes to make sure they exist
	for i, dtype := range service.Datatypes {
		logger.Debugf("dtype: %v", dtype)

		service.Datatypes[i].ServiceID = service.ServiceID

		valid, err := ValidateDatatype(stub, callerObj, dtype)
		if err != nil {
			customErr := &ValidateDatatypeError{Datatype: dtype.DatatypeID}
			logger.Errorf("%v: %v", customErr, err)
			return nil, errors.Wrap(err, customErr.Error())
		}

		if !valid {
			customErr := &ValidateDatatypeError{Datatype: dtype.DatatypeID}
			logger.Errorf("%v: %v", customErr, err)
			return nil, errors.Wrap(err, customErr.Error())
		}
	}

	// Validate service summary
	if utils.IsStringEmpty(service.Summary) {
		customErr := &custom_errors.LengthCheckingError{Type: "service.Summary"}
		logger.Errorf(customErr.Error())
		return nil, errors.WithStack(customErr)
	}

	// Validate payment required
	if service.PaymentRequired != "yes" && service.PaymentRequired != "no" {
		logger.Error("Payment required must be yes or no")
		return nil, errors.New("Payment required must be yes or no")
	}

	// Validate service status, must be active
	if service.Status != "active" {
		logger.Error("Invalid status, must be active")
		return nil, errors.New("Invalid status, must be active")
	}

	// check that createDate is within 10 mins of current time
	currTime := time.Now().Unix()
	if currTime-service.CreateDate > 10*60 || currTime-service.CreateDate < -10*60 {
		logger.Errorf("Invalid create date (current time: %v)  %v", currTime, service.CreateDate)
		return nil, errors.New("Invalid create date, not within possible time range")
	}

	// Set service update date
	service.UpdateDate = service.CreateDate

	// ==============================================================
	// Call user mgmt to register service as subgroup of org
	// ==============================================================

	serviceSubgroup, err := convertToSubgroup(args)
	if err != nil {
		var errMsg = "Failed converting service to subgroup"
		logger.Errorf("%v: %v", errMsg, err)
		return nil, errors.Wrap(err, errMsg)
	}

	if utils.IsStringEmpty(serviceSubgroup.ID) {
		customErr := &custom_errors.LengthCheckingError{Type: "serviceSubgroup.ID"}
		logger.Errorf(customErr.Error())
		return nil, errors.WithStack(customErr)
	}

	subgroupBytes, err := json.Marshal(serviceSubgroup)
	if err != nil {
		customErr := &custom_errors.MarshalError{Type: "Service subgroup"}
		logger.Errorf("%v: %v", customErr, err)
		return nil, errors.Wrap(err, customErr.Error())
	}

	subgroupArgs := []string{string(subgroupBytes), service.OrgID}
	_, err = user_groups.RegisterSubgroup(stub, caller, subgroupArgs)
	if err != nil {
		customErr := &RegisterOrgError{Org: service.ServiceID}
		logger.Errorf("%v: %v", customErr, err)
		return nil, errors.Wrap(err, customErr.Error())
	}

	// Convert to Asset
	serviceAsset, err := convertServiceToAsset(stub, service, callerObj.ID)
	if err != nil {
		customErr := &ConvertToAssetError{Asset: "serviceAsset"}
		logger.Errorf("%v: %v", customErr, err)
		return nil, errors.Wrap(err, customErr.Error())
	}
	serviceAsset.AssetKeyId = serviceSubgroup.GetSymKeyId()

	serviceAssetKey := data_model.Key{}
	serviceAssetKey.ID = serviceSubgroup.GetSymKeyId()
	serviceAssetKey.Type = key_mgmt.KEY_TYPE_SYM
	serviceAssetKey.KeyBytes = serviceSubgroup.SymKey

	am := asset_mgmt.GetAssetManager(stub, callerObj)
	err = am.AddAsset(serviceAsset, serviceAssetKey, false)
	if err != nil {
		customErr := &PutAssetError{Asset: service.ServiceID}
		logger.Errorf("%v: %v", customErr, err)
		return nil, errors.Wrap(err, customErr.Error())
	}

	// ==============================================================
	// Create datatype symkey between service and all datatypes
	// ==============================================================
	for _, dtype := range service.Datatypes {
		_, err = datatype.AddDatatypeSymKey(stub, callerObj, dtype.DatatypeID, service.ServiceID)
		if err != nil {
			customErr := &custom_errors.AddAccessError{Key: "service pub key to datatype key"}
			logger.Errorf("%v: %v", customErr, err)
			return nil, errors.Wrap(err, customErr.Error())
		}
	}

	return nil, nil
}

// UpdateService
// 1) update service in user_mgmt
// 2) update service asset
// args = [ serviceBytes ]
func UpdateService(stub cached_stub.CachedStubInterface, caller data_model.User, args []string) ([]byte, error) {
	defer utils.ExitFnLog(utils.EnterFnLog())
	logger.Debugf("args: %v", args)

	if len(args) != 1 {
		customErr := &custom_errors.LengthCheckingError{Type: "UpdateService arguments length"}
		logger.Errorf(customErr.Error())
		return nil, errors.New(customErr.Error())
	}

	// ==============================================================
	// Validate incoming service object
	// ==============================================================

	service := Service{}
	serviceBytes := []byte(args[0])
	err := json.Unmarshal(serviceBytes, &service)
	if err != nil {
		customErr := &custom_errors.UnmarshalError{Type: "Service"}
		logger.Errorf("%v: %v", customErr, err)
		return nil, errors.Wrap(err, customErr.Error())
	}

	if utils.IsStringEmpty(service.ServiceID) {
		customErr := &custom_errors.LengthCheckingError{Type: "ServiceID"}
		logger.Errorf(customErr.Error())
		return nil, errors.WithStack(customErr)
	}

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
		serviceCaller, err := user_mgmt.GetUserData(stub, caller, service.ServiceID, true, false)
		if err != nil {
			customErr := &GetOrgError{Org: service.ServiceID}
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

	// Check for existing service ID, return error if not found
	existingService, err := GetServiceInternal(stub, callerObj, service.ServiceID, true)
	if err != nil {
		customErr := &GetServiceError{Service: service.ServiceID}
		logger.Errorf("%v: %v", customErr, err)
		return nil, errors.Wrap(err, customErr.Error())
	}

	if utils.IsStringEmpty(existingService.ServiceID) {
		customErr := &custom_errors.LengthCheckingError{Type: "existingService.ServiceID"}
		logger.Errorf(customErr.Error())
		return nil, errors.WithStack(customErr)
	}

	// Validate service name
	if utils.IsStringEmpty(service.ServiceName) {
		customErr := &custom_errors.LengthCheckingError{Type: "ServiceName"}
		logger.Errorf(customErr.Error())
		return nil, errors.WithStack(customErr)
	}

	if service.OrgID != existingService.OrgID {
		logger.Error("Service org cannot be changed")
		return nil, errors.New("Service org cannot be changed")
	}

	// Validate datatypes
	if len(service.Datatypes) == 0 {
		logger.Error("Must have at least one datatype")
		return nil, errors.New("Must have at least one datatype")
	}

	// Get datatypes to make sure they exist
	for i, dtype := range service.Datatypes {
		logger.Debugf("dtype: %v", dtype)

		service.Datatypes[i].ServiceID = service.ServiceID

		valid, err := ValidateDatatype(stub, callerObj, dtype)
		if err != nil {
			customErr := &ValidateDatatypeError{Datatype: dtype.DatatypeID}
			logger.Errorf("%v: %v", customErr, err)
			return nil, errors.Wrap(err, customErr.Error())
		}

		if !valid {
			customErr := &ValidateDatatypeError{Datatype: dtype.DatatypeID}
			logger.Errorf("%v: %v", customErr, err)
			return nil, errors.Wrap(err, customErr.Error())
		}

		// create datatype symkey between datatype and service
		_, err = datatype.AddDatatypeSymKey(stub, callerObj, dtype.DatatypeID, service.ServiceID)
		if err != nil {
			customErr := &custom_errors.AddAccessError{Key: "service pub key to datatype key"}
			logger.Errorf("%v: %v", customErr, err)
			return nil, errors.Wrap(err, customErr.Error())
		}
	}

	// Validate service summary
	if utils.IsStringEmpty(service.Summary) {
		customErr := &custom_errors.LengthCheckingError{Type: "service.Summary"}
		logger.Errorf(customErr.Error())
		return nil, errors.WithStack(customErr)
	}

	// Validate payment required
	if service.PaymentRequired != "yes" && service.PaymentRequired != "no" {
		logger.Error("Payment required must be yes or no")
		return nil, errors.New("Payment required must be yes or no")
	}

	// Validate payment required
	if service.Status != "active" && service.Status != "inactive" {
		logger.Error("Payment status must be active or inactive")
		return nil, errors.New("Payment status must be active or inactive")
	}

	// check that updateDate is within 10 mins of current time
	currTime := time.Now().Unix()
	if currTime-service.UpdateDate > 10*60 || currTime-service.UpdateDate < -10*60 {
		logger.Errorf("Invalid create date (current time: %v)  %v", currTime, service.UpdateDate)
		return nil, errors.New("Invalid create date, not within possible time range")
	}

	// Update service fields
	existingService.UpdateDate = service.UpdateDate
	existingService.PaymentRequired = service.PaymentRequired
	existingService.Datatypes = service.Datatypes
	existingService.SolutionPrivateData = service.SolutionPrivateData
	existingService.Terms = service.Terms
	existingService.Summary = service.Summary
	existingService.ServiceName = service.ServiceName
	existingService.Status = service.Status

	// ==============================================================
	// Call user mgmt to update subgroup
	// ==============================================================

	serviceSubgroup, err := convertToSubgroup(args)
	if err != nil {
		var errMsg = "Failed converting service to subgroup"
		logger.Errorf("%v: %v", errMsg, err)
		return nil, errors.Wrap(err, errMsg)
	}

	if utils.IsStringEmpty(serviceSubgroup.ID) {
		customErr := &custom_errors.LengthCheckingError{Type: "serviceSubgroup.ID"}
		logger.Errorf(customErr.Error())
		return nil, errors.WithStack(customErr)
	}

	subgroupBytes, err := json.Marshal(serviceSubgroup)
	if err != nil {
		customErr := &custom_errors.MarshalError{Type: "Service subgroup"}
		logger.Errorf("%v: %v", customErr, err)
		return nil, errors.Wrap(err, customErr.Error())
	}

	args = []string{string(subgroupBytes), "false"}
	_, err = user_mgmt.UpdateOrg(stub, callerObj, args)
	if err != nil {
		customErr := &RegisterOrgError{Org: service.ServiceID}
		logger.Errorf("%v: %v", customErr, err)
		return nil, errors.Wrap(err, customErr.Error())
	}

	// Convert to Asset
	serviceAsset, err := convertServiceToAsset(stub, existingService, callerObj.ID)
	if err != nil {
		customErr := &ConvertToAssetError{Asset: "serviceAsset"}
		logger.Errorf("%v: %v", customErr, err)
		return nil, errors.Wrap(err, customErr.Error())
	}
	serviceAsset.AssetKeyId = serviceSubgroup.GetSymKeyId()

	assetManager := asset_mgmt.GetAssetManager(stub, callerObj)
	serviceAssetKey, err := assetManager.GetAssetKey(serviceAsset.AssetId, GetKeyPathFromCallerToServiceAsset(stub, callerObj, serviceAsset.AssetKeyId))
	if err != nil {
		logger.Errorf("Failed to GetAssetKey for serviceAssetKey: %v", err)
		return nil, errors.Wrap(err, "Failed to GetAssetKey for serviceAssetKey")
	}

	err = assetManager.UpdateAsset(serviceAsset, serviceAssetKey, true)
	if err != nil {
		customErr := &PutAssetError{Asset: service.ServiceID}
		logger.Errorf("%v: %v", customErr, err)
		return nil, errors.Wrap(err, customErr.Error())
	}

	return nil, nil
}

// Get service
// args = [ serviceID ]
func GetService(stub cached_stub.CachedStubInterface, caller data_model.User, args []string) ([]byte, error) {
	defer utils.ExitFnLog(utils.EnterFnLog())
	logger.Debugf("args: %v", args)

	if len(args) != 1 {
		customErr := &custom_errors.LengthCheckingError{Type: "GetService arguments length"}
		logger.Errorf(customErr.Error())
		return nil, errors.New(customErr.Error())
	}

	serviceID := args[0]
	if utils.IsStringEmpty(serviceID) {
		customErr := &custom_errors.LengthCheckingError{Type: "ServiceID"}
		logger.Errorf(customErr.Error())
		return nil, errors.WithStack(customErr)
	}

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
		serviceCaller, err := user_mgmt.GetUserData(stub, caller, serviceID, true, false)
		if err != nil {
			customErr := &GetOrgError{Org: serviceID}
			logger.Errorf("%v: %v", customErr, err)
			return nil, errors.Wrap(err, customErr.Error())
		}

		if serviceCaller.PrivateKey == nil {
			logger.Errorf("Caller does not have access to service private key")
		} else {
			callerObj = serviceCaller
		}
	}

	assetManager := asset_mgmt.GetAssetManager(stub, callerObj)
	serviceAssetID := asset_mgmt.GetAssetId(ServiceAssetNamespace, serviceID)
	serviceAssetKeyID, err := asset_mgmt.GetAssetKeyId(stub, serviceAssetID)
	if err != nil {
		logger.Errorf("Failed to get serviceAssetKeyID: %v", err)
		return nil, errors.Wrap(err, "Failed to get serviceAssetKeyID")
	}

	serviceAssetKeyPath := GetKeyPathFromCallerToServiceAsset(stub, callerObj, serviceAssetKeyID)

	if len(serviceAssetKeyPath) <= 0 {
		customErr := &GetKeyPathError{Caller: callerObj.ID, AssetID: serviceAssetKeyID}
		logger.Errorf(customErr.Error())
		return nil, errors.New(customErr.Error())
	}

	serviceAssetKey := data_model.Key{}
	// Ignore any errors here because we still return service public data if failed to get service key
	serviceAssetKey, _ = assetManager.GetAssetKey(serviceAssetID, serviceAssetKeyPath)
	serviceAsset, err := assetManager.GetAsset(serviceAssetID, serviceAssetKey)
	if err != nil {
		customErr := &custom_errors.GetAssetDataError{AssetId: serviceAssetID}
		logger.Errorf("%v: %v", customErr, err)
		return nil, errors.Wrap(err, customErr.Error())
	}

	if utils.IsStringEmpty(serviceAsset.AssetId) {
		logger.Errorf("Failed to get service, no existing service asset found")
		return nil, errors.New("Failed to get service, no existing service asset found")
	}

	service := convertServiceFromAsset(serviceAsset)
	if utils.IsStringEmpty(service.ServiceName) {
		customErr := &custom_errors.LengthCheckingError{Type: "Service converted"}
		logger.Errorf(customErr.Error())
		return nil, errors.WithStack(customErr)
	}

	return json.Marshal(service)
}

// GetServiceInternal returns a service
// It will not return an error if getPrivateData flag is set to true but caller does not have access
// It will only return public data if caller does not have access to service private data
func GetServiceInternal(stub cached_stub.CachedStubInterface, caller data_model.User, serviceID string, getPrivateData bool) (Service, error) {
	defer utils.ExitFnLog(utils.EnterFnLog())

	if utils.IsStringEmpty(serviceID) {
		customErr := &custom_errors.LengthCheckingError{Type: "ServiceID"}
		logger.Errorf(customErr.Error())
		return Service{}, errors.WithStack(customErr)
	}

	serviceAssetKey := data_model.Key{}
	assetManager := asset_mgmt.GetAssetManager(stub, caller)
	serviceAssetID := asset_mgmt.GetAssetId(ServiceAssetNamespace, serviceID)
	if getPrivateData {
		callerObj := caller
		serviceAssetKeyID, err := asset_mgmt.GetAssetKeyId(stub, serviceAssetID)
		if err != nil {
			logger.Errorf("Failed to get serviceAssetKeyID: %v", err)
			return Service{}, errors.Wrap(err, "Failed to get serviceAssetKeyID")
		}

		// if caller is org admin, use org as caller
		solutionCaller := convertToSolutionUser(caller)
		if solutionCaller.SolutionInfo.IsOrgAdmin {
			orgCaller, err := user_mgmt.GetUserData(stub, caller, solutionCaller.Org, true, false)
			if err != nil {
				customErr := &GetOrgError{Org: solutionCaller.Org}
				logger.Errorf("%v: %v", customErr, err)
				return Service{}, errors.Wrap(err, customErr.Error())
			}

			if orgCaller.PrivateKey == nil {
				errMsg := "Caller does not have access to org private key"
				logger.Errorf(errMsg)
				return Service{}, errors.New(errMsg)
			} else {
				callerObj = orgCaller
			}
		} else {
			// if caller is service admin, use service as caller
			// construct new caller object representing service itself
			serviceCaller, err := user_mgmt.GetUserData(stub, caller, serviceID, true, false)
			if err != nil {
				customErr := &GetOrgError{Org: serviceID}
				logger.Errorf("%v: %v", customErr, err)
				return Service{}, errors.Wrap(err, customErr.Error())
			}

			if serviceCaller.PrivateKey == nil {
				logger.Errorf("Caller does not have access to service private key")
			} else {
				callerObj = serviceCaller
			}
		}

		assetManager = asset_mgmt.GetAssetManager(stub, callerObj)
		serviceAssetKeyPath := GetKeyPathFromCallerToServiceAsset(stub, callerObj, serviceAssetKeyID)
		// GetAssetKey will not return an error if getPrivateData flag is set to true but caller does not have access
		serviceAssetKey, _ = assetManager.GetAssetKey(serviceAssetID, serviceAssetKeyPath)
	}

	serviceAsset, err := assetManager.GetAsset(serviceAssetID, serviceAssetKey)
	if err != nil {
		customErr := &custom_errors.GetAssetDataError{AssetId: serviceAssetID}
		logger.Errorf("%v: %v", customErr, err)
		return Service{}, errors.Wrap(err, customErr.Error())
	}

	if utils.IsStringEmpty(serviceAsset.AssetId) {
		customErr := &GetServiceError{Service: serviceID}
		logger.Errorf(customErr.Error())
		return Service{}, customErr
	}

	service := convertServiceFromAsset(serviceAsset)
	if utils.IsStringEmpty(service.ServiceName) {
		customErr := &custom_errors.LengthCheckingError{Type: "Service converted"}
		logger.Errorf(customErr.Error())
		return Service{}, errors.WithStack(customErr)
	}

	return service, nil
}

// AddDatatypeToService adds a datatype to existing service
// Note that within 1 chaincode transaction, this function can only be called once
// args = [ serviceID, datatype ]
func AddDatatypeToService(stub cached_stub.CachedStubInterface, caller data_model.User, args []string) ([]byte, error) {
	defer utils.ExitFnLog(utils.EnterFnLog())
	logger.Debugf("args: %v", args)

	if len(args) != 2 {
		customErr := &custom_errors.LengthCheckingError{Type: "AddDatatypeToService arguments length"}
		logger.Errorf(customErr.Error())
		return nil, errors.New(customErr.Error())
	}

	// ==============================================================
	// Validate serviceID and incoming datatype object
	// ==============================================================
	serviceID := args[0]
	if utils.IsStringEmpty(serviceID) {
		customErr := &custom_errors.LengthCheckingError{Type: "ServiceID"}
		logger.Errorf(customErr.Error())
		return nil, errors.WithStack(customErr)
	}

	callerObj := data_model.User{}
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
		}
		callerObj = orgCaller
	} else {
		// if caller is service admin, use service as caller
		// construct new caller object representing service itself
		serviceCaller, err := user_mgmt.GetUserData(stub, caller, serviceID, true, false)
		if err != nil {
			customErr := &GetOrgError{Org: serviceID}
			logger.Errorf("%v: %v", customErr, err)
			return nil, errors.Wrap(err, customErr.Error())
		}

		if serviceCaller.PrivateKey == nil {
			logger.Errorf("Caller does not have access to service private key")
			return nil, errors.New("Caller does not have access to service private key")
		}

		callerObj = serviceCaller
	}

	// Get service to make sure it exists
	existingService, err := GetServiceInternal(stub, callerObj, serviceID, true)
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

	// Validate datatype
	serviceDatatype := ServiceDatatype{}
	serviceDatatypeBytes := []byte(args[1])
	err = json.Unmarshal(serviceDatatypeBytes, &serviceDatatype)
	if err != nil {
		customErr := &custom_errors.UnmarshalError{Type: "ServiceDatatype"}
		logger.Errorf("%v: %v", customErr, err)
		return nil, errors.Wrap(err, customErr.Error())
	}

	serviceDatatype.ServiceID = existingService.ServiceID

	valid, err := ValidateDatatype(stub, callerObj, serviceDatatype)
	if err != nil {
		customErr := &ValidateDatatypeError{Datatype: serviceDatatype.DatatypeID}
		logger.Errorf("%v: %v", customErr, err)
		return nil, errors.Wrap(err, customErr.Error())
	}

	if !valid {
		customErr := &ValidateDatatypeError{Datatype: serviceDatatype.DatatypeID}
		logger.Errorf(customErr.Error())
		return nil, errors.WithStack(customErr)
	}

	// Check if existing service already contains this datatype
	if existingService.hasDatatype(serviceDatatype.DatatypeID) {
		logger.Errorf("Failed to add datatype, service already contains this datatype: %v", serviceDatatype.DatatypeID)
		return nil, errors.New("Failed to add datatype, service " + serviceID + " already contains the datatype " + serviceDatatype.DatatypeID)
	}

	// ==============================================================
	// Add datatype to service object
	// ==============================================================
	existingService.Datatypes = append(existingService.Datatypes, serviceDatatype)

	assetManager := asset_mgmt.GetAssetManager(stub, callerObj)
	serviceAsset, err := convertServiceToAsset(stub, existingService, callerObj.ID)
	if err != nil {
		customErr := &ConvertToAssetError{Asset: "serviceAsset"}
		logger.Errorf("%v: %v", customErr, err)
		return nil, errors.Wrap(err, customErr.Error())
	}
	serviceAssetKeyID, err := asset_mgmt.GetAssetKeyId(stub, serviceAsset.AssetId)
	if err != nil {
		logger.Errorf("Failed to get serviceAssetKeyID: %v", err)
		return nil, errors.Wrap(err, "Failed to get serviceAssetKeyID")
	}
	serviceAssetKey, err := assetManager.GetAssetKey(serviceAsset.AssetId, GetKeyPathFromCallerToServiceAsset(stub, callerObj, serviceAssetKeyID))
	if err != nil {
		logger.Errorf("Failed to GetAssetKey for serviceAssetKey: %v", err)
		return nil, errors.Wrap(err, "Failed to GetAssetKey for serviceAssetKey")
	}
	err = assetManager.UpdateAsset(serviceAsset, serviceAssetKey, true)
	if err != nil {
		customErr := &PutAssetError{Asset: serviceID}
		logger.Errorf("%v: %v", customErr, err)
		return nil, errors.Wrap(err, customErr.Error())
	}

	// Create datatype symkey between service and datatype
	_, err = datatype.AddDatatypeSymKey(stub, callerObj, serviceDatatype.DatatypeID, existingService.ServiceID)
	if err != nil {
		customErr := &custom_errors.AddAccessError{Key: "service pub key to datatype key"}
		logger.Errorf("%v: %v", customErr, err)
		return nil, errors.Wrap(err, customErr.Error())
	}

	return nil, nil
}

// RemoveDatatypeFromService removes a datatype from existing service
// args = [ serviceID, datatypeID ]
func RemoveDatatypeFromService(stub cached_stub.CachedStubInterface, caller data_model.User, args []string) ([]byte, error) {
	defer utils.ExitFnLog(utils.EnterFnLog())
	logger.Debugf("args: %v", args)

	if len(args) != 2 {
		customErr := &custom_errors.LengthCheckingError{Type: "RemoveDatatypeFromService arguments length"}
		logger.Errorf(customErr.Error())
		return nil, errors.New(customErr.Error())
	}

	// ==============================================================
	// Validate serviceID and incoming datatype object
	// ==============================================================
	serviceID := args[0]
	if utils.IsStringEmpty(serviceID) {
		customErr := &custom_errors.LengthCheckingError{Type: "ServiceID"}
		logger.Errorf(customErr.Error())
		return nil, errors.WithStack(customErr)
	}

	callerObj := data_model.User{}
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
		}
		callerObj = orgCaller
	} else {
		// if caller is service admin, use service as caller
		// construct new caller object representing service itself
		serviceCaller, err := user_mgmt.GetUserData(stub, caller, serviceID, true, false)
		if err != nil {
			customErr := &GetOrgError{Org: serviceID}
			logger.Errorf("%v: %v", customErr, err)
			return nil, errors.Wrap(err, customErr.Error())
		}

		if serviceCaller.PrivateKey == nil {
			logger.Errorf("Caller does not have access to service private key")
			return nil, errors.New("Caller does not have access to service private key")
		}

		callerObj = serviceCaller
	}

	// Get service to make sure it exists
	existingService, err := GetServiceInternal(stub, callerObj, serviceID, true)
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

	// Check if existing service actually contains this datatype
	datatypeID := args[1]
	if !existingService.hasDatatype(datatypeID) {
		logger.Errorf("Failed to remove datatype, service does not contains this datatype: %v", datatypeID)
		return nil, errors.New("Failed to remove datatype, service " + serviceID + " does not contains the datatype " + datatypeID)
	}

	// ==============================================================
	// Remove datatype from service object
	// ==============================================================
	existingService.Datatypes = RemoveDatatypeFromList(existingService.Datatypes, datatypeID)

	assetManager := asset_mgmt.GetAssetManager(stub, callerObj)
	serviceAsset, err := convertServiceToAsset(stub, existingService, callerObj.ID)
	if err != nil {
		customErr := &ConvertToAssetError{Asset: "serviceAsset"}
		logger.Errorf("%v: %v", customErr, err)
		return nil, errors.Wrap(err, customErr.Error())
	}

	serviceAssetKeyID, err := asset_mgmt.GetAssetKeyId(stub, serviceAsset.AssetId)
	if err != nil {
		logger.Errorf("Failed to get serviceAssetKeyID: %v", err)
		return nil, errors.Wrap(err, "Failed to get serviceAssetKeyID")
	}

	serviceAssetKey, err := assetManager.GetAssetKey(serviceAsset.AssetId, GetKeyPathFromCallerToServiceAsset(stub, callerObj, serviceAssetKeyID))
	if err != nil {
		logger.Errorf("Failed to GetAssetKey for serviceAssetKey: %v", err)
		return nil, errors.Wrap(err, "Failed to GetAssetKey for serviceAssetKey")
	}
	err = assetManager.UpdateAsset(serviceAsset, serviceAssetKey, true)
	if err != nil {
		customErr := &PutAssetError{Asset: serviceID}
		logger.Errorf("%v: %v", customErr, err)
		return nil, errors.Wrap(err, customErr.Error())
	}

	return nil, nil
}

// GetServicesOfOrg gets all services of an organization
// args = [ orgID ]
func GetServicesOfOrg(stub cached_stub.CachedStubInterface, caller data_model.User, args []string) ([]byte, error) {
	defer utils.ExitFnLog(utils.EnterFnLog())
	logger.Debugf("args: %v", args)

	if len(args) != 1 {
		customErr := &custom_errors.LengthCheckingError{Type: "GetServicesOfOrg arguments length"}
		logger.Errorf(customErr.Error())
		return nil, errors.New(customErr.Error())
	}

	// ==============================================================
	// Validation
	// ==============================================================
	orgID := args[0]
	if utils.IsStringEmpty(orgID) {
		customErr := &custom_errors.LengthCheckingError{Type: "OrgID"}
		logger.Errorf(customErr.Error())
		return nil, errors.WithStack(customErr)
	}

	// ==============================================================
	// Get all services of this org using index table
	// ==============================================================
	services := []Service{}

	// TODO: add limit to parameter
	iter, err := asset_mgmt.GetAssetManager(stub, caller).GetAssetIter(ServiceAssetNamespace, IndexService, []string{"org_id"}, []string{orgID}, []string{orgID}, true, false, OMRServiceAssetKeyPathFunc, "", 1000, nil)
	if err != nil {
		logger.Errorf("GetServiceAssets failed: %v", err)
		return nil, errors.Wrap(err, "GetServiceAssets failed")
	}
	// Iterate over all services
	defer iter.Close()
	for iter.HasNext() {
		serviceAsset, err := iter.Next()
		if err != nil {
			customErr := &custom_errors.IterError{}
			logger.Errorf("%v: %v", customErr, err)
			return nil, errors.Wrap(err, customErr.Error())
		}
		service := convertServiceFromAsset(serviceAsset)
		services = append(services, service)
	}

	return json.Marshal(&services)
}

func convertServiceToAsset(stub cached_stub.CachedStubInterface, service Service, callerID string) (data_model.Asset, error) {
	defer utils.ExitFnLog(utils.EnterFnLog())

	asset := data_model.Asset{}
	asset.AssetId = asset_mgmt.GetAssetId(ServiceAssetNamespace, service.ServiceID)
	asset.Datatypes = []string{}

	metaData := make(map[string]string)
	metaData["namespace"] = ServiceAssetNamespace
	asset.Metadata = metaData
	serviceAssetPublicData, err := getServiceAssetPublicData(service)
	if err != nil {
		logger.Errorf("getServiceAssetPublicData failed: %v", err)
		return data_model.Asset{}, errors.Wrap(err, "getServiceAssetPublicData failed")
	}
	asset.PublicData = serviceAssetPublicData
	asset.PrivateData, err = getServiceAssetPrivateData(service)
	if err != nil {
		logger.Errorf("getServiceAssetPrivateData failed: %v", err)
		return data_model.Asset{}, errors.Wrap(err, "getServiceAssetPrivateData failed")
	}

	asset.OwnerIds = []string{service.OrgID}

	asset.IndexTableName = IndexService

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

func convertServiceFromAsset(asset *data_model.Asset) Service {
	defer utils.ExitFnLog(utils.EnterFnLog())

	service := Service{}
	var publicData ServicePublicData
	var privateData servicePrivateData

	json.Unmarshal(asset.PublicData, &publicData)
	json.Unmarshal(asset.PrivateData, &privateData)

	service.ServiceID = publicData.ServiceID
	service.ServiceName = publicData.ServiceName
	service.Datatypes = publicData.Datatypes
	service.OrgID = publicData.OrgID
	service.Summary = publicData.Summary
	service.Terms = publicData.Terms
	service.PaymentRequired = publicData.PaymentRequired
	service.Status = publicData.Status
	service.CreateDate = publicData.CreateDate
	service.UpdateDate = publicData.UpdateDate

	service.Email = privateData.Email
	service.SolutionPrivateData = privateData.SolutionPrivateData

	return service
}

func convertToSubgroup(args []string) (data_model.User, error) {
	defer utils.ExitFnLog(utils.EnterFnLog())
	logger.Debugf("args: %v", args)

	subgroup := data_model.User{}

	service := Service{}
	serviceBytes := []byte(args[0])
	err := json.Unmarshal(serviceBytes, &service)
	if err != nil {
		customErr := &custom_errors.UnmarshalError{Type: "Service"}
		logger.Errorf("%v: %v", customErr, err)
		return subgroup, errors.Wrap(err, customErr.Error())
	}

	subgroupBytes := []byte(args[0])
	err = json.Unmarshal(subgroupBytes, &subgroup)
	if err != nil {
		customErr := &custom_errors.UnmarshalError{Type: "Subgroup"}
		logger.Errorf("%v: %v", customErr, err)
		return subgroup, errors.Wrap(err, customErr.Error())
	}

	subgroup.ID = service.ServiceID
	subgroup.Name = service.ServiceName

	publicData := make(map[string]interface{})
	publicData["org"] = service.OrgID
	subgroup.SolutionPublicData = publicData

	privateData := make(map[string]interface{})
	privateData["services"] = []string{service.ServiceID}
	privateData["is_org_admin"] = false
	privateData["data"] = service.SolutionPrivateData

	subgroup.SolutionPrivateData = privateData

	subgroup.SymKey, err = crypto.ParseSymKeyB64(subgroup.SymKeyB64)
	if err != nil {
		logger.Errorf("Failed to parse SymKeyB64")
		return subgroup, errors.Wrap(err, "Failed to parse SymKeyB64")
	}

	subgroup.PublicKey, err = crypto.ParsePublicKeyB64(subgroup.PublicKeyB64)
	if err != nil {
		logger.Errorf("Failed to parse PublicKeyB64")
		return subgroup, errors.Wrap(err, "Failed to parse PublicKeyB64")
	}

	subgroup.PrivateKey, err = crypto.ParsePrivateKeyB64(subgroup.PrivateKeyB64)
	if err != nil {
		logger.Errorf("Failed to parse PrivateKeyB64")
		return subgroup, errors.Wrap(err, "Failed to parse PrivateKeyB64")
	}

	return subgroup, nil
}

func getServiceAssetPublicData(service Service) ([]byte, error) {
	defer utils.ExitFnLog(utils.EnterFnLog())

	publicData := ServicePublicData{}
	publicData.ServiceID = service.ServiceID
	publicData.ServiceName = service.ServiceName
	publicData.Datatypes = service.Datatypes
	publicData.OrgID = service.OrgID
	publicData.Summary = service.Summary
	publicData.Terms = service.Terms
	publicData.PaymentRequired = service.PaymentRequired
	publicData.Status = service.Status
	publicData.CreateDate = service.CreateDate
	publicData.UpdateDate = service.UpdateDate

	publicBytes, err := json.Marshal(&publicData)
	if err != nil {
		customErr := &custom_errors.MarshalError{Type: "publicData"}
		logger.Errorf("%v: %v", customErr, err)
		return nil, errors.Wrap(err, customErr.Error())
	}

	return publicBytes, nil
}

func getServiceAssetPrivateData(service Service) ([]byte, error) {
	defer utils.ExitFnLog(utils.EnterFnLog())

	privateData := servicePrivateData{}
	privateData.SolutionPrivateData = service.SolutionPrivateData
	privateData.Email = service.Email
	privateBytes, err := json.Marshal(&privateData)
	if err != nil {
		customErr := &custom_errors.MarshalError{Type: "privateData"}
		logger.Errorf("%v: %v", customErr, err)
		return nil, errors.Wrap(err, customErr.Error())
	}

	return privateBytes, nil
}

func (s Service) hasDatatype(datatype string) bool {
	defer utils.ExitFnLog(utils.EnterFnLog())

	for _, dt := range s.Datatypes {
		if dt.DatatypeID == datatype {
			return true
		}
	}
	return false
}

func RemoveDatatypeFromList(list []ServiceDatatype, datatypeID string) []ServiceDatatype {
	defer utils.ExitFnLog(utils.EnterFnLog())

	for i, datatype := range list {
		if datatype.DatatypeID == datatypeID {
			list = append(list[:i], list[i+1:]...)
			break
		}
	}
	return list
}

func SetupServiceIndex(stub cached_stub.CachedStubInterface) error {
	defer utils.ExitFnLog(utils.EnterFnLog())

	// Service Indices
	serviceTable := index.GetTable(stub, IndexService, "service_id")
	serviceTable.AddIndex([]string{"org_id", "service_id"}, false)
	err := serviceTable.SaveToLedger()

	if err != nil {
		return err
	}

	return nil
}

// private function for validating a single datatype
func ValidateDatatype(stub cached_stub.CachedStubInterface, caller data_model.User, dtype ServiceDatatype) (bool, error) {
	defer utils.ExitFnLog(utils.EnterFnLog())

	if utils.IsStringEmpty(dtype.DatatypeID) {
		customErr := &custom_errors.LengthCheckingError{Type: "DatatypeID"}
		logger.Errorf(customErr.Error())
		return false, errors.WithStack(customErr)
	}

	if len(dtype.Access) <= 0 {
		customErr := &custom_errors.LengthCheckingError{Type: "Access"}
		logger.Errorf(customErr.Error())
		return false, errors.WithStack(customErr)
	}

	datatype, err := GetDatatypeWithParams(stub, caller, dtype.DatatypeID)
	if err != nil {
		logger.Errorf("Failed to GetDatatypeWithParams: %v, %v", dtype.DatatypeID, err)
		return false, errors.Wrap(err, "Failed to GetDatatypeWithParams: "+dtype.DatatypeID)
	}

	if utils.IsStringEmpty(datatype.DatatypeID) {
		customErr := &custom_errors.LengthCheckingError{Type: dtype.DatatypeID}
		logger.Errorf(customErr.Error())
		return false, errors.WithStack(customErr)
	}

	return true, nil
}

// CheckAccessToService is an internal function for checking a caller's access to a service
func CheckAccessToService(stub cached_stub.CachedStubInterface, caller data_model.User, serviceID string, access string) (bool, error) {
	defer utils.ExitFnLog(utils.EnterFnLog())

	if access != consent_mgmt.ACCESS_WRITE && access != consent_mgmt.ACCESS_READ {
		logger.Errorf("Access must be READ or WRITE")
		return false, errors.New("Access must be READ or WRITE")
	}

	serviceAssetID := asset_mgmt.GetAssetId(ServiceAssetNamespace, serviceID)

	userAccessManager := user_access_ctrl.GetUserAccessManager(stub, caller)
	accessControl := data_model.AccessControl{}
	accessControl.UserId = caller.ID
	accessControl.AssetId = serviceAssetID
	accessControl.Access = access

	return userAccessManager.CheckAccess(accessControl)
}

func GetOrgIDFromServiceSubgroup(subgroup data_model.User) string {
	if subgroup.SolutionPublicData != nil {
		if subgroup.SolutionPublicData.(map[string]interface{})["org"] != nil {
			return subgroup.SolutionPublicData.(map[string]interface{})["org"].(string)
		}
	}
	return ""
}
