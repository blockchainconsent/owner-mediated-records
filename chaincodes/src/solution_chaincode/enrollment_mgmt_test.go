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
	"common/bchcls/data_model"
	"common/bchcls/datastore/datastore_manager"
	"common/bchcls/init_common"
	"common/bchcls/test_utils"
	"common/bchcls/user_mgmt"
	"encoding/json"
	"testing"
	"time"

	"github.com/hyperledger/fabric/core/chaincode/shim"
)

func GenerateEnrollmentTest(userID string, serviceID string) Enrollment {
	enrollment := Enrollment{}
	enrollment.UserID = userID
	enrollment.ServiceID = serviceID
	enrollment.EnrollDate = time.Now().Unix()
	enrollment.Status = "active"

	return enrollment
}

func TestEnrollPatient(t *testing.T) {
	logger.SetLevel(shim.LogDebug)
	logger.Info("TestEnrollPatient function called")

	// create a MockStub
	mstub := SetupIndexesAndGetStub(t)

	// register admin user
	mstub.MockTransactionStart("t123")
	stub := cached_stub.NewCachedStub(mstub)
	init_common.Init(stub)
	systemAdmin := test_utils.CreateTestUser("systemAdmin")
	systemAdmin.Role = SOLUTION_ROLE_SYSTEM
	systemAdminBytes, _ := json.Marshal(&systemAdmin)
	_, err := RegisterUser(stub, systemAdmin, []string{string(systemAdminBytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterUser to succeed")
	mstub.MockTransactionEnd("t123")

	// Register system datatypes
	mstub.MockTransactionStart("init")
	stub = cached_stub.NewCachedStub(mstub)
	RegisterSystemDatatypeTest(t, stub, systemAdmin)
	mstub.MockTransactionEnd("init")

	// register org
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	org1 := test_utils.CreateTestGroup("org1")
	org1Bytes, _ := json.Marshal(&org1)
	_, err = RegisterOrg(stub, org1, []string{string(org1Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterOrg to succeed")
	mstub.MockTransactionEnd("t123")

	// get org back
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	org1OMR, err := GetSolutionUserWithParams(stub, org1, org1.ID, true, true)
	test_utils.AssertTrue(t, err == nil, "Expected GetUserData to succeed")
	test_utils.AssertTrue(t, org1OMR.Role == "org", "Expected GetSolutionUserWithParams to succeed")
	test_utils.AssertTrue(t, org1OMR.Org == org1.ID, "Expected GetUserData to succeed")
	mstub.MockTransactionEnd("t123")

	//  register datatype
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	datatype1 := Datatype{DatatypeID: "datatype1", Description: "datatype1"}
	datatype1Bytes, _ := json.Marshal(&datatype1)
	org1Caller, _ := user_mgmt.GetUserData(stub, org1, org1.ID, true, true)
	_, err = RegisterDatatype(stub, org1Caller, []string{string(datatype1Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterDatatype to succeed")
	mstub.MockTransactionEnd("t123")

	//  register service
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub, true, true, true)
	serviceDatatype1 := GenerateServiceDatatypeForTesting("datatype1", "service1", []string{consentOptionWrite, consentOptionRead})
	service1 := GenerateServiceForTesting("service1", "org1", []ServiceDatatype{serviceDatatype1})
	service1Bytes, _ := json.Marshal(&service1)
	_, err = RegisterService(stub, org1Caller, []string{string(service1Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterService to succeed")
	mstub.MockTransactionEnd("t123")

	// get service to make sure it was registered
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	serviceBytes, err := GetService(stub, org1Caller, []string{"service1"})
	test_utils.AssertTrue(t, err == nil, "GetService succcessful")
	var service = Service{}
	_ = json.Unmarshal(serviceBytes, &service)
	test_utils.AssertTrue(t, service.ServiceName == "service1 Name", "Got service name correctly")
	mstub.MockTransactionEnd("t123")

	// register patient
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	patient1 := test_utils.CreateTestUser("patient1")
	patient1Bytes, _ := json.Marshal(&patient1)
	_, err = user_mgmt.RegisterUser(stub, org1Caller, []string{string(patient1Bytes), "false"})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterUser to succeed")
	mstub.MockTransactionEnd("t123")

	// enroll patient
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	enrollment1 := GenerateEnrollmentTest("patient1", "service1")
	enrollment1Bytes, _ := json.Marshal(&enrollment1)
	enrollmentKey1 := test_utils.GenerateSymKey()
	enrollmentKey1B64 := crypto.EncodeToB64String(enrollmentKey1)
	_, err = EnrollPatient(stub, org1Caller, []string{string(enrollment1Bytes), enrollmentKey1B64})
	test_utils.AssertTrue(t, err == nil, "Expected EnrollPatient to succeed")
	mstub.MockTransactionEnd("t123")

	// get enrollment back
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	enrollmentResult, err := GetEnrollmentInternal(stub, org1Caller, "patient1", "service1", "org1")
	test_utils.AssertTrue(t, err == nil, "Expected GetEnrollmentInternal to succeed")
	test_utils.AssertTrue(t, enrollmentResult.ServiceName == "service1 Name", "Got service name correctly")
	test_utils.AssertTrue(t, enrollmentResult.Status == "active", "Expected enrollment status to be active")
	mstub.MockTransactionEnd("t123")

	// enroll patient1 to service 1 as service1 default admin
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	service1Subgroup, _ := user_mgmt.GetUserData(stub, org1Caller, "service1", true, true)
	_, err = EnrollPatient(stub, service1Subgroup, []string{string(enrollment1Bytes), enrollmentKey1B64})
	test_utils.AssertTrue(t, err == nil, "Expected EnrollPatient to succeed")
	mstub.MockTransactionEnd("t123")

	// get enrollment back
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	enrollmentResult, err = GetEnrollmentInternal(stub, service1Subgroup, "patient1", "service1")
	test_utils.AssertTrue(t, err == nil, "Expected GetEnrollmentInternal to succeed")
	test_utils.AssertTrue(t, enrollmentResult.ServiceName == "service1 Name", "Got service name correctly")
	test_utils.AssertTrue(t, enrollmentResult.Status == "active", "Expected enrollment status to be active")
	mstub.MockTransactionEnd("t123")

	// create an org user
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	orgUser1 := CreateTestSolutionUser("orgUser1")
	orgUser1.Org = "org1"
	orgUser1Bytes, _ := json.Marshal(&orgUser1)
	_, err = RegisterUser(stub, org1Caller, []string{string(orgUser1Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterUser to succeed")
	mstub.MockTransactionEnd("t123")

	// put orgUser1 in org1
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = PutUserInOrg(stub, org1Caller, []string{orgUser1.ID, org1.ID, "false"})
	test_utils.AssertTrue(t, err == nil, "Expected PutUserInOrg to succeed")
	mstub.MockTransactionEnd("t123")

	// enroll as orgUser1 without admin permission
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	orgUser1Caller, _ := user_mgmt.GetUserData(stub, org1Caller, orgUser1.ID, true, true)
	_, err = EnrollPatient(stub, orgUser1Caller, []string{string(enrollment1Bytes), enrollmentKey1B64})
	test_utils.AssertTrue(t, err != nil, "Expected EnrollPatient to fail")
	mstub.MockTransactionEnd("t123")

	// give org admin permission
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = AddPermissionOrgAdmin(stub, org1Caller, []string{orgUser1.ID, org1.ID})
	test_utils.AssertTrue(t, err == nil, "Expected AddPermissionOrgAdmin to succeed")
	mstub.MockTransactionEnd("t123")

	// enroll as orgUser1 with admin permission
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	orgUser1Caller, _ = user_mgmt.GetUserData(stub, org1Caller, orgUser1.ID, true, true)
	_, err = EnrollPatient(stub, orgUser1Caller, []string{string(enrollment1Bytes), enrollmentKey1B64})
	test_utils.AssertTrue(t, err == nil, "Expected EnrollPatient to succeed")
	mstub.MockTransactionEnd("t123")

	// get enrollment back
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	enrollmentResult, err = GetEnrollmentInternal(stub, orgUser1Caller, "patient1", "service1", "org1")
	test_utils.AssertTrue(t, err == nil, "Expected GetEnrollmentInternal to succeed")
	test_utils.AssertTrue(t, enrollmentResult.ServiceName == "service1 Name", "Expected ServiceName: service1 Name")
	test_utils.AssertTrue(t, enrollmentResult.Status == "active", "Expected enrollment status to be active")
	mstub.MockTransactionEnd("t123")

	// remove org admin permission
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = RemovePermissionOrgAdmin(stub, org1Caller, []string{orgUser1.ID, org1.ID})
	test_utils.AssertTrue(t, err == nil, "Expected RemovePermissionOrgAdmin to succeed")
	mstub.MockTransactionEnd("t123")

	// enroll as orgUser1 without admin permission
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	orgUser1Caller, _ = user_mgmt.GetUserData(stub, org1Caller, orgUser1.ID, true, true)
	_, err = EnrollPatient(stub, orgUser1Caller, []string{string(enrollment1Bytes), enrollmentKey1B64})
	test_utils.AssertTrue(t, err != nil, "Expected EnrollPatient to fail")
	mstub.MockTransactionEnd("t123")

	// get enrollment back
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	enrollmentResult, err = GetEnrollmentInternal(stub, orgUser1Caller, "patient1", "service1", "org1")
	test_utils.AssertTrue(t, err != nil, "Expected GetEnrollmentInternal to fail")
	mstub.MockTransactionEnd("t123")

	// give service admin permission
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = AddPermissionServiceAdmin(stub, org1Caller, []string{orgUser1.ID, "service1"})
	test_utils.AssertTrue(t, err == nil, "Expected AddPermissionServiceAdmin to succeed")
	mstub.MockTransactionEnd("t123")

	// enroll as orgUser1 with service admin permission
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	orgUser1Caller, _ = user_mgmt.GetUserData(stub, org1Caller, orgUser1.ID, true, true)
	_, err = EnrollPatient(stub, orgUser1Caller, []string{string(enrollment1Bytes), enrollmentKey1B64})
	test_utils.AssertTrue(t, err == nil, "Expected EnrollPatient to succeed")
	mstub.MockTransactionEnd("t123")

	// get enrollment back
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	enrollmentResult, err = GetEnrollmentInternal(stub, orgUser1Caller, "patient1", "service1", "org1")
	test_utils.AssertTrue(t, err == nil, "Expected GetEnrollmentInternal to succeed")
	test_utils.AssertTrue(t, enrollmentResult.ServiceName == "service1 Name", "Got service name correctly")
	test_utils.AssertTrue(t, enrollmentResult.Status == "active", "Expected enrollment status to be active")
	mstub.MockTransactionEnd("t123")
}

func TestEnrollPatient_OffChain(t *testing.T) {
	logger.SetLevel(shim.LogDebug)
	logger.Info("TestEnrollPatient_OffChain function called")

	// create a MockStub
	mstub := SetupIndexesAndGetStub(t)

	// register admin user
	mstub.MockTransactionStart("t123")
	stub := cached_stub.NewCachedStub(mstub)
	init_common.Init(stub)
	systemAdmin := test_utils.CreateTestUser("systemAdmin")
	systemAdmin.Role = SOLUTION_ROLE_SYSTEM
	systemAdminBytes, _ := json.Marshal(&systemAdmin)
	_, err := RegisterUser(stub, systemAdmin, []string{string(systemAdminBytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterUser to succeed")
	mstub.MockTransactionEnd("t123")

	// setup off-chain datastore
	datastoreConnectionID := "cloudant1"
	err = setupDatastore(mstub, systemAdmin, datastoreConnectionID)
	test_utils.AssertTrue(t, err == nil, "Expected setupDatastore to succeed")

	// Register system datatypes
	mstub.MockTransactionStart("init")
	stub = cached_stub.NewCachedStub(mstub)
	RegisterSystemDatatypeTest(t, stub, systemAdmin)
	mstub.MockTransactionEnd("init")

	// register org
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	org1 := test_utils.CreateTestGroup("org1")
	org1Bytes, _ := json.Marshal(&org1)
	_, err = RegisterOrg(stub, org1, []string{string(org1Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterOrg to succeed")
	mstub.MockTransactionEnd("t123")

	// register datatype
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	datatype1 := Datatype{DatatypeID: "datatype1", Description: "datatype1"}
	datatype1Bytes, _ := json.Marshal(&datatype1)
	org1Caller, _ := user_mgmt.GetUserData(stub, org1, org1.ID, true, true)
	_, err = RegisterDatatype(stub, org1Caller, []string{string(datatype1Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterDatatype to succeed")
	mstub.MockTransactionEnd("t123")

	// register service
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub, true, true, true)
	serviceDatatype1 := GenerateServiceDatatypeForTesting("datatype1", "service1", []string{consentOptionWrite, consentOptionRead})
	service1 := GenerateServiceForTesting("service1", "org1", []ServiceDatatype{serviceDatatype1})
	service1Bytes, _ := json.Marshal(&service1)
	_, err = RegisterService(stub, org1Caller, []string{string(service1Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterService to succeed")
	mstub.MockTransactionEnd("t123")

	// register patient
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	patient1 := test_utils.CreateTestUser("patient1")
	patient1Bytes, _ := json.Marshal(&patient1)
	_, err = user_mgmt.RegisterUser(stub, org1Caller, []string{string(patient1Bytes), "false"})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterUser to succeed")
	mstub.MockTransactionEnd("t123")

	// enroll patient
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	enrollment1 := GenerateEnrollmentTest("patient1", "service1")
	enrollment1Bytes, _ := json.Marshal(&enrollment1)
	enrollmentKey1 := test_utils.GenerateSymKey()
	enrollmentKey1B64 := crypto.EncodeToB64String(enrollmentKey1)
	_, err = EnrollPatient(stub, org1Caller, []string{string(enrollment1Bytes), enrollmentKey1B64})
	test_utils.AssertTrue(t, err == nil, "Expected EnrollPatient to succeed")
	mstub.MockTransactionEnd("t123")

	// get enrollment back as org1
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	enrollmentResult, err := GetEnrollmentInternal(stub, org1Caller, "patient1", "service1", "org1")
	test_utils.AssertTrue(t, err == nil, "Expected GetEnrollmentInternal to succeed")
	test_utils.AssertTrue(t, enrollmentResult.ServiceID == "service1", "Expected ServiceID")
	test_utils.AssertTrue(t, enrollmentResult.ServiceName == "service1 Name", "Expected ServiceName")
	test_utils.AssertTrue(t, enrollmentResult.UserID == "patient1", "Expected UserID")
	test_utils.AssertTrue(t, enrollmentResult.UserName != "", "Expected UserName")
	test_utils.AssertTrue(t, enrollmentResult.EnrollDate > 0, "Expected EnrollDate")
	test_utils.AssertTrue(t, enrollmentResult.Status == "active", "Expected Status")
	mstub.MockTransactionEnd("t123")

	// verify data access for org1
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	assetManager := asset_mgmt.GetAssetManager(stub, org1Caller)
	enrollmentID := GetEnrollmentID("patient1", "service1")
	enrollmentAssetID := asset_mgmt.GetAssetId(EnrollmentAssetNamespace, enrollmentID)
	keyPath, err := GetKeyPath(stub, org1Caller, enrollmentAssetID)
	test_utils.AssertTrue(t, err == nil, "Expected GetKeyPath user data to succeed")
	enrollmentAssetKey, err := assetManager.GetAssetKey(enrollmentAssetID, keyPath)
	test_utils.AssertTrue(t, err == nil, "Expected GetAssetKey to succeed")
	enrollmentAsset, err := assetManager.GetAsset(enrollmentAssetID, enrollmentAssetKey)
	test_utils.AssertTrue(t, err == nil, "Expected GetAsset to succeed")
	mstub.MockTransactionEnd("t123")

	// org1 has access to public data
	enrollmentPublicData := Enrollment{}
	json.Unmarshal(enrollmentAsset.PublicData, &enrollmentPublicData)
	test_utils.AssertTrue(t, enrollmentPublicData.EnrollmentID != "", "Expected EnrollmentID")
	test_utils.AssertTrue(t, enrollmentPublicData.UserID != "", "Expected UserID")
	test_utils.AssertTrue(t, enrollmentPublicData.UserName != "", "Expected UserName")
	test_utils.AssertTrue(t, enrollmentPublicData.ServiceID != "", "Expected ServiceID")
	test_utils.AssertTrue(t, enrollmentPublicData.ServiceName != "", "Expected ServiceName")

	// org1 has access to private data
	enrollmentPrivateData := Enrollment{}
	json.Unmarshal(enrollmentAsset.PrivateData, &enrollmentPrivateData)
	test_utils.AssertTrue(t, enrollmentPrivateData.EnrollDate > 0, "Expected EnrollDate")
	test_utils.AssertTrue(t, enrollmentPrivateData.Status != "", "Expected Status")

	// create an org user
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	orgUser1 := CreateTestSolutionUser("orgUser1")
	orgUser1.Org = "org1"
	orgUser1Bytes, _ := json.Marshal(&orgUser1)
	_, err = RegisterUser(stub, org1Caller, []string{string(orgUser1Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterUser to succeed")
	mstub.MockTransactionEnd("t123")

	// put orgUser1 in org1
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = PutUserInOrg(stub, org1Caller, []string{orgUser1.ID, org1.ID, "false"})
	test_utils.AssertTrue(t, err == nil, "Expected PutUserInOrg to succeed")
	mstub.MockTransactionEnd("t123")

	// give org admin permission
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = AddPermissionOrgAdmin(stub, org1Caller, []string{orgUser1.ID, org1.ID})
	test_utils.AssertTrue(t, err == nil, "Expected AddPermissionOrgAdmin to succeed")
	mstub.MockTransactionEnd("t123")

	// get enrollment back as org1 user (org admin)
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	orgUser1Caller, _ := user_mgmt.GetUserData(stub, org1Caller, orgUser1.ID, true, true)
	enrollmentResult, err = GetEnrollmentInternal(stub, orgUser1Caller, "patient1", "service1", "org1")
	test_utils.AssertTrue(t, err == nil, "Expected GetEnrollmentInternal to succeed")
	test_utils.AssertTrue(t, enrollmentResult.ServiceID == "service1", "Expected ServiceID")
	test_utils.AssertTrue(t, enrollmentResult.Status == "active", "Expected Status")
	mstub.MockTransactionEnd("t123")

	// verify data access for org1 user (org admin)
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	assetManager = asset_mgmt.GetAssetManager(stub, orgUser1Caller)
	keyPath, err = GetKeyPath(stub, orgUser1Caller, enrollmentAssetID, "org1")
	test_utils.AssertTrue(t, err == nil, "Expected GetKeyPath user data to succeed")
	enrollmentAssetKey, err = assetManager.GetAssetKey(enrollmentAssetID, keyPath)
	test_utils.AssertTrue(t, err == nil, "Expected GetAssetKey to succeed")
	enrollmentAsset, err = assetManager.GetAsset(enrollmentAssetID, enrollmentAssetKey)
	test_utils.AssertTrue(t, err == nil, "Expected GetAsset to succeed")
	mstub.MockTransactionEnd("t123")

	// org1 user has access to public data
	enrollmentPublicData = Enrollment{}
	json.Unmarshal(enrollmentAsset.PublicData, &enrollmentPublicData)
	test_utils.AssertTrue(t, enrollmentPublicData.EnrollmentID != "", "Expected EnrollmentID")
	test_utils.AssertTrue(t, enrollmentPublicData.UserID != "", "Expected UserID")
	test_utils.AssertTrue(t, enrollmentPublicData.UserName != "", "Expected UserName")
	test_utils.AssertTrue(t, enrollmentPublicData.ServiceID != "", "Expected ServiceID")
	test_utils.AssertTrue(t, enrollmentPublicData.ServiceName != "", "Expected ServiceName")

	// org1 user has access to private data
	enrollmentPrivateData = Enrollment{}
	json.Unmarshal(enrollmentAsset.PrivateData, &enrollmentPrivateData)
	test_utils.AssertTrue(t, enrollmentPrivateData.EnrollDate > 0, "Expected EnrollDate")
	test_utils.AssertTrue(t, enrollmentPrivateData.Status != "", "Expected Status")

	// remove org admin permission
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = RemovePermissionOrgAdmin(stub, org1Caller, []string{orgUser1.ID, org1.ID})
	test_utils.AssertTrue(t, err == nil, "Expected RemovePermissionOrgAdmin to succeed")
	mstub.MockTransactionEnd("t123")

	// attempt to get enrollment back as org1 user (no permission)
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	enrollmentResult, err = GetEnrollmentInternal(stub, orgUser1Caller, "patient1", "service1", "org1")
	test_utils.AssertTrue(t, err != nil, "Expected GetEnrollmentInternal to fail")
	mstub.MockTransactionEnd("t123")

	// verify data access for org1 user (no permission)
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	assetManager = asset_mgmt.GetAssetManager(stub, orgUser1Caller)
	keyPath, err = GetKeyPath(stub, orgUser1Caller, enrollmentAssetID, "org1")
	test_utils.AssertTrue(t, err == nil, "Expected GetKeyPath user data to succeed")
	_, err = assetManager.GetAssetKey(enrollmentAssetID, keyPath)
	test_utils.AssertTrue(t, err != nil, "Expected GetAssetKey to fail")
	enrollmentAsset, err = assetManager.GetAsset(enrollmentAssetID, data_model.Key{})
	test_utils.AssertTrue(t, err == nil, "Expected GetAsset to succeed")
	mstub.MockTransactionEnd("t123")

	// org1 user has access to public data
	enrollmentPublicData = Enrollment{}
	json.Unmarshal(enrollmentAsset.PublicData, &enrollmentPublicData)
	test_utils.AssertTrue(t, enrollmentPublicData.EnrollmentID != "", "Expected EnrollmentID")
	test_utils.AssertTrue(t, enrollmentPublicData.UserID != "", "Expected UserID")
	test_utils.AssertTrue(t, enrollmentPublicData.UserName != "", "Expected UserName")
	test_utils.AssertTrue(t, enrollmentPublicData.ServiceID != "", "Expected ServiceID")
	test_utils.AssertTrue(t, enrollmentPublicData.ServiceName != "", "Expected ServiceName")

	// org1 user has no access to private data
	enrollmentPrivateData = Enrollment{}
	json.Unmarshal(enrollmentAsset.PrivateData, &enrollmentPrivateData)
	test_utils.AssertTrue(t, enrollmentPrivateData.EnrollDate == 0, "Expected no EnrollDate")
	test_utils.AssertTrue(t, enrollmentPrivateData.Status == "", "Expected no Status")

	// give service admin permission
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = AddPermissionServiceAdmin(stub, org1Caller, []string{orgUser1.ID, "service1"})
	test_utils.AssertTrue(t, err == nil, "Expected AddPermissionServiceAdmin to succeed")
	mstub.MockTransactionEnd("t123")

	// get enrollment back as org1 user (service admin)
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	enrollmentResult, err = GetEnrollmentInternal(stub, orgUser1Caller, "patient1", "service1", "org1")
	test_utils.AssertTrue(t, err == nil, "Expected GetEnrollmentInternal to succeed")
	test_utils.AssertTrue(t, enrollmentResult.ServiceID == "service1", "Expected ServiceID")
	test_utils.AssertTrue(t, enrollmentResult.Status == "active", "Expected Status")
	mstub.MockTransactionEnd("t123")

	// verify data access for org1 user (service admin)
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	assetManager = asset_mgmt.GetAssetManager(stub, orgUser1Caller)
	keyPath, err = GetKeyPath(stub, orgUser1Caller, enrollmentAssetID, "org1")
	test_utils.AssertTrue(t, err == nil, "Expected GetKeyPath user data to succeed")
	enrollmentAssetKey, err = assetManager.GetAssetKey(enrollmentAssetID, keyPath)
	test_utils.AssertTrue(t, err == nil, "Expected GetAssetKey to succeed")
	enrollmentAsset, err = assetManager.GetAsset(enrollmentAssetID, enrollmentAssetKey)
	test_utils.AssertTrue(t, err == nil, "Expected GetAsset to succeed")
	mstub.MockTransactionEnd("t123")

	// org1 user has access to public data
	enrollmentPublicData = Enrollment{}
	json.Unmarshal(enrollmentAsset.PublicData, &enrollmentPublicData)
	test_utils.AssertTrue(t, enrollmentPublicData.EnrollmentID != "", "Expected EnrollmentID")
	test_utils.AssertTrue(t, enrollmentPublicData.UserID != "", "Expected UserID")
	test_utils.AssertTrue(t, enrollmentPublicData.UserName != "", "Expected UserName")
	test_utils.AssertTrue(t, enrollmentPublicData.ServiceID != "", "Expected ServiceID")
	test_utils.AssertTrue(t, enrollmentPublicData.ServiceName != "", "Expected ServiceName")

	// org1 user has access to private data
	enrollmentPrivateData = Enrollment{}
	json.Unmarshal(enrollmentAsset.PrivateData, &enrollmentPrivateData)
	test_utils.AssertTrue(t, enrollmentPrivateData.EnrollDate > 0, "Expected EnrollDate")
	test_utils.AssertTrue(t, enrollmentPrivateData.Status != "", "Expected Status")

	// register unrelatedUser
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	unrelatedUser := test_utils.CreateTestUser("unrelatedUser")
	unrelatedUserBytes, _ := json.Marshal(&unrelatedUser)
	_, err = user_mgmt.RegisterUser(stub, unrelatedUser, []string{string(unrelatedUserBytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterUser to succeed")
	mstub.MockTransactionEnd("t123")

	// attempt to get enrollment back as unrelatedUser
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	enrollmentResult, err = GetEnrollmentInternal(stub, unrelatedUser, "patient1", "service1", "org1")
	test_utils.AssertTrue(t, err != nil, "Expected GetEnrollmentInternal to fail")
	mstub.MockTransactionEnd("t123")

	// verify data access for unrelatedUser
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	assetManager = asset_mgmt.GetAssetManager(stub, unrelatedUser)
	keyPath, err = GetKeyPath(stub, unrelatedUser, enrollmentAssetID, "org1")
	test_utils.AssertTrue(t, err == nil, "Expected GetKeyPath user data to succeed")
	_, err = assetManager.GetAssetKey(enrollmentAssetID, keyPath)
	test_utils.AssertTrue(t, err != nil, "Expected GetAssetKey to fail")
	enrollmentAsset, err = assetManager.GetAsset(enrollmentAssetID, data_model.Key{})
	test_utils.AssertTrue(t, err == nil, "Expected GetAsset to succeed")
	mstub.MockTransactionEnd("t123")

	// unrelatedUser has access to public data
	enrollmentPublicData = Enrollment{}
	json.Unmarshal(enrollmentAsset.PublicData, &enrollmentPublicData)
	test_utils.AssertTrue(t, enrollmentPublicData.EnrollmentID != "", "Expected EnrollmentID")
	test_utils.AssertTrue(t, enrollmentPublicData.UserID != "", "Expected UserID")
	test_utils.AssertTrue(t, enrollmentPublicData.UserName != "", "Expected UserName")
	test_utils.AssertTrue(t, enrollmentPublicData.ServiceID != "", "Expected ServiceID")
	test_utils.AssertTrue(t, enrollmentPublicData.ServiceName != "", "Expected ServiceName")

	// unrelatedUser has no access to private data
	enrollmentPrivateData = Enrollment{}
	json.Unmarshal(enrollmentAsset.PrivateData, &enrollmentPrivateData)
	test_utils.AssertTrue(t, enrollmentPrivateData.EnrollDate == 0, "Expected no EnrollDate")
	test_utils.AssertTrue(t, enrollmentPrivateData.Status == "", "Expected no Status")

	// remove datastore connection
	mstub.MockTransactionStart("t123")
	err = datastore_manager.DeleteDatastoreConnection(stub, systemAdmin, datastoreConnectionID)
	test_utils.AssertTrue(t, err == nil, "Expected DeleteDatastoreConnection to succeed")
	mstub.MockTransactionEnd("t123")

	// attempt to get enrollment back as org1 user
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	enrollmentResult, err = GetEnrollmentInternal(stub, orgUser1Caller, "patient1", "service1", "org1")
	test_utils.AssertTrue(t, err != nil, "Expected GetEnrollmentInternal to fail")
	mstub.MockTransactionEnd("t123")

	// verify data access for org1 user
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	assetManager = asset_mgmt.GetAssetManager(stub, orgUser1Caller)
	keyPath, err = GetKeyPath(stub, orgUser1Caller, enrollmentAssetID, "org1")
	test_utils.AssertTrue(t, err == nil, "Expected GetKeyPath user data to succeed")
	enrollmentAssetKey, err = assetManager.GetAssetKey(enrollmentAssetID, keyPath)
	test_utils.AssertTrue(t, err == nil, "Expected GetAssetKey to succeed")
	_, err = assetManager.GetAsset(enrollmentAssetID, enrollmentAssetKey)
	test_utils.AssertTrue(t, err != nil, "Expected GetAsset to fail")
	mstub.MockTransactionEnd("t123")
}

func TestUnenrollPatient(t *testing.T) {
	logger.SetLevel(shim.LogDebug)
	logger.Info("TestUnenrollPatient function called")

	// create a MockStub
	mstub := SetupIndexesAndGetStub(t)

	// register admin user
	mstub.MockTransactionStart("t123")
	stub := cached_stub.NewCachedStub(mstub)
	init_common.Init(stub)
	systemAdmin := test_utils.CreateTestUser("systemAdmin")
	systemAdmin.Role = SOLUTION_ROLE_SYSTEM
	systemAdminBytes, _ := json.Marshal(&systemAdmin)
	_, err := RegisterUser(stub, systemAdmin, []string{string(systemAdminBytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterUser to succeed")
	mstub.MockTransactionEnd("t123")

	// Register system datatypes
	mstub.MockTransactionStart("init")
	stub = cached_stub.NewCachedStub(mstub)
	RegisterSystemDatatypeTest(t, stub, systemAdmin)
	mstub.MockTransactionEnd("init")

	// register org
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	org1 := test_utils.CreateTestGroup("org1")
	org1Bytes, _ := json.Marshal(&org1)
	_, err = RegisterOrg(stub, org1, []string{string(org1Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterOrg to succeed")
	mstub.MockTransactionEnd("t123")

	// get org back
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	org1OMR, err := GetSolutionUserWithParams(stub, org1, org1.ID, true, true)
	test_utils.AssertTrue(t, err == nil, "Expected GetUserData to succeed")
	test_utils.AssertTrue(t, org1OMR.Role == "org", "Expected GetSolutionUserWithParams to succeed")
	test_utils.AssertTrue(t, org1OMR.Org == org1.ID, "Expected GetUserData to succeed")
	mstub.MockTransactionEnd("t123")

	//  register datatype
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	datatype1 := Datatype{DatatypeID: "datatype1", Description: "datatype1"}
	datatype1Bytes, _ := json.Marshal(&datatype1)
	org1Caller, _ := user_mgmt.GetUserData(stub, org1, org1.ID, true, true)
	_, err = RegisterDatatype(stub, org1Caller, []string{string(datatype1Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterDatatype to succeed")
	mstub.MockTransactionEnd("t123")

	//  register service
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub, true, true, true)
	serviceDatatype1 := GenerateServiceDatatypeForTesting("datatype1", "service1", []string{consentOptionWrite, consentOptionRead})
	service1 := GenerateServiceForTesting("service1", "org1", []ServiceDatatype{serviceDatatype1})
	service1Bytes, _ := json.Marshal(&service1)
	_, err = RegisterService(stub, org1Caller, []string{string(service1Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterService to succeed")
	mstub.MockTransactionEnd("t123")

	// get service to make sure it was registered
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	serviceBytes, err := GetService(stub, org1Caller, []string{"service1"})
	test_utils.AssertTrue(t, err == nil, "GetService succcessful")
	var service = Service{}
	_ = json.Unmarshal(serviceBytes, &service)
	test_utils.AssertTrue(t, service.ServiceName == "service1 Name", "Got service name correctly")
	mstub.MockTransactionEnd("t123")

	// register patient
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	patient1 := test_utils.CreateTestUser("patient1")
	patient1Bytes, _ := json.Marshal(&patient1)
	_, err = user_mgmt.RegisterUser(stub, org1Caller, []string{string(patient1Bytes), "false"})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterUser to succeed")
	mstub.MockTransactionEnd("t123")

	// enroll patient
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	enrollment1 := GenerateEnrollmentTest("patient1", "service1")
	enrollment1Bytes, _ := json.Marshal(&enrollment1)
	enrollmentKey1 := test_utils.GenerateSymKey()
	enrollmentKey1B64 := crypto.EncodeToB64String(enrollmentKey1)
	_, err = EnrollPatient(stub, org1Caller, []string{string(enrollment1Bytes), enrollmentKey1B64})
	test_utils.AssertTrue(t, err == nil, "Expected EnrollPatient to succeed")
	mstub.MockTransactionEnd("t123")

	// Unenroll patient
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = UnenrollPatient(stub, org1Caller, []string{"service1", "patient1"})
	test_utils.AssertTrue(t, err == nil, "Expected UnenrollPatient to succeed")
	mstub.MockTransactionEnd("t123")

	// get enrollment back, see status is inactive
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	enrollmentResult, err := GetEnrollmentInternal(stub, org1, "patient1", "service1")
	test_utils.AssertTrue(t, err == nil, "Expected GetEnrollmentInternal to succeed")
	test_utils.AssertTrue(t, enrollmentResult.Status == "inactive", "Expected enrollment status to be inactive")
	mstub.MockTransactionEnd("t123")
}

func TestGetPatientEnrollments(t *testing.T) {
	logger.SetLevel(shim.LogDebug)
	logger.Info("TestGetPatientEnrollments function called")

	// create a MockStub
	mstub := SetupIndexesAndGetStub(t)

	// register admin user
	mstub.MockTransactionStart("t123")
	stub := cached_stub.NewCachedStub(mstub)
	init_common.Init(stub)
	err := SetupEnrollmentIndex(stub)
	test_utils.AssertTrue(t, err == nil, "Expected SetupEnrollmentIndex to succeed")
	systemAdmin := test_utils.CreateTestUser("systemAdmin")
	systemAdmin.Role = SOLUTION_ROLE_SYSTEM
	systemAdminBytes, _ := json.Marshal(&systemAdmin)
	_, err = RegisterUser(stub, systemAdmin, []string{string(systemAdminBytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterUser to succeed")
	mstub.MockTransactionEnd("t123")

	// Register system datatypes
	mstub.MockTransactionStart("init")
	stub = cached_stub.NewCachedStub(mstub)
	RegisterSystemDatatypeTest(t, stub, systemAdmin)
	mstub.MockTransactionEnd("init")

	// register org
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	org1 := test_utils.CreateTestGroup("org1")
	org1Bytes, _ := json.Marshal(&org1)
	_, err = RegisterOrg(stub, org1, []string{string(org1Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterOrg to succeed")
	mstub.MockTransactionEnd("t123")

	//  register datatype
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	datatype1 := Datatype{DatatypeID: "datatype1", Description: "datatype1"}
	datatype1Bytes, _ := json.Marshal(&datatype1)
	org1Caller, _ := user_mgmt.GetUserData(stub, org1, org1.ID, true, true)
	_, err = RegisterDatatype(stub, org1Caller, []string{string(datatype1Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterDatatype to succeed")
	mstub.MockTransactionEnd("t123")

	//  register service
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub, true, true, true)
	serviceDatatype1 := GenerateServiceDatatypeForTesting("datatype1", "service1", []string{consentOptionWrite, consentOptionRead})
	service1 := GenerateServiceForTesting("service1", "org1", []ServiceDatatype{serviceDatatype1})
	service1Bytes, _ := json.Marshal(&service1)
	_, err = RegisterService(stub, org1Caller, []string{string(service1Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterService to succeed")
	mstub.MockTransactionEnd("t123")

	// register service2
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub, true, true, true)
	service2 := GenerateServiceForTesting("service2", "org1", []ServiceDatatype{serviceDatatype1})
	service2Bytes, _ := json.Marshal(&service2)
	_, err = RegisterService(stub, org1Caller, []string{string(service2Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterService to succeed")
	mstub.MockTransactionEnd("t123")

	// register patient1
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	patient1 := test_utils.CreateTestUser("patient1")
	patient1Bytes, _ := json.Marshal(&patient1)
	_, err = user_mgmt.RegisterUser(stub, org1Caller, []string{string(patient1Bytes), "false"})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterUser to succeed")
	mstub.MockTransactionEnd("t123")

	// register patient2
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	patient2 := test_utils.CreateTestUser("patient2")
	patient2Bytes, _ := json.Marshal(&patient2)
	_, err = user_mgmt.RegisterUser(stub, org1Caller, []string{string(patient2Bytes), "false"})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterUser to succeed")
	mstub.MockTransactionEnd("t123")

	// enroll patient to service 1 and service 2
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	enrollment1 := GenerateEnrollmentTest("patient1", "service1")
	enrollment1Bytes, _ := json.Marshal(&enrollment1)
	enrollmentKey1 := test_utils.GenerateSymKey()
	enrollmentKey1B64 := crypto.EncodeToB64String(enrollmentKey1)
	_, err = EnrollPatient(stub, org1Caller, []string{string(enrollment1Bytes), enrollmentKey1B64})
	test_utils.AssertTrue(t, err == nil, "Expected EnrollPatient to succeed")
	enrollment2 := GenerateEnrollmentTest("patient1", "service2")
	enrollment2Bytes, _ := json.Marshal(&enrollment2)
	enrollmentKey2 := test_utils.GenerateSymKey()
	enrollmentKey2B64 := crypto.EncodeToB64String(enrollmentKey2)
	_, err = EnrollPatient(stub, org1Caller, []string{string(enrollment2Bytes), enrollmentKey2B64})
	test_utils.AssertTrue(t, err == nil, "Expected EnrollPatient to succeed")
	mstub.MockTransactionEnd("t123")

	// get patient enrollments as org1
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	enrollmentsBytes, err := GetPatientEnrollments(stub, org1Caller, []string{"patient1"})
	test_utils.AssertTrue(t, err == nil, "Expected GetPatientEnrollments to succeed")
	var enrollments = []Enrollment{}
	_ = json.Unmarshal(enrollmentsBytes, &enrollments)
	test_utils.AssertTrue(t, enrollments[0].ServiceID == "service1", "Got service 1 correctly")
	test_utils.AssertTrue(t, enrollments[1].ServiceID == "service2", "Got service 2 correctly")
	// get patient enrollments that are inactive
	// expecting none
	enrollmentsBytes, err = GetPatientEnrollments(stub, org1Caller, []string{"patient1", "inactive"})
	test_utils.AssertTrue(t, err == nil, "Expected GetPatientEnrollments to succeed")
	_ = json.Unmarshal(enrollmentsBytes, &enrollments)
	test_utils.AssertTrue(t, len(enrollments) == 0, "Expected 0 inactive enrollments")
	mstub.MockTransactionEnd("t123")

	// create an org user
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	orgUser1 := CreateTestSolutionUser("orgUser1")
	orgUser1.Org = "org1"
	orgUser1Bytes, _ := json.Marshal(&orgUser1)
	_, err = RegisterUser(stub, org1Caller, []string{string(orgUser1Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterUser to succeed")
	mstub.MockTransactionEnd("t123")

	// put orgUser1 in org1
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = PutUserInOrg(stub, org1Caller, []string{orgUser1.ID, org1.ID, "false"})
	test_utils.AssertTrue(t, err == nil, "Expected PutUserInOrg to succeed")
	mstub.MockTransactionEnd("t123")

	// give org admin permission
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = AddPermissionOrgAdmin(stub, org1Caller, []string{orgUser1.ID, org1.ID})
	test_utils.AssertTrue(t, err == nil, "Expected AddPermissionOrgAdmin to succeed")
	mstub.MockTransactionEnd("t123")

	// get patient enrollments as orgUser1 (org admin)
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	orgUser1Caller, _ := user_mgmt.GetUserData(stub, org1Caller, orgUser1.ID, true, true)
	enrollmentsBytes, err = GetPatientEnrollments(stub, orgUser1Caller, []string{"patient1"})
	test_utils.AssertTrue(t, err == nil, "Expected GetPatientEnrollments to succeed")
	_ = json.Unmarshal(enrollmentsBytes, &enrollments)
	test_utils.AssertTrue(t, enrollments[0].ServiceID == "service1", "Got service 1 correctly")
	test_utils.AssertTrue(t, enrollments[1].ServiceID == "service2", "Got service 2 correctly")
	mstub.MockTransactionEnd("t123")

	// remove org admin permission
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = RemovePermissionOrgAdmin(stub, org1Caller, []string{orgUser1.ID, org1.ID})
	test_utils.AssertTrue(t, err == nil, "Expected RemovePermissionOrgAdmin to succeed")
	mstub.MockTransactionEnd("t123")

	// attempt to get enrollments without org admin permission
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	orgUser1Caller, _ = user_mgmt.GetUserData(stub, org1Caller, orgUser1.ID, true, true)
	enrollmentsBytes, err = GetPatientEnrollments(stub, orgUser1Caller, []string{"patient1"})
	test_utils.AssertTrue(t, err == nil, "Expected GetPatientEnrollments to succeed")
	_ = json.Unmarshal(enrollmentsBytes, &enrollments)
	test_utils.AssertTrue(t, len(enrollments) == 0, "Expected 0 enrollments")
	mstub.MockTransactionEnd("t123")

	// give service admin permission
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = AddPermissionServiceAdmin(stub, org1Caller, []string{orgUser1.ID, "service1"})
	test_utils.AssertTrue(t, err == nil, "Expected AddPermissionServiceAdmin to succeed")
	mstub.MockTransactionEnd("t123")

	// get enrollments, as service admin, should only get 1 enrollment back
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	orgUser1Caller, _ = user_mgmt.GetUserData(stub, org1Caller, orgUser1.ID, true, true)
	enrollmentsBytes, err = GetPatientEnrollments(stub, orgUser1Caller, []string{"patient1"})
	test_utils.AssertTrue(t, err == nil, "Expected GetPatientEnrollments to succeed")
	_ = json.Unmarshal(enrollmentsBytes, &enrollments)
	test_utils.AssertTrue(t, len(enrollments) == 1, "Expected 1 enrollment")
	test_utils.AssertTrue(t, enrollments[0].ServiceID == "service1", "Got service 1 correctly")
	mstub.MockTransactionEnd("t123")

	// remove service admin permission
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = RemovePermissionServiceAdmin(stub, org1Caller, []string{orgUser1.ID, "service1"})
	test_utils.AssertTrue(t, err == nil, "Expected AddPermissionServiceAdmin to succeed")
	mstub.MockTransactionEnd("t123")

	// attempt to get enrollments without service admin permission
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	orgUser1Caller, _ = user_mgmt.GetUserData(stub, org1Caller, orgUser1.ID, true, true)
	enrollmentsBytes, err = GetPatientEnrollments(stub, orgUser1Caller, []string{"patient1"})
	test_utils.AssertTrue(t, err == nil, "Expected GetPatientEnrollments to succeed")
	_ = json.Unmarshal(enrollmentsBytes, &enrollments)
	test_utils.AssertTrue(t, len(enrollments) == 0, "Expected 0 enrollments")
	mstub.MockTransactionEnd("t123")

	// get enrollments, as patient1
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	enrollmentsBytes, err = GetPatientEnrollments(stub, patient1, []string{"patient1"})
	test_utils.AssertTrue(t, err == nil, "Expected GetPatientEnrollments to succeed")
	_ = json.Unmarshal(enrollmentsBytes, &enrollments)
	test_utils.AssertTrue(t, len(enrollments) == 2, "Expected 2 enrollments")
	test_utils.AssertTrue(t, enrollments[0].ServiceID == "service1", "Got service 1 correctly")
	test_utils.AssertTrue(t, enrollments[1].ServiceID == "service2", "Got service 2 correctly")
	mstub.MockTransactionEnd("t123")

	// get enrollments, as patient2
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	enrollmentsBytes, err = GetPatientEnrollments(stub, patient2, []string{"patient1"})
	test_utils.AssertTrue(t, err == nil, "Expected GetPatientEnrollments to succeed")
	_ = json.Unmarshal(enrollmentsBytes, &enrollments)
	test_utils.AssertTrue(t, len(enrollments) == 0, "Expected 0 enrollments")
	mstub.MockTransactionEnd("t123")
}

func TestGetServiceEnrollments(t *testing.T) {
	logger.SetLevel(shim.LogDebug)
	logger.Info("TestGetServiceEnrollments function called")

	// create a MockStub
	mstub := SetupIndexesAndGetStub(t)

	// register admin user
	mstub.MockTransactionStart("t123")
	stub := cached_stub.NewCachedStub(mstub)
	init_common.Init(stub)
	err := SetupEnrollmentIndex(stub)
	test_utils.AssertTrue(t, err == nil, "Expected SetupEnrollmentIndex to succeed")
	systemAdmin := test_utils.CreateTestUser("systemAdmin")
	systemAdmin.Role = SOLUTION_ROLE_SYSTEM
	systemAdminBytes, _ := json.Marshal(&systemAdmin)
	_, err = RegisterUser(stub, systemAdmin, []string{string(systemAdminBytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterUser to succeed")
	mstub.MockTransactionEnd("t123")

	// Register system datatypes
	mstub.MockTransactionStart("init")
	stub = cached_stub.NewCachedStub(mstub)
	RegisterSystemDatatypeTest(t, stub, systemAdmin)
	mstub.MockTransactionEnd("init")

	// register org
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	org1 := test_utils.CreateTestGroup("org1")
	org1Bytes, _ := json.Marshal(&org1)
	_, err = RegisterOrg(stub, org1, []string{string(org1Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterOrg to succeed")
	mstub.MockTransactionEnd("t123")

	//  register datatype
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	datatype1 := Datatype{DatatypeID: "datatype1", Description: "datatype1"}
	datatype1Bytes, _ := json.Marshal(&datatype1)
	org1Caller, _ := user_mgmt.GetUserData(stub, org1, org1.ID, true, true)
	_, err = RegisterDatatype(stub, org1Caller, []string{string(datatype1Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterDatatype to succeed")
	mstub.MockTransactionEnd("t123")

	//  register service
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub, true, true, true)
	serviceDatatype1 := GenerateServiceDatatypeForTesting("datatype1", "service1", []string{consentOptionWrite, consentOptionRead})
	service1 := GenerateServiceForTesting("service1", "org1", []ServiceDatatype{serviceDatatype1})
	service1Bytes, _ := json.Marshal(&service1)
	_, err = RegisterService(stub, org1Caller, []string{string(service1Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterService to succeed")
	mstub.MockTransactionEnd("t123")

	// register patient1
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	patient1 := test_utils.CreateTestUser("patient1")
	patient1Bytes, _ := json.Marshal(&patient1)
	_, err = user_mgmt.RegisterUser(stub, org1Caller, []string{string(patient1Bytes), "false"})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterUser to succeed")
	mstub.MockTransactionEnd("t123")

	// register patient2
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	patient2 := test_utils.CreateTestUser("patient2")
	patient2Bytes, _ := json.Marshal(&patient2)
	_, err = user_mgmt.RegisterUser(stub, org1Caller, []string{string(patient2Bytes), "false"})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterUser to succeed")
	mstub.MockTransactionEnd("t123")

	// enroll patient1 to service 1 as org1
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	enrollment1 := GenerateEnrollmentTest("patient1", "service1")
	enrollment1Bytes, _ := json.Marshal(&enrollment1)
	enrollmentKey1 := test_utils.GenerateSymKey()
	enrollmentKey1B64 := crypto.EncodeToB64String(enrollmentKey1)
	_, err = EnrollPatient(stub, org1Caller, []string{string(enrollment1Bytes), enrollmentKey1B64})
	test_utils.AssertTrue(t, err == nil, "Expected EnrollPatient to succeed")
	mstub.MockTransactionEnd("t123")

	// enroll patient2 to service 1
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	enrollment2 := GenerateEnrollmentTest("patient2", "service1")
	enrollment2Bytes, _ := json.Marshal(&enrollment2)
	enrollmentKey2 := test_utils.GenerateSymKey()
	enrollmentKey2B64 := crypto.EncodeToB64String(enrollmentKey2)
	_, err = EnrollPatient(stub, org1Caller, []string{string(enrollment2Bytes), enrollmentKey2B64})
	test_utils.AssertTrue(t, err == nil, "Expected EnrollPatient to succeed")
	mstub.MockTransactionEnd("t123")

	// get service's patient enrollments
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	enrollmentsBytes, err := GetServiceEnrollments(stub, org1Caller, []string{"service1"})
	test_utils.AssertTrue(t, err == nil, "Expected GetServiceEnrollments to succeed")
	var enrollments = []Enrollment{}
	_ = json.Unmarshal(enrollmentsBytes, &enrollments)
	test_utils.AssertTrue(t, enrollments[0].UserID == "patient1", "Got patient 1 correctly")
	test_utils.AssertTrue(t, enrollments[1].UserID == "patient2", "Got patient 2 correctly")
	// get enrollments that are inactive
	// expecting none
	enrollmentsBytes, err = GetServiceEnrollments(stub, org1Caller, []string{"service1", "inactive"})
	test_utils.AssertTrue(t, err == nil, "Expected GetServiceEnrollments to succeed")
	_ = json.Unmarshal(enrollmentsBytes, &enrollments)
	test_utils.AssertTrue(t, len(enrollments) == 0, "Expected 0 inactive enrollments")
	mstub.MockTransactionEnd("t123")

	// create an org user
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	orgUser1 := CreateTestSolutionUser("orgUser1")
	orgUser1.Org = "org1"
	orgUser1Bytes, _ := json.Marshal(&orgUser1)
	_, err = RegisterUser(stub, org1Caller, []string{string(orgUser1Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterUser to succeed")
	mstub.MockTransactionEnd("t123")

	// put orgUser1 in org1
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = PutUserInOrg(stub, org1Caller, []string{orgUser1.ID, org1.ID, "false"})
	test_utils.AssertTrue(t, err == nil, "Expected PutUserInOrg to succeed")
	mstub.MockTransactionEnd("t123")

	// give org admin permission
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = AddPermissionOrgAdmin(stub, org1Caller, []string{orgUser1.ID, org1.ID})
	test_utils.AssertTrue(t, err == nil, "Expected AddPermissionOrgAdmin to succeed")
	mstub.MockTransactionEnd("t123")

	// get service's patient enrollments
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	orgUser1Caller, _ := user_mgmt.GetUserData(stub, org1Caller, orgUser1.ID, true, true)
	enrollmentsBytes, err = GetServiceEnrollments(stub, orgUser1Caller, []string{"service1"})
	test_utils.AssertTrue(t, err == nil, "Expected GetServiceEnrollments to succeed")
	_ = json.Unmarshal(enrollmentsBytes, &enrollments)
	test_utils.AssertTrue(t, enrollments[0].UserID == "patient1", "Got patient 1 correctly")
	test_utils.AssertTrue(t, enrollments[1].UserID == "patient2", "Got patient 2 correctly")
	mstub.MockTransactionEnd("t123")

	//  register service2
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub, true, true, true)
	service2 := GenerateServiceForTesting("service2", "org1", []ServiceDatatype{serviceDatatype1})
	service2Bytes, _ := json.Marshal(&service2)
	_, err = RegisterService(stub, org1Caller, []string{string(service2Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterService to succeed")
	mstub.MockTransactionEnd("t123")

	// enroll patient1 to service 2
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	enrollment3 := GenerateEnrollmentTest("patient1", "service2")
	enrollment3Bytes, _ := json.Marshal(&enrollment3)
	enrollmentKey3 := test_utils.GenerateSymKey()
	enrollmentKey3B64 := crypto.EncodeToB64String(enrollmentKey3)
	_, err = EnrollPatient(stub, org1Caller, []string{string(enrollment3Bytes), enrollmentKey3B64})
	test_utils.AssertTrue(t, err == nil, "Expected EnrollPatient to succeed")
	mstub.MockTransactionEnd("t123")

	// enroll patient2 to service2
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	enrollment4 := GenerateEnrollmentTest("patient2", "service2")
	enrollment4Bytes, _ := json.Marshal(&enrollment4)
	enrollmentKey4 := test_utils.GenerateSymKey()
	enrollmentKey4B64 := crypto.EncodeToB64String(enrollmentKey4)
	_, err = EnrollPatient(stub, org1Caller, []string{string(enrollment4Bytes), enrollmentKey4B64})
	test_utils.AssertTrue(t, err == nil, "Expected EnrollPatient to succeed")
	mstub.MockTransactionEnd("t123")

	// get service2's patient enrollments
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	enrollmentsBytes, err = GetServiceEnrollments(stub, org1Caller, []string{"service2"})
	test_utils.AssertTrue(t, err == nil, "Expected GetServiceEnrollments to succeed")
	_ = json.Unmarshal(enrollmentsBytes, &enrollments)
	test_utils.AssertTrue(t, enrollments[0].UserID == "patient1", "Got patient 1 correctly")
	test_utils.AssertTrue(t, enrollments[1].UserID == "patient2", "Got patient 2 correctly")
	test_utils.AssertTrue(t, len(enrollments) == 2, "Expected 0 inactive enrollments")
	mstub.MockTransactionEnd("t123")

	// remove orgUser1's org admin permission
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = RemovePermissionOrgAdmin(stub, org1Caller, []string{orgUser1.ID, org1.ID})
	test_utils.AssertTrue(t, err == nil, "Expected RemovePermissionOrgAdmin to succeed")
	mstub.MockTransactionEnd("t123")

	// get enrollments, should fail
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	orgUser1Caller, _ = user_mgmt.GetUserData(stub, org1Caller, orgUser1.ID, true, true)
	enrollmentsBytes, err = GetServiceEnrollments(stub, orgUser1Caller, []string{"service1"})
	test_utils.AssertTrue(t, err == nil, "Expected GetPatientEnrollments to succeed")
	_ = json.Unmarshal(enrollmentsBytes, &enrollments)
	test_utils.AssertTrue(t, len(enrollments) == 0, "Expected 0 enrollments")
	enrollmentsBytes, err = GetServiceEnrollments(stub, orgUser1Caller, []string{"service2"})
	test_utils.AssertTrue(t, err == nil, "Expected GetPatientEnrollments to succeed")
	_ = json.Unmarshal(enrollmentsBytes, &enrollments)
	test_utils.AssertTrue(t, len(enrollments) == 0, "Expected 0 enrollments")
	mstub.MockTransactionEnd("t123")

	// give service admin permission
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = AddPermissionServiceAdmin(stub, org1Caller, []string{orgUser1.ID, "service1"})
	test_utils.AssertTrue(t, err == nil, "Expected AddPermissionServiceAdmin to succeed")
	mstub.MockTransactionEnd("t123")

	// get enrollments for service1, should pass
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	orgUser1Caller, _ = user_mgmt.GetUserData(stub, org1Caller, orgUser1.ID, true, true)
	enrollmentsBytes, err = GetServiceEnrollments(stub, orgUser1Caller, []string{"service1"})
	test_utils.AssertTrue(t, err == nil, "Expected GetPatientEnrollments to succeed")
	_ = json.Unmarshal(enrollmentsBytes, &enrollments)
	test_utils.AssertTrue(t, len(enrollments) == 2, "Got service enrollments")
	// get enrollments for service2, should fail
	enrollmentsBytes, err = GetServiceEnrollments(stub, orgUser1Caller, []string{"service2"})
	test_utils.AssertTrue(t, err == nil, "Expected GetPatientEnrollments to succeed")
	_ = json.Unmarshal(enrollmentsBytes, &enrollments)
	test_utils.AssertTrue(t, len(enrollments) == 0, "Expected 0 enrollments")
	mstub.MockTransactionEnd("t123")
}
