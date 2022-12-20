/*******************************************************************************
 *
 *
 * (c) Copyright Merative US L.P. and others 2020-2022 
 *
 * SPDX-Licence-Identifier: Apache 2.0
 *
 *******************************************************************************/

/*
 * GET home page.
 */

var express = require('express');
var router = express.Router();
var fs = require("fs");
var path = require('path');
var util = require('util');

// Load our modules.
var user_manager = require('common-utils/user_manager.js');
var req_handler = require('common-utils/request_handler.js');
var solution_api_handler = require('../solution/solution_api_handler');
//var omr_req_handler = require('../utils1/request_handler.js');
var ums = require('common-utils/ums.js');

// Use tags to make logs easier to find
var TAG = "site_router.js";
var log4js = require('log4js');
var logger = log4js.getLogger(TAG);
logger.level = 'DEBUG';

var hfc = require('fabric-client');
var config_path = path.join(process.env.config_file);
var setup = require(config_path);

// ============================================================================================================================
// Home
// ============================================================================================================================

router.use(function (req, res, next) {
	logger.debug(req.path);
	logger.debug(req.headers);

	var authorization = req.headers['authorization'];
	var b64auth = (authorization || '').split(' ')[1] || ':';
	var basicAuth = new Buffer(b64auth, 'base64').toString().split(':');
	logger.debug(basicAuth);
	if (!req.path.startsWith("/omr/api/") || !req.path.startsWith("/common/api/")) {
		req.session.error_msg = req.session.error_msg ? req.query.error_msg : null;
		req.session.success_msg = req.query.success_msg ? req.query.success_msg : null;
		req.session.login_error_msg = req.query.login_error_msg ? req.query.login_error_msg : null;
		req.session.reg_error_msg = req.query.reg_error_msg ? req.query.reg_error_msg : null;
		req.session.change_password_url = ums.getChangePasswordUrl();
		req.session.logout_url = ums.getLogoutUrl();
		req.session.login_url = ums.getLoginUrl();
		ums.loginUserByHeaders(req, res, next);
	} else {
		next();
	}

});

router.get('/', function (req, res) {
	res.redirect("/api-docs");
});

module.exports = router;
