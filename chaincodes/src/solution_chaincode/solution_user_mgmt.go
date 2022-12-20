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
	"common/bchcls/user_access_ctrl"
	"common/bchcls/user_mgmt"
	"common/bchcls/user_mgmt/user_groups"
	"common/bchcls/utils"
	"crypto/rsa"
	"encoding/json"
	"strconv"

	"github.com/pkg/errors"
)

// Solution level role can be system, audit, org, or user
const SOLUTION_ROLE_USER = "user"
const SOLUTION_ROLE_ORG = "org"
const SOLUTION_ROLE_SYSTEM = "system"
const SOLUTION_ROLE_AUDIT = "audit"

// When an organization is registered, a default org admin is created
// The admin's role type is org
// De-identified fields:
//   - ID
//   - Name
//   - Org
type SolutionUser struct {
	ID                 string          `json:"id"`
	Name               string          `json:"name"`
	Role               string          `json:"role"`
	PublicKey          *rsa.PublicKey  `json:"-"`
	PublicKeyB64       string          `json:"public_key"`
	PrivateKey         *rsa.PrivateKey `json:"-"`
	PrivateKeyB64      string          `json:"private_key"`
	SymKey             []byte          `json:"-"`
	SymKeyB64          string          `json:"sym_key"`
	IsGroup            bool            `json:"is_group"`
	Org                string          `json:"org"`
	Status             string          `json:"status"`
	SolutionPublicData interface{}     `json:"solution_public_data"`

	// private data
	Email               string       `json:"email"`
	KmsPublicKeyId      string       `json:"kms_public_key_id"`
	KmsPrivateKeyId     string       `json:"kms_private_key_id"`
	KmsSymKeyId         string       `json:"kms_sym_key_id"`
	Secret              string       `json:"secret"`
	SolutionPrivateData interface{}  `json:"solution_private_data"`
	SolutionInfo        SolutionInfo `json:"solution_info"`
}

type SolutionInfo struct {
	Services   []string `json:"services"`
	IsOrgAdmin bool     `json:"is_org_admin"`
}

// RegisterUser is the solution level register user function
// It maintains solution level user object and calls core chaincode funtion user_mgmt.RegisterUser
// During an update, it does not overwrite org admin flag and services fields in solution private data
// and does not overwrite org field.
// When registering an org user, admins of that org should call PutUserInOrg function immediately
// after to put user in the organization.
//
// Auditors can only be registered by system admins
// System admins can only be registered by system admin
// args = [ SolutionUser ]
func RegisterUser(stub cached_stub.CachedStubInterface, caller data_model.User, args []string) ([]byte, error) {
	defer utils.ExitFnLog(utils.EnterFnLog())
	logger.Debugf("RegisterUser args: %v", args)

	if len(args) != 1 {
		customErr := &custom_errors.LengthCheckingError{Type: "Solution RegisterUser arguments length"}
		logger.Errorf(customErr.Error())
		return nil, errors.New(customErr.Error())
	}

	var solutionUser = SolutionUser{}
	solutionUserBytes := []byte(args[0])
	err := json.Unmarshal(solutionUserBytes, &solutionUser)
	if err != nil {
		customErr := &custom_errors.UnmarshalError{Type: "solutionUser"}
		logger.Errorf("%v: %v", customErr, err)
		return nil, errors.Wrap(err, customErr.Error())
	}

	return RegisterSolutionUserWithParams(stub, caller, solutionUser)
}

// Internal function for registering solution user
func RegisterSolutionUserWithParams(stub cached_stub.CachedStubInterface, caller data_model.User, solutionUser SolutionUser) ([]byte, error) {
	defer utils.ExitFnLog(utils.EnterFnLog())
	logger.Debug("RegisterSolutionUserWithParams")

	existingUser, err := GetSolutionUserWithParams(stub, caller, solutionUser.ID, true, true)
	if err != nil {
		customErr := &GetUserError{User: solutionUser.ID}
		logger.Errorf("%v: %v", customErr, err)
		return nil, errors.Wrap(err, customErr.Error())
	}
	// If user already exists, this is an update
	if !utils.IsStringEmpty(existingUser.ID) {
		// If caller is not user, check that caller has access to user's private key
		if caller.ID != solutionUser.ID && existingUser.PrivateKey == nil {
			logger.Errorf("Caller does not have access to user")
			return nil, errors.New("Caller does not have access to user")
		}

		// Keep org the same during update
		solutionUser.Org = existingUser.Org

		// during an update, keep org admin flag and services fields in solution private data
		solutionUser.SolutionInfo.IsOrgAdmin = existingUser.SolutionInfo.IsOrgAdmin
		solutionUser.SolutionInfo.Services = existingUser.SolutionInfo.Services
	} else {
		// If org field has value, check that caller is org admin
		if !utils.IsStringEmpty(solutionUser.Org) {
			isAdmin, _, err := user_groups.IsUserAdminOfGroup(stub, caller.ID, solutionUser.Org)
			if err != nil {
				customErr := &custom_errors.NotGroupAdminError{UserID: caller.ID, GroupID: solutionUser.Org}
				logger.Errorf("%v: %v", customErr, err)
				return nil, errors.Wrap(err, customErr.Error())
			}

			if !isAdmin {
				customErr := &custom_errors.NotGroupAdminError{UserID: caller.ID, GroupID: solutionUser.Org}
				logger.Errorf("%v", customErr)
				return nil, errors.WithStack(customErr)
			}
		}
	}

	// must marshall to register using user_mgmt
	platformUser, err := convertToPlatformUser(stub, solutionUser)
	if err != nil {
		errMsg := "Failed to convertToPlatformUser"
		logger.Errorf("%v: %v", errMsg, err)
		return nil, errors.Wrap(err, errMsg)
	}
	platformUserBytes, err := json.Marshal(&platformUser)
	if err != nil {
		customErr := &custom_errors.MarshalError{Type: "platformUser"}
		logger.Errorf("%v: %v", customErr, err)
		return nil, errors.Wrap(err, customErr.Error())
	}

	giveAccess := "false"
	// If we are registering an org user, caller (org admin) will always have access to this user
	// This creates an edge in user_mgmt.RegisterUser from caller priv key to org user's priKeyHashSym
	if !utils.IsStringEmpty(solutionUser.Org) {
		giveAccess = "true"
	}

	if solutionUser.Role == user_mgmt.ROLE_USER {
		return user_mgmt.RegisterUser(stub, caller, []string{string(platformUserBytes), giveAccess})
	} else if solutionUser.Role == SOLUTION_ROLE_SYSTEM {
		return user_mgmt.RegisterSystemAdmin(stub, caller, []string{string(platformUserBytes), giveAccess})
	} else if solutionUser.Role == SOLUTION_ROLE_AUDIT {
		return user_mgmt.RegisterAuditor(stub, caller, []string{string(platformUserBytes), giveAccess})
	} else {
		// no other role types are allowed, return error
		return nil, errors.New("Error: invalid role type")
	}
}

// GetSolutionUser gets a SolutionUser object given user ID
// Will always attempt to get private data
// args = [userID]
func GetSolutionUser(stub cached_stub.CachedStubInterface, caller data_model.User, args []string) ([]byte, error) {
	defer utils.ExitFnLog(utils.EnterFnLog())
	logger.Debugf("GetSolutionUser args: %v", args)

	if len(args) != 1 {
		customErr := &custom_errors.LengthCheckingError{Type: "Solution GetUser arguments length"}
		logger.Errorf(customErr.Error())
		return nil, errors.New(customErr.Error())
	}

	userID := args[0]

	solutionUser, err := GetSolutionUserWithParams(stub, caller, userID, true, true)
	if err != nil {
		customErr := &GetUserError{User: userID}
		logger.Errorf("%v: %v", customErr, err)
		return nil, errors.Wrap(err, customErr.Error())
	}

	return json.Marshal(&solutionUser)
}

// GetSolutionUsers gets SolutionUser objects for a given orgID
// When getting all auditors and system admins, pass in * for orgID
// Attempts to get private data of users; will return private data if caller has access
// args = [orgID, maxNum, role]
// maxNum is maximum number of results to be returned, default is 20
func GetSolutionUsers(stub cached_stub.CachedStubInterface, caller data_model.User, args []string) ([]byte, error) {
	defer utils.ExitFnLog(utils.EnterFnLog())
	logger.Debugf("GetSolutionUsers args: %v", args)

	if len(args) != 2 && len(args) != 3 {
		customErr := &custom_errors.LengthCheckingError{Type: "GetSolutionUsers arguments length"}
		logger.Errorf(customErr.Error())
		return nil, errors.New(customErr.Error())
	}

	orgID := args[0]
	if utils.IsStringEmpty(orgID) {
		customErr := &custom_errors.LengthCheckingError{Type: "orgID"}
		logger.Errorf(customErr.Error())
		return nil, errors.WithStack(customErr)
	}

	maxNum, err := strconv.ParseInt(args[1], 10, 32)
	if err != nil {
		logger.Errorf("Error converting num to type int32")
		return nil, errors.Wrap(err, "Error converting num to type int32")
	}

	num := int(maxNum)
	if num < 0 {
		logger.Error("Max num must be greater than 0")
		return nil, errors.New("Max num must be greater than 0")
	}

	if num == 0 {
		logger.Error("Max num is 0, defaulting to 20")
		num = 20
	}

	userList := []SolutionUser{}

	getOrgUsers := true
	if len(args) == 3 {
		role := args[2]
		if !utils.IsStringEmpty(role) {
			if orgID == "*" {
				getOrgUsers = false
				if role == SOLUTION_ROLE_ORG {
					getOrgUsers = true
				} else {
					if role != SOLUTION_ROLE_AUDIT && role != SOLUTION_ROLE_SYSTEM {
						logger.Errorf("Can only get org users, auditors, and system admins using this function")
						return nil, errors.New("Can only get org users, auditors, and system admins using this function")
					}
					userList, err = GetSolutionUsersOfRole(stub, caller, role, num)
					if err != nil {
						logger.Errorf("Failed to get solution users for role type, %v", err)
						return nil, errors.WithStack(err)
					}
				}
			}
		}
	}

	if getOrgUsers {
		// Get all memberIds of this group, then search for each one and append to list
		memberIds, err := user_groups.SlowGetGroupMemberIDs(stub, orgID)
		if err != nil {
			logger.Errorf("GetGroupMemberIDS returned error: %v", err)
			return nil, errors.WithStack(err)
		}
		for _, userId := range memberIds {

			// Get the user
			user, err := user_mgmt.GetUserData(stub, caller, userId, false, true)
			if user.ID == "" {
				msg := "Failed to get orgUser: " + userId
				logger.Errorf(msg)
				return nil, errors.Wrap(err, msg)
			}

			solutionUser := convertToSolutionUser(user)
			userList = append(userList, solutionUser)
		}
	}

	return json.Marshal(&userList)
}

// GetSolutionUsersOfRole gets all solution users of a role type
// If caller has access to a user, user private data will be returned
// parameters: role, maxNum
// maxNum is maximum number of results to be returned
func GetSolutionUsersOfRole(stub cached_stub.CachedStubInterface, caller data_model.User, role string, maxNum int) ([]SolutionUser, error) {
	defer utils.ExitFnLog(utils.EnterFnLog())

	if role != SOLUTION_ROLE_SYSTEM && role != SOLUTION_ROLE_AUDIT {
		logger.Errorf("Role must be system or audit")
		return nil, errors.New("Role must be system or audit")
	}

	if maxNum < 0 {
		logger.Error("Max num must be greater than 0")
		return nil, errors.New("Max num must be greater than 0")
	}

	if maxNum == 0 {
		logger.Error("Max num is 0, defaulting to 20")
		maxNum = 20
	}

	// ==============================================================
	// GetAssets with page
	// ==============================================================
	users := []SolutionUser{}

	userIter, err := user_mgmt.GetUserIter(stub, caller, []string{"false", role}, []string{"false", role}, true, false, KeyPathFunc, "", maxNum, nil)
	if err != nil {
		logger.Errorf("GetUserIter failed: %v", err)
		return nil, errors.Wrap(err, "GetUserIter failed")
	}

	// TODO: return ledger key in the future
	assets, _, err := userIter.GetAssetPage()
	if err != nil {
		errMsg := "Failed to get asset page"
		logger.Errorf("%v: %v", errMsg, err)
		return nil, errors.Wrap(err, errMsg)
	}

	for _, asset := range assets {
		if utils.IsStringEmpty(asset.AssetId) {
			continue
		}

		commonUser := user_mgmt.ConvertFromAsset(&asset)
		user := convertToSolutionUser(commonUser)
		// simpleUser := convertSimpleUserFromAsset(asset)
		// if utils.IsStringEmpty(simpleUser.ID) {
		// 	continue
		// }

		// user, err := GetSolutionUserWithParams(stub, caller, simpleUser.ID, false, true)
		// if err != nil {
		// 	continue
		// }

		users = append(users, user)
	}

	return users, nil
}

// GetSolutionUserWithParams is a helper function that returns a solution user object given userID
// If no user is found, does not return error; returns empty user object
func GetSolutionUserWithParams(stub cached_stub.CachedStubInterface, caller data_model.User, userID string, includePrivateKeys bool, includePrivateData bool) (SolutionUser, error) {
	defer utils.ExitFnLog(utils.EnterFnLog())

	if utils.IsStringEmpty(userID) {
		customErr := &custom_errors.LengthCheckingError{Type: "userID"}
		logger.Errorf(customErr.Error())
		return SolutionUser{}, errors.WithStack(customErr)
	}

	org, isOrgAdmin := getCallerOrgAdminInfo(caller)

	// if caller is regular user
	// use GetKeyPathFromCallerToAsset
	symKeyPath := []string{}
	prvKeyPath := []string{}
	// if caller is org user (admin) or org
	if !utils.IsStringEmpty(org) && isOrgAdmin {
		symKeyPath = GetKeyPathFromOrgAdminToOrgUser(caller, org, userID, true)
		prvKeyPath = GetKeyPathFromOrgAdminToOrgUser(caller, org, userID, false)
	}

	platformUser, err := user_mgmt.GetUserData(stub, caller, userID, includePrivateKeys, includePrivateData, symKeyPath, prvKeyPath)
	if err != nil {
		customErr := &GetUserError{User: userID}
		logger.Errorf("%v: %v", customErr, err)
		return SolutionUser{}, errors.Wrap(err, customErr.Error())
	}

	if utils.IsStringEmpty(platformUser.ID) {
		return SolutionUser{}, nil
	}

	return convertToSolutionUser(platformUser), nil
}

// GetOrgs returns all organizations
// args = []
func GetOrgs(stub cached_stub.CachedStubInterface, caller data_model.User, args []string) ([]byte, error) {
	defer utils.ExitFnLog(utils.EnterFnLog())
	logger.Debugf("GetOrgs args: %v", args)

	orgBytes, err := user_mgmt.GetOrgs(stub, caller, args)
	if err != nil {
		customErr := &GetOrgsError{}
		logger.Errorf("%v: %v", customErr, err)
		return nil, errors.Wrap(err, customErr.Error())
	}

	var orgs = []data_model.User{}
	err = json.Unmarshal(orgBytes, &orgs)
	if err != nil {
		customErr := &custom_errors.UnmarshalError{Type: "orgs"}
		logger.Errorf("%v: %v", customErr, err)
		return nil, errors.Wrap(err, customErr.Error())
	}

	orgsReturnList := []data_model.User{}
	for _, org := range orgs {
		solutionPublicData := org.SolutionPublicData.(map[string]interface{})
		orgType, _ := solutionPublicData["type"]

		if orgType == "organization" {
			orgsReturnList = append(orgsReturnList, org)
		}
	}

	return json.Marshal(&orgsReturnList)
}

// AddPermissionServiceAdmin adds service admin permission to an org user
// This function can only be called by org admins
// args = [ userID, serviceID ]
func AddPermissionServiceAdmin(stub cached_stub.CachedStubInterface, caller data_model.User, args []string) ([]byte, error) {
	defer utils.ExitFnLog(utils.EnterFnLog())
	logger.Debugf("args: %v", args)

	if len(args) != 2 {
		customErr := &custom_errors.LengthCheckingError{Type: "AddPermissionServiceAdmin arguments length"}
		logger.Errorf(customErr.Error())
		return nil, errors.New(customErr.Error())
	}

	userID := args[0]
	serviceID := args[1]

	if utils.IsStringEmpty(userID) {
		customErr := &custom_errors.LengthCheckingError{Type: "userID"}
		logger.Errorf(customErr.Error())
		return nil, errors.WithStack(customErr)
	}

	if utils.IsStringEmpty(serviceID) {
		customErr := &custom_errors.LengthCheckingError{Type: "serviceID"}
		logger.Errorf(customErr.Error())
		return nil, errors.WithStack(customErr)
	}

	// Check if user is already admin of service
	isAdmin, _, err := user_groups.IsUserAdminOfGroup(stub, userID, serviceID)
	if err != nil {
		customErr := &custom_errors.NotGroupAdminError{UserID: userID, GroupID: serviceID}
		logger.Errorf("%v: %v", customErr, err)
		return nil, errors.Wrap(err, customErr.Error())
	}

	// if is admin already, return
	if isAdmin {
		return nil, nil
	}

	// convert caller to solution user object to get org field
	solutionCaller := convertToSolutionUser(caller)

	// construct new caller object representing org itself
	org, err := user_mgmt.GetUserData(stub, caller, solutionCaller.Org, true, false)
	if err != nil {
		customErr := &GetOrgError{Org: solutionCaller.Org}
		logger.Errorf("%v: %v", customErr, err)
		return nil, errors.Wrap(err, customErr.Error())
	}

	if org.PrivateKey == nil {
		errMsg := "Caller does not have access to org private key"
		logger.Errorf(errMsg)
		return nil, errors.New(errMsg)
	}
	orgCaller := org

	// Check caller has access to user's private keys
	// Needed to update user private data's list of services later
	solutionUser, err := GetSolutionUserWithParams(stub, orgCaller, userID, true, true)
	if err != nil {
		customErr := &GetUserError{User: userID}
		logger.Errorf("%v: %v", customErr, err)
		return nil, errors.Wrap(err, customErr.Error())
	}

	if solutionUser.PrivateKey == nil {
		logger.Errorf("Caller does not have access to add permission")
		return nil, errors.New("Caller does not have access to add permission")
	}

	// Check user role and org
	if solutionUser.Role != SOLUTION_ROLE_USER {
		logger.Errorf("Org user role must be user")
		return nil, errors.New("Org user role must be user")
	}

	if utils.IsStringEmpty(solutionUser.Org) {
		logger.Errorf("User org field is empty")
		return nil, errors.New("User org field is empty")
	}

	// Put user in service
	_, err = user_mgmt.PutUserInOrg(stub, caller, []string{userID, serviceID, "true"})
	if err != nil {
		logger.Errorf("Failed to put user in org")
		return nil, errors.Wrapf(err, "Failed to put user in org")
	}

	// append service ID to list of services
	solutionUser.SolutionInfo.Services = append(solutionUser.SolutionInfo.Services, serviceID)
	platformUser, err := convertToPlatformUser(stub, solutionUser)
	if err != nil {
		errMsg := "Failed to convertToPlatformUser"
		logger.Errorf("%v: %v", errMsg, err)
		return nil, errors.Wrap(err, errMsg)
	}

	// if caller is org admin, act as org
	return nil, user_mgmt.RegisterUserWithParams(stub, orgCaller, platformUser, false)
}

// RemovePermissionServiceAdmin, only org admins can remove service permission
// args = [ userID, serviceID ]
func RemovePermissionServiceAdmin(stub cached_stub.CachedStubInterface, caller data_model.User, args []string) ([]byte, error) {
	defer utils.ExitFnLog(utils.EnterFnLog())
	logger.Debugf("args: %v", args)

	if len(args) != 2 {
		customErr := &custom_errors.LengthCheckingError{Type: "RemovePermissionServiceAdmin arguments length"}
		logger.Errorf(customErr.Error())
		return nil, errors.New(customErr.Error())
	}

	userID := args[0]
	serviceID := args[1]

	if utils.IsStringEmpty(userID) {
		customErr := &custom_errors.LengthCheckingError{Type: "userID"}
		logger.Errorf(customErr.Error())
		return nil, errors.WithStack(customErr)
	}

	if utils.IsStringEmpty(serviceID) {
		customErr := &custom_errors.LengthCheckingError{Type: "serviceID"}
		logger.Errorf(customErr.Error())
		return nil, errors.WithStack(customErr)
	}

	// Check if user is already not an admin of service
	isAdmin, _, err := user_groups.IsUserAdminOfGroup(stub, userID, serviceID)
	if err != nil {
		customErr := &custom_errors.NotGroupAdminError{UserID: userID, GroupID: serviceID}
		logger.Errorf("%v: %v", customErr, err)
		return nil, errors.Wrap(err, customErr.Error())
	}

	if !isAdmin {
		return nil, nil
	}

	// convert caller to solution user object to get org field
	solutionCaller := convertToSolutionUser(caller)

	// construct new caller object representing org itself
	org, err := user_mgmt.GetUserData(stub, caller, solutionCaller.Org, true, false)
	if err != nil {
		customErr := &GetOrgError{Org: solutionCaller.Org}
		logger.Errorf("%v: %v", customErr, err)
		return nil, errors.Wrap(err, customErr.Error())
	}

	if org.PrivateKey == nil {
		errMsg := "Caller does not have access to org private key"
		logger.Errorf(errMsg)
		return nil, errors.New(errMsg)
	}
	orgCaller := org

	solutionUser, err := GetSolutionUserWithParams(stub, orgCaller, userID, true, true)
	if err != nil {
		customErr := &GetUserError{User: userID}
		logger.Errorf("%v: %v", customErr, err)
		return nil, errors.Wrap(err, customErr.Error())
	}

	if solutionUser.PrivateKey == nil {
		logger.Errorf("Caller does not have access to add permission")
		return nil, errors.New("Caller does not have access to add permission")
	}

	// Check user role and org
	if solutionUser.Role != SOLUTION_ROLE_USER {
		logger.Errorf("Org user role must be user")
		return nil, errors.New("Org user role must be user")
	}

	if utils.IsStringEmpty(solutionUser.Org) {
		logger.Errorf("User org field is empty")
		return nil, errors.New("User org field is empty")
	}

	// remove serviceID from list of serivces
	solutionUser.SolutionInfo.Services = utils.RemoveItemFromList(solutionUser.SolutionInfo.Services, serviceID)
	platformUser, err := convertToPlatformUser(stub, solutionUser)
	if err != nil {
		errMsg := "Failed to convertToPlatformUser"
		logger.Errorf("%v: %v", errMsg, err)
		return nil, errors.Wrap(err, errMsg)
	}

	err = user_mgmt.RegisterUserWithParams(stub, orgCaller, platformUser, false)
	if err != nil {
		logger.Errorf("Failed to update SolutionPrivateData for user")
		return nil, errors.New("Failed to update SolutionPrivateData for user")
	}

	// Then remove admin permission
	return user_groups.RemoveAdminPermissionOfGroup(stub, orgCaller, args)
}

// AddPermissionOrgAdmin adds org admin permission to user of role type org
// This function can only be called by org admin
// args = [ userID, orgID ]
func AddPermissionOrgAdmin(stub cached_stub.CachedStubInterface, caller data_model.User, args []string) ([]byte, error) {
	defer utils.ExitFnLog(utils.EnterFnLog())
	logger.Debugf("args: %v", args)

	if len(args) != 2 {
		customErr := &custom_errors.LengthCheckingError{Type: "AddPermissionOrgAdmin arguments length"}
		logger.Errorf(customErr.Error())
		return nil, errors.New(customErr.Error())
	}

	userID := args[0]
	orgID := args[1]

	if utils.IsStringEmpty(userID) {
		customErr := &custom_errors.LengthCheckingError{Type: "userID"}
		logger.Errorf(customErr.Error())
		return nil, errors.WithStack(customErr)
	}

	if utils.IsStringEmpty(orgID) {
		customErr := &custom_errors.LengthCheckingError{Type: "orgID"}
		logger.Errorf(customErr.Error())
		return nil, errors.WithStack(customErr)
	}

	// Check if user is already admin of org
	isAdmin, _, err := user_groups.IsUserAdminOfGroup(stub, userID, orgID)
	if err != nil {
		customErr := &custom_errors.NotGroupAdminError{UserID: userID, GroupID: orgID}
		logger.Errorf("%v: %v", customErr, err)
		return nil, errors.Wrap(err, customErr.Error())
	}

	// if user is an admin already, return
	if isAdmin {
		return nil, nil
	}

	// convert caller to solution user object to get org field
	solutionCaller := convertToSolutionUser(caller)

	// construct new caller object representing org itself
	org, err := user_mgmt.GetUserData(stub, caller, solutionCaller.Org, true, false)
	if err != nil {
		customErr := &GetOrgError{Org: solutionCaller.Org}
		logger.Errorf("%v: %v", customErr, err)
		return nil, errors.Wrap(err, customErr.Error())
	}

	if org.PrivateKey == nil {
		errMsg := "Caller does not have access to org private key"
		logger.Errorf(errMsg)
		return nil, errors.New(errMsg)
	}
	orgCaller := org

	solutionUser, err := GetSolutionUserWithParams(stub, orgCaller, userID, true, true)
	if err != nil {
		customErr := &GetUserError{User: userID}
		logger.Errorf("%v: %v", customErr, err)
		return nil, errors.Wrap(err, customErr.Error())
	}

	if solutionUser.PrivateKey == nil {
		logger.Errorf("Caller does not have access to add permission")
		return nil, errors.New("Caller does not have access to add permission")
	}

	// Check user role and org
	if solutionUser.Role != SOLUTION_ROLE_USER {
		logger.Errorf("Org user role must be user")
		return nil, errors.New("Org user role must be user")
	}

	if orgID != solutionUser.Org {
		logger.Errorf("Org mismatch")
		return nil, errors.New("Org mismatch")
	}

	// Put user in org
	_, err = user_mgmt.PutUserInOrg(stub, orgCaller, []string{userID, orgID, "true"})
	if err != nil {
		logger.Errorf("Failed to put user in org")
		return nil, errors.Wrapf(err, "Failed to put user in org")
	}

	solutionUser.SolutionInfo.IsOrgAdmin = true
	platformUser, err := convertToPlatformUser(stub, solutionUser)
	if err != nil {
		errMsg := "Failed to convertToPlatformUser"
		logger.Errorf("%v: %v", errMsg, err)
		return nil, errors.Wrap(err, errMsg)
	}

	return nil, user_mgmt.RegisterUserWithParams(stub, orgCaller, platformUser, false)
}

// RemovePermissionOrgAdmin removes org admin permission from an org user
// This function can only be called by org admin
// args = [ userID, orgID ]
func RemovePermissionOrgAdmin(stub cached_stub.CachedStubInterface, caller data_model.User, args []string) ([]byte, error) {
	defer utils.ExitFnLog(utils.EnterFnLog())
	logger.Debugf("args: %v", args)

	if len(args) != 2 {
		customErr := &custom_errors.LengthCheckingError{Type: "RemovePermissionOrgAdmin arguments length"}
		logger.Errorf(customErr.Error())
		return nil, errors.New(customErr.Error())
	}

	userID := args[0]
	orgID := args[1]

	if utils.IsStringEmpty(userID) {
		customErr := &custom_errors.LengthCheckingError{Type: "userID"}
		logger.Errorf(customErr.Error())
		return nil, errors.WithStack(customErr)
	}

	if utils.IsStringEmpty(orgID) {
		customErr := &custom_errors.LengthCheckingError{Type: "orgID"}
		logger.Errorf(customErr.Error())
		return nil, errors.WithStack(customErr)
	}

	// Check if user is already admin of org
	isAdmin, _, err := user_groups.IsUserAdminOfGroup(stub, userID, orgID)
	if err != nil {
		customErr := &custom_errors.NotGroupAdminError{UserID: userID, GroupID: orgID}
		logger.Errorf("%v: %v", customErr, err)
		return nil, errors.Wrap(err, customErr.Error())
	}

	// if user is not an admin already, return
	if !isAdmin {
		return nil, nil
	}

	// convert caller to solution user object to get org field
	solutionCaller := convertToSolutionUser(caller)

	// construct new caller object representing org itself
	org, err := user_mgmt.GetUserData(stub, caller, solutionCaller.Org, true, false)
	if err != nil {
		customErr := &GetOrgError{Org: solutionCaller.Org}
		logger.Errorf("%v: %v", customErr, err)
		return nil, errors.Wrap(err, customErr.Error())
	}

	if org.PrivateKey == nil {
		errMsg := "Caller does not have access to org private key"
		logger.Errorf(errMsg)
		return nil, errors.New(errMsg)
	}
	orgCaller := org

	solutionUser, err := GetSolutionUserWithParams(stub, orgCaller, userID, true, true)
	if err != nil {
		customErr := &GetUserError{User: userID}
		logger.Errorf("%v: %v", customErr, err)
		return nil, errors.Wrap(err, customErr.Error())
	}

	if solutionUser.PrivateKey == nil {
		logger.Errorf("Caller does not have access to add permission")
		return nil, errors.New("Caller does not have access to add permission")
	}

	// Check user role and org
	if solutionUser.Role != SOLUTION_ROLE_USER {
		logger.Errorf("Org user role must be user")
		return nil, errors.New("Org user role must be user")
	}

	if orgID != solutionUser.Org {
		logger.Errorf("Org mismatch")
		return nil, errors.New("Org mismatch")
	}

	solutionUser.SolutionInfo.IsOrgAdmin = false
	platformUser, err := convertToPlatformUser(stub, solutionUser)
	if err != nil {
		errMsg := "Failed to convertToPlatformUser"
		logger.Errorf("%v: %v", errMsg, err)
		return nil, errors.Wrap(err, errMsg)
	}

	err = user_mgmt.RegisterUserWithParams(stub, orgCaller, platformUser, false)
	if err != nil {
		logger.Errorf("Failed to update SolutionPrivateData for user")
		return nil, errors.Wrap(err, "Failed to update SolutionPrivateData for user")
	}

	// Then remove admin permission
	return user_groups.RemoveAdminPermissionOfGroup(stub, orgCaller, args)
}

// PutUserInOrg adds user to an org, also updates solution user object
// Can only be called by org admins
// args = [ userID, orgID, permissionFlag ]
func PutUserInOrg(stub cached_stub.CachedStubInterface, caller data_model.User, args []string) ([]byte, error) {
	defer utils.ExitFnLog(utils.EnterFnLog())
	logger.Debugf("args: %v", args)

	if len(args) != 3 {
		customErr := &custom_errors.LengthCheckingError{Type: "PutUserInOrg arguments length"}
		logger.Errorf(customErr.Error())
		return nil, errors.New(customErr.Error())
	}

	userID := args[0]
	orgID := args[1]
	adminFlag := args[2]

	if utils.IsStringEmpty(userID) {
		customErr := &custom_errors.LengthCheckingError{Type: "userID"}
		logger.Errorf(customErr.Error())
		return nil, errors.WithStack(customErr)
	}

	if utils.IsStringEmpty(orgID) {
		customErr := &custom_errors.LengthCheckingError{Type: "orgID"}
		logger.Errorf(customErr.Error())
		return nil, errors.WithStack(customErr)
	}

	// Check if futher actions are needed
	if adminFlag == "false" {
		// Check if user is already in group
		inGroup, err := user_groups.IsUserInGroup(stub, userID, orgID)
		if err != nil {
			var errMsg = "IsUserInGroup Error"
			logger.Errorf("%v: %v", errMsg, err)
			return nil, errors.Wrap(err, errMsg)
		}

		// if user is already in group, return
		// Do not care if admin permission is set
		// TODO: update permission even if already in group
		if inGroup {
			return nil, nil
		}
	} else if adminFlag == "true" {
		// Check if user is already admin of group
		isAdmin, _, err := user_groups.IsUserAdminOfGroup(stub, userID, orgID)
		if err != nil {
			customErr := &custom_errors.NotGroupAdminError{UserID: userID, GroupID: orgID}
			logger.Errorf("%v: %v", customErr, err)
			return nil, errors.Wrap(err, customErr.Error())
		}

		// if user is admin already, return
		if isAdmin {
			return nil, nil
		}
	} else {
		logger.Errorf("Error: admin flag must be true or false")
		return nil, errors.New("Error: permissionFlag must be true or false")
	}

	// Get user
	orgUser, err := user_mgmt.GetUserData(stub, caller, userID, true, true)
	if err != nil {
		customErr := &GetUserError{User: userID}
		logger.Errorf("%v: %v", customErr, err)
		return nil, errors.Wrap(err, customErr.Error())
	}

	if orgUser.PrivateKey == nil {
		logger.Errorf("Caller does not have access to add permission")
		return nil, errors.New("Caller does not have access to add permission")
	}

	solutionUser := convertToSolutionUser(orgUser)

	// Check user role and org
	if solutionUser.Role != SOLUTION_ROLE_USER {
		logger.Errorf("Org user role must be user")
		return nil, errors.New("Org user role must be user")
	}

	if utils.IsStringEmpty(solutionUser.Org) {
		logger.Errorf("User org field is empty")
		return nil, errors.New("User org field is empty")
	}

	// We need to add access from org to user so org admin can update user private data when updating permissions
	// user_mgmt.PutUserInOrg does not do this
	// First get group sym key
	group, err := user_mgmt.GetUserData(stub, caller, solutionUser.Org, true, false)
	if err != nil {
		customErr := &GetUserError{User: solutionUser.ID}
		logger.Errorf("%v: %v", customErr, err)
		return nil, errors.Wrap(err, customErr.Error())
	}

	if group.SymKey == nil {
		logger.Errorf("Caller does not have access to group sym key")
		return nil, errors.New("Caller does not have access to group sym key")
	}

	userAccessManager := user_access_ctrl.GetUserAccessManager(stub, caller)
	err = userAccessManager.AddAccessByKey(group.GetPrivateKey(), orgUser.GetPrivateKeyHashSymKey())
	if err != nil {
		customErr := &custom_errors.AddAccessError{Key: "Org privateKey to orgUser symPrvKey"}
		logger.Errorf("%v: %v", customErr, err)
		return nil, errors.Wrap(err, customErr.Error())
	}

	// Put user in group
	// construct new caller object representing group itself
	groupCaller := group
	_, err = user_mgmt.PutUserInOrg(stub, groupCaller, args)
	if err != nil {
		logger.Errorf("Failed to put user in org")
		return nil, errors.Wrapf(err, "Failed to put user in org")
	}

	// If admin flag is set to true, update orgAdmin field
	if adminFlag == "true" {
		solutionUser.SolutionInfo.IsOrgAdmin = true
		orgUser, err = convertToPlatformUser(stub, solutionUser)
		if err != nil {
			errMsg := "Failed to convertToPlatformUser"
			logger.Errorf("%v: %v", errMsg, err)
			return nil, errors.Wrap(err, errMsg)
		}
		return nil, user_mgmt.RegisterUserWithParams(stub, groupCaller, orgUser, false)
	} else {
		return nil, nil
	}
}

// RemoveUserFromOrg removes org user from org, also updates solution user object
// Can only be called by org admins
// args = [ userID, orgID ]
func RemoveUserFromOrg(stub cached_stub.CachedStubInterface, caller data_model.User, args []string) ([]byte, error) {
	defer utils.ExitFnLog(utils.EnterFnLog())
	logger.Debugf("args: %v", args)

	if len(args) != 2 {
		customErr := &custom_errors.LengthCheckingError{Type: "RemoveUserFromOrg arguments length"}
		logger.Errorf(customErr.Error())
		return nil, errors.New(customErr.Error())
	}

	userID := args[0]
	orgID := args[1]

	if utils.IsStringEmpty(userID) {
		customErr := &custom_errors.LengthCheckingError{Type: "userID"}
		logger.Errorf(customErr.Error())
		return nil, errors.WithStack(customErr)
	}

	if utils.IsStringEmpty(orgID) {
		customErr := &custom_errors.LengthCheckingError{Type: "orgID"}
		logger.Errorf(customErr.Error())
		return nil, errors.WithStack(customErr)
	}

	// Check if user is already in group
	inGroup, err := user_groups.IsUserInGroup(stub, userID, orgID)
	if err != nil {
		var errMsg = "IsUserInGroup Error"
		logger.Errorf("%v: %v", errMsg, err)
		return nil, errors.Wrap(err, errMsg)
	}

	// if user is already not in group, return
	// Do not care if admin permission is set
	if !inGroup {
		return nil, nil
	}

	solutionUser, err := GetSolutionUserWithParams(stub, caller, userID, true, true)
	if err != nil {
		customErr := &GetUserError{User: userID}
		logger.Errorf("%v: %v", customErr, err)
		return nil, errors.Wrap(err, customErr.Error())
	}

	if solutionUser.PrivateKey == nil {
		logger.Errorf("Caller does not have access to add permission")
		return nil, errors.New("Caller does not have access to add permission")
	}

	// construct new caller object representing org itself
	group, err := user_mgmt.GetUserData(stub, caller, orgID, true, false)
	if err != nil {
		customErr := &GetUserError{User: solutionUser.ID}
		logger.Errorf("%v: %v", customErr, err)
		return nil, errors.Wrap(err, customErr.Error())
	}

	if group.SymKey == nil {
		logger.Errorf("Caller does not have access to group sym key")
		return nil, errors.New("Caller does not have access to group sym key")
	}

	groupCaller := group

	solutionUser.SolutionInfo.IsOrgAdmin = false
	platformUser, err := convertToPlatformUser(stub, solutionUser)
	if err != nil {
		errMsg := "Failed to convertToPlatformUser"
		logger.Errorf("%v: %v", errMsg, err)
		return nil, errors.Wrap(err, errMsg)
	}
	err = user_mgmt.RegisterUserWithParams(stub, groupCaller, platformUser, false)
	if err != nil {
		logger.Errorf("Failed to update SolutionPrivateData for user")
		return nil, errors.New("Failed to update SolutionPrivateData for user")
	}

	// Then remove admin permission
	return user_groups.RemoveUserFromGroup(stub, groupCaller, args)
}

// RegisterOrg is the solution level register org function
// args = [ org ]
func RegisterOrg(stub cached_stub.CachedStubInterface, caller data_model.User, args []string) ([]byte, error) {
	defer utils.ExitFnLog(utils.EnterFnLog())
	logger.Debugf("RegisterOrg args: %v", args)

	if len(args) != 1 {
		customErr := &custom_errors.LengthCheckingError{Type: "Solution RegisterOrg arguments length"}
		logger.Errorf(customErr.Error())
		return nil, errors.New(customErr.Error())
	}

	var org = data_model.User{}
	orgBytes := []byte(args[0])
	err := json.Unmarshal(orgBytes, &org)
	if err != nil {
		customErr := &custom_errors.UnmarshalError{Type: "org"}
		logger.Errorf("%v: %v", customErr, err)
		return nil, errors.Wrap(err, customErr.Error())
	}

	valid, err := ValidateOrg(stub, caller, org)
	if err != nil {
		logger.Errorf("Error validating org: %v", err)
		return nil, errors.Wrap(err, "Error validating org")
	} else if !valid {
		logger.Errorf("Invalid org!")
		return nil, errors.New("Invalid org!")
	}

	solutionOrg := convertToSolutionOrg(org)

	publicData := make(map[string]interface{})
	publicData["type"] = "organization"
	solutionOrg.SolutionPublicData = publicData

	return RegisterOrgInternal(stub, caller, solutionOrg)
}

// UpdateOrg is the solution level update org function
// args = [ org ]
func UpdateOrg(stub cached_stub.CachedStubInterface, caller data_model.User, args []string) ([]byte, error) {
	defer utils.ExitFnLog(utils.EnterFnLog())
	logger.Debugf("UpdateOrg args: %v", args)

	if len(args) != 1 {
		customErr := &custom_errors.LengthCheckingError{Type: "Solution UpdateOrg arguments length"}
		logger.Errorf(customErr.Error())
		return nil, errors.New(customErr.Error())
	}

	var org = data_model.User{}
	orgBytes := []byte(args[0])
	err := json.Unmarshal(orgBytes, &org)
	if err != nil {
		customErr := &custom_errors.UnmarshalError{Type: "org"}
		logger.Errorf("%v: %v", customErr, err)
		return nil, errors.Wrap(err, customErr.Error())
	}

	solutionOrg := convertToSolutionOrg(org)

	return RegisterOrgInternal(stub, caller, solutionOrg)
}

// RegisterOrgInternal validates and creates/updates an org
// The caller will be added as an admin of the org if this is a new org and makeCallerAdmin == true
func RegisterOrgInternal(stub cached_stub.CachedStubInterface, caller data_model.User, solutionOrg SolutionUser) ([]byte, error) {
	defer utils.ExitFnLog(utils.EnterFnLog())
	logger.Debugf("solutionOrg: %v", solutionOrg)

	existingOrg, err := GetSolutionUserWithParams(stub, caller, solutionOrg.ID, true, true)
	if err != nil {
		customErr := &GetUserError{User: solutionOrg.ID}
		logger.Errorf("%v: %v", customErr, err)
		return nil, errors.Wrap(err, customErr.Error())
	}

	// If org already exists, this is an update
	if !utils.IsStringEmpty(existingOrg.ID) {
		// Check that caller has access to org
		if caller.ID != solutionOrg.ID && existingOrg.PrivateKey == nil {
			logger.Errorf("Caller does not have access to org")
			return nil, errors.New("Caller does not have access to user")
		}

		// Keep org the same during update
		solutionOrg.Org = existingOrg.Org

		// during an update, keep org admin flag and services fields in solution private data
		solutionOrg.SolutionInfo.IsOrgAdmin = existingOrg.SolutionInfo.IsOrgAdmin
		solutionOrg.SolutionInfo.Services = existingOrg.SolutionInfo.Services
	} else {
		if utils.IsStringEmpty(solutionOrg.Org) {
			logger.Errorf("org field cannot be empty")
			return nil, errors.New("org field cannot be empty")
		}
	}

	// must marshall to register using user_mgmt
	platformOrg, err := convertToPlatformUser(stub, solutionOrg)
	if err != nil {
		errMsg := "Failed to convertToPlatformUser"
		logger.Errorf("%v: %v", errMsg, err)
		return nil, errors.Wrap(err, errMsg)
	}
	return nil, user_mgmt.RegisterOrgWithParams(stub, caller, platformOrg, false)
}

// ValidateOrg is the solution equivalent of user_mgmt validateOrg which checks that the org is valid
// Makes sure role is org
func ValidateOrg(stub cached_stub.CachedStubInterface, caller data_model.User, org data_model.User) (bool, error) {
	defer utils.ExitFnLog(utils.EnterFnLog())
	logger.Debugf("org: %v", org)

	// org.IsGroup should be true
	if org.IsGroup != true {
		customErr := &custom_errors.RegisterOrgInvalidFieldError{ID: org.ID, Field: "IsGroup"}
		logger.Errorf("%v: %v", customErr.Error(), org.IsGroup)
		return false, errors.New(customErr.Error())
	}
	// org.Role should be "org"
	if org.Role != SOLUTION_ROLE_ORG {
		customErr := &custom_errors.RegisterOrgInvalidFieldError{ID: org.ID, Field: "Role"}
		logger.Errorf("%v: %v", customErr.Error(), org.Role)
		return false, errors.New(customErr.Error())
	}

	return true, nil
}

// helper function to convert service of type []interface{} to a slice
func ConvertServicesToSlice(services []interface{}) []string {
	defer utils.ExitFnLog(utils.EnterFnLog())

	var list []string
	for _, v := range services {
		list = append(list, v.(string))
	}
	return list
}

// Caller must pass in solution level role and org
// user type is optional; user type is used to distinguish organization and services
func convertToPlatformUser(stub cached_stub.CachedStubInterface, solutionUser SolutionUser) (data_model.User, error) {
	defer utils.ExitFnLog(utils.EnterFnLog())

	user := data_model.User{}
	user.ID = solutionUser.ID
	user.Name = solutionUser.Name
	user.Role = solutionUser.Role
	user.PublicKey = solutionUser.PublicKey
	user.PublicKeyB64 = solutionUser.PublicKeyB64
	user.PrivateKey = solutionUser.PrivateKey
	user.PrivateKeyB64 = solutionUser.PrivateKeyB64
	user.SymKey = solutionUser.SymKey
	user.SymKeyB64 = solutionUser.SymKeyB64
	user.IsGroup = solutionUser.IsGroup
	user.Status = solutionUser.Status

	publicData := make(map[string]interface{})
	publicData["org"] = solutionUser.Org
	if solutionUser.SolutionPublicData != nil {
		if solutionUser.SolutionPublicData.(map[string]interface{})["type"] != nil {
			orgType := solutionUser.SolutionPublicData.(map[string]interface{})["type"]
			publicData["type"] = orgType
		}
	}
	user.SolutionPublicData = publicData

	//private data
	user.Email = solutionUser.Email
	user.KmsPublicKeyId = solutionUser.KmsPublicKeyId
	user.KmsPrivateKeyId = solutionUser.KmsPrivateKeyId
	user.KmsSymKeyId = solutionUser.KmsSymKeyId
	user.Secret = solutionUser.Secret

	privateData := make(map[string]interface{})
	privateData["services"] = solutionUser.SolutionInfo.Services
	privateData["is_org_admin"] = solutionUser.SolutionInfo.IsOrgAdmin
	privateData["data"] = solutionUser.SolutionPrivateData
	if solutionUser.SolutionPrivateData.(map[string]interface{})["tax_id"] != nil {
		privateData["tax_id"] = solutionUser.SolutionPrivateData.(map[string]interface{})["tax_id"]
	}
	if solutionUser.SolutionPrivateData.(map[string]interface{})["address"] != nil {
		privateData["address"] = solutionUser.SolutionPrivateData.(map[string]interface{})["address"]
	}

	user.SolutionPrivateData = privateData

	// get off-chain datastore connection id, if one is setup
	dsConnectionID, err := GetActiveConnectionID(stub)
	if err != nil {
		errMsg := "Failed to GetActiveConnectionID"
		logger.Errorf("%v: %v", errMsg, err)
		return data_model.User{}, errors.Wrap(err, errMsg)
	}
	user.ConnectionID = dsConnectionID

	return user, nil
}

func convertToSolutionOrg(org data_model.User) SolutionUser {
	defer utils.ExitFnLog(utils.EnterFnLog())

	solutionOrg := SolutionUser{}

	solutionOrg.ID = org.ID
	solutionOrg.Name = org.Name
	solutionOrg.Role = org.Role
	solutionOrg.Org = org.ID
	solutionOrg.PublicKey = org.PublicKey
	solutionOrg.PublicKeyB64 = org.PublicKeyB64
	solutionOrg.PrivateKey = org.PrivateKey
	solutionOrg.PrivateKeyB64 = org.PrivateKeyB64
	solutionOrg.SymKey = org.SymKey
	solutionOrg.SymKeyB64 = org.SymKeyB64
	solutionOrg.IsGroup = org.IsGroup
	solutionOrg.Status = org.Status
	if org.SolutionPublicData == nil {
		solutionOrg.SolutionPublicData = make(map[string]interface{})
	} else {
		solutionOrg.SolutionPublicData = org.SolutionPublicData
	}

	//private data
	solutionOrg.Email = org.Email
	solutionOrg.KmsPublicKeyId = org.KmsPublicKeyId
	solutionOrg.KmsPrivateKeyId = org.KmsPrivateKeyId
	solutionOrg.KmsSymKeyId = org.KmsSymKeyId
	solutionOrg.Secret = org.Secret
	solutionOrg.SolutionInfo.Services = []string{}
	solutionOrg.SolutionInfo.IsOrgAdmin = true
	solutionOrg.SolutionPrivateData = org.SolutionPrivateData

	return solutionOrg
}

func convertToSolutionUser(user data_model.User) SolutionUser {
	defer utils.ExitFnLog(utils.EnterFnLog())

	solutionUser := SolutionUser{}
	solutionUser.ID = user.ID
	solutionUser.Name = user.Name
	solutionUser.Role = user.Role

	if user.SolutionPublicData != nil {
		if user.SolutionPublicData.(map[string]interface{})["org"] != nil {
			solutionUser.Org = user.SolutionPublicData.(map[string]interface{})["org"].(string)
		}
	}
	solutionUser.PublicKey = user.PublicKey
	solutionUser.PublicKeyB64 = user.PublicKeyB64
	solutionUser.PrivateKey = user.PrivateKey
	solutionUser.PrivateKeyB64 = user.PrivateKeyB64
	solutionUser.SymKey = user.SymKey
	solutionUser.SymKeyB64 = user.SymKeyB64
	solutionUser.IsGroup = user.IsGroup
	solutionUser.Status = user.Status
	solutionUser.SolutionPublicData = make(map[string]interface{})

	//private data
	solutionUser.Email = user.Email
	solutionUser.KmsPublicKeyId = user.KmsPublicKeyId
	solutionUser.KmsPrivateKeyId = user.KmsPrivateKeyId
	solutionUser.KmsSymKeyId = user.KmsSymKeyId
	solutionUser.Secret = user.Secret

	if user.SolutionPrivateData != nil {
		if user.SolutionPrivateData.(map[string]interface{})["services"] != nil {
			platformUserServices := user.SolutionPrivateData.(map[string]interface{})["services"]
			// future improvement, find a way to update services without copying it to a slice first; currently it is O(n)
			solutionUser.SolutionInfo.Services = ConvertServicesToSlice(platformUserServices.([]interface{}))
			// convert slice to array if slice is empty
			if len(solutionUser.SolutionInfo.Services) <= 0 {
				solutionUser.SolutionInfo.Services = []string{}
			}
		} else {
			solutionUser.SolutionInfo.Services = []string{}
		}

		if user.SolutionPrivateData.(map[string]interface{})["is_org_admin"] != nil {
			platformUserOrgAdminFlag := user.SolutionPrivateData.(map[string]interface{})["is_org_admin"]
			solutionUser.SolutionInfo.IsOrgAdmin = platformUserOrgAdminFlag.(bool)
		} else {
			solutionUser.SolutionInfo.IsOrgAdmin = false
		}

		// In Golang, if the map string interface only has 1 key and value, it only stores the value, so we have to check
		// if the other 2 possible keys are nil
		if user.SolutionPrivateData.(map[string]interface{})["services"] == nil && user.SolutionPrivateData.(map[string]interface{})["is_org_admin"] == nil {
			solutionUser.SolutionPrivateData = user.SolutionPrivateData
		} else {
			if user.SolutionPrivateData.(map[string]interface{})["data"] != nil {
				solutionUser.SolutionPrivateData = user.SolutionPrivateData.(map[string]interface{})["data"]
			}
		}
	} else {
		solutionUser.SolutionInfo.Services = []string{}
		solutionUser.SolutionInfo.IsOrgAdmin = false
	}

	return solutionUser
}

func getCallerOrgAdminInfo(caller data_model.User) (string, bool) {
	orgID := ""
	isOrgAdmin := false
	if caller.SolutionPublicData != nil {
		if caller.SolutionPublicData.(map[string]interface{})["org"] != nil {
			orgID = caller.SolutionPublicData.(map[string]interface{})["org"].(string)
		}
	}

	if caller.SolutionPrivateData != nil {
		if caller.SolutionPrivateData.(map[string]interface{})["is_org_admin"] != nil {
			platformUserOrgAdminFlag := caller.SolutionPrivateData.(map[string]interface{})["is_org_admin"]
			isOrgAdmin = platformUserOrgAdminFlag.(bool)
		}
	}

	return orgID, isOrgAdmin
}
