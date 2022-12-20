/*******************************************************************************
 *
 *
 * (c) Copyright Merative US L.P. and others 2020-2022 
 *
 * SPDX-Licence-Identifier: Apache 2.0
 *
 *******************************************************************************/

let userManager = require('common-utils/user_manager');
let kms = require("common-utils/kms");
let ums = require("common-utils/ums");
let solutionChaincodeOps = require("./solution_chaincode_ops.js");
const hfc = require('fabric-client');
const { getPii, deIdentifyPii, getDeIdentifiedPii } = require('common-utils/deIdentifierUtils/deIdentifierService');
const { logPhiAccess, logFailedPhiAccess } = require('../phiAccessLogging/phiAccessLogger');

let TAG = "solution_api_handler.js";
let log4js = require('log4js');
let logger = log4js.getLogger(TAG);
logger.level = 'DEBUG';

const deIdentifierConfig = hfc.getConfigSetting('de_identifier');
const isDeIdentifierServiceEnabled = deIdentifierConfig.enabled;

// Set to true to allow relevant functions to broadcast on success; enables refresh notifications
var allowBroadcast = false;

module.exports.getChaincode = getChaincode;
function getChaincode() {
    return chaincode;
}

module.exports.clearMessage = clearMessage;
function clearMessage(req, res) {
    var msg = req.query.msg;

    if (!msg || msg == "") {
        msg = "all";
    }

    if (msg == "error_msg") {
        req.session.error_msg = null;
    }
    else if (msg == "success_msg") {
        req.session.success_msg = null;
    }
    else if (msg == "login_error_msg") {
        req.session.login_error_msg = null;
    }
    else if (msg == "reg_error_msg") {
        req.session.reg_error_msg = null;
    }
    else if (msg == "all") {
        req.session.error_msg = null;
        req.session.success_msg = null;
        req.session.login_error_msg = null;
        req.session.reg_error_msg = null;
    } else if (msg == "all_err") {
        req.session.error_msg = null;
        req.session.login_error_msg = null;
        req.session.reg_error_msg = null;
    }

    res.json({ message: "clearMessage Done: " + msg });
}

/*
 * if redirectUrl == true then next() else redirect to redirectUrl
 *    if not defined, default is /index
 * if logoutUrl is not defined, default is ums.getLoginUrl()
 */
module.exports.login = login;
function login(req, res, timeoutsec) {
    logger.info('site_router.js login() - fired');
    req.session.error_msg = 'Invalid username or password';
    req.session.reg_error_msg = null;
    req.session.success_msg = null;

    // Registering the user against a peer can serve as a login checker, for now
    logger.info("Attempting login for:", req.body.username + " token timeoutsec=" + timeoutsec);
    try {
        let callerData = {
            org: req.body.ca_org,
            channel: req.body.channel,
        }

        let caller = {
            id: req.body.username,
            secret: req.body.password,
            org: req.body.ca_org,
            channel: req.body.channel
        };

        userManager.getLoginToken(caller.id, caller.secret, callerData, function (err, token, user) {
            if (err) {
                req.session.login_error_msg = "User login failed:\n" + err.message;
                logger.warn("User login failed:", err);
                req.session.loginID = req.body.username;
                res.render('index', {
                    title: 'Login',
                    bag: {
                        e: process.error,
                        session: req.session
                    }
                });
            } else {
                logger.info("User login successful:", caller.id);

                getUser(caller, caller.id).then(([err, user]) => {
                    if (err != null) {
                        res.render('index', {
                            title: 'Login',
                            bag: {
                                //setup: setup,
                                e: process.error,
                                session: req.session
                            }
                        });
                    }
                    else {
                        req.session.callerID = caller.id
                        req.session.callerSecret = caller.secret
                        req.session.caOrg = caller.org
                        req.session.channel = caller.channel
                        req.session.token = token;
                        req.session.username = user.id;
                        req.session.name = user.name;
                        if (user.role == "org" && user.solution_info.is_org_admin == true) {
                            req.session.role = "org";
                        } else if (user.role == "org" && user.solution_info.is_org_admin == false) {
                            req.session.role = "service";
                        } else {
                            req.session.role = user.role;
                        }
                        if (user.solution_info.services) {
                            req.session.services = user.solution_info.services;
                        } else {
                            req.session.services = [];
                        }
                        req.session.org = user.org;
                        req.session.status = user.status;
                        req.session.error_msg = null;
                        req.session.success_msg = null;

                        logger.info("Setting session:");
                        logger.debug(req.session);

                        if (typeof redirectUrl === 'string' || redirectUrl instanceof String) {
                            logger.info("====> redirecting to " + redirectUrl);
                            res.redirect(redirectUrl);
                        } else if (typeof redirectUrl === "function") {
                            logger.info("do next()");
                            redirectUrl();
                        } else {
                            res.redirect("/index");
                        }
                    }
                }).catch(err => {
                    logger.error('User login failed', err);
                    req.session.login_error_msg = "User login failed with error:" + err;
                    res.redirect(logoutUrl);
                });
            }
        }, timeoutsec);
    } catch (err) {
        logger.debug("error: ", err);
        req.session.login_error_msg = "User login failed with error:" + err;
        res.redirect(logoutUrl);
    }
}

function onSuccess(data) {
    if (allowBroadcast) {
        if (wss) {
            logger.info("wss active");
            if (data.type.indexOf('get') == -1) {
                var broadcast_msg = { msg: "broadcast_refresh", data: data };
                wss.broadcast(broadcast_msg);
            }
        }
    }
}

async function getOrgsFromChaincode(caller) {
    return new Promise((resolve) => {
        solutionChaincodeOps.getOrgs(caller, function (err, orgs) {
            resolve([err, orgs]);
        });
    });
}

async function identifyOrganizationPii(organisation) {
    const [id, name] = await Promise.all([getPii(organisation.id), getPii(organisation.name)]);
    organisation.id = id;
    organisation.name = name;

    return organisation;
}

async function identifyOrganizationsPii(organisations) {
    if (!Array.isArray(organisations)) throw new Error('Cannot identify Organization records. Wrong datatype. Expected Array type.');

    const identifiedOrgPromises = organisations.map(identifyOrganizationPii);

    const successPropmises = identifiedOrgPromises.map((identifiedOrgPromise) => {
        return identifiedOrgPromise.catch((err) => {
            const errorMessage = 'Failed to get organization';
            logger.error(errorMessage, err);
        })
    });

    return Promise.all(successPropmises);
}

async function getOrgs(caller) {
    let [err, organizations] = await getOrgsFromChaincode(caller);

    if (err) return [err, organizations];

    if (isDeIdentifierServiceEnabled && organizations) {
        organizations = await (await identifyOrganizationsPii(organizations)).filter((org) => org != null);
    }

    return [null, organizations];
}

function getOrgsApiHandler(caller, data, req, res) {
    logger.debug('getOrgs');

    getOrgs(caller).then(([err, orgs]) => {
        if (err != null) {
            logger.error(err);
            res.json([]);
        }
        else {
            res.json(orgs);
        }
    }).catch(err => {
        const message = 'Failed to get organizations';
        logger.error(message, err);
        res.status(500).json({ msg: message, status: 500 });
    });
}

async function getOrgFromChaincodeOps(caller, orgId) {
    logger.debug('Call solutionChaincodeOps.getOrg');
    return new Promise((resolve => {
        solutionChaincodeOps.getOrg(caller, orgId, function (err, user) {
            logger.debug('solutionChaincodeOps.getOrg callback');
            resolve([err, user]);
        });
    }));
}

async function getOrg(caller, orgId) {
    logger.debug('getOrg');

    let deIdentifiedId = orgId;
    if (isDeIdentifierServiceEnabled) {
        deIdentifiedId = await getDeIdentifiedPii(orgId);
    }

    if (!deIdentifiedId) return [new Error('Org not found'), null];

    const [err, organization] = await getOrgFromChaincodeOps(caller, deIdentifiedId);

    if (err) return [err, organization];

    if (isDeIdentifierServiceEnabled && organization) {
        organization.id = orgId;
        organization.name = await getPii(organization.name)
    }
    return [null, organization];
}

function getOrgApiHandler(caller, data, req, res) {
    logger.debug('getOrgApiHandler');

    getOrg(caller, data.id).then(([err, org]) => {
        if (err != null) {
            logger.error(err);
            res.json({});
        }
        else {
            res.json(org);
        }
    }).catch((err) => {
        logger.error(err);
        res.status(500).json('Failed to get Organization');
    });
}

function updateOrgApiHandler(caller, data, req, res) {
    logger.debug('update org');

    getOrg(caller, data.id).then(([err, user]) => {
        if (err != null || !user) {
            var errmsg = "Org not found";
            logger.error(errmsg, err);
            res.status(404).json({ msg: errmsg, status: 404 });
        } else if (!user.secret) {
            var errmsg = "Unauthorized to update the org";
            logger.error(errmsg, err);
            res.status(401).json({ msg: errmsg, status: 401 });
        } else if (data.secret && user.secret && user.secret != data.secret) {
            var errmsg = "Org admin's secret cannot be changed";
            logger.error(errmsg, err);
            res.status(400).json({ msg: errmsg, status: 400 });
        } else if (!data.email) {
            var errmsg = "email missing";
            logger.error(errmsg);
            res.status(400).json({ msg: "Invalid data:" + errmsg, status: 400 });
        } else {
            logger.debug("Existing Org found: ", data.id);
            data.secret = user.secret;
            return registerOrg(caller, data, req, res);
        }
    }).catch(err => {
        const errorMessage = 'Failed to update org';
        logger.error(errorMessage, err);
        res.status(500).json({ msg: errorMessage, status: 500 });
    });
}

function registerOrgApiHandler(caller, data, req, res) {
    logger.debug('register org');

    getOrg(caller, data.id).then(([err, user]) => {
        if (err == null && user) {
            var errmsg = "Existing org with same id found";
            logger.error(errmsg, err);
            res.status(200).send("Org was already registered");
        } else {
            logger.debug("Existing org not found: ", data.id);
            return registerOrg(caller, data, req, res);
        }
    }).catch(err => {
        const errorMessage = 'Failed to register org';
        logger.error(errorMessage, err);
        res.status(500).json({ msg: errorMessage, status: 500 });
    });
}

async function registerOrg(caller, data, req, res) {
    logger.debug('registerOrg');
    logger.debug(data);

    var failIfExist = caller.id ? false : true;
    try {
        let { id, name } = data;

        if (isDeIdentifierServiceEnabled) {
            [id, name] = await Promise.all([deIdentifyPii(id), deIdentifyPii(name)]);
        }

        // 1. register user (CA) & enroll
        userManager.registerUser(id, data.secret, data.role, data.ca_org, data.verify_key, data.private_key, data.public_key, data.sym_key, failIfExist).then((attrList) => {
            // 2. register user in chaincode
            var is_group = true
            var solution_private_data = data.data;

            var userInfo = {
                id,
                name,
                role: data.role,
                is_group,
                status: data.status,
                email: data.email,
                secret: data.secret,
                solution_private_data,
            };

            //keys
            for (let i = 0; i < attrList.length; i++) {
                let attr = attrList[i];
                if (attr["name"] == "prvkey") {
                    userInfo["private_key"] = attr["value"];
                } else if (attr["name"] == "pubkey") {
                    userInfo["public_key"] = attr["value"];
                } else if (attr["name"] == "symkey") {
                    userInfo["sym_key"] = attr["value"];
                }
            }

            solutionChaincodeOps.registerOrg(caller, userInfo, function (err, result) {
                if (err != null) {
                    var errmsg = "Org is registered to CA, but failed to register org in Blockchain:" + err.message;
                    logger.error(errmsg);
                    res.status(500).json({ msg: errmsg, status: 500 });
                } else {
                    logger.debug('org (CC) registered successfully:', result);
                    logger.info('org registration completed successfully');
                    res.json({
                        id: data.id,
                        secret: data.secret,
                        msg: "org registration completed successfully",
                        tx_id: result.tx_id
                    });
                }
            });

        }).catch((err) => {
            var errmsg = "Failed to register org (CA):" + err.message;
            logger.error(errmsg, err);
            res.status(500).json({ msg: errmsg, status: 500 });
        });

    } catch (err) {
        var errmsg = "Failed to register org";
        logger.error(errmsg, err);
        res.status(500).json({ msg: errmsg, status: 500 });
    }

}

async function getUserFromChaincode(caller, userId) {
    logger.debug('Call getUserFromChaincode');
    return new Promise(resolve => {
        solutionChaincodeOps.getUser(caller, userId, function (err, user) {
            logger.debug('Call solutionChaincodeOps.getUser callback');
            resolve([err, user]);
        });
    });
}

async function getUser(caller, userId) {
    logger.debug('getUser');
    let deIdentifiedId = userId;

    if (isDeIdentifierServiceEnabled) {
        deIdentifiedId = await getDeIdentifiedPii(userId);
    }
    if (!deIdentifiedId) return [new Error('User not found'), null];

    const [err, user] = await getUserFromChaincode(caller, deIdentifiedId);

    if (err) return [err, user];

    if (isDeIdentifierServiceEnabled && user && user.id) {
        user.id = userId;
        user.name = await getPii(user.name);

        if (user.org) {
            user.org = await getPii(user.org);
        }
    }

    return [null, user];
}

function getUserApiHandler(caller, data, req, res) {
    logger.debug('getUserApiHandler');
    getUser(caller, data.userid).then(([err, user]) => {
        if (err != null) {
            logger.error(err);
            res.json({});
        }
        else {
            if (user.id != "") {
                res.json(user);
            } else {
                res.json({});
            }
        }
    }).catch(err => {
        const errorMessage = 'Failed to find user';
        logger.error(errorMessage, err);
        res.status(500).json({ msg: errorMessage, status: 500 });
    });
}

async function registerUser(caller, data, req, res) {
    logger.debug('registerUser');

    //system can register system, auditor, org, patient
    //org can register org user, patient
    //service can register patient
    //patient can self register

    var failIfExist = caller.id ? false : true;
    try {
        let { id, name, org } = data;

        if (isDeIdentifierServiceEnabled && org) {
            [id, name, org] = await Promise.all([deIdentifyPii(id), deIdentifyPii(name), deIdentifyPii(org)]);
        } else if (isDeIdentifierServiceEnabled) {
            [id, name] = await Promise.all([deIdentifyPii(id), deIdentifyPii(name)]);
        }

        try {
            // 1. register user (CA) & enroll
            const attrList = await userManager.registerUser(id, data.secret, data.role, data.ca_org, data.verify_key, data.private_key, data.public_key, data.sym_key, failIfExist);
            // 2. register user in chaincode
            var is_group = false
            var solution_private_data = data.data;

            var userInfo = {
                id,
                name,
                role: data.role,
                is_group,
                org,
                status: data.status,
                email: data.email,
                secret: data.secret,
                solution_private_data,
            };

            //keys
            for (let i = 0; i < attrList.length; i++) {
                let attr = attrList[i];
                if (attr["name"] == "prvkey") {
                    userInfo["private_key"] = attr["value"];
                } else if (attr["name"] == "pubkey") {
                    userInfo["public_key"] = attr["value"];
                } else if (attr["name"] == "symkey") {
                    userInfo["sym_key"] = attr["value"];
                }
            }


            solutionChaincodeOps.registerUser(caller, userInfo, function (err, result) {
                if (err != null) {
                    var errmsg = "User is registered to CA, but failed to register user in Blockchain:" + err.message;
                    logger.error(errmsg);
                    res.status(500).json({ msg: errmsg, status: 500 });
                } else {
                    logger.debug('user (CC) registered successfully:', result);
                    logger.info('user registration completed successfully');
                    res.json({
                        id: data.id,
                        secret: data.secret,
                        msg: 'user registration completed successfully',
                        tx_id: result.tx_id
                    });
                }
            });


        } catch(err)  {
            var errmsg = "Failed to register user in CA:" + err.message;
            logger.error(errmsg, err);
            res.status(500).json({ msg: errmsg, status: 500 });
        };

    } catch (err) {
        var errmsg = "Failed to register user";
        logger.error(errmsg, err);
        res.status(500).json({ msg: errmsg, status: 500 });
    }
}

function registerUserApiHandler(caller, data, req, res) {
    logger.debug('register user');

    getUser(caller, data.id).then(([err, user]) => {
        if (err == null && user.id != "") {
            var errmsg = "Existing user with same id found";
            logger.error(errmsg, err);
            res.status(200).send("User was already registered");
        } else {
            logger.debug("Existing User not found: ", data.id);
            return registerUser(caller, data, req, res);
        }
    }).catch(err => {
        const message = 'Failed to register user';
        logger.error(message, err);
        res.status(500).json({ msg: message, status: 500 });
    });
}

async function enrollAndConsent(caller, data, req, res) {

    kms.getSymKeyAes(function (err, enrollment_sym_key) {
        if (err) {
            var errmsg = "error creating enrollment sym key";
            logger.error(errmsg, err);
            res.status(500).json({ msg: errmsg, successes: [], failures: [], status: 500, failure_type: enrollError });
        } else {
            const enrollData = {user_id: data.id, service_id: data.service_id, status: "active"};

            enrollPatient(caller, enrollData, enrollment_sym_key.keyBase64).then(result => {
                logger.info('Enroll patient success');

                caller.id = data.id;
                caller.secret = data.secret;
                caller.org = data.ca_org;

                const consentData = {consents: data.consents};
                consentData.consents.forEach(consent => {
                    consent.owner_id = data.id;
                    consent.service_id = data.service_id;
                });

                putMultiConsentPatientData(caller, consentData, req, res)
                .catch((err) => {
                    const errorMessage = 'putConsentPatientData error';
                    logger.error(errorMessage, err);
                    res.status(500).json({ msg: errorMessage, successes: [], failures: [], status: 500, failure_type: consentError });
                });;

            }).catch((err) => {
                const errorMessage = 'Enroll patient error';
                logger.error(errorMessage, err);
                res.status(500).json({ msg: errorMessage, successes: [], failures: [], status: 500, failure_type: enrollError });
            });
        }
    }, data.enrollment_sym_key);
}

async function registerEnrollAndConsent(caller, data, req, res) {
    logger.debug('registerEnrollAndConsent');

    var failIfExist = caller.id ? false : true;
    try {
        let { id, name, org } = data;

        if (isDeIdentifierServiceEnabled && org) {
            [id, name, org] = await Promise.all([deIdentifyPii(id), deIdentifyPii(name), deIdentifyPii(org)]);
        } else if (isDeIdentifierServiceEnabled) {
            [id, name] = await Promise.all([deIdentifyPii(id), deIdentifyPii(name)]);
        }

        // 1. register user (CA) & enroll
        userManager.registerUser(id, data.secret, data.role, data.ca_org, "", "", "", "", failIfExist).then(attrList => {
            // 2. register user in chaincode
            var is_group = false
            var solution_private_data = data.data;

            var userInfo = {
                id,
                name,
                role: data.role,
                is_group,
                org,
                status: data.status,
                email: data.email,
                secret: data.secret,
                solution_private_data,
            };

            //keys
            for (let i = 0; i < attrList.length; i++) {
                let attr = attrList[i];
                if (attr["name"] == "prvkey") {
                    userInfo["private_key"] = attr["value"];
                } else if (attr["name"] == "pubkey") {
                    userInfo["public_key"] = attr["value"];
                } else if (attr["name"] == "symkey") {
                    userInfo["sym_key"] = attr["value"];
                }
            }


            solutionChaincodeOps.registerUser(caller, userInfo, function (err, result) {
                if (err != null) {
                    var errmsg = "User is registered to CA, but failed to register user in Blockchain:" + err.message;
                    logger.error(errmsg);
                    res.status(500).json({ msg: errmsg, successes: [], failures: [], status: 500, failure_type: registerError });
                } else {
                    logger.debug('user (CC) registered successfully:', result);
                    logger.info('user registration completed successfully');
                    return enrollAndConsent(caller, data, req, res);
                }
            });


        }).catch((err) => {
            var errmsg = "Failed to register user in CA:" + err.message;
            logger.error(errmsg, err);
            res.status(500).json({ msg: errmsg, successes: [], failures: [], status: 500, failure_type: registerError });
        });

    } catch (err) {
        var errmsg = "Failed to register user";
        logger.error(errmsg, err);
        res.status(500).json({ msg: errmsg, successes: [], failures: [], status: 500, failure_type: registerError });
    }
}

function registerEnrollAndConsentApiHandler(caller, data, req, res) {
    logger.debug('register,enroll user and add consent');

    if (caller.id === data.id) {
        var errmsg = "error enrolling patient, caller and enrolled patient must be different";
        logger.error(errmsg);
        res.status(500).json({ msg: "Invalid data:" + errmsg, successes: [], failures: [], status: 500, failure_type: enrollError });
    }

    getUser(caller, data.id).then(([err, user]) => {
        if (err == null && user.id != "") {
            logger.debug("User was already registered, enroll user and add consent");
            return enrollAndConsent(caller, data, req, res);
        } else {
            logger.debug("Existing User not found: ", data.id);
            return registerEnrollAndConsent(caller, data, req, res);
        }
    }).catch(err => {
        const message = 'Failed to register user';
        logger.error(message, err);
        res.status(500).json({ msg: message, successes: [], failures: [], status: 500, failure_type: registerError });
    });
}

function putUserInOrgApiHandler(caller, data, req, res) {
    logger.debug("put user in group");

    chaincodeOps.putUserInOrg(
        caller,
        data.user_id,
        data.org_id,
        data.isAdmin,
        function (err, result) {
            if (err != null) {
                var errmsg = "Failed to put user in group";
                logger.error(errmsg, err);
                res.status(500).json({ msg: errmsg, status: 500 });
            } else {
                logger.info("put user in group successfully:", result);
                res.json({
                    msg: "put user in group completed successfully",
                    tx_id: result.tx_id
                });
            }
        }
    );
}

function removeUserFromOrgApiHandler(caller, data, req, res) {
    logger.debug("removeUserFromOrg");

    chaincodeOps.removeUserFromOrg(caller, data.user_id, data.org_id, function (
        err,
        result
    ) {
        if (err != null) {
            var errmsg = "Failed to remove user from group";
            logger.error(errmsg, err);
            res.status(500).json({ msg: errmsg, status: 500 });
        } else {
            logger.info("removed user from group successfully:", result);
            res.json({
                msg: "removed user from group completed successfully",
                tx_id: result.tx_id
            });
        }
    });
}

async function addPermissionOrgAdminInChaincode(caller, userId, orgId) {
    return new Promise((resolve, reject) => {
        solutionChaincodeOps.addPermissionOrgAdmin(caller, userId, orgId, function (err, result) {
            if (err) return reject(err);

            return resolve(result);
        });
    });
}

async function addPermissionOrgAdmin(caller, data) {
    let userId = data.user_id;
    let orgId = data.org_id;

    if (isDeIdentifierServiceEnabled) {
        [userId, orgId] = await Promise.all([getDeIdentifiedPii(data.user_id), getDeIdentifiedPii(data.org_id)]);
    }

    return addPermissionOrgAdminInChaincode(caller, userId, orgId);
}

function addPermissionOrgAdminApiHandler(caller, data, req, res) {
    logger.debug('addPermissionOrgAdmin');
    addPermissionOrgAdmin(caller, data).then(result => {
        const message = 'adding org admin permission completed successfully';
        res.status(200).json({ msg: message, tx_id: result.tx_id });
    }).catch(err => {
        const errorMessage = 'addPermissionOrgAdmin error';
        logger.error(errorMessage, err);
        res.status(500).json({ msg: errorMessage, status: 500 });
    });
}

async function deletePermissionOrgAdminInChaincode(caller, userId, orgId) {
    return new Promise((resolve, reject) => {
        solutionChaincodeOps.deletePermissionOrgAdmin(caller, userId, orgId, function (err, result) {
            if (err) return reject(err);

            return resolve(result);
        });
    });
}

async function deletePermissionOrg(caller, data) {
    let userId = data.user_id;
    let orgId = data.org_id;

    if (isDeIdentifierServiceEnabled) {
        [userId, orgId] = await Promise.all([getDeIdentifiedPii(data.user_id), getDeIdentifiedPii(data.org_id)]);
    }

    return deletePermissionOrgAdminInChaincode(caller, userId, orgId);
}

function deletePermissionOrgAdminApiHandler(caller, data, req, res) {
    logger.debug('deletePermissionOrgAdmin');

    deletePermissionOrg(caller, data).then(result => {
        const message = 'removing org admin permission completed successfully';
        res.status(200).json({ msg: message, tx_id: result.tx_id });
    }).catch(err => {
        const errorMessage = 'deletePermissionOrgAdmin error';
        logger.error(errorMessage, err);
        res.status(500).json({ msg: errorMessage, status: 500 });
    });
}

async function addPermissionServiceAdminInChaincode(caller, userId, serviceId) {
    return new Promise((resolve, reject) => {
        solutionChaincodeOps.addPermissionServiceAdmin(caller, userId, serviceId, function (err, result) {
            if (err) return reject(err);

            return resolve(result);
        });
    });
}

async function addPermissionServiceAdmin(caller, data) {
    let userId = data.user_id;
    let serviceId = data.service_id;

    if (isDeIdentifierServiceEnabled) {
        [userId, serviceId] = await Promise.all([getDeIdentifiedPii(data.user_id), getDeIdentifiedPii(data.service_id)]);
    }

    return addPermissionServiceAdminInChaincode(caller, userId, serviceId);
}

function addPermissionServiceAdminApiHandler(caller, data, req, res) {
    logger.debug('addPermissionServiceAdmin');
    addPermissionServiceAdmin(caller, data).then(result => {
        const message = 'adding service admin permission completed successfully';
        res.status(200).json({ msg: message, tx_id: result.tx_id });
    }).catch(err => {
        const errorMessage = 'addPermissionServiceAdmin error';
        logger.error(errorMessage, err);
        res.status(500).json({ msg: errorMessage, status: 500 });
    });
}

async function deletePermissionServiceFromChaincode(caller, userId, serviceId,) {
    return new Promise((resolve, reject) => {
        solutionChaincodeOps.deletePermissionServiceAdmin(caller, userId, serviceId, function (err, result) {
            if (err) return reject(err);

            return resolve(result);
        });
    });
}

async function deletePermissionService(caller, data) {
    let userId = data.user_id;
    let serviceId = data.service_id;

    if (isDeIdentifierServiceEnabled) {
        [userId, serviceId] = await Promise.all([getDeIdentifiedPii(data.user_id), getDeIdentifiedPii(data.service_id)]);
    }

    return deletePermissionServiceFromChaincode(caller, userId, serviceId);
}

function deletePermissionServiceAdminApiHandler(caller, data, req, res) {
    logger.debug('deletePermissionServiceAdmin');

    deletePermissionService(caller, data).then(result => {
        const message = 'removing service admin permission completed successfully';
        res.status(200).json({ msg: message, tx_id: result.tx_id });
    }).catch(err => {
        const errorMessage = 'deletePermissionServiceAdmin error';
        logger.error(errorMessage, err);
        res.status(500).json({ msg: errorMessage, status: 500 });
    });
}

async function addPermissionAuditorInChaincode(caller, userId, serviceId, auditPermissionKeyB64) {
    return new Promise((resolve, reject) => {
        solutionChaincodeOps.addPermissionAuditor(caller, userId, serviceId, auditPermissionKeyB64, function (err, result) {
            if (err) return reject(err);

            return resolve(result);
        });
    });
}

async function addPermissionAuditor(caller, data, auditPermissionKeyB64) {
    let userId = data.user_id;
    let serviceId = data.service_id;

    if (isDeIdentifierServiceEnabled) {
        [userId, serviceId] = await Promise.all(
            [getDeIdentifiedPii(data.user_id), getDeIdentifiedPii(data.service_id)]);
    }

    return addPermissionAuditorInChaincode(caller, userId, serviceId, auditPermissionKeyB64);
}

function addPermissionAuditorApiHandler(caller, data, req, res) {
    logger.debug('addPermissionAuditor');

    kms.getSymKeyAes(function (err, auditPermissionSymKey) {
        if (err) {
            let errMsg = "Error creating auditPermissionSymKey in KMS";
            logger.error(errMsg, err);
            res.status(500).json({ msg: errMsg, status: 500 });
        } else {
            addPermissionAuditor(caller, data, auditPermissionSymKey.keyBase64).then(result => {
                const message = 'adding auditor permission completed successfully';
                res.status(200).json({ msg: message, tx_id: result.tx_id });
            }).catch((err) => {
                const errorMessage = 'addPermissionAuditor error';
                logger.error(errorMessage, err);
                res.status(500).json({ msg: errorMessage, status: 500 });
            });
        }
    }, null)
}

async function deletePermissionAuditorInChaincode(caller, userId, serviceId) {
    return new Promise((resolve, reject) => {
        solutionChaincodeOps.deletePermissionAuditor(caller, userId, serviceId, function (err, result) {
            if (err) return reject(err);

            return resolve(result);
        });
    });
}

async function deletePermissionAuditor(caller, data) {
    let userId = data.user_id;
    let serviceId = data.service_id;

    if (isDeIdentifierServiceEnabled) {
        [userId, serviceId] = await Promise.all(
            [getDeIdentifiedPii(data.user_id), getDeIdentifiedPii(data.service_id)]);
    }

    return deletePermissionAuditorInChaincode(caller, userId, serviceId);
}

function deletePermissionAuditorApiHandler(caller, data, req, res) {
    logger.debug('deletePermissionAuditor');

    deletePermissionAuditor(caller, data).then(result => {
        const message = 'removing auditor permission completed successfully';
        res.status(200).json({ msg: message, tx_id: result.tx_id });
    }).catch(err => {
        const errorMessage = 'deletePermissionAuditor error';
        logger.error(errorMessage, err);
        res.status(500).json({ msg: errorMessage, status: 500 });
    });
}

function updateUserApiHandler(caller, data, req, res) {
    logger.debug('update user');
    getUser(caller, data.id).then(([err, user]) => {
        if (err != null || !user) {
            var errmsg = "User not found";
            logger.error(errmsg, err);
            res.status(404).json({ msg: errmsg, status: 404 });
        } else if (!user.secret) {
            var errmsg = "Unauthorized to update the user";
            logger.error(errmsg, err);
            res.status(401).json({ msg: errmsg, status: 401 });
        } else if (user.role != data.role) {
            var errmsg = "User's role cannot be changed";
            logger.error(errmsg, err);
            res.status(400).json({ msg: errmsg, status: 400 });
        } else if (data.secret && user.secret && user.secret != data.secret) {
            var errmsg = "User's secret cannot be changed";
            logger.error(errmsg, err);
            res.status(400).json({ msg: errmsg, status: 400 });
        } else {
            logger.debug("Existing User found: ", data.id)
            if (user.secret) {
                data.secret = user.secret;
            }
            return registerUser(caller, data, req, res);
        }
    }).catch(err => {
        const errorMessage = 'Failed to update user';
        logger.error(errorMessage, err);
        res.status(500).json({ msg: errorMessage, status: 500 });
    });
}

async function getUsersInChaincode(caller, orgId, maxNum, role) {
    return new Promise(resolve => {
        solutionChaincodeOps.getUsers(caller, orgId, maxNum, role, function (err, users) {
            resolve([err, users]);
        });
    })
}

async function getUsers(caller, data) {
    let orgId = data.org;

    if (isDeIdentifierServiceEnabled && orgId !== "*") {
        orgId = await getDeIdentifiedPii(data.org);
    }

    let [err, users] = await getUsersInChaincode(caller, orgId, "" + data.maxNum, data.role);

    if (err) return [err, users];

    if (isDeIdentifierServiceEnabled && Array.isArray(users) && users.length) {
        const usersWithPii = users.map(async (user) => {
            const [userId, name, org] = await Promise.all(
                [getPii(user.id), getPii(user.name), user.org && getPii(user.org)]);
            user.id = userId;
            user.name = name;
            user.org = org;

            return user;
        });

        const successPropmises = usersWithPii.map((userWithPii) => {
            return userWithPii.catch((err) => {
                const errorMessage = 'Failed to get user';
                logger.error(errorMessage, err);
            })
        });

        users = Promise.all(successPropmises);
    }
    return [err, users];
}

function getUsersApiHandler(caller, data, req, res) {
    logger.debug('getUsers');
    getUsers(caller, data).then(([err, users]) => {
        if (err != null) {
            logger.error(err);
            res.json([]);
        }
        else {
            res.json(users.filter((user) => user != null));
        }
    }).catch(err => {
        const message = 'Failed to get users';
        logger.log(message, err);
        res.status(500).json({ msg: message, status: 500 });
    });
}

function registerDatatypeApiHandler(caller, data, req, res) {
    logger.debug('register datatype');

    solutionChaincodeOps.getDatatype(caller, data.id, function (err, datatype) {
        // we have to check for empty datatype ID here because chaincode
        // returns empty datatype if does not exist
        if (err != null || (err == null && datatype.datatype_id != "")) {
            var errmsg = "Existing datatype with same id found";
            logger.error(errmsg, err);
            res.status(400).json({ msg: errmsg, status: 400 });
        }
        else {
            logger.debug("Existing datatype not found, proceed to register: ", data.id);

            var datatype = {
                datatype_id: data.id,
                description: data.description
            }

            solutionChaincodeOps.registerDatatype(caller, datatype, function (err, result) {
                if (err != null) {
                    var errmsg = "registerDatatype error";
                    logger.error(errmsg, err);
                    res.status(500).json({ msg: errmsg, status: 500 });
                }
                res.json({
                    id: data.id,
                    msg: "register datatype completed successfully",
                    tx_id: result.tx_id
                });
            });
        }
    });
}

function updateDatatypeApiHandler(caller, data, req, res) {
    logger.debug('update datatype');

    solutionChaincodeOps.getDatatype(caller, data.id, function (err, datatype) {
        if (err != null || (err == null && datatype.datatype_id == "")) {
            var errmsg = "Existing datatype with this id not found";
            logger.error(errmsg, err);
            res.status(400).json({ msg: errmsg, status: 400 });
        }
        else {
            var datatype = {
                datatype_id: data.id,
                description: data.description
            }

            solutionChaincodeOps.updateDatatype(caller, datatype, function (err, result) {
                if (err != null) {
                    var errmsg = "updateDatatype error";
                    logger.error(errmsg, err);
                    res.status(500).json({ msg: errmsg, status: 500 });
                }
                else {
                    res.json({
                        id: data.id,
                        msg: "update datatype completed successfully",
                        tx_id: result.tx_id
                    });
                }
            });
        }
    });
}

function getDatatypeApiHandler(caller, data, req, res) {
    logger.debug('getDatatype');
    solutionChaincodeOps.getDatatype(caller, data.id, function (err, datatype) {
        if (err != null) {
            logger.error(err);
            res.json({});
        }
        else {
            res.json(datatype);
        }
    });
}

function getAllDatatypesApiHandler(caller, data, req, res) {
    logger.debug('getAllDatatypes');
    solutionChaincodeOps.getAllDatatypes(caller, function (err, datatypes) {
        if (err != null) {
            logger.error(err);
            res.json([]);
        }
        else {
            res.json(datatypes);
        }
    });
}

function registerServiceApiHandler(caller, data, req, res) {
    logger.debug('register service');

    getOrg(caller, data.id).then(([err, service]) => {
        if (err == null && service.service_id != "") {
            var errmsg = "Existing service with same id found";
            logger.error(errmsg, err);
            res.status(400).json({ msg: errmsg, status: 400 });
        }
        else {
            logger.debug("Existing service not found: ", data.id);
            return registerService(caller, data, req, res);
        }
    }).catch(err => {
        const errorMessage = 'Failed to register service';
        logger.error(errorMessage, err);
        res.status(500).json({ msg: errorMessage, status: 500 });
    });
}

async function registerService(caller, data, req, res) {
    logger.debug('registerService');
    var failIfExist = caller.id ? false : true;
    try {
        let { id, name, org_id } = data;

        if (isDeIdentifierServiceEnabled) {
            [id, name, org_id] = await Promise.all([deIdentifyPii(id), deIdentifyPii(name), deIdentifyPii(org_id)]);
        }

        // 1. register user (CA) & enroll
        userManager.registerUser(id, data.secret, data.role, data.ca_org, data.verify_key, data.private_key, data.public_key, data.sym_key, failIfExist).then(attrList => {
            // 2. register user in chaincode
            var is_group = data.is_group == "true" || data.is_group == true

            var serviceInfo = {
                service_id: id,
                service_name: name,
                role: data.role,
                is_group,
                status: data.status,
                email: data.email,
                secret: data.secret,
                solution_private_data: data.solution_private_data,
                org_id,
                datatypes: data.datatypes,
                summary: data.summary,
                terms: data.terms,
                payment_required: data.payment_required,
                create_date: Math.floor(new Date().getTime() / 1000)
            };

            //keys
            for (let i = 0; i < attrList.length; i++) {
                let attr = attrList[i];
                if (attr["name"] == "prvkey") {
                    serviceInfo["private_key"] = attr["value"];
                } else if (attr["name"] == "pubkey") {
                    serviceInfo["public_key"] = attr["value"];
                } else if (attr["name"] == "symkey") {
                    serviceInfo["sym_key"] = attr["value"];
                }
            }

            solutionChaincodeOps.registerService(caller, serviceInfo, function (err, result) {
                if (err != null) {
                    var errmsg = "Service is registered to CA, but failed to register service in Blockchain:" + err.message;
                    logger.error(errmsg);
                    res.status(500).json({ msg: errmsg, status: 500 });
                } else {
                    logger.debug('service (CC) registered successfully:', result);
                    logger.info('service registration completed successfully');
                    res.json({
                        id: data.id,
                        secret: data.secret,
                        msg: "service registration completed successfully",
                        tx_id: result.tx_id
                    });
                }
            });


        }).catch((err) => {
            var errmsg = "Failed to register service in CA:" + err.message;
            logger.error(errmsg, err);
            res.status(500).json({ msg: errmsg, status: 500 });
        });

    } catch (err) {
        var errmsg = "Failed to register service";
        logger.error(errmsg, err);
        res.status(500).json({ msg: errmsg, status: 500 });
    }

}

function updateServiceApiHandler(caller, data, req, res) {
    logger.debug('update service');

    getOrg(caller, data.id).then(([err, service]) => {
        if (err != null || (err == null && service.service_id == "")) {
            var errmsg = "Service not found";
            logger.error(errmsg, err);
            res.status(404).json({ msg: errmsg, status: 404 });
        } else if (!service.secret) {
            var errmsg = "Unauthorized to update the service";
            logger.error(errmsg, err);
            res.status(401).json({ msg: errmsg, status: 401 });
        } else if (data.secret && service.secret && service.secret != data.secret) {
            var errmsg = "Service admin's secret cannot be changed";
            logger.error(errmsg, err);
            res.status(400).json({ msg: errmsg, status: 400 });
        } else {
            logger.debug("Existing service found: ", data.id);
            data.secret = service.secret;
            return updateService(caller, data, req, res);
        }
    }).catch(err => {
        const message = 'Failed to update service';
        logger.error(message, err);
        res.status(500).json({ msg: message, status: 500 });
    });
}

async function updateService(caller, data, req, res) {
    logger.debug('updateService');

    var failIfExist = caller.id ? false : true;

    try {
        let { id, name, org_id } = data;

        if (isDeIdentifierServiceEnabled) {
            [id, name, org_id] = await Promise.all([deIdentifyPii(id), deIdentifyPii(name), deIdentifyPii(org_id)]);
        }

        // 1. register user (CA) & enroll
        userManager.registerUser(id, data.secret, data.role, data.ca_org, data.verify_key, data.private_key, data.public_key, data.sym_key, failIfExist).then(attrList => {
            // 2. register user in chaincode
            var is_group = data.is_group == "true" || data.is_group == true

            var serviceInfo = {
                service_id: id,
                service_name: name,
                role: data.role,
                is_group,
                status: data.status,
                email: data.email,
                secret: data.secret,
                solution_private_data: data.solution_private_data,
                org_id,
                datatypes: data.datatypes,
                summary: data.summary,
                terms: data.terms,
                payment_required: data.payment_required,
                update_date: Math.floor(new Date().getTime() / 1000)
            };

            //keys
            for (let i = 0; i < attrList.length; i++) {
                let attr = attrList[i];
                if (attr["name"] == "prvkey") {
                    serviceInfo["private_key"] = attr["value"];
                } else if (attr["name"] == "pubkey") {
                    serviceInfo["public_key"] = attr["value"];
                } else if (attr["name"] == "symkey") {
                    serviceInfo["sym_key"] = attr["value"];
                }
            }

            solutionChaincodeOps.updateService(caller, serviceInfo, function (err, result) {
                if (err != null) {
                    var errmsg = "Service is registered to CA, but failed to update service (CC):" + err.message;
                    logger.error(errmsg);
                    res.status(500).json({ msg: errmsg, status: 500 });
                } else {
                    logger.debug('service (CC) updated successfully:', result);
                    logger.info('service update completed successfully');
                    res.json({
                        id: data.id,
                        secret: data.secret,
                        msg: "service update completed successfully",
                        tx_id: result.tx_id
                    });
                }
            });


        }).catch((err) => {
            var errmsg = "Failed to update service (CA):" + err.message;
            logger.error(errmsg, err);
            res.status(500).json({ msg: errmsg, status: 500 });
        });

    } catch (err) {
        var errmsg = "Failed to update service";
        logger.error(errmsg, err);
        res.status(500).json({ msg: errmsg, status: 500 });
    }

}

async function addDatatypeToServiceInChaincode(caller, datatype, serviceId) {
    return new Promise((resolve, reject) => {
        solutionChaincodeOps.addDatatypeToService(caller, datatype, serviceId, function (err, result) {
            if (err) return reject(err);

            return resolve(result);
        });
    });
}

async function addDatatypeToService(caller, data, datatype) {
    let serviceId = data.service_id;

    if (isDeIdentifierServiceEnabled) {
        serviceId = await getDeIdentifiedPii(data.service_id);
    }

    return addDatatypeToServiceInChaincode(caller, datatype, serviceId);
}

function addDatatypeToServiceApiHandler(caller, data, req, res) {
    logger.debug('addDatatypeToService');
    getService(caller, data.service_id).then(([err, service]) => {
        if (err != null || (err == null && service.service_id == "")) {
            var errmsg = "Invalid id: Service with id " + data.service_id + " does not exist";
            logger.error(errmsg, err);
            res.status(400).json({ msg: errmsg, status: 400 });
        } else {
            solutionChaincodeOps.getDatatype(caller, data.datatype_id, function (err, datatype) {
                if (err != null || (err == null && datatype.datatype_id == "")) {
                    var errmsg = "Datatype not found";
                    logger.error(errmsg, err);
                    res.status(404).json({ msg: errmsg, status: 404 });
                } else {
                    var datatype = {
                        datatype_id: data.datatype_id,
                        access: data.access
                    }

                    addDatatypeToService(caller, data, datatype).then(result => {
                        const message = 'adding datatype to service completed successfully';
                        logger.debug(message);
                        res.status(200).json({ msg: message, tx_id: result.tx_id });
                    }).catch(err => {
                        const message = 'addDatatypeToService error';
                        logger.error(message, err);
                        res.status(500).json({ msg: message, status: 500 });
                    });
                }
            })
        }
    }).catch(err => {
        const errorMessage = 'addDatatypeToService error';
        logger.error(errorMessage, err);
        res.status(500).json({ msg: errorMessage, status: 500 });
    });
}

async function removeDataTypeFromServiceInChaincode(caller, datatypeId, serviceId) {
    return new Promise((resolve, reject) => {
        solutionChaincodeOps.removeDatatypeFromService(caller, datatypeId, serviceId, function (err, result) {
            if (err) return reject(err);

            return resolve(result);
        });
    });
}

async function removeDatatypeFromService(caller, data) {
    let { service_id } = data;

    if (isDeIdentifierServiceEnabled) {
        service_id = await getDeIdentifiedPii(service_id);
    }

    return removeDataTypeFromServiceInChaincode(caller, data.datatype_id, service_id);
}

function removeDatatypeFromServiceApiHandler(caller, data, req, res) {
    logger.debug('removeDatatypeFromServiceApiHandler');
    getService(caller, data.service_id).then(([err, service]) => {
        if (err != null || (err == null && service.service_id == "")) {
            var errmsg = "Invalid id: Service with id " + data.service_id + " does not exist";
            logger.error(errmsg, err);
            res.status(400).json({ msg: errmsg, status: 400 });
        } else {
            solutionChaincodeOps.getDatatype(caller, data.datatype_id, function (err, datatype) {
                if (err != null || (err == null && datatype.datatype_id == "")) {
                    var errmsg = "Datatype not found";
                    logger.error(errmsg, err);
                    res.status(404).json({ msg: errmsg, status: 404 });
                } else {
                    removeDatatypeFromService(caller, data).then(result => {
                        const message = 'removing datatype from service completed successfully';
                        logger.debug(message);
                        res.status(200).json({ msg: message, tx_id: result.tx_id });
                    }).catch(err => {
                        const message = 'removeDatatypeFromService error';
                        logger.error(message, err);
                        res.status(500).json({ msg: message, status: 500 });
                    });
                }
            })
        }
    }).catch(err => {
        const errorMessage = 'removeDatatypeFromService error';
        logger.error(errorMessage, err);
        res.status(500).json({ msg: errorMessage, status: 500 });
    });
}

async function getServiceFromChaincode(caller, serviceId) {
    return new Promise(resolve => {
        solutionChaincodeOps.getService(caller, serviceId, function (err, service) {
            resolve([err, service]);
        });
    })
}

async function getService(caller, serviceId) {
    let deIdentifiedId = serviceId;

    if (isDeIdentifierServiceEnabled) {
        deIdentifiedId = await getDeIdentifiedPii(serviceId);
    }
    if (!deIdentifiedId) return [new Error('Failed to find service'), null];

    const [err, service] = await getServiceFromChaincode(caller, deIdentifiedId);
    if (isDeIdentifierServiceEnabled && service) {
        await identifyServicePii(service);
    }
    return [err, service];
}

function getServiceApiHandler(caller, data, req, res) {
    logger.debug('getService');
    getService(caller, data.service_id).then(([err, service]) => {
        if (err != null) {
            logger.error(err);
            res.json({});
        }
        else {
            res.json(service);
        }
    }).catch(err => {
        const message = 'getService error';
        logger.error(message, err);
        res.status(500).json({ msg: message, status: 500 });
    });
}

async function getServicesOfOrgInChaincode(caller, orgId) {
    return new Promise(resolve => {
        solutionChaincodeOps.getServicesOfOrg(caller, orgId, function (err, services) {
            resolve([err, services]);
        });
    });
}

async function identifyServicePii(service) {
    service.service_id = await getPii(service.service_id);

    if (service.service_name) {
        service.service_name = await getPii(service.service_name);
    }
    if (service.org_id) {
        service.org_id = await getPii(service.org_id);
    }
    if (service.datatypes) {
        const identifiedDatatypePromises = service.datatypes.map(async (datatype) => {
            datatype.service_id = await getPii(datatype.service_id);
            return datatype;
        });

        const successPropmises = identifiedDatatypePromises.map((identifiedDatatypePromise) => {
            return identifiedDatatypePromise.catch((err) => {
                const errorMessage = 'Failed to get datatype';
                logger.error(errorMessage, err);
            })
        });

        service.datatypes = await Promise.all(successPropmises);
    }

    return service;
}

async function identifyServicesPii(deIdentifiedServices) {
    const identifiedServicePromises = deIdentifiedServices.map(identifyServicePii);

    const successPropmises = identifiedServicePromises.map((identifiedServicePromise) => {
        return identifiedServicePromise.catch((err) => {
            const errorMessage = 'Failed to get service';
            logger.error(errorMessage, err);
        })
    });

    return Promise.all(successPropmises);
}

async function getServicesOfOrg(caller, orgId) {
    let deIdentifiedOrgId = orgId;

    if (isDeIdentifierServiceEnabled) {
        deIdentifiedOrgId = await getDeIdentifiedPii(orgId);
    }

    let [err, services] = await getServicesOfOrgInChaincode(caller, deIdentifiedOrgId);

    if (err) return [err, services];

    if (isDeIdentifierServiceEnabled && Array.isArray(services) && services.length) {
        services = await (await identifyServicesPii(services)).filter((service) => service != null);
    }

    return [err, services];
}

function getServicesOfOrgApiHandler(caller, data, req, res) {
    logger.debug('getServicesOfOrg');
    getServicesOfOrg(caller, data.org).then(([err, services]) => {
        if (err != null) {
            logger.error(err);
            res.json([]);
        }
        else {
            res.json(services);
        }
    }).catch(err => {
        const errorMessage = 'Failed to get services of organization';
        logger.error(errorMessage, err);
        res.status(500).json({ msg: errorMessage, status: 500 });
    });
}

async function enrollPatientInChaincode(caller, enrollment, enrollmentKeyB64) {
    return new Promise((resolve, reject) => {
        solutionChaincodeOps.enrollPatient(caller, enrollment, enrollmentKeyB64, function (err, result) {
            if (err) return reject(err);

            return resolve(result);
        });
    });
}

async function enrollPatient(caller, data, enrollmentKeyB64) {
    let serviceId = data.service_id;
    let userId = data.user_id;

    if (isDeIdentifierServiceEnabled) {
        [serviceId, userId] = await Promise.all(
            [getDeIdentifiedPii(data.service_id), getDeIdentifiedPii(data.user_id)]);
    }

    const enrollment = {
        service_id: serviceId,
        user_id: userId,
        enroll_date: Math.floor(new Date().getTime() / 1000),
        status: data.status
    };

    return enrollPatientInChaincode(caller, enrollment, enrollmentKeyB64);
}

function enrollPatientApiHandler(caller, data, req, res) {
    logger.debug('enrollPatient');

    if (!data.user_id) {
        var errmsg = "error user_id missing";
        logger.error(errmsg);
        res.status(500).json({ msg: errmsg, status: 500 });
    }

    if (caller.id === data.user_id) {
        var errmsg = "error enrolling patient, caller and enrolled patient must be different";
        logger.error(errmsg);
        res.status(500).json({ msg: errmsg, status: 500 });
    }

    kms.getSymKeyAes(function (err, enrollment_sym_key) {
        if (err) {
            var errmsg = "error creating enrollment sym key";
            logger.error(errmsg, err);
            res.status(500).json({ msg: errmsg, status: 500 });
        } else {
            enrollPatient(caller, data, enrollment_sym_key.keyBase64).then(result => {
                logger.info('Enroll patient success');
                const message = 'Patient enrollment completed successfully';
                res.json({ msg: message, tx_id: result.tx_id });
                onSuccess(data);
            }).catch((err) => {
                const errorMessage = 'Enroll patient error';
                logger.error(errorMessage, err);
                res.status(500).json({ msg: errorMessage, status: 500 });
            });
        }
    }, data.enrollment_sym_key);
}

async function unenrollPatientInChaincode(caller, serviceId, userId) {
    return new Promise((resolve, reject) => {
        solutionChaincodeOps.unenrollPatient(caller, serviceId, userId, function (err, result) {
            if (err) return reject(err);

            return resolve(result);
        });
    });
}

async function unenrollPatient(caller, data) {
    let serviceId = data.service_id;
    let userId = data.user_id;

    if (isDeIdentifierServiceEnabled) {
        [serviceId, userId] = await Promise.all(
            [getDeIdentifiedPii(data.service_id), getDeIdentifiedPii(data.user_id)]);
    }

    return unenrollPatientInChaincode(caller, serviceId, userId);
}

function unenrollPatientApiHandler(caller, data, req, res) {
    logger.debug('unenrollPatient');
    unenrollPatient(caller, data).then(result => {
        logger.info('Unenroll patient success');
        const message = 'Patient unenrollment completed successfully';
        res.status(200).json({ msg: message, tx_id: result.tx_id });
        onSuccess(data);
    }).catch((err) => {
        const errorMessage = 'Unenroll patient error';
        logger.error(errorMessage, err);
        res.status(500).json({ msg: errorMessage, status: 500 });
    });
}

async function getPatientEnrollmentsInChaincode(caller, userId, status) {
    return new Promise((resolve, reject) => {
        solutionChaincodeOps.getPatientEnrollments(caller, userId, status, function (err, enrollments) {
            if (err) return reject(err);

            return resolve(enrollments);
        });
    });
}

async function getPatientEnrollments(caller, data) {
    let userId = data.user_id;

    if (isDeIdentifierServiceEnabled) {
        userId = await getDeIdentifiedPii(data.user_id);
    }

    const enrollments = await getPatientEnrollmentsInChaincode(caller, userId, data.status);

    if (isDeIdentifierServiceEnabled && enrollments) {
        return identifyEnrollmentsPii(enrollments);
    }

    return enrollments;
}

function getPatientEnrollmentsApiHandler(caller, data, req, res) {
    logger.debug('getPatientEnrollments');

    const phiAccessDetails = {
        initiatorId: caller.id,
        targetId: data.user_id,
        targetName: 'patientEnrollment',
        message: 'Get patient enrollments.',
        action: 'getPatientEnrollments',
    };

    getPatientEnrollments(caller, data).then(enrollments => {
        logPhiAccess(phiAccessDetails, 200);
        res.status(200).json(enrollments);
    }).catch(err => {
        const errorMessage = 'getPatientEnrollments error';
        logger.error(errorMessage, err);
        logFailedPhiAccess(phiAccessDetails, 500);

        res.status(500).json({ msg: errorMessage, status: 500 });
    });
}

async function getServiceEnrollmentsInChaincode(caller, serviceId, status) {
    return new Promise((resolve, reject) => {
        solutionChaincodeOps.getServiceEnrollments(caller, serviceId, status, function (err, enrollments) {
            if (err) return reject(err);

            return resolve(enrollments);
        });
    });
}

async function identifyEnrollmentsPii(enrollments) {
    if (!Array.isArray(enrollments)) throw new Error('Cannot identify Enrollment records. Expected Array type.');

    const identifiedEnrollmentPromises = enrollments.map(async (enrollment) => {
        const [serviceId, userId] = await Promise.all(
            [getPii(enrollment.service_id), getPii(enrollment.user_id)]);
        enrollment.service_id = serviceId;
        enrollment.user_id = userId;
        return enrollment;
    });

    const successPropmises = identifiedEnrollmentPromises.map((identifiedEnrollmentPromise) => {
        return identifiedEnrollmentPromise.catch((err) => {
            const errorMessage = 'Failed to get enrollment';
            logger.error(errorMessage, err);
        })
    });

    return Promise.all(successPropmises);
}

async function getServiceEnrollments(caller, data) {
    let serviceId = data.service_id;

    if (isDeIdentifierServiceEnabled) {
        serviceId = await getDeIdentifiedPii(data.service_id);
    }

    let enrollments = await getServiceEnrollmentsInChaincode(caller, serviceId, data.status);

    if (isDeIdentifierServiceEnabled && enrollments) {
        enrollments = identifyEnrollmentsPii(enrollments);
    }

    return enrollments;
}

function getServiceEnrollmentsApiHandler(caller, data, req, res) {
    logger.debug('getServiceEnrollments');

    const phiAccessDetails = {
        initiatorId: caller.id,
        targetId: data.service_id,
        targetName: 'serviceEnrollments',
        message: 'Get service enrollments.',
        action: 'getServiceEnrollments',
    };

    getServiceEnrollments(caller, data).then(enrollments => {
        logPhiAccess(phiAccessDetails, 200);
        res.status(200).json(enrollments.filter((enrollment) => enrollment != null));
    }).catch(err => {
        const errorMessage = 'getServiceEnrollments error';
        logger.error(errorMessage, err);
        logFailedPhiAccess(phiAccessDetails, 500);

        res.status(500).json({ msg: errorMessage, status: 500 });
    });
}

async function putConsentPatientDataInChaincode(caller, consent, consentKeyBase64) {
    return new Promise((resolve, reject) => {
        solutionChaincodeOps.putConsentPatientData(caller, consent, consentKeyBase64, function (err, result) {
            if (err) return reject(err);

            return resolve(result);
        });
    });
}

async function putConsentPatientData(data, caller, consentSymKeyBase64) {
    let ownerId = data.owner_id;
    let serviceId = data.service_id;
    let targetServiceId = data.target_service_id;

    if (isDeIdentifierServiceEnabled) {
        [ownerId, serviceId, targetServiceId] = await Promise.all(
            [getDeIdentifiedPii(data.owner_id), getDeIdentifiedPii(data.service_id), getDeIdentifiedPii(data.target_service_id)]);
    }

    const consent = {
        owner: ownerId,
        service: serviceId,
        target: targetServiceId,
        datatype: data.datatype_id,
        expiration: data.expiration,
        option: data.option,
        timestamp: Math.floor(new Date().getTime() / 1000)
    };

    return putConsentPatientDataInChaincode(caller, consent, consentSymKeyBase64);
}

function putConsentPatientDataApiHandler(caller, data, req, res) {
    logger.debug('putConsentPatientDataApiHandler');

    kms.getSymKeyAes(function (err, consentSymKey) {
        if (err) {
            var errmsg = "error creating consent sym key in KMS";
            logger.error(errmsg, err);
            res.status(500).json({ msg: errmsg, status: 500 });
        } else {
            putConsentPatientData(data, caller, consentSymKey.keyBase64).then(result => {
                const message = 'Adding consent for patient data completed successfully';
                res.status(200).json({ msg: message, tx_id: result.tx_id });
            }).catch((err) => {
                const errorMessage = 'putConsentPatientData error';
                logger.error(errorMessage, err);
                res.status(500).json({ msg: errorMessage, status: 500 });
            });
        }
    }, data.consent_sym_key);
}

async function putMultiConsentPatientDataApiHandler(caller, data, req, res) {
    logger.debug('putMultiConsentPatientDataApiHandler');

    let consents = [];

    // sanity check
    for (let consentInput of data.consents) {
        let [isValid, errMsg] = await validateConsentInput(consentInput);

        if (!isValid) {
            res.status(400).json({ msg: errMsg, failure: consentInput, status: 400 });
        }

        var ts = consentInput.expiration;
        if (!ts) {
            ts = 0
        } else if (ts && typeof ts == 'string') {
            let match = ts.match(/\D/g);
            if (match != null) {
                let tsdate = new Date(ts);
                ts = Math.floor(tsdate.getTime() / 1000);
            } else {
                ts = parseInt(ts, 10);
            }
        }

        consents.push({
            owner_id: consentInput.patient_id,
            service_id: consentInput.service_id,
            target_service_id: consentInput.target_id,
            datatype_id: consentInput.datatype_id,
            option: consentInput.option,
            expiration: ts
        });
    }

    let promises = [];
    let successes = [];
    let failures = [];

    for (let consent of consents) {
        let consentSymKey = await kms.getSymKeyAesPromise("");

        let consentPromise = putConsentPatientData(consent, caller, consentSymKey.keyBase64).then(_ => {
            successes.push(consent);
        }).catch((err) => {
            const errorMessage = 'putConsentPatientData error';
            logger.error(errorMessage, err);
            failures.push(consent);
        });

        promises.push(consentPromise);
    }

    await Promise.all(promises);

    const message = 'Adding multiple consents for patient data completed';
    res.status(200).json({ msg: message, successes, failures });
}

async function putMultiConsentPatientData(caller, data, req, res) {
    logger.debug('putMultiConsentPatientData');

    let consents = [];
    let promises = [];
    let successes = [];
    let failures = [];

    // sanity check
    for (let consentInput of data.consents) {
        let [isValid, errMsg] = await validateConsentInputSingleApi(consentInput);

        if (!isValid) {
            res.status(400).json({ msg: errMsg, successes, failures, status: 400, failure_type: consentError });
        }

        var ts = consentInput.expirationTimestamp;
        if (!ts) {
            ts = 0
        } else if (ts && typeof ts == 'string') {
            let match = ts.match(/\D/g);
            if (match != null) {
                let tsdate = new Date(ts);
                ts = Math.floor(tsdate.getTime() / 1000);
            } else {
                ts = parseInt(ts, 10);
            }
        }

        consents.push({
            owner_id: consentInput.owner_id,
            service_id: consentInput.service_id,
            target_service_id: consentInput.target_id,
            datatype_id: consentInput.datatype_id,
            option: consentInput.option,
            expirationTimestamp: ts
        });
    }

    for (let consent of consents) {
        let consentSymKey = await kms.getSymKeyAesPromise("");

        let consentPromise = putConsentPatientData(consent, caller, consentSymKey.keyBase64).then(_ => {
            successes.push(consent);
        }).catch((err) => {
            const errorMessage = 'putConsentPatientData error';
            logger.error(errorMessage, err);
            failures.push(consent);
        });

        promises.push(consentPromise);
    }

    await Promise.all(promises);

    if (failures.length < 1) {
        const message = 'Register, enroll user and adding multiple consents for patient data completed';
        res.status(200).json({ msg: message, successes, failures, status: 200, failure_type: "" });
    } else {
        const errorMessage = 'putMultiConsentPatientData error';
        res.status(500).json({ msg: errorMessage, successes, failures, status: 500, failure_type: consentError });
    }
}

async function validateConsentInput(consentInput) {
    if (!consentInput.patient_id) {
        var errmsg = "Missing patient ID";
        return [false, errmsg];
    } else if (!consentInput.service_id) {
        var errmsg = "Missing service ID";
        return [false, errmsg];
    } else if (!consentInput.target_id) {
        var errmsg = "Missing target ID";
        return [false, errmsg];
    } else if (!consentInput.datatype_id) {
        var errmsg = "Missing datatype ID";
        return [false, errmsg];
    } else if (consentInput.option.length < 1) {
        var errmsg = "Must specify at least one consent option";
        return [false, errmsg];
    } else if (consentInput.option.length > 2) {
        var errmsg = "Too many consent options";
        return [false, errmsg];
    } else if ((consentInput.option.includes("write") && consentInput.option.includes("deny"))
        || (consentInput.option.includes("read") && consentInput.option.includes("deny"))) {
        var errmsg = "Deny cannot be paired with another consent option";
        return [false, errmsg];
    } else if (consentInput.option.filter((option) => option === "read" || option === "write" || option === "deny").length < 1) {
        var errmsg = "Invalid consent option, the option can only be write, read, or deny, or a combination of them.";
        return [false, errmsg];
    }

    return [true, null];
}

async function validateConsentInputSingleApi(consentInput) {
    if (!consentInput.owner_id) {
        var errmsg = "Missing owner_id";
        return [false, errmsg];
    } else if (!consentInput.service_id) {
        var errmsg = "Missing service_id";
        return [false, errmsg];
    } else if (!consentInput.target_id) {
        var errmsg = "Missing target_id";
        return [false, errmsg];
    } else if (!consentInput.datatype_id) {
        var errmsg = "Missing datatype_id";
        return [false, errmsg];
    } else if (consentInput.option.length < 1) {
        var errmsg = "Must specify at least one consent option";
        return [false, errmsg];
    } else if (consentInput.option.length > 2) {
        var errmsg = "Too many consent options";
        return [false, errmsg];
    } else if ((consentInput.option.includes("write") && consentInput.option.includes("deny"))
        || (consentInput.option.includes("read") && consentInput.option.includes("deny"))) {
        var errmsg = "Deny cannot be paired with another consent option";
        return [false, errmsg];
    } else if (consentInput.option.filter((option) => option === "read" || option === "write" || option === "deny").length < 1) {
        var errmsg = "Invalid consent option, the option can only be write, read, or deny, or a combination of them.";
        return [false, errmsg];
    }

    return [true, null];
}

async function putConsentOwnerDataInChaincode(caller, consent, consentSymKeyBase64) {
    return new Promise((resolve, reject) => {
        solutionChaincodeOps.putConsentOwnerData(caller, consent, consentSymKeyBase64, function (err, result) {
            if (err) return reject(err);

            return resolve(result);
        });
    });
}

async function putConsentOwnerData(caller, data, consentSymKeyBase64) {
    let ownerId = data.owner_id;
    let targetId = data.target_id;

    if (isDeIdentifierServiceEnabled) {
        [ownerId, targetId] = await Promise.all(
            [deIdentifyPii(data.owner_id), deIdentifyPii(data.target_id)]);
    }

    const consent = {
        owner: ownerId,
        target: targetId,
        datatype: data.datatype_id,
        expiration: data.expiration,
        option: data.option,
        timestamp: Math.floor(new Date().getTime() / 1000)
    }

    return putConsentOwnerDataInChaincode(caller, consent, consentSymKeyBase64);
}

function putConsentOwnerDataApiHandler(caller, data, req, res) {
    logger.debug('putConsentOwnerDataApiHandler');

    kms.getSymKeyAes(function (err, consentSymKey) {
        if (err) {
            var errmsg = "error creating consent sym key in KMS";
            logger.error(errmsg, err);
            res.status(500).json({ msg: errmsg, status: 500 });
        } else {
            putConsentOwnerData(caller, data, consentSymKey.keyBase64).then(result => {
                const message = 'adding consent for owner data completed successfully';
                res.status(200).json({ msg: message, tx_id: result.tx_id });
            }).catch((err) => {
                const errorMessage = "putConsentOwnerData error";
                logger.error(errorMessage, err);
                res.status(500).json({ msg: errorMessage, status: 500 });
            });
        }
    }, data.consent_sym_key);
}

async function getConsentFromChaincode(caller, ownerId, targetId, datatypeId) {
    return new Promise(resolve => {
        solutionChaincodeOps.getConsent(caller, ownerId, targetId, datatypeId, function (err, consent) {
            resolve([err, consent]);
        });
    });
}

async function getConsent(caller, data) {
    let ownerId = data.owner_id;
    let targetId = data.target_id;

    if (isDeIdentifierServiceEnabled) {
        [ownerId, targetId] = await Promise.all(
            [getDeIdentifiedPii(data.owner_id), getDeIdentifiedPii(data.target_id)]);
    }

    let [err, consent] = await getConsentFromChaincode(caller, ownerId, targetId, data.datatype_id);

    if (err) {
        return [err, consent];
    }

    if (isDeIdentifierServiceEnabled) {
        consent = await identifyConsentPii(consent);
    }

    return [null, consent];
}

function getConsentApiHandler(caller, data, req, res) {
    logger.debug('getConsent');

    const phiAccessDetails = {
        initiatorId: caller.id,
        targetId: `${data.owner_id}-${data.target_id}`,
        targetName: 'consent',
        message: 'Get consent.',
        action: 'getConsent',
    };

    getConsent(caller, data).then(([err, consent]) => {
        if (err != null) {
            logger.error(err);
            logFailedPhiAccess(phiAccessDetails, 404);
            res.json({});
        }
        else {
            logPhiAccess(phiAccessDetails, 200);
            res.json(consent);
        }
    }).catch(err => {
        const errorMessage = "getConsent error";
        logger.error(errorMessage, err);
        logFailedPhiAccess(phiAccessDetails, 500);

        res.status(500).json({ msg: errorMessage, status: 500 });
    });
}

async function getConsentOwnerDataFromChaincode(caller, ownerService, service, datatype) {
    return new Promise(resolve => {
        solutionChaincodeOps.getConsentOwnerData(caller, ownerService, service, datatype, function (err, consent) {
            resolve([err, consent]);
        });
    });
}

async function identifyConsentPii(consent) {
    const [ownerId, targetId] = await Promise.all(
        [getPii(consent.owner), getPii(consent.target)]);
    consent.owner = ownerId;
    consent.target = targetId;

    if (consent.service) {
        consent.service = await getPii(consent.service);
    }
    if (consent.requester) {
        consent.requester = await getPii(consent.requester);
    }

    return consent;
}

async function getConsentOwnerData(caller, data) {
    let ownerService = data.owner_service;
    let service = data.service;

    if (isDeIdentifierServiceEnabled) {
        [ownerService, service] = await Promise.all(
            [getDeIdentifiedPii(data.owner_service), getDeIdentifiedPii(data.service)]);
    }

    let [err, consent] = await getConsentOwnerDataFromChaincode(caller, ownerService, service, data.datatype);

    if (err) return [err, consent];

    if (isDeIdentifierServiceEnabled) {
        consent = await identifyConsentPii(consent);
    }

    return [null, consent];
}

function getConsentOwnerDataApiHandler(caller, data, req, res) {
    logger.debug('getConsentOwnerData');

    const phiAccessDetails = {
        initiatorId: caller.id,
        targetId: `${data.owner_service}-${data.service}-${data.datatype}`,
        targetName: 'consentOwnerData',
        message: 'Get consent owner data.',
        action: 'getConsentOwnerData',
    };

    getConsentOwnerData(caller, data).then(([err, consent]) => {
        if (err != null) {
            logger.error(err);
            logFailedPhiAccess(phiAccessDetails, 404);
            res.json({});
        } else {
            logPhiAccess(phiAccessDetails, 200);
            res.json(consent);
        }
    }).catch(err => {
        const errorMessage = 'getConsentOwnerData error';
        logger.error(errorMessage, err);
        logFailedPhiAccess(phiAccessDetails, 500);

        res.status(500).json({ msg: errorMessage, status: 500 });
    });
}

async function getConsentsInChaincode(caller, serviceId, userId) {
    return new Promise(resolve => {
        solutionChaincodeOps.getConsents(caller, serviceId, userId, function (err, consents) {
            resolve([err, consents]);
        });
    });
}

async function identifyConsentsPii(consents) {
    if (!Array.isArray(consents)) throw new Error('Cannot identify Consent records. Expected Array type.');

    const identifiedConsentPromises = consents.map(identifyConsentPii);
    return Promise.all(identifiedConsentPromises);
}

async function getConsents(caller, data) {
    let serviceId = data.service;
    let userId = data.user;

    if (isDeIdentifierServiceEnabled) {
        [serviceId, userId] = await Promise.all(
            [getDeIdentifiedPii(data.service), getDeIdentifiedPii(data.user)]);
    }

    let [err, consents] = await getConsentsInChaincode(caller, serviceId, userId);

    if (isDeIdentifierServiceEnabled && consents) {
        consents = await identifyConsentsPii(consents);
    }

    return [err, consents];
}

function getConsentsApiHandler(caller, data, req, res) {
    logger.debug('getConsents');

    const phiAccessDetails = {
        initiatorId: caller.id,
        targetId: `${data.service}-${data.user}`,
        targetName: 'consents',
        message: 'Get consents.',
        action: 'getConsents',
    };

    getConsents(caller, data).then(([err, consents]) => {
        if (err != null) {
            logger.error(err);
            logFailedPhiAccess(phiAccessDetails, 404);
            res.json([]);
        }
        else {
            logPhiAccess(phiAccessDetails, 200);
            res.json(consents);
        }
    }).catch(err => {
        const errorMessage = 'Failed to get consents';
        logger.error(errorMessage, err);
        logFailedPhiAccess(phiAccessDetails, 500);

        res.status(500).json({ msg: errorMessage, status: 500 });
    });
}

async function getConsentsWithOwnerIdFromChaincode(caller, ownerId) {
    return new Promise(resolve => {
        solutionChaincodeOps.getConsentsWithOwnerID(caller, ownerId, function (err, consent) {
            return resolve([err, consent]);
        });
    });
}

async function getConsentsWithOwnerId(caller, data) {
    let ownerId = data.owner_id;

    if (isDeIdentifierServiceEnabled) {
        ownerId = await getDeIdentifiedPii(data.owner_id);
    }

    let [err, consents] = await getConsentsWithOwnerIdFromChaincode(caller, ownerId);

    if (err) return [err, null];

    if (isDeIdentifierServiceEnabled && consents) {
        consents = await identifyConsentsPii(consents);
    }

    return [err, consents];
}

function getConsentsWithOwnerIDApiHandler(caller, data, req, res) {
    logger.debug('getConsentsWithOwnerID');

    const phiAccessDetails = {
        initiatorId: caller.id,
        targetId: data.owner_id,
        targetName: 'consentsWithOwnerId',
        message: 'Get consents with owner id.',
        action: 'getConsentsWithOwnerID',
    };

    getConsentsWithOwnerId(caller, data).then(([err, consents]) => {
        if (err != null) {
            logger.error(err);
            logFailedPhiAccess(phiAccessDetails, 404);
            res.json({});
        }
        else {
            logPhiAccess(phiAccessDetails, 200);
            res.json(consents);
        }
    }).catch(err => {
        const errorMessage = 'Failed to get consents with owner ID';
        logger.error(errorMessage, err);
        logFailedPhiAccess(phiAccessDetails, 500);

        res.status(500).json({ msg: errorMessage, status: 500 });
    });
}

async function getConsentsWithTargetIdInChaincode(caller, targetId) {
    return new Promise(resolve => {
        solutionChaincodeOps.getConsentsWithTargetID(caller, targetId, function (err, consent) {
            resolve([err, consent]);
        });
    });
}

async function getConsentsWithTargetId(caller, data) {
    let targetId = data.target_id;

    if (isDeIdentifierServiceEnabled) {
        targetId = await getDeIdentifiedPii(data.target_id);
    }

    let [err, consents] = await getConsentsWithTargetIdInChaincode(caller, targetId);

    if (err) return [err, consents];

    if (isDeIdentifierServiceEnabled && consents) {
        consents = await identifyConsentsPii(consents);
    }

    return [err, consents];
}

function getConsentsWithTargetIDApiHandler(caller, data, req, res) {
    logger.debug('getConsentsWithTargetID');

    const phiAccessDetails = {
        initiatorId: caller.id,
        targetId: data.target_id,
        targetName: 'consentsWithTargetId',
        message: 'Get consents with Target ID.',
        action: 'getConsentsWithTargetID',
    };

    getConsentsWithTargetId(caller, data).then(([err, consent]) => {
        if (err != null) {
            logger.error(err);
            logFailedPhiAccess(phiAccessDetails, 404);
            res.json({});
        }
        else {
            logPhiAccess(phiAccessDetails, 200);
            res.json(consent);
        }
    }).catch(err => {
        const errorMessage = 'Failed to get consent with target id';
        logger.error(errorMessage, err);
        logFailedPhiAccess(phiAccessDetails, 500);

        res.status(500).json({ msg: errorMessage, status: 500 });
    });
}

async function identifyPIIConsentRequests(consentRequests) {
  if (!Array.isArray(consentRequests)) throw new Error('Cannot identify Consent Request records. Expect Array type.');

  const consentRequestPromises = consentRequests.map(async (request) => {
    const { user, org, service, service_name } = request;
    [request.user, request.org, request.service] = await Promise.all([getPii(user), getPii(org), getPii(service)]);

    if (service_name) {
      request.service_name = await getPii(service_name);
    }
    return request;
  });

  return Promise.all(consentRequestPromises);
}

async function getConsentRequestsFromChaincode(caller, patient, service) {
    return new Promise(resolve => {
        solutionChaincodeOps.getConsentRequests(caller, patient, service, function (err, consentRequests) {
            resolve([err, consentRequests]);
        });
    });
}

async function getConsentRequests(caller, data) {
    let { patient, service } = data;

    if (isDeIdentifierServiceEnabled) {
        [patient, service] = await Promise.all([getDeIdentifiedPii(patient), service && getDeIdentifiedPii(service)]);
    }

    let [err, consentRequests] = await getConsentRequestsFromChaincode(caller, patient, service);

    if (err) return [err, null];

    if (isDeIdentifierServiceEnabled && consentRequests) {
        consentRequests = await identifyPIIConsentRequests(consentRequests);
    }

    return [err, consentRequests];
}

function getConsentRequestsApiHandler(caller, data, req, res) {
    logger.debug('getConsentRequests');

    const phiAccessDetails = {
        initiatorId: caller.id,
        targetId: `${data.patient}-${data.service}`,
        targetName: 'consentsRequests',
        message: 'Get current and pending consent requests for given patient.',
        action: 'getConsentRequests',
    };

    getConsentRequests(caller, data).then(([err, consentRequests]) => {
        if (err != null) {
            logger.error(err);
            logFailedPhiAccess(phiAccessDetails, 404);
            res.json([]);
        }
        else {
            logPhiAccess(phiAccessDetails, 200);
            res.json(consentRequests);
        }
    }).catch(err => {
        const errorMessage = 'Failed to get consent requests (getConsentRequests error).';
        logger.error(errorMessage, err);
        logFailedPhiAccess(phiAccessDetails, 500);

        res.status(500).json({ msg: errorMessage, status: 500 });
    });
}

async function validateConsentInChaincode(caller, ownerId, serviceId, datatype, access, timestamp) {
    return new Promise(resolve => {
        solutionChaincodeOps.validateConsent(caller, ownerId, serviceId, datatype, access, timestamp, function (err, val) {
            resolve([err, val]);
        });
    });
}

async function validateConsent(caller, data) {
    let { owner, service } = data;

    if (isDeIdentifierServiceEnabled) {
        [owner, service] = await Promise.all([deIdentifyPii(owner), deIdentifyPii(service)]);
    }

    const timestamp = '' + Math.floor(new Date().getTime() / 1000);
    let [err, consent] = await validateConsentInChaincode(caller, owner, service, data.datatype, data.access, timestamp);

    if (err) return [err, null];

    if (isDeIdentifierServiceEnabled && consent) {
        consent = await identifyConsentPii(consent);
    }

    return [err, consent];
}

function validateConsentApiHandler(caller, data, req, res) {
    logger.debug('validateConsent');

    const phiAccessDetails = {
        initiatorId: caller.id,
        targetId: `${data.owner}-${data.service}-${data.datatype}-${data.access}`,
        targetName: 'consentValidation',
        message: 'Validates consent for accessing the data.',
        action: 'validateConsent',
    };

    validateConsent(caller, data).then(([err, val]) => {
        if (err != null) {
            logger.error(err);
            logFailedPhiAccess(phiAccessDetails, 404);
            res.json({});
        }
        else {
            logPhiAccess(phiAccessDetails, 200);
            res.json(val);
        }
    }).catch(err => {
        const errorMessage = 'Validate consent error';
        logger.error(errorMessage, err);
        logFailedPhiAccess(phiAccessDetails, 500);

        res.status(500).json({ msg: errorMessage, status: 500 });
    });
}

async function uploadUserDataInChaincode(caller, userdata, dataSymKeyB64) {
    return new Promise((resolve, reject) => {
        solutionChaincodeOps.uploadUserData(caller, userdata, dataSymKeyB64, function (err, result) {
            if (err) return reject(err);

            return resolve(result);
        });
    });
}

async function uploadUserData(caller, data, data_sym_key, res) {
    let service = data.service;
    let user = data.user;
    if (isDeIdentifierServiceEnabled) {
        [service, user] = await Promise.all([deIdentifyPii(data.service), deIdentifyPii(data.user)]);
    }

    const timestamp = Math.floor(new Date().getTime() / 1000);
    const userdata = {
        service: service,
        owner: user,
        datatype: data.datatype,
        timestamp: timestamp,
        data: data.userdata
    }

    return uploadUserDataInChaincode(caller, userdata, data_sym_key);
}

function uploadUserDataApiHandler(caller, data, req, res) {
    logger.debug('uploadUserData');

    kms.getSymKeyAes(function (err, symkey) {

        if (err) {
            var errmsg = "error creating contract sym key";
            logger.error(errmsg, err);
            res.status(500).json({ msg: errmsg, status: 500 });
        } else {
            var data_sym_key = symkey.keyBase64;
            uploadUserData(caller, data, data_sym_key, res).then(result => {
                const message = 'Upload user data completed successfully';
                res.status(200).json({ msg: message, tx_id: result.tx_id });
            }).catch(err => {
                const errorMessage = 'uploadUserData error';
                logger.error(errorMessage, err);
                res.status(500).json({ msg: errorMessage, status: 500 });
            });
        }
    }, data.symkey);
}

async function identifyUserData(userDataArray) {
    const userDataPromises = userDataArray.map(async userData => {
        [userData.owner, userData.service] = await Promise.all([getPii(userData.owner), getPii(userData.service)]);
        return userData;
    });

    return Promise.all(userDataPromises);
}

async function downloadUserDataInChaincode(caller, serviceId, userId, datatypeId, startTimestamp, endTimestamp, latestOnly, maxNum, timestamp) {
    return new Promise((resolve, reject) => {
        solutionChaincodeOps.downloadUserData(
            caller,
            serviceId,
            userId,
            datatypeId,
            startTimestamp,
            endTimestamp,
            latestOnly,
            maxNum,
            timestamp,
            function (err, result) {
                if (err) return reject(err);

                return resolve(result);
            });
    });
}

async function downloadUserData(caller, data) {
    let serviceId = data.service_id;
    let userId = data.user_id;
    if (isDeIdentifierServiceEnabled) {
        [serviceId, userId] = await Promise.all([getDeIdentifiedPii(data.service_id), getDeIdentifiedPii(data.user_id)]);
    }

    const timestamp = '' + Math.floor(new Date().getTime() / 1000);
    const userDataArray = await downloadUserDataInChaincode(caller, serviceId, userId, data.datatype_id, data.start_timestamp + '', data.end_timestamp + '', data.latest_only + '', data.maxNum, timestamp);

    if (isDeIdentifierServiceEnabled) {
        return identifyUserData(userDataArray);
    }

    return userDataArray;
}

function downloadUserDataApiHandler(caller, data, req, res) {
    logger.debug('downloadUserData', data);
    if (!data.start_timestamp) {
        data.start_timestamp = "0";
    }
    if (!data.end_timestamp) {
        data.end_timestamp = "0";
    }
    if (!data.latest_only) {
        data.latest_only = "false"
    }
    if (!data.maxNum) {
        data.maxNum = "1000"
    }

    const phiAccessDetails = {
        initiatorId: caller.id,
        targetId: `userDataByServiceUserDatatype-${data.service_id}-${data.user_id}-${data.datatype_id}`,
        targetName: 'userData',
        message: 'Get User Data by Service Id, User Id, and Datatype Id.',
        action: 'downloadUserData',
        requestData: {
            startTimestamp: data.start_timestamp,
            endTimestamp: data.end_timestamp,
            latestOnly: data.latest_only,
            maxNum: data.maxNum,
        },
    };

    downloadUserData(caller, data).then(result => {
        logPhiAccess(phiAccessDetails, 200);
        logger.debug('Success downloading user data');
        res.status(200).json(result);
    }).catch(err => {
        const errorMessage = 'download user data error';
        logger.error(errorMessage, err);
        logFailedPhiAccess(phiAccessDetails, 500);

        res.status(500).json({ msg: errorMessage, status: 500 });
    });
}

async function downloadUserDataConsentToken(caller, data) {
  const timestamp = Math.floor(new Date().getTime() / 1000) + '';
  const userData = await new Promise((resolve, reject) => {
    solutionChaincodeOps.downloadUserDataConsentToken(caller, data.start_timestamp + '', data.end_timestamp + '', data.latest_only + '', data.maxNum, timestamp, data.token, (err, userData) => {
      if (err) return reject(err);
      return resolve(userData);
    });
  });

  if (isDeIdentifierServiceEnabled) {
    return identifyUserData(userData);
  }

  return userData;
}

function downloadUserDataConsentTokenApiHandler(caller, data, req, res) {
    logger.debug('downloadUserDataConsentToken', data);
    if (!data.start_timestamp) {
        data.start_timestamp = "0";
    }
    if (!data.end_timestamp) {
        data.end_timestamp = "0";
    }
    if (!data.latest_only) {
        data.latest_only = "false"
    }
    if (!data.maxNum) {
        data.maxNum = "1000"
    }

    const phiAccessDetails = {
        initiatorId: caller.id,
        targetId: `userDataByConsentToken`,
        targetName: 'userData',
        message: 'Get User Data by Consent Validation Token.',
        action: 'downloadUserDataConsentToken',
        requestData: {
            startTimestamp: data.start_timestamp,
            endTimestamp: data.end_timestamp,
            latestOnly: data.latest_only,
            maxNum: data.maxNum,
        },
    };

    downloadUserDataConsentToken(caller, data).then(userData => {
        logger.debug('Success downloading data with consent token');
        logPhiAccess(phiAccessDetails, 200);
        res.status(200).json(userData);
    }).catch(err => {
        const message = 'download data with consent token error';
        logger.error(message, err);
        logFailedPhiAccess(phiAccessDetails, 500);
        res.status(500).json({ msg: message, status: 500 });
    });
}

async function deleteUserDataInChaincode(caller, user, datatype, service, timestamp) {
    return new Promise((resolve, reject) => {
        solutionChaincodeOps.deleteUserData(caller, user, datatype, service, timestamp, function (err, result) {
            if (err) return reject(err);

            return resolve(result);
        });
    });
}

async function deleteUserData(caller, data) {
    let user = data.user;
    let service = data.service;
    if (isDeIdentifierServiceEnabled) {
        [user, service] = await Promise.all(
            [getDeIdentifiedPii(data.user), getDeIdentifiedPii(data.service)]);
    }

    return deleteUserDataInChaincode(caller, user, data.datatype, service, data.timestamp + '');
}

function deleteUserDataApiHandler(caller, data, req, res) {
    logger.debug('deleteUserData');
    deleteUserData(caller, data).then(result => {
        logger.debug('Success deleting user data');
        const message = 'delete user data completed successfully';
        res.status(200).json({ msg: message, tx_id: result.tx_id });
    }).catch(err => {
        const errorMessage = 'deleteUserData error';
        logger.error(errorMessage, err);
        res.status(500).json({ msg: errorMessage, status: 500 });
    });
}

async function uploadOwnerDataInChaincode(caller, ownerData, data_sym_key) {
    return new Promise((resolve, reject) => {
        solutionChaincodeOps.uploadOwnerData(caller, ownerData, data_sym_key, function (err, result) {
            if (err) return reject(err);

            return resolve(result);
        });
    });
}

async function uploadOwnerData(caller, data, data_sym_key, res) {
    let service = data.service;

    if (isDeIdentifierServiceEnabled) {
        service = await deIdentifyPii(data.service);
    }

    const timestamp = Math.floor(new Date().getTime() / 1000);
    const ownerData = {
        service: service,
        owner: service,
        datatype: data.datatype,
        timestamp: timestamp,
        data: data.ownerData
    }

    return uploadOwnerDataInChaincode(caller, ownerData, data_sym_key);
}

function uploadOwnerDataApiHandler(caller, data, req, res) {
    logger.debug('uploadOwnerData');
    kms.getSymKeyAes(function (err, symkey) {

        if (err) {
            var errmsg = "error creating contract sym key";
            logger.error(errmsg, err);
            res.status(500).json({ msg: errmsg, status: 500 });
        } else {
            var data_sym_key = symkey.keyBase64;
            uploadOwnerData(caller, data, data_sym_key, res).then(result => {
                const message = 'upload owner data completed successfully';
                logger.debug(message);
                res.status(200).json({ msg: message, tx_id: result.tx_id });
            }).catch(err => {
                const errorMessage = 'uploadOwnerData error';
                logger.error(errorMessage, err);
                res.status(500).json({ msg: errorMessage, status: 500 });
            });
        }
    }, data.symkey);
}

async function identifyOwnerData(ownerDataArray) {
  const ownerDataPromises = ownerDataArray.map(async ownerData => {
    [ ownerData.owner, ownerData.service ] = await Promise.all([getPii(ownerData.owner), getPii(ownerData.service)]);

    return ownerData;
  });

  return Promise.all(ownerDataPromises);
}

async function downloadOwnerDataAsOwnerInChaincode(
    caller, serviceId, datatypeId, startTimestamp, endTimestamp, latestOnly, maxNum, timestamp) {
    return new Promise((resolve, reject) => {
        solutionChaincodeOps.downloadOwnerDataAsOwner(
            caller,
            serviceId,
            datatypeId,
            startTimestamp,
            endTimestamp,
            latestOnly,
            maxNum,
            timestamp,
            function (err, result) {
                if (err) return reject(err);

                return resolve(result);
            });
    });
}

async function downloadOwnerDataAsOwner(caller, data) {
    let serviceId = data.service_id;

    if (isDeIdentifierServiceEnabled) {
        serviceId = await getDeIdentifiedPii(data.service_id);
    }

    const timestamp = "" + Math.floor(new Date().getTime() / 1000);
    const ownerData = await downloadOwnerDataAsOwnerInChaincode(caller,
        serviceId,
        data.datatype_id,
        data.start_timestamp + '',
        data.end_timestamp + '',
        data.latest_only + '',
        data.maxNum + '',
        timestamp);

    if (isDeIdentifierServiceEnabled) {
        return identifyOwnerData(ownerData);
    }

    return ownerData;
}

function downloadOwnerDataAsOwnerApiHandler(caller, data, req, res) {
    logger.debug('downloadOwnerDataAsOwner');
    if (!data.start_timestamp) {
        data.start_timestamp = "0";
    }
    if (!data.end_timestamp) {
        data.end_timestamp = "0";
    }
    if (!data.latest_only) {
        data.latest_only = "false"
    }
    if (!data.maxNum) {
        data.maxNum = "1000"
    }

    const phiAccessDetails = {
        initiatorId: caller.id,
        targetId: `ownerDataByServiceIdAndDatatype-${data.service_id}-${data.datatype_id}`,
        targetName: 'ownerData',
        message: 'Get Owner Data by Service ID and datatype as data owner.',
        action: 'downloadOwnerDataAsOwner',
        requestData: {
            startTimestamp: data.start_timestamp,
            endTimestamp: data.end_timestamp,
            latestOnly: data.latest_only,
            maxNum: data.maxNum,
        },
    };

    downloadOwnerDataAsOwner(caller, data).then(result => {
        logPhiAccess(phiAccessDetails, 200);
        res.status(200).json(result);
    }).catch(err => {
        const errorMessage = 'get owner data error';
        logger.error(errorMessage, err);
        logFailedPhiAccess(phiAccessDetails, 500);

        res.status(500).json({ msg: errorMessage, status: 500 });
    });
}

async function downloadOwnerDataAsRequester(caller, data) {
  const timestamp = Math.floor(new Date().getTime() / 1000) + '';
  const ownerData = await  new Promise((resolve, reject) => {
    solutionChaincodeOps.downloadOwnerDataAsRequester(caller, data.contract_id, data.datatype_id, data.start_timestamp + '', data.end_timestamp + '', data.latest_only + '', data.maxNum + '', timestamp, function (err, result) {
      if (err) return reject(err);
      return resolve(result);
    });
  });

  if (isDeIdentifierServiceEnabled) {
    return identifyOwnerData(ownerData);
  }

  return ownerData;
}

function downloadOwnerDataAsRequesterApiHandler(caller, data, req, res) {
    logger.debug('downloadOwnerDataAsRequester');
    if (!data.start_timestamp) {
        data.start_timestamp = "0";
    }
    if (!data.end_timestamp) {
        data.end_timestamp = "0";
    }
    if (!data.latest_only) {
        data.latest_only = "false"
    }
    if (!data.maxNum) {
        data.maxNum = "1000"
    }

    const phiAccessDetails = {
        initiatorId: caller.id,
        targetId: `ownerDataByContractAndDatatype-${data.contract_id}-${data.datatype_id}`,
        targetName: 'ownerData',
        message: 'Download Owner Data by contract ID and data type as the data requester.',
        action: 'downloadOwnerDataAsRequester',
        requestData: {
            startTimestamp: data.start_timestamp,
            endTimestamp: data.end_timestamp,
            latestOnly: data.latest_only,
            maxNum: data.maxNum,
        },
    };

    downloadOwnerDataAsRequester(caller, data).then(result => {
        logPhiAccess(phiAccessDetails, 200);
        res.status(200).json(result);
    }).catch(err => {
        const message = 'download owner data error';
        logger.error(message, err);
        logFailedPhiAccess(phiAccessDetails, 500);
        res.status(500).json({ msg: message, status: 500 });
    });
}

async function downloadOwnerDataWithConsentInChaincode(
    caller,
    serviceId,
    ownerServiceId,
    datatypeId,
    startTimestamp,
    endTimestamp,
    latestOnly,
    maxNum,
    timestamp,
    token) {
    return new Promise((resolve, reject) => {
        solutionChaincodeOps.downloadOwnerDataWithConsent(
            caller,
            serviceId,
            ownerServiceId,
            datatypeId,
            startTimestamp,
            endTimestamp,
            latestOnly,
            maxNum,
            timestamp,
            token,
            function (err, result) {
                if (err) return reject(err);

                return resolve(result);
            });
    });
}

async function downloadOwnerDataWithConsent(caller, data) {
    let serviceId = data.service_id;
    let ownerServiceId = data.owner_service_id;

    if (isDeIdentifierServiceEnabled) {
        [serviceId, ownerServiceId] = await Promise.all(
            [getDeIdentifiedPii(data.service_id), getDeIdentifiedPii(data.owner_service_id)]);
    }

    const timestamp = "" + Math.floor(new Date().getTime() / 1000);
    const ownerData = await downloadOwnerDataWithConsentInChaincode(
        caller,
        serviceId,
        ownerServiceId,
        data.datatype_id,
        data.start_timestamp + '',
        data.end_timestamp + '',
        data.latest_only + '',
        data.maxNum,
        timestamp,
        data.token);

    if (isDeIdentifierServiceEnabled) {
      return identifyOwnerData(ownerData);
    }

    return ownerData;
}

function downloadOwnerDataWithConsentApiHandler(caller, data, req, res) {
    logger.debug('downloadOwnerDataWithConsent', data);
    if (!data.start_timestamp) {
        data.start_timestamp = "0";
    }
    if (!data.end_timestamp) {
        data.end_timestamp = "0";
    }
    if (!data.latest_only) {
        data.latest_only = "false"
    }
    if (!data.maxNum) {
        data.maxNum = "1000"
    }

    const phiAccessDetails = {
        initiatorId: caller.id,
        targetId: `ownerData-${data.service_id}-${data.owner_service_id}-${data.datatype_id}`,
        targetName: 'ownerData',
        message: 'Get Owner Data with consent from data owner.',
        action: 'downloadOwnerDataWithConsent',
        requestData: {
            startTimestamp: data.start_timestamp,
            endTimestamp: data.end_timestamp,
            latestOnly: data.latest_only,
            maxNum: data.maxNum,
        },
    };

    downloadOwnerDataWithConsent(caller, data).then(result => {
        logger.debug('Success downloading owner data with consent');
        logPhiAccess(phiAccessDetails, 200);
        res.status(200).json(result);
    }).catch(err => {
        const errorMessage = 'download owner data with consent error';
        logger.error(errorMessage, err);
        logFailedPhiAccess(phiAccessDetails, 500);

        res.status(500).json({ msg: errorMessage, status: 500 });
    });
}

async function downloadOwnerDataConsentToken(caller, data) {
  const timestamp = Math.floor(new Date().getTime() / 1000) + '';
  const ownerData = await new Promise((resolve, reject) => {
    solutionChaincodeOps.downloadOwnerDataConsentToken(caller, data.start_timestamp + '', data.end_timestamp + '', data.latest_only + '', data.maxNum, timestamp, data.token, function (err, result) {
      if (err) return reject(err);
      return resolve(result);
    });
  });

  if (isDeIdentifierServiceEnabled) {
    return identifyOwnerData(ownerData);
  }

  return ownerData;
}

function downloadOwnerDataConsentTokenApiHandler(caller, data, req, res) {
    logger.debug('downloadOwnerDataConsentToken', data);
    if (!data.start_timestamp) {
        data.start_timestamp = "0";
    }
    if (!data.end_timestamp) {
        data.end_timestamp = "0";
    }
    if (!data.latest_only) {
        data.latest_only = "false"
    }
    if (!data.maxNum) {
        data.maxNum = "1000"
    }

    const phiAccessDetails = {
        initiatorId: caller.id,
        targetId: `ownerDataByConsentToken`,
        targetName: 'ownerData',
        message: 'Get Owner Data by Consent Validation Token.',
        action: 'downloadOwnerDataConsentToken',
        requestData: {
            startTimestamp: data.start_timestamp,
            endTimestamp: data.end_timestamp,
            latestOnly: data.latest_only,
            maxNum: data.maxNum,
        },
    };

    downloadOwnerDataConsentToken(caller, data).then(result => {
        logger.debug('Success downloading data with consent token');
        logPhiAccess(phiAccessDetails, 200);
        res.status(200).json(result);
    }).catch(err => {
        const message = 'download data with consent token error';
        logger.error(message, err);
        logFailedPhiAccess(phiAccessDetails, 500);
        res.status(500).json({ msg: message, status: 500 });
    });
}

async function createContractInChaincode(caller, contract, contractSymKey) {
    return new Promise((resolve, reject) => {
        solutionChaincodeOps.createContract(caller, contract, contractSymKey, function (err, result) {
            if (err) return reject(err);

            return resolve(result);
        });
    });
}

async function createContract(caller, data, contractUUID, contractSymKey) {
    let { owner_org_id, owner_service_id, requester_org_id, requester_service_id } = data;

    if (isDeIdentifierServiceEnabled) {
        [owner_org_id, owner_service_id, requester_org_id, requester_service_id] = await Promise.all(
            [getDeIdentifiedPii(owner_org_id),
            getDeIdentifiedPii(owner_service_id),
            getDeIdentifiedPii(requester_org_id),
            getDeIdentifiedPii(requester_service_id)]);
    }

    const timestamp = Math.floor(new Date().getTime() / 1000);
    const contract = {
        contract_id: contractUUID,
        owner_org_id,
        owner_service_id,
        requester_org_id,
        requester_service_id,
        contract_terms: data.terms || {},
        state: 'new',
        create_date: timestamp,
        update_date: timestamp,
        contract_detail: [],
        payment_required: 'no',
        payment_verified: 'no',
        max_num_download: 0,
        num_download: 0
    };

    return createContractInChaincode(caller, contract, contractSymKey);
}

function createContractApiHandler(caller, data, req, res) {
    logger.debug('createContract');

    kms.getSymKeyAes(function (err, symkey) {

        if (err) {
            var errmsg = "error creating contract sym key";
            logger.error(errmsg, err);
            res.status(500).json({ msg: errmsg, status: 500 });
        } else {
            var contract_sym_key = symkey.keyBase64;
            const contractUUID = 'contract-' + kms.getUuid();
            createContract(caller, data, contractUUID, contract_sym_key).then(result => {
                res.json({
                    contract_id: contractUUID,
                    msg: 'contract created successfully',
                    tx_id: result.tx_id
                });

                onSuccess(data);
            }).catch(err => {
                const message = 'createContract error';
                logger.error(message, err);
                res.status(500).json({ msg: message, status: 500 });
            });
        }
    }, data.symkey);
}

function payContractApiHandler(caller, data, req, res) {
    logger.debug('payContract');
    var timestamp = "" + Math.floor(new Date().getTime() / 1000);
    var detailType = "payment";
    solutionChaincodeOps.addContractDetail(caller, data.contract_id, detailType, data.contract_terms, timestamp, function (err, result) {
        if (err != null) {
            var errmsg = "pay contract error";
            logger.error(errmsg, err);
            res.status(500).json({ msg: errmsg, status: 500 });
        }
        else {
            res.json({
                contract_id: data.contract_id,
                msg: "contract payment recorded successfully",
                tx_id: result.tx_id
            });
            onSuccess(data);
        }
    });
}

function verifyContractPayment(caller, data, req, res) {
    logger.debug('verifyContractPayment');
    var timestamp = "" + Math.floor(new Date().getTime() / 1000);
    var detailType = "verify";
    solutionChaincodeOps.addContractDetail(caller, data.contract_id, detailType, data.contract_terms, timestamp, function (err, result) {
        if (err != null) {
            var errmsg = "verify contract error";
            logger.error(errmsg, err);
            res.status(500).json({ msg: errmsg, status: 500 });
        }
        else {
            res.json({
                contract_id: data.contract_id,
                msg: "contract payment verified successfully",
                tx_id: result.tx_id
            });
            onSuccess(data);
        }
    });
}

function changeContractTermsApiHandler(caller, data, req, res) {
    logger.debug('changeContractTerms');
    var timestamp = "" + Math.floor(new Date().getTime() / 1000);
    var detailType = "terms";
    solutionChaincodeOps.addContractDetail(caller, data.contract_id, detailType, data.contract_terms, timestamp, function (err, result) {
        if (err != null) {
            var errmsg = "add contract terms error";
            logger.error(errmsg, err);
            res.status(500).json({ msg: errmsg, status: 500 });
        }
        else {
            res.json({
                contract_id: data.contract_id,
                msg: "contract terms changed successfully",
                tx_id: result.tx_id
            });
            onSuccess(data);
        }
    });
}

function signContractApiHandler(caller, data, req, res) {
    logger.debug('signContract');
    var timestamp = "" + Math.floor(new Date().getTime() / 1000);
    var detailType = "sign";
    solutionChaincodeOps.addContractDetail(caller, data.contract_id, detailType, data.contract_terms, timestamp, function (err, result) {
        if (err != null) {
            var errmsg = "sign contract error";
            logger.error(errmsg, err);
            res.status(500).json({ msg: errmsg, status: 500 });
        }
        else {
            res.json({
                contract_id: data.contract_id,
                msg: "contract signed successfully",
                tx_id: result.tx_id
            });
            onSuccess(data);
        }
    });
}

function terminateContractApiHandler(caller, data, req, res) {
    logger.debug('terminateContract');
    var timestamp = "" + Math.floor(new Date().getTime() / 1000);
    var detailType = "terminate";
    solutionChaincodeOps.addContractDetail(caller, data.contract_id, detailType, data.contract_terms, timestamp, function (err, result) {
        if (err != null) {
            var errmsg = "Terminate contract error";
            logger.error(errmsg, err);
            res.status(500).json({ msg: errmsg, status: 500 });
        }
        else {
            res.json({
                contract_id: data.contract_id,
                msg: "contract terminated successfully",
                tx_id: result.tx_id
            });
            onSuccess(data);
        }
    });
}

function givePermissionByContractApiHandler(caller, data, req, res) {
    logger.debug('givePermissionByContract');
    var contract_id = data.contract_id;
    var max_num_download = data.max_num_download;
    var timestamp = "" + Math.floor(new Date().getTime() / 1000);
    var datatype = data.datatype;
    solutionChaincodeOps.givePermissionByContract(caller, contract_id, max_num_download, timestamp, datatype, function (err, result) {
        if (err != null) {
            var errmsg = "givePermissionByContract error";
            logger.error(errmsg, err);
            res.status(500).json({ msg: errmsg, status: 500 });
        }
        else {
            res.json({
                contract_id: data.contract_id,
                msg: "permission to download given successfully",
                tx_id: result.tx_id
            });
            onSuccess(data);
        }
    });
}

async function getContractFromChaincode(caller, contractId) {
    return new Promise(resolve => {
        solutionChaincodeOps.getContract(caller, contractId, function (err, contract) {
            resolve([err, contract]);
        });
    });
}

async function identifyContractPii(contract) {
    const [ownerOrgId, ownerServiceId, requesterOrgId, requesterServiceId] = await Promise.all(
        [contract.owner_org_id && getPii(contract.owner_org_id),
        contract.owner_service_id && getPii(contract.owner_service_id),
        contract.requester_org_id && getPii(contract.requester_org_id),
        contract.requester_service_id && getPii(contract.requester_service_id)]);

    contract.owner_org_id = ownerOrgId;
    contract.owner_service_id = ownerServiceId;
    contract.requester_org_id = requesterOrgId;
    contract.requester_service_id = requesterServiceId;

    return contract;
}

async function getContract(caller, contractId) {
    let [err, contract] = await getContractFromChaincode(caller, contractId);

    if (err) return [err, null];

    if (isDeIdentifierServiceEnabled) {
        contract = await identifyContractPii(contract);
    }

    return [err, contract];
}

function getContractApiHandler(caller, data, req, res) {
    logger.debug('getContract');
    getContract(caller, data.contract_id).then(([err, contract]) => {
        if (err != null) {
            logger.error(err);
            res.json({});
        }
        else {
            res.json(contract);
        }
    }).catch(err => {
        const message = 'getContract error';
        logger.error(message, err);
        res.status(500).json({ msg: message, status: 500 });
    });
}

async function identifyContractsPii(contracts) {
    if (!Array.isArray(contracts)) throw new Error('Cannot identify Contract Records PII. Expect Array datatype.');

    const identifiedContractPromises = contracts.map(identifyContractPii);
    return Promise.all(identifiedContractPromises);
}

async function getOwnerContractsFromChaincode(caller, ownerServiceId, state) {
    return new Promise(resolve => {
        solutionChaincodeOps.getOwnerContracts(caller, ownerServiceId, state, function (err, contracts) {
            resolve([err, contracts]);
        });
    });
}

async function getOwnerContracts(caller, data) {
    let { service_id } = data;

    if (isDeIdentifierServiceEnabled) {
        service_id = await getDeIdentifiedPii(service_id);
    }

    let [err, contracts] = await getOwnerContractsFromChaincode(caller, service_id, data.state);

    if (err) return [err, contracts];

    if (isDeIdentifierServiceEnabled && contracts) {
        contracts = await identifyContractsPii(contracts);
    }

    return [err, contracts];
}

function getOwnerContractsApiHandler(caller, data, req, res) {
    logger.debug('getOwnerContracts');
    getOwnerContracts(caller, data).then(([err, contracts]) => {
        if (err != null) {
            logger.error(err);
            res.json({});
        }
        else {
            res.json(contracts);
        }
    }).catch(err => {
        const message = 'getOwnerContracts error';
        logger.error(message, err);
        res.status(500).json({ msg: message, status: 500 });
    });
}

async function getRequesterContractsInChaincode(caller, requesterServiceId, state) {
    return new Promise(resolve => {
        solutionChaincodeOps.getRequesterContracts(caller, requesterServiceId, state, (err, contracts) => {
            resolve([err, contracts]);
        });
    });
}

async function getRequesterContracts(caller, data) {
    let { service_id } = data;

    if (isDeIdentifierServiceEnabled) {
        service_id = await getDeIdentifiedPii(service_id);
    }

    let [err, contracts] = await getRequesterContractsInChaincode(caller, service_id, data.state);

    if (err) return [err, contracts];

    if (isDeIdentifierServiceEnabled && contracts) {
        contracts = await identifyContractsPii(contracts);
    }

    return [err, contracts];
}

function getRequesterContractsApiHandler(caller, data, req, res) {
    logger.debug('getRequesterContracts');
    getRequesterContracts(caller, data).then(([err, contracts]) => {
        if (err != null) {
            logger.error(err);
            res.json({});
        }
        else {
            res.json(contracts);
        }
    }).catch(err => {
        const message = 'getRequesterContracts error';
        logger.error(message, err);
        res.status(500).json({ msg: message, status: 500 });
    });
}

async function getLogsFromChaincode(caller, contract, patient, service, datatype, contractOrgId, consentOwnerTargetId, startTimestamp, endTimestamp, latestOnly, maxNum) {
    return new Promise(resolve => {
        solutionChaincodeOps.getLogs(caller, contract, patient, service, datatype, contractOrgId, consentOwnerTargetId, startTimestamp,
            endTimestamp, latestOnly, maxNum, (err, logs) => {
                resolve([err, logs]);
            });
    });
}

async function identifyLogsPii(logs) {
  if (!Array.isArray(logs)) throw new Error('Cannot identify PII for Log records. Expect Array type.');

  const logPromises = logs.map(async (log) => {
    if (log.service) {
      log.service = await getPii(log.service);
    }
    if (log.owner) {
      log.owner = await getPii(log.owner);
    }
    if (log.target) {
      log.target = await getPii(log.target);
    }
    if (log.caller) {
      log.caller = await getPii(log.caller);
    }

    return log;
  });

  return Promise.all(logPromises);
}

async function getLogs(caller, data) {
    let { patient, service, contract_org_id, consent_owner_target_id } = data;

    if (isDeIdentifierServiceEnabled) {
        if (patient) {
            patient = await getDeIdentifiedPii(patient);
        }
        if (service) {
            service = await getDeIdentifiedPii(service);
        }
        if (contract_org_id) {
            contract_org_id = await getDeIdentifiedPii(contract_org_id);
        }
        if (consent_owner_target_id) {
            consent_owner_target_id = await getDeIdentifiedPii(consent_owner_target_id);
        }
    }

    let [err, logs] = await getLogsFromChaincode(caller, data.contract, patient, service, data.datatype, contract_org_id, consent_owner_target_id, data.start_timestamp + '', data.end_timestamp + '', data.latest_only + '', data.maxNum + '');

    if (err) return [err, logs];

    if (isDeIdentifierServiceEnabled) {
      logs = await identifyLogsPii(logs);
    }

    return [null, logs];
}

function getLogsApiHandler(caller, data, req, res) {
    logger.debug('getLogs');
    if (!data.start_timestamp) {
        data.start_timestamp = "0";
    }
    if (!data.end_timestamp) {
        data.end_timestamp = "0";
    }
    if (!data.latest_only) {
        data.latest_only = "false"
    }
    if (!data.maxNum) {
        data.maxNum = "1000"
    }
    getLogs(caller, data).then(([err, logs]) => {
        if (err != null) {
            logger.error(err);
            res.json([]);
        }
        else {
            res.json(logs);
        }
    }).catch(err => {
        const message = 'Failed to get Logs';
        logger.error(message, err);
        res.status(500).json({ msg: message, status: 500 });
    });
}

module.exports.process_api = process_api;
function process_api(data, req, res) {
    logger.debug('received api:', data);
    var enrollId = req.headers["enroll-id"];
    var enrollSecret = req.headers["enroll-secret"];
    var org = req.headers["ca-org"];
    var channel = req.headers["channel"];

    var caller = {
        id: enrollId,
        secret: enrollSecret,
        org: org,
        channel: channel
    };

    try {
        //orgs
        if (data.type == 'getOrgs') {
            getOrgsApiHandler(caller, data, req, res);
        } else if (data.type == 'getOrg') {
            getOrgApiHandler(caller, data, req, res);
        } else if (data.type == 'registerOrg') {
            registerOrgApiHandler(caller, data, req, res);
        } else if (data.type == 'updateOrg') {
            updateOrgApiHandler(caller, data, req, res);
        }

        //users
        else if (data.type == 'getUser') {
            getUserApiHandler(caller, data, req, res);
        } else if (data.type == 'registerUser') {
            registerUserApiHandler(caller, data, req, res);
        } else if (data.type == 'registerEnrollAndConsent') {
            registerEnrollAndConsentApiHandler(caller, data, req, res);
        } else if (data.type == 'updateUser') {
            updateUserApiHandler(caller, data, req, res);
        } else if (data.type == 'getUsers') {
            getUsersApiHandler(caller, data, req, res);
        } else if (data.type == 'putUserInOrg') {
            putUserInOrgApiHandler(caller, data, req, res);
        } else if (data.type == 'removeUserFromOrg') {
            removeUserFromOrgApiHandler(caller, data, req, res);
        } else if (data.type == 'addPermissionOrgAdmin') {
            addPermissionOrgAdminApiHandler(caller, data, req, res);
        } else if (data.type == 'deletePermissionOrgAdmin') {
            deletePermissionOrgAdminApiHandler(caller, data, req, res);
        } else if (data.type == 'addPermissionServiceAdmin') {
            addPermissionServiceAdminApiHandler(caller, data, req, res);
        } else if (data.type == 'deletePermissionServiceAdmin') {
            deletePermissionServiceAdminApiHandler(caller, data, req, res);
        }

        // datatype
        else if (data.type == 'registerDatatype') {
            registerDatatypeApiHandler(caller, data, req, res);
        } else if (data.type == 'updateDatatype') {
            updateDatatypeApiHandler(caller, data, req, res);
        } else if (data.type == 'getDatatype') {
            getDatatypeApiHandler(caller, data, req, res);
        } else if (data.type == 'getAllDatatypes') {
            getAllDatatypesApiHandler(caller, data, req, res);

            // service
        } else if (data.type == 'registerService') {
            registerServiceApiHandler(caller, data, req, res);
        } else if (data.type == 'updateService') {
            updateServiceApiHandler(caller, data, req, res);
        } else if (data.type == 'getService') {
            getServiceApiHandler(caller, data, req, res);
        } else if (data.type == 'addDatatypeToService') {
            addDatatypeToServiceApiHandler(caller, data, req, res);
        } else if (data.type == 'removeDatatypeFromService') {
            removeDatatypeFromServiceApiHandler(caller, data, req, res);
        } else if (data.type == 'getServicesOfOrg') {
            getServicesOfOrgApiHandler(caller, data, req, res);
        } else if (data.type == 'addPermissionAuditor') {
            addPermissionAuditorApiHandler(caller, data, req, res);
        } else if (data.type == 'deletePermissionAuditor') {
            deletePermissionAuditorApiHandler(caller, data, req, res);


            // enrollment
        } else if (data.type == 'enrollPatient') {
            enrollPatientApiHandler(caller, data, req, res);
        } else if (data.type == 'unenrollPatient') {
            unenrollPatientApiHandler(caller, data, req, res);
        } else if (data.type == 'getPatientEnrollments') {
            getPatientEnrollmentsApiHandler(caller, data, req, res);
        } else if (data.type == 'getServiceEnrollments') {
            getServiceEnrollmentsApiHandler(caller, data, req, res);

            // consent
        } else if (data.type == 'putConsentPatientData') {
            putConsentPatientDataApiHandler(caller, data, req, res);
        } else if (data.type == 'putMultiConsentPatientData') {
            putMultiConsentPatientDataApiHandler(caller, data, req, res);
        } else if (data.type == 'putConsentOwnerData') {
            putConsentOwnerDataApiHandler(caller, data, req, res);
        } else if (data.type == 'getConsent') {
            getConsentApiHandler(caller, data, req, res);
        } else if (data.type == 'getConsentOwnerData') {
            getConsentOwnerDataApiHandler(caller, data, req, res);
        } else if (data.type == 'getConsents') {
            getConsentsApiHandler(caller, data, req, res);
        } else if (data.type == 'getConsentsWithOwnerID') {
            getConsentsWithOwnerIDApiHandler(caller, data, req, res);
        } else if (data.type == 'getConsentsWithTargetID') {
            getConsentsWithTargetIDApiHandler(caller, data, req, res);
        } else if (data.type == 'validateConsent') {
            validateConsentApiHandler(caller, data, req, res);
        } else if (data.type == 'getConsentRequests') {
            getConsentRequestsApiHandler(caller, data, req, res);

            // user data
        } else if (data.type == 'uploadUserData') {
            uploadUserDataApiHandler(caller, data, req, res);
        } else if (data.type == 'downloadUserData') {
            downloadUserDataApiHandler(caller, data, req, res);
        } else if (data.type == 'downloadUserDataConsentToken') {
            downloadUserDataConsentTokenApiHandler(caller, data, req, res);
        } else if (data.type == 'deleteUserData') {
            deleteUserDataApiHandler(caller, data, req, res);

            // owner data
        } else if (data.type == 'uploadOwnerData') {
            uploadOwnerDataApiHandler(caller, data, req, res);
        } else if (data.type == 'downloadOwnerDataAsOwner') {
            downloadOwnerDataAsOwnerApiHandler(caller, data, req, res);
        } else if (data.type == 'downloadOwnerDataAsRequester') {
            downloadOwnerDataAsRequesterApiHandler(caller, data, req, res);
        } else if (data.type == 'downloadOwnerDataWithConsent') {
            downloadOwnerDataWithConsentApiHandler(caller, data, req, res);
        } else if (data.type == 'downloadOwnerDataConsentToken') {
            downloadOwnerDataConsentTokenApiHandler(caller, data, req, res);

            // contract life cycle
        } else if (data.type == 'createContract') {
            createContractApiHandler(caller, data, req, res);
        } else if (data.type == 'changeContractTerms') {
            changeContractTermsApiHandler(caller, data, req, res);
        } else if (data.type == 'payContract') {
            payContractApiHandler(caller, data, req, res);
        } else if (data.type == 'verifyContractPayment') {
            verifyContractPayment(caller, data, req, res);
        } else if (data.type == 'signContract') {
            signContractApiHandler(caller, data, req, res);
        } else if (data.type == 'terminateContract') {
            terminateContractApiHandler(caller, data, req, res);
        } else if (data.type == 'givePermissionByContract') {
            givePermissionByContractApiHandler(caller, data, req, res);
        } else if (data.type == 'getContract') {
            getContractApiHandler(caller, data, req, res);
        } else if (data.type == 'getOwnerContracts') {
            getOwnerContractsApiHandler(caller, data, req, res);
        } else if (data.type == 'getRequesterContracts') {
            getRequesterContractsApiHandler(caller, data, req, res);

            // logging
        } else if (data.type == 'getLogs') {
            getLogsApiHandler(caller, data, req, res);
        } else {
            var errmsg = "Unknown API end point";
            logger.error(errmsg, data.type);
            res.json({ msg: errmsg + req.path, status: 404 });
            res.status(404);
        }
    }
    catch (err) {
        var errmsg = "process api error";
        logger.error(errmsg, err);

        res.json({ msg: errmsg, status: 500 });
        res.status(500);
    }
};
