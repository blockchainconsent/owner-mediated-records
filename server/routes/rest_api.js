/*******************************************************************************
 *
 *
 * (c) Copyright Merative US L.P. and others 2020-2022 
 *
 * SPDX-Licence-Identifier: Apache 2.0
 *
 *******************************************************************************/

var express = require('express');
var router = express.Router();
var crypto = require('crypto');

// load common util node modules
var userManager = require('common-utils/user_manager');
var chainHelper = require('common-utils/chain_helper');
var ums = require('common-utils/ums');
var kms = require('common-utils/kms');

// make sure request_handler is pointing to the solution specific request_handler
var req_handler = require('common-utils/request_handler.js');

// solution api handler
var solution_req_handler = require('../solution/solution_api_handler.js');

var TAG = "rest_api.js";
var log4js = require('log4js');
var logger = log4js.getLogger(TAG);
logger.level = 'DEBUG';

// ============================================================================================================================
// Home
// ============================================================================================================================

// 1. if path is /login skip authentication,
// 2. check token
// 3. if authorization (basic auth) exist in headers, remove token in session
// 4. check token in session
// 5. check headers
// 6. apilogin (require same headers as login) - basic auth first and then headers
router.use(function (req, res, next) {
    // do logging
    logger.info('Rest API ' + req.path);
    logger.info('URL ' + req.originalUrl);
    logger.debug(req.headers);
    logger.debug(req.session);

    //clear header
    req.headers["enroll-id"] = null;
    req.headers["enroll-secret"] = null;
    req.headers["ca-org"] = null;
    req.headers["channel"] = null;

    // check token
    var token = req.headers["token"];
    var auth = req.headers.authorization;
    if (auth) {
        logger.info("basic auth, remove token from session if exists");
        req.session["token"] = null;
    }
    if (token) {
        logger.info("token:", token);
        // UNCOMMENT THIS LINE WHEN SWITCHING TO REACT
        // req.session["token"] = null;
    } else {
        // UNCOMMENT THIS BLOCK WHEN SWITCHING TO REACT
        // if basic auth, remove token from session
        //var auth = req.headers.authorization;

        //check token in session
        token = req.session["token"];
        if (token) {
            logger.info("got token from session:", token);
        }
    }

    if (req.path == "/sign") {
        //sign doesn't need to verify signature
        next();
    } else if (req.path == "/login") {
        //login doesn't need to authenticate, but
        //still need to verify signature
        var login_info = getLoginInfo(req);
        var enrollId = login_info["id"];
        var enrollSecret = login_info["password"];
        var caOrg = login_info["org"];
        verifySignatureAndNext(enrollId, enrollSecret, caOrg, req, res, next);
    } else {
        //authenticate token
        userManager.validateLoginToken(token, function (err, tokenList) {
            if (err) {
                if (token) {
                    logger.warn("Login token failed");
                    logger.debug(err)
                }

                //authenticate by headers (ums specific)
                userManager.validateUserByHeaders(req.headers, function (err, tokenList) {
                    if (err) {
                        logger.warn("Authentication by Headers failed");
                        logger.debug(err);

                        //authenticate by apilogin (same as /login parameters)
                        apiLogin(req, res, next);

                    } else {
                        logger.info("Authentication by Headers validated");
                        let enrollId = tokenList && tokenList[1] ? tokenList[1] : "";
                        let enrollSecret = tokenList && tokenList[2] ? tokenList[2] : "";
                        let caOrg = tokenList && tokenList[3] ? tokenList[3] : "";
                        let channel = tokenList && tokenList[4] ? tokenList[4] : "";
                        req.headers["enroll-id"] = enrollId;
                        req.headers["enroll-secret"] = enrollSecret;
                        req.headers["ca-org"] = caOrg;
                        req.headers["channel"] = channel;
                        logger.debug("enroll-id:", enrollId);
                        //logger.debug("enroll-secret:", enrollSecret);
                        logger.debug("ca-org:", caOrg);
                        logger.debug("channel:", channel);
                        //next();
                        verifySignatureAndNext(enrollId, enrollSecret, caOrg, req, res, next);
                    }
                });

            } else {
                logger.info("login token validated");
                let enrollId = tokenList && tokenList[1] ? tokenList[1] : "";
                let enrollSecret = tokenList && tokenList[2] ? tokenList[2] : "";
                let caOrg = tokenList && tokenList[3] ? tokenList[3] : "";
                let channel = tokenList && tokenList[4] ? tokenList[4] : "";
                req.headers["enroll-id"] = enrollId;
                req.headers["enroll-secret"] = enrollSecret;
                req.headers["ca-org"] = caOrg;
                req.headers["channel"] = channel;
                logger.debug("enroll-id:", enrollId);
                //logger.debug("enroll-secret:", enrollSecret);
                logger.debug("ca-org:", caOrg);
                logger.debug("channel:", channel);
                //next();
                verifySignatureAndNext(enrollId, enrollSecret, caOrg, req, res, next);
            }
        });
    }
});

//this function checks signature and checks if payload is encrypted
function verifySignatureAndNext(username, secret, org, req, res, next) {
    logger.debug(req_handler.solutionConfig);
    let verify_signature = req_handler.solutionConfig["verify_user_signature"];
    let disable_verify_signature = req_handler.solutionConfig["disable_verify_user_signature_for_localhost"];
    let disable_verify_signature_no_key = req_handler.solutionConfig["skip_verify_user_signature_if_no_key_found"];
    //don't need to signature if request is from local host
    if (verify_signature && disable_verify_signature) {
        //logger.debug("request.headers.host:",req.headers.host);
        //logger.debug("request.get(host):",req.get('host'));
        let host = req.get('host');
        if (host.startsWith("localhost")) {
            verify_signature = false;
            logger.debug("verify signature disabled for localhost:", host);
        }
    }

    //decrypt payload if payload is encrypted
    if (req.body && req.body._key_ && req.body._data_) {
        logger.debug("payload is encrypted");
        logger.debug(req.body);
        //decrypt key -- it should be encrypted with the public key
        let prvkey = req_handler.solutionConfig["app_admin"]["private_key"];
        let keyHex = Buffer.from(req.body._key_, "base64").toString("hex");
        let symKeyB64 = kms.decryptRSA(keyHex, prvkey);
        //logger.debug("B64:",symKeyB64);
        let symKey = Buffer.from(symKeyB64, "base64");
        let iv = null;
        if (req.body._iv_) {
            iv = Buffer.from(req.body._iv_, "base64");
        }
        decBody = kms.decryptAesSymKey(symKey, req.body._data_, "base64", "utf8", iv);
        logger.debug(decBody);
        try {
            let bodyJson = JSON.parse(decBody);
            req.body = bodyJson;
        } catch (err) {
            logger.error("Decrypting payload failed:", err);
        }
    }

    //verify signature
    if (!verify_signature) {
        next();
    } else {
        logger.info("verify signature");
        var signature = req.headers["signature"];

        if (!signature) {
            var errmsg = "signature header is missing";
            logger.error(errmsg);
            res.status(401).json({ msg: "Invalid signature:" + errmsg, status: 401 });
        } else {

            let signatureList = signature.split(":");
            let signAlgorithm = "";
            if (signatureList.length == 2) {
                signAlgorithm = signatureList[1];
            }

            const url = req.originalUrl;
            var message = username + url;
            var method = req.method;
            let payload = "";
            if (method == "POST" || method == "PUT") {
                payload = req.body;
                if (typeof payload === 'string' || payload instanceof String) {
                    message = message + payload;
                } else {
                    payload = JSON.stringify(payload);
                    message = message + payload;
                }
            }

            if (signAlgorithm == "solution-ui") {
                logger.debug("From UI: signature verification with ui-verify_key");
                // UI signature verification
                let verifyKey = req_handler.solutionConfig["ui_verify_key"];
                //verify signature
                let verified = verifySignature(verifyKey, signature, message);
                if (!verified) {
                    var errmsg = "UI signature verification failed";
                    logger.error(errmsg);
                    res.status(401).json({ msg: "Invalid signature:" + errmsg, status: 401 });
                } else {
                    logger.info("UI Signature verification succeeded");
                    next();
                }


            } else {
                var myclient = null;
                chainHelper.getClientForOrg(org, username, secret).then((client) => {
                    myclient = client;
                    var caClient = myclient.getCertificateAuthority();
                    let attrs = [{ name: "id" }, { name: "verifykey" }];
                    return chainHelper.getUserAttributes(username, secret, attrs, caClient);
                }).then((amap) => {
                    if (amap && amap["verifykey"] && amap["verifykey"] != userManager.NO_VERIFY_KEY) {
                        let verifyKey = amap["verifykey"];
                        //verify signature
                        let verified = verifySignature(verifyKey, signature, message);
                        if (!verified) {
                            var errmsg = "signature verification failed";
                            logger.error(errmsg);
                            res.status(401).json({ msg: "Invalid signature:" + errmsg, status: 401 });
                        } else {
                            logger.info("Signature verification succeeded");
                            next();
                        }

                    } else {
                        var errmsg = "verify key not found on the CA server";
                        if (disable_verify_signature_no_key) {
                            logger.warn(errmsg + ":", "skip verify signature");
                            next();
                        } else {
                            logger.error(errmsg);
                            res.status(401).json({ msg: "Invalid signature:" + errmsg, status: 401 });
                        }
                    }
                }).catch((err) => {
                    var errmsg = "Unable to verify signature";
                    logger.error(errmsg, err);
                    res.status(401).json({ msg: "Invalid signature:" + errmsg, status: 401 });
                });
            }

        }

    }
}

function verifySignature(verifyKey, signature, message) {
    if (!verifyKey.startsWith("-----BEGIN")) {
        verifyKey = "-----BEGIN PUBLIC KEY-----\n" + verifyKey + "\n-----END PUBLIC KEY-----\n";
    }
    //verify signature
    signatureList = signature.split(":");
    let algorithm = "sha256";
    if (signatureList.length == 2) {
        if (signatureList[1] != "solution-ui") {
            algorithm = signatureList[1];
        }
        signature = signatureList[0];
    }

    //logger.debug("verifyKey:", verifyKey);
    //logger.debug("message:", message);
    //logger.debug("signature", signature);
    const verifier = crypto.createVerify(algorithm);
    verifier.update(message);
    verifier.end();
    let verified = verifier.verify(verifyKey, signature, "base64");
    return verified;
}

function apiLogin(req, res, next) {
    logger.info("attempting api login");
    var login_info = getLoginInfo(req);
    logger.info(login_info)
    var id = login_info["id"];
    var password = login_info["password"];
    var channel = login_info["channel"];
    var org = login_info["org"];

    if (!id || !password || !org || !channel) {
        logger.error("Unauthorized");
        res.status(401).json({ msg: "Unauthorized.", status: 401 });
    } else {
        let user_data = {
            org: org,
            channel: channel
        };
        ums.validateLoginUser(id, password, user_data, function (err, user) {
            if (err) {
                logger.error("Unauthorized:", err);
                res.status(401).json({ msg: "Unauthorized:" + err, status: 401 });
            } else {
                logger.info("User successfully login via ApiLogin")
                req.headers["enroll-id"] = user.enrollId;
                req.headers["enroll-secret"] = user.enrollSecret;
                req.headers["ca-org"] = user.caOrg;
                req.headers["channel"] = user.channel;
                logger.debug("enroll-id:", user.enrollId);
                //logger.debug("enroll-secret:", user.enrollSecret);
                logger.debug("ca-org:", user.caOrg);
                logger.debug("channel:", user.channel);
                //next();
                verifySignatureAndNext(user.enrollId, user.enrollSecret, user.caOrg, req, res, next);
            }
        });
    }
}

function getLoginInfo(req) {
    var b64auth = (req.headers.authorization || '').split(' ')[1] || ':';
    var basicAuth = new Buffer(b64auth, 'base64').toString().split(':');
    var ba_id = null;
    var ba_org = null;
    var ba_channel = null;
    var ba_pass = null;
    if (basicAuth[0]) {
        let val1 = basicAuth[0].split('/');
        if (val1.length >= 1) {
            ba_id = val1[0];
        }
        if (val1.length >= 2) {
            ba_org = val1[1];
        }
        if (val1.length >= 3) {
            ba_channel = val1[2];
        }
    }
    if (basicAuth[1]) {
        ba_pass = basicAuth[1];
    }

    //use user-id and password header if basicauth header does not exist
    var id = ba_id ? ba_id : req.headers["user-id"];
    var password = ba_pass ? ba_pass : req.headers["password"];
    var channel = ba_channel ? ba_channel : req.headers["login-channel"];
    var org = ba_org ? ba_org : req.headers["login-org"];

    var user_info = {
        id: id,
        password: password,
        channel: channel,
        org: org
    };
    return user_info;
}

//===========================================
// data definitions
//===========================================

/**
 * @swagger
 * definitions:
 *
 *   InvokeResponse:
 *     properties:
 *       msg:
 *         type: string
 *       tx_id:
 *         type: string
 *
 *   InvokeResponseWithID:
 *     properties:
 *       id:
 *         type: string
 *       msg:
 *         type: string
 *       tx_id:
 *         type: string
 *
 *   InvokeResponseWithIDAndSecret:
 *     properties:
 *       id:
 *         type: string
 *       secret:
 *         type: string
 *       msg:
 *         type: string
 *       tx_id:
 *         type: string
 *
 *   SignData:
 *     properties:
 *       user_id:
 *         type: string
 *         default: ""
 *       method:
 *         type: string
 *         enum: ["GET", "POST", "PUT"]
 *         default: "GET"
 *       api_path:
 *         type: string
 *         default: "omr/api/v1/"
 *       sign_key:
 *         type: string
 *         default: ""
 *       sign_algorithm:
 *         type: string
 *         default: "sha256"
 *       payload:
 *         type: object
 *         default: ""
 *       is_proxy:
 *         type: boolean
 *         default: false
 *       encrypt_payload:
 *         type: boolean
 *         default: false
 *
 *   OrgNew:
 *     properties:
 *       id:
 *         type: string
 *         default: "required"
 *       secret:
 *         type: string
 *         default: "required"
 *       name:
 *         type: string
 *         default: "required"
 *       ca_org:
 *         type: string
 *         default: "required"
 *       email:
 *         type: string
 *         default: "required"
 *       data:
 *         type: object
 *         default: {}
 *
 *   OrgUpdate:
 *     properties:
 *       id:
 *         type: string
 *         default: "required"
 *       name:
 *         type: string
 *         default: "required"
 *       ca_org:
 *         type: string
 *         default: "required"
 *       data:
 *         type: object
 *         default: {}
 *       status:
 *         type: string
 *         default: "active"
 *
 *   Org:
 *     properties:
 *       id:
 *         type: string
 *       name:
 *         type: string
 *       role:
 *         type: string
 *       public_key:
 *         type: string
 *       private_key:
 *         type: string
 *       sym_key:
 *         type: string
 *       is_group:
 *         type: boolean
 *       status:
 *         type: string
 *       solution_public_data:
 *         type: object
 *       email:
 *         type: string
 *       kms_public_key_id:
 *         type: string
 *       kms_private_key_id:
 *         type: string
 *       kms_sym_key_id:
 *         type: string
 *       secret:
 *         type: string
 *       solution_private_data:
 *         type: object
 *
 *   UserNew:
 *     properties:
 *       id:
 *         type: string
 *         default: "required"
 *       secret:
 *         type: string
 *         default: "required"
 *       name:
 *         type: string
 *         default: "required"
 *       role:
 *         type: string
 *         enum: ["user", "audit", "system"]
 *         default: "user"
 *       org:
 *         type: string
 *         default: "optional"
 *       email:
 *         type: string
 *         default: "required"
 *       ca_org:
 *         type: string
 *         default: "required"
 *       data:
 *         type: object
 *         default: {}
 *
 *   PatientRegisterAndConsent:
 *     properties:
 *       id:
 *         type: string
 *         default: "required"
 *       secret:
 *         type: string
 *         default: "required"
 *       name:
 *         type: string
 *         default: "required"
 *       email:
 *         type: string
 *         default: "required"
 *       ca_org:
 *         type: string
 *         default: "required"
 *       data:
 *         type: object
 *         default: {}
 *       service_id:
 *         type: string
 *         default: "required"
 *       consents:
 *         type: array
 *         default: "required"
 *         items:
 *             $ref: '#/definitions/PatientRegisterAndConsentInput'
 *
 *   PatientRegisterAndConsentResponseObject:
 *     properties:
 *       msg:
 *         type: string
 *       successes:
 *         type: array
 *         items:
 *             $ref: '#/definitions/ConsentOutputPatientData'
 *       failures:
 *         type: array
 *         items:
 *             $ref: '#/definitions/ConsentOutputPatientData'
 *       status:
 *         type: integer
 *         format: int64
 *       failure_type:
 *         type: string
 *
 *   UserUpdate:
 *     properties:
 *       id:
 *         type: string
 *         default: "required"
 *       name:
 *         type: string
 *         default: "required"
 *       ca_org:
 *         type: string
 *         default: "required"
 *       role:
 *         type: string
 *         enum: ["user", "audit", "system"]
 *         default: "user"
 *       email:
 *         type: string
 *         default: "required"
 *       data:
 *         type: object
 *         default: {}
 *       status:
 *         type: string
 *         default: "active"
 *
 *   User:
 *     properties:
 *       id:
 *         type: string
 *       name:
 *         type: string
 *       role:
 *         type: string
 *         enum: ["user", "audit", "system"]
 *         default: "user"
 *       is_group:
 *         type: boolean
 *         default: false
 *       org:
 *         type: string
 *       solution_info:
 *         $ref: '#/definitions/SolutionInfo'
 *       email:
 *         type: string
 *       solution_private_data:
 *         type: object
 *         default: {}
 *       status:
 *         type: string
 *         default: "active"
 *       public_key:
 *         type: string
 *       private_key:
 *         type: string
 *       sym_key:
 *         type: string
 *       kms_private_key_id:
 *         type: string
 *       kms_public_key_id:
 *         type: string
 *       kms_sym_key_id:
 *         type: string
 *       secret:
 *         type: string
 *
 *   SolutionInfo:
 *     properties:
 *       is_org_admin:
 *         type: string
 *       services:
 *         type: string
 *
 *   SimpleUser:
 *     properties:
 *       id:
 *         type: string
 *       name:
 *         type: string
 *       role:
 *         type: string
 *         enum: ["user", "org", "audit", "system"]
 *         default: "user"
 *       is_group:
 *         type: boolean
 *         default: false
 *       status:
 *         type: string
 *         default: "active"
 *
 *   UserResponse:
 *     properties:
 *       id:
 *         type: string
 *       key_id:
 *         type: object
 *       secret:
 *         type: string
 *       msg:
 *         type: object
 *         properties:
 *           result:
 *             type: string
 *           error:
 *             type: string
 *       token:
 *         type: string
 *
 *   LoginResponse:
 *     properties:
 *       id:
 *         type: string
 *       name:
 *         type: string
 *       secret:
 *         type: string
 *       msg:
 *         type: object
 *         properties:
 *           result:
 *             type: string
 *           error:
 *             type: string
 *       token:
 *         type: string
 *
 *   Datatype:
 *     properties:
 *       id:
 *         type: string
 *       description:
 *         type: string
 *
 *   DatatypeNew:
 *     properties:
 *       id:
 *         type: string
 *         default: "required"
 *       description:
 *         type: string
 *         default: "required"
 *
 *   DatatypeUpdate:
 *     properties:
 *       id:
 *         type: string
 *         default: "required"
 *       description:
 *         type: string
 *         default: "required"
 *
 *   DatatypeResponse:
 *     properties:
 *       datatype_id:
 *         type: string
 *       description:
 *         type: string
 *
 *   ServiceDataType:
 *     properties:
 *       datatype_id:
 *         type: string
 *       access:
 *         type: array
 *         items:
 *           type: string
 *
 *   Service:
 *     properties:
 *       service_id:
 *         type: string
 *       service_name:
 *         type: string
 *       email:
 *         type: string
 *       datatypes:
 *         type: object
 *       org_id:
 *         type: string
 *       summary:
 *         type: string
 *       terms:
 *         type: object
 *       payment_required:
 *         type: string
 *       solution_private_data:
 *         type: object
 *       status:
 *         type: string
 *       create_date:
 *         type: integer
 *         format: int64
 *       update_date:
 *         type: integer
 *         format: int64
 *
 *   ServiceNew:
 *     properties:
 *       id:
 *         type: string
 *         default: "required"
 *       name:
 *         type: string
 *         default: "required"
 *       secret:
 *         type: string
 *         default: "required"
 *       ca_org:
 *         type: string
 *         default: "required"
 *       email:
 *         type: string
 *         default: "required"
 *       org_id:
 *         type: string
 *         default: "required"
 *       summary:
 *         type: string
 *         default: "required"
 *       terms:
 *         type: object
 *       payment_required:
 *         type: string
 *         enum: ["no", "yes"]
 *         default: "no"
 *       datatypes:
 *         type: array
 *         items:
 *           $ref: '#/definitions/ServiceDataType'
 *       solution_private_data:
 *         type: object
 *         default: {}
 *
 *   ServiceUpdate:
 *     properties:
 *       id:
 *         type: string
 *         default: "required"
 *       name:
 *         type: string
 *         default: "required"
 *       ca_org:
 *         type: string
 *         default: "required"
 *       org_id:
 *         type: string
 *         default: "required"
 *       summary:
 *         type: string
 *         default: "Description of the service"
 *       terms:
 *         type: object
 *       payment_required:
 *         type: string
 *         enum: ["no", "yes"]
 *         default: "no"
 *       datatypes:
 *         type: array
 *         items:
 *           $ref: '#/definitions/ServiceDataType'
 *       solution_private_data:
 *         type: object
 *         default: {}
 *       status:
 *         type: string
 *         enum: ["active", "inactive"]
 *         default: "active"
 *
 *   PatientEnrollment:
 *     properties:
 *       enrollment_id:
 *         type: string
 *       user_id:
 *         type: string
 *       user_name:
 *         type: string
 *       service_id:
 *         type: string
 *       service_name:
 *         type: string
 *       enroll_date:
 *         type: integer
 *         format: int64
 *       status:
 *         type: string
 *
 *   PatientEnrollmentNew:
 *     properties:
 *       user:
 *         type: string
 *       service:
 *         type: string
 *
 *   PatientUnenroll:
 *     properties:
 *       user:
 *         type: string
 *       service:
 *         type: string
 *
 *   ConsentInputPatientData:
 *     properties:
 *       patient_id:
 *         type: string
 *         default: "required"
 *       service_id:
 *         type: string
 *         default: "required"
 *       target_id:
 *         type: string
 *         default: "required"
 *       datatype_id:
 *         type: string
 *         default: "required"
 *       option:
 *         type: array
 *         default: ["write", "read", "deny"]
 *         items:
 *           type: string
 *       expiration:
 *         type: integer
 *         format: int64
 *         default: 0
 *
 *   ConsentOutputPatientData:
 *     properties:
 *       owner_id:
 *         type: string
 *         default: "required"
 *       service_id:
 *         type: string
 *         default: "required"
 *       target_id:
 *         type: string
 *         default: "required"
 *       datatype_id:
 *         type: string
 *         default: "required"
 *       option:
 *         type: array
 *         default: ["write", "read", "deny"]
 *         items:
 *           type: string
 *       expirationTimestamp:
 *         type: integer
 *         format: int64
 *         default: 0
 *
 *   MultiConsentInputPatientData:
 *     type: array
 *     items:
 *       $ref: '#/definitions/ConsentInputPatientData'
 *
 *   PatientRegisterAndConsentInput:
 *     properties:
 *       datatype_id:
 *         type: string
 *         default: "required"
 *       target_id:
 *         type: string
 *         default: "required"
 *       option:
 *         type: array
 *         default: ["write", "read", "deny"]
 *         items:
 *           type: string
 *       expirationTimestamp:
 *         type: integer
 *         format: int64
 *         default: 0
 *
 *   ConsentInputOwnerData:
 *     properties:
 *       service_id:
 *         type: string
 *         default: "required"
 *       datatype_id:
 *         type: string
 *         default: "required"
 *       target_id:
 *         type: string
 *         default: "required"
 *       option:
 *         type: array
 *         default: ["write", "read", "deny"]
 *         items:
 *           type: string
 *       expiration:
 *         type: integer
 *         format: int64
 *         default: 0
 *
 *   Consent:
 *     properties:
 *       owner:
 *         type: string
 *       service:
 *         type: string
 *       datatype:
 *         type: string
 *       target:
 *         type: string
 *       option:
 *         type: array
 *         items:
 *           type: string
 *       timestamp:
 *         type: integer
 *         format: int64
 *       expiration:
 *         type: integer
 *         format: int64
 *         default: 0
 *
 *   ConsentRequest:
 *     properties:
 *       org:
 *         type: string
 *       user:
 *         type: string
 *       service:
 *         type: string
 *       service_name:
 *         type: string
 *       datatypes:
 *         type: array
 *         items:
 *           $ref: '#/definitions/ServiceDataType'
 *       enroll_date:
 *         type: integer
 *
 *   Validation:
 *     properties:
 *       owner:
 *         type: string
 *       target:
 *         type: string
 *       datatype:
 *         type: string
 *       requester:
 *         type: string
 *       requested_access:
 *         type: string
 *       permission_granted:
 *         type: boolean
 *         default: false
 *       message:
 *         type: string
 *       timestamp:
 *         type: integer
 *         format: int64
 *       token:
 *         type: string
 *
 *   UserData:
 *     properties:
 *       data_id:
 *         type: string
 *       owner:
 *         type: string
 *       service:
 *         type: string
 *       datatype:
 *         type: string
 *       timestamp:
 *         type: integer
 *         format: int64
 *         default: 0
 *       data:
 *         type: object
 *
 *   OwnerData:
 *     properties:
 *       data_id:
 *         type: string
 *       owner:
 *         type: string
 *       service:
 *         type: string
 *       datatype:
 *         type: string
 *       timestamp:
 *         type: integer
 *         format: int64
 *         default: 0
 *       data:
 *         type: object
 *
 *   ContractDetail:
 *     properties:
 *       contract_id:
 *         type: string
 *       contract_detail_type:
 *         type: string
 *       contract_detail_terms:
 *         type: object
 *       create_date:
 *         type: integer
 *         format: int64
 *       created_by:
 *         type: string
 *
 *   ContractTerms:
 *     properties:
 *       contract_id:
 *         type: string
 *         default: "required"
 *       contract_terms:
 *         type: object
 *
 *   ContractTerminate:
 *     properties:
 *       contract_id:
 *         type: string
 *         default: "required"
 *       contract_terms:
 *         type: object
 *
 *   ContractSign:
 *     properties:
 *       contract_id:
 *         type: string
 *         default: "required"
 *       contract_terms:
 *         type: object
 *       signed_by:
 *         type: string
 *
 *   ContractPayment:
 *     properties:
 *       contract_id:
 *         type: string
 *         default: "required"
 *       contract_terms:
 *         type: object
 *
 *   ContractPaymentVerification:
 *     properties:
 *       contract_id:
 *         type: string
 *         default: "required"
 *       contract_terms:
 *         type: object
 *
 *   Contract:
 *     properties:
 *       contract_id:
 *         type: string
 *       owner_org_id:
 *         type: string
 *       owner_service_id:
 *         type: string
 *       requester_org_id:
 *         type: string
 *       requester_service_id:
 *         type: string
 *       contract_terms:
 *         type: object
 *       state:
 *         type: string
 *       create_date:
 *         type: integer
 *         format: int64
 *       update_date:
 *         type: integer
 *         format: int64
 *       contract_details:
 *         type: array
 *         items:
 *           $ref: '#/definitions/ContractDetail'
 *       payment_required:
 *         type: string
 *         enum: ["yes", "no"]
 *         default: "no"
 *       payment_verified:
 *         type: string
 *         enum: ["yes", "no"]
 *         default: "no"
 *       max_num_download:
 *         type: integer
 *       num_download:
 *         type: integer
 *
 *   ContractRequest:
 *     properties:
 *       owner_org_id:
 *         type: string
 *         default: "required"
 *       owner_service_id:
 *         type: string
 *         default: "required"
 *       requester_org_id:
 *         type: string
 *         default: "required"
 *       requester_service_id:
 *         type: string
 *         default: "required"
 *       contract_terms:
 *         type: object
 *
 *   ContractInput:
 *     properties:
 *       contract_id:
 *         type: string
 *         default: "required"
 *       owner_org_id:
 *         type: string
 *         default: "required"
 *       owner_service_id:
 *         type: string
 *         default: "required"
 *       requester_org_id:
 *         type: string
 *         default: "required"
 *       requester_service_id:
 *         type: string
 *         default: "required"
 *       contract_terms:
 *         type: object
 *
 *   ContractInvokeResponse:
 *     properties:
 *       contract_id:
 *         type: string
 *         default: "required"
 *
 *   Log:
 *     properties:
 *       transaction_id:
 *         type: string
 *       contract:
 *         type: string
 *       datatype:
 *         type: string
 *       service:
 *         type: string
 *       owner:
 *         type: string
 *       target:
 *         type: string
 *       contract_requester_service:
 *         type: string
 *       contract_owner_service:
 *         type: string
 *       contract_requester_org:
 *         type: string
 *       contract_owner_org:
 *         type: string
 *       data:
 *         type: object
 *       type:
 *         type: string
 *       caller:
 *         type: string
 *       timestamp:
 *         type: integer
 *         format: int64
 */


/**
 * @swagger
 * securityDefinitions:
 *   basicAuth:
 *     type: basic
 *     description: HTTP Basic Authentication.
 */

//===========================================
//Orgs
//===========================================

/**
 * @swagger
 * /omr/api/v1/orgs:
 *   post:
 *     tags:
 *       - Solution Organizations
 *     description: "Register a new org"
 *     produces:
 *       - application/json
 *     parameters:
 *       - name: token
 *         in: header
 *         description: login token
 *         required: false
 *         type: string
 *         format: string
 *       - name: Org
 *         description: Org Object
 *         in: body
 *         required: true
 *         schema:
 *           $ref: '#/definitions/OrgNew'
 *     security:
 *       - basicAuth: []
 *     responses:
 *       200:
 *         description: Org registration response object
 *         schema:
 *           type: array
 *           items:
 *             $ref: '#/definitions/InvokeResponseWithIDAndSecret'
 */
router.route('/orgs').post(function (req, res) {

    const verify_signature = req_handler.solutionConfig["verify_user_signature"];

    if (!req.body.id) {
        var errmsg = "id missing";
        logger.error(errmsg);
        res.status(400).json({ msg: "Invalid data:" + errmsg, status: 400 });
    } else if (!req.body.secret) {
        var errmsg = "secret missing";
        logger.error(errmsg);
        res.status(400).json({ msg: "Invalid data:" + errmsg, status: 400 });
    } else if (!req.body.name) {
        var errmsg = "name missing";
        logger.error(errmsg);
        res.status(400).json({ msg: "Invalid data:" + errmsg, status: 400 });
    } else if (!req.body.ca_org) {
        var errmsg = "ca_org missing";
        logger.error(errmsg);
        res.status(400).json({ msg: "Invalid data:" + errmsg, status: 400 });
    } else if (!req.body.email) {
        var errmsg = "email missing";
        logger.error(errmsg);
        res.status(400).json({ msg: "Invalid data:" + errmsg, status: 400 });
    } else {
        var orgData = req.body.data && typeof req.body.data == 'object' ? req.body.data : {};
        var data = {
            type: 'registerOrg',
            id: req.body.id,
            secret: req.body.secret,
            ca_org: req.body.ca_org,
            name: req.body.name,
            role: "org",
            email: req.body.email,
            status: "active",
            is_group: true,
            data: orgData,
            public_key: "",
            private_key: "",
            sym_key: "",
            verify_key: "",
            action: "register"
        };
        solution_req_handler.process_api(data, req, res);
    }
});

/**
 * @swagger
 * /omr/api/v1/orgs/{org_id}:
 *   put:
 *     tags:
 *       - Solution Organizations
 *     description: Update an organization
 *     produces:
 *       - application/json
 *     parameters:
 *       - name: token
 *         in: header
 *         description: login token
 *         required: false
 *         type: string
 *         format: string
 *       - name: org_id
 *         description: Organization id
 *         in: path
 *         required: true
 *         type: string
 *       - name: Org
 *         description: Organization detail
 *         in: body
 *         required: true
 *         schema:
 *           $ref: '#/definitions/OrgUpdate'
 *     security:
 *       - basicAuth: []
 *     responses:
 *       200:
 *         description: Org update response object
 *         schema:
 *           type: array
 *           items:
 *             $ref: '#/definitions/InvokeResponseWithIDAndSecret'
 */
router.route('/orgs/:org_id').put(function (req, res) {
    if (!req.body.id) {
        var errmsg = "id missing";
        logger.error(errmsg);
        res.status(400).json({ msg: "Invalid data:" + errmsg, status: 400 });
    } else if (req.body.id != req.params.org_id) {
        var errmsg = "ID in path and data body does not match";
        logger.error(errmsg);
        res.status(400).json({ msg: errmsg, status: 400 });
    } else if (!req.body.name) {
        var errmsg = "name missing";
        logger.error(errmsg);
        res.status(400).json({ msg: "Invalid data:" + errmsg, status: 400 });
    } else if (!req.body.ca_org) {
        var errmsg = "ca_org missing";
        logger.error(errmsg);
        res.status(400).json({ msg: "Invalid data:" + errmsg, status: 400 });
    } else if (!req.body.status || req.body.status != "active") {
        var errmsg = "status has to be \"active\"";
        logger.error(errmsg);
        res.status(400).json({ msg: "Invalid data:" + errmsg, status: 400 });
    } else if (!req.body.email) {
        var errmsg = "email missing";
        logger.error(errmsg);
        res.status(400).json({ msg: "Invalid data:" + errmsg, status: 400 });
    } else {
        var orgData = req.body.data && typeof req.body.data == 'object' ? req.body.data : {};
        var data = {
            type: 'updateOrg',
            id: req.body.id,
            name: req.body.name,
            secret: req.body.secret,
            role: "org",
            ca_org: req.body.ca_org,
            status: req.body.status,
            email: req.body.email,
            is_group: true,
            data: orgData,
        };
        solution_req_handler.process_api(data, req, res);
    }
});

/**
 * @swagger
 * /omr/api/v1/orgs:
 *   get:
 *     tags:
 *       - Solution Organizations
 *     description: Returns all organizations
 *     produces:
 *       - application/json
 *     parameters:
 *       - name: token
 *         in: header
 *         description: login token
 *         required: false
 *         type: string
 *         format: string
 *     security:
 *       - basicAuth: []
 *     responses:
 *       200:
 *         description: An array of organization objects
 *         schema:
 *           type: array
 *           items:
 *             $ref: '#/definitions/Org'
 */
router.route('/orgs').get(function (req, res) {
    var data = {
        type: 'getOrgs'
    };
    solution_req_handler.process_api(data, req, res);
});

/**
 * @swagger
 * /omr/api/v1/orgs/{org_id}:
 *   get:
 *     tags:
 *       - Solution Organizations
 *     description: Returns a organization detail
 *     produces:
 *       - application/json
 *     parameters:
 *       - name: token
 *         in: header
 *         description: login token
 *         required: false
 *         type: string
 *         format: string
 *       - name: org_id
 *         description: Organization name (id)
 *         in: path
 *         required: true
 *         type: string
 *     security:
 *       - basicAuth: []
 *     responses:
 *       200:
 *         description: A single organization
 *         schema:
 *           $ref: '#/definitions/Org'
 */
router.route('/orgs/:org_id').get(function (req, res) {
    var data = {
        type: 'getOrg',
        id: req.params.org_id
    };
    solution_req_handler.process_api(data, req, res);
});

//===========================================
// User
//===========================================

/**
 * @swagger
 * /omr/api/v1/users:
 *   post:
 *     tags:
 *       - Solution Users
 *     description: "Register a new user<br>Note: Org admin user is registered by Register Org API"
 *     produces:
 *       - application/json
 *     parameters:
 *       - name: token
 *         in: header
 *         description: login token
 *         required: false
 *         type: string
 *         format: string
 *       - name: User
 *         description: User Object
 *         in: body
 *         required: true
 *         schema:
 *           $ref: '#/definitions/UserNew'
 *     security:
 *       - basicAuth: []
 *     responses:
 *       200:
 *         description: User Registration response object with enroll id and secret
 *         schema:
 *           type: array
 *           items:
 *             $ref: '#/definitions/InvokeResponseWithIDAndSecret'
 */
router.route('/users').post(function (req, res) {
    const verify_signature = req_handler.solutionConfig["verify_user_signature"];
    var role = req.body.role;
    if (!req.body.id) {
        var errmsg = "id missing";
        logger.error(errmsg);
        res.status(400).json({ msg: "Invalid data:" + errmsg, status: 400 });
    } else if (!/^[a-z0-9]+$/i.test(req.body.id)) {
        var errmsg = "id may only contain alphanumeric characters";
        logger.error(errmsg);
        res.status(400).json({ msg: "Invalid data:" + errmsg, status: 400 });
    } else if (!req.body.secret) {
        var errmsg = "secret missing";
        logger.error(errmsg);
        res.status(400).json({ msg: "Invalid data:" + errmsg, status: 400 });
    } else if (!req.body.ca_org) {
        var errmsg = "CA Org missing";
        logger.error(errmsg);
        res.status(400).json({ msg: "Invalid data:" + errmsg, status: 400 });
    } else if (!req.body.name) {
        var errmsg = "name missing";
        logger.error(errmsg);
        res.status(400).json({ msg: "Invalid data:" + errmsg, status: 400 });
    } else if (!req.body.role) {
        var errmsg = "role missing";
        logger.error(errmsg);
        res.status(400).json({ msg: "Invalid data:" + errmsg, status: 400 });
    } else if (role != "user" && role != "audit" && role != "system") {
        var errmsg = "invalid user role " + role;
        logger.error(errmsg);
        res.status(500).json({ msg: "Invalid data:" + errmsg, status: 400 });
    } else if (!req.body.email) {
        var errmsg = "email missing";
        logger.error(errmsg);
        res.status(400).json({ msg: "Invalid data:" + errmsg, status: 400 });
    } else {
        var userData = req.body.data && typeof req.body.data == 'object' ? req.body.data : {};
        var data = {
            type: 'registerUser',
            id: req.body.id,
            secret: req.body.secret,
            ca_org: req.body.ca_org,
            name: req.body.name,
            role: req.body.role,
            org: req.body.org,
            email: req.body.email,
            status: "active",
            data: userData,
            public_key: "",
            private_key: "",
            sym_key: "",
            verify_key: "",
            action: "register"
        };
        solution_req_handler.process_api(data, req, res);
    }
});

/**
 * @swagger
 * /omr/api/v1/users/patient_register_and_consent:
 *   post:
 *     tags:
 *       - Solution Users
 *     description: "Register, enroll a new user and add consent. Only for patients"
 *     produces:
 *       - application/json
 *     parameters:
 *       - name: token
 *         in: header
 *         description: login token
 *         required: false
 *         type: string
 *         format: string
 *       - name: PatientRegisterAndConsent
 *         description: PatientRegisterAndConsent Object
 *         in: body
 *         required: true
 *         schema:
 *           $ref: '#/definitions/PatientRegisterAndConsent'
 *     security:
 *       - basicAuth: []
 *     responses:
 *       200:
 *         description: PatientRegisterAndConsent response object
 *         schema:
 *           $ref: '#/definitions/PatientRegisterAndConsentResponseObject'
 */
 router.route('/users/patient_register_and_consent').post(function (req, res) {
    global.registerError = "Registration";
    global.enrollError = "Enrollment";
    global.consentError = "Consent";
    if (!req.body.id) {
        var errmsg = "id missing";
        logger.error(errmsg);
        res.status(400).json({ msg: "Invalid data:" + errmsg, successes: [], failures: [], status: 400, failure_type: registerError });
    } else if (!/^[a-z0-9]+$/i.test(req.body.id)) {
        var errmsg = "id may only contain alphanumeric characters";
        logger.error(errmsg);
        res.status(400).json({ msg: "Invalid data:" + errmsg, successes: [], failures: [], status: 400, failure_type: registerError });
    } else if (!req.body.secret) {
        var errmsg = "secret missing";
        logger.error(errmsg);
        res.status(400).json({ msg: "Invalid data:" + errmsg, successes: [], failures: [], status: 400, failure_type: registerError });
    } else if (!req.body.ca_org) {
        var errmsg = "CA Org missing";
        logger.error(errmsg);
        res.status(400).json({ msg: "Invalid data:" + errmsg, successes: [], failures: [], status: 400, failure_type: registerError });
    } else if (!req.body.name) {
        var errmsg = "name missing";
        logger.error(errmsg);
        res.status(400).json({ msg: "Invalid data:" + errmsg, successes: [], failures: [], status: 400, failure_type: registerError });
    } else if (!req.body.email) {
        var errmsg = "email missing";
        logger.error(errmsg);
        res.status(400).json({ msg: "Invalid data:" + errmsg, successes: [], failures: [], status: 400, failure_type: registerError });
    } else if (!req.body.service_id) {
        var errmsg = "service ID is missing";
        logger.error(errmsg);
        res.status(400).json({ msg: "Invalid data:" + errmsg, successes: [], failures: [], status: 400, failure_type: registerError });
    } else if (!req.body.consents || !Array.isArray(req.body.consents) || req.body.consents < 1) {
        var errmsg = "consents missing";
        logger.error(errmsg);
        res.status(400).json({ msg: "Invalid data:" + errmsg, successes: [], failures: [], status: 400, failure_type: registerError });
    } else {
        var userData = req.body.data && typeof req.body.data == 'object' ? req.body.data : {};

        var data = {
            type: 'registerEnrollAndConsent',
            id: req.body.id,
            secret: req.body.secret,
            ca_org: req.body.ca_org,
            name: req.body.name,
            role: "user",
            email: req.body.email,
            data: userData,
            service_id: req.body.service_id,
            consents: req.body.consents,
            action: "patientRegisterEnrollConsent"
        };
        solution_req_handler.process_api(data, req, res);
    }
});

/**
 * @swagger
 * /omr/api/v1/users/{user_id}:
 *   put:
 *     tags:
 *       - Solution Users
 *     description: "Update an user"
 *     produces:
 *       - application/json
 *     parameters:
 *       - name: token
 *         in: header
 *         description: login token
 *         required: false
 *         type: string
 *         format: string
 *       - name: user_id
 *         description: User id
 *         in: path
 *         required: true
 *         type: string
 *       - name: User
 *         description: User Object
 *         in: body
 *         required: true
 *         schema:
 *           $ref: '#/definitions/UserUpdate'
 *     security:
 *       - basicAuth: []
 *     responses:
 *       200:
 *         description: User Registration response object with enroll id and secret
 *         schema:
 *           type: array
 *           items:
 *             $ref: '#/definitions/InvokeResponseWithIDAndSecret'
 */
router.route('/users/:user_id').put(function (req, res) {
    var role = req.body.role;

    if (!req.body.id) {
        var errmsg = "id missing";
        logger.error(errmsg);
        res.status(400).json({ msg: "Invalid data:" + errmsg, status: 400 });
    } else if (req.params.user_id != req.body.id) {
        var errmsg = "id mismatch";
        logger.error(errmsg);
        res.status(400).json({ msg: "Invalid id: " + errmsg, status: 400 });
    } else if (!req.body.name) {
        var errmsg = "name missing";
        logger.error(errmsg);
        res.status(400).json({ msg: "Invalid data:" + errmsg, status: 400 });
    } else if (!req.body.ca_org) {
        var errmsg = "ca_org missing";
        logger.error(errmsg);
        res.status(400).json({ msg: "Invalid data:" + errmsg, status: 400 });
    } else if (!req.body.role) {
        var errmsg = "role missing";
        logger.error(errmsg);
        res.status(400).json({ msg: "Invalid data:" + errmsg, status: 400 });
    } else if (role != "user" && role != "audit" && role != "system") {
        var errmsg = "invalid role " + role;
        logger.error(errmsg);
        res.status(500).json({ msg: "Invalid data:" + errmsg, status: 400 });
    } else if (!req.body.email) {
        var errmsg = "email missing";
        logger.error(errmsg);
        res.status(400).json({ msg: "Invalid data:" + errmsg, status: 400 });
    } else if (!req.body.status || (req.body.status != "active" && req.body.status != "inactive")) {
        var errmsg = "status has to be either \"active\" or \"inactive\"";
        logger.error(errmsg);
        res.status(400).json({ msg: "Invalid data:" + errmsg, status: 400 });
    } else {
        var userData = req.body.data && typeof req.body.data == 'object' ? req.body.data : {};
        var data = {
            type: 'updateUser',
            id: req.body.id,
            name: req.body.name,
            role: req.body.role,
            ca_org: req.body.ca_org,
            email: req.body.email,
            status: req.body.status,
            data: userData
        };
        solution_req_handler.process_api(data, req, res);
    }
});

/**
 * @swagger
 * /omr/api/v1/users/{user_id}:
 *   get:
 *     tags:
 *       - Solution Users
 *     description: Returns user details
 *     produces:
 *       - application/json
 *     parameters:
 *       - name: token
 *         in: header
 *         description: login token
 *         required: false
 *         type: string
 *         format: string
 *       - name: user_id
 *         description: User name (id)
 *         in: path
 *         required: true
 *         type: string
 *     security:
 *       - basicAuth: []
 *     responses:
 *       200:
 *         description: User Object
 *         schema:
 *           $ref: '#/definitions/User'
 */
router.route('/users/:user_id').get(function (req, res) {
    var data = {
        type: 'getUser',
        userid: req.params.user_id
    };
    solution_req_handler.process_api(data, req, res);
});

/**
 * @swagger
 * /omr/api/v1/orgs/{org_id}/users:
 *   get:
 *     tags:
 *       - Solution Users
 *     description: "Returns all users for an organization; when getting auditors and system admin, pass * for org ID"
 *     produces:
 *       - application/json
 *     parameters:
 *       - name: token
 *         in: header
 *         description: login token
 *         required: false
 *         type: string
 *         format: string
 *       - name: org_id
 *         description: Organization name (id)
 *         in: path
 *         required: true
 *         type: string
 *       - name: num
 *         description: max number of users to be returned
 *         in: query
 *         required: false
 *         type: integer
 *         format: int32
 *         default: 20
 *       - name: role
 *         description: "role filter, must be org, audit, or system"
 *         in: query
 *         required: false
 *         type: string
 *     security:
 *       - basicAuth: []
 *     responses:
 *       200:
 *         description: Array of user object
 *         schema:
 *           type: array
 *           items:
 *             $ref: '#/definitions/User'
 */
router.route('/orgs/:org_id/users').get(function (req, res) {
    if (!req.params.org_id) {
        var errmsg = "org_id is missing";
        logger.error(errmsg);
        res.status(400).json({ msg: "Invalid query: " + errmsg, status: 400 });
    } else if (req.params.org_id == "*" && !req.query.role) {
        var errmsg = "role is missing, org ID is *";
        logger.error(errmsg);
        res.status(400).json({ msg: "Invalid query: " + errmsg, status: 400 });
    } else {
        var data = {
            type: 'getUsers',
            org: req.params.org_id,
            role: req.query.role || "",
            maxNum: req.query.num || 20,
        };
        solution_req_handler.process_api(data, req, res);
    }
});

/**
 * @swagger
 * /omr/api/v1/users/{user_id}/permissions/admin/{org_id}:
 *   put:
 *     tags:
 *       - Solution Users
 *     description: "Add org admin permission to the user"
 *     produces:
 *       - application/json
 *     parameters:
 *       - name: token
 *         in: header
 *         description: login token
 *         required: false
 *         type: string
 *         format: string
 *       - name: user_id
 *         description: User ID
 *         in: path
 *         required: true
 *         type: string
 *       - name: org_id
 *         description: Org ID
 *         in: path
 *         required: true
 *         type: string
 *     security:
 *       - basicAuth: []
 *     responses:
 *       200:
 *         description: Response object
 *         schema:
 *           type: array
 *           items:
 *             $ref: '#/definitions/InvokeResponse'
 */
router.route('/users/:user_id/permissions/admin/:org_id').put(function (req, res) {
    if (!req.params.org_id) {
        var errmsg = "org_id missing";
        logger.error(errmsg);
        res.status(400).json({ msg: "Invalid request: " + errmsg, status: 400 });
    } else if (!req.params.user_id) {
        var errmsg = "user_id missing";
        logger.error(errmsg);
        res.status(400).json({ msg: "Invalid request: " + errmsg, status: 400 });
    } else {
        var data = {
            type: 'addPermissionOrgAdmin',
            user_id: req.params.user_id,
            org_id: req.params.org_id
        };
        solution_req_handler.process_api(data, req, res);
    }
});

/**
 * @swagger
 * /omr/api/v1/users/{user_id}/permissions/admin/{org_id}:
 *   delete:
 *     tags:
 *       - Solution Users
 *     description: "Remove org admin permission from user"
 *     produces:
 *       - application/json
 *     parameters:
 *       - name: token
 *         in: header
 *         description: login token
 *         required: false
 *         type: string
 *         format: string
 *       - name: user_id
 *         description: User ID
 *         in: path
 *         required: true
 *         type: string
 *       - name: org_id
 *         description: Org ID
 *         in: path
 *         required: true
 *         type: string
 *     security:
 *       - basicAuth: []
 *     responses:
 *       200:
 *         description: Response object
 *         schema:
 *           type: array
 *           items:
 *             $ref: '#/definitions/InvokeResponse'
 */
router.route('/users/:user_id/permissions/admin/:org_id').delete(function (req, res) {
    if (!req.params.org_id) {
        var errmsg = "org_id missing";
        logger.error(errmsg);
        res.status(400).json({ msg: "Invalid request: " + errmsg, status: 400 });
    } else if (!req.params.user_id) {
        var errmsg = "user_id missing";
        logger.error(errmsg);
        res.status(400).json({ msg: "Invalid request: " + errmsg, status: 400 });
    } else {
        var data = {
            type: 'deletePermissionOrgAdmin',
            org_id: req.params.org_id,
            user_id: req.params.user_id
        };
        solution_req_handler.process_api(data, req, res);
    }
});

/**
 * @swagger
 * /omr/api/v1/users/{user_id}/permissions/services/{service_id}:
 *   put:
 *     tags:
 *       - Solution Users
 *     description: "Add service admin permission to the user"
 *     produces:
 *       - application/json
 *     parameters:
 *       - name: token
 *         in: header
 *         description: login token
 *         required: false
 *         type: string
 *         format: string
 *       - name: user_id
 *         description: User ID
 *         in: path
 *         required: true
 *         type: string
 *       - name: service_id
 *         description: Service ID
 *         in: path
 *         required: true
 *         type: string
 *     security:
 *       - basicAuth: []
 *     responses:
 *       200:
 *         description: Response object
 *         schema:
 *           type: array
 *           items:
 *             $ref: '#/definitions/InvokeResponse'
 */
router.route('/users/:user_id/permissions/services/:service_id').put(function (req, res) {
    if (!req.params.service_id) {
        var errmsg = "service_id missing";
        logger.error(errmsg);
        res.status(400).json({ msg: "Invalid request: " + errmsg, status: 400 });
    } else if (!req.params.user_id) {
        var errmsg = "user_id missing";
        logger.error(errmsg);
        res.status(400).json({ msg: "Invalid request: " + errmsg, status: 400 });
    } else {
        var data = {
            type: 'addPermissionServiceAdmin',
            user_id: req.params.user_id,
            service_id: req.params.service_id
        };
        solution_req_handler.process_api(data, req, res);
    }
});

/**
 * @swagger
 * /omr/api/v1/users/{user_id}/permissions/services/{service_id}:
 *   delete:
 *     tags:
 *       - Solution Users
 *     description: "Remove service admin permission from user"
 *     produces:
 *       - application/json
 *     parameters:
 *       - name: token
 *         in: header
 *         description: login token
 *         required: false
 *         type: string
 *         format: string
 *       - name: user_id
 *         description: User ID
 *         in: path
 *         required: true
 *         type: string
 *       - name: service_id
 *         description: Service ID
 *         in: path
 *         required: true
 *         type: string
 *     security:
 *       - basicAuth: []
 *     responses:
 *       200:
 *         description: Response object
 *         schema:
 *           type: array
 *           items:
 *             $ref: '#/definitions/InvokeResponse'
 */
router.route('/users/:user_id/permissions/services/:service_id').delete(function (req, res) {
    if (!req.params.service_id) {
        var errmsg = "service_id missing";
        logger.error(errmsg);
        res.status(400).json({ msg: "Invalid request: " + errmsg, status: 400 });
    } else if (!req.params.user_id) {
        var errmsg = "user_id missing";
        logger.error(errmsg);
        res.status(400).json({ msg: "Invalid request: " + errmsg, status: 400 });
    } else {
        var data = {
            type: 'deletePermissionServiceAdmin',
            service_id: req.params.service_id,
            user_id: req.params.user_id
        };
        solution_req_handler.process_api(data, req, res);
    }
});

/**
 * @swagger
 * /omr/api/v1/users/{user_id}/permissions/audit/{service_id}:
 *   put:
 *     tags:
 *       - Solution Users
 *     description: "Add audit permission to user for a service"
 *     produces:
 *       - application/json
 *     parameters:
 *       - name: token
 *         in: header
 *         description: login token
 *         required: false
 *         type: string
 *         format: string
 *       - name: user_id
 *         description: User Id
 *         in: path
 *         required: true
 *         type: string
 *       - name: service_id
 *         description: Service Id
 *         in: path
 *         required: true
 *         type: string
 *     security:
 *       - basicAuth: []
 *     responses:
 *       200:
 *         description: Response object
 *         schema:
 *           type: array
 *           items:
 *             $ref: '#/definitions/InvokeResponse'
 */
router.route('/users/:user_id/permissions/audit/:service_id').put(function (req, res) {
    if (!req.params.user_id) {
        var errmsg = "user_id missing";
        logger.error(errmsg);
        res.status(400).json({ msg: "Invalid params: " + errmsg, status: 400 });
    } else if (!req.params.service_id) {
        var errmsg = "service_id missing";
        logger.error(errmsg);
        res.status(400).json({ msg: "Invalid params: " + errmsg, status: 400 });
    } else {
        var data = {
            type: 'addPermissionAuditor',
            user_id: req.params.user_id,
            service_id: req.params.service_id
        };
        solution_req_handler.process_api(data, req, res);
    }
});

/**
 * @swagger
 *  /omr/api/v1/users/{user_id}/permissions/audit/{service_id}:
 *   delete:
 *     tags:
 *       - Solution Users
 *     description: "Delete audit permission to user for a service"
 *     produces:
 *       - application/json
 *     parameters:
 *       - name: token
 *         in: header
 *         description: login token
 *         required: false
 *         type: string
 *         format: string
 *       - name: user_id
 *         description: User id
 *         in: path
 *         required: true
 *         type: string
 *       - name: service_id
 *         description: Service id
 *         in: path
 *         required: true
 *         type: string
 *     security:
 *       - basicAuth: []
 *     responses:
 *       200:
 *         description: Response object
 *         schema:
 *           type: array
 *           items:
 *             $ref: '#/definitions/InvokeResponse'
 */
router.route('/users/:user_id/permissions/audit/:service_id').delete(function (req, res) {
    if (!req.params.user_id) {
        var errmsg = "user_id missing";
        logger.error(errmsg);
        res.status(400).json({ msg: "Invalid params: " + errmsg, status: 400 });
    } else if (!req.params.service_id) {
        var errmsg = "service_id missing";
        logger.error(errmsg);
        res.status(400).json({ msg: "Invalid params: " + errmsg, status: 400 });
    } else {
        var data = {
            type: 'deletePermissionAuditor',
            user_id: req.params.user_id,
            service_id: req.params.service_id
        };
        solution_req_handler.process_api(data, req, res);
    }
});

//===========================================
// Datatypes
//===========================================

/**
 * @swagger
 * /omr/api/v1/datatypes:
 *   post:
 *     tags:
 *       - Datatypes
 *     description: "Register a new datatype"
 *     produces:
 *       - application/json
 *     parameters:
 *       - name: token
 *         in: header
 *         description: login token
 *         required: false
 *         type: string
 *         format: string
 *       - name: Datatype
 *         description: Datatype Object
 *         in: body
 *         required: true
 *         schema:
 *           $ref: '#/definitions/DatatypeNew'
 *     security:
 *       - basicAuth: []
 *     responses:
 *       200:
 *         description: Datatype registration response object
 *         schema:
 *           type: array
 *           items:
 *             $ref: '#/definitions/InvokeResponseWithID'
 */
router.route('/datatypes').post(function (req, res) {
    if (!req.body.id) {
        var errmsg = "id missing";
        logger.error(errmsg);
        res.status(400).json({ msg: "Invalid datatype:" + errmsg, status: 400 });
    } else if (!req.body.description) {
        var errmsg = "description missing";
        logger.error(errmsg);
        res.status(400).json({ msg: "Invalid datatype:" + errmsg, status: 400 });
    } else {
        var data = {
            type: 'registerDatatype',
            id: req.body.id,
            description: req.body.description
        };
        solution_req_handler.process_api(data, req, res);
    }
});

/**
 * @swagger
 * /omr/api/v1/datatypes/{datatype_id}:
 *   put:
 *     tags:
 *       - Datatypes
 *     description: "Update a datatype"
 *     produces:
 *       - application/json
 *     parameters:
 *       - name: token
 *         in: header
 *         description: login token
 *         required: false
 *         type: string
 *         format: string
 *       - name: datatype_id
 *         description: Datatype ID
 *         in: path
 *         required: true
 *         type: string
 *       - name: Datatype
 *         description: Datatype Object
 *         in: body
 *         required: true
 *         schema:
 *           $ref: '#/definitions/DatatypeUpdate'
 *     security:
 *       - basicAuth: []
 *     responses:
 *       200:
 *         description: Datatype update response object
 *         schema:
 *           type: array
 *           items:
 *             $ref: '#/definitions/InvokeResponseWithID'
 */
router.route('/datatypes/:datatype_id').put(function (req, res) {
    if (!req.body.id || !req.params.datatype_id) {
        var errmsg = "id missing";
        logger.error(errmsg);
        res.status(400).json({ msg: "Invalid datatype:" + errmsg, status: 400 });
    } else if (req.params.datatype_id != req.body.id) {
        var errmsg = "id mismatch";
        logger.error(errmsg);
        res.status(400).json({ msg: "Invalid datatype: " + errmsg, status: 400 });
    } else if (!req.body.description) {
        var errmsg = "description missing";
        logger.error(errmsg);
        res.status(400).json({ msg: "Invalid datatype:" + errmsg, status: 400 });
    } else {
        var data = {
            type: 'updateDatatype',
            id: req.body.id,
            description: req.body.description
        };
        solution_req_handler.process_api(data, req, res);
    }
});

/**
 * @swagger
 * /omr/api/v1/datatypes/{datatype_id}:
 *   get:
 *     tags:
 *       - Datatypes
 *     description: Returns a datatype detail
 *     produces:
 *       - application/json
 *     parameters:
 *       - name: token
 *         in: header
 *         description: login token
 *         required: false
 *         type: string
 *         format: string
 *       - name: datatype_id
 *         description: Datatype ID
 *         in: path
 *         required: true
 *         type: string
 *     security:
 *       - basicAuth: []
 *     responses:
 *       200:
 *         description: A single datatype
 *         schema:
 *           $ref: '#/definitions/Datatype'
 */
router.route('/datatypes/:datatype_id').get(function (req, res) {
    if (!req.params.datatype_id) {
        var errmsg = "id missing";
        logger.error(errmsg);
        res.status(400).json({ msg: "Invalid datatype:" + errmsg, status: 400 });
    } else {
        var data = {
            type: 'getDatatype',
            id: req.params.datatype_id
        };
        solution_req_handler.process_api(data, req, res);
    }
});

/**
 * @swagger
 * /omr/api/v1/datatypes:
 *   get:
 *     tags:
 *       - Datatypes
 *     description: Returns all datatypes
 *     produces:
 *       - application/json
 *     parameters:
 *       - name: token
 *         in: header
 *         description: login token
 *         required: false
 *         type: string
 *         format: string
 *     security:
 *       - basicAuth: []
 *     responses:
 *       200:
 *         description: Array of datatype objects
 *         schema:
 *           type: array
 *           items:
 *             $ref: '#/definitions/Datatype'
 */
router.route('/datatypes').get(function (req, res) {
    var data = {
        type: 'getAllDatatypes'
    };
    solution_req_handler.process_api(data, req, res);
});

//===========================================
// Services
//===========================================

/**
 * @swagger
 * /omr/api/v1/services:
 *   post:
 *     tags:
 *       - Services
 *     description: "Register a new service"
 *     produces:
 *       - application/json
 *     parameters:
 *       - name: token
 *         in: header
 *         description: login token
 *         required: false
 *         type: string
 *         format: string
 *       - name: Service
 *         description: Service Object
 *         in: body
 *         required: true
 *         schema:
 *           $ref: '#/definitions/ServiceNew'
 *     security:
 *       - basicAuth: []
 *     responses:
 *       200:
 *         description: Service registration response object
 *         schema:
 *           type: array
 *           items:
 *             $ref: '#/definitions/InvokeResponseWithIDAndSecret'
 */
router.route('/services').post(function (req, res) {
    const verify_signature = req_handler.solutionConfig["verify_user_signature"];

    if (!req.body.id) {
        var errmsg = "id missing";
        logger.error(errmsg);
        res.status(400).json({ msg: "Invalid data:" + errmsg, status: 400 });
    } else if (!req.body.name) {
        var errmsg = "name missing";
        logger.error(errmsg);
        res.status(400).json({ msg: "Invalid data:" + errmsg, status: 400 });
    } else if (!req.body.secret) {
        var errmsg = "secret missing";
        logger.error(errmsg);
        res.status(400).json({ msg: "Invalid data:" + errmsg, status: 400 });
    } else if (!req.body.ca_org) {
        var errmsg = "ca_org missing";
        logger.error(errmsg);
        res.status(400).json({ msg: "Invalid data:" + errmsg, status: 400 });
    } else if (!req.body.email) {
        var errmsg = "email missing";
        logger.error(errmsg);
        res.status(400).json({ msg: "Invalid data:" + errmsg, status: 400 });
    } else if (!req.body.org_id) {
        var errmsg = "org_id missing";
        logger.error(errmsg);
        res.status(400).json({ msg: "Invalid data:" + errmsg, status: 400 });
    } else if (!req.body.datatypes || !Array.isArray(req.body.datatypes)) {
        var errmsg = "must have a list of datatypes";
        logger.error(errmsg);
        res.status(400).json({ msg: "Invalid data: " + errmsg, status: 400 });
    } else if (req.body.datatypes.length == 0) {
        var errmsg = "must include at least one datatype";
        logger.error(errmsg);
        res.status(400).json({ msg: "Invalid data: " + errmsg, status: 400 });
    } else if (!req.body.summary) {
        var errmsg = "summary missing";
        logger.error(errmsg);
        res.status(400).json({ msg: "Invalid data: " + errmsg, status: 400 });
    } else if (req.body.payment_required != "yes" && req.body.payment_required != "no") {
        var errmsg = "payment_required must be either 'yes' or 'no'";
        logger.error(errmsg);
        res.status(400).json({ msg: "Invalid data: " + errmsg, status: 400 });
    } else {
        var serviceData = req.body.solution_private_data && typeof req.body.solution_private_data == 'object' ? req.body.solution_private_data : {};
        var data = {
            type: 'registerService',
            id: req.body.id,
            name: req.body.name,
            secret: req.body.secret,
            ca_org: req.body.ca_org,
            role: "org",
            org_id: req.body.org_id,
            email: req.body.email,
            summary: req.body.summary,
            terms: req.body.terms || {},
            datatypes: req.body.datatypes,
            status: "active",
            payment_required: req.body.payment_required,
            is_group: true,
            solution_private_data: serviceData,
            action: "register"
        };
        solution_req_handler.process_api(data, req, res);
    }
});

/**
 * @swagger
 * /omr/api/v1/services/{service_id}:
 *   put:
 *     tags:
 *       - Services
 *     description: "Update an existing service"
 *     produces:
 *       - application/json
 *     parameters:
 *       - name: token
 *         in: header
 *         description: login token
 *         required: false
 *         type: string
 *         format: string
 *       - name: service_id
 *         description: Service (id)
 *         in: path
 *         required: true
 *         type: string
 *       - name: Service
 *         description: Service Object
 *         in: body
 *         required: true
 *         schema:
 *           $ref: '#/definitions/ServiceUpdate'
 *     security:
 *       - basicAuth: []
 *     responses:
 *       200:
 *         description: Successfully updated
 */
router.route('/services/:service_id').put(function (req, res) {
    var errmsg = "";

    if (!req.body.id || !req.params.service_id) {
        errmsg = "id missing";
        logger.error(errmsg);
        res.status(400).json({ msg: "Invalid data: " + errmsg, status: 400 });
    } else if (req.params.service_id != req.body.id) {
        errmsg = "id mismatch";
        logger.error(errmsg);
        res.status(400).json({ msg: "Invalid id: " + errmsg + " | param ID: " + req.params.service_id + ", data ID: " + req.body.id, status: 400 });
    } else if (!req.body.name) {
        errmsg = "name missing";
        logger.error(errmsg);
        res.status(400).json({ msg: "Invalid data: " + errmsg, status: 400 });
    } else if (!req.body.ca_org) {
        var errmsg = "ca_org missing";
        logger.error(errmsg);
        res.status(400).json({ msg: "Invalid data:" + errmsg, status: 400 });
    } else if (!req.body.org_id) {
        errmsg = "org_id missing";
        logger.error(errmsg);
        res.status(400).json({ msg: "Invalid data: " + errmsg, status: 400 });
    } else if (!req.body.summary) {
        errmsg = "summary missing";
        logger.error(errmsg);
        res.status(400).json({ msg: "Invalid data: " + errmsg, status: 400 });
    } else if (!req.body.datatypes || !Array.isArray(req.body.datatypes)) {
        var msg = "must have a list of datatypes";
        logger.error(msg);
        res.status(400).json({ msg: "Invalid data: " + msg, status: 400 });
    } else if (req.body.datatypes.length == 0) {
        var msg = "must include at least one datatype";
        logger.error(msg);
        res.status(400).json({ msg: "Invalid data: " + msg, status: 400 });
    } else if (req.body.payment_required != "yes" && req.body.payment_required != "no") {
        errmsg = "payment required must be yes or no";
        logger.error(errmsg);
        res.status(400).json({ msg: "Invalid data: " + errmsg, status: 400 });
    } else if (req.body.status != "active" && req.body.status != "inactive") {
        errmsg = "invalid status, must be 'active' or 'inactive'";
        logger.error(errmsg);
        res.status(400).json({ msg: "Invalid data: " + errmsg, status: 400 });
    } else {
        var serviceData = req.body.solution_private_data && typeof req.body.solution_private_data == 'object' ? req.body.solution_private_data : {};

        var data = {
            type: 'updateService',
            id: req.body.id,
            name: req.body.name,
            secret: req.body.secret,
            ca_org: req.body.ca_org,
            org_id: req.body.org_id,
            role: "org",
            summary: req.body.summary,
            terms: req.body.terms || {},
            datatypes: req.body.datatypes,
            status: req.body.status,
            payment_required: req.body.payment_required,
            is_group: true,
            solution_private_data: serviceData
        };
        solution_req_handler.process_api(data, req, res);
    }
});

/**
 * @swagger
 * /omr/api/v1/services/{service_id}/removeDatatype/{datatype_id}:
 *   delete:
 *     tags:
 *       - Services
 *     description: "Remove a datatype from a service"
 *     produces:
 *       - application/json
 *     parameters:
 *       - name: token
 *         in: header
 *         description: login token
 *         required: false
 *         type: string
 *         format: string
 *       - name: service_id
 *         description: Service ID
 *         in: path
 *         required: true
 *         type: string
 *       - name: datatype_id
 *         description: Datatype ID
 *         in: path
 *         required: true
 *         type: string
 *     security:
 *       - basicAuth: []
 *     responses:
 *       200:
 *         description: Response object
 *         schema:
 *           type: array
 *           items:
 *             $ref: '#/definitions/InvokeResponse'
 */
router.route('/services/:service_id/removeDatatype/:datatype_id').delete(function (req, res) {
    if (!req.params.service_id) {
        errmsg = "service_id";
        logger.error(errmsg);
        res.status(400).json({ msg: "Missing id: " + errmsg, status: 400 });
    } else if (!req.params.datatype_id) {
        errmsg = "datatype_id";
        logger.error(errmsg);
        res.status(400).json({ msg: "Missing id: " + errmsg, status: 400 });
    } else {
        var data = {
            type: 'removeDatatypeFromService',
            service_id: req.params.service_id,
            datatype_id: req.params.datatype_id
        };
        solution_req_handler.process_api(data, req, res);
    }
});

/**
 * @swagger
 * /omr/api/v1/services/{service_id}:
 *   get:
 *     tags:
 *       - Services
 *     description: Returns a service's detail
 *     produces:
 *       - application/json
 *     parameters:
 *       - name: token
 *         in: header
 *         description: login token
 *         required: false
 *         type: string
 *         format: string
 *       - name: service_id
 *         description: Service ID
 *         in: path
 *         required: true
 *         type: string
 *     security:
 *       - basicAuth: []
 *     responses:
 *       200:
 *         description: Service object
 *         schema:
 *           $ref: '#/definitions/Service'
 */
router.route('/services/:service_id').get(function (req, res) {
    if (!req.params.service_id) {
        errmsg = "id missing";
        logger.error(errmsg);
        res.status(400).json({ msg: "Invalid data: " + errmsg, status: 400 });
    } else {
        var data = {
            type: 'getService',
            service_id: req.params.service_id
        };
        solution_req_handler.process_api(data, req, res);
    }
});

/**
 * @swagger
 * /omr/api/v1/services/{service_id}/addDatatype/{datatype_id}:
 *   post:
 *     tags:
 *       - Services
 *     description: "Add a datatype to a service"
 *     produces:
 *       - application/json
 *     parameters:
 *       - name: token
 *         in: header
 *         description: login token
 *         required: false
 *         type: string
 *         format: string
 *       - name: service_id
 *         description: Service ID
 *         in: path
 *         required: true
 *         type: string
 *       - name: datatype_id
 *         description: Datatype ID
 *         in: path
 *         required: true
 *         type: string
 *       - name: info
 *         description: Access options and reference info
 *         in: body
 *         required: true
 *         schema:
 *           $ref: '#/definitions/ServiceDataType'
 *     security:
 *       - basicAuth: []
 *     responses:
 *       200:
 *         description: Response object
 *         schema:
 *           type: array
 *           items:
 *             $ref: '#/definitions/InvokeResponse'
 */
router.route('/services/:service_id/addDatatype/:datatype_id').post(function (req, res) {
    if (!req.params.service_id) {
        errmsg = "service_id";
        logger.error(errmsg);
        res.status(400).json({ msg: "Missing id: " + errmsg, status: 400 });
    } else if (!req.params.datatype_id) {
        errmsg = "datatype_id";
        logger.error(errmsg);
        res.status(400).json({ msg: "Missing id: " + errmsg, status: 400 });
    } else if (!req.body.datatype_id) {
        errmsg = "datatype_id";
        logger.error(errmsg);
        res.status(400).json({ msg: "Missing id: " + errmsg, status: 400 });
    } else if (req.body.datatype_id != req.params.datatype_id) {
        errmsg = "datatype_id mismatch";
        logger.error(errmsg);
        res.status(400).json({ msg: "Invalid request: " + errmsg, status: 400 });
    } else if (!req.body.access) {
        var errmsg = "access missing";
        logger.error(errmsg);
        res.status(400).json({ msg: "Invalid data:" + errmsg, status: 400 });
    } if (req.body.access.length == 0) {
        var errmsg = "must include at least one access option";
        logger.error(errmsg);
        res.status(400).json({ msg: "Invalid data:" + errmsg, status: 400 });
    } else {
        var data = {
            type: 'addDatatypeToService',
            service_id: req.params.service_id,
            datatype_id: req.body.datatype_id,
            access: req.body.access
        };
        solution_req_handler.process_api(data, req, res);
    }
});

/**
 * @swagger
 * /omr/api/v1/orgs/{org_id}/services:
 *   get:
 *     tags:
 *       - Services
 *     description: Returns all services for an organization
 *     produces:
 *       - application/json
 *     parameters:
 *       - name: token
 *         in: header
 *         description: login token
 *         required: false
 *         type: string
 *         format: string
 *       - name: org_id
 *         description: Organization ID
 *         in: path
 *         required: true
 *         type: string
 *     security:
 *       - basicAuth: []
 *     responses:
 *       200:
 *         description: Array of service objects
 *         schema:
 *           type: array
 *           items:
 *             $ref: '#/definitions/Service'
 */
router.route('/orgs/:org_id/services').get(function (req, res) {
    var data = {
        type: 'getServicesOfOrg',
        org: req.params.org_id
    };
    solution_req_handler.process_api(data, req, res);
});

//===========================================
// Enrollment
//===========================================

/**
 * @swagger
 * /omr/api/v1/services/{service_id}/user/enroll:
 *   post:
 *     tags:
 *       - Enrollments
 *     description: Enroll a user to a service
 *     produces:
 *       - application/json
 *     parameters:
 *       - name: token
 *         in: header
 *         description: login token
 *         required: false
 *         type: string
 *         format: string
 *       - name: service_id
 *         description: Service ID
 *         in: path
 *         required: true
 *         type: string
 *       - name: Patient enrollment
 *         description: "user to be enrolled."
 *         in: body
 *         required: true
 *         schema:
 *           $ref: '#/definitions/PatientEnrollmentNew'
 *     security:
 *       - basicAuth: []
 *     responses:
 *       200:
 *         description: Response object
 *         schema:
 *           type: array
 *           items:
 *             $ref: '#/definitions/InvokeResponse'
 */
router.route('/services/:service_id/user/enroll').post(function (req, res) {
    if (!req.body.user) {
        var errmsg = "user missing";
        logger.error(errmsg);
        res.status(400).json({ msg: "Invalid data: " + errmsg, status: 400 });
    } else if (!req.params.service_id) {
        var errmsg = "service ID is missing";
        logger.error(errmsg);
        res.status(400).json({ msg: "Invalid data: " + errmsg, status: 400 });
    } else if (!req.body.service || req.body.service != req.params.service_id) {
        var errmsg = "service is missing or mismatched";
        logger.error(errmsg);
        res.status(400).json({ msg: "Invalid data: " + errmsg, status: 400 });
    } else if (req.headers["enroll-id"] === req.body.user) {
        var errmsg = "error enrolling patient, caller and enrolled patient must be different";
        logger.error(errmsg);
        res.status(500).json({ msg: errmsg, status: 500 });
    } else {
        var data = {
            type: 'enrollPatient',
            service_id: req.params.service_id,
            user_id: req.body.user,
            status: "active"
        };
        solution_req_handler.process_api(data, req, res);
    }
});

/**
 * @swagger
 * /omr/api/v1/services/{service_id}/user/unenroll:
 *   post:
 *     tags:
 *       - Enrollments
 *     description: Unenroll a user from a service
 *     produces:
 *       - application/json
 *     parameters:
 *       - name: token
 *         in: header
 *         description: login token
 *         required: false
 *         type: string
 *         format: string
 *       - name: service_id
 *         description: Service ID
 *         in: path
 *         required: true
 *         type: string
 *       - name: patient unenrollment
 *         description: "user to be unenrolled."
 *         in: body
 *         required: true
 *         schema:
 *           $ref: '#/definitions/PatientUnenroll'
 *     security:
 *       - basicAuth: []
 *     responses:
 *       200:
 *         description: Response object
 *         schema:
 *           type: array
 *           items:
 *             $ref: '#/definitions/InvokeResponse'
 */
router.route('/services/:service_id/user/unenroll').post(function (req, res) {
    if (!req.params.service_id) {
        var errmsg = "service id missing";
        logger.error(errmsg);
        res.status(400).json({ msg: "Invalid request: " + errmsg, status: 400 });
    } else if (!req.body.user) {
        var errmsg = "user ID missing";
        logger.error(errmsg);
        res.status(400).json({ msg: "Invalid data: " + errmsg, status: 400 });
    } else if (!req.body.service || req.body.service != req.params.service_id) {
        var errmsg = "service is missing or mismatched";
        logger.error(errmsg);
        res.status(400).json({ msg: "Invalid data: " + errmsg, status: 400 });
    } else {
        var data = {
            type: 'unenrollPatient',
            service_id: req.params.service_id,
            user_id: req.body.user,
            status: "inactive"
        };
        solution_req_handler.process_api(data, req, res);
    }
});

/**
 * @swagger
 * /omr/api/v1/users/{user_id}/enrollments:
 *   get:
 *     tags:
 *       - Enrollments
 *     description: Returns a list of patient enrollments
 *     produces:
 *       - application/json
 *     parameters:
 *       - name: token
 *         in: header
 *         description: login token
 *         required: false
 *         type: string
 *         format: string
 *       - name: user_id
 *         description: User ID
 *         in: path
 *         required: true
 *         type: string
 *       - name: status
 *         description: Active, Inactive
 *         in: query
 *         required: false
 *         type: string
 *         enum: ["active", "inactive"]
 *     security:
 *       - basicAuth: []
 *     responses:
 *       200:
 *         description: Patient enrollment object
 *         schema:
 *           $ref: '#/definitions/PatientEnrollment'
 */
router.route('/users/:user_id/enrollments').get(function (req, res) {
    if (!req.params.user_id) {
        var errmsg = "user_id missing";
        logger.error(errmsg);
        res.status(400).json({ msg: errmsg, status: 400 });
    } else {
        var data = {
            type: 'getPatientEnrollments',
            user_id: req.params.user_id,
            status: req.query.status || ""
        };
        solution_req_handler.process_api(data, req, res);
    }
});

/**
 * @swagger
 * /omr/api/v1/services/{service_id}/enrollments:
 *   get:
 *     tags:
 *       - Enrollments
 *     description: Returns a list of patient enrollments for a service
 *     produces:
 *       - application/json
 *     parameters:
 *       - name: token
 *         in: header
 *         description: login token
 *         required: false
 *         type: string
 *         format: string
 *       - name: service_id
 *         description: Service ID
 *         in: path
 *         required: true
 *         type: string
 *       - name: status
 *         description: Active, Inactive
 *         in: query
 *         required: false
 *         type: string
 *         enum: ["active", "inactive"]
 *     security:
 *       - basicAuth: []
 *     responses:
 *       200:
 *         description: Patient enrollment object
 *         schema:
 *           $ref: '#/definitions/PatientEnrollment'
 */
router.route('/services/:service_id/enrollments').get(function (req, res) {
    if (!req.params.service_id) {
        var errmsg = "service_id missing";
        logger.error(errmsg);
        res.status(400).json({ msg: errmsg, status: 400 });
    } else {
        var data = {
            type: 'getServiceEnrollments',
            service_id: req.params.service_id,
            status: req.query.status || ""
        };
        solution_req_handler.process_api(data, req, res);
    }
});

//=============================================
// Consent
//=============================================

/**
 * @swagger
 * /omr/api/v1/consents/patientData:
 *   post:
 *     tags:
 *       - Consents
 *     description: Add or update a consent for patient data
 *     produces:
 *       - application/json
 *     parameters:
 *       - name: token
 *         in: header
 *         description: login token
 *         required: false
 *         type: string
 *         format: string
 *       - name: Consent
 *         description: Consent Object
 *         in: body
 *         required: true
 *         schema:
 *           $ref: '#/definitions/ConsentInputPatientData'
 *     security:
 *       - basicAuth: []
 *     responses:
 *       200:
 *         description: Response object
 *         schema:
 *           type: array
 *           items:
 *             $ref: '#/definitions/InvokeResponse'
 */
router.route('/consents/patientData').post(function (req, res) {
    if (!req.body.patient_id) {
        var errmsg = "patient_id missing";
        logger.error(errmsg);
        res.status(400).json({ msg: "Invalid data: " + errmsg, status: 400 });
    } else if (!req.body.service_id) {
        var errmsg = "service_id missing";
        logger.error(errmsg);
        res.status(400).json({ msg: "Invalid data:" + errmsg, status: 400 });
    } else if (!req.body.datatype_id) {
        var errmsg = "datatype_id missing";
        logger.error(errmsg);
        res.status(400).json({ msg: "Invalid data: " + errmsg, status: 400 });
    } else if (!req.body.target_id) {
        var errmsg = "target_id missing";
        logger.error(errmsg);
        res.status(400).json({ msg: "Invalid data: " + errmsg, status: 400 });
    } else if (!req.body.option) {
        var errmsg = "option missing";
        logger.error(errmsg);
        res.status(400).json({ msg: "Invalid data: " + errmsg, status: 400 });
    } else if (req.body.option.length < 1) {
        var errmsg = "must specify at least one consent option";
        logger.error(errmsg);
        res.status(400).json({ msg: "Invalid data: " + errmsg, status: 400 });
    } else if (req.body.option.length > 2) {
        var errmsg = "too many consent options";
        logger.error(errmsg);
        res.status(400).json({ msg: "Invalid data: " + errmsg, status: 400 });
    } else if ((req.body.option.includes("write") && req.body.option.includes("deny"))
        || (req.body.option.includes("read") && req.body.option.includes("deny"))) {
        var errmsg = "deny cannot be paired with another consent option";
        logger.error(errmsg);
        res.status(400).json({ msg: "Invalid data: " + errmsg, status: 400 });
    } else if (req.body.option.filter((option) => option === "read" || option === "write" || option === "deny").length < 1) {
        var errmsg = "invalid consent option";
        logger.error(errmsg);
        res.status(400).json({ msg: "Invalid data: " + errmsg, status: 400 });
    } else {
        var ts = req.body.expiration;
        if (!ts) {
            ts = 0
        } else if (ts && typeof ts == 'string') {
            var match = ts.match(/\D/g);
            if (match != null) {
                var tsdate = new Date(ts);
                ts = Math.floor(tsdate.getTime() / 1000);
          } else {
            ts = parseInt(ts, 10);
          }
      }
        var data = {
            type: "putConsentPatientData",
            owner_id: req.body.patient_id,
            service_id: req.body.service_id,
            target_service_id: req.body.target_id,
            datatype_id: req.body.datatype_id,
            option: req.body.option,
            expiration: ts
        };
        solution_req_handler.process_api(data, req, res);
    }
});

/**
 * @swagger
 * /omr/api/v1/consents/multi-consent/patientData:
 *   post:
 *     tags:
 *       - Consents
 *     description: Add multiple consents for patient data
 *     produces:
 *       - application/json
 *     parameters:
 *       - name: token
 *         in: header
 *         description: login token
 *         required: false
 *         type: string
 *         format: string
 *       - name: consents
 *         description: One or many consent objects. Each consent must be for a different datatype and must contain all fields (only expiration is optional).
 *         in: body
 *         required: true
 *         schema:
 *           $ref: '#/definitions/MultiConsentInputPatientData'
 *     security:
 *       - basicAuth: []
 *     responses:
 *       200:
 *         description: Response object
 *         schema:
 *           type: array
 *           items:
 *             $ref: '#/definitions/InvokeResponse'
 */
router.route('/consents/multi-consent/patientData').post(function (req, res) {
    if (!req.body || req.body.length < 1) {
        var errmsg = "consents missing";
        logger.error(errmsg);
        res.status(400).json({ msg: "Invalid data: " + errmsg, status: 400 });
    }

    var data = {
        type: "putMultiConsentPatientData",
        consents: req.body
    };
    solution_req_handler.process_api(data, req, res);
});

/**
 * @swagger
 * /omr/api/v1/services/{service_id}/users/{user_id}/consents:
 *   get:
 *     tags:
 *       - Consents
 *     description: Returns all consents for the service/user pair
 *     produces:
 *       - application/json
 *     parameters:
 *       - name: token
 *         in: header
 *         description: login token
 *         required: false
 *         type: string
 *         format: string
 *       - name: service_id
 *         description: Service id
 *         in: path
 *         required: true
 *         type: string
 *       - name: user_id
 *         description: User id
 *         in: path
 *         required: true
 *         type: string
 *     security:
 *       - basicAuth: []
 *     responses:
 *       200:
 *         description: Consent object
 *         schema:
 *           type: array
 *           items:
 *             $ref: '#/definitions/Consent'
 */
router.route('/services/:service_id/users/:user_id/consents').get(function (req, res) {
    if (!req.params.service_id) {
        var errmsg = "service_id missing";
        logger.error(errmsg);
        res.status(400).json({ msg: "Invalid data: " + errmsg, status: 400 });
    } else if (!req.params.user_id) {
        var errmsg = "user_id missing";
        logger.error(errmsg);
        res.status(400).json({ msg: "Invalid data: " + errmsg, status: 400 });
    } else {
        var data = {
            type: 'getConsents',
            service: req.params.service_id,
            user: req.params.user_id
        };
        solution_req_handler.process_api(data, req, res);
    }
});

/**
 * @swagger
 * /omr/api/v1/consents/ownerData:
 *   post:
 *     tags:
 *       - Consents
 *     description: Add or update a consent for owner data
 *     produces:
 *       - application/json
 *     parameters:
 *       - name: token
 *         in: header
 *         description: login token
 *         required: false
 *         type: string
 *         format: string
 *       - name: Consent
 *         description: Consent Object
 *         in: body
 *         required: true
 *         schema:
 *           $ref: '#/definitions/ConsentInputOwnerData'
 *     security:
 *       - basicAuth: []
 *     responses:
 *       200:
 *         description: Response object
 *         schema:
 *           type: array
 *           items:
 *             $ref: '#/definitions/InvokeResponse'
 */
router.route('/consents/ownerData').post(function (req, res) {
    if (!req.body.service_id) {
        var errmsg = "service_id missing";
        logger.error(errmsg);
        res.status(400).json({ msg: "Invalid data: " + errmsg, status: 400 });
    } else if (!req.body.target_id) {
        var errmsg = "target_id missing";
        logger.error(errmsg);
        res.status(400).json({ msg: "Invalid data: " + errmsg, status: 400 });
    } else if (!req.body.datatype_id) {
        var errmsg = "datatype_id missing";
        logger.error(errmsg);
        res.status(400).json({ msg: "Invalid data: " + errmsg, status: 400 });
    } else if (req.body.option.length < 1) {
        var errmsg = "must specify at least one consent option";
        logger.error(errmsg);
        res.status(400).json({ msg: "Invalid data: " + errmsg, status: 400 });
    } else if (req.body.option.length > 2) {
        var errmsg = "too many consent options";
        logger.error(errmsg);
        res.status(400).json({ msg: "Invalid data: " + errmsg, status: 400 });
    } else if (req.body.option.includes("")) {
        var errmsg = "invalid consent option";
        logger.error(errmsg);
        res.status(400).json({ msg: "Invalid data: " + errmsg, status: 400 });
    } else if ((req.body.option.includes("write") && req.body.option.includes("deny"))
        || (req.body.option.includes("read") && req.body.option.includes("deny"))) {
        var errmsg = "deny cannot be paired with another consent option";
        logger.error(errmsg);
        res.status(400).json({ msg: "Invalid data: " + errmsg, status: 400 });
    } else {
        var ts = req.body.expiration;
        if (!ts) {
            ts = 0
        } else if (ts && typeof ts == 'string') {
            var match = ts.match(/\D/g);
            if (match != null) {
                var tsdate = new Date(ts);
                ts = Math.floor(tsdate.getTime() / 1000);
            }
            else {
                ts = parseInt(ts, 10);
            }
        }
        var data = {
            type: 'putConsentOwnerData',
            owner_id: req.body.service_id,
            datatype_id: req.body.datatype_id,
            target_id: req.body.target_id,
            option: req.body.option,
            expiration: ts
        };
        solution_req_handler.process_api(data, req, res);
    }
});

/**
 * @swagger
 * /omr/api/v1/services/{service_id}/users/{user_id}/datatype/{datatype_id}/consents:
 *   get:
 *     tags:
 *       - Consents
 *     description: Returns a consent
 *     produces:
 *       - application/json
 *     parameters:
 *       - name: token
 *         in: header
 *         description: login token
 *         required: false
 *         type: string
 *         format: string
 *       - name: user_id
 *         description: User/Owner ID
 *         in: path
 *         required: true
 *         type: string
 *       - name: service_id
 *         description: Service/Target ID
 *         in: path
 *         required: true
 *         type: string
 *       - name: datatype_id
 *         description: Datatype ID
 *         in: path
 *         required: true
 *         type: string
 *     security:
 *       - basicAuth: []
 *     responses:
 *       200:
 *         description: Consent object
 *         schema:
 *           type: array
 *           items:
 *             $ref: '#/definitions/Consent'
 */
router.route('/services/:service_id/users/:user_id/datatype/:datatype_id/consents').get(function (req, res) {
    if (!req.params.service_id) {
        var errmsg = "service_id missing";
        logger.error(errmsg);
        res.status(400).json({ msg: "Invalid data: " + errmsg, status: 400 });
    } else if (!req.params.user_id) {
        var errmsg = "user_id missing";
        logger.error(errmsg);
        res.status(400).json({ msg: "Invalid data: " + errmsg, status: 400 });
    } else if (!req.params.datatype_id) {
        var errmsg = "datatype_id missing";
        logger.error(errmsg);
        res.status(400).json({ msg: "Invalid data: " + errmsg, status: 400 });
    } else {
        var data = {
            type: 'getConsent',
            owner_id: req.params.user_id,
            target_id: req.params.service_id,
            datatype_id: req.params.datatype_id
        };
        solution_req_handler.process_api(data, req, res);
    }
});

/**
 * @swagger
 * /omr/api/v1/users/{patient_id}/consents:
 *   get:
 *     tags:
 *       - Consents
 *     description: Returns consents for given owner/patient
 *     produces:
 *       - application/json
 *     parameters:
 *       - name: token
 *         in: header
 *         description: login token
 *         required: false
 *         type: string
 *         format: string
 *       - name: patient_id
 *         description: Patient/owner ID
 *         in: path
 *         required: true
 *         type: string
 *     security:
 *       - basicAuth: []
 *     responses:
 *       200:
 *         description: Consent object
 *         schema:
 *           type: array
 *           items:
 *             $ref: '#/definitions/Consent'
 */
router.route('/users/:patient_id/consents').get(function (req, res) {
    if (!req.params.patient_id) {
        var errmsg = "patient_id missing";
        logger.error(errmsg);
        res.status(400).json({ msg: "Invalid data: " + errmsg, status: 400 });
    } else {
        var data = {
            type: 'getConsentsWithOwnerID',
            owner_id: req.params.patient_id
        };
        solution_req_handler.process_api(data, req, res);
    }
});

/**
 * @swagger
 * /omr/api/v1/users/{patient_id}/consents/requests:
 *   get:
 *     tags:
 *       - Consents
 *     description: Returns consent requests (current and pending requests) for given patient
 *     produces:
 *       - application/json
 *     parameters:
 *       - name: token
 *         in: header
 *         description: login token
 *         required: false
 *         type: string
 *         format: string
 *       - name: patient_id
 *         description: Patient ID
 *         in: path
 *         required: true
 *         type: string
 *       - name: service_id
 *         description: Service ID
 *         in: query
 *         required: false
 *         type: string
 *     security:
 *       - basicAuth: []
 *     responses:
 *       200:
 *         description: Consent object
 *         schema:
 *           type: array
 *           items:
 *             $ref: '#/definitions/ConsentRequest'
 */
router.route('/users/:patient_id/consents/requests').get(function (req, res) {
    if (!req.params.patient_id) {
        var errmsg = "patient_id missing";
        logger.error(errmsg);
        res.status(400).json({ msg: "Invalid data: " + errmsg, status: 400 });
    } else {
        var data = {
            type: 'getConsentRequests',
            patient: req.params.patient_id,
            service: req.query.service_id || ""
        };
        solution_req_handler.process_api(data, req, res);
    }
});

/**
 * @swagger
 * /omr/api/v1/services/{owner_service_id}/ownerdata/consents:
 *   get:
 *     tags:
 *       - Consents
 *     description: Returns all consents that an owner service has given
 *     produces:
 *       - application/json
 *     parameters:
 *       - name: token
 *         in: header
 *         description: login token
 *         required: false
 *         type: string
 *         format: string
 *       - name: owner_service_id
 *         description: Owner Service ID
 *         in: path
 *         required: true
 *         type: string
 *     security:
 *       - basicAuth: []
 *     responses:
 *       200:
 *         description: Consent object
 *         schema:
 *           type: array
 *           items:
 *             $ref: '#/definitions/Consent'
 */
router.route('/services/:owner_service_id/ownerdata/consents').get(function (req, res) {
    if (!req.params.owner_service_id) {
        var errmsg = "owner_service_id missing";
        logger.error(errmsg);
        res.status(400).json({ msg: "Invalid data: " + errmsg, status: 400 });
    }

    var data = {
        type: 'getConsentsWithOwnerID',
        owner_id: req.params.owner_service_id
    };
    solution_req_handler.process_api(data, req, res);
});

/**
 * @swagger
 * /omr/api/v1/services/{service_id}/consents:
 *   get:
 *     tags:
 *       - Consents
 *     description: Returns consents for given target
 *     produces:
 *       - application/json
 *     parameters:
 *       - name: token
 *         in: header
 *         description: login token
 *         required: false
 *         type: string
 *         format: string
 *       - name: service_id
 *         description: Service/target ID
 *         in: path
 *         required: true
 *         type: string
 *     security:
 *       - basicAuth: []
 *     responses:
 *       200:
 *         description: Consent object
 *         schema:
 *           type: array
 *           items:
 *             $ref: '#/definitions/Consent'
 */
router.route('/services/:service_id/consents').get(function (req, res) {
    if (!req.params.service_id) {
        var errmsg = "service_id missing";
        logger.error(errmsg);
        res.status(400).json({ msg: "Invalid data: " + errmsg, status: 400 });
    } else {
        var data = {
            type: 'getConsentsWithTargetID',
            target_id: req.params.service_id
        };
        solution_req_handler.process_api(data, req, res);
    }
});

/**
 * @swagger
 * /omr/api/v1/services/{service_id}/owner_service/{owner_service_id}/datatype/{datatype_id}/ownerdata/consents:
 *   get:
 *     tags:
 *       - Consents
 *     description: Returns an owner data consent
 *     produces:
 *       - application/json
 *     parameters:
 *       - name: token
 *         in: header
 *         description: login token
 *         required: false
 *         type: string
 *         format: string
 *       - name: owner_service_id
 *         description: Owner Service ID
 *         in: path
 *         required: true
 *         type: string
 *       - name: service_id
 *         description: Target Service ID
 *         in: path
 *         required: true
 *         type: string
 *       - name: datatype_id
 *         description: Datatype ID
 *         in: path
 *         required: true
 *         type: string
 *     security:
 *       - basicAuth: []
 *     responses:
 *       200:
 *         description: Consent object
 *         schema:
 *           $ref: '#/definitions/Consent'
 */
router.route('/services/:service_id/owner_service/:owner_service_id/datatype/:datatype_id/ownerdata/consents').get(function (req, res) {
    if (!req.params.service_id) {
        var errmsg = "service missing";
        logger.error(errmsg);
        res.status(400).json({ msg: "Invalid data: " + errmsg, status: 400 });
    } else if (!req.params.owner_service_id) {
        var errmsg = "owner service id missing";
        logger.error(errmsg);
        res.status(400).json({ msg: "Invalid data: " + errmsg, status: 400 });
    } else if (!req.params.datatype_id) {
        var errmsg = "datatype id missing";
        logger.error(errmsg);
        res.status(400).json({ msg: "Invalid data: " + errmsg, status: 400 });
    }

    var data = {
        type: 'getConsentOwnerData',
        owner_service: req.params.owner_service_id,
        service: req.params.service_id,
        datatype: req.params.datatype_id
    };
    solution_req_handler.process_api(data, req, res);
});

/**
 * @swagger
 * /omr/api/v1/services/{service_id}/users/{user_id}/datatype/{datatype_id}/validation/{access}:
 *   get:
 *     tags:
 *       - Consents
 *     description: Currently this function can only be called by the admin of the consent target service. Validates consent for accessing the data
 *     produces:
 *       - application/json
 *     parameters:
 *       - name: token
 *         in: header
 *         description: login token
 *         required: false
 *         type: string
 *         format: string
 *       - name: service_id
 *         description: Target/Service ID
 *         in: path
 *         required: true
 *         type: string
 *       - name: user_id
 *         description: Owner ID
 *         in: path
 *         required: true
 *         type: string
 *       - name: datatype_id
 *         description: DataType defined in the service
 *         in: path
 *         required: true
 *         type: string
 *       - name: access
 *         description: Requested access
 *         in: path
 *         required: true
 *         type: string
 *         enum: [ "write", "read"]
 *     security:
 *       - basicAuth: []
 *     responses:
 *       200:
 *         description: Consent Validation object
 *         schema:
 *           $ref: '#/definitions/Validation'
 */
router.route('/services/:service_id/users/:user_id/datatype/:datatype_id/validation/:access').get(function (req, res) {
    // TODO: validateConsent type is overloaded
    if (!req.params.user_id) {
        var errmsg = "user_id missing";
        logger.error(errmsg);
        res.status(400).json({ msg: "Invalid data:" + errmsg, status: 400 });
    } else if (!req.params.service_id) {
        var errmsg = "service_id missing";
        logger.error(errmsg);
        res.status(400).json({ msg: "Invalid data:" + errmsg, status: 400 });
    } else if (!req.params.datatype_id) {
        var errmsg = "datatype_id missing";
        logger.error(errmsg);
        res.status(400).json({ msg: "Invalid data:" + errmsg, status: 400 });
    } else if (!req.params.access) {
        var errmsg = "access missing";
        logger.error(errmsg);
        res.status(400).json({ msg: "Invalid data:" + errmsg, status: 400 });
    }

    var data = {
        type: 'validateConsent',
        owner: req.params.user_id,
        service: req.params.service_id,
        datatype: req.params.datatype_id,
        access: req.params.access
    };
    solution_req_handler.process_api(data, req, res);
});

/**
 * @swagger
 * /omr/api/v1/services/{service_id}/owner_service/{owner_service_id}/datatype/{datatype_id}/validation/{access}/ownerdata:
 *   get:
 *     tags:
 *       - Consents
 *     description: Validates consents for accessing owner data.
 *     produces:
 *       - application/json
 *     parameters:
 *       - name: token
 *         in: header
 *         description: login token
 *         required: false
 *         type: string
 *         format: string
 *       - name: service_id
 *         description: Target Service ID
 *         in: path
 *         required: true
 *         type: string
 *       - name: owner_service_id
 *         description: Owner Service ID
 *         in: path
 *         required: true
 *         type: string
 *       - name: datatype_id
 *         description: DataType item defined in the service
 *         in: path
 *         required: true
 *         type: string
 *       - name: access
 *         description: Requested access
 *         in: path
 *         required: true
 *         type: string
 *         enum: [ "write", "read"]
 *     security:
 *       - basicAuth: []
 *     responses:
 *       200:
 *         description: Consent Validation object
 *         schema:
 *           $ref: '#/definitions/Validation'
 */
router.route('/services/:service_id/owner_service/:owner_service_id/datatype/:datatype_id/validation/:access/ownerdata').get(function (req, res) {
    if (!req.params.service_id) {
        var errmsg = "service missing";
        logger.error(errmsg);
        res.status(400).json({ msg: "Invalid data: " + errmsg, status: 400 });
    } else if (!req.params.datatype_id) {
        var errmsg = "datatype missing";
        logger.error(errmsg);
        res.status(400).json({ msg: "Invalid data: " + errmsg, status: 400 });
    } else if (!req.params.access) {
        var errmsg = "access missing";
        logger.error(errmsg);
        res.status(400).json({ msg: "Invalid data: " + errmsg, status: 400 });
    } else if (!req.params.owner_service_id) {
        var errmsg = "owner service id missing";
        logger.error(errmsg);
        res.status(400).json({ msg: "Invalid data: " + errmsg, status: 400 });
    }

    // TODO: validateConsent type is overloaded
    var data = {
        type: 'validateConsent',
        owner: req.params.owner_service_id,
        service: req.params.service_id,
        datatype: req.params.datatype_id,
        access: req.params.access
    };
    solution_req_handler.process_api(data, req, res);
});


//===========================================
// User Data
//===========================================

/**
 * @swagger
 * /omr/api/v1/services/{service_id}/users/{user_id}/datatype/{datatype}/upload:
 *   post:
 *     tags:
 *       - User Data
 *     description: Upload User Data. User consents for write access will be checked.
 *     produces:
 *       - application/json
 *     parameters:
 *       - name: token
 *         in: header
 *         description: login token
 *         required: false
 *         type: string
 *         format: string
 *       - name: service_id
 *         description: Service ID
 *         in: path
 *         required: true
 *         type: string
 *       - name: user_id
 *         description: User ID
 *         in: path
 *         required: true
 *         type: string
 *       - name: datatype
 *         description: Datatype defined in the service
 *         in: path
 *         required: true
 *         type: string
 *       - name: data
 *         description: "User's data in JSON format"
 *         in: body
 *         required: true
 *         schema:
 *           type: object
 *     security:
 *       - basicAuth: []
 *     responses:
 *       200:
 *         description: Response object
 *         schema:
 *           type: array
 *           items:
 *             $ref: '#/definitions/InvokeResponse'
 */
router.route('/services/:service_id/users/:user_id/datatype/:datatype/upload').post(function (req, res) {
    var userdata = req.body;

    if (!userdata || typeof userdata != 'object') {
        var errmsg = "invalid data";
        logger.error(errmsg);
        res.status(400).json({ msg: "Invalid data: " + errmsg, status: 400 });
        return;
    }

    var data = {
        type: 'uploadUserData',
        service: req.params.service_id,
        user: req.params.user_id,
        datatype: req.params.datatype,
        userdata: userdata
    };
    solution_req_handler.process_api(data, req, res);
});

/**
 * @swagger
 * /omr/api/v1/services/{service_id}/users/{user_id}/datatypes/{datatype_id}/download:
 *   get:
 *     tags:
 *       - User Data
 *     description: "Get User Data by Service Id, User Id, and Datatype Id. Only user data stored after the timestamp will be returned."
 *     produces:
 *       - application/json
 *     parameters:
 *       - name: token
 *         in: header
 *         description: login token
 *         required: false
 *         type: string
 *         format: string
 *       - name: service_id
 *         description: Consent target ID
 *         in: path
 *         required: true
 *         type: string
 *       - name: user_id
 *         description: User ID
 *         in: path
 *         required: true
 *         type: string
 *       - name: datatype_id
 *         description: Datatype ID
 *         in: path
 *         required: true
 *         type: string
 *       - name: start_timestamp
 *         description: starting timestamp (optional). Cull for data uploaded before at specified time.  Can enter timestamp or date (mm/dd/yyyy)
 *         in: query
 *         required: false
 *         type: integer
 *         format: int64
 *       - name: end_timestamp
 *         description: end timestamp (optional). Cull for data uploaded after specified time.  Can enter timestamp or date (mm/dd/yyyy)
 *         in: query
 *         required: false
 *         type: integer
 *         format: int64
 *       - name: latest_only
 *         description: If true, start_timestamp, end_timestamp and maxNum are ignored and only latest data will be returned.
 *         in: query
 *         required: false
 *         type: string
 *         enum: [ "true", "false"]
 *         default: "false"
 *       - name: maxNum
 *         description: Max number of results returned. If start timestamp is not specified most resent results are returned. If 0 is given, maxNum will default to 1000.
 *         in: query
 *         required: false
 *         type: string
 *         default: 1000
 *     security:
 *       - basicAuth: []
 *     responses:
 *       200:
 *         description: List of user data objects
 *         schema:
 *           type: array
 *           items:
 *             $ref: '#/definitions/UserData'
 */
router.route('/services/:service_id/users/:user_id/datatypes/:datatype_id/download').get(function (req, res) {
    if (!req.params.service_id) {
        var errmsg = "service_id";
        logger.error(errmsg);
        res.status(400).json({ msg: "Missing id: " + errmsg, status: 400 });
    } else if (!req.params.user_id) {
        var errmsg = "user_id";
        logger.error(errmsg);
        res.status(400).json({ msg: "Missing id: " + errmsg, status: 400 });
    } else if (!req.params.datatype_id) {
        var errmsg = "datatype_id";
        logger.error(errmsg);
        res.status(400).json({ msg: "Missing id: " + errmsg, status: 400 });
    } else {
        var ts1 = req.query.start_timestamp;
        if (!ts1) {
            ts1 = 0
        } else if (ts1 && typeof ts1 == 'string') {
            var match = ts1.match(/\D/g);
            if (match != null) {
                var tsdate = new Date(ts1);
                ts1 = Math.floor(tsdate.getTime() / 1000);
            }
            else {
                ts1 = parseInt(ts1, 10);
            }
        }

        var ts2 = req.query.end_timestamp;
        if (!ts2) {
            ts2 = 0
        } else if (ts2 && typeof ts2 == 'string') {
            var match = ts2.match(/\D/g);
            if (match != null) {
                var tsdate = new Date(ts2);
                ts2 = Math.floor(tsdate.getTime() / 1000);
            }
            else {
                ts2 = parseInt(ts2, 10);
            }
        }

        var data = {
            type: 'downloadUserData',
            service_id: req.params.service_id,
            user_id: req.params.user_id,
            datatype_id: req.params.datatype_id,
            latest_only: req.query.latest_only,
            start_timestamp: ts1,
            end_timestamp: ts2,
            maxNum: req.query.maxNum
        };
        solution_req_handler.process_api(data, req, res);
    }
});

/**
* @swagger
* /omr/api/v1/userdata/download/{consent_token}:
*   get:
*     tags:
*       - User Data
*     description: "Currently this function can only be called by default service admin of consent target service. Get User Data by Consent Validation Token. Token is only valid for 10 min, and is also only valid for the user who originally obtained it. Only user data stored after the timestamp will be returned."
*     produces:
*       - application/json
*     parameters:
*       - name: token
*         in: header
*         description: login token
*         required: false
*         type: string
*         format: string
*       - name: consent_token
*         description: consent validaiton token
*         in: path
*         required: true
*         type: string
*       - name: start_timestamp
*         description: starting timestamp (optional). Cull for data uploaded before at specified time.  Can enter timestamp or date (mm/dd/yyyy)
*         in: query
*         required: false
*         type: integer
*         format: int64
*       - name: end_timestamp
*         description: end timestamp (optional). Cull for data uploaded after specified time.  Can enter timestamp or date (mm/dd/yyyy)
*         in: query
*         required: false
*         type: integer
*         format: int64
*       - name: latest_only
*         description: If true, start_timestamp, end_timestamp and maxNum are ignored and only latest data will be returned.
*         in: query
*         required: false
*         type: string
*         enum: [ "true", "false"]
*         default: "false"
*       - name: maxNum
*         description: Max number of results returned. If start timestamp is not specified most resent results are returned. If 0 is given, maxNum will default to 1000.
*         in: query
*         required: false
*         type: string
*         default: 1000
*     security:
*       - basicAuth: []
*     responses:
*       200:
*         description: List of user data objects
*         schema:
*           type: array
*           items:
*             $ref: '#/definitions/UserData'
*/
router.route('/userdata/download/:consent_token').get(function (req, res) {
    if (!req.params.consent_token) {
        var errmsg = "consent token";
        logger.error(errmsg);
        res.status(400).json({ msg: "Missing: " + errmsg, status: 400 });

    } else {
        var ts1 = req.query.start_timestamp;
        if (!ts1) {
            ts1 = 0
        } else if (ts1 && typeof ts1 == 'string') {
            var match = ts1.match(/\D/g);
            if (match != null) {
                var tsdate = new Date(ts1);
                ts1 = Math.floor(tsdate.getTime() / 1000);
            }
            else {
                ts1 = parseInt(ts1, 10);
            }
        }

        var ts2 = req.query.end_timestamp;
        if (!ts2) {
            ts2 = 0
        } else if (ts2 && typeof ts2 == 'string') {
            var match = ts2.match(/\D/g);
            if (match != null) {
                var tsdate = new Date(ts2);
                ts2 = Math.floor(tsdate.getTime() / 1000);
            }
            else {
                ts2 = parseInt(ts2, 10);
            }
        }

        var data = {
            type: 'downloadUserDataConsentToken',
            latest_only: req.query.latest_only,
            start_timestamp: ts1,
            end_timestamp: ts2,
            maxNum: req.query.maxNum,
            token: req.params.consent_token
        };
        solution_req_handler.process_api(data, req, res);
    }
});


// /**
//  * @swagger
//  * /omr/api/v1/users/{user_id}/services/{service_id}/datatypes/{datatype_id}/timestamps/{timestamp}:
//  *   delete:
//  *     tags:
//  *       - User Data
//  *     description: Delete user data. Only available for data owner.
//  *     produces:
//  *       - application/json
//  *     parameters:
//  *       - name: token
//  *         in: header
//  *         description: login token
//  *         required: false
//  *         type: string
//  *         format: string
//  *       - name: user_id
//  *         description: Patient ID
//  *         in: path
//  *         required: true
//  *         type: string
//  *       - name: service_id
//  *         description: Service ID
//  *         in: path
//  *         required: true
//  *         type: string
//  *       - name: datatype_id
//  *         description: Datatype defined in the service
//  *         in: path
//  *         required: true
//  *         type: string
//  *       - name: timestamp
//  *         description: Unix timestamp (sec)
//  *         in: path
//  *         type: integer
//  *         format: int64
//  *         required: true
//  *     security:
//  *       - basicAuth: []
//  *     responses:
//  *       200:
//  *         description: Successfully deleted
//  */
// router.route('/users/:user_id/services/:service_id/datatypes/:datatype_id/timestamps/:timestamp').delete(function(req, res) {
//     if (!req.params.user_id ) {
//         var errmsg = "user ID";
//         logger.error(errmsg);
//         res.status(400).json({msg: "Missing: " + errmsg, status: 400});

//     } else if (!req.params.service_id ) {
//         var errmsg = "service ID";
//         logger.error(errmsg);
//         res.status(400).json({msg: "Missing: " + errmsg, status: 400});

//     } else if (!req.params.datatype_id ) {
//         var errmsg = "datatype ID";
//         logger.error(errmsg);
//         res.status(400).json({msg: "Missing: " + errmsg, status: 400});

//     } else if (!req.params.timestamp ) {
//         var errmsg = "timestamp";
//         logger.error(errmsg);
//         res.status(400).json({msg: "Missing: " + errmsg, status: 400});

//     } else {
//         var data = {
//             type: 'deleteUserData',
//             user: req.params.user_id,
//             service: req.params.service_id,
//             datatype: req.params.datatype_id,
//             timestamp: req.params.timestamp
//         };
//         req_handler.process_api(data, req, res);
//     }
// });

//===========================================
// Owner Data
//===========================================

/**
 * @swagger
 * /omr/api/v1/services/{service_id}/datatype/{datatype}/upload:
 *   post:
 *     tags:
 *       - Owner Data
 *     description: Upload owner Data. Owner can be service or user. In the case of user giving uploading data for his own datatype, substitute User ID for Service ID
 *     produces:
 *       - application/json
 *     parameters:
 *       - name: token
 *         in: header
 *         description: login token
 *         required: false
 *         type: string
 *         format: string
 *       - name: service_id
 *         description: Service ID
 *         in: path
 *         required: true
 *         type: string
 *       - name: datatype
 *         description: Datatype defined in the service
 *         in: path
 *         required: true
 *         type: string
 *       - name: data
 *         description: Owner data in json format
 *         in: body
 *         required: true
 *         schema:
 *           type: object
 *     security:
 *       - basicAuth: []
 *     responses:
 *       200:
 *         description: Response object
 *         schema:
 *           type: array
 *           items:
 *             $ref: '#/definitions/InvokeResponse'
 */
router.route('/services/:service_id/datatype/:datatype/upload').post(function (req, res) {
    var ownerData = req.body;

    if (!ownerData || typeof ownerData != 'object') {
        var msg = "invalid data";
        logger.error(msg);
        res.status(400).json({ msg: "Invalid data: " + msg, status: 400 });
        return;
    }

    var data = {
        type: 'uploadOwnerData',
        service: req.params.service_id,
        datatype: req.params.datatype,
        ownerData: ownerData,
    };
    solution_req_handler.process_api(data, req, res);
});

/**
 * @swagger
 * /omr/api/v1/services/{service_id}/ownerdata/{datatype_id}/downloadAsOwner:
 *   get:
 *     tags:
 *       - Owner Data
 *     description: "Get Owner Data by Service Id and datatype.  This is used by the data owner.  Data requester should use /omr/api/v1/contracts/{contract_id}/ownerdata/{datatype_id}/download:"
 *     produces:
 *       - application/json
 *     parameters:
 *       - name: token
 *         in: header
 *         description: login token
 *         required: false
 *         type: string
 *         format: string
 *       - name: service_id
 *         description: Owner ID
 *         in: path
 *         required: true
 *         type: string
 *       - name: datatype_id
 *         description: Datatype ID
 *         in: path
 *         required: true
 *         type: string
*       - name: start_timestamp
*         description: starting timestamp (inclusive). Cull for data uploaded before at specified time.  Can enter timestamp or date (mm/dd/yyyy)
*         in: query
*         required: false
*         type: integer
*         format: int64
*       - name: end_timestamp
*         description: end timestamp (inclusive). Cull results data uploaded after specified time.  Can enter timestamp or date (mm/dd/yyyy)
*         in: query
*         required: false
*         type: integer
*         format: int64
*       - name: latest_only
*         description: If true, start_timestamp, end_timestamp and maxNum are ignored and only latest data will be returned.
*         in: query
*         required: false
*         type: string
*         enum: [ "true", "false"]
*         default: "false"
*       - name: maxNum
*         description: Max number of results returned. If start and end timestamps are not specified, then the most recent results are returned.
*         in: query
*         required: false
*         type: string
*         default: 1000
 *     security:
 *       - basicAuth: []
 *     responses:
 *       200:
 *         description: List of owner data objects
 *         schema:
 *           type: array
 *           items:
 *             $ref: '#/definitions/OwnerData'
 */
router.route('/services/:service_id/ownerdata/:datatype_id/downloadAsOwner').get(function (req, res) {
    if (!req.params.service_id) {
        var errmsg = "service_id";
        logger.error(errmsg);
        res.status(400).json({ msg: "Missing id: " + errmsg, status: 400 });
    } else if (!req.params.datatype_id) {
        var errmsg = "datatype_id";
        logger.error(errmsg);
        res.status(400).json({ msg: "Missing id: " + errmsg, status: 400 });
    }

    //support date format input for timestamp
    var ts1 = req.query.start_timestamp;
    if (!ts1) {
        ts1 = 0
    } else if (ts1 && typeof ts1 == 'string') {
        var match = ts1.match(/\D/g);
        if (match != null) {
            var tsdate = new Date(ts1);
            ts1 = Math.floor(tsdate.getTime() / 1000);
        }
        else {
            ts1 = parseInt(ts1, 10);
        }
    }
    var ts2 = req.query.end_timestamp;
    if (!ts2) {
        ts2 = 0
    } else if (ts2 && typeof ts2 == 'string') {
        var match = ts2.match(/\D/g);
        if (match != null) {
            var tsdate = new Date(ts2);
            ts2 = Math.floor(tsdate.getTime() / 1000);
        }
        else {
            ts2 = parseInt(ts2, 10);
        }
    }

    var data = {
        type: 'downloadOwnerDataAsOwner',
        service_id: req.params.service_id,
        datatype_id: req.params.datatype_id,
        latest_only: req.query.latest_only,
        start_timestamp: ts1,
        end_timestamp: ts2,
        maxNum: req.query.maxNum
    };
    solution_req_handler.process_api(data, req, res);
});

/**
 * @swagger
 * /omr/api/v1/contracts/{contract_id}/ownerdata/{datatype_id}/downloadAsRequester:
 *   get:
 *     tags:
 *       - Owner Data
 *     description: "Download Owner Data by contract ID and data type. This is used by the data requester.  Data owner should use /omr/api/v1/services/{service_id}/ownerdata/{datatype_id}/download:"
 *     produces:
 *       - application/json
 *     parameters:
 *       - name: token
 *         in: header
 *         description: login token
 *         required: false
 *         type: string
 *         format: string
 *       - name: contract_id
 *         description: Contract ID
 *         in: path
 *         required: true
 *         type: string
 *       - name: datatype_id
 *         description: Data item defined in the service
 *         in: path
 *         required: true
 *         type: string
*       - name: start_timestamp
*         description: starting timestamp (inclusive). Cull for data uploaded before at specified time.  Can enter timestamp or date (mm/dd/yyyy)
*         in: query
*         required: false
*         type: integer
*         format: int64
*       - name: end_timestamp
*         description: end timestamp (inclusive). Cull for data uploaded after specified time.  Can enter timestamp or date (mm/dd/yyyy)
*         in: query
*         required: false
*         type: integer
*         format: int64
*       - name: latest_only
*         description: If true, start_timestamp, end_timestamp and maxNum are ignored and only latest data will be returned.
*         in: query
*         required: false
*         type: string
*         enum: [ "true", "false"]
*         default: "false"
*       - name: maxNum
*         description: Max number of results returned. If start and end timestamps are not specified, then the most recent results are returned.
*         in: query
*         required: false
*         type: string
*         default: 1000
 *     security:
 *       - basicAuth: []
 *     responses:
 *       200:
 *         description: List of owner data objects
 *         schema:
 *           type: array
 *           items:
 *             $ref: '#/definitions/OwnerData'
 */
router.route('/contracts/:contract_id/ownerdata/:datatype_id/downloadAsRequester').get(function (req, res) {
    if (!req.params.contract_id) {
        var errmsg = "contract_id";
        logger.error(errmsg);
        res.status(400).json({ msg: "Missing id: " + errmsg, status: 400 });

    } else if (!req.params.datatype_id) {
        var errmsg = "datatype_id";
        logger.error(errmsg);
        res.status(400).json({ msg: "Missing id: " + errmsg, status: 400 });
    }

    //support date format input for timestamp
    var ts1 = req.query.start_timestamp;
    if (!ts1) {
        ts1 = 0
    } else if (ts1 && typeof ts1 == 'string') {
        var match = ts1.match(/\D/g);
        if (match != null) {
            var tsdate = new Date(ts1);
            ts1 = Math.floor(tsdate.getTime() / 1000);
        }
        else {
            ts1 = parseInt(ts1, 10);
        }
    }
    var ts2 = req.query.end_timestamp;
    if (!ts2) {
        ts2 = 0
    } else if (ts2 && typeof ts2 == 'string') {
        var match = ts2.match(/\D/g);
        if (match != null) {
            var tsdate = new Date(ts2);
            ts2 = Math.floor(tsdate.getTime() / 1000);
        }
        else {
            ts2 = parseInt(ts2, 10);
        }
    }

    var data = {
        type: 'downloadOwnerDataAsRequester',
        contract_id: req.params.contract_id,
        datatype_id: req.params.datatype_id,
        latest_only: req.query.latest_only,
        start_timestamp: ts1,
        end_timestamp: ts2,
        maxNum: req.query.maxNum
    };
    solution_req_handler.process_api(data, req, res);
});

/**
* @swagger
* /omr/api/v1/services/{service_id}/owner_service/{owner_service_id}/datatypes/{datatype_id}/downloadWithConsent:
*   get:
*     tags:
*       - Owner Data
*     description: "Get Owner Data with consent from data owner."
*     produces:
*       - application/json
*     parameters:
*       - name: token
*         in: header
*         description: login token
*         required: false
*         type: string
*         format: string
*       - name: service_id
*         description: Target Service Id
*         in: path
*         required: true
*         type: string
*       - name: owner_service_id
*         description: Owner Service Id
*         in: path
*         required: true
*         type: string
*       - name: datatype_id
*         description: Datatype Id
*         in: path
*         required: true
*         type: string
*       - name: start_timestamp
*         description: starting timestamp (optional). Cull for data uploaded before at specified time.  Can enter timestamp or date (mm/dd/yyyy)
*         in: query
*         required: false
*         type: integer
*         format: int64
*       - name: end_timestamp
*         description: end timestamp (optional). Cull for data uploaded after specified time.  Can enter timestamp or date (mm/dd/yyyy)
*         in: query
*         required: false
*         type: integer
*         format: int64
*       - name: latest_only
*         description: If true, start_timestamp, end_timestamp and maxNum are ignored and only latest data will be returned.
*         in: query
*         required: false
*         type: string
*         enum: [ "true", "false"]
*         default: "false"
*       - name: maxNum
*         description: Max number of results returned. If start timestamp is not specified most resent results are returned. If 0 is given, maxNum will default to 1000.
*         in: query
*         required: false
*         type: string
*         default: 1000
*     security:
*       - basicAuth: []
*     responses:
*       200:
*         description: List of owner data objects
*         schema:
*           type: array
*           items:
*             $ref: '#/definitions/OwnerData'
*/
router.route('/services/:service_id/owner_service/:owner_service_id/datatypes/:datatype_id/downloadWithConsent').get(function (req, res) {
    if (!req.params.service_id) {
        var errmsg = "service_id";
        logger.error(errmsg);
        res.status(400).json({ msg: "Missing id: " + errmsg, status: 400 });
    } else if (!req.params.owner_service_id) {
        var errmsg = "owner_service_id";
        logger.error(errmsg);
        res.status(400).json({ msg: "Missing id: " + errmsg, status: 400 });
    } else if (!req.params.datatype_id) {
        var errmsg = "datatype_id";
        logger.error(errmsg);
        res.status(400).json({ msg: "Missing id: " + errmsg, status: 400 });
    } else {
        var ts1 = req.query.start_timestamp;
        if (!ts1) {
            ts1 = 0
        }

        var ts2 = req.query.end_timestamp;
        if (!ts2) {
            ts2 = 0
        }

        var data = {
            type: 'downloadOwnerDataWithConsent',
            service_id: req.params.service_id,
            owner_service_id: req.params.owner_service_id,
            datatype_id: req.params.datatype_id,
            latest_only: req.query.latest_only,
            start_timestamp: ts1,
            end_timestamp: ts2,
            maxNum: req.query.maxNum
        };
        solution_req_handler.process_api(data, req, res);
    }
});

/**
* @swagger
* /omr/api/v1/ownerdata/download/{consent_token}:
*   get:
*     tags:
*       - Owner Data
*     description: "Get Owner Data by Consent Validation Token. Token is only valid for 10 min, and is also only valid for the user who originally obtained it."
*     produces:
*       - application/json
*     parameters:
*       - name: token
*         in: header
*         description: login token
*         required: false
*         type: string
*         format: string
*       - name: consent_token
*         description: consent validaiton token
*         in: path
*         required: true
*         type: string
*       - name: start_timestamp
*         description: starting timestamp (optional). Cull for data uploaded before at specified time.  Can enter timestamp or date (mm/dd/yyyy)
*         in: query
*         required: false
*         type: integer
*         format: int64
*       - name: end_timestamp
*         description: end timestamp (optional). Cull for data uploaded after specified time.  Can enter timestamp or date (mm/dd/yyyy)
*         in: query
*         required: false
*         type: integer
*         format: int64
*       - name: latest_only
*         description: If true, start_timestamp, end_timestamp and maxNum are ignored and only latest data will be returned.
*         in: query
*         required: false
*         type: string
*         enum: [ "true", "false"]
*         default: "false"
*       - name: maxNum
*         description: Max number of results returned. If start timestamp is not specified most resent results are returned. If 0 is given, maxNum will default to 1000.
*         in: query
*         required: false
*         type: string
*         default: 1000
*     security:
*       - basicAuth: []
*     responses:
*       200:
*         description: List of owner data objects
*         schema:
*           type: array
*           items:
*             $ref: '#/definitions/OwnerData'
*/
router.route('/ownerdata/download/:consent_token').get(function (req, res) {
    if (!req.params.consent_token) {
        var errmsg = "consent token";
        logger.error(errmsg);
        res.status(400).json({ msg: "Missing: " + errmsg, status: 400 });
    } else {
        var ts1 = req.query.start_timestamp;
        if (!ts1) {
            ts1 = 0
        } else if (ts1 && typeof ts1 == 'string') {
            var match = ts1.match(/\D/g);
            if (match != null) {
                var tsdate = new Date(ts1);
                ts1 = Math.floor(tsdate.getTime() / 1000);
            }
            else {
                ts1 = parseInt(ts1, 10);
            }
        }

        var ts2 = req.query.end_timestamp;
        if (!ts2) {
            ts2 = 0
        } else if (ts2 && typeof ts2 == 'string') {
            var match = ts2.match(/\D/g);
            if (match != null) {
                var tsdate = new Date(ts2);
                ts2 = Math.floor(tsdate.getTime() / 1000);
            }
            else {
                ts2 = parseInt(ts2, 10);
            }
        }

        var data = {
            type: 'downloadOwnerDataConsentToken',
            latest_only: req.query.latest_only,
            start_timestamp: ts1,
            end_timestamp: ts2,
            maxNum: req.query.maxNum,
            token: req.params.consent_token
        };
        solution_req_handler.process_api(data, req, res);
    }
});

//=============================================
// Contract
//=============================================

/**
 * @swagger
 * /omr/api/v1/contracts:
 *   post:
 *     tags:
 *       - Contracts
 *     description: Create a new contract (Requesting owner data).  The owner or requester of the data can create the
 *                  contract.  If the requester creates the contract, the owner must use changeTerms to approve of the terms
 *                  before the contract can be signed (by the requester).  If the owner creates the contract, the requester
 *                  can sign the contract if they agree to the terms, otherwise the requester can use changeTerms to reset
 *                  the process (now owner must approve the new terms before the contract can be signed).
 *     produces:
 *       - application/json
 *     parameters:
 *       - name: token
 *         in: header
 *         description: login token
 *         required: false
 *         type: string
 *         format: string
 *       - name: ContractRequest
 *         description: "Contract Object<br><br>leave contract_term as {} to get the default value from the contract template (service). Or copy the terms from the contract template and modify if needed."
 *         in: body
 *         required: true
 *         schema:
 *           $ref: '#/definitions/ContractRequest'
 *     security:
 *       - basicAuth: []
 *     responses:
 *       200:
 *         description: Response object
 *         schema:
 *           type: array
 *           items:
 *             $ref: '#/definitions/ContractInvokeResponse'
 */
router.route('/contracts').post(function (req, res) {
    var serviceid = req.body.requester_service_id;
    var orgid = req.body.requester_org_id;
    var ownserviceid = req.body.owner_service_id;
    var ownorgid = req.body.owner_org_id;

    if (!serviceid) {
        var errmsg = "You missed a required element: requester_service_id"
        logger.error(errmsg);
        res.status(400).json({ msg: "Error:" + errmsg, status: 400 });
    } else if (!orgid) {
        var errmsg = "You missed a required element: requester_org_id"
        logger.error(errmsg);
        res.status(400).json({ msg: "Error:" + errmsg, status: 400 });
    } else if (!ownserviceid) {
        var errmsg = "You missed a required element: owner_service_id"
        logger.error(errmsg);
        res.status(400).json({ msg: "Error:" + errmsg, status: 400 });
    } else if (!ownorgid) {
        var errmsg = "You missed a required element: owner_org_id"
        logger.error(errmsg);
        res.status(400).json({ msg: "Error:" + errmsg, status: 400 });
    } else {
        var data = {
            type: 'createContract',
            requester_org_id: orgid,
            requester_service_id: serviceid,
            owner_org_id: ownorgid,
            owner_service_id: ownserviceid,
            terms: req.body.contract_terms || {},
            symkey: req.body.symkey ? req.body.symkey : null,
            symkey: req.body.logsymkey ? req.body.logsymkey : null
        };
        solution_req_handler.process_api(data, req, res);
    }
});

/**
 * @swagger
 * /omr/api/v1/contracts/{contract_id}:
 *   get:
 *     tags:
 *       - Contracts
 *     description: Returns a contract
 *     produces:
 *       - application/json
 *     parameters:
 *       - name: token
 *         in: header
 *         description: login token
 *         required: false
 *         type: string
 *         format: string
 *       - name: contract_id
 *         description: Contract ID
 *         in: path
 *         required: true
 *         type: string
 *     security:
 *       - basicAuth: []
 *     responses:
 *       200:
 *         description: Contract object
 *         schema:
 *           type: array
 *           items:
 *             $ref: '#/definitions/Contract'
 */
router.route('/contracts/:contract_id').get(function (req, res) {
    var data = {
        type: 'getContract',
        contract_id: req.params.contract_id
    };
    solution_req_handler.process_api(data, req, res);
});

/**
 * @swagger
 * /omr/api/v1/contracts/{contract_id}/changeTerms:
 *   post:
 *     tags:
 *       - Contracts
 *     description: Change contract terms by the owner or requester of the data.  If the requester changes the terms, the owner
 *                  can approve these changes by calling this and leaving terms {}.  Each time the requester changes the terms,
 *                  the contract state is reset so the owner needs to approve the changed terms before the requester can
 *                  sign the contract.
 *     produces:
 *       - application/json
 *     parameters:
 *       - name: token
 *         in: header
 *         description: login token
 *         required: false
 *         type: string
 *         format: string
 *       - name: contract_id
 *         description: Contract id
 *         in: path
 *         required: true
 *         type: string
 *       - name: Terms
 *         description: Contract Terms
 *         in: body
 *         required: true
 *         schema:
 *           $ref: '#/definitions/ContractTerms'
 *     security:
 *       - basicAuth: []
 *     responses:
 *       200:
 *         description: Response object
 *         schema:
 *           type: array
 *           items:
 *             $ref: '#/definitions/ContractInvokeResponse'
 */
router.route('/contracts/:contract_id/changeTerms').post(function (req, res) {
    if (!req.body.contract_id || req.body.contract_id != req.params.contract_id) {
        var errmsg = "Contract ID does not match";
        logger.error(errmsg);
        res.status(400).json({ msg: "Error:" + errmsg, status: 400 });
    }
    else if (!req.body.contract_terms || req.body.contract_terms == [] || req.body.contract_terms == {}) {
        var errmsg = "You need to add contract_terms";
        logger.error(errmsg);
        res.status(400).json({ msg: "Error:" + errmsg, status: 400 });
    }
    else {
        var data = {
            type: 'changeContractTerms',
            contract_id: req.params.contract_id,
            contract_terms: req.body.contract_terms
        };
        solution_req_handler.process_api(data, req, res);
    }
});

/**
 * @swagger
 * /omr/api/v1/contracts/{contract_id}/sign:
 *   post:
 *     tags:
 *       - Contracts
 *     description: Sign the contract by the requester
 *     produces:
 *       - application/json
 *     parameters:
 *       - name: token
 *         in: header
 *         description: login token
 *         required: false
 *         type: string
 *         format: string
 *       - name: contract_id
 *         description: Contract id
 *         in: path
 *         required: true
 *         type: string
 *       - name: Detail
 *         description: Sign Contract
 *         in: body
 *         required: true
 *         schema:
 *           $ref: '#/definitions/ContractSign'
 *     security:
 *       - basicAuth: []
 *     responses:
 *       200:
 *         description: Response object
 *         schema:
 *           type: array
 *           items:
 *             $ref: '#/definitions/ContractInvokeResponse'
 */
router.route('/contracts/:contract_id/sign').post(function (req, res) {
    if (!req.body.contract_id || req.body.contract_id != req.params.contract_id) {
        var errmsg = "Contract ID does not match"
        logger.error(errmsg);
        res.status(400).json({ msg: "Error:" + errmsg, status: 400 });
    } else {
        var data = {
            type: 'signContract',
            contract_id: req.params.contract_id,
            contract_terms: req.body.contract_terms || {}
        };
        solution_req_handler.process_api(data, req, res);
    }
});

/**
 * @swagger
 * /omr/api/v1/contracts/{contract_id}/terminate:
 *   post:
 *     tags:
 *       - Contracts
 *     description: Terminate the contract by the requester
 *     produces:
 *       - application/json
 *     parameters:
 *       - name: token
 *         in: header
 *         description: login token
 *         required: false
 *         type: string
 *         format: string
 *       - name: contract_id
 *         description: Contract id
 *         in: path
 *         required: true
 *         type: string
 *       - name: Detail
 *         description: Terminate Contract Detail
 *         in: body
 *         required: true
 *         schema:
 *           $ref: '#/definitions/ContractTerminate'
 *     security:
 *       - basicAuth: []
 *     responses:
 *       200:
 *         description: Response object
 *         schema:
 *           type: array
 *           items:
 *             $ref: '#/definitions/ContractInvokeResponse'
 */
router.route('/contracts/:contract_id/terminate').post(function (req, res) {
    if (!req.body.contract_id || req.body.contract_id != req.params.contract_id) {
        var errmsg = "Contract ID does not match";
        logger.error(errmsg);
        res.status(400).json({ msg: "Error:" + errmsg, status: 400 });
    } else {
        var data = {
            type: 'terminateContract',
            contract_id: req.params.contract_id,
            contract_terms: req.body.contract_terms || {}
        };
        solution_req_handler.process_api(data, req, res);
    }
});

/**
 * @swagger
 * /omr/api/v1/contracts/{contract_id}/payment:
 *   post:
 *     tags:
 *       - Contracts
 *     description: Payment of contract by the requester
 *     produces:
 *       - application/json
 *     parameters:
 *       - name: token
 *         in: header
 *         description: login token
 *         required: false
 *         type: string
 *         format: string
 *       - name: contract_id
 *         description: Contract id
 *         in: path
 *         required: true
 *         type: string
 *       - name: Detail
 *         description: Payment Contract Detail
 *         in: body
 *         required: true
 *         schema:
 *           $ref: '#/definitions/ContractPayment'
 *     security:
 *       - basicAuth: []
 *     responses:
 *       200:
 *         description: Response object
 *         schema:
 *           type: array
 *           items:
 *             $ref: '#/definitions/ContractInvokeResponse'
 */

router.route('/contracts/:contract_id/payment').post(function (req, res) {
    if (!req.body.contract_id || req.body.contract_id != req.params.contract_id) {
        var errmsg = "Contract ID does not match"
        logger.error(errmsg);
        res.status(400).json({ msg: "Error:" + errmsg, status: 400 });
    }
    else {
        var data = {
            type: 'payContract',
            contract_id: req.params.contract_id,
            contract_terms: req.body.contract_terms || {}
        };
        solution_req_handler.process_api(data, req, res);
    }
});


/**
 * @swagger
 * /omr/api/v1/contracts/{contract_id}/verify:
 *   post:
 *     tags:
 *       - Contracts
 *     description: Verify the payment of contract by the owner
 *     produces:
 *       - application/json
 *     parameters:
 *       - name: token
 *         in: header
 *         description: login token
 *         required: false
 *         type: string
 *         format: string
 *       - name: contract_id
 *         description: Contract id
 *         in: path
 *         required: true
 *         type: string
 *       - name: Detail
 *         description: Contract Payment Verification Detail
 *         in: body
 *         required: true
 *         schema:
 *           $ref: '#/definitions/ContractPaymentVerification'
 *     security:
 *       - basicAuth: []
 *     responses:
 *       200:
 *         description: Response object
 *         schema:
 *           type: array
 *           items:
 *             $ref: '#/definitions/ContractInvokeResponse'
 */
router.route('/contracts/:contract_id/verify').post(function (req, res) {
    if (!req.body.contract_id || req.body.contract_id != req.params.contract_id) {
        var errmsg = "Contract ID does not match"
        logger.error(errmsg);
        res.status(400).json({ msg: "Error:" + errmsg, status: 400 });
    }
    else {
        var data = {
            type: 'verifyContractPayment',
            contract_id: req.params.contract_id,
            contract_terms: req.body.contract_terms || {}
        };
        solution_req_handler.process_api(data, req, res);
    }
});

/**
 * @swagger
 * /omr/api/v1/contracts/{contract_id}/max_num_download/{max_num_download}/permission:
 *   post:
 *     tags:
 *       - Contracts
 *     description: Give permission to download to a contract
 *     produces:
 *       - application/json
 *     parameters:
 *       - name: token
 *         in: header
 *         description: login token
 *         required: false
 *         type: string
 *         format: string
 *       - name: contract_id
 *         description: Contract ID
 *         in: path
 *         required: true
 *         type: string
 *       - name: max_num_download
 *         description: Maximum number of downloads for this contract
 *         in: path
 *         type: integer
 *         required: true
 *       - name: datatype_id
 *         description: "datatype ID"
 *         in: query
 *         required: true
 *         type: string
 *         default: ""
 *     security:
 *       - basicAuth: []
 *     responses:
 *       200:
 *         description: Response object
 *         schema:
 *           type: array
 *           items:
 *             $ref: '#/definitions/ContractInvokeResponse'
 */
router.route('/contracts/:contract_id/max_num_download/:max_num_download/permission').post(function (req, res) {
    if (!req.params.contract_id) {
        var errmsg = "Contract ID missing"
        logger.error(errmsg);
        res.status(400).json({ msg: "Error:" + errmsg, status: 400 });
    } else if (!req.query.datatype_id) {
        var errmsg = "Datatype missing"
        logger.error(errmsg);
        res.status(400).json({ msg: "Error:" + errmsg, status: 400 });
    } else {
        var data = {
            type: 'givePermissionByContract',
            contract_id: req.params.contract_id,
            max_num_download: req.params.max_num_download,
            datatype: req.query.datatype_id
        };
        solution_req_handler.process_api(data, req, res);
    }
});

/**
 * @swagger
 * /omr/api/v1/contracts/owner/{owner_service_id}/status/{state}:
 *   get:
 *     tags:
 *       - Contracts
 *     description: Returns contracts for an owner, filtered by status
 *     produces:
 *       - application/json
 *     parameters:
 *       - name: token
 *         in: header
 *         description: login token
 *         required: false
 *         type: string
 *         format: string
 *       - name: owner_service_id
 *         description: Owner service id
 *         in: path
 *         required: true
 *         type: string
 *       - name: state
 *         description: Current status of contract
 *         in: path
 *         required: true
 *         type: string
 *         enum: ["*","requested", "contractReady", "contractSigned", "paymentDone", "paymentVerified", "downloadReady", "downloadDone", "terminated"]
 *     security:
 *       - basicAuth: []
 *     responses:
 *       200:
 *         description: Contract object
 *         schema:
 *           type: array
 *           items:
 *             $ref: '#/definitions/Contract'
 */
router.route('/contracts/owner/:owner_service_id/status/:state').get(function (req, res) {
    var state = req.params.state;
    if (state == '*') state = "";
    var data = {
        type: 'getOwnerContracts',
        service_id: req.params.owner_service_id,
        state: state
    };
    solution_req_handler.process_api(data, req, res);
});

/**
 * @swagger
 * /omr/api/v1/contracts/requester/{requester_service_id}/status/{state}:
 *   get:
 *     tags:
 *       - Contracts
 *     description: Returns contracts for a requester, filtered by status
 *     produces:
 *       - application/json
 *     parameters:
 *       - name: token
 *         in: header
 *         description: login token
 *         required: false
 *         type: string
 *         format: string
 *       - name: requester_service_id
 *         description: Requester service id
 *         in: path
 *         required: true
 *         type: string
 *       - name: state
 *         description: Current status of contract
 *         in: path
 *         required: true
 *         type: string
 *         enum: ["*", "requested", "contractReady", "contractSigned", "paymentDone", "paymentVerified", "downloadReady", "downloadDone", "terminated"]
 *     security:
 *       - basicAuth: []
 *     responses:
 *       200:
 *         description: Contract object
 *         schema:
 *           type: array
 *           items:
 *             $ref: '#/definitions/Contract'
 */
router.route('/contracts/requester/:requester_service_id/status/:state').get(function (req, res) {
    var state = req.params.state;
    if (state == '*') state = "";

    var data = {
        type: 'getRequesterContracts',
        service_id: req.params.requester_service_id,
        state: state
    };
    solution_req_handler.process_api(data, req, res);
});

//=============================================
// Logs
//=============================================

/**
* @swagger
* /omr/api/v1/history:
*   get:
*     tags:
*       - Logs
*     description: Must provide at least one of the following parameters -  patient_id, service_id, datatype_id, org_id, contract_id, or data.  The rest of the parameters are optional.
*     produces:
*       - application/json
*     parameters:
*       - name: token
*         in: header
*         description: login token
*         required: false
*         type: string
*         format: string
*       - name: contract_id
*         description: Contract ID
*         in: query
*         required: false
*         type: string
*       - name: datatype_id
*         description: Datatype ID
*         in: query
*         required: false
*         type: string
*       - name: patient_id
*         description: Patient ID where patient can be consent owner, consent target, data owner, and data target
*         in: query
*         required: false
*         type: string
*       - name: service_id
*         description: Service ID where service can be consent owner, consent target, data owner, data target, contract requester, or contract owner
*         in: query
*         required: false
*         type: string
*       - name: contract_org_id
*         description: Contract Org ID where org can be Contract Owner Org or Contract Requester Org.
*         in: query
*         required: false
*         type: string
*       - name: consent_owner_target_id
*         description: Consent Owner Target ID where consent owner can be Patient, Service or where Consent target is a Service.
*         in: query
*         required: false
*         type: string
*       - name: start_timestamp
*         description: starting timestamp (inclusive). Cull for data uploaded before at specified time.  Can enter timestamp or date (mm/dd/yyyy)
*         in: query
*         required: false
*         type: integer
*         format: int64
*       - name: end_timestamp
*         description: end timestamp (exclusive). Cull results data uploaded after specified time.  Can enter timestamp or date (mm/dd/yyyy)
*         in: query
*         required: false
*         type: integer
*         format: int64
*       - name: latest_only
*         description: If true, maxNum is ignored and only latest data is returned.  If start and end timestamp are also specified, they will not be ignored.
*         in: query
*         required: false
*         type: string
*         enum: [ "true", "false"]
*         default: "false"
*       - name: maxNum
*         description: Max number of results returned. If start and end timestamps are not specified, then the most recent results are returned.  If 0 is given maxNum will default to 20.
*         in: query
*         required: false
*         type: string
*         default: 20
*     security:
*       - basicAuth: []
*     responses:
*       200:
*         description: List of Log object
*         schema:
*           type: array
*           items:
*             $ref: '#/definitions/Log'
*/
router.route('/history').get(function (req, res) {
    const isSearchFieldsProvided = req.query.patient_id || req.query.service_id || req.query.contract_id ||
        req.query.datatype_id || req.query.contract_id || req.query.contract_org_id || req.query.consent_owner_target_id;

    if (!isSearchFieldsProvided) {
        var errmsg = "Must provide at least one of the fields";
        logger.error(errmsg);
        res.status(400).json({ msg: errmsg, status: 400 });
    } else {
        var log_type = (req.query.log_type == "All Types") ? "" : req.query.log_type

        //support date format input for timestamp
        var ts1 = req.query.start_timestamp;
        if (!ts1) {
            ts1 = 0
        } else if (ts1 && typeof ts1 == 'string') {
            var match = ts1.match(/\D/g);
            if (match != null) {
                var tsdate = new Date(ts1);
                ts1 = Math.floor(tsdate.getTime() / 1000);
            }
            else {
                ts1 = parseInt(ts1, 10);
            }
        }
        var ts2 = req.query.end_timestamp;
        if (!ts2) {
            ts2 = 0
        } else if (ts2 && typeof ts2 == 'string') {
            var match = ts2.match(/\D/g);
            if (match != null) {
                var tsdate = new Date(ts2);
                ts2 = Math.floor(tsdate.getTime() / 1000);
            }
            else {
                ts2 = parseInt(ts2, 10);
            }
        }

        var data = {
            type: 'getLogs',
            patient: req.query.patient_id || "",
            service: req.query.service_id || "",
            contract_org_id: req.query.contract_org_id || '',
            consent_owner_target_id: req.query.consent_owner_target_id || '',
            contract: req.query.contract_id || "",
            datatype: req.query.datatype_id || "",
            latest_only: req.query.latest_only || "",
            start_timestamp: ts1 || "",
            end_timestamp: ts2 || "",
            maxNum: req.query.maxNum || ""
        };

        solution_req_handler.process_api(data, req, res);
    }
});

router.get('*', function (req, res) {
    res.redirect("/api-docs");
});

module.exports = router;
