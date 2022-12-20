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

describe('Checking operation logs by auditor', function() {

    let timestampStart;
    let timestampStop;

    let timeStampTwoStart;
    let timeStampTwoStop;

    let orgAdminToken1;
    let orgAdminToken2;

    const org1 = {
        "id": idGenerator(),
        "secret": "pass1",
        "name": "org1",
        "ca_org": Config.caOrg,
        "email": "org1@email.com",
        "status": "active"
    };
        
    const org2 = {
      "id": idGenerator(),
      "secret": "pass2",
      "name": "org2",
      "ca_org": Config.caOrg,
      "email": "org2@email.com",
      "status": "active"
    };

    // register org1
    before((done) => {  
        chai.request(Config.server).post('omr/api/v1/orgs') 
            .set('Accept',  'application/json')
            .set('token', Config.sysAdminToken)
            .send(org1)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("success");
                done();
            });
    });

    // register org2
    before((done) => {  
        chai.request(Config.server).post('omr/api/v1/orgs') 
            .set('Accept',  'application/json')
            .set('token', Config.sysAdminToken)
            .send(org2)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("success");
                done();
            });
    });

    // org1admin token
    before((done) => {
        chai.request(Config.server).get('common/api/v1/login')
            .set('Accept', 'application/json')
            .set('user-id', org1.id)
            .set('password', org1.secret)
            .set('login-org', org1.ca_org)
            .set('login-channel', Config.channel)
            .set('signature', "")
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.not.deep.equal({})
                expect(res.body.token).to.not.equal("")
                orgAdminToken1 = res.body.token;
                done();
            });
    });

    // org2admin token
    before((done) => {
        chai.request(Config.server).get('common/api/v1/login')
            .set('Accept', 'application/json')
            .set('user-id', org2.id)
            .set('password', org2.secret)
            .set('login-org', org2.ca_org)
            .set('login-channel', Config.channel)
            .set('signature', "")
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.not.deep.equal({})
                expect(res.body.token).to.not.equal("")
                orgAdminToken2 = res.body.token;
                done();
            });
    });

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
            "org_id": org1.id,
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
            .set('token', orgAdminToken1)
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
            "org_id": org2.id,
            "summary": "New service under org 2. Has two datatypes",
            "terms": {
                "term1" : "example term",
                "term2" : "example term"
            },
            "payment_required": "yes",
            "datatypes": [
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
                        "write"
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
            .set('token', orgAdminToken2)
            .send(org2Service1)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("success");
                done();
                });
    });

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

    // create newPatient2
    const newPatient2 = {
        "id": idGenerator(),
        "secret": "newPatient2",
        "name": "newPatient2",
        "role": "user",
        "org" : "",
        "email":  "newPatient2email@example.com",
        "ca_org": Config.caOrg,
        "data": {
          "address": "2 User St"
        }
    };

    // create newAuditorOrg1Service1
    const newAuditorOrg1Service1 = {
        "id": idGenerator(),
        "secret": "newAuditorOrg1Service1",
        "name": "newAuditorOrg1Service1",
        "role": "audit",
        "org" : "",
        "email":  "newAuditorOrg1Service1email@example.com",
        "ca_org": Config.caOrg,
        "data": {
          "address": "2 User St"
        }
    };

    // create newAuditorOrg2Service1
    const newAuditorOrg2Service1 = {
        "id": idGenerator(),
        "secret": "newAuditorOrg2Service1",
        "name": "newAuditorOrg2Service1",
        "role": "audit",
        "org" : "",
        "email":  "newAuditorOrg2Service1email@example.com",
        "ca_org": Config.caOrg,
        "data": {
          "address": "2 User St"
        }
    };

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

    // register newPatient2 (doesn't belong to any Org, with role user)
    before((done) => {
        chai.request(Config.server).post('omr/api/v1/users') 
            .set('Accept',  'application/json')
            .set('token', Config.sysAdminToken)
            .send(newPatient2)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("success");
                done();
            });
    });

    // register newAuditorOrg1Service1
    before((done) => {
        chai.request(Config.server).post('omr/api/v1/users') 
            .set('Accept',  'application/json')
            .set('token', Config.sysAdminToken)
            .send(newAuditorOrg1Service1)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("success");
                done();
            });
    });

    // register newAuditorOrg2Service1
    before((done) => {
        chai.request(Config.server).post('omr/api/v1/users') 
            .set('Accept',  'application/json')
            .set('token', Config.sysAdminToken)
            .send(newAuditorOrg2Service1)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("success");
                done();
            });
    });

    // grant auditor permission to newAuditorOrg1Service1 by org admin
    before((done) => {
        chai.request(Config.server).put('omr/api/v1/users/' + newAuditorOrg1Service1.id + '/permissions/audit/' + org1Service1.id) 
            .set('Accept',  'application/json')
            .set('token', orgAdminToken1)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("success");
                done();
            });
    });

    // grant auditor permission to newAuditorOrg2Service1 by org admin
    before((done) => {
        chai.request(Config.server).put('omr/api/v1/users/' + newAuditorOrg2Service1.id + '/permissions/audit/' + org2Service1.id) 
            .set('Accept',  'application/json')
            .set('token', orgAdminToken2)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("success");
                done();
            });
    });

    let newAuditorOrg1Service1Token = '';
    before((done) => {
        chai.request(Config.server).get('common/api/v1/login')
            .set('Accept', 'application/json')
            .set('user-id', newAuditorOrg1Service1.id)
            .set('password', newAuditorOrg1Service1.secret)
            .set('login-org', newAuditorOrg1Service1.ca_org)
            .set('login-channel', Config.channel)
            .set('signature', "")
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("token");
                expect(res.body.token).to.not.equal("");
                newAuditorOrg1Service1Token = res.body.token;
                done();
            });
    });

    let newAuditorOrg2Service1Token = '';
    before((done) => {
        chai.request(Config.server).get('common/api/v1/login')
            .set('Accept', 'application/json')
            .set('user-id', newAuditorOrg2Service1.id)
            .set('password', newAuditorOrg2Service1.secret)
            .set('login-org', newAuditorOrg2Service1.ca_org)
            .set('login-channel', Config.channel)
            .set('signature', "")
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("token");
                expect(res.body.token).to.not.equal("");
                newAuditorOrg2Service1Token = res.body.token;
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
            .set('token', orgAdminToken1)
            .send(bodyRequest)
            .end(function(err, res){           
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("success");
                done();
            });
    });

    // enroll newPatient2 to org1Service1
    before((done) => {
        const bodyRequest = {
            "user": newPatient2.id,
            "service": org1Service1.id,
            "status": "active"
        }
        chai.request(Config.server).post('omr/api/v1/services/' + org1Service1.id + '/user/enroll')
            .set('Accept', 'application/json')
            .set('token', orgAdminToken1)
            .send(bodyRequest)
            .end(function(err, res){           
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("success");
                done();
            });
    });

    // enroll newPatient2 to org2Service1
    before((done) => {
        const bodyRequest = {
            "user": newPatient2.id,
            "service": org2Service1.id,
            "status": "active"
        }
        chai.request(Config.server).post('omr/api/v1/services/' + org2Service1.id + '/user/enroll')
            .set('Accept', 'application/json')
            .set('token', orgAdminToken2)
            .send(bodyRequest)
            .end(function(err, res){           
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("success");
                done();
            });
    });

    // enroll newPatient1 to org2Service1
    before((done) => {
        const bodyRequest = {
            "user": newPatient1.id,
            "service": org2Service1.id,
            "status": "active"
        }
        chai.request(Config.server).post('omr/api/v1/services/' + org2Service1.id + '/user/enroll')
            .set('Accept', 'application/json')
            .set('token', orgAdminToken2)
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

    let newPatient2Token = '';
    before((done) => {
        chai.request(Config.server).get('common/api/v1/login')
            .set('Accept', 'application/json')
            .set('user-id', newPatient2.id)
            .set('password', newPatient2.secret)
            .set('login-org', newPatient2.ca_org)
            .set('login-channel', Config.channel)
            .set('signature', "")
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("token");
                expect(res.body.token).to.not.equal("");
                newPatient2Token = res.body.token;
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

    // create consent to o1d2
    before((done) => {
        const bodyRequest = {
            "patient_id": newPatient1.id,
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

    // create consent to o2d3
    before((done) => {
        const bodyRequest = {
            "patient_id": newPatient1.id,
            "service_id": org2Service1.id,
            "target_id": org2Service1.id,
            "datatype_id": org2Service1.datatypes[0].datatype_id,
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

    // create consent to o1d1
    before((done) => {
        const bodyRequest = {
            "patient_id": newPatient2.id,
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
            .set('token', newPatient2Token)
            .send(bodyRequest)
            .end(function(err, res){           
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("success");
                done();
            });
    });

    // create consent to o2d4
    before((done) => {
        const bodyRequest = {
            "patient_id": newPatient2.id,
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
            .set('token', newPatient2Token)
            .send(bodyRequest)
            .end(function(err, res){           
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("success");
                done();
            });
    });

    // upload user data
    before((done) => {
        const bodyRequest = {
            "test data 1": "some test data 1 for o1s1d1"
        }
        chai.request(Config.server).post('omr/api/v1/services/' + org1Service1.id + '/users/' +
         newPatient1.id + '/datatype/' + org1Service1.datatypes[0].datatype_id + '/upload')
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

    before((done) => {
        const bodyRequest = {
            "test data 2": "some test data 2 for o1s1d1"
        }
        chai.request(Config.server).post('omr/api/v1/services/' + org1Service1.id + '/users/' +
         newPatient1.id + '/datatype/' + org1Service1.datatypes[0].datatype_id + '/upload')
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

    // upload user data for newPatient2
    before((done) => {
        const bodyRequest = {
            "test data 1": "some test data 1 for o1s1d1 for newPatient2"
        }
        chai.request(Config.server).post('omr/api/v1/services/' + org1Service1.id + '/users/' +
         newPatient2.id + '/datatype/' + org1Service1.datatypes[0].datatype_id + '/upload')
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

    // upload user data for newPatient1
    before((done) => {
        const bodyRequest = {
            "test data 1": "some test data 1 for o2s1d1"
        }
        chai.request(Config.server).post('omr/api/v1/services/' + org2Service1.id + '/users/' +
         newPatient1.id + '/datatype/' + org2Service1.datatypes[0].datatype_id + '/upload')
            .set('Accept', 'application/json')
            .set('token', defaultOrg2Service1AdminToken)
            .send(bodyRequest)
            .end(function(err, res){
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("success");
                done();
            });
    });

    // upload user data for newPatient1
    before((done) => {
        const bodyRequest = {
            "test data 2": "some test data 2 for o2s1d1"
        }
        chai.request(Config.server).post('omr/api/v1/services/' + org2Service1.id + '/users/' +
         newPatient1.id + '/datatype/' + org2Service1.datatypes[0].datatype_id + '/upload')
            .set('Accept', 'application/json')
            .set('token', defaultOrg2Service1AdminToken)
            .send(bodyRequest)
            .end(function(err, res){
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("success");
                timestampStart = Math.floor(Date.now() / 1000);
                done();
            });
    });

    // upload user data for newPatient1
    before((done) => {
        const bodyRequest = {
            "test data 3": "some test data 3 for o2s1d1"
        }
        chai.request(Config.server).post('omr/api/v1/services/' + org2Service1.id + '/users/' +
         newPatient1.id + '/datatype/' + org2Service1.datatypes[0].datatype_id + '/upload')
            .set('Accept', 'application/json')
            .set('token', defaultOrg2Service1AdminToken)
            .send(bodyRequest)
            .end(function(err, res){
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("success");
                timeStampTwoStart = Math.floor(Date.now() / 1000);
                done();
            });
    });

    // download user data by default service admin of org1Service1
    before((done) => {
        chai.request(Config.server).get('omr/api/v1/services/' + org1Service1.id + '/users/' +
         newPatient1.id + '/datatypes/' + org1Service1.datatypes[0].datatype_id + '/download')
            .set('Accept', 'application/json')
            .set('token', defaultOrg1Service1AdminToken)
            .end(function(err, res){
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.be.an('array');
                expect(res.body).to.have.lengthOf(2);
                expect(JSON.stringify(res.body[0].data)).to.equal(JSON.stringify({
                    "test data 1": "some test data 1 for o1s1d1"
                }));
                expect(JSON.stringify(res.body[1].data)).to.equal(JSON.stringify({
                    "test data 2": "some test data 2 for o1s1d1"
                }));
                done();
            });
    });

    // download user data by default service admin of org1Service1
    before((done) => {
        chai.request(Config.server).get('omr/api/v1/services/' + org1Service1.id + '/users/' +
         newPatient2.id + '/datatypes/' + org1Service1.datatypes[0].datatype_id + '/download')
            .set('Accept', 'application/json')
            .set('token', defaultOrg1Service1AdminToken)
            .end(function(err, res){
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.be.an('array');
                expect(res.body).to.have.lengthOf(1);
                expect(JSON.stringify(res.body[0].data)).to.equal(JSON.stringify({
                    "test data 1": "some test data 1 for o1s1d1 for newPatient2"
                }));
                timestampStop = Math.floor(Date.now() / 1000);
                done();
            });
    });

    // download user data by default service admin of org2Service1
    before((done) => {
        chai.request(Config.server).get('omr/api/v1/services/' + org2Service1.id + '/users/' +
         newPatient1.id + '/datatypes/' + org2Service1.datatypes[0].datatype_id + '/download')
            .set('Accept', 'application/json')
            .set('token', defaultOrg2Service1AdminToken)
            .end(function(err, res){
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.be.an('array');
                expect(res.body).to.have.lengthOf(3);
                expect(JSON.stringify(res.body[0].data)).to.equal(JSON.stringify({
                    "test data 1": "some test data 1 for o2s1d1"
                }));
                expect(JSON.stringify(res.body[1].data)).to.equal(JSON.stringify({
                    "test data 2": "some test data 2 for o2s1d1"
                }));
                expect(JSON.stringify(res.body[2].data)).to.equal(JSON.stringify({
                    "test data 3": "some test data 3 for o2s1d1"
                }));
                done();
            });
    });

    // revoke consent from org2Service1 by newPatient2
    before((done) => {
        const bodyRequest = {
            "patient_id": newPatient2.id,
            "service_id": org2Service1.id,
            "target_id": org2Service1.id,
            "datatype_id": org2Service1.datatypes[1].datatype_id,
            "option": [
                "deny"
            ],
            "expiration": 0
        }
        chai.request(Config.server).post('omr/api/v1/consents/patientData')
            .set('Accept', 'application/json')
            .set('token', newPatient2Token)
            .send(bodyRequest)
            .end(function(err, res){
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("success");
                timeStampTwoStop = Math.floor(Date.now() / 1000);
                done();
            });
    });

    // check logs of org1service1 by newAuditorOrg1Service1 searched by service_id
    it('Should return logs searched by service_id', function (done) {
        chai.request(Config.server).get('omr/api/v1/history?service_id=' + org1Service1.id + '&latest_only=false&maxNum=20') 
            .set('Accept',  'application/json')
            .set('token', newAuditorOrg1Service1Token)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.body).to.be.an('array');
                expect(res.body).to.have.lengthOf(8);
                expect(res.body[0].type).to.equal('PutConsentPatientData');
                expect(JSON.stringify(res.body[0].data)).to.equal(JSON.stringify({"option":["write"]}));
                expect(res.body[1].type).to.equal('PutConsentPatientData');
                expect(JSON.stringify(res.body[1].data)).to.equal(JSON.stringify({"option":["write"]}));
                expect(res.body[2].type).to.equal('PutConsentPatientData');
                expect(JSON.stringify(res.body[2].data)).to.equal(JSON.stringify({"option":["write"]}));
                expect(res.body[3].type).to.equal('UploadUserData');
                expect(JSON.stringify(res.body[3].data)).to.equal(JSON.stringify({}));
                expect(res.body[4].type).to.equal('UploadUserData');
                expect(JSON.stringify(res.body[4].data)).to.equal(JSON.stringify({}));
                expect(res.body[5].type).to.equal('UploadUserData');
                expect(JSON.stringify(res.body[5].data)).to.equal(JSON.stringify({}));
                expect(res.body[6].type).to.equal('DownloadUserData');
                expect(JSON.stringify(res.body[6].data)).to.equal(JSON.stringify({}));
                expect(res.body[7].type).to.equal('DownloadUserData');
                expect(JSON.stringify(res.body[7].data)).to.equal(JSON.stringify({}));
                done();
            });
    });

    // check logs of org1service1 by newAuditorOrg1Service1 searched by datatype_id
    it('Should return logs searched by datatype_id', function (done) {
        chai.request(Config.server).get('omr/api/v1/history?datatype_id=' + org1Service1.datatypes[1].datatype_id + '&latest_only=false&maxNum=20') 
            .set('Accept',  'application/json')
            .set('token', newAuditorOrg1Service1Token)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.body).to.be.an('array');
                expect(res.body).to.have.lengthOf(1);
                expect(res.body[0].type).to.equal('PutConsentPatientData');
                expect(JSON.stringify(res.body[0].data)).to.equal(JSON.stringify({"option":["write"]}));
                done();
            });
    });

    // check logs of org1service1 by newAuditorOrg1Service1 searched by datatype_id
    it('Should return logs searched by datatype_id', function (done) {
        chai.request(Config.server).get('omr/api/v1/history?datatype_id=' + org1Service1.datatypes[0].datatype_id + '&latest_only=false&maxNum=20') 
            .set('Accept',  'application/json')
            .set('token', newAuditorOrg1Service1Token)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.body).to.be.an('array');
                expect(res.body).to.have.lengthOf(7);
                expect(res.body[0].type).to.equal('PutConsentPatientData');
                expect(JSON.stringify(res.body[0].data)).to.equal(JSON.stringify({"option":["write"]}));
                expect(res.body[1].type).to.equal('PutConsentPatientData');
                expect(JSON.stringify(res.body[1].data)).to.equal(JSON.stringify({"option":["write"]}));
                expect(res.body[2].type).to.equal('UploadUserData');
                expect(JSON.stringify(res.body[2].data)).to.equal(JSON.stringify({}));
                expect(res.body[3].type).to.equal('UploadUserData');
                expect(JSON.stringify(res.body[3].data)).to.equal(JSON.stringify({}));
                expect(res.body[4].type).to.equal('UploadUserData');
                expect(JSON.stringify(res.body[4].data)).to.equal(JSON.stringify({}));
                expect(res.body[5].type).to.equal('DownloadUserData');
                expect(JSON.stringify(res.body[5].data)).to.equal(JSON.stringify({}));
                expect(res.body[6].type).to.equal('DownloadUserData');
                expect(JSON.stringify(res.body[6].data)).to.equal(JSON.stringify({}));
                done();
            });
    });

    // check logs of org1service1 by newAuditorOrg1Service1 searched by patient_id
    it('Should return logs searched by patient_id', function (done) {
        chai.request(Config.server).get('omr/api/v1/history?patient_id=' + newPatient2.id + '&latest_only=false&maxNum=20') 
            .set('Accept',  'application/json')
            .set('token', newAuditorOrg1Service1Token)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.body).to.be.an('array');
                expect(res.body).to.have.lengthOf(3);
                expect(res.body[0].type).to.equal('PutConsentPatientData');
                expect(JSON.stringify(res.body[0].data)).to.equal(JSON.stringify({"option":["write"]}));
                expect(res.body[1].type).to.equal('UploadUserData');
                expect(JSON.stringify(res.body[1].data)).to.equal(JSON.stringify({}));
                expect(res.body[2].type).to.equal('DownloadUserData');
                expect(JSON.stringify(res.body[2].data)).to.equal(JSON.stringify({}));
                done();
            });
    });

    // check logs of org1service1 by newAuditorOrg1Service1 searched by service_id with latest_only flag
    it('Should return logs searched by service_id with latest_only flag', function (done) {
        chai.request(Config.server).get('omr/api/v1/history?service_id=' + org1Service1.id + '&latest_only=true&maxNum=20') 
            .set('Accept',  'application/json')
            .set('token', newAuditorOrg1Service1Token)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.body).to.be.an('array');
                expect(res.body).to.have.lengthOf(1);
                expect(res.body[0].type).to.equal('DownloadUserData');
                expect(JSON.stringify(res.body[0].data)).to.equal(JSON.stringify({}));
                done();
            });
    });

    // check logs of org1service1 by newAuditorOrg1Service1 searched by service_id with maxNum flag
    it('Should return logs searched by service_id with maxNum flag', function (done) {
        chai.request(Config.server).get('omr/api/v1/history?service_id=' + org1Service1.id + '&latest_only=false&maxNum=2') 
            .set('Accept',  'application/json')
            .set('token', newAuditorOrg1Service1Token)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.body).to.be.an('array');
                expect(res.body).to.have.lengthOf(2);
                expect(res.body[0].type).to.equal('PutConsentPatientData');
                expect(JSON.stringify(res.body[0].data)).to.equal(JSON.stringify({"option":["write"]}));
                expect(res.body[1].type).to.equal('PutConsentPatientData');
                expect(JSON.stringify(res.body[1].data)).to.equal(JSON.stringify({"option":["write"]}));
                done();
            });
    });

    // check logs of org2service1 by newAuditorOrg2Service1 searched by service_id
    it('Should return logs of org2Service1 searched by service_id', function (done) {
        chai.request(Config.server).get('omr/api/v1/history?service_id=' + org2Service1.id + '&latest_only=false&maxNum=20') 
            .set('Accept',  'application/json')
            .set('token', newAuditorOrg2Service1Token)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.body).to.be.an('array');
                expect(res.body).to.have.lengthOf(7);
                expect(res.body[0].type).to.equal('PutConsentPatientData');
                expect(JSON.stringify(res.body[0].data)).to.equal(JSON.stringify({"option":["write"]}));
                expect(res.body[1].type).to.equal('PutConsentPatientData');
                expect(JSON.stringify(res.body[1].data)).to.equal(JSON.stringify({"option":["read"]}));
                expect(res.body[2].type).to.equal('UploadUserData');
                expect(JSON.stringify(res.body[2].data)).to.equal(JSON.stringify({}));
                expect(res.body[3].type).to.equal('UploadUserData');
                expect(JSON.stringify(res.body[3].data)).to.equal(JSON.stringify({}));
                expect(res.body[4].type).to.equal('UploadUserData');
                expect(JSON.stringify(res.body[4].data)).to.equal(JSON.stringify({}));
                expect(res.body[5].type).to.equal('DownloadUserData');
                expect(JSON.stringify(res.body[5].data)).to.equal(JSON.stringify({}));
                expect(res.body[6].type).to.equal('PutConsentPatientData');
                expect(JSON.stringify(res.body[6].data)).to.equal(JSON.stringify({"option":["deny"]}));
                done();
            });
    });

    // revoke auditor admin permission from newAuditorOrg2Service1 by service admin
    it('Should successfully revoke auditor permission from user', function (done) {
        chai.request(Config.server).delete('omr/api/v1/users/' + newAuditorOrg2Service1.id + '/permissions/audit/' + org2Service1.id) 
            .set('Accept',  'application/json')
            .set('token', defaultOrg2Service1AdminToken)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("success");
                done();
            });
    });

    // check logs of org2service1 by newAuditorOrg2Service1 after auditor permission was revoked
    it('Should return a blank list of logs after permission of auditor admin was revoked', function (done) {
        chai.request(Config.server).get('omr/api/v1/history?service_id=' + org2Service1.id + '&latest_only=false&maxNum=20') 
            .set('Accept',  'application/json')
            .set('token', newAuditorOrg2Service1Token)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.body).to.be.an('array');
                expect(res.body).to.have.lengthOf(0);
                done();
            });
    });

    // check logs of org1service1 by newAuditorOrg1Service1 searched by several options: service_id, patient_id, datatype_id
    it('Should return logs searched by several options', function (done) {
        chai.request(Config.server).get('omr/api/v1/history?service_id=' + org1Service1.id +
         '&patient_id=' + newPatient1.id + '&datatype_id=' + org1Service1.datatypes[0].datatype_id + '&latest_only=false&maxNum=20') 
            .set('Accept',  'application/json')
            .set('token', newAuditorOrg1Service1Token)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.body).to.be.an('array');
                expect(res.body).to.have.lengthOf(4);
                expect(res.body[0].type).to.equal('PutConsentPatientData');
                expect(JSON.stringify(res.body[0].data)).to.equal(JSON.stringify({"option":["write"]}));
                expect(res.body[1].type).to.equal('UploadUserData');
                expect(JSON.stringify(res.body[1].data)).to.equal(JSON.stringify({}));
                expect(res.body[2].type).to.equal('UploadUserData');
                expect(JSON.stringify(res.body[2].data)).to.equal(JSON.stringify({}));
                expect(res.body[3].type).to.equal('DownloadUserData');
                expect(JSON.stringify(res.body[3].data)).to.equal(JSON.stringify({}));
                done();
            });
    });

    // check logs of org1service1 by newAuditorOrg1Service1 searched by several options: patient_id, timestamp
    it('Should return logs searched by several options', function (done) {
        chai.request(Config.server).get('omr/api/v1/history?&patient_id=' + newPatient1.id + '&start_timestamp='
         + timestampStart + '&end_timestamp=' + timestampStop) 
            .set('Accept',  'application/json')
            .set('token', newAuditorOrg1Service1Token)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.body).to.be.an('array');
                expect(res.body).to.have.lengthOf(1);
                expect(res.body[0].type).to.equal('DownloadUserData');
                expect(JSON.stringify(res.body[0].data)).to.equal(JSON.stringify({}));
                done();
            });
    });

    // checking logs by auditor searched by service_id with timestamp option
    it('Should return logs of org1Service1 searched by service_id with timestamp option', function (done) {
        chai.request(Config.server).get('omr/api/v1/history?service_id=' + org1Service1.id + 
        '&start_timestamp=' + timestampStart + '&end_timestamp=' + timestampStop + '&latest_only=false&maxNum=20') 
            .set('Accept',  'application/json')
            .set('token', newAuditorOrg1Service1Token)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.body).to.be.an('array');
                expect(res.body).to.have.lengthOf(2);
                expect(res.body[0].type).to.equal('DownloadUserData');
                expect(JSON.stringify(res.body[0].data)).to.equal(JSON.stringify({}));
                expect(res.body[1].type).to.equal('DownloadUserData');
                expect(JSON.stringify(res.body[1].data)).to.equal(JSON.stringify({}));
                done();
            });
    });

    // checking logs by newAuditorOrg1Service1Token searched by consent_owner_target_id
    it('Should return logs of org1 searched by consent_owner_target_id = service_id by auditor', function (done) {
        chai.request(Config.server).get('omr/api/v1/history?consent_owner_target_id=' + org1Service1.id + '&latest_only=false&maxNum=20') 
            .set('Accept',  'application/json')
            .set('token', newAuditorOrg1Service1Token)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.body).to.be.an('array');
                expect(res.body).to.have.lengthOf(5);
                expect(res.body[0].type).to.equal('PutConsentPatientData');
                expect(JSON.stringify(res.body[0].data)).to.equal(JSON.stringify({"option":["write"]}));
                expect(res.body[1].type).to.equal('PutConsentPatientData');
                expect(JSON.stringify(res.body[1].data)).to.equal(JSON.stringify({"option":["write"]}));
                expect(res.body[2].type).to.equal('PutConsentPatientData');
                expect(JSON.stringify(res.body[2].data)).to.equal(JSON.stringify({"option":["write"]}));
                expect(res.body[3].type).to.equal('DownloadUserData');
                expect(JSON.stringify(res.body[3].data)).to.equal(JSON.stringify({}));
                expect(res.body[4].type).to.equal('DownloadUserData');
                expect(JSON.stringify(res.body[4].data)).to.equal(JSON.stringify({}));
                done();
            });
    });
});