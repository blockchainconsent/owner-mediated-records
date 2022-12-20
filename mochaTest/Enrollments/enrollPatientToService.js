const Config = require('../ConfigFile.js');
const chai = require('chai');
const { expect } = chai;
const chaiHttp = require('chai-http')
chai.use(chaiHttp);
chai.use(require('chai-like'));
const chaiThings = require('chai-things')
chai.use(chaiThings)
const idGenerator = require('../Utils/helper');

require('../Login/loginSysAdmin'); //login as sysadmin
require("../Organizations/OrgSetup/RegisterOrg1"); //creation of new org
require("../DataTypes/RegisterDatatypesSysAdmin"); // creation of datatypes as sysadmin
require("../Services/RegisterServiceOrg1Admin"); // creation of services
require("../Users/RegisterUsers"); //creation of 2 users(patients)
require("../Login/OrgLogin/Org1Login"); // login as OrgAdmin1

describe('Enroll a patient to service as org admin', function() {

    it('Should successfully be enrolled to service 2', function (done) {
        const bodyRequest = {
            "user": Config.user1.id,
            "service": Config.service2.id,
            "status": "active"
        }
        chai.request(Config.server).post('omr/api/v1/services/' + Config.service2.id + '/user/enroll')
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

    it('Should successfully be enrolled to service 2 one more time', function (done) {
        const bodyRequest = {
            "user": Config.user1.id,
            "service": Config.service2.id,
            "status": "active"
        }

        chai.request(Config.server).post('omr/api/v1/services/' + Config.service2.id + '/user/enroll')
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
});

require("../Login/User2Login");
describe('Enroll a patient to service as unauthorized user', function() {
    
    it('Should return 500 error when enroll as an unauthorized user', function (done) {
        const bodyRequest = {
            "user": Config.user2.id,
            "service": Config.service1.id,
            "status": "active"
        }
        
        chai.request(Config.server).post('omr/api/v1/services/' + Config.service1.id + '/user/enroll')
                .set('Accept', 'application/json')
                .set('token', Config.user2Token)
                .send(bodyRequest)
                .end(function(err, res){           
                    expect(err).to.be.null;
                    expect(res.status).to.equal(500);
                    expect(res.body).to.have.property("msg");
                    expect(res.body.msg).to.include("error");
                    done();
                });
    });
});

describe('Enroll a patient to service as sys admin', function() {

    it('Should return 500 error when enroll as a sys admin', function (done) {
        const bodyRequest = {
            "user": Config.user1.id,
            "service": Config.service1.id,
            "status": "active"
        }
        const msgForSysAdmin = 'Enroll patient error'
        chai.request(Config.server).post('omr/api/v1/services/' + Config.service1.id + '/user/enroll')
                .set('Accept', 'application/json')
                .set('token', Config.sysAdminToken)
                .send(bodyRequest)
                .end(function(err, res){          
                    expect(err).to.be.null;
                    expect(res.status).to.equal(500);
                    expect(res.body).to.have.property("msg");
                    expect(res.body.msg).to.include("error");
                    done();
                });
    });
});

// enroll a patient as service admin
describe('Enroll a patient to the service as service admin', function () {
    // create two new random users
    const newUser = {
        "id": idGenerator(),
        "secret": "newuser",
        "name": "New User: newUser",
        "role": "user",
        "org" : Config.org1.id,
        "email":  "newuseremail@example.com",
        "ca_org": Config.caOrg,
        "data": {
          "address": "2 User St"
        }
      };
    const newUser2 = {
        "id": idGenerator(),
        "secret": "newuser2",
        "name": "New User: newuser2",
        "role": "user",
        "org" : Config.org1.id,
        "email":  "newuser2email@example.com",
        "ca_org": Config.caOrg,
        "data": {
          "address": "2 User St"
        }
      };
    
    before((done) => {
        chai.request(Config.server).post('omr/api/v1/users') 
            .set('Accept',  'application/json')
            .set('token', Config.orgAdminToken1)
            .send(newUser)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("success");
                done();
            });
    });

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

    // and grant service admin permission to one of them
    before((done) => {
        
        chai.request(Config.server).put('omr/api/v1/users/' + newUser2.id + '/permissions/services/' + Config.service1.id) 
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
    var serviceAdminToken = '';
    it('Should successfully login as service admin', function (done) {
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
                expect(res.body).to.not.deep.equal({});
                expect(res.body.token).to.not.equal("");
                serviceAdminToken = res.body.token;
                done();
            });
    });

    it('Should be enrolled to service 1 successfully', function (done) {
        const bodyRequest = {
            "user": newUser.id,
            "service": Config.service1.id,
            "status": "active"
        }
        
        chai.request(Config.server).post('omr/api/v1/services/' + Config.service1.id + '/user/enroll')
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

    // login as default service admin
    var defaultServiceAdminToken = '';
    it('Should successfully login as service admin', function (done) {
        chai.request(Config.server).get('common/api/v1/login')
            .set('Accept', 'application/json')
            .set('user-id', Config.service1.id)
            .set('password', Config.service1.secret)
            .set('login-org', Config.service1.ca_org)
            .set('login-channel', Config.channel)
            .set('signature', "")
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.not.deep.equal({});
                expect(res.body.token).to.not.equal("");
                defaultServiceAdminToken = res.body.token;
                done();
            });
    });

    it('Should be enrolled to service 1 successfully as default service admin', function (done) {
        const bodyRequest = {
            "user": newUser.id,
            "service": Config.service1.id,
            "status": "active"
        }
        
        chai.request(Config.server).post('omr/api/v1/services/' + Config.service1.id + '/user/enroll')
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
});
