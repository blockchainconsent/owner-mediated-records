const ConfigParser = require('./ConfigParser.js');
exports.app = require('../server/server'); // start the app server

//accessing api
exports.server = ConfigParser.server;
exports.channel = ConfigParser.channel;
exports.caOrg = ConfigParser.caOrg;
//Sys Admin
exports.sysAdminId = ConfigParser.sysAdminId; 
exports.sysAdminPass = ConfigParser.sysAdminPass;

//testRun returns the last 5 digits of epoch time stamp
const testRun = ((Math.floor(Date.now())).toString());
exports.testRun = testRun;

exports.org1 = {
    "id": "org1ID",
    "secret": "pass1",
    "name": "org1",
    "ca_org": exports.caOrg,
    "email": "org1@email.com",
    "status": "active"
};

exports.orgAdminToken1 = "";

exports.org2 = {
  "id": "org2ID",
  "secret": "pass2",
  "name": "org2",
  "ca_org": exports.caOrg,
  "email": "org2@email.com",
  "status": "active"
};

exports.orgAdminToken2 = "";


exports.sysAdminToken = "";
exports.BasicAuthSysAdmin = "Basic " + new Buffer.from(exports.sysAdminId+"/"+ exports.caOrg +"/"+exports.channel + ":" + exports.sysAdminPass).toString("base64");
const systemAdminId2 = 'systemadmin'+testRun;
exports.systemAdminId2 = systemAdminId2;

//Datatypes
exports.datatype1 = {
  "id": "data1",
  "description": "sample data 1"
}

exports.datatype2 = {
  "id": "data2",
  "description": "sample data 2"
}

exports.datatype3 = {
  "id": "data3",
  "description": "sample data 3"
}

exports.datatype4 = {
  "id": "data4",
  "description": "sample data 4"
}
//Services
exports.service1 = {
  "id": "service1ID",
  "name": "service1",
  "secret": "service1pass",
  "ca_org": exports.caOrg,
  "email": "service1email@example.com",
  "org_id": this.org1.id,
  "summary": "Service 1 under org 1. Has one datatype",
  "terms": {
    "term1" : "example term",
    "term2" : "example term",
  },
  "payment_required": "yes",
  "datatypes": [
      {
          "datatype_id": this.datatype1.id,
          "access":[
              "write",
              "read"
          ]
      }
  ],
  "solution_private_data": {
    "sample data 1": "service 3 sample data 1",
    "sample data 2": "service 3 sample data 2"
  }
};

exports.service2 = {
  "id": "service2ID",
  "name": "service2",
  "secret": "service2pass",
  "ca_org": exports.caOrg,
  "email": "service2email@example.com",
  "org_id": this.org1.id,
  "summary": "Service 2 under org 1. Has multiple datatype",
  "terms": {
    "term1" : "example term",
    "term2" : "example term",
  },
  "payment_required": "yes",
  "datatypes": [
      {
          "datatype_id": this.datatype1.id,
          "access":[
              "write",
              "read"
          ]
      },

      {
        "datatype_id": this.datatype2.id,
        "access":[
            "write",
            "read"
        ]
    }

  ],
  "solution_private_data": {
    "sample data 1": "service 2 sample data 1",
    "sample data 2": "service 2 sample data 2"
  }
};

exports.service3 = {
  "id": "service3ID",
  "name": "service3",
  "secret": "service3pass",
  "ca_org": exports.caOrg,
  "email": "service3email@example.com",
  "org_id": this.org2.id,
  "summary": "Service 3 under org 2. Has multiple datatype",

  "terms": {
    "term1" : "service 3: example term",
    "term2" : "service 3: example term",
  },
  "payment_required": "yes",
  "datatypes": [
      {
          "datatype_id": this.datatype1.id,
          "access":[
              "write",
              "read"
          ]
      },

      {
        "datatype_id": this.datatype2.id,
        "access":[
            "write",
            "read"
        ]
    }

  ],
  "solution_private_data": {
      "sample data 1": "service 3 sample data 1",
      "sample data 2": "service 3 sample data 2"
  }
};

//not part of any org
exports.user1 = {
  "id": "user1",
  "secret": "user1pass",
  "name": "User 1",
  "role": "user",
  "email":  "user1email@example.com",
  "ca_org": exports.caOrg,
  "data": {
    "address": "1 User St"
  }
};

exports.user1Token = "";

// register with org1 admin
exports.user2 = {
  "id": "user2",
  "secret": "user2pass",
  "name": "User 2",
  "role": "user",
  "org" : this.org1.id,
  "email":  "user2email@example.com",
  "ca_org": exports.caOrg,
  "data": {
    "address": "2 User St"
  }
};

exports.user2Token = "";

exports.newServiceOneID = "";
exports.newServiceTwoID = "";
