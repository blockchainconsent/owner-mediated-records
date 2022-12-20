let Config = require('../ConfigFile.js');
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
require("../Users/RegisterUsers");
require("../Login/OrgLogin/Org1Login");

describe('Get all enrollments for patient', function() {

    // registration of first new service
    let newService = {};    
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
    let newService2 = {};    
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

    before((done) => {
        Config.newServiceOneID = newService.id;
        const bodyRequest = {
            "user": newUser.id,
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
            "user": newUser.id,
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
            "user": newUser2.id,
            "service": newService2.id,
            "status": "active"
        };
        
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

    it('Should return a list of enrollments for user1', function (done) {
        
        chai.request(Config.server).get('omr/api/v1/users/' + newUser.id + '/enrollments')
                .set('Accept', 'application/json')
                .set('token', Config.orgAdminToken1)
                .end(function(err, res){
                    expect(err).to.be.null;
                    expect(res.status).to.equal(200);
                    expect(res.body).to.be.an('array');
                    expect(res.body).to.have.lengthOf(2);
                    expect(res.body[0].service_id).to.be.oneOf([newService.id, newService2.id]);
                    expect(res.body[1].service_id).to.be.oneOf([newService.id, newService2.id]);
                    done();
                });
    });

    it('Should return a list of enrollments for user2', function (done) {
        
        chai.request(Config.server).get('omr/api/v1/users/' + newUser2.id + '/enrollments')
                .set('Accept', 'application/json')
                .set('token', Config.orgAdminToken1)
                .end(function(err, res){
                    expect(err).to.be.null;
                    expect(res.status).to.equal(200);
                    expect(res.body).to.be.an('array');
                    expect(res.body).to.have.lengthOf(1);
                    expect(res.body[0].service_id).to.equal(newService2.id);
                    done();
                });
    });
});