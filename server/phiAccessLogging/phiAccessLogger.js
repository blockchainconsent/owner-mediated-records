/*******************************************************************************
 *
 *
 * (c) Copyright Merative US L.P. and others 2020-2022 
 *
 * SPDX-Licence-Identifier: Apache 2.0
 *
 *******************************************************************************/

const log4js = require('log4js');
const { v4: uuidV4 } = require('uuid');

const hfc = require('fabric-client');
const { storeEvent } = require('./cadfEventStorage');

const cadfEventLoggingConfig = hfc.getConfigSetting('cadfEventLogging');

const log = log4js.getLogger('phiAccessLogger');
log.level = 'debug';

const successOutcome = 'success';
const failureOutcome = 'failure';

log.info(`PHI access logging is ${cadfEventLoggingConfig.enabled ? 'enabled' : 'disabled'}`);

/**
 * Create CADF event frame.
 * Log record should help to understand 'who', 'what', 'when' accessed the PHI data.
 */
const createCadfEvent = (initiatorId, targetId, targetName, outcome, httpStatusCode, resourceTypeAction, message, requestData) => {
  return {
    id: uuidV4(),
    typeURI: 'http://schemas.dmtf.org/cloud/audit/1.0/event',
    eventTime: new Date().toISOString(),
    eventType: 'activity',
    severity: 'normal',
    /**
     * Action attempted by the Initiator.
     */
    action: `omr-app.${resourceTypeAction}.read`,
    /**
     * Outcome of the action. Must be either "success" or "failure".
     */
    outcome,
    reason: {
      reasonType: 'HTTP',
      reasonCode: httpStatusCode,
    },
    /**
     * Provide a user-friendly description of the event.
     */
    message,
    /**
     * Should be set to "true" for data access events and to "false" for other
     * (e.g. management and state change) events.
     * Always true since we are monitor requests for PHI data.
     */
    dataEvent: true,
    /**
     * The entity that attempts to perform the Action upon the Target.
     * Initiator contains information about the 'who'.
     */
    initiator: {
      /**
       * User ID of user who requests PHI.
       */
      id: initiatorId,
      typeURI: 'data/security/account/user',
      credential: {
        type: 'token',
      },
    },
    /**
     * Target describes the 'what' and contains information about requested PHI data.
     */
    target: {
      /**
       * A universally unique id. This ID must remain consistent over time.
       * Different event records must not provide different IDs for the same underlying resource.
       */
      id: targetId,
      typeURI: 'data/storage/userdata',
      name: targetName,
    },
    observer: {
      id: 'owner-mediated-records',
      typeURI: 'service/compute/servers',
      name: 'owner-mediated-records',
    },
    /**
     * An object (key-valye map) containing arbitrary details about the request sent by Initiator.
     */
    requestData: requestData || {},
  };
};

const logCadfEvent = (phiAccessDetails, statusCode, outcome) => {
  const {initiatorId, targetId, targetName, action, message, requestData} = phiAccessDetails;
  const cadfEvent = createCadfEvent(initiatorId, targetId, targetName, outcome, statusCode, action, message, requestData);

  storeEvent(cadfEvent).catch(e => {
    log.error('Failed to log PHI access event.', e)
  });
};

module.exports.logPhiAccess = (phiAccessDetails, statusCode) => {
  if (!cadfEventLoggingConfig.enabled) return;

  log.debug('Log successful PHI access action.');
  logCadfEvent(phiAccessDetails, successOutcome, statusCode);
};

module.exports.logFailedPhiAccess = (phiAccessDetails, statusCode) => {
  if (!cadfEventLoggingConfig.enabled) return;

  log.debug('Log failed PHI access action.');
  logCadfEvent(phiAccessDetails, failureOutcome, statusCode)
};
