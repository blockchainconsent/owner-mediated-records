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

describe('Check of uploading and downloading of user data', function() {

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
            "secret": "newservice2pass",
            "ca_org": Config.caOrg,
            "email": "newservice2email@example.com",
            "org_id": Config.org1.id,
            "summary": "New service under org 1. Has two datatypes",
            "terms": {
                "term1" : "example term",
                "term2" : "example term",
            },
            "payment_required": "yes",
            "datatypes":  [
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
                "sample data 3": "service sample data 3",
                "sample data 4": "service sample data 4"
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

    // enroll newUser1 to org1Service2
    before((done) => {
        const bodyRequest = {
            "user": newUser1.id,
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

    //enroll newUser1 to o2s1
    before((done) => {
        const bodyRequest = {
            "user": newUser1.id,
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

    // create consent to o1d1
    before((done) => {
        const bodyRequest = {
            "patient_id": newUser1.id,
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

    // create consent to o1d2
    before((done) => {
        const bodyRequest = {
            "patient_id": newUser1.id,
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
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("success");
                done();
            });
    });

    // create consent to datatype belongs to different services
    before((done) => {
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

    it('Should successfully upload user data as target service org1Service1', function (done) {
        const bodyRequest = {
            "test data": "some test data 1 for s1d1"
        }
        chai.request(Config.server).post('omr/api/v1/services/' + org1Service1.id + '/users/' +
         newUser1.id + '/datatype/' + org1Service1.datatypes[0].datatype_id + '/upload')
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

    // check that service admin of target service from another org is not able to upload user data
    //  in case when consent was given to datatype belongs to services from different orgs
    it('Should not successfully upload user data as service admin of another org', function (done) {
        const bodyRequest = {
            "test data": "test data for s1d2 uploaded by o2s1 admin"
        }
        chai.request(Config.server).post('omr/api/v1/services/' + org1Service1.id + '/users/' +
         newUser1.id + '/datatype/' + org1Service1.datatypes[1].datatype_id + '/upload')
            .set('Accept', 'application/json')
            .set('token', defaultOrg2Service1AdminToken)
            .send(bodyRequest)
            .end(function(err, res){
                expect(err).to.be.null;
                expect(res.status).to.equal(500);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("error");
                done();
            });
    });

    // check that org admin of another org is not able to upload user data
    //  in case when consent was given to datatype belongs to services from different orgs
    it('Should not successfully upload user data as org admin of another org', function (done) {
        const bodyRequest = {
            "test data": "test data for s1d2 uploaded by o2 org admin"
        }
        chai.request(Config.server).post('omr/api/v1/services/' + org1Service1.id + '/users/' +
         newUser1.id + '/datatype/' + org1Service1.datatypes[1].datatype_id + '/upload')
            .set('Accept', 'application/json')
            .set('token', Config.orgAdminToken2)
            .send(bodyRequest)
            .end(function(err, res){
                expect(err).to.be.null;
                expect(res.status).to.equal(500);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("error");
                done();
            });
    });

    it('Should not upload user data when consent is not given yet', function (done) {
        const bodyRequest = {
            "test data": "some test data 1 for s2d3"
        }
        chai.request(Config.server).post('omr/api/v1/services/' + org1Service2.id + '/users/' +
         newUser1.id + '/datatype/' + org1Service2.datatypes[0].datatype_id + '/upload')
            .set('Accept', 'application/json')
            .set('token', defaultOrg1Service1AdminToken)
            .send(bodyRequest)
            .end(function(err, res){
                expect(err).to.be.null;
                expect(res.status).to.equal(500);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("error");
                done();
            });
    });

    it('Should not upload user data as data owner (patient)', function (done) {
        const bodyRequest = {
            "test data": "some test data"
        }
        chai.request(Config.server).post('omr/api/v1/services/' + org1Service1.id + '/users/' +
         newUser1.id + '/datatype/' + org1Service1.datatypes[0].datatype_id + '/upload')
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

    it('Should successfully upload user data as target service org1Service1', function (done) {
        const bodyRequest = {
            "test data": "some test data 2 for s1d2"
        }
        chai.request(Config.server).post('omr/api/v1/services/' + org1Service1.id + '/users/' +
         newUser1.id + '/datatype/' + org1Service1.datatypes[1].datatype_id + '/upload')
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

    it('Should successfully upload user data as org admin', function (done) {
        const bodyRequest = {
            "test data 2": "another test data for s1d2"
        }
        chai.request(Config.server).post('omr/api/v1/services/' + org1Service1.id + '/users/' +
         newUser1.id + '/datatype/' + org1Service1.datatypes[1].datatype_id + '/upload')
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

    it('Should successfully download user data as default service admin when an option is write', function (done) {
        chai.request(Config.server).get('omr/api/v1/services/' + org1Service1.id + '/users/' +
         newUser1.id + '/datatypes/' + org1Service1.datatypes[1].datatype_id + '/download')
            .set('Accept', 'application/json')
            .set('token', defaultOrg1Service1AdminToken)
            .end(function(err, res){
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(JSON.stringify(res.body[0].data)).to.equal(JSON.stringify({
                    "test data": "some test data 2 for s1d2"
                }));
                expect(JSON.stringify(res.body[1].data)).to.equal(JSON.stringify({
                    "test data 2": "another test data for s1d2"
                }));
                done();
            });
    });

    it('Should successfully download user data as data owner when an option is write', function (done) {
        chai.request(Config.server).get('omr/api/v1/services/' + org1Service1.id + '/users/' +
         newUser1.id + '/datatypes/' + org1Service1.datatypes[1].datatype_id + '/download')
            .set('Accept', 'application/json')
            .set('token', newUser1Token)
            .end(function(err, res){
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(JSON.stringify(res.body[0].data)).to.equal(JSON.stringify({
                    "test data": "some test data 2 for s1d2"
                }));
                expect(JSON.stringify(res.body[1].data)).to.equal(JSON.stringify({
                    "test data 2": "another test data for s1d2"
                }));
                done();
            });
    });

    // download as org admin
    it('Should successfully download user data as org admin', function (done) {
        chai.request(Config.server).get('omr/api/v1/services/' + org1Service1.id + '/users/' +
         newUser1.id + '/datatypes/' + org1Service1.datatypes[1].datatype_id + '/download')
            .set('Accept', 'application/json')
            .set('token', Config.orgAdminToken1)
            .end(function(err, res){
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(JSON.stringify(res.body[0].data)).to.equal(JSON.stringify({
                    "test data": "some test data 2 for s1d2"
                }));
                expect(JSON.stringify(res.body[1].data)).to.equal(JSON.stringify({
                    "test data 2": "another test data for s1d2"
                }));
                done();
            });
    });

    // download as org admin of another org (when datatype belongs to services from different orgs and
    // consent was given to o2s1
    it('Should successfully download user data as org admin of another org with given consent', function (done) {
        chai.request(Config.server).get('omr/api/v1/services/' + org2Service1.id + '/users/' +
         newUser1.id + '/datatypes/' + org1Service1.datatypes[1].datatype_id + '/download')
            .set('Accept', 'application/json')
            .set('token', Config.orgAdminToken2)
            .end(function(err, res){
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(JSON.stringify(res.body[0].data)).to.equal(JSON.stringify({
                    "test data": "some test data 2 for s1d2"
                }));
                expect(JSON.stringify(res.body[1].data)).to.equal(JSON.stringify({
                    "test data 2": "another test data for s1d2"
                }));
                done();
            });
    });

    // download as service admin of another org (when datatype belongs to services from different orgs and
    // consent was given to o2s1
    it('Should successfully download user data as service admin of another org with given consent', function (done) {
        chai.request(Config.server).get('omr/api/v1/services/' + org2Service1.id + '/users/' +
         newUser1.id + '/datatypes/' + org1Service1.datatypes[1].datatype_id + '/download')
            .set('Accept', 'application/json')
            .set('token', defaultOrg2Service1AdminToken)
            .end(function(err, res){
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(JSON.stringify(res.body[0].data)).to.equal(JSON.stringify({
                    "test data": "some test data 2 for s1d2"
                }));
                expect(JSON.stringify(res.body[1].data)).to.equal(JSON.stringify({
                    "test data 2": "another test data for s1d2"
                }));
                done();
            });
    });

    it('Should revoke consent from s1d2', function (done) {
        const bodyRequest = {
            "patient_id": newUser1.id,
            "service_id": org1Service1.id,
            "target_id": org1Service1.id,
            "datatype_id": org1Service1.datatypes[1].datatype_id,
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

    it('Should not successfully download user data as default service admin when an option is deny', function (done) {
        chai.request(Config.server).get('omr/api/v1/services/' + org1Service1.id + '/users/' +
         newUser1.id + '/datatypes/' + org1Service1.datatypes[1].datatype_id + '/download')
            .set('Accept', 'application/json')
            .set('token', defaultOrg1Service1AdminToken)
            .end(function(err, res){          
                expect(err).to.be.null;
                expect(res.status).to.equal(500);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("error");
                done();
            });
    });

    // Check that service admin cannot upload data after consent changes to "deny"
    it('Should not successfully upload user data as service admin after consent changes to "deny"', function (done) {
        const bodyRequest = {
            "test data 3": "some test data 3 for s1d2"
        }
        chai.request(Config.server).post('omr/api/v1/services/' + org1Service1.id + '/users/' +
         newUser1.id + '/datatype/' + org1Service1.datatypes[1].datatype_id + '/upload')
            .set('Accept', 'application/json')
            .set('token', defaultOrg1Service1AdminToken)
            .send(bodyRequest)
            .end(function(err, res){
                expect(err).to.be.null;
                expect(res.status).to.equal(500);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("error");
                done();
            });
    });

    // Check that org admin cannot upload data after consent changes to "deny"
    it('Should not successfully upload user data as org admin after consent changes to "deny"', function (done) {
        const bodyRequest = {
            "test data 3": "some test data 3 for s1d2"
        }
        chai.request(Config.server).post('omr/api/v1/services/' + org1Service1.id + '/users/' +
         newUser1.id + '/datatype/' + org1Service1.datatypes[1].datatype_id + '/upload')
            .set('Accept', 'application/json')
            .set('token', Config.orgAdminToken1)
            .send(bodyRequest)
            .end(function(err, res){
                expect(err).to.be.null;
                expect(res.status).to.equal(500);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("error");
                done();
            });
    });

    it('Should still successfully download user data as data owner when an option is deny', function (done) {
        chai.request(Config.server).get('omr/api/v1/services/' + org1Service1.id + '/users/' +
         newUser1.id + '/datatypes/' + org1Service1.datatypes[1].datatype_id + '/download')
            .set('Accept', 'application/json')
            .set('token', newUser1Token)
            .end(function(err, res){
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(JSON.stringify(res.body[0].data)).to.equal(JSON.stringify({
                    "test data": "some test data 2 for s1d2"
                }));
                expect(JSON.stringify(res.body[1].data)).to.equal(JSON.stringify({
                    "test data 2": "another test data for s1d2"
                }));
                done();
            });
    });
});


// Check that org users with permissions can also access consent just like default admins
// Check that another patient cannot change the consent of a different patient
describe('Check of uploading and downloading of user data as org users and another patients', function() {

    // register org1Service1
    let org1Service3 = {};    
    before((done) => {
        org1Service3 = {
            "id": idGenerator(),
            "name": "o1s3",
            "secret": "newservice3pass",
            "ca_org": Config.caOrg,
            "email": "newservice3email@example.com",
            "org_id": Config.org1.id,
            "summary": "New service under org 1. Has 3 datatypes",
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
                        "write"
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
            .send(org1Service3)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("success");
                done();
                });
    });

    // register org1Service2
    let org1Service4 = {};    
    before((done) => {
        org1Service4 = {
            "id": idGenerator(),
            "name": "o1s4",
            "secret": "newservice4pass",
            "ca_org": Config.caOrg,
            "email": "newservice4email@example.com",
            "org_id": Config.org1.id,
            "summary": "New service under org 1. Has 1 datatype",
            "terms": {
                "term4" : "example term"
            },
            "payment_required": "yes",
            "datatypes":  [
                {
                    "datatype_id": Config.datatype4.id,
                    "access":[
                        "write",
                        "read"
                    ]
                }
            ],
            "solution_private_data": {
                "sample data 4": "service sample data 4"
            }
        };

        chai.request(Config.server).post('omr/api/v1/services') 
            .set('Accept',  'application/json')
            .set('token', Config.orgAdminToken1)
            .send(org1Service4)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("success");
                done();
                });
    });

    // create new random newUser1 which doesn't belong to any org
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

    // create newUser2 which belongs to org1
    const newUser2 = {
        "id": idGenerator(),
        "secret": "newuser2",
        "name": "OrgAdmin",
        "role": "user",
        "org" : Config.org1.id,
        "email":  "newuser2email@example.com",
        "ca_org": Config.caOrg,
        "data": {
          "address": "2 User St"
        }
    };

    // create newUser3 which belongs to org1
    const newUser3 = {
        "id": idGenerator(),
        "secret": "newuser3",
        "name": "ServiceAdmin",
        "role": "user",
        "org" : Config.org1.id,
        "email":  "newuser4email@example.com",
        "ca_org": Config.caOrg,
        "data": {
          "address": "2 User St"
        }
    };

    // create random user4 which doesn't belong to any org
    const newUser4 = {
        "id": idGenerator(),
        "secret": "newuser4",
        "name": "New User4",
        "role": "user",
        "org" : "",
        "email":  "newuser4email@example.com",
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

    // register new user2 (belongs to Org1, with role user)
    before((done) => {
        chai.request(Config.server).post('omr/api/v1/users') 
            .set('Accept',  'application/json')
            .set('token', Config.orgAdminToken1)
            .send(newUser2)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("success");
                done();
            });
    });

    // register new user3 (belongs to Org1, with role user)
    before((done) => {
        chai.request(Config.server).post('omr/api/v1/users') 
            .set('Accept',  'application/json')
            .set('token', Config.orgAdminToken1)
            .send(newUser3)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("success");
                done();
            });
    });

    // register newUser4 (doesn't belong to any Org, with role user)
    before((done) => {
        chai.request(Config.server).post('omr/api/v1/users') 
            .set('Accept',  'application/json')
            .set('token', Config.sysAdminToken)
            .send(newUser4)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("success");
                done();
            });
    });

    // grant service admin permission to newUser3
    before((done) => {
        chai.request(Config.server).put('omr/api/v1/users/' + newUser3.id + '/permissions/services/' + org1Service3.id) 
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

    // grant org admin permission to newUser2
    before((done) => {  
        chai.request(Config.server).put('omr/api/v1/users/' + newUser2.id + '/permissions/admin/' + Config.org1.id) 
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

    // enroll newUser1 to org1Service3
    before((done) => {
        const bodyRequest = {
            "user": newUser1.id,
            "service": org1Service3.id,
            "status": "active"
        }
        chai.request(Config.server).post('omr/api/v1/services/' + org1Service3.id + '/user/enroll')
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

    // enroll newUser4 also to org1Service3
    before((done) => {
        const bodyRequest = {
            "user": newUser1.id,
            "service": org1Service3.id,
            "status": "active"
        }
        chai.request(Config.server).post('omr/api/v1/services/' + org1Service3.id + '/user/enroll')
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

    let newUser4Token = '';
    before((done) => {
        chai.request(Config.server).get('common/api/v1/login')
            .set('Accept', 'application/json')
            .set('user-id', newUser4.id)
            .set('password', newUser4.secret)
            .set('login-org', newUser4.ca_org)
            .set('login-channel', Config.channel)
            .set('signature', "")
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("token");
                expect(res.body.token).to.not.equal("");
                newUser4Token = res.body.token;
                done();
            });
    });

    let serviceAdminToken = '';
    before((done) => {
        chai.request(Config.server).get('common/api/v1/login')
            .set('Accept', 'application/json')
            .set('user-id', newUser3.id)
            .set('password', newUser3.secret)
            .set('login-org', newUser3.ca_org)
            .set('login-channel', Config.channel)
            .set('signature', "")
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("token");
                expect(res.body.token).to.not.equal("");
                serviceAdminToken = res.body.token;
                done();
            });
    });

    let orgAdminToken = '';
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
                orgAdminToken = res.body.token;
                done();
            });
    });

    // create write consent to s1d3
    before((done) => {
        const bodyRequest = {
            "patient_id": newUser1.id,
            "service_id": org1Service3.id,
            "target_id": org1Service3.id,
            "datatype_id": org1Service3.datatypes[2].datatype_id,
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
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("success");
                done();
            });
    });

    it('Should successfully upload user data as org admin', function (done) {
        const bodyRequest = {
            "org admin test data": "test data 1 for s3d3"
        }
        chai.request(Config.server).post('omr/api/v1/services/' + org1Service3.id + '/users/' +
         newUser1.id + '/datatype/' + org1Service3.datatypes[2].datatype_id + '/upload')
            .set('Accept', 'application/json')
            .set('token', orgAdminToken)
            .send(bodyRequest)
            .end(function(err, res){
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("success");
                done();
            });
    });

    it('Should successfully upload user data as service admin', function (done) {
        const bodyRequest = {
            "service admin test data": "test data 2 for s3d3"
        }
        chai.request(Config.server).post('omr/api/v1/services/' + org1Service3.id + '/users/' +
         newUser1.id + '/datatype/' + org1Service3.datatypes[2].datatype_id + '/upload')
            .set('Accept', 'application/json')
            .set('token', serviceAdminToken)
            .send(bodyRequest)
            .end(function(err, res){
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("success");
                done();
            });
    });

    it('Should not have access to upload user data as another user enrolled to the same service', function (done) {
        const bodyRequest = {
            "service admin test data": "test data 2 for s3d3"
        }
        chai.request(Config.server).post('omr/api/v1/services/' + org1Service3.id + '/users/' +
         newUser1.id + '/datatype/' + org1Service3.datatypes[2].datatype_id + '/upload')
            .set('Accept', 'application/json')
            .set('token', newUser4Token)
            .send(bodyRequest)
            .end(function(err, res){
                expect(err).to.be.null;
                expect(res.status).to.equal(500);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("error");
                done();
            });
    });
   
    // download as org user with service admin permission
    it('Should successfully download user data as org user with service admin permission when an option is write', function (done) {
        chai.request(Config.server).get('omr/api/v1/services/' + org1Service3.id + '/users/' +
         newUser1.id + '/datatypes/' + org1Service3.datatypes[2].datatype_id + '/download')
            .set('Accept', 'application/json')
            .set('token', serviceAdminToken)
            .end(function(err, res){
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(JSON.stringify(res.body[0].data)).to.equal(JSON.stringify({
                    "org admin test data": "test data 1 for s3d3"
                }));
                expect(JSON.stringify(res.body[1].data)).to.equal(JSON.stringify({
                    "service admin test data": "test data 2 for s3d3"
                }));
                done();
            });
    });

    // download as org user with org admin permission
    it('Should successfully download user data as org user with org admin permission', function (done) {
        chai.request(Config.server).get('omr/api/v1/services/' + org1Service3.id + '/users/' +
         newUser1.id + '/datatypes/' + org1Service3.datatypes[2].datatype_id + '/download')
            .set('Accept', 'application/json')
            .set('token', orgAdminToken)
            .end(function(err, res){
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(JSON.stringify(res.body[0].data)).to.equal(JSON.stringify({
                    "org admin test data": "test data 1 for s3d3"
                }));
                expect(JSON.stringify(res.body[1].data)).to.equal(JSON.stringify({
                    "service admin test data": "test data 2 for s3d3"
                }));
                done();
            });
    });
});