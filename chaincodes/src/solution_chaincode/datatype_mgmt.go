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
	"common/bchcls/custom_errors"
	"common/bchcls/data_model"
	"common/bchcls/datatype"
	"common/bchcls/datatype/datatype_interface"
	"common/bchcls/utils"
	"encoding/json"

	"github.com/pkg/errors"
)

// Datatype object
type Datatype struct {
	DatatypeID  string `json:"datatype_id"`
	Description string `json:"description"`
}

// RegisterDatatype registers a new datatype by calling Common's RegisterDatatype function
// Only org admins and system admins can create datatype
//
// args = [ datatypeBytes ]
func RegisterDatatype(stub cached_stub.CachedStubInterface, caller data_model.User, args []string) ([]byte, error) {
	defer utils.ExitFnLog(utils.EnterFnLog())
	logger.Debugf("args: %v", args)

	if len(args) != 1 {
		customErr := &custom_errors.LengthCheckingError{Type: "RegisterDatatype arguments length"}
		logger.Errorf(customErr.Error())
		return nil, errors.New(customErr.Error())
	}

	// ==============================================================
	// Validation
	// ==============================================================
	var datatypeOMR = Datatype{}
	datatypeOMRBytes := []byte(args[0])
	err := json.Unmarshal(datatypeOMRBytes, &datatypeOMR)
	if err != nil {
		customErr := &custom_errors.UnmarshalError{Type: "datatypeOMR"}
		logger.Errorf("%v: %v", customErr, err)
		return nil, errors.Wrap(err, customErr.Error())
	}

	if utils.IsStringEmpty(datatypeOMR.DatatypeID) {
		customErr := &custom_errors.LengthCheckingError{Type: "datatypeOMR.DatatypeID"}
		logger.Errorf(customErr.Error())
		return nil, errors.WithStack(customErr)
	}

	solutionCaller := convertToSolutionUser(caller)
	if !solutionCaller.SolutionInfo.IsOrgAdmin && !caller.IsSystemAdmin() {
		logger.Errorf("Caller is not org admin or system admin")
		return nil, errors.New("Caller is not org admin or system admin")
	}

	// ==============================================================
	// Register datatype in Common SDK
	// ==============================================================
	_, err = datatype.RegisterDatatypeWithParams(stub, datatypeOMR.DatatypeID, datatypeOMR.Description, true, datatype.ROOT_DATATYPE_ID)
	if err != nil {
		errMsg := "Failed to register datatype using SDK"
		logger.Errorf("%v: %v", errMsg, err)
		return nil, errors.Wrap(err, errMsg)
	}

	return nil, nil
}

// RegisterDatatypeWithParams registers a new datatype by calling Common's RegisterDatatypeWithParams function
// Only org admins and system admins can create datatype
func RegisterDatatypeWithParams(stub cached_stub.CachedStubInterface, caller data_model.User, datatypeID string, description string) ([]byte, error) {
	defer utils.ExitFnLog(utils.EnterFnLog())

	// ==============================================================
	// Validation
	// ==============================================================
	if utils.IsStringEmpty(datatypeID) {
		customErr := &custom_errors.LengthCheckingError{Type: "datatypeID"}
		logger.Errorf(customErr.Error())
		return nil, errors.WithStack(customErr)
	}

	solutionCaller := convertToSolutionUser(caller)
	if !solutionCaller.SolutionInfo.IsOrgAdmin && !caller.IsSystemAdmin() {
		logger.Errorf("Caller is not org admin or system admin")
		return nil, errors.New("Caller is not org admin or system admin")
	}

	// ==============================================================
	// Register datatype in Common SDK
	// ==============================================================
	_, err := datatype.RegisterDatatypeWithParams(stub, datatypeID, description, true, datatype.ROOT_DATATYPE_ID)
	if err != nil {
		errMsg := "Failed to register datatype using SDK"
		logger.Errorf("%v: %v", errMsg, err)
		return nil, errors.Wrap(err, errMsg)
	}

	return nil, nil
}

// GetDatatype returns a datatype object given datatype ID
// args = [ datatypeID ]
// returns empty datatype if no datatype with this ID is found
func GetDatatype(stub cached_stub.CachedStubInterface, caller data_model.User, args []string) ([]byte, error) {
	defer utils.ExitFnLog(utils.EnterFnLog())
	logger.Debugf("args: %v", args)

	if len(args) != 1 {
		customErr := &custom_errors.LengthCheckingError{Type: "GetDatatype arguments length"}
		logger.Errorf(customErr.Error())
		return nil, errors.New(customErr.Error())
	}

	datatypeID := args[0]
	if utils.IsStringEmpty(datatypeID) {
		customErr := &custom_errors.LengthCheckingError{Type: "datatypeID"}
		logger.Errorf(customErr.Error())
		return nil, errors.WithStack(customErr)
	}

	datatypeCommon, err := datatype.GetDatatypeWithParams(stub, datatypeID)
	if err != nil {
		errMsg := "Failed to get datatype"
		logger.Errorf("%v: %v", errMsg, err)
		return nil, errors.Wrap(err, errMsg)
	}

	datatypeOMR := convertDatatypeInterfaceToDatatypeOMR(datatypeCommon)
	return json.Marshal(datatypeOMR)
}

// GetDatatypeWithParams returns a datatype object given datatype ID
// returns empty datatype if no datatype with this ID is found
func GetDatatypeWithParams(stub cached_stub.CachedStubInterface, caller data_model.User, datatypeID string) (Datatype, error) {
	defer utils.ExitFnLog(utils.EnterFnLog())

	if utils.IsStringEmpty(datatypeID) {
		customErr := &custom_errors.LengthCheckingError{Type: "datatypeID"}
		logger.Errorf(customErr.Error())
		return Datatype{}, errors.WithStack(customErr)
	}

	datatypeCommon, err := datatype.GetDatatypeWithParams(stub, datatypeID)
	if err != nil {
		errMsg := "Failed to get datatype"
		logger.Errorf("%v: %v", errMsg, err)
		return Datatype{}, errors.Wrap(err, errMsg)
	}

	datatypeOMR := convertDatatypeInterfaceToDatatypeOMR(datatypeCommon)
	return datatypeOMR, nil
}

// UpdateDatatypeDescription updates an existing datatype's description by calling Common's UpdateDatatype function
// Only org admins and system admins can update datatype
//
// args = [ datatype ]
// datatype is existing datatype with new description
func UpdateDatatypeDescription(stub cached_stub.CachedStubInterface, caller data_model.User, args []string) ([]byte, error) {
	defer utils.ExitFnLog(utils.EnterFnLog())
	logger.Debugf("args: %v", args)

	if len(args) != 1 {
		customErr := &custom_errors.LengthCheckingError{Type: "UpdateDatatypeDescription arguments length"}
		logger.Errorf(customErr.Error())
		return nil, errors.New(customErr.Error())
	}

	// ==============================================================
	// Validation
	// ==============================================================
	var datatypeOMR = Datatype{}
	datatypeOMRBytes := []byte(args[0])
	err := json.Unmarshal(datatypeOMRBytes, &datatypeOMR)
	if err != nil {
		customErr := &custom_errors.UnmarshalError{Type: "datatypeOMR"}
		logger.Errorf("%v: %v", customErr, err)
		return nil, errors.Wrap(err, customErr.Error())
	}

	if utils.IsStringEmpty(datatypeOMR.DatatypeID) {
		customErr := &custom_errors.LengthCheckingError{Type: "datatypeOMR.DatatypeID"}
		logger.Errorf(customErr.Error())
		return nil, errors.WithStack(customErr)
	}

	if utils.IsStringEmpty(datatypeOMR.Description) {
		customErr := &custom_errors.LengthCheckingError{Type: "datatypeOMR.Description"}
		logger.Errorf(customErr.Error())
		return nil, errors.WithStack(customErr)
	}

	solutionCaller := convertToSolutionUser(caller)
	if !solutionCaller.SolutionInfo.IsOrgAdmin && !caller.IsSystemAdmin() {
		logger.Errorf("Caller is not org admin or system admin")
		return nil, errors.New("Caller is not org admin or system admin")
	}

	// ==============================================================
	// Update datatype in Common SDK
	// ==============================================================
	datatypeCommon := convertDatatypeOMRToCommonDatatype(datatypeOMR)
	datatypeCommonBytes, err := json.Marshal(&datatypeCommon)
	if err != nil {
		customErr := &custom_errors.MarshalError{Type: "Datatype [Common]"}
		logger.Errorf("%v: %v", customErr, err)
		return nil, errors.Wrap(err, customErr.Error())
	}

	_, err = datatype.UpdateDatatype(stub, caller, []string{string(datatypeCommonBytes)})
	if err != nil {
		errMsg := "Failed to update datatype using SDK"
		logger.Errorf("%v: %v", errMsg, err)
		return nil, errors.Wrap(err, errMsg)
	}

	return nil, nil
}

// GetAllDatatypes gets datatypes in the Blockchain (except for ROOT)
// args = [ ]
func GetAllDatatypes(stub cached_stub.CachedStubInterface, caller data_model.User, args []string) ([]byte, error) {
	defer utils.ExitFnLog(utils.EnterFnLog())
	logger.Debugf("args: %v", args)

	allDatatypesBytes, err := datatype.GetAllDatatypes(stub, caller, args)

	allDatatypes := []data_model.Datatype{}
	err = json.Unmarshal(allDatatypesBytes, &allDatatypes)
	if err != nil {
		customErr := &custom_errors.UnmarshalError{Type: "allDatatypes"}
		logger.Errorf("%v: %v", customErr, err)
		return nil, errors.Wrap(err, customErr.Error())
	}

	allOMRDatatypes := []Datatype{}
	for _, datatype := range allDatatypes {
		omrDatatype := convertDatatypeCommonToDatatypeOMR(datatype)
		allOMRDatatypes = append(allOMRDatatypes, omrDatatype)
	}

	return json.Marshal(&allOMRDatatypes)
}

// GetDatatypeSymKey returns a datatype sym key for the given datatypeID and ownerID
// returns empty key if datatype with datatypeID is not found
func GetDatatypeSymKey(stub cached_stub.CachedStubInterface, caller data_model.User, datatypeID string, ownerID string, keyPath ...[]string) (data_model.Key, error) {
	defer utils.ExitFnLog(utils.EnterFnLog())

	if utils.IsStringEmpty(datatypeID) {
		customErr := &custom_errors.LengthCheckingError{Type: "datatypeID"}
		logger.Errorf(customErr.Error())
		return data_model.Key{}, errors.WithStack(customErr)
	}

	if utils.IsStringEmpty(ownerID) {
		customErr := &custom_errors.LengthCheckingError{Type: "ownerID"}
		logger.Errorf(customErr.Error())
		return data_model.Key{}, errors.WithStack(customErr)
	}

	datatypeKey, err := datatype.GetDatatypeSymKey(stub, caller, datatypeID, ownerID, keyPath...)
	if err != nil {
		logger.Errorf("Failed to GetDatatypeSymkey: %v", err)
		return data_model.Key{}, errors.Wrap(err, "Failed to GetDatatypeSymKey")
	}

	return datatypeKey, nil
}

// private function that converts solution datatype to Common SDK datatype
func convertDatatypeOMRToCommonDatatype(datatype Datatype) data_model.Datatype {
	defer utils.ExitFnLog(utils.EnterFnLog())
	datatypeCommon := data_model.Datatype{}
	datatypeCommon.DatatypeID = datatype.DatatypeID
	datatypeCommon.Description = datatype.Description
	datatypeCommon.IsActive = true
	return datatypeCommon
}

// private function that converts Common SDK's datatype interface to solution level datatype
func convertDatatypeInterfaceToDatatypeOMR(datatypeCommon datatype_interface.DatatypeInterface) Datatype {
	defer utils.ExitFnLog(utils.EnterFnLog())
	datatypeOMR := Datatype{}
	datatypeOMR.DatatypeID = datatypeCommon.GetDatatypeID()
	datatypeOMR.Description = datatypeCommon.GetDescription()
	return datatypeOMR
}

// private function that converts Common SDK's data_modal.datatype to solution level datatype
func convertDatatypeCommonToDatatypeOMR(datatypeCommon data_model.Datatype) Datatype {
	defer utils.ExitFnLog(utils.EnterFnLog())
	datatypeOMR := Datatype{}
	datatypeOMR.DatatypeID = datatypeCommon.DatatypeID
	datatypeOMR.Description = datatypeCommon.Description
	return datatypeOMR
}
