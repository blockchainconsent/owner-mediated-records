/*******************************************************************************
 *
 *
 * (c) Copyright Merative US L.P. and others 2020-2022 
 *
 * SPDX-Licence-Identifier: Apache 2.0
 *
 *******************************************************************************/

// Swagger_filter hides the specified API endpoint operations from being 
// displayed in the Swagger UI by removing all Swagger documentation for the API
// endpoint operations with the specified tags from the passed-in Swagger doc.

const recurse = require('reftools/lib/recurse.js').recurse;
const clone = require('reftools/lib/clone.js').clone;

function filter(obj,tags) {

    let src = clone(obj);
    let filteredpaths = [];

    recurse(src,{},function(obj,key,state){
        for (let tag of tags) {
            if (obj[key] && obj[key].tags && Array.isArray(obj[key].tags) && obj[key].tags.includes(tag)) {
                filteredpaths.push(state.path);
                delete obj[key];
                break;
            }
        }
    });

    // remove undefined properties (important for YAML output)
    recurse(src,{},function(obj,key,state){
        if (Array.isArray(obj[key])) {
            obj[key] = obj[key].filter(function(e){
                return typeof e !== 'undefined';
            });
        }
    });

    recurse(src,{},function(obj,key,state){
        if (obj.hasOwnProperty('$ref') && filteredpaths.includes(obj.$ref)) {
            if (Array.isArray(state.parent)) {
                state.parent.splice(state.pkey, 1);
            }
        }
    });

    // tidy up any paths where we have removed all the operations
    for (let p in src.paths) {
        if (Object.keys(src.paths[p]).length === 0) {
            delete src.paths[p];
        }
    }

    return src;
}

module.exports = {
    filter : filter
};