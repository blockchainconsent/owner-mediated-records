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
	"fmt"
)

type RegisterOrgError struct {
	Org string
}

func (e *RegisterOrgError) Error() string {
	return fmt.Sprintf("Failed to register org %v", e.Org)
}

type GetOrgError struct {
	Org string
}

func (e *GetOrgError) Error() string {
	return fmt.Sprintf("Failed to get org %v", e.Org)
}

type GetOrgsError struct {
}

func (e *GetOrgsError) Error() string {
	return fmt.Sprintf("Failed to get orgs")
}

type GetSubgroupsError struct {
	Org string
}

func (e *GetSubgroupsError) Error() string {
	return fmt.Sprintf("Failed to get subgroups for org %v", e.Org)
}

type GetUserError struct {
	User string
}

func (e *GetUserError) Error() string {
	return fmt.Sprintf("Failed to get user %v", e.User)
}

type GetUsersError struct {
	Group string
}

func (e *GetUsersError) Error() string {
	return fmt.Sprintf("Failed to get users for %v", e.Group)
}

type GetDatatypeError struct {
	Datatype string
}

func (e *GetDatatypeError) Error() string {
	return fmt.Sprintf("Failed to get datatype %v", e.Datatype)
}

type PutAssetError struct {
	Asset string
}

func (e *PutAssetError) Error() string {
	return fmt.Sprintf("Failed to put asset for %v", e.Asset)
}

type DeleteAssetError struct {
	Asset string
}

func (e *DeleteAssetError) Error() string {
	return fmt.Sprintf("Failed to delete asset for %v", e.Asset)
}

type ValidateDatatypeError struct {
	Datatype string
}

func (e *ValidateDatatypeError) Error() string {
	return fmt.Sprintf("Failed to validate datatype:  %v", e.Datatype)
}

type GetServiceError struct {
	Service string
}

func (e *GetServiceError) Error() string {
	return fmt.Sprintf("Failed to get service %v", e.Service)
}

type AddDatatypeError struct {
	Datatype string
}

func (e *AddDatatypeError) Error() string {
	return fmt.Sprintf("Failed to add datatype for %v", e.Datatype)
}

type ConvertToAssetError struct {
	Asset string
}

func (e *ConvertToAssetError) Error() string {
	return fmt.Sprintf("Failed to convert to asset %v", e.Asset)
}

type GetEnrollmentError struct {
	Enrollment string
}

func (e *GetEnrollmentError) Error() string {
	return fmt.Sprintf("Failed to get enrollment %v", e.Enrollment)
}

type GetConsentError struct {
	Consent string
}

func (e *GetConsentError) Error() string {
	return fmt.Sprintf("Failed to get consent %v", e.Consent)
}

type ValidateConsentError struct{}

func (e *ValidateConsentError) Error() string {
	return "Failed to vaidate consent"
}

type GetContractError struct {
	ContractID string
}

func (e *GetContractError) Error() string {
	return fmt.Sprintf("Failed to get contract %v", e.ContractID)
}

type GetOwnerDataError struct {
	Owner    string
	Datatype string
}

func (e *GetOwnerDataError) Error() string {
	return fmt.Sprintf("Failed to get owner data %v, %v", e.Owner, e.Datatype)
}

type GetDatasError struct {
	FieldNames []string
	Values     []string
}

func (e *GetDatasError) Error() string {
	return fmt.Sprintf("Data not found with sort order: %v and partial key list: %v", e.FieldNames, e.Values)
}

type PutUserInOrgError struct {
	User string
	Org  string
}

func (e *PutUserInOrgError) Error() string {
	return fmt.Sprintf("Failed to get owner data %v, %v", e.User, e.Org)
}

type DecryptTokenError struct {
}

func (e *DecryptTokenError) Error() string {
	return fmt.Sprintf("Failed to decrypt validation token")
}

type GetDataKeyError struct {
	KeyID string
}

func (e *GetDataKeyError) Error() string {
	return fmt.Sprintf("Failed to get data key %v", e.KeyID)
}

type PutTransactionLogError struct {
	Function string
}

func (e *PutTransactionLogError) Error() string {
	return fmt.Sprintf("Failed to add transaction log for %v", e.Function)
}

type GenerateExportableTransactionLogError struct {
	Function string
}

func (e *GenerateExportableTransactionLogError) Error() string {
	return fmt.Sprintf("Failed to generate exportable transaction log for %v", e.Function)
}

type GetKeyPathError struct {
	Caller  string
	AssetID string
}

func (e *GetKeyPathError) Error() string {
	return fmt.Sprintf("Failed to get key path from %v to %v", e.Caller, e.AssetID)
}

type GetDatatypeKeyPathError struct {
	Caller     string
	DatatypeID string
}

func (e *GetDatatypeKeyPathError) Error() string {
	return fmt.Sprintf("Failed to get key path from %v to datatype %v", e.Caller, e.DatatypeID)
}

type AddSolutionLogError struct {
	FunctionName string
}

func (e *AddSolutionLogError) Error() string {
	return fmt.Sprintf("Failed to add solution log for %v", e.FunctionName)
}
