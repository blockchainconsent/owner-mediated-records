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

	"github.com/hyperledger/fabric/core/chaincode/shim"
)

// InitApp is called by main.Init(), which is called during chaincode Instantiation
func InitApp(stub cached_stub.CachedStubInterface, logLevel shim.LoggingLevel) error {
	SetLogLevel(logLevel)

	return InitIndices(stub)
}

// SetLogLevel set the app's logging level and is called during instantiation
func SetLogLevel(logLevel shim.LoggingLevel) {
	logger.SetLevel(logLevel)
	shim.SetLoggingLevel(logLevel)
	logger.Infof("Setting logging level to %v", logLevel)
}
