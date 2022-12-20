/*******************************************************************************
 * Some helpful functions for tests
 *******************************************************************************/

const { v4: uuidv4 } = require('uuid');

// generate random string with alphanumeric characters only
let idGenerator = function idGenerator() {
    return uuidv4().replace(/[^0-9a-z]/g, '');
}

module.exports = idGenerator;