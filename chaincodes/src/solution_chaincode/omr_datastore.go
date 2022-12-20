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
	"common/bchcls/custom_errors"
	"common/bchcls/data_model"
	"common/bchcls/datastore"
	"common/bchcls/datastore/datastore_manager"
	"common/bchcls/init_common"
	"common/bchcls/utils"

	"github.com/pkg/errors"

	"net/url"
)

const setupDatastoreArgsLength = 5
const dsConnectionKey = "OMR.DatastoreConnectionID"

// SetupDatastore initializes an offchain-datastore
func SetupDatastore(stub cached_stub.CachedStubInterface, caller data_model.User, args []string) error {
	if len(args) != setupDatastoreArgsLength {
		customErr := &custom_errors.LengthCheckingError{Type: "SetupDatastore arguments length"}
		logger.Errorf(customErr.Error())
		return errors.New(customErr.Error())
	}

	dsConnectionID := args[0]
	username := args[1]
	password := args[2]
	database := args[3]
	host := args[4]

	if utils.IsStringEmpty(dsConnectionID) {
		errMsg := "Datastore ConnectionID cannot be empty"
		logger.Error(errMsg)
		return errors.New(errMsg)
	}

	switch dsConnectionID {
	case datastore.DEFAULT_LEDGER_DATASTORE_ID:
		logger.Info("Skipping datastore init - save to Ledger")
		return nil
	case datastore.DEFAULT_CLOUDANT_DATASTORE_ID:
		logger.Info("Instantiating default Cloudant datastore")
		return setupDefaultDatastore(stub, username, password, database, host)
	default:
		logger.Info("Instantiating custom datastore: " + dsConnectionID)
		return setupCustomDatastore(stub, caller, dsConnectionID, username, password, database, host)
	}
}

func setupDefaultDatastore(stub cached_stub.CachedStubInterface, username, password, database, host string) error {
	// init default datastore
	_, err := init_common.InitDatastore(stub, username, password, database, host)
	if err != nil {
		errMsg := "Failed to InitDatastore"
		logger.Errorf("%v: %v", errMsg, err)
		return errors.Wrap(err, errMsg)
	}

	// verify setup
	_, err = datastore_manager.GetDatastoreImpl(stub, datastore.DEFAULT_CLOUDANT_DATASTORE_ID)
	if err != nil {
		errMsg := "Failed to instantiate default Cloudant datastore"
		logger.Errorf("%v: %v", errMsg, err)
		return errors.Wrap(err, errMsg)
	}

	// save connectionID to ledger, so we know where to store and retrieve assets
	err = stub.PutState(dsConnectionKey, []byte(datastore.DEFAULT_CLOUDANT_DATASTORE_ID))
	if err != nil {
		customErr := &custom_errors.PutLedgerError{LedgerKey: dsConnectionKey}
		logger.Errorf("%v: %v", customErr, err)
		return errors.Wrap(err, customErr.Error())
	}

	return nil
}

func setupCustomDatastore(stub cached_stub.CachedStubInterface, caller data_model.User, datastoreID, username, password, database, host string) error {
	params := url.Values{}
	params.Add("username", username)
	params.Add("password", password)
	params.Add("database", database)
	params.Add("host", host)
	params.Add("create_database", "true")

	dsConnection := datastore.DatastoreConnection{
		ID:         datastoreID,
		Type:       datastore.DATASTORE_TYPE_DEFAULT_CLOUDANT,
		ConnectStr: params.Encode(),
	}

	// save custom datastore
	err := datastore_manager.PutDatastoreConnection(stub, caller, dsConnection)
	if err != nil {
		errMsg := "Failed to PutDatastoreConnection"
		logger.Errorf("%v: %v", errMsg, err)
		return errors.Wrap(err, errMsg)
	}

	// verify by instantiating datastore
	_, err = datastore_manager.GetDatastoreImpl(stub, datastoreID)
	if err != nil {
		errMsg := "Failed to instantiate custom datastore"
		logger.Errorf("%v: %v", errMsg, err)
		return errors.Wrap(err, errMsg)
	}

	// save connectionID to ledger, so we know where to store and retrieve assets
	err = stub.PutState(dsConnectionKey, []byte(datastoreID))
	if err != nil {
		customErr := &custom_errors.PutLedgerError{LedgerKey: dsConnectionKey}
		logger.Errorf("%v: %v", customErr, err)
		return errors.Wrap(err, customErr.Error())
	}

	return nil
}

// GetActiveConnectionID returns the connectionID of the datastore currently in use
func GetActiveConnectionID(stub cached_stub.CachedStubInterface) (string, error) {
	idBytes, err := stub.GetState(dsConnectionKey)
	if err != nil {
		customErr := &custom_errors.GetLedgerError{LedgerKey: dsConnectionKey, LedgerItem: "DatastoreConnectionID"}
		logger.Errorf("%v: %v", customErr, err)
		return "", errors.Wrap(err, customErr.Error())
	}

	// empty connectionID means no datastore was setup
	if len(idBytes) == 0 {
		errMsg := "No active ConnectionID found"
		logger.Error(errMsg)
		return "", nil
	}

	return string(idBytes), nil
}
