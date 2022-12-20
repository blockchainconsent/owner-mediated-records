/*******************************************************************************
 *
 *
 * (c) Copyright Merative US L.P. and others 2020-2022 
 *
 * SPDX-Licence-Identifier: Apache 2.0
 *
 *******************************************************************************/

'use strict';

// For logging
var TAG = 'solution_chaincode_ops.js';
var log4js = require('log4js');
var logger = log4js.getLogger(TAG);
logger.level = 'DEBUG';

var request = require('request');
var util = require('util');
var chaincodeOps = require("common-utils/chaincode_ops");

var peers = null;
var isHttps = false;

/**
 * Get all orgs
 */
module.exports.getOrgs = getOrgs;
function getOrgs(caller, cb) {
    logger.debug('get orgs:');

    let fcn = 'getOrgs';
    let args = [];


    chaincodeOps._query(caller, fcn, args, function (err, result) {
        if (err) {
            var errmsg = "Failed to get orgs";
            logger.error(errmsg, err);
            try {
                cb && cb(new Error(errmsg));
            } catch (err) {
                logger.error(util.format("callback for " + fcn + " failed: error: %s", err));
            }
        }
        else {
            logger.debug('got orgs successfully:', result);
            try {
                cb && cb(null, result);
            } catch (err) {
                logger.error(util.format("callback for " + fcn + " failed: error: %s", err));
            }
        }
    });
};

/**
 * Get an org
 * @param orgId Id of org
 */
module.exports.getOrg = getOrg;
function getOrg(caller, orgId, cb) {
    logger.debug('get org:');

    let fcn = 'getOrg';
    let args = [orgId];

    chaincodeOps._query(caller, fcn, args, function (err, value) {
        if (err) {
            var errmsg = "failed to get org:";
            logger.error(errmsg, err);
            try {
                cb && cb(new Error(errmsg));
            } catch (err) {
                logger.error(util.format("callback for " + fcn + " failed: error: %s", err));
            }
        }
        else {
            logger.debug('got org successfully:', value);
            try {
                cb && cb(null, value);
            } catch (err) {
                logger.error(util.format("callback for " + fcn + " failed: error: %s", err));
            }
        }
    });
};

/**
 * Register a org
 * @param orgInfo JSON
 */
module.exports.registerOrg = registerOrg;
function registerOrg(caller, orgInfo, cb) {
    logger.debug('Register org:', orgInfo);

    var args = [];
    var orgInfoStr = "";
    try {
        orgInfoStr = JSON.stringify(orgInfo);
    } catch (err) {
        var errmsg = "Invalid org info";
        logger.error(errmsg, err);
        cb && cb(new Error(errmsg));
        return;
    }

    args.push(orgInfoStr);

    var fcn = "registerOrg";
    if (orgInfo.role != "org") {
        cb && cb(new Error('Invalid org role: ' + orgInfo.role));
    } else {
        chaincodeOps._invoke(caller, fcn, args, function (err, result) {
            if (err) {
                var errmsg = "Failed to register org";
                logger.error(errmsg, err);
                try {
                    cb && cb(new Error(errmsg));
                } catch (err) {
                    logger.error(util.format("callback for " + fcn + " failed: error: %s", err));
                }
            }
            else {
                logger.debug('Registered org successfully:', result);
                try {
                    cb && cb(null, result);
                } catch (err) {
                    logger.error(util.format("callback for " + fcn + " failed: error: %s", err));
                }
            }
        });
    }
}

/**
 * Update organization
 * @param org Org object
 */
module.exports.updateOrg = updateOrg;
function updateOrg(caller, orgData, cb) {
    logger.debug('Update org');

    var org = {
        id: orgData.id,
        name: orgData.name,
        data: orgData.data
    };

    let fcn = 'updateOrg';
    let args = [JSON.stringify(org)];

    chaincodeOps._invoke(caller, fcn, args, function (err, result) {
        if (err) {
            var errmsg = "Failed to update org";
            logger.error(errmsg, err);
            try {
                cb && cb(new Error(errmsg));
            } catch (err) {
                logger.error(util.format("callback for " + fcn + " failed: error: %s", err));
            }
        }
        else {
            logger.debug('Updated org successfully:', result);
            try {
                cb && cb(null, result);
            } catch (err) {
                logger.error(util.format("callback for " + fcn + " failed: error: %s", err));
            }
        }
    });
}

/**
 * Get a user
 */
module.exports.getUser = getUser;
function getUser(caller, userId, cb) {
    logger.debug('get user:', userId);

    var fcn = 'getUser';
    var args = [userId];

    chaincodeOps._query(caller, fcn, args, function (err, value) {
        if (err) {
            var errmsg = "Failed to get user";
            logger.error(errmsg, err);
            try {
                cb && cb(new Error(errmsg));
            } catch (err) {
                logger.error(util.format("callback for " + fcn + " failed: error: %s", err));
            }
        }
        else {
            logger.debug('retrieved user:', value);
            try {
                cb && cb(null, value);
            } catch (err) {
                logger.error(util.format("callback for " + fcn + " failed: error: %s", err));
            }
        }
    });
}

/**
 * Get a login user
 * @param userId Login User Id
 */
module.exports.getLoginUser = getLoginUser;
function getLoginUser(caller, userId, cb) {
    logger.debug('get login user:', userId);

    let fcn = 'getLoginUser';
    let args = [userId];

    chaincodeOps._query(caller, fcn, args, function (err, result) {
        if (err) {
            var errmsg = "Failed to get login user";
            logger.error(errmsg, err);
            try {
                cb && cb(new Error(errmsg));
            } catch (err) {
                logger.error(util.format("callback for " + fcn + " failed: error: %s", err));
            }
        }
        else {
            logger.debug('got login user successfully:', result);
            try {
                cb && cb(null, result);
            } catch (err) {
                logger.error(util.format("callback for " + fcn + " failed: error: %s", err));
            }
        }
    });
};

/**
 * Update a login user
 */
module.exports.updateLoginUser = updateLoginUser;
function updateLoginUser(caller, userData, cb) {
    logger.debug('Update login user:', userData.id);

    var userDataStr = "";

    try {
        userDataStr = JSON.stringify(userData);
    } catch (err) {
        var errmsg = "Failed to convert user data to string";
        logger.error(errmsg, err);
        cb && cb(new Error(errmsg));
        return;
    }

    var fcn = 'updateLoginUser';
    var args = [userDataStr];

    chaincodeOps._invoke(caller, fcn, args, function (err, result) {
        if (err) {
            var errmsg = "Failed to update login user";
            logger.error(errmsg, err);
            try {
                cb && cb(new Error(errmsg));
            } catch (err) {
                logger.error(util.format("callback for " + fcn + " failed: error: %s", err));
            }
        }
        else {
            logger.debug('Updated login user successfully:', result);
            try {
                cb && cb(null, result);
            } catch (err) {
                logger.error(util.format("callback for " + fcn + " failed: error: %s", err));
            }
        }
    });
}

/*
 * Register a login user
 */
module.exports.registerLoginUser = registerLoginUser;
function registerLoginUser(caller, userInfo, cb) {
    logger.debug('Register login user:', userInfo);

    var userInfoStr = "";
    try {
        userInfoStr = JSON.stringify(userInfo);
    } catch (err) {
        var errmsg = "Invalid login user";
        logger.error(errmsg, err);
        cb && cb(new Error(errmsg));
    }

    let fcn = 'registerLoginUser';
    let args = [userInfoStr];

    chaincodeOps._invoke(caller, fcn, args, function (err, result) {
        if (err) {
            var errmsg = "Failed to register login user";
            logger.error(errmsg, err);
            try {
                cb && cb(new Error(errmsg));
            } catch (err) {
                logger.error(util.format("callback for " + fcn + " failed: error: %s", err));
            }
        }
        else {
            logger.info('Registered login user successfully:', result);
            try {
                cb && cb(null, result);
            } catch (err) {
                logger.error(util.format("callback for " + fcn + " failed: error: %s", err));
            }
        }
    });
}

/**
 * Register a user
 * @param userInfo JSON
 */
module.exports.registerUser = registerUser;
function registerUser(caller, userInfo, cb) {
    logger.debug('Register user:', userInfo);

    var userInfoStr = "";
    try {
        userInfoStr = JSON.stringify(userInfo);
    } catch (err) {
        var errmsg = "Invalid user info";
        logger.error(errmsg, err);
        cb && cb(new Error(errmsg));
        return;
    }

    let fcn = 'registerUser';
    let args = [userInfoStr];

    var updateUser = false

    chaincodeOps._invoke(caller, fcn, args, function (err, result) {
        if (err) {
            var errmsg = "Failed to register user";
            logger.error(errmsg, err);
            try {
                cb && cb(new Error(errmsg));
            } catch (err) {
                logger.error(util.format("callback for " + fcn + " failed: error: %s", err));
            }
        }
        else {
            logger.debug('Registered user successfully:', result);
            // If user role is org and this is a new user registration
            if (userInfo.role == "org" && !updateUser) {
                var fcn = 'putUserInOrg';
                var args = [userInfo.id, userInfo.org, "false"];

                chaincodeOps._invoke(caller, fcn, args, function (err) {
                    if (err) {
                        var errmsg = "Failed to add org user to org";
                        logger.error(errmsg, err);
                        try {
                            cb && cb(new Error(errmsg));
                        } catch (err) {
                            logger.error(util.format("callback for " + fcn + " failed: error: %s", err));
                        }
                    }
                    else {
                        logger.debug('Added org user to org successfully:');
                        logger.debug('successfully added org user to org:', result);
                        try {
                            cb && cb(null, result);
                        } catch (err) {
                            logger.error(util.format("callback for " + fcn + " failed: error: %s", err));
                        }
                    }
                });
            } else {
                try {
                    cb && cb(null, result);
                } catch (err) {
                    logger.error(util.format("callback for " + fcn + " failed: error: %s", err));
                }
            }
        }
    });
}

/**
 * Put user in group
 */
module.exports.putUserInOrg = putUserInOrg;
function putUserInOrg(caller, userID, orgID, isAdmin, cb) {
    logger.debug('Put user in group:', userID, orgID, isAdmin);

    var fcn = 'putUserInOrg';
    var args = [userID, orgID, isAdmin];

    chaincodeOps._invoke(caller, fcn, args, function (err, result) {
        if (err) {
            var errmsg = "Failed to put user in group";
            logger.error(errmsg, err);
            try {
                cb && cb(new Error(errmsg));
            } catch (err) {
                logger.error(util.format("callback for " + fcn + " failed: error: %s", err));
            }
        }
        else {
            logger.debug('Put user in group successfully:', result);
            try {
                cb && cb(null, result);
            } catch (err) {
                logger.error(util.format("callback for " + fcn + " failed: error: %s", err));
            }
        }
    });
}

/**
 * Remove user from org
 */
module.exports.removeUserFromOrg = removeUserFromOrg;
function removeUserFromOrg(caller, userID, orgID, cb) {
    logger.debug('Remove user from group:', userID, orgID);

    var fcn = 'removeUserFromOrg';
    var args = [userID, orgID];

    chaincodeOps._invoke(caller, fcn, args, function (err, result) {
        if (err) {
            var errmsg = "Failed to remove user from group";
            logger.error(errmsg, err);
            try {
                cb && cb(new Error(errmsg));
            } catch (err) {
                logger.error(util.format("callback for " + fcn + " failed: error: %s", err));
            }
        }
        else {
            logger.debug('Removed user from group successfully:', result);
            try {
                cb && cb(null, result);
            } catch (err) {
                logger.error(util.format("callback for " + fcn + " failed: error: %s", err));
            }
        }
    });
}

/**
 * Add org admin permission
 * @param userID
 * @param orgID
 */
module.exports.addPermissionOrgAdmin = addPermissionOrgAdmin;
function addPermissionOrgAdmin(caller, userId, orgId, cb) {
    logger.debug('add org admin permission:', userId, orgId);

    var fcn = 'addPermissionOrgAdmin';
    var args = [userId, orgId];

    chaincodeOps._invoke(caller, fcn, args, function (err, result) {
        if (err) {
            var errmsg = "Failed to add org admin permission:";
            logger.error(errmsg, err);
            try {
                cb && cb(new Error(errmsg));
            } catch (err) {
                logger.error(util.format("callback for " + fcn + " failed: error: %s", err));
            }
        }
        else {
            logger.debug('Added org admin permission successfully:', result);
            try {
                cb && cb(null, result);
            } catch (err) {
                logger.error(util.format("callback for " + fcn + " failed: error: %s", err));
            }
        }
    });
}

/**
 * Delete org admin permission
 * @param userID
 * @param orgID
 */
module.exports.deletePermissionOrgAdmin = deletePermissionOrgAdmin;
function deletePermissionOrgAdmin(caller, userId, orgId, cb) {
    logger.debug('delete org admin permission:', userId, orgId);

    var fcn = 'deletePermissionOrgAdmin';
    var args = [userId, orgId];

    chaincodeOps._invoke(caller, fcn, args, function (err, result) {
        if (err) {
            var errmsg = "Failed to delete org admin permission:";
            logger.error(errmsg, err);
            try {
                cb && cb(new Error(errmsg));
            } catch (err) {
                logger.error(util.format("callback for " + fcn + " failed: error: %s", err));
            }
        }
        else {
            logger.debug('Removed org admin permission successfully:', result);
            try {
                cb && cb(null, result);
            } catch (err) {
                logger.error(util.format("callback for " + fcn + " failed: error: %s", err));
            }
        }
    });
}

/**
 * Add service admin permission
 * @param userID
 * @param serviceID
 */
module.exports.addPermissionServiceAdmin = addPermissionServiceAdmin;
function addPermissionServiceAdmin(caller, userId, serviceId, cb) {
    logger.debug('add service admin permission:', userId, serviceId);

    var fcn = 'addPermissionServiceAdmin';
    var args = [userId, serviceId];

    chaincodeOps._invoke(caller, fcn, args, function (err, result) {
        if (err) {
            var errmsg = "Failed to add service admin permission:";
            logger.error(errmsg, err);
            try {
                cb && cb(new Error(errmsg));
            } catch (err) {
                logger.error(util.format("callback for " + fcn + " failed: error: %s", err));
            }
        }
        else {
            logger.debug('Added service admin permission successfully:', result);
            try {
                cb && cb(null, result);
            } catch (err) {
                logger.error(util.format("callback for " + fcn + " failed: error: %s", err));
            }
        }
    });
}

/**
 * Delete service admin permission
 * @param userID
 * @param serviceID
 */
module.exports.deletePermissionServiceAdmin = deletePermissionServiceAdmin;
function deletePermissionServiceAdmin(caller, userId, serviceId, cb) {
    logger.debug('delete service admin permission:', userId, serviceId);

    var fcn = 'deletePermissionServiceAdmin';
    var args = [userId, serviceId];

    chaincodeOps._invoke(caller, fcn, args, function (err, result) {
        if (err) {
            var errmsg = "Failed to delete service admin permission:";
            logger.error(errmsg, err);
            try {
                cb && cb(new Error(errmsg));
            } catch (err) {
                logger.error(util.format("callback for " + fcn + " failed: error: %s", err));
            }
        }
        else {
            logger.debug('Removed service admin permission successfully:', result);
            try {
                cb && cb(null, result);
            } catch (err) {
                logger.error(util.format("callback for " + fcn + " failed: error: %s", err));
            }
        }
    });
}

/**
 * Add audit permission
 * @param userID
 * @param serviceID
 */
module.exports.addPermissionAuditor = addPermissionAuditor;
function addPermissionAuditor(caller, userId, serviceId, auditPermissionKeyB64, cb) {
    logger.debug('add auditor permission:', userId, serviceId);

    var fcn = 'addPermissionAuditor';
    var args = [userId, serviceId, auditPermissionKeyB64];

    chaincodeOps._invoke(caller, fcn, args, function (err, result) {
        if (err) {
            var errmsg = "Failed to add auditor admin permission:";
            logger.error(errmsg, err);
            try {
                cb && cb(new Error(errmsg));
            } catch (err) {
                logger.error(util.format("callback for " + fcn + " failed: error: %s", err));
            }
        } else {
            logger.debug('Added auditor admin permission successfully:', result);
            try {
                cb && cb(null, result);
            } catch (err) {
                logger.error(util.format("callback for " + fcn + " failed: error: %s", err));
            }
        }
    });
}

/**
 * Delete audit permission
 * @param userID
 * @param serviceID
 */
module.exports.deletePermissionAuditor = deletePermissionAuditor;
function deletePermissionAuditor(caller, userId, serviceId, cb) {
    logger.debug('delete auditor permission:', userId, serviceId);

    var fcn = 'deletePermissionAuditor';
    var args = [userId, serviceId];

    chaincodeOps._invoke(caller, fcn, args, function (err, result) {
        if (err) {
            var errmsg = "Failed to delete auditor permission";
            logger.error(errmsg, err);
            try {
                cb && cb(new Error(errmsg));
            } catch (err) {
                logger.error(util.format("callback for " + fcn + " failed: error: %s", err));
            }
        }
        else {
            logger.debug('Deleted auditor permission successfully:', result);
            try {
                cb && cb(null, result);
            } catch (err) {
                logger.error(util.format("callback for " + fcn + " failed: error: %s", err));
            }
        }
    });
}

/**
 * Get users
 * @param orgId org id
 * @param role  user role (optional)
 */
module.exports.getUsers = getUsers;
function getUsers(caller, orgId, maxNum, role, cb) {
    logger.debug('get users for:', orgId, maxNum, role);

    var fcn = 'getUsers';
    var args = [orgId, maxNum, role];

    chaincodeOps._query(caller, fcn, args, function (err, value) {
        if (err) {
            var errmsg = "Failed to get users:";
            logger.error(errmsg, err);
            try {
                cb && cb(new Error(errmsg));
            } catch (err) {
                logger.error(util.format("callback for " + fcn + " failed: error: %s", err));
            }
        }
        else {
            logger.debug('got result for get users:', value);
            try {
                cb && cb(null, value);
            } catch (err) {
                logger.error(util.format("callback for " + fcn + " failed: error: %s", err));
            }
        }
    });
};

/**
 * Register data type
 */
module.exports.registerDatatype = registerDatatype;
function registerDatatype(caller, dataType, cb) {
    logger.debug('register datatype:', dataType.id);
    var dataTypeStr = "";
    try {
        dataTypeStr = JSON.stringify(dataType);
    } catch (err) {
        var errmsg = "Invalid data type";
        logger.error(errmsg, err);
        cb && cb(new Error(errmsg));
    }

    var fcn = 'registerDatatype';
	var args = [dataTypeStr];

    chaincodeOps._invoke(caller, fcn, args, function (err, result) {
        if (err) {
            var errmsg = "Failed to register data type";
            logger.error(errmsg, err);
            try {
                cb && cb(new Error(errmsg));
            } catch (err) {
                logger.error(util.format("callback for " + fcn + " failed: error: %s", err));
            }
        }
        else {
            logger.debug('Registered data type successfully:', result);
            try {
                cb && cb(null, result);
            } catch (err) {
                logger.error(util.format("callback for " + fcn + " failed: error: %s", err));
            }
        }
    });
}

/**
 * Update data type
 */
module.exports.updateDatatype = updateDatatype;
function updateDatatype(caller, dataType, cb) {
    logger.debug('update datatype:', dataType.id);

    var datatypeStr = "";

    try {
        datatypeStr = JSON.stringify(dataType);
    } catch (err) {
        var errmsg = "Failed to convert datatype to string";
        logger.error(errmsg, err);
        cb && cb(new Error(errmsg));
        return;
    }

    var fcn = 'updateDatatype';
    var args = [datatypeStr];

    chaincodeOps._invoke(caller, fcn, args, function (err, result) {
        if (err) {
            var errmsg = "Failed to update datatype";
            logger.error(errmsg, err);
            try {
                cb && cb(new Error(errmsg));
            } catch (err) {
                logger.error(util.format("callback for " + fcn + " failed: error: %s", err));
            }
        }
        else {
            logger.debug('Updated datatype successfully:', result);
            try {
                cb && cb(null, result);
            } catch (err) {
                logger.error(util.format("callback for " + fcn + " failed: error: %s", err));
            }
        }
    });
}

/**
 * Get data type
 * @param dataTypeId Datatype ID
 */
module.exports.getDatatype = getDatatype;
function getDatatype(caller, dataTypeId, cb) {
    logger.debug('get data type for:', dataTypeId);

    var fcn = 'getDatatype';
    var args = [dataTypeId];

    chaincodeOps._query(caller, fcn, args, function (err, value) {
        if (err) {
            var errmsg = "Failed to get datatype";
            logger.error(errmsg, err);
            try {
                cb && cb(new Error(errmsg));
            } catch (err) {
                logger.error(util.format("callback for " + fcn + " failed: error: %s", err));
            }
        }
        else {
            logger.debug('successfully got datatype:', value);
            try {
                cb && cb(null, value);
            } catch (err) {
                logger.error(util.format("callback for " + fcn + " failed: error: %s", err));
            }
        }
    });
}

/**
 * Get datatypes
 */
module.exports.getAllDatatypes = getAllDatatypes;
function getAllDatatypes(caller, cb) {
    logger.debug('get datatypes');

    var fcn = 'getAllDatatypes';
    var args = [];

    chaincodeOps._query(caller, fcn, args, function (err, value) {
        if (err) {
            var errmsg = "Failed to get datatypes";
            logger.error(errmsg, err);
            try {
                cb && cb(new Error(errmsg));
            } catch (err) {
                logger.error(util.format("callback for " + fcn + " failed: error: %s", err));
            }
        }
        else {
            logger.debug('successfully got datatypes:', value);
            try {
                cb && cb(null, value);
            } catch (err) {
                logger.error(util.format("callback for " + fcn + " failed: error: %s", err));
            }
        }
    });
}

/**
 * Register a service
 * @param orgInfo JSON
 */
module.exports.registerService = registerService;
function registerService(caller, serviceInfo, cb) {
    logger.debug('Register service:', serviceInfo);

    var args = [];

    var serviceInfoStr = "";
    try {
        serviceInfoStr = JSON.stringify(serviceInfo);
    } catch (err) {
        var errmsg = "Invalid service info";
        logger.error(errmsg, err);
        cb && cb(new Error(errmsg));
        return;
    }

    args.push(serviceInfoStr);

    var fcn = "registerService";

    if (serviceInfo.role != "org") {
        cb && cb(new Error('Invalid service role: ' + serviceInfo.role));
    } else {
        chaincodeOps._invoke(caller, fcn, args, function (err, result) {
            if (err) {
                var errmsg = "Failed to register service";
                logger.error(errmsg, err);
                try {
                    cb && cb(new Error(errmsg));
                } catch (err) {
                    logger.error(util.format("callback for " + fcn + " failed: error: %s", err));
                }
            }
            else {
                logger.debug('Registered service successfully:', result);
                try {
                    cb && cb(null, result);
                } catch (err) {
                    logger.error(util.format("callback for " + fcn + " failed: error: %s", err));
                }
            }
        });
    }
}

/**
 * Update a service
 * @param orgInfo JSON
 */
module.exports.updateService = updateService;
function updateService(caller, serviceInfo, cb) {
    logger.debug('Update service:', serviceInfo);

    var args = [];

    var serviceInfoStr = "";
    try {
        serviceInfoStr = JSON.stringify(serviceInfo);
    } catch (err) {
        var errmsg = "Invalid service info";
        logger.error(errmsg, err);
        cb && cb(new Error(errmsg));
        return;
    }

    args.push(serviceInfoStr);

    var fcn = "updateService";

    if (serviceInfo.role != "org") {
        cb && cb(new Error('Invalid service role: ' + serviceInfo.role));
    } else {
        chaincodeOps._invoke(caller, fcn, args, function (err, result) {
            if (err) {
                var errmsg = "Failed to update service";
                logger.error(errmsg, err);
                try {
                    cb && cb(new Error(errmsg));
                } catch (err) {
                    logger.error(util.format("callback for " + fcn + " failed: error: %s", err));
                }
            }
            else {
                logger.debug('Updated service successfully:', result);
                try {
                    cb && cb(null, result);
                } catch (err) {
                    logger.error(util.format("callback for " + fcn + " failed: error: %s", err));
                }
            }
        });
    }
}

/**
 * Get a service
 * @param service Id of service
 */
module.exports.getService = getService;
function getService(caller, serviceID, cb) {
    logger.debug('get service:');

    let fcn = 'getService';
    let args = [serviceID];

    chaincodeOps._query(caller, fcn, args, function (err, value) {
        if (err) {
            var errmsg = "failed to get service:";
            logger.error(errmsg, err);
            try {
                cb && cb(new Error(errmsg));
            } catch (err) {
                logger.error(util.format("callback for " + fcn + " failed: error: %s", err));
            }
        }
        else {
            logger.debug('got service successfully:', value);
            try {
                cb && cb(null, value);
            } catch (err) {
                logger.error(util.format("callback for " + fcn + " failed: error: %s", err));
            }
        }
    });
};

/**
 * Add datatype to a service
 */
module.exports.addDatatypeToService = addDatatypeToService;
function addDatatypeToService(caller, datatype, serviceID, cb) {
    logger.debug('addDatatypeToService:', datatype);

    var args = [];

    var datatypeStr = "";
    try {
        datatypeStr = JSON.stringify(datatype);
    } catch (err) {
        var errmsg = "Invalid datatype";
        logger.error(errmsg, err);
        cb && cb(new Error(errmsg));
        return;
    }

    args.push(serviceID);
    args.push(datatypeStr);

    var fcn = "addDatatypeToService";
    chaincodeOps._invoke(caller, fcn, args, function (err, result) {
        if (err) {
            var errmsg = "Failed to add datatype to service";
            logger.error(errmsg, err);
            try {
                cb && cb(new Error(errmsg));
            } catch (err) {
                logger.error(util.format("callback for " + fcn + " failed: error: %s", err));
            }
        }
        else {
            logger.debug('Added datatype to service successfully:', result);
            try {
                cb && cb(null, result);
            } catch (err) {
                logger.error(util.format("callback for " + fcn + " failed: error: %s", err));
            }
        }
    });
}

/**
 * Remove datatype from a service
 */
module.exports.removeDatatypeFromService = removeDatatypeFromService;
function removeDatatypeFromService(caller, datatypeID, serviceID, cb) {
    logger.debug('removeDatatypeFromService:', datatypeID);

    var args = [];

    args.push(serviceID);
    args.push(datatypeID);

    var fcn = "removeDatatypeFromService";
    chaincodeOps._invoke(caller, fcn, args, function (err, result) {
        if (err) {
            var errmsg = "Failed to remove datatype from service";
            logger.error(errmsg, err);
            try {
                cb && cb(new Error(errmsg));
            } catch (err) {
                logger.error(util.format("callback for " + fcn + " failed: error: %s", err));
            }
        }
        else {
            logger.debug('Removed datatype from service successfully:', result);
            try {
                cb && cb(null, result);
            } catch (err) {
                logger.error(util.format("callback for " + fcn + " failed: error: %s", err));
            }
        }
    });
}

/**
 * Get all services of org
 * @param orgId ID of organization.
 */
module.exports.getServicesOfOrg = getServicesOfOrg;
function getServicesOfOrg(caller, orgId, cb) {
    logger.debug('get getServicesOfOrg for:', orgId);

    let fcn = 'getServicesOfOrg';
    let args = [orgId];

    chaincodeOps._query(caller, fcn, args, function (err, value) {
        if (err) {
            var errmsg = "Failed to get services for org";
            logger.error(errmsg, err);
            try {
                cb && cb(new Error(errmsg));
            } catch (err) {
                logger.error(util.format("callback for " + fcn + " failed: error: %s", err));
            }
        }
        else {
            logger.debug('got services successfully:', value);
            try {
                cb && cb(null, value);
            } catch (err) {
                logger.error(util.format("callback for " + fcn + " failed: error: %s", err));
            }
        }
    });
}

/**
 * Enroll patient
 * @param enrollemnt The enrollment object
 * @param enrollmentKeyB64 The enrollment key in B64
 */
module.exports.enrollPatient = enrollPatient;
function enrollPatient(caller, enrollment, enrollmentKeyB64, cb) {
    logger.debug('Enroll patient:', enrollment);
    var enrollmentStr = "";
    try {
        enrollmentStr = JSON.stringify(enrollment);
    } catch (err) {
        var errmsg = "Invalid enrollment";
        logger.error(errmsg, err);
        cb && cb(new Error(errmsg));
        return;
    }

    var fcn = 'enrollPatient';
    var args = [enrollmentStr];
    args.push(enrollmentKeyB64);

    chaincodeOps._invoke(caller, fcn, args, function (err, result) {
        if (err) {
            var errmsg = "Failed to enroll patient:";
            logger.error(errmsg, err);
            try {
                cb && cb(new Error(errmsg));
            } catch (err) {
                logger.error(util.format("callback for " + fcn + " failed: error: %s", err));
            }
        }
        else {
            logger.debug('Enrolled patient successfully:', result);
            try {
                cb && cb(null, result);
            } catch (err) {
                logger.error(util.format("callback for " + fcn + " failed: error: %s", err));
            }
        }
    });
};

/**
 * Remove datatype from a service
 */
module.exports.unenrollPatient = unenrollPatient;
function unenrollPatient(caller, serviceID, userID, cb) {
    logger.debug('unenrollPatient:', userID);

    var args = [];

    args.push(serviceID);
    args.push(userID);

    var fcn = "unenrollPatient";
    chaincodeOps._invoke(caller, fcn, args, function (err, result) {
        if (err) {
            var errmsg = "Failed to unenroll patient";
            logger.error(errmsg, err);
            try {
                cb && cb(new Error(errmsg));
            } catch (err) {
                logger.error(util.format("callback for " + fcn + " failed: error: %s", err));
            }
        }
        else {
            logger.debug('Unenrolled patient successfully:', result);
            try {
                cb && cb(null, result);
            } catch (err) {
                logger.error(util.format("callback for " + fcn + " failed: error: %s", err));
            }
        }
    });
}

/**
 * Get patient enrollments
 */
module.exports.getPatientEnrollments = getPatientEnrollments;
function getPatientEnrollments(caller, userID, statusFilter, cb) {
    logger.debug('get patient enrollments: ', userID);
    logger.debug('get patient enrollments (statusFilter): ', statusFilter);

    var fcn = 'getPatientEnrollments';
    var args = [userID];
    if (statusFilter != "") {
        args.push(statusFilter)
    }

    chaincodeOps._query(caller, fcn, args, function (err, value) {
        if (err) {
            var errmsg = "Failed to get patient enrollments:";
            logger.error(errmsg, err);
            try {
                cb && cb(new Error(errmsg));
            } catch (err) {
                logger.error(util.format("callback for " + fcn + " failed: error: %s", err));
            }
        }
        else {
            logger.debug('successfully got patient enrollments:', value);
            try {
                cb && cb(null, value);
            } catch (err) {
                logger.error(util.format("callback for " + fcn + " failed: error: %s", err));
            }
        }
    });
};

/**
 * Get service enrollments
 */
module.exports.getServiceEnrollments = getServiceEnrollments;
function getServiceEnrollments(caller, serviceID, statusFilter, cb) {
    logger.debug('get service enrollments: ', serviceID);
    logger.debug('get service enrollments (statusFilter): ', statusFilter);

    var fcn = 'getServiceEnrollments';
    var args = [serviceID];
    if (statusFilter != "") {
        args.push(statusFilter)
    }

    chaincodeOps._query(caller, fcn, args, function (err, value) {
        if (err) {
            var errmsg = "Failed to get service enrollments:";
            logger.error(errmsg, err);
            try {
                cb && cb(new Error(errmsg));
            } catch (err) {
                logger.error(util.format("callback for " + fcn + " failed: error: %s", err));
            }
        }
        else {
            logger.debug('successfully got service enrollments:', value);
            try {
                cb && cb(null, value);
            } catch (err) {
                logger.error(util.format("callback for " + fcn + " failed: error: %s", err));
            }
        }
    });
};

/**
 * Put consent patient data
 * @param consent
 */
module.exports.putConsentPatientData = putConsentPatientData;
function putConsentPatientData(caller, consent, consentKey, cb) {
    logger.debug('put consent patient data:', consent);

    if (!consent || typeof consent != 'object') {
        cb && cb(new Error("Consent must be a JSON format"));
        return;
    }

    var consentStr = "";
    try {
        consentStr = JSON.stringify(consent);
    } catch (err) {
        var errmsg = "Invalid consent object";
        logger.error(errmsg, err);
        cb && cb(new Error(errmsg));
        return
    }

    var fcn = 'putConsentPatientData';
    var args = [consentStr, consentKey];

    chaincodeOps._invoke(caller, fcn, args, function (err, result) {
        if (err) {
            var errmsg = "Failed to put consent";
            logger.error(errmsg, err);
            cb && cb(new Error(errmsg));
        }
        else {
            logger.debug('Put consent successfully:', result);
            cb && cb(null, result);
        }
    });
};

/**
 * Put consent owner data
 * @param consent
 */
module.exports.putConsentOwnerData = putConsentOwnerData;
function putConsentOwnerData(caller, consent, consentKey, cb) {
    logger.debug('put consent owner data:', consent);

    if (!consent || typeof consent != 'object') {
        cb && cb(new Error("Consent must be a JSON format"));
        return;
    }

    var consentStr = "";
    try {
        consentStr = JSON.stringify(consent);
    } catch (err) {
        var errmsg = "Invalid consent object";
        logger.error(errmsg, err);
        cb && cb(new Error(errmsg));
        return
    }

    var fcn = 'putConsentOwnerData';
    var args = [consentStr, consentKey];

    chaincodeOps._invoke(caller, fcn, args, function (err, result) {
        if (err) {
            var errmsg = "Failed to put consent";
            logger.error(errmsg, err);
            cb && cb(new Error(errmsg));
        }
        else {
            logger.debug('Put consent successfully:', result);
            cb && cb(null, result);
        }
    });
};

/**
 * Get consent
 */
module.exports.getConsent = getConsent;
function getConsent(caller, ownerID, targetID, datatypeID, cb) {
    logger.debug('getConsent for:', ownerID, targetID, datatypeID);

    var fcn = 'getConsent';
    var args = [ownerID, targetID, datatypeID];

    chaincodeOps._query(caller, fcn, args, function (err, value) {
        if (err) {
            var errmsg = "Failed to get consent";
            logger.error(errmsg, err);
            try {
                cb && cb(new Error(errmsg));
            } catch (err) {
                logger.error(util.format("callback for " + fcn + " failed: error: %s", err));
            }
        }
        else {
            logger.debug('successfully got consent:', value);
            try {
                cb && cb(null, value);
            } catch (err) {
                logger.error(util.format("callback for " + fcn + " failed: error: %s", err));
            }
        }
    });
};

/**
 * Get consent for owner data
 */
module.exports.getConsentOwnerData = getConsentOwnerData;
function getConsentOwnerData(caller, ownerServiceId, serviceId, datatypeId, cb) {
    logger.debug('get consent for:', ownerServiceId, serviceId, datatypeId);

    var fcn = 'getConsentOwnerData';
    var args = [ownerServiceId, serviceId, datatypeId];

    chaincodeOps._query(caller, fcn, args, function (err, value) {
        if (err) {
            var errmsg = "Failed to get owner data consent";
            logger.error(errmsg, err);
            try {
                cb && cb(new Error(errmsg));
            } catch (err) {
                logger.error(util.format("callback for " + fcn + " failed: error: %s", err));
            }
        }
        else {
            logger.debug('successfully got owner data consent:', value);
            try {
                cb && cb(null, value);
            } catch (err) {
                logger.error(util.format("callback for " + fcn + " failed: error: %s", err));
            }
        }
    });
};

/**
 * Get all consents for service / user
 * @param serviceId  service ID
 * @param userId Login User Id
 */
module.exports.getConsents = getConsents;
function getConsents(caller, serviceId, userId, cb) {
    logger.debug('get consents for:', serviceId, userId);

    var fcn = 'getConsents';
    var args = [serviceId, userId];

    chaincodeOps._query(caller, fcn, args, function (err, value) {
        if (err) {
            var errmsg = "Failed to get consents";
            logger.error(errmsg, err);
            try {
                cb && cb(new Error(errmsg));
            } catch (err) {
                logger.error(util.format("callback for " + fcn + " failed: error: %s", err));
            }
        }
        else {
            logger.debug('successfully got consents:', value);
            try {
                cb && cb(null, value);
            } catch (err) {
                logger.error(util.format("callback for " + fcn + " failed: error: %s", err));
            }
        }
    });
};

/**
 * Get consents with owner ID
 */
module.exports.getConsentsWithOwnerID = getConsentsWithOwnerID;
function getConsentsWithOwnerID(caller, ownerID, cb) {
    logger.debug('getConsentsWithOwnerID for:', ownerID);

    var fcn = 'getConsentsWithOwnerID';
    var args = [ownerID];

    chaincodeOps._query(caller, fcn, args, function (err, value) {
        if (err) {
            var errmsg = "Failed to get consents";
            logger.error(errmsg, err);
            try {
                cb && cb(new Error(errmsg));
            } catch (err) {
                logger.error(util.format("callback for " + fcn + " failed: error: %s", err));
            }
        }
        else {
            logger.debug('successfully got consents:', value);
            try {
                cb && cb(null, value);
            } catch (err) {
                logger.error(util.format("callback for " + fcn + " failed: error: %s", err));
            }
        }
    });
};

/**
 * Get consents with target ID
 */
module.exports.getConsentsWithTargetID = getConsentsWithTargetID;
function getConsentsWithTargetID(caller, targetID, cb) {
    logger.debug('getConsentsWithTargetID for:', targetID);

    var fcn = 'getConsentsWithTargetID';
    var args = [targetID];

    chaincodeOps._query(caller, fcn, args, function (err, value) {
        if (err) {
            var errmsg = "Failed to get consents";
            logger.error(errmsg, err);
            try {
                cb && cb(new Error(errmsg));
            } catch (err) {
                logger.error(util.format("callback for " + fcn + " failed: error: %s", err));
            }
        }
        else {
            logger.debug('successfully got consents:', value);
            try {
                cb && cb(null, value);
            } catch (err) {
                logger.error(util.format("callback for " + fcn + " failed: error: %s", err));
            }
        }
    });
};

/**
 * Get consent requests
 */
module.exports.getConsentRequests = getConsentRequests;
function getConsentRequests(caller, userId, serviceId, cb) {
    logger.debug('get consent requests for:', userId, serviceId);

    var fcn = 'getConsentRequests';
    var args = [userId, serviceId];

    chaincodeOps._query(caller, fcn, args, function (err, value) {
        if (err) {
            var errmsg = "Failed to get consent reqeusts";
            logger.error(errmsg, err);
            try {
                cb && cb(new Error(errmsg));
            } catch (err) {
                logger.error(util.format("callback for " + fcn + " failed: error: %s", err));
            }
        }
        else {
            logger.debug('successfully got consent requests:', value);
            try {
                cb && cb(null, value);
            } catch (err) {
                logger.error(util.format("callback for " + fcn + " failed: error: %s", err));
            }
        }
    });
};

/**
 * uploadUserData
 * @param userdata User data to be uploaded
 */
module.exports.uploadUserData = uploadUserData;
function uploadUserData(caller, userdata, dataKeyB64, cb) {
    logger.debug('upload user data:', userdata);

    var userdataStr = "";
    try {
        userdataStr = JSON.stringify(userdata);
    } catch (err) {
        logger.error('Invalid data: ', err);
        cb && cb(new Error("Invalid data"));
        return;
    }

    var fcn = 'uploadUserData';
    var args = [userdataStr, dataKeyB64];

    chaincodeOps._invoke(caller, fcn, args, function (err, result) {
        if (err) {
            var errmsg = "Failed to upload user data";
            logger.error(errmsg, err);
            try {
                cb && cb(new Error(errmsg));
            } catch (err) {
                logger.error(util.format("callback for " + fcn + " failed: error: %s", err));
            }
        }
        else {
            logger.debug('Upload user data successfully:', result);
            try {
                cb && cb(null, result);
            } catch (err) {
                logger.error(util.format("callback for " + fcn + " failed: error: %s", err));
            }
        }
    });
};

/**
 * downloadUserData
 * @param caller for the user submitting the transaction.
 */
module.exports.downloadUserData = downloadUserData;
function downloadUserData(caller, serviceId, userId, datatypeId, start_timestamp, end_timestamp, latest_only, maxNum, timestamp, cb) {
    logger.debug('download user data: ', serviceId, datatypeId);

    var fcn = 'downloadUserData';
    var args = [serviceId, userId, datatypeId, latest_only, start_timestamp, end_timestamp, maxNum, timestamp];

    chaincodeOps._query(caller, fcn, args, function (err, value) {
        if (err) {
            var errmsg = "Failed to download user data";
            logger.error(errmsg, err);
            try {
                cb && cb(new Error(errmsg));
            } catch (err) {
                logger.error(util.format("callback for " + fcn + " failed: error: %s", err));
            }
        }
        else {
            var ownerDatas = value.owner_datas;
            var transactionLog = JSON.stringify(value.transaction_log);

            if (ownerDatas.length <= 0) {
                logger.debug('successfully got patient datas:', ownerDatas);
                try {
                    cb && cb(null, ownerDatas);
                } catch (err) {
                    logger.error(util.format("callback for " + fcn + " failed: error: %s", err));
                }
            } else {
                var fcn = 'addQueryTransactionLog';
                var args = [transactionLog];

                chaincodeOps._invoke(caller, fcn, args, function (err) {
                    if (err) {
                        var errmsg = "Failed to add transaction log for query";
                        logger.error(errmsg, err);
                        try {
                            cb && cb(new Error(errmsg));
                        } catch (err) {
                            logger.error(util.format("callback for " + fcn + " failed: error: %s", err));
                        }
                    }
                    else {
                        logger.debug('Added transaction log for query successfully:');
                        logger.debug('successfully got patient datas:', ownerDatas);
                        try {
                            cb && cb(null, ownerDatas);
                        } catch (err) {
                            logger.error(util.format("callback for " + fcn + " failed: error: %s", err));
                        }
                    }
                });
            }
        }
    });
}

/**
 * downloadUserDataConsentToken
 * @param caller for the user submitting the transaction.
 */
module.exports.downloadUserDataConsentToken = downloadUserDataConsentToken;
function downloadUserDataConsentToken(caller, start_timestamp, end_timestamp, latest_only, maxNum, timestamp, token, cb) {
    logger.debug('download user data with consent token');

    var fcn = 'downloadUserDataConsentToken';
    var args = [latest_only, start_timestamp, end_timestamp, maxNum, timestamp, token];

    chaincodeOps._query(caller, fcn, args, function (err, value) {
        if (err) {
            var errmsg = "Failed to download user data";
            logger.error(errmsg, err);
            try {
                cb && cb(new Error(errmsg));
            } catch (err) {
                logger.error(util.format("callback for " + fcn + " failed: error: %s", err));
            }
        }
        else {
            var ownerDatas = value.owner_datas;
            var transactionLog = JSON.stringify(value.transaction_log);

            if (ownerDatas.length <= 0) {
                logger.debug('successfully got patient datas:', ownerDatas);
                try {
                    cb && cb(null, ownerDatas);
                } catch (err) {
                    logger.error(util.format("callback for " + fcn + " failed: error: %s", err));
                }
            } else {
                var fcn = 'addQueryTransactionLog';
                var args = [transactionLog];

                chaincodeOps._invoke(caller, fcn, args, function (err) {
                    if (err) {
                        var errmsg = "Failed to add transaction log for query";
                        logger.error(errmsg, err);
                        try {
                            cb && cb(new Error(errmsg));
                        } catch (err) {
                            logger.error(util.format("callback for " + fcn + " failed: error: %s", err));
                        }
                    }
                    else {
                        logger.debug('Added transaction log for query successfully:');
                        logger.debug('successfully got patient datas:', ownerDatas);
                        try {
                            cb && cb(null, ownerDatas);
                        } catch (err) {
                            logger.error(util.format("callback for " + fcn + " failed: error: %s", err));
                        }
                    }
                });
            }
        }
    });
}

/**
 * deleteUserData
 */
module.exports.deleteUserData = deleteUserData;
function deleteUserData(caller, userId, datatype, service, timestamp, cb) {
    logger.debug('delete user data:', userId, datatype, service, timestamp);

    var fcn = 'deleteUserData';
    var args = [userId, datatype, service, timestamp];

    chaincodeOps._invoke(caller, fcn, args, function (err, result) {
        if (err) {
            var errmsg = "Failed to delete user data";
            logger.error(errmsg, err);
            try {
                cb && cb(new Error(errmsg));
            } catch (err) {
                logger.error(util.format("callback for " + fcn + " failed: error: %s", err));
            }
        }
        else {
            logger.debug('Deleted user data successfully:', result);
            try {
                cb && cb(null, result);
            } catch (err) {
                logger.error(util.format("callback for " + fcn + " failed: error: %s", err));
            }
        }
    });
};

/**
 * downloadOwnerDataConsentToken
 * @param caller for the user submitting the transaction.
 */
module.exports.downloadOwnerDataConsentToken = downloadOwnerDataConsentToken;
function downloadOwnerDataConsentToken(caller, start_timestamp, end_timestamp, latest_only, maxNum, timestamp, token, cb) {
    logger.debug('download owner data with consent token');

    var fcn = 'downloadOwnerDataConsentToken';
    var args = [latest_only, start_timestamp, end_timestamp, maxNum, timestamp, token];

    chaincodeOps._query(caller, fcn, args, function (err, value) {
        if (err) {
            var errmsg = "Failed to download owner data";
            logger.error(errmsg, err);
            try {
                cb && cb(new Error(errmsg));
            } catch (err) {
                logger.error(util.format("callback for " + fcn + " failed: error: %s", err));
            }
        }
        else {
            var ownerDatas = value.owner_datas;
            var transactionLog = JSON.stringify(value.transaction_log);

            if (ownerDatas.length <= 0) {
                logger.debug('successfully got patient datas:', ownerDatas);
                try {
                    cb && cb(null, ownerDatas);
                } catch (err) {
                    logger.error(util.format("callback for " + fcn + " failed: error: %s", err));
                }
            } else {
                var fcn = 'addQueryTransactionLog';
                var args = [transactionLog];

                chaincodeOps._invoke(caller, fcn, args, function (err) {
                    if (err) {
                        var errmsg = "Failed to add transaction log for query";
                        logger.error(errmsg, err);
                        try {
                            cb && cb(new Error(errmsg));
                        } catch (err) {
                            logger.error(util.format("callback for " + fcn + " failed: error: %s", err));
                        }
                    }
                    else {
                        logger.debug('Added transaction log for query successfully:');
                        logger.debug('successfully got patient datas:', ownerDatas);
                        try {
                            cb && cb(null, ownerDatas);
                        } catch (err) {
                            logger.error(util.format("callback for " + fcn + " failed: error: %s", err));
                        }
                    }
                });
            }
        }
    });
}

/**
 * uploadOwnerData
 */
module.exports.uploadOwnerData = uploadOwnerData;
function uploadOwnerData(caller, ownerData, dataKeyB64, cb) {
    logger.debug('upload owner data:', ownerData);

    var ownerDataStr = "";
    try {
        ownerDataStr = JSON.stringify(ownerData);
    } catch (err) {
        var errmsg = "Invalid data";
        logger.error(errmsg, err);
        cb && cb(new Error(errmsg));
        return;
    }

    var fcn = 'uploadOwnerData';
    var args = [ownerDataStr, dataKeyB64];

    chaincodeOps._invoke(caller, fcn, args, function (err, result) {
        if (err) {
            var errmsg = "Failed to upload owner data";
            logger.error(errmsg, err);
            try {
                cb && cb(new Error(errmsg));
            } catch (err) {
                logger.error(util.format("callback for " + fcn + " failed: error: %s", err));
            }
        }
        else {
            logger.debug('Uploaded owner data successfully:', result);
            try {
                cb && cb(null, result);
            } catch (err) {
                logger.error(util.format("callback for " + fcn + " failed: error: %s", err));
            }
        }
    });
}

/**
 * downloadOwnerDataAsOwner
 */
module.exports.downloadOwnerDataAsOwner = downloadOwnerDataAsOwner;
function downloadOwnerDataAsOwner(caller, serviceId, datatypeId, startTimestamp, endTimestamp, latestOnly, maxNum, timestamp, cb) {
    logger.debug('get owner data:', serviceId, datatypeId);
    logger.debug('start time:', startTimestamp);
    logger.debug('end time:', endTimestamp);

    var fcn = 'downloadOwnerDataAsOwner';
    var args = [serviceId, datatypeId, latestOnly, startTimestamp, endTimestamp, maxNum, timestamp];

    chaincodeOps._query(caller, fcn, args, function (err, value) {
        if (err) {
            var errmsg = "Failed to get owner data";
            logger.error(errmsg, err);
            try {
                cb && cb(new Error(errmsg));
            } catch (err) {
                logger.error(util.format("callback for " + fcn + " failed: error: %s", err));
            }
        }
        else {
            logger.debug('successfully got owner datas:', value.owner_datas);
            var ownerDatas = value.owner_datas;
            var transactionLog = JSON.stringify(value.transaction_log);

            if (ownerDatas.length <= 0) {
                logger.debug('successfully got ownerdatas:', ownerDatas);
                try {
                    cb && cb(null, ownerDatas);
                } catch (err) {
                    logger.error(util.format("callback for " + fcn + " failed: error: %s", err));
                }
            } else {
                var fcn = 'addQueryTransactionLog';
                var args = [transactionLog];

                chaincodeOps._invoke(caller, fcn, args, function (err) {
                    if (err) {
                        var errmsg = "Failed to add transaction log for query";
                        logger.error(errmsg, err);
                        try {
                            cb && cb(new Error(errmsg));
                        } catch (err) {
                            logger.error(util.format("callback for " + fcn + " failed: error: %s", err));
                        }
                    }
                    else {
                        logger.debug('Added transaction log for query successfully:');
                        logger.debug('successfully got ownerdatas:', ownerDatas);
                        try {
                            cb && cb(null, ownerDatas);
                        } catch (err) {
                            logger.error(util.format("callback for " + fcn + " failed: error: %s", err));
                        }
                    }
                });
            }
        }
    });
}

/**
 * downloadOwnerDataAsRequester
 */
module.exports.downloadOwnerDataAsRequester = downloadOwnerDataAsRequester;
function downloadOwnerDataAsRequester(caller, contractId, datatypeId, startTimestamp, endTimestamp, latestOnly, maxNum, timestamp, cb) {
    logger.debug('download owner data:', contractId, datatypeId);

    var fcn = 'downloadOwnerDataAsRequester';
    var args = [contractId, datatypeId, latestOnly, startTimestamp, endTimestamp, maxNum, timestamp];

    chaincodeOps._query(caller, fcn, args, function (err, value) {
        if (err) {
            var errmsg = "Failed to get owner data";
            logger.error(errmsg, err);
            try {
                cb && cb(new Error(errmsg));
            } catch (err) {
                logger.error(util.format("callback for " + fcn + " failed: error: %s", err));
            }
        }
        else {
            var encContract = value.encrypted_contract;
            var ownerDatas = value.owner_datas;
            var datatype = value.datatype;

            if (ownerDatas.length <= 0) {
                logger.debug('successfully get ownerdatas:', ownerDatas);
                try {
                    cb && cb(null, ownerDatas);
                } catch (err) {
                    logger.error(util.format("callback for " + fcn + " failed: error: %s", err));
                }
            } else {
                var fcn = 'addContractDetailDownload';
                var args = [contractId, encContract, datatype];

                chaincodeOps._invoke(caller, fcn, args, function (err, result) {
                    if (err) {
                        var errmsg = "Failed to add download contract detail";
                        logger.error(errmsg, err);
                        try {
                            cb && cb(new Error(errmsg));
                        } catch (err) {
                            logger.error(util.format("callback for " + fcn + " failed: error: %s", err));
                        }
                    }
                    else {
                        logger.debug('Added download contract detail successfully:', result);
                        logger.debug('successfully get ownerdatas:', ownerDatas);
                        try {
                            cb && cb(null, ownerDatas);
                        } catch (err) {
                            logger.error(util.format("callback for " + fcn + " failed: error: %s", err));
                        }
                    }
                });
            }
        }
    });
}

/**
 * downloadOwnerDataWithConsent
 * @param caller for the user submitting the transaction.
 */
module.exports.downloadOwnerDataWithConsent = downloadOwnerDataWithConsent;
function downloadOwnerDataWithConsent(caller, targetServiceId, ownerServiceId, datatypeId, start_timestamp, end_timestamp, latest_only, maxNum, timestamp, token, cb) {
    logger.debug('download owner data with consent (target, owner): ', targetServiceId, ownerServiceId);

    var fcn = 'downloadOwnerDataWithConsent';
    var args = [targetServiceId, ownerServiceId, datatypeId, latest_only, start_timestamp, end_timestamp, maxNum, timestamp];
    if (token) {
        args.push(token);
    }
    chaincodeOps._query(caller, fcn, args, function (err, value) {
        if (err) {
            var errmsg = "Failed to download user data";
            logger.error(errmsg, err);
            try {
                cb && cb(new Error(errmsg));
            } catch (err) {
                logger.error(util.format("callback for " + fcn + " failed: error: %s", err));
            }
        }
        else {
            var ownerDatas = value.owner_datas;
            var transactionLog = JSON.stringify(value.transaction_log);

            if (ownerDatas.length <= 0) {
                logger.debug('successfully got ownerdatas:', ownerDatas);
                try {
                    cb && cb(null, ownerDatas);
                } catch (err) {
                    logger.error(util.format("callback for " + fcn + " failed: error: %s", err));
                }
            } else {
                var fcn = 'addQueryTransactionLog';
                var args = [transactionLog];

                chaincodeOps._invoke(caller, fcn, args, function (err) {
                    if (err) {
                        var errmsg = "Failed to add transaction log for query";
                        logger.error(errmsg, err);
                        try {
                            cb && cb(new Error(errmsg));
                        } catch (err) {
                            logger.error(util.format("callback for " + fcn + " failed: error: %s", err));
                        }
                    }
                    else {
                        logger.debug('Added transaction log for query successfully:');
                        logger.debug('successfully got ownerdatas:', ownerDatas);
                        try {
                            cb && cb(null, ownerDatas);
                        } catch (err) {
                            logger.error(util.format("callback for " + fcn + " failed: error: %s", err));
                        }
                    }
                });
            }
        }
    });
}

/**
 * downloadOwnerDataConsentToken
 * @param caller for the user submitting the transaction.
 */
module.exports.downloadOwnerDataConsentToken = downloadOwnerDataConsentToken;
function downloadOwnerDataConsentToken(caller, start_timestamp, end_timestamp, latest_only, maxNum, timestamp, token, cb) {
    logger.debug('download owner data with consent token');

    var fcn = 'downloadOwnerDataConsentToken';
    var args = [latest_only, start_timestamp, end_timestamp, maxNum, timestamp, token];

    chaincodeOps._query(caller, fcn, args, function (err, value) {
        if (err) {
            var errmsg = "Failed to download owner data";
            logger.error(errmsg, err);
            try {
                cb && cb(new Error(errmsg));
            } catch (err) {
                logger.error(util.format("callback for " + fcn + " failed: error: %s", err));
            }
        }
        else {
            var ownerDatas = value.owner_datas;
            var transactionLog = JSON.stringify(value.transaction_log);

            if (ownerDatas.length <= 0) {
                logger.debug('successfully got ownerdatas:', ownerDatas);
                try {
                    cb && cb(null, ownerDatas);
                } catch (err) {
                    logger.error(util.format("callback for " + fcn + " failed: error: %s", err));
                }
            } else {
                var fcn = 'addQueryTransactionLog';
                var args = [transactionLog];

                chaincodeOps._invoke(caller, fcn, args, function (err) {
                    if (err) {
                        var errmsg = "Failed to add transaction log for query";
                        logger.error(errmsg, err);
                        try {
                            cb && cb(new Error(errmsg));
                        } catch (err) {
                            logger.error(util.format("callback for " + fcn + " failed: error: %s", err));
                        }
                    }
                    else {
                        logger.debug('Added transaction log for query successfully:');
                        logger.debug('successfully got ownerdatas:', ownerDatas);
                        try {
                            cb && cb(null, ownerDatas);
                        } catch (err) {
                            logger.error(util.format("callback for " + fcn + " failed: error: %s", err));
                        }
                    }
                });
            }
        }
    });
}

/**
 * Validate consent
 */
module.exports.validateConsent = validateConsent;
function validateConsent(caller, ownerID, targetID, datatypeID, access, timestamp, cb) {
    logger.debug('validate consent for:', ownerID, targetID, datatypeID, access);

    var fcn = 'validateConsent';
    var args = [ownerID, targetID, datatypeID, access, timestamp];

    chaincodeOps._query(caller, fcn, args, function (err, value) {
        if (err) {
            var errmsg = "Failed to validate consent";
            logger.error(errmsg, err);
            try {
                cb && cb(new Error(errmsg));
            } catch (err) {
                logger.error(util.format("callback for " + fcn + " failed: error: %s", err));
            }
        }
        else {
            var validation = value.validation;
            var transactionLog = JSON.stringify(value.transaction_log);

            var fcn = 'addValidateConsentQueryLog';
            var args = [JSON.stringify(validation), transactionLog];

            chaincodeOps._invoke(caller, fcn, args, function (err) {
                if (err) {
                    var errmsg = "Failed to add transaction log for query";
                    logger.error(errmsg, err);
                    try {
                        cb && cb(new Error(errmsg));
                    } catch (err) {
                        logger.error(util.format("callback for " + fcn + " failed: error: %s", err));
                    }
                }
                else {
                    logger.debug('Added transaction log for query successfully:');
                    logger.debug('successfully got consent validation:', validation);
                    try {
                        cb && cb(null, validation);
                    } catch (err) {
                        logger.error(util.format("callback for " + fcn + " failed: error: %s", err));
                    }
                }
            });
        }
    });
};

/**
 * Request contract
 */
module.exports.createContract = createContract;
function createContract(caller, contract, contractKeyB64, cb) {
    logger.debug('request contract:', contract.contract_id);

    var contractStr = "";
    try {
        contractStr = JSON.stringify(contract);
    } catch (err) {
        var errmsg = "Failed to convert contract to string";
        logger.error(errmsg, err);
        cb && cb(new Error(errmsg));
        return;
    }

    let fcn = 'createContract';
    let args = [contractStr, contractKeyB64];

    chaincodeOps._invoke(caller, fcn, args, function (err, result) {
        if (err) {
            var errmsg = "Failed to create contract";
            logger.error(errmsg, err);
            try {
                cb && cb(new Error(errmsg));
            } catch (err) {
                logger.error(util.format("callback for " + fcn + " failed: error: %s", err));
            }
        }
        else {
            logger.debug('Contract created successfully:', result);
            try {
                cb && cb(null, result);
            } catch (err) {
                logger.error(util.format("callback for " + fcn + " failed: error: %s", err));
            }
        }
    });
}

/**
 * Add contract details
 * @param contractId The unique contract identifier
 * @param detailType The type of update to be made to the contract
 * @param contractTerms The json form terms to be added to the contract
 */
module.exports.addContractDetail = addContractDetail;
function addContractDetail(caller, contractId, detailType, contractTerms, timestamp, cb) {
    logger.debug('add contract detail:', contractId, detailType);
    if (!contractTerms) contractTerms = {};
    if (typeof contractTerms != 'object') {
        cb && cb(new Error("contract_terms must be a JSON format"));
        return;
    }
    else {
        contractTerms = JSON.stringify(contractTerms);
    }

    let fcn = 'addContractDetail';
    let args = [contractId, detailType, contractTerms, timestamp];

    chaincodeOps._invoke(caller, fcn, args, function (err, result) {
        if (err) {
            var errmsg = "Failed to add contract details";
            logger.error(errmsg, err);
            try {
                cb && cb(new Error(errmsg));
            } catch (err) {
                logger.error(util.format("callback for " + fcn + " failed: error: %s", err));
            }
        }
        else {
            logger.debug('Added contract details successfully:', result);
            try {
                cb && cb(null, result);
            } catch (err) {
                logger.error(util.format("callback for " + fcn + " failed: error: %s", err));
            }
        }
    });
}

/**
 * Give permission to download based on contract
 * @param contract_id the id of the contract to give permission
 * @param max_num_download the maximum number of downloads
 */
module.exports.givePermissionByContract = givePermissionByContract;
function givePermissionByContract(caller, contract_id, max_num_download, timestamp, datatype, cb) {
    logger.debug('give download permission for contract:', contract_id, "with max:", max_num_download);

    let fcn = 'givePermissionByContract';
    let args = [contract_id, String(max_num_download), timestamp, datatype];

    chaincodeOps._invoke(caller, fcn, args, function (err, result) {
        if (err) {
            var errmsg = "Failed to give permission to contract";
            logger.error(errmsg, err);
            try {
                cb && cb(new Error(errmsg));
            } catch (err) {
                logger.error(util.format("callback for " + fcn + " failed: error: %s", err));
            }
        }
        else {
            logger.debug('Gave permission to contract successfully:', result);
            try {
                cb && cb(null, result);
            } catch (err) {
                logger.error(util.format("callback for " + fcn + " failed: error: %s", err));
            }
        }
    });
}

/**
 * Get contract
 */
module.exports.getContract = getContract;
function getContract(caller, contractId, cb) {
    logger.debug('get contract:', contractId);

    let fcn = 'getContract';
    let args = [contractId];

    chaincodeOps._query(caller, fcn, args, function (err, value) {
        if (err) {
            var errmsg = "Failed to get contract";
            logger.error(errmsg, err);
            try {
                cb && cb(new Error(errmsg));
            } catch (err) {
                logger.error(util.format("callback for " + fcn + " failed: error: %s", err));
            }
        }
        else {
            logger.debug('successfully got contract: ', value);
            try {
                cb && cb(null, value);
            } catch (err) {
                logger.error(util.format("callback for " + fcn + " failed: error: %s", err));
            }
        }
    });
};

/**
 * Get contracts
 */
module.exports.getOwnerContracts = getOwnerContracts;
function getOwnerContracts(caller, ownerServiceId, status, cb) {
    logger.debug('get owner contracts:', ownerServiceId, status);
    if (!status) status = "";

    let fcn = 'getOwnerContracts';
    let args = [ownerServiceId, status];

    chaincodeOps._query(caller, fcn, args, function (err, value) {
        if (err) {
            var errmsg = "Failed to get owner contracts";
            logger.error(errmsg, err);
            try {
                cb && cb(new Error(errmsg));
            } catch (err) {
                logger.error(util.format("callback for " + fcn + " failed: error: %s", err));
            }
        }
        else {
            logger.debug('successfully got owner contracts: ', value);
            try {
                cb && cb(null, value);
            } catch (err) {
                logger.error(util.format("callback for " + fcn + " failed: error: %s", err));
            }
        }
    });
};

/**
 * Get requester contracts
 */
module.exports.getRequesterContracts = getRequesterContracts;
function getRequesterContracts(caller, requesterServiceId, status, cb) {
    logger.debug('get requester contracts:', requesterServiceId, status);
    if (!status) status = "";

    let fcn = 'getRequesterContracts';
    let args = [requesterServiceId, status];

    chaincodeOps._query(caller, fcn, args, function (err, value) {
        if (err) {
            var errmsg = "Failed to get requester contracts";
            logger.error(errmsg, err);
            try {
                cb && cb(new Error(errmsg));
            } catch (err) {
                logger.error(util.format("callback for " + fcn + " failed: error: %s", err));
            }
        }
        else {
            logger.debug('successfully got requester contracts: ', value);
            try {
                cb && cb(null, value);
            } catch (err) {
                logger.error(util.format("callback for " + fcn + " failed: error: %s", err));
            }
        }
    });
};

/**
 * Get logs
 */
module.exports.getLogs = getLogs;
function getLogs(caller, contractId, patientId, serviceId, datatypeId, contractOrgId, consentOwnerTargetId, start_timestamp, end_timestamp, latest_only, maxNum, cb) {
    logger.debug('get logs for:', contractId, patientId, serviceId, datatypeId, contractOrgId, consentOwnerTargetId, start_timestamp, end_timestamp, latest_only, maxNum);

    var fcn = 'getLogs';
    var args = [contractId, patientId, serviceId, datatypeId, contractOrgId, consentOwnerTargetId, start_timestamp, end_timestamp, latest_only, maxNum];

    chaincodeOps._query(caller, fcn, args, function (err, value) {
        if (err) {
            var errmsg = "Failed to get logs";
            logger.error(errmsg, err);
            try {
                cb && cb(new Error(errmsg));
            } catch (err) {
                logger.error(util.format("callback for " + fcn + " failed: error: %s", err));
            }
        }
        else {
            logger.debug('successfully got logs:', value);
            try {
                cb && cb(null, value);
            } catch (err) {
                logger.error(util.format("callback for " + fcn + " failed: error: %s", err));
            }
        }
    });
};
