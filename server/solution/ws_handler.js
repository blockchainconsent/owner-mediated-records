/*******************************************************************************
 *
 *
 * (c) Copyright Merative US L.P. and others 2020-2022 
 *
 * SPDX-Licence-Identifier: Apache 2.0
 *
 *******************************************************************************/

/* global bag */
/* global $ */
"use strict";

var TAG = 'solution_ws_handler.js';
var log4js = require('log4js');
var logger = log4js.getLogger(TAG);
logger.level = 'DEBUG';

var userManager = require('common-utils/user_manager.js');
var kms = require('common-utils/kms.js');
var ums = require('common-utils/ums.js');
var chaincodeOps = require('common-utils/chaincode_ops.js');
let solutionChaincodeOps = require("./solution_chaincode_ops.js");

var clients = {};
var channels = {};
var caClients = {};
var adminUser = null;
var appAdminUser = null;
var orgName =null;
var wss = {};

module.exports.setup = function (pclients, pchannels, pcaClients, padminUser, pappAdminUser, porgName, pwss) {
	clients = pclients;
	channels = pchannels;
	caClients = pcaClients;
	adminUser = padminUser;
	appAdminUser = pappAdminUser;
	orgName = porgName;
	wss = pwss;
};

// Set to true to allow relevant functions to broadcast on success; enables refresh notifications
var allowBroadcast = false;

/*
module.exports.setup = function (cc, ccID, prs, reg, admin, cert, helper) {
	chaincode = cc;
	peers = prs;
	chaincodeID = ccID;
	registrar = reg;
	appadmin = admin;
	certificate = cert;
	chaincodeOps = helper;
};
*/

//send a message, socket might be closed...
function sendMsg(ws, json, data) {
	if (ws) {
		try {
			ws.send(JSON.stringify(json));
			if (allowBroadcast) {
				if (data) {
					if (wss) {
						if (data.type.indexOf('get') == -1) {
							var broadcast_msg = {msg:"broadcast_refresh", data:data};
							wss.broadcast(broadcast_msg);
						}
					}
				}
			}
		}
		catch (err) {
			logger.error(err);
		}
	}
	else {
		logger.warn("Socket closed. message not sent ");
	}
}

function registerUserMsgHandler(ws, data){
    logger.debug('register user');

    solutionChaincodeOps.getUser(data.caller, data.id, function(err, user) {
        if (err == null && user.id != "") {
            var errmsg = "Existing user with same id found";
            logger.error(errmsg, err);
            sendMsg(ws, { msg: 'message', message: errmsg, 'append':'false'});
        } else {
            logger.debug("Existing User not found: ", data.id);
            registerUser(ws, data, false);
        }
    });
}

function registerUser(ws, data, isUpdate) {
	logger.debug('registerUser');

	// On the UI
    // system admin can register system, auditor, patient
    // org can register org user and patient

    var failIfExist = data.caller.id ? false : true;
    try {
        // 1. register user (CA) & enroll
        userManager.registerUser(data.id, data.secret, data.role, data.ca_org, data.verify_key, data.private_key, data.public_key, data.sym_key, failIfExist).then((attrList) => {
            // 2. register user in chaincode
            var is_group = data.is_group == "true" || data.is_group == true
            var solution_private_data = data.data;

            var userInfo = {
                id: data.id,
                name: data.name,
                role: data.role,
				is_group: is_group,
				org: data.org,
                status: data.status,
                email: data.email,
                secret: data.secret,
                solution_private_data: solution_private_data
            };

            //keys
            for (let i = 0; i < attrList.length; i++) {
                let attr = attrList[i];
                if (attr["name"] === "prvkey") {
                    userInfo["private_key"] = attr["value"];
                } else if (attr["name"] === "pubkey") {
                    userInfo["public_key"] = attr["value"];
                } else if (attr["name"] === "symkey") {
                    userInfo["sym_key"] = attr["value"];
                }
            }

            solutionChaincodeOps.registerUser(data.caller, userInfo, isUpdate, function(err, result) {
                if (err != null) {
                    var errmsg = "User is registered to CA, but failed to update user (CC):" + err.message;
                    logger.error(errmsg);
                    sendMsg(ws, { msg: 'message', message: errmsg, 'append':'false'});
                } else {
                    logger.debug('user (CC) registered successfully:', result);
					logger.info('user registration completed successfully');
					var successMsg = data.id + " registration request has been successfully submitted"
                    sendMsg(ws, { msg: 'success_message', message: successMsg, 'append':'false'});
                }
            });

        }).catch((err) => {
            var errmsg = "Failed to register user (CA):" + err.message;
            logger.error(errmsg, err);
            sendMsg(ws, { msg: 'message', message: errmsg, 'append':'false'});
        });

    } catch (err) {
        var errmsg = "Failed to register user";
        logger.error(errmsg, err);
        sendMsg(ws, { msg: 'message', message: errmsg, 'append':'false'});
    }
}

function changePasswordMsgHandler(ws, data) {
	logger.info('changePassword');

	try {
		Promise.resolve().then(()=>{
			//1. get usertoken
			var p;
			if (data.is_token)  {
				p = new Promise((resolve, reject) => {
					resolve(data.old_password);
				});
			} else {
				p = new Promise((resolve, reject) => {
					userManager.getLoginToken(data.id, data.old_password, function(err, token, user) {
						if (err) {
							logger.error("faild to get token", err);
							reject(err);
						} else  {
							//logger.debug("got token: ", token);
							resolve(token);
						}
					}, 300);
				});
			}
			return p;
		}). then((usertoken) => {
			logger.debug("got token: ", usertoken)
			// 2. change pasword

			ums.changePassword(data.id, data.password, usertoken, function(err, msg) {
				if (err) {
					//failed to change password
					var errmsg = "Failed to change password: "+ err.message;
					logger.error(errmsg, err);
					sendMsg(ws, { msg: 'message', message: errmsg, 'append':'false'});
				}
				else {
					logger.info('Password change completed successfully');
					logger.debug("Sending password change message: ", msg);
					sendMsg(ws, { msg: 'success_message', message: msg, 'append':'false'});
				}
			});
		}, (err) => {
			var errmsg = "Failed to update password for the user " + data.id;
			logger.error(errmsg, err);
			sendMsg(ws, { msg: 'message', message: errmsg, 'append':'false'});
		}).catch((err) => {
			var errmsg = "Failed to update password for the user: "+err.message;
			logger.error(errmsg, err);
			sendMsg(ws, { msg: 'message', message: errmsg, 'append':'false'});
		});

	} catch (err) {
		var errmsg = "Failed to update password for the user " + data.id;
		logger.error(errmsg, err);
		sendMsg(ws, { msg: 'message', message: errmsg, 'append':'false'});
	}

}

function chainstatsMsgHandler(ws, data){
	logger.info('chainstat');
	chaincodeOps.chainstat(data.user, function (err, stats) {
		if (err != null) {
			var errmsg = "Failed to get chainstat";
			logger.error(errmsg, err);
			sendMsg(ws, { msg: 'message', message: errmsg, 'append':'false'});
		}
		else {
			logger.debug('chainstat successful: ', stats);
			sendMsg(ws,{ msg: 'chainstats', e: null, chainstats: stats});
		}
	});
}

function registerOrgMsgHandler(ws, data) {
	logger.info('registerOrgMsgHandler');

	var failIfExist = data.caller.id ? false : true;
	try {
		// 1. register user (CA) & enroll
		userManager.registerUser(data.id, data.secret, data.role, data.ca_org, data.verify_key, data.private_key, data.public_key, data.sym_key, failIfExist).then((attrList) => {
			// 2. register user in chaincode
			var userInfo = {
				id: data.id,
				name: data.name,
				role: data.role,
				is_group: true,
				status: "active",
				email: data.email,
				secret: data.secret,
				solution_public_data: "{}",
				solution_private_data: data.data
			};

			//keys
			for (let attr of attrList) {
				if (attr["name"] === "prvkey") {
					userInfo["private_key"] = attr["value"];
				} else if (attr["name"] === "pubkey") {
					userInfo["public_key"] = attr["value"];
				} else if (attr["name"] === "symkey") {
					userInfo["sym_key"] = attr["value"];
				}
			}

			solutionChaincodeOps.registerOrg(data.caller, userInfo, function(err, result) {
				if (err != null) {
					var errmsg = "Org is registered to CA, but failed to update org (CC):" + err.message;
					logger.error(errmsg);
					sendMsg(ws, { msg: 'message', message: errmsg, 'append':'false'});
				} else {
					var successMsg = data.id + " registration request has been successfully submitted"
					logger.debug('org (CC) registered successfully:', result);
					logger.info('org registration completed successfully');
					sendMsg(ws, { msg: 'success_message', message: successMsg});
				}
			});


		}).catch((err) => {
			var errmsg = "Failed to register org (CA):" + err.message;
			logger.error(errmsg, err);
			sendMsg(ws, { msg: 'message', message: errmsg, 'append':'false'});
		});

	} catch (err) {
		var errmsg = "Failed to register org";
		logger.error(errmsg, err);
		sendMsg(ws, { msg: 'message', message: errmsg, 'append':'false'});
	}
}

function updateUserMsgHandler(ws, data) {
    logger.info('updateUser');
    solutionChaincodeOps.getUser(data.caller, data.id, function (err, user) {
        if (err != null) {
            var errmsg = "User not found";
            logger.error(errmsg, err);
            sendMsg(ws, { msg: 'message', message: errmsg, 'append':'false'});
        }
        else if (user.role != data.role){
            var errmsg =  "User's role cannot be changed";
            logger.error(errmsg, err);
            sendMsg(ws, { msg: 'message', message: errmsg, 'append':'false'});
		}
        else {
            logger.debug("Existing User not found: ", data.id);
            registerUser(ws, data, true);
        }
    });
}

function updateOrgMsgHandler(ws, data) {
    logger.debug('update org');

    solutionChaincodeOps.getOrg(data.caller, data.id, function(err, user) {
        if (err != null || !user) {
            var errmsg = "Org not found";
            logger.error(errmsg, err);
            sendMsg(ws, { msg: 'message', message: errmsg, 'append':'false'});
        } else if (!user.secret) {
            var errmsg = "Unauthorized to update the org";
            logger.error(errmsg, err);
            sendMsg(ws, { msg: 'message', message: errmsg, 'append':'false'});
        } else if (data.secret && user.secret && user.secret !== data.secret) {
            var errmsg = "Org admin's secret cannot be changed";
            logger.error(errmsg, err);
            sendMsg(ws, { msg: 'message', message: errmsg, 'append':'false'});
        }else {
            logger.debug("Existing Org found: ", data.id);
            data.secret = user.secret;
            registerOrgMsgHandler(ws, data);
        }
    });
}

function registerServiceMsgHandler(ws, data) {
	logger.info('registerService');

	// if no data ID is passed in, then we assume it's a new service
	if (!data.id || data.id.toLowerCase() === "new") {
		data.id = "service-"+kms.getUuid();
	}

    var failIfExist = data.caller.id ? false : true;
    try {
        // 1. register user (CA) & enroll
        userManager.registerUser(data.id, data.secret, data.role, data.ca_org, data.verify_key, data.private_key, data.public_key, data.sym_key, failIfExist).then((attrList) => {
            // 2. register user in chaincode
            var is_group = data.is_group == "true" || data.is_group == true
            var serviceInfo = {
                service_id: data.id,
                service_name: data.name,
                role: data.role,
                is_group: is_group,
                status: data.status,
                email: data.email,
                secret: data.secret,
                solution_public_data: "{}",
                solution_private_data: data.solution_private_data,
                org_id: data.org_id,
                datatypes: data.datatypes,
                summary: data.summary,
                terms: data.terms,
                payment_required: data.payment_required,
                create_date: Math.floor(new Date().getTime() / 1000)
            };

            //keys
            for (let i = 0; i < attrList.length; i++) {
                let attr = attrList[i];
                if (attr["name"] === "prvkey") {
                    serviceInfo["private_key"] = attr["value"];
                } else if (attr["name"] === "pubkey") {
                    serviceInfo["public_key"] = attr["value"];
                } else if (attr["name"] === "symkey") {
                    serviceInfo["sym_key"] = attr["value"];
                }
            }

            solutionChaincodeOps.registerService(data.caller, serviceInfo, function(err, result) {
                if (err != null) {
                    var errmsg = "Service is registered to CA, but failed to update service (CC):" + err.message;
                    logger.error(errmsg);
                    sendMsg(ws, { msg: 'message', message: errmsg, 'append':'false'});
                } else {
                    logger.info('service registration completed successfully');
                    var successMsg = data.id + " registration request has been successfully submitted";
					sendMsg(ws, { msg: 'success_message', message: successMsg});
                }
            });

        }).catch((err) => {
            var errmsg = "Failed to register service (CA):" + err.message;
            logger.error(errmsg, err);
            sendMsg(ws, { msg: 'message', message: errmsg, 'append':'false'});
        });

    } catch (err) {
        var errmsg = "Failed to register service";
        logger.error(errmsg, err);
        sendMsg(ws, { msg: 'message', message: errmsg, 'append':'false'});
    }
}

function updateServiceMsgHandler(ws, data) {
	logger.info('updateService');

	solutionChaincodeOps.getOrg( data.caller, data.id , function (err, service) {
        if (err != null || (err == null && service.service_id == "")) {
            var errmsg =  "Service not found";
            logger.error(errmsg, err);
            sendMsg(ws, { msg: 'message', message: errmsg, 'append':'false'});
        } else if (!service.secret) {
            var errmsg = "Unauthorized to update the service";
            logger.error(errmsg, err);
            sendMsg(ws, { msg: 'message', message: errmsg, 'append':'false'});
        } else if (data.secret && service.secret && service.secret != data.secret) {
            var errmsg = "Service admin's secret cannot be changed";
            logger.error(errmsg, err);
            sendMsg(ws, { msg: 'message', message: errmsg, 'append':'false'});
        } else {
            logger.debug("Existing service found: ", data.id);
            data.secret = service.secret;
            updateService(ws, data);
        }
    });
}

function updateService(ws, data) {
    logger.debug('updateService');

    logger.debug('registerService');
    var failIfExist = data.caller.id ? false : true;

    try {
        // 1. register user (CA) & enroll
        userManager.registerUser(data.id, data.secret, data.role, data.ca_org, data.verify_key, data.private_key, data.public_key, data.sym_key, failIfExist).then((attrList) => {
            // 2. register user in chaincode
            var is_group = data.is_group == "true" || data.is_group == true
            var serviceInfo = {
                service_id: data.id,
                service_name: data.name,
                role: data.role,
                is_group: is_group,
                status: data.status,
                email: data.email,
                secret: data.secret,
                solution_public_data: "{}",
                solution_private_data: data.solution_private_data,
                org_id: data.org_id,
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

            solutionChaincodeOps.updateService(data.caller, serviceInfo, function(err, result) {
                if (err != null) {
                    var errmsg = "Service is registered to CA, but failed to update service (CC):" + err.message;
                    logger.error(errmsg);
                    sendMsg(ws, { msg: 'message', message: errmsg, 'append':'false'});
                } else {
                    logger.debug('service (CC) updated successfully:', result);
                    logger.info('service update completed successfully');
                    var successMsg = data.id + " update request has been successfully submitted";
					sendMsg(ws, { msg: 'success_message', message: successMsg});
                }
            });

        }).catch((err) => {
            var errmsg = "Failed to update service (CA):" + err.message;
            logger.error(errmsg, err);
            sendMsg(ws, { msg: 'message', message: errmsg, 'append':'false'});
        });

    } catch (err) {
        var errmsg = "Failed to update service";
        logger.error(errmsg, err);
        sendMsg(ws, { msg: 'message', message: errmsg, 'append':'false'});
    }
}

function getServiceMsgHandler(ws, data) {
	logger.info('getService');
	solutionChaincodeOps.getService(data.caller, data.id, function (err, service) {
		if (err != null) {
			logger.error(err);
			sendMsg(ws, { msg: 'message', message: err, 'append':'false'});
		}
		else {
			sendMsg(ws, { msg: 'getService', data: service});
		}
	});
}

function getAllServicesOfOrgMsgHandler(ws, data) {
	logger.info('getServices');
	solutionChaincodeOps.getServicesOfOrg(data.caller, data.org, function (err, services) {
		if (err != null) {
			logger.error(err);
			sendMsg(ws, { msg: 'message', message: err, 'append':'false'});
		}
		else {
			sendMsg(ws, { msg: 'allServices', data: services, org: data.org });
		}
	});
}

function registerDataTypeMsgHandler(ws, data) {
	logger.info('registerDataType');

	solutionChaincodeOps.getDatatype(data.caller, data.id, function (err, datatype) {
        // we have to check for empty datatype ID here because chaincode
        // returns empty datatype if does not exist
        if (err != null || (err == null && datatype.datatype_id != "")) {
            var errmsg =  "Existing datatype with same id found";
            logger.error(errmsg, err);
            sendMsg(ws, { msg: 'registerDatatype error', message: errmsg, 'append':'false'});
        }
        else {
            logger.debug("Existing datatype not found, proceed to register: ", data.id);

            kms.getSymKeyAes(function(err, datatypeSymkey) {
                if (err) {
                    var errmsg = "error creating datatype sym key";
                    logger.error(errmsg, err);
                    sendMsg(ws, { msg: 'registerDatatype error', message: errmsg, 'append':'false'});
                } else {
                    var datatype = {
                        datatype_id: data.id,
                        description: data.description
                    }

                    solutionChaincodeOps.registerDatatype(data.caller, datatype, datatypeSymkey.keyBase64, data.parent_datatype_id, function (err, result) {
                        if (err != null) {
                            var errmsg = "registerDatatype error";
                            logger.error(errmsg, err);
                            sendMsg(ws, { msg: 'registerDatatype error', message: errmsg, 'append':'false'});
                        }
                        var successMsg = data.id + " registration request has been successfully submitted"
						logger.debug('datatype (CC) registered successfully:', result);
						logger.info('datatype registration completed successfully');
						sendMsg(ws, { msg: 'success_message', message: successMsg});
                    });
                }

            }, data.datatypeSymkey);
        }
    });
}

function updateDataTypeMsgHandler(ws, data) {
	logger.info('updateDataType');
	var connect_string = data.connect_string? data.connect_string : "";
	var dataType = {
		id: data.id,
		name: data.name,
		org: data.org,
		owner_type: data.owner_type,
		status: data.status,
		data: {
		db_type: data.db_type,
		connect_string: connect_string}
	};

	solutionChaincodeOps.updateDataType(data.user, dataType, function (err, result) {
		if (err != null) {
			var errmsg = "updateDataType error";
			logger.error(errmsg, err);
			sendMsg(ws, { msg: 'message', message: errmsg, 'append':'false'});
		}
		else {
			logger.debug('updateDataType:', result);
			var successMsg = data.id + " update request has been successfully submitted";
			sendMsg(ws, { msg: 'success_message', message: successMsg});
		}
	});
}

function createContractMsgHandler(ws, data) {
	logger.info('createContract');
	var contractUUID = "contract-"+kms.getUuid();
	var timestamp = Math.floor(new Date().getTime() / 1000);
	var contract = {
		contract_id: contractUUID,
		owner_org_id: data.owner_org_id,
		owner_service_id: data.owner_service_id,
		requester_org_id: data.requester_org_id,
		requester_service_id: data.requester_service_id,
		contract_terms: data.terms ? data.terms : {},
		state: "new",
		create_date: timestamp,
		update_date: timestamp,
		contract_detail: [],
		payment_required: "no",
		payment_verified: "no",
		max_num_download: 0,
		num_download: 0
	};

	kms.getSymKeyAes( function(err, symkey) {
		if (err) {
			var errmsg = "error creating contract sym key";
			logger.error(errmsg, err);
			sendMsg(ws, { msg: 'message', message: errmsg, 'append':'false'});
		} else {
			var contract_sym_key = symkey.keyBase64;
			solutionChaincodeOps.createContract(data.caller, contract, contract_sym_key, function (err, result) {
				if (err != null) {
					var errmsg = "failed to create contract : " + contractUUID;
					logger.error(errmsg, err);
					sendMsg(ws, { msg: 'message', message: errmsg, 'append':'false'});
				} else {
					logger.debug('createContract:', result);
					sendMsg(ws, { msg: 'success_createContract_message', message: contractUUID + " has been successfully created.", 'append':'false' }, data);
				}
			});
		}
	}, data.symkey);
}

function changeContractTermsMsgHandler(ws, data) {
	logger.info('changeContractTerms');
	var timestamp = "" + Math.floor(new Date().getTime() / 1000);
	var detailType = "terms";
	solutionChaincodeOps.addContractDetail(data.caller, data.contract_id, detailType, data.contract_terms, timestamp, function (err, result) {
		if (err != null) {
			var errmsg = "add contract terms error";
			logger.error(errmsg, err);
			sendMsg(ws, { msg: 'add contract terms error: ', message: errmsg, 'append':'false'});
		}
		else {
			sendMsg(ws, { msg: 'success_contract_message', contract_id: data.contract_id, message: "Successfully added terms to " + data.contract_id, 'append':'false' }, data);
		}
	});
}

function signContractMsgHandler(ws, data) {
	logger.info('signContract');
	var timestamp = "" + Math.floor(new Date().getTime() / 1000);
	var detailType = "sign";
	solutionChaincodeOps.addContractDetail(data.caller, data.contract_id, detailType, data.contract_terms, timestamp, function (err, result) {
		if (err != null) {
			var errmsg = "sign contract error";
			logger.error(errmsg, err);
			sendMsg(ws, { msg: 'sign contract error: ', message: errmsg, 'append':'false'});
		}
		else {
			sendMsg(ws, { msg: 'success_contract_message', contract_id: data.contract_id, message: data.contract_id + " has been successfully signed.", 'append':'false' }, data);
		}
	});
}

function payContractMsgHandler(ws, data) {
	logger.info('payContract');
	var timestamp = "" + Math.floor(new Date().getTime() / 1000);
	var detailType = "payment";
	solutionChaincodeOps.addContractDetail(data.caller, data.contract_id, detailType, data.contract_terms, timestamp, function (err, result) {
		if (err != null) {
			var errmsg = "failed to pay contract : " + data.contract_id;
			logger.error(errmsg, err);
			sendMsg(ws, { msg: 'message', message: errmsg, 'append':'false'});
		}
		else {
			logger.debug('payContract:', result);
			sendMsg(ws, { msg: 'success_contract_message', contract_id: data.contract_id, message: data.contract_id + " has been successfully paid.", 'append':'false' }, data);
		}
	});
}

function terminateContractMsgHandler(ws, data) {
	logger.info('terminateContract');
	var timestamp = "" + Math.floor(new Date().getTime() / 1000);
	var detailType = "terminate";
	solutionChaincodeOps.addContractDetail(data.caller, data.contract_id, detailType, data.contract_terms, timestamp, function (err, result) {
		if (err != null) {
			var errmsg = "failed to terminate contract : " + data.contract_id;
			logger.error(errmsg, err);
			sendMsg(ws, { msg: 'message', message: errmsg, 'append':'false'});
		}
		else {
			logger.debug('terminateContract:', result);
			sendMsg(ws, { msg: 'success_contract_message', contract_id: data.contract_id, message: data.contract_id + " has been successfully terminated.", 'append':'false' }, data);
		}
	});
}

function verifyContractPaymentMsgHandler(ws, data) {
	logger.info('verifyContractPayment');
	var timestamp = "" + Math.floor(new Date().getTime() / 1000);
	var detailType = "verify";
	solutionChaincodeOps.addContractDetail(data.caller, data.contract_id, detailType, data.contract_terms, timestamp, function (err, result) {
		if (err != null) {
			var errmsg = "failed to verify payment : " + data.contract_id;
			logger.error(errmsg, err);
			sendMsg(ws, { msg: 'message', message: errmsg, 'append':'false'});
		}
		else {
			logger.debug('verify payment:', result);
			sendMsg(ws, { msg: 'success_contract_message', contract_id: data.contract_id, message: "Successfully verified payment for " + data.contract_id, 'append':'false' }, data);
		}
	});
}

function givePermissionByContractMsgHandler(ws, data) {
	logger.info('givePermissionByContract');
	var contract_id = data.contract_id;
	var max_num_download = data.max_num_download;
	var timestamp = "" + Math.floor(new Date().getTime() / 1000);
	var datatype = data.datatype;
	solutionChaincodeOps.givePermissionByContract(data.caller, contract_id, "" + max_num_download, timestamp, datatype, function (err, result) {
		if (err != null) {
			var errmsg = "failed to give permission : " + data.contract_id;
			logger.error(errmsg, err);
			sendMsg(ws, { msg: 'message', message: errmsg, 'append':'false'});
		}
		else {
			logger.debug('verify payment:', result);
			sendMsg(ws, { msg: 'success_contract_message', contract_id: data.contract_id, message: "Successfully given permission for " + data.contract_id, 'append':'false' }, data);
		}
	});
}

function addMultipleConsentsMsgHandler(ws, data) {
	logger.info('add multiple consents');

	var promises = [];
	var success = [];
	var failure = [];

	for (var i = 0; i < data.consents.length; i++) {
		var p = new Promise(function(resolve, reject) {
			kms.getSymKeyAes( function(err, consentSymKey) {
				if (err) {
					var errmsg = "Error creating consent sym key";
					logger.error(errmsg, err);
					failure.push(request.datatype);
					resolve(err);
				} else {
					var consent = data.consents[i];
					var request = {
						owner: consent.patient_id,
						target: consent.target_id,
						datatype: consent.datatype_id,
						service: consent.service_id,
						option: consent.option,
						expiration: 0,
						timestamp: Math.floor(new Date().getTime() / 1000)
					}
					solutionChaincodeOps.putConsentPatientData(data.caller, request, consentSymKey.keyBase64, function (err, result) {
						if (err != null) {
							var errmsg = "Error adding consent for " + request.service;
							logger.error(errmsg, err);
							failure.push(request.datatype);
							resolve(err);
						}
						else {
							logger.debug('addConsent:', result);
							success.push(request.datatype);
							resolve(result);
						}
					});
				}
			}, data.consent_sym_key);
		});
		promises.push(p);
	}

	Promise.all(promises).then(result => {
		if (failure.length > 0 && success.length > 0) {
			sendMsg(ws, { msg: 'addConsent_response_message', message: "Successfully submitted consent change request for " + success + " and failed for " + failure, 'append':'false'  });
		} else if (failure.length <= 0 && success.length > 0) {
			sendMsg(ws, { msg: 'addConsent_response_message', message: "Successfully submitted consent change request for " + success, 'append':'false'  });
		} else if (failure.length > 0 && success.length <= 0) {
			sendMsg(ws, { msg: 'addConsent_response_message', message: "Failed submitted consent change request for " + failure, 'append':'false'  });
		} else {
			var errmsg = "Add consents error";
			logger.error(errmsg, err);
			sendMsg(ws, { msg: 'message', message: errmsg, 'append':'false' });
		}
	});
}

module.exports.process_msg = process_msg;
function process_msg(ws, data) {

	logger.debug('ws_handler received message:', data);

	try {
		// Process the message
		if (data.type == 'chainstats') {
			chainstatsMsgHandler(ws, data);
		}
		else if (data.type == 'registerUser') {
			registerUserMsgHandler(ws, data);
		}
		else if (data.type == 'changePassword') {
			changePasswordMsgHandler(ws, data);
		}
		else if (data.type == 'registerOrg') {
			registerOrgMsgHandler(ws, data);
		}
        else if (data.type == 'updateUser') {
			updateUserMsgHandler(ws, data);
        }
		else if (data.type == 'updateOrg') {
			updateOrgMsgHandler(ws, data);
		}
		else if (data.type == 'registerService') {
			registerServiceMsgHandler(ws, data);
		}
		else if (data.type == 'updateService') {
			updateServiceMsgHandler(ws, data);
		}
		else if (data.type == 'registerDataType') {
			registerDataTypeMsgHandler(ws, data);
		}
		else if (data.type == 'updateDataType') {
			updateDataTypeMsgHandler(ws, data);
		}
		else if (data.type == 'createContract') {
			createContractMsgHandler(ws, data);
		}
		else if (data.type == 'requestContract') {
			requestContractMsgHandler(ws, data);
		}
		else if (data.type == 'signContract') {
			signContractMsgHandler(ws, data);
		}
		else if (data.type == 'payContract') {
			payContractMsgHandler(ws, data);
		}
		else if (data.type == 'terminateContract') {
			terminateContractMsgHandler(ws, data);
		}
		else if (data.type == 'changeContractTerms') {
			changeContractTermsMsgHandler(ws, data);
		}
		else if (data.type == 'verifyContractPayment') {
			verifyContractPaymentMsgHandler(ws, data);
		}
		else if (data.type == 'givePermissionByContract') {
			givePermissionByContractMsgHandler(ws, data);
		}
		else if (data.type == 'addMultipleConsents') {
			addMultipleConsentsMsgHandler(ws, data);
		}
		else if (data.type == 'getServiceDetails') {
			getServiceDetailsMsgHandler(ws, data);
		}
		else if (data.type == 'getService') {
			getServiceMsgHandler(ws, data);
		}
		else if (data.type == 'getAllServicesOfOrg') {
			getAllServicesOfOrgMsgHandler(ws, data);
		}
		else if (data.type == 'queryTest') {
			queryTestMsgHandler(ws, data);
		}
		else if (data.type == 'invokeTest') {
			invokeTestMsgHandler(ws, data);
		}
		else if (data.type == 'keyTest') {
			keyTestMsgHandler(ws, data);
		}
		else {
			logger.debug("error: unknown message type ", data.type);
		}


	} catch (err) {
		var errmsg = "ws handler message processing error";
		logger.error(errmsg, err);
		/*
		Object.getOwnPropertyNames(err).forEach(function (key) {
			logger.debug("===>", key, err[key]);
			}, err);
		*/
		sendMsg(ws, {type: "error", error: errmsg});
	}


};



