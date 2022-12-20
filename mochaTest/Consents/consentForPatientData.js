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
        "name": "New User: newUser1",
        "role": "user",
        "org" : "",
        "email":  "newuser1email@example.com",
        "ca_org": Config.caOrg,
        "data": {
          "address": "2 User St"
        }
    };
    
    // register newUser1 (doesn't belong to any Org, with role user)
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

    it('Should successfully create consent for org1Service1', function (done) {
        const bodyRequest = {
            "patient_id": newUser1.id,
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
            .set('token', newUser1Token)
            .send(bodyRequest)
            .end(function(err, res){           
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("success");
                done();
            });
    });

    it('Should return an error when an unauthorized user tries to create a consent', function (done) {
        const bodyRequest = {
            "patient_id": newUser1.id,
            "service_id": org1Service1.id,
            "target_id": org1Service1.id,
            "datatype_id": org1Service1.datatypes[0].datatype_id,
            "option": [
                "read"
            ],
            "expiration": 0
        }
        chai.request(Config.server).post('omr/api/v1/consents/patientData')
            .set('Accept', 'application/json')
            .set('token', Config.orgAdminToken1)
            .send(bodyRequest)
            .end(function(err, res){
                expect(err).to.be.null;
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("error");
                expect(res.status).to.equal(500);
                done();
            });
    });

    it('Should successfully revoke consent from org1Service1', function (done) {
        const bodyRequest = {
            "patient_id": newUser1.id,
            "service_id": org1Service1.id,
            "target_id": org1Service1.id,
            "datatype_id": org1Service1.datatypes[0].datatype_id,
            "option": [
                "deny"
            ],
            "expiration": 0
        }
        chai.request(Config.server).post('omr/api/v1/consents/patientData')
            .set('Accept', 'application/json')
            .set('token', newUser1Token)
            .send(bodyRequest)
            .end(function(err, res){
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("success");
                done();
            });
    });

    it('Should successfully create consent to datatype belongs to different services with read consent option', function (done) {
        const bodyRequest = {
            "patient_id": newUser1.id,
            "service_id": org1Service1.id,
            "target_id": org2Service1.id,
            "datatype_id": org1Service1.datatypes[1].datatype_id,
            "option": [
                "read"
            ],
            "expiration": 0
        }
        chai.request(Config.server).post('omr/api/v1/consents/patientData')
            .set('Accept', 'application/json')
            .set('token', newUser1Token)
            .send(bodyRequest)
            .end(function(err, res){
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("success");
                done();
            });
    });

    it('Should not successfully create consent to datatype belongs to different services with write consent option', function (done) {
        const bodyRequest = {
            "patient_id": newUser1.id,
            "service_id": org1Service1.id,
            "target_id": org2Service1.id,
            "datatype_id": org1Service1.datatypes[1].datatype_id,
            "option": [
                "write"
            ],
            "expiration": 0
        }
        chai.request(Config.server).post('omr/api/v1/consents/patientData')
            .set('Accept', 'application/json')
            .set('token', newUser1Token)
            .send(bodyRequest)
            .end(function(err, res){
                expect(err).to.be.null;
                expect(res.status).to.equal(500);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("error");
                done();
            });
    });

    // negative cases - patient_id does not exist
    it('Should not successfully create consent in case when field patient_id does not exist', function (done) {
        const bodyRequest = {
            // "patient_id": newUser1.id,
            "service_id": org1Service1.id,
            "target_id": org1Service1.id,
            "datatype_id": org1Service1.datatypes[1].datatype_id,
            "option": [
                "read"
            ],
            "expiration": 0
        }
        chai.request(Config.server).post('omr/api/v1/consents/patientData')
            .set('Accept', 'application/json')
            .set('token', newUser1Token)
            .send(bodyRequest)
            .end(function(err, res){
                expect(err).to.be.null;
                expect(res.status).to.equal(400);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.equal('Invalid data: patient_id missing');
                done();
            });
    });

    // negative cases - service_id does not exist
    it('Should not successfully create consent in case when field service_id does not exist', function (done) {
        const bodyRequest = {
            "patient_id": newUser1.id,
            // "service_id": org1Service1.id,
            "target_id": org1Service1.id,
            "datatype_id": org1Service1.datatypes[1].datatype_id,
            "option": [
                "read"
            ],
            "expiration": 0
        }
        chai.request(Config.server).post('omr/api/v1/consents/patientData')
            .set('Accept', 'application/json')
            .set('token', newUser1Token)
            .send(bodyRequest)
            .end(function(err, res){
                expect(err).to.be.null;
                expect(res.status).to.equal(400);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.equal('Invalid data:service_id missing');
                done();
            });
    });

    // negative cases - target_id does not exist
    it('Should not successfully create consent in case when field target_id does not exist', function (done) {
        const bodyRequest = {
            "patient_id": newUser1.id,
            "service_id": org1Service1.id,
            // "target_id": org2Service1.id,
            "datatype_id": org1Service1.datatypes[1].datatype_id,
            "option": [
                "read"
            ],
            "expiration": 0
        }
        chai.request(Config.server).post('omr/api/v1/consents/patientData')
            .set('Accept', 'application/json')
            .set('token', newUser1Token)
            .send(bodyRequest)
            .end(function(err, res){
                expect(err).to.be.null;
                expect(res.status).to.equal(400);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.equal('Invalid data: target_id missing');
                done();
            });
    });

    // negative cases - datatype_id does not exist
    it('Should not successfully create consent in case when field datatype_id does not exist', function (done) {
        const bodyRequest = {
            "patient_id": newUser1.id,
            "service_id": org1Service1.id,
            "target_id": org1Service1.id,
            // "datatype_id": org1Service1.datatypes[1].datatype_id,
            "option": [
                "read"
            ],
            "expiration": 0
        }
        chai.request(Config.server).post('omr/api/v1/consents/patientData')
            .set('Accept', 'application/json')
            .set('token', newUser1Token)
            .send(bodyRequest)
            .end(function(err, res){
                expect(err).to.be.null;
                expect(res.status).to.equal(400);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.equal('Invalid data: datatype_id missing');
                done();
            });
    });

    // negative cases - user with provided patient_id does not exist
    it('Should not successfully create consent in case when provided patient_id does not exist', function (done) {
        const bodyRequest = {
            "patient_id": "12dwed445",
            "service_id": org1Service1.id,
            "target_id": org1Service1.id,
            "datatype_id": org1Service1.datatypes[1].datatype_id,
            "option": [
                "write"
            ],
            "expiration": 0
        }
        chai.request(Config.server).post('omr/api/v1/consents/patientData')
            .set('Accept', 'application/json')
            .set('token', newUser1Token)
            .send(bodyRequest)
            .end(function(err, res){
                expect(err).to.be.null;
                expect(res.body).to.have.property("msg");
                expect(res.status).to.equal(500);
                expect(res.body.msg).to.include("error");
                done();
            });
    });

    // negative cases - service with provided service_id does not exist
    it('Should not successfully create consent in case when provided service_id does not exist', function (done) {
        const bodyRequest = {
            "patient_id": newUser1.id,
            "service_id": "12dwed445",
            "target_id": org1Service1.id,
            "datatype_id": org1Service1.datatypes[1].datatype_id,
            "option": [
                "read"
            ],
            "expiration": 0
        }
        chai.request(Config.server).post('omr/api/v1/consents/patientData')
            .set('Accept', 'application/json')
            .set('token', newUser1Token)
            .send(bodyRequest)
            .end(function(err, res){
                expect(err).to.be.null;
                expect(res.body).to.have.property("msg");
                expect(res.status).to.equal(500);
                expect(res.body.msg).to.include("error");
                done();
            });
    });

    // negative cases - target with provided target_id does not exist
    it('Should not successfully create consent in case when provided target_id does not exist', function (done) {
        const bodyRequest = {
            "patient_id": newUser1.id,
            "service_id": org1Service1.id,
            "target_id": "12dwed445",
            "datatype_id": org1Service1.datatypes[1].datatype_id,
            "option": [
                "read"
            ],
            "expiration": 0
        }
        chai.request(Config.server).post('omr/api/v1/consents/patientData')
            .set('Accept', 'application/json')
            .set('token', newUser1Token)
            .send(bodyRequest)
            .end(function(err, res){
                expect(err).to.be.null;
                expect(res.status).to.equal(500);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("error");
                done();
            });
    });

    // negative cases - datatype with provided datatype_id does not exist
    it('Should not successfully create consent in case when provided datatype_id does not exist', function (done) {
        const bodyRequest = {
            "patient_id": newUser1.id,
            "service_id": org1Service1.id,
            "target_id": org1Service1.id,
            "datatype_id": "12dwed445",
            "option": [
                "write"
            ],
            "expiration": 0
        }
        chai.request(Config.server).post('omr/api/v1/consents/patientData')
            .set('Accept', 'application/json')
            .set('token', newUser1Token)
            .send(bodyRequest)
            .end(function(err, res){
                expect(err).to.be.null;
                expect(res.status).to.equal(500);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("error");
                done();
            });
    });

    // negative cases - provided option is not allowed
    it('Should not successfully create consent in case when provided options are not in (write, read, deny)', function (done) {
        const bodyRequest = {
            "patient_id": newUser1.id,
            "service_id": org1Service1.id,
            "target_id": org1Service1.id,
            "datatype_id": org1Service1.datatypes[1].datatype_id,
            "option": [
                "scfcfwfw",
                "wdwd"
            ],
            "expiration": 0
        }
        chai.request(Config.server).post('omr/api/v1/consents/patientData')
            .set('Accept', 'application/json')
            .set('token', newUser1Token)
            .send(bodyRequest)
            .end(function(err, res){
                expect(err).to.be.null;
                expect(res.status).to.equal(400);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.equal('Invalid data: invalid consent option');
                done();
            });
    });
});
