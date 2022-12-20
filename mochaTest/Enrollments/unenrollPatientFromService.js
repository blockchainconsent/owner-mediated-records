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
require("../DataTypes/RegisterDatatypesSysAdmin");
require("../Services/RegisterServiceOrg1Admin");
require("../Users/RegisterUsers");
require("../Login/OrgLogin/Org1Login");

describe('Unenroll a patient from service as org admin', function() {

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

    it('Should successfully be enrolled to service 2', function (done) {
        const bodyRequest = {
            "user": newUser.id,
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
                    expect(res.body.tx_id).to.not.equal("")
                    done();
                });
    });

    it('Should return an enrollment entry for new user with active status', function (done) {
        
        chai.request(Config.server).get('omr/api/v1/users/' + newUser.id + '/enrollments')
                .set('Accept', 'application/json')
                .set('token', Config.orgAdminToken1)
                .end(function(err, res){          
                    expect(err).to.be.null;
                    expect(res.status).to.equal(200);
                    expect(res.body).to.be.an('array');
                    expect(res.body).to.have.lengthOf(1);
                    expect(res.body[0].status).to.equal('active');
                    done();
                });
    });

    it('Should return an error when unenroll as an unauthorized user from a service', function (done) {
        const bodyRequest = {
            "user": newUser.id,
            "service": Config.service2.id,
            "status": "inactive"
        }
        chai.request(Config.server).post('omr/api/v1/services/' + Config.service2.id + '/user/unenroll')
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

    it('Should successfully be unenrolled from service 2', function (done) {
        const bodyRequest = {
            "user": newUser.id,
            "service": Config.service2.id,
            "status": "inactive"
        }
        chai.request(Config.server).post('omr/api/v1/services/' + Config.service2.id + '/user/unenroll')
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

    it('Should return an enrollment entry for new user with inactive status', function (done) {  
        chai.request(Config.server).get('omr/api/v1/users/' + newUser.id + '/enrollments')
                .set('Accept', 'application/json')
                .set('token', Config.orgAdminToken1)
                .end(function(err, res){          
                    expect(err).to.be.null;
                    expect(res.status).to.equal(200);
                    expect(res.body).to.be.an('array');
                    expect(res.body).to.have.lengthOf(1);
                    expect(res.body[0].status).to.equal('inactive');
                    done();
                });
    });

    it('Should successfully be unenrolled from service 2', function (done) {
        const bodyRequest = {
            "user": newUser.id,
            "service": Config.service2.id,
            "status": "inactive"
        }
        chai.request(Config.server).post('omr/api/v1/services/' + Config.service2.id + '/user/unenroll')
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

// unenroll a patient from a service as service admin
describe('Unenroll a patient from the service as service admin', function () {
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
        "name": "New User: newUser2",
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

    // grant service admin permission to one of them
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
    let serviceAdminToken = '';
    it('Should return a 200 test response', function (done) {
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

    // unenroll patient from service
    it('Should successfully be unenrolled from service 2', function (done) {
        const bodyRequest = {
            "user": newUser.id,
            "service": Config.service1.id,
            "status": "inactive"
        }   
        chai.request(Config.server).post('omr/api/v1/services/' + Config.service1.id + '/user/unenroll')
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

    it('Should return an enrollment entry for newUser with inactive status', function (done) {
        chai.request(Config.server).get('omr/api/v1/users/' + newUser.id + '/enrollments')
                .set('Accept', 'application/json')
                .set('token', serviceAdminToken)
                .end(function(err, res){ 
                    expect(err).to.be.null;
                    expect(res.status).to.equal(200);
                    expect(res.body).to.be.an('array');
                    expect(res.body).to.have.lengthOf(1);
                    expect(res.body[0].status).to.equal('inactive');
                    done();
                });
    });

    // get access token for default service admin
    let defaultServiceAdminToken = '';
    it('Should return access token for default service admin', function (done) {
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

    // unenroll patient from service
    it('Should successfully be unenrolled from service 2 as default service admin', function (done) {
        const bodyRequest = {
            "user": newUser.id,
            "service": Config.service1.id,
            "status": "inactive"
        }   
        chai.request(Config.server).post('omr/api/v1/services/' + Config.service1.id + '/user/unenroll')
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

    it('Should return an enrollment entry for newUser with inactive status as default service admin', function (done) {
        chai.request(Config.server).get('omr/api/v1/users/' + newUser.id + '/enrollments')
                .set('Accept', 'application/json')
                .set('token', defaultServiceAdminToken)
                .end(function(err, res){ 
                    expect(err).to.be.null;
                    expect(res.status).to.equal(200);
                    expect(res.body).to.be.an('array');
                    expect(res.body).to.have.lengthOf(1);
                    expect(res.body[0].status).to.equal('inactive');
                    done();
                });
    });
});
