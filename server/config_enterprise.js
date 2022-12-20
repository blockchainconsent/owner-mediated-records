/*******************************************************************************
 * 
 * 
 * (c) Copyright Merative US L.P. and others 2020-2022 
 *
 * SPDX-Licence-Identifier: Apache 2.0
 * 
 *******************************************************************************/

var path = require('path');
var hfc = require('fabric-client');

var config_file = path.join(__dirname, 'config','config_enterprise.json');
hfc.addConfigFile(config_file);

process.env.config_file = config_file;

if ( 
        !hfc.getConfigSetting('host')||
        !hfc.getConfigSetting('port') ||
        !hfc.getConfigSetting('swagger_host') ||
        !hfc.getConfigSetting('kms_module') ||
        !hfc.getConfigSetting('ums_module') ||
        !hfc.getConfigSetting('network_config_file') ||
        !hfc.getConfigSetting('solution_config_file') ||
        !hfc.getConfigSetting('key_value_store_module') 
    ) {
    throw new Error('Missing config item in config.js');
}
