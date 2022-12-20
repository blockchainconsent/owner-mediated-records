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
	"common/bchcls/asset_mgmt"
	"common/bchcls/cached_stub"
	"common/bchcls/custom_errors"
	"common/bchcls/data_model"
	"common/bchcls/history"
	"common/bchcls/key_mgmt"
	"common/bchcls/simple_rule"
	"common/bchcls/utils"
	"encoding/json"
	"strconv"
	"time"

	"github.com/pkg/errors"
)

// Log is the query response definition for transaction logs
// TransactionID is the ID of the transaction
// Contract is the contract ID for contract life cycle functions
// Datatype is the datatype ID
// Service is the service ID for consent and data upload/download functions
// Owner is the owner ID for consent and data upload/download functions
// Target is the target ID for consent and data upload/download functions
// ContractRequesterService is the service ID of contract requester for contract life cycle functions
// ContractOwnerService is the service ID of contract owner for contract life cycle functions
// ContractRequesterOrg is the org ID of contract requester for contract life cycle functions
// ContractOwnerOrg is the org ID of contract owner for contract life cycle functions
// Timestamp is the timestamp of when the log was sent to the endorsing peer
// Type is name of the function
// Caller is ID of the caller
// Data contains arbitrary data added by the caller during each instance of logging
type Log struct {
	TransactionID            string      `json:"transaction_id"`
	Contract                 string      `json:"contract"`
	Datatype                 string      `json:"datatype"`
	Service                  string      `json:"service"`
	Owner                    string      `json:"owner"`
	Target                   string      `json:"target"`
	ContractRequesterService string      `json:"contract_requester_service"`
	ContractOwnerService     string      `json:"contract_owner_service"`
	ContractOwnerOrg         string      `json:"contract_owner_org"`
	ContractRequesterOrg     string      `json:"contract_requester_org"`
	Timestamp                int64       `json:"timestamp"`
	Type                     string      `json:"type"`
	Caller                   string      `json:"caller"`
	Data                     interface{} `json:"data"`
}

// SolutionLog is the solution level log object used when adding a transaction log
type SolutionLog struct {
	TransactionID string      `json:"transaction_id"`
	Namespace     string      `json:"namespace"`
	FunctionName  string      `json:"function_name"`
	CallerID      string      `json:"caller_id"`
	Timestamp     int64       `json:"timestamp"`
	Data          interface{} `json:"data"`
}

// internal function to check if a solution log object is valid
// only optional field is ConnectionID
func isValidSolutionLog(solutionLog SolutionLog) (bool, error) {
	if utils.IsStringEmpty(solutionLog.TransactionID) {
		custom_err := &custom_errors.LengthCheckingError{Type: "solutionLog.TransactionID"}
		logger.Errorf(custom_err.Error())
		return false, errors.WithStack(custom_err)
	}

	if utils.IsStringEmpty(solutionLog.CallerID) {
		custom_err := &custom_errors.LengthCheckingError{Type: "solutionLog.CallerID"}
		logger.Errorf(custom_err.Error())
		return false, errors.WithStack(custom_err)
	}

	if utils.IsStringEmpty(solutionLog.FunctionName) {
		custom_err := &custom_errors.LengthCheckingError{Type: "solutionLog.FunctionName"}
		logger.Errorf(custom_err.Error())
		return false, errors.WithStack(custom_err)
	}

	if utils.IsStringEmpty(solutionLog.Namespace) {
		custom_err := &custom_errors.LengthCheckingError{Type: "solutionLog.Namespace"}
		logger.Errorf(custom_err.Error())
		return false, errors.WithStack(custom_err)
	}

	// Check solutionLog's timestamp is within 10 mins of current time
	currTime := time.Now().Unix()
	if currTime-solutionLog.Timestamp > 10*60 || currTime-solutionLog.Timestamp < -10*60 {
		logger.Errorf("Invalid Timestamp (current time: %v)  %v", currTime, solutionLog.Timestamp)
		return false, errors.New("Invalid Timestamp, not within possible time range")
	}

	return true, nil
}

// AddLogWithParams is the internal function to add transaction log
func AddLogWithParams(stub cached_stub.CachedStubInterface, caller data_model.User, solutionLog SolutionLog, logSymKey data_model.Key) error {
	defer utils.ExitFnLog(utils.EnterFnLog())

	isValid, err := isValidSolutionLog(solutionLog)
	if err != nil {
		customErr := &PutTransactionLogError{Function: "AddLogWithParams"}
		logger.Errorf("%v: %v", customErr, err)
		return errors.Wrap(err, customErr.Error())
	}

	if !isValid {
		errMsg := "Solution log is not valid"
		logger.Errorf(errMsg)
		return errors.New(errMsg)
	}

	if logSymKey.Type != key_mgmt.KEY_TYPE_SYM {
		errMsg := "Log key type is not symmetric key"
		logger.Errorf(errMsg)
		return errors.New(errMsg)
	}

	if logSymKey.KeyBytes == nil {
		errMsg := "Log sym key bytes cannot be nil"
		logger.Errorf(errMsg)
		return errors.New(errMsg)
	}

	transactionLog, err := convertLogToHistory(stub, solutionLog)
	if err != nil {
		errMsg := "Failed to convert log to history"
		logger.Errorf("%v: %v", errMsg, err)
		return errors.Wrap(err, errMsg)
	}

	assetManager := asset_mgmt.GetAssetManager(stub, caller)
	historyManager := history.GetHistoryManager(assetManager)
	err = historyManager.PutInvokeTransactionLog(transactionLog, logSymKey)
	if err != nil {
		customErr := &PutTransactionLogError{Function: transactionLog.FunctionName}
		logger.Errorf("%v: %v", customErr, err)
		return errors.Wrap(err, customErr.Error())
	}

	return nil
}

// GenerateExportableSolutionLog generates exportable logs
// This function adds connection ID for offchain log storage if there is an active connectionID
func GenerateExportableSolutionLog(stub cached_stub.CachedStubInterface, caller data_model.User, solutionLog SolutionLog, logSymKey data_model.Key) (data_model.ExportableTransactionLog, error) {
	isValid, err := isValidSolutionLog(solutionLog)
	if err != nil {
		customErr := &PutTransactionLogError{Function: "GenerateExportableSolutionLog"}
		logger.Errorf("%v: %v", customErr, err)
		return data_model.ExportableTransactionLog{}, errors.Wrap(err, customErr.Error())
	}

	if !isValid {
		errMsg := "Solution log is not valid"
		logger.Errorf(errMsg)
		return data_model.ExportableTransactionLog{}, errors.New(errMsg)
	}

	transactionLog, err := convertLogToHistory(stub, solutionLog)
	if err != nil {
		errMsg := "Failed to convert log to history"
		logger.Errorf("%v: %v", errMsg, err)
		return data_model.ExportableTransactionLog{}, errors.Wrap(err, errMsg)
	}

	exportableLog, err := history.GenerateExportableTransactionLog(stub, caller, transactionLog, logSymKey)
	if err != nil {
		customErr := &GenerateExportableTransactionLogError{Function: transactionLog.FunctionName}
		logger.Errorf("%v: %v", customErr, err)
		return data_model.ExportableTransactionLog{}, errors.Wrap(err, customErr.Error())
	}

	return exportableLog, nil
}

// GetLogs returns logs
// args = [contractID, patientID, serviceID, datatypeID, orgID, data, startTimestamp, endTimestamp, latestOnly, maxNum]
// Pass "" for fields if not used
// Pass 0 for timestamps if not used
// Default for maxNum is 20
func GetLogs(stub cached_stub.CachedStubInterface, caller data_model.User, args []string) ([]byte, error) {
	defer utils.ExitFnLog(utils.EnterFnLog())
	logger.Debugf("args: %v", args)

	if len(args) < 8 {
		customErr := &custom_errors.LengthCheckingError{Type: "GetLogs arguments length"}
		logger.Errorf(customErr.Error())
		return nil, errors.New(customErr.Error())
	}

	// ==============================================================
	// Validation
	// ==============================================================

	contractID := args[0]
	patientID := args[1]
	serviceID := args[2]
	datatypeID := args[3]
	contractOrgID := args[4]
	consentOwnerTargetID := args[5]

	startTimestamp, err := strconv.ParseInt(args[6], 10, 64)
	if err != nil {
		logger.Errorf("Error converting startTimestamp to type int64")
		return nil, errors.Wrap(err, "Error converting startTimestamp to type int64")
	}

	if startTimestamp <= 0 {
		startTimestamp = -1
	}

	endTimestamp, err := strconv.ParseInt(args[7], 10, 64)
	if err != nil {
		logger.Errorf("Error converting endTimestamp to type int64")
		return nil, errors.Wrap(err, "Error converting endTimestamp to type int64")
	}

	if endTimestamp <= 0 {
		endTimestamp = -1
	}

	latestOnly := args[8]
	if latestOnly != "true" && latestOnly != "false" {
		logger.Errorf("Error: Latest only flag must be true or false")
		return nil, errors.New("Error: Latest only flag must be true or false")
	}

	maxNum, err := strconv.ParseInt(args[9], 10, 64)
	if err != nil {
		logger.Errorf("Error converting maxNum to type int")
		return nil, errors.Wrap(err, "Error converting maxNum to type int")
	}

	if maxNum < 0 {
		logger.Errorf("Max num must be greater than 0")
		return nil, errors.New("Max num must be greater than 0")
	}

	if maxNum == 0 {
		maxNum = 20
	}

	// ==============================================================
	// GetLogs
	// ==============================================================

	assetManager := asset_mgmt.GetAssetManager(stub, caller)
	historyManager := history.GetHistoryManager(assetManager)

	rulesMap := make(map[string]interface{})

	var contractRule map[string]interface{}
	if !utils.IsStringEmpty(contractID) {
		contractRule = simple_rule.R("==", simple_rule.R("var", "private_data.data.contract"), contractID)
		rulesMap["contractRule"] = contractRule
	}

	var patientRule map[string]interface{}
	if !utils.IsStringEmpty(patientID) {
		simplePatientRule := simple_rule.R("==", simple_rule.R("var", "private_data.data.owner"), patientID)
		targetPatientRule := simple_rule.R("==", simple_rule.R("var", "private_data.data.target"), patientID)
		patientRule = simple_rule.R("or", simplePatientRule, targetPatientRule)
		rulesMap["patientRule"] = patientRule
	}

	var datatypeRule map[string]interface{}
	if !utils.IsStringEmpty(datatypeID) {
		datatypeRule = simple_rule.R("==", simple_rule.R("var", "private_data.data.datatype"), datatypeID)
		rulesMap["datatypeRule"] = datatypeRule
	}

	var serviceRule map[string]interface{}
	if !utils.IsStringEmpty(serviceID) {
		simpleServiceRule := simple_rule.R("==", simple_rule.R("var", "private_data.data.service"), serviceID)
		targetServiceRule := simple_rule.R("==", simple_rule.R("var", "private_data.data.target"), serviceID)
		ownerServiceRule := simple_rule.R("==", simple_rule.R("var", "private_data.data.owner_service"), serviceID)
		requesterServiceRule := simple_rule.R("==", simple_rule.R("var", "private_data.data.requester_service"), serviceID)
		serviceRule = simple_rule.R("or", simpleServiceRule, targetServiceRule, ownerServiceRule, requesterServiceRule)
		rulesMap["serviceRule"] = serviceRule
	}

	var contractOrgRule map[string]interface{}
	if !utils.IsStringEmpty(contractOrgID) {
		ownerOrgRule := simple_rule.R("==", simple_rule.R("var", "private_data.data.owner_org"), contractOrgID)
		requesterOrgRule := simple_rule.R("==", simple_rule.R("var", "private_data.data.requester_org"), contractOrgID)
		contractOwnerOrgRule := simple_rule.R("==", simple_rule.R("var", "private_data.data.contract_owner_org"), contractOrgID)
		contractRequesterOwnerRule := simple_rule.R("==", simple_rule.R("var", "private_data.data.contract_requester_org"), contractOrgID)
		contractOrgRule = simple_rule.R("or", ownerOrgRule, requesterOrgRule, contractOwnerOrgRule, contractRequesterOwnerRule)
		rulesMap["contractOrgRule"] = contractOrgRule
	}

	var contractOwnerTargetRule map[string]interface{}
	if !utils.IsStringEmpty(consentOwnerTargetID) {
		targetRule := simple_rule.R("==", simple_rule.R("var", "private_data.data.target"), consentOwnerTargetID)
		ownerRule := simple_rule.R("==", simple_rule.R("var", "private_data.data.owner"), consentOwnerTargetID)
		contractOwnerTargetRule = simple_rule.R("or", targetRule, ownerRule)
		rulesMap["contractOwnerTargetRule"] = contractOwnerTargetRule
	}

	// combine rules
	andPredicate := simple_rule.R("and")
	for _, ruleComponent := range rulesMap {
		if ruleComponent != nil {
			andPredicate["and"] = append(andPredicate["and"].([]interface{}), ruleComponent)
		}
	}
	rule := simple_rule.NewRule(andPredicate)

	var logs []data_model.TransactionLog
	logs, _, err = historyManager.GetTransactionLogs("OMR", "field_1", "", startTimestamp, endTimestamp, "", int(maxNum), &rule, OMRAssetKeyPathFuncForLogging)

	// if latest only flag is set to true, then only return the last element
	if latestOnly == "true" {
		logs = logs[len(logs)-1:]
	}

	logResults := convertLogFromHistory(logs)
	return json.Marshal(&logResults)
}

func convertLogToHistory(stub cached_stub.CachedStubInterface, solutionLog SolutionLog) (data_model.TransactionLog, error) {
	defer utils.ExitFnLog(utils.EnterFnLog())

	history := data_model.TransactionLog{}
	history.TransactionID = solutionLog.TransactionID
	history.Namespace = solutionLog.Namespace
	history.FunctionName = solutionLog.FunctionName
	history.CallerID = solutionLog.CallerID
	history.Timestamp = solutionLog.Timestamp
	history.Data = solutionLog.Data

	dsConnectionID, err := GetActiveConnectionID(stub)
	if err != nil {
		errMsg := "Failed to GetActiveConnectionID for history"
		logger.Errorf("%v: %v", errMsg, err)
		return data_model.TransactionLog{}, errors.Wrap(err, errMsg)
	}
	history.ConnectionID = dsConnectionID

	return history, nil
}

func convertLogFromHistory(logs []data_model.TransactionLog) []Log {
	defer utils.ExitFnLog(utils.EnterFnLog())

	logsOMR := []Log{}
	logOMR := Log{}
	for _, logCommon := range logs {
		contract := logCommon.Data.(map[string]interface{})["contract"]
		if contract != nil {
			logOMR.Contract = contract.(string)
		}

		owner := logCommon.Data.(map[string]interface{})["owner"]
		if owner != nil {
			logOMR.Owner = owner.(string)
		}

		target := logCommon.Data.(map[string]interface{})["target"]
		if owner != nil {
			logOMR.Target = target.(string)
		}

		service := logCommon.Data.(map[string]interface{})["service"]
		if service != nil {
			logOMR.Service = service.(string)
		}

		requesterService := logCommon.Data.(map[string]interface{})["requester_service"]
		if requesterService != nil {
			logOMR.ContractRequesterService = requesterService.(string)
		}

		ownerService := logCommon.Data.(map[string]interface{})["owner_service"]
		if ownerService != nil {
			logOMR.ContractOwnerService = ownerService.(string)
		}

		ownerOrg := logCommon.Data.(map[string]interface{})["owner_org"]
		if ownerOrg != nil {
			logOMR.ContractOwnerOrg = ownerOrg.(string)
		}

		requesterOrg := logCommon.Data.(map[string]interface{})["requester_org"]
		if requesterOrg != nil {
			logOMR.ContractOwnerOrg = requesterOrg.(string)
		}

		datatype := logCommon.Data.(map[string]interface{})["datatype"]
		if datatype != nil {
			logOMR.Datatype = datatype.(string)
		}

		data := logCommon.Data.(map[string]interface{})["data"]
		if data != nil {
			logOMR.Data = data
		} else {
			logOMR.Data = make(map[string]interface{})
		}

		logOMR.Timestamp = logCommon.Timestamp
		logOMR.Type = logCommon.FunctionName
		logOMR.TransactionID = logCommon.TransactionID
		logOMR.Caller = logCommon.CallerID

		logsOMR = append(logsOMR, logOMR)
	}

	return logsOMR
}
