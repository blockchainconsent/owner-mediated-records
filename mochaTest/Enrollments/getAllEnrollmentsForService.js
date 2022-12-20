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

describe('Get all Enrollments for Service as org admin', function() {

    // registration of first new service
    var newService = {};    
    before((done) => {
        newService = {
            "id": idGenerator(),
            "name": "newService",
            "secret": "newservicepass",
            "ca_org": Config.caOrg,
            "email": "newserviceemail@example.com",
            "org_id": Config.org1.id,
            "summary": "New service under org 1. Has one datatype",
            "terms": {
                "term1" : "example term",
                "term2" : "example term",
            },
            "payment_required": "yes",
            "datatypes": [
                {
                    "datatype_id": Config.datatype1.id,
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

        chai.request(Config.server).post('omr/api/v1/services') 
            .set('Accept',  'application/json')
            .set('token', Config.orgAdminToken1)
            .send(newService)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("success");
                done();
                });
    });

    // registration of second new service
    var newService2 = {};    
    before((done) => {
        newService2 = {
            "id": idGenerator(),
            "name": "newService2",
            "secret": "newservicepass",
            "ca_org": Config.caOrg,
            "email": "newservice2email@example.com",
            "org_id": Config.org1.id,
            "summary": "New service under org 1. Has one datatype",
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
                }
            ],
            "solution_private_data": {
                "sample data 1": "service 3 sample data 1",
                "sample data 2": "service 3 sample data 2"
            }
        };

        chai.request(Config.server).post('omr/api/v1/services') 
            .set('Accept',  'application/json')
            .set('token', Config.orgAdminToken1)
            .send(newService2)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("success");
                done();
                });
    });

    before((done) => {
        Config.newServiceOneID = newService.id;
        const bodyRequest = {
            "user": Config.user1.id,
            "service": newService.id,
            "status": "active"
        }
        chai.request(Config.server).post('omr/api/v1/services/' + newService.id + '/user/enroll')
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

    before((done) => {
        Config.newServiceTwoID = newService2.id;
        const bodyRequest = {
            "user": Config.user1.id,
            "service": newService2.id,
            "status": "active"
        }
        chai.request(Config.server).post('omr/api/v1/services/' + newService2.id + '/user/enroll')
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

    before((done) => {
        const bodyRequest = {
            "user": Config.user2.id,
            "service": newService2.id,
            "status": "active"
        }
        chai.request(Config.server).post('omr/api/v1/services/' + newService2.id + '/user/enroll')
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

    it('Should return a list of users enrolled to newservice', function (done) {
        
        chai.request(Config.server).get('omr/api/v1/services/' + newService.id + '/enrollments')
                .set('Accept', 'application/json')
                .set('token', Config.orgAdminToken1)
                .end(function(err, res){
                    expect(err).to.be.null;
                    expect(res.status).to.equal(200);
                    expect(res.body).to.be.an('array');
                    expect(res.body).to.have.lengthOf(1);
                    expect(res.body[0].service_id).to.equal(newService.id);
                    done();
                });
    });

    it('Should return a list of users enrolled to newservice2', function (done) {
        
        chai.request(Config.server).get('omr/api/v1/services/' + newService2.id + '/enrollments')
                .set('Accept', 'application/json')
                .set('token', Config.orgAdminToken1)
                .end(function(err, res){
                    expect(err).to.be.null;
                    expect(res.status).to.equal(200);
                    expect(res.body).to.be.an('array');
                    expect(res.body).to.have.lengthOf(2);
                    expect(res.body[0].service_id).to.equal(newService2.id);
                    expect(res.body[1].service_id).to.equal(newService2.id);
                    done();
                });
    });

    // login as default service admin
    let defaultService2AdminToken = '';
    it('Should return a 200 test response', function (done) {
        chai.request(Config.server).get('common/api/v1/login')
            .set('Accept', 'application/json')
            .set('user-id', newService2.id)
            .set('password', newService2.secret)
            .set('login-org', newService2.ca_org)
            .set('login-channel', Config.channel)
            .set('signature', "")
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.not.deep.equal({});
                expect(res.body.token).to.not.equal("");
                defaultService2AdminToken = res.body.token;
                done();
            });
    });

    it('Should return a list of enrollments to newService2 as default service admin', function (done) {
        chai.request(Config.server).get('omr/api/v1/services/' + newService2.id + '/enrollments')
                .set('Accept', 'application/json')
                .set('token', defaultService2AdminToken)
                .end(function(err, res){
                    expect(err).to.be.null;
                    expect(res.status).to.equal(200);
                    expect(res.body).to.be.an('array');
                    expect(res.body).to.have.lengthOf(2);
                    expect(res.body[0].service_id).to.equal(newService2.id);
                    expect(res.body[1].service_id).to.equal(newService2.id);
                    expect(res.body[0].user_id).to.equal(Config.user1.id);
                    expect(res.body[1].user_id).to.equal(Config.user2.id);
                    done();
                });
    });
});

require("../Login/User2Login");

describe('Get all enrollments for service as user 2', function() {

    it('Should return a list of enrollments to new service', function (done) {
        
        chai.request(Config.server).get('omr/api/v1/services/' + Config.newServiceOneID + '/enrollments')
                .set('Accept', 'application/json')
                .set('token', Config.user2Token)
                .end(function(err, res){
                    expect(err).to.be.null;
                    expect(res.status).to.equal(200);
                    expect(res.body).to.be.an('array');
                    expect(res.body).to.have.lengthOf(0);
                    done();
                });
    });

    it('Should return a list of enrollments to new service 2', function (done) {
        
        chai.request(Config.server).get('omr/api/v1/services/' + Config.newServiceTwoID + '/enrollments')
                .set('Accept', 'application/json')
                .set('token', Config.user2Token)
                .end(function(err, res){
                    expect(err).to.be.null;
                    expect(res.status).to.equal(200);
                    expect(res.body).to.be.an('array');
                    expect(res.body).to.have.lengthOf(1);
                    expect(res.body[0].service_id).to.equal(Config.newServiceTwoID);
                    done();
                });
    });
});

require("../Login/User1Login");

describe('Get all enrollments for service as user 1', function() {

    it('Should return a list of enrollments to new service 1', function (done) {
        
        chai.request(Config.server).get('omr/api/v1/services/' + Config.newServiceOneID + '/enrollments')
                .set('Accept', 'application/json')
                .set('token', Config.user1Token)
                .end(function(err, res){
                    expect(err).to.be.null;
                    expect(res.status).to.equal(200);
                    expect(res.body).to.be.an('array');
                    expect(res.body).to.have.lengthOf(1);
                    expect(res.body[0].service_id).to.equal(Config.newServiceOneID);
                    done();
                });
    });

    it('Should return a list of enrollments to new service 2', function (done) {
        
        chai.request(Config.server).get('omr/api/v1/services/' + Config.newServiceTwoID + '/enrollments')
                .set('Accept', 'application/json')
                .set('token', Config.user1Token)
                .end(function(err, res){
                    expect(err).to.be.null;
                    expect(res.status).to.equal(200);
                    expect(res.body).to.be.an('array');
                    expect(res.body).to.have.lengthOf(1);
                    expect(res.body[0].service_id).to.equal(Config.newServiceTwoID);
                    done();
                });
    });
});

require("../Organizations/OrgSetup/RegisterOrg2");
require("../Login/OrgLogin/Org2Login");
describe('Get all enrollments for service as org admin of another organization', function() {

    it('Should return 200 code and blank array of users been enrolled to service 1', function (done) {
        
        chai.request(Config.server).get('omr/api/v1/services/' + Config.service1.id + '/enrollments')
                .set('Accept', 'application/json')
                .set('token', Config.orgAdminToken2)
                .end(function(err, res){
                    expect(err).to.be.null;
                    expect(res.status).to.equal(200);
                    expect(res.body).to.be.an('array');
                    expect(res.body).to.have.lengthOf(0);
                    done();
                });
    });

    it('Should return 200 code and blank array of users been enrolled to service 2', function (done) {
        
        chai.request(Config.server).get('omr/api/v1/services/' + Config.service2.id + '/enrollments')
                .set('Accept', 'application/json')
                .set('token', Config.orgAdminToken2)
                .end(function(err, res){
                    expect(err).to.be.null;
                    expect(res.status).to.equal(200);
                    expect(res.body).to.be.an('array');
                    expect(res.body).to.have.lengthOf(0);
                    done();
                });
    });
});

describe('Grant permissions of service admin to new user and get the list of enrollments for service', function () {

    // create new random user
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
            .send(newUser2)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("success");
                done();
            });
    });

    // and grant service admin permission to him
    before((done) => {
        chai.request(Config.server).put('omr/api/v1/users/' + newUser2.id + '/permissions/services/' + Config.newServiceOneID) 
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

    it('Should return a list of enrollments to new service 1', function (done) {
        chai.request(Config.server).get('omr/api/v1/services/' + Config.newServiceOneID + '/enrollments')
                .set('Accept', 'application/json')
                .set('token', serviceAdminToken)
                .end(function(err, res){
                    expect(err).to.be.null;
                    expect(res.status).to.equal(200);
                    expect(res.body).to.be.an('array');
                    expect(res.body).to.have.lengthOf(1);
                    expect(res.body[0].service_id).to.equal(Config.newServiceOneID);
                    done();
                });
    });
});
