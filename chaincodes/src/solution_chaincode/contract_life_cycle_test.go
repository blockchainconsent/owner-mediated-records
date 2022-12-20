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

func GenerateContractTest(contractID string, ownerOrgID string, ownerServiceID string, requesterOrgID string, requesterServiceID string) Contract {
	contract := Contract{}
	contract.ContractID = contractID
	contract.OwnerOrgID = ownerOrgID
	contract.OwnerServiceID = ownerServiceID
	contract.RequesterOrgID = requesterOrgID
	contract.RequesterServiceID = requesterServiceID
	terms := make(map[string]string)
	terms["need consent"] = "yes"
	terms["time"] = "first of each month"
	contract.ContractTerms = terms
	contract.State = "new"
	contract.CreateDate = time.Now().Unix()
	contract.PaymentRequired = "no"

	return contract
}

func TestCreateContract(t *testing.T) {
	logger.SetLevel(shim.LogDebug)
	logger.Info("TestCreateContract function called")

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

	// create requester org
	mstub.MockTransactionStart("1")
	stub = cached_stub.NewCachedStub(mstub)
	requesterOrg1 := test_utils.CreateTestGroup("requesterOrg1")
	requesterOrg1Bytes, _ := json.Marshal(&requesterOrg1)
	_, err = RegisterOrg(stub, requesterOrg1, []string{string(requesterOrg1Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterOrg to succeed")
	mstub.MockTransactionEnd("1")

	// create requester org datatype
	mstub.MockTransactionStart("2")
	stub = cached_stub.NewCachedStub(mstub)
	reqOrgDatatype1 := Datatype{DatatypeID: "reqOrgDatatype1", Description: "reqOrgDatatype1"}
	reqOrgDatatype1Bytes, _ := json.Marshal(&reqOrgDatatype1)
	requesterOrg1Caller, _ := user_mgmt.GetUserData(stub, requesterOrg1, requesterOrg1.ID, true, true)
	_, err = RegisterDatatype(stub, requesterOrg1Caller, []string{string(reqOrgDatatype1Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterDatatype to succeed")
	mstub.MockTransactionEnd("2")

	// create requester service
	mstub.MockTransactionStart("3")
	stub = cached_stub.NewCachedStub(mstub, true, true, true)
	reqServiceDatatype := GenerateServiceDatatypeForTesting("reqOrgDatatype1", "reqService1", []string{consentOptionWrite, consentOptionRead})
	reqService1 := GenerateServiceForTesting("reqService1", "requesterOrg1", []ServiceDatatype{reqServiceDatatype})
	reqService1Bytes, _ := json.Marshal(&reqService1)
	_, err = RegisterService(stub, requesterOrg1Caller, []string{string(reqService1Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterService to succeed")
	mstub.MockTransactionEnd("3")

	// create owner org
	mstub.MockTransactionStart("4")
	stub = cached_stub.NewCachedStub(mstub)
	ownerOrg1 := test_utils.CreateTestGroup("ownerOrg1")
	ownerOrg1Bytes, _ := json.Marshal(&ownerOrg1)
	_, err = RegisterOrg(stub, ownerOrg1, []string{string(ownerOrg1Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterOrg to succeed")
	mstub.MockTransactionEnd("4")

	// create owner org datatype
	mstub.MockTransactionStart("5")
	stub = cached_stub.NewCachedStub(mstub)
	ownerOrg1Caller, _ := user_mgmt.GetUserData(stub, ownerOrg1, ownerOrg1.ID, true, true)
	ownerOrgDatatype1 := Datatype{DatatypeID: "ownerOrgDatatype1", Description: "ownerOrgDatatype1"}
	ownerOrgDatatype1Bytes, _ := json.Marshal(&ownerOrgDatatype1)
	_, err = RegisterDatatype(stub, ownerOrg1Caller, []string{string(ownerOrgDatatype1Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterDatatype to succeed")
	mstub.MockTransactionEnd("5")

	// create owner service
	mstub.MockTransactionStart("6")
	stub = cached_stub.NewCachedStub(mstub, true, true, true)
	ownerServiceDatatype := GenerateServiceDatatypeForTesting("ownerOrgDatatype1", "ownerService1", []string{consentOptionWrite, consentOptionRead})
	ownerService1 := GenerateServiceForTesting("ownerService1", "ownerOrg1", []ServiceDatatype{ownerServiceDatatype})
	ownerService1Bytes, _ := json.Marshal(&ownerService1)
	_, err = RegisterService(stub, ownerOrg1Caller, []string{string(ownerService1Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterService to succeed")
	mstub.MockTransactionEnd("6")

	// create contract as requester
	mstub.MockTransactionStart("7")
	stub = cached_stub.NewCachedStub(mstub)
	contract1 := GenerateContractTest("contract1", "ownerOrg1", "ownerService1", "requesterOrg1", "reqService1")
	contract1Bytes, _ := json.Marshal(&contract1)
	contractKey := test_utils.GenerateSymKey()
	contractKeyB64 := crypto.EncodeToB64String(contractKey)
	args := []string{string(contract1Bytes), contractKeyB64}
	reqService1Subgroup, _ := user_mgmt.GetUserData(stub, requesterOrg1, "reqService1", true, true)
	_, err = CreateContract(stub, reqService1Subgroup, args)
	test_utils.AssertTrue(t, err == nil, "Expected CreateContract to succeed")
	mstub.MockTransactionEnd("7")

	// create contract as owner
	mstub.MockTransactionStart("8")
	stub = cached_stub.NewCachedStub(mstub)
	contract2 := GenerateContractTest("contract2", "ownerOrg1", "ownerService1", "requesterOrg1", "reqService1")
	contract2Bytes, _ := json.Marshal(&contract2)
	contractKey = test_utils.GenerateSymKey()
	contractKeyB64 = crypto.EncodeToB64String(contractKey)
	args = []string{string(contract2Bytes), contractKeyB64}
	ownerService1Subgroup, _ := user_mgmt.GetUserData(stub, ownerOrg1, "ownerService1", true, true)
	_, err = CreateContract(stub, ownerService1Subgroup, args)
	test_utils.AssertTrue(t, err == nil, "Expected CreateContract to succeed")
	mstub.MockTransactionEnd("8")

	// create an org user
	mstub.MockTransactionStart("11")
	stub = cached_stub.NewCachedStub(mstub)
	orgUser1 := CreateTestSolutionUser("orgUser1")
	orgUser1.Org = requesterOrg1.ID
	orgUser1Bytes, _ := json.Marshal(&orgUser1)
	_, err = RegisterUser(stub, requesterOrg1Caller, []string{string(orgUser1Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterUser to succeed")
	mstub.MockTransactionEnd("11")

	// put orgUser1 in org1
	mstub.MockTransactionStart("12")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = PutUserInOrg(stub, requesterOrg1Caller, []string{orgUser1.ID, requesterOrg1.ID, "false"})
	test_utils.AssertTrue(t, err == nil, "Expected PutUserInOrg to succeed")
	mstub.MockTransactionEnd("12")

	// give org admin permission
	mstub.MockTransactionStart("13")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = AddPermissionOrgAdmin(stub, requesterOrg1Caller, []string{orgUser1.ID, requesterOrg1.ID})
	test_utils.AssertTrue(t, err == nil, "Expected AddPermissionOrgAdmin to succeed")
	mstub.MockTransactionEnd("13")

	// create contract as org admin of requester service
	mstub.MockTransactionStart("14")
	stub = cached_stub.NewCachedStub(mstub)
	contract3 := GenerateContractTest("contract3", "ownerOrg1", "ownerService1", "requesterOrg1", "reqService1")
	contract3Bytes, _ := json.Marshal(&contract3)
	contractKey = test_utils.GenerateSymKey()
	contractKeyB64 = crypto.EncodeToB64String(contractKey)
	args = []string{string(contract3Bytes), contractKeyB64}
	orgUser1Caller, _ := user_mgmt.GetUserData(stub, requesterOrg1Caller, orgUser1.ID, true, true)
	_, err = CreateContract(stub, orgUser1Caller, args)
	test_utils.AssertTrue(t, err == nil, "Expected CreateContract to succeed")
	mstub.MockTransactionEnd("14")

	// create another org user
	mstub.MockTransactionStart("15")
	stub = cached_stub.NewCachedStub(mstub)
	orgUser2 := CreateTestSolutionUser("orgUser2")
	orgUser2.Org = requesterOrg1.ID
	orgUser2Bytes, _ := json.Marshal(&orgUser2)
	_, err = RegisterUser(stub, requesterOrg1Caller, []string{string(orgUser2Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterUser to succeed")
	mstub.MockTransactionEnd("15")

	// put orgUser2 in org1
	mstub.MockTransactionStart("16")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = PutUserInOrg(stub, requesterOrg1Caller, []string{orgUser2.ID, requesterOrg1.ID, "false"})
	test_utils.AssertTrue(t, err == nil, "Expected PutUserInOrg to succeed")
	mstub.MockTransactionEnd("16")

	// give service admin permission
	mstub.MockTransactionStart("17")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = AddPermissionServiceAdmin(stub, requesterOrg1Caller, []string{orgUser2.ID, "reqService1"})
	test_utils.AssertTrue(t, err == nil, "Expected AddPermissionServiceAdmin to succeed")
	mstub.MockTransactionEnd("17")

	// create contract as service admin of requester service
	mstub.MockTransactionStart("18")
	stub = cached_stub.NewCachedStub(mstub)
	contract4 := GenerateContractTest("contract4", "ownerOrg1", "ownerService1", "requesterOrg1", "reqService1")
	contract4Bytes, _ := json.Marshal(&contract4)
	contractKey = test_utils.GenerateSymKey()
	contractKeyB64 = crypto.EncodeToB64String(contractKey)
	args = []string{string(contract4Bytes), contractKeyB64}
	orgUser2Caller, _ := user_mgmt.GetUserData(stub, requesterOrg1Caller, orgUser2.ID, true, true)
	_, err = CreateContract(stub, orgUser2Caller, args)
	test_utils.AssertTrue(t, err == nil, "Expected CreateContract to succeed")
	mstub.MockTransactionEnd("18")
}

func TestGetContract(t *testing.T) {
	logger.SetLevel(shim.LogDebug)
	logger.Info("TestGetContract function called")

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

	// create requester org
	mstub.MockTransactionStart("1")
	stub = cached_stub.NewCachedStub(mstub)
	requesterOrg1 := test_utils.CreateTestGroup("requesterOrg1")
	requesterOrg1Bytes, _ := json.Marshal(&requesterOrg1)
	_, err = RegisterOrg(stub, requesterOrg1, []string{string(requesterOrg1Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterOrg to succeed")
	mstub.MockTransactionEnd("1")

	// create requester org datatype
	mstub.MockTransactionStart("2")
	stub = cached_stub.NewCachedStub(mstub)
	requesterOrg1Caller, _ := user_mgmt.GetUserData(stub, requesterOrg1, requesterOrg1.ID, true, true)
	reqOrgDatatype1 := Datatype{DatatypeID: "reqOrgDatatype1", Description: "reqOrgDatatype1"}
	reqOrgDatatype1Bytes, _ := json.Marshal(&reqOrgDatatype1)
	_, err = RegisterDatatype(stub, requesterOrg1Caller, []string{string(reqOrgDatatype1Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterDatatype to succeed")
	mstub.MockTransactionEnd("2")

	// create requester service
	mstub.MockTransactionStart("3")
	stub = cached_stub.NewCachedStub(mstub, true, true, true)
	reqServiceDatatype := GenerateServiceDatatypeForTesting("reqOrgDatatype1", "reqService1", []string{consentOptionWrite, consentOptionRead})
	reqService1 := GenerateServiceForTesting("reqService1", "requesterOrg1", []ServiceDatatype{reqServiceDatatype})
	reqService1Bytes, _ := json.Marshal(&reqService1)
	_, err = RegisterService(stub, requesterOrg1Caller, []string{string(reqService1Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterService to succeed")
	mstub.MockTransactionEnd("3")

	// create owner org
	mstub.MockTransactionStart("4")
	stub = cached_stub.NewCachedStub(mstub)
	ownerOrg1 := test_utils.CreateTestGroup("ownerOrg1")
	ownerOrg1Bytes, _ := json.Marshal(&ownerOrg1)
	_, err = RegisterOrg(stub, ownerOrg1, []string{string(ownerOrg1Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterOrg to succeed")
	mstub.MockTransactionEnd("4")

	// create owner org datatype
	mstub.MockTransactionStart("5")
	stub = cached_stub.NewCachedStub(mstub)
	ownerOrg1Caller, _ := user_mgmt.GetUserData(stub, ownerOrg1, ownerOrg1.ID, true, true)
	ownerOrgDatatype1 := Datatype{DatatypeID: "ownerOrgDatatype1", Description: "ownerOrgDatatype1"}
	ownerOrgDatatype1Bytes, _ := json.Marshal(&ownerOrgDatatype1)
	_, err = RegisterDatatype(stub, ownerOrg1Caller, []string{string(ownerOrgDatatype1Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterDatatype to succeed")
	mstub.MockTransactionEnd("5")

	// create owner service
	mstub.MockTransactionStart("6")
	stub = cached_stub.NewCachedStub(mstub, true, true, true)
	ownerServiceDatatype := GenerateServiceDatatypeForTesting("ownerOrgDatatype1", "ownerService1", []string{consentOptionWrite, consentOptionRead})
	ownerService1 := GenerateServiceForTesting("ownerService1", "ownerOrg1", []ServiceDatatype{ownerServiceDatatype})
	ownerService1Bytes, _ := json.Marshal(&ownerService1)
	_, err = RegisterService(stub, ownerOrg1Caller, []string{string(ownerService1Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterService to succeed")
	mstub.MockTransactionEnd("6")

	// create contract as requester
	mstub.MockTransactionStart("7")
	stub = cached_stub.NewCachedStub(mstub)
	contract1 := GenerateContractTest("contract1", "ownerOrg1", "ownerService1", "requesterOrg1", "reqService1")
	contract1Bytes, _ := json.Marshal(&contract1)
	contractKey := test_utils.GenerateSymKey()
	contractKeyB64 := crypto.EncodeToB64String(contractKey)
	args := []string{string(contract1Bytes), contractKeyB64}
	reqService1Subgroup, _ := user_mgmt.GetUserData(stub, requesterOrg1, "reqService1", true, true)
	_, err = CreateContract(stub, reqService1Subgroup, args)
	test_utils.AssertTrue(t, err == nil, "Expected CreateContract to succeed")
	mstub.MockTransactionEnd("7")

	// get contract by ID as default requester org admin
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	contractBytes, err := GetContract(stub, requesterOrg1Caller, []string{contract1.ContractID})
	test_utils.AssertTrue(t, err == nil, "GetContract succcessful")
	var contract = Contract{}
	_ = json.Unmarshal(contractBytes, &contract)
	test_utils.AssertTrue(t, contract.OwnerServiceID == contract1.OwnerServiceID, "Got service ID correctly")
	mstub.MockTransactionEnd("t123")

	// get contract by ID as default requester service admin
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	contractBytes, err = GetContract(stub, reqService1Subgroup, []string{contract1.ContractID})
	test_utils.AssertTrue(t, err == nil, "GetContract succcessful")
	_ = json.Unmarshal(contractBytes, &contract)
	test_utils.AssertTrue(t, contract.OwnerServiceID == contract1.OwnerServiceID, "Got service ID correctly")
	mstub.MockTransactionEnd("t123")

	// create an org user
	mstub.MockTransactionStart("11")
	stub = cached_stub.NewCachedStub(mstub)
	orgUser1 := CreateTestSolutionUser("orgUser1")
	orgUser1.Org = requesterOrg1.ID
	orgUser1Bytes, _ := json.Marshal(&orgUser1)
	_, err = RegisterUser(stub, requesterOrg1Caller, []string{string(orgUser1Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterUser to succeed")
	mstub.MockTransactionEnd("11")

	// put orgUser1 in requester org
	mstub.MockTransactionStart("12")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = PutUserInOrg(stub, requesterOrg1Caller, []string{orgUser1.ID, requesterOrg1.ID, "false"})
	test_utils.AssertTrue(t, err == nil, "Expected PutUserInOrg to succeed")
	mstub.MockTransactionEnd("12")

	// give org admin permission
	mstub.MockTransactionStart("13")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = AddPermissionOrgAdmin(stub, requesterOrg1Caller, []string{orgUser1.ID, requesterOrg1.ID})
	test_utils.AssertTrue(t, err == nil, "Expected AddPermissionOrgAdmin to succeed")
	mstub.MockTransactionEnd("13")

	// get contract by ID as org user with org admin permission
	mstub.MockTransactionStart("14")
	stub = cached_stub.NewCachedStub(mstub)
	orgUser1Caller, _ := user_mgmt.GetUserData(stub, requesterOrg1Caller, orgUser1.ID, true, true)
	contractBytes, err = GetContract(stub, orgUser1Caller, []string{contract1.ContractID})
	test_utils.AssertTrue(t, err == nil, "GetContract succcessful")
	_ = json.Unmarshal(contractBytes, &contract)
	test_utils.AssertTrue(t, contract.OwnerServiceID == contract1.OwnerServiceID, "Got service ID correctly")
	mstub.MockTransactionEnd("14")

	// create another org user
	mstub.MockTransactionStart("15")
	stub = cached_stub.NewCachedStub(mstub)
	orgUser2 := CreateTestSolutionUser("orgUser2")
	orgUser2.Org = requesterOrg1.ID
	orgUser2Bytes, _ := json.Marshal(&orgUser2)
	_, err = RegisterUser(stub, requesterOrg1Caller, []string{string(orgUser2Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterUser to succeed")
	mstub.MockTransactionEnd("15")

	// put orgUser2 in requester org
	mstub.MockTransactionStart("16")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = PutUserInOrg(stub, requesterOrg1Caller, []string{orgUser2.ID, requesterOrg1.ID, "false"})
	test_utils.AssertTrue(t, err == nil, "Expected PutUserInOrg to succeed")
	mstub.MockTransactionEnd("16")

	// give service admin permission
	mstub.MockTransactionStart("17")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = AddPermissionServiceAdmin(stub, requesterOrg1Caller, []string{orgUser2.ID, "reqService1"})
	test_utils.AssertTrue(t, err == nil, "Expected AddPermissionServiceAdmin to succeed")
	mstub.MockTransactionEnd("17")

	// get contract by ID as org user with service admin permission
	mstub.MockTransactionStart("14")
	stub = cached_stub.NewCachedStub(mstub)
	orgUser2Caller, _ := user_mgmt.GetUserData(stub, requesterOrg1Caller, orgUser2.ID, true, true)
	contractBytes, err = GetContract(stub, orgUser2Caller, []string{contract1.ContractID})
	test_utils.AssertTrue(t, err == nil, "GetContract succcessful")
	_ = json.Unmarshal(contractBytes, &contract)
	test_utils.AssertTrue(t, contract.OwnerServiceID == contract1.OwnerServiceID, "Got service ID correctly")
	mstub.MockTransactionEnd("14")

	// get contract by ID as default owner service admin
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	ownerService1Subgroup, _ := user_mgmt.GetUserData(stub, ownerOrg1, "ownerService1", true, true)
	contractBytes, err = GetContract(stub, ownerService1Subgroup, []string{contract1.ContractID})
	test_utils.AssertTrue(t, err == nil, "GetContract succcessful")
	_ = json.Unmarshal(contractBytes, &contract)
	test_utils.AssertTrue(t, contract.OwnerServiceID == contract1.OwnerServiceID, "Got service ID correctly")
	mstub.MockTransactionEnd("t123")

	// get contract by ID as default owner org admin
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	contractBytes, err = GetContract(stub, ownerOrg1Caller, []string{contract1.ContractID})
	test_utils.AssertTrue(t, err == nil, "GetContract succcessful")
	_ = json.Unmarshal(contractBytes, &contract)
	test_utils.AssertTrue(t, contract.OwnerServiceID == contract1.OwnerServiceID, "Got service ID correctly")
	mstub.MockTransactionEnd("t123")

	// create an org user
	mstub.MockTransactionStart("11")
	stub = cached_stub.NewCachedStub(mstub)
	orgUser3 := CreateTestSolutionUser("orgUser3")
	orgUser3.Org = ownerOrg1.ID
	orgUser3Bytes, _ := json.Marshal(&orgUser3)
	_, err = RegisterUser(stub, ownerOrg1Caller, []string{string(orgUser3Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterUser to succeed")
	mstub.MockTransactionEnd("11")

	// put orgUser3 in owner org
	mstub.MockTransactionStart("12")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = PutUserInOrg(stub, ownerOrg1Caller, []string{orgUser3.ID, ownerOrg1.ID, "false"})
	test_utils.AssertTrue(t, err == nil, "Expected PutUserInOrg to succeed")
	mstub.MockTransactionEnd("12")

	// give org admin permission
	mstub.MockTransactionStart("13")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = AddPermissionOrgAdmin(stub, ownerOrg1Caller, []string{orgUser3.ID, ownerOrg1.ID})
	test_utils.AssertTrue(t, err == nil, "Expected AddPermissionOrgAdmin to succeed")
	mstub.MockTransactionEnd("13")

	// get contract by ID as org user with org admin permission
	mstub.MockTransactionStart("14")
	stub = cached_stub.NewCachedStub(mstub)
	orgUser3Caller, _ := user_mgmt.GetUserData(stub, ownerOrg1Caller, orgUser3.ID, true, true)
	contractBytes, err = GetContract(stub, orgUser3Caller, []string{contract1.ContractID})
	test_utils.AssertTrue(t, err == nil, "GetContract succcessful")
	_ = json.Unmarshal(contractBytes, &contract)
	test_utils.AssertTrue(t, contract.OwnerServiceID == contract1.OwnerServiceID, "Got service ID correctly")
	mstub.MockTransactionEnd("14")

	// create another org user
	mstub.MockTransactionStart("15")
	stub = cached_stub.NewCachedStub(mstub)
	orgUser4 := CreateTestSolutionUser("orgUser4")
	orgUser4.Org = ownerOrg1.ID
	orgUser4Bytes, _ := json.Marshal(&orgUser4)
	_, err = RegisterUser(stub, ownerOrg1Caller, []string{string(orgUser4Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterUser to succeed")
	mstub.MockTransactionEnd("15")

	// put orgUser4 in owner org
	mstub.MockTransactionStart("16")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = PutUserInOrg(stub, ownerOrg1Caller, []string{orgUser4.ID, ownerOrg1.ID, "false"})
	test_utils.AssertTrue(t, err == nil, "Expected PutUserInOrg to succeed")
	mstub.MockTransactionEnd("16")

	// give service admin permission
	mstub.MockTransactionStart("17")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = AddPermissionServiceAdmin(stub, ownerOrg1Caller, []string{orgUser4.ID, "ownerService1"})
	test_utils.AssertTrue(t, err == nil, "Expected AddPermissionServiceAdmin to succeed")
	mstub.MockTransactionEnd("17")

	// get contract by ID as org user with service admin permission
	mstub.MockTransactionStart("14")
	stub = cached_stub.NewCachedStub(mstub)
	orgUser4Caller, _ := user_mgmt.GetUserData(stub, ownerOrg1Caller, orgUser4.ID, true, true)
	contractBytes, err = GetContract(stub, orgUser4Caller, []string{contract1.ContractID})
	test_utils.AssertTrue(t, err == nil, "GetContract succcessful")
	_ = json.Unmarshal(contractBytes, &contract)
	test_utils.AssertTrue(t, contract.OwnerServiceID == contract1.OwnerServiceID, "Got service ID correctly")
	mstub.MockTransactionEnd("14")
}

func TestContractLifeCycle(t *testing.T) {
	logger.SetLevel(shim.LogDebug)
	logger.Info("TestContractLifeCycle function called")

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

	// create requester org
	mstub.MockTransactionStart("1")
	stub = cached_stub.NewCachedStub(mstub)
	requesterOrg1 := test_utils.CreateTestGroup("requesterOrg1")
	requesterOrg1Bytes, _ := json.Marshal(&requesterOrg1)
	_, err = RegisterOrg(stub, requesterOrg1, []string{string(requesterOrg1Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterOrg to succeed")
	mstub.MockTransactionEnd("1")

	// create requester org datatype
	mstub.MockTransactionStart("2")
	stub = cached_stub.NewCachedStub(mstub)
	requesterOrg1Caller, _ := user_mgmt.GetUserData(stub, requesterOrg1, requesterOrg1.ID, true, true)
	reqOrgDatatype1 := Datatype{DatatypeID: "reqOrgDatatype1", Description: "reqOrgDatatype1"}
	reqOrgDatatype1Bytes, _ := json.Marshal(&reqOrgDatatype1)
	_, err = RegisterDatatype(stub, requesterOrg1Caller, []string{string(reqOrgDatatype1Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterDatatype to succeed")
	mstub.MockTransactionEnd("2")

	// create requester service
	mstub.MockTransactionStart("3")
	stub = cached_stub.NewCachedStub(mstub, true, true, true)
	reqServiceDatatype := GenerateServiceDatatypeForTesting("reqOrgDatatype1", "reqService1", []string{consentOptionWrite, consentOptionRead})
	reqService1 := GenerateServiceForTesting("reqService1", "requesterOrg1", []ServiceDatatype{reqServiceDatatype})
	reqService1Bytes, _ := json.Marshal(&reqService1)
	_, err = RegisterService(stub, requesterOrg1Caller, []string{string(reqService1Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterService to succeed")
	mstub.MockTransactionEnd("3")

	// create owner org
	mstub.MockTransactionStart("4")
	stub = cached_stub.NewCachedStub(mstub)
	ownerOrg1 := test_utils.CreateTestGroup("ownerOrg1")
	ownerOrg1Bytes, _ := json.Marshal(&ownerOrg1)
	_, err = RegisterOrg(stub, ownerOrg1, []string{string(ownerOrg1Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterOrg to succeed")
	mstub.MockTransactionEnd("4")

	// create owner org datatype
	mstub.MockTransactionStart("5")
	stub = cached_stub.NewCachedStub(mstub)
	ownerOrg1Caller, _ := user_mgmt.GetUserData(stub, ownerOrg1, ownerOrg1.ID, true, true)
	ownerOrgDatatype1 := Datatype{DatatypeID: "ownerOrgDatatype1", Description: "ownerOrgDatatype1"}
	ownerOrgDatatype1Bytes, _ := json.Marshal(&ownerOrgDatatype1)
	_, err = RegisterDatatype(stub, ownerOrg1Caller, []string{string(ownerOrgDatatype1Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterDatatype to succeed")
	mstub.MockTransactionEnd("5")

	// create owner service
	mstub.MockTransactionStart("6")
	stub = cached_stub.NewCachedStub(mstub, true, true, true)
	ownerServiceDatatype := GenerateServiceDatatypeForTesting("ownerOrgDatatype1", "ownerService1", []string{consentOptionWrite, consentOptionRead})
	ownerService1 := GenerateServiceForTesting("ownerService1", "ownerOrg1", []ServiceDatatype{ownerServiceDatatype})
	ownerService1Bytes, _ := json.Marshal(&ownerService1)
	_, err = RegisterService(stub, ownerOrg1Caller, []string{string(ownerService1Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterService to succeed")
	mstub.MockTransactionEnd("6")

	// owner service upload owner data
	mstub.MockTransactionStart("7")
	stub = cached_stub.NewCachedStub(mstub, true, true, true)
	ownerData := GenerateOwnerData("ownerService1", "ownerOrgDatatype1")
	ownerDataBytes, _ := json.Marshal(&ownerData)
	dataKey := test_utils.GenerateSymKey()
	dataKeyB64 := crypto.EncodeToB64String(dataKey)
	args := []string{string(ownerDataBytes), dataKeyB64}
	ownerService1Subgroup, _ := user_mgmt.GetUserData(stub, ownerOrg1Caller, "ownerService1", true, true)
	_, err = UploadOwnerData(stub, ownerService1Subgroup, args)
	test_utils.AssertTrue(t, err == nil, "Expected UploadOwnerData to succeed")
	mstub.MockTransactionEnd("7")

	// owner download data
	mstub.MockTransactionStart("8")
	stub = cached_stub.NewCachedStub(mstub)
	args = []string{"ownerService1", "ownerOrgDatatype1", "false", "0", "0", "1000", strconv.FormatInt(time.Now().Unix(), 10)}
	_, err = DownloadOwnerDataAsOwner(stub, ownerService1Subgroup, args)
	test_utils.AssertTrue(t, err == nil, "DownloadOwnerDataAsOwner succcessful")
	mstub.MockTransactionEnd("8")

	// create contract as requester default service admin
	mstub.MockTransactionStart("9")
	stub = cached_stub.NewCachedStub(mstub)
	contract1 := GenerateContractTest("contract1", "ownerOrg1", "ownerService1", "requesterOrg1", "reqService1")
	contract1.PaymentRequired = "yes"
	contract1Bytes, _ := json.Marshal(&contract1)
	contractKey := test_utils.GenerateSymKey()
	contractKeyB64 := crypto.EncodeToB64String(contractKey)
	args = []string{string(contract1Bytes), contractKeyB64}
	reqService1Subgroup, _ := user_mgmt.GetUserData(stub, requesterOrg1, "reqService1", true, true)
	_, err = CreateContract(stub, reqService1Subgroup, args)
	test_utils.AssertTrue(t, err == nil, "Expected CreateContract to succeed")
	mstub.MockTransactionEnd("9")

	// get contract by requester org default admin
	mstub.MockTransactionStart("10")
	stub = cached_stub.NewCachedStub(mstub)
	contractBytes, err := GetContract(stub, requesterOrg1Caller, []string{contract1.ContractID})
	test_utils.AssertTrue(t, err == nil, "GetContract succcessful")
	var contract = Contract{}
	_ = json.Unmarshal(contractBytes, &contract)
	test_utils.AssertTrue(t, contract.State == "requested", "Got contract state correctly")
	mstub.MockTransactionEnd("10")

	// add contract terms as requester
	mstub.MockTransactionStart("11")
	stub = cached_stub.NewCachedStub(mstub)
	terms := make(map[string]string)
	terms["owner update"] = "new terms"
	contractTermsBytes, err := json.Marshal(&terms)
	args = []string{contract1.ContractID, "terms", string(contractTermsBytes), strconv.FormatInt(time.Now().Unix(), 10)}
	_, err = AddContractDetail(stub, reqService1Subgroup, args)
	test_utils.AssertTrue(t, err == nil, "AddContractDetail succcessful")
	mstub.MockTransactionEnd("11")

	// add contract terms as owner
	mstub.MockTransactionStart("12")
	stub = cached_stub.NewCachedStub(mstub)
	terms["requester update"] = "new terms"
	contractTermsBytes, err = json.Marshal(&terms)
	args = []string{contract1.ContractID, "terms", string(contractTermsBytes), strconv.FormatInt(time.Now().Unix(), 10)}
	_, err = AddContractDetail(stub, ownerService1Subgroup, args)
	test_utils.AssertTrue(t, err == nil, "AddContractDetail succcessful")
	mstub.MockTransactionEnd("12")

	// requester sign contract
	mstub.MockTransactionStart("13")
	stub = cached_stub.NewCachedStub(mstub)
	args = []string{contract1.ContractID, "sign", "{}", strconv.FormatInt(time.Now().Unix(), 10)}
	_, err = AddContractDetail(stub, reqService1Subgroup, args)
	test_utils.AssertTrue(t, err == nil, "AddContractDetail succcessful")
	mstub.MockTransactionEnd("13")

	// requester pay contract
	mstub.MockTransactionStart("14")
	stub = cached_stub.NewCachedStub(mstub)
	args = []string{contract1.ContractID, "payment", "{}", strconv.FormatInt(time.Now().Unix(), 10)}
	_, err = AddContractDetail(stub, reqService1Subgroup, args)
	test_utils.AssertTrue(t, err == nil, "AddContractDetail succcessful")
	mstub.MockTransactionEnd("14")

	// owner verify payment
	mstub.MockTransactionStart("15")
	stub = cached_stub.NewCachedStub(mstub)
	args = []string{contract1.ContractID, "verify", "{}", strconv.FormatInt(time.Now().Unix(), 10)}
	_, err = AddContractDetail(stub, ownerService1Subgroup, args)
	test_utils.AssertTrue(t, err == nil, "AddContractDetail succcessful")
	mstub.MockTransactionEnd("15")

	// get contract
	mstub.MockTransactionStart("16")
	stub = cached_stub.NewCachedStub(mstub)
	contractBytes, err = GetContract(stub, requesterOrg1Caller, []string{contract1.ContractID})
	test_utils.AssertTrue(t, err == nil, "GetContract succcessful")
	_ = json.Unmarshal(contractBytes, &contract)
	test_utils.AssertTrue(t, contract.State == "paymentVerified", "Got contract state correctly")
	mstub.MockTransactionEnd("16")

	// owner give permission to download
	mstub.MockTransactionStart("17")
	stub = cached_stub.NewCachedStub(mstub)
	args = []string{contract1.ContractID, "10", strconv.FormatInt(time.Now().Unix(), 10), "ownerOrgDatatype1"}
	_, err = GivePermissionByContract(stub, ownerService1Subgroup, args)
	test_utils.AssertTrue(t, err == nil, "GivePermissionByContract succcessful")
	mstub.MockTransactionEnd("17")

	// requester download data
	mstub.MockTransactionStart("18")
	stub = cached_stub.NewCachedStub(mstub)
	args = []string{contract1.ContractID, "ownerOrgDatatype1", "false", "0", "0", "1000", strconv.FormatInt(time.Now().Unix(), 10)}
	_, err = DownloadOwnerDataAsRequester(stub, reqService1Subgroup, args)
	test_utils.AssertTrue(t, err == nil, "DownloadOwnerDataAsRequester succcessful")
	mstub.MockTransactionEnd("18")

	// create org user
	mstub.MockTransactionStart("19")
	stub = cached_stub.NewCachedStub(mstub)
	orgUser1 := CreateTestSolutionUser("orgUser1")
	orgUser1.Org = requesterOrg1.ID
	orgUser1Bytes, _ := json.Marshal(&orgUser1)
	_, err = RegisterUser(stub, requesterOrg1Caller, []string{string(orgUser1Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterUser to succeed")
	mstub.MockTransactionEnd("19")

	// put orgUser1 in requester org
	mstub.MockTransactionStart("20")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = PutUserInOrg(stub, requesterOrg1Caller, []string{orgUser1.ID, requesterOrg1.ID, "false"})
	test_utils.AssertTrue(t, err == nil, "Expected PutUserInOrg to succeed")
	mstub.MockTransactionEnd("20")

	// give service admin permission
	mstub.MockTransactionStart("21")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = AddPermissionServiceAdmin(stub, requesterOrg1Caller, []string{orgUser1.ID, "reqService1"})
	test_utils.AssertTrue(t, err == nil, "Expected AddPermissionServiceAdmin to succeed")
	mstub.MockTransactionEnd("21")

	// create org user
	mstub.MockTransactionStart("19")
	stub = cached_stub.NewCachedStub(mstub)
	orgUser2 := CreateTestSolutionUser("orgUser2")
	orgUser2.Org = ownerOrg1.ID
	orgUser2Bytes, _ := json.Marshal(&orgUser2)
	_, err = RegisterUser(stub, ownerOrg1Caller, []string{string(orgUser2Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterUser to succeed")
	mstub.MockTransactionEnd("19")

	// put orgUser2 in owner org
	mstub.MockTransactionStart("20")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = PutUserInOrg(stub, ownerOrg1Caller, []string{orgUser2.ID, ownerOrg1.ID, "false"})
	test_utils.AssertTrue(t, err == nil, "Expected PutUserInOrg to succeed")
	mstub.MockTransactionEnd("20")

	// give service admin permission
	mstub.MockTransactionStart("21")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = AddPermissionServiceAdmin(stub, ownerOrg1Caller, []string{orgUser2.ID, "ownerService1"})
	test_utils.AssertTrue(t, err == nil, "Expected AddPermissionServiceAdmin to succeed")
	mstub.MockTransactionEnd("21")

	// create contract as requester org user with service admin permission
	mstub.MockTransactionStart("22")
	stub = cached_stub.NewCachedStub(mstub)
	orgUser1Caller, _ := user_mgmt.GetUserData(stub, requesterOrg1Caller, orgUser1.ID, true, true)
	contract2 := GenerateContractTest("contract2", "ownerOrg1", "ownerService1", "requesterOrg1", "reqService1")
	contract2.PaymentRequired = "yes"
	contract2Bytes, _ := json.Marshal(&contract2)
	contractKey = test_utils.GenerateSymKey()
	contractKeyB64 = crypto.EncodeToB64String(contractKey)
	args = []string{string(contract2Bytes), contractKeyB64}
	_, err = CreateContract(stub, orgUser1Caller, args)
	test_utils.AssertTrue(t, err == nil, "CreateContract succcessful")
	_ = json.Unmarshal(contractBytes, &contract)
	test_utils.AssertTrue(t, contract.OwnerServiceID == contract2.OwnerServiceID, "Got service ID correctly")
	mstub.MockTransactionEnd("22")

	// get contract by requester org user with service admin permission
	mstub.MockTransactionStart("23")
	stub = cached_stub.NewCachedStub(mstub)
	contractBytes, err = GetContract(stub, orgUser1Caller, []string{contract2.ContractID})
	test_utils.AssertTrue(t, err == nil, "GetContract succcessful")
	_ = json.Unmarshal(contractBytes, &contract)
	test_utils.AssertTrue(t, contract.State == "requested", "Got contract state correctly")
	mstub.MockTransactionEnd("23")

	// add contract terms as requester org user with service admin permission
	mstub.MockTransactionStart("24")
	stub = cached_stub.NewCachedStub(mstub)
	terms = make(map[string]string)
	terms["owner update"] = "new terms"
	contractTermsBytes, err = json.Marshal(&terms)
	args = []string{contract2.ContractID, "terms", string(contractTermsBytes), strconv.FormatInt(time.Now().Unix(), 10)}
	_, err = AddContractDetail(stub, orgUser1Caller, args)
	test_utils.AssertTrue(t, err == nil, "AddContractDetail succcessful")
	mstub.MockTransactionEnd("24")

	// add contract terms as owner org user with service admin permission
	mstub.MockTransactionStart("25")
	stub = cached_stub.NewCachedStub(mstub)
	orgUser2Caller, _ := user_mgmt.GetUserData(stub, ownerOrg1Caller, orgUser2.ID, true, true)
	terms["requester update"] = "new terms"
	contractTermsBytes, err = json.Marshal(&terms)
	args = []string{contract2.ContractID, "terms", string(contractTermsBytes), strconv.FormatInt(time.Now().Unix(), 10)}
	_, err = AddContractDetail(stub, orgUser2Caller, args)
	test_utils.AssertTrue(t, err == nil, "AddContractDetail succcessful")
	mstub.MockTransactionEnd("25")

	// sign contract as requester org user with service admin permission
	mstub.MockTransactionStart("26a")
	stub = cached_stub.NewCachedStub(mstub)
	args = []string{contract2.ContractID, "sign", "{}", strconv.FormatInt(time.Now().Unix(), 10)}
	_, err = AddContractDetail(stub, orgUser1Caller, args)
	test_utils.AssertTrue(t, err == nil, "AddContractDetail succcessful")
	mstub.MockTransactionEnd("26a")

	// requester download data, should fail
	mstub.MockTransactionStart("26b")
	stub = cached_stub.NewCachedStub(mstub)
	args = []string{contract2.ContractID, "ownerOrgDatatype1", "false", "0", "0", "1000", strconv.FormatInt(time.Now().Unix(), 10)}
	_, err = DownloadOwnerDataAsRequester(stub, orgUser1Caller, args)
	test_utils.AssertTrue(t, err != nil, "DownloadOwnerDataAsRequester should fail")
	mstub.MockTransactionEnd("26b")

	// pay contract as requester org user with service admin permission
	mstub.MockTransactionStart("27")
	stub = cached_stub.NewCachedStub(mstub)
	args = []string{contract2.ContractID, "payment", "{}", strconv.FormatInt(time.Now().Unix(), 10)}
	_, err = AddContractDetail(stub, orgUser1Caller, args)
	test_utils.AssertTrue(t, err == nil, "AddContractDetail succcessful")
	mstub.MockTransactionEnd("27")

	// owner verify payment as owner org user with service admin permission
	mstub.MockTransactionStart("28")
	stub = cached_stub.NewCachedStub(mstub)
	args = []string{contract2.ContractID, "verify", "{}", strconv.FormatInt(time.Now().Unix(), 10)}
	_, err = AddContractDetail(stub, orgUser2Caller, args)
	test_utils.AssertTrue(t, err == nil, "AddContractDetail succcessful")
	mstub.MockTransactionEnd("28")

	// get contract as requester org user with service admin permission
	mstub.MockTransactionStart("29")
	stub = cached_stub.NewCachedStub(mstub)
	contractBytes, err = GetContract(stub, orgUser1Caller, []string{contract2.ContractID})
	test_utils.AssertTrue(t, err == nil, "GetContract succcessful")
	_ = json.Unmarshal(contractBytes, &contract)
	test_utils.AssertTrue(t, contract.State == "paymentVerified", "Got contract state correctly")
	mstub.MockTransactionEnd("29")

	// give permission to download as owner org user with service admin permission
	mstub.MockTransactionStart("30")
	stub = cached_stub.NewCachedStub(mstub)
	args = []string{contract2.ContractID, "10", strconv.FormatInt(time.Now().Unix(), 10), "ownerOrgDatatype1"}
	_, err = GivePermissionByContract(stub, orgUser2Caller, args)
	test_utils.AssertTrue(t, err == nil, "GivePermissionByContract succcessful")
	mstub.MockTransactionEnd("30")

	// requester download data
	mstub.MockTransactionStart("31")
	stub = cached_stub.NewCachedStub(mstub)
	args = []string{contract2.ContractID, "ownerOrgDatatype1", "false", "0", "0", "1000", strconv.FormatInt(time.Now().Unix(), 10)}
	_, err = DownloadOwnerDataAsRequester(stub, orgUser1Caller, args)
	test_utils.AssertTrue(t, err == nil, "DownloadOwnerDataAsRequester succcessful")
	mstub.MockTransactionEnd("31")

	// create org user
	mstub.MockTransactionStart("32")
	stub = cached_stub.NewCachedStub(mstub)
	orgUser3 := CreateTestSolutionUser("orgUser3")
	orgUser3.Org = requesterOrg1.ID
	orgUser3Bytes, _ := json.Marshal(&orgUser3)
	_, err = RegisterUser(stub, requesterOrg1Caller, []string{string(orgUser3Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterUser to succeed")
	mstub.MockTransactionEnd("32")

	// put orgUser3 in requester org
	mstub.MockTransactionStart("33")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = PutUserInOrg(stub, requesterOrg1Caller, []string{orgUser3.ID, requesterOrg1.ID, "false"})
	test_utils.AssertTrue(t, err == nil, "Expected PutUserInOrg to succeed")
	mstub.MockTransactionEnd("33")

	// give org admin permission
	mstub.MockTransactionStart("34")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = AddPermissionOrgAdmin(stub, requesterOrg1Caller, []string{orgUser3.ID, requesterOrg1.ID})
	test_utils.AssertTrue(t, err == nil, "Expected AddPermissionOrgAdmin to succeed")
	mstub.MockTransactionEnd("34")

	// create org user
	mstub.MockTransactionStart("35")
	stub = cached_stub.NewCachedStub(mstub)
	orgUser4 := CreateTestSolutionUser("orgUser4")
	orgUser4.Org = ownerOrg1.ID
	orgUser4Bytes, _ := json.Marshal(&orgUser4)
	_, err = RegisterUser(stub, ownerOrg1Caller, []string{string(orgUser4Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterUser to succeed")
	mstub.MockTransactionEnd("35")

	// put orgUser4 in owner org
	mstub.MockTransactionStart("36")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = PutUserInOrg(stub, ownerOrg1Caller, []string{orgUser4.ID, ownerOrg1.ID, "false"})
	test_utils.AssertTrue(t, err == nil, "Expected PutUserInOrg to succeed")
	mstub.MockTransactionEnd("36")

	// give org admin permission
	mstub.MockTransactionStart("37")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = AddPermissionOrgAdmin(stub, ownerOrg1Caller, []string{orgUser4.ID, ownerOrg1.ID})
	test_utils.AssertTrue(t, err == nil, "Expected AddPermissionOrgAdmin to succeed")
	mstub.MockTransactionEnd("37")

	// create contract as requester org user with org admin permission
	mstub.MockTransactionStart("38")
	stub = cached_stub.NewCachedStub(mstub)
	orgUser3Caller, _ := user_mgmt.GetUserData(stub, requesterOrg1Caller, orgUser3.ID, true, true)
	contract3 := GenerateContractTest("contract3", "ownerOrg1", "ownerService1", "requesterOrg1", "reqService1")
	contract3.PaymentRequired = "yes"
	contract3Bytes, _ := json.Marshal(&contract3)
	contractKey = test_utils.GenerateSymKey()
	contractKeyB64 = crypto.EncodeToB64String(contractKey)
	args = []string{string(contract3Bytes), contractKeyB64}
	_, err = CreateContract(stub, orgUser3Caller, args)
	test_utils.AssertTrue(t, err == nil, "CreateContract succcessful")
	_ = json.Unmarshal(contractBytes, &contract)
	test_utils.AssertTrue(t, contract.OwnerServiceID == contract3.OwnerServiceID, "Got service ID correctly")
	mstub.MockTransactionEnd("38")

	// get contract by requester org user with org admin permission
	mstub.MockTransactionStart("39a")
	stub = cached_stub.NewCachedStub(mstub)
	contractBytes, err = GetContract(stub, orgUser3Caller, []string{contract3.ContractID})
	test_utils.AssertTrue(t, err == nil, "GetContract succcessful")
	_ = json.Unmarshal(contractBytes, &contract)
	test_utils.AssertTrue(t, contract.State == "requested", "Got contract state correctly")
	mstub.MockTransactionEnd("39a")

	// requester download data, should fail
	mstub.MockTransactionStart("39b")
	stub = cached_stub.NewCachedStub(mstub)
	args = []string{contract3.ContractID, "ownerOrgDatatype1", "false", "0", "0", "1000", strconv.FormatInt(time.Now().Unix(), 10)}
	_, err = DownloadOwnerDataAsRequester(stub, orgUser3Caller, args)
	test_utils.AssertTrue(t, err != nil, "DownloadOwnerDataAsRequester should fail")
	mstub.MockTransactionEnd("39b")

	// add contract terms as requester org user with org admin permission
	mstub.MockTransactionStart("40")
	stub = cached_stub.NewCachedStub(mstub)
	terms = make(map[string]string)
	terms["owner update"] = "new terms"
	contractTermsBytes, err = json.Marshal(&terms)
	args = []string{contract3.ContractID, "terms", string(contractTermsBytes), strconv.FormatInt(time.Now().Unix(), 10)}
	_, err = AddContractDetail(stub, orgUser3Caller, args)
	test_utils.AssertTrue(t, err == nil, "AddContractDetail succcessful")
	mstub.MockTransactionEnd("40")

	// add contract terms as owner org user with org admin permission
	mstub.MockTransactionStart("41")
	stub = cached_stub.NewCachedStub(mstub)
	orgUser4Caller, _ := user_mgmt.GetUserData(stub, ownerOrg1Caller, orgUser4.ID, true, true)
	terms["requester update"] = "new terms"
	contractTermsBytes, err = json.Marshal(&terms)
	args = []string{contract3.ContractID, "terms", string(contractTermsBytes), strconv.FormatInt(time.Now().Unix(), 10)}
	_, err = AddContractDetail(stub, orgUser4Caller, args)
	test_utils.AssertTrue(t, err == nil, "AddContractDetail succcessful")
	mstub.MockTransactionEnd("41")

	// sign contract as requester org user with org admin permission
	mstub.MockTransactionStart("42")
	stub = cached_stub.NewCachedStub(mstub)
	args = []string{contract3.ContractID, "sign", "{}", strconv.FormatInt(time.Now().Unix(), 10)}
	_, err = AddContractDetail(stub, orgUser3Caller, args)
	test_utils.AssertTrue(t, err == nil, "AddContractDetail succcessful")
	mstub.MockTransactionEnd("42")

	// pay contract as requester org user with org admin permission
	mstub.MockTransactionStart("43")
	stub = cached_stub.NewCachedStub(mstub)
	args = []string{contract3.ContractID, "payment", "{}", strconv.FormatInt(time.Now().Unix(), 10)}
	_, err = AddContractDetail(stub, orgUser3Caller, args)
	test_utils.AssertTrue(t, err == nil, "AddContractDetail succcessful")
	mstub.MockTransactionEnd("43")

	// owner verify payment as owner org user with org admin permission
	mstub.MockTransactionStart("44")
	stub = cached_stub.NewCachedStub(mstub)
	args = []string{contract3.ContractID, "verify", "{}", strconv.FormatInt(time.Now().Unix(), 10)}
	_, err = AddContractDetail(stub, orgUser4Caller, args)
	test_utils.AssertTrue(t, err == nil, "AddContractDetail succcessful")
	mstub.MockTransactionEnd("44")

	// get contract as requester org user with org admin permission
	mstub.MockTransactionStart("45")
	stub = cached_stub.NewCachedStub(mstub)
	contractBytes, err = GetContract(stub, orgUser3Caller, []string{contract3.ContractID})
	test_utils.AssertTrue(t, err == nil, "GetContract succcessful")
	_ = json.Unmarshal(contractBytes, &contract)
	test_utils.AssertTrue(t, contract.State == "paymentVerified", "Got contract state correctly")
	mstub.MockTransactionEnd("45")

	// give permission to download as owner org user with org admin permission
	mstub.MockTransactionStart("46")
	stub = cached_stub.NewCachedStub(mstub)
	args = []string{contract3.ContractID, "10", strconv.FormatInt(time.Now().Unix(), 10), "ownerOrgDatatype1"}
	_, err = GivePermissionByContract(stub, orgUser4Caller, args)
	test_utils.AssertTrue(t, err == nil, "GivePermissionByContract succcessful")
	mstub.MockTransactionEnd("46")

	// requester download data
	mstub.MockTransactionStart("47")
	stub = cached_stub.NewCachedStub(mstub)
	args = []string{contract3.ContractID, "ownerOrgDatatype1", "false", "0", "0", "1000", strconv.FormatInt(time.Now().Unix(), 10)}
	_, err = DownloadOwnerDataAsRequester(stub, orgUser3Caller, args)
	test_utils.AssertTrue(t, err == nil, "DownloadOwnerDataAsRequester succcessful")
	mstub.MockTransactionEnd("47")
}

func TestContractMaxNumDownload(t *testing.T) {
	logger.SetLevel(shim.LogDebug)
	logger.Info("TestContractMaxNumDownload function called")

	// create a MockStub
	mstub := SetupIndexesAndGetStub(t)

	// register admin user
	mstub.MockTransactionStart("0")
	stub := cached_stub.NewCachedStub(mstub)
	init_common.Init(stub)
	err := SetupDataIndex(stub)
	systemAdmin := test_utils.CreateTestUser("systemAdmin")
	systemAdmin.Role = SOLUTION_ROLE_SYSTEM
	systemAdminBytes, _ := json.Marshal(&systemAdmin)
	_, err = RegisterUser(stub, systemAdmin, []string{string(systemAdminBytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterUser to succeed")
	mstub.MockTransactionEnd("0")

	// Register system datatypes
	mstub.MockTransactionStart("init")
	stub = cached_stub.NewCachedStub(mstub)
	RegisterSystemDatatypeTest(t, stub, systemAdmin)
	mstub.MockTransactionEnd("init")

	// create requester org
	mstub.MockTransactionStart("1")
	stub = cached_stub.NewCachedStub(mstub)
	requesterOrg1 := test_utils.CreateTestGroup("requesterOrg1")
	requesterOrg1Bytes, _ := json.Marshal(&requesterOrg1)
	_, err = RegisterOrg(stub, requesterOrg1, []string{string(requesterOrg1Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterOrg to succeed")
	mstub.MockTransactionEnd("1")

	// create requester org datatype
	mstub.MockTransactionStart("2")
	stub = cached_stub.NewCachedStub(mstub)
	requesterOrg1Caller, _ := user_mgmt.GetUserData(stub, requesterOrg1, requesterOrg1.ID, true, true)
	reqOrgDatatype1 := Datatype{DatatypeID: "reqOrgDatatype1", Description: "reqOrgDatatype1"}
	reqOrgDatatype1Bytes, _ := json.Marshal(&reqOrgDatatype1)
	_, err = RegisterDatatype(stub, requesterOrg1Caller, []string{string(reqOrgDatatype1Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterDatatype to succeed")
	mstub.MockTransactionEnd("2")

	// create requester service
	mstub.MockTransactionStart("3")
	stub = cached_stub.NewCachedStub(mstub, true, true, true)
	reqServiceDatatype := GenerateServiceDatatypeForTesting("reqOrgDatatype1", "reqService1", []string{consentOptionWrite, consentOptionRead})
	reqService1 := GenerateServiceForTesting("reqService1", "requesterOrg1", []ServiceDatatype{reqServiceDatatype})
	reqService1Bytes, _ := json.Marshal(&reqService1)
	_, err = RegisterService(stub, requesterOrg1Caller, []string{string(reqService1Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterService to succeed")
	mstub.MockTransactionEnd("3")

	// create owner org
	mstub.MockTransactionStart("4")
	stub = cached_stub.NewCachedStub(mstub)
	ownerOrg1 := test_utils.CreateTestGroup("ownerOrg1")
	ownerOrg1Bytes, _ := json.Marshal(&ownerOrg1)
	_, err = RegisterOrg(stub, ownerOrg1, []string{string(ownerOrg1Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterOrg to succeed")
	mstub.MockTransactionEnd("4")

	// create owner org datatype
	mstub.MockTransactionStart("5")
	stub = cached_stub.NewCachedStub(mstub)
	ownerOrg1Caller, _ := user_mgmt.GetUserData(stub, ownerOrg1, ownerOrg1.ID, true, true)
	ownerOrgDatatype1 := Datatype{DatatypeID: "ownerOrgDatatype1", Description: "ownerOrgDatatype1"}
	ownerOrgDatatype1Bytes, _ := json.Marshal(&ownerOrgDatatype1)
	_, err = RegisterDatatype(stub, ownerOrg1Caller, []string{string(ownerOrgDatatype1Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterDatatype to succeed")
	mstub.MockTransactionEnd("5")

	// create owner service
	mstub.MockTransactionStart("6")
	stub = cached_stub.NewCachedStub(mstub, true, true, true)
	ownerServiceDatatype := GenerateServiceDatatypeForTesting("ownerOrgDatatype1", "ownerService1", []string{consentOptionWrite, consentOptionRead})
	ownerService1 := GenerateServiceForTesting("ownerService1", "ownerOrg1", []ServiceDatatype{ownerServiceDatatype})
	ownerService1Bytes, _ := json.Marshal(&ownerService1)
	_, err = RegisterService(stub, ownerOrg1Caller, []string{string(ownerService1Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterService to succeed")
	mstub.MockTransactionEnd("6")

	// owner service upload owner data
	mstub.MockTransactionStart("7")
	stub = cached_stub.NewCachedStub(mstub, true, true, true)
	ownerData := GenerateOwnerData("ownerService1", "ownerOrgDatatype1")
	ownerDataBytes, _ := json.Marshal(&ownerData)
	dataKey := test_utils.GenerateSymKey()
	dataKeyB64 := crypto.EncodeToB64String(dataKey)
	args := []string{string(ownerDataBytes), dataKeyB64}
	ownerService1Subgroup, _ := user_mgmt.GetUserData(stub, ownerOrg1Caller, "ownerService1", true, true)
	_, err = UploadOwnerData(stub, ownerService1Subgroup, args)
	test_utils.AssertTrue(t, err == nil, "Expected UploadOwnerData to succeed")
	mstub.MockTransactionEnd("7")

	// owner download data
	mstub.MockTransactionStart("8")
	stub = cached_stub.NewCachedStub(mstub)
	args = []string{"ownerService1", "ownerOrgDatatype1", "false", "0", "0", "1000", strconv.FormatInt(time.Now().Unix(), 10)}
	_, err = DownloadOwnerDataAsOwner(stub, ownerService1Subgroup, args)
	test_utils.AssertTrue(t, err == nil, "DownloadOwnerDataAsOwner succcessful")
	mstub.MockTransactionEnd("8")

	// create contract as requester default service admin
	mstub.MockTransactionStart("9")
	stub = cached_stub.NewCachedStub(mstub)
	contract1 := GenerateContractTest("contract1", "ownerOrg1", "ownerService1", "requesterOrg1", "reqService1")
	contract1.PaymentRequired = "yes"
	contract1Bytes, _ := json.Marshal(&contract1)
	contractKey := test_utils.GenerateSymKey()
	contractKeyB64 := crypto.EncodeToB64String(contractKey)
	args = []string{string(contract1Bytes), contractKeyB64}
	reqService1Subgroup, _ := user_mgmt.GetUserData(stub, requesterOrg1, "reqService1", true, true)
	_, err = CreateContract(stub, reqService1Subgroup, args)
	test_utils.AssertTrue(t, err == nil, "Expected CreateContract to succeed")
	mstub.MockTransactionEnd("9")

	// get contract by requester org default admin
	mstub.MockTransactionStart("10")
	stub = cached_stub.NewCachedStub(mstub)
	contractBytes, err := GetContract(stub, requesterOrg1Caller, []string{contract1.ContractID})
	test_utils.AssertTrue(t, err == nil, "GetContract succcessful")
	var contract = Contract{}
	_ = json.Unmarshal(contractBytes, &contract)
	test_utils.AssertTrue(t, contract.State == "requested", "Got contract state correctly")
	mstub.MockTransactionEnd("10")

	// add contract terms as requester
	mstub.MockTransactionStart("11")
	stub = cached_stub.NewCachedStub(mstub)
	terms := make(map[string]string)
	terms["owner update"] = "new terms"
	contractTermsBytes, err := json.Marshal(&terms)
	args = []string{contract1.ContractID, "terms", string(contractTermsBytes), strconv.FormatInt(time.Now().Unix(), 10)}
	_, err = AddContractDetail(stub, reqService1Subgroup, args)
	test_utils.AssertTrue(t, err == nil, "AddContractDetail succcessful")
	mstub.MockTransactionEnd("11")

	// add contract terms as owner
	mstub.MockTransactionStart("12")
	stub = cached_stub.NewCachedStub(mstub)
	terms["requester update"] = "new terms"
	contractTermsBytes, err = json.Marshal(&terms)
	args = []string{contract1.ContractID, "terms", string(contractTermsBytes), strconv.FormatInt(time.Now().Unix(), 10)}
	_, err = AddContractDetail(stub, ownerService1Subgroup, args)
	test_utils.AssertTrue(t, err == nil, "AddContractDetail succcessful")
	mstub.MockTransactionEnd("12")

	// requester sign contract
	mstub.MockTransactionStart("13")
	stub = cached_stub.NewCachedStub(mstub)
	args = []string{contract1.ContractID, "sign", "{}", strconv.FormatInt(time.Now().Unix(), 10)}
	_, err = AddContractDetail(stub, reqService1Subgroup, args)
	test_utils.AssertTrue(t, err == nil, "AddContractDetail succcessful")
	mstub.MockTransactionEnd("13")

	// requester pay contract
	mstub.MockTransactionStart("14")
	stub = cached_stub.NewCachedStub(mstub)
	args = []string{contract1.ContractID, "payment", "{}", strconv.FormatInt(time.Now().Unix(), 10)}
	_, err = AddContractDetail(stub, reqService1Subgroup, args)
	test_utils.AssertTrue(t, err == nil, "AddContractDetail succcessful")
	mstub.MockTransactionEnd("14")

	// owner verify payment
	mstub.MockTransactionStart("15")
	stub = cached_stub.NewCachedStub(mstub)
	args = []string{contract1.ContractID, "verify", "{}", strconv.FormatInt(time.Now().Unix(), 10)}
	_, err = AddContractDetail(stub, ownerService1Subgroup, args)
	test_utils.AssertTrue(t, err == nil, "AddContractDetail succcessful")
	mstub.MockTransactionEnd("15")

	// get contract
	mstub.MockTransactionStart("16")
	stub = cached_stub.NewCachedStub(mstub)
	contractBytes, err = GetContract(stub, requesterOrg1Caller, []string{contract1.ContractID})
	test_utils.AssertTrue(t, err == nil, "GetContract succcessful")
	_ = json.Unmarshal(contractBytes, &contract)
	test_utils.AssertTrue(t, contract.State == "paymentVerified", "Got contract state correctly")
	mstub.MockTransactionEnd("16")

	// check no datatype sym key access
	mstub.MockTransactionStart("17")
	stub = cached_stub.NewCachedStub(mstub)
	datatypeSymKeyPath, err := GetDatatypeKeyPath(stub, reqService1Subgroup, ownerServiceDatatype.DatatypeID, contract.OwnerServiceID)
	test_utils.AssertTrue(t, err == nil, "GetDatatypeKeyPath succcessful")
	test_utils.AssertTrue(t, len(datatypeSymKeyPath) == 0, "Expected no datatypeSymKeyPath")
	mstub.MockTransactionEnd("17")

	// owner give permission to download (max 2 times)
	mstub.MockTransactionStart("18")
	stub = cached_stub.NewCachedStub(mstub)
	args = []string{contract1.ContractID, "2", strconv.FormatInt(time.Now().Unix(), 10), "ownerOrgDatatype1"}
	_, err = GivePermissionByContract(stub, ownerService1Subgroup, args)
	test_utils.AssertTrue(t, err == nil, "GivePermissionByContract succcessful")
	mstub.MockTransactionEnd("18")

	// check datatype sym key access given
	mstub.MockTransactionStart("19")
	stub = cached_stub.NewCachedStub(mstub)
	datatypeSymKeyPath, err = GetDatatypeKeyPath(stub, reqService1Subgroup, ownerServiceDatatype.DatatypeID, contract.OwnerServiceID)
	test_utils.AssertTrue(t, err == nil, "GetDatatypeKeyPath succcessful")
	test_utils.AssertTrue(t, len(datatypeSymKeyPath) != 0, "Expected datatypeSymKeyPath")
	mstub.MockTransactionEnd("19")

	// requester download data (1st time)
	mstub.MockTransactionStart("20")
	stub = cached_stub.NewCachedStub(mstub)
	args = []string{contract1.ContractID, "ownerOrgDatatype1", "false", "0", "0", "1000", strconv.FormatInt(time.Now().Unix(), 10)}
	downloadBytes, err := DownloadOwnerDataAsRequester(stub, reqService1Subgroup, args)
	test_utils.AssertTrue(t, err == nil, "DownloadOwnerDataAsRequester succcessful")
	var downloadResult = OwnerDataDownloadResult{}
	json.Unmarshal(downloadBytes, &downloadResult)
	mstub.MockTransactionEnd("20")

	// add contract detail download
	mstub.MockTransactionStart("21")
	stub = cached_stub.NewCachedStub(mstub)
	args = []string{contract1.ContractID, downloadResult.EncryptedContract, "ownerOrgDatatype1"}
	_, err = AddContractDetailDownload(stub, reqService1Subgroup, args)
	test_utils.AssertTrue(t, err == nil, "AddContractDetailDownload succcessful")
	mstub.MockTransactionEnd("21")

	// requester download data (2nd time)
	mstub.MockTransactionStart("22")
	stub = cached_stub.NewCachedStub(mstub)
	args = []string{contract1.ContractID, "ownerOrgDatatype1", "false", "0", "0", "1000", strconv.FormatInt(time.Now().Unix(), 10)}
	downloadBytes, err = DownloadOwnerDataAsRequester(stub, reqService1Subgroup, args)
	test_utils.AssertTrue(t, err == nil, "DownloadOwnerDataAsRequester succcessful")
	downloadResult = OwnerDataDownloadResult{}
	json.Unmarshal(downloadBytes, &downloadResult)
	mstub.MockTransactionEnd("22")

	// add contract detail download
	mstub.MockTransactionStart("23")
	stub = cached_stub.NewCachedStub(mstub)
	args = []string{contract1.ContractID, downloadResult.EncryptedContract, "ownerOrgDatatype1"}
	_, err = AddContractDetailDownload(stub, reqService1Subgroup, args)
	test_utils.AssertTrue(t, err == nil, "AddContractDetailDownload succcessful")
	mstub.MockTransactionEnd("23")

	// check datatype sym key access removed
	mstub.MockTransactionStart("24")
	stub = cached_stub.NewCachedStub(mstub)
	datatypeSymKeyPath, err = GetDatatypeKeyPath(stub, reqService1Subgroup, ownerServiceDatatype.DatatypeID, contract.OwnerServiceID)
	test_utils.AssertTrue(t, err == nil, "GetDatatypeKeyPath succcessful")
	test_utils.AssertTrue(t, len(datatypeSymKeyPath) == 0, "Expected no datatypeSymKeyPath")
	mstub.MockTransactionEnd("24")

	// requester attempts to download data (3rd time)
	mstub.MockTransactionStart("25")
	stub = cached_stub.NewCachedStub(mstub)
	args = []string{contract1.ContractID, "ownerOrgDatatype1", "false", "0", "0", "1000", strconv.FormatInt(time.Now().Unix(), 10)}
	_, err = DownloadOwnerDataAsRequester(stub, reqService1Subgroup, args)
	test_utils.AssertTrue(t, err != nil, "Expected DownloadOwnerDataAsRequester to fail")
	mstub.MockTransactionEnd("25")
}
