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
	"common/bchcls/consent_mgmt"
	"common/bchcls/crypto"
	"common/bchcls/data_model"
	"common/bchcls/datastore/datastore_manager"
	"common/bchcls/init_common"
	"common/bchcls/test_utils"
	"common/bchcls/user_access_ctrl"
	"common/bchcls/user_mgmt"
	"net/url"

	"encoding/json"
	"strconv"
	"testing"
	"time"

	"github.com/hyperledger/fabric/core/chaincode/shim"
)

func GeneratePatientData(Owner string, Datatype string, Service string) OwnerData {
	patientData := OwnerData{}
	patientData.Owner = Owner
	patientData.Datatype = Datatype
	patientData.Service = Service
	patientData.Timestamp = time.Now().Unix()
	data := make(map[string]string)
	data["age"] = "23"
	data["address"] = "123 park street"
	patientData.Data = data

	return patientData
}

func GenerateOwnerData(Owner string, Datatype string) OwnerData {
	ownerData := OwnerData{}
	ownerData.Owner = Owner
	ownerData.Datatype = Datatype
	ownerData.Service = Owner
	ownerData.Timestamp = time.Now().Unix()
	data := make(map[string]string)
	data["tax code"] = "3456"
	data["address"] = "567 binney street"
	ownerData.Data = data

	return ownerData
}

func TestUploadUserData(t *testing.T) {
	logger.SetLevel(shim.LogDebug)
	logger.Info("TestUploadUserData function called")

	// create a MockStub
	mstub := SetupIndexesAndGetStub(t)

	// register admin user
	mstub.MockTransactionStart("t123")
	stub := cached_stub.NewCachedStub(mstub)
	init_common.Init(stub)
	err := SetupDataIndex(stub)
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

	// register patient
	mstub.MockTransactionStart("4")
	stub = cached_stub.NewCachedStub(mstub)
	patient1 := test_utils.CreateTestUser("patient1")
	patient1Bytes, _ := json.Marshal(&patient1)
	_, err = user_mgmt.RegisterUser(stub, org1Caller, []string{string(patient1Bytes), "false"})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterUser to succeed")
	mstub.MockTransactionEnd("4")

	// enroll patient
	mstub.MockTransactionStart("5")
	stub = cached_stub.NewCachedStub(mstub)
	enrollment1 := GenerateEnrollmentTest("patient1", "service1")
	enrollment1Bytes, _ := json.Marshal(&enrollment1)
	enrollmentKey1 := test_utils.GenerateSymKey()
	enrollmentKey1B64 := crypto.EncodeToB64String(enrollmentKey1)
	_, err = EnrollPatient(stub, org1Caller, []string{string(enrollment1Bytes), enrollmentKey1B64})
	test_utils.AssertTrue(t, err == nil, "Expected EnrollPatient to succeed")
	mstub.MockTransactionEnd("5")

	// get enrollment back to make sure it was enrolled
	mstub.MockTransactionStart("6")
	stub = cached_stub.NewCachedStub(mstub)
	serviceSubgroup, _ := user_mgmt.GetUserData(stub, org1Caller, "service1", true, true)
	enrollmentResult, err := GetEnrollmentInternal(stub, serviceSubgroup, "patient1", "service1")
	test_utils.AssertTrue(t, err == nil, "Expected GetEnrollmentInternal to succeed")
	test_utils.AssertTrue(t, enrollmentResult.ServiceName == "service1 Name", "Got service name correctly")
	test_utils.AssertTrue(t, enrollmentResult.Status == "active", "Expected enrollment status to be active")
	mstub.MockTransactionEnd("6")

	// patient gives consent
	mstub.MockTransactionStart("7")
	stub = cached_stub.NewCachedStub(mstub, true, true, true)
	consent := Consent{}
	consent.Owner = "patient1"
	consent.Service = "service1"
	consent.Target = "service1"
	consent.Datatype = "datatype1"
	consent.Option = []string{consentOptionWrite}
	consent.Timestamp = time.Now().Unix()
	consent.Expiration = 0
	consentBytes, _ := json.Marshal(&consent)
	consentKey := test_utils.GenerateSymKey()
	consentKeyB64 := crypto.EncodeToB64String(consentKey)
	patient1Caller, err := user_mgmt.GetUserData(stub, patient1, "patient1", true, true)
	_, err = PutConsentPatientData(stub, patient1Caller, []string{string(consentBytes), consentKeyB64})
	test_utils.AssertTrue(t, err == nil, "Expected PutConsentPatientData to succeed")
	mstub.MockTransactionEnd("7")

	// get consent back to make sure it was added
	mstub.MockTransactionStart("7")
	stub = cached_stub.NewCachedStub(mstub)
	consentResultBytes, err := GetConsent(stub, patient1Caller, []string{"patient1", "service1", "datatype1"})
	test_utils.AssertTrue(t, err == nil, "Expected GetConsent to succeed")
	test_utils.AssertTrue(t, consentResultBytes != nil, "Expected GetConsent to succeed")
	consentResult := Consent{}
	json.Unmarshal(consentResultBytes, &consentResult)
	test_utils.AssertTrue(t, consentResult.Owner == "patient1", "Got consent owner correctly")
	test_utils.AssertTrue(t, consentResult.Target == "service1", "Got consent service correctly")
	mstub.MockTransactionEnd("7")

	// upload patient data
	mstub.MockTransactionStart("8")
	stub = cached_stub.NewCachedStub(mstub)
	patientData := GeneratePatientData("patient1", "datatype1", "service1")
	patientDataBytes, _ := json.Marshal(&patientData)
	dataKey := test_utils.GenerateSymKey()
	dataKeyB64 := crypto.EncodeToB64String(dataKey)
	args := []string{string(patientDataBytes), dataKeyB64}
	_, err = UploadUserData(stub, serviceSubgroup, args)
	test_utils.AssertTrue(t, err == nil, "Expected UploadUserData to succeed")
	mstub.MockTransactionEnd("8")

	// upload patient data again
	mstub.MockTransactionStart("9")
	stub = cached_stub.NewCachedStub(mstub)
	patientData = GeneratePatientData("patient1", "datatype1", "service1")
	patientData.Timestamp++
	patientDataBytes, _ = json.Marshal(&patientData)
	args = []string{string(patientDataBytes)}
	_, err = UploadUserData(stub, serviceSubgroup, args)
	test_utils.AssertTrue(t, err == nil, "Expected UploadUserData to succeed")
	mstub.MockTransactionEnd("9")

	// upload patient data as org admin
	mstub.MockTransactionStart("10")
	stub = cached_stub.NewCachedStub(mstub)
	patientData = GeneratePatientData("patient1", "datatype1", "service1")
	patientData.Timestamp += 2
	patientDataBytes, _ = json.Marshal(&patientData)
	args = []string{string(patientDataBytes)}
	_, err = UploadUserData(stub, org1Caller, args)
	test_utils.AssertTrue(t, err == nil, "Expected UploadUserData to succeed")
	mstub.MockTransactionEnd("10")

	// create an org user
	mstub.MockTransactionStart("11")
	stub = cached_stub.NewCachedStub(mstub)
	orgUser1 := CreateTestSolutionUser("orgUser1")
	orgUser1.Org = "org1"
	orgUser1Bytes, _ := json.Marshal(&orgUser1)
	_, err = RegisterUser(stub, org1Caller, []string{string(orgUser1Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterUser to succeed")
	mstub.MockTransactionEnd("11")

	// put orgUser1 in org1
	mstub.MockTransactionStart("12")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = PutUserInOrg(stub, org1Caller, []string{orgUser1.ID, org1.ID, "false"})
	test_utils.AssertTrue(t, err == nil, "Expected PutUserInOrg to succeed")
	mstub.MockTransactionEnd("12")

	// give org admin permission
	mstub.MockTransactionStart("13")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = AddPermissionOrgAdmin(stub, org1Caller, []string{orgUser1.ID, org1.ID})
	test_utils.AssertTrue(t, err == nil, "Expected AddPermissionOrgAdmin to succeed")
	mstub.MockTransactionEnd("13")

	// upload patient data as org user (org admin)
	mstub.MockTransactionStart("14")
	stub = cached_stub.NewCachedStub(mstub)
	patientData = GeneratePatientData("patient1", "datatype1", "service1")
	patientData.Timestamp += 3
	patientDataBytes, _ = json.Marshal(&patientData)
	args = []string{string(patientDataBytes)}
	orgUser1Caller, _ := user_mgmt.GetUserData(stub, org1Caller, orgUser1.ID, true, true)
	_, err = UploadUserData(stub, orgUser1Caller, args)
	test_utils.AssertTrue(t, err == nil, "Expected UploadUserData to succeed")
	mstub.MockTransactionEnd("14")

	// create another org user
	mstub.MockTransactionStart("15")
	stub = cached_stub.NewCachedStub(mstub)
	orgUser2 := CreateTestSolutionUser("orgUser2")
	orgUser2.Org = "org1"
	orgUser2Bytes, _ := json.Marshal(&orgUser2)
	_, err = RegisterUser(stub, org1Caller, []string{string(orgUser2Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterUser to succeed")
	mstub.MockTransactionEnd("15")

	// put orgUser2 in org1
	mstub.MockTransactionStart("16")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = PutUserInOrg(stub, org1Caller, []string{orgUser2.ID, org1.ID, "false"})
	test_utils.AssertTrue(t, err == nil, "Expected PutUserInOrg to succeed")
	mstub.MockTransactionEnd("16")

	// give service admin permission
	mstub.MockTransactionStart("17")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = AddPermissionServiceAdmin(stub, org1Caller, []string{orgUser2.ID, "service1"})
	test_utils.AssertTrue(t, err == nil, "Expected AddPermissionServiceAdmin to succeed")
	mstub.MockTransactionEnd("17")

	// upload patient data as org user (service admin)
	mstub.MockTransactionStart("18")
	stub = cached_stub.NewCachedStub(mstub)
	patientData = GeneratePatientData("patient1", "datatype1", "service1")
	patientData.Timestamp += 4
	patientDataBytes, _ = json.Marshal(&patientData)
	args = []string{string(patientDataBytes)}
	orgUser2Caller, _ := user_mgmt.GetUserData(stub, org1Caller, orgUser2.ID, true, true)
	_, err = UploadUserData(stub, orgUser2Caller, args)
	test_utils.AssertTrue(t, err == nil, "Expected UploadUserData to succeed")
	mstub.MockTransactionEnd("18")

	// register another patient and upload data to the same service datatype pair
	mstub.MockTransactionStart("19")
	stub = cached_stub.NewCachedStub(mstub)
	patient2 := test_utils.CreateTestUser("patient2")
	patient2Bytes, _ := json.Marshal(&patient2)
	_, err = user_mgmt.RegisterUser(stub, org1Caller, []string{string(patient2Bytes), "false"})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterUser to succeed")
	mstub.MockTransactionEnd("19")

	// enroll patient
	mstub.MockTransactionStart("20")
	stub = cached_stub.NewCachedStub(mstub)
	enrollment2 := GenerateEnrollmentTest("patient2", "service1")
	enrollment2Bytes, _ := json.Marshal(&enrollment2)
	enrollmentKey2 := test_utils.GenerateSymKey()
	enrollmentKey2B64 := crypto.EncodeToB64String(enrollmentKey2)
	_, err = EnrollPatient(stub, serviceSubgroup, []string{string(enrollment2Bytes), enrollmentKey2B64})
	test_utils.AssertTrue(t, err == nil, "Expected EnrollPatient to succeed")
	mstub.MockTransactionEnd("20")

	// get enrollment back to make sure it was enrolled
	mstub.MockTransactionStart("21")
	stub = cached_stub.NewCachedStub(mstub)
	enrollmentResult, err = GetEnrollmentInternal(stub, serviceSubgroup, "patient2", "service1")
	test_utils.AssertTrue(t, err == nil, "Expected GetEnrollmentInternal to succeed")
	test_utils.AssertTrue(t, enrollmentResult.ServiceName == "service1 Name", "Got service name correctly")
	test_utils.AssertTrue(t, enrollmentResult.Status == "active", "Expected enrollment status to be active")
	mstub.MockTransactionEnd("21")

	// patient gives consent
	mstub.MockTransactionStart("22")
	stub = cached_stub.NewCachedStub(mstub, true, true, true)
	consent.Owner = "patient2"
	consent.Service = "service1"
	consent.Target = "service1"
	consent.Datatype = "datatype1"
	consent.Option = []string{consentOptionWrite}
	consent.Timestamp = time.Now().Unix()
	consent.Expiration = 0
	consentBytes, _ = json.Marshal(&consent)
	consentKey = test_utils.GenerateSymKey()
	consentKeyB64 = crypto.EncodeToB64String(consentKey)
	patient2Caller, err := user_mgmt.GetUserData(stub, patient2, "patient2", true, true)
	_, err = PutConsentPatientData(stub, patient2Caller, []string{string(consentBytes), consentKeyB64})
	test_utils.AssertTrue(t, err == nil, "Expected PutConsentPatientData to succeed")
	mstub.MockTransactionEnd("22")

	// upload patient data
	mstub.MockTransactionStart("23")
	stub = cached_stub.NewCachedStub(mstub)
	patientData = GeneratePatientData("patient2", "datatype1", "service1")
	patientDataBytes, _ = json.Marshal(&patientData)
	dataKey = test_utils.GenerateSymKey()
	dataKeyB64 = crypto.EncodeToB64String(dataKey)
	args = []string{string(patientDataBytes), dataKeyB64}
	_, err = UploadUserData(stub, serviceSubgroup, args)
	test_utils.AssertTrue(t, err == nil, "Expected UploadUserData to succeed")
	mstub.MockTransactionEnd("23")

	// patient update consent to DENY
	mstub.MockTransactionStart("24")
	stub = cached_stub.NewCachedStub(mstub, true, true, true)
	consent = Consent{}
	consent.Owner = "patient1"
	consent.Service = "service1"
	consent.Target = "service1"
	consent.Datatype = "datatype1"
	consent.Option = []string{consentOptionDeny}
	consent.Timestamp = time.Now().Unix()
	consent.Expiration = 0
	consentBytes, _ = json.Marshal(&consent)
	_, err = PutConsentPatientData(stub, patient1Caller, []string{string(consentBytes)})
	test_utils.AssertTrue(t, err == nil, "Expected PutConsentPatientData to succeed")
	mstub.MockTransactionEnd("24")

	// try upload patient data, should fail
	mstub.MockTransactionStart("25")
	stub = cached_stub.NewCachedStub(mstub)
	patientData = GeneratePatientData("patient1", "datatype1", "service1")
	patientData.Timestamp++
	patientDataBytes, _ = json.Marshal(&patientData)
	args = []string{string(patientDataBytes)}
	_, err = UploadUserData(stub, serviceSubgroup, args)
	test_utils.AssertTrue(t, err != nil, "Expected UploadUserData to fail")
	mstub.MockTransactionEnd("25")
}

func TestUploadUserData_OffChain(t *testing.T) {
	logger.SetLevel(shim.LogDebug)
	logger.Info("TestUploadUserData_OffChain function called")

	// create a MockStub
	mstub := SetupIndexesAndGetStub(t)

	// register admin user
	mstub.MockTransactionStart("t123")
	stub := cached_stub.NewCachedStub(mstub)
	init_common.Init(stub)
	err := SetupDataIndex(stub)
	systemAdmin := test_utils.CreateTestUser("systemAdmin")
	systemAdmin.Role = SOLUTION_ROLE_SYSTEM
	systemAdminBytes, _ := json.Marshal(&systemAdmin)
	_, err = RegisterUser(stub, systemAdmin, []string{string(systemAdminBytes)})
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

	// register patient
	mstub.MockTransactionStart("4")
	stub = cached_stub.NewCachedStub(mstub)
	patient1 := test_utils.CreateTestUser("patient1")
	patient1Bytes, _ := json.Marshal(&patient1)
	_, err = user_mgmt.RegisterUser(stub, org1Caller, []string{string(patient1Bytes), "false"})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterUser to succeed")
	mstub.MockTransactionEnd("4")

	// enroll patient
	mstub.MockTransactionStart("5")
	stub = cached_stub.NewCachedStub(mstub)
	enrollment1 := GenerateEnrollmentTest("patient1", "service1")
	enrollment1Bytes, _ := json.Marshal(&enrollment1)
	enrollmentKey1 := test_utils.GenerateSymKey()
	enrollmentKey1B64 := crypto.EncodeToB64String(enrollmentKey1)
	_, err = EnrollPatient(stub, org1Caller, []string{string(enrollment1Bytes), enrollmentKey1B64})
	test_utils.AssertTrue(t, err == nil, "Expected EnrollPatient to succeed")
	mstub.MockTransactionEnd("5")

	// patient gives consent
	mstub.MockTransactionStart("7")
	stub = cached_stub.NewCachedStub(mstub, true, true, true)
	consent := Consent{}
	consent.Owner = "patient1"
	consent.Service = "service1"
	consent.Target = "service1"
	consent.Datatype = "datatype1"
	consent.Option = []string{consentOptionWrite}
	consent.Timestamp = time.Now().Unix()
	consent.Expiration = 0
	consentBytes, _ := json.Marshal(&consent)
	consentKey := test_utils.GenerateSymKey()
	consentKeyB64 := crypto.EncodeToB64String(consentKey)
	patient1Caller, err := user_mgmt.GetUserData(stub, patient1, "patient1", true, true)
	_, err = PutConsentPatientData(stub, patient1Caller, []string{string(consentBytes), consentKeyB64})
	test_utils.AssertTrue(t, err == nil, "Expected PutConsentPatientData to succeed")
	mstub.MockTransactionEnd("7")

	// upload patient data
	mstub.MockTransactionStart("8")
	stub = cached_stub.NewCachedStub(mstub)
	patientData := GeneratePatientData("patient1", "datatype1", "service1")
	patientDataBytes, _ := json.Marshal(&patientData)
	dataKey := test_utils.GenerateSymKey()
	dataKeyB64 := crypto.EncodeToB64String(dataKey)
	args := []string{string(patientDataBytes), dataKeyB64}
	service1Subgroup, _ := user_mgmt.GetUserData(stub, org1Caller, "service1", true, true)
	_, err = UploadUserData(stub, service1Subgroup, args)
	test_utils.AssertTrue(t, err == nil, "Expected UploadUserData to succeed")
	mstub.MockTransactionEnd("8")

	// get user data to make sure it was registered
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	ownerDatas, err := GetDataInternal(stub, service1Subgroup, []string{"owner", "datatype", "timestamp"}, []string{patientData.Owner, patientData.Datatype}, []string{patientData.Owner, patientData.Datatype}, 1)
	test_utils.AssertTrue(t, err == nil, "Expected GetDataInternal to succeed")
	test_utils.AssertTrue(t, len(ownerDatas) == 1, "Expected 1 user data")
	test_utils.AssertTrue(t, ownerDatas[0].Owner == patientData.Owner, "Expected Owner")
	mstub.MockTransactionEnd("t123")

	// verify data access for service
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	assetManager := asset_mgmt.GetAssetManager(stub, service1Subgroup)
	userDataID := GetPatientDataID(patientData.Owner, patientData.Datatype, patientData.Timestamp)
	userDataAssetID := asset_mgmt.GetAssetId(OwnerDataNamespace, userDataID)
	keyPath, err := GetKeyPath(stub, service1Subgroup, userDataAssetID)
	test_utils.AssertTrue(t, err == nil, "Expected GetKeyPath user data to succeed")
	userDataAssetKey, err := assetManager.GetAssetKey(userDataAssetID, keyPath)
	test_utils.AssertTrue(t, err == nil, "Expected GetAssetKey to succeed")
	userDataAsset, err := assetManager.GetAsset(userDataAssetID, userDataAssetKey)
	test_utils.AssertTrue(t, err == nil, "Expected GetAsset to succeed")
	mstub.MockTransactionEnd("t123")

	// service has access to private data
	userDataPrivateData := OwnerData{}
	json.Unmarshal(userDataAsset.PrivateData, &userDataPrivateData)
	test_utils.AssertTrue(t, userDataPrivateData.DataID != "", "Expected DataID")
	test_utils.AssertTrue(t, userDataPrivateData.Owner != "", "Expected Owner")
	test_utils.AssertTrue(t, userDataPrivateData.Service != "", "Expected Service")
	test_utils.AssertTrue(t, userDataPrivateData.Datatype != "", "Expected Datatype")
	test_utils.AssertTrue(t, userDataPrivateData.Timestamp > 0, "Expected Timestamp")
	test_utils.AssertTrue(t, userDataPrivateData.Data != nil, "Expected Data")

	// verify data access for patient
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	assetManager = asset_mgmt.GetAssetManager(stub, patient1)
	keyPath, err = GetKeyPath(stub, patient1, userDataAssetID)
	test_utils.AssertTrue(t, err == nil, "Expected GetKeyPath user data to succeed")
	userDataAssetKey, err = assetManager.GetAssetKey(userDataAssetID, keyPath)
	test_utils.AssertTrue(t, err == nil, "Expected GetAssetKey to succeed")
	userDataAsset, err = assetManager.GetAsset(userDataAssetID, userDataAssetKey)
	test_utils.AssertTrue(t, err == nil, "Expected GetAsset to succeed")
	mstub.MockTransactionEnd("t123")

	// patient has access to private data
	userDataPrivateData = OwnerData{}
	json.Unmarshal(userDataAsset.PrivateData, &userDataPrivateData)
	test_utils.AssertTrue(t, userDataPrivateData.DataID != "", "Expected DataID")
	test_utils.AssertTrue(t, userDataPrivateData.Owner != "", "Expected Owner")
	test_utils.AssertTrue(t, userDataPrivateData.Service != "", "Expected Service")
	test_utils.AssertTrue(t, userDataPrivateData.Datatype != "", "Expected Datatype")
	test_utils.AssertTrue(t, userDataPrivateData.Timestamp > 0, "Expected Timestamp")
	test_utils.AssertTrue(t, userDataPrivateData.Data != nil, "Expected Data")

	// register unrelatedUser
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	unrelatedUser := test_utils.CreateTestUser("unrelatedUser")
	unrelatedUserBytes, _ := json.Marshal(&unrelatedUser)
	_, err = user_mgmt.RegisterUser(stub, unrelatedUser, []string{string(unrelatedUserBytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterUser to succeed")
	mstub.MockTransactionEnd("t123")

	// verify data access for unrelatedUser
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	assetManager = asset_mgmt.GetAssetManager(stub, unrelatedUser)
	keyPath, err = GetKeyPath(stub, unrelatedUser, userDataAssetID)
	test_utils.AssertTrue(t, err == nil, "Expected GetKeyPath user data to succeed")
	_, err = assetManager.GetAssetKey(userDataAssetID, keyPath)
	test_utils.AssertTrue(t, err != nil, "Expected GetAssetKey to fail")
	userDataAsset, err = assetManager.GetAsset(userDataAssetID, data_model.Key{})
	test_utils.AssertTrue(t, err == nil, "Expected GetAsset to succeed")
	mstub.MockTransactionEnd("t123")

	// unrelatedUser has no access to private data
	userDataPrivateData = OwnerData{}
	json.Unmarshal(userDataAsset.PrivateData, &userDataPrivateData)
	test_utils.AssertTrue(t, userDataPrivateData.DataID == "", "Expected no DataID")
	test_utils.AssertTrue(t, userDataPrivateData.Owner == "", "Expected no Owner")
	test_utils.AssertTrue(t, userDataPrivateData.Service == "", "Expected no Service")
	test_utils.AssertTrue(t, userDataPrivateData.Datatype == "", "Expected no Datatype")
	test_utils.AssertTrue(t, userDataPrivateData.Timestamp == 0, "Expected no Timestamp")
	test_utils.AssertTrue(t, userDataPrivateData.Data == nil, "Expected no Data")

	// remove datastore connection
	mstub.MockTransactionStart("t123")
	err = datastore_manager.DeleteDatastoreConnection(stub, systemAdmin, datastoreConnectionID)
	test_utils.AssertTrue(t, err == nil, "Expected DeleteDatastoreConnection to succeed")
	mstub.MockTransactionEnd("t123")

	// attempt to get user data
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	ownerDatas, err = GetDataInternal(stub, service1Subgroup, []string{"owner", "datatype", "timestamp"}, []string{patientData.Owner, patientData.Datatype}, []string{patientData.Owner, patientData.Datatype}, 1)
	test_utils.AssertTrue(t, err == nil, "Expected GetDataInternal to succeed")
	test_utils.AssertTrue(t, len(ownerDatas) == 0, "Expected no user data")
	mstub.MockTransactionEnd("t123")
}

func TestDownloadUserData(t *testing.T) {
	logger.SetLevel(shim.LogDebug)
	logger.Info("TestDownloadUserData function called")

	// create a MockStub
	mstub := SetupIndexesAndGetStub(t)

	// register admin user
	mstub.MockTransactionStart("t123")
	stub := cached_stub.NewCachedStub(mstub)
	init_common.Init(stub)
	err := SetupDataIndex(stub)
	err = SetupAuditPermissionIndex(stub)
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

	// register auditor
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	auditor := CreateTestSolutionUser("auditor1")
	auditor.Role = SOLUTION_ROLE_AUDIT
	auditorBytes, _ := json.Marshal(&auditor)
	_, err = RegisterUser(stub, systemAdmin, []string{string(auditorBytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterUser to succeed")
	mstub.MockTransactionEnd("t1")

	// give auditor audit permission
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	auditPermissionAssetKey := test_utils.GenerateSymKey()
	auditPermissionAssetKeyB64 := crypto.EncodeToB64String(auditPermissionAssetKey)
	_, err = AddAuditorPermission(stub, org1Caller, []string{auditor.ID, "service1", auditPermissionAssetKeyB64})
	test_utils.AssertTrue(t, err == nil, "Expected AddAuditorPermission to succeed")
	mstub.MockTransactionEnd("t1")

	// register patient
	mstub.MockTransactionStart("4")
	stub = cached_stub.NewCachedStub(mstub)
	patient1 := test_utils.CreateTestUser("patient1")
	patient1Bytes, _ := json.Marshal(&patient1)
	_, err = user_mgmt.RegisterUser(stub, org1Caller, []string{string(patient1Bytes), "false"})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterUser to succeed")
	mstub.MockTransactionEnd("4")

	// enroll patient
	mstub.MockTransactionStart("5")
	stub = cached_stub.NewCachedStub(mstub)
	enrollment1 := GenerateEnrollmentTest("patient1", "service1")
	enrollment1Bytes, _ := json.Marshal(&enrollment1)
	enrollmentKey1 := test_utils.GenerateSymKey()
	enrollmentKey1B64 := crypto.EncodeToB64String(enrollmentKey1)
	_, err = EnrollPatient(stub, org1Caller, []string{string(enrollment1Bytes), enrollmentKey1B64})
	test_utils.AssertTrue(t, err == nil, "Expected EnrollPatient to succeed")
	mstub.MockTransactionEnd("5")

	// patient give consent
	mstub.MockTransactionStart("6")
	stub = cached_stub.NewCachedStub(mstub, true, true, true)
	consent := Consent{}
	consent.Owner = "patient1"
	consent.Service = "service1"
	consent.Target = "service1"
	consent.Datatype = "datatype1"
	consent.Option = []string{consentOptionWrite, consentOptionRead}
	consent.Timestamp = time.Now().Unix()
	consent.Expiration = 0
	consentBytes, _ := json.Marshal(&consent)
	consentKey := test_utils.GenerateSymKey()
	consentKeyB64 := crypto.EncodeToB64String(consentKey)
	patient1Caller, err := user_mgmt.GetUserData(stub, patient1, "patient1", true, true)
	_, err = PutConsentPatientData(stub, patient1Caller, []string{string(consentBytes), consentKeyB64})
	test_utils.AssertTrue(t, err == nil, "Expected PutConsentPatientData to succeed")
	mstub.MockTransactionEnd("6")

	// get consent back to make sure it was added
	mstub.MockTransactionStart("7")
	stub = cached_stub.NewCachedStub(mstub)
	consentResultBytes, err := GetConsent(stub, patient1Caller, []string{"patient1", "service1", "datatype1"})
	test_utils.AssertTrue(t, err == nil, "Expected GetConsent to succeed")
	test_utils.AssertTrue(t, consentResultBytes != nil, "Expected GetConsent to succeed")
	consentResult := Consent{}
	json.Unmarshal(consentResultBytes, &consentResult)
	test_utils.AssertTrue(t, consentResult.Owner == "patient1", "Got consent owner correctly")
	test_utils.AssertTrue(t, consentResult.Target == "service1", "Got consent service correctly")
	mstub.MockTransactionEnd("7")

	// upload patient data
	mstub.MockTransactionStart("8")
	stub = cached_stub.NewCachedStub(mstub)
	service1Subgroup, _ := user_mgmt.GetUserData(stub, org1Caller, "service1", true, true)
	patientData := GeneratePatientData("patient1", "datatype1", "service1")
	patientDataBytes, _ := json.Marshal(&patientData)
	dataKey := test_utils.GenerateSymKey()
	dataKeyB64 := crypto.EncodeToB64String(dataKey)
	args := []string{string(patientDataBytes), dataKeyB64}
	_, err = UploadUserData(stub, service1Subgroup, args)
	test_utils.AssertTrue(t, err == nil, "Expected UploadUserData to succeed")
	mstub.MockTransactionEnd("8")

	// Check access to patient data
	mstub.MockTransactionStart("9")
	stub = cached_stub.NewCachedStub(mstub)
	userAccessManager := user_access_ctrl.GetUserAccessManager(stub, service1Subgroup)
	accessControl := data_model.AccessControl{}
	accessControl.UserId = service1Subgroup.ID
	accessControl.AssetId = asset_mgmt.GetAssetId(OwnerDataNamespace, GetPatientDataID("patient1", "datatype1", patientData.Timestamp))
	accessControl.Access = consent_mgmt.ACCESS_READ
	hasAccess, err := userAccessManager.CheckAccess(accessControl)
	test_utils.AssertTrue(t, err == nil, "Expected CheckAccess to succeed")
	test_utils.AssertTrue(t, hasAccess == true, "Expected hasAccess to be true")
	mstub.MockTransactionEnd("9")

	// upload patient data2
	mstub.MockTransactionStart("10")
	stub = cached_stub.NewCachedStub(mstub)
	patientData = GeneratePatientData("patient1", "datatype1", "service1")
	// Need to add 1 so timestamps are different
	patientData.Timestamp++
	patientDataBytes, _ = json.Marshal(&patientData)
	args = []string{string(patientDataBytes)}
	_, err = UploadUserData(stub, service1Subgroup, args)
	test_utils.AssertTrue(t, err == nil, "Expected UploadUserData to succeed")
	mstub.MockTransactionEnd("10")

	// upload patient data3
	mstub.MockTransactionStart("11")
	stub = cached_stub.NewCachedStub(mstub)
	patientData = GeneratePatientData("patient1", "datatype1", "service1")
	// Need to add 2 so timestamps are different
	patientData.Timestamp += 2
	patientDataBytes, _ = json.Marshal(&patientData)
	args = []string{string(patientDataBytes)}
	_, err = UploadUserData(stub, service1Subgroup, args)
	test_utils.AssertTrue(t, err == nil, "Expected UploadUserData to succeed")
	mstub.MockTransactionEnd("11")

	// Download patient data as consent target
	mstub.MockTransactionStart("12")
	stub = cached_stub.NewCachedStub(mstub)
	dataResultBytes, err := DownloadUserData(stub, service1Subgroup, []string{"service1", "patient1", "datatype1", "false", strconv.FormatInt(0, 10), strconv.FormatInt(0, 10), "1000", strconv.FormatInt(time.Now().Unix(), 10)})
	test_utils.AssertTrue(t, err == nil, "Expected DownloadUserData to succeed")
	test_utils.AssertTrue(t, dataResultBytes != nil, "Expected DownloadUserData to succeed")
	dataResult := OwnerDataResultWithLog{}
	json.Unmarshal(dataResultBytes, &dataResult)
	test_utils.AssertTrue(t, len(dataResult.OwnerDatas) == 3, "Got patient data correctly")
	test_utils.AssertTrue(t, dataResult.OwnerDatas[0].Owner == "patient1", "Got owner data correctly")
	mstub.MockTransactionEnd("12")

	// Download patient data latest only
	mstub.MockTransactionStart("13")
	stub = cached_stub.NewCachedStub(mstub)
	dataResultBytes, err = DownloadUserData(stub, service1Subgroup, []string{"service1", "patient1", "datatype1", "true", strconv.FormatInt(0, 10), strconv.FormatInt(0, 10), "1000", strconv.FormatInt(time.Now().Unix(), 10)})
	test_utils.AssertTrue(t, err == nil, "Expected DownloadUserData to succeed")
	test_utils.AssertTrue(t, dataResultBytes != nil, "Expected DownloadUserData to succeed")
	json.Unmarshal(dataResultBytes, &dataResult)
	test_utils.AssertTrue(t, len(dataResult.OwnerDatas) == 1, "Got patient data correctly")
	test_utils.AssertTrue(t, dataResult.OwnerDatas[0].Owner == "patient1", "Got owner data correctly")
	mstub.MockTransactionEnd("13")

	// Download as patient himself
	mstub.MockTransactionStart("14")
	stub = cached_stub.NewCachedStub(mstub)
	dataResultBytes, err = DownloadUserData(stub, patient1Caller, []string{"service1", "patient1", "datatype1", "false", strconv.FormatInt(0, 10), strconv.FormatInt(0, 10), "1000", strconv.FormatInt(time.Now().Unix(), 10)})
	test_utils.AssertTrue(t, err == nil, "Expected DownloadUserData to succeed")
	test_utils.AssertTrue(t, dataResultBytes != nil, "Expected DownloadUserData to succeed")
	json.Unmarshal(dataResultBytes, &dataResult)
	test_utils.AssertTrue(t, len(dataResult.OwnerDatas) == 3, "Got patient data correctly")
	test_utils.AssertTrue(t, dataResult.OwnerDatas[0].Owner == "patient1", "Got owner data correctly")
	mstub.MockTransactionEnd("14")

	// attempt to download as auditor
	mstub.MockTransactionStart("14")
	stub = cached_stub.NewCachedStub(mstub)
	auditorPlatformUser, err := convertToPlatformUser(stub, auditor)
	test_utils.AssertTrue(t, err == nil, "Expected convertToPlatformUser to succeed")
	auditorCaller, _ := user_mgmt.GetUserData(stub, auditorPlatformUser, auditor.ID, true, true)
	dataResultBytes, err = DownloadUserData(stub, auditorCaller, []string{"service1", "patient1", "datatype1", "false", strconv.FormatInt(0, 10), strconv.FormatInt(0, 10), "1000", strconv.FormatInt(time.Now().Unix(), 10)})
	test_utils.AssertTrue(t, err != nil, "Expected DownloadUserData to fail")
	mstub.MockTransactionEnd("14")

	// Download as org of consent target
	mstub.MockTransactionStart("14")
	stub = cached_stub.NewCachedStub(mstub)
	dataResultBytes, err = DownloadUserData(stub, org1Caller, []string{"service1", "patient1", "datatype1", "false", strconv.FormatInt(0, 10), strconv.FormatInt(0, 10), "1000", strconv.FormatInt(time.Now().Unix(), 10)})
	test_utils.AssertTrue(t, err == nil, "Expected DownloadUserData to succeed")
	test_utils.AssertTrue(t, dataResultBytes != nil, "Expected DownloadUserData to succeed")
	json.Unmarshal(dataResultBytes, &dataResult)
	test_utils.AssertTrue(t, len(dataResult.OwnerDatas) == 3, "Got patient data correctly")
	test_utils.AssertTrue(t, dataResult.OwnerDatas[0].Owner == "patient1", "Got owner data correctly")
	mstub.MockTransactionEnd("14")

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

	// Download as orgUser1 (non admin of consent target)
	mstub.MockTransactionStart("14")
	stub = cached_stub.NewCachedStub(mstub)
	orgUser1Caller, _ := user_mgmt.GetUserData(stub, org1Caller, orgUser1.ID, true, true)
	dataResultBytes, err = DownloadUserData(stub, orgUser1Caller, []string{"service1", "patient1", "datatype1", "false", strconv.FormatInt(0, 10), strconv.FormatInt(0, 10), "1000", strconv.FormatInt(time.Now().Unix(), 10)})
	test_utils.AssertTrue(t, err != nil, "Expected DownloadUserData to fail")
	mstub.MockTransactionEnd("14")

	// give org admin permission
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = AddPermissionOrgAdmin(stub, org1Caller, []string{orgUser1.ID, org1.ID})
	test_utils.AssertTrue(t, err == nil, "Expected AddPermissionOrgAdmin to succeed")
	mstub.MockTransactionEnd("t123")

	// Download as orgUser1 (admin of consent target)
	mstub.MockTransactionStart("14")
	stub = cached_stub.NewCachedStub(mstub)
	orgUser1Caller, _ = user_mgmt.GetUserData(stub, org1Caller, orgUser1.ID, true, true)
	dataResultBytes, err = DownloadUserData(stub, orgUser1Caller, []string{"service1", "patient1", "datatype1", "false", strconv.FormatInt(0, 10), strconv.FormatInt(0, 10), "1000", strconv.FormatInt(time.Now().Unix(), 10)})
	test_utils.AssertTrue(t, err == nil, "Expected DownloadUserData to succeed")
	test_utils.AssertTrue(t, dataResultBytes != nil, "Expected DownloadUserData to succeed")
	json.Unmarshal(dataResultBytes, &dataResult)
	test_utils.AssertTrue(t, len(dataResult.OwnerDatas) == 3, "Got patient data correctly")
	test_utils.AssertTrue(t, dataResult.OwnerDatas[0].Owner == "patient1", "Got owner data correctly")
	mstub.MockTransactionEnd("14")

	// create another org user
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	orgUser2 := CreateTestSolutionUser("orgUser2")
	orgUser2.Org = "org1"
	orgUser2Bytes, _ := json.Marshal(&orgUser2)
	_, err = RegisterUser(stub, org1Caller, []string{string(orgUser2Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterUser to succeed")
	mstub.MockTransactionEnd("t123")

	// put orgUser2 in org1
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = PutUserInOrg(stub, org1Caller, []string{orgUser2.ID, org1.ID, "false"})
	test_utils.AssertTrue(t, err == nil, "Expected PutUserInOrg to succeed")
	mstub.MockTransactionEnd("t123")

	// Download as orgUser2 (non admin of consent target)
	mstub.MockTransactionStart("14")
	stub = cached_stub.NewCachedStub(mstub)
	orgUser2Caller, _ := user_mgmt.GetUserData(stub, org1Caller, orgUser2.ID, true, true)
	dataResultBytes, err = DownloadUserData(stub, orgUser2Caller, []string{"service1", "patient1", "datatype1", "false", strconv.FormatInt(0, 10), strconv.FormatInt(0, 10), "1000", strconv.FormatInt(time.Now().Unix(), 10)})
	test_utils.AssertTrue(t, err != nil, "Expected DownloadUserData to fail")
	mstub.MockTransactionEnd("14")

	// give service admin permission
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = AddPermissionServiceAdmin(stub, org1Caller, []string{orgUser2.ID, "service1"})
	test_utils.AssertTrue(t, err == nil, "Expected AddPermissionServiceAdmin to succeed")
	mstub.MockTransactionEnd("t123")

	// Download as orgUser2 (admin of consent target)
	mstub.MockTransactionStart("14")
	stub = cached_stub.NewCachedStub(mstub)
	orgUser2Caller, _ = user_mgmt.GetUserData(stub, org1Caller, orgUser2.ID, true, true)
	dataResultBytes, err = DownloadUserData(stub, orgUser2Caller, []string{"service1", "patient1", "datatype1", "false", strconv.FormatInt(0, 10), strconv.FormatInt(0, 10), "1000", strconv.FormatInt(time.Now().Unix(), 10)})
	test_utils.AssertTrue(t, err == nil, "Expected DownloadUserData to succeed")
	test_utils.AssertTrue(t, dataResultBytes != nil, "Expected DownloadUserData to succeed")
	json.Unmarshal(dataResultBytes, &dataResult)
	test_utils.AssertTrue(t, len(dataResult.OwnerDatas) == 3, "Got patient data correctly")
	test_utils.AssertTrue(t, dataResult.OwnerDatas[0].Owner == "patient1", "Got owner data correctly")
	mstub.MockTransactionEnd("14")

	// update consent to DENY
	mstub.MockTransactionStart("15")
	stub = cached_stub.NewCachedStub(mstub, true, true, true)
	consent = Consent{}
	consent.Owner = "patient1"
	consent.Service = "service1"
	consent.Target = "service1"
	consent.Datatype = "datatype1"
	consent.Option = []string{consentOptionDeny}
	consent.Timestamp = time.Now().Unix()
	consent.Expiration = 0
	consentBytes, _ = json.Marshal(&consent)
	_, err = PutConsentPatientData(stub, patient1Caller, []string{string(consentBytes)})
	test_utils.AssertTrue(t, err == nil, "Expected PutConsentPatientData to succeed")
	mstub.MockTransactionEnd("15")

	// Download patient data
	mstub.MockTransactionStart("16")
	stub = cached_stub.NewCachedStub(mstub)
	dataResultBytes, err = DownloadUserData(stub, service1Subgroup, []string{"service1", "patient1", "datatype1", "false", strconv.FormatInt(0, 10), strconv.FormatInt(0, 10), "1000", strconv.FormatInt(time.Now().Unix(), 10)})
	test_utils.AssertTrue(t, err != nil, "Expected DownloadUserData to fail")
	mstub.MockTransactionEnd("16")
}

func TestDownloadUserDataConsentToken(t *testing.T) {
	logger.SetLevel(shim.LogDebug)
	logger.Info("TestDownloadUserDataConsentToken function called")

	// create a MockStub
	mstub := SetupIndexesAndGetStub(t)

	// register admin user
	mstub.MockTransactionStart("t123")
	stub := cached_stub.NewCachedStub(mstub)
	init_common.Init(stub)
	err := SetupDataIndex(stub)
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

	// register patient
	mstub.MockTransactionStart("4")
	stub = cached_stub.NewCachedStub(mstub)
	patient1 := test_utils.CreateTestUser("patient1")
	patient1Bytes, _ := json.Marshal(&patient1)
	_, err = user_mgmt.RegisterUser(stub, org1Caller, []string{string(patient1Bytes), "false"})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterUser to succeed")
	mstub.MockTransactionEnd("4")

	// enroll patient
	mstub.MockTransactionStart("5")
	stub = cached_stub.NewCachedStub(mstub)
	enrollment1 := GenerateEnrollmentTest("patient1", "service1")
	enrollment1Bytes, _ := json.Marshal(&enrollment1)
	enrollmentKey1 := test_utils.GenerateSymKey()
	enrollmentKey1B64 := crypto.EncodeToB64String(enrollmentKey1)
	_, err = EnrollPatient(stub, org1Caller, []string{string(enrollment1Bytes), enrollmentKey1B64})
	test_utils.AssertTrue(t, err == nil, "Expected EnrollPatient to succeed")
	mstub.MockTransactionEnd("5")

	// patient give consent
	mstub.MockTransactionStart("6")
	stub = cached_stub.NewCachedStub(mstub, true, true, true)
	consent := Consent{}
	consent.Owner = "patient1"
	consent.Service = "service1"
	consent.Target = "service1"
	consent.Datatype = "datatype1"
	consent.Option = []string{consentOptionWrite}
	consent.Timestamp = time.Now().Unix()
	consent.Expiration = 0
	consentBytes, _ := json.Marshal(&consent)
	consentKey := test_utils.GenerateSymKey()
	consentKeyB64 := crypto.EncodeToB64String(consentKey)
	patient1Caller, err := user_mgmt.GetUserData(stub, patient1, "patient1", true, true)
	_, err = PutConsentPatientData(stub, patient1Caller, []string{string(consentBytes), consentKeyB64})
	test_utils.AssertTrue(t, err == nil, "Expected PutConsentPatientData to succeed")
	mstub.MockTransactionEnd("6")

	// get consent back to make sure it was added
	mstub.MockTransactionStart("7")
	stub = cached_stub.NewCachedStub(mstub)
	consentResultBytes, err := GetConsent(stub, patient1Caller, []string{"patient1", "service1", "datatype1"})
	test_utils.AssertTrue(t, err == nil, "Expected GetConsent to succeed")
	test_utils.AssertTrue(t, consentResultBytes != nil, "Expected GetConsent to succeed")
	consentResult := Consent{}
	json.Unmarshal(consentResultBytes, &consentResult)
	test_utils.AssertTrue(t, consentResult.Owner == "patient1", "Got consent owner correctly")
	test_utils.AssertTrue(t, consentResult.Target == "service1", "Got consent service correctly")
	mstub.MockTransactionEnd("7")

	// upload patient data
	mstub.MockTransactionStart("8")
	stub = cached_stub.NewCachedStub(mstub)
	service1Subgroup, _ := user_mgmt.GetUserData(stub, org1, "service1", true, true)
	patientData := GeneratePatientData("patient1", "datatype1", "service1")
	patientDataBytes, _ := json.Marshal(&patientData)
	dataKey := test_utils.GenerateSymKey()
	dataKeyB64 := crypto.EncodeToB64String(dataKey)
	args := []string{string(patientDataBytes), dataKeyB64}
	_, err = UploadUserData(stub, service1Subgroup, args)
	test_utils.AssertTrue(t, err == nil, "Expected UploadUserData to succeed")
	mstub.MockTransactionEnd("8")

	// upload patient data 2
	mstub.MockTransactionStart("9")
	stub = cached_stub.NewCachedStub(mstub)
	patientData = GeneratePatientData("patient1", "datatype1", "service1")
	patientData.Timestamp++
	patientDataBytes, _ = json.Marshal(&patientData)
	dataKey = test_utils.GenerateSymKey()
	dataKeyB64 = crypto.EncodeToB64String(dataKey)
	args = []string{string(patientDataBytes), dataKeyB64}
	_, err = UploadUserData(stub, service1Subgroup, args)
	test_utils.AssertTrue(t, err == nil, "Expected UploadUserData to succeed")
	mstub.MockTransactionEnd("9")

	// validate consent as service1
	mstub.MockTransactionStart("10")
	stub = cached_stub.NewCachedStub(mstub)
	cvResultBytes, err := ValidateConsent(stub, service1Subgroup, []string{"patient1", "service1", "datatype1", consentOptionWrite, strconv.FormatInt(time.Now().Unix(), 10)})
	test_utils.AssertTrue(t, err == nil, "Expected ValidateConsent to succeed")
	test_utils.AssertTrue(t, cvResultBytes != nil, "Expected ValidateConsent to succeed")
	cvResult := ValidationResultWithLog{}
	json.Unmarshal(cvResultBytes, &cvResult)
	test_utils.AssertTrue(t, cvResult.ConsentValidation.PermissionGranted == true, "Got consent validation correctly")
	test_utils.AssertTrue(t, cvResult.ConsentValidation.Message == "permission granted", "Got consent validation correctly")
	mstub.MockTransactionEnd("10")

	// download user data
	mstub.MockTransactionStart("11")
	stub = cached_stub.NewCachedStub(mstub)
	// NOTE: We need to call url.QueryUnescape() function because on working environment Node.js automatically
	// replace URL escape characters with original values
	consentToken, _ := url.QueryUnescape(cvResult.ConsentValidation.Token)
	dataResultBytes, err := DownloadUserDataConsentToken(stub, service1Subgroup, []string{"false", strconv.FormatInt(0, 10), strconv.FormatInt(0, 10), "1000", strconv.FormatInt(time.Now().Unix(), 10), consentToken})
	test_utils.AssertTrue(t, err == nil, "Expected DownloadUserDataConsentToken to succeed")
	test_utils.AssertTrue(t, dataResultBytes != nil, "Expected DownloadUserDataConsentToken to succeed")
	dataResult := OwnerDataResultWithLog{}
	json.Unmarshal(dataResultBytes, &dataResult)
	test_utils.AssertTrue(t, len(dataResult.OwnerDatas) == 2, "Got patient data correctly")
	test_utils.AssertTrue(t, dataResult.OwnerDatas[0].Owner == "patient1", "Got patient data correctly")
	mstub.MockTransactionEnd("11")

	// Download user data latest only
	mstub.MockTransactionStart("12")
	stub = cached_stub.NewCachedStub(mstub)
	consentToken, _ = url.QueryUnescape(cvResult.ConsentValidation.Token)
	dataResultBytes, err = DownloadUserDataConsentToken(stub, service1Subgroup, []string{"true", strconv.FormatInt(0, 10), strconv.FormatInt(0, 10), "1000", strconv.FormatInt(time.Now().Unix(), 10), consentToken})
	test_utils.AssertTrue(t, err == nil, "Expected DownloadUserDataConsentToken to succeed")
	test_utils.AssertTrue(t, dataResultBytes != nil, "Expected DownloadUserDataConsentToken to succeed")
	json.Unmarshal(dataResultBytes, &dataResult)
	test_utils.AssertTrue(t, len(dataResult.OwnerDatas) == 1, "Got patient data correctly")
	test_utils.AssertTrue(t, dataResult.OwnerDatas[0].Owner == "patient1", "Got patient data correctly")
	mstub.MockTransactionEnd("12")

	// create an org user
	mstub.MockTransactionStart("13")
	stub = cached_stub.NewCachedStub(mstub)
	orgUser1 := CreateTestSolutionUser("orgUser1")
	orgUser1.Org = "org1"
	orgUser1Bytes, _ := json.Marshal(&orgUser1)
	_, err = RegisterUser(stub, org1Caller, []string{string(orgUser1Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterUser to succeed")
	mstub.MockTransactionEnd("13")

	// put orgUser1 in org1
	mstub.MockTransactionStart("14")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = PutUserInOrg(stub, org1Caller, []string{orgUser1.ID, org1.ID, "false"})
	test_utils.AssertTrue(t, err == nil, "Expected PutUserInOrg to succeed")
	mstub.MockTransactionEnd("14")

	// validate consent as orgUser1
	mstub.MockTransactionStart("15")
	stub = cached_stub.NewCachedStub(mstub)
	orgUser1Caller, _ := user_mgmt.GetUserData(stub, org1Caller, orgUser1.ID, true, true)
	cvResultBytes, err = ValidateConsent(stub, orgUser1Caller, []string{"patient1", "service1", "datatype1", consentOptionWrite, strconv.FormatInt(time.Now().Unix(), 10)})
	test_utils.AssertTrue(t, err == nil, "Expected ValidateConsent to succeed")
	test_utils.AssertTrue(t, cvResultBytes != nil, "Expected ValidateConsent to succeed")
	json.Unmarshal(cvResultBytes, &cvResult)
	test_utils.AssertTrue(t, cvResult.ConsentValidation.PermissionGranted == false, "Got consent validation correctly")
	test_utils.AssertTrue(t, cvResult.ConsentValidation.Message != "permission granted", "Got consent validation correctly")
	mstub.MockTransactionEnd("15")

	// give org admin permission
	mstub.MockTransactionStart("16")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = AddPermissionOrgAdmin(stub, org1Caller, []string{orgUser1.ID, org1.ID})
	test_utils.AssertTrue(t, err == nil, "Expected AddPermissionOrgAdmin to succeed")
	mstub.MockTransactionEnd("16")

	// validate consent as orgUser1
	mstub.MockTransactionStart("17")
	stub = cached_stub.NewCachedStub(mstub)
	orgUser1Caller, _ = user_mgmt.GetUserData(stub, org1Caller, orgUser1.ID, true, true)
	cvResultBytes, err = ValidateConsent(stub, orgUser1Caller, []string{"patient1", "service1", "datatype1", consentOptionWrite, strconv.FormatInt(time.Now().Unix(), 10)})
	test_utils.AssertTrue(t, err == nil, "Expected ValidateConsent to succeed")
	test_utils.AssertTrue(t, cvResultBytes != nil, "Expected ValidateConsent to succeed")
	json.Unmarshal(cvResultBytes, &cvResult)
	test_utils.AssertTrue(t, cvResult.ConsentValidation.PermissionGranted == true, "Got consent validation correctly")
	test_utils.AssertTrue(t, cvResult.ConsentValidation.Message == "permission granted", "Got consent validation correctly")
	mstub.MockTransactionEnd("17")

	// download user data
	mstub.MockTransactionStart("18")
	stub = cached_stub.NewCachedStub(mstub)
	consentToken, _ = url.QueryUnescape(cvResult.ConsentValidation.Token)
	dataResultBytes, err = DownloadUserDataConsentToken(stub, orgUser1Caller, []string{"false", strconv.FormatInt(0, 10), strconv.FormatInt(0, 10), "1000", strconv.FormatInt(time.Now().Unix(), 10), consentToken})
	test_utils.AssertTrue(t, err == nil, "Expected DownloadUserDataConsentToken to succeed")
	test_utils.AssertTrue(t, dataResultBytes != nil, "Expected DownloadUserDataConsentToken to succeed")
	json.Unmarshal(dataResultBytes, &dataResult)
	test_utils.AssertTrue(t, len(dataResult.OwnerDatas) == 2, "Got patient data correctly")
	test_utils.AssertTrue(t, dataResult.OwnerDatas[0].Owner == "patient1", "Got patient data correctly")
	mstub.MockTransactionEnd("18")

	// create another org user
	mstub.MockTransactionStart("19")
	stub = cached_stub.NewCachedStub(mstub)
	orgUser2 := CreateTestSolutionUser("orgUser2")
	orgUser2.Org = "org1"
	orgUser2Bytes, _ := json.Marshal(&orgUser2)
	_, err = RegisterUser(stub, org1Caller, []string{string(orgUser2Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterUser to succeed")
	mstub.MockTransactionEnd("19")

	// put orgUser2 in org1
	mstub.MockTransactionStart("20")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = PutUserInOrg(stub, org1Caller, []string{orgUser2.ID, org1.ID, "false"})
	test_utils.AssertTrue(t, err == nil, "Expected PutUserInOrg to succeed")
	mstub.MockTransactionEnd("20")

	// give service admin permission
	mstub.MockTransactionStart("21")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = AddPermissionServiceAdmin(stub, org1Caller, []string{orgUser2.ID, "service1"})
	test_utils.AssertTrue(t, err == nil, "Expected AddPermissionServiceAdmin to succeed")
	mstub.MockTransactionEnd("21")

	// validate consent as orgUser2
	mstub.MockTransactionStart("22")
	stub = cached_stub.NewCachedStub(mstub)
	orgUser2Caller, _ := user_mgmt.GetUserData(stub, org1Caller, orgUser2.ID, true, true)
	cvResultBytes, err = ValidateConsent(stub, orgUser2Caller, []string{"patient1", "service1", "datatype1", consentOptionWrite, strconv.FormatInt(time.Now().Unix(), 10)})
	test_utils.AssertTrue(t, err == nil, "Expected ValidateConsent to succeed")
	test_utils.AssertTrue(t, cvResultBytes != nil, "Expected ValidateConsent to succeed")
	json.Unmarshal(cvResultBytes, &cvResult)
	test_utils.AssertTrue(t, cvResult.ConsentValidation.PermissionGranted == true, "Got consent validation correctly")
	test_utils.AssertTrue(t, cvResult.ConsentValidation.Message == "permission granted", "Got consent validation correctly")
	mstub.MockTransactionEnd("22")

	// download user data
	mstub.MockTransactionStart("23")
	stub = cached_stub.NewCachedStub(mstub)
	consentToken, _ = url.QueryUnescape(cvResult.ConsentValidation.Token)
	dataResultBytes, err = DownloadUserDataConsentToken(stub, orgUser2Caller, []string{"false", strconv.FormatInt(0, 10), strconv.FormatInt(0, 10), "1000", strconv.FormatInt(time.Now().Unix(), 10), consentToken})
	test_utils.AssertTrue(t, err == nil, "Expected DownloadUserDataConsentToken to succeed")
	test_utils.AssertTrue(t, dataResultBytes != nil, "Expected DownloadUserDataConsentToken to succeed")
	json.Unmarshal(dataResultBytes, &dataResult)
	test_utils.AssertTrue(t, len(dataResult.OwnerDatas) == 2, "Got patient data correctly")
	test_utils.AssertTrue(t, dataResult.OwnerDatas[0].Owner == "patient1", "Got patient data correctly")
	mstub.MockTransactionEnd("23")
}

func TestUploadOwnerData(t *testing.T) {
	logger.SetLevel(shim.LogDebug)
	logger.Info("TestPutConsentOwnerData function called")

	// create a MockStub
	mstub := SetupIndexesAndGetStub(t)

	// register admin user
	mstub.MockTransactionStart("t123")
	stub := cached_stub.NewCachedStub(mstub)
	init_common.Init(stub)
	err := SetupDataIndex(stub)
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

	// upload owner data as default service admin
	mstub.MockTransactionStart("4")
	stub = cached_stub.NewCachedStub(mstub, true, true, true)
	ownerData := GenerateOwnerData("service1", "datatype1")
	ownerDataBytes, _ := json.Marshal(&ownerData)
	dataKey := test_utils.GenerateSymKey()
	dataKeyB64 := crypto.EncodeToB64String(dataKey)
	args := []string{string(ownerDataBytes), dataKeyB64}
	serviceSubgroup, _ := user_mgmt.GetUserData(stub, org1Caller, "service1", true, true)
	_, err = UploadOwnerData(stub, serviceSubgroup, args)
	test_utils.AssertTrue(t, err == nil, "Expected UploadOwnerData to succeed")
	mstub.MockTransactionEnd("4")

	// register patient and use patient as owner
	mstub.MockTransactionStart("5")
	stub = cached_stub.NewCachedStub(mstub)
	patient1 := test_utils.CreateTestUser("patient1")
	patient1Bytes, _ := json.Marshal(&patient1)
	_, err = user_mgmt.RegisterUser(stub, org1Caller, []string{string(patient1Bytes), "false"})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterUser to succeed")
	mstub.MockTransactionEnd("5")

	// upload owner data
	mstub.MockTransactionStart("6")
	stub = cached_stub.NewCachedStub(mstub, true, true, true)
	ownerData = GenerateOwnerData("patient1", "datatype1")
	ownerDataBytes, _ = json.Marshal(&ownerData)
	dataKey = test_utils.GenerateSymKey()
	dataKeyB64 = crypto.EncodeToB64String(dataKey)
	args = []string{string(ownerDataBytes), dataKeyB64}
	patient1Caller, _ := user_mgmt.GetUserData(stub, patient1, patient1.ID, true, true)
	_, err = UploadOwnerData(stub, patient1Caller, args)
	test_utils.AssertTrue(t, err == nil, "Expected UploadOwnerData to succeed")
	mstub.MockTransactionEnd("6")

	// create an org user
	mstub.MockTransactionStart("7")
	stub = cached_stub.NewCachedStub(mstub)
	orgUser1 := CreateTestSolutionUser("orgUser1")
	orgUser1.Org = "org1"
	orgUser1Bytes, _ := json.Marshal(&orgUser1)
	_, err = RegisterUser(stub, org1Caller, []string{string(orgUser1Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterUser to succeed")
	mstub.MockTransactionEnd("7")

	// put orgUser1 in org1
	mstub.MockTransactionStart("8")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = PutUserInOrg(stub, org1Caller, []string{orgUser1.ID, org1.ID, "false"})
	test_utils.AssertTrue(t, err == nil, "Expected PutUserInOrg to succeed")
	mstub.MockTransactionEnd("8")

	// upload owner data where owner is org as orgUser1 (non-admin)
	mstub.MockTransactionStart("9")
	stub = cached_stub.NewCachedStub(mstub, true, true, true)
	ownerData = GenerateOwnerData("org1", "datatype1")
	ownerDataBytes, _ = json.Marshal(&ownerData)
	dataKey = test_utils.GenerateSymKey()
	dataKeyB64 = crypto.EncodeToB64String(dataKey)
	args = []string{string(ownerDataBytes), dataKeyB64}
	orgUser1Caller, _ := user_mgmt.GetUserData(stub, org1Caller, orgUser1.ID, true, true)
	_, err = UploadOwnerData(stub, orgUser1Caller, args)
	test_utils.AssertTrue(t, err != nil, "Expected UploadOwnerData to fail")
	mstub.MockTransactionEnd("9")

	// give org admin permission
	mstub.MockTransactionStart("10")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = AddPermissionOrgAdmin(stub, org1Caller, []string{orgUser1.ID, org1.ID})
	test_utils.AssertTrue(t, err == nil, "Expected AddPermissionOrgAdmin to succeed")
	mstub.MockTransactionEnd("10")

	// upload owner data where owner is org as orgUser1 (org admin)
	mstub.MockTransactionStart("11")
	stub = cached_stub.NewCachedStub(mstub, true, true, true)
	ownerData = GenerateOwnerData("org1", "datatype1")
	ownerDataBytes, _ = json.Marshal(&ownerData)
	dataKey = test_utils.GenerateSymKey()
	dataKeyB64 = crypto.EncodeToB64String(dataKey)
	args = []string{string(ownerDataBytes), dataKeyB64}
	orgUser1Caller, _ = user_mgmt.GetUserData(stub, org1Caller, orgUser1.ID, true, true)
	_, err = UploadOwnerData(stub, orgUser1Caller, args)
	test_utils.AssertTrue(t, err == nil, "Expected UploadOwnerData to succeed")
	mstub.MockTransactionEnd("11")

	// upload owner data again where owner is org as orgUser1 (org admin)
	mstub.MockTransactionStart("12")
	stub = cached_stub.NewCachedStub(mstub, true, true, true)
	ownerData = GenerateOwnerData("org1", "datatype1")
	ownerData.Timestamp++
	ownerDataBytes, _ = json.Marshal(&ownerData)
	args = []string{string(ownerDataBytes)}
	_, err = UploadOwnerData(stub, orgUser1Caller, args)
	test_utils.AssertTrue(t, err == nil, "Expected UploadOwnerData to succeed")
	mstub.MockTransactionEnd("12")

	// upload owner data again where owner is default org admin
	mstub.MockTransactionStart("13")
	stub = cached_stub.NewCachedStub(mstub, true, true, true)
	ownerData = GenerateOwnerData("org1", "datatype1")
	ownerData.Timestamp += 2
	ownerDataBytes, _ = json.Marshal(&ownerData)
	args = []string{string(ownerDataBytes)}
	_, err = UploadOwnerData(stub, org1Caller, args)
	test_utils.AssertTrue(t, err == nil, "Expected UploadOwnerData to succeed")
	mstub.MockTransactionEnd("13")

	// upload owner data again where owner is service as orgUser1 (org admin)
	// ownerdata for service1 datatype1 already uploaded, this time caller is different
	mstub.MockTransactionStart("14")
	stub = cached_stub.NewCachedStub(mstub, true, true, true)
	ownerData = GenerateOwnerData("service1", "datatype1")
	ownerData.Timestamp++
	ownerDataBytes, _ = json.Marshal(&ownerData)
	args = []string{string(ownerDataBytes)}
	_, err = UploadOwnerData(stub, orgUser1Caller, args)
	test_utils.AssertTrue(t, err == nil, "Expected UploadOwnerData to succeed")
	mstub.MockTransactionEnd("14")

	// create another org user
	mstub.MockTransactionStart("15")
	stub = cached_stub.NewCachedStub(mstub)
	orgUser2 := CreateTestSolutionUser("orgUser2")
	orgUser2.Org = "org1"
	orgUser2Bytes, _ := json.Marshal(&orgUser2)
	_, err = RegisterUser(stub, org1Caller, []string{string(orgUser2Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterUser to succeed")
	mstub.MockTransactionEnd("15")

	// put orgUser2 in org1
	mstub.MockTransactionStart("16")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = PutUserInOrg(stub, org1Caller, []string{orgUser2.ID, org1.ID, "false"})
	test_utils.AssertTrue(t, err == nil, "Expected PutUserInOrg to succeed")
	mstub.MockTransactionEnd("16")

	// give service admin permission
	mstub.MockTransactionStart("17")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = AddPermissionServiceAdmin(stub, org1Caller, []string{orgUser2.ID, "service1"})
	test_utils.AssertTrue(t, err == nil, "Expected AddPermissionServiceAdmin to succeed")
	mstub.MockTransactionEnd("17")

	// upload owner data again where owner is service as orgUser2 (service admin)
	// ownerdata for service1 datatype1 already uploaded, this time caller is different
	mstub.MockTransactionStart("18")
	stub = cached_stub.NewCachedStub(mstub, true, true, true)
	ownerData = GenerateOwnerData("service1", "datatype1")
	ownerData.Timestamp += 2
	ownerDataBytes, _ = json.Marshal(&ownerData)
	args = []string{string(ownerDataBytes)}
	orgUser2Caller, _ := user_mgmt.GetUserData(stub, org1Caller, orgUser2.ID, true, true)
	_, err = UploadOwnerData(stub, orgUser2Caller, args)
	test_utils.AssertTrue(t, err == nil, "Expected UploadOwnerData to succeed")
	mstub.MockTransactionEnd("18")
}

func TestDownloadOwnerDataConsentToken(t *testing.T) {
	logger.SetLevel(shim.LogDebug)
	logger.Info("TestDownloadOwnerDataConsentToken function called")

	// create a MockStub
	mstub := SetupIndexesAndGetStub(t)

	// register admin user
	mstub.MockTransactionStart("t123")
	stub := cached_stub.NewCachedStub(mstub)
	init_common.Init(stub)
	err := SetupDataIndex(stub)
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

	// upload owner data
	mstub.MockTransactionStart("4")
	stub = cached_stub.NewCachedStub(mstub, true, true, true)
	ownerData := GenerateOwnerData("service1", "datatype1")
	ownerDataBytes, _ := json.Marshal(&ownerData)
	dataKey := test_utils.GenerateSymKey()
	dataKeyB64 := crypto.EncodeToB64String(dataKey)
	args := []string{string(ownerDataBytes), dataKeyB64}
	serviceSubgroup, _ := user_mgmt.GetUserData(stub, org1, "service1", true, true)
	_, err = UploadOwnerData(stub, serviceSubgroup, args)
	test_utils.AssertTrue(t, err == nil, "Expected UploadOwnerData to succeed")
	mstub.MockTransactionEnd("4")

	// upload owner data again
	mstub.MockTransactionStart("5")
	stub = cached_stub.NewCachedStub(mstub, true, true, true)
	ownerData = GenerateOwnerData("service1", "datatype1")
	ownerData.Timestamp++
	ownerDataBytes, _ = json.Marshal(&ownerData)
	dataKey = test_utils.GenerateSymKey()
	dataKeyB64 = crypto.EncodeToB64String(dataKey)
	args = []string{string(ownerDataBytes), dataKeyB64}
	_, err = UploadOwnerData(stub, serviceSubgroup, args)
	test_utils.AssertTrue(t, err == nil, "Expected UploadOwnerData to succeed")
	mstub.MockTransactionEnd("5")

	// register org2
	mstub.MockTransactionStart("6")
	stub = cached_stub.NewCachedStub(mstub)
	org2 := test_utils.CreateTestGroup("org2")
	org2Bytes, _ := json.Marshal(&org2)
	_, err = RegisterOrg(stub, org2, []string{string(org2Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterOrg to succeed")
	mstub.MockTransactionEnd("6")

	// register datatype2
	mstub.MockTransactionStart("7")
	stub = cached_stub.NewCachedStub(mstub)
	datatype2 := Datatype{DatatypeID: "datatype2", Description: "datatype2"}
	datatype2Bytes, _ := json.Marshal(&datatype2)
	org2Caller, _ := user_mgmt.GetUserData(stub, org2, org2.ID, true, true)
	_, err = RegisterDatatype(stub, org2Caller, []string{string(datatype2Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterDatatype to succeed")
	mstub.MockTransactionEnd("7")

	// register service2
	mstub.MockTransactionStart("8")
	stub = cached_stub.NewCachedStub(mstub, true, true, true)
	serviceDatatype2 := GenerateServiceDatatypeForTesting("datatype2", "service2", []string{consentOptionWrite, consentOptionRead})
	service2 := GenerateServiceForTesting("service2", "org2", []ServiceDatatype{serviceDatatype2})
	service2Bytes, _ := json.Marshal(&service2)
	_, err = RegisterService(stub, org2Caller, []string{string(service2Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterService to succeed")
	mstub.MockTransactionEnd("8")

	// service1 give consent to service2
	mstub.MockTransactionStart("9")
	stub = cached_stub.NewCachedStub(mstub, true, true, true)
	consent := Consent{}
	consent.Owner = "service1"
	consent.Service = "service1"
	consent.Datatype = "datatype1"
	consent.Target = "service2"
	consent.Option = []string{consentOptionWrite}
	consent.Timestamp = time.Now().Unix()
	consent.Expiration = 0
	consentBytes, _ := json.Marshal(&consent)
	consentKey := test_utils.GenerateSymKey()
	consentKeyB64 := crypto.EncodeToB64String(consentKey)
	service1Subgroup, _ := user_mgmt.GetUserData(stub, org1Caller, "service1", true, true)
	_, err = PutConsentOwnerData(stub, service1Subgroup, []string{string(consentBytes), consentKeyB64})
	test_utils.AssertTrue(t, err == nil, "Expected PutConsentOwnerData to succeed")
	mstub.MockTransactionEnd("9")

	// get consent back to make sure it was added
	mstub.MockTransactionStart("10")
	stub = cached_stub.NewCachedStub(mstub)
	consentResultBytes, err := GetConsent(stub, service1Subgroup, []string{"service1", "service2", "datatype1"})
	test_utils.AssertTrue(t, err == nil, "Expected GetConsent to succeed")
	test_utils.AssertTrue(t, consentResultBytes != nil, "Expected GetConsent to succeed")
	consentResult := Consent{}
	json.Unmarshal(consentResultBytes, &consentResult)
	test_utils.AssertTrue(t, consentResult.Owner == "service1", "Got consent owner correctly")
	test_utils.AssertTrue(t, consentResult.Target == "service2", "Got consent target correctly")
	mstub.MockTransactionEnd("10")

	// validate consent as service2
	mstub.MockTransactionStart("11")
	stub = cached_stub.NewCachedStub(mstub)
	service2Subgroup, _ := user_mgmt.GetUserData(stub, org2Caller, "service2", true, true)
	cvResultBytes, err := ValidateConsent(stub, service2Subgroup, []string{"service1", "service2", "datatype1", consentOptionWrite, strconv.FormatInt(time.Now().Unix(), 10)})
	test_utils.AssertTrue(t, err == nil, "Expected ValidateConsent to succeed")
	test_utils.AssertTrue(t, cvResultBytes != nil, "Expected ValidateConsent to succeed")
	cvResult := ValidationResultWithLog{}
	json.Unmarshal(cvResultBytes, &cvResult)
	test_utils.AssertTrue(t, cvResult.ConsentValidation.PermissionGranted == true, "Got consent validation correctly")
	test_utils.AssertTrue(t, cvResult.ConsentValidation.Message == "permission granted", "Got consent validation correctly")
	mstub.MockTransactionEnd("11")

	// download
	mstub.MockTransactionStart("12")
	stub = cached_stub.NewCachedStub(mstub)
	consentToken, _ := url.QueryUnescape(cvResult.ConsentValidation.Token)
	dataResultBytes, err := DownloadOwnerDataConsentToken(stub, service2Subgroup, []string{"false", strconv.FormatInt(0, 10), strconv.FormatInt(0, 10), "1000", strconv.FormatInt(time.Now().Unix(), 10), consentToken})
	test_utils.AssertTrue(t, err == nil, "Expected DownloadOwnerDataConsentToken to succeed")
	test_utils.AssertTrue(t, dataResultBytes != nil, "Expected DownloadOwnerDataConsentToken to succeed")
	dataResult := OwnerDataResultWithLog{}
	json.Unmarshal(dataResultBytes, &dataResult)
	test_utils.AssertTrue(t, len(dataResult.OwnerDatas) == 2, "Got owner data correctly")
	test_utils.AssertTrue(t, dataResult.OwnerDatas[0].Owner == "service1", "Got owner data correctly")
	mstub.MockTransactionEnd("12")

	// download latest only
	mstub.MockTransactionStart("13")
	stub = cached_stub.NewCachedStub(mstub)
	consentToken, _ = url.QueryUnescape(cvResult.ConsentValidation.Token)
	dataResultBytes, err = DownloadOwnerDataConsentToken(stub, service2Subgroup, []string{"true", strconv.FormatInt(0, 10), strconv.FormatInt(0, 10), "1000", strconv.FormatInt(time.Now().Unix(), 10), consentToken})
	test_utils.AssertTrue(t, err == nil, "Expected DownloadOwnerDataConsentToken to succeed")
	test_utils.AssertTrue(t, dataResultBytes != nil, "Expected DownloadOwnerDataConsentToken to succeed")
	json.Unmarshal(dataResultBytes, &dataResult)
	test_utils.AssertTrue(t, len(dataResult.OwnerDatas) == 1, "Got owner data correctly")
	test_utils.AssertTrue(t, dataResult.OwnerDatas[0].Owner == "service1", "Got owner data correctly")
	mstub.MockTransactionEnd("13")

	// create an org user
	mstub.MockTransactionStart("13")
	stub = cached_stub.NewCachedStub(mstub)
	orgUser1 := CreateTestSolutionUser("orgUser1")
	orgUser1.Org = "org2"
	orgUser1Bytes, _ := json.Marshal(&orgUser1)
	_, err = RegisterUser(stub, org2Caller, []string{string(orgUser1Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterUser to succeed")
	mstub.MockTransactionEnd("13")

	// put orgUser1 in org2
	mstub.MockTransactionStart("14")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = PutUserInOrg(stub, org2Caller, []string{orgUser1.ID, org2.ID, "false"})
	test_utils.AssertTrue(t, err == nil, "Expected PutUserInOrg to succeed")
	mstub.MockTransactionEnd("14")

	// validate consent as orgUser1
	mstub.MockTransactionStart("15")
	stub = cached_stub.NewCachedStub(mstub)
	orgUser1Caller, _ := user_mgmt.GetUserData(stub, org2Caller, orgUser1.ID, true, true)
	cvResultBytes, err = ValidateConsent(stub, orgUser1Caller, []string{"service1", "service2", "datatype1", consentOptionWrite, strconv.FormatInt(time.Now().Unix(), 10)})
	test_utils.AssertTrue(t, err == nil, "Expected ValidateConsent to succeed")
	test_utils.AssertTrue(t, cvResultBytes != nil, "Expected ValidateConsent to succeed")
	json.Unmarshal(cvResultBytes, &cvResult)
	test_utils.AssertTrue(t, cvResult.ConsentValidation.PermissionGranted == false, "Got consent validation correctly")
	test_utils.AssertTrue(t, cvResult.ConsentValidation.Message != "permission granted", "Got consent validation correctly")
	mstub.MockTransactionEnd("15")

	// give org admin permission
	mstub.MockTransactionStart("16")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = AddPermissionOrgAdmin(stub, org2Caller, []string{orgUser1.ID, org2.ID})
	test_utils.AssertTrue(t, err == nil, "Expected AddPermissionOrgAdmin to succeed")
	mstub.MockTransactionEnd("16")

	// validate consent as orgUser1
	mstub.MockTransactionStart("17")
	stub = cached_stub.NewCachedStub(mstub)
	orgUser1Caller, _ = user_mgmt.GetUserData(stub, org2Caller, orgUser1.ID, true, true)
	cvResultBytes, err = ValidateConsent(stub, orgUser1Caller, []string{"service1", "service2", "datatype1", consentOptionWrite, strconv.FormatInt(time.Now().Unix(), 10)})
	test_utils.AssertTrue(t, err == nil, "Expected ValidateConsent to succeed")
	test_utils.AssertTrue(t, cvResultBytes != nil, "Expected ValidateConsent to succeed")
	json.Unmarshal(cvResultBytes, &cvResult)
	test_utils.AssertTrue(t, cvResult.ConsentValidation.PermissionGranted == true, "Got consent validation correctly")
	test_utils.AssertTrue(t, cvResult.ConsentValidation.Message == "permission granted", "Got consent validation correctly")
	mstub.MockTransactionEnd("17")

	// download owner data
	mstub.MockTransactionStart("18")
	stub = cached_stub.NewCachedStub(mstub)
	consentToken, _ = url.QueryUnescape(cvResult.ConsentValidation.Token)
	dataResultBytes, err = DownloadOwnerDataConsentToken(stub, orgUser1Caller, []string{"false", strconv.FormatInt(0, 10), strconv.FormatInt(0, 10), "1000", strconv.FormatInt(time.Now().Unix(), 10), consentToken})
	test_utils.AssertTrue(t, err == nil, "Expected DownloadUserDataConsentToken to succeed")
	test_utils.AssertTrue(t, dataResultBytes != nil, "Expected DownloadUserDataConsentToken to succeed")
	json.Unmarshal(dataResultBytes, &dataResult)
	test_utils.AssertTrue(t, len(dataResult.OwnerDatas) == 2, "Got patient data correctly")
	test_utils.AssertTrue(t, dataResult.OwnerDatas[0].Owner == "service1", "Got patient data correctly")
	mstub.MockTransactionEnd("18")

	// create another org user
	mstub.MockTransactionStart("19")
	stub = cached_stub.NewCachedStub(mstub)
	orgUser2 := CreateTestSolutionUser("orgUser2")
	orgUser2.Org = "org2"
	orgUser2Bytes, _ := json.Marshal(&orgUser2)
	_, err = RegisterUser(stub, org2Caller, []string{string(orgUser2Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterUser to succeed")
	mstub.MockTransactionEnd("19")

	// put orgUser2 in org2
	mstub.MockTransactionStart("20")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = PutUserInOrg(stub, org2Caller, []string{orgUser2.ID, org2.ID, "false"})
	test_utils.AssertTrue(t, err == nil, "Expected PutUserInOrg to succeed")
	mstub.MockTransactionEnd("20")

	// give service admin permission
	mstub.MockTransactionStart("21")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = AddPermissionServiceAdmin(stub, org2Caller, []string{orgUser2.ID, "service2"})
	test_utils.AssertTrue(t, err == nil, "Expected AddPermissionServiceAdmin to succeed")
	mstub.MockTransactionEnd("21")

	// validate consent as orgUser2
	mstub.MockTransactionStart("11")
	stub = cached_stub.NewCachedStub(mstub)
	orgUser2Caller, _ := user_mgmt.GetUserData(stub, org2Caller, orgUser2.ID, true, true)
	cvResultBytes, err = ValidateConsent(stub, orgUser2Caller, []string{"service1", "service2", "datatype1", consentOptionWrite, strconv.FormatInt(time.Now().Unix(), 10)})
	test_utils.AssertTrue(t, err == nil, "Expected ValidateConsent to succeed")
	test_utils.AssertTrue(t, cvResultBytes != nil, "Expected ValidateConsent to succeed")
	json.Unmarshal(cvResultBytes, &cvResult)
	test_utils.AssertTrue(t, cvResult.ConsentValidation.PermissionGranted == true, "Got consent validation correctly")
	test_utils.AssertTrue(t, cvResult.ConsentValidation.Message == "permission granted", "Got consent validation correctly")
	mstub.MockTransactionEnd("11")

	// download
	mstub.MockTransactionStart("12")
	stub = cached_stub.NewCachedStub(mstub)
	consentToken, _ = url.QueryUnescape(cvResult.ConsentValidation.Token)
	dataResultBytes, err = DownloadOwnerDataConsentToken(stub, orgUser2Caller, []string{"false", strconv.FormatInt(0, 10), strconv.FormatInt(0, 10), "1000", strconv.FormatInt(time.Now().Unix(), 10), consentToken})
	test_utils.AssertTrue(t, err == nil, "Expected DownloadOwnerDataConsentToken to succeed")
	test_utils.AssertTrue(t, dataResultBytes != nil, "Expected DownloadOwnerDataConsentToken to succeed")
	json.Unmarshal(dataResultBytes, &dataResult)
	test_utils.AssertTrue(t, len(dataResult.OwnerDatas) == 2, "Got owner data correctly")
	test_utils.AssertTrue(t, dataResult.OwnerDatas[0].Owner == "service1", "Got owner data correctly")
	mstub.MockTransactionEnd("12")
}

func TestDownloadOwnerDataWithConsent(t *testing.T) {
	logger.SetLevel(shim.LogDebug)
	logger.Info("TestDownloadOwnerDataWithConsent function called")

	// create a MockStub
	mstub := SetupIndexesAndGetStub(t)

	// register admin user
	mstub.MockTransactionStart("t123")
	stub := cached_stub.NewCachedStub(mstub)
	init_common.Init(stub)
	err := SetupDataIndex(stub)
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

	// upload owner data
	mstub.MockTransactionStart("4")
	stub = cached_stub.NewCachedStub(mstub, true, true, true)
	ownerData := GenerateOwnerData("service1", "datatype1")
	ownerDataBytes, _ := json.Marshal(&ownerData)
	dataKey := test_utils.GenerateSymKey()
	dataKeyB64 := crypto.EncodeToB64String(dataKey)
	args := []string{string(ownerDataBytes), dataKeyB64}
	serviceSubgroup, _ := user_mgmt.GetUserData(stub, org1, "service1", true, true)
	_, err = UploadOwnerData(stub, serviceSubgroup, args)
	test_utils.AssertTrue(t, err == nil, "Expected UploadOwnerData to succeed")
	mstub.MockTransactionEnd("4")

	// upload owner data again
	mstub.MockTransactionStart("5")
	stub = cached_stub.NewCachedStub(mstub, true, true, true)
	ownerData = GenerateOwnerData("service1", "datatype1")
	ownerData.Timestamp++
	ownerDataBytes, _ = json.Marshal(&ownerData)
	dataKey = test_utils.GenerateSymKey()
	dataKeyB64 = crypto.EncodeToB64String(dataKey)
	args = []string{string(ownerDataBytes), dataKeyB64}
	_, err = UploadOwnerData(stub, serviceSubgroup, args)
	test_utils.AssertTrue(t, err == nil, "Expected UploadOwnerData to succeed")
	mstub.MockTransactionEnd("5")

	// register org2
	mstub.MockTransactionStart("6")
	stub = cached_stub.NewCachedStub(mstub)
	org2 := test_utils.CreateTestGroup("org2")
	org2Bytes, _ := json.Marshal(&org2)
	_, err = RegisterOrg(stub, org2, []string{string(org2Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterOrg to succeed")
	mstub.MockTransactionEnd("6")

	// register datatype2
	mstub.MockTransactionStart("7")
	stub = cached_stub.NewCachedStub(mstub)
	datatype2 := Datatype{DatatypeID: "datatype2", Description: "datatype2"}
	datatype2Bytes, _ := json.Marshal(&datatype2)
	org2Caller, _ := user_mgmt.GetUserData(stub, org2, org2.ID, true, true)
	_, err = RegisterDatatype(stub, org2Caller, []string{string(datatype2Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterDatatype to succeed")
	mstub.MockTransactionEnd("7")

	// register service2
	mstub.MockTransactionStart("8")
	stub = cached_stub.NewCachedStub(mstub, true, true, true)
	serviceDatatype2 := GenerateServiceDatatypeForTesting("datatype2", "service2", []string{consentOptionWrite, consentOptionRead})
	service2 := GenerateServiceForTesting("service2", "org2", []ServiceDatatype{serviceDatatype2})
	service2Bytes, _ := json.Marshal(&service2)
	_, err = RegisterService(stub, org2Caller, []string{string(service2Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterService to succeed")
	mstub.MockTransactionEnd("8")

	// service1 give consent to service2
	mstub.MockTransactionStart("9")
	stub = cached_stub.NewCachedStub(mstub, true, true, true)
	consent := Consent{}
	consent.Owner = "service1"
	consent.Service = "service1"
	consent.Datatype = "datatype1"
	consent.Target = "service2"
	consent.Option = []string{consentOptionWrite, consentOptionRead}
	consent.Timestamp = time.Now().Unix()
	consent.Expiration = 0
	consentBytes, _ := json.Marshal(&consent)
	consentKey := test_utils.GenerateSymKey()
	consentKeyB64 := crypto.EncodeToB64String(consentKey)
	service1Subgroup, _ := user_mgmt.GetUserData(stub, org1Caller, "service1", true, true)
	_, err = PutConsentOwnerData(stub, service1Subgroup, []string{string(consentBytes), consentKeyB64})
	test_utils.AssertTrue(t, err == nil, "Expected PutConsentOwnerData to succeed")
	mstub.MockTransactionEnd("9")

	// get consent back to make sure it was added
	mstub.MockTransactionStart("10")
	stub = cached_stub.NewCachedStub(mstub)
	consentResultBytes, err := GetConsent(stub, service1Subgroup, []string{"service1", "service2", "datatype1"})
	test_utils.AssertTrue(t, err == nil, "Expected GetConsent to succeed")
	test_utils.AssertTrue(t, consentResultBytes != nil, "Expected GetConsent to succeed")
	consentResult := Consent{}
	json.Unmarshal(consentResultBytes, &consentResult)
	test_utils.AssertTrue(t, consentResult.Owner == "service1", "Got consent owner correctly")
	test_utils.AssertTrue(t, consentResult.Target == "service2", "Got consent target correctly")
	mstub.MockTransactionEnd("10")

	// download as service2
	mstub.MockTransactionStart("10")
	stub = cached_stub.NewCachedStub(mstub)
	service2Subgroup, _ := user_mgmt.GetUserData(stub, org2Caller, "service2", true, true)
	dataResultBytes, err := DownloadOwnerDataWithConsent(stub, service2Subgroup, []string{"service2", "service1", "datatype1", "false", strconv.FormatInt(0, 10), strconv.FormatInt(0, 10), "1000", strconv.FormatInt(time.Now().Unix(), 10)})
	test_utils.AssertTrue(t, err == nil, "Expected DownloadOwnerDataWithConsent to succeed")
	test_utils.AssertTrue(t, dataResultBytes != nil, "Expected DownloadOwnerDataWithConsent to succeed")
	dataResult := OwnerDataResultWithLog{}
	json.Unmarshal(dataResultBytes, &dataResult)
	test_utils.AssertTrue(t, len(dataResult.OwnerDatas) == 2, "Got owner data correctly")
	test_utils.AssertTrue(t, dataResult.OwnerDatas[0].Owner == "service1", "Got owner data correctly")
	mstub.MockTransactionEnd("10")

	// download latest only
	mstub.MockTransactionStart("13")
	stub = cached_stub.NewCachedStub(mstub)
	dataResultBytes, err = DownloadOwnerDataWithConsent(stub, service2Subgroup, []string{"service2", "service1", "datatype1", "true", strconv.FormatInt(0, 10), strconv.FormatInt(0, 10), "1000", strconv.FormatInt(time.Now().Unix(), 10)})
	test_utils.AssertTrue(t, err == nil, "Expected DownloadOwnerDataConsentToken to succeed")
	test_utils.AssertTrue(t, dataResultBytes != nil, "Expected DownloadOwnerDataConsentToken to succeed")
	json.Unmarshal(dataResultBytes, &dataResult)
	test_utils.AssertTrue(t, len(dataResult.OwnerDatas) == 1, "Got owner data correctly")
	test_utils.AssertTrue(t, dataResult.OwnerDatas[0].Owner == "service1", "Got owner data correctly")
	mstub.MockTransactionEnd("13")

	// create an org user
	mstub.MockTransactionStart("13")
	stub = cached_stub.NewCachedStub(mstub)
	orgUser1 := CreateTestSolutionUser("orgUser1")
	orgUser1.Org = "org2"
	orgUser1Bytes, _ := json.Marshal(&orgUser1)
	_, err = RegisterUser(stub, org2Caller, []string{string(orgUser1Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterUser to succeed")
	mstub.MockTransactionEnd("13")

	// put orgUser1 in org2
	mstub.MockTransactionStart("14")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = PutUserInOrg(stub, org2Caller, []string{orgUser1.ID, org2.ID, "false"})
	test_utils.AssertTrue(t, err == nil, "Expected PutUserInOrg to succeed")
	mstub.MockTransactionEnd("14")

	// download as orgUser1
	mstub.MockTransactionStart("10")
	stub = cached_stub.NewCachedStub(mstub)
	orgUser1Caller, _ := user_mgmt.GetUserData(stub, org2Caller, orgUser1.ID, true, true)
	dataResultBytes, err = DownloadOwnerDataWithConsent(stub, orgUser1Caller, []string{"service2", "service1", "datatype1", "false", strconv.FormatInt(0, 10), strconv.FormatInt(0, 10), "1000", strconv.FormatInt(time.Now().Unix(), 10)})
	test_utils.AssertTrue(t, err != nil, "Expected DownloadOwnerDataWithConsent to succeed")
	test_utils.AssertTrue(t, dataResultBytes == nil, "Expected DownloadOwnerDataWithConsent to succeed")
	mstub.MockTransactionEnd("10")

	// give org admin permission
	mstub.MockTransactionStart("16")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = AddPermissionOrgAdmin(stub, org2Caller, []string{orgUser1.ID, org2.ID})
	test_utils.AssertTrue(t, err == nil, "Expected AddPermissionOrgAdmin to succeed")
	mstub.MockTransactionEnd("16")

	// download as orgUser1 (org admin)
	mstub.MockTransactionStart("10")
	stub = cached_stub.NewCachedStub(mstub)
	orgUser1Caller, _ = user_mgmt.GetUserData(stub, org2Caller, orgUser1.ID, true, true)
	dataResultBytes, err = DownloadOwnerDataWithConsent(stub, orgUser1Caller, []string{"service2", "service1", "datatype1", "false", strconv.FormatInt(0, 10), strconv.FormatInt(0, 10), "1000", strconv.FormatInt(time.Now().Unix(), 10)})
	test_utils.AssertTrue(t, err == nil, "Expected DownloadOwnerDataWithConsent to succeed")
	test_utils.AssertTrue(t, dataResultBytes != nil, "Expected DownloadOwnerDataWithConsent to succeed")
	json.Unmarshal(dataResultBytes, &dataResult)
	test_utils.AssertTrue(t, len(dataResult.OwnerDatas) == 2, "Got owner data correctly")
	test_utils.AssertTrue(t, dataResult.OwnerDatas[0].Owner == "service1", "Got owner data correctly")
	mstub.MockTransactionEnd("10")

	// create another org user
	mstub.MockTransactionStart("19")
	stub = cached_stub.NewCachedStub(mstub)
	orgUser2 := CreateTestSolutionUser("orgUser2")
	orgUser2.Org = "org2"
	orgUser2Bytes, _ := json.Marshal(&orgUser2)
	_, err = RegisterUser(stub, org2Caller, []string{string(orgUser2Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterUser to succeed")
	mstub.MockTransactionEnd("19")

	// put orgUser2 in org2
	mstub.MockTransactionStart("20")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = PutUserInOrg(stub, org2Caller, []string{orgUser2.ID, org2.ID, "false"})
	test_utils.AssertTrue(t, err == nil, "Expected PutUserInOrg to succeed")
	mstub.MockTransactionEnd("20")

	// give service admin permission
	mstub.MockTransactionStart("21")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = AddPermissionServiceAdmin(stub, org2Caller, []string{orgUser2.ID, "service2"})
	test_utils.AssertTrue(t, err == nil, "Expected AddPermissionServiceAdmin to succeed")
	mstub.MockTransactionEnd("21")

	// download as orgUser2
	mstub.MockTransactionStart("10")
	stub = cached_stub.NewCachedStub(mstub)
	orgUser2Caller, _ := user_mgmt.GetUserData(stub, org2Caller, orgUser2.ID, true, true)
	dataResultBytes, err = DownloadOwnerDataWithConsent(stub, orgUser2Caller, []string{"service2", "service1", "datatype1", "false", strconv.FormatInt(0, 10), strconv.FormatInt(0, 10), "1000", strconv.FormatInt(time.Now().Unix(), 10)})
	test_utils.AssertTrue(t, err == nil, "Expected DownloadOwnerDataWithConsent to succeed")
	test_utils.AssertTrue(t, dataResultBytes != nil, "Expected DownloadOwnerDataWithConsent to succeed")
	json.Unmarshal(dataResultBytes, &dataResult)
	test_utils.AssertTrue(t, len(dataResult.OwnerDatas) == 2, "Got owner data correctly")
	test_utils.AssertTrue(t, dataResult.OwnerDatas[0].Owner == "service1", "Got owner data correctly")
	mstub.MockTransactionEnd("10")
}

func TestDownloadOwnerDataAsOwner(t *testing.T) {
	logger.SetLevel(shim.LogDebug)
	logger.Info("TestDownloadOwnerDataAsOwner function called")

	// create a MockStub
	mstub := SetupIndexesAndGetStub(t)

	// register admin user
	mstub.MockTransactionStart("t123")
	stub := cached_stub.NewCachedStub(mstub)
	init_common.Init(stub)
	err := SetupDataIndex(stub)
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

	// upload owner data
	mstub.MockTransactionStart("4")
	stub = cached_stub.NewCachedStub(mstub, true, true, true)
	ownerData := GenerateOwnerData("service1", "datatype1")
	ownerDataBytes, _ := json.Marshal(&ownerData)
	dataKey := test_utils.GenerateSymKey()
	dataKeyB64 := crypto.EncodeToB64String(dataKey)
	args := []string{string(ownerDataBytes), dataKeyB64}
	serviceSubgroup, _ := user_mgmt.GetUserData(stub, org1, "service1", true, true)
	_, err = UploadOwnerData(stub, serviceSubgroup, args)
	test_utils.AssertTrue(t, err == nil, "Expected UploadOwnerData to succeed")
	mstub.MockTransactionEnd("4")

	// upload owner data again
	mstub.MockTransactionStart("5")
	stub = cached_stub.NewCachedStub(mstub, true, true, true)
	ownerData = GenerateOwnerData("service1", "datatype1")
	ownerData.Timestamp++
	ownerDataBytes, _ = json.Marshal(&ownerData)
	dataKey = test_utils.GenerateSymKey()
	dataKeyB64 = crypto.EncodeToB64String(dataKey)
	args = []string{string(ownerDataBytes), dataKeyB64}
	_, err = UploadOwnerData(stub, serviceSubgroup, args)
	test_utils.AssertTrue(t, err == nil, "Expected UploadOwnerData to succeed")
	mstub.MockTransactionEnd("5")

	// download as default service admin
	mstub.MockTransactionStart("10")
	stub = cached_stub.NewCachedStub(mstub)
	dataResultBytes, err := DownloadOwnerDataAsOwner(stub, serviceSubgroup, []string{"service1", "datatype1", "false", strconv.FormatInt(0, 10), strconv.FormatInt(0, 10), "1000", strconv.FormatInt(time.Now().Unix(), 10)})
	test_utils.AssertTrue(t, err == nil, "Expected DownloadOwnerDataAsOwner to succeed")
	test_utils.AssertTrue(t, dataResultBytes != nil, "Expected DownloadOwnerDataAsOwner to succeed")
	dataResult := OwnerDataResultWithLog{}
	json.Unmarshal(dataResultBytes, &dataResult)
	test_utils.AssertTrue(t, len(dataResult.OwnerDatas) == 2, "Got owner data correctly")
	test_utils.AssertTrue(t, dataResult.OwnerDatas[0].Owner == "service1", "Got owner data correctly")
	mstub.MockTransactionEnd("10")

	// download latest only
	mstub.MockTransactionStart("13")
	stub = cached_stub.NewCachedStub(mstub)
	dataResultBytes, err = DownloadOwnerDataAsOwner(stub, serviceSubgroup, []string{"service1", "datatype1", "true", strconv.FormatInt(0, 10), strconv.FormatInt(0, 10), "1000", strconv.FormatInt(time.Now().Unix(), 10)})
	test_utils.AssertTrue(t, err == nil, "Expected DownloadOwnerDataAsOwner to succeed")
	test_utils.AssertTrue(t, dataResultBytes != nil, "Expected DownloadOwnerDataAsOwner to succeed")
	json.Unmarshal(dataResultBytes, &dataResult)
	test_utils.AssertTrue(t, len(dataResult.OwnerDatas) == 1, "Got owner data correctly")
	test_utils.AssertTrue(t, dataResult.OwnerDatas[0].Owner == "service1", "Got owner data correctly")
	mstub.MockTransactionEnd("13")

	// register org2
	mstub.MockTransactionStart("6")
	stub = cached_stub.NewCachedStub(mstub)
	org2 := test_utils.CreateTestGroup("org2")
	org2Bytes, _ := json.Marshal(&org2)
	_, err = RegisterOrg(stub, org2, []string{string(org2Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterOrg to succeed")
	mstub.MockTransactionEnd("6")

	// register datatype2
	mstub.MockTransactionStart("7")
	stub = cached_stub.NewCachedStub(mstub)
	datatype2 := Datatype{DatatypeID: "datatype2", Description: "datatype2"}
	datatype2Bytes, _ := json.Marshal(&datatype2)
	org2Caller, _ := user_mgmt.GetUserData(stub, org2, org2.ID, true, true)
	_, err = RegisterDatatype(stub, org2Caller, []string{string(datatype2Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterDatatype to succeed")
	mstub.MockTransactionEnd("7")

	// register service2
	mstub.MockTransactionStart("8")
	stub = cached_stub.NewCachedStub(mstub, true, true, true)
	serviceDatatype2 := GenerateServiceDatatypeForTesting("datatype2", "service2", []string{consentOptionWrite, consentOptionRead})
	service2 := GenerateServiceForTesting("service2", "org2", []ServiceDatatype{serviceDatatype2})
	service2Bytes, _ := json.Marshal(&service2)
	_, err = RegisterService(stub, org2Caller, []string{string(service2Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterService to succeed")
	mstub.MockTransactionEnd("8")

	// download as service2, should fail
	mstub.MockTransactionStart("10")
	stub = cached_stub.NewCachedStub(mstub)
	service2Subgroup, _ := user_mgmt.GetUserData(stub, org2Caller, "service2", true, true)
	dataResultBytes, err = DownloadOwnerDataAsOwner(stub, service2Subgroup, []string{"service1", "datatype1", "false", strconv.FormatInt(0, 10), strconv.FormatInt(0, 10), "1000", strconv.FormatInt(time.Now().Unix(), 10)})
	test_utils.AssertTrue(t, err != nil, "Expected DownloadOwnerDataAsOwner to fail")
	test_utils.AssertTrue(t, dataResultBytes == nil, "Expected DownloadOwnerDataAsOwner to succeed")
	mstub.MockTransactionEnd("10")

	// create an org user
	mstub.MockTransactionStart("13")
	stub = cached_stub.NewCachedStub(mstub)
	orgUser1 := CreateTestSolutionUser("orgUser1")
	orgUser1.Org = "org1"
	orgUser1Bytes, _ := json.Marshal(&orgUser1)
	_, err = RegisterUser(stub, org1Caller, []string{string(orgUser1Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterUser to succeed")
	mstub.MockTransactionEnd("13")

	// put orgUser1 in org1
	mstub.MockTransactionStart("14")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = PutUserInOrg(stub, org1Caller, []string{orgUser1.ID, org1.ID, "false"})
	test_utils.AssertTrue(t, err == nil, "Expected PutUserInOrg to succeed")
	mstub.MockTransactionEnd("14")

	// download as orgUser1 (non admin)
	mstub.MockTransactionStart("10")
	stub = cached_stub.NewCachedStub(mstub)
	orgUser1Caller, _ := user_mgmt.GetUserData(stub, org1Caller, orgUser1.ID, true, true)
	dataResultBytes, err = DownloadOwnerDataAsOwner(stub, orgUser1Caller, []string{"service1", "datatype1", "false", strconv.FormatInt(0, 10), strconv.FormatInt(0, 10), "1000", strconv.FormatInt(time.Now().Unix(), 10)})
	test_utils.AssertTrue(t, err != nil, "Expected DownloadOwnerDataAsOwner to fail")
	test_utils.AssertTrue(t, dataResultBytes == nil, "Expected DownloadOwnerDataAsOwner to fail")
	mstub.MockTransactionEnd("10")

	// give org admin permission
	mstub.MockTransactionStart("16")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = AddPermissionOrgAdmin(stub, org1Caller, []string{orgUser1.ID, org1.ID})
	test_utils.AssertTrue(t, err == nil, "Expected AddPermissionOrgAdmin to succeed")
	mstub.MockTransactionEnd("16")

	// download as orgUser1 (org admin)
	mstub.MockTransactionStart("10")
	stub = cached_stub.NewCachedStub(mstub)
	orgUser1Caller, _ = user_mgmt.GetUserData(stub, org1Caller, orgUser1.ID, true, true)
	dataResultBytes, err = DownloadOwnerDataAsOwner(stub, orgUser1Caller, []string{"service1", "datatype1", "false", strconv.FormatInt(0, 10), strconv.FormatInt(0, 10), "1000", strconv.FormatInt(time.Now().Unix(), 10)})
	test_utils.AssertTrue(t, err == nil, "Expected DownloadOwnerDataAsOwner to succeed")
	test_utils.AssertTrue(t, dataResultBytes != nil, "Expected DownloadOwnerDataAsOwner to succeed")
	json.Unmarshal(dataResultBytes, &dataResult)
	test_utils.AssertTrue(t, len(dataResult.OwnerDatas) == 2, "Got owner data correctly")
	test_utils.AssertTrue(t, dataResult.OwnerDatas[0].Owner == "service1", "Got owner data correctly")
	mstub.MockTransactionEnd("10")

	// create another org user
	mstub.MockTransactionStart("19")
	stub = cached_stub.NewCachedStub(mstub)
	orgUser2 := CreateTestSolutionUser("orgUser2")
	orgUser2.Org = "org1"
	orgUser2Bytes, _ := json.Marshal(&orgUser2)
	_, err = RegisterUser(stub, org1Caller, []string{string(orgUser2Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterUser to succeed")
	mstub.MockTransactionEnd("19")

	// put orgUser2 in org2
	mstub.MockTransactionStart("20")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = PutUserInOrg(stub, org1Caller, []string{orgUser2.ID, org1.ID, "false"})
	test_utils.AssertTrue(t, err == nil, "Expected PutUserInOrg to succeed")
	mstub.MockTransactionEnd("20")

	// give service admin permission
	mstub.MockTransactionStart("21")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = AddPermissionServiceAdmin(stub, org1Caller, []string{orgUser2.ID, "service1"})
	test_utils.AssertTrue(t, err == nil, "Expected AddPermissionServiceAdmin to succeed")
	mstub.MockTransactionEnd("21")

	// download as orgUser2
	mstub.MockTransactionStart("10")
	stub = cached_stub.NewCachedStub(mstub)
	orgUser2Caller, _ := user_mgmt.GetUserData(stub, org1Caller, orgUser2.ID, true, true)
	dataResultBytes, err = DownloadOwnerDataAsOwner(stub, orgUser2Caller, []string{"service1", "datatype1", "false", strconv.FormatInt(0, 10), strconv.FormatInt(0, 10), "1000", strconv.FormatInt(time.Now().Unix(), 10)})
	test_utils.AssertTrue(t, err == nil, "Expected DownloadOwnerDataAsOwner to succeed")
	test_utils.AssertTrue(t, dataResultBytes != nil, "Expected DownloadOwnerDataAsOwner to succeed")
	json.Unmarshal(dataResultBytes, &dataResult)
	test_utils.AssertTrue(t, len(dataResult.OwnerDatas) == 2, "Got owner data correctly")
	test_utils.AssertTrue(t, dataResult.OwnerDatas[0].Owner == "service1", "Got owner data correctly")
	mstub.MockTransactionEnd("10")
}
