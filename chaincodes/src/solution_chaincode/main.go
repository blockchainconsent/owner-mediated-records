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
	"github.com/hyperledger/fabric/core/chaincode/shim"
)

var logger = shim.NewLogger("owner-mediated-records")

func main() {
	logger.SetLevel(shim.LogDebug)
	shim.SetLoggingLevel(shim.LogDebug)

	logger.Info("---Starting Chaincode---")
	err := shim.Start(new(Chaincode))
	if err != nil {
		logger.Errorf("Error starting Chaincode: %v", err)
		panic(err)
	}
}
