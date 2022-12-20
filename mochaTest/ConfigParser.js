const path = require('path');
const yaml = require('js-yaml');
const fs = require('fs-extra');


const configFile = process.env.MOCHA_CM_CONFIG_JSON || path.join(__dirname, '../server/config','config.json');
const configJsonData = fs.readFileSync(configFile);
const configJson = JSON.parse(configJsonData);

const solutionFilePath = process.env.MOCHA_CM_SOLUTION || path.join(__dirname, '..', configJson['solution_config_file']);
const fileData = fs.readFileSync(solutionFilePath);
const fileExt = path.extname(solutionFilePath);

let solutionConfig;
if(fileExt.indexOf('y') > -1) {
    solutionConfig = yaml.safeLoad(fileData);
} else {
    solutionConfig = JSON.parse(fileData);
}

let protocol = 'http';
if (configJson['enable_https']) {
	protocol = 'https';
}
protocol = process.env.MOCHA_CM_PROTOCOL || protocol;
let host = process.env.MOCHA_CM_HOST || configJson['host'] || 'localhost';
let port = process.env.MOCHA_CM_PORT || configJson['port'] || '3000';


const appAdmin = getAppAdmJson(solutionConfig);
if (!appAdmin) {
	throw new Error("Unable to find app_admin field in solution config file.");
}
const channelsJson = getChannelsJson(solutionConfig);
if (!channelsJson) {
	throw new Error("Unable to find channels field in solution config file." );
}

//accessing api
exports.server = protocol + "://" + host + ":" + port + "/";
exports.channel = Object.keys(channelsJson)[0] || 'mychannel';
exports.caOrg = appAdmin['org'] || 'Org1';
//Sys Admin
exports.sysAdminId = appAdmin['username'] || 'AppAdmin';
exports.sysAdminPass = appAdmin['secret'] || 'pass0';


function getOmrJson(solutionConfig) {
	let solutions = solutionConfig['solutions'];
	return solutions && solutions['owner-mediated-records'] ? solutions['owner-mediated-records'] : null;
}
function getChannelsJson(solutionConfig) {
    let omr = getOmrJson(solutionConfig);
    return omr && omr['channels'] ? omr['channels'] : null;
}

function getAppAdmJson(solutionConfig) {
    let omr = getOmrJson(solutionConfig);
    return omr && omr['app_admin'] ? omr['app_admin'] : null;
}
