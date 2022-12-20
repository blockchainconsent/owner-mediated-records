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
	"common/bchcls/data_model"
	"common/bchcls/test_utils"
	"common/bchcls/user_mgmt"

	"encoding/json"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"testing"
)

func TestRegisterDatatype(t *testing.T) {
	logger.SetLevel(shim.LogDebug)
	logger.Info("TestRegisterDatatype function called")

	// create a MockStub
	mstub := SetupIndexesAndGetStub(t)

	// register admin user
	mstub.MockTransactionStart("t123")
	stub := cached_stub.NewCachedStub(mstub)
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
	org1Caller, _ := user_mgmt.GetUserData(stub, org1, org1.ID, true, true)
	datatype1 := Datatype{DatatypeID: "datatype1", Description: "datatype1 description"}
	datatype1Bytes, _ := json.Marshal(&datatype1)
	_, err = RegisterDatatype(stub, org1Caller, []string{string(datatype1Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterDatatype to succeed")
	mstub.MockTransactionEnd("t123")
}

func RegisterSystemDatatypeTest(t *testing.T, stub cached_stub.CachedStubInterface, caller data_model.User) {
	datatypeID := "ConsentTemplate"
	description := "ConsentTemplate datatype description"

	// caller is the datatype owner
	_, err := RegisterDatatypeWithParams(stub, caller, datatypeID, description)
	test_utils.AssertTrue(t, err == nil, "Expected RegisterDatatypeWithParams to succeed")
}
