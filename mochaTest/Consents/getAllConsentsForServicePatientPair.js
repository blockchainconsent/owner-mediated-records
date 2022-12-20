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

describe('Check of getting all consents for a service-datatype-patient and validate consent for patient data access', function() {

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

    // register org1Service2
    let org1Service2 = {};    
    before((done) => {
        org1Service2 = {
            "id": idGenerator(),
            "name": "o1s2",
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
                    "datatype_id": Config.datatype3.id,
                    "access":[
                        "write",
                        "read"
                    ]
                }
            ],
            "solution_private_data": {
                "sample data": "service sample data 3"
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

    // register org2Service1
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
                "sample data 1": "service sample data 2",
                "sample data 2": "service sample data 4"
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

    // register org1Service3
    let org1Service3 = {};    
    before((done) => {
        org1Service3 = {
            "id": idGenerator(),
            "name": "o1s3",
            "secret": "newservice3pass",
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
                    "datatype_id": Config.datatype4.id,
                    "access":[
                        "write",
                        "read"
                    ]
                }
            ],
            "solution_private_data": {
                "sample data 1": "service sample data 1",
                "sample data 4": "service sample data 4"
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


    // create new random newUser1
    const newUser1 = {
        "id": idGenerator(),
        "secret": "newuser1",
        "name": "New User: " + this.id,
        "role": "user",
        "org" : "",
        "email":  "newuser1email@example.com",
        "ca_org": Config.caOrg,
        "data": {
          "address": "2 User St"
        }
    };

    // create new random newUser2
    const newUser2 = {
        "id": idGenerator(),
        "secret": "newuser2",
        "name": "New User: " + this.id,
        "role": "user",
        "org" : "",
        "email":  "newuser2email@example.com",
        "ca_org": Config.caOrg,
        "data": {
          "address": "2 User St"
        }
    };

    // create new org1Admin
    const org1Admin = {
        "id": idGenerator(),
        "secret": "newuser3",
        "name": "New User: " + this.id,
        "role": "user",
        "org" : Config.org1.id,
        "email":  "newuser3email@example.com",
        "ca_org": Config.caOrg,
        "data": {
          "address": "2 User St"
        }
    };

    // create new service1Org1Admin
    const service1Org1Admin = {
        "id": idGenerator(),
        "secret": "newuser4",
        "name": "New User: " + this.id,
        "role": "user",
        "org" : Config.org1.id,
        "email":  "newuser4email@example.com",
        "ca_org": Config.caOrg,
        "data": {
          "address": "2 User St"
        }
    };
    
    // register new newUser1 (doesn't belong to any Org, with role user)
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

    // register new newUser2 (doesn't belong to any Org, with role user)
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

    // register new org1Admin (belongs to Org1, with role user)
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

    // register new service1Org1Admin (belongs to Org1, with role user)
    before((done) => {
        chai.request(Config.server).post('omr/api/v1/users') 
            .set('Accept',  'application/json')
            .set('token', Config.orgAdminToken1)
            .send(service1Org1Admin)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("success");
                done();
            });
    });

    // grant service admin permission to service1Org1Admin
    before((done) => {
        chai.request(Config.server).put('omr/api/v1/users/' + service1Org1Admin.id + '/permissions/services/' + org1Service1.id) 
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

    // enroll newUser2 to org2Service1
    before((done) => {
        const bodyRequest = {
            "user": newUser2.id,
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

    // Get login token for newUser1
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

    // Get login token for newUser2
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

    // Get login token for org1Admin
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

    // Get login token for service1Org1Admin
    let service1Org1AdminToken = '';
    before((done) => {
        chai.request(Config.server).get('common/api/v1/login')
            .set('Accept', 'application/json')
            .set('user-id', service1Org1Admin.id)
            .set('password', service1Org1Admin.secret)
            .set('login-org', service1Org1Admin.ca_org)
            .set('login-channel', Config.channel)
            .set('signature', "")
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("token");
                expect(res.body.token).to.not.equal("");
                service1Org1AdminToken = res.body.token;
                done();
            });
    });

    // Get login token for default service admin of org2Service1
    let defaultOrg2Service1Token = '';
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
                defaultOrg2Service1Token = res.body.token;
                done();
            });
    });

    // Should successfully create consent to org1Service1 for datatype1
    before((done) => {
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

    // Should successfully create consent to org1Service1 for datatype2
    before((done) => {
        const bodyRequest = {
            "patient_id": newUser1.id,
            "service_id": org1Service1.id,
            "target_id": org1Service1.id,
            "datatype_id": org1Service1.datatypes[1].datatype_id,
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

    // Should successfully create consent to org1Service2 for datatype3
    before((done) => {
        const bodyRequest = {
            "patient_id": newUser1.id,
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
 
    // Should successfully create consent to org2Service1 for datatype
    before((done) => {
        const bodyRequest = {
            "patient_id": newUser2.id,
            "service_id": org2Service1.id,
            "target_id": org2Service1.id,
            "datatype_id": org2Service1.datatypes[0].datatype_id,
            "option": [
                "write",
                "read"
            ],
            "expiration": 0
        }
        chai.request(Config.server).post('omr/api/v1/consents/patientData')
            .set('Accept', 'application/json')
            .set('token', newUser2Token)
            .send(bodyRequest)
            .end(function(err, res){
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("success");
                done();
            });
    });

    // Should successfully create consent to org1Service1 for datatype2
    before((done) => {
        const bodyRequest = {
            "patient_id": newUser2.id,
            "service_id": org2Service1.id,
            "target_id": org1Service1.id,
            "datatype_id": org2Service1.datatypes[0].datatype_id,
            "option": [
                "read"
            ],
            "expiration": 0
        }
        chai.request(Config.server).post('omr/api/v1/consents/patientData')
            .set('Accept', 'application/json')
            .set('token', newUser2Token)
            .send(bodyRequest)
            .end(function(err, res){
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("success");
                done();
            });
    });

    // get all consents for org1Service1/newUser1 pair as newUser1
    it('Should successfully get all consents from newUser1 for org1Service1', function (done) {
        let servicePatientPair1 = {
            "owner":newUser1.id,
            "service":org1Service1.id,
            "datatype":org1Service1.datatypes[0].datatype_id,
            "target":org1Service1.id,
            "option":["write","read"],
            "expiration":0,
        }
    
        let servicePatientPair2 = {
            "owner":newUser1.id,
            "service":org1Service1.id,
            "datatype":org1Service1.datatypes[1].datatype_id,
            "target":org1Service1.id,
            "option":["write","read"],
            "expiration":0,
        }
        chai.request(Config.server).get('omr/api/v1/services/' + org1Service1.id + '/users/' + newUser1.id + '/consents')
            .set('Accept', 'application/json')
            .set('token', newUser1Token)
            .end(function(err, res){
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.be.an('array');
                expect(res.body).to.have.lengthOf(2);
                expect(res.body).to.contain.something.like(servicePatientPair1);
                expect(res.body).to.contain.something.like(servicePatientPair2);
                done();
            });
    });

    // get all consents for org2Service1/newUser2 pair as newUser2
    it('Should successfully get all consents from newUser2 for org2Service1', function (done) {
        let servicePatientPair = {
            "owner":newUser2.id,
            "service":org2Service1.id,
            "datatype":org2Service1.datatypes[0].datatype_id,
            "target":org2Service1.id,
            "option":[
                "write",
                "read"
            ],
            "expiration":0,
        }
        chai.request(Config.server).get('omr/api/v1/services/' + org2Service1.id + '/users/' + newUser2.id + '/consents')
            .set('Accept', 'application/json')
            .set('token', newUser2Token)
            .end(function(err, res){
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.be.an('array');
                expect(res.body).to.have.lengthOf(1);
                expect(res.body).to.contain.something.like(servicePatientPair);
                done();
            });
    });

    // get all consents for org1Service1/newUser1 pair as org1Admin
    it('Should successfully get all consents from newUser1 for org1Service1 as org1Admin', function (done) {
        let servicePatientPair1 = {
            "owner":newUser1.id,
            "service":org1Service1.id,
            "datatype":org1Service1.datatypes[0].datatype_id,
            "target":org1Service1.id,
            "option":["write","read"],
            "expiration":0,
        }
    
        let servicePatientPair2 = {
            "owner":newUser1.id,
            "service":org1Service1.id,
            "datatype":org1Service1.datatypes[1].datatype_id,
            "target":org1Service1.id,
            "option":["write","read"],
            "expiration":0,
        }
        chai.request(Config.server).get('omr/api/v1/services/' + org1Service1.id + '/users/' + newUser1.id + '/consents')
            .set('Accept', 'application/json')
            .set('token', org1AdminToken)
            .end(function(err, res){
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.be.an('array');
                expect(res.body).to.have.lengthOf(2);
                expect(res.body).to.contain.something.like(servicePatientPair1);
                expect(res.body).to.contain.something.like(servicePatientPair2);
                done();
            });
    });

    // get all consents for org1Service1/newUser1 pair as service1Org1Admin
    it('Should successfully get all consents from newUser1 for org1Service1 as service1Org1Admin', function (done) {
        let servicePatientPair1 = {
            "owner":newUser1.id,
            "service":org1Service1.id,
            "datatype":org1Service1.datatypes[0].datatype_id,
            "target":org1Service1.id,
            "option":["write","read"],
            "expiration":0,
        }
    
        let servicePatientPair2 = {
            "owner":newUser1.id,
            "service":org1Service1.id,
            "datatype":org1Service1.datatypes[1].datatype_id,
            "target":org1Service1.id,
            "option":["write","read"],
            "expiration":0,
        }
        chai.request(Config.server).get('omr/api/v1/services/' + org1Service1.id + '/users/' + newUser1.id + '/consents')
            .set('Accept', 'application/json')
            .set('token', service1Org1AdminToken)
            .end(function(err, res){
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.be.an('array');
                expect(res.body).to.have.lengthOf(2);
                expect(res.body).to.contain.something.like(servicePatientPair1);
                expect(res.body).to.contain.something.like(servicePatientPair2);
                done();
            });
    });

    // get all consents for org2Service1/newUser2 pair as default org admin of org2
    it('Should successfully get all consents from newUser2 for org2Service1 as default org admin of org2', function (done) {
        let servicePatientPair = {
            "owner":newUser2.id,
            "service":org2Service1.id,
            "datatype":org2Service1.datatypes[0].datatype_id,
            "target":org2Service1.id,
            "option":[
                "write",
                "read"
            ],
            "expiration":0,
        }
        chai.request(Config.server).get('omr/api/v1/services/' + org2Service1.id + '/users/' + newUser2.id + '/consents')
            .set('Accept', 'application/json')
            .set('token', Config.orgAdminToken2)
            .end(function(err, res){
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.be.an('array');
                expect(res.body).to.have.lengthOf(1);
                expect(res.body).to.contain.something.like(servicePatientPair);
                done();
            });
    });

    // get all consents for org2Service1/newUser2 pair as default service admin of org2Service1
    it('Should successfully get all consents from newUser2 for org2Service1 as default service admin of org2Service1', function (done) {
        let servicePatientPair = {
            "owner":newUser2.id,
            "service":org2Service1.id,
            "datatype":org2Service1.datatypes[0].datatype_id,
            "target":org2Service1.id,
            "option":[
                "write",
                "read"
            ],
            "expiration":0,
        }
        chai.request(Config.server).get('omr/api/v1/services/' + org2Service1.id + '/users/' + newUser2.id + '/consents')
            .set('Accept', 'application/json')
            .set('token', defaultOrg2Service1Token)
            .end(function(err, res){
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.be.an('array');
                expect(res.body).to.have.lengthOf(1);
                expect(res.body).to.contain.something.like(servicePatientPair);
                done();
            });
    });

    // Get consent for a Service-Datatype-Patient as serviceAdmin
    it('Should successfully get consent for newUser1/org1Service1/datatype2 as service1Org1Admin', function (done) {
        chai.request(Config.server).get('omr/api/v1/services/' + org1Service1.id + '/users/' + newUser1.id + '/datatype/' + org1Service1.datatypes[1].datatype_id + '/consents')
            .set('Accept', 'application/json')
            .set('token', service1Org1AdminToken)
            .end(function(err, res){
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body.owner).to.equal(newUser1.id);
                expect(res.body.service).to.equal(org1Service1.id);
                expect(res.body.datatype).to.equal(org1Service1.datatypes[1].datatype_id);
                expect(res.body.target).to.equal(org1Service1.id);
                done();
            });
    });

    // Get consent for a Service-Datatype-Patient as user
    it('Should successfully get consent for newUser1/org1Service1/datatype2 as newUser2', function (done) {
        chai.request(Config.server).get('omr/api/v1/services/' + org2Service1.id + '/users/' + newUser2.id + '/datatype/' + org2Service1.datatypes[0].datatype_id + '/consents')
            .set('Accept', 'application/json')
            .set('token', newUser2Token)
            .end(function(err, res){
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body.owner).to.equal(newUser2.id);
                expect(res.body.service).to.equal(org2Service1.id);
                expect(res.body.datatype).to.equal(org2Service1.datatypes[0].datatype_id);
                expect(res.body.target).to.equal(org2Service1.id);
                done();
            });
    });

    // Validate сonsent for Patient Data Access as as org admin of this org
    it('Should successfully return access token which can be used to download data as org admin of this org', function (done) {
        chai.request(Config.server).get('omr/api/v1/services/' + org2Service1.id + '/users/' + newUser2.id + '/datatype/' + org2Service1.datatypes[0].datatype_id + '/validation/read')
            .set('Accept', 'application/json')
            .set('token', Config.orgAdminToken2)
            .end(function(err, res){
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body.owner).to.equal(newUser2.id);
                expect(res.body.datatype).to.equal(org2Service1.datatypes[0].datatype_id);
                expect(res.body.target).to.equal(org2Service1.id);
                expect(res.body.requested_access).to.equal("read");
                expect(res.body.message).to.equal("permission granted");
                expect(res.body).to.have.property("token");
                expect(res.body.token).to.not.equal("");
                done();
            });
    });

    // Validate сonsent for Patient Data Access as service admin of this service
    it('Should successfully return access token which can be used to download data as service admin of this service', function (done) {
        chai.request(Config.server).get('omr/api/v1/services/' + org2Service1.id + '/users/' + newUser2.id + '/datatype/' + org2Service1.datatypes[0].datatype_id + '/validation/read')
            .set('Accept', 'application/json')
            .set('token', defaultOrg2Service1Token)
            .end(function(err, res){
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body.owner).to.equal(newUser2.id);
                expect(res.body.datatype).to.equal(org2Service1.datatypes[0].datatype_id);
                expect(res.body.target).to.equal(org2Service1.id);
                expect(res.body.requested_access).to.equal("read");
                expect(res.body.message).to.equal("permission granted");
                expect(res.body).to.have.property("token");
                expect(res.body.token).to.not.equal("");
                done();
            });
    });

    // Validate сonsent for Patient Data Access as a service admin of another org
    it('Should not return access token when service admin of another service tries to call this API', function (done) {
        chai.request(Config.server).get('omr/api/v1/services/' + org2Service1.id + '/users/' + newUser2.id + '/datatype/' + org2Service1.datatypes[0].datatype_id + '/validation/read')
            .set('Accept', 'application/json')
            .set('token', service1Org1AdminToken)
            .end(function(err, res){
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body.owner).to.equal(newUser2.id);
                expect(res.body.datatype).to.equal(org2Service1.datatypes[0].datatype_id);
                expect(res.body.target).to.equal(org2Service1.id);
                expect(res.body.requested_access).to.equal("read");
                expect(res.body.permission_granted).to.be.false;
                expect(res.body.message).to.equal("permission denied");
                expect(res.body).to.have.property("token");
                expect(res.body.token).to.equal("");
                done();
            });
    });

    // Validate сonsent for Patient Data Access as an org admin of another org
    it('Should not return access token when org admin of another org tries to call this API', function (done) {
        chai.request(Config.server).get('omr/api/v1/services/' + org2Service1.id + '/users/' + newUser2.id + '/datatype/' + org2Service1.datatypes[0].datatype_id + '/validation/read')
            .set('Accept', 'application/json')
            .set('token', org1AdminToken)
            .end(function(err, res){
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body.owner).to.equal(newUser2.id);
                expect(res.body.datatype).to.equal(org2Service1.datatypes[0].datatype_id);
                expect(res.body.target).to.equal(org2Service1.id);
                expect(res.body.requested_access).to.equal("read");
                expect(res.body.permission_granted).to.be.false;
                expect(res.body.message).to.equal("permission denied");
                expect(res.body).to.have.property("token");
                expect(res.body.token).to.equal("");
                done();
            });
    });

    // get list of consents for newUser1 as service1Org1AdminToken
    it('Should get list of consents for newUser1 as service1Org1AdminToken', function (done) {
        let consent1 = {
            "owner":newUser1.id,
            "service":org1Service1.id,
            "datatype":org1Service1.datatypes[0].datatype_id,
            "target":org1Service1.id,
            "option":["write","read"],
            "expiration":0,
        }
    
        let consent2 = {
            "owner":newUser1.id,
            "service":org1Service1.id,
            "datatype":org1Service1.datatypes[1].datatype_id,
            "target":org1Service1.id,
            "option":["write","read"],
            "expiration":0,
        }
        chai.request(Config.server).get('omr/api/v1/users/' + newUser1.id + '/consents')
            .set('Accept', 'application/json')
            .set('token', service1Org1AdminToken)
            .end(function(err, res){
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.be.an('array');
                expect(res.body).to.have.lengthOf(2);
                expect(res.body).to.contain.something.like(consent1);
                expect(res.body).to.contain.something.like(consent2);
                done();
            });
    });

    // get list of consents for newUser1 as org1AdminToken
    it('Should get list of consents for newUser1 as org1AdminToken', function (done) {
        let consent1 = {
            "owner":newUser1.id,
            "service":org1Service1.id,
            "datatype":org1Service1.datatypes[0].datatype_id,
            "target":org1Service1.id,
            "option":["write","read"],
            "expiration":0,
        }
    
        let consent2 = {
            "owner":newUser1.id,
            "service":org1Service1.id,
            "datatype":org1Service1.datatypes[1].datatype_id,
            "target":org1Service1.id,
            "option":["write","read"],
            "expiration":0,
        }
        let consent3 = {
            "owner":newUser1.id,
            "service":org1Service2.id,
            "datatype":org1Service2.datatypes[0].datatype_id,
            "target":org1Service2.id,
            "option":["write","read"],
            "expiration":0,
        }
        chai.request(Config.server).get('omr/api/v1/users/' + newUser1.id + '/consents')
            .set('Accept', 'application/json')
            .set('token', org1AdminToken)
            .end(function(err, res){
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.be.an('array');
                expect(res.body).to.have.lengthOf(3);
                expect(res.body).to.contain.something.like(consent1);
                expect(res.body).to.contain.something.like(consent2);
                expect(res.body).to.contain.something.like(consent3);
                done();
            });
    });

    // revoke org admin permission from org user org1Admin
    it('Should revoke org admin permission from org user org1Admin', function (done) {  
        chai.request(Config.server).delete('omr/api/v1/users/' + org1Admin.id + '/permissions/admin/' + Config.org1.id) 
            .set('Accept',  'application/json')
            .set('token', Config.orgAdminToken1)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body.msg).to.include("success");
                done();
            });
        
    });

    // revoke service admin permission from org user service1Org1Admin
    it('Should revoke service admin permission from org user service1Org1Admin', function (done) {
        chai.request(Config.server).delete('omr/api/v1/users/' + service1Org1Admin.id + '/permissions/services/' + org1Service1.id) 
            .set('Accept',  'application/json')
            .set('token', Config.orgAdminToken1)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body.msg).to.include("success");
                done();
            });
    });

    // get blank list of consents for org1Service1/newUser1 pair as org1Admin without rights
    it('Should not get all consents from newUser1 for org1Service1 as org1Admin anymore', function (done) {
        chai.request(Config.server).get('omr/api/v1/services/' + org1Service1.id + '/users/' + newUser1.id + '/consents')
            .set('Accept', 'application/json')
            .set('token', org1AdminToken)
            .end(function(err, res){
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.be.an('array');
                expect(res.body).to.have.lengthOf(0);
                done();
            });
    });

    // get blank list of consents for org1Service1/newUser1 pair as service1Org1Admin without rights
    it('Should not get all consents from newUser1 for org1Service1 as service1Org1Admin anymore', function (done) {
        chai.request(Config.server).get('omr/api/v1/services/' + org1Service1.id + '/users/' + newUser1.id + '/consents')
            .set('Accept', 'application/json')
            .set('token', service1Org1AdminToken)
            .end(function(err, res){
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.be.an('array');
                expect(res.body).to.have.lengthOf(0);
                done();
            });
    });

    // get list of consents for newUser1 as newUser1
    it('Should get list of consents for newUser1 as newUser1', function (done) {
        let consent1 = {
            "owner":newUser1.id,
            "service":org1Service1.id,
            "datatype":org1Service1.datatypes[0].datatype_id,
            "target":org1Service1.id,
            "option":["write","read"],
            "expiration":0,
        }
    
        let consent2 = {
            "owner":newUser1.id,
            "service":org1Service1.id,
            "datatype":org1Service1.datatypes[1].datatype_id,
            "target":org1Service1.id,
            "option":["write","read"],
            "expiration":0,
        }
        let consent3 = {
            "owner":newUser1.id,
            "service":org1Service2.id,
            "datatype":org1Service2.datatypes[0].datatype_id,
            "target":org1Service2.id,
            "option":["write","read"],
            "expiration":0,
        }
        chai.request(Config.server).get('omr/api/v1/users/' + newUser1.id + '/consents')
            .set('Accept', 'application/json')
            .set('token', newUser1Token)
            .end(function(err, res){
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.be.an('array');
                expect(res.body).to.have.lengthOf(3);
                expect(res.body).to.contain.something.like(consent1);
                expect(res.body).to.contain.something.like(consent2);
                expect(res.body).to.contain.something.like(consent3);
                done();
            });
    });

    it('Should successfully revoke consent from org1Service1 datatype2', function (done) {
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

    // get list of consents for newUser1 as newUser1 one more time after one of consents was removed
    it('Should get list of consents for newUser1 as newUser1 after one of consents was removed', function (done) {
        let consent1 = {
            "owner":newUser1.id,
            "service":org1Service1.id,
            "datatype":org1Service1.datatypes[0].datatype_id,
            "target":org1Service1.id,
            "option":["write","read"],
            "expiration":0,
        }
        let consent2 = {
            "owner":newUser1.id,
            "service":org1Service1.id,
            "datatype":org1Service1.datatypes[1].datatype_id,
            "target":org1Service1.id,
            "option":["deny"],
            "expiration":0,
        }
        let consent3 = {
            "owner":newUser1.id,
            "service":org1Service2.id,
            "datatype":org1Service2.datatypes[0].datatype_id,
            "target":org1Service2.id,
            "option":["write","read"],
            "expiration":0,
        }
        chai.request(Config.server).get('omr/api/v1/users/' + newUser1.id + '/consents')
            .set('Accept', 'application/json')
            .set('token', newUser1Token)
            .end(function(err, res){
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.be.an('array');
                expect(res.body).to.have.lengthOf(3);
                expect(res.body).to.contain.something.like(consent1);
                expect(res.body).to.contain.something.like(consent2);
                expect(res.body).to.contain.something.like(consent3);
                done();
            });
    });

    // get list of requests for newUser1 as newUser1
    it('Should get list of requests for newUser1 as newUser1', function (done) {
        let request3 = {
            "user":newUser1.id,
            "org":Config.org1.id,
            "service":org1Service1.id,
            "service_name":org1Service1.name,
            "status": "active"
        }
    
        let request2 = {
            "user":newUser1.id,
            "org":Config.org1.id,
            "service":org1Service2.id,
            "service_name":org1Service2.name,
            "status": "active"
        }

        let request1 = {
            "user":newUser1.id,
            "org":Config.org1.id,
            "service":org1Service3.id,
            "service_name":org1Service3.name,
            "status":"active"
        }

        chai.request(Config.server).get('omr/api/v1/users/' + newUser1.id + '/consents/requests')
            .set('Accept', 'application/json')
            .set('token', newUser1Token)
            .end(function(err, res){
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.be.an('array');
                expect(res.body).to.have.lengthOf(3);
                expect(res.body).to.contain.something.like(request1);
                expect(res.body).to.contain.something.like(request2);
                expect(res.body).to.contain.something.like(request3);
                done();
            });
    });

    // get list of requests for newUser1 as org admin
    it('Should get list of requests for newUser1 as org admin', function (done) {
        let request3 = {
            "user":newUser1.id,
            "org":Config.org1.id,
            "service":org1Service1.id,
            "service_name":org1Service1.name,
            "status": "active"
        }
    
        let request2 = {
            "user":newUser1.id,
            "org":Config.org1.id,
            "service":org1Service2.id,
            "service_name":org1Service2.name,
            "status": "active"
        }

        let request1 = {
            "user":newUser1.id,
            "org":Config.org1.id,
            "service":org1Service3.id,
            "service_name":org1Service3.name,
            "status":"active"
        }

        chai.request(Config.server).get('omr/api/v1/users/' + newUser1.id + '/consents/requests')
            .set('Accept', 'application/json')
            .set('token', Config.orgAdminToken1)
            .end(function(err, res){
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.be.an('array');
                expect(res.body).to.have.lengthOf(3);
                expect(res.body).to.contain.something.like(request1);
                expect(res.body).to.contain.something.like(request2);
                expect(res.body).to.contain.something.like(request3);
                done();
            });
    });

    // enroll newUser1 to org2Service1
    it('Should enroll newUser1 to org2Service1', function (done) {
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

    // Should successfully create consent to org2Service1 for datatype
    it('Should create consent to org2Service1 for datatype4', function (done) {
        const bodyRequest = {
            "patient_id": newUser1.id,
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

    // get list of requests for newUser1 as newUser1
    it('Should get list of requests for newUser1 as newUser1', function (done) {

        let request4 = {
            "user":newUser1.id,
            "org":Config.org2.id,
            "service":org2Service1.id,
            "service_name":org2Service1.name,
            "status": "active"
        }

        let request3 = {
            "user":newUser1.id,
            "org":Config.org1.id,
            "service":org1Service1.id,
            "service_name":org1Service1.name,
            "status": "active"
        }
    
        let request2 = {
            "user":newUser1.id,
            "org":Config.org1.id,
            "service":org1Service2.id,
            "service_name":org1Service2.name,
            "status": "active"
        }

        let request1 = {
            "user":newUser1.id,
            "org":Config.org1.id,
            "service":org1Service3.id,
            "service_name":org1Service3.name,
            "status":"active"
        }

        chai.request(Config.server).get('omr/api/v1/users/' + newUser1.id + '/consents/requests')
            .set('Accept', 'application/json')
            .set('token', newUser1Token)
            .end(function(err, res){
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.be.an('array');
                expect(res.body).to.have.lengthOf(4);
                expect(res.body).to.contain.something.like(request1);
                expect(res.body).to.contain.something.like(request2);
                expect(res.body).to.contain.something.like(request3);
                expect(res.body).to.contain.something.like(request4);
                done();
            });
    });

    // get list of requests for newUser1 as org admin of org2
    it('Should get list of requests for newUser1 as org admin of org2', function (done) {

        let request1 = {
            "user":newUser1.id,
            "org":Config.org2.id,
            "service":org2Service1.id,
            "service_name":org2Service1.name,
            "status":"active"
        }

        chai.request(Config.server).get('omr/api/v1/users/' + newUser1.id + '/consents/requests')
            .set('Accept', 'application/json')
            .set('token', Config.orgAdminToken2)
            .end(function(err, res){
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.be.an('array');
                expect(res.body).to.have.lengthOf(1);
                expect(res.body).to.contain.something.like(request1);
                done();
            });
    });
});
