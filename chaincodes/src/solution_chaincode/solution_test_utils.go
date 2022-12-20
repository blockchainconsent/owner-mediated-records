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
	"common/bchcls/init_common"
	"common/bchcls/test_utils"
	"testing"
)

// SetupIndexesAndGetStub inits common for tests
func SetupIndexesAndGetStub(t *testing.T) *test_utils.NewMockStub {
	stub := test_utils.CreateNewMockStub(t)
	stub.MockTransactionStart("init")
	cstub := cached_stub.NewCachedStub(stub)
	init_common.Init(cstub)
	stub.MockTransactionEnd("init")
	return stub
}
