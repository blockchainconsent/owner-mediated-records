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
	"common/bchcls/crypto"
	"common/bchcls/data_model"
	"common/bchcls/init_common"
	"common/bchcls/test_utils"
	"common/bchcls/user_mgmt"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"testing"
	"time"

	"github.com/hyperledger/fabric/core/chaincode/shim"
)

func querySetup(t *testing.T) (*test_utils.NewMockStub, data_model.User) {
	mstub := SetupIndexesAndGetStub(t)

	mstub.MockTransactionStart("t1")
	stub := cached_stub.NewCachedStub(mstub)
	init_common.Init(stub)
	err := SetupAuditPermissionIndex(stub)

	systemAdmin := test_utils.CreateTestUser("systemAdmin")
	systemAdmin.Role = SOLUTION_ROLE_SYSTEM
	systemAdminBytes, _ := json.Marshal(&systemAdmin)
	_, err = RegisterUser(stub, systemAdmin, []string{string(systemAdminBytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterUser to succeed")
	mstub.MockTransactionEnd("t1")

	return mstub, systemAdmin
}

func registerOrgUser(t *testing.T, mstub *test_utils.NewMockStub, org data_model.User, seqID string) data_model.User {
	orgID := org.ID
	orgUserID := orgID + "User" + seqID

	// register org user1
	mstub.MockTransactionStart("t1")
	stub := cached_stub.NewCachedStub(mstub)
	org1User1 := CreateTestSolutionUser(orgUserID)
	org1User1.Org = org.ID
	org1User1Bytes, _ := json.Marshal(&org1User1)
	_, err := RegisterUser(stub, org, []string{string(org1User1Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterUser to succeed")
	mstub.MockTransactionEnd("t1")

	// put org user1 in org
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = PutUserInOrg(stub, org, []string{orgUserID, org.ID, "false"})
	test_utils.AssertTrue(t, err == nil, "Expected PutUserInOrg to succeed")
	mstub.MockTransactionEnd("t1")

	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	orgUser1Caller, _ := user_mgmt.GetUserData(stub, org, org1User1.ID, true, true)
	mstub.MockTransactionEnd("t1")
	return orgUser1Caller
}

func registerServiceAdmin(t *testing.T, mstub *test_utils.NewMockStub, org data_model.User, systemAdmin data_model.User, seqID string) (data_model.User, data_model.User) {
	orgID := org.ID
	datatypeID := orgID + "Datatype" + seqID
	serviceID := orgID + "Service" + seqID
	serviceAdminID := orgID + "ServiceAdmin" + seqID

	// register auditor
	mstub.MockTransactionStart("t1")
	stub := cached_stub.NewCachedStub(mstub)
	auditor := CreateTestSolutionUser(orgID + serviceID + "auditor" + seqID)
	auditor.Role = SOLUTION_ROLE_AUDIT
	auditorBytes, _ := json.Marshal(&auditor)
	_, err := RegisterUser(stub, systemAdmin, []string{string(auditorBytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterUser to succeed")
	mstub.MockTransactionEnd("t1")

	// register org datatype
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	orgCaller, _ := user_mgmt.GetUserData(stub, org, org.ID, true, true)

	orgDatatype := Datatype{DatatypeID: datatypeID, Description: datatypeID}
	orgDatatypeBytes, _ := json.Marshal(&orgDatatype)
	_, err = RegisterDatatype(stub, orgCaller, []string{string(orgDatatypeBytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterDatatype to succeed")
	mstub.MockTransactionEnd("t1")

	// register org service
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub, true, true, true)
	orgServiceDatatype := GenerateServiceDatatypeForTesting(datatypeID, serviceID, []string{consentOptionWrite, consentOptionRead})
	orgService := GenerateServiceForTesting(serviceID, org.ID, []ServiceDatatype{orgServiceDatatype})
	orgServiceBytes, _ := json.Marshal(&orgService)
	_, err = RegisterService(stub, orgCaller, []string{string(orgServiceBytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterService to succeed")
	mstub.MockTransactionEnd("t1")

	// register org service admin
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	orgServiceAdmin := CreateTestSolutionUser(serviceAdminID)
	orgServiceAdmin.Org = org.ID
	orgServiceAdminBytes, _ := json.Marshal(&orgServiceAdmin)
	_, err = RegisterUser(stub, org, []string{string(orgServiceAdminBytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterUser to succeed")
	mstub.MockTransactionEnd("t1")

	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = PutUserInOrg(stub, org, []string{serviceAdminID, org.ID, "false"})
	test_utils.AssertTrue(t, err == nil, "Expected PutUserInOrg to succeed")
	mstub.MockTransactionEnd("t1")

	// give service admin permission
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = AddPermissionServiceAdmin(stub, orgCaller, []string{orgServiceAdmin.ID, serviceID})
	test_utils.AssertTrue(t, err == nil, "Expected AddPermissionServiceAdmin to succeed")
	mstub.MockTransactionEnd("t1")

	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	orgServiceAdminCaller, _ := user_mgmt.GetUserData(stub, orgCaller, orgServiceAdmin.ID, true, true)
	mstub.MockTransactionEnd("t1")

	// give auditor audit permission
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	auditPermissionAssetKey := test_utils.GenerateSymKey()
	auditPermissionAssetKeyB64 := crypto.EncodeToB64String(auditPermissionAssetKey)
	serviceCaller, _ := user_mgmt.GetUserData(stub, orgCaller, serviceID, true, true)
	_, err = AddAuditorPermission(stub, serviceCaller, []string{auditor.ID, serviceID, auditPermissionAssetKeyB64})
	test_utils.AssertTrue(t, err == nil, "Expected AddAuditorPermission to succeed")
	mstub.MockTransactionEnd("t1")

	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	auditorPlatformUser, err := convertToPlatformUser(stub, auditor)
	test_utils.AssertTrue(t, err == nil, "Expected convertToPlatformUser to succeed")
	auditorCaller, _ := user_mgmt.GetUserData(stub, auditorPlatformUser, auditor.ID, true, true)
	mstub.MockTransactionEnd("t1")

	return orgServiceAdminCaller, auditorCaller
}

func TestQueryOrg(t *testing.T) {
	logger.SetLevel(shim.LogDebug)
	logger.Info("TestQueryOrg function called")

	mstub, systemAdmin := querySetup(t)

	// ------- ORG 1 -------

	// register org1
	mstub.MockTransactionStart("t1")
	stub := cached_stub.NewCachedStub(mstub)
	org1 := test_utils.CreateTestGroup("org1")
	org1Bytes, _ := json.Marshal(&org1)
	_, err := RegisterOrg(stub, org1, []string{string(org1Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterOrg to succeed")
	mstub.MockTransactionEnd("t1")

	org1UserCaller := registerOrgUser(t, mstub, org1, "1")

	org1Service1AdminCaller, auditor1Caller := registerServiceAdmin(t, mstub, org1, systemAdmin, "1")
	org1Service2AdminCaller, _ := registerServiceAdmin(t, mstub, org1, systemAdmin, "2")

	// ------- ORG 2 -------

	// register org2
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	org2 := test_utils.CreateTestGroup("org2")
	org2Bytes, _ := json.Marshal(&org2)
	_, err = RegisterOrg(stub, org2, []string{string(org2Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterOrg to succeed")
	mstub.MockTransactionEnd("t1")

	org2UserCaller := registerOrgUser(t, mstub, org2, "1")

	org2Service1AdminCaller, _ := registerServiceAdmin(t, mstub, org2, systemAdmin, "1")

	// ------- PATIENTS -------

	// register patient1
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	patient1 := CreateTestSolutionUser("patient1")
	patient1PlatformUser, err := convertToPlatformUser(stub, patient1)
	test_utils.AssertTrue(t, err == nil, "Expected convertToPlatformUser to succeed")
	patient1Bytes, _ := json.Marshal(&patient1)
	_, err = RegisterUser(stub, patient1PlatformUser, []string{string(patient1Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterUser to succeed")
	mstub.MockTransactionEnd("t1")

	// register patient2
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	patient2 := CreateTestSolutionUser("patient2")
	patient2Caller, err := convertToPlatformUser(stub, patient2)
	test_utils.AssertTrue(t, err == nil, "Expected convertToPlatformUser to succeed")
	patient2Bytes, _ := json.Marshal(&patient2)
	_, err = RegisterUser(stub, patient2Caller, []string{string(patient2Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterUser to succeed")
	mstub.MockTransactionEnd("t1")

	// ------- GET ORG -------

	// get org1 as org1 admin
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	org1Caller, _ := user_mgmt.GetUserData(stub, org1, org1.ID, true, true)
	getSolutionUserResult, err := GetSolutionUserWithParams(stub, org1Caller, org1.ID, true, true)
	test_utils.AssertTrue(t, err == nil, "Expected GetUserData to succeed")
	test_utils.AssertTrue(t, getSolutionUserResult.ID == org1.ID, "Expected org ID")
	test_utils.AssertTrue(t, getSolutionUserResult.Email != "", "Expected org Email")
	mstub.MockTransactionEnd("t1")

	// get org1 as org1 user
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	getSolutionUserResult, err = GetSolutionUserWithParams(stub, org1UserCaller, org1.ID, true, true)
	test_utils.AssertTrue(t, err == nil, "Expected GetUserData to succeed")
	test_utils.AssertTrue(t, getSolutionUserResult.ID == org1.ID, "Expected org ID")
	test_utils.AssertTrue(t, getSolutionUserResult.Email != "", "Expected org Email")
	mstub.MockTransactionEnd("t1")

	// get org1 as org1 service1 admin
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	getSolutionUserResult, err = GetSolutionUserWithParams(stub, org1Service1AdminCaller, org1.ID, true, true)
	test_utils.AssertTrue(t, err == nil, "Expected GetUserData to succeed")
	test_utils.AssertTrue(t, getSolutionUserResult.ID == org1.ID, "Expected org ID")
	test_utils.AssertTrue(t, getSolutionUserResult.Email != "", "Expected org Email")
	mstub.MockTransactionEnd("t1")

	// get org1 as org1 service2 admin
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	getSolutionUserResult, err = GetSolutionUserWithParams(stub, org1Service2AdminCaller, org1.ID, true, true)
	test_utils.AssertTrue(t, err == nil, "Expected GetUserData to succeed")
	test_utils.AssertTrue(t, getSolutionUserResult.ID == org1.ID, "Expected org ID")
	test_utils.AssertTrue(t, getSolutionUserResult.Email != "", "Expected org Email")
	mstub.MockTransactionEnd("t1")

	// get org1 as org2 admin (X)
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	org2Caller, _ := user_mgmt.GetUserData(stub, org2, org2.ID, true, true)
	getSolutionUserResult, err = GetSolutionUserWithParams(stub, org2Caller, org1.ID, true, true)
	test_utils.AssertTrue(t, err == nil, "Expected GetUserData to succeed")
	test_utils.AssertTrue(t, getSolutionUserResult.ID == org1.ID, "Expected org ID")
	test_utils.AssertTrue(t, getSolutionUserResult.Email == "", "Expected no org Email")
	mstub.MockTransactionEnd("t1")

	// get org1 as org2 user (X)
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	getSolutionUserResult, err = GetSolutionUserWithParams(stub, org2UserCaller, org1.ID, true, true)
	test_utils.AssertTrue(t, err == nil, "Expected GetUserData to succeed")
	test_utils.AssertTrue(t, getSolutionUserResult.ID == org1.ID, "Expected org ID")
	test_utils.AssertTrue(t, getSolutionUserResult.Email == "", "Expected no org Email")
	mstub.MockTransactionEnd("t1")

	// get org1 as org2 service admin (X)
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	getSolutionUserResult, err = GetSolutionUserWithParams(stub, org2Service1AdminCaller, org1.ID, true, true)
	test_utils.AssertTrue(t, err == nil, "Expected GetUserData to succeed")
	test_utils.AssertTrue(t, getSolutionUserResult.ID == org1.ID, "Expected org ID")
	test_utils.AssertTrue(t, getSolutionUserResult.Email == "", "Expected no org Email")
	mstub.MockTransactionEnd("t1")

	// get org1 as patient1 (X)
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	patient1Caller, _ := user_mgmt.GetUserData(stub, patient1PlatformUser, patient1.ID, true, true)
	getSolutionUserResult, err = GetSolutionUserWithParams(stub, patient1Caller, org1.ID, true, true)
	test_utils.AssertTrue(t, err == nil, "Expected GetUserData to succeed")
	test_utils.AssertTrue(t, getSolutionUserResult.ID == org1.ID, "Expected org ID")
	test_utils.AssertTrue(t, getSolutionUserResult.Email == "", "Expected no org Email")
	mstub.MockTransactionEnd("t1")

	// get org1 as patient2 (X)
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	getSolutionUserResult, err = GetSolutionUserWithParams(stub, patient2Caller, org1.ID, true, true)
	test_utils.AssertTrue(t, err == nil, "Expected GetUserData to succeed")
	test_utils.AssertTrue(t, getSolutionUserResult.ID == org1.ID, "Expected org ID")
	test_utils.AssertTrue(t, getSolutionUserResult.Email == "", "Expected no org Email")
	mstub.MockTransactionEnd("t1")

	// get org1 as auditor1 (X)
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	getSolutionUserResult, err = GetSolutionUserWithParams(stub, auditor1Caller, org1.ID, true, true)
	test_utils.AssertTrue(t, err == nil, "Expected GetUserData to succeed")
	test_utils.AssertTrue(t, getSolutionUserResult.ID == org1.ID, "Expected org ID")
	test_utils.AssertTrue(t, getSolutionUserResult.Email == "", "Expected no org Email")
	mstub.MockTransactionEnd("t1")
}

func TestQueryOrgUser(t *testing.T) {
	logger.SetLevel(shim.LogDebug)
	logger.Info("TestQueryOrgUser function called")

	mstub, systemAdmin := querySetup(t)

	// ------- ORG 1 -------

	// register org1
	mstub.MockTransactionStart("t1")
	stub := cached_stub.NewCachedStub(mstub)
	org1 := test_utils.CreateTestGroup("org1")
	org1Bytes, _ := json.Marshal(&org1)
	_, err := RegisterOrg(stub, org1, []string{string(org1Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterOrg to succeed")
	mstub.MockTransactionEnd("t1")

	org1User1Caller := registerOrgUser(t, mstub, org1, "1")
	org1User2Caller := registerOrgUser(t, mstub, org1, "2")

	org1Service1AdminCaller, auditor1Caller := registerServiceAdmin(t, mstub, org1, systemAdmin, "1")
	org1Service2AdminCaller, _ := registerServiceAdmin(t, mstub, org1, systemAdmin, "2")

	// ------- ORG 2 -------

	// register org2
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	org2 := test_utils.CreateTestGroup("org2")
	org2Bytes, _ := json.Marshal(&org2)
	_, err = RegisterOrg(stub, org2, []string{string(org2Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterOrg to succeed")
	mstub.MockTransactionEnd("t1")

	org2UserCaller := registerOrgUser(t, mstub, org2, "1")

	org2Service1AdminCaller, _ := registerServiceAdmin(t, mstub, org2, systemAdmin, "1")

	// ------- PATIENTS -------

	// register patient1
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	patient1 := CreateTestSolutionUser("patient1")
	patient1PlatformUser, err := convertToPlatformUser(stub, patient1)
	test_utils.AssertTrue(t, err == nil, "Expected convertToPlatformUser to succeed")
	patient1Bytes, _ := json.Marshal(&patient1)
	_, err = RegisterUser(stub, patient1PlatformUser, []string{string(patient1Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterUser to succeed")
	mstub.MockTransactionEnd("t1")

	// register patient2
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	patient2 := CreateTestSolutionUser("patient2")
	patient2Caller, err := convertToPlatformUser(stub, patient2)
	test_utils.AssertTrue(t, err == nil, "Expected convertToPlatformUser to succeed")
	patient2Bytes, _ := json.Marshal(&patient2)
	_, err = RegisterUser(stub, patient2Caller, []string{string(patient2Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterUser to succeed")
	mstub.MockTransactionEnd("t1")

	// ------- GET USER -------

	// get org1 user as org1 admin
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	org1Caller, _ := user_mgmt.GetUserData(stub, org1, org1.ID, true, true)
	getSolutionUserResult, err := GetSolutionUserWithParams(stub, org1Caller, org1User1Caller.ID, true, true)
	test_utils.AssertTrue(t, err == nil, "Expected GetUserData to succeed")
	test_utils.AssertTrue(t, getSolutionUserResult.ID == org1User1Caller.ID, "Expected user ID")
	test_utils.AssertTrue(t, getSolutionUserResult.Email != "", "Expected user Email")
	mstub.MockTransactionEnd("t1")

	// get org1 user as org1 user1
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	getSolutionUserResult, err = GetSolutionUserWithParams(stub, org1User1Caller, org1User1Caller.ID, true, true)
	test_utils.AssertTrue(t, err == nil, "Expected GetUserData to succeed")
	test_utils.AssertTrue(t, getSolutionUserResult.ID == org1User1Caller.ID, "Expected user ID")
	test_utils.AssertTrue(t, getSolutionUserResult.Email != "", "Expected user Email")
	mstub.MockTransactionEnd("t1")

	// get org1 user as org1 user2
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	getSolutionUserResult, err = GetSolutionUserWithParams(stub, org1User2Caller, org1User1Caller.ID, true, true)
	test_utils.AssertTrue(t, err == nil, "Expected GetUserData to succeed")
	test_utils.AssertTrue(t, getSolutionUserResult.ID == org1User1Caller.ID, "Expected user ID")
	test_utils.AssertTrue(t, getSolutionUserResult.Email == "", "Expected no user Email")
	mstub.MockTransactionEnd("t1")

	// get org1 user as org1 service1 admin (X)
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	getSolutionUserResult, err = GetSolutionUserWithParams(stub, org1Service1AdminCaller, org1User1Caller.ID, true, true)
	test_utils.AssertTrue(t, err == nil, "Expected GetUserData to succeed")
	test_utils.AssertTrue(t, getSolutionUserResult.ID == org1User1Caller.ID, "Expected user ID")
	test_utils.AssertTrue(t, getSolutionUserResult.Email == "", "Expected no user Email")
	mstub.MockTransactionEnd("t1")

	// get org1 user as org1 service2 admin (X)
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	getSolutionUserResult, err = GetSolutionUserWithParams(stub, org1Service2AdminCaller, org1User1Caller.ID, true, true)
	test_utils.AssertTrue(t, err == nil, "Expected GetUserData to succeed")
	test_utils.AssertTrue(t, getSolutionUserResult.ID == org1User1Caller.ID, "Expected user ID")
	test_utils.AssertTrue(t, getSolutionUserResult.Email == "", "Expected no user Email")
	mstub.MockTransactionEnd("t1")

	// get org1 user as org2 admin (X)
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	org2Caller, _ := user_mgmt.GetUserData(stub, org2, org2.ID, true, true)
	getSolutionUserResult, err = GetSolutionUserWithParams(stub, org2Caller, org1User1Caller.ID, true, true)
	test_utils.AssertTrue(t, err == nil, "Expected GetUserData to succeed")
	test_utils.AssertTrue(t, getSolutionUserResult.ID == org1User1Caller.ID, "Expected user ID")
	test_utils.AssertTrue(t, getSolutionUserResult.Email == "", "Expected no user Email")
	mstub.MockTransactionEnd("t1")

	// get org1 user as org2 user (X)
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	getSolutionUserResult, err = GetSolutionUserWithParams(stub, org2UserCaller, org1User1Caller.ID, true, true)
	test_utils.AssertTrue(t, err == nil, "Expected GetUserData to succeed")
	test_utils.AssertTrue(t, getSolutionUserResult.ID == org1User1Caller.ID, "Expected user ID")
	test_utils.AssertTrue(t, getSolutionUserResult.Email == "", "Expected no user Email")
	mstub.MockTransactionEnd("t1")

	// get org1 user as org2 service admin (X)
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	getSolutionUserResult, err = GetSolutionUserWithParams(stub, org2Service1AdminCaller, org1User1Caller.ID, true, true)
	test_utils.AssertTrue(t, err == nil, "Expected GetUserData to succeed")
	test_utils.AssertTrue(t, getSolutionUserResult.ID == org1User1Caller.ID, "Expected user ID")
	test_utils.AssertTrue(t, getSolutionUserResult.Email == "", "Expected no user Email")
	mstub.MockTransactionEnd("t1")

	// get org1 user as patient1 (X)
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	patient1Caller, _ := user_mgmt.GetUserData(stub, patient1PlatformUser, patient1.ID, true, true)
	getSolutionUserResult, err = GetSolutionUserWithParams(stub, patient1Caller, org1User1Caller.ID, true, true)
	test_utils.AssertTrue(t, err == nil, "Expected GetUserData to succeed")
	test_utils.AssertTrue(t, getSolutionUserResult.ID == org1User1Caller.ID, "Expected user ID")
	test_utils.AssertTrue(t, getSolutionUserResult.Email == "", "Expected no user Email")
	mstub.MockTransactionEnd("t1")

	// get org1 user as patient2 (X)
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	getSolutionUserResult, err = GetSolutionUserWithParams(stub, patient2Caller, org1User1Caller.ID, true, true)
	test_utils.AssertTrue(t, err == nil, "Expected GetUserData to succeed")
	test_utils.AssertTrue(t, getSolutionUserResult.ID == org1User1Caller.ID, "Expected user ID")
	test_utils.AssertTrue(t, getSolutionUserResult.Email == "", "Expected no user Email")
	mstub.MockTransactionEnd("t1")

	// get org1 user as auditor1 (X)
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	getSolutionUserResult, err = GetSolutionUserWithParams(stub, auditor1Caller, org1User1Caller.ID, true, true)
	test_utils.AssertTrue(t, err == nil, "Expected GetUserData to succeed")
	test_utils.AssertTrue(t, getSolutionUserResult.ID == org1User1Caller.ID, "Expected user ID")
	test_utils.AssertTrue(t, getSolutionUserResult.Email == "", "Expected no user Email")
	mstub.MockTransactionEnd("t1")
}

func TestQueryUnenrolledPatient(t *testing.T) {
	logger.SetLevel(shim.LogDebug)
	logger.Info("TestQueryUnenrolledPatient function called")

	mstub, systemAdmin := querySetup(t)

	// ------- ORG 1 -------

	// register org1
	mstub.MockTransactionStart("t1")
	stub := cached_stub.NewCachedStub(mstub)
	org1 := test_utils.CreateTestGroup("org1")
	org1Bytes, _ := json.Marshal(&org1)
	_, err := RegisterOrg(stub, org1, []string{string(org1Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterOrg to succeed")
	mstub.MockTransactionEnd("t1")

	org1UserCaller := registerOrgUser(t, mstub, org1, "1")

	org1Service1AdminCaller, auditor1Caller := registerServiceAdmin(t, mstub, org1, systemAdmin, "1")
	org1Service2AdminCaller, _ := registerServiceAdmin(t, mstub, org1, systemAdmin, "2")

	// ------- ORG 2 -------

	// register org2
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	org2 := test_utils.CreateTestGroup("org2")
	org2Bytes, _ := json.Marshal(&org2)
	_, err = RegisterOrg(stub, org2, []string{string(org2Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterOrg to succeed")
	mstub.MockTransactionEnd("t1")

	org2UserCaller := registerOrgUser(t, mstub, org2, "1")

	org2Service1AdminCaller, _ := registerServiceAdmin(t, mstub, org2, systemAdmin, "1")

	// ------- PATIENTS -------

	// register patient1
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	patient1 := CreateTestSolutionUser("patient1")
	patient1PlatformUser, err := convertToPlatformUser(stub, patient1)
	test_utils.AssertTrue(t, err == nil, "Expected convertToPlatformUser to succeed")
	patient1Bytes, _ := json.Marshal(&patient1)
	_, err = RegisterUser(stub, patient1PlatformUser, []string{string(patient1Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterUser to succeed")
	mstub.MockTransactionEnd("t1")

	// register patient2
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	patient2 := CreateTestSolutionUser("patient2")
	patient2Caller, err := convertToPlatformUser(stub, patient2)
	test_utils.AssertTrue(t, err == nil, "Expected convertToPlatformUser to succeed")
	patient2Bytes, _ := json.Marshal(&patient2)
	_, err = RegisterUser(stub, patient2Caller, []string{string(patient2Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterUser to succeed")
	mstub.MockTransactionEnd("t1")

	// ------- GET UNENROLLED PATIENT -------

	// get patient1 as org1 admin (X)
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	org1Caller, _ := user_mgmt.GetUserData(stub, org1, org1.ID, true, true)
	getSolutionUserResult, err := GetSolutionUserWithParams(stub, org1Caller, patient1.ID, true, true)
	test_utils.AssertTrue(t, err == nil, "Expected GetUserData to succeed")
	test_utils.AssertTrue(t, getSolutionUserResult.ID == patient1.ID, "Expected user ID")
	test_utils.AssertTrue(t, getSolutionUserResult.Email == "", "Expected no user Email")
	mstub.MockTransactionEnd("t1")

	// get patient1 as org1 user (X)
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	getSolutionUserResult, err = GetSolutionUserWithParams(stub, org1UserCaller, patient1.ID, true, true)
	test_utils.AssertTrue(t, err == nil, "Expected GetUserData to succeed")
	test_utils.AssertTrue(t, getSolutionUserResult.ID == patient1.ID, "Expected user ID")
	test_utils.AssertTrue(t, getSolutionUserResult.Email == "", "Expected no user Email")
	mstub.MockTransactionEnd("t1")

	// get patient1 as org1 service1 admin (X)
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	getSolutionUserResult, err = GetSolutionUserWithParams(stub, org1Service1AdminCaller, patient1.ID, true, true)
	test_utils.AssertTrue(t, err == nil, "Expected GetUserData to succeed")
	test_utils.AssertTrue(t, getSolutionUserResult.ID == patient1.ID, "Expected user ID")
	test_utils.AssertTrue(t, getSolutionUserResult.Email == "", "Expected no user Email")
	mstub.MockTransactionEnd("t1")

	// get patient1 as org1 service2 admin (X)
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	getSolutionUserResult, err = GetSolutionUserWithParams(stub, org1Service2AdminCaller, patient1.ID, true, true)
	test_utils.AssertTrue(t, err == nil, "Expected GetUserData to succeed")
	test_utils.AssertTrue(t, getSolutionUserResult.ID == patient1.ID, "Expected user ID")
	test_utils.AssertTrue(t, getSolutionUserResult.Email == "", "Expected no user Email")
	mstub.MockTransactionEnd("t1")

	// get patient1 as org2 admin (X)
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	org2Caller, _ := user_mgmt.GetUserData(stub, org2, org2.ID, true, true)
	getSolutionUserResult, err = GetSolutionUserWithParams(stub, org2Caller, patient1.ID, true, true)
	test_utils.AssertTrue(t, err == nil, "Expected GetUserData to succeed")
	test_utils.AssertTrue(t, getSolutionUserResult.ID == patient1.ID, "Expected user ID")
	test_utils.AssertTrue(t, getSolutionUserResult.Email == "", "Expected no user Email")
	mstub.MockTransactionEnd("t1")

	// get patient1 as org2 user (X)
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	getSolutionUserResult, err = GetSolutionUserWithParams(stub, org2UserCaller, patient1.ID, true, true)
	test_utils.AssertTrue(t, err == nil, "Expected GetUserData to succeed")
	test_utils.AssertTrue(t, getSolutionUserResult.ID == patient1.ID, "Expected user ID")
	test_utils.AssertTrue(t, getSolutionUserResult.Email == "", "Expected no user Email")
	mstub.MockTransactionEnd("t1")

	// get patient1 as org2 service admin (X)
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	getSolutionUserResult, err = GetSolutionUserWithParams(stub, org2Service1AdminCaller, patient1.ID, true, true)
	test_utils.AssertTrue(t, err == nil, "Expected GetUserData to succeed")
	test_utils.AssertTrue(t, getSolutionUserResult.ID == patient1.ID, "Expected user ID")
	test_utils.AssertTrue(t, getSolutionUserResult.Email == "", "Expected no user Email")
	mstub.MockTransactionEnd("t1")

	// get patient1 as patient1
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	patient1Caller, _ := user_mgmt.GetUserData(stub, patient1PlatformUser, patient1.ID, true, true)
	getSolutionUserResult, err = GetSolutionUserWithParams(stub, patient1Caller, patient1.ID, true, true)
	test_utils.AssertTrue(t, err == nil, "Expected GetUserData to succeed")
	test_utils.AssertTrue(t, getSolutionUserResult.ID == patient1.ID, "Expected user ID")
	test_utils.AssertTrue(t, getSolutionUserResult.Email != "", "Expected user Email")
	mstub.MockTransactionEnd("t1")

	// get patient1 as patient2 (X)
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	getSolutionUserResult, err = GetSolutionUserWithParams(stub, patient2Caller, patient1.ID, true, true)
	test_utils.AssertTrue(t, err == nil, "Expected GetUserData to succeed")
	test_utils.AssertTrue(t, getSolutionUserResult.ID == patient1.ID, "Expected user ID")
	test_utils.AssertTrue(t, getSolutionUserResult.Email == "", "Expected no user Email")
	mstub.MockTransactionEnd("t1")

	// get patient1 as auditor1 (X)
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	getSolutionUserResult, err = GetSolutionUserWithParams(stub, auditor1Caller, patient1.ID, true, true)
	test_utils.AssertTrue(t, err == nil, "Expected GetUserData to succeed")
	test_utils.AssertTrue(t, getSolutionUserResult.ID == patient1.ID, "Expected user ID")
	test_utils.AssertTrue(t, getSolutionUserResult.Email == "", "Expected no user Email")
	mstub.MockTransactionEnd("t1")
}

func TestQueryEnrolledPatient(t *testing.T) {
	logger.SetLevel(shim.LogDebug)
	logger.Info("TestQueryEnrolledPatient function called")

	mstub, systemAdmin := querySetup(t)

	// ------- ORG 1 -------

	// register org1
	mstub.MockTransactionStart("t1")
	stub := cached_stub.NewCachedStub(mstub)
	org1 := test_utils.CreateTestGroup("org1")
	org1Bytes, _ := json.Marshal(&org1)
	_, err := RegisterOrg(stub, org1, []string{string(org1Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterOrg to succeed")
	mstub.MockTransactionEnd("t1")

	org1UserCaller := registerOrgUser(t, mstub, org1, "1")

	org1Service1AdminCaller, auditor1Caller := registerServiceAdmin(t, mstub, org1, systemAdmin, "1")
	org1Service2AdminCaller, _ := registerServiceAdmin(t, mstub, org1, systemAdmin, "2")

	// ------- ORG 2 -------

	// register org2
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	org2 := test_utils.CreateTestGroup("org2")
	org2Bytes, _ := json.Marshal(&org2)
	_, err = RegisterOrg(stub, org2, []string{string(org2Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterOrg to succeed")
	mstub.MockTransactionEnd("t1")

	org2UserCaller := registerOrgUser(t, mstub, org2, "1")

	org2Service1AdminCaller, _ := registerServiceAdmin(t, mstub, org2, systemAdmin, "1")

	// ------- PATIENTS -------

	// register patient1
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	patient1 := CreateTestSolutionUser("patient1")
	patient1PlatformUser, err := convertToPlatformUser(stub, patient1)
	test_utils.AssertTrue(t, err == nil, "Expected convertToPlatformUser to succeed")
	patient1Bytes, _ := json.Marshal(&patient1)
	_, err = RegisterUser(stub, patient1PlatformUser, []string{string(patient1Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterUser to succeed")
	mstub.MockTransactionEnd("t1")

	// register patient2
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	patient2 := CreateTestSolutionUser("patient2")
	patient2Caller, err := convertToPlatformUser(stub, patient2)
	test_utils.AssertTrue(t, err == nil, "Expected convertToPlatformUser to succeed")
	patient2Bytes, _ := json.Marshal(&patient2)
	_, err = RegisterUser(stub, patient2Caller, []string{string(patient2Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterUser to succeed")
	mstub.MockTransactionEnd("t1")

	// ------- ENROLLMENTS -------

	// enroll patient1 to org1 service1
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	enrollment1 := GenerateEnrollmentTest(patient1.ID, "org1Service1")
	enrollment1Bytes, _ := json.Marshal(&enrollment1)
	enrollmentKey1 := test_utils.GenerateSymKey()
	enrollmentKey1B64 := crypto.EncodeToB64String(enrollmentKey1)
	_, err = EnrollPatient(stub, org1Service1AdminCaller, []string{string(enrollment1Bytes), enrollmentKey1B64})
	test_utils.AssertTrue(t, err == nil, "Expected EnrollPatient to succeed")
	mstub.MockTransactionEnd("t1")

	// ------- GET ENROLLED PATIENT -------

	// get patient1 as org1 admin (X)
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	org1Caller, _ := user_mgmt.GetUserData(stub, org1, org1.ID, true, true)
	getSolutionUserResult, err := GetSolutionUserWithParams(stub, org1Caller, patient1.ID, true, true)
	test_utils.AssertTrue(t, err == nil, "Expected GetUserData to succeed")
	test_utils.AssertTrue(t, getSolutionUserResult.ID == patient1.ID, "Expected user ID")
	test_utils.AssertTrue(t, getSolutionUserResult.Email == "", "Expected no user Email")
	mstub.MockTransactionEnd("t1")

	// get patient1 as org1 user (X)
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	getSolutionUserResult, err = GetSolutionUserWithParams(stub, org1UserCaller, patient1.ID, true, true)
	test_utils.AssertTrue(t, err == nil, "Expected GetUserData to succeed")
	test_utils.AssertTrue(t, getSolutionUserResult.ID == patient1.ID, "Expected user ID")
	test_utils.AssertTrue(t, getSolutionUserResult.Email == "", "Expected no user Email")
	mstub.MockTransactionEnd("t1")

	// get patient1 as org1 service1 admin (X)
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	getSolutionUserResult, err = GetSolutionUserWithParams(stub, org1Service1AdminCaller, patient1.ID, true, true)
	test_utils.AssertTrue(t, err == nil, "Expected GetUserData to succeed")
	test_utils.AssertTrue(t, getSolutionUserResult.ID == patient1.ID, "Expected user ID")
	test_utils.AssertTrue(t, getSolutionUserResult.Email == "", "Expected no user Email")
	mstub.MockTransactionEnd("t1")

	// get patient1 as org1 service2 admin (X)
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	getSolutionUserResult, err = GetSolutionUserWithParams(stub, org1Service2AdminCaller, patient1.ID, true, true)
	test_utils.AssertTrue(t, err == nil, "Expected GetUserData to succeed")
	test_utils.AssertTrue(t, getSolutionUserResult.ID == patient1.ID, "Expected user ID")
	test_utils.AssertTrue(t, getSolutionUserResult.Email == "", "Expected no user Email")
	mstub.MockTransactionEnd("t1")

	// get patient1 as org2 admin (X)
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	org2Caller, _ := user_mgmt.GetUserData(stub, org2, org2.ID, true, true)
	getSolutionUserResult, err = GetSolutionUserWithParams(stub, org2Caller, patient1.ID, true, true)
	test_utils.AssertTrue(t, err == nil, "Expected GetUserData to succeed")
	test_utils.AssertTrue(t, getSolutionUserResult.ID == patient1.ID, "Expected user ID")
	test_utils.AssertTrue(t, getSolutionUserResult.Email == "", "Expected no user Email")
	mstub.MockTransactionEnd("t1")

	// get patient1 as org2 user (X)
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	getSolutionUserResult, err = GetSolutionUserWithParams(stub, org2UserCaller, patient1.ID, true, true)
	test_utils.AssertTrue(t, err == nil, "Expected GetUserData to succeed")
	test_utils.AssertTrue(t, getSolutionUserResult.ID == patient1.ID, "Expected user ID")
	test_utils.AssertTrue(t, getSolutionUserResult.Email == "", "Expected no user Email")
	mstub.MockTransactionEnd("t1")

	// get patient1 as org2 service admin (X)
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	getSolutionUserResult, err = GetSolutionUserWithParams(stub, org2Service1AdminCaller, patient1.ID, true, true)
	test_utils.AssertTrue(t, err == nil, "Expected GetUserData to succeed")
	test_utils.AssertTrue(t, getSolutionUserResult.ID == patient1.ID, "Expected user ID")
	test_utils.AssertTrue(t, getSolutionUserResult.Email == "", "Expected no user Email")
	mstub.MockTransactionEnd("t1")

	// get patient1 as patient1
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	patient1Caller, _ := user_mgmt.GetUserData(stub, patient1PlatformUser, patient1.ID, true, true)
	getSolutionUserResult, err = GetSolutionUserWithParams(stub, patient1Caller, patient1.ID, true, true)
	test_utils.AssertTrue(t, err == nil, "Expected GetUserData to succeed")
	test_utils.AssertTrue(t, getSolutionUserResult.ID == patient1.ID, "Expected user ID")
	test_utils.AssertTrue(t, getSolutionUserResult.Email != "", "Expected user Email")
	mstub.MockTransactionEnd("t1")

	// get patient1 as patient2 (X)
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	getSolutionUserResult, err = GetSolutionUserWithParams(stub, patient2Caller, patient1.ID, true, true)
	test_utils.AssertTrue(t, err == nil, "Expected GetUserData to succeed")
	test_utils.AssertTrue(t, getSolutionUserResult.ID == patient1.ID, "Expected user ID")
	test_utils.AssertTrue(t, getSolutionUserResult.Email == "", "Expected no user Email")
	mstub.MockTransactionEnd("t1")

	// get patient1 as auditor1 (X)
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	getSolutionUserResult, err = GetSolutionUserWithParams(stub, auditor1Caller, patient1.ID, true, true)
	test_utils.AssertTrue(t, err == nil, "Expected GetUserData to succeed")
	test_utils.AssertTrue(t, getSolutionUserResult.ID == patient1.ID, "Expected user ID")
	test_utils.AssertTrue(t, getSolutionUserResult.Email == "", "Expected no user Email")
	mstub.MockTransactionEnd("t1")
}

func TestQueryService(t *testing.T) {
	logger.SetLevel(shim.LogDebug)
	logger.Info("TestQueryService function called")

	mstub, systemAdmin := querySetup(t)

	// ------- ORG 1 -------

	// register org1
	mstub.MockTransactionStart("t1")
	stub := cached_stub.NewCachedStub(mstub)
	org1 := test_utils.CreateTestGroup("org1")
	org1Bytes, _ := json.Marshal(&org1)
	_, err := RegisterOrg(stub, org1, []string{string(org1Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterOrg to succeed")
	mstub.MockTransactionEnd("t1")

	org1UserCaller := registerOrgUser(t, mstub, org1, "1")

	org1Service1AdminCaller, auditor1Caller := registerServiceAdmin(t, mstub, org1, systemAdmin, "1")
	org1Service2AdminCaller, _ := registerServiceAdmin(t, mstub, org1, systemAdmin, "2")

	// ------- ORG 2 -------

	// register org2
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	org2 := test_utils.CreateTestGroup("org2")
	org2Bytes, _ := json.Marshal(&org2)
	_, err = RegisterOrg(stub, org2, []string{string(org2Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterOrg to succeed")
	mstub.MockTransactionEnd("t1")

	org2UserCaller := registerOrgUser(t, mstub, org2, "1")

	org2Service1AdminCaller, _ := registerServiceAdmin(t, mstub, org2, systemAdmin, "1")

	// ------- PATIENTS -------

	// register patient1
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	patient1 := CreateTestSolutionUser("patient1")
	patient1PlatformUser, err := convertToPlatformUser(stub, patient1)
	test_utils.AssertTrue(t, err == nil, "Expected convertToPlatformUser to succeed")
	patient1Bytes, _ := json.Marshal(&patient1)
	_, err = RegisterUser(stub, patient1PlatformUser, []string{string(patient1Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterUser to succeed")
	mstub.MockTransactionEnd("t1")

	// register patient2
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	patient2 := CreateTestSolutionUser("patient2")
	patient2PlatformUser, err := convertToPlatformUser(stub, patient2)
	test_utils.AssertTrue(t, err == nil, "Expected convertToPlatformUser to succeed")
	patient2Bytes, _ := json.Marshal(&patient2)
	_, err = RegisterUser(stub, patient2PlatformUser, []string{string(patient2Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterUser to succeed")
	mstub.MockTransactionEnd("t1")

	// ------- ENROLLMENTS -------

	// enroll patient1 to org1 service1
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	enrollment1 := GenerateEnrollmentTest(patient1.ID, "org1Service1")
	enrollment1Bytes, _ := json.Marshal(&enrollment1)
	enrollmentKey1 := test_utils.GenerateSymKey()
	enrollmentKey1B64 := crypto.EncodeToB64String(enrollmentKey1)
	_, err = EnrollPatient(stub, org1Service1AdminCaller, []string{string(enrollment1Bytes), enrollmentKey1B64})
	test_utils.AssertTrue(t, err == nil, "Expected EnrollPatient to succeed")
	mstub.MockTransactionEnd("t1")

	// ------- GET SERVICE -------

	// get org1 service as org1 admin
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	org1Caller, _ := user_mgmt.GetUserData(stub, org1, org1.ID, true, true)
	service, err := GetServiceInternal(stub, org1Caller, "org1Service1", true)
	test_utils.AssertTrue(t, err == nil, "GetServiceInternal succcessful")
	test_utils.AssertTrue(t, service.ServiceID != "", "Expected service ServiceID")
	test_utils.AssertTrue(t, service.SolutionPrivateData != nil, "Expected service SolutionPrivateData")
	mstub.MockTransactionEnd("t1")

	// get org1 service as org1 user (X)
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	service, err = GetServiceInternal(stub, org1UserCaller, "org1Service1", true)
	test_utils.AssertTrue(t, err == nil, "GetServiceInternal succcessful")
	test_utils.AssertTrue(t, service.ServiceID != "", "Expected service ServiceID")
	test_utils.AssertTrue(t, service.SolutionPrivateData == nil, "Expected no service SolutionPrivateData")
	mstub.MockTransactionEnd("t1")

	// get org1 service as org1 service1 admin
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	service, err = GetServiceInternal(stub, org1Service1AdminCaller, "org1Service1", true)
	test_utils.AssertTrue(t, err == nil, "GetServiceInternal succcessful")
	test_utils.AssertTrue(t, service.ServiceID != "", "Expected service ServiceID")
	test_utils.AssertTrue(t, service.SolutionPrivateData != nil, "Expected service SolutionPrivateData")
	mstub.MockTransactionEnd("t1")

	// get org1 service as org1 service2 admin (X)
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	service, err = GetServiceInternal(stub, org1Service2AdminCaller, "org1Service1", true)
	test_utils.AssertTrue(t, err == nil, "GetServiceInternal succcessful")
	test_utils.AssertTrue(t, service.ServiceID != "", "Expected service ServiceID")
	test_utils.AssertTrue(t, service.SolutionPrivateData == nil, "Expected no service SolutionPrivateData")
	mstub.MockTransactionEnd("t1")

	// get org1 service as org2 admin (X)
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	org2Caller, _ := user_mgmt.GetUserData(stub, org2, org2.ID, true, true)
	service, err = GetServiceInternal(stub, org2Caller, "org1Service1", true)
	test_utils.AssertTrue(t, err == nil, "GetServiceInternal succcessful")
	test_utils.AssertTrue(t, service.ServiceID != "", "Expected service ServiceID")
	test_utils.AssertTrue(t, service.SolutionPrivateData == nil, "Expected no service SolutionPrivateData")
	mstub.MockTransactionEnd("t1")

	// get org1 service as org2 user (X)
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	service, err = GetServiceInternal(stub, org2UserCaller, "org1Service1", true)
	test_utils.AssertTrue(t, err == nil, "GetServiceInternal succcessful")
	test_utils.AssertTrue(t, service.ServiceID != "", "Expected service ServiceID")
	test_utils.AssertTrue(t, service.SolutionPrivateData == nil, "Expected no service SolutionPrivateData")
	mstub.MockTransactionEnd("t1")

	// get org1 service as org2 service admin (X)
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	service, err = GetServiceInternal(stub, org2Service1AdminCaller, "org1Service1", true)
	test_utils.AssertTrue(t, err == nil, "GetServiceInternal succcessful")
	test_utils.AssertTrue(t, service.ServiceID != "", "Expected service ServiceID")
	test_utils.AssertTrue(t, service.SolutionPrivateData == nil, "Expected no service SolutionPrivateData")
	mstub.MockTransactionEnd("t1")

	// get org1 service as patient1, enrolled (X)
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	patient1Caller, _ := user_mgmt.GetUserData(stub, patient1PlatformUser, patient1.ID, true, true)
	service, err = GetServiceInternal(stub, patient1Caller, "org1Service1", true)
	test_utils.AssertTrue(t, err == nil, "GetServiceInternal succcessful")
	test_utils.AssertTrue(t, service.ServiceID != "", "Expected service ServiceID")
	test_utils.AssertTrue(t, service.SolutionPrivateData == nil, "Expected service SolutionPrivateData")
	mstub.MockTransactionEnd("t1")

	// get org1 service as patient2 (X)
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	patient2Caller, _ := user_mgmt.GetUserData(stub, patient2PlatformUser, patient2.ID, true, true)
	service, err = GetServiceInternal(stub, patient2Caller, "org1Service1", true)
	test_utils.AssertTrue(t, err == nil, "GetServiceInternal succcessful")
	test_utils.AssertTrue(t, service.ServiceID != "", "Expected service ServiceID")
	test_utils.AssertTrue(t, service.SolutionPrivateData == nil, "Expected no service SolutionPrivateData")
	mstub.MockTransactionEnd("t1")

	// get org1 service as auditor1 (X)
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	service, err = GetServiceInternal(stub, auditor1Caller, "org1Service1", true)
	test_utils.AssertTrue(t, err == nil, "GetServiceInternal succcessful")
	test_utils.AssertTrue(t, service.ServiceID != "", "Expected service ServiceID")
	test_utils.AssertTrue(t, service.SolutionPrivateData == nil, "Expected no service SolutionPrivateData")
	mstub.MockTransactionEnd("t1")
}

func TestQueryEnrollment(t *testing.T) {
	logger.SetLevel(shim.LogDebug)
	logger.Info("TestQueryEnrollment function called")

	mstub, systemAdmin := querySetup(t)

	// ------- ORG 1 -------

	// register org1
	mstub.MockTransactionStart("t1")
	stub := cached_stub.NewCachedStub(mstub)
	org1 := test_utils.CreateTestGroup("org1")
	org1Bytes, _ := json.Marshal(&org1)
	_, err := RegisterOrg(stub, org1, []string{string(org1Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterOrg to succeed")
	mstub.MockTransactionEnd("t1")

	org1UserCaller := registerOrgUser(t, mstub, org1, "1")

	org1Service1AdminCaller, auditor1Caller := registerServiceAdmin(t, mstub, org1, systemAdmin, "1")
	org1Service2AdminCaller, _ := registerServiceAdmin(t, mstub, org1, systemAdmin, "2")

	// ------- ORG 2 -------

	// register org2
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	org2 := test_utils.CreateTestGroup("org2")
	org2Bytes, _ := json.Marshal(&org2)
	_, err = RegisterOrg(stub, org2, []string{string(org2Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterOrg to succeed")
	mstub.MockTransactionEnd("t1")

	org2UserCaller := registerOrgUser(t, mstub, org2, "1")

	org2Service1AdminCaller, _ := registerServiceAdmin(t, mstub, org2, systemAdmin, "1")

	// ------- PATIENTS -------

	// register patient1
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	patient1 := CreateTestSolutionUser("patient1")
	patient1PlatformUser, err := convertToPlatformUser(stub, patient1)
	test_utils.AssertTrue(t, err == nil, "Expected convertToPlatformUser to succeed")
	patient1Bytes, _ := json.Marshal(&patient1)
	_, err = RegisterUser(stub, patient1PlatformUser, []string{string(patient1Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterUser to succeed")
	mstub.MockTransactionEnd("t1")

	// register patient2
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	patient2 := CreateTestSolutionUser("patient2")
	patient2Caller, err := convertToPlatformUser(stub, patient2)
	test_utils.AssertTrue(t, err == nil, "Expected convertToPlatformUser to succeed")
	patient2Bytes, _ := json.Marshal(&patient2)
	_, err = RegisterUser(stub, patient2Caller, []string{string(patient2Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterUser to succeed")
	mstub.MockTransactionEnd("t1")

	// ------- ENROLLMENT -------

	// enroll patient1 to org1 service1
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	enrollment1 := GenerateEnrollmentTest(patient1.ID, "org1Service1")
	enrollment1Bytes, _ := json.Marshal(&enrollment1)
	enrollmentKey1 := test_utils.GenerateSymKey()
	enrollmentKey1B64 := crypto.EncodeToB64String(enrollmentKey1)
	_, err = EnrollPatient(stub, org1Service1AdminCaller, []string{string(enrollment1Bytes), enrollmentKey1B64})
	test_utils.AssertTrue(t, err == nil, "Expected EnrollPatient to succeed")
	mstub.MockTransactionEnd("t1")

	// ------- GET ENROLLMENT -------

	// get enrollment as org1 admin
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	org1Caller, _ := user_mgmt.GetUserData(stub, org1, org1.ID, true, true)
	enrollmentResult, err := GetEnrollmentInternal(stub, org1Caller, "patient1", "org1Service1", "org1")
	test_utils.AssertTrue(t, err == nil, "Expected GetEnrollmentInternal to succeed")
	test_utils.AssertTrue(t, enrollmentResult.ServiceName != "", "Expected enrollment ServiceName")
	test_utils.AssertTrue(t, enrollmentResult.Status == "active", "Expected enrollment status to be active")
	mstub.MockTransactionEnd("t1")

	// get enrollment as org1 user (X)
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	enrollmentResult, err = GetEnrollmentInternal(stub, org1UserCaller, "patient1", "org1Service1", "org1")
	test_utils.AssertTrue(t, err != nil, "Expected GetEnrollmentInternal to fail")
	mstub.MockTransactionEnd("t1")

	// get enrollment as org1 service1 admin
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	enrollmentResult, err = GetEnrollmentInternal(stub, org1Service1AdminCaller, "patient1", "org1Service1", "org1")
	test_utils.AssertTrue(t, err == nil, "Expected GetEnrollmentInternal to succeed")
	test_utils.AssertTrue(t, enrollmentResult.ServiceName != "", "Expected enrollment ServiceName")
	test_utils.AssertTrue(t, enrollmentResult.Status == "active", "Expected enrollment status to be active")
	mstub.MockTransactionEnd("t1")

	// get enrollment as org1 service2 admin (X)
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	enrollmentResult, err = GetEnrollmentInternal(stub, org1Service2AdminCaller, "patient1", "org1Service1", "org1")
	test_utils.AssertTrue(t, err != nil, "Expected GetEnrollmentInternal to fail")
	mstub.MockTransactionEnd("t1")

	// get enrollment as org2 admin (X)
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	org2Caller, _ := user_mgmt.GetUserData(stub, org2, org2.ID, true, true)
	enrollmentResult, err = GetEnrollmentInternal(stub, org2Caller, "patient1", "org1Service1", "org1")
	test_utils.AssertTrue(t, err != nil, "Expected GetEnrollmentInternal to fail")
	mstub.MockTransactionEnd("t1")

	// get enrollment as org2 user (X)
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	enrollmentResult, err = GetEnrollmentInternal(stub, org2UserCaller, "patient1", "org1Service1", "org1")
	test_utils.AssertTrue(t, err != nil, "Expected GetEnrollmentInternal to fail")
	mstub.MockTransactionEnd("t1")

	// get enrollment as org2 service admin (X)
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	enrollmentResult, err = GetEnrollmentInternal(stub, org2Service1AdminCaller, "patient1", "org1Service1", "org1")
	test_utils.AssertTrue(t, err != nil, "Expected GetEnrollmentInternal to fail")
	mstub.MockTransactionEnd("t1")

	// get enrollment as patient1
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	patient1Caller, _ := user_mgmt.GetUserData(stub, patient1PlatformUser, patient1.ID, true, true)
	enrollmentResult, err = GetEnrollmentInternal(stub, patient1Caller, "patient1", "org1Service1", "org1")
	test_utils.AssertTrue(t, err == nil, "Expected GetEnrollmentInternal to succeed")
	test_utils.AssertTrue(t, enrollmentResult.ServiceName != "", "Expected enrollment ServiceName")
	test_utils.AssertTrue(t, enrollmentResult.Status == "active", "Expected enrollment status to be active")
	mstub.MockTransactionEnd("t1")

	// get enrollment as patient2 (X)
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	enrollmentResult, err = GetEnrollmentInternal(stub, patient2Caller, "patient1", "org1Service1", "org1")
	test_utils.AssertTrue(t, err != nil, "Expected GetEnrollmentInternal to fail")
	mstub.MockTransactionEnd("t1")

	// get enrollment as auditor1 (X)
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	enrollmentResult, err = GetEnrollmentInternal(stub, auditor1Caller, "patient1", "org1Service1", "org1")
	test_utils.AssertTrue(t, err != nil, "Expected GetEnrollmentInternal to fail")
	mstub.MockTransactionEnd("t1")
}

func TestQueryConsent(t *testing.T) {
	logger.SetLevel(shim.LogDebug)
	logger.Info("TestQueryConsent function called")

	mstub, systemAdmin := querySetup(t)

	// ------- ORG 1 -------

	// register org1
	mstub.MockTransactionStart("t1")
	stub := cached_stub.NewCachedStub(mstub)
	org1 := test_utils.CreateTestGroup("org1")
	org1Bytes, _ := json.Marshal(&org1)
	_, err := RegisterOrg(stub, org1, []string{string(org1Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterOrg to succeed")
	mstub.MockTransactionEnd("t1")

	org1UserCaller := registerOrgUser(t, mstub, org1, "1")

	org1Service1AdminCaller, auditor1Caller := registerServiceAdmin(t, mstub, org1, systemAdmin, "1")
	org1Service2AdminCaller, _ := registerServiceAdmin(t, mstub, org1, systemAdmin, "2")

	// ------- ORG 2 -------

	// register org2
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	org2 := test_utils.CreateTestGroup("org2")
	org2Bytes, _ := json.Marshal(&org2)
	_, err = RegisterOrg(stub, org2, []string{string(org2Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterOrg to succeed")
	mstub.MockTransactionEnd("t1")

	org2UserCaller := registerOrgUser(t, mstub, org2, "1")

	org2Service1AdminCaller, _ := registerServiceAdmin(t, mstub, org2, systemAdmin, "1")

	// ------- PATIENTS -------

	// register patient1
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	patient1 := CreateTestSolutionUser("patient1")
	patient1PlatformUser, err := convertToPlatformUser(stub, patient1)
	test_utils.AssertTrue(t, err == nil, "Expected convertToPlatformUser to succeed")
	patient1Bytes, _ := json.Marshal(&patient1)
	_, err = RegisterUser(stub, patient1PlatformUser, []string{string(patient1Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterUser to succeed")
	mstub.MockTransactionEnd("t1")

	// register patient2
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	patient2 := CreateTestSolutionUser("patient2")
	patient2Caller, err := convertToPlatformUser(stub, patient2)
	test_utils.AssertTrue(t, err == nil, "Expected convertToPlatformUser to succeed")
	patient2Bytes, _ := json.Marshal(&patient2)
	_, err = RegisterUser(stub, patient2Caller, []string{string(patient2Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterUser to succeed")
	mstub.MockTransactionEnd("t1")

	// ------- ENROLLMENT -------

	// enroll patient1 to org1 service1
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	enrollment1 := GenerateEnrollmentTest(patient1.ID, "org1Service1")
	enrollment1Bytes, _ := json.Marshal(&enrollment1)
	enrollmentKey1 := test_utils.GenerateSymKey()
	enrollmentKey1B64 := crypto.EncodeToB64String(enrollmentKey1)
	_, err = EnrollPatient(stub, org1Service1AdminCaller, []string{string(enrollment1Bytes), enrollmentKey1B64})
	test_utils.AssertTrue(t, err == nil, "Expected EnrollPatient to succeed")
	mstub.MockTransactionEnd("t1")

	// ------- CONSENT -------

	// patient1 gives consent
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub, true, true, true)
	consent := Consent{}
	consent.Owner = "patient1"
	consent.Service = "org1Service1"
	consent.Target = "org1Service1"
	consent.Datatype = "org1Datatype1"
	consent.Option = []string{consentOptionWrite, consentOptionRead}
	consent.Timestamp = time.Now().Unix()
	consent.Expiration = 0
	consentBytes, _ := json.Marshal(&consent)
	consentKey := test_utils.GenerateSymKey()
	consentKeyB64 := crypto.EncodeToB64String(consentKey)
	patient1Caller, _ := user_mgmt.GetUserData(stub, patient1PlatformUser, patient1.ID, true, true)
	_, err = PutConsentPatientData(stub, patient1Caller, []string{string(consentBytes), consentKeyB64})
	test_utils.AssertTrue(t, err == nil, "Expected PutConsentPatientData to succeed")
	mstub.MockTransactionEnd("t1")

	// ------- GET CONSENT -------

	// get consent as org1 admin
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	org1Caller, _ := user_mgmt.GetUserData(stub, org1, org1.ID, true, true)
	consentResult, err := GetConsentInternal(stub, org1Caller, "org1Service1", "org1Datatype1", "patient1")
	test_utils.AssertTrue(t, err == nil, "Expected GetConsentInternal to succeed")
	test_utils.AssertTrue(t, consentResult.Owner == "patient1", "Got consent owner correctly")
	test_utils.AssertTrue(t, consentResult.Owner != "", "Got consent Owner")
	test_utils.AssertTrue(t, consentResult.Timestamp > 0, "Expected consent Timestamp")
	mstub.MockTransactionEnd("t1")

	// get consent as org1 user (X)
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	consentResult, err = GetConsentInternal(stub, org1UserCaller, "org1Service1", "org1Datatype1", "patient1")
	test_utils.AssertTrue(t, err != nil, "Expected GetConsentInternal to fail")
	mstub.MockTransactionEnd("t1")

	// get consent as org1 service1 admin
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	consentResult, err = GetConsentInternal(stub, org1Service1AdminCaller, "org1Service1", "org1Datatype1", "patient1")
	test_utils.AssertTrue(t, err == nil, "Expected GetConsentInternal to succeed")
	test_utils.AssertTrue(t, consentResult.Owner == "patient1", "Got consent owner correctly")
	test_utils.AssertTrue(t, consentResult.Owner != "", "Got consent Owner")
	test_utils.AssertTrue(t, consentResult.Timestamp > 0, "Expected consent Timestamp")
	mstub.MockTransactionEnd("t1")

	// get consent as org1 service2 admin (X)
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	consentResult, err = GetConsentInternal(stub, org1Service2AdminCaller, "org1Service1", "org1Datatype1", "patient1")
	test_utils.AssertTrue(t, err != nil, "Expected GetConsentInternal to fail")
	mstub.MockTransactionEnd("t1")

	// get consent as org2 admin (X)
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	org2Caller, _ := user_mgmt.GetUserData(stub, org2, org2.ID, true, true)
	consentResult, err = GetConsentInternal(stub, org2Caller, "org1Service1", "org1Datatype1", "patient1")
	test_utils.AssertTrue(t, err != nil, "Expected GetConsentInternal to fail")
	mstub.MockTransactionEnd("t1")

	// get consent as org2 user (X)
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	consentResult, err = GetConsentInternal(stub, org2UserCaller, "org1Service1", "org1Datatype1", "patient1")
	test_utils.AssertTrue(t, err != nil, "Expected GetConsentInternal to fail")
	mstub.MockTransactionEnd("t1")

	// get consent as org2 service admin (X)
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	consentResult, err = GetConsentInternal(stub, org2Service1AdminCaller, "org1Service1", "org1Datatype1", "patient1")
	test_utils.AssertTrue(t, err != nil, "Expected GetConsentInternal to fail")
	mstub.MockTransactionEnd("t1")

	// get consent as patient1
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	consentResult, err = GetConsentInternal(stub, patient1Caller, "org1Service1", "org1Datatype1", "patient1")
	test_utils.AssertTrue(t, err == nil, "Expected GetConsentInternal to succeed")
	test_utils.AssertTrue(t, consentResult.Owner == "patient1", "Got consent owner correctly")
	test_utils.AssertTrue(t, consentResult.Owner != "", "Got consent Owner")
	test_utils.AssertTrue(t, consentResult.Timestamp > 0, "Expected consent Timestamp")
	mstub.MockTransactionEnd("t1")

	// get consent as patient2 (X)
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	consentResult, err = GetConsentInternal(stub, patient2Caller, "org1Service1", "org1Datatype1", "patient1")
	test_utils.AssertTrue(t, err != nil, "Expected GetConsentInternal to fail")
	mstub.MockTransactionEnd("t1")

	// get consent as auditor1 (X)
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	consentResult, err = GetConsentInternal(stub, auditor1Caller, "org1Service1", "org1Datatype1", "patient1")
	test_utils.AssertTrue(t, err != nil, "Expected GetConsentInternal to fail")
	mstub.MockTransactionEnd("t1")
}

func TestQueryValidateConsent(t *testing.T) {
	logger.SetLevel(shim.LogDebug)
	logger.Info("TestQueryValidateConsent function called")

	mstub, systemAdmin := querySetup(t)

	// ------- ORG 1 -------

	// register org1
	mstub.MockTransactionStart("t1")
	stub := cached_stub.NewCachedStub(mstub)
	org1 := test_utils.CreateTestGroup("org1")
	org1Bytes, _ := json.Marshal(&org1)
	_, err := RegisterOrg(stub, org1, []string{string(org1Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterOrg to succeed")
	mstub.MockTransactionEnd("t1")

	org1UserCaller := registerOrgUser(t, mstub, org1, "1")

	org1Service1AdminCaller, auditor1Caller := registerServiceAdmin(t, mstub, org1, systemAdmin, "1")
	org1Service2AdminCaller, _ := registerServiceAdmin(t, mstub, org1, systemAdmin, "2")

	// ------- ORG 2 -------

	// register org2
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	org2 := test_utils.CreateTestGroup("org2")
	org2Bytes, _ := json.Marshal(&org2)
	_, err = RegisterOrg(stub, org2, []string{string(org2Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterOrg to succeed")
	mstub.MockTransactionEnd("t1")

	org2UserCaller := registerOrgUser(t, mstub, org2, "1")

	org2Service1AdminCaller, _ := registerServiceAdmin(t, mstub, org2, systemAdmin, "1")

	// ------- PATIENTS -------

	// register patient1
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	patient1 := CreateTestSolutionUser("patient1")
	patient1PlatformUser, err := convertToPlatformUser(stub, patient1)
	test_utils.AssertTrue(t, err == nil, "Expected convertToPlatformUser to succeed")
	patient1Bytes, _ := json.Marshal(&patient1)
	_, err = RegisterUser(stub, patient1PlatformUser, []string{string(patient1Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterUser to succeed")
	mstub.MockTransactionEnd("t1")

	// ------- ENROLLMENT -------

	// enroll patient1 to org1 service1
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	enrollment1 := GenerateEnrollmentTest(patient1.ID, "org1Service1")
	enrollment1Bytes, _ := json.Marshal(&enrollment1)
	enrollmentKey1 := test_utils.GenerateSymKey()
	enrollmentKey1B64 := crypto.EncodeToB64String(enrollmentKey1)
	_, err = EnrollPatient(stub, org1Service1AdminCaller, []string{string(enrollment1Bytes), enrollmentKey1B64})
	test_utils.AssertTrue(t, err == nil, "Expected EnrollPatient to succeed")
	mstub.MockTransactionEnd("t1")

	// ------- CONSENT -------

	// patient1 gives consent
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub, true, true, true)
	consent := Consent{}
	consent.Owner = "patient1"
	consent.Service = "org1Service1"
	consent.Target = "org1Service1"
	consent.Datatype = "org1Datatype1"
	consent.Option = []string{consentOptionWrite, consentOptionRead}
	consent.Timestamp = time.Now().Unix()
	consent.Expiration = 0
	consentBytes, _ := json.Marshal(&consent)
	consentKey := test_utils.GenerateSymKey()
	consentKeyB64 := crypto.EncodeToB64String(consentKey)
	patient1Caller, _ := user_mgmt.GetUserData(stub, patient1PlatformUser, patient1.ID, true, true)
	_, err = PutConsentPatientData(stub, patient1Caller, []string{string(consentBytes), consentKeyB64})
	test_utils.AssertTrue(t, err == nil, "Expected PutConsentPatientData to succeed")
	mstub.MockTransactionEnd("t1")

	// ------- VALIDATE CONSENT -------

	// validate consent as org1 admin
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	org1Caller, _ := user_mgmt.GetUserData(stub, org1, org1.ID, true, true)
	cvResultBytes, err := ValidateConsent(stub, org1Caller, []string{"patient1", "org1Service1", "org1Datatype1", consentOptionWrite, strconv.FormatInt(time.Now().Unix(), 10)})
	test_utils.AssertTrue(t, err == nil, "Expected ValidateConsent to succeed")
	test_utils.AssertTrue(t, cvResultBytes != nil, "Expected ValidateConsent to succeed")
	cvResult := ValidationResultWithLog{}
	json.Unmarshal(cvResultBytes, &cvResult)
	consentToken, _ := url.QueryUnescape(cvResult.ConsentValidation.Token)
	token, err := DecryptConsentValidationToken(stub, org1, consentToken)
	test_utils.AssertTrue(t, err == nil, "Expected to get token successfully")
	test_utils.AssertTrue(t, token.ConsentKey != nil, "Expected to get token successfully")
	test_utils.AssertTrue(t, cvResult.ConsentValidation.PermissionGranted == true, "Got consent validation correctly")
	test_utils.AssertTrue(t, cvResult.ConsentValidation.Requester != "", "Expected Requester")
	test_utils.AssertTrue(t, cvResult.ConsentValidation.Owner != "", "Got Owner")
	test_utils.AssertTrue(t, cvResult.ConsentValidation.Target != "", "Got Target")
	mstub.MockTransactionEnd("t1")

	// validate consent as org1 user
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	cvResultBytes, err = ValidateConsent(stub, org1UserCaller, []string{"patient1", "org1Service1", "org1Datatype1", consentOptionWrite, strconv.FormatInt(time.Now().Unix(), 10)})
	test_utils.AssertTrue(t, err == nil, "Expected ValidateConsent to succeed")
	test_utils.AssertTrue(t, cvResultBytes != nil, "Expected ValidateConsent to succeed")
	cvResult = ValidationResultWithLog{}
	json.Unmarshal(cvResultBytes, &cvResult)
	test_utils.AssertTrue(t, cvResult.ConsentValidation.PermissionGranted == false, "Expected no PermissionGranted")
	fmt.Printf("CONSENT VALID: %+v\n", cvResult.ConsentValidation)
	test_utils.AssertTrue(t, cvResult.ConsentValidation.Requester != "", "Expected Requester")
	test_utils.AssertTrue(t, cvResult.ConsentValidation.Owner != "", "Got Owner")
	test_utils.AssertTrue(t, cvResult.ConsentValidation.Target != "", "Got Target")
	mstub.MockTransactionEnd("t1")

	// validate consent as org1 service1 admin
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	cvResultBytes, err = ValidateConsent(stub, org1Service1AdminCaller, []string{"patient1", "org1Service1", "org1Datatype1", consentOptionWrite, strconv.FormatInt(time.Now().Unix(), 10)})
	test_utils.AssertTrue(t, err == nil, "Expected ValidateConsent to succeed")
	test_utils.AssertTrue(t, cvResultBytes != nil, "Expected ValidateConsent to succeed")
	cvResult = ValidationResultWithLog{}
	json.Unmarshal(cvResultBytes, &cvResult)
	test_utils.AssertTrue(t, cvResult.ConsentValidation.PermissionGranted == true, "Got consent validation correctly")
	test_utils.AssertTrue(t, cvResult.ConsentValidation.Requester != "", "Expected Requester")
	test_utils.AssertTrue(t, cvResult.ConsentValidation.Owner != "", "Got Owner")
	test_utils.AssertTrue(t, cvResult.ConsentValidation.Target != "", "Got Target")
	mstub.MockTransactionEnd("t1")

	// validate consent as org1 service2 admin
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	cvResultBytes, err = ValidateConsent(stub, org1Service2AdminCaller, []string{"patient1", "org1Service1", "org1Datatype1", consentOptionWrite, strconv.FormatInt(time.Now().Unix(), 10)})
	test_utils.AssertTrue(t, err == nil, "Expected ValidateConsent to succeed")
	test_utils.AssertTrue(t, cvResultBytes != nil, "Expected ValidateConsent to succeed")
	cvResult = ValidationResultWithLog{}
	json.Unmarshal(cvResultBytes, &cvResult)
	consentToken, _ = url.QueryUnescape(cvResult.ConsentValidation.Token)
	_, err = DecryptConsentValidationToken(stub, org1, consentToken)
	test_utils.AssertTrue(t, err != nil, "Expected DecryptConsentValidationToken to fail")
	test_utils.AssertTrue(t, cvResult.ConsentValidation.PermissionGranted == false, "Expected no PermissionGranted")
	test_utils.AssertTrue(t, cvResult.ConsentValidation.Requester != "", "Expected Requester")
	test_utils.AssertTrue(t, cvResult.ConsentValidation.Owner != "", "Got Owner")
	test_utils.AssertTrue(t, cvResult.ConsentValidation.Target != "", "Got Target")
	mstub.MockTransactionEnd("t1")

	// validate consent as org2 admin
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	cvResultBytes, err = ValidateConsent(stub, org2, []string{"patient1", "org1Service1", "org1Datatype1", consentOptionWrite, strconv.FormatInt(time.Now().Unix(), 10)})
	test_utils.AssertTrue(t, err == nil, "Expected ValidateConsent to succeed")
	test_utils.AssertTrue(t, cvResultBytes != nil, "Expected ValidateConsent to succeed")
	cvResult = ValidationResultWithLog{}
	json.Unmarshal(cvResultBytes, &cvResult)
	consentToken, _ = url.QueryUnescape(cvResult.ConsentValidation.Token)
	_, err = DecryptConsentValidationToken(stub, org1, consentToken)
	test_utils.AssertTrue(t, err != nil, "Expected DecryptConsentValidationToken to fail")
	test_utils.AssertTrue(t, cvResult.ConsentValidation.PermissionGranted == false, "Expected no PermissionGranted")
	test_utils.AssertTrue(t, cvResult.ConsentValidation.Requester != "", "Expected Requester")
	test_utils.AssertTrue(t, cvResult.ConsentValidation.Owner != "", "Got Owner")
	test_utils.AssertTrue(t, cvResult.ConsentValidation.Target != "", "Got Target")
	mstub.MockTransactionEnd("t1")

	// validate consent as org2 user
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	cvResultBytes, err = ValidateConsent(stub, org2UserCaller, []string{"patient1", "org1Service1", "org1Datatype1", consentOptionWrite, strconv.FormatInt(time.Now().Unix(), 10)})
	test_utils.AssertTrue(t, err == nil, "Expected ValidateConsent to succeed")
	test_utils.AssertTrue(t, cvResultBytes != nil, "Expected ValidateConsent to succeed")
	cvResult = ValidationResultWithLog{}
	json.Unmarshal(cvResultBytes, &cvResult)
	consentToken, _ = url.QueryUnescape(cvResult.ConsentValidation.Token)
	_, err = DecryptConsentValidationToken(stub, org1, consentToken)
	test_utils.AssertTrue(t, err != nil, "Expected DecryptConsentValidationToken to fail")
	test_utils.AssertTrue(t, cvResult.ConsentValidation.PermissionGranted == false, "Expected no PermissionGranted")
	test_utils.AssertTrue(t, cvResult.ConsentValidation.Requester != "", "Expected Requester")
	test_utils.AssertTrue(t, cvResult.ConsentValidation.Owner != "", "Got Owner")
	test_utils.AssertTrue(t, cvResult.ConsentValidation.Target != "", "Got Target")
	mstub.MockTransactionEnd("t1")

	// validate consent as org2 service admin
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	cvResultBytes, err = ValidateConsent(stub, org2Service1AdminCaller, []string{"patient1", "org1Service1", "org1Datatype1", consentOptionWrite, strconv.FormatInt(time.Now().Unix(), 10)})
	test_utils.AssertTrue(t, err == nil, "Expected ValidateConsent to succeed")
	test_utils.AssertTrue(t, cvResultBytes != nil, "Expected ValidateConsent to succeed")
	cvResult = ValidationResultWithLog{}
	json.Unmarshal(cvResultBytes, &cvResult)
	test_utils.AssertTrue(t, cvResult.ConsentValidation.PermissionGranted == false, "Expected no PermissionGranted")
	test_utils.AssertTrue(t, cvResult.ConsentValidation.Requester != "", "Expected Requester")
	test_utils.AssertTrue(t, cvResult.ConsentValidation.Owner != "", "Got Owner")
	test_utils.AssertTrue(t, cvResult.ConsentValidation.Target != "", "Got Target")
	mstub.MockTransactionEnd("t1")

	// validate consent as patient1
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	cvResultBytes, err = ValidateConsent(stub, patient1Caller, []string{"patient1", "org1Service1", "org1Datatype1", consentOptionWrite, strconv.FormatInt(time.Now().Unix(), 10)})
	test_utils.AssertTrue(t, err == nil, "Expected ValidateConsent to succeed")
	test_utils.AssertTrue(t, cvResultBytes != nil, "Expected ValidateConsent to succeed")
	cvResult = ValidationResultWithLog{}
	json.Unmarshal(cvResultBytes, &cvResult)
	test_utils.AssertTrue(t, cvResult.ConsentValidation.PermissionGranted == true, "Got consent validation correctly")
	test_utils.AssertTrue(t, cvResult.ConsentValidation.Requester != "", "Expected Requester")
	test_utils.AssertTrue(t, cvResult.ConsentValidation.Owner != "", "Got Owner")
	test_utils.AssertTrue(t, cvResult.ConsentValidation.Target != "", "Got Target")
	mstub.MockTransactionEnd("t1")

	// validate consent as auditor1
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	cvResultBytes, err = ValidateConsent(stub, auditor1Caller, []string{"patient1", "org1Service1", "org1Datatype1", consentOptionWrite, strconv.FormatInt(time.Now().Unix(), 10)})
	test_utils.AssertTrue(t, err == nil, "Expected ValidateConsent to succeed")
	test_utils.AssertTrue(t, cvResultBytes != nil, "Expected ValidateConsent to succeed")
	cvResult = ValidationResultWithLog{}
	json.Unmarshal(cvResultBytes, &cvResult)
	test_utils.AssertTrue(t, cvResult.ConsentValidation.PermissionGranted == false, "Got consent validation correctly")
	test_utils.AssertTrue(t, cvResult.ConsentValidation.Requester != "", "Expected Requester")
	test_utils.AssertTrue(t, cvResult.ConsentValidation.Owner != "", "Got Owner")
	test_utils.AssertTrue(t, cvResult.ConsentValidation.Target != "", "Got Target")
	mstub.MockTransactionEnd("t1")
}

func TestQueryConsentRequests(t *testing.T) {
	logger.SetLevel(shim.LogDebug)
	logger.Info("TestQueryConsentRequests function called")

	mstub, systemAdmin := querySetup(t)

	// ------- ORG 1 -------

	// register org1
	mstub.MockTransactionStart("t1")
	stub := cached_stub.NewCachedStub(mstub)
	org1 := test_utils.CreateTestGroup("org1")
	org1Bytes, _ := json.Marshal(&org1)
	_, err := RegisterOrg(stub, org1, []string{string(org1Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterOrg to succeed")
	mstub.MockTransactionEnd("t1")

	org1UserCaller := registerOrgUser(t, mstub, org1, "1")

	org1Service1AdminCaller, auditor1Caller := registerServiceAdmin(t, mstub, org1, systemAdmin, "1")
	org1Service2AdminCaller, _ := registerServiceAdmin(t, mstub, org1, systemAdmin, "2")

	// ------- ORG 2 -------

	// register org2
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	org2 := test_utils.CreateTestGroup("org2")
	org2Bytes, _ := json.Marshal(&org2)
	_, err = RegisterOrg(stub, org2, []string{string(org2Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterOrg to succeed")
	mstub.MockTransactionEnd("t1")

	org2UserCaller := registerOrgUser(t, mstub, org2, "1")

	org2Service1AdminCaller, _ := registerServiceAdmin(t, mstub, org2, systemAdmin, "1")

	// ------- PATIENTS -------

	// register patient1
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	patient1 := CreateTestSolutionUser("patient1")
	patient1PlatformUser, err := convertToPlatformUser(stub, patient1)
	test_utils.AssertTrue(t, err == nil, "Expected convertToPlatformUser to succeed")
	patient1Bytes, _ := json.Marshal(&patient1)
	_, err = RegisterUser(stub, patient1PlatformUser, []string{string(patient1Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterUser to succeed")
	mstub.MockTransactionEnd("t1")

	// register patient2
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	patient2 := CreateTestSolutionUser("patient2")
	patient2PlatformUser, err := convertToPlatformUser(stub, patient2)
	test_utils.AssertTrue(t, err == nil, "Expected convertToPlatformUser to succeed")
	patient2Bytes, _ := json.Marshal(&patient2)
	_, err = RegisterUser(stub, patient2PlatformUser, []string{string(patient2Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterUser to succeed")
	mstub.MockTransactionEnd("t1")

	// ------- ENROLLMENT -------

	// enroll patient1 to org1 service1
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	enrollment1 := GenerateEnrollmentTest(patient1.ID, "org1Service1")
	enrollment1Bytes, _ := json.Marshal(&enrollment1)
	enrollmentKey1 := test_utils.GenerateSymKey()
	enrollmentKey1B64 := crypto.EncodeToB64String(enrollmentKey1)
	_, err = EnrollPatient(stub, org1Service1AdminCaller, []string{string(enrollment1Bytes), enrollmentKey1B64})
	test_utils.AssertTrue(t, err == nil, "Expected EnrollPatient to succeed")
	mstub.MockTransactionEnd("t1")

	// ------- CONSENT -------

	// patient1 gives consent
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub, true, true, true)
	consent := Consent{}
	consent.Owner = "patient1"
	consent.Service = "org1Service1"
	consent.Target = "org1Service1"
	consent.Datatype = "org1Datatype1"
	consent.Option = []string{consentOptionWrite, consentOptionRead}
	consent.Timestamp = time.Now().Unix()
	consent.Expiration = 0
	consentBytes, _ := json.Marshal(&consent)
	consentKey := test_utils.GenerateSymKey()
	consentKeyB64 := crypto.EncodeToB64String(consentKey)
	patient1Caller, _ := user_mgmt.GetUserData(stub, patient1PlatformUser, patient1.ID, true, true)
	_, err = PutConsentPatientData(stub, patient1Caller, []string{string(consentBytes), consentKeyB64})
	test_utils.AssertTrue(t, err == nil, "Expected PutConsentPatientData to succeed")
	mstub.MockTransactionEnd("t1")

	// ------- GET CONSENT REQUESTS -------

	// func GetAllConsentRequests(stub cached_stub.CachedStubInterface, caller data_model.User, args []string) ([]byte, error) {

	// get consent request as org1 admin
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	org1Caller, _ := user_mgmt.GetUserData(stub, org1, org1.ID, true, true)
	consentRequestsBytes, err := GetAllConsentRequests(stub, org1Caller, []string{"patient1", "org1Service1"})
	test_utils.AssertTrue(t, err == nil, "Expected GetAllConsentRequests to succeed")
	test_utils.AssertTrue(t, consentRequestsBytes != nil, "Expected GetAllConsentRequests to succeed")
	consentRequestsResult := []ConsentRequest{}
	json.Unmarshal(consentRequestsBytes, &consentRequestsResult)
	test_utils.AssertTrue(t, consentRequestsResult[0].Owner != "", "Expected Owner")
	test_utils.AssertTrue(t, consentRequestsResult[0].Org != "", "Expected Org")
	test_utils.AssertTrue(t, consentRequestsResult[0].Service != "", "Expected Service")
	test_utils.AssertTrue(t, consentRequestsResult[0].ServiceName != "", "Expected ServiceName")
	mstub.MockTransactionEnd("t1")

	// validate consent as org1 user
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = GetAllConsentRequests(stub, org1UserCaller, []string{"patient1", "org1Service1"})
	test_utils.AssertTrue(t, err != nil, "Expected GetAllConsentRequests to fail")
	mstub.MockTransactionEnd("t1")

	// validate consent as org1 service1 admin
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	consentRequestsBytes, err = GetAllConsentRequests(stub, org1Service1AdminCaller, []string{"patient1", "org1Service1"})
	test_utils.AssertTrue(t, err == nil, "Expected GetAllConsentRequests to succeed")
	test_utils.AssertTrue(t, consentRequestsBytes != nil, "Expected GetAllConsentRequests to succeed")
	consentRequestsResult = []ConsentRequest{}
	json.Unmarshal(consentRequestsBytes, &consentRequestsResult)
	test_utils.AssertTrue(t, consentRequestsResult[0].Owner != "", "Expected Owner")
	test_utils.AssertTrue(t, consentRequestsResult[0].Org != "", "Expected Org")
	test_utils.AssertTrue(t, consentRequestsResult[0].Service != "", "Expected Service")
	test_utils.AssertTrue(t, consentRequestsResult[0].ServiceName != "", "Expected ServiceName")
	mstub.MockTransactionEnd("t1")

	// validate consent as org1 service2 admin
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = GetAllConsentRequests(stub, org1Service2AdminCaller, []string{"patient1", "org1Service1"})
	test_utils.AssertTrue(t, err != nil, "Expected GetAllConsentRequests to fail")
	mstub.MockTransactionEnd("t1")

	// validate consent as org2 admin
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = GetAllConsentRequests(stub, org2, []string{"patient1", "org1Service1"})
	test_utils.AssertTrue(t, err != nil, "Expected GetAllConsentRequests to fail")
	mstub.MockTransactionEnd("t1")

	// validate consent as org2 user
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = GetAllConsentRequests(stub, org2UserCaller, []string{"patient1", "org1Service1"})
	test_utils.AssertTrue(t, err != nil, "Expected GetAllConsentRequests to fail")
	mstub.MockTransactionEnd("t1")

	// validate consent as org2 service admin
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = GetAllConsentRequests(stub, org2Service1AdminCaller, []string{"patient1", "org1Service1"})
	test_utils.AssertTrue(t, err != nil, "Expected GetAllConsentRequests to fail")
	mstub.MockTransactionEnd("t1")

	// validate consent as patient1
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	consentRequestsBytes, err = GetAllConsentRequests(stub, patient1Caller, []string{"patient1", "org1Service1"})
	test_utils.AssertTrue(t, err == nil, "Expected GetAllConsentRequests to succeed")
	consentRequestsResult = []ConsentRequest{}
	json.Unmarshal(consentRequestsBytes, &consentRequestsResult)
	test_utils.AssertTrue(t, consentRequestsResult[0].Owner != "", "Expected Owner")
	test_utils.AssertTrue(t, consentRequestsResult[0].Org != "", "Expected Org")
	test_utils.AssertTrue(t, consentRequestsResult[0].Service != "", "Expected Service")
	test_utils.AssertTrue(t, consentRequestsResult[0].ServiceName != "", "Expected ServiceName")
	mstub.MockTransactionEnd("t1")

	// validate consent as patient2
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	patient2Caller, _ := user_mgmt.GetUserData(stub, patient2PlatformUser, patient2.ID, true, true)
	_, err = GetAllConsentRequests(stub, patient2Caller, []string{"patient1", "org1Service1"})
	test_utils.AssertTrue(t, err != nil, "Expected GetAllConsentRequests to fail")
	mstub.MockTransactionEnd("t1")

	// validate consent as auditor1
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = GetAllConsentRequests(stub, auditor1Caller, []string{"patient1", "org1Service1"})
	test_utils.AssertTrue(t, err != nil, "Expected GetAllConsentRequests to fail")
	mstub.MockTransactionEnd("t1")
}

func TestQueryLogs(t *testing.T) {
	logger.SetLevel(shim.LogDebug)
	logger.Info("TestQueryLogs function called")

	mstub, systemAdmin := querySetup(t)

	// ------- ORG 1 -------

	// register org1
	mstub.MockTransactionStart("t1")
	stub := cached_stub.NewCachedStub(mstub)
	org1 := test_utils.CreateTestGroup("org1")
	org1Bytes, _ := json.Marshal(&org1)
	_, err := RegisterOrg(stub, org1, []string{string(org1Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterOrg to succeed")
	mstub.MockTransactionEnd("t1")

	org1UserCaller := registerOrgUser(t, mstub, org1, "1")

	org1Service1AdminCaller, auditor1Caller := registerServiceAdmin(t, mstub, org1, systemAdmin, "1")
	org1Service2AdminCaller, _ := registerServiceAdmin(t, mstub, org1, systemAdmin, "2")

	// ------- ORG 2 -------

	// register org2
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	org2 := test_utils.CreateTestGroup("org2")
	org2Bytes, _ := json.Marshal(&org2)
	_, err = RegisterOrg(stub, org2, []string{string(org2Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterOrg to succeed")
	mstub.MockTransactionEnd("t1")

	org2UserCaller := registerOrgUser(t, mstub, org2, "1")

	org2Service1AdminCaller, _ := registerServiceAdmin(t, mstub, org2, systemAdmin, "1")

	// ------- PATIENTS -------

	// register patient1
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	patient1 := CreateTestSolutionUser("patient1")
	patient1PlatformUser, err := convertToPlatformUser(stub, patient1)
	test_utils.AssertTrue(t, err == nil, "Expected convertToPlatformUser to succeed")
	patient1Bytes, _ := json.Marshal(&patient1)
	_, err = RegisterUser(stub, patient1PlatformUser, []string{string(patient1Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterUser to succeed")
	mstub.MockTransactionEnd("t1")

	// register patient2
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	patient2 := CreateTestSolutionUser("patient2")
	patient2PlatformUser, err := convertToPlatformUser(stub, patient2)
	test_utils.AssertTrue(t, err == nil, "Expected convertToPlatformUser to succeed")
	patient2Bytes, _ := json.Marshal(&patient2)
	_, err = RegisterUser(stub, patient2PlatformUser, []string{string(patient2Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterUser to succeed")
	mstub.MockTransactionEnd("t1")

	// ------- ENROLLMENT -------

	// enroll patient1 to org1 service1
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	enrollment1 := GenerateEnrollmentTest(patient1.ID, "org1Service1")
	enrollment1Bytes, _ := json.Marshal(&enrollment1)
	enrollmentKey1 := test_utils.GenerateSymKey()
	enrollmentKey1B64 := crypto.EncodeToB64String(enrollmentKey1)
	_, err = EnrollPatient(stub, org1Service1AdminCaller, []string{string(enrollment1Bytes), enrollmentKey1B64})
	test_utils.AssertTrue(t, err == nil, "Expected EnrollPatient to succeed")
	mstub.MockTransactionEnd("t1")

	// ------- CONSENT -------

	// patient1 gives consent
	mstub.MockTransactionStart("giveConsent")
	stub = cached_stub.NewCachedStub(mstub, true, true, true)
	consent := Consent{}
	consent.Owner = "patient1"
	consent.Service = "org1Service1"
	consent.Target = "org1Service1"
	consent.Datatype = "org1Datatype1"
	consent.Option = []string{consentOptionWrite, consentOptionRead}
	consent.Timestamp = time.Now().Unix()
	consent.Expiration = 0
	consentBytes, _ := json.Marshal(&consent)
	consentKey := test_utils.GenerateSymKey()
	consentKeyB64 := crypto.EncodeToB64String(consentKey)
	patient1Caller, _ := user_mgmt.GetUserData(stub, patient1PlatformUser, patient1.ID, true, true)
	_, err = PutConsentPatientData(stub, patient1Caller, []string{string(consentBytes), consentKeyB64})
	test_utils.AssertTrue(t, err == nil, "Expected PutConsentPatientData to succeed")
	mstub.MockTransactionEnd("giveConsent")

	// ------- UPLOAD PATIENT DATA -------

	// upload patient data as service1 admin
	mstub.MockTransactionStart("uploadPatientData")
	stub = cached_stub.NewCachedStub(mstub)
	patientData := GeneratePatientData("patient1", "org1Datatype1", "org1Service1")
	patientDataBytes, _ := json.Marshal(&patientData)
	dataKey := test_utils.GenerateSymKey()
	dataKeyB64 := crypto.EncodeToB64String(dataKey)
	args := []string{string(patientDataBytes), dataKeyB64}
	_, err = UploadUserData(stub, org1Service1AdminCaller, args)
	test_utils.AssertTrue(t, err == nil, "Expected UploadUserData to succeed")
	mstub.MockTransactionEnd("uploadPatientData")

	// ------- GET LOGS -------

	// get logs as org1 admin
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	org1Caller, _ := user_mgmt.GetUserData(stub, org1, org1.ID, true, true)
	logBytes, err := GetLogs(stub, org1Caller, []string{"", "patient1", "", "", "", "", strconv.FormatInt(0, 10), strconv.FormatInt(0, 10), "false", "20"})
	test_utils.AssertTrue(t, err == nil, "Expected GetLogs to succeed")
	logsOMR := []Log{}
	json.Unmarshal(logBytes, &logsOMR)
	test_utils.AssertTrue(t, len(logsOMR) == 2, "Expected 2 logs")
	test_utils.AssertTrue(t, logsOMR[0].Type == "PutConsentPatientData", "Expected PutConsentPatientData log")
	test_utils.AssertTrue(t, logsOMR[0].Owner != "", "Expected Owner")
	test_utils.AssertTrue(t, logsOMR[1].Type == "UploadUserData", "Expected UploadUserData log")
	test_utils.AssertTrue(t, logsOMR[1].Owner != "", "Expected Owner")
	mstub.MockTransactionEnd("t1")

	// get logs as org1 user (X)
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	logBytes, err = GetLogs(stub, org1UserCaller, []string{"", "patient1", "", "", "", "", strconv.FormatInt(0, 10), strconv.FormatInt(0, 10), "false", "20"})
	test_utils.AssertTrue(t, err == nil, "Expected GetLogs to succeed")
	logsOMR = []Log{}
	json.Unmarshal(logBytes, &logsOMR)
	test_utils.AssertTrue(t, len(logsOMR) == 0, "Expected 0 logs")
	mstub.MockTransactionEnd("t1")

	// get logs as org1 service1 admin
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	logBytes, err = GetLogs(stub, org1Service1AdminCaller, []string{"", "patient1", "", "", "", "", strconv.FormatInt(0, 10), strconv.FormatInt(0, 10), "false", "20"})
	test_utils.AssertTrue(t, err == nil, "Expected GetLogs to succeed")
	logsOMR = []Log{}
	json.Unmarshal(logBytes, &logsOMR)
	test_utils.AssertTrue(t, len(logsOMR) == 2, "Expected 2 logs")
	test_utils.AssertTrue(t, logsOMR[0].Type == "PutConsentPatientData", "Expected PutConsentPatientData log")
	test_utils.AssertTrue(t, logsOMR[0].Owner != "", "Expected Owner")
	test_utils.AssertTrue(t, logsOMR[1].Type == "UploadUserData", "Expected UploadUserData log")
	test_utils.AssertTrue(t, logsOMR[1].Owner != "", "Expected Owner")
	mstub.MockTransactionEnd("t1")

	// get logs as org1 service2 admin (X)
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	logBytes, err = GetLogs(stub, org1Service2AdminCaller, []string{"", "patient1", "", "", "", "", strconv.FormatInt(0, 10), strconv.FormatInt(0, 10), "false", "20"})
	test_utils.AssertTrue(t, err == nil, "Expected GetLogs to succeed")
	logsOMR = []Log{}
	json.Unmarshal(logBytes, &logsOMR)
	test_utils.AssertTrue(t, len(logsOMR) == 0, "Expected 0 logs")
	mstub.MockTransactionEnd("t1")

	// get logs as org2 admin (X)
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	logBytes, err = GetLogs(stub, org2, []string{"", "patient1", "", "", "", "", strconv.FormatInt(0, 10), strconv.FormatInt(0, 10), "false", "20"})
	test_utils.AssertTrue(t, err == nil, "Expected GetLogs to succeed")
	logsOMR = []Log{}
	json.Unmarshal(logBytes, &logsOMR)
	test_utils.AssertTrue(t, len(logsOMR) == 0, "Expected 0 logs")
	mstub.MockTransactionEnd("t1")

	// get logs as org2 user (X)
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	logBytes, err = GetLogs(stub, org2UserCaller, []string{"", "patient1", "", "", "", "", strconv.FormatInt(0, 10), strconv.FormatInt(0, 10), "false", "20"})
	test_utils.AssertTrue(t, err == nil, "Expected GetLogs to succeed")
	logsOMR = []Log{}
	json.Unmarshal(logBytes, &logsOMR)
	test_utils.AssertTrue(t, len(logsOMR) == 0, "Expected 0 logs")
	mstub.MockTransactionEnd("t1")

	// get logs as org2 service admin (X)
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	logBytes, err = GetLogs(stub, org2Service1AdminCaller, []string{"", "patient1", "", "", "", "", strconv.FormatInt(0, 10), strconv.FormatInt(0, 10), "false", "20"})
	test_utils.AssertTrue(t, err == nil, "Expected GetLogs to succeed")
	logsOMR = []Log{}
	json.Unmarshal(logBytes, &logsOMR)
	test_utils.AssertTrue(t, len(logsOMR) == 0, "Expected 0 logs")
	mstub.MockTransactionEnd("t1")

	// get logs as patient1
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	logBytes, err = GetLogs(stub, patient1Caller, []string{"", "patient1", "", "", "", "", strconv.FormatInt(0, 10), strconv.FormatInt(0, 10), "false", "20"})
	test_utils.AssertTrue(t, err == nil, "Expected GetLogs to succeed")
	logsOMR = []Log{}
	json.Unmarshal(logBytes, &logsOMR)
	test_utils.AssertTrue(t, len(logsOMR) == 2, "Expected 2 logs")
	test_utils.AssertTrue(t, logsOMR[0].Type == "PutConsentPatientData", "Expected PutConsentPatientData log")
	test_utils.AssertTrue(t, logsOMR[0].Owner != "", "Expected Owner")
	test_utils.AssertTrue(t, logsOMR[1].Type == "UploadUserData", "Expected UploadUserData log")
	test_utils.AssertTrue(t, logsOMR[1].Owner != "", "Expected Owner")
	mstub.MockTransactionEnd("t1")

	// get logs as patient2 (X)
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	patient2Caller, _ := user_mgmt.GetUserData(stub, patient2PlatformUser, patient2.ID, true, true)
	logBytes, err = GetLogs(stub, patient2Caller, []string{"", "patient1", "", "", "", "", strconv.FormatInt(0, 10), strconv.FormatInt(0, 10), "false", "20"})
	test_utils.AssertTrue(t, err == nil, "Expected GetLogs to succeed")
	logsOMR = []Log{}
	json.Unmarshal(logBytes, &logsOMR)
	test_utils.AssertTrue(t, len(logsOMR) == 0, "Expected 0 logs")
	mstub.MockTransactionEnd("t1")

	// get logs as auditor1
	mstub.MockTransactionStart("t1")
	stub = cached_stub.NewCachedStub(mstub)
	logBytes, err = GetLogs(stub, auditor1Caller, []string{"", "patient1", "", "", "", "", strconv.FormatInt(0, 10), strconv.FormatInt(0, 10), "false", "20"})
	test_utils.AssertTrue(t, err == nil, "Expected GetLogs to succeed")
	logsOMR = []Log{}
	json.Unmarshal(logBytes, &logsOMR)
	test_utils.AssertTrue(t, len(logsOMR) == 2, "Expected 2 logs")
	test_utils.AssertTrue(t, logsOMR[0].Type == "PutConsentPatientData", "Expected PutConsentPatientData log")
	test_utils.AssertTrue(t, logsOMR[0].Owner == patient1.ID, "Expected Owner")
	test_utils.AssertTrue(t, logsOMR[1].Type == "UploadUserData", "Expected UploadUserData log")
	test_utils.AssertTrue(t, logsOMR[1].Owner != "", "Expected Owner")
	mstub.MockTransactionEnd("t1")
}
