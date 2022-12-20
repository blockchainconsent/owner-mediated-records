/*******************************************************************************
 *
 *
 * (c) Copyright Merative US L.P. and others 2020-2022 
 *
 * SPDX-Licence-Identifier: Apache 2.0
 *
 *******************************************************************************/

const log4js = require('log4js');
const { address } = require('ip');

const hfc = require('fabric-client');
const CloudantHelper = require('common-utils/cloudant_helper');

const cadfEventLoggingConfig = hfc.getConfigSetting('cadfEventLogging');

const log = log4js.getLogger('cadfEventStorage');
log.level = 'debug';

const createCloudantHelper = () =>  {
  const account = process.env.CLOUDANT_PHI_LOG_ACCOUNT || 'admin';
  const password = process.env.CLOUDANT_PHI_LOG_PASSWORD || 'pass';
  const url= process.env.CLOUDANT_PHI_LOG_URL || `http://${address()}:9080`;
  const dbName = process.env.CLOUDANT_PHI_LOG_DBNAME || 'phi-access-cadf-events';

  const cloudantHelper = new CloudantHelper(account, password, dbName, url);

  cloudantHelper.cloudantDB.server.db.list().then(dbList => {
    if (dbList.includes(dbName)) {
      log.debug(`PHI Access Logging database "${dbName}" exists`);
    } else {
      log.debug(`Create PHI Access Logging database "${dbName}"`);
      return cloudantHelper.cloudantDB.server.db.create(dbName);
    }
  }).catch(err => {
    log.error('Failed to create PHI Access Logging database', err);
  });

  return cloudantHelper;
};

let cloudantHelper;

if (cadfEventLoggingConfig.enabled) {
  cloudantHelper = createCloudantHelper();
}

module.exports.storeEvent = async (event) => {
  log.debug('Save CADF event to database.');
  return cloudantHelper.createDocument(event.id, event);
}
