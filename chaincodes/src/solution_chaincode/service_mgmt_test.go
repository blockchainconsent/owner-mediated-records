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
	"time"

	"github.com/hyperledger/fabric/core/chaincode/shim"
)

func GenerateServiceDatatypeForTesting(datatypeID string, serviceID string, access []string) ServiceDatatype {
	serviceDatatype := ServiceDatatype{}
	serviceDatatype.DatatypeID = datatypeID
	serviceDatatype.ServiceID = serviceID
	serviceDatatype.Access = access

	return serviceDatatype
}

func GenerateServiceForTesting(serviceID string, orgID string, serviceDatatypes []ServiceDatatype) map[string]interface{} {
	service := make(map[string]interface{})

	service["service_id"] = serviceID
	service["service_name"] = serviceID + " Name"
	service["datatypes"] = serviceDatatypes
	service["org_id"] = orgID
	service["email"] = serviceID + "@services.com"
	service["summary"] = serviceID + " summary"
	service["terms"] = make(map[string]interface{})
	service["payment_required"] = "yes"
	service["status"] = "active"
	data := make(map[string]interface{})
	service["solution_private_data"] = data
	service["create_date"] = time.Now().Unix()
	service["update_date"] = time.Now().Unix()

	service["is_group"] = true
	service["role"] = "org"

	privateKey := test_utils.GeneratePrivateKey()
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	service["private_key"] = base64.StdEncoding.EncodeToString(privateKeyBytes)

	publicKey := privateKey.Public().(*rsa.PublicKey)
	publicKeyBytes, _ := x509.MarshalPKIXPublicKey(publicKey)
	service["public_key"] = base64.StdEncoding.EncodeToString(publicKeyBytes)

	symKey := test_utils.GenerateSymKey()
	service["sym_key"] = base64.StdEncoding.EncodeToString(symKey)

	return service
}

func TestRegisterService(t *testing.T) {
	logger.SetLevel(shim.LogDebug)
	logger.Info("TestRegisterService function called")

	// create a MockStub
	mstub := SetupIndexesAndGetStub(t)

	// register admin user
	mstub.MockTransactionStart("t123")
	stub := cached_stub.NewCachedStub(mstub)
	init_common.Init(stub)
	err := SetupServiceIndex(stub)
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

	// get service to make sure it was registered
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	serviceBytes, err := GetService(stub, org1Caller, []string{"service1"})
	test_utils.AssertTrue(t, err == nil, "GetService succcessful")
	var service = Service{}
	_ = json.Unmarshal(serviceBytes, &service)
	test_utils.AssertTrue(t, service.ServiceName == "service1 Name", "Got service name correctly")
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

	// register service as orgUser1 without org admin permission
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub, true, true, true)
	orgUser1Caller, _ := user_mgmt.GetUserData(stub, org1Caller, orgUser1.ID, true, true)
	serviceDatatype2 := GenerateServiceDatatypeForTesting("datatype1", "service2", []string{consentOptionWrite, consentOptionRead})
	service2 := GenerateServiceForTesting("service2", "org1", []ServiceDatatype{serviceDatatype2})
	service.ServiceName = "service2 Name"
	service2Bytes, _ := json.Marshal(&service2)
	_, err = RegisterService(stub, orgUser1Caller, []string{string(service2Bytes)})
	test_utils.AssertTrue(t, err != nil, "Expected RegisterService to fail")
	mstub.MockTransactionEnd("t123")

	// give org admin permission
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = AddPermissionOrgAdmin(stub, org1Caller, []string{orgUser1.ID, org1.ID})
	test_utils.AssertTrue(t, err == nil, "Expected AddPermissionOrgAdmin to succeed")
	mstub.MockTransactionEnd("t123")

	// register service as orgUser1 with org admin permission
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub, true, true, true)
	orgUser1Caller, _ = user_mgmt.GetUserData(stub, org1Caller, orgUser1.ID, true, true)
	_, err = RegisterService(stub, orgUser1Caller, []string{string(service2Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterService to succeed")
	mstub.MockTransactionEnd("t123")

	// get service to make sure it was registered
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	serviceBytes, err = GetService(stub, org1Caller, []string{"service2"})
	test_utils.AssertTrue(t, err == nil, "GetService succcessful")
	_ = json.Unmarshal(serviceBytes, &service)
	test_utils.AssertTrue(t, service.ServiceName == "service2 Name", "Got service name correctly")
	mstub.MockTransactionEnd("t123")

	// remove org admin permission
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = RemovePermissionOrgAdmin(stub, org1Caller, []string{orgUser1.ID, org1.ID})
	test_utils.AssertTrue(t, err == nil, "Expected RemovePermissionOrgAdmin to succeed")
	mstub.MockTransactionEnd("t123")

	// register service as orgUser1 without org admin permission
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub, true, true, true)
	orgUser1Caller, _ = user_mgmt.GetUserData(stub, org1Caller, orgUser1.ID, true, true)
	serviceDatatype3 := GenerateServiceDatatypeForTesting("datatype1", "service3", []string{consentOptionWrite, consentOptionRead})
	service3 := GenerateServiceForTesting("service3", "org1", []ServiceDatatype{serviceDatatype3})
	service.ServiceName = "service3 Name"
	service3Bytes, _ := json.Marshal(&service3)
	_, err = RegisterService(stub, orgUser1Caller, []string{string(service3Bytes)})
	test_utils.AssertTrue(t, err != nil, "Expected RegisterService to fail")
	mstub.MockTransactionEnd("t123")
}

func TestRegisterService_OffChain(t *testing.T) {
	logger.SetLevel(shim.LogDebug)
	logger.Info("TestRegisterService_OffChain function called")

	// create a MockStub
	mstub := SetupIndexesAndGetStub(t)

	// register admin user
	mstub.MockTransactionStart("t123")
	stub := cached_stub.NewCachedStub(mstub)
	init_common.Init(stub)
	err := SetupServiceIndex(stub)
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
	serviceID := "service1"
	serviceDatatype1 := GenerateServiceDatatypeForTesting("datatype1", serviceID, []string{consentOptionWrite, consentOptionRead})
	service1 := GenerateServiceForTesting(serviceID, "org1", []ServiceDatatype{serviceDatatype1})
	service1Bytes, _ := json.Marshal(&service1)
	_, err = RegisterService(stub, org1Caller, []string{string(service1Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterService to succeed")
	mstub.MockTransactionEnd("t123")

	// get service to make sure it was registered
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	serviceBytes, err := GetService(stub, org1Caller, []string{serviceID})
	test_utils.AssertTrue(t, err == nil, "GetService succcessful")
	var service = Service{}
	_ = json.Unmarshal(serviceBytes, &service)
	test_utils.AssertTrue(t, service.ServiceName == "service1 Name", "Got service name correctly")
	mstub.MockTransactionEnd("t123")

	// verify data access for org1Caller
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	assetManager := asset_mgmt.GetAssetManager(stub, org1Caller)
	serviceAssetID := asset_mgmt.GetAssetId(ServiceAssetNamespace, serviceID)
	keyPath, err := GetKeyPath(stub, org1Caller, serviceAssetID)
	test_utils.AssertTrue(t, err == nil, "Expected GetKeyPath consent to succeed")
	serviceAssetKey, err := assetManager.GetAssetKey(serviceAssetID, keyPath)
	test_utils.AssertTrue(t, err == nil, "Expected GetAssetKey to succeed")
	serviceAsset, err := assetManager.GetAsset(serviceAssetID, serviceAssetKey)
	test_utils.AssertTrue(t, err == nil, "Expected GetAsset to succeed")
	mstub.MockTransactionEnd("t123")

	// org1Caller has access to public data
	servicePublicData := Service{}
	json.Unmarshal(serviceAsset.PublicData, &servicePublicData)
	test_utils.AssertTrue(t, servicePublicData.ServiceID != "", "Expected ServiceID")
	test_utils.AssertTrue(t, servicePublicData.ServiceName != "", "Expected ServiceName")
	test_utils.AssertTrue(t, len(servicePublicData.Datatypes) > 0, "Expected Datatypes")
	test_utils.AssertTrue(t, servicePublicData.OrgID != "", "Expected OrgID")
	test_utils.AssertTrue(t, servicePublicData.Summary != "", "Expected Summary")
	test_utils.AssertTrue(t, servicePublicData.Terms != nil, "Expected Terms")
	test_utils.AssertTrue(t, servicePublicData.PaymentRequired != "", "Expected PaymentRequired")
	test_utils.AssertTrue(t, servicePublicData.Status != "", "Expected Status")
	test_utils.AssertTrue(t, servicePublicData.CreateDate > 0, "Expected CreateDate")
	test_utils.AssertTrue(t, servicePublicData.UpdateDate > 0, "Expected UpdateDate")

	// org1Caller has access to private data
	servicePrivateData := Service{}
	json.Unmarshal(serviceAsset.PrivateData, &servicePrivateData)
	test_utils.AssertTrue(t, servicePrivateData.Email != "", "Expected Email")
	test_utils.AssertTrue(t, servicePrivateData.SolutionPrivateData != nil, "Expected SolutionPrivateData")

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
	keyPath, err = GetKeyPath(stub, unrelatedUser, serviceAssetID)
	test_utils.AssertTrue(t, err == nil, "Expected GetKeyPath consent to succeed")
	_, err = assetManager.GetAssetKey(serviceAssetID, keyPath)
	test_utils.AssertTrue(t, err != nil, "Expected GetAssetKey to fail")
	serviceAsset, err = assetManager.GetAsset(serviceAssetID, data_model.Key{})
	test_utils.AssertTrue(t, err == nil, "Expected GetAsset to succeed")
	mstub.MockTransactionEnd("t123")

	// unrelatedUser has access to public data
	servicePublicData = Service{}
	json.Unmarshal(serviceAsset.PublicData, &servicePublicData)
	test_utils.AssertTrue(t, servicePublicData.ServiceID != "", "Expected ServiceID")
	test_utils.AssertTrue(t, servicePublicData.ServiceName != "", "Expected ServiceName")
	test_utils.AssertTrue(t, len(servicePublicData.Datatypes) > 0, "Expected Datatypes")
	test_utils.AssertTrue(t, servicePublicData.OrgID != "", "Expected OrgID")
	test_utils.AssertTrue(t, servicePublicData.Summary != "", "Expected Summary")
	test_utils.AssertTrue(t, servicePublicData.Terms != nil, "Expected Terms")
	test_utils.AssertTrue(t, servicePublicData.PaymentRequired != "", "Expected PaymentRequired")
	test_utils.AssertTrue(t, servicePublicData.Status != "", "Expected Status")
	test_utils.AssertTrue(t, servicePublicData.CreateDate > 0, "Expected CreateDate")
	test_utils.AssertTrue(t, servicePublicData.UpdateDate > 0, "Expected UpdateDate")

	// unrelatedUser has no access to private data
	servicePrivateData = Service{}
	json.Unmarshal(serviceAsset.PrivateData, &servicePrivateData)
	test_utils.AssertTrue(t, servicePrivateData.Email == "", "Expected no Email")
	test_utils.AssertTrue(t, servicePrivateData.SolutionPrivateData == nil, "Expected no SolutionPrivateData")

	// remove datastore connection
	mstub.MockTransactionStart("t123")
	err = datastore_manager.DeleteDatastoreConnection(stub, systemAdmin, datastoreConnectionID)
	test_utils.AssertTrue(t, err == nil, "Expected DeleteDatastoreConnection to succeed")
	mstub.MockTransactionEnd("t123")

	// attempt to get service
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	serviceBytes, err = GetService(stub, org1Caller, []string{serviceID})
	test_utils.AssertTrue(t, err != nil, "Expected GetService to fail")
	mstub.MockTransactionEnd("t123")
}

func TestUpdateService(t *testing.T) {
	logger.SetLevel(shim.LogDebug)
	logger.Info("TestUpdateService function called")

	// create a MockStub
	mstub := SetupIndexesAndGetStub(t)

	// register admin user
	mstub.MockTransactionStart("t123")
	stub := cached_stub.NewCachedStub(mstub)
	init_common.Init(stub)
	err := SetupServiceIndex(stub)
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

	// register datatypes
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

	// register another datatype and update service with this datatype
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	datatype2 := Datatype{DatatypeID: "datatype2", Description: "datatype2"}
	datatype2Bytes, _ := json.Marshal(&datatype2)
	_, err = RegisterDatatype(stub, org1Caller, []string{string(datatype2Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterDatatype to succeed")
	mstub.MockTransactionEnd("t123")

	// update service
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	serviceDatatype2 := GenerateServiceDatatypeForTesting("datatype2", "service1", []string{consentOptionWrite})
	service1 = GenerateServiceForTesting("service1", "org1", []ServiceDatatype{serviceDatatype2})
	service1["service_name"] = "service1 new name"
	service1Bytes, _ = json.Marshal(&service1)
	_, err = UpdateService(stub, org1Caller, []string{string(service1Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected UpdateService to succeed")
	mstub.MockTransactionEnd("t123")

	// get service to make sure it was updated
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	serviceBytes, err := GetService(stub, org1Caller, []string{"service1"})
	test_utils.AssertTrue(t, err == nil, "GetService succcessful")
	var service = Service{}
	_ = json.Unmarshal(serviceBytes, &service)
	test_utils.AssertTrue(t, service.ServiceName == "service1 new name", "Got service name correctly")
	test_utils.AssertTrue(t, service.Datatypes[0].DatatypeID == "datatype2", "Got datatype correctly")
	mstub.MockTransactionEnd("t123")

	// update service as service1 default admin
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	service1 = GenerateServiceForTesting("service1", "org1", []ServiceDatatype{serviceDatatype2})
	service1["service_name"] = "service1 updated name"
	service1Bytes, _ = json.Marshal(&service1)
	service1Caller, _ := user_mgmt.GetUserData(stub, org1Caller, "service1", true, true)
	_, err = UpdateService(stub, service1Caller, []string{string(service1Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected UpdateService to succeed")
	mstub.MockTransactionEnd("t123")

	// get service to make sure it was updated
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	serviceBytes, err = GetService(stub, service1Caller, []string{"service1"})
	test_utils.AssertTrue(t, err == nil, "GetService succcessful")
	_ = json.Unmarshal(serviceBytes, &service)
	test_utils.AssertTrue(t, service.ServiceName == "service1 updated name", "Got service name correctly")
	test_utils.AssertTrue(t, service.Datatypes[0].DatatypeID == "datatype2", "Got datatype correctly")
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

	// update service as orgUser1 without admin permission
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	orgUser1Caller, _ := user_mgmt.GetUserData(stub, org1Caller, orgUser1.ID, true, true)
	service1 = GenerateServiceForTesting("service1", "org1", []ServiceDatatype{serviceDatatype2})
	service1["service_name"] = "service1 newest updated name"
	service1Bytes, _ = json.Marshal(&service1)
	_, err = UpdateService(stub, orgUser1Caller, []string{string(service1Bytes)})
	test_utils.AssertTrue(t, err != nil, "Expected UpdateService to fail")
	mstub.MockTransactionEnd("t123")

	// get service to make sure it was not updated
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	serviceBytes, err = GetService(stub, org1Caller, []string{"service1"})
	test_utils.AssertTrue(t, err == nil, "GetService succcessful")
	_ = json.Unmarshal(serviceBytes, &service)
	test_utils.AssertTrue(t, service.ServiceName == "service1 updated name", "Got service name correctly")
	test_utils.AssertTrue(t, service.Email != "", "Expected service email")
	mstub.MockTransactionEnd("t123")

	// get service (as org user)
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	serviceBytes, err = GetService(stub, orgUser1Caller, []string{"service1"})
	test_utils.AssertTrue(t, err == nil, "GetService succcessful")
	_ = json.Unmarshal(serviceBytes, &service)
	test_utils.AssertTrue(t, service.ServiceName != "", "Got service name correctly")
	test_utils.AssertTrue(t, service.Email == "", "Expected no service email")
	mstub.MockTransactionEnd("t123")

	// org admin give orgUser1 service admin permission
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = AddPermissionServiceAdmin(stub, org1Caller, []string{orgUser1.ID, "service1"})
	test_utils.AssertTrue(t, err == nil, "Expected AddPermissionServiceAdmin to succeed")
	mstub.MockTransactionEnd("t123")

	// update service as orgUser1 (service admin)
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	service1 = GenerateServiceForTesting("service1", "org1", []ServiceDatatype{serviceDatatype2})
	service1["service_name"] = "service1 newest updated name"
	service1Bytes, _ = json.Marshal(&service1)
	_, err = UpdateService(stub, orgUser1Caller, []string{string(service1Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected UpdateService to succeed")
	mstub.MockTransactionEnd("t123")

	// get service to make sure it was updated
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	serviceBytes, err = GetService(stub, orgUser1Caller, []string{"service1"})
	test_utils.AssertTrue(t, err == nil, "GetService succcessful")
	_ = json.Unmarshal(serviceBytes, &service)
	test_utils.AssertTrue(t, service.ServiceName == "service1 newest updated name", "Got service name correctly")
	test_utils.AssertTrue(t, service.Email != "", "Expected service email")
	mstub.MockTransactionEnd("t123")

	// create another org user for negative test
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	orgUser2 := CreateTestSolutionUser("orgUser2")
	orgUser2.Org = "org1"
	orgUser2Bytes, _ := json.Marshal(&orgUser2)
	_, err = RegisterUser(stub, org1Caller, []string{string(orgUser2Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterUser to succeed")
	mstub.MockTransactionEnd("t123")

	// update service as orgUser2 (who is not service admin)
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	orgUser2Caller, _ := user_mgmt.GetUserData(stub, org1Caller, orgUser2.ID, true, true)
	service1 = GenerateServiceForTesting("service1", "org1", []ServiceDatatype{serviceDatatype2})
	service1["service_name"] = "service1 name (should fail)"
	service1Bytes, _ = json.Marshal(&service1)
	_, err = UpdateService(stub, orgUser2Caller, []string{string(service1Bytes)})
	test_utils.AssertTrue(t, err != nil, "Expected UpdateService to fail")
	mstub.MockTransactionEnd("t123")

	// get service to make sure it was not updated
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	serviceBytes, err = GetService(stub, org1Caller, []string{"service1"})
	test_utils.AssertTrue(t, err == nil, "GetService succcessful")
	_ = json.Unmarshal(serviceBytes, &service)
	test_utils.AssertTrue(t, service.ServiceName == "service1 newest updated name", "Got service name correctly, same as before")
	mstub.MockTransactionEnd("t123")

	// org admin give orgUser2 org admin permission
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = AddPermissionOrgAdmin(stub, org1Caller, []string{orgUser2.ID, "org1"})
	test_utils.AssertTrue(t, err == nil, "Expected AddPermissionOrgAdmin to succeed")
	mstub.MockTransactionEnd("t123")

	// update service as orgUser2 (who is org admin)
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	orgUser2Caller, _ = user_mgmt.GetUserData(stub, org1Caller, orgUser2.ID, true, true)
	service1 = GenerateServiceForTesting("service1", "org1", []ServiceDatatype{serviceDatatype2})
	service1["service_name"] = "service1 name newest"
	service1Bytes, _ = json.Marshal(&service1)
	_, err = UpdateService(stub, orgUser2Caller, []string{string(service1Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected UpdateService to succeed")
	mstub.MockTransactionEnd("t123")

	// get service to make sure it was not updated
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	serviceBytes, err = GetService(stub, org1Caller, []string{"service1"})
	test_utils.AssertTrue(t, err == nil, "GetService succcessful")
	_ = json.Unmarshal(serviceBytes, &service)
	test_utils.AssertTrue(t, service.ServiceName == "service1 name newest", "Got service name correctly, updated")
	mstub.MockTransactionEnd("t123")
}

func TestGetService(t *testing.T) {
	logger.SetLevel(shim.LogDebug)
	logger.Info("TestGetService function called")

	// create a MockStub
	mstub := SetupIndexesAndGetStub(t)

	// register admin user
	mstub.MockTransactionStart("t123")
	stub := cached_stub.NewCachedStub(mstub)
	init_common.Init(stub)
	err := SetupServiceIndex(stub)
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

	// get service as default org admin to make sure it was registered
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	serviceBytes, err := GetService(stub, org1Caller, []string{"service1"})
	test_utils.AssertTrue(t, err == nil, "GetService succcessful")
	var service = Service{}
	_ = json.Unmarshal(serviceBytes, &service)
	test_utils.AssertTrue(t, service.ServiceName == "service1 Name", "Got service name correctly")
	test_utils.AssertTrue(t, service.SolutionPrivateData != nil, "Got service private data as expected")
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

	// get service1 as org user with no admin permission
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	orgUser1Caller, _ := user_mgmt.GetUserData(stub, org1Caller, orgUser1.ID, true, true)
	serviceBytes, err = GetService(stub, orgUser1Caller, []string{"service1"})
	test_utils.AssertTrue(t, err == nil, "GetService succcessful")
	_ = json.Unmarshal(serviceBytes, &service)
	// can get public data
	test_utils.AssertTrue(t, service.ServiceName == "service1 Name", "Got service name correctly")
	// but no private data
	test_utils.AssertTrue(t, service.SolutionPrivateData == nil, "Got service private data as expected")
	mstub.MockTransactionEnd("t123")

	// register service as orgUser1 without org admin permission
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub, true, true, true)
	serviceDatatype2 := GenerateServiceDatatypeForTesting("datatype1", "service2", []string{consentOptionWrite, consentOptionRead})
	service2 := GenerateServiceForTesting("service2", "org1", []ServiceDatatype{serviceDatatype2})
	service.ServiceName = "service2 Name"
	service2Bytes, _ := json.Marshal(&service2)
	_, err = RegisterService(stub, orgUser1Caller, []string{string(service2Bytes)})
	test_utils.AssertTrue(t, err != nil, "Expected RegisterService to fail")
	mstub.MockTransactionEnd("t123")

	// give org admin permission
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = AddPermissionOrgAdmin(stub, org1Caller, []string{orgUser1.ID, org1.ID})
	test_utils.AssertTrue(t, err == nil, "Expected AddPermissionOrgAdmin to succeed")
	mstub.MockTransactionEnd("t123")

	// register service as orgUser1 with org admin permission
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub, true, true, true)
	orgUser1Caller, _ = user_mgmt.GetUserData(stub, org1Caller, orgUser1.ID, true, true)
	_, err = RegisterService(stub, orgUser1Caller, []string{string(service2Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterService to succeed")
	mstub.MockTransactionEnd("t123")

	// get service to make sure it was registered
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	serviceBytes, err = GetService(stub, org1Caller, []string{"service2"})
	test_utils.AssertTrue(t, err == nil, "GetService succcessful")
	_ = json.Unmarshal(serviceBytes, &service)
	test_utils.AssertTrue(t, service.ServiceName == "service2 Name", "Got service name correctly")
	test_utils.AssertTrue(t, service.SolutionPrivateData != nil, "Got service private data as expected")
	mstub.MockTransactionEnd("t123")

	// get service as orgUser1
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	serviceBytes, err = GetService(stub, orgUser1Caller, []string{"service2"})
	test_utils.AssertTrue(t, err == nil, "GetService succcessful")
	_ = json.Unmarshal(serviceBytes, &service)
	test_utils.AssertTrue(t, service.ServiceName == "service2 Name", "Got service name correctly")
	test_utils.AssertTrue(t, service.SolutionPrivateData != nil, "Got service private data as expected")
	mstub.MockTransactionEnd("t123")

	// remove org admin permission
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = RemovePermissionOrgAdmin(stub, org1Caller, []string{orgUser1.ID, org1.ID})
	test_utils.AssertTrue(t, err == nil, "Expected RemovePermissionOrgAdmin to succeed")
	mstub.MockTransactionEnd("t123")

	// get service as orgUser1 (now without org admin permission)
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	orgUser1Caller, _ = user_mgmt.GetUserData(stub, org1Caller, orgUser1.ID, true, true)
	serviceBytes, err = GetService(stub, orgUser1Caller, []string{"service2"})
	test_utils.AssertTrue(t, err == nil, "GetService succcessful")
	_ = json.Unmarshal(serviceBytes, &service)
	test_utils.AssertTrue(t, service.ServiceName == "service2 Name", "Got service name correctly")
	test_utils.AssertTrue(t, service.SolutionPrivateData == nil, "Got service private data as expected")
	mstub.MockTransactionEnd("t123")

	// create an orgUser2
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

	// get service as orgUser2
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	orgUser2Caller, _ := user_mgmt.GetUserData(stub, org1Caller, orgUser2.ID, true, true)
	serviceBytes, err = GetService(stub, orgUser2Caller, []string{"service2"})
	test_utils.AssertTrue(t, err == nil, "GetService succcessful")
	_ = json.Unmarshal(serviceBytes, &service)
	test_utils.AssertTrue(t, service.ServiceName == "service2 Name", "Got service name correctly")
	test_utils.AssertTrue(t, service.SolutionPrivateData == nil, "Got service private data as expected")
	mstub.MockTransactionEnd("t123")

	// get service as default service admin
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	service1Caller, _ := user_mgmt.GetUserData(stub, org1Caller, "service1", true, true)
	serviceBytes, err = GetService(stub, service1Caller, []string{"service1"})
	test_utils.AssertTrue(t, err == nil, "GetService succcessful")
	_ = json.Unmarshal(serviceBytes, &service)
	test_utils.AssertTrue(t, service.ServiceName == "service1 Name", "Got service name correctly")
	test_utils.AssertTrue(t, service.SolutionPrivateData != nil, "Got service private data as expected")
	mstub.MockTransactionEnd("t123")

	// org admin give orgUser2 service admin permission
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = AddPermissionServiceAdmin(stub, org1Caller, []string{orgUser2.ID, "service1"})
	test_utils.AssertTrue(t, err == nil, "Expected AddPermissionServiceAdmin to succeed")
	mstub.MockTransactionEnd("t123")

	// get service1 as orgUser2
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	orgUser2Caller, _ = user_mgmt.GetUserData(stub, org1Caller, orgUser2.ID, true, true)
	serviceBytes, err = GetService(stub, orgUser2Caller, []string{"service1"})
	test_utils.AssertTrue(t, err == nil, "GetService succcessful")
	_ = json.Unmarshal(serviceBytes, &service)
	test_utils.AssertTrue(t, service.ServiceName == "service1 Name", "Got service name correctly")
	test_utils.AssertTrue(t, service.SolutionPrivateData != nil, "Got service private data as expected")
	mstub.MockTransactionEnd("t123")

	// get service2 as orgUser2
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	orgUser2Caller, _ = user_mgmt.GetUserData(stub, org1Caller, orgUser2.ID, true, true)
	serviceBytes, err = GetService(stub, orgUser2Caller, []string{"service2"})
	test_utils.AssertTrue(t, err == nil, "GetService succcessful")
	_ = json.Unmarshal(serviceBytes, &service)
	test_utils.AssertTrue(t, service.ServiceName == "service2 Name", "Got service name correctly")
	test_utils.AssertTrue(t, service.SolutionPrivateData == nil, "Got service private data as expected")
	mstub.MockTransactionEnd("t123")
}

func TestGetServiceInternal(t *testing.T) {
	logger.SetLevel(shim.LogDebug)
	logger.Info("TestGetServiceInternal function called")

	// create a MockStub
	mstub := SetupIndexesAndGetStub(t)

	// register admin user
	mstub.MockTransactionStart("t123")
	stub := cached_stub.NewCachedStub(mstub)
	init_common.Init(stub)
	err := SetupServiceIndex(stub)
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

	// get service as default org admin to make sure it was registered
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	service, err := GetServiceInternal(stub, org1Caller, "service1", true)
	test_utils.AssertTrue(t, err == nil, "GetServiceInternal succcessful")
	test_utils.AssertTrue(t, service.ServiceName == "service1 Name", "Got service name correctly")
	test_utils.AssertTrue(t, service.SolutionPrivateData != nil, "Got service private data as expected")
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

	// get service1 as org user with no admin permission
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	orgUser1Caller, _ := user_mgmt.GetUserData(stub, org1Caller, orgUser1.ID, true, true)
	service, err = GetServiceInternal(stub, orgUser1Caller, "service1", true)
	test_utils.AssertTrue(t, err == nil, "GetServiceInternal succcessful")
	// can get public data
	test_utils.AssertTrue(t, service.ServiceName == "service1 Name", "Got service name correctly")
	// but no private data
	test_utils.AssertTrue(t, service.SolutionPrivateData == nil, "Got service private data as expected")
	mstub.MockTransactionEnd("t123")

	// register service as orgUser1 without org admin permission
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub, true, true, true)
	serviceDatatype2 := GenerateServiceDatatypeForTesting("datatype1", "service2", []string{consentOptionWrite, consentOptionRead})
	service2 := GenerateServiceForTesting("service2", "org1", []ServiceDatatype{serviceDatatype2})
	service.ServiceName = "service2 Name"
	service2Bytes, _ := json.Marshal(&service2)
	_, err = RegisterService(stub, orgUser1Caller, []string{string(service2Bytes)})
	test_utils.AssertTrue(t, err != nil, "Expected RegisterService to fail")
	mstub.MockTransactionEnd("t123")

	// give org admin permission
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = AddPermissionOrgAdmin(stub, org1Caller, []string{orgUser1.ID, org1.ID})
	test_utils.AssertTrue(t, err == nil, "Expected AddPermissionOrgAdmin to succeed")
	mstub.MockTransactionEnd("t123")

	// register service as orgUser1 with org admin permission
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub, true, true, true)
	orgUser1Caller, _ = user_mgmt.GetUserData(stub, org1Caller, orgUser1.ID, true, true)
	_, err = RegisterService(stub, orgUser1Caller, []string{string(service2Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterService to succeed")
	mstub.MockTransactionEnd("t123")

	// get service to make sure it was registered
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	service, err = GetServiceInternal(stub, org1Caller, "service2", true)
	test_utils.AssertTrue(t, err == nil, "GetServiceInternal succcessful")
	test_utils.AssertTrue(t, service.ServiceName == "service2 Name", "Got service name correctly")
	test_utils.AssertTrue(t, service.SolutionPrivateData != nil, "Got service private data as expected")
	mstub.MockTransactionEnd("t123")

	// get service as orgUser1
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	service, err = GetServiceInternal(stub, orgUser1Caller, "service2", true)
	test_utils.AssertTrue(t, err == nil, "GetServiceInternal succcessful")
	test_utils.AssertTrue(t, service.ServiceName == "service2 Name", "Got service name correctly")
	test_utils.AssertTrue(t, service.SolutionPrivateData != nil, "Got service private data as expected")
	mstub.MockTransactionEnd("t123")

	// remove org admin permission
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = RemovePermissionOrgAdmin(stub, org1Caller, []string{orgUser1.ID, org1.ID})
	test_utils.AssertTrue(t, err == nil, "Expected RemovePermissionOrgAdmin to succeed")
	mstub.MockTransactionEnd("t123")

	// get service as orgUser1 (now without org admin permission)
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	orgUser1Caller, _ = user_mgmt.GetUserData(stub, org1Caller, orgUser1.ID, true, true)
	service, err = GetServiceInternal(stub, orgUser1Caller, "service2", true)
	test_utils.AssertTrue(t, err == nil, "GetServiceInternal succcessful")
	test_utils.AssertTrue(t, service.ServiceName == "service2 Name", "Got service name correctly")
	test_utils.AssertTrue(t, service.SolutionPrivateData == nil, "Got service private data as expected")
	mstub.MockTransactionEnd("t123")

	// create an orgUser2
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

	// get service as orgUser2
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	orgUser2Caller, _ := user_mgmt.GetUserData(stub, org1Caller, orgUser2.ID, true, true)
	service, err = GetServiceInternal(stub, orgUser2Caller, "service2", true)
	test_utils.AssertTrue(t, err == nil, "GetServiceInternal succcessful")
	test_utils.AssertTrue(t, service.ServiceName == "service2 Name", "Got service name correctly")
	test_utils.AssertTrue(t, service.SolutionPrivateData == nil, "Got service private data as expected")
	mstub.MockTransactionEnd("t123")

	// get service as default service admin
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	service1Caller, _ := user_mgmt.GetUserData(stub, org1Caller, "service1", true, true)
	service, err = GetServiceInternal(stub, service1Caller, "service1", true)
	test_utils.AssertTrue(t, err == nil, "GetServiceInternal succcessful")
	test_utils.AssertTrue(t, service.ServiceName == "service1 Name", "Got service name correctly")
	test_utils.AssertTrue(t, service.SolutionPrivateData != nil, "Got service private data as expected")
	mstub.MockTransactionEnd("t123")

	// org admin give orgUser2 service admin permission
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = AddPermissionServiceAdmin(stub, org1Caller, []string{orgUser2.ID, "service1"})
	test_utils.AssertTrue(t, err == nil, "Expected AddPermissionServiceAdmin to succeed")
	mstub.MockTransactionEnd("t123")

	// get service1 as orgUser2
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	orgUser2Caller, _ = user_mgmt.GetUserData(stub, org1Caller, orgUser2.ID, true, true)
	service, err = GetServiceInternal(stub, orgUser2Caller, "service1", true)
	test_utils.AssertTrue(t, err == nil, "GetServiceInternal succcessful")
	test_utils.AssertTrue(t, service.ServiceName == "service1 Name", "Got service name correctly")
	test_utils.AssertTrue(t, service.SolutionPrivateData != nil, "Got service private data as expected")
	mstub.MockTransactionEnd("t123")

	// get service2 as orgUser2
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	orgUser2Caller, _ = user_mgmt.GetUserData(stub, org1Caller, orgUser2.ID, true, true)
	service, err = GetServiceInternal(stub, orgUser2Caller, "service2", true)
	test_utils.AssertTrue(t, err == nil, "GetServiceInternal succcessful")
	test_utils.AssertTrue(t, service.ServiceName == "service2 Name", "Got service name correctly")
	test_utils.AssertTrue(t, service.SolutionPrivateData == nil, "Got service private data as expected")
	mstub.MockTransactionEnd("t123")
}

func TestAddDatatypeToService(t *testing.T) {
	logger.SetLevel(shim.LogDebug)
	logger.Info("TestAddDatatypeToService function called")

	// create a MockStub
	mstub := SetupIndexesAndGetStub(t)

	// register admin user
	mstub.MockTransactionStart("t123")
	stub := cached_stub.NewCachedStub(mstub)
	init_common.Init(stub)
	err := SetupServiceIndex(stub)
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

	// get service to make sure it was registered
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	serviceBytes, err := GetService(stub, org1Caller, []string{"service1"})
	test_utils.AssertTrue(t, err == nil, "GetService succcessful")
	var service = Service{}
	_ = json.Unmarshal(serviceBytes, &service)
	test_utils.AssertTrue(t, service.ServiceName == "service1 Name", "Got service name correctly")
	mstub.MockTransactionEnd("t123")

	// register another datatype and add this datatype to service
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	datatype2 := Datatype{DatatypeID: "datatype2", Description: "datatype2"}
	datatype2Bytes, _ := json.Marshal(&datatype2)
	_, err = RegisterDatatype(stub, org1Caller, []string{string(datatype2Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterDatatype to succeed")
	mstub.MockTransactionEnd("t123")

	// add datatype to service as org admin
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	serviceDatatype2 := GenerateServiceDatatypeForTesting("datatype2", "service1", []string{consentOptionWrite})
	serviceDatatype2Bytes, _ := json.Marshal(&serviceDatatype2)
	_, err = AddDatatypeToService(stub, org1Caller, []string{"service1", string(serviceDatatype2Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected AddDatatypeToService to succeed")
	mstub.MockTransactionEnd("t123")

	// get service to make sure it was updated
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	serviceBytes, err = GetService(stub, org1Caller, []string{"service1"})
	test_utils.AssertTrue(t, err == nil, "GetService succcessful")
	_ = json.Unmarshal(serviceBytes, &service)
	test_utils.AssertTrue(t, service.Datatypes[0].DatatypeID == "datatype1", "Got datatype1 correctly")
	test_utils.AssertTrue(t, service.Datatypes[1].DatatypeID == "datatype2", "Got datatype2 correctly")
	mstub.MockTransactionEnd("t123")

	// create an org user for negative test
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	orgUser1 := CreateTestSolutionUser("orgUser1")
	orgUser1.Org = "org1"
	orgUser1Bytes, _ := json.Marshal(&orgUser1)
	_, err = RegisterUser(stub, org1Caller, []string{string(orgUser1Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterUser to succeed")
	mstub.MockTransactionEnd("t123")

	// put org user in org
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = PutUserInOrg(stub, org1Caller, []string{orgUser1.ID, org1.ID, "false"})
	test_utils.AssertTrue(t, err == nil, "Expected PutUserInOrg to succeed")
	mstub.MockTransactionEnd("t123")

	// register another datatype
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	datatype3 := Datatype{DatatypeID: "datatype3", Description: "datatype3"}
	datatype3Bytes, _ := json.Marshal(&datatype3)
	_, err = RegisterDatatype(stub, org1Caller, []string{string(datatype3Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterDatatype to succeed")
	mstub.MockTransactionEnd("t123")

	// add datatype to service as orgUser1 without service admin permission, should fail
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	orgUser1Caller, _ := user_mgmt.GetUserData(stub, org1Caller, orgUser1.ID, true, true)
	serviceDatatype3 := GenerateServiceDatatypeForTesting("datatype3", "service1", []string{consentOptionRead})
	serviceDatatype3Bytes, _ := json.Marshal(&serviceDatatype3)
	_, err = AddDatatypeToService(stub, orgUser1Caller, []string{"service1", string(serviceDatatype3Bytes)})
	test_utils.AssertTrue(t, err != nil, "Expected AddDatatypeToService to fail")
	mstub.MockTransactionEnd("t123")

	// get service
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	serviceBytes, err = GetService(stub, org1Caller, []string{"service1"})
	test_utils.AssertTrue(t, err == nil, "GetService succcessful")
	_ = json.Unmarshal(serviceBytes, &service)
	test_utils.AssertTrue(t, len(service.Datatypes) == 2, "Got datatypes as expected")
	mstub.MockTransactionEnd("t123")

	// org admin give service admin permission to org user
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = AddPermissionServiceAdmin(stub, org1Caller, []string{orgUser1.ID, "service1"})
	test_utils.AssertTrue(t, err == nil, "Expected AddPermissionServiceAdmin to succeed")
	mstub.MockTransactionEnd("t123")

	// add datatype to service as org user
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = AddDatatypeToService(stub, orgUser1Caller, []string{"service1", string(serviceDatatype3Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected AddDatatypeToService to succeed")
	mstub.MockTransactionEnd("t123")

	// get service
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	serviceBytes, err = GetService(stub, org1Caller, []string{"service1"})
	test_utils.AssertTrue(t, err == nil, "GetService succcessful")
	_ = json.Unmarshal(serviceBytes, &service)
	test_utils.AssertTrue(t, len(service.Datatypes) == 3, "Got datatypes as expected")
	mstub.MockTransactionEnd("t123")

	// create orgUser2
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	orgUser2 := CreateTestSolutionUser("orgUser2")
	orgUser2.Org = "org1"
	orgUser2Bytes, _ := json.Marshal(&orgUser2)
	_, err = RegisterUser(stub, org1Caller, []string{string(orgUser2Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterUser to succeed")
	mstub.MockTransactionEnd("t123")

	// put orgUser2 in org
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = PutUserInOrg(stub, org1Caller, []string{orgUser2.ID, org1.ID, "false"})
	test_utils.AssertTrue(t, err == nil, "Expected PutUserInOrg to succeed")
	mstub.MockTransactionEnd("t123")

	// org admin give org admin permission to orgUser2
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = AddPermissionOrgAdmin(stub, org1Caller, []string{orgUser2.ID, "org1"})
	test_utils.AssertTrue(t, err == nil, "Expected AddPermissionOrgAdmin to succeed")
	mstub.MockTransactionEnd("t123")

	// register another datatype
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	datatype4 := Datatype{DatatypeID: "datatype4", Description: "datatype4"}
	datatype4Bytes, _ := json.Marshal(&datatype4)
	_, err = RegisterDatatype(stub, org1Caller, []string{string(datatype4Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterDatatype to succeed")
	mstub.MockTransactionEnd("t123")

	// add datatype to service as orgUser2
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	orgUser2Caller, _ := user_mgmt.GetUserData(stub, org1Caller, orgUser2.ID, true, true)
	serviceDatatype4 := GenerateServiceDatatypeForTesting("datatype4", "service1", []string{consentOptionRead})
	serviceDatatype4Bytes, _ := json.Marshal(&serviceDatatype4)
	_, err = AddDatatypeToService(stub, orgUser2Caller, []string{"service1", string(serviceDatatype4Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected AddDatatypeToService to succeed")
	mstub.MockTransactionEnd("t123")

	// get service
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	serviceBytes, err = GetService(stub, org1Caller, []string{"service1"})
	test_utils.AssertTrue(t, err == nil, "GetService succcessful")
	_ = json.Unmarshal(serviceBytes, &service)
	test_utils.AssertTrue(t, len(service.Datatypes) == 4, "Got datatypes as expected")
	mstub.MockTransactionEnd("t123")
}

func TestRemoveDatatypeFromService(t *testing.T) {
	logger.SetLevel(shim.LogDebug)
	logger.Info("TestRemoveDatatypeFromService function called")

	// create a MockStub
	mstub := SetupIndexesAndGetStub(t)

	// register admin user
	mstub.MockTransactionStart("t123")
	stub := cached_stub.NewCachedStub(mstub)
	init_common.Init(stub)
	err := SetupServiceIndex(stub)
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

	// get service to make sure it was registered
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	serviceBytes, err := GetService(stub, org1Caller, []string{"service1"})
	test_utils.AssertTrue(t, err == nil, "GetService succcessful")
	var service = Service{}
	_ = json.Unmarshal(serviceBytes, &service)
	test_utils.AssertTrue(t, service.ServiceName == "service1 Name", "Got service name correctly")
	mstub.MockTransactionEnd("t123")

	// remove datatype from service
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = RemoveDatatypeFromService(stub, org1Caller, []string{"service1", "datatype1"})
	test_utils.AssertTrue(t, err == nil, "Expected RemoveDatatypeFromService to succeed")
	mstub.MockTransactionEnd("t123")

	// get service to make sure it was updated
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	serviceBytes, err = GetService(stub, org1Caller, []string{"service1"})
	test_utils.AssertTrue(t, err == nil, "GetService succcessful")
	_ = json.Unmarshal(serviceBytes, &service)
	test_utils.AssertTrue(t, len(service.Datatypes) == 0, "Removal worked correctly")
	mstub.MockTransactionEnd("t123")

	// create an org user for negative test
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	orgUser1 := CreateTestSolutionUser("orgUser1")
	orgUser1.Org = "org1"
	orgUser1Bytes, _ := json.Marshal(&orgUser1)
	_, err = RegisterUser(stub, org1Caller, []string{string(orgUser1Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterUser to succeed")
	mstub.MockTransactionEnd("t123")

	// register another datatype
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	datatype2 := Datatype{DatatypeID: "datatype2", Description: "datatype3"}
	datatype2Bytes, _ := json.Marshal(&datatype2)
	_, err = RegisterDatatype(stub, org1Caller, []string{string(datatype2Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterDatatype to succeed")
	mstub.MockTransactionEnd("t123")

	// add datatype2 to service, now service has one datatype
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	serviceDatatype2 := GenerateServiceDatatypeForTesting("datatype2", "service1", []string{consentOptionWrite})
	serviceDatatype2Bytes, _ := json.Marshal(&serviceDatatype2)
	_, err = AddDatatypeToService(stub, org1Caller, []string{"service1", string(serviceDatatype2Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected AddDatatypeToService to succeed")
	mstub.MockTransactionEnd("t123")

	// get service to make sure it was updated
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	serviceBytes, err = GetService(stub, org1Caller, []string{"service1"})
	test_utils.AssertTrue(t, err == nil, "GetService succcessful")
	_ = json.Unmarshal(serviceBytes, &service)
	test_utils.AssertTrue(t, len(service.Datatypes) == 1, "Got datatypes as expected")
	mstub.MockTransactionEnd("t123")

	// register one more datatype
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	datatype3 := Datatype{DatatypeID: "datatype3", Description: "datatype3"}
	datatype3Bytes, _ := json.Marshal(&datatype3)
	_, err = RegisterDatatype(stub, org1Caller, []string{string(datatype3Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected RegisterDatatype to succeed")
	mstub.MockTransactionEnd("t123")

	// add datatypes to service
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	serviceDatatype3 := GenerateServiceDatatypeForTesting("datatype3", "service1", []string{consentOptionRead})
	serviceDatatype3Bytes, _ := json.Marshal(&serviceDatatype3)
	_, err = AddDatatypeToService(stub, org1Caller, []string{"service1", string(serviceDatatype3Bytes)})
	test_utils.AssertTrue(t, err == nil, "Expected AddDatatypeToService to fail")
	mstub.MockTransactionEnd("t123")

	// get service
	mstub.MockTransactionStart("t124")
	stub = cached_stub.NewCachedStub(mstub)
	serviceBytes, err = GetService(stub, org1Caller, []string{"service1"})
	test_utils.AssertTrue(t, err == nil, "GetService succcessful")
	_ = json.Unmarshal(serviceBytes, &service)
	test_utils.AssertTrue(t, len(service.Datatypes) == 2, "Got datatypes as expected")
	mstub.MockTransactionEnd("t124")

	// remove datatype from service as org user
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	orgUser1Caller, _ := user_mgmt.GetUserData(stub, org1Caller, orgUser1.ID, true, true)
	_, err = RemoveDatatypeFromService(stub, orgUser1Caller, []string{"service1", "datatype2"})
	test_utils.AssertTrue(t, err != nil, "Expected RemoveDatatypeFromService to fail")
	mstub.MockTransactionEnd("t123")

	// get service
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	serviceBytes, err = GetService(stub, org1Caller, []string{"service1"})
	test_utils.AssertTrue(t, err == nil, "GetService succcessful")
	_ = json.Unmarshal(serviceBytes, &service)
	test_utils.AssertTrue(t, len(service.Datatypes) == 2, "Got 3 datatypes correctly")
	mstub.MockTransactionEnd("t123")

	// org admin give service admin permission to org user
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = AddPermissionServiceAdmin(stub, org1Caller, []string{orgUser1.ID, "service1"})
	test_utils.AssertTrue(t, err == nil, "Expected AddPermissionServiceAdmin to succeed")
	mstub.MockTransactionEnd("t123")

	// remove datatype from service as org user
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = RemoveDatatypeFromService(stub, orgUser1Caller, []string{"service1", "datatype3"})
	test_utils.AssertTrue(t, err == nil, "Expected RemoveDatatypeFromService to succeed")
	mstub.MockTransactionEnd("t123")

	// get service
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	serviceBytes, err = GetService(stub, org1Caller, []string{"service1"})
	test_utils.AssertTrue(t, err == nil, "GetService succcessful")
	_ = json.Unmarshal(serviceBytes, &service)
	test_utils.AssertTrue(t, len(service.Datatypes) == 1, "Got datatypes as expected")
	mstub.MockTransactionEnd("t123")
}

func TestGetServicesOfOrg(t *testing.T) {
	logger.SetLevel(shim.LogDebug)
	logger.Info("TestGetServicesOfOrg function called")

	// create a MockStub
	mstub := SetupIndexesAndGetStub(t)

	// register admin user
	mstub.MockTransactionStart("t123")
	stub := cached_stub.NewCachedStub(mstub)
	init_common.Init(stub)
	err := SetupServiceIndex(stub)
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

	//  register service1
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

	// get services of org as default org admin
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	servicesBytes, err := GetServicesOfOrg(stub, org1Caller, []string{"org1"})
	test_utils.AssertTrue(t, err == nil, "Expected GetServicesOfOrg to succeed")
	var services = []Service{}
	_ = json.Unmarshal(servicesBytes, &services)
	test_utils.AssertTrue(t, services[0].ServiceName == "service1 Name", "Got service 1 name correctly")
	test_utils.AssertTrue(t, services[1].ServiceName == "service2 Name", "Got service 2 name correctly")
	test_utils.AssertTrue(t, services[0].SolutionPrivateData != nil, "Got service private data")
	test_utils.AssertTrue(t, services[1].SolutionPrivateData != nil, "Got service private data")
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

	// get services of org as orgUser1
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	orgUser1Caller, _ := user_mgmt.GetUserData(stub, org1Caller, orgUser1.ID, true, true)
	servicesBytes, err = GetServicesOfOrg(stub, orgUser1Caller, []string{"org1"})
	test_utils.AssertTrue(t, err == nil, "Expected GetServicesOfOrg to succeed")
	_ = json.Unmarshal(servicesBytes, &services)
	test_utils.AssertTrue(t, services[0].ServiceName == "service1 Name", "Got service 1 name correctly")
	test_utils.AssertTrue(t, services[1].ServiceName == "service2 Name", "Got service 2 name correctly")
	test_utils.AssertTrue(t, services[0].SolutionPrivateData == nil, "Did not get service private data")
	test_utils.AssertTrue(t, services[1].SolutionPrivateData == nil, "Did not get service private data")
	mstub.MockTransactionEnd("t123")

	// give org admin permission
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	_, err = AddPermissionOrgAdmin(stub, org1Caller, []string{orgUser1.ID, org1.ID})
	test_utils.AssertTrue(t, err == nil, "Expected AddPermissionOrgAdmin to succeed")
	mstub.MockTransactionEnd("t123")

	// get service private data
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	orgUser1Caller, _ = user_mgmt.GetUserData(stub, org1Caller, orgUser1.ID, true, true)
	service, err := GetServiceInternal(stub, orgUser1Caller, "service1", true)
	test_utils.AssertTrue(t, err == nil, "Expected GetServiceInternal to succeed")
	test_utils.AssertTrue(t, service.SolutionPrivateData != nil, "Expected GetServiceInternal to succeed")
	mstub.MockTransactionEnd("t123")

	// get services of org as orgUser1
	mstub.MockTransactionStart("t123")
	stub = cached_stub.NewCachedStub(mstub)
	orgUser1Caller, _ = user_mgmt.GetUserData(stub, org1Caller, orgUser1.ID, true, true)
	servicesBytes, err = GetServicesOfOrg(stub, orgUser1Caller, []string{"org1"})
	test_utils.AssertTrue(t, err == nil, "Expected GetServicesOfOrg to succeed")
	_ = json.Unmarshal(servicesBytes, &services)
	test_utils.AssertTrue(t, services[0].ServiceName == "service1 Name", "Got service 1 name correctly")
	test_utils.AssertTrue(t, services[1].ServiceName == "service2 Name", "Got service 2 name correctly")
	test_utils.AssertTrue(t, services[0].SolutionPrivateData != nil, "Got service private data")
	test_utils.AssertTrue(t, services[1].SolutionPrivateData != nil, "Got service private data")
	mstub.MockTransactionEnd("t123")
}
