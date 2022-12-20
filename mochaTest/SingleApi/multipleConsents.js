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

describe('Check of creating and updating consents by user', function() {

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
            "name": "newService",
            "secret": "newservicepass",
            "ca_org": Config.caOrg,
            "email": "newserviceemail@example.com",
            "org_id": Config.org1.id,
            "summary": "New service under org 1. Has 4 datatypes",
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
                },
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
                        "write",
                        "read"
                    ]
                }
            ],
            "solution_private_data": {
                "sample data 1": "service sample data 1",
                "sample data 2": "service sample data 2",
                "sample data 3": "service sample data 3",
                "sample data 4": "service sample data 4"
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
            "summary": "New service under org 2. Has two datatypes",
            "terms": {
                "term1" : "example term",
                "term2" : "example term",
            },
            "payment_required": "yes",
            "datatypes": [
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
                        "write"
                    ]
                }
            ],
            "solution_private_data": {
                "sample data 1": "service sample data 2",
                "sample data 2": "service sample data 3"
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

    // create new random user
    const newUser1 = {
        "id": idGenerator(),
        "secret": "newuser1",
        "name": "newUser1",
        "role": "user",
        "org" : "",
        "email":  "newuser1email@example.com",
        "ca_org": Config.caOrg,
        "data": {
          "address": "2 User St"
        }
    };

    // create new random user2
    const newUser2 = {
        "id": idGenerator(),
        "secret": "newuser2",
        "name": "newUser2",
        "role": "user",
        "org" : "",
        "email":  "newuser2email@example.com",
        "ca_org": Config.caOrg,
        "data": {
          "address": "2 User St 1"
        }
    };
    
    // register newUser1
    before((done) => {
        chai.request(Config.server).post('omr/api/v1/users') 
            .set('Accept',  'application/json')
            .set('token', Config.sysAdminToken)
            .send(newUser1)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("success");
                done();
            });
    });

    // register newUser2
    before((done) => {
        chai.request(Config.server).post('omr/api/v1/users') 
            .set('Accept',  'application/json')
            .set('token', Config.sysAdminToken)
            .send(newUser2)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("success");
                done();
            });
    });

    // enroll newUser1 to org1Service1
    before((done) => {
        const bodyRequest = {
            "user": newUser1.id,
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

    // enroll newUser2 to org1Service1
    before((done) => {
        const bodyRequest = {
            "user": newUser2.id,
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

    let newUser1Token = '';
    before((done) => {
        chai.request(Config.server).get('common/api/v1/login')
            .set('Accept', 'application/json')
            .set('user-id', newUser1.id)
            .set('password', newUser1.secret)
            .set('login-org', newUser1.ca_org)
            .set('login-channel', Config.channel)
            .set('signature', "")
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("token");
                expect(res.body.token).to.not.equal("");
                newUser1Token = res.body.token;
                done();
            });
    });

    let newUser2Token = '';
    before((done) => {
        chai.request(Config.server).get('common/api/v1/login')
            .set('Accept', 'application/json')
            .set('user-id', newUser2.id)
            .set('password', newUser2.secret)
            .set('login-org', newUser2.ca_org)
            .set('login-channel', Config.channel)
            .set('signature', "")
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("token");
                expect(res.body.token).to.not.equal("");
                newUser2Token = res.body.token;
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

    // according to documentation: "A patient can create consent for a service to view its patient data".
    // So org users (org admin, service admin) should not be able to call this API too.
    it('Should NOT successfully give consents via multiple consents API by org admin', function (done) {
        const bodyRequest = [
        {
            "patient_id": newUser1.id,
            "service_id": org1Service1.id,
            "target_id": org1Service1.id,
            "datatype_id": org1Service1.datatypes[0].datatype_id,
            "option": [
                "write",
                "read"
            ]
        }
        ]
        chai.request(Config.server).post('omr/api/v1/consents/multi-consent/patientData')
            .set('Accept', 'application/json')
            .set('token', Config.orgAdminToken1)
            .send(bodyRequest)
            .end(function(err, res){          
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body.successes).to.be.an('array');
                expect(res.body.successes).to.have.length(0);
                expect(res.body.failures).to.be.an('array');
                expect(res.body.failures).to.have.length(1);
                expect(res.body.failures[0].owner_id).to.equal(newUser1.id);
                expect(res.body.failures[0].service_id).to.equal(org1Service1.id);
                expect(res.body.failures[0].target_service_id).to.equal(org1Service1.id);
                expect(res.body.failures[0].datatype_id).to.equal(org1Service1.datatypes[0].datatype_id);
                expect(res.body.failures[0].option[0]).to.equal("write");
                expect(res.body.failures[0].option[1]).to.equal("read");
                done();
            });
    });

    // Only patient itself should be able to give consent. Another patient should not be able to give consent on behalf of somebody.
    it('Should NOT successfully give consents via multiple consents API by another user', function (done) {
        const bodyRequest = [
        {
            "patient_id": newUser1.id,
            "service_id": org1Service1.id,
            "target_id": org1Service1.id,
            "datatype_id": org1Service1.datatypes[0].datatype_id,
            "option": [
                "write",
                "read"
            ]
        }
        ]
        chai.request(Config.server).post('omr/api/v1/consents/multi-consent/patientData')
            .set('Accept', 'application/json')
            .set('token', newUser2Token)
            .send(bodyRequest)
            .end(function(err, res){ 
                expect(res.status).to.equal(200);
                expect(res.body.successes).to.be.an('array');
                expect(res.body.successes).to.have.length(0);
                expect(res.body.failures).to.be.an('array');
                expect(res.body.failures).to.have.length(1);
                expect(res.body.failures[0].owner_id).to.equal(newUser1.id);
                expect(res.body.failures[0].service_id).to.equal(org1Service1.id);
                expect(res.body.failures[0].target_service_id).to.equal(org1Service1.id);
                expect(res.body.failures[0].datatype_id).to.equal(org1Service1.datatypes[0].datatype_id);
                expect(res.body.failures[0].option[0]).to.equal("write");
                expect(res.body.failures[0].option[1]).to.equal("read");
                done();
            });
    });

    // according to documentation: "A patient can create consent for a service to view its patient data".
    // So org users (org admin, service admin) should not be able to call this API too.
    it('Should NOT successfully give consents via multiple consents API by service admin', function (done) {
        const bodyRequest = [
        {
            "patient_id": newUser1.id,
            "service_id": org1Service1.id,
            "target_id": org1Service1.id,
            "datatype_id": org1Service1.datatypes[0].datatype_id,
            "option": [
                "write",
                "read"
            ]
        }
        ]
        chai.request(Config.server).post('omr/api/v1/consents/multi-consent/patientData')
            .set('Accept', 'application/json')
            .set('token', o1s1adminToken)
            .send(bodyRequest)
            .end(function(err, res){ 
                expect(res.status).to.equal(200);
                expect(res.body.successes).to.be.an('array');
                expect(res.body.successes).to.have.length(0);
                expect(res.body.failures).to.be.an('array');
                expect(res.body.failures).to.have.length(1);
                expect(res.body.failures[0].owner_id).to.equal(newUser1.id);
                expect(res.body.failures[0].service_id).to.equal(org1Service1.id);
                expect(res.body.failures[0].target_service_id).to.equal(org1Service1.id);
                expect(res.body.failures[0].datatype_id).to.equal(org1Service1.datatypes[0].datatype_id);
                expect(res.body.failures[0].option[0]).to.equal("write");
                expect(res.body.failures[0].option[1]).to.equal("read");
                done();
            });
    });

    it('Should successfully give 1 consent via multiple consents API by newUser1', function (done) {
        const bodyRequest = [
        {
            "patient_id": newUser1.id,
            "service_id": org1Service1.id,
            "target_id": org1Service1.id,
            "datatype_id": org1Service1.datatypes[0].datatype_id,
            "option": [
                "write"
            ]
        }
        ]
        chai.request(Config.server).post('omr/api/v1/consents/multi-consent/patientData')
            .set('Accept', 'application/json')
            .set('token', newUser1Token)
            .send(bodyRequest)
            .end(function(err, res){ 
                expect(res.status).to.equal(200);
                expect(res.body.failures).to.be.an('array');
                expect(res.body.failures).to.have.length(0);
                expect(res.body.successes).to.be.an('array');
                expect(res.body.successes).to.have.length(1);
                expect(res.body.successes[0].owner_id).to.equal(newUser1.id);
                expect(res.body.successes[0].service_id).to.equal(org1Service1.id);
                expect(res.body.successes[0].target_service_id).to.equal(org1Service1.id);
                expect(res.body.successes[0].datatype_id).to.equal(org1Service1.datatypes[0].datatype_id);
                expect(res.body.successes[0].option[0]).to.equal("write");
                done();
            });
    });

    it('Should successfully create multiple consents equal to number of datatypes of this service by newUser1', function (done) {
        const bodyRequest = [
        {
            "patient_id": newUser1.id,
            "service_id": org1Service1.id,
            "target_id": org1Service1.id,
            "datatype_id": org1Service1.datatypes[0].datatype_id,
            "option": [
                "write"
            ]
        },
        {
            "patient_id": newUser1.id,
            "service_id": org1Service1.id,
            "target_id": org1Service1.id,
            "datatype_id": org1Service1.datatypes[1].datatype_id,
            "option": [
                "read"
            ],
            "expiration": 0
        },
        {
            "patient_id": newUser1.id,
            "service_id": org1Service1.id,
            "target_id": org1Service1.id,
            "datatype_id": org1Service1.datatypes[2].datatype_id,
            "option": [
                "read"
            ],
            "expiration": 0
        },
        {
            "patient_id": newUser1.id,
            "service_id": org1Service1.id,
            "target_id": org1Service1.id,
            "datatype_id": org1Service1.datatypes[3].datatype_id,
            "option": [
                "write",
                "read"
            ],
            "expiration": 0
        }
        ]

        const consent1 = {
            "owner_id":newUser1.id,
            "service_id":org1Service1.id,
            "target_service_id":org1Service1.id,
            "datatype_id":org1Service1.datatypes[0].datatype_id,
            "option":["write"],
            "expiration":0
        }
        const consent2 = {
            "owner_id":newUser1.id,
            "service_id":org1Service1.id,
            "target_service_id":org1Service1.id,
            "datatype_id":org1Service1.datatypes[1].datatype_id,
            "option":["read"],
            "expiration":0
        }
        const consent3 = {
            "owner_id":newUser1.id,
            "service_id":org1Service1.id,
            "target_service_id":org1Service1.id,
            "datatype_id":org1Service1.datatypes[2].datatype_id,
            "option":["read"],
            "expiration":0
        }
        const consent4 = {
            "owner_id":newUser1.id,
            "service_id":org1Service1.id,
            "target_service_id":org1Service1.id,
            "datatype_id":org1Service1.datatypes[3].datatype_id,
            "option":["write","read"],
            "expiration":0
        }
        chai.request(Config.server).post('omr/api/v1/consents/multi-consent/patientData')
            .set('Accept', 'application/json')
            .set('token', newUser1Token)
            .send(bodyRequest)
            .end(function(err, res){
                expect(res.status).to.equal(200);
                expect(res.body.failures).to.be.an('array');
                expect(res.body.failures).to.have.length(0);
                expect(res.body.successes).to.be.an('array');
                expect(res.body.successes).to.have.length(4);
                expect(res.body.successes).to.contain.something.like(consent1);
                expect(res.body.successes).to.contain.something.like(consent2);
                expect(res.body.successes).to.contain.something.like(consent3);
                expect(res.body.successes).to.contain.something.like(consent4);
                done();
            });
    });

    it('Should successfully deny all consents given before via multiple consents API by newUser1', function (done) {
        const bodyRequest = [{
            "patient_id": newUser1.id,
            "service_id": org1Service1.id,
            "target_id": org1Service1.id,
            "datatype_id": org1Service1.datatypes[0].datatype_id,
            "option": [
                "deny"
            ]
        },
        {
            "patient_id": newUser1.id,
            "service_id": org1Service1.id,
            "target_id": org1Service1.id,
            "datatype_id": org1Service1.datatypes[1].datatype_id,
            "option": [
                "deny"
            ],
            "expiration": 0
        },
        {
            "patient_id": newUser1.id,
            "service_id": org1Service1.id,
            "target_id": org1Service1.id,
            "datatype_id": org1Service1.datatypes[2].datatype_id,
            "option": [
                "deny"
            ],
            "expiration": 0
        },
        {
            "patient_id": newUser1.id,
            "service_id": org1Service1.id,
            "target_id": org1Service1.id,
            "datatype_id": org1Service1.datatypes[3].datatype_id,
            "option": [
                "deny"
            ],
            "expiration": 0
        }
        ]
        const consent1 = {
            "owner_id":newUser1.id,
            "service_id":org1Service1.id,
            "target_service_id":org1Service1.id,
            "datatype_id":org1Service1.datatypes[0].datatype_id,
            "option":["deny"],
            "expiration":0
        }
        const consent2 = {
            "owner_id":newUser1.id,
            "service_id":org1Service1.id,
            "target_service_id":org1Service1.id,
            "datatype_id":org1Service1.datatypes[1].datatype_id,
            "option":["deny"],
            "expiration":0
        }
        const consent3 = {
            "owner_id":newUser1.id,
            "service_id":org1Service1.id,
            "target_service_id":org1Service1.id,
            "datatype_id":org1Service1.datatypes[2].datatype_id,
            "option":["deny"],
            "expiration":0
        }
        const consent4 = {
            "owner_id":newUser1.id,
            "service_id":org1Service1.id,
            "target_service_id":org1Service1.id,
            "datatype_id":org1Service1.datatypes[3].datatype_id,
            "option":["deny"],
            "expiration":0
        }
        chai.request(Config.server).post('omr/api/v1/consents/multi-consent/patientData')
            .set('Accept', 'application/json')
            .set('token', newUser1Token)
            .send(bodyRequest)
            .end(function(err, res){
                expect(res.status).to.equal(200); 
                expect(res.body.failures).to.be.an('array');
                expect(res.body.failures).to.have.length(0);
                expect(res.body.successes).to.be.an('array');
                expect(res.body.successes).to.have.length(4);
                expect(res.body.successes).to.contain.something.like(consent1);
                expect(res.body.successes).to.contain.something.like(consent2);
                expect(res.body.successes).to.contain.something.like(consent3);
                expect(res.body.successes).to.contain.something.like(consent4);
                done();
            });
    });
});
