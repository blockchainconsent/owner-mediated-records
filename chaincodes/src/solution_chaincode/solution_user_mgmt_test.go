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
	"common/bchcls/asset_mgmt"
	"common/bchcls/cached_stub"
	"common/bchcls/data_model"
	"common/bchcls/datastore/datastore_manager"
	"common/bchcls/init_common"
	"common/bchcls/test_utils"
	"common/bchcls/user_mgmt"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"testing"

	"github.com/hyperledger/fabric/core/chaincode/shim"
)

func TestRegisterOrgUser(t *testing.T) {
	logger.SetLevel(shim.LogDebug)
	logger.Info("TestRegisterOrgUser function called")

	mstub := SetupIndexesAndGetStub(t)
	mstub.MockTransactionStart("t1")
	stub := cached_stub.NewCachedStub(mstub)
	init_common.Init(stub)
	mstub.MockTransactionEnd("t1")

	// register org
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	org1 := test_utils.CreateTestGroup("org1")
	org1Bytes, _ := json.Marshal(&org1)
	_, err := RegisterOrg(stub, org1, []string{string(org1Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterOrg to succeed")
	mstub.MockTransactionEnd("t123")

	// get org back
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	org1Caller, _ := user_mgmt.GetUserData(stub, org1, org1.ID, true, true)
	getSolutionUserResult, err := GetSolutionUserWithParams(stub, org1Caller, org1.ID, true, true)
	test_utils.AssertTrue(t, err == nil, "Expected GetUserData to succeed")
	test_utils.AssertTrue(t, getSolutionUserResult.Role == "org", "Expected GetSolutionUserWithParams to succeed")
	test_utils.AssertTrue(t, getSolutionUserResult.Org == org1.ID, "Expected GetUserData to succeed")
	mstub.MockTransactionEnd("t123")

	// register org user
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	orgUser1 := CreateTestSolutionUser("orgUser1")
	orgUser1.Org = org1.ID
	orgUser1Bytes, _ := json.Marshal(&orgUser1)
	_, err = RegisterUser(stub, org1Caller, []string{string(orgUser1Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterUser OMR to succeed")
	mstub.MockTransactionEnd("t123")

	// put orgUser1 in org1
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = PutUserInOrg(stub, org1Caller, []string{"orgUser1", org1.ID, "false"})
	test_utils.AssertTrue(t, err == nil, "Expected PutUserInOrg to succeed")
	mstub.MockTransactionEnd("t123")

	// get user back
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	getSolutionUserResult, err = GetSolutionUserWithParams(stub, org1Caller, orgUser1.ID, true, true)
	test_utils.AssertTrue(t, err == nil, "Expected GetUserData to succeed")
	test_utils.AssertTrue(t, getSolutionUserResult.Role == SOLUTION_ROLE_USER, "Expected GetSolutionUserWithParams to succeed")
	test_utils.AssertTrue(t, getSolutionUserResult.Org == org1.ID, "Expected GetUserData to succeed")
	mstub.MockTransactionEnd("t123")
}

func TestUpdateOrgUser(t *testing.T) {
	logger.SetLevel(shim.LogDebug)
	logger.Info("TestUpdateOrgUser function called")

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

	// register org user
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	org1Caller, _ := user_mgmt.GetUserData(stub, org1, org1.ID, true, true)
	orgUser1 := CreateTestSolutionUser("orgUser1")
	orgUser1.Org = org1.ID
	data := make(map[string]interface{})
	data["test"] = "data"
	orgUser1.SolutionPrivateData = data
	orgUser1Bytes, _ := json.Marshal(&orgUser1)
	_, err = RegisterUser(stub, org1Caller, []string{string(orgUser1Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterUser OMR to succeed")
	mstub.MockTransactionEnd("t123")

	// get user object back
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	orgUser1Caller, err := user_mgmt.GetUserData(stub, org1, "orgUser1", true, true)
	test_utils.AssertTrue(t, err == nil, "Expected GetUserData to succeed")
	mstub.MockTransactionEnd("t123")

	// put orgUser1 in org1
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = PutUserInOrg(stub, org1Caller, []string{orgUser1Caller.ID, org1.ID, "false"})
	test_utils.AssertTrue(t, err == nil, "Expected PutUserInOrg to succeed")
	mstub.MockTransactionEnd("t123")

	// give org admin permission
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = AddPermissionOrgAdmin(stub, org1Caller, []string{orgUser1Caller.ID, org1.ID})
	test_utils.AssertTrue(t, err == nil, "Expected AddPermissionOrgAdmin to succeed")
	mstub.MockTransactionEnd("t123")

	// get user back
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	getSolutionUserResult, err := GetSolutionUserWithParams(stub, org1Caller, orgUser1Caller.ID, true, true)
	test_utils.AssertTrue(t, err == nil, "Expected GetUserData to succeed")
	test_utils.AssertTrue(t, getSolutionUserResult.SolutionInfo.IsOrgAdmin == true, "Expected GetSolutionUserWithParams to succeed")
	test_utils.AssertTrue(t, getSolutionUserResult.SolutionPrivateData.(map[string]interface{})["test"] == "data", "Expected GetUserData to succeed")
	mstub.MockTransactionEnd("t123")

	// update org user
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	orgUser1.Name = "new name"
	orgUser1Bytes, _ = json.Marshal(&orgUser1)
	_, err = RegisterUser(stub, orgUser1Caller, []string{string(orgUser1Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected UpdateUser OMR to succeed")
	mstub.MockTransactionEnd("t123")

	// get user back
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	getSolutionUserResult, err = GetSolutionUserWithParams(stub, org1Caller, orgUser1Caller.ID, true, true)
	test_utils.AssertTrue(t, err == nil, "Expected GetUserData to succeed")
	test_utils.AssertTrue(t, getSolutionUserResult.Name == "new name", "Expected GetUserData to succeed")
	test_utils.AssertTrue(t, getSolutionUserResult.SolutionInfo.IsOrgAdmin == true, "Expected GetSolutionUserWithParams to succeed")
	test_utils.AssertTrue(t, getSolutionUserResult.SolutionPrivateData.(map[string]interface{})["test"] == "data", "Expected GetUserData to succeed")
	mstub.MockTransactionEnd("t123")

	// remove org admin permission
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = RemovePermissionOrgAdmin(stub, org1Caller, []string{orgUser1Caller.ID, org1.ID})
	test_utils.AssertTrue(t, err == nil, "Expected RemovePermissionOrgAdmin to succeed")
	mstub.MockTransactionEnd("t123")

	// give org admin permission again
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = AddPermissionOrgAdmin(stub, org1Caller, []string{orgUser1Caller.ID, org1.ID})
	test_utils.AssertTrue(t, err == nil, "Expected AddPermissionOrgAdmin to succeed")
	mstub.MockTransactionEnd("t123")

	//  register datatype
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	datatype1 := Datatype{DatatypeID: "datatype1", Description: "datatype1"}
	datatype1Bytes, _ := json.Marshal(&datatype1)
	_, err = RegisterDatatype(stub, org1Caller, []string{string(datatype1Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterDatatype to succeed")
	mstub.MockTransactionEnd("t123")

	// register service1 as org
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub, true, true, true)
	serviceDatatype1 := GenerateServiceDatatypeForTesting("datatype1", "service1", []string{consentOptionWrite, consentOptionRead})
	service1 := GenerateServiceForTesting("service1", org1.ID, []ServiceDatatype{serviceDatatype1})
	service1Bytes, _ := json.Marshal(&service1)
	_, err = RegisterService(stub, org1Caller, []string{string(service1Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterService to succeed")
	mstub.MockTransactionEnd("t123")

	// register service2 as org user who has admin permission
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub, true, true, true)
	serviceDatatype2 := GenerateServiceDatatypeForTesting("datatype1", "service2", []string{consentOptionWrite, consentOptionRead})
	service2 := GenerateServiceForTesting("service2", org1.ID, []ServiceDatatype{serviceDatatype2})
	service2Bytes, _ := json.Marshal(&service2)
	orgUser1Caller, _ = user_mgmt.GetUserData(stub, org1Caller, orgUser1.ID, true, true)
	_, err = RegisterService(stub, orgUser1Caller, []string{string(service2Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterService to succeed")
	mstub.MockTransactionEnd("t123")

	// update org user
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	data = make(map[string]interface{})
	data["test"] = "new data"
	orgUser1.SolutionPrivateData = data
	orgUser1Bytes, _ = json.Marshal(&orgUser1)
	_, err = RegisterUser(stub, orgUser1Caller, []string{string(orgUser1Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected UpdateUser OMR to succeed")
	mstub.MockTransactionEnd("t123")

	// get user back
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	getSolutionUserResult, err = GetSolutionUserWithParams(stub, orgUser1Caller, orgUser1Caller.ID, true, true)
	test_utils.AssertTrue(t, err == nil, "Expected GetUserData to succeed")
	test_utils.AssertTrue(t, getSolutionUserResult.SolutionInfo.IsOrgAdmin == true, "Expected GetSolutionUserWithParams to succeed")
	test_utils.AssertTrue(t, getSolutionUserResult.SolutionPrivateData.(map[string]interface{})["test"] == "new data", "Expected GetUserData to succeed")
	mstub.MockTransactionEnd("t123")
}

func TestUpdatePatient(t *testing.T) {
	logger.SetLevel(shim.LogDebug)
	logger.Info("TestUpdatePatient function called")

	mstub := SetupIndexesAndGetStub(t)
	mstub.MockTransactionStart("t1")
	stub := cached_stub.NewCachedStub(mstub)
	init_common.Init(stub)
	mstub.MockTransactionEnd("t1")

	// register system admin
	mstub.MockTransactionStart("1")
	stub = cached_stub.NewCachedStub(mstub)
	systemAdmin := CreateTestSolutionUser("systemAdmin")
	systemAdmin.Role = SOLUTION_ROLE_SYSTEM
	systemAdminBytes, _ := json.Marshal(&systemAdmin)
	sysAdminCaller, err := convertToPlatformUser(stub, systemAdmin)
	test_utils.AssertTrue(t, err == nil, "Expected convertToPlatformUser to succeed")
	_, err = RegisterUser(stub, sysAdminCaller, []string{string(systemAdminBytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterSystemAdmin to succeed")
	mstub.MockTransactionEnd("1")

	// register patient
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	patient1 := CreateTestSolutionUser("patient1")
	data := make(map[string]interface{})
	data["test"] = "data"
	patient1.SolutionPrivateData = data
	patient1Bytes, _ := json.Marshal(&patient1)
	_, err = user_mgmt.RegisterUser(stub, sysAdminCaller, []string{string(patient1Bytes), "false"})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterUser to succeed")
	mstub.MockTransactionEnd("t123")

	// get user back
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	patient1Caller, err := convertToPlatformUser(stub, patient1)
	test_utils.AssertTrue(t, err == nil, "Expected convertToPlatformUser to succeed")
	getSolutionUserResult, err := GetSolutionUserWithParams(stub, patient1Caller, patient1.ID, true, true)
	test_utils.AssertTrue(t, err == nil, "Expected GetSolutionUserWithParams to succeed")
	test_utils.AssertTrue(t, getSolutionUserResult.SolutionPrivateData.(map[string]interface{})["test"] == "data", "Expected GetUserData to succeed")
	mstub.MockTransactionEnd("t123")

	// update patient
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	patient1.Name = "new name"
	patient1Bytes, _ = json.Marshal(&patient1)
	_, err = RegisterUser(stub, patient1Caller, []string{string(patient1Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected UpdateUser OMR to succeed")
	mstub.MockTransactionEnd("t123")

	// get user back
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	getSolutionUserResult, err = GetSolutionUserWithParams(stub, patient1Caller, patient1.ID, true, true)
	test_utils.AssertTrue(t, err == nil, "Expected GetSolutionUserWithParams to succeed")
	test_utils.AssertTrue(t, getSolutionUserResult.SolutionPrivateData.(map[string]interface{})["test"] == "data", "Expected GetUserData to succeed")
	test_utils.AssertTrue(t, getSolutionUserResult.Name == "new name", "Expected GetUserData to succeed")
	mstub.MockTransactionEnd("t123")

	// update patient
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	data = make(map[string]interface{})
	data["new test"] = "data"
	patient1.SolutionPrivateData = data
	patient1Bytes, _ = json.Marshal(&patient1)
	_, err = RegisterUser(stub, patient1Caller, []string{string(patient1Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected UpdateUser OMR to succeed")
	mstub.MockTransactionEnd("t123")

	// get user back
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	getSolutionUserResult, err = GetSolutionUserWithParams(stub, patient1Caller, patient1.ID, true, true)
	test_utils.AssertTrue(t, err == nil, "Expected GetSolutionUserWithParams to succeed")
	test_utils.AssertTrue(t, getSolutionUserResult.SolutionPrivateData.(map[string]interface{})["new test"] == "data", "Expected GetUserData to succeed")
	mstub.MockTransactionEnd("t123")
}

func TestAddAndRemoveOrgAdminPermission(t *testing.T) {
	logger.SetLevel(shim.LogDebug)
	logger.Info("TestAddAndRemoveOrgAdminPermission function called")

	mstub := SetupIndexesAndGetStub(t)
	mstub.MockTransactionStart("t1")
	stub := cached_stub.NewCachedStub(mstub)
	init_common.Init(stub)
	mstub.MockTransactionEnd("t1")

	// register org
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	org1 := test_utils.CreateTestGroup("org1")
	org1Bytes, _ := json.Marshal(&org1)
	_, err := RegisterOrg(stub, org1, []string{string(org1Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterOrg to succeed")
	mstub.MockTransactionEnd("t123")

	// register org user
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	org1Caller, _ := user_mgmt.GetUserData(stub, org1, org1.ID, true, true)
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

	// give org admin permission
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = AddPermissionOrgAdmin(stub, org1Caller, []string{orgUser1.ID, org1.ID})
	test_utils.AssertTrue(t, err == nil, "Expected AddPermissionOrgAdmin to succeed")
	mstub.MockTransactionEnd("t123")

	// get user back
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	getSolutionUserResult, err := GetSolutionUserWithParams(stub, org1Caller, orgUser1.ID, true, true)
	test_utils.AssertTrue(t, err == nil, "Expected GetSolutionUserWithParams to succeed")
	test_utils.AssertTrue(t, getSolutionUserResult.SolutionInfo.IsOrgAdmin == true, "Expected GetUserData to succeed")
	mstub.MockTransactionEnd("t123")

	// remove org admin permission
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = RemovePermissionOrgAdmin(stub, org1Caller, []string{orgUser1.ID, org1.ID})
	test_utils.AssertTrue(t, err == nil, "Expected RemovePermissionOrgAdmin to succeed")
	mstub.MockTransactionEnd("t123")

	// get user back
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	getSolutionUserResult, err = GetSolutionUserWithParams(stub, org1Caller, orgUser1.ID, true, true)
	test_utils.AssertTrue(t, err == nil, "Expected GetSolutionUserWithParams to succeed")
	test_utils.AssertTrue(t, getSolutionUserResult.SolutionInfo.IsOrgAdmin == false, "Expected GetUserData to succeed")
	mstub.MockTransactionEnd("t123")

	// give org admin permission again for additional testing
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = AddPermissionOrgAdmin(stub, org1Caller, []string{orgUser1.ID, org1.ID})
	test_utils.AssertTrue(t, err == nil, "Expected AddPermissionOrgAdmin to succeed")
	mstub.MockTransactionEnd("t123")

	// register additional org user as orgUser1
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	orgUser1Caller, _ := user_mgmt.GetUserData(stub, org1Caller, orgUser1.ID, true, true)
	orgUser2 := CreateTestSolutionUser("orgUser2")
	orgUser2.Org = "org1"
	orgUser2Bytes, _ := json.Marshal(&orgUser2)
	_, err = RegisterUser(stub, orgUser1Caller, []string{string(orgUser2Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterUser OMR to succeed")
	mstub.MockTransactionEnd("t123")

	// put orgUser2 in org1 (non-admin) as orgUser1
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = PutUserInOrg(stub, orgUser1Caller, []string{orgUser2.ID, org1.ID, "false"})
	test_utils.AssertTrue(t, err == nil, "Expected PutUserInOrg to succeed")
	mstub.MockTransactionEnd("t123")

	// give org admin permission to orgUser2 as orgUser1
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = AddPermissionOrgAdmin(stub, orgUser1Caller, []string{orgUser2.ID, org1.ID})
	test_utils.AssertTrue(t, err == nil, "Expected AddPermissionOrgAdmin to succeed")
	mstub.MockTransactionEnd("t123")

	// get user back as orgUser1
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	getSolutionUserResult, err = GetSolutionUserWithParams(stub, orgUser1Caller, orgUser2.ID, true, true)
	test_utils.AssertTrue(t, err == nil, "Expected GetSolutionUserWithParams to succeed")
	test_utils.AssertTrue(t, getSolutionUserResult.SolutionInfo.IsOrgAdmin == true, "Expected GetUserData to succeed")
	mstub.MockTransactionEnd("t123")

	// get user back as org1
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	getSolutionUserResult, err = GetSolutionUserWithParams(stub, org1Caller, orgUser2.ID, true, true)
	test_utils.AssertTrue(t, err == nil, "Expected GetSolutionUserWithParams to succeed")
	test_utils.AssertTrue(t, getSolutionUserResult.SolutionInfo.IsOrgAdmin == true, "Expected GetUserData to succeed")
	mstub.MockTransactionEnd("t123")

	// remove orgUser2's admin permission as orgUser1
	// now orgUser1 is not org admin
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	orgUser2Caller, _ := user_mgmt.GetUserData(stub, org1Caller, orgUser2.ID, true, true)
	_, err = RemovePermissionOrgAdmin(stub, orgUser2Caller, []string{orgUser1.ID, org1.ID})
	test_utils.AssertTrue(t, err == nil, "Expected RemovePermissionOrgAdmin to succeed")
	mstub.MockTransactionEnd("t123")

	// get orgUser1 back
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	getSolutionUserResult, err = GetSolutionUserWithParams(stub, org1Caller, orgUser1.ID, true, true)
	test_utils.AssertTrue(t, err == nil, "Expected GetSolutionUserWithParams to succeed")
	test_utils.AssertTrue(t, getSolutionUserResult.SolutionInfo.IsOrgAdmin == false, "Expected GetUserData to succeed")
	mstub.MockTransactionEnd("t123")

	// remove orgUser2's admin permission as org1
	// now orgUser2 is not org admin
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = RemovePermissionOrgAdmin(stub, org1Caller, []string{orgUser2.ID, org1.ID})
	test_utils.AssertTrue(t, err == nil, "Expected RemovePermissionOrgAdmin to succeed")
	mstub.MockTransactionEnd("t123")

	// get orgUser2 back
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	getSolutionUserResult, err = GetSolutionUserWithParams(stub, org1Caller, orgUser2.ID, true, true)
	test_utils.AssertTrue(t, err == nil, "Expected GetSolutionUserWithParams to succeed")
	test_utils.AssertTrue(t, getSolutionUserResult.SolutionInfo.IsOrgAdmin == false, "Expected GetUserData to succeed")
	mstub.MockTransactionEnd("t123")

	// give org admin permission to orgUser2 as orgUser1, should fail
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = AddPermissionOrgAdmin(stub, orgUser1Caller, []string{orgUser2.ID, org1.ID})
	test_utils.AssertTrue(t, err != nil, "Expected AddPermissionOrgAdmin to fail")
	mstub.MockTransactionEnd("t123")

	// get orgUser1 back
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	getSolutionUserResult, err = GetSolutionUserWithParams(stub, org1Caller, orgUser1.ID, true, true)
	test_utils.AssertTrue(t, err == nil, "Expected GetSolutionUserWithParams to succeed")
	test_utils.AssertTrue(t, getSolutionUserResult.SolutionInfo.IsOrgAdmin == false, "Expected GetUserData to succeed")
	mstub.MockTransactionEnd("t123")
}

func TestAddAndRemoveServiceAdminPermission(t *testing.T) {
	logger.SetLevel(shim.LogDebug)
	logger.Info("TestAddAndRemoveServiceAdminPermission function called")

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

	// register org user
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	orgUser1 := CreateTestSolutionUser("orgUser1")
	orgUser1.Org = "org1"
	orgUser1Bytes, _ := json.Marshal(&orgUser1)
	org1Caller, _ := user_mgmt.GetUserData(stub, org1, org1.ID, true, true)
	_, err = RegisterUser(stub, org1Caller, []string{string(orgUser1Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterUser OMR to succeed")
	mstub.MockTransactionEnd("t123")

	// put orgUser1 in org1
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	orgUser1Caller, _ := user_mgmt.GetUserData(stub, org1Caller, orgUser1.ID, true, true)
	_, err = PutUserInOrg(stub, org1Caller, []string{orgUser1.ID, org1.ID, "false"})
	test_utils.AssertTrue(t, err == nil, "Expected PutUserInOrg to succeed")
	mstub.MockTransactionEnd("t123")

	//  register datatype
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	datatype1 := Datatype{DatatypeID: "datatype1", Description: "datatype1"}
	datatype1Bytes, _ := json.Marshal(&datatype1)
	_, err = RegisterDatatype(stub, org1Caller, []string{string(datatype1Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterDatatype to succeed")
	mstub.MockTransactionEnd("t123")

	// register service1
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub, true, true, true)
	serviceDatatype1 := GenerateServiceDatatypeForTesting("datatype1", "service1", []string{consentOptionWrite, consentOptionRead})
	service1 := GenerateServiceForTesting("service1", org1.ID, []ServiceDatatype{serviceDatatype1})
	service1Bytes, _ := json.Marshal(&service1)
	_, err = RegisterService(stub, org1Caller, []string{string(service1Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterService to succeed")
	mstub.MockTransactionEnd("t123")

	// register service2
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub, true, true, true)
	service2 := GenerateServiceForTesting("service2", org1.ID, []ServiceDatatype{serviceDatatype1})
	service2Bytes, _ := json.Marshal(&service2)
	_, err = RegisterService(stub, org1Caller, []string{string(service2Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterService to succeed")
	mstub.MockTransactionEnd("t123")

	// orgUser1 is not service admin of service 1 yet
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	service, err := GetServiceInternal(stub, orgUser1Caller, "service1", true)
	test_utils.AssertTrue(t, err == nil, "Expected GetServiceInternal to succeed")
	test_utils.AssertTrue(t, service.SolutionPrivateData == nil, "Expected GetServiceInternal to succeed")
	mstub.MockTransactionEnd("t123")

	// give service1 admin permission to orguser1
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	service1Subgroup, _ := user_mgmt.GetUserData(stub, org1, "service1", true, true)
	_, err = AddPermissionServiceAdmin(stub, org1Caller, []string{orgUser1.ID, service1Subgroup.ID})
	test_utils.AssertTrue(t, err == nil, "Expected AddPermissionServiceAdmin to succeed")
	mstub.MockTransactionEnd("t123")

	// orgUser1 is service admin of service 1, should see private data
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	service, err = GetServiceInternal(stub, orgUser1Caller, "service1", true)
	test_utils.AssertTrue(t, err == nil, "Expected GetServiceInternal to succeed")
	test_utils.AssertTrue(t, service.SolutionPrivateData != nil, "Expected GetServiceInternal to succeed")
	mstub.MockTransactionEnd("t123")

	// give service1 admin permission to orguser1 again
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = AddPermissionServiceAdmin(stub, org1Caller, []string{orgUser1.ID, service1Subgroup.ID})
	test_utils.AssertTrue(t, err == nil, "Expected AddPermissionServiceAdmin to succeed")
	mstub.MockTransactionEnd("t123")

	// give service2 admin permission to orguser1
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	service2Subgroup, _ := user_mgmt.GetUserData(stub, org1Caller, "service2", true, true)
	_, err = AddPermissionServiceAdmin(stub, org1Caller, []string{orgUser1.ID, service2Subgroup.ID})
	test_utils.AssertTrue(t, err == nil, "Expected AddPermissionServiceAdmin to succeed")
	mstub.MockTransactionEnd("t123")

	// get user back
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	getSolutionUserResult, err := GetSolutionUserWithParams(stub, org1Caller, orgUser1.ID, true, true)
	test_utils.AssertTrue(t, err == nil, "Expected GetUserData to succeed")
	test_utils.AssertTrue(t, getSolutionUserResult.SolutionInfo.IsOrgAdmin == false, "Expected GetSolutionUserWithParams to succeed")
	test_utils.AssertTrue(t, len(getSolutionUserResult.SolutionInfo.Services) == 2, "Expected GetSolutionUserWithParams to succeed")
	mstub.MockTransactionEnd("t123")

	// remove service1 admin permission from orgUser1
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = RemovePermissionServiceAdmin(stub, org1Caller, []string{orgUser1.ID, service1Subgroup.ID})
	test_utils.AssertTrue(t, err == nil, "Expected RemovePermissionServiceAdmin to succeed")
	mstub.MockTransactionEnd("t123")

	// get user back
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	getSolutionUserResult, err = GetSolutionUserWithParams(stub, org1Caller, orgUser1.ID, true, true)
	test_utils.AssertTrue(t, err == nil, "Expected GetUserData to succeed")
	// only service2
	test_utils.AssertTrue(t, len(getSolutionUserResult.SolutionInfo.Services) == 1, "Expected GetSolutionUserWithParams to succeed")
	mstub.MockTransactionEnd("t123")

	// register orgUser2
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	orgUser2 := CreateTestSolutionUser("orgUser2")
	orgUser2.Org = "org1"
	orgUser2Bytes, _ := json.Marshal(&orgUser2)
	_, err = RegisterUser(stub, org1Caller, []string{string(orgUser2Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterUser OMR to succeed")
	mstub.MockTransactionEnd("t123")

	// put orgUser2 in org1
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = PutUserInOrg(stub, org1Caller, []string{orgUser2.ID, org1.ID, "false"})
	test_utils.AssertTrue(t, err == nil, "Expected PutUserInOrg to succeed")
	mstub.MockTransactionEnd("t123")

	// give org admin permission
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = AddPermissionOrgAdmin(stub, org1Caller, []string{orgUser2.ID, org1.ID})
	test_utils.AssertTrue(t, err == nil, "Expected AddPermissionOrgAdmin to succeed")
	mstub.MockTransactionEnd("t123")

	// get user back to make sure admin is true
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	getSolutionUserResult, err = GetSolutionUserWithParams(stub, org1Caller, orgUser2.ID, true, true)
	test_utils.AssertTrue(t, err == nil, "Expected GetSolutionUserWithParams to succeed")
	test_utils.AssertTrue(t, getSolutionUserResult.SolutionInfo.IsOrgAdmin == true, "Expected GetUserData to succeed")
	mstub.MockTransactionEnd("t123")

	// give service1 admin permission to orgUser1 again as orgUser2 (org admin)
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	orgUser2Caller, _ := user_mgmt.GetUserData(stub, org1, orgUser2.ID, true, true)
	_, err = AddPermissionServiceAdmin(stub, orgUser2Caller, []string{orgUser1.ID, service1Subgroup.ID})
	test_utils.AssertTrue(t, err == nil, "Expected AddPermissionServiceAdmin to succeed")
	mstub.MockTransactionEnd("t123")

	// get orgUser1 back, should have service admin permission to 2 services now
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	getSolutionUserResult, err = GetSolutionUserWithParams(stub, orgUser2Caller, orgUser1.ID, true, true)
	test_utils.AssertTrue(t, err == nil, "Expected GetUserData to succeed")
	test_utils.AssertTrue(t, len(getSolutionUserResult.SolutionInfo.Services) == 2, "Expected GetSolutionUserWithParams to succeed")
	mstub.MockTransactionEnd("t123")

	// remove service1 admin permission from orgUser1 as orgUser2 (org admin)
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = RemovePermissionServiceAdmin(stub, orgUser2Caller, []string{orgUser1.ID, service1Subgroup.ID})
	test_utils.AssertTrue(t, err == nil, "Expected RemovePermissionServiceAdmin to succeed")
	mstub.MockTransactionEnd("t123")

	// get orgUser1 back, should have service admin permission to 1 service
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	getSolutionUserResult, err = GetSolutionUserWithParams(stub, orgUser2Caller, orgUser1.ID, true, true)
	test_utils.AssertTrue(t, err == nil, "Expected GetUserData to succeed")
	test_utils.AssertTrue(t, len(getSolutionUserResult.SolutionInfo.Services) == 1, "Expected GetSolutionUserWithParams to succeed")
	mstub.MockTransactionEnd("t123")
}

func TestPutAndRemoveUserInGroup(t *testing.T) {
	logger.SetLevel(shim.LogDebug)
	logger.Info("TestPutAndRemoveUserInGroup function called")

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

	// register org user
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	org1Caller, _ := user_mgmt.GetUserData(stub, org1, org1.ID, true, true)
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

	//  register datatype
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	datatype1 := Datatype{DatatypeID: "datatype1", Description: "datatype1"}
	datatype1Bytes, _ := json.Marshal(&datatype1)
	_, err = RegisterDatatype(stub, org1Caller, []string{string(datatype1Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterDatatype to succeed")
	mstub.MockTransactionEnd("t123")

	// register service1
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub, true, true, true)
	serviceDatatype1 := GenerateServiceDatatypeForTesting("datatype1", "service1", []string{consentOptionWrite, consentOptionRead})
	service1 := GenerateServiceForTesting("service1", org1.ID, []ServiceDatatype{serviceDatatype1})
	service1Bytes, _ := json.Marshal(&service1)
	_, err = RegisterService(stub, org1Caller, []string{string(service1Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterService to succeed")
	mstub.MockTransactionEnd("t123")

	// register service2
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub, true, true, true)
	service2 := GenerateServiceForTesting("service2", org1.ID, []ServiceDatatype{serviceDatatype1})
	service2Bytes, _ := json.Marshal(&service2)
	_, err = RegisterService(stub, org1Caller, []string{string(service2Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterService to succeed")
	mstub.MockTransactionEnd("t123")

	// put orgUser1 in service 1 as admin
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	service1Caller, _ := user_mgmt.GetUserData(stub, org1Caller, "service1", true, true)
	_, err = PutUserInOrg(stub, org1Caller, []string{orgUser1.ID, service1Caller.ID, "true"})
	test_utils.AssertTrue(t, err == nil, "Expected PutUserInOrg to succeed")
	mstub.MockTransactionEnd("t123")

	// put orgUser1 in org1 as org admin
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = PutUserInOrg(stub, org1Caller, []string{orgUser1.ID, org1.ID, "true"})
	test_utils.AssertTrue(t, err == nil, "Expected PutUserInOrg to succeed")
	mstub.MockTransactionEnd("t123")

	// get user back
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	getSolutionUserResult, err := GetSolutionUserWithParams(stub, org1Caller, orgUser1.ID, true, true)
	test_utils.AssertTrue(t, err == nil, "Expected GetUserData to succeed")
	test_utils.AssertTrue(t, getSolutionUserResult.SolutionInfo.IsOrgAdmin == true, "Expected GetSolutionUserWithParams to succeed")
	mstub.MockTransactionEnd("t123")

	// remove org1 admin permission by removing user from org
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = RemoveUserFromOrg(stub, org1Caller, []string{orgUser1.ID, org1.ID})
	test_utils.AssertTrue(t, err == nil, "Expected RemoveUserFromOrg to succeed")
	mstub.MockTransactionEnd("t123")

	// get user back
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	getSolutionUserResult, err = GetSolutionUserWithParams(stub, org1Caller, orgUser1.ID, true, true)
	test_utils.AssertTrue(t, err == nil, "Expected GetUserData to succeed")
	test_utils.AssertTrue(t, getSolutionUserResult.SolutionInfo.IsOrgAdmin == false, "Expected GetSolutionUserWithParams to succeed")
	mstub.MockTransactionEnd("t123")
}

func TestRegisterAndUpdateOrg(t *testing.T) {
	logger.SetLevel(shim.LogDebug)
	logger.Info("TestRegisterAndUpdateOrg function called")

	mstub := SetupIndexesAndGetStub(t)
	mstub.MockTransactionStart("t1")
	stub := cached_stub.NewCachedStub(mstub)
	init_common.Init(stub)
	mstub.MockTransactionEnd("t1")

	// register org
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	org1 := test_utils.CreateTestGroup("org1")
	org1Bytes, _ := json.Marshal(&org1)
	_, err := RegisterOrg(stub, org1, []string{string(org1Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterOrg to succeed")
	mstub.MockTransactionEnd("t123")

	// get org back
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	org1Caller, _ := user_mgmt.GetUserData(stub, org1, org1.ID, true, true)
	getSolutionUserResult, err := GetSolutionUserWithParams(stub, org1Caller, org1.ID, true, true)
	test_utils.AssertTrue(t, err == nil, "Expected GetUserData to succeed")
	test_utils.AssertTrue(t, getSolutionUserResult.Role == "org", "Expected GetSolutionUserWithParams to succeed")
	test_utils.AssertTrue(t, getSolutionUserResult.Org == org1.ID, "Expected GetUserData to succeed")
	mstub.MockTransactionEnd("t123")

	// Update org
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	org1.Name = "new name"
	newData := make(map[string]interface{})
	newData["new data"] = "data"
	org1.SolutionPrivateData = newData
	org1Bytes, _ = json.Marshal(&org1)
	_, err = UpdateOrg(stub, org1Caller, []string{string(org1Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected UpdateOrg to succeed")
	mstub.MockTransactionEnd("t123")

	// get org back
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	getSolutionUserResult, err = GetSolutionUserWithParams(stub, org1Caller, org1.ID, true, true)
	test_utils.AssertTrue(t, err == nil, "Expected GetUserData to succeed")
	test_utils.AssertTrue(t, getSolutionUserResult.Role == "org", "Expected GetSolutionUserWithParams to succeed")
	test_utils.AssertTrue(t, getSolutionUserResult.Name == "new name", "Expected GetSolutionUserWithParams to succeed")
	test_utils.AssertTrue(t, getSolutionUserResult.Org == org1.ID, "Expected GetUserData to succeed")
	mstub.MockTransactionEnd("t123")

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

	// give org admin permission
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = AddPermissionOrgAdmin(stub, org1Caller, []string{orgUser1.ID, org1.ID})
	test_utils.AssertTrue(t, err == nil, "Expected AddPermissionOrgAdmin to succeed")
	mstub.MockTransactionEnd("t123")

	// get user back to make sure admin is true
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	getSolutionUserResult2, err := GetSolutionUserWithParams(stub, org1Caller, orgUser1.ID, true, true)
	test_utils.AssertTrue(t, err == nil, "Expected GetSolutionUserWithParams to succeed")
	test_utils.AssertTrue(t, getSolutionUserResult2.SolutionInfo.IsOrgAdmin == true, "Expected GetUserData to succeed")
	mstub.MockTransactionEnd("t123")

	// Update org as orgUser1
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	org1.Name = "newest name"
	org1Bytes, _ = json.Marshal(&org1)
	orgUser1Caller, _ := user_mgmt.GetUserData(stub, org1Caller, orgUser1.ID, true, true)
	_, err = UpdateOrg(stub, orgUser1Caller, []string{string(org1Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected UpdateOrg to succeed")
	mstub.MockTransactionEnd("t123")

	// get org back
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	getSolutionUserResult, err = GetSolutionUserWithParams(stub, org1Caller, org1.ID, true, true)
	test_utils.AssertTrue(t, err == nil, "Expected GetUserData to succeed")
	test_utils.AssertTrue(t, getSolutionUserResult.Role == "org", "Expected GetSolutionUserWithParams to succeed")
	test_utils.AssertTrue(t, getSolutionUserResult.Name == "newest name", "Expected GetSolutionUserWithParams to succeed")
	test_utils.AssertTrue(t, getSolutionUserResult.Org == org1.ID, "Expected GetUserData to succeed")
	mstub.MockTransactionEnd("t123")
}

func TestRegisterOrg_OffChain(t *testing.T) {
	logger.SetLevel(shim.LogDebug)
	logger.Info("TestRegisterOrg_OffChain function called")

	mstub := SetupIndexesAndGetStub(t)
	mstub.MockTransactionStart("t1")
	stub := cached_stub.NewCachedStub(mstub)
	init_common.Init(stub)
	mstub.MockTransactionEnd("t1")

	// register system admin user
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
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

	// register org
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	org1 := test_utils.CreateTestGroup("org1")
	org1Bytes, _ := json.Marshal(&org1)
	_, err = RegisterOrg(stub, org1, []string{string(org1Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterOrg to succeed")
	mstub.MockTransactionEnd("t123")

	// get org
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	org1Caller, _ := user_mgmt.GetUserData(stub, org1, org1.ID, true, true)
	getSolutionUserResult, err := GetSolutionUserWithParams(stub, org1Caller, org1.ID, true, true)
	test_utils.AssertTrue(t, err == nil, "Expected GetUserData to succeed")
	test_utils.AssertTrue(t, getSolutionUserResult.Role == "org", "Expected GetSolutionUserWithParams to succeed")
	test_utils.AssertTrue(t, getSolutionUserResult.Org == org1.ID, "Expected GetUserData to succeed")
	mstub.MockTransactionEnd("t123")

	// verify data access for org1Caller
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	assetManager := asset_mgmt.GetAssetManager(stub, org1Caller)
	orgAssetID := user_mgmt.GetUserAssetID(org1.ID)
	keyPath, err := GetKeyPath(stub, org1Caller, orgAssetID)
	test_utils.AssertTrue(t, err == nil, "Expected GetKeyPath consent to succeed")
	orgAssetKey, err := assetManager.GetAssetKey(orgAssetID, keyPath)
	test_utils.AssertTrue(t, err == nil, "Expected GetAssetKey to succeed")
	orgAsset, err := assetManager.GetAsset(orgAssetID, orgAssetKey)
	test_utils.AssertTrue(t, err == nil, "Expected GetAsset to succeed")
	mstub.MockTransactionEnd("t123")

	// org1Caller has access to public data
	orgPublicData := data_model.UserPublicData{}
	json.Unmarshal(orgAsset.PublicData, &orgPublicData)
	test_utils.AssertTrue(t, orgPublicData.ID != "", "Expected ID")
	test_utils.AssertTrue(t, orgPublicData.Name != "", "Expected Name")
	test_utils.AssertTrue(t, orgPublicData.Role != "", "Expected Role")
	test_utils.AssertTrue(t, orgPublicData.PublicKeyB64 != "", "Expected PublicKeyB64")
	test_utils.AssertTrue(t, orgPublicData.IsGroup == true, "Expected IsGroup")
	test_utils.AssertTrue(t, orgPublicData.Status != "", "Expected Status")
	test_utils.AssertTrue(t, orgPublicData.ConnectionID == datastoreConnectionID, "Expected ConnectionID")
	test_utils.AssertTrue(t, orgPublicData.SolutionPublicData != nil, "Expected SolutionPublicData")

	// org1Caller has access to private data
	orgPrivateData := data_model.UserPrivateData{}
	json.Unmarshal(orgAsset.PrivateData, &orgPrivateData)
	test_utils.AssertTrue(t, orgPrivateData.Email != "", "Expected Email")
	test_utils.AssertTrue(t, orgPrivateData.KmsPublicKeyId != "", "Expected KmsPublicKeyId")
	test_utils.AssertTrue(t, orgPrivateData.KmsPrivateKeyId != "", "Expected KmsPrivateKeyId")
	test_utils.AssertTrue(t, orgPrivateData.KmsSymKeyId != "", "Expected KmsSymKeyId")
	test_utils.AssertTrue(t, orgPrivateData.Secret != "", "Expected Secret")
	test_utils.AssertTrue(t, orgPrivateData.SolutionPrivateData != nil, "Expected SolutionPrivateData")

	// register unrelated user
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	unrelatedUser := test_utils.CreateTestUser("di-unrelatedUser")
	unrelatedUserBytes, _ := json.Marshal(&unrelatedUser)
	_, err = RegisterUser(stub, unrelatedUser, []string{string(unrelatedUserBytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterUser to succeed")
	mstub.MockTransactionEnd("t123")

	// verify data access for unrelatedUser
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	assetManager = asset_mgmt.GetAssetManager(stub, unrelatedUser)
	keyPath, err = GetKeyPath(stub, unrelatedUser, orgAssetID)
	test_utils.AssertTrue(t, err == nil, "Expected GetKeyPath consent to succeed")
	_, err = assetManager.GetAssetKey(orgAssetID, keyPath)
	test_utils.AssertTrue(t, err != nil, "Expected GetAssetKey to fail")
	orgAsset, err = assetManager.GetAsset(orgAssetID, data_model.Key{})
	test_utils.AssertTrue(t, err == nil, "Expected GetAsset to succeed")
	mstub.MockTransactionEnd("t123")

	// unrelatedUser has access to public data
	orgPublicData = data_model.UserPublicData{}
	json.Unmarshal(orgAsset.PublicData, &orgPublicData)
	test_utils.AssertTrue(t, orgPublicData.ID != "", "Expected ID")
	test_utils.AssertTrue(t, orgPublicData.Name != "", "Expected Name")
	test_utils.AssertTrue(t, orgPublicData.Role != "", "Expected Role")
	test_utils.AssertTrue(t, orgPublicData.PublicKeyB64 != "", "Expected PublicKeyB64")
	test_utils.AssertTrue(t, orgPublicData.IsGroup == true, "Expected IsGroup")
	test_utils.AssertTrue(t, orgPublicData.Status != "", "Expected Status")
	test_utils.AssertTrue(t, orgPublicData.ConnectionID == datastoreConnectionID, "Expected ConnectionID")
	test_utils.AssertTrue(t, orgPublicData.SolutionPublicData != nil, "Expected SolutionPublicData")

	// unrelatedUser has no access to private data
	orgPrivateData = data_model.UserPrivateData{}
	json.Unmarshal(orgAsset.PrivateData, &orgPrivateData)
	test_utils.AssertTrue(t, orgPrivateData.Email == "", "Expected no Email")
	test_utils.AssertTrue(t, orgPrivateData.KmsPublicKeyId == "", "Expected no KmsPublicKeyId")
	test_utils.AssertTrue(t, orgPrivateData.KmsPrivateKeyId == "", "Expected no KmsPrivateKeyId")
	test_utils.AssertTrue(t, orgPrivateData.KmsSymKeyId == "", "Expected no KmsSymKeyId")
	test_utils.AssertTrue(t, orgPrivateData.Secret == "", "Expected no Secret")
	test_utils.AssertTrue(t, orgPrivateData.SolutionPrivateData == nil, "Expected no SolutionPrivateData")

	// remove datastore connection
	mstub.MockTransactionStart("t123")
	err = datastore_manager.DeleteDatastoreConnection(stub, systemAdmin, datastoreConnectionID)
	test_utils.AssertTrue(t, err == nil, "Expected DeleteDatastoreConnection to succeed")
	mstub.MockTransactionEnd("t123")

	// attempt to get org
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	getSolutionUserResult, err = GetSolutionUserWithParams(stub, org1Caller, org1.ID, true, true)
	test_utils.AssertTrue(t, err != nil, "Expected GetSolutionUserWithParams to fail")
	mstub.MockTransactionEnd("t123")
}

func TestRegisterUser_OffChain(t *testing.T) {
	logger.SetLevel(shim.LogDebug)
	logger.Info("TestRegisterUser_OffChain function called")

	mstub := SetupIndexesAndGetStub(t)
	mstub.MockTransactionStart("t1")
	stub := cached_stub.NewCachedStub(mstub)
	init_common.Init(stub)
	mstub.MockTransactionEnd("t1")

	// register system admin user
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
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

	// register org
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	org1 := test_utils.CreateTestGroup("org1")
	org1Bytes, _ := json.Marshal(&org1)
	_, err = RegisterOrg(stub, org1, []string{string(org1Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterOrg to succeed")
	mstub.MockTransactionEnd("t123")

	// register user
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	user1 := CreateTestSolutionUser("user1")
	user1Caller, err := convertToPlatformUser(stub, user1)
	test_utils.AssertTrue(t, err == nil, "Expected convertToPlatformUser to succeed")

	user1.Org = "org1"
	user1Bytes, _ := json.Marshal(&user1)
	_, err = RegisterUser(stub, org1, []string{string(user1Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterUser to succeed")
	mstub.MockTransactionEnd("t123")

	// get user
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	getSolutionUserResult, err := GetSolutionUserWithParams(stub, user1Caller, user1.ID, true, true)
	test_utils.AssertTrue(t, err == nil, "Expected GetSolutionUserWithParams to succeed")
	test_utils.AssertTrue(t, getSolutionUserResult.Role == "user", "Expected Role user")
	test_utils.AssertTrue(t, getSolutionUserResult.Org == org1.ID, "Expected Org org1")
	mstub.MockTransactionEnd("t123")

	// verify data access for user1Caller
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	assetManager := asset_mgmt.GetAssetManager(stub, user1Caller)
	user1AssetID := user_mgmt.GetUserAssetID(user1.ID)
	keyPath, err := GetKeyPath(stub, user1Caller, user1AssetID)
	test_utils.AssertTrue(t, err == nil, "Expected GetKeyPath consent to succeed")
	user1AssetKey, err := assetManager.GetAssetKey(user1AssetID, keyPath)
	test_utils.AssertTrue(t, err == nil, "Expected GetAssetKey to succeed")
	user1Asset, err := assetManager.GetAsset(user1AssetID, user1AssetKey)
	test_utils.AssertTrue(t, err == nil, "Expected GetAsset to succeed")
	mstub.MockTransactionEnd("t123")

	// user1Caller has access to public data
	user1PublicData := data_model.UserPublicData{}
	json.Unmarshal(user1Asset.PublicData, &user1PublicData)
	test_utils.AssertTrue(t, user1PublicData.ID != "", "Expected ID")
	test_utils.AssertTrue(t, user1PublicData.Name != "", "Expected Name")
	test_utils.AssertTrue(t, user1PublicData.Role != "", "Expected Role")
	test_utils.AssertTrue(t, user1PublicData.PublicKeyB64 != "", "Expected PublicKeyB64")
	test_utils.AssertTrue(t, user1PublicData.IsGroup == false, "Expected IsGroup")
	test_utils.AssertTrue(t, user1PublicData.Status != "", "Expected Status")
	test_utils.AssertTrue(t, user1PublicData.ConnectionID == datastoreConnectionID, "Expected ConnectionID")
	test_utils.AssertTrue(t, user1PublicData.SolutionPublicData != nil, "Expected SolutionPublicData")

	// user1Caller has access to private data
	user1PrivateData := data_model.UserPrivateData{}
	json.Unmarshal(user1Asset.PrivateData, &user1PrivateData)
	test_utils.AssertTrue(t, user1PrivateData.Email != "", "Expected Email")
	test_utils.AssertTrue(t, user1PrivateData.KmsPublicKeyId != "", "Expected KmsPublicKeyId")
	test_utils.AssertTrue(t, user1PrivateData.KmsPrivateKeyId != "", "Expected KmsPrivateKeyId")
	test_utils.AssertTrue(t, user1PrivateData.KmsSymKeyId != "", "Expected KmsSymKeyId")
	test_utils.AssertTrue(t, user1PrivateData.Secret != "", "Expected Secret")
	test_utils.AssertTrue(t, user1PrivateData.SolutionPrivateData != nil, "Expected SolutionPrivateData")

	// register unrelated user
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	unrelatedUser := test_utils.CreateTestUser("di-unrelatedUser")
	unrelatedUserBytes, _ := json.Marshal(&unrelatedUser)
	_, err = RegisterUser(stub, unrelatedUser, []string{string(unrelatedUserBytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterUser to succeed")
	mstub.MockTransactionEnd("t123")

	// verify data access for unrelatedUser
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	assetManager = asset_mgmt.GetAssetManager(stub, unrelatedUser)
	keyPath, err = GetKeyPath(stub, unrelatedUser, user1AssetID)
	test_utils.AssertTrue(t, err == nil, "Expected GetKeyPath consent to succeed")
	_, err = assetManager.GetAssetKey(user1AssetID, keyPath)
	test_utils.AssertTrue(t, err != nil, "Expected GetAssetKey to fail")
	user1Asset, err = assetManager.GetAsset(user1AssetID, data_model.Key{})
	test_utils.AssertTrue(t, err == nil, "Expected GetAsset to succeed")
	mstub.MockTransactionEnd("t123")

	// user1Caller has access to public data
	user1PublicData = data_model.UserPublicData{}
	json.Unmarshal(user1Asset.PublicData, &user1PublicData)
	test_utils.AssertTrue(t, user1PublicData.ID != "", "Expected ID")
	test_utils.AssertTrue(t, user1PublicData.Name != "", "Expected Name")
	test_utils.AssertTrue(t, user1PublicData.Role != "", "Expected Role")
	test_utils.AssertTrue(t, user1PublicData.PublicKeyB64 != "", "Expected PublicKeyB64")
	test_utils.AssertTrue(t, user1PublicData.IsGroup == false, "Expected IsGroup")
	test_utils.AssertTrue(t, user1PublicData.Status != "", "Expected Status")
	test_utils.AssertTrue(t, user1PublicData.ConnectionID == datastoreConnectionID, "Expected ConnectionID")
	test_utils.AssertTrue(t, user1PublicData.SolutionPublicData != nil, "Expected SolutionPublicData")

	// user1Caller has no access to private data
	user1PrivateData = data_model.UserPrivateData{}
	json.Unmarshal(user1Asset.PrivateData, &user1PrivateData)
	test_utils.AssertTrue(t, user1PrivateData.Email == "", "Expected no Email")
	test_utils.AssertTrue(t, user1PrivateData.KmsPublicKeyId == "", "Expected no KmsPublicKeyId")
	test_utils.AssertTrue(t, user1PrivateData.KmsPrivateKeyId == "", "Expected no KmsPrivateKeyId")
	test_utils.AssertTrue(t, user1PrivateData.KmsSymKeyId == "", "Expected no KmsSymKeyId")
	test_utils.AssertTrue(t, user1PrivateData.Secret == "", "Expected no Secret")
	test_utils.AssertTrue(t, user1PrivateData.SolutionPrivateData == nil, "Expected no SolutionPrivateData")

	// remove datastore connection
	mstub.MockTransactionStart("t123")
	err = datastore_manager.DeleteDatastoreConnection(stub, systemAdmin, datastoreConnectionID)
	test_utils.AssertTrue(t, err == nil, "Expected DeleteDatastoreConnection to succeed")
	mstub.MockTransactionEnd("t123")

	// attempt to get user
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	getSolutionUserResult, err = GetSolutionUserWithParams(stub, user1Caller, user1.ID, true, true)
	test_utils.AssertTrue(t, err != nil, "Expected GetSolutionUserWithParams to fail")
	mstub.MockTransactionEnd("t123")
}

func TestGetOrg(t *testing.T) {
	logger.SetLevel(shim.LogDebug)
	logger.Info("TestGetOrg function called")

	mstub := SetupIndexesAndGetStub(t)
	mstub.MockTransactionStart("t1")
	stub := cached_stub.NewCachedStub(mstub)
	init_common.Init(stub)
	mstub.MockTransactionEnd("t1")

	// register org
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	org1 := test_utils.CreateTestGroup("org1")
	org1Bytes, _ := json.Marshal(&org1)
	_, err := RegisterOrg(stub, org1, []string{string(org1Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterOrg to succeed")
	mstub.MockTransactionEnd("t123")

	// get org back, as org
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	orgResultBytes, err := user_mgmt.GetOrg(stub, org1, []string{org1.ID})
	org := data_model.User{}
	_ = json.Unmarshal(orgResultBytes, &org)
	test_utils.AssertTrue(t, err == nil, "Expected GetUserData to succeed")
	test_utils.AssertTrue(t, org.ID == "org1", "Expected ID to be org1")
	test_utils.AssertTrue(t, org.Role == "org", "Expected role to be org")
	test_utils.AssertTrue(t, org.Email != "", "Expected Email")
	mstub.MockTransactionEnd("t123")

	// register org user
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	org1Caller, _ := user_mgmt.GetUserData(stub, org1, org1.ID, true, true)
	orgUser1 := CreateTestSolutionUser("orgUser1")
	orgUser1.Org = org1.ID
	orgUser1Bytes, _ := json.Marshal(&orgUser1)
	_, err = RegisterUser(stub, org1Caller, []string{string(orgUser1Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterUser to succeed")
	mstub.MockTransactionEnd("t123")

	// put orgUser1 in org1
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = PutUserInOrg(stub, org1Caller, []string{"orgUser1", org1.ID, "false"})
	test_utils.AssertTrue(t, err == nil, "Expected PutUserInOrg to succeed")
	mstub.MockTransactionEnd("t123")

	// get org back, as org user
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	orgUser1Caller, err := convertToPlatformUser(stub, orgUser1)
	test_utils.AssertTrue(t, err == nil, "Expected convertToPlatformUser to succeed")
	orgResultBytes, err = user_mgmt.GetOrg(stub, orgUser1Caller, []string{org1.ID})
	org = data_model.User{}
	_ = json.Unmarshal(orgResultBytes, &org)
	test_utils.AssertTrue(t, err == nil, "Expected GetUserData to succeed")
	test_utils.AssertTrue(t, org.ID == "org1", "Expected ID to be org1")
	test_utils.AssertTrue(t, org.Role == "org", "Expected role to be org")
	test_utils.AssertTrue(t, org.Email != "", "Expected Email")
	mstub.MockTransactionEnd("t123")

	// register unrelatedUser
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	unrelatedUser := CreateTestSolutionUser("unrelatedUser")
	unrelatedUserCaller, err := convertToPlatformUser(stub, unrelatedUser)
	test_utils.AssertTrue(t, err == nil, "Expected convertToPlatformUser to succeed")
	unrelatedUserBytes, _ := json.Marshal(&unrelatedUser)
	_, err = RegisterUser(stub, unrelatedUserCaller, []string{string(unrelatedUserBytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterUser to succeed")
	mstub.MockTransactionEnd("t123")

	// attempt to get org back, as unrelated user
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	orgResultBytes, err = user_mgmt.GetOrg(stub, unrelatedUserCaller, []string{org1.ID})
	org = data_model.User{}
	_ = json.Unmarshal(orgResultBytes, &org)
	test_utils.AssertTrue(t, err == nil, "Expected GetUserData to succeed")
	test_utils.AssertTrue(t, org.ID == "org1", "Expected ID to be org1")
	test_utils.AssertTrue(t, org.Role == "org", "Expected role to be org")
	test_utils.AssertTrue(t, org.Email == "", "Expected no Email")
	mstub.MockTransactionEnd("t123")

	//  register datatype
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	datatype1 := Datatype{DatatypeID: "datatype1", Description: "datatype1"}
	datatype1Bytes, _ := json.Marshal(&datatype1)
	_, err = RegisterDatatype(stub, org1Caller, []string{string(datatype1Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterDatatype to succeed")
	mstub.MockTransactionEnd("t123")

	// register service1
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub, true, true, true)
	serviceDatatype1 := GenerateServiceDatatypeForTesting("datatype1", "service1", []string{consentOptionWrite, consentOptionRead})
	service1 := GenerateServiceForTesting("service1", org1.ID, []ServiceDatatype{serviceDatatype1})
	service1Bytes, _ := json.Marshal(&service1)
	_, err = RegisterService(stub, org1Caller, []string{string(service1Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterService to succeed")
	mstub.MockTransactionEnd("t123")

	// give service1 admin permission to orguser1
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	service1Subgroup, _ := user_mgmt.GetUserData(stub, org1, "service1", true, true)
	_, err = AddPermissionServiceAdmin(stub, org1Caller, []string{orgUser1.ID, service1Subgroup.ID})
	test_utils.AssertTrue(t, err == nil, "Expected AddPermissionServiceAdmin to succeed")
	mstub.MockTransactionEnd("t123")

	// get org back, as service admin
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	orgResultBytes, err = user_mgmt.GetOrg(stub, orgUser1Caller, []string{org1.ID})
	org = data_model.User{}
	_ = json.Unmarshal(orgResultBytes, &org)
	test_utils.AssertTrue(t, err == nil, "Expected GetUserData to succeed")
	test_utils.AssertTrue(t, org.ID == "org1", "Expected ID to be org1")
	test_utils.AssertTrue(t, org.Role == "org", "Expected role to be org")
	test_utils.AssertTrue(t, org.Email != "", "Expected Email")
	mstub.MockTransactionEnd("t123")
}

func TestGetUsers(t *testing.T) {
	logger.SetLevel(shim.LogDebug)
	logger.Info("TestGetUsers function called")

	mstub := SetupIndexesAndGetStub(t)
	mstub.MockTransactionStart("t1")
	stub := cached_stub.NewCachedStub(mstub)
	init_common.Init(stub)
	mstub.MockTransactionEnd("t1")

	// register org
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	org1 := test_utils.CreateTestGroup("org1")
	org1Bytes, _ := json.Marshal(&org1)
	_, err := RegisterOrg(stub, org1, []string{string(org1Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterOrg to succeed")
	mstub.MockTransactionEnd("t123")

	// register org user
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	org1Caller, _ := user_mgmt.GetUserData(stub, org1, org1.ID, true, true)
	orgUser1 := CreateTestSolutionUser("orgUser1")
	orgUser1.Org = org1.ID
	orgUser1Bytes, _ := json.Marshal(&orgUser1)
	_, err = RegisterUser(stub, org1Caller, []string{string(orgUser1Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterUser OMR to succeed")
	mstub.MockTransactionEnd("t123")

	// put orgUser1 in org1
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = PutUserInOrg(stub, org1Caller, []string{"orgUser1", org1.ID, "false"})
	test_utils.AssertTrue(t, err == nil, "Expected PutUserInOrg to succeed")
	mstub.MockTransactionEnd("t123")

	// Search for all users of org 1
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	resultsBytes, err := GetSolutionUsers(stub, org1Caller, []string{"org1", "20"})
	test_utils.AssertTrue(t, err == nil, "Expected GetSolutionUsers to succeed")
	test_utils.AssertTrue(t, resultsBytes != nil, "Expected GetSolutionUsers to succeed")
	users := []SolutionUser{}
	json.Unmarshal(resultsBytes, &users)
	test_utils.AssertTrue(t, len(users) == 2, "Got search result correctly")
	mstub.MockTransactionEnd("t123")

	// Search for all users of org 1, specifying role as well
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	resultsBytes, err = GetSolutionUsers(stub, org1Caller, []string{"org1", "20", "org"})
	test_utils.AssertTrue(t, err == nil, "Expected GetSolutionUsers to succeed")
	test_utils.AssertTrue(t, resultsBytes != nil, "Expected GetSolutionUsers to succeed")
	json.Unmarshal(resultsBytes, &users)
	test_utils.AssertTrue(t, len(users) == 2, "Got search result correctly")
	mstub.MockTransactionEnd("t123")

	// register system admin
	mstub.MockTransactionStart("1")
	stub = cached_stub.NewCachedStub(mstub)
	systemAdmin := CreateTestSolutionUser("systemAdmin")
	systemAdmin.Role = SOLUTION_ROLE_SYSTEM
	systemAdminBytes, _ := json.Marshal(&systemAdmin)
	sysAdminCaller, err := convertToPlatformUser(stub, systemAdmin)
	test_utils.AssertTrue(t, err == nil, "Expected convertToPlatformUser to succeed")
	_, err = RegisterUser(stub, sysAdminCaller, []string{string(systemAdminBytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterSystemAdmin to succeed")
	mstub.MockTransactionEnd("1")

	// Search for all uses of role type system
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	resultsBytes, err = GetSolutionUsers(stub, org1Caller, []string{"*", "20", "system"})
	test_utils.AssertTrue(t, err == nil, "Expected GetSolutionUsers to succeed")
	test_utils.AssertTrue(t, resultsBytes != nil, "Expected GetSolutionUsers to succeed")
	json.Unmarshal(resultsBytes, &users)
	test_utils.AssertTrue(t, len(users) == 1, "Got search result correctly")
	mstub.MockTransactionEnd("t123")

	// register auditors
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	audit1 := CreateTestSolutionUser("audit1")
	audit1.Role = SOLUTION_ROLE_AUDIT
	audit1Bytes, _ := json.Marshal(&audit1)
	_, err = RegisterUser(stub, sysAdminCaller, []string{string(audit1Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterUser OMR to succeed")
	mstub.MockTransactionEnd("t123")

	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	audit2 := CreateTestSolutionUser("audit2")
	audit2.Role = SOLUTION_ROLE_AUDIT
	audit2Bytes, _ := json.Marshal(&audit2)
	_, err = RegisterUser(stub, sysAdminCaller, []string{string(audit2Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterUser OMR to succeed")
	mstub.MockTransactionEnd("t123")

	// Search for all uses of role type audit
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	resultsBytes, err = GetSolutionUsers(stub, org1Caller, []string{"*", "20", "audit"})
	test_utils.AssertTrue(t, err == nil, "Expected GetSolutionUsers to succeed")
	test_utils.AssertTrue(t, resultsBytes != nil, "Expected GetSolutionUsers to succeed")
	json.Unmarshal(resultsBytes, &users)
	test_utils.AssertTrue(t, len(users) == 2, "Got search result correctly")
	mstub.MockTransactionEnd("t123")
}

// Creates Test Solution User with random keys
func CreateTestSolutionUser(userID string) SolutionUser {
	testUser := SolutionUser{}
	testUser.ID = userID
	testUser.PrivateKey = test_utils.GeneratePrivateKey()
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(testUser.PrivateKey)
	testUser.PrivateKeyB64 = base64.StdEncoding.EncodeToString(privateKeyBytes)
	testUser.PublicKey = testUser.PrivateKey.Public().(*rsa.PublicKey)
	publicKeyBytes, _ := x509.MarshalPKIXPublicKey(testUser.PublicKey)
	testUser.PublicKeyB64 = base64.StdEncoding.EncodeToString(publicKeyBytes)
	testUser.SymKey = test_utils.GenerateSymKey()
	testUser.SymKeyB64 = base64.StdEncoding.EncodeToString(testUser.SymKey)
	testUser.KmsPublicKeyId = "kmspubkeyid"
	testUser.KmsPrivateKeyId = "kmsprivkeyid"
	testUser.KmsSymKeyId = "kmssymkeyid"
	testUser.Email = "none"
	testUser.Org = ""
	testUser.Name = userID
	testUser.IsGroup = false
	testUser.Status = "active"
	testUser.Secret = "pass0"
	testUser.Role = SOLUTION_ROLE_USER
	testUser.SolutionPublicData = make(map[string]interface{})
	testUser.SolutionPrivateData = make(map[string]interface{})
	return testUser
}
