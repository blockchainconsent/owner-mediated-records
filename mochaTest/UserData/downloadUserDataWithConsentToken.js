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

describe('Check of downloading of user data with consent token', function() {

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

    // create new org1Service1 admin
    const org1Service1Admin = {
        "id": idGenerator(),
        "secret": "org1Service1Admin",
        "name": "org1Service1Admin",
        "role": "user",
        "org" : Config.org1.id,
        "email":  "org1Service1Adminemail@example.com",
        "ca_org": Config.caOrg,
        "data": {
          "address": "2 User St"
        }
    };

    // create new org1Admin
    const org1Admin = {
        "id": idGenerator(),
        "secret": "org1Admin",
        "name": "org1Admin",
        "role": "user",
        "org" : Config.org1.id,
        "email":  "org1Adminemail@example.com",
        "ca_org": Config.caOrg,
        "data": {
          "address": "2 User St"
        }
    };

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

    // register org1Service1Admin
    before((done) => {
        chai.request(Config.server).post('omr/api/v1/users') 
            .set('Accept',  'application/json')
            .set('token', Config.orgAdminToken1)
            .send(org1Service1Admin)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("success");
                done();
            });
    });

    // register org1Admin
    before((done) => {
        chai.request(Config.server).post('omr/api/v1/users') 
            .set('Accept',  'application/json')
            .set('token', Config.orgAdminToken1)
            .send(org1Admin)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("success");
                done();
            });
    });

    // grant service admin permission to org1Service1Admin
    before((done) => {
        chai.request(Config.server).put('omr/api/v1/users/' + org1Service1Admin.id + '/permissions/services/' + org1Service1.id) 
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

    // grant org admin permission to org1Admin
    before((done) => {  
        chai.request(Config.server).put('omr/api/v1/users/' + org1Admin.id + '/permissions/admin/' + Config.org1.id) 
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

    let defaultServiceAdminToken = '';
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
                defaultServiceAdminToken = res.body.token;
                done();
            });
    });

    let org1Service1AdminToken = '';
    before((done) => {
        chai.request(Config.server).get('common/api/v1/login')
            .set('Accept', 'application/json')
            .set('user-id', org1Service1Admin.id)
            .set('password', org1Service1Admin.secret)
            .set('login-org', org1Service1Admin.ca_org)
            .set('login-channel', Config.channel)
            .set('signature', "")
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("token");
                expect(res.body.token).to.not.equal("");
                org1Service1AdminToken = res.body.token;
                done();
            });
    });

    let org1AdminToken = '';
    before((done) => {
        chai.request(Config.server).get('common/api/v1/login')
            .set('Accept', 'application/json')
            .set('user-id', org1Admin.id)
            .set('password', org1Admin.secret)
            .set('login-org', org1Admin.ca_org)
            .set('login-channel', Config.channel)
            .set('signature', "")
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("token");
                expect(res.body.token).to.not.equal("");
                org1AdminToken = res.body.token;
                done();
            });
    });

    // create consent to s1d1
    before((done) => {
        const bodyRequest = {
            "patient_id": newPatient1.id,
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

    // create consent to s1d2
    before((done) => {
        const bodyRequest = {
            "patient_id": newPatient1.id,
            "service_id": org1Service1.id,
            "target_id": org1Service1.id,
            "datatype_id": org1Service1.datatypes[1].datatype_id,
            "option": [
                "read","write"
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

    // Validate сonsent for Patient Data Access as default service admin of this service
    let dataAccessToken = '';
    it('Should successfully return access token which can be used to download data as service admin of this service', function (done) {
        chai.request(Config.server).get('omr/api/v1/services/' + org1Service1.id + '/users/' + newPatient1.id 
        + '/datatype/' + org1Service1.datatypes[0].datatype_id + '/validation/write')
            .set('Accept', 'application/json')
            .set('token', defaultServiceAdminToken)
            .end(function(err, res){
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body.owner).to.equal(newPatient1.id);
                expect(res.body.datatype).to.equal(org1Service1.datatypes[0].datatype_id);
                expect(res.body.target).to.equal(org1Service1.id);
                expect(res.body.requested_access).to.equal("write");
                expect(res.body.message).to.equal("permission granted");
                expect(res.body).to.have.property("token");
                expect(res.body.token).to.not.equal("");
                dataAccessToken = res.body.token;
                done();
            });
    });

    it('Should successfully upload user data as default service admin', function (done) {
        const bodyRequest = {
            "test data": "some test data text"
        }
        chai.request(Config.server).post('omr/api/v1/services/' + org1Service1.id + '/users/' +
         newPatient1.id + '/datatype/' + org1Service1.datatypes[0].datatype_id + '/upload')
            .set('Accept', 'application/json')
            .set('token', defaultServiceAdminToken)
            .send(bodyRequest)
            .end(function(err, res){
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("success");
                done();
            });
    });

    it('Should successfully download user data as default service admin with data access token', function (done) {
        chai.request(Config.server).get('omr/api/v1/userdata/download/' + dataAccessToken)
            .set('Accept', 'application/json')
            .set('token', defaultServiceAdminToken)
            .end(function(err, res){
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(JSON.stringify(res.body[0].data)).to.equal(JSON.stringify({
                    "test data": "some test data text"
                }));
                done();
            });
    });

    // Validate сonsent for Patient Data Access as patient
    let dataAccessTokenPatient = '';
    it('Should successfully return data access token as patient', function (done) {
        chai.request(Config.server).get('omr/api/v1/services/' + org1Service1.id + '/users/' + newPatient1.id 
        + '/datatype/' + org1Service1.datatypes[0].datatype_id + '/validation/read')
            .set('Accept', 'application/json')
            .set('token', newPatient1Token)
            .end(function(err, res){
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body.owner).to.equal(newPatient1.id);
                expect(res.body.datatype).to.equal(org1Service1.datatypes[0].datatype_id);
                expect(res.body.target).to.equal(org1Service1.id);
                expect(res.body.requested_access).to.equal("read");
                expect(res.body.message).to.equal("permission granted");
                expect(res.body).to.have.property("token");
                expect(res.body.token).to.not.equal("");
                dataAccessTokenPatient = res.body.token;
                done();
            });
    });

    it('Should successfully download user data as patient with data access token', function (done) {
        chai.request(Config.server).get('omr/api/v1/userdata/download/' + dataAccessTokenPatient)
            .set('Accept', 'application/json')
            .set('token', newPatient1Token)
            .end(function(err, res){
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(JSON.stringify(res.body[0].data)).to.equal(JSON.stringify({
                    "test data": "some test data text"
                }));
                done();
            });
    });

    // Validate сonsent for Patient Data Access as org user with service admin permission with read consent
    let org1Service1AdminDataAccessToken = '';
    it('Should successfully return access token which can be used to download data as service admin of this service', function (done) {
        chai.request(Config.server).get('omr/api/v1/services/' + org1Service1.id + '/users/' + newPatient1.id 
        + '/datatype/' + org1Service1.datatypes[1].datatype_id + '/validation/read')
            .set('Accept', 'application/json')
            .set('token', org1Service1AdminToken)
            .end(function(err, res){
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body.owner).to.equal(newPatient1.id);
                expect(res.body.datatype).to.equal(org1Service1.datatypes[1].datatype_id);
                expect(res.body.target).to.equal(org1Service1.id);
                expect(res.body.requested_access).to.equal("read");
                expect(res.body.message).to.equal("permission granted");
                expect(res.body).to.have.property("token");
                expect(res.body.token).to.not.equal("");
                org1Service1AdminDataAccessToken = res.body.token;
                done();
            });
    });

    it('Should NOT successfully upload user data as newPatient1', function (done) {
        const bodyRequest = {
            "test data": "some test data text org1Service1Admin"
        }
        chai.request(Config.server).post('omr/api/v1/services/' + org1Service1.id + '/users/' +
         newPatient1.id + '/datatype/' + org1Service1.datatypes[1].datatype_id + '/upload')
            .set('Accept', 'application/json')
            .set('token', newPatient1Token)
            .send(bodyRequest)
            .end(function(err, res){
                expect(err).to.be.null;
                expect(res.status).to.equal(500);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("error");
                done();
            });
    });

    it('Should successfully upload user data as default service admin', function (done) {
        const bodyRequest = {
            "test data": "some test data text org1Service1Admin"
        }
        chai.request(Config.server).post('omr/api/v1/services/' + org1Service1.id + '/users/' +
         newPatient1.id + '/datatype/' + org1Service1.datatypes[1].datatype_id + '/upload')
            .set('Accept', 'application/json')
            .set('token', defaultServiceAdminToken)
            .send(bodyRequest)
            .end(function(err, res){
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("success");
                done();
            });
    });

    it('Should successfully download user data as service admin with data access token', function (done) {
        chai.request(Config.server).get('omr/api/v1/userdata/download/' + org1Service1AdminDataAccessToken)
            .set('Accept', 'application/json')
            .set('token', org1Service1AdminToken)
            .end(function(err, res){
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(JSON.stringify(res.body[0].data)).to.equal(JSON.stringify({
                    "test data": "some test data text org1Service1Admin"
                }));
                done();
            });
    });

    // Validate сonsent for Patient Data Access as org user with org admin permission with read consent
    let org1AdminDataAccessToken = '';
    it('Should successfully return access token which can be used to download data as org admin of this org', function (done) {
        chai.request(Config.server).get('omr/api/v1/services/' + org1Service1.id + '/users/' + newPatient1.id 
        + '/datatype/' + org1Service1.datatypes[1].datatype_id + '/validation/read')
            .set('Accept', 'application/json')
            .set('token', org1AdminToken)
            .end(function(err, res){
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body.owner).to.equal(newPatient1.id);
                expect(res.body.datatype).to.equal(org1Service1.datatypes[1].datatype_id);
                expect(res.body.target).to.equal(org1Service1.id);
                expect(res.body.requested_access).to.equal("read");
                expect(res.body.message).to.equal("permission granted");
                expect(res.body).to.have.property("token");
                expect(res.body.token).to.not.equal("");
                org1AdminDataAccessToken = res.body.token;
                done();
            });
    });

    it('Should successfully download user data as org admin with data access token', function (done) {
        chai.request(Config.server).get('omr/api/v1/userdata/download/' + org1AdminDataAccessToken)
            .set('Accept', 'application/json')
            .set('token', org1AdminToken)
            .end(function(err, res){
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(JSON.stringify(res.body[0].data)).to.equal(JSON.stringify({
                    "test data": "some test data text org1Service1Admin"
                }));
                done();
            });
    });
});