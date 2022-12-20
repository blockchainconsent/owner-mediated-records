const Config = require('../ConfigFile.js');
const chai = require('chai');
const { expect } = chai;
const chaiHttp = require('chai-http');
chai.use(chaiHttp);
chai.use(require('chai-like'));
const chaiThings = require('chai-things');
chai.use(chaiThings);
const idGenerator = require('../Utils/helper');

require('../Login/loginSysAdmin');
require("../Organizations/OrgSetup/RegisterOrg1");
require("../Organizations/OrgSetup/RegisterOrg2");
require("../Login/OrgLogin/Org1Login");
require("../Login/OrgLogin/Org2Login");

describe('Checking operation logs by patient', function() {

    // register datatype1
    before((done) => {
        chai.request(Config.server).post('omr/api/v1/datatypes') 
            .set('Accept',  'application/json')
            .set('token', Config.sysAdminToken)
            .send({id: idGenerator(), description: Config.datatype1.description})
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                Config.datatype1.id = res.body.id;
                done();
            });
    })

    // register datatype2
    before((done) => {
        chai.request(Config.server).post('omr/api/v1/datatypes') 
            .set('Accept',  'application/json')
            .set('token', Config.sysAdminToken)
            .send({id: idGenerator(), description: Config.datatype2.description})
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                Config.datatype2.id = res.body.id;
                done();
            });
    })

    // register datatype3
    before((done) => {
        chai.request(Config.server).post('omr/api/v1/datatypes') 
            .set('Accept',  'application/json')
            .set('token', Config.sysAdminToken)
            .send({id: idGenerator(), description: Config.datatype3.description})
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                Config.datatype3.id = res.body.id;
                done();
            });
    })

    // register datatype4
    before((done) => {
        chai.request(Config.server).post('omr/api/v1/datatypes') 
            .set('Accept',  'application/json')
            .set('token', Config.sysAdminToken)
            .send({id: idGenerator(), description: Config.datatype4.description})
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                Config.datatype4.id = res.body.id;
                done();
            });
    })

    // register org1Service1
    let org1Service1 = {};    
    before((done) => {
        org1Service1 = {
            "id": idGenerator(),
            "name": "o1s1",
            "secret": "newservice1pass",
            "ca_org": Config.caOrg,
            "email": "newservice1email@example.com",
            "org_id": Config.org1.id,
            "summary": "New service under org 1. Has two datatypes",
            "terms": {
                "term1" : "example term",
                "term2" : "example term",
            },
            "payment_required": "yes",
            "datatypes":  [
                {
                    "datatype_id": Config.datatype1.id,
                    "access":[
                        "write",
                        "read"
                    ]
                },
                {
                    "datatype_id": Config.datatype2.id,
                    "access":[
                        "write",
                        "read"
                    ]
                }
            ],
            "solution_private_data": {
                "sample data 1": "service sample data 1",
                "sample data 2": "service sample data 2"
            }
        };

        chai.request(Config.server).post('omr/api/v1/services') 
            .set('Accept',  'application/json')
            .set('token', Config.orgAdminToken1)
            .send(org1Service1)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("success");
                done();
                });
    });

    // register org1Service2
    let org1Service2 = {};    
    before((done) => {
        org1Service2 = {
            "id": idGenerator(),
            "name": "o1s2",
            "secret": "org1Service2pass",
            "ca_org": Config.caOrg,
            "email": "org1Service2email@example.com",
            "org_id": Config.org1.id,
            "summary": "New service under org 1. Has one datatype",
            "terms": {
                "term1" : "example term"
            },
            "payment_required": "yes",
            "datatypes":  [
                {
                    "datatype_id": Config.datatype3.id,
                    "access":[
                        "write",
                        "read"
                    ]
                }
            ],
            "solution_private_data": {
                "sample data 1": "service sample data 1"
            }
        };

        chai.request(Config.server).post('omr/api/v1/services') 
            .set('Accept',  'application/json')
            .set('token', Config.orgAdminToken1)
            .send(org1Service2)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("success");
                done();
                });
    });

    // register of org2Service1
    let org2Service1 = {};    
    before((done) => {
        org2Service1 = {
            "id": idGenerator(),
            "name": "org2Service1",
            "secret": "org2Service1pass",
            "ca_org": Config.caOrg,
            "email": "org2Service1email@example.com",
            "org_id": Config.org2.id,
            "summary": "New service under org 2. Has three datatypes",
            "terms": {
                "term1" : "example term",
                "term2" : "example term",
                "term3" : "example term"
            },
            "payment_required": "yes",
            "datatypes": [
                {
                    "datatype_id": Config.datatype3.id,
                    "access":[
                        "write",
                        "read"
                    ]
                },
                {
                    "datatype_id": Config.datatype4.id,
                    "access":[
                        "write"
                    ]
                }
            ],
            "solution_private_data": {
                "sample data 1": "service sample data 1",
                "sample data 2": "service sample data 3",
                "sample data 3": "service sample data 4"
            }
        };

        chai.request(Config.server).post('omr/api/v1/services') 
            .set('Accept',  'application/json')
            .set('token', Config.orgAdminToken2)
            .send(org2Service1)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("success");
                done();
                });
    });

    // create newPatient1
    const newPatient1 = {
        "id": idGenerator(),
        "secret": "newPatient1",
        "name": "newPatient1",
        "role": "user",
        "org" : "",
        "email":  "newPatient1email@example.com",
        "ca_org": Config.caOrg,
        "data": {
          "address": "2 User St"
        }
    };

    // create newPatient2
    const newPatient2 = {
        "id": idGenerator(),
        "secret": "newPatient2",
        "name": "newPatient2",
        "role": "user",
        "org" : "",
        "email":  "newPatient2email@example.com",
        "ca_org": Config.caOrg,
        "data": {
          "address": "2 User St"
        }
    };

    // create newPatient3
    const newPatient3 = {
        "id": idGenerator(),
        "secret": "newPatient3",
        "name": "newPatient3",
        "role": "user",
        "org" : "",
        "email":  "newPatient3email@example.com",
        "ca_org": Config.caOrg,
        "data": {
          "address": "2 User St"
        }
    };

    // create newPatient4
    const newPatient4 = {
        "id": idGenerator(),
        "secret": "newPatient4",
        "name": "newPatient4",
        "role": "user",
        "org" : "",
        "email":  "newPatient4email@example.com",
        "ca_org": Config.caOrg,
        "data": {
          "address": "2 User St"
        }
    };

    // register newPatient1 (doesn't belong to any Org, with role user)
    before((done) => {
        chai.request(Config.server).post('omr/api/v1/users') 
            .set('Accept',  'application/json')
            .set('token', Config.sysAdminToken)
            .send(newPatient1)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("success");
                done();
            });
    });

    // register newPatient2 (doesn't belong to any Org, with role user)
    before((done) => {
        chai.request(Config.server).post('omr/api/v1/users') 
            .set('Accept',  'application/json')
            .set('token', Config.sysAdminToken)
            .send(newPatient2)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("success");
                done();
            });
    });

    // register newPatient3 (doesn't belong to any Org, with role user)
    before((done) => {
        chai.request(Config.server).post('omr/api/v1/users') 
            .set('Accept',  'application/json')
            .set('token', Config.sysAdminToken)
            .send(newPatient3)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("success");
                done();
            });
    });

    // register newPatient4 (doesn't belong to any Org, with role user)
    before((done) => {
        chai.request(Config.server).post('omr/api/v1/users') 
            .set('Accept',  'application/json')
            .set('token', Config.sysAdminToken)
            .send(newPatient4)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("success");
                done();
            });
    });

    // enroll newPatient1 to org1Service1
    before((done) => {
        const bodyRequest = {
            "user": newPatient1.id,
            "service": org1Service1.id,
            "status": "active"
        }
        chai.request(Config.server).post('omr/api/v1/services/' + org1Service1.id + '/user/enroll')
            .set('Accept', 'application/json')
            .set('token', Config.orgAdminToken1)
            .send(bodyRequest)
            .end(function(err, res){           
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("success");
                done();
            });
    });

    // enroll newPatient2 to org1Service1
    before((done) => {
        const bodyRequest = {
            "user": newPatient2.id,
            "service": org1Service1.id,
            "status": "active"
        }
        chai.request(Config.server).post('omr/api/v1/services/' + org1Service1.id + '/user/enroll')
            .set('Accept', 'application/json')
            .set('token', Config.orgAdminToken1)
            .send(bodyRequest)
            .end(function(err, res){           
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("success");
                done();
            });
    });

    // enroll newPatient2 to org2Service1
    before((done) => {
        const bodyRequest = {
            "user": newPatient2.id,
            "service": org2Service1.id,
            "status": "active"
        }
        chai.request(Config.server).post('omr/api/v1/services/' + org2Service1.id + '/user/enroll')
            .set('Accept', 'application/json')
            .set('token', Config.orgAdminToken2)
            .send(bodyRequest)
            .end(function(err, res){           
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("success");
                done();
            });
    });

    // enroll newPatient1 to org1Service2
    before((done) => {
        const bodyRequest = {
            "user": newPatient1.id,
            "service": org1Service2.id,
            "status": "active"
        }
        chai.request(Config.server).post('omr/api/v1/services/' + org1Service2.id + '/user/enroll')
            .set('Accept', 'application/json')
            .set('token', Config.orgAdminToken1)
            .send(bodyRequest)
            .end(function(err, res){           
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("success");
                done();
            });
    });

    // enroll newPatient1 to org2Service1
    before((done) => {
        const bodyRequest = {
            "user": newPatient1.id,
            "service": org2Service1.id,
            "status": "active"
        }
        chai.request(Config.server).post('omr/api/v1/services/' + org2Service1.id + '/user/enroll')
            .set('Accept', 'application/json')
            .set('token', Config.orgAdminToken2)
            .send(bodyRequest)
            .end(function(err, res){           
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("success");
                done();
            });
    });

    // enroll newPatient3 to org2Service1
    before((done) => {
        const bodyRequest = {
            "user": newPatient3.id,
            "service": org2Service1.id,
            "status": "active"
        }
        chai.request(Config.server).post('omr/api/v1/services/' + org2Service1.id + '/user/enroll')
            .set('Accept', 'application/json')
            .set('token', Config.orgAdminToken2)
            .send(bodyRequest)
            .end(function(err, res){           
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("success");
                done();
            });
    });

    // enroll newPatient4 to org1Service2
    before((done) => {
        const bodyRequest = {
            "user": newPatient4.id,
            "service": org1Service2.id,
            "status": "active"
        }
        chai.request(Config.server).post('omr/api/v1/services/' + org1Service2.id + '/user/enroll')
            .set('Accept', 'application/json')
            .set('token', Config.orgAdminToken1)
            .send(bodyRequest)
            .end(function(err, res){           
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("success");
                done();
            });
    });

    let newPatient1Token = '';
    before((done) => {
        chai.request(Config.server).get('common/api/v1/login')
            .set('Accept', 'application/json')
            .set('user-id', newPatient1.id)
            .set('password', newPatient1.secret)
            .set('login-org', newPatient1.ca_org)
            .set('login-channel', Config.channel)
            .set('signature', "")
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("token");
                expect(res.body.token).to.not.equal("");
                newPatient1Token = res.body.token;
                done();
            });
    });

    let newPatient2Token = '';
    before((done) => {
        chai.request(Config.server).get('common/api/v1/login')
            .set('Accept', 'application/json')
            .set('user-id', newPatient2.id)
            .set('password', newPatient2.secret)
            .set('login-org', newPatient2.ca_org)
            .set('login-channel', Config.channel)
            .set('signature', "")
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("token");
                expect(res.body.token).to.not.equal("");
                newPatient2Token = res.body.token;
                done();
            });
    });

    let newPatient3Token = '';
    before((done) => {
        chai.request(Config.server).get('common/api/v1/login')
            .set('Accept', 'application/json')
            .set('user-id', newPatient3.id)
            .set('password', newPatient3.secret)
            .set('login-org', newPatient3.ca_org)
            .set('login-channel', Config.channel)
            .set('signature', "")
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("token");
                expect(res.body.token).to.not.equal("");
                newPatient3Token = res.body.token;
                done();
            });
    });

    let newPatient4Token = '';
    before((done) => {
        chai.request(Config.server).get('common/api/v1/login')
            .set('Accept', 'application/json')
            .set('user-id', newPatient4.id)
            .set('password', newPatient4.secret)
            .set('login-org', newPatient4.ca_org)
            .set('login-channel', Config.channel)
            .set('signature', "")
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("token");
                expect(res.body.token).to.not.equal("");
                newPatient4Token = res.body.token;
                done();
            });
    });

    let defaultOrg1Service1AdminToken = '';
    before((done) => {
        chai.request(Config.server).get('common/api/v1/login')
            .set('Accept', 'application/json')
            .set('user-id', org1Service1.id)
            .set('password', org1Service1.secret)
            .set('login-org', org1Service1.ca_org)
            .set('login-channel', Config.channel)
            .set('signature', "")
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("token");
                expect(res.body.token).to.not.equal("");
                defaultOrg1Service1AdminToken = res.body.token;
                done();
            });
    });

    let defaultOrg2Service1AdminToken = '';
    before((done) => {
        chai.request(Config.server).get('common/api/v1/login')
            .set('Accept', 'application/json')
            .set('user-id', org2Service1.id)
            .set('password', org2Service1.secret)
            .set('login-org', org2Service1.ca_org)
            .set('login-channel', Config.channel)
            .set('signature', "")
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("token");
                expect(res.body.token).to.not.equal("");
                defaultOrg2Service1AdminToken = res.body.token;
                done();
            });
    });

    // create consent to o1d1 as newPatient1 
    before((done) => {
        const bodyRequest = {
            "patient_id": newPatient1.id,
            "service_id": org1Service1.id,
            "target_id": org1Service1.id,
            "datatype_id": org1Service1.datatypes[0].datatype_id,
            "option": [
                "write",
                "read"
            ],
            "expiration": 0
        }
        chai.request(Config.server).post('omr/api/v1/consents/patientData')
            .set('Accept', 'application/json')
            .set('token', newPatient1Token)
            .send(bodyRequest)
            .end(function(err, res){           
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("success");
                done();
            });
    });

    // create consent to o1s2d3 by newPatient1
    before((done) => {
        const bodyRequest = {
            "patient_id": newPatient1.id,
            "service_id": org1Service2.id,
            "target_id": org1Service2.id,
            "datatype_id": org1Service2.datatypes[0].datatype_id,
            "option": [
                "write",
                "read"
            ],
            "expiration": 0
        }
        chai.request(Config.server).post('omr/api/v1/consents/patientData')
            .set('Accept', 'application/json')
            .set('token', newPatient1Token)
            .send(bodyRequest)
            .end(function(err, res){           
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("success");
                done();
            });
    });

    // create consent to o2d3
    before((done) => {
        const bodyRequest = {
            "patient_id": newPatient1.id,
            "service_id": org2Service1.id,
            "target_id": org2Service1.id,
            "datatype_id": org2Service1.datatypes[0].datatype_id,
            "option": [
                "write"
            ],
            "expiration": 0
        }
        chai.request(Config.server).post('omr/api/v1/consents/patientData')
            .set('Accept', 'application/json')
            .set('token', newPatient1Token)
            .send(bodyRequest)
            .end(function(err, res){           
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("success");
                done();
            });
    });

    // create consent to o1d1 by newPatient2
    before((done) => {
        const bodyRequest = {
            "patient_id": newPatient2.id,
            "service_id": org1Service1.id,
            "target_id": org1Service1.id,
            "datatype_id": org1Service1.datatypes[0].datatype_id,
            "option": [
                "write"
            ],
            "expiration": 0
        }
        chai.request(Config.server).post('omr/api/v1/consents/patientData')
            .set('Accept', 'application/json')
            .set('token', newPatient2Token)
            .send(bodyRequest)
            .end(function(err, res){           
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("success");
                done();
            });
    });

    // create consent to o2d4 by newPatient2
    before((done) => {
        const bodyRequest = {
            "patient_id": newPatient2.id,
            "service_id": org2Service1.id,
            "target_id": org2Service1.id,
            "datatype_id": org2Service1.datatypes[1].datatype_id,
            "option": [
                "read"
            ],
            "expiration": 0
        }
        chai.request(Config.server).post('omr/api/v1/consents/patientData')
            .set('Accept', 'application/json')
            .set('token', newPatient2Token)
            .send(bodyRequest)
            .end(function(err, res){           
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("success");
                done();
            });
    });

    // upload user data by target service org1Service1
    before((done) => {
        const bodyRequest = {
            "test data 1": "some test data 1 for o1s1d1"
        }
        chai.request(Config.server).post('omr/api/v1/services/' + org1Service1.id + '/users/' +
         newPatient1.id + '/datatype/' + org1Service1.datatypes[0].datatype_id + '/upload')
            .set('Accept', 'application/json')
            .set('token', defaultOrg1Service1AdminToken)
            .send(bodyRequest)
            .end(function(err, res){           
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("success");
                done();
            });
    });

    before((done) => {
        const bodyRequest = {
            "test data 1": "some test data 1 for o1s1d1 for newPatient2"
        }
        chai.request(Config.server).post('omr/api/v1/services/' + org1Service1.id + '/users/' +
         newPatient2.id + '/datatype/' + org1Service1.datatypes[0].datatype_id + '/upload')
            .set('Accept', 'application/json')
            .set('token', defaultOrg1Service1AdminToken)
            .send(bodyRequest)
            .end(function(err, res){           
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("success");
                done();
            });
    });

    before((done) => {
        const bodyRequest = {
            "test data 1": "some test data 1 for o2s1d3"
        }
        chai.request(Config.server).post('omr/api/v1/services/' + org2Service1.id + '/users/' +
         newPatient1.id + '/datatype/' + org2Service1.datatypes[0].datatype_id + '/upload')
            .set('Accept', 'application/json')
            .set('token', defaultOrg2Service1AdminToken)
            .send(bodyRequest)
            .end(function(err, res){
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("success");
                done();
            });
    });

    // download user data by default service admin of org1Service1
    before((done) => {
        chai.request(Config.server).get('omr/api/v1/services/' + org1Service1.id + '/users/' +
         newPatient1.id + '/datatypes/' + org1Service1.datatypes[0].datatype_id + '/download')
            .set('Accept', 'application/json')
            .set('token', defaultOrg1Service1AdminToken)
            .end(function(err, res){
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.be.an('array');
                expect(res.body).to.have.lengthOf(1);
                expect(JSON.stringify(res.body[0].data)).to.equal(JSON.stringify({
                    "test data 1": "some test data 1 for o1s1d1"
                }));
                done();
            });
    });

    // download user data by default service admin of org1Service1
    before((done) => {
        chai.request(Config.server).get('omr/api/v1/services/' + org1Service1.id + '/users/' +
         newPatient2.id + '/datatypes/' + org1Service1.datatypes[0].datatype_id + '/download')
            .set('Accept', 'application/json')
            .set('token', defaultOrg1Service1AdminToken)
            .end(function(err, res){
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.be.an('array');
                expect(res.body).to.have.lengthOf(1);
                expect(JSON.stringify(res.body[0].data)).to.equal(JSON.stringify({
                    "test data 1": "some test data 1 for o1s1d1 for newPatient2"
                }));
                done();
            });
    });

    // download user data by default service admin of org2Service1
    before((done) => {
        chai.request(Config.server).get('omr/api/v1/services/' + org2Service1.id + '/users/' +
         newPatient1.id + '/datatypes/' + org2Service1.datatypes[0].datatype_id + '/download')
            .set('Accept', 'application/json')
            .set('token', defaultOrg2Service1AdminToken)
            .end(function(err, res){
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.be.an('array');
                expect(res.body).to.have.lengthOf(1);
                expect(JSON.stringify(res.body[0].data)).to.equal(JSON.stringify({
                    "test data 1": "some test data 1 for o2s1d3"
                }));
                done();
            });
    });

    // revoke consent from org2Service1 by newPatient2
    before((done) => {
        const bodyRequest = {
            "patient_id": newPatient2.id,
            "service_id": org2Service1.id,
            "target_id": org2Service1.id,
            "datatype_id": org2Service1.datatypes[1].datatype_id,
            "option": [
                "deny"
            ],
            "expiration": 0
        }
        chai.request(Config.server).post('omr/api/v1/consents/patientData')
            .set('Accept', 'application/json')
            .set('token', newPatient2Token)
            .send(bodyRequest)
            .end(function(err, res){
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("success");
                done();
            });
    });

    // checking logs by newPatient1 searched by datatype_id
    it('Should return logs searched by datatype_id', function (done) {
        chai.request(Config.server).get('omr/api/v1/history?datatype_id=' + org1Service1.datatypes[1].datatype_id + '&latest_only=false&maxNum=20') 
            .set('Accept',  'application/json')
            .set('token', newPatient1Token)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.body).to.be.an('array');
                expect(res.body).to.have.lengthOf(0);
                done();
            });
    });

    // checking logs by newPatient1 searched by datatype_id
    it('Should return logs searched by datatype_id', function (done) {
        chai.request(Config.server).get('omr/api/v1/history?datatype_id=' + org1Service1.datatypes[0].datatype_id + '&latest_only=false&maxNum=20') 
            .set('Accept',  'application/json')
            .set('token', newPatient1Token)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.body).to.be.an('array');
                expect(res.body).to.have.lengthOf(3);
                expect(res.body[0].type).to.equal('PutConsentPatientData');
                expect(JSON.stringify(res.body[0].data)).to.equal(JSON.stringify({"option":["write","read"]}));
                expect(res.body[1].type).to.equal('UploadUserData');
                expect(JSON.stringify(res.body[1].data)).to.equal(JSON.stringify({}));
                expect(res.body[2].type).to.equal('DownloadUserData');
                expect(JSON.stringify(res.body[2].data)).to.equal(JSON.stringify({}));
                done();
            });
    });

    // checking logs by newPatient1 searched by patient_id
    it('Should return logs searched by patient_id', function (done) {
        chai.request(Config.server).get('omr/api/v1/history?patient_id=' + newPatient1.id + '&latest_only=false&maxNum=20') 
            .set('Accept',  'application/json')
            .set('token', newPatient1Token)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.body).to.be.an('array');
                expect(res.body).to.have.lengthOf(7);
                expect(res.body[0].type).to.equal('PutConsentPatientData');
                expect(JSON.stringify(res.body[0].data)).to.equal(JSON.stringify({"option":["write","read"]}));
                expect(res.body[1].type).to.equal('PutConsentPatientData');
                expect(JSON.stringify(res.body[1].data)).to.equal(JSON.stringify({"option":["write","read"]}));
                expect(res.body[2].type).to.equal('PutConsentPatientData');
                expect(JSON.stringify(res.body[2].data)).to.equal(JSON.stringify({"option":["write"]}));
                expect(res.body[3].type).to.equal('UploadUserData');
                expect(JSON.stringify(res.body[3].data)).to.equal(JSON.stringify({}));
                expect(res.body[4].type).to.equal('UploadUserData');
                expect(JSON.stringify(res.body[4].data)).to.equal(JSON.stringify({}));
                expect(res.body[5].type).to.equal('DownloadUserData');
                expect(JSON.stringify(res.body[5].data)).to.equal(JSON.stringify({}));
                expect(res.body[6].type).to.equal('DownloadUserData');
                expect(JSON.stringify(res.body[6].data)).to.equal(JSON.stringify({}));
                done();
            });
    });

    // checking logs by newPatient1 searched by service_id with latest_only flag
    it('Should return logs searched by service_id with latest_only flag', function (done) {
        chai.request(Config.server).get('omr/api/v1/history?service_id=' + org1Service1.id + '&latest_only=true&maxNum=20') 
            .set('Accept',  'application/json')
            .set('token', newPatient1Token)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.body).to.be.an('array');
                expect(res.body).to.have.lengthOf(1);
                expect(res.body[0].type).to.equal('DownloadUserData');
                expect(JSON.stringify(res.body[0].data)).to.equal(JSON.stringify({}));
                done();
            });
    });

    // checking logs by newPatient1 searched by service_id with maxNum flag
    it('Should return logs searched by service_id with maxNum flag', function (done) {
        chai.request(Config.server).get('omr/api/v1/history?service_id=' + org1Service1.id + '&latest_only=false&maxNum=2') 
            .set('Accept',  'application/json')
            .set('token', newPatient1Token)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.body).to.be.an('array');
                expect(res.body).to.have.lengthOf(2);
                expect(res.body[0].type).to.equal('PutConsentPatientData');
                expect(JSON.stringify(res.body[0].data)).to.equal(JSON.stringify({"option":["write","read"]}));
                expect(res.body[1].type).to.equal('UploadUserData');
                expect(JSON.stringify(res.body[1].data)).to.equal(JSON.stringify({}));
                done();
            });
    });

    // checking logs by newPatient2 searched by consent_owner_target_id = service_id
    it('Should return logs searched by consent_owner_target_id = service_id', function (done) {
        chai.request(Config.server).get('omr/api/v1/history?consent_owner_target_id=' + org2Service1.id + '&latest_only=false&maxNum=20') 
            .set('Accept',  'application/json')
            .set('token', newPatient2Token)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.body).to.be.an('array');
                expect(res.body).to.have.lengthOf(2);
                expect(res.body[0].type).to.equal('PutConsentPatientData');
                expect(JSON.stringify(res.body[0].data)).to.equal(JSON.stringify({"option":["read"]}));
                expect(res.body[1].type).to.equal('PutConsentPatientData');
                expect(JSON.stringify(res.body[1].data)).to.equal(JSON.stringify({"option":["deny"]}));
                done();
            });
    });

    // checking logs by newPatient3 searched by service_id
    it('Should return logs searched by service_id', function (done) {
        chai.request(Config.server).get('omr/api/v1/history?service_id=' + org2Service1.id + '&latest_only=false&maxNum=20') 
            .set('Accept',  'application/json')
            .set('token', newPatient3Token)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.body).to.be.an('array');
                expect(res.body).to.have.lengthOf(0);
                done();
            });
    });

    // checking logs by newPatient3 searched by patient_id
    it('Should return logs searched by patient_id', function (done) {
        chai.request(Config.server).get('omr/api/v1/history?patient_id=' + newPatient1.id + '&latest_only=false&maxNum=20') 
            .set('Accept',  'application/json')
            .set('token', newPatient3Token)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.body).to.be.an('array');
                expect(res.body).to.have.lengthOf(0);
                done();
            });
    });

    // checking logs by newPatient4 searched by datatype_id
    it('Should return logs searched by datatype_id', function (done) {
        chai.request(Config.server).get('omr/api/v1/history?datatype_id=' + org2Service1.datatypes[0].datatype_id + '&latest_only=false&maxNum=20') 
            .set('Accept',  'application/json')
            .set('token', newPatient4Token)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.body).to.be.an('array');
                expect(res.body).to.have.lengthOf(0);
                done();
            });
    });

    // checking logs by newPatient4 searched by patient_id
    it('Should return logs searched by patient_id', function (done) {
        chai.request(Config.server).get('omr/api/v1/history?patient_id=' + newPatient1.id + '&latest_only=false&maxNum=20') 
            .set('Accept',  'application/json')
            .set('token', newPatient4Token)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.body).to.be.an('array');
                expect(res.body).to.have.lengthOf(0);
                done();
            });
    });

    // checking logs of newPatient2 searched by several options: patient_id, datatype_id
    it('Should return logs searched by several options', function (done) {
        chai.request(Config.server).get('omr/api/v1/history?patient_id=' + newPatient2.id + '&datatype_id=' + org2Service1.datatypes[1].datatype_id) 
            .set('Accept',  'application/json')
            .set('token', newPatient2Token)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.body).to.be.an('array');
                expect(res.body).to.have.lengthOf(2);
                expect(res.body[0].type).to.equal('PutConsentPatientData');
                expect(JSON.stringify(res.body[0].data)).to.equal(JSON.stringify({"option":["read"]}));
                expect(res.body[1].type).to.equal('PutConsentPatientData');
                expect(JSON.stringify(res.body[1].data)).to.equal(JSON.stringify({"option":["deny"]}));
                done();
            });
    });
});