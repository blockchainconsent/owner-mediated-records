/*******************************************************************************
 *
 *
 * (c) Copyright Merative US L.P. and others 2020-2022 
 *
 * SPDX-Licence-Identifier: Apache 2.0
 *
 *******************************************************************************/

'use strict';
/*
 * config.js is the wrapper for the main configuration files for the app
 * It identifies two main configuration files within ./config folder:
 * config.json - app level configurations
 * network-config.json - chaincode network configuration
 *
 *
 */

var TAG = 'server.js:';
var log4js = require('log4js');
var logger = log4js.getLogger(TAG);
logger.level = 'DEBUG';

// =====================================================================================================================
//                                                 Node.js Setup
// =====================================================================================================================
var express = require('express');
var app = express();
var session = require('cookie-session');
var compression = require('compression');
var swaggerJSDoc = require('swagger-jsdoc');
var serve_static = require('serve-static');
var path = require('path');
var morgan = require('morgan');
var cookieParser = require('cookie-parser');
var bodyParser = require('body-parser');
var http = require('http');
var https = require('https');
var cors = require('cors');
var fs = require('fs');
var ip = require("ip");
var swaggerUI = require('swagger-ui-express');
var swaggerFilter = require('./swagger_filter.js');

require('./config.js');
// UNCOMMENT THIS LINE BELOW IF CONNECTING TO IBP
// require('./config_enterprise.js');
var hfc = require('fabric-client');

// Things that require the network to be set up
var chainHelper = require('common-utils/chain_helper');
var userManager = require('common-utils/user_manager');
var kms = require('common-utils/kms');
var ums = require('common-utils/ums');

var ws_handler = require('./solution/ws_handler');

logger.info('------------------------------------------------');
logger.info(' SERVER INIT');
logger.info('------------------------------------------------');

//Fix for the SDK.  Need to make sure a `/tmp` directory exists to tarball chaincode
//TODO: check if this is neede for v1.0
try {
    if (!fs.existsSync('/tmp')) {
        logger.info('No /tmp directory. Creating /tmp directory');
        fs.mkdirSync('/tmp');
    }
} catch (err) {
    logger.error('Error creating /tmp directory for chaincode:', err);
}

// crate tmp
var appTmpDir = hfc.getConfigSetting('tmp_dir');
try {
    if (!fs.existsSync(appTmpDir)) {
        logger.info('No tmp directory. Creating tmp directory');
        fs.mkdirSync(appTmpDir);
    }
} catch (err) {
    logger.error('Error creating tmp directory for chaincode:', err);
}

const deIdentifierConfig = hfc.getConfigSetting('de_identifier');
const isDevMode = hfc.getConfigSetting('dev_mode');

if (!isDevMode && !deIdentifierConfig.enabled) {
    logger.fatal('PII/PHI de-identification is disabled on environment without enabled dev_mode');
    logger.fatal('------------------------------------------------');
    logger.fatal(' SERVER FAILED TO START SUCCESSFULLY');
    logger.fatal('------------------------------------------------');
    process.exit(1);
}

//=====================================================================================================================
//                                                 Default Datastore Setup
//=====================================================================================================================
// Set up default datastore env variables
// This needs to be done before `chainHelper.setup()` in order to correctly populate values in solutions.yaml
// If using a custom datastore, update these values as appropriate or set the env variables directly in the system
if (!process.env.CLOUDANT_USERNAME || process.env.CLOUDANT_USERNAME == 'undefined') {
    process.env.CLOUDANT_USERNAME = 'admin';
}
if (!process.env.CLOUDANT_PASSWORD || process.env.CLOUDANT_PASSWORD == 'undefined') {
    process.env.CLOUDANT_PASSWORD = 'pass';
}
if (!process.env.CLOUDANT_DATABASE || process.env.CLOUDANT_DATABASE == 'undefined') {
    process.env.CLOUDANT_DATABASE = 'owner_mediated_records';
}
if (!process.env.CLOUDANT_HOST || process.env.CLOUDANT_HOST == 'undefined') {
    process.env.CLOUDANT_HOST = 'http://' + ip.address() + ':9080';
}

// invoke setup for offchain-datastore
async function setupDatastore() {
    let solutionConfig = defaultSolutionConfig["solutions"]["owner-mediated-records"];
    let chaincodeConfig = solutionConfig["chaincode"];

    let datastoreConfig = chaincodeConfig["datastore_config"];
    if (!datastoreConfig) {
        logger.info("Skipping datastore setup - datastore credentials not provided")
        return;
    }

    let connectionID = datastoreConfig["connection_id"] || "";
    let username = datastoreConfig["username"] || "";
    let password = datastoreConfig["password"] || "";
    let database = datastoreConfig["database"] || "";
    let host = datastoreConfig["host"] || "";

    // call setup with AppAdmin
    let adminConfig = solutionConfig["app_admin"];
    let adminOrg = adminConfig["org"];
    let adminID = adminConfig["username"];
    let adminSecret = adminConfig["secret"];
    let adminClient = await chainHelper.getClientForOrg(adminOrg, adminID, adminSecret);

    let chaincodeName = chaincodeConfig["name"];
    let fcn = "setupDatastore";
    let args = [connectionID, username, password, database, host];

    let channelConfig = solutionConfig["channels"];
    for (let channelName in channelConfig) {
        logger.info("Invoke setupDatastore for channel: " + channelName);
        await chainHelper.invoke(adminID, adminSecret, channelName, chaincodeName, fcn, args, null, adminOrg, adminClient);
    }

    logger.info("Invoke setupDatastore successful");
    server.emit('server_started');
}

// initialize util libraries
kms.setup();
ums.setup();
chainHelper.setup();
userManager.setup();

//=====================================================================================================================
//                                                SWAGGER Setup
//=====================================================================================================================
logger.info('Configuring Swagger');
//swagger definition
const swagger_host = hfc.getConfigSetting('swagger_host');
const swagger_scheme = hfc.getConfigSetting('swagger_scheme');
logger.info("swagger_host : " + swagger_host);
var swaggerDefinition = {
    info: {
        title: 'Watson Health Blockchain owner-mediated-records API',
        version: '4.6.0',
        description: 'RESTful APIs for the WH Blockchain owner-mediated-records',
    },
    host: swagger_host,
    schemes: [swagger_scheme],
    basePath: '/',
};

var defaultSolutionConfig = chainHelper.solutionConfig();
logger.info(defaultSolutionConfig);

var swaggerApiDocs = [];

// add common api routes to swagger
let swagger_common_option = "../node_modules/common-utils/routes/common_rest_api.js";
swaggerApiDocs.push(swagger_common_option);

for (let solutionName in defaultSolutionConfig["solutions"]) {
    logger.info("Setting Router for Solution: " + solutionName);
    let solutionConfig = defaultSolutionConfig["solutions"][solutionName];

    let rest_api_routes = solutionConfig["solution_api_routes"];
    swaggerApiDocs.push(rest_api_routes + ".js");
}

logger.debug("Swagger API Docs:", swaggerApiDocs)

// options for the swagger docs
var options = {
    // import swaggerDefinitions
    swaggerDefinition: swaggerDefinition,
    // path to the API docs
    apis: swaggerApiDocs
};

// initialize swagger-jsdoc
var hiddenSwaggerTags = ['Organizations', 'Users', 'MFA', 'Network Connection Profile', 'Sign'];
var swaggerSpec = swaggerFilter.filter(swaggerJSDoc(options), hiddenSwaggerTags);

//serve swagger
app.use('/api-docs', swaggerUI.serve, swaggerUI.setup(swaggerSpec));

//=====================================================================================================================
//                                                Express Setup
//=====================================================================================================================
logger.info('Configuring Express app');

// ADDITION TO SOLUTION TEMPLATE
// app.set('views', path.join(__dirname, '../src/views'));
// app.set('view engine', 'pug');
// app.engine('.html', require('pug').__express);
// ADDITION TO SOLUTION TEMPLATE

app.use(compression());
app.use(morgan('dev'));
app.use(bodyParser.json());
app.use(bodyParser.urlencoded({ extended: true }));
app.use(cookieParser());

// static folders
// COMMENTED OUT SOLUTION TEMPLATE CODE
// app.use(express.static('../build'));
// app.use(express.static('../build/stylesheets/'));
// END COMMENT

// ADDITION TO SOLUTION TEMPLATE
app.use(express.static('../public'));
app.use(express.static('../public/stylesheets/'));
// ADDITION TO SOLUTION TEMPLATE

var maxAge = parseInt(hfc.getConfigSetting('sessionMaxAge'));
// These two variables define whether or not to secure the cookie, which we should.
// and the depth of the proxy "bounce". That is found by running traceroute <ip of one of the albs> and adding one.

var cookieSecure = hfc.getConfigSetting('cookieSecure');
var proxyDepth = parseInt(hfc.getConfigSetting('proxyDepth'));

logger.info('Session MaxAge: ', maxAge);
app.use(session({
    secret: 'Somethignsomething1234!test',
    maxAge: maxAge,
    cookie: {
        httpOnly: true,
        secure: cookieSecure
    }
}));
app.set("trust proxy", proxyDepth);

//Enable CORS preflight across the board so browser will let the app make REST requests
app.options('*', cors());
app.use(cors());

// passing information back to the client
app.use(function (req, res, next) {
    logger.info('------------------------------------------------');
    logger.info(' incoming request');
    logger.info('------------------------------------------------');
    logger.info('New ' + req.method + ' request for', req.url);
    req.bag = {};
    req.bag.session = req.session;
    logger.debug("SESSION: " + JSON.stringify(req.session));
    next();
});

// router
var router = require('./routes/site_router');
app.use('/', router);

// initialize solution
// routes and initializing request_handlder
for (let solutionName in defaultSolutionConfig["solutions"]) {
    logger.info("Setting Router and Request Handler for Solution: " + solutionName);
    let solutionConfig = defaultSolutionConfig["solutions"][solutionName];

    // setup common request handler
    let common_req_handler_path = solutionConfig["common_request_handler"];
    let common_req_handler = require(common_req_handler_path);
    common_req_handler.setup(solutionConfig);

    // setup common routes
    let common_api_routes = solutionConfig["common_api_routes"];
    let common_restapi = require(common_api_routes);
    app.use(common_restapi.common_api_base, common_restapi.router);

    // set up solution routes
    let rest_api_path = solutionConfig["rest_api_path"];
    let rest_api_routes = solutionConfig["solution_api_routes"];
    var restapi = require(rest_api_routes);
    app.use(rest_api_path, restapi);
    logger.debug("route " + rest_api_path + ": " + rest_api_routes);
}

//If the request is not process by this point, their are 2 possibilities:
//1. We don't have a route for handling the request
app.use(function (req, res, next) {
    var err = new Error('Not Found');
    err.status = 404;
    next(err);
});
//2. Something else went wrong
app.use(function (err, req, res, next) { // = development error handler, print
    // stack trace
    logger.error('Error Handler -', req.url);
    var errorCode = err.status || 500;
    res.status(errorCode);
    logger.error(err);
    var errorMsg = {
        msg: err.message,
        status: errorCode
    };

    logger.error(errorMsg);
    res.json(errorMsg);
});

//=====================================================================================================================
//                                                Launch Webserver
//=====================================================================================================================
// server.key and server.crt are required for HTTPS
// these certs can be created by the following steps with openssl:
//
//openssl genrsa -out key.pem
//openssl req -new -key key.pem -out csr.pem
//openssl x509 -req -days 9999 -in csr.pem -signkey key.pem -out cert.pem
//rm csr.pem

//openssl genrsa -des3 -out server.key 1024
//openssl req -new -key server.key -out server.csr
//cp server.key server.key.org
//openssl rsa -in server.key.org -out server.key
//openssl x509 -req -days 365 -in server.csr -signkey server.key -out server.crt
//rm -rf server.key.org
//rm -rf server.csr
/////////////////////////////////////////////////////////////////////////

var host = hfc.getConfigSetting('host');
var port = parseInt(hfc.getConfigSetting('port'));
var server = null;
//var server = http.createServer(app).listen(3000, function() { });

if (hfc.getConfigSetting('enable_https')) {
    var options = {
        key: fs.readFileSync('server.key'),
        cert: fs.readFileSync('server.crt')
    };
    logger.info("HTTPS enabled");
    logger.info('Staring https server on: ' + host + ':' + port);
    server = https.createServer(options, app).listen(port, function () {
        logger.info('------------------------------------------------');
        logger.info(' Server Up - ' + host + ':' + port);
        logger.info('------------------------------------------------');
    });

}
else {
    logger.info('Staring http server on: ' + host + ':' + port);
    server = http.createServer(app).listen(port, function () {
        logger.info('------------------------------------------------');
        logger.info(' Server Up - ' + host + ':' + port);
        logger.info('------------------------------------------------');
    });
}

// some setting needed for some node modeule
process.env.NODE_TLS_REJECT_UNAUTHORIZED = '0';
process.env.NODE_ENV = 'production';
server.timeout = 240000;

var ws = require('ws');
var wss = {};

// =====================================================================================================================
//                                                     Blockchain Setup
// =====================================================================================================================
logger.info('configuring the chain object and its dependencies');
logger.info('------------------------------------------------');
logger.info(' Blockchain Setup');
logger.info('------------------------------------------------');

// stup blockchian network and deploy chaincode
chainHelper.setupChain().then(() => {
    logger.info('------------------------------------------------');
    logger.info(' SERVER READY');
    logger.info('------------------------------------------------');
    setupDatastore();
    start_websocket_server();
}, (err) => {
    logger.error(err);
    logger.error('------------------------------------------------');
    logger.error(' SERVER FAILED TO START SUCCESSFULLY');
    logger.error('------------------------------------------------');
    server.close(function () {
        logger.error("server closed");
        process.exit(1);
    });
});

//=====================================================================================================================
//                                            WebSocket Communication Madness
//=====================================================================================================================

function start_websocket_server(err, registrarCred, chaincodeHelper) {
    if (err != null) {
        //look at tutorial_part1.md in the trouble shooting section for help
        logger.error('! looks like the final configuration failed, holding off on the starting the socket\n', err);
        if (!process.error) process.error = { type: 'deploy', msg: err };
    }
    else {
        logger.info('------------------------------------------------');
        logger.info(' Websocket Up');
        logger.info('------------------------------------------------');
        wss = new ws.Server({ server: server });
        //start the websocket now
        wss.on('connection', function connection(ws) {
            ws.on('message', function incoming(message) {
                logger.info('------------------------------------------------');
                logger.info(' incoming ws message');
                logger.info('------------------------------------------------');
                try {
                    var data = JSON.parse(message);
                    logger.info('ws message:', data.type);
                    ws_handler.process_msg(ws, data);
                } catch (err) {
                    logger.error('ws message error', err);
                }
            });
            // Pass in wss to handlers to allow them to broadcast
            ws_handler.wss = wss;

            ws.on('error', function (err) {
                logger.error('ws error', err);
            });
            ws.on('close', function () {
                logger.info('ws closed');
            });
        });

        // This makes it easier to contact our clients
        wss.broadcast = function broadcast(data) {
            wss.clients.forEach(function each(client) {
                try {
                    data.v = '2';
                    client.send(JSON.stringify(data));
                }
                catch (err) {
                    logger.error('error broadcast ws', err);
                }
            });
        };
    }
}

module.exports = server;