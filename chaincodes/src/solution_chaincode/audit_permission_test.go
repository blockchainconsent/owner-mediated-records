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
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"testing"
)

func TestAddAndRemoveAuditorPermissionAsServiceAdmin(t *testing.T) {
	logger.SetLevel(shim.LogDebug)
	logger.Info("TestAddAndRemoveAuditorPermissionAsServiceAdmin function called")

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
	_, err = RegisterOrg(stub, systemAdmin, []string{string(org1Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterOrg to succeed")
	mstub.MockTransactionEnd("t123")

	// register auditor
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	audit1 := CreateTestSolutionUser("audit1")
	audit1.Role = SOLUTION_ROLE_AUDIT
	audit1Bytes, _ := json.Marshal(&audit1)
	_, err = RegisterUser(stub, systemAdmin, []string{string(audit1Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterUser OMR to succeed")
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

	// register service1
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub, true, true, true)
	serviceDatatype1 := GenerateServiceDatatypeForTesting("datatype1", "service1", []string{consentOptionWrite, consentOptionRead})
	service1 := GenerateServiceForTesting("service1", org1.ID, []ServiceDatatype{serviceDatatype1})
	service1Bytes, _ := json.Marshal(&service1)
	_, err = RegisterService(stub, org1Caller, []string{string(service1Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterService to succeed")
	mstub.MockTransactionEnd("t123")

	// give auditor audit permission
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	service1Caller, _ := user_mgmt.GetUserData(stub, org1Caller, "service1", true, true)
	auditPermissionAssetKey := test_utils.GenerateSymKey()
	auditPermissionAssetKeyB64 := crypto.EncodeToB64String(auditPermissionAssetKey)
	_, err = AddAuditorPermission(stub, service1Caller, []string{audit1.ID, "service1", auditPermissionAssetKeyB64})
	test_utils.AssertTrue(t, err == nil, "Expected AddAuditorPermission to succeed")
	mstub.MockTransactionEnd("t1")

	//remove audit permission for auditor
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = RemoveAuditorPermission(stub, service1Caller, []string{audit1.ID, "service1"})
	test_utils.AssertTrue(t, err == nil, "Expected RemoveAuditorPermission to succeed")
	mstub.MockTransactionEnd("t1")

	// give auditor audit permission one more time
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = AddAuditorPermission(stub, service1Caller, []string{audit1.ID, "service1", auditPermissionAssetKeyB64})
	test_utils.AssertTrue(t, err == nil, "Expected AddAuditorPermission to succeed")
	mstub.MockTransactionEnd("t1")
}

func TestAddAndRemoveAuditorPermissionAsOrgAdmin(t *testing.T) {
	logger.SetLevel(shim.LogDebug)
	logger.Info("TestAddAndRemoveAuditorPermissionAsOrgAdmin function called")

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
	_, err = RegisterOrg(stub, systemAdmin, []string{string(org1Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterOrg to succeed")
	mstub.MockTransactionEnd("t123")

	// register auditor
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	audit1 := CreateTestSolutionUser("audit1")
	audit1.Role = SOLUTION_ROLE_AUDIT
	audit1Bytes, _ := json.Marshal(&audit1)
	_, err = RegisterUser(stub, systemAdmin, []string{string(audit1Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterUser OMR to succeed")
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

	// register service1
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub, true, true, true)
	serviceDatatype1 := GenerateServiceDatatypeForTesting("datatype1", "service1", []string{consentOptionWrite, consentOptionRead})
	service1 := GenerateServiceForTesting("service1", org1.ID, []ServiceDatatype{serviceDatatype1})
	service1Bytes, _ := json.Marshal(&service1)
	_, err = RegisterService(stub, org1Caller, []string{string(service1Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterService to succeed")
	mstub.MockTransactionEnd("t123")

	// give auditor audit permission
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	auditPermissionAssetKey := test_utils.GenerateSymKey()
	auditPermissionAssetKeyB64 := crypto.EncodeToB64String(auditPermissionAssetKey)
	_, err = AddAuditorPermission(stub, org1Caller, []string{audit1.ID, "service1", auditPermissionAssetKeyB64})
	test_utils.AssertTrue(t, err == nil, "Expected AddAuditorPermission to succeed")
	mstub.MockTransactionEnd("t1")

	//remove audit permission for auditor
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = RemoveAuditorPermission(stub, org1Caller, []string{audit1.ID, "service1"})
	test_utils.AssertTrue(t, err == nil, "Expected RemoveAuditorPermission to succeed")
	mstub.MockTransactionEnd("t1")

	// give auditor audit permission one more time
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = AddAuditorPermission(stub, org1Caller, []string{audit1.ID, "service1", auditPermissionAssetKeyB64})
	test_utils.AssertTrue(t, err == nil, "Expected AddAuditorPermission to succeed")
	mstub.MockTransactionEnd("t1")
}

func TestAddAndRemoveAuditorPermissionAsOrgUser(t *testing.T) {
	logger.SetLevel(shim.LogDebug)
	logger.Info("TestAddAndRemoveAuditorPermissionAsOrgUser function called")

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
	_, err = RegisterOrg(stub, systemAdmin, []string{string(org1Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterOrg to succeed")
	mstub.MockTransactionEnd("t123")

	// register auditor
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	audit1 := CreateTestSolutionUser("audit1")
	audit1.Role = SOLUTION_ROLE_AUDIT
	audit1Bytes, _ := json.Marshal(&audit1)
	_, err = RegisterUser(stub, systemAdmin, []string{string(audit1Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterUser OMR to succeed")
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

	// register service1
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub, true, true, true)
	serviceDatatype1 := GenerateServiceDatatypeForTesting("datatype1", "service1", []string{consentOptionWrite, consentOptionRead})
	service1 := GenerateServiceForTesting("service1", org1.ID, []ServiceDatatype{serviceDatatype1})
	service1Bytes, _ := json.Marshal(&service1)
	_, err = RegisterService(stub, org1Caller, []string{string(service1Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterService to succeed")
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

	// give auditor audit permission
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	orgUser1Caller, _ := user_mgmt.GetUserData(stub, org1Caller, orgUser1.ID, true, true)
	auditPermissionAssetKey := test_utils.GenerateSymKey()
	auditPermissionAssetKeyB64 := crypto.EncodeToB64String(auditPermissionAssetKey)
	_, err = AddAuditorPermission(stub, orgUser1Caller, []string{audit1.ID, "service1", auditPermissionAssetKeyB64})
	test_utils.AssertTrue(t, err == nil, "Expected AddAuditorPermission to succeed")
	mstub.MockTransactionEnd("t1")

	//remove audit permission for auditor
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = RemoveAuditorPermission(stub, orgUser1Caller, []string{audit1.ID, "service1"})
	test_utils.AssertTrue(t, err == nil, "Expected RemoveAuditorPermission to succeed")
	mstub.MockTransactionEnd("t1")

	// give auditor audit permission one more time
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = AddAuditorPermission(stub, orgUser1Caller, []string{audit1.ID, "service1", auditPermissionAssetKeyB64})
	test_utils.AssertTrue(t, err == nil, "Expected AddAuditorPermission to succeed")
	mstub.MockTransactionEnd("t1")

	// remove org admin permission
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = RemovePermissionOrgAdmin(stub, org1Caller, []string{orgUser1.ID, org1.ID})
	test_utils.AssertTrue(t, err == nil, "Expected RemovePermissionOrgAdmin to succeed")
	mstub.MockTransactionEnd("t123")

	// give auditor audit permission without org/service admin permissions
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = AddAuditorPermission(stub, orgUser1Caller, []string{audit1.ID, "service1", auditPermissionAssetKeyB64})
	test_utils.AssertTrue(t, err != nil, "Expected AddAuditorPermission to fail")
	mstub.MockTransactionEnd("t1")

	// give service admin permission
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = AddPermissionServiceAdmin(stub, org1Caller, []string{orgUser1.ID, "service1"})
	test_utils.AssertTrue(t, err == nil, "Expected AddPermissionServiceAdmin to succeed")
	mstub.MockTransactionEnd("t123")

	// give auditor audit permission
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	orgUser1Caller, _ = user_mgmt.GetUserData(stub, org1Caller, orgUser1.ID, true, true)
	auditPermissionAssetKey = test_utils.GenerateSymKey()
	auditPermissionAssetKeyB64 = crypto.EncodeToB64String(auditPermissionAssetKey)
	_, err = AddAuditorPermission(stub, orgUser1Caller, []string{audit1.ID, "service1", auditPermissionAssetKeyB64})
	test_utils.AssertTrue(t, err == nil, "Expected AddAuditorPermission to succeed")
	mstub.MockTransactionEnd("t1")

	//remove audit permission for auditor
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = RemoveAuditorPermission(stub, orgUser1Caller, []string{audit1.ID, "service1"})
	test_utils.AssertTrue(t, err == nil, "Expected RemoveAuditorPermission to succeed")
	mstub.MockTransactionEnd("t1")

	// give auditor audit permission one more time
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = AddAuditorPermission(stub, orgUser1Caller, []string{audit1.ID, "service1", auditPermissionAssetKeyB64})
	test_utils.AssertTrue(t, err == nil, "Expected AddAuditorPermission to succeed")
	mstub.MockTransactionEnd("t1")
}
