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
	"common/bchcls/utils"

	"github.com/pkg/errors"
)

// InitIndices initializes index tables
func InitIndices(stub cached_stub.CachedStubInterface) error {
	defer utils.ExitFnLog(utils.EnterFnLog())

	err := SetupServiceIndex(stub)
	if err != nil {
		err = errors.Wrap(err, "Failed to create service indices")
		logger.Error(err.Error())
		return err
	}

	err = SetupAuditPermissionIndex(stub)
	if err != nil {
		err = errors.Wrap(err, "Failed to create audit permission indices")
		logger.Error(err.Error())
		return err
	}

	err = SetupEnrollmentIndex(stub)
	if err != nil {
		err = errors.Wrap(err, "Failed to create enrollment indices")
		logger.Error(err.Error())
		return err
	}

	err = SetupDataIndex(stub)
	if err != nil {
		err = errors.Wrap(err, "Failed to create owner data indices")
		logger.Error(err.Error())
		return err
	}

	err = SetupContractIndex(stub)
	if err != nil {
		err = errors.Wrap(err, "Failed to create contract indices")
		logger.Error(err.Error())
		return err
	}

	return nil
}
