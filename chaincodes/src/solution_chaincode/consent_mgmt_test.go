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
	"common/bchcls/user_mgmt"
	"common/bchcls/user_mgmt/user_groups"
	"common/bchcls/utils"
	"encoding/json"
	"net/url"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/hyperledger/fabric/core/chaincode/shim"
)

func setupDatastore(mstub *test_utils.NewMockStub, caller data_model.User, datastoreConnectionID string) error {
	mstub.MockTransactionStart("t1")
	stub := cached_stub.NewCachedStub(mstub, true, true)

	username := "admin"
	password := "pass"
	database := "test"
	host := "http://127.0.0.1:9080"
	// Get values from environment variables
	if !utils.IsStringEmpty(os.Getenv("CLOUDANT_USERNAME")) {
		username = os.Getenv("CLOUDANT_USERNAME")
	}
	if !utils.IsStringEmpty(os.Getenv("CLOUDANT_PASSWORD")) {
		password = os.Getenv("CLOUDANT_PASSWORD")
	}
	if !utils.IsStringEmpty(os.Getenv("CLOUDANT_DATABASE")) {
		database = os.Getenv("CLOUDANT_DATABASE")
	}
	if !utils.IsStringEmpty(os.Getenv("CLOUDANT_HOST")) {
		host = os.Getenv("CLOUDANT_HOST")
	}

	params := url.Values{}
	params.Add("username", username)
	params.Add("password", password)
	params.Add("database", database)
	params.Add("host", host)

	err := SetupDatastore(stub, caller, []string{datastoreConnectionID, username, password, database, host})
	mstub.MockTransactionEnd("t1")

	return err
}

func TestPutConsentPatientData(t *testing.T) {
	logger.SetLevel(shim.LogDebug)
	logger.Info("TestPutConsentPatientData function called")

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

	// update consent to DENY
	mstub.MockTransactionStart("8")
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
	mstub.MockTransactionEnd("8")

	// get consent back to make sure it was added
	mstub.MockTransactionStart("9")
	stub = cached_stub.NewCachedStub(mstub)
	consentResultBytes, err = GetConsent(stub, patient1Caller, []string{"patient1", "service1", "datatype1"})
	test_utils.AssertTrue(t, err == nil, "Expected GetConsent to succeed")
	test_utils.AssertTrue(t, consentResultBytes != nil, "Expected GetConsent to succeed")
	json.Unmarshal(consentResultBytes, &consentResult)
	test_utils.AssertSetsEqual(t, []string{consentOptionDeny}, consentResult.Option)
	mstub.MockTransactionEnd("9")

	// try adding consent as service admin
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub, true, true, true)
	service1AdminCaller, err := user_mgmt.GetUserData(stub, org1Caller, "service1", true, true)
	test_utils.AssertTrue(t, err == nil, "Expected GetUserData to succeed")
	_, err = PutConsentPatientData(stub, service1AdminCaller, []string{string(consentBytes)})
	test_utils.AssertTrue(t, err != nil, "Expected PutConsentPatientData to fail")
	mstub.MockTransactionEnd("t123")

	// try adding consent as org admin
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub, true, true, true)
	_, err = PutConsentPatientData(stub, org1Caller, []string{string(consentBytes)})
	test_utils.AssertTrue(t, err != nil, "Expected PutConsentPatientData to fail")
	mstub.MockTransactionEnd("t123")

	// register new patient
	mstub.MockTransactionStart("10")
	stub = cached_stub.NewCachedStub(mstub)
	patient2 := test_utils.CreateTestUser("patient2")
	patient2Bytes, _ := json.Marshal(&patient2)
	_, err = user_mgmt.RegisterUser(stub, org1Caller, []string{string(patient2Bytes), "false"})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterUser to succeed")
	mstub.MockTransactionEnd("10")

	// enroll to service
	mstub.MockTransactionStart("11")
	stub = cached_stub.NewCachedStub(mstub)
	enrollment2 := GenerateEnrollmentTest("patient2", "service1")
	enrollment2Bytes, _ := json.Marshal(&enrollment2)
	enrollmentKey2 := test_utils.GenerateSymKey()
	enrollmentKey2B64 := crypto.EncodeToB64String(enrollmentKey2)
	_, err = EnrollPatient(stub, org1Caller, []string{string(enrollment2Bytes), enrollmentKey2B64})
	test_utils.AssertTrue(t, err == nil, "Expected EnrollPatient to succeed")
	mstub.MockTransactionEnd("11")

	// try updating consent as wrong patient
	mstub.MockTransactionStart("12")
	stub = cached_stub.NewCachedStub(mstub, true, true, true)
	consent = Consent{}
	consent.Owner = "patient1"
	consent.Service = "service1"
	consent.Target = "service1"
	consent.Datatype = "datatype1"
	consent.Option = []string{consentOptionRead}
	consent.Timestamp = time.Now().Unix()
	consent.Expiration = 0
	consentBytes, _ = json.Marshal(&consent)
	consentKey = test_utils.GenerateSymKey()
	consentKeyB64 = crypto.EncodeToB64String(consentKey)
	patient2Caller, err := user_mgmt.GetUserData(stub, patient2, "patient2", true, true)
	_, err = PutConsentPatientData(stub, patient2Caller, []string{string(consentBytes), consentKeyB64})
	test_utils.AssertTrue(t, err != nil, "Expected PutConsentPatientData to fail")
	mstub.MockTransactionEnd("12")

	// patient2 give consent for himself
	mstub.MockTransactionStart("13")
	stub = cached_stub.NewCachedStub(mstub, true, true, true)
	consent = Consent{}
	consent.Owner = "patient2"
	consent.Service = "service1"
	consent.Target = "service1"
	consent.Datatype = "datatype1"
	consent.Option = []string{consentOptionRead}
	consent.Timestamp = time.Now().Unix()
	consent.Expiration = 0
	consentBytes, _ = json.Marshal(&consent)
	consentKey = test_utils.GenerateSymKey()
	consentKeyB64 = crypto.EncodeToB64String(consentKey)
	_, err = PutConsentPatientData(stub, patient2Caller, []string{string(consentBytes), consentKeyB64})
	test_utils.AssertTrue(t, err == nil, "Expected PutConsentPatientData to succeed")
	mstub.MockTransactionEnd("13")

	// get patient2's consent back
	mstub.MockTransactionStart("14")
	stub = cached_stub.NewCachedStub(mstub)
	consentResultBytes, err = GetConsent(stub, patient2Caller, []string{"patient2", "service1", "datatype1"})
	test_utils.AssertTrue(t, err == nil, "Expected GetConsent to succeed")
	mstub.MockTransactionEnd("14")
}

func TestPutConsentPatientData_ReferenceService(t *testing.T) {
	logger.SetLevel(shim.LogDebug)
	logger.Info("TestPutConsentPatientData_ReferenceService function called")

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

	// register datatype2
	mstub.MockTransactionStart("6")
	stub = cached_stub.NewCachedStub(mstub)
	datatype2 := Datatype{DatatypeID: "datatype2", Description: "datatype2"}
	datatype2Bytes, _ := json.Marshal(&datatype2)
	_, err = RegisterDatatype(stub, org1Caller, []string{string(datatype2Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterDatatype to succeed")
	mstub.MockTransactionEnd("6")

	// register "reference service" that uses datatype2, service2 belongs to org1
	mstub.MockTransactionStart("7")
	stub = cached_stub.NewCachedStub(mstub, true, true, true)
	serviceDatatype2 := GenerateServiceDatatypeForTesting("datatype2", "service2", []string{consentOptionWrite, consentOptionRead})
	service2 := GenerateServiceForTesting("service2", "org1", []ServiceDatatype{serviceDatatype2})
	service2Bytes, _ := json.Marshal(&service2)
	_, err = RegisterService(stub, org1Caller, []string{string(service2Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterService to succeed")
	mstub.MockTransactionEnd("7")

	// patient give consent
	mstub.MockTransactionStart("8")
	stub = cached_stub.NewCachedStub(mstub, true, true, true)
	consent := Consent{}
	consent.Owner = "patient1"
	consent.Service = "service1"
	consent.Target = "service2"
	consent.Datatype = "datatype1"
	consent.Option = []string{consentOptionRead}
	consent.Timestamp = time.Now().Unix()
	consent.Expiration = 0
	consentBytes, _ := json.Marshal(&consent)
	consentKey := test_utils.GenerateSymKey()
	consentKeyB64 := crypto.EncodeToB64String(consentKey)
	patient1Caller, err := user_mgmt.GetUserData(stub, patient1, "patient1", true, true)
	_, err = PutConsentPatientData(stub, patient1Caller, []string{string(consentBytes), consentKeyB64})
	test_utils.AssertTrue(t, err == nil, "Expected PutConsentPatientData to succeed")
	mstub.MockTransactionEnd("8")

	// get consent back as patient
	mstub.MockTransactionStart("9")
	stub = cached_stub.NewCachedStub(mstub)
	consentResultBytes, err := GetConsent(stub, patient1Caller, []string{"patient1", "service2", "datatype1"})
	test_utils.AssertTrue(t, err == nil, "Expected GetConsent to succeed")
	test_utils.AssertTrue(t, consentResultBytes != nil, "Expected GetConsent to succeed")
	consentResult := Consent{}
	json.Unmarshal(consentResultBytes, &consentResult)
	test_utils.AssertTrue(t, consentResult.Owner == "patient1", "Got consent owner correctly")
	test_utils.AssertTrue(t, consentResult.Target == "service2", "Got consent target correctly")
	mstub.MockTransactionEnd("9")

	// get consent back as service target
	mstub.MockTransactionStart("9")
	stub = cached_stub.NewCachedStub(mstub)
	service2AdminCaller, err := user_mgmt.GetUserData(stub, org1Caller, "service2", true, true)
	consentResultBytes, err = GetConsent(stub, service2AdminCaller, []string{"patient1", "service2", "datatype1"})
	test_utils.AssertTrue(t, err == nil, "Expected GetConsent to succeed")
	test_utils.AssertTrue(t, consentResultBytes != nil, "Expected GetConsent to succeed")
	json.Unmarshal(consentResultBytes, &consentResult)
	test_utils.AssertTrue(t, consentResult.Owner == "patient1", "Got consent owner correctly")
	test_utils.AssertTrue(t, consentResult.Target == "service2", "Got consent target correctly")
	mstub.MockTransactionEnd("9")

	// get consent back as org admin of service target
	mstub.MockTransactionStart("9")
	stub = cached_stub.NewCachedStub(mstub)
	consentResultBytes, err = GetConsent(stub, org1Caller, []string{"patient1", "service2", "datatype1"})
	test_utils.AssertTrue(t, err == nil, "Expected GetConsent to succeed")
	test_utils.AssertTrue(t, consentResultBytes != nil, "Expected GetConsent to succeed")
	json.Unmarshal(consentResultBytes, &consentResult)
	test_utils.AssertTrue(t, consentResult.Owner == "patient1", "Got consent owner correctly")
	test_utils.AssertTrue(t, consentResult.Target == "service2", "Got consent target correctly")
	mstub.MockTransactionEnd("9")

	// get consent back as a service without access
	mstub.MockTransactionStart("9")
	stub = cached_stub.NewCachedStub(mstub)
	service1AdminCaller, err := user_mgmt.GetUserData(stub, org1Caller, "service1", true, true)
	consentResultBytes, err = GetConsent(stub, service1AdminCaller, []string{"patient1", "service2", "datatype1"})
	test_utils.AssertTrue(t, err != nil, "Expected GetConsent to fail")
	test_utils.AssertTrue(t, consentResultBytes == nil, "Expected GetConsent to fail")
	mstub.MockTransactionEnd("9")

	// update consent to DENY
	mstub.MockTransactionStart("10")
	stub = cached_stub.NewCachedStub(mstub, true, true, true)
	consent = Consent{}
	consent.Owner = "patient1"
	consent.Service = "service1"
	consent.Target = "service2"
	consent.Datatype = "datatype1"
	consent.Option = []string{consentOptionDeny}
	consent.Timestamp = time.Now().Unix()
	consent.Expiration = 0
	consentBytes, _ = json.Marshal(&consent)
	_, err = PutConsentPatientData(stub, patient1Caller, []string{string(consentBytes)})
	test_utils.AssertTrue(t, err == nil, "Expected PutConsentPatientData to succeed")
	mstub.MockTransactionEnd("10")

	// get consent back as patient
	mstub.MockTransactionStart("11")
	stub = cached_stub.NewCachedStub(mstub)
	consentResultBytes, err = GetConsent(stub, patient1Caller, []string{"patient1", "service2", "datatype1"})
	test_utils.AssertTrue(t, err == nil, "Expected GetConsent to succeed")
	test_utils.AssertTrue(t, consentResultBytes != nil, "Expected GetConsent to succeed")
	json.Unmarshal(consentResultBytes, &consentResult)
	test_utils.AssertSetsEqual(t, []string{consentOptionDeny}, consentResult.Option)
	mstub.MockTransactionEnd("11")

	// try adding consent as service1 admin
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub, true, true, true)
	test_utils.AssertTrue(t, err == nil, "Expected GetUserData to succeed")
	_, err = PutConsentPatientData(stub, service1AdminCaller, []string{string(consentBytes)})
	test_utils.AssertTrue(t, err != nil, "Expected PutConsentPatientData to fail")
	mstub.MockTransactionEnd("t123")

	// try adding consent as org admin
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub, true, true, true)
	_, err = PutConsentPatientData(stub, org1Caller, []string{string(consentBytes)})
	test_utils.AssertTrue(t, err != nil, "Expected PutConsentPatientData to fail")
	mstub.MockTransactionEnd("t123")

	// try adding consent as service2 admin
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub, true, true, true)
	_, err = PutConsentPatientData(stub, service2AdminCaller, []string{string(consentBytes)})
	test_utils.AssertTrue(t, err != nil, "Expected PutConsentPatientData to fail")
	mstub.MockTransactionEnd("t123")

	// register reference service that belongs to a different org
	// register org
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	org2 := test_utils.CreateTestGroup("org2")
	org2Bytes, _ := json.Marshal(&org2)
	_, err = RegisterOrg(stub, org2, []string{string(org2Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterOrg to succeed")
	mstub.MockTransactionEnd("t123")

	//  register service
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub, true, true, true)
	org2Caller, _ := user_mgmt.GetUserData(stub, org2, org2.ID, true, true)
	o2s1 := GenerateServiceForTesting("o2s1", "org2", []ServiceDatatype{serviceDatatype1})
	o2s1Bytes, _ := json.Marshal(&o2s1)
	_, err = RegisterService(stub, org2Caller, []string{string(o2s1Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterService to succeed")
	mstub.MockTransactionEnd("t123")

	// enroll patient1 to o2s1
	mstub.MockTransactionStart("5")
	stub = cached_stub.NewCachedStub(mstub)
	enrollment2 := GenerateEnrollmentTest("patient1", "o2s1")
	enrollment2Bytes, _ := json.Marshal(&enrollment2)
	enrollmentKey2 := test_utils.GenerateSymKey()
	enrollmentKey2B64 := crypto.EncodeToB64String(enrollmentKey2)
	_, err = EnrollPatient(stub, org2Caller, []string{string(enrollment2Bytes), enrollmentKey2B64})
	test_utils.AssertTrue(t, err == nil, "Expected EnrollPatient to succeed")
	mstub.MockTransactionEnd("5")

	// try adding consent as o2s1 admin
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub, true, true, true)
	consent = Consent{}
	consent.Owner = "patient1"
	consent.Service = "service1"
	consent.Target = "o2s1"
	consent.Datatype = "datatype1"
	consent.Option = []string{consentOptionRead}
	consent.Timestamp = time.Now().Unix()
	consent.Expiration = 0
	consentBytes, _ = json.Marshal(&consent)
	consentKey = test_utils.GenerateSymKey()
	consentKeyB64 = crypto.EncodeToB64String(consentKey)
	o2s1AdminCaller, err := user_mgmt.GetUserData(stub, org2Caller, "o2s1", true, true)
	test_utils.AssertTrue(t, err == nil, "Expected GetUserData to succeed")
	_, err = PutConsentPatientData(stub, o2s1AdminCaller, []string{string(consentBytes), consentKeyB64})
	test_utils.AssertTrue(t, err != nil, "Expected PutConsentPatientData to fail")
	mstub.MockTransactionEnd("t123")

	// try adding consent as org admin
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub, true, true, true)
	_, err = PutConsentPatientData(stub, org2Caller, []string{string(consentBytes), consentKeyB64})
	test_utils.AssertTrue(t, err != nil, "Expected PutConsentPatientData to fail")
	mstub.MockTransactionEnd("t123")

	// try adding consent as patient1
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub, true, true, true)
	_, err = PutConsentPatientData(stub, patient1Caller, []string{string(consentBytes), consentKeyB64})
	test_utils.AssertTrue(t, err == nil, "Expected PutConsentPatientData to succeed")
	mstub.MockTransactionEnd("t123")

	// get consent back as service target
	mstub.MockTransactionStart("9")
	stub = cached_stub.NewCachedStub(mstub)
	consentResultBytes, err = GetConsent(stub, o2s1AdminCaller, []string{"patient1", "o2s1", "datatype1"})
	test_utils.AssertTrue(t, err == nil, "Expected GetConsent to succeed")
	test_utils.AssertTrue(t, consentResultBytes != nil, "Expected GetConsent to succeed")
	json.Unmarshal(consentResultBytes, &consentResult)
	test_utils.AssertTrue(t, consentResult.Owner == "patient1", "Got consent owner correctly")
	test_utils.AssertTrue(t, consentResult.Target == "o2s1", "Got consent target correctly")
	mstub.MockTransactionEnd("9")

	// get consent back as org admin of service target
	mstub.MockTransactionStart("9")
	stub = cached_stub.NewCachedStub(mstub)
	consentResultBytes, err = GetConsent(stub, org2Caller, []string{"patient1", "o2s1", "datatype1"})
	test_utils.AssertTrue(t, err == nil, "Expected GetConsent to succeed")
	test_utils.AssertTrue(t, consentResultBytes != nil, "Expected GetConsent to succeed")
	json.Unmarshal(consentResultBytes, &consentResult)
	test_utils.AssertTrue(t, consentResult.Owner == "patient1", "Got consent owner correctly")
	test_utils.AssertTrue(t, consentResult.Target == "o2s1", "Got consent target correctly")
	mstub.MockTransactionEnd("9")

	// try to get consent as service1 admin, service 1 is original service of datatype1
	mstub.MockTransactionStart("9")
	stub = cached_stub.NewCachedStub(mstub)
	consentResultBytes, err = GetConsent(stub, service1AdminCaller, []string{"patient1", "o2s1", "datatype1"})
	test_utils.AssertTrue(t, err != nil, "Expected GetConsent to fail")
	test_utils.AssertTrue(t, consentResultBytes == nil, "Expected GetConsent to fail")
	mstub.MockTransactionEnd("9")
}

func TestPutConsentOffChain(t *testing.T) {
	logger.SetLevel(shim.LogDebug)
	logger.Info("TestPutConsentOffChain function called")

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
	serviceDatatype1 := GenerateServiceDatatypeForTesting("datatype1", "di-service1", []string{consentOptionWrite, consentOptionRead})
	service1 := GenerateServiceForTesting("di-service1", "org1", []ServiceDatatype{serviceDatatype1})
	service1Bytes, _ := json.Marshal(&service1)
	_, err = RegisterService(stub, org1Caller, []string{string(service1Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterService to succeed")
	mstub.MockTransactionEnd("t123")

	// register patient
	mstub.MockTransactionStart("4")
	stub = cached_stub.NewCachedStub(mstub)
	patient1 := test_utils.CreateTestUser("di-patient1")
	patient1Bytes, _ := json.Marshal(&patient1)
	_, err = user_mgmt.RegisterUser(stub, org1Caller, []string{string(patient1Bytes), "false"})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterUser to succeed")
	mstub.MockTransactionEnd("4")

	// register unrelated patient
	mstub.MockTransactionStart("5")
	stub = cached_stub.NewCachedStub(mstub)
	patient2 := test_utils.CreateTestUser("di-patient2")
	patient2Bytes, _ := json.Marshal(&patient2)
	_, err = user_mgmt.RegisterUser(stub, org1Caller, []string{string(patient2Bytes), "false"})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterUser to succeed")
	mstub.MockTransactionEnd("5")

	// enroll patient
	mstub.MockTransactionStart("6")
	stub = cached_stub.NewCachedStub(mstub)
	enrollment1 := GenerateEnrollmentTest("di-patient1", "di-service1")
	enrollment1Bytes, _ := json.Marshal(&enrollment1)
	enrollmentKey1 := test_utils.GenerateSymKey()
	enrollmentKey1B64 := crypto.EncodeToB64String(enrollmentKey1)
	_, err = EnrollPatient(stub, org1Caller, []string{string(enrollment1Bytes), enrollmentKey1B64})
	test_utils.AssertTrue(t, err == nil, "Expected EnrollPatient to succeed")
	mstub.MockTransactionEnd("6")

	// patient give consent
	mstub.MockTransactionStart("7")
	stub = cached_stub.NewCachedStub(mstub, true, true, true)
	consent := Consent{}
	consent.Owner = "di-patient1"
	consent.Service = "di-service1"
	consent.Target = "di-service1"
	consent.Datatype = "datatype1"
	consent.Option = []string{consentOptionWrite, consentOptionRead}
	consent.Timestamp = time.Now().Unix()
	consent.Expiration = time.Now().Unix() + 50000
	consentBytes, _ := json.Marshal(&consent)
	consentKey := test_utils.GenerateSymKey()
	consentKeyB64 := crypto.EncodeToB64String(consentKey)
	patient1Caller, err := user_mgmt.GetUserData(stub, patient1, "di-patient1", true, true)
	_, err = PutConsentPatientData(stub, patient1Caller, []string{string(consentBytes), consentKeyB64})
	test_utils.AssertTrue(t, err == nil, "Expected PutConsentPatientData to succeed")
	mstub.MockTransactionEnd("7")

	// get consent to make sure it was added
	mstub.MockTransactionStart("8")
	stub = cached_stub.NewCachedStub(mstub)
	consentResultBytes, err := GetConsent(stub, patient1Caller, []string{"di-patient1", "di-service1", "datatype1"})
	test_utils.AssertTrue(t, err == nil, "Expected GetConsent to succeed")
	test_utils.AssertTrue(t, consentResultBytes != nil, "Expected GetConsent to succeed")
	consentResult := Consent{}
	json.Unmarshal(consentResultBytes, &consentResult)
	test_utils.AssertTrue(t, consentResult.Owner == "di-patient1", "Expected public Owner")
	test_utils.AssertTrue(t, consentResult.Target == "di-service1", "Expected public Target")
	test_utils.AssertTrue(t, consentResult.Expiration > 0, "Expected private ExpirationDate")
	mstub.MockTransactionEnd("8")

	// verify data access for patient1
	mstub.MockTransactionStart("9")
	stub = cached_stub.NewCachedStub(mstub)
	assetManager := asset_mgmt.GetAssetManager(stub, patient1Caller)
	consentID := consent_mgmt.GetConsentID("datatype1", "di-service1", "di-patient1")
	consentAssetID, err := consent_mgmt.GetConsentAssetID(stub, consentID)
	test_utils.AssertTrue(t, err == nil, "Expected GetConsentAssetID consent to succeed")
	keyPath, err := GetKeyPath(stub, patient1Caller, consentAssetID)
	test_utils.AssertTrue(t, err == nil, "Expected GetKeyPath consent to succeed")
	consentAssetKey, err := assetManager.GetAssetKey(consentAssetID, keyPath)
	test_utils.AssertTrue(t, err == nil, "Expected GetAssetKey to succeed")
	consentAsset, err := assetManager.GetAsset(consentAssetID, consentAssetKey)
	test_utils.AssertTrue(t, err == nil, "Expected GetAsset to succeed")

	// patient1 has access to public data
	consentPublicData := data_model.Consent{}
	json.Unmarshal(consentAsset.PublicData, &consentPublicData)
	test_utils.AssertTrue(t, consentPublicData.ConsentID != "", "Expected ConsentID")
	test_utils.AssertTrue(t, consentPublicData.ConsentAssetID != "", "Expected ConsentAssetID")
	test_utils.AssertTrue(t, consentPublicData.CreatorID == "di-patient1", "Expected CreatorID")
	test_utils.AssertTrue(t, consentPublicData.OwnerID == "di-patient1", "Expected Owner")
	test_utils.AssertTrue(t, consentPublicData.TargetID == "di-service1", "Expected Target")
	test_utils.AssertTrue(t, consentPublicData.DatatypeID == "datatype1", "Expected DatatypeID")
	test_utils.AssertTrue(t, consentPublicData.ConnectionID == "cloudant1", "Expected ConnectionID")

	// patient1 has access to private data
	consentPrivateData := data_model.Consent{}
	json.Unmarshal(consentAsset.PrivateData, &consentPrivateData)
	test_utils.AssertTrue(t, consentPrivateData.Access == consentOptionWrite, "Expected Access")
	test_utils.AssertTrue(t, consentPrivateData.ExpirationDate == consent.Expiration, "Expected ExpirationDate")
	test_utils.AssertTrue(t, consentPrivateData.ConsentDate > 0, "Expected ConsentData")
	test_utils.AssertTrue(t, consentPrivateData.Data != nil, "Expected Data")
	mstub.MockTransactionEnd("9")

	// verify data access for patient2 (no access to private data)
	mstub.MockTransactionStart("10")
	stub = cached_stub.NewCachedStub(mstub)
	patient2Caller, err := user_mgmt.GetUserData(stub, patient2, "di-patient2", true, true)
	assetManager = asset_mgmt.GetAssetManager(stub, patient2Caller)
	consentID = consent_mgmt.GetConsentID("datatype1", "di-service1", "di-patient1")
	consentAssetID, err = consent_mgmt.GetConsentAssetID(stub, consentID)
	test_utils.AssertTrue(t, err == nil, "Expected GetConsentAssetID consent to succeed")
	keyPath, err = GetKeyPath(stub, patient2Caller, consentAssetID)
	test_utils.AssertTrue(t, err == nil, "Expected GetKeyPath consent to succeed")
	_, err = assetManager.GetAssetKey(consentAssetID, keyPath)
	test_utils.AssertTrue(t, err != nil, "Expected GetAssetKey to fail")
	consentAsset, err = assetManager.GetAsset(consentAssetID, data_model.Key{})
	test_utils.AssertTrue(t, err == nil, "Expected GetAsset to succeed")

	// patient2 has access to public data
	consentPublicData = data_model.Consent{}
	json.Unmarshal(consentAsset.PublicData, &consentPublicData)
	test_utils.AssertTrue(t, consentPublicData.ConsentID != "", "Expected ConsentID")
	test_utils.AssertTrue(t, consentPublicData.ConsentAssetID != "", "Expected ConsentAssetID")
	test_utils.AssertTrue(t, consentPublicData.CreatorID == "di-patient1", "Expected CreatorID")
	test_utils.AssertTrue(t, consentPublicData.OwnerID == "di-patient1", "Expected Owner")
	test_utils.AssertTrue(t, consentPublicData.TargetID == "di-service1", "Expected Target")
	test_utils.AssertTrue(t, consentPublicData.DatatypeID == "datatype1", "Expected DatatypeID")
	test_utils.AssertTrue(t, consentPublicData.ConnectionID == "cloudant1", "Expected ConnectionID")

	// patient2 does not have access to private data
	consentPrivateData = data_model.Consent{}
	json.Unmarshal(consentAsset.PrivateData, &consentPrivateData)
	test_utils.AssertTrue(t, consentPrivateData.Access == "", "Expected no Access")
	test_utils.AssertTrue(t, consentPrivateData.ExpirationDate == 0, "Expected no ExpirationDate")
	test_utils.AssertTrue(t, consentPrivateData.ConsentDate == 0, "Expected no ConsentData")
	test_utils.AssertTrue(t, consentPrivateData.Data == nil, "Expected no Data")
	mstub.MockTransactionEnd("10")

	// remove datastore connection
	mstub.MockTransactionStart("11")
	err = datastore_manager.DeleteDatastoreConnection(stub, systemAdmin, datastoreConnectionID)
	test_utils.AssertTrue(t, err == nil, "Expected DeleteDatastoreConnection to succeed")
	mstub.MockTransactionEnd("11")

	// attempt to get consent
	mstub.MockTransactionStart("12")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = GetConsent(stub, patient1Caller, []string{"patient1", "service1", "datatype1"})
	test_utils.AssertTrue(t, err != nil, "Expected GetConsent to fail")
	mstub.MockTransactionEnd("12")
}

func TestPutConsentOwnerData(t *testing.T) {
	logger.SetLevel(shim.LogDebug)
	logger.Info("TestPutConsentOwnerData function called")

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

	// register orgUser and make him service admin for service1
	mstub.MockTransactionStart("4")
	stub = cached_stub.NewCachedStub(mstub)
	service1Admin := test_utils.CreateTestUser("service1Admin")
	service1AdminBytes, _ := json.Marshal(&service1Admin)
	_, err = user_mgmt.RegisterUser(stub, org1Caller, []string{string(service1AdminBytes), "false"})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterUser to succeed")
	mstub.MockTransactionEnd("4")
	mstub.MockTransactionStart("5")
	stub = cached_stub.NewCachedStub(mstub)
	err = user_groups.PutUserInGroup(stub, org1Caller, "service1Admin", "service1", true)
	test_utils.AssertTrue(t, err == nil, "Expected PutUserInGroup to succeed")
	mstub.MockTransactionEnd("5")

	// register datatype2
	mstub.MockTransactionStart("6")
	stub = cached_stub.NewCachedStub(mstub)
	datatype2 := Datatype{DatatypeID: "datatype2", Description: "datatype2"}
	datatype2Bytes, _ := json.Marshal(&datatype2)
	_, err = RegisterDatatype(stub, org1Caller, []string{string(datatype2Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterDatatype to succeed")
	mstub.MockTransactionEnd("6")

	// register service2
	mstub.MockTransactionStart("7")
	stub = cached_stub.NewCachedStub(mstub, true, true, true)
	serviceDatatype2 := GenerateServiceDatatypeForTesting("datatype2", "service2", []string{consentOptionWrite, consentOptionRead})
	service2 := GenerateServiceForTesting("service2", "org1", []ServiceDatatype{serviceDatatype2})
	service2Bytes, _ := json.Marshal(&service2)
	_, err = RegisterService(stub, org1Caller, []string{string(service2Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterService to succeed")
	mstub.MockTransactionEnd("7")

	// register another orgUser and make him service admin for service2
	mstub.MockTransactionStart("8")
	stub = cached_stub.NewCachedStub(mstub)
	service2Admin := test_utils.CreateTestUser("service2Admin")
	service2AdminBytes, _ := json.Marshal(&service2Admin)
	_, err = user_mgmt.RegisterUser(stub, org1Caller, []string{string(service2AdminBytes), "false"})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterUser to succeed")
	mstub.MockTransactionEnd("8")
	mstub.MockTransactionStart("9")
	stub = cached_stub.NewCachedStub(mstub)
	err = user_groups.PutUserInGroup(stub, org1Caller, "service2Admin", "service2", true)
	test_utils.AssertTrue(t, err == nil, "Expected PutUserInGroup to succeed")
	mstub.MockTransactionEnd("9")

	// service 2 try to give consent to service 1, should fail because owner of consent is service1
	mstub.MockTransactionStart("10")
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
	service2AdminCaller, err := user_mgmt.GetUserData(stub, service2Admin, "service2Admin", true, true)
	_, err = PutConsentOwnerData(stub, service2AdminCaller, []string{string(consentBytes), consentKeyB64})
	test_utils.AssertTrue(t, err != nil, "Expected PutConsentOwnerData to fail")
	mstub.MockTransactionEnd("10")

	// service1 give consent to service2
	mstub.MockTransactionStart("11")
	stub = cached_stub.NewCachedStub(mstub, true, true, true)
	consent.Owner = "service1"
	consent.Service = "service1"
	consent.Datatype = "datatype1"
	consent.Target = "service2"
	consent.Option = []string{consentOptionWrite}
	consent.Timestamp = time.Now().Unix()
	consent.Expiration = 0
	consentBytes, _ = json.Marshal(&consent)
	consentKey = test_utils.GenerateSymKey()
	consentKeyB64 = crypto.EncodeToB64String(consentKey)
	service1AdminCaller, err := user_mgmt.GetUserData(stub, service1Admin, "service1Admin", true, true)
	test_utils.AssertTrue(t, err == nil, "Expected GetUserData to succeed")
	_, err = PutConsentOwnerData(stub, service1AdminCaller, []string{string(consentBytes), consentKeyB64})
	test_utils.AssertTrue(t, err == nil, "Expected PutConsentOwnerData to succeed")
	mstub.MockTransactionEnd("11")

	// get consent back as service1 admin to make sure it was added
	mstub.MockTransactionStart("12")
	stub = cached_stub.NewCachedStub(mstub)
	consentResultBytes, err := GetConsent(stub, service1AdminCaller, []string{"service1", "service2", "datatype1"})
	test_utils.AssertTrue(t, err == nil, "Expected GetConsent to succeed")
	test_utils.AssertTrue(t, consentResultBytes != nil, "Expected GetConsent to succeed")
	consentResult := Consent{}
	json.Unmarshal(consentResultBytes, &consentResult)
	test_utils.AssertTrue(t, consentResult.Owner == "service1", "Got consent owner correctly")
	test_utils.AssertTrue(t, consentResult.Target == "service2", "Got consent target correctly")
	mstub.MockTransactionEnd("12")

	// get consent back as service2 admin to make sure it was added
	mstub.MockTransactionStart("12")
	stub = cached_stub.NewCachedStub(mstub)
	consentResultBytes, err = GetConsent(stub, service2AdminCaller, []string{"service1", "service2", "datatype1"})
	test_utils.AssertTrue(t, err == nil, "Expected GetConsent to succeed")
	test_utils.AssertTrue(t, consentResultBytes != nil, "Expected GetConsent to succeed")
	json.Unmarshal(consentResultBytes, &consentResult)
	test_utils.AssertTrue(t, consentResult.Owner == "service1", "Got consent owner correctly")
	test_utils.AssertTrue(t, consentResult.Target == "service2", "Got consent target correctly")
	mstub.MockTransactionEnd("12")

	// update consent to DENY
	mstub.MockTransactionStart("13")
	stub = cached_stub.NewCachedStub(mstub, true, true, true)
	consent = Consent{}
	consent.Owner = "service1"
	consent.Service = "service1"
	consent.Datatype = "datatype1"
	consent.Target = "service2"
	consent.Option = []string{consentOptionDeny}
	consent.Timestamp = time.Now().Unix()
	consent.Expiration = 0
	consentBytes, _ = json.Marshal(&consent)
	_, err = PutConsentOwnerData(stub, service1AdminCaller, []string{string(consentBytes)})
	test_utils.AssertTrue(t, err == nil, "Expected PutConsentPatientData to succeed")
	mstub.MockTransactionEnd("13")

	// get consent back as service1 admin
	mstub.MockTransactionStart("14")
	stub = cached_stub.NewCachedStub(mstub)
	consentResultBytes, err = GetConsent(stub, service1AdminCaller, []string{"service1", "service2", "datatype1"})
	test_utils.AssertTrue(t, err == nil, "Expected GetConsent to succeed")
	test_utils.AssertTrue(t, consentResultBytes != nil, "Expected GetConsent to succeed")
	json.Unmarshal(consentResultBytes, &consentResult)
	test_utils.AssertSetsEqual(t, []string{consentOptionDeny}, consentResult.Option)
	mstub.MockTransactionEnd("14")

	// get consent back as service2 admin
	mstub.MockTransactionStart("14")
	stub = cached_stub.NewCachedStub(mstub)
	consentResultBytes, err = GetConsent(stub, service2AdminCaller, []string{"service1", "service2", "datatype1"})
	test_utils.AssertTrue(t, err == nil, "Expected GetConsent to succeed")
	test_utils.AssertTrue(t, consentResultBytes != nil, "Expected GetConsent to succeed")
	json.Unmarshal(consentResultBytes, &consentResult)
	test_utils.AssertSetsEqual(t, []string{consentOptionDeny}, consentResult.Option)
	mstub.MockTransactionEnd("14")

	// register patient and use patient as owner
	mstub.MockTransactionStart("15")
	stub = cached_stub.NewCachedStub(mstub)
	patient1 := test_utils.CreateTestUser("patient1")
	patient1Bytes, _ := json.Marshal(&patient1)
	_, err = user_mgmt.RegisterUser(stub, org1Caller, []string{string(patient1Bytes), "false"})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterUser to succeed")
	mstub.MockTransactionEnd("15")

	// patient give consent as owner
	mstub.MockTransactionStart("16")
	stub = cached_stub.NewCachedStub(mstub, true, true, true)
	consent.Owner = "patient1"
	consent.Service = ""
	consent.Target = "service1"
	consent.Datatype = "datatype1"
	consent.Option = []string{consentOptionWrite}
	consent.Timestamp = time.Now().Unix()
	consent.Expiration = 0
	consentBytes, _ = json.Marshal(&consent)
	consentKey = test_utils.GenerateSymKey()
	consentKeyB64 = crypto.EncodeToB64String(consentKey)
	patient1Caller, err := user_mgmt.GetUserData(stub, patient1, "patient1", true, true)
	_, err = PutConsentOwnerData(stub, patient1Caller, []string{string(consentBytes), consentKeyB64})
	test_utils.AssertTrue(t, err == nil, "Expected PutConsentPatientData to succeed")
	mstub.MockTransactionEnd("16")

	// get consent back as patient owner
	mstub.MockTransactionStart("17")
	stub = cached_stub.NewCachedStub(mstub)
	consentResultBytes, err = GetConsent(stub, patient1Caller, []string{"patient1", "service1", "datatype1"})
	test_utils.AssertTrue(t, err == nil, "Expected GetConsent to succeed")
	test_utils.AssertTrue(t, consentResultBytes != nil, "Expected GetConsent to succeed")
	json.Unmarshal(consentResultBytes, &consentResult)
	test_utils.AssertTrue(t, consentResult.Owner == "patient1", "Got consent owner correctly")
	test_utils.AssertTrue(t, consentResult.Target == "service1", "Got consent target correctly")
	mstub.MockTransactionEnd("17")

	// get consent back as service target
	mstub.MockTransactionStart("17")
	stub = cached_stub.NewCachedStub(mstub)
	consentResultBytes, err = GetConsent(stub, service1AdminCaller, []string{"patient1", "service1", "datatype1"})
	test_utils.AssertTrue(t, err == nil, "Expected GetConsent to succeed")
	test_utils.AssertTrue(t, consentResultBytes != nil, "Expected GetConsent to succeed")
	json.Unmarshal(consentResultBytes, &consentResult)
	test_utils.AssertTrue(t, consentResult.Owner == "patient1", "Got consent owner correctly")
	test_utils.AssertTrue(t, consentResult.Target == "service1", "Got consent target correctly")
	mstub.MockTransactionEnd("17")

	// update consent to DENY
	mstub.MockTransactionStart("18")
	stub = cached_stub.NewCachedStub(mstub, true, true, true)
	consent.Owner = "patient1"
	consent.Service = ""
	consent.Target = "service1"
	consent.Datatype = "datatype1"
	consent.Option = []string{consentOptionDeny}
	consent.Timestamp = time.Now().Unix()
	consent.Expiration = 0
	consentBytes, _ = json.Marshal(&consent)
	_, err = PutConsentOwnerData(stub, patient1Caller, []string{string(consentBytes)})
	test_utils.AssertTrue(t, err == nil, "Expected PutConsentPatientData to succeed")
	mstub.MockTransactionEnd("18")

	// get consent back as patient owner
	mstub.MockTransactionStart("19")
	stub = cached_stub.NewCachedStub(mstub)
	consentResultBytes, err = GetConsent(stub, patient1Caller, []string{"patient1", "service1", "datatype1"})
	test_utils.AssertTrue(t, err == nil, "Expected GetConsent to succeed")
	test_utils.AssertTrue(t, consentResultBytes != nil, "Expected GetConsent to succeed")
	json.Unmarshal(consentResultBytes, &consentResult)
	test_utils.AssertSetsEqual(t, []string{consentOptionDeny}, consentResult.Option)
	mstub.MockTransactionEnd("19")

	// get consent back as service target
	mstub.MockTransactionStart("19")
	stub = cached_stub.NewCachedStub(mstub)
	consentResultBytes, err = GetConsent(stub, service1AdminCaller, []string{"patient1", "service1", "datatype1"})
	test_utils.AssertTrue(t, err == nil, "Expected GetConsent to succeed")
	test_utils.AssertTrue(t, consentResultBytes != nil, "Expected GetConsent to succeed")
	json.Unmarshal(consentResultBytes, &consentResult)
	test_utils.AssertSetsEqual(t, []string{consentOptionDeny}, consentResult.Option)
	mstub.MockTransactionEnd("19")

	// try to get consent as a different service who should not have access
	mstub.MockTransactionStart("19")
	stub = cached_stub.NewCachedStub(mstub)
	consentResultBytes, err = GetConsent(stub, service2AdminCaller, []string{"patient1", "service1", "datatype1"})
	test_utils.AssertTrue(t, err != nil, "Expected GetConsent to fail")
	mstub.MockTransactionEnd("19")

	// register reference service that belongs to a different org
	// register org
	mstub.MockTransactionStart("20")
	stub = cached_stub.NewCachedStub(mstub)
	org2 := test_utils.CreateTestGroup("org2")
	org2Bytes, _ := json.Marshal(&org2)
	_, err = RegisterOrg(stub, org2, []string{string(org2Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterOrg to succeed")
	mstub.MockTransactionEnd("20")

	//  register service
	mstub.MockTransactionStart("21")
	stub = cached_stub.NewCachedStub(mstub, true, true, true)
	org2Caller, _ := user_mgmt.GetUserData(stub, org2, org2.ID, true, true)
	o2s1 := GenerateServiceForTesting("o2s1", "org2", []ServiceDatatype{serviceDatatype1})
	o2s1Bytes, _ := json.Marshal(&o2s1)
	_, err = RegisterService(stub, org2Caller, []string{string(o2s1Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterService to succeed")
	mstub.MockTransactionEnd("21")

	// patient give consent to reference service as owner
	mstub.MockTransactionStart("22")
	stub = cached_stub.NewCachedStub(mstub, true, true, true)
	consent.Owner = "patient1"
	consent.Service = ""
	consent.Target = "o2s1"
	consent.Datatype = "datatype1"
	consent.Option = []string{consentOptionWrite}
	consent.Timestamp = time.Now().Unix()
	consent.Expiration = 0
	consentBytes, _ = json.Marshal(&consent)
	consentKey = test_utils.GenerateSymKey()
	consentKeyB64 = crypto.EncodeToB64String(consentKey)
	_, err = PutConsentOwnerData(stub, patient1Caller, []string{string(consentBytes), consentKeyB64})
	test_utils.AssertTrue(t, err == nil, "Expected PutConsentPatientData to succeed")
	mstub.MockTransactionEnd("22")

	// get consent back as patient owner
	mstub.MockTransactionStart("23")
	stub = cached_stub.NewCachedStub(mstub)
	consentResultBytes, err = GetConsent(stub, patient1Caller, []string{"patient1", "o2s1", "datatype1"})
	test_utils.AssertTrue(t, err == nil, "Expected GetConsent to succeed")
	test_utils.AssertTrue(t, consentResultBytes != nil, "Expected GetConsent to succeed")
	json.Unmarshal(consentResultBytes, &consentResult)
	test_utils.AssertTrue(t, consentResult.Owner == "patient1", "Got consent owner correctly")
	test_utils.AssertTrue(t, consentResult.Target == "o2s1", "Got consent target correctly")
	mstub.MockTransactionEnd("23")

	// get consent back as service target
	mstub.MockTransactionStart("23")
	stub = cached_stub.NewCachedStub(mstub)
	o2s1AdminCaller, err := user_mgmt.GetUserData(stub, org2Caller, "o2s1", true, true)
	consentResultBytes, err = GetConsent(stub, o2s1AdminCaller, []string{"patient1", "o2s1", "datatype1"})
	test_utils.AssertTrue(t, err == nil, "Expected GetConsent to succeed")
	test_utils.AssertTrue(t, consentResultBytes != nil, "Expected GetConsent to succeed")
	json.Unmarshal(consentResultBytes, &consentResult)
	test_utils.AssertTrue(t, consentResult.Owner == "patient1", "Got consent owner correctly")
	test_utils.AssertTrue(t, consentResult.Target == "o2s1", "Got consent target correctly")
	mstub.MockTransactionEnd("23")

	// get consent back as org admin of service target
	mstub.MockTransactionStart("9")
	stub = cached_stub.NewCachedStub(mstub)
	consentResultBytes, err = GetConsent(stub, org2Caller, []string{"patient1", "o2s1", "datatype1"})
	test_utils.AssertTrue(t, err == nil, "Expected GetConsent to succeed")
	test_utils.AssertTrue(t, consentResultBytes != nil, "Expected GetConsent to succeed")
	json.Unmarshal(consentResultBytes, &consentResult)
	test_utils.AssertTrue(t, consentResult.Owner == "patient1", "Got consent owner correctly")
	test_utils.AssertTrue(t, consentResult.Target == "o2s1", "Got consent target correctly")
	mstub.MockTransactionEnd("9")

	// try to get consent as service1 admin, service 1 is original service of datatype1
	mstub.MockTransactionStart("9")
	stub = cached_stub.NewCachedStub(mstub)
	consentResultBytes, err = GetConsent(stub, service1AdminCaller, []string{"patient1", "o2s1", "datatype1"})
	test_utils.AssertTrue(t, err != nil, "Expected GetConsent to fail")
	test_utils.AssertTrue(t, consentResultBytes == nil, "Expected GetConsent to fail")
	mstub.MockTransactionEnd("9")
}

func TestGetConsentInternal(t *testing.T) {
	logger.SetLevel(shim.LogDebug)
	logger.Info("TestGetConsentInternal function called")

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

	// 8 get consent as org1
	mstub.MockTransactionStart("8")
	stub = cached_stub.NewCachedStub(mstub)
	consentResult, err = GetConsentInternal(stub, org1Caller, "service1", "datatype1", "patient1")
	test_utils.AssertTrue(t, err == nil, "Expected GetConsentInternal to succeed")
	test_utils.AssertTrue(t, consentResult.Owner == "patient1", "Got consent owner correctly")
	test_utils.AssertTrue(t, consentResult.Target == "service1", "Got consent service correctly")
	test_utils.AssertSetsEqual(t, []string{consentOptionWrite, consentOptionRead}, utils.GetSetFromList(consentResult.Option))
	mstub.MockTransactionEnd("8")

	// register org user and make him ServiceAdmin
	mstub.MockTransactionStart("9")
	stub = cached_stub.NewCachedStub(mstub)
	serviceAdmin := test_utils.CreateTestUser("ServiceAdmin")
	serviceAdminBytes, _ := json.Marshal(&serviceAdmin)
	_, err = user_mgmt.RegisterUser(stub, org1Caller, []string{string(serviceAdminBytes), "false"})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterUser to succeed")
	mstub.MockTransactionEnd("9")

	// give ServiceAdmin permission as service admin of service1
	mstub.MockTransactionStart("10")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = user_mgmt.PutUserInOrg(stub, org1Caller, []string{"ServiceAdmin", "service1", "true"})
	test_utils.AssertTrue(t, err == nil, "Expected PutUserInOrg to succeed")
	mstub.MockTransactionEnd("10")

	// 11 get consent as service admin
	mstub.MockTransactionStart("11")
	stub = cached_stub.NewCachedStub(mstub)
	serviceAdminCaller, err := user_mgmt.GetUserData(stub, serviceAdmin, "ServiceAdmin", true, true)
	test_utils.AssertTrue(t, err == nil, "Expected GetUserData to succeed")
	consentResult, err = GetConsentInternal(stub, serviceAdminCaller, "service1", "datatype1", "patient1")
	test_utils.AssertTrue(t, err == nil, "Expected GetConsentInternal to succeed")
	test_utils.AssertTrue(t, consentResult.Owner == "patient1", "Got consent owner correctly")
	test_utils.AssertTrue(t, consentResult.Target == "service1", "Got consent service correctly")
	test_utils.AssertSetsEqual(t, []string{consentOptionWrite, consentOptionRead}, utils.GetSetFromList(consentResult.Option))
	mstub.MockTransactionEnd("11")

	// register another org user
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	orgAdminUser1 := CreateTestSolutionUser("orgAdminUser1")
	orgAdminUser1.Org = org1.ID
	data := make(map[string]interface{})
	data["test"] = "data"
	orgAdminUser1.SolutionPrivateData = data
	orgAdminUser1Bytes, _ := json.Marshal(&orgAdminUser1)
	_, err = RegisterUser(stub, org1Caller, []string{string(orgAdminUser1Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterUser OMR to succeed")
	mstub.MockTransactionEnd("t123")

	// put orgUser1 in org1
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = PutUserInOrg(stub, org1Caller, []string{"orgAdminUser1", org1.ID, "false"})
	test_utils.AssertTrue(t, err == nil, "Expected PutUserInOrg to succeed")
	mstub.MockTransactionEnd("t123")

	// give org admin permission
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = AddPermissionOrgAdmin(stub, org1Caller, []string{"orgAdminUser1", org1.ID})
	test_utils.AssertTrue(t, err == nil, "Expected AddPermissionOrgAdmin to succeed")
	mstub.MockTransactionEnd("t123")

	// get consent as org user with org admin permission
	mstub.MockTransactionStart("11")
	stub = cached_stub.NewCachedStub(mstub)
	orgAdminUser1Caller, err := user_mgmt.GetUserData(stub, org1Caller, "orgAdminUser1", true, true)
	test_utils.AssertTrue(t, err == nil, "Expected GetUserData to succeed")
	consentResult, err = GetConsentInternal(stub, orgAdminUser1Caller, "service1", "datatype1", "patient1")
	test_utils.AssertTrue(t, err == nil, "Expected GetConsentInternal to succeed")
	test_utils.AssertTrue(t, consentResult.Owner == "patient1", "Got consent owner correctly")
	test_utils.AssertTrue(t, consentResult.Target == "service1", "Got consent service correctly")
	test_utils.AssertSetsEqual(t, []string{consentOptionWrite, consentOptionRead}, utils.GetSetFromList(consentResult.Option))
	mstub.MockTransactionEnd("11")
}

func TestValidateConsent(t *testing.T) {
	logger.SetLevel(shim.LogDebug)
	logger.Info("TestValidateConsent function called")

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

	// validate consent as service1 for write access
	mstub.MockTransactionStart("8")
	stub = cached_stub.NewCachedStub(mstub)
	service1Subgroup, _ := user_mgmt.GetUserData(stub, org1, "service1", true, true)
	cvResultBytes, err := ValidateConsent(stub, service1Subgroup, []string{"patient1", "service1", "datatype1", consentOptionWrite, strconv.FormatInt(time.Now().Unix(), 10)})
	test_utils.AssertTrue(t, err == nil, "Expected ValidateConsent to succeed")
	test_utils.AssertTrue(t, cvResultBytes != nil, "Expected ValidateConsent to succeed")
	cvResult := ValidationResultWithLog{}
	json.Unmarshal(cvResultBytes, &cvResult)
	// NOTE: We need to call url.QueryUnescape() function because on working environment Node.js automatically
	// replace URL escape characters with original values
	base64UrlEncodedToken, _ := url.QueryUnescape(cvResult.ConsentValidation.Token)
	token, err := DecryptConsentValidationToken(stub, service1Subgroup, base64UrlEncodedToken)
	test_utils.AssertTrue(t, err == nil, "Expected to decrypt token successfully")
	test_utils.AssertTrue(t, token.ConsentKey != nil, "Expected to get ConsentKey from token successfully")
	test_utils.AssertTrue(t, cvResult.ConsentValidation.PermissionGranted == true, "Got consent validation correctly")
	test_utils.AssertTrue(t, cvResult.ConsentValidation.Message == "permission granted", "Got consent validation correctly")
	mstub.MockTransactionEnd("8")

	// validate consent as service1 for read access
	mstub.MockTransactionStart("9")
	stub = cached_stub.NewCachedStub(mstub)
	cvResultBytes, err = ValidateConsent(stub, service1Subgroup, []string{"patient1", "service1", "datatype1", consentOptionRead, strconv.FormatInt(time.Now().Unix(), 10)})
	test_utils.AssertTrue(t, err == nil, "Expected ValidateConsent to succeed")
	test_utils.AssertTrue(t, cvResultBytes != nil, "Expected ValidateConsent to succeed")
	json.Unmarshal(cvResultBytes, &cvResult)
	test_utils.AssertTrue(t, cvResult.ConsentValidation.PermissionGranted == true, "Got consent validation correctly")
	test_utils.AssertTrue(t, cvResult.ConsentValidation.Message == "permission granted", "Got consent validation correctly")
	mstub.MockTransactionEnd("9")

	// register org user
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	orgUser1 := CreateTestSolutionUser("orgUser1")
	orgUser1.Org = "org1"
	orgUser1Bytes, _ := json.Marshal(&orgUser1)
	_, err = RegisterUser(stub, org1Caller, []string{string(orgUser1Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterUser OMR to succeed")
	mstub.MockTransactionEnd("t123")

	// put orgUser1 in org1
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = PutUserInOrg(stub, org1Caller, []string{orgUser1.ID, org1.ID, "false"})
	test_utils.AssertTrue(t, err == nil, "Expected PutUserInOrg to succeed")
	mstub.MockTransactionEnd("t123")

	// give service1 admin permission to orgUser1
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = AddPermissionServiceAdmin(stub, org1Caller, []string{orgUser1.ID, "service1"})
	test_utils.AssertTrue(t, err == nil, "Expected AddPermissionServiceAdmin to succeed")
	mstub.MockTransactionEnd("t123")

	// validate consent as orgUser1
	mstub.MockTransactionStart("15")
	stub = cached_stub.NewCachedStub(mstub)
	orgUser1Caller, _ := user_mgmt.GetUserData(stub, org1Caller, orgUser1.ID, true, true)
	cvResultBytes, err = ValidateConsent(stub, orgUser1Caller, []string{"patient1", "service1", "datatype1", consentOptionWrite, strconv.FormatInt(time.Now().Unix(), 10)})
	test_utils.AssertTrue(t, err == nil, "Expected ValidateConsent to succeed")
	test_utils.AssertTrue(t, cvResultBytes != nil, "Expected ValidateConsent to succeed")
	json.Unmarshal(cvResultBytes, &cvResult)
	test_utils.AssertTrue(t, cvResult.ConsentValidation.PermissionGranted == true, "Got consent validation correctly")
	test_utils.AssertTrue(t, cvResult.ConsentValidation.Message == "permission granted", "Got consent validation correctly")
	mstub.MockTransactionEnd("15")

	// validate consent as org admin for read access
	mstub.MockTransactionStart("9")
	stub = cached_stub.NewCachedStub(mstub)
	cvResultBytes, err = ValidateConsent(stub, org1Caller, []string{"patient1", "service1", "datatype1", consentOptionRead, strconv.FormatInt(time.Now().Unix(), 10)})
	test_utils.AssertTrue(t, err == nil, "Expected ValidateConsent to succeed")
	test_utils.AssertTrue(t, cvResultBytes != nil, "Expected ValidateConsent to succeed")
	json.Unmarshal(cvResultBytes, &cvResult)
	test_utils.AssertTrue(t, cvResult.ConsentValidation.PermissionGranted == true, "Got consent validation correctly")
	test_utils.AssertTrue(t, cvResult.ConsentValidation.Message == "permission granted", "Got consent validation correctly")
	mstub.MockTransactionEnd("9")

	// register another org user
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	orgUser2 := CreateTestSolutionUser("orgUser2")
	orgUser2.Org = "org1"
	orgUser2Bytes, _ := json.Marshal(&orgUser2)
	_, err = RegisterUser(stub, org1Caller, []string{string(orgUser2Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterUser OMR to succeed")
	mstub.MockTransactionEnd("t123")

	// put orgUser1 in org1
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = PutUserInOrg(stub, org1Caller, []string{orgUser2.ID, org1.ID, "false"})
	test_utils.AssertTrue(t, err == nil, "Expected PutUserInOrg to succeed")
	mstub.MockTransactionEnd("t123")

	// give org admin permission to orgUser1
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = AddPermissionOrgAdmin(stub, org1Caller, []string{orgUser2.ID, "org1"})
	test_utils.AssertTrue(t, err == nil, "Expected AddPermissionServiceAdmin to succeed")
	mstub.MockTransactionEnd("t123")

	// validate consent as orgUser2
	mstub.MockTransactionStart("15")
	stub = cached_stub.NewCachedStub(mstub)
	orgUser2Caller, _ := user_mgmt.GetUserData(stub, org1Caller, orgUser2.ID, true, true)
	cvResultBytes, err = ValidateConsent(stub, orgUser2Caller, []string{"patient1", "service1", "datatype1", consentOptionWrite, strconv.FormatInt(time.Now().Unix(), 10)})
	test_utils.AssertTrue(t, err == nil, "Expected ValidateConsent to succeed")
	test_utils.AssertTrue(t, cvResultBytes != nil, "Expected ValidateConsent to succeed")
	json.Unmarshal(cvResultBytes, &cvResult)
	test_utils.AssertTrue(t, cvResult.ConsentValidation.PermissionGranted == true, "Got consent validation correctly")
	test_utils.AssertTrue(t, cvResult.ConsentValidation.Message == "permission granted", "Got consent validation correctly")
	mstub.MockTransactionEnd("15")

	// register datatype2
	mstub.MockTransactionStart("6")
	stub = cached_stub.NewCachedStub(mstub)
	datatype2 := Datatype{DatatypeID: "datatype2", Description: "datatype2"}
	datatype2Bytes, _ := json.Marshal(&datatype2)
	_, err = RegisterDatatype(stub, org1Caller, []string{string(datatype2Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterDatatype to succeed")
	mstub.MockTransactionEnd("6")

	// register service2
	mstub.MockTransactionStart("7")
	stub = cached_stub.NewCachedStub(mstub, true, true, true)
	serviceDatatype2 := GenerateServiceDatatypeForTesting("datatype2", "service2", []string{consentOptionWrite, consentOptionRead})
	service2 := GenerateServiceForTesting("service2", "org1", []ServiceDatatype{serviceDatatype2})
	service2Bytes, _ := json.Marshal(&service2)
	_, err = RegisterService(stub, org1Caller, []string{string(service2Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterService to succeed")
	mstub.MockTransactionEnd("7")

	// validate consent as service2 for write access, should fail
	mstub.MockTransactionStart("8")
	stub = cached_stub.NewCachedStub(mstub)
	service2Subgroup, _ := user_mgmt.GetUserData(stub, org1, "service2", true, true)
	cvResultBytes, err = ValidateConsent(stub, service2Subgroup, []string{"patient1", "service1", "datatype1", consentOptionWrite, strconv.FormatInt(time.Now().Unix(), 10)})
	test_utils.AssertTrue(t, err == nil, "Expected ValidateConsent to succeed")
	test_utils.AssertTrue(t, cvResultBytes != nil, "Expected ValidateConsent to succeed")
	json.Unmarshal(cvResultBytes, &cvResult)
	test_utils.AssertTrue(t, cvResult.ConsentValidation.PermissionGranted == false, "Got consent validation correctly")
	test_utils.AssertTrue(t, cvResult.ConsentValidation.Message != "permission granted", "Got consent validation correctly")
	mstub.MockTransactionEnd("8")

	// update consent to DENY
	mstub.MockTransactionStart("10")
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
	mstub.MockTransactionEnd("10")

	// get consent back to make sure it was added
	mstub.MockTransactionStart("11")
	stub = cached_stub.NewCachedStub(mstub)
	consentResultBytes, err = GetConsent(stub, patient1Caller, []string{"patient1", "service1", "datatype1"})
	test_utils.AssertTrue(t, err == nil, "Expected GetConsent to succeed")
	test_utils.AssertTrue(t, consentResultBytes != nil, "Expected GetConsent to succeed")
	json.Unmarshal(consentResultBytes, &consentResult)
	test_utils.AssertSetsEqual(t, []string{consentOptionDeny}, consentResult.Option)
	mstub.MockTransactionEnd("11")

	// validate consent as service1
	mstub.MockTransactionStart("12")
	stub = cached_stub.NewCachedStub(mstub)
	cvResultBytes, err = ValidateConsent(stub, service1Subgroup, []string{"patient1", "service1", "datatype1", consentOptionWrite, strconv.FormatInt(time.Now().Unix(), 10)})
	test_utils.AssertTrue(t, err == nil, "Expected ValidateConsent to fail")
	test_utils.AssertTrue(t, cvResultBytes != nil, "Expected ValidateConsent to succeed")
	json.Unmarshal(cvResultBytes, &cvResult)
	test_utils.AssertTrue(t, cvResult.ConsentValidation.PermissionGranted == false, "Got consent validation correctly")
	test_utils.AssertTrue(t, cvResult.ConsentValidation.Message != "permission granted", "Got consent validation correctly")
	mstub.MockTransactionEnd("12")

	// validate consent as orgUser1
	mstub.MockTransactionStart("15")
	stub = cached_stub.NewCachedStub(mstub)
	cvResultBytes, err = ValidateConsent(stub, orgUser1Caller, []string{"patient1", "service1", "datatype1", consentOptionWrite, strconv.FormatInt(time.Now().Unix(), 10)})
	test_utils.AssertTrue(t, err == nil, "Expected ValidateConsent to succeed")
	test_utils.AssertTrue(t, cvResultBytes != nil, "Expected ValidateConsent to succeed")
	json.Unmarshal(cvResultBytes, &cvResult)
	test_utils.AssertTrue(t, cvResult.ConsentValidation.PermissionGranted == false, "Got consent validation correctly")
	test_utils.AssertTrue(t, cvResult.ConsentValidation.Message != "permission granted", "Got consent validation correctly")
	mstub.MockTransactionEnd("15")
}

func TestValidateConsentAsServiceAdmin(t *testing.T) {
	logger.SetLevel(shim.LogDebug)
	logger.Info("TestValidateConsent function called")

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

	// Register Service Admin user
	mstub.MockTransactionStart("RegisterServiceAdminUser")
	stub = cached_stub.NewCachedStub(mstub)
	serviceAdminUser := CreateTestSolutionUser("ServiceAdmin")
	serviceAdminUser.Org = org1Caller.ID
	serviceAdminUserBytes, _ := json.Marshal(&serviceAdminUser)
	_, err = RegisterUser(stub, org1Caller, []string{string(serviceAdminUserBytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterUser to succeed")
	mstub.MockTransactionEnd("RegisterServiceAdminUser")

	// put ServiceAdmin user in org1
	mstub.MockTransactionStart("161")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = PutUserInOrg(stub, org1Caller, []string{serviceAdminUser.ID, org1Caller.ID, "false"})
	test_utils.AssertTrue(t, err == nil, "Expected PutUserInOrg to succeed")
	mstub.MockTransactionEnd("161")

	// give service admin permission
	mstub.MockTransactionStart("171")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = AddPermissionServiceAdmin(stub, org1Caller, []string{serviceAdminUser.ID, "service1"})
	test_utils.AssertTrue(t, err == nil, "Expected AddPermissionServiceAdmin to succeed")
	mstub.MockTransactionEnd("171")

	// Validate Consent as Service Admin user
	mstub.MockTransactionStart("ValidateConsentForServiceAdmin")
	stub = cached_stub.NewCachedStub(mstub)
	serviceAdminCaller, _ := user_mgmt.GetUserData(stub, org1Caller, serviceAdminUser.ID, true, true)
	cvResultBytes, err := ValidateConsent(stub, serviceAdminCaller, []string{"patient1", "service1", "datatype1", consentOptionRead, strconv.FormatInt(time.Now().Unix(), 10)})
	test_utils.AssertTrue(t, err == nil, "Expect ValidateConsent under Service Admin without errors")
	test_utils.AssertTrue(t, cvResultBytes != nil, "Expect ValidateConsent returns validation result")
	cvResult := ValidationResultWithLog{}
	json.Unmarshal(cvResultBytes, &cvResult)
	test_utils.AssertTrue(t, cvResult.ConsentValidation.PermissionGranted == true, "Expect permission to be granted")
	test_utils.AssertTrue(t, cvResult.ConsentValidation.Message == "permission granted", "Expect 'permission granted' message")
	mstub.MockTransactionEnd("ValidateConsentForServiceAdmin")

	// AddValidateConsentQueryLog after Validate Consent as Service Admin.
	mstub.MockTransactionStart("addValidateConsentQueryLog")
	stub = cached_stub.NewCachedStub(mstub)
	cvJsonBytes, _ := json.Marshal(cvResult.ConsentValidation)
	transactionLogJsonBytes, _ := json.Marshal(cvResult.TransactionLog)
	err = AddValidateConsentQueryLog(stub, serviceAdminCaller, []string{string(cvJsonBytes), string(transactionLogJsonBytes)})
	test_utils.AssertTrue(t, err == nil, "AddValidateConsentQueryLog should not return any error")
	mstub.MockTransactionEnd("addValidateConsentQueryLog")
}
