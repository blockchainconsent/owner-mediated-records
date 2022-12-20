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
	"common/bchcls/cached_stub"
	"common/bchcls/history"
	"common/bchcls/init_common"
	"common/bchcls/user_mgmt"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos/peer"

	"errors"
)

// Chaincode is a sample definition of chaincode structure
type Chaincode struct {
}

// Init initializes the chaincode
func (t *Chaincode) Init(chaincodeStub shim.ChaincodeStubInterface) peer.Response {
	stub := cached_stub.NewCachedStub(chaincodeStub, true, true, true)

	// init common chaincode
	// 1. run a simple self test by writing to ledger and printing any errors
	// 2. initialize common packages by calling init_common.Init() with logLevel
	// 3. initialize default datastore if the first args is "_cloudant"
	logLevel, err := init_common.InitSetup(stub)
	if err != nil {
		logger.Errorf("InitSetup failed: %v", err)
		return shim.Error("InitSetup failed")
	}

	err = InitApp(stub, logLevel)
	if err != nil {
		logger.Errorf("Failed to run InitApp: %v", err)
		return shim.Error("Failed to run InitApp")
	}

	logger.Info("Initialization complete")

	return shim.Success(nil)
}

// Invoke is the entry point for chaincode functions
func (t *Chaincode) Invoke(chaincodeStub shim.ChaincodeStubInterface) (result peer.Response) {
	stub := cached_stub.NewCachedStub(chaincodeStub)

	function, args := stub.GetFunctionAndParameters()
	logger.Infof("=====> Invoke %v %v", function, args)
	defer logger.Debugf("<===== Invoke %v", function)

	// InvokeSetup performs the following:
	// - checks caller's identity and keys
	// - performs chaincode login
	// - decrypts args if needed
	// - retrieves and parses phi_args
	// - runs InitByInvoke if "Init" function is called
	caller, function, args, toReturn, err := init_common.InvokeSetup(stub)
	if err != nil {
		logger.Errorf("InvokeSetup failed: %v", err)
		return shim.Error("InvokeSetup failed")
	}
	if toReturn {
		return shim.Success(nil)
	}

	// error handling
	defer func() {
		if r := recover(); r != nil {
			var ok bool
			result, ok = r.(peer.Response)
			if !ok {
				logger.Errorf("pkg err %v", r)
				result = shim.Error("pkg err %v")
			}
		}
	}()

	logger.Debug("starting Invoke for: " + function)

	var returnBytes []byte
	returnError := errors.New("Unknown function: " + function)

	if function == "setupDatastore" {
		stub2 := cached_stub.NewCachedStub(chaincodeStub, true, true, true)
		returnError = SetupDatastore(stub2, caller, args)
	} else if function == "registerSystemAdmin" {
		returnBytes, returnError = user_mgmt.RegisterSystemAdmin(stub, caller, args)
	} else if function == "registerAuditor" {
		returnBytes, returnError = user_mgmt.RegisterAuditor(stub, caller, args)

		// ==============================================================
		// Refactored functions
		// ==============================================================

		// Users & Permissions
	} else if function == "registerUser" {
		returnBytes, returnError = RegisterUser(stub, caller, args)
	} else if function == "getUser" {
		returnBytes, returnError = GetSolutionUser(stub, caller, args)
	} else if function == "getUsers" {
		returnBytes, returnError = GetSolutionUsers(stub, caller, args)
	} else if function == "registerOrg" {
		returnBytes, returnError = RegisterOrg(stub, caller, args)
	} else if function == "updateOrg" {
		returnBytes, returnError = UpdateOrg(stub, caller, args)
	} else if function == "getOrg" {
		returnBytes, returnError = user_mgmt.GetOrg(stub, caller, args)
	} else if function == "getOrgs" {
		returnBytes, returnError = GetOrgs(stub, caller, args)
	} else if function == "PutUserInOrg" {
		returnBytes, returnError = PutUserInOrg(stub, caller, args)
	} else if function == "RemoveUserFromOrg" {
		returnBytes, returnError = RemoveUserFromOrg(stub, caller, args)
	} else if function == "addPermissionOrgAdmin" {
		returnBytes, returnError = AddPermissionOrgAdmin(stub, caller, args)
	} else if function == "deletePermissionOrgAdmin" {
		returnBytes, returnError = RemovePermissionOrgAdmin(stub, caller, args)
	} else if function == "addPermissionServiceAdmin" {
		returnBytes, returnError = AddPermissionServiceAdmin(stub, caller, args)
	} else if function == "deletePermissionServiceAdmin" {
		returnBytes, returnError = RemovePermissionServiceAdmin(stub, caller, args)
	} else if function == "addPermissionAuditor" {
		returnBytes, returnError = AddAuditorPermission(stub, caller, args)
	} else if function == "deletePermissionAuditor" {
		returnBytes, returnError = RemoveAuditorPermission(stub, caller, args)

		// Datatypes
	} else if function == "registerDatatype" {
		returnBytes, returnError = RegisterDatatype(stub, caller, args)
	} else if function == "updateDatatype" {
		returnBytes, returnError = UpdateDatatypeDescription(stub, caller, args)
	} else if function == "getDatatype" {
		returnBytes, returnError = GetDatatype(stub, caller, args)
	} else if function == "getAllDatatypes" {
		returnBytes, returnError = GetAllDatatypes(stub, caller, args)

		// Services
	} else if function == "registerService" {
		stub2 := cached_stub.NewCachedStub(chaincodeStub, true, true, true)
		returnBytes, returnError = RegisterService(stub2, caller, args)
	} else if function == "updateService" {
		returnBytes, returnError = UpdateService(stub, caller, args)
	} else if function == "getService" {
		returnBytes, returnError = GetService(stub, caller, args)
	} else if function == "addDatatypeToService" {
		returnBytes, returnError = AddDatatypeToService(stub, caller, args)
	} else if function == "removeDatatypeFromService" {
		returnBytes, returnError = RemoveDatatypeFromService(stub, caller, args)
	} else if function == "getServicesOfOrg" {
		returnBytes, returnError = GetServicesOfOrg(stub, caller, args)

		// Enrollment
	} else if function == "enrollPatient" {
		returnBytes, returnError = EnrollPatient(stub, caller, args)
	} else if function == "unenrollPatient" {
		returnBytes, returnError = UnenrollPatient(stub, caller, args)
	} else if function == "getPatientEnrollments" {
		returnBytes, returnError = GetPatientEnrollments(stub, caller, args)
	} else if function == "getServiceEnrollments" {
		returnBytes, returnError = GetServiceEnrollments(stub, caller, args)

		// Consent
	} else if function == "putConsentPatientData" {
		// get cached stub from chaincode stub, enabling putCache
		// because adding datatype key and put asset will be in the same transaction
		stub2 := cached_stub.NewCachedStub(chaincodeStub, true, true, true)
		returnBytes, returnError = PutConsentPatientData(stub2, caller, args)
	} else if function == "putConsentOwnerData" {
		// get cached stub from chaincode stub, enabling putCache
		// because adding datatype key and put asset will be in the same transaction
		stub2 := cached_stub.NewCachedStub(chaincodeStub, true, true, true)
		returnBytes, returnError = PutConsentOwnerData(stub2, caller, args)
	} else if function == "getConsent" {
		returnBytes, returnError = GetConsent(stub, caller, args)
	} else if function == "getConsentOwnerData" {
		returnBytes, returnError = GetConsent(stub, caller, args)
	} else if function == "getConsents" {
		returnBytes, returnError = GetConsents(stub, caller, args)
	} else if function == "getConsentsWithOwnerID" {
		returnBytes, returnError = GetConsentsWithOwnerID(stub, caller, args)
	} else if function == "getConsentsWithTargetID" {
		returnBytes, returnError = GetConsentsWithTargetID(stub, caller, args)
	} else if function == "validateConsent" {
		returnBytes, returnError = ValidateConsent(stub, caller, args)
	} else if function == "getConsentRequests" {
		returnBytes, returnError = GetAllConsentRequests(stub, caller, args)

		// User data
	} else if function == "uploadUserData" {
		// get cached stub from chaincode stub, enabling putCache
		// because adding datatype key and put asset will be in the same transaction
		stub2 := cached_stub.NewCachedStub(chaincodeStub, true, true, true)
		returnBytes, returnError = UploadUserData(stub2, caller, args)
	} else if function == "downloadUserData" {
		returnBytes, returnError = DownloadUserData(stub, caller, args)
	} else if function == "downloadUserDataConsentToken" {
		returnBytes, returnError = DownloadUserDataConsentToken(stub, caller, args)
	} else if function == "deleteUserData" {
		returnBytes, returnError = DeleteUserData(stub, caller, args)

		// Owner data
	} else if function == "uploadOwnerData" {
		// get cached stub from chaincode stub, enabling putCache
		// because adding datatype key and put asset will be in the same transaction
		stub2 := cached_stub.NewCachedStub(chaincodeStub, true, true, true)
		returnBytes, returnError = UploadOwnerData(stub2, caller, args)
	} else if function == "downloadOwnerDataAsOwner" {
		returnBytes, returnError = DownloadOwnerDataAsOwner(stub, caller, args)
	} else if function == "downloadOwnerDataAsRequester" {
		returnBytes, returnError = DownloadOwnerDataAsRequester(stub, caller, args)
	} else if function == "downloadOwnerDataWithConsent" {
		returnBytes, returnError = DownloadOwnerDataWithConsent(stub, caller, args)
	} else if function == "downloadOwnerDataConsentToken" {
		returnBytes, returnError = DownloadOwnerDataConsentToken(stub, caller, args)

		// Contract life cycle
	} else if function == "createContract" {
		returnBytes, returnError = CreateContract(stub, caller, args)
	} else if function == "addContractDetail" {
		returnBytes, returnError = AddContractDetail(stub, caller, args)
	} else if function == "addContractDetailDownload" {
		returnBytes, returnError = AddContractDetailDownload(stub, caller, args)
	} else if function == "givePermissionByContract" {
		returnBytes, returnError = GivePermissionByContract(stub, caller, args)
	} else if function == "getContract" {
		returnBytes, returnError = GetContract(stub, caller, args)
	} else if function == "getOwnerContracts" {
		returnBytes, returnError = GetContractsAsOwner(stub, caller, args)
	} else if function == "getRequesterContracts" {
		returnBytes, returnError = GetContractsAsRequester(stub, caller, args)

		// Logging
	} else if function == "getLogs" {
		returnBytes, returnError = GetLogs(stub, caller, args)
	} else if function == "addQueryTransactionLog" {
		returnError = history.PutQueryTransactionLog(stub, caller, args)
	} else if function == "addValidateConsentQueryLog" {
		returnError = AddValidateConsentQueryLog(stub, caller, args)
	}

	if returnError != nil {
		logger.Errorf("Invoke %v Error: %v", function, returnError)
		return shim.Error(returnError.Error())
	}

	logger.Debugf("Invoke %v Success", function)
	return shim.Success(returnBytes)
}

// Query is a legacy function and should not be used anymore
func (t *Chaincode) Query(stub shim.ChaincodeStubInterface) peer.Response {
	return shim.Error("Unknown supported call - Query()")
}
