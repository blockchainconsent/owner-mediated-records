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
require("../Login/OrgLogin/Org1Login");

describe('Checking of SingleApi feature', function() {

    const patient1Failed = idGenerator();
    const patientSuccess = idGenerator();
    const patient3consent = idGenerator();

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
                "term3" : "example term"
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
                },
                {
                    "datatype_id": Config.datatype3.id,
                    "access":[
                        "write",
                        "read"
                    ]
                }
            ],
            "solution_private_data": {
                "sample data 1": "service sample data 1",
                "sample data 2": "service sample data 2",
                "sample data 3": "service sample data 3"
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

    let o1s1adminToken = '';
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
                o1s1adminToken = res.body.token;
                done();
            });
    });

    // create orgUserServiceAdmin
    const orgUserServiceAdmin = {
        "id": idGenerator(),
        "secret": "orgUserServiceAdmin",
        "name": "orgUserServiceAdmin",
        "role": "user",
        "org" : Config.org1.id,
        "email":  "orgUserServiceAdminEmail@example.com",
        "ca_org": Config.caOrg,
        "data": {
          "address": "2 User St"
        }
    };

    // register new orgUserServiceAdmin
    before((done) => {
        chai.request(Config.server).post('omr/api/v1/users') 
            .set('Accept',  'application/json')
            .set('token', Config.orgAdminToken1)
            .send(orgUserServiceAdmin)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("success");
                done();
            });
    });

    // grant service admin permission to orgUserServiceAdmin
    before((done) => {
        chai.request(Config.server).put('omr/api/v1/users/' + orgUserServiceAdmin.id + '/permissions/services/' + org1Service1.id) 
            .set('Accept',  'application/json')
            .set('token', Config.orgAdminToken1)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("success");
                done();
            });
    });

    // login as service admin
    let orgUserServiceAdminToken = '';
    before((done) => {
        chai.request(Config.server).get('common/api/v1/login')
            .set('Accept', 'application/json')
            .set('user-id', orgUserServiceAdmin.id)
            .set('password', orgUserServiceAdmin.secret)
            .set('login-org', orgUserServiceAdmin.ca_org)
            .set('login-channel', Config.channel)
            .set('signature', "")
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("token");
                expect(res.body.token).to.not.equal("");
                orgUserServiceAdminToken = res.body.token;
                done();
            });
    });

    it('Should successfully create consent via SingleApi feature', function (done) {
        const bodyRequest = {
            "id": patient3consent,
            "secret": "secret1",
            "name": "patient3consent",
            "email": "patient3consent@example.com",
            "ca_org": Config.caOrg,
            "data": {},
            "service_id": org1Service1.id,
            "consents": 
            [
                {
                    "datatype_id": org1Service1.datatypes[0].datatype_id,
                    "target_id": org1Service1.id,
                    "option": [
                    "write"
                    ],
                    "expirationTimestamp": 0
                },
                {
                    "datatype_id": org1Service1.datatypes[1].datatype_id,
                    "target_id": org1Service1.id,
                    "option": [
                    "read"
                    ],
                    "expirationTimestamp": 0
                },
                {
                    "datatype_id": org1Service1.datatypes[2].datatype_id,
                    "target_id": org1Service1.id,
                    "option": [
                    "write","read"
                    ],
                    "expirationTimestamp": 0
                }
            ]
        }
        const consent1 = {
            "owner_id":patient3consent,
            "service_id":org1Service1.id,
            "target_service_id":org1Service1.id,
            "datatype_id":org1Service1.datatypes[0].datatype_id,
            "option":["write"],
            "expirationTimestamp":0
        }
        const consent2 = {
            "owner_id":patient3consent,
            "service_id":org1Service1.id,
            "target_service_id":org1Service1.id,
            "datatype_id":org1Service1.datatypes[1].datatype_id,
            "option":["read"],
            "expirationTimestamp":0
        }
        const consent3 = {
            "owner_id":patient3consent,
            "service_id":org1Service1.id,
            "target_service_id":org1Service1.id,
            "datatype_id":org1Service1.datatypes[2].datatype_id,
            "option":["write","read"],
            "expirationTimestamp":0
        }
        chai.request(Config.server).post('omr/api/v1/users/patient_register_and_consent')
            .set('Accept', 'application/json')
            .set('token', Config.orgAdminToken1)
            .send(bodyRequest)
            .end(function(err, res){
                expect(res.status).to.equal(200);
                expect(res.body.failures).to.be.an('array');
                expect(res.body.failures).to.have.length(0);
                expect(res.body.successes).to.be.an('array');
                expect(res.body.successes).to.have.length(3);
                expect(res.body.successes).to.contain.something.like(consent1);
                expect(res.body.successes).to.contain.something.like(consent2);
                expect(res.body.successes).to.contain.something.like(consent3);
                expect(res.body.status).to.equal(200);
                expect(res.body.failure_type).to.equal("");
                done();
            });
    });

    // get all consents for org1Service1/patient3consent pair as org Admin
    it('Should successfully get all consents from patient3consent to org1Service1 as org admin', function (done) {
        let servicePatientPair1 = {
            "owner":patient3consent,
            "service":org1Service1.id,
            "datatype":org1Service1.datatypes[0].datatype_id,
            "target":org1Service1.id,
            "option":["write"],
            "expiration":0,
        }
        let servicePatientPair2 = {
            "owner":patient3consent,
            "service":org1Service1.id,
            "datatype":org1Service1.datatypes[1].datatype_id,
            "target":org1Service1.id,
            "option":["read"],
            "expiration":0,
        }
        let servicePatientPair3 = {
            "owner":patient3consent,
            "service":org1Service1.id,
            "datatype":org1Service1.datatypes[2].datatype_id,
            "target":org1Service1.id,
            "option":["write","read"],
            "expiration":0,
        }
        chai.request(Config.server).get('omr/api/v1/services/' + org1Service1.id + '/users/' + patient3consent + '/consents')
            .set('Accept', 'application/json')
            .set('token', Config.orgAdminToken1)
            .end(function(err, res){
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.be.an('array');
                expect(res.body).to.have.lengthOf(3);
                expect(res.body).to.contain.something.like(servicePatientPair1);
                expect(res.body).to.contain.something.like(servicePatientPair2);
                expect(res.body).to.contain.something.like(servicePatientPair3);
                done();
            });
    });

    it('Should successfully create single consent via SingleApi feature as org admin', function (done) {
        const patient1 = idGenerator();
        const bodyRequest = {
            "id": patient1,
            "secret": "secret1",
            "name": "patient1",
            "email": "patient1@example.com",
            "ca_org": Config.caOrg,
            "data": {},
            "service_id": org1Service1.id,
            "consents": 
            [
                {
                    "datatype_id": org1Service1.datatypes[0].datatype_id,
                    "target_id": org1Service1.id,
                    "option": [
                    "write"
                    ],
                    "expirationTimestamp": 0
                }
            ]
        }
        const consent1 = {
            "owner_id":patient1,
            "service_id":org1Service1.id,
            "target_service_id":org1Service1.id,
            "datatype_id":org1Service1.datatypes[0].datatype_id,
            "option":["write"],
            "expirationTimestamp":0
        }
        chai.request(Config.server).post('omr/api/v1/users/patient_register_and_consent')
            .set('Accept', 'application/json')
            .set('token', Config.orgAdminToken1)
            .send(bodyRequest)
            .end(function(err, res){
                expect(res.status).to.equal(200);
                expect(res.body.failures).to.be.an('array');
                expect(res.body.failures).to.have.length(0);
                expect(res.body.successes).to.be.an('array');
                expect(res.body.successes).to.have.length(1);
                expect(res.body.successes).to.contain.something.like(consent1);
                expect(res.body.status).to.equal(200);
                expect(res.body.failure_type).to.equal("");
                done();
            });
    });

    it('Should successfully create single consent via SingleApi feature as service admin', function (done) {
        const patient1 = idGenerator();
        const bodyRequest = {
            "id": patient1,
            "secret": "secret1",
            "name": "patient1",
            "email": "patient1@example.com",
            "ca_org": Config.caOrg,
            "data": {},
            "service_id": org1Service1.id,
            "consents": 
            [
                {
                    "datatype_id": org1Service1.datatypes[0].datatype_id,
                    "target_id": org1Service1.id,
                    "option": ["read"],
                    "expirationTimestamp": 0
                }
            ]
        }
        const consent1 = {
            "owner_id":patient1,
            "service_id":org1Service1.id,
            "target_service_id":org1Service1.id,
            "datatype_id":org1Service1.datatypes[0].datatype_id,
            "option":["read"],
            "expirationTimestamp":0
        }
        chai.request(Config.server).post('omr/api/v1/users/patient_register_and_consent')
            .set('Accept', 'application/json')
            .set('token', o1s1adminToken)
            .send(bodyRequest)
            .end(function(err, res){
                expect(res.status).to.equal(200);
                expect(res.body.failures).to.be.an('array');
                expect(res.body.failures).to.have.length(0);
                expect(res.body.successes).to.be.an('array');
                expect(res.body.successes).to.have.length(1);
                expect(res.body.successes).to.contain.something.like(consent1);
                expect(res.body.status).to.equal(200);
                expect(res.body.failure_type).to.equal("");
                done();
            });
    });

    it('Should successfully create single consent via SingleApi feature as org user with service admin permission', function (done) {
        const bodyRequest = {
            "id": patientSuccess,
            "secret": "secret1",
            "name": "patientSuccess",
            "email": "patientSuccess@example.com",
            "ca_org": Config.caOrg,
            "data": {},
            "service_id": org1Service1.id,
            "consents": 
            [
                {
                    "datatype_id": org1Service1.datatypes[0].datatype_id,
                    "target_id": org1Service1.id,
                    "option": ["write"],
                    "expirationTimestamp": 0
                }
            ]
        }
        const consent1 = {
            "owner_id":patientSuccess,
            "service_id":org1Service1.id,
            "target_service_id":org1Service1.id,
            "datatype_id":org1Service1.datatypes[0].datatype_id,
            "option":["write"],
            "expirationTimestamp":0
        }
        chai.request(Config.server).post('omr/api/v1/users/patient_register_and_consent')
            .set('Accept', 'application/json')
            .set('token', orgUserServiceAdminToken)
            .send(bodyRequest)
            .end(function(err, res){
                expect(res.status).to.equal(200);
                expect(res.body.failures).to.be.an('array');
                expect(res.body.failures).to.have.length(0);
                expect(res.body.successes).to.be.an('array');
                expect(res.body.successes).to.have.length(1);
                expect(res.body.successes).to.contain.something.like(consent1);
                expect(res.body.status).to.equal(200);
                expect(res.body.failure_type).to.equal("");
                done();
            });
    });

    // get all consents for org1Service1/patientSuccess pair as org1Admin
    it('Should successfully get all consents from patientSuccess to org1Service1 as service admin', function (done) {
        let servicePatientPair1 = {
            "owner":patientSuccess,
            "service":org1Service1.id,
            "datatype":org1Service1.datatypes[0].datatype_id,
            "target":org1Service1.id,
            "option":["write"],
            "expiration":0,
        }
        chai.request(Config.server).get('omr/api/v1/services/' + org1Service1.id + '/users/' + patientSuccess + '/consents')
            .set('Accept', 'application/json')
            .set('token', orgUserServiceAdminToken)
            .end(function(err, res){
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.be.an('array');
                expect(res.body).to.have.lengthOf(1);
                expect(res.body).to.contain.something.like(servicePatientPair1);
                done();
            });
    });

    it('Should have Enroll fail (invalid service_id)', function (done) {
        const patient1 = idGenerator();
        const bodyRequest = {
            "id": patient1,
            "secret": "secret1",
            "name": "patient1",
            "email": "patient1@example.com",
            "ca_org": Config.caOrg,
            "data": {},
            "service_id": "invalid_service_id",
            "consents": 
            [
                {
                    "datatype_id": org1Service1.datatypes[0].datatype_id,
                    "target_id": org1Service1.id,
                    "option": [
                    "write"
                    ],
                    "expirationTimestamp": 0
                },
                {
                    "datatype_id": org1Service1.datatypes[1].datatype_id,
                    "target_id": org1Service1.id,
                    "option": [
                    "read"
                    ],
                    "expirationTimestamp": 0
                },
                {
                    "datatype_id": org1Service1.datatypes[2].datatype_id,
                    "target_id": org1Service1.id,
                    "option": [
                    "write","read"
                    ],
                    "expirationTimestamp": 0
                }
            ]
        }
        chai.request(Config.server).post('omr/api/v1/users/patient_register_and_consent')
            .set('Accept', 'application/json')
            .set('token', Config.orgAdminToken1)
            .send(bodyRequest)
            .end(function(err, res){
                expect(res.status).to.equal(500);
                expect(res.body.failures).to.be.an('array');
                expect(res.body.failures).to.have.length(0);
                expect(res.body.successes).to.be.an('array');
                expect(res.body.successes).to.have.length(0);
                expect(res.body.status).to.equal(500);
                expect(res.body.failure_type).to.equal("Enrollment");
                done();
            });
    });

    it('Should have Registration fail (empty secret)', function (done) {
        const patient1 = idGenerator();
        const bodyRequest = {
            "id": patient1,
            "secret": "",
            "name": "patient1",
            "email": "patient1@example.com",
            "ca_org": Config.caOrg,
            "data": {},
            "service_id": org1Service1.id,
            "consents": 
            [
                {
                    "datatype_id": org1Service1.datatypes[0].datatype_id,
                    "target_id": org1Service1.id,
                    "option": [
                    "write"
                    ],
                    "expirationTimestamp": 0
                },
                {
                    "datatype_id": org1Service1.datatypes[1].datatype_id,
                    "target_id": org1Service1.id,
                    "option": [
                    "read"
                    ],
                    "expirationTimestamp": 0
                },
                {
                    "datatype_id": org1Service1.datatypes[2].datatype_id,
                    "target_id": org1Service1.id,
                    "option": [
                    "write","read"
                    ],
                    "expirationTimestamp": 0
                }
            ]
        }
        chai.request(Config.server).post('omr/api/v1/users/patient_register_and_consent')
            .set('Accept', 'application/json')
            .set('token', Config.orgAdminToken1)
            .send(bodyRequest)
            .end(function(err, res){
                expect(res.status).to.equal(400);
                expect(res.body.failures).to.be.an('array');
                expect(res.body.failures).to.have.length(0);
                expect(res.body.successes).to.be.an('array');
                expect(res.body.successes).to.have.length(0);
                expect(res.body.status).to.equal(400);
                expect(res.body.failure_type).to.equal("Registration");
                done();
            });
    });

    it('Should have Consent fail (invalid datatype_id)', function (done) {
        
        const bodyRequest = {
            "id": patient1Failed,
            "secret": "secret1",
            "name": "patient1Failed",
            "email": "patient1Failed@example.com",
            "ca_org": Config.caOrg,
            "data": {},
            "service_id": org1Service1.id,
            "consents": 
            [
                {
                    "datatype_id": "invalid_datatype_id",
                    "target_id": org1Service1.id,
                    "option": ["write"],
                    "expirationTimestamp": 0
                },
                {
                    "datatype_id": "invalid_datatype_id",
                    "target_id": org1Service1.id,
                    "option": ["read"],
                    "expirationTimestamp": 0
                },
                {
                    "datatype_id": "invalid_datatype_id",
                    "target_id": org1Service1.id,
                    "option": ["write","read"],
                    "expirationTimestamp": 0
                }
            ]
        }
        const consent1 = {
            "owner_id":patient1Failed,
            "service_id":org1Service1.id,
            "target_service_id":org1Service1.id,
            "datatype_id":"invalid_datatype_id",
            "option":["write"],
            "expirationTimestamp":0
        }
        const consent2 = {
            "owner_id":patient1Failed,
            "service_id":org1Service1.id,
            "target_service_id":org1Service1.id,
            "datatype_id":"invalid_datatype_id",
            "option":["read"],
            "expirationTimestamp":0
        }
        const consent3 = {
            "owner_id":patient1Failed,
            "service_id":org1Service1.id,
            "target_service_id":org1Service1.id,
            "datatype_id":"invalid_datatype_id",
            "option":["write","read"],
            "expirationTimestamp":0
        }
        chai.request(Config.server).post('omr/api/v1/users/patient_register_and_consent')
            .set('Accept', 'application/json')
            .set('token', Config.orgAdminToken1)
            .send(bodyRequest)
            .end(function(err, res){
                expect(res.status).to.equal(500);
                expect(res.body.successes).to.be.an('array');
                expect(res.body.successes).to.have.length(0);
                expect(res.body.failures).to.be.an('array');
                expect(res.body.failures).to.have.length(3);
                expect(res.body.failures).to.contain.something.like(consent1);
                expect(res.body.failures).to.contain.something.like(consent2);
                expect(res.body.failures).to.contain.something.like(consent3);
                expect(res.body.status).to.equal(500);
                expect(res.body.failure_type).to.equal("Consent");
                done();
            });
    });

    // get all consents for org1Service1/patient1Failed pair as org admin
    it('Should get blank list of consents for patient1Failed', function (done) {
        chai.request(Config.server).get('omr/api/v1/services/' + org1Service1.id + '/users/' + patient1Failed + '/consents')
            .set('Accept', 'application/json')
            .set('token', Config.orgAdminToken1)
            .end(function(err, res){
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.be.an('array');
                expect(res.body).to.have.lengthOf(0);
                done();
            });
    });

    // get all enrollments patient1Failed pair as org admin
    it('Should return a list of enrollments for patient1Failed', function (done) {
        chai.request(Config.server).get('omr/api/v1/users/' + patient1Failed + '/enrollments')
            .set('Accept', 'application/json')
            .set('token', Config.orgAdminToken1)
            .end(function(err, res){
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.be.an('array');
                expect(res.body).to.have.lengthOf(1);
                expect(res.body[0].service_id).to.equal(org1Service1.id);
                done();
            });
    });

    it('Should have Consent fail (Missing target ID)', function (done) {
        const patient1 = idGenerator();
        const bodyRequest = {
            "id": patient1,
            "secret": "secret1",
            "name": "patient1",
            "email": "patient1@example.com",
            "ca_org": Config.caOrg,
            "data": {},
            "service_id": org1Service1.id,
            "consents": 
            [
                {}
            ]
        }
            
        chai.request(Config.server).post('omr/api/v1/users/patient_register_and_consent')
            .set('Accept', 'application/json')
            .set('token', Config.orgAdminToken1)
            .send(bodyRequest)
            .end(function(err, res){
                expect(res.status).to.equal(400);
                expect(res.body.successes).to.be.an('array');
                expect(res.body.successes).to.have.length(0);
                expect(res.body.failures).to.be.an('array');
                expect(res.body.failures).to.have.length(0);
                expect(res.body.status).to.equal(400);
                expect(res.body.failure_type).to.equal("Consent");
                done();
            });
    });
});