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
	"common/bchcls/cached_stub"
	"common/bchcls/crypto"
	"common/bchcls/init_common"
	"common/bchcls/test_utils"
	"common/bchcls/user_mgmt"
	"encoding/json"
	"strconv"
	"testing"
	"time"

	"github.com/hyperledger/fabric/core/chaincode/shim"
)

func TestGetLogs(t *testing.T) {
	logger.SetLevel(shim.LogDebug)
	logger.Info("TestGetLogs function called")

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

	// register patient1
	mstub.MockTransactionStart("4")
	stub = cached_stub.NewCachedStub(mstub)
	patient1 := test_utils.CreateTestUser("patient1")
	patient1Bytes, _ := json.Marshal(&patient1)
	_, err = user_mgmt.RegisterUser(stub, org1Caller, []string{string(patient1Bytes), "false"})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterUser to succeed")
	mstub.MockTransactionEnd("4")

	// register patient2
	mstub.MockTransactionStart("4")
	stub = cached_stub.NewCachedStub(mstub)
	patient2 := test_utils.CreateTestUser("patient2")
	patient2Bytes, _ := json.Marshal(&patient2)
	_, err = user_mgmt.RegisterUser(stub, org1Caller, []string{string(patient2Bytes), "false"})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterUser to succeed")
	mstub.MockTransactionEnd("4")

	// enroll patient
	mstub.MockTransactionStart("5")
	stub = cached_stub.NewCachedStub(mstub)
	service1Subgroup, _ := user_mgmt.GetUserData(stub, org1, "service1", true, true)
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

	// get logs as default service admin
	mstub.MockTransactionStart("7")
	stub = cached_stub.NewCachedStub(mstub)
	logBytes, err := GetLogs(stub, service1Subgroup, []string{"", "patient1", "", "", "", "", strconv.FormatInt(0, 10), strconv.FormatInt(0, 10), "false", "20"})
	test_utils.AssertTrue(t, err == nil, "Expected GetLogs to succeed")
	logsOMR := []Log{}
	json.Unmarshal(logBytes, &logsOMR)
	// only log so far should be consent log
	test_utils.AssertTrue(t, len(logsOMR) == 1, "Got logs correctly")
	test_utils.AssertTrue(t, logsOMR[0].Owner != "", "Expected Owner")
	mstub.MockTransactionEnd("7")

	// upload patient data as default service admin
	mstub.MockTransactionStart("8")
	stub = cached_stub.NewCachedStub(mstub)
	patientData := GeneratePatientData("patient1", "datatype1", "service1")
	patientDataBytes, _ := json.Marshal(&patientData)
	dataKey := test_utils.GenerateSymKey()
	dataKeyB64 := crypto.EncodeToB64String(dataKey)
	args := []string{string(patientDataBytes), dataKeyB64}
	_, err = UploadUserData(stub, service1Subgroup, args)
	test_utils.AssertTrue(t, err == nil, "Expected UploadUserData to succeed")
	mstub.MockTransactionEnd("8")

	// upload patient data2 as default service admin
	mstub.MockTransactionStart("10")
	stub = cached_stub.NewCachedStub(mstub)
	patientData = GeneratePatientData("patient1", "datatype1", "service1")
	// Need to add 1 so timestamps are different
	patientData.Timestamp++
	patientDataBytes, _ = json.Marshal(&patientData)
	dataKey = test_utils.GenerateSymKey()
	dataKeyB64 = crypto.EncodeToB64String(dataKey)
	args = []string{string(patientDataBytes), dataKeyB64}
	_, err = UploadUserData(stub, service1Subgroup, args)
	test_utils.AssertTrue(t, err == nil, "Expected UploadUserData to succeed")
	mstub.MockTransactionEnd("10")

	// upload patient data3 as service admin
	mstub.MockTransactionStart("11")
	stub = cached_stub.NewCachedStub(mstub)
	patientData = GeneratePatientData("patient1", "datatype1", "service1")
	// Need to add 2 so timestamps are different
	patientData.Timestamp += 2
	patientDataBytes, _ = json.Marshal(&patientData)
	dataKey = test_utils.GenerateSymKey()
	dataKeyB64 = crypto.EncodeToB64String(dataKey)
	args = []string{string(patientDataBytes), dataKeyB64}
	_, err = UploadUserData(stub, service1Subgroup, args)
	test_utils.AssertTrue(t, err == nil, "Expected UploadUserData to succeed")
	mstub.MockTransactionEnd("11")

	// get logs as default service admin
	mstub.MockTransactionStart("12")
	stub = cached_stub.NewCachedStub(mstub)
	logBytes, err = GetLogs(stub, service1Subgroup, []string{"", "patient1", "", "", "", "", strconv.FormatInt(0, 10), strconv.FormatInt(0, 10), "false", "20"})
	test_utils.AssertTrue(t, err == nil, "Expected GetLogs to succeed")
	json.Unmarshal(logBytes, &logsOMR)
	// 4 logs - add consent + 3 upload patient data
	test_utils.AssertTrue(t, len(logsOMR) == 4, "Got logs correctly")
	test_utils.AssertTrue(t, logsOMR[0].Owner != "", "Expected Owner")
	mstub.MockTransactionEnd("12")

	// get logs as patient1
	mstub.MockTransactionStart("13")
	stub = cached_stub.NewCachedStub(mstub)
	logBytes, err = GetLogs(stub, patient1Caller, []string{"", "patient1", "", "", "", "", strconv.FormatInt(0, 10), strconv.FormatInt(0, 10), "false", "20"})
	test_utils.AssertTrue(t, err == nil, "Expected GetLogs to succeed")
	json.Unmarshal(logBytes, &logsOMR)
	// 4 logs - add consent + 3 upload patient data
	test_utils.AssertTrue(t, len(logsOMR) == 4, "Got logs correctly")
	test_utils.AssertTrue(t, logsOMR[0].Owner != "", "Expected Owner")
	mstub.MockTransactionEnd("13")

	// get logs as patient2
	mstub.MockTransactionStart("13")
	stub = cached_stub.NewCachedStub(mstub)
	patient2Caller, err := user_mgmt.GetUserData(stub, patient2, "patient2", true, true)
	logBytes, err = GetLogs(stub, patient2Caller, []string{"", "patient1", "", "", "", "", strconv.FormatInt(0, 10), strconv.FormatInt(0, 10), "false", "20"})
	test_utils.AssertTrue(t, err == nil, "Expected GetLogs to succeed")
	json.Unmarshal(logBytes, &logsOMR)
	// 4 logs - add consent + 3 upload patient data
	test_utils.AssertTrue(t, len(logsOMR) == 0, "Got logs correctly")
	mstub.MockTransactionEnd("13")

	// get logs as auditor1
	mstub.MockTransactionStart("13")
	stub = cached_stub.NewCachedStub(mstub)
	auditorPlatformUser, err := convertToPlatformUser(stub, auditor)
	test_utils.AssertTrue(t, err == nil, "Expected convertToPlatformUser to succeed")
	auditorCaller, _ := user_mgmt.GetUserData(stub, auditorPlatformUser, auditor.ID, true, true)
	logBytes, err = GetLogs(stub, auditorCaller, []string{"", "patient1", "", "", "", "", strconv.FormatInt(0, 10), strconv.FormatInt(0, 10), "false", "20"})
	test_utils.AssertTrue(t, err == nil, "Expected GetLogs to succeed")
	json.Unmarshal(logBytes, &logsOMR)
	// 4 logs - add consent + 3 upload patient data
	test_utils.AssertTrue(t, len(logsOMR) == 4, "Got logs correctly")
	test_utils.AssertTrue(t, logsOMR[0].Caller != "", "Expected Owner")
	test_utils.AssertTrue(t, logsOMR[0].Service != "", "Expected Service")
	test_utils.AssertTrue(t, logsOMR[0].Owner != "", "Expected Owner")
	test_utils.AssertTrue(t, logsOMR[0].Target != "", "Expected Target")
	mstub.MockTransactionEnd("13")

	mstub.MockTransactionStart("14")
	stub = cached_stub.NewCachedStub(mstub)
	logBytes, err = GetLogs(stub, patient1Caller, []string{"", "patient1", "service1", "", "", "", strconv.FormatInt(0, 10), strconv.FormatInt(0, 10), "false", "20"})
	test_utils.AssertTrue(t, err == nil, "Expected GetLogs to succeed")
	json.Unmarshal(logBytes, &logsOMR)
	// 4 logs - add consent + 3 upload patient data
	test_utils.AssertTrue(t, len(logsOMR) == 4, "Got logs correctly")
	test_utils.AssertTrue(t, logsOMR[0].Owner != "", "Expected Owner")
	mstub.MockTransactionEnd("14")

	mstub.MockTransactionStart("15")
	stub = cached_stub.NewCachedStub(mstub)
	logBytes, err = GetLogs(stub, patient1Caller, []string{"", "patient1", "service1", "datatype1", "", "", strconv.FormatInt(0, 10), strconv.FormatInt(0, 10), "false", "20"})
	test_utils.AssertTrue(t, err == nil, "Expected GetLogs to succeed")
	json.Unmarshal(logBytes, &logsOMR)
	// 4 logs - add consent + 3 upload patient data
	test_utils.AssertTrue(t, len(logsOMR) == 4, "Got logs correctly")
	test_utils.AssertTrue(t, logsOMR[0].Owner != "", "Expected Owner")
	mstub.MockTransactionEnd("15")

	// Download patient data as service admin
	mstub.MockTransactionStart("16")
	stub = cached_stub.NewCachedStub(mstub)
	dataResultBytes, err := DownloadUserData(stub, service1Subgroup, []string{"service1", "patient1", "datatype1", "false", strconv.FormatInt(0, 10), strconv.FormatInt(0, 10), "1000", strconv.FormatInt(time.Now().Unix(), 10)})
	test_utils.AssertTrue(t, err == nil, "Expected DownloadUserData to succeed")
	test_utils.AssertTrue(t, dataResultBytes != nil, "Expected DownloadUserData to succeed")
	dataResult := OwnerDataResultWithLog{}
	json.Unmarshal(dataResultBytes, &dataResult)
	test_utils.AssertTrue(t, len(dataResult.OwnerDatas) == 3, "Got patient data correctly")
	test_utils.AssertTrue(t, dataResult.OwnerDatas[0].Owner == "patient1", "Got owner data correctly")
	mstub.MockTransactionEnd("16")

	// Download patient data as patient
	mstub.MockTransactionStart("17")
	stub = cached_stub.NewCachedStub(mstub)
	dataResultBytes, err = DownloadUserData(stub, patient1Caller, []string{"service1", "patient1", "datatype1", "false", strconv.FormatInt(0, 10), strconv.FormatInt(0, 10), "1000", strconv.FormatInt(time.Now().Unix(), 10)})
	test_utils.AssertTrue(t, err == nil, "Expected DownloadUserData to succeed")
	test_utils.AssertTrue(t, dataResultBytes != nil, "Expected DownloadUserData to succeed")
	json.Unmarshal(dataResultBytes, &dataResult)
	test_utils.AssertTrue(t, len(dataResult.OwnerDatas) == 3, "Got patient data correctly")
	test_utils.AssertTrue(t, dataResult.OwnerDatas[0].Owner == "patient1", "Got owner data correctly")
	mstub.MockTransactionEnd("17")

	// create an org user
	mstub.MockTransactionStart("18")
	stub = cached_stub.NewCachedStub(mstub)
	orgUser1 := CreateTestSolutionUser("orgUser1")
	orgUser1.Org = "org1"
	orgUser1Bytes, _ := json.Marshal(&orgUser1)
	_, err = RegisterUser(stub, org1Caller, []string{string(orgUser1Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterUser to succeed")
	mstub.MockTransactionEnd("18")

	// put orgUser1 in org1
	mstub.MockTransactionStart("19")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = PutUserInOrg(stub, org1Caller, []string{orgUser1.ID, org1.ID, "false"})
	test_utils.AssertTrue(t, err == nil, "Expected PutUserInOrg to succeed")
	mstub.MockTransactionEnd("19")

	// give org admin permission
	mstub.MockTransactionStart("20")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = AddPermissionOrgAdmin(stub, org1Caller, []string{orgUser1.ID, org1.ID})
	test_utils.AssertTrue(t, err == nil, "Expected AddPermissionOrgAdmin to succeed")
	mstub.MockTransactionEnd("20")

	// get logs as orgUser1
	mstub.MockTransactionStart("21")
	stub = cached_stub.NewCachedStub(mstub)
	orgUser1Caller, _ := user_mgmt.GetUserData(stub, org1Caller, orgUser1.ID, true, true)
	logBytes, err = GetLogs(stub, orgUser1Caller, []string{"", "patient1", "", "", "", "", strconv.FormatInt(0, 10), strconv.FormatInt(0, 10), "false", "20"})
	test_utils.AssertTrue(t, err == nil, "Expected GetLogs to succeed")
	json.Unmarshal(logBytes, &logsOMR)
	// 4 logs - 1 add consent + 3 upload patient data
	// The 2 download patient data logs will not show up here because this is GO test and we need JS to invoke query transaction
	test_utils.AssertTrue(t, len(logsOMR) == 4, "Got logs correctly")
	test_utils.AssertTrue(t, logsOMR[0].Owner != "", "Expected Owner")
	mstub.MockTransactionEnd("21")

	// create another org user
	mstub.MockTransactionStart("22")
	stub = cached_stub.NewCachedStub(mstub)
	orgUser2 := CreateTestSolutionUser("orgUser2")
	orgUser2.Org = "org1"
	orgUser2Bytes, _ := json.Marshal(&orgUser2)
	_, err = RegisterUser(stub, org1Caller, []string{string(orgUser2Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterUser to succeed")
	mstub.MockTransactionEnd("22")

	// put orgUser2 in org1
	mstub.MockTransactionStart("23")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = PutUserInOrg(stub, org1Caller, []string{orgUser2.ID, org1.ID, "false"})
	test_utils.AssertTrue(t, err == nil, "Expected PutUserInOrg to succeed")
	mstub.MockTransactionEnd("23")

	// give service admin permission
	mstub.MockTransactionStart("24")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = AddPermissionServiceAdmin(stub, org1Caller, []string{orgUser2.ID, "service1"})
	test_utils.AssertTrue(t, err == nil, "Expected AddPermissionServiceAdmin to succeed")
	mstub.MockTransactionEnd("24")

	// get logs as orgUser2
	mstub.MockTransactionStart("25")
	stub = cached_stub.NewCachedStub(mstub)
	orgUser2Caller, _ := user_mgmt.GetUserData(stub, org1Caller, orgUser2.ID, true, true)
	logBytes, err = GetLogs(stub, orgUser2Caller, []string{"", "patient1", "", "", "", "", strconv.FormatInt(0, 10), strconv.FormatInt(0, 10), "false", "20"})
	test_utils.AssertTrue(t, err == nil, "Expected GetLogs to succeed")
	json.Unmarshal(logBytes, &logsOMR)
	// 4 logs - 1 add consent + 3 upload patient data
	// The 2 download patient data logs will not show up here because this is GO test and we need JS to invoke query transaction
	test_utils.AssertTrue(t, len(logsOMR) == 4, "Got logs correctly")
	test_utils.AssertTrue(t, logsOMR[0].Owner != "", "Expected Owner")
	mstub.MockTransactionEnd("25")

	// get logs by contract owner/target
	mstub.MockTransactionStart("26")
	stub = cached_stub.NewCachedStub(mstub)
	logBytes, err = GetLogs(stub, patient1Caller, []string{"", "", "", "", "", "patient1", strconv.FormatInt(0, 10), strconv.FormatInt(0, 10), "false", "20"})
	test_utils.AssertTrue(t, err == nil, "Expected GetLogs to succeed")
	json.Unmarshal(logBytes, &logsOMR)
	// 4 logs - add consent + 3 upload patient data
	test_utils.AssertTrue(t, len(logsOMR) == 4, "Expected 4 log records")
	test_utils.AssertTrue(t, logsOMR[0].Owner != "", "Expected Owner")
	mstub.MockTransactionEnd("26")

	// get logs by contract owner/target
	mstub.MockTransactionStart("27")
	stub = cached_stub.NewCachedStub(mstub)
	logBytes, err = GetLogs(stub, patient1Caller, []string{"", "", "", "", "", "notRegisteredPatient", strconv.FormatInt(0, 10), strconv.FormatInt(0, 10), "false", "20"})
	test_utils.AssertTrue(t, err == nil, "Expected GetLogs to succeed")
	json.Unmarshal(logBytes, &logsOMR)
	// 4 logs - add consent + 3 upload patient data
	test_utils.AssertTrue(t, len(logsOMR) == 0, "Expected 0 log records.")
	mstub.MockTransactionEnd("27")

	// register org
	mstub.MockTransactionStart("createContract")
	stub = cached_stub.NewCachedStub(mstub)
	org2 := test_utils.CreateTestGroup("org2")
	org2Bytes, _ := json.Marshal(&org2)
	_, err = RegisterOrg(stub, org2, []string{string(org2Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterOrg to succeed")
	mstub.MockTransactionEnd("createContract")

	//  register datatype
	mstub.MockTransactionStart("createContract")
	stub = cached_stub.NewCachedStub(mstub)
	datatype2 := Datatype{DatatypeID: "datatype2", Description: "datatype2"}
	datatype2Bytes, _ := json.Marshal(&datatype2)
	org2Caller, _ := user_mgmt.GetUserData(stub, org2, org2.ID, true, true)
	_, err = RegisterDatatype(stub, org2Caller, []string{string(datatype2Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterDatatype to succeed")
	mstub.MockTransactionEnd("createContract")

	//  register service
	mstub.MockTransactionStart("createContract")
	stub = cached_stub.NewCachedStub(mstub, true, true, true)
	serviceDatatype2 := GenerateServiceDatatypeForTesting("datatype2", "service2", []string{consentOptionWrite, consentOptionRead})
	service2 := GenerateServiceForTesting("service2", "org2", []ServiceDatatype{serviceDatatype2})
	service2Bytes, _ := json.Marshal(&service2)
	_, err = RegisterService(stub, org2Caller, []string{string(service2Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterService to succeed")
	mstub.MockTransactionEnd("createContract")

	mstub.MockTransactionStart("createContract")
	stub = cached_stub.NewCachedStub(mstub)
	contract := GenerateContractTest("contract1", "org1", "service1", "org2", "service2")
	contractBytes, _ := json.Marshal(&contract)
	contractKey := test_utils.GenerateSymKey()
	contractKeyB64 := crypto.EncodeToB64String(contractKey)
	args = []string{string(contractBytes), contractKeyB64}
	ownerService1Subgroup, _ := user_mgmt.GetUserData(stub, org1Caller, "service1", true, true)
	_, err = CreateContract(stub, ownerService1Subgroup, args)
	mstub.MockTransactionEnd("createContract")

	// get logs by contract owner/target
	mstub.MockTransactionStart("28")
	stub = cached_stub.NewCachedStub(mstub)
	org1Caller, _ = user_mgmt.GetUserData(stub, org1, org1.ID, true, true)
	logBytes, err = GetLogs(stub, org1Caller, []string{"", "", "", "", "org2", "", strconv.FormatInt(0, 10), strconv.FormatInt(0, 10), "false", "20"})
	test_utils.AssertTrue(t, err == nil, "Expected GetLogs to succeed")
	json.Unmarshal(logBytes, &logsOMR)
	// 4 logs - add consent + 3 upload patient data
	test_utils.AssertTrue(t, len(logsOMR) == 1, "Expected 1 log records.")
	test_utils.AssertTrue(t, logsOMR[0].ContractOwnerOrg == "org2", "Expected contract owner org")
	test_utils.AssertTrue(t, logsOMR[0].ContractOwnerService == "service1", "Expected contract owner service")
	mstub.MockTransactionEnd("28")

	// get logs by contract owner/target org should return empty logs
	mstub.MockTransactionStart("29")
	stub = cached_stub.NewCachedStub(mstub)
	org1Caller, _ = user_mgmt.GetUserData(stub, org1, org1.ID, true, true)
	logBytes, err = GetLogs(stub, org1Caller, []string{"", "", "", "", "org3", "", strconv.FormatInt(0, 10), strconv.FormatInt(0, 10), "false", "20"})
	test_utils.AssertTrue(t, err == nil, "Expected GetLogs to succeed")
	json.Unmarshal(logBytes, &logsOMR)
	// 4 logs - add consent + 3 upload patient data
	test_utils.AssertTrue(t, len(logsOMR) == 0, "Expected 0 log records.")
	mstub.MockTransactionEnd("29")

	// Get logs by contract ID. Should return logs for contract.
	mstub.MockTransactionStart("searchByContractId")
	stub = cached_stub.NewCachedStub(mstub)
	org1Caller, _ = user_mgmt.GetUserData(stub, org1, org1.ID, true, true)
	logBytes, err = GetLogs(stub, org1Caller, []string{"contract1", "", "", "", "", "", strconv.FormatInt(0, 10), strconv.FormatInt(0, 10), "false", "20"})
	test_utils.AssertTrue(t, err == nil, "Expected GetLogs to succeed")
	json.Unmarshal(logBytes, &logsOMR)
	// 4 logs - add consent + 3 upload patient data
	test_utils.AssertTrue(t, len(logsOMR) == 1, "Expected 1 log records")
	test_utils.AssertTrue(t, logsOMR[0].Contract == "contract1", "Expected contract ID to match search parameter")
	mstub.MockTransactionEnd("searchByContractId")
}

func TestGetLogs_Offchain(t *testing.T) {
	logger.SetLevel(shim.LogDebug)
	logger.Info("TestGetLogs_Offchain function called")

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

	// register patient1
	mstub.MockTransactionStart("4")
	stub = cached_stub.NewCachedStub(mstub)
	patient1 := test_utils.CreateTestUser("patient1")
	patient1Bytes, _ := json.Marshal(&patient1)
	_, err = user_mgmt.RegisterUser(stub, org1Caller, []string{string(patient1Bytes), "false"})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterUser to succeed")
	mstub.MockTransactionEnd("4")

	// register patient2
	mstub.MockTransactionStart("4")
	stub = cached_stub.NewCachedStub(mstub)
	patient2 := test_utils.CreateTestUser("patient2")
	patient2Bytes, _ := json.Marshal(&patient2)
	_, err = user_mgmt.RegisterUser(stub, org1Caller, []string{string(patient2Bytes), "false"})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterUser to succeed")
	mstub.MockTransactionEnd("4")

	// enroll patient
	mstub.MockTransactionStart("5")
	stub = cached_stub.NewCachedStub(mstub)
	service1Subgroup, _ := user_mgmt.GetUserData(stub, org1, "service1", true, true)
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

	// get logs as default service admin
	mstub.MockTransactionStart("7")
	stub = cached_stub.NewCachedStub(mstub)
	logBytes, err := GetLogs(stub, service1Subgroup, []string{"", "patient1", "", "", "", "", strconv.FormatInt(0, 10), strconv.FormatInt(0, 10), "false", "20"})
	test_utils.AssertTrue(t, err == nil, "Expected GetLogs to succeed")
	logsOMR := []Log{}
	json.Unmarshal(logBytes, &logsOMR)
	// only log so far should be consent log
	test_utils.AssertTrue(t, len(logsOMR) == 1, "Got logs correctly")
	test_utils.AssertTrue(t, logsOMR[0].Owner != "", "Expected Owner")
	mstub.MockTransactionEnd("7")

	// upload patient data as default service admin
	mstub.MockTransactionStart("8")
	stub = cached_stub.NewCachedStub(mstub)
	patientData := GeneratePatientData("patient1", "datatype1", "service1")
	patientDataBytes, _ := json.Marshal(&patientData)
	dataKey := test_utils.GenerateSymKey()
	dataKeyB64 := crypto.EncodeToB64String(dataKey)
	args := []string{string(patientDataBytes), dataKeyB64}
	_, err = UploadUserData(stub, service1Subgroup, args)
	test_utils.AssertTrue(t, err == nil, "Expected UploadUserData to succeed")
	mstub.MockTransactionEnd("8")

	// upload patient data2 as default service admin
	mstub.MockTransactionStart("10")
	stub = cached_stub.NewCachedStub(mstub)
	patientData = GeneratePatientData("patient1", "datatype1", "service1")
	// Need to add 1 so timestamps are different
	patientData.Timestamp++
	patientDataBytes, _ = json.Marshal(&patientData)
	dataKey = test_utils.GenerateSymKey()
	dataKeyB64 = crypto.EncodeToB64String(dataKey)
	args = []string{string(patientDataBytes), dataKeyB64}
	_, err = UploadUserData(stub, service1Subgroup, args)
	test_utils.AssertTrue(t, err == nil, "Expected UploadUserData to succeed")
	mstub.MockTransactionEnd("10")

	// upload patient data3 as service admin
	mstub.MockTransactionStart("11")
	stub = cached_stub.NewCachedStub(mstub)
	patientData = GeneratePatientData("patient1", "datatype1", "service1")
	// Need to add 2 so timestamps are different
	patientData.Timestamp += 2
	patientDataBytes, _ = json.Marshal(&patientData)
	dataKey = test_utils.GenerateSymKey()
	dataKeyB64 = crypto.EncodeToB64String(dataKey)
	args = []string{string(patientDataBytes), dataKeyB64}
	_, err = UploadUserData(stub, service1Subgroup, args)
	test_utils.AssertTrue(t, err == nil, "Expected UploadUserData to succeed")
	mstub.MockTransactionEnd("11")

	// get logs as default service admin
	mstub.MockTransactionStart("12")
	stub = cached_stub.NewCachedStub(mstub)
	logBytes, err = GetLogs(stub, service1Subgroup, []string{"", "patient1", "", "", "", "", strconv.FormatInt(0, 10), strconv.FormatInt(0, 10), "false", "20"})
	test_utils.AssertTrue(t, err == nil, "Expected GetLogs to succeed")
	json.Unmarshal(logBytes, &logsOMR)
	// 4 logs - add consent + 3 upload patient data
	test_utils.AssertTrue(t, len(logsOMR) == 4, "Got logs correctly")
	test_utils.AssertTrue(t, logsOMR[0].Owner != "", "Expected Owner")
	mstub.MockTransactionEnd("12")

	// get logs as patient1
	mstub.MockTransactionStart("13")
	stub = cached_stub.NewCachedStub(mstub)
	logBytes, err = GetLogs(stub, patient1Caller, []string{"", "patient1", "", "", "", "", strconv.FormatInt(0, 10), strconv.FormatInt(0, 10), "false", "20"})
	test_utils.AssertTrue(t, err == nil, "Expected GetLogs to succeed")
	json.Unmarshal(logBytes, &logsOMR)
	// 4 logs - add consent + 3 upload patient data
	test_utils.AssertTrue(t, len(logsOMR) == 4, "Got logs correctly")
	test_utils.AssertTrue(t, logsOMR[0].Owner != "", "Expected Owner")
	mstub.MockTransactionEnd("13")

	// get logs as patient2
	mstub.MockTransactionStart("13")
	stub = cached_stub.NewCachedStub(mstub)
	patient2Caller, err := user_mgmt.GetUserData(stub, patient2, "patient2", true, true)
	logBytes, err = GetLogs(stub, patient2Caller, []string{"", "patient1", "", "", "", "", strconv.FormatInt(0, 10), strconv.FormatInt(0, 10), "false", "20"})
	test_utils.AssertTrue(t, err == nil, "Expected GetLogs to succeed")
	json.Unmarshal(logBytes, &logsOMR)
	// 4 logs - add consent + 3 upload patient data
	test_utils.AssertTrue(t, len(logsOMR) == 0, "Got logs correctly")
	mstub.MockTransactionEnd("13")

	// get logs as auditor1
	mstub.MockTransactionStart("13")
	stub = cached_stub.NewCachedStub(mstub)
	auditorPlatformUser, err := convertToPlatformUser(stub, auditor)
	test_utils.AssertTrue(t, err == nil, "Expected convertToPlatformUser to succeed")
	auditorCaller, _ := user_mgmt.GetUserData(stub, auditorPlatformUser, auditor.ID, true, true)
	logBytes, err = GetLogs(stub, auditorCaller, []string{"", "patient1", "", "", "", "", strconv.FormatInt(0, 10), strconv.FormatInt(0, 10), "false", "20"})
	test_utils.AssertTrue(t, err == nil, "Expected GetLogs to succeed")
	json.Unmarshal(logBytes, &logsOMR)
	// 4 logs - add consent + 3 upload patient data
	test_utils.AssertTrue(t, len(logsOMR) == 4, "Got logs correctly")
	test_utils.AssertTrue(t, logsOMR[0].Caller != "", "Expected Owner")
	test_utils.AssertTrue(t, logsOMR[0].Service != "", "Expected Service")
	test_utils.AssertTrue(t, logsOMR[0].Owner != "", "Expected Owner")
	test_utils.AssertTrue(t, logsOMR[0].Target != "", "Expected Target")
	mstub.MockTransactionEnd("13")

	mstub.MockTransactionStart("14")
	stub = cached_stub.NewCachedStub(mstub)
	logBytes, err = GetLogs(stub, patient1Caller, []string{"", "patient1", "service1", "", "", "", strconv.FormatInt(0, 10), strconv.FormatInt(0, 10), "false", "20"})
	test_utils.AssertTrue(t, err == nil, "Expected GetLogs to succeed")
	json.Unmarshal(logBytes, &logsOMR)
	// 4 logs - add consent + 3 upload patient data
	test_utils.AssertTrue(t, len(logsOMR) == 4, "Got logs correctly")
	test_utils.AssertTrue(t, logsOMR[0].Owner != "", "Expected Owner")
	mstub.MockTransactionEnd("14")

	mstub.MockTransactionStart("15")
	stub = cached_stub.NewCachedStub(mstub)
	logBytes, err = GetLogs(stub, patient1Caller, []string{"", "patient1", "service1", "datatype1", "", "", strconv.FormatInt(0, 10), strconv.FormatInt(0, 10), "false", "20"})
	test_utils.AssertTrue(t, err == nil, "Expected GetLogs to succeed")
	json.Unmarshal(logBytes, &logsOMR)
	// 4 logs - add consent + 3 upload patient data
	test_utils.AssertTrue(t, len(logsOMR) == 4, "Got logs correctly")
	test_utils.AssertTrue(t, logsOMR[0].Owner != "", "Expected Owner")
	mstub.MockTransactionEnd("15")

	// Download patient data as service admin
	mstub.MockTransactionStart("16")
	stub = cached_stub.NewCachedStub(mstub)
	dataResultBytes, err := DownloadUserData(stub, service1Subgroup, []string{"service1", "patient1", "datatype1", "false", strconv.FormatInt(0, 10), strconv.FormatInt(0, 10), "1000", strconv.FormatInt(time.Now().Unix(), 10)})
	test_utils.AssertTrue(t, err == nil, "Expected DownloadUserData to succeed")
	test_utils.AssertTrue(t, dataResultBytes != nil, "Expected DownloadUserData to succeed")
	dataResult := OwnerDataResultWithLog{}
	json.Unmarshal(dataResultBytes, &dataResult)
	test_utils.AssertTrue(t, len(dataResult.OwnerDatas) == 3, "Got patient data correctly")
	test_utils.AssertTrue(t, dataResult.OwnerDatas[0].Owner == "patient1", "Got owner data correctly")
	mstub.MockTransactionEnd("16")

	// Download patient data as patient
	mstub.MockTransactionStart("17")
	stub = cached_stub.NewCachedStub(mstub)
	dataResultBytes, err = DownloadUserData(stub, patient1Caller, []string{"service1", "patient1", "datatype1", "false", strconv.FormatInt(0, 10), strconv.FormatInt(0, 10), "1000", strconv.FormatInt(time.Now().Unix(), 10)})
	test_utils.AssertTrue(t, err == nil, "Expected DownloadUserData to succeed")
	test_utils.AssertTrue(t, dataResultBytes != nil, "Expected DownloadUserData to succeed")
	json.Unmarshal(dataResultBytes, &dataResult)
	test_utils.AssertTrue(t, len(dataResult.OwnerDatas) == 3, "Got patient data correctly")
	test_utils.AssertTrue(t, dataResult.OwnerDatas[0].Owner == "patient1", "Got owner data correctly")
	mstub.MockTransactionEnd("17")

	// create an org user
	mstub.MockTransactionStart("18")
	stub = cached_stub.NewCachedStub(mstub)
	orgUser1 := CreateTestSolutionUser("orgUser1")
	orgUser1.Org = "org1"
	orgUser1Bytes, _ := json.Marshal(&orgUser1)
	_, err = RegisterUser(stub, org1Caller, []string{string(orgUser1Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterUser to succeed")
	mstub.MockTransactionEnd("18")

	// put orgUser1 in org1
	mstub.MockTransactionStart("19")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = PutUserInOrg(stub, org1Caller, []string{orgUser1.ID, org1.ID, "false"})
	test_utils.AssertTrue(t, err == nil, "Expected PutUserInOrg to succeed")
	mstub.MockTransactionEnd("19")

	// give org admin permission
	mstub.MockTransactionStart("20")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = AddPermissionOrgAdmin(stub, org1Caller, []string{orgUser1.ID, org1.ID})
	test_utils.AssertTrue(t, err == nil, "Expected AddPermissionOrgAdmin to succeed")
	mstub.MockTransactionEnd("20")

	// get logs as orgUser1
	mstub.MockTransactionStart("21")
	stub = cached_stub.NewCachedStub(mstub)
	orgUser1Caller, _ := user_mgmt.GetUserData(stub, org1Caller, orgUser1.ID, true, true)
	logBytes, err = GetLogs(stub, orgUser1Caller, []string{"", "patient1", "", "", "", "", strconv.FormatInt(0, 10), strconv.FormatInt(0, 10), "false", "20"})
	test_utils.AssertTrue(t, err == nil, "Expected GetLogs to succeed")
	json.Unmarshal(logBytes, &logsOMR)
	// 4 logs - 1 add consent + 3 upload patient data
	// The 2 download patient data logs will not show up here because this is GO test and we need JS to invoke query transaction
	test_utils.AssertTrue(t, len(logsOMR) == 4, "Got logs correctly")
	test_utils.AssertTrue(t, logsOMR[0].Owner != "", "Expected Owner")
	mstub.MockTransactionEnd("21")

	// create another org user
	mstub.MockTransactionStart("22")
	stub = cached_stub.NewCachedStub(mstub)
	orgUser2 := CreateTestSolutionUser("orgUser2")
	orgUser2.Org = "org1"
	orgUser2Bytes, _ := json.Marshal(&orgUser2)
	_, err = RegisterUser(stub, org1Caller, []string{string(orgUser2Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterUser to succeed")
	mstub.MockTransactionEnd("22")

	// put orgUser2 in org1
	mstub.MockTransactionStart("23")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = PutUserInOrg(stub, org1Caller, []string{orgUser2.ID, org1.ID, "false"})
	test_utils.AssertTrue(t, err == nil, "Expected PutUserInOrg to succeed")
	mstub.MockTransactionEnd("23")

	// give service admin permission
	mstub.MockTransactionStart("24")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = AddPermissionServiceAdmin(stub, org1Caller, []string{orgUser2.ID, "service1"})
	test_utils.AssertTrue(t, err == nil, "Expected AddPermissionServiceAdmin to succeed")
	mstub.MockTransactionEnd("24")

	// get logs as orgUser2
	mstub.MockTransactionStart("25")
	stub = cached_stub.NewCachedStub(mstub)
	orgUser2Caller, _ := user_mgmt.GetUserData(stub, org1Caller, orgUser2.ID, true, true)
	logBytes, err = GetLogs(stub, orgUser2Caller, []string{"", "patient1", "", "", "", "", strconv.FormatInt(0, 10), strconv.FormatInt(0, 10), "false", "20"})
	test_utils.AssertTrue(t, err == nil, "Expected GetLogs to succeed")
	json.Unmarshal(logBytes, &logsOMR)
	// 4 logs - 1 add consent + 3 upload patient data
	// The 2 download patient data logs will not show up here because this is GO test and we need JS to invoke query transaction
	test_utils.AssertTrue(t, len(logsOMR) == 4, "Got logs correctly")
	test_utils.AssertTrue(t, logsOMR[0].Owner != "", "Expected Owner")
	mstub.MockTransactionEnd("25")
}
