/*******************************************************************************
 *
 *
 * (c) Copyright Merative US L.P. and others 2020-2022 
 *
 * SPDX-Licence-Identifier: Apache 2.0
 *
 *******************************************************************************/

const Redis = require('ioredis');



const client = new Redis(process.argv[2]);
console.log("Redis: Start flushdb");
client.flushdb( function (err, succeeded) {
    console.log(succeeded); // will be true if successfull
    client.disconnect();
});

