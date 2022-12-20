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

describe('Checking contract logs by org admin, service admin, auditor', function() {

    let contractId;
    let datatype5 = {};

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

    const org3 = {
        "id": idGenerator(),
        "secret": "pass3",
        "name": "org3",
        "ca_org": Config.caOrg,
        "email": "org3@email.com",
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

    // register org3
    before((done) => {  
        chai.request(Config.server).post('omr/api/v1/orgs') 
            .set('Accept',  'application/json')
            .set('token', Config.sysAdminToken)
            .send(org3)
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

    let orgAdminToken3;
    // org3admin token
    before((done) => {
        chai.request(Config.server).get('common/api/v1/login')
            .set('Accept', 'application/json')
            .set('user-id', org3.id)
            .set('password', org3.secret)
            .set('login-org', org3.ca_org)
            .set('login-channel', Config.channel)
            .set('signature', "")
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.not.deep.equal({})
                expect(res.body.token).to.not.equal("")
                orgAdminToken3 = res.body.token;
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

    // register datatype5
    before((done) => {
        chai.request(Config.server).post('omr/api/v1/datatypes') 
            .set('Accept',  'application/json')
            .set('token', Config.sysAdminToken)
            .send({id: idGenerator(), description: "sample data 5"})
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                datatype5.id = res.body.id;
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

    // register org1Service2
    let org1Service2 = {};    
    before((done) => {
        org1Service2 = {
            "id": idGenerator(),
            "name": "o1s2",
            "secret": "org1Service2pass",
            "ca_org": Config.caOrg,
            "email": "org1Service2email@example.com",
            "org_id": org1.id,
            "summary": "New service under org 1. Has one datatype",
            "terms": {
                "term1" : "example term"
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
                "sample data 1": "service sample data 1"
            }
        };

        chai.request(Config.server).post('omr/api/v1/services') 
            .set('Accept',  'application/json')
            .set('token', orgAdminToken1)
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
            "org_id": org2.id,
            "summary": "New service under org 2. Has three datatypes",
            "terms": {
                "term1" : "example term",
                "term2" : "example term",
                "term3" : "example term"
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
                "sample data 2": "service sample data 3",
                "sample data 3": "service sample data 4"
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

    // register of org3Service1
    let org3Service1 = {};    
    before((done) => {
        org3Service1 = {
            "id": idGenerator(),
            "name": "org3Service1",
            "secret": "org3Service1pass",
            "ca_org": Config.caOrg,
            "email": "org3Service1email@example.com",
            "org_id": org3.id,
            "summary": "New service under org 3. Has one datatype",
            "terms": {
                "term1" : "example term"
            },
            "payment_required": "yes",
            "datatypes": [
                {
                    "datatype_id": datatype5.id,
                    "access":[
                        "write",
                        "read"
                    ]
                }
            ],
            "solution_private_data": {
                "sample data 1": "service sample data 1"
            }
        };

        chai.request(Config.server).post('omr/api/v1/services') 
            .set('Accept',  'application/json')
            .set('token', orgAdminToken3)
            .send(org3Service1)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("success");
                done();
                });
    });

    // create org1Admin
    const org1Admin = {
        "id": idGenerator(),
        "secret": "org1Admin",
        "name": "org1Admin",
        "role": "user",
        "org" : org1.id,
        "email":  "org1Adminemail@example.com",
        "ca_org": Config.caOrg,
        "data": {
          "address": "2 User St"
        }
    };

    // create org2Admin
    const org2Admin = {
        "id": idGenerator(),
        "secret": "org2Admin",
        "name": "org2Admin",
        "role": "user",
        "org" : org2.id,
        "email":  "org2Adminemail@example.com",
        "ca_org": Config.caOrg,
        "data": {
          "address": "2 User St"
        }
    };

    // create org3Admin
    const org3Admin = {
        "id": idGenerator(),
        "secret": "org3Admin",
        "name": "org3Admin",
        "role": "user",
        "org" : org3.id,
        "email":  "org3Adminemail@example.com",
        "ca_org": Config.caOrg,
        "data": {
          "address": "3 User St"
        }
    };

    // register org1Admin
    before((done) => {
        chai.request(Config.server).post('omr/api/v1/users') 
            .set('Accept',  'application/json')
            .set('token', orgAdminToken1)
            .send(org1Admin)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("success");
                done();
            });
    });

    // register org2Admin
    before((done) => {
        chai.request(Config.server).post('omr/api/v1/users') 
            .set('Accept',  'application/json')
            .set('token', orgAdminToken2)
            .send(org2Admin)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("success");
                done();
            });
    });

    // register org3Admin
    before((done) => {
        chai.request(Config.server).post('omr/api/v1/users') 
            .set('Accept',  'application/json')
            .set('token', orgAdminToken3)
            .send(org3Admin)
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
        chai.request(Config.server).put('omr/api/v1/users/' + org1Admin.id + '/permissions/admin/' + org1.id) 
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

    // grant org admin permission to org2Admin
    before((done) => {
        chai.request(Config.server).put('omr/api/v1/users/' + org2Admin.id + '/permissions/admin/' + org2.id) 
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

    // grant org admin permission to org3Admin
    before((done) => {
        chai.request(Config.server).put('omr/api/v1/users/' + org3Admin.id + '/permissions/admin/' + org3.id) 
            .set('Accept',  'application/json')
            .set('token', orgAdminToken3)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("success");
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

    let org2AdminToken = '';
    before((done) => {
        chai.request(Config.server).get('common/api/v1/login')
            .set('Accept', 'application/json')
            .set('user-id', org2Admin.id)
            .set('password', org2Admin.secret)
            .set('login-org', org2Admin.ca_org)
            .set('login-channel', Config.channel)
            .set('signature', "")
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("token");
                expect(res.body.token).to.not.equal("");
                org2AdminToken = res.body.token;
                done();
            });
    });

    let org3AdminToken = '';
    before((done) => {
        chai.request(Config.server).get('common/api/v1/login')
            .set('Accept', 'application/json')
            .set('user-id', org3Admin.id)
            .set('password', org3Admin.secret)
            .set('login-org', org3Admin.ca_org)
            .set('login-channel', Config.channel)
            .set('signature', "")
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("token");
                expect(res.body.token).to.not.equal("");
                org3AdminToken = res.body.token;
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

    let defaultOrg3Service1AdminToken = '';
    before((done) => {
        chai.request(Config.server).get('common/api/v1/login')
            .set('Accept', 'application/json')
            .set('user-id', org3Service1.id)
            .set('password', org3Service1.secret)
            .set('login-org', org3Service1.ca_org)
            .set('login-channel', Config.channel)
            .set('signature', "")
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("token");
                expect(res.body.token).to.not.equal("");
                defaultOrg3Service1AdminToken = res.body.token;
                done();
            });
    });

    // upload owner data to o1d1 by o1s1 admin
    before((done) => {
        const bodyRequest = {
            "test owner data": "owner data o1s1d1"
        }
        chai.request(Config.server).post('omr/api/v1/services/' + org1Service1.id + '/datatype/'
         + org1Service1.datatypes[0].datatype_id + '/upload')
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

    it('Should successfully create contract by data owner', function (done) {
        const bodyRequest = {
            "owner_org_id": org1.id,
            "owner_service_id": org1Service1.id,
            "requester_org_id": org2.id,
            "requester_service_id": org2Service1.id,
            "contract_terms": {"contract terms":"testing"}
          }
        chai.request(Config.server).post('omr/api/v1/contracts')
            .set('Accept', 'application/json')
            .set('token', defaultOrg1Service1AdminToken)
            .send(bodyRequest)
            .end(function(err, res){           
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("success");
                contractId = res.body.contract_id;
                done();
            });
    });
    
    it('Should successfully change contract terms by data requester', function (done) {
        const newTermsByRequester = {
            "contract_id": contractId,
            "contract_terms": {"contract terms":"updated terms by requester"}
          }
        chai.request(Config.server).post('omr/api/v1/contracts/' + contractId + '/changeTerms')
            .set('Accept', 'application/json')
            .set('token', defaultOrg2Service1AdminToken)
            .send(newTermsByRequester)
            .end(function(err, res){           
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("success");
                done();
            });
    });

    it('Should successfully change contract terms by data owner', function (done) {
        const newTermsByOwner = {
            "contract_id": contractId,
            "contract_terms": {"contract terms":"updated terms by owner"}
          }
        chai.request(Config.server).post('omr/api/v1/contracts/' + contractId + '/changeTerms')
            .set('Accept', 'application/json')
            .set('token', defaultOrg1Service1AdminToken)
            .send(newTermsByOwner)
            .end(function(err, res){          
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("success");
                done();
            });
    });

    it('Should successfully sign contract by requester', function (done) {
        let bodyRequest = {
            "contract_id": contractId,
            "contract_terms": {},
            "signed_by": org2Service1.id
          }
        chai.request(Config.server).post('omr/api/v1/contracts/' + contractId + '/sign')
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

    it('Should successfully pay contract by requester', function (done) {
        let bodyRequest = {
            "contract_id": contractId,
            "contract_terms": {}
          }
        chai.request(Config.server).post('omr/api/v1/contracts/' + contractId + '/payment')
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

    it('Should successfully verify payment of contract by data owner', function (done) {
        let bodyRequest = {
            "contract_id": contractId,
            "contract_terms": {}
          }
        chai.request(Config.server).post('omr/api/v1/contracts/' + contractId + '/verify')
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

    // download before permission is granted - should fail
    it('Should NOT successfully download owner data by requester before permission is granted', function (done) {
        chai.request(Config.server).get('omr/api/v1/contracts/' + contractId + 
        '/ownerdata/' + org1Service1.datatypes[0].datatype_id + '/downloadAsRequester')
            .set('Accept', 'application/json')
            .set('token', defaultOrg2Service1AdminToken)
            .end(function(err, res){
                expect(err).to.be.null;
                expect(res.status).to.equal(500);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("error");
                done();
            });
    });

    it('Should successfully grant download permission by data owner', function (done) {
        chai.request(Config.server).post('omr/api/v1/contracts/' + contractId +
         '/max_num_download/3/permission?datatype_id=' + org1Service1.datatypes[0].datatype_id)
            .set('Accept', 'application/json')
            .set('token', defaultOrg1Service1AdminToken)
            .end(function(err, res){
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("success");
                done();
            });
    });

    it('Should successfully download owner data by requester', function (done) {
        chai.request(Config.server).get('omr/api/v1/contracts/' + contractId + 
        '/ownerdata/' + org1Service1.datatypes[0].datatype_id + '/downloadAsRequester')
            .set('Accept', 'application/json')
            .set('token', defaultOrg2Service1AdminToken)
            .end(function(err, res){                           
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.be.an('array');
                expect(res.body).to.have.lengthOf(1);
                expect(JSON.stringify(res.body[0].data)).to.equal(JSON.stringify({
                    "test owner data":"owner data o1s1d1"
                }));
                done();
            });
    });

    it('Should successfully get contract by requester', function (done) {
        chai.request(Config.server).get('omr/api/v1/contracts/' + contractId)
            .set('Accept', 'application/json')
            .set('token', defaultOrg2Service1AdminToken)
            .end(function(err, res){
                expect(err).to.be.null;
                expect(res.body).to.have.property("contract_id");
                expect(res.body).to.have.property("owner_org_id");
                expect(res.body.owner_org_id).to.equal(org1.id);
                expect(res.body).to.have.property("owner_service_id");
                expect(res.body.owner_service_id).to.equal(org1Service1.id);
                expect(res.body).to.have.property("requester_org_id");
                expect(res.body.requester_org_id).to.equal(org2.id);
                expect(res.body).to.have.property("requester_service_id");
                expect(res.body.requester_service_id).to.equal(org2Service1.id);
                expect(res.body).to.have.property("contract_terms");
                expect(JSON.stringify(res.body.contract_terms)).to.equal(JSON.stringify({
                    "contract terms":"updated terms by owner"
                }));
                expect(res.body).to.have.property("state");
                expect(res.body.state).to.equal("downloadReady");
                expect(res.body).to.have.property("create_date");
                expect(res.body).to.have.property("update_date");
                expect(res.body).to.have.property("contract_details");
                expect(res.body.contract_details).to.be.an('array');
                expect(res.body.contract_details).to.have.lengthOf(8);
                expect(res.body.contract_details[0].contract_detail_type).to.equal("request");
                expect(JSON.stringify(res.body.contract_details[0].contract_detail_terms)).to.equal(JSON.stringify({
                    "contract terms":"testing"
                }));
                expect(res.body.contract_details[1].contract_detail_type).to.equal("terms");
                expect(JSON.stringify(res.body.contract_details[1].contract_detail_terms)).to.equal(JSON.stringify({
                    "contract terms":"updated terms by requester"
                }));
                expect(res.body.contract_details[2].contract_detail_type).to.equal("terms");
                expect(JSON.stringify(res.body.contract_details[2].contract_detail_terms)).to.equal(JSON.stringify({
                    "contract terms":"updated terms by owner"
                }));
                expect(res.body.contract_details[3].contract_detail_type).to.equal("sign");
                expect(res.body.contract_details[4].contract_detail_type).to.equal("payment");
                expect(res.body.contract_details[5].contract_detail_type).to.equal("verify");
                expect(res.body.contract_details[6].contract_detail_type).to.equal("permission");
                expect(res.body.contract_details[7].contract_detail_type).to.equal("download");
                expect(res.body).to.have.property("payment_required");
                expect(res.body.payment_required).to.equal("yes");
                expect(res.body).to.have.property("payment_verified");
                expect(res.body.payment_verified).to.equal("yes");
                expect(res.body).to.have.property("max_num_download");
                expect(res.body.max_num_download).to.equal(3);
                expect(res.body).to.have.property("num_download");
                expect(res.body.num_download).to.equal(1);
                done();
            });
    });

    it('Should successfully get contract by data owner', function (done) {
        chai.request(Config.server).get('omr/api/v1/contracts/' + contractId)
            .set('Accept', 'application/json')
            .set('token', defaultOrg1Service1AdminToken)
            .end(function(err, res){
                expect(err).to.be.null;
                expect(res.body).to.have.property("contract_id");
                expect(res.body).to.have.property("owner_org_id");
                expect(res.body.owner_org_id).to.equal(org1.id);
                expect(res.body).to.have.property("owner_service_id");
                expect(res.body.owner_service_id).to.equal(org1Service1.id);
                expect(res.body).to.have.property("requester_org_id");
                expect(res.body.requester_org_id).to.equal(org2.id);
                expect(res.body).to.have.property("requester_service_id");
                expect(res.body.requester_service_id).to.equal(org2Service1.id);
                expect(res.body).to.have.property("contract_terms");
                expect(JSON.stringify(res.body.contract_terms)).to.equal(JSON.stringify({
                    "contract terms":"updated terms by owner"
                }));
                expect(res.body).to.have.property("state");
                expect(res.body.state).to.equal("downloadReady");
                expect(res.body).to.have.property("create_date");
                expect(res.body).to.have.property("update_date");
                expect(res.body).to.have.property("contract_details");
                expect(res.body.contract_details).to.be.an('array');
                expect(res.body.contract_details).to.have.lengthOf(8);
                expect(res.body.contract_details[0].contract_detail_type).to.equal("request");
                expect(JSON.stringify(res.body.contract_details[0].contract_detail_terms)).to.equal(JSON.stringify({
                    "contract terms":"testing"
                }));
                expect(res.body.contract_details[1].contract_detail_type).to.equal("terms");
                expect(JSON.stringify(res.body.contract_details[1].contract_detail_terms)).to.equal(JSON.stringify({
                    "contract terms":"updated terms by requester"
                }));
                expect(res.body.contract_details[2].contract_detail_type).to.equal("terms");
                expect(JSON.stringify(res.body.contract_details[2].contract_detail_terms)).to.equal(JSON.stringify({
                    "contract terms":"updated terms by owner"
                }));
                expect(res.body.contract_details[3].contract_detail_type).to.equal("sign");
                expect(res.body.contract_details[4].contract_detail_type).to.equal("payment");
                expect(res.body.contract_details[5].contract_detail_type).to.equal("verify");
                expect(res.body.contract_details[6].contract_detail_type).to.equal("permission");
                expect(res.body.contract_details[7].contract_detail_type).to.equal("download");
                expect(res.body).to.have.property("payment_required");
                expect(res.body.payment_required).to.equal("yes");
                expect(res.body).to.have.property("payment_verified");
                expect(res.body.payment_verified).to.equal("yes");
                expect(res.body).to.have.property("max_num_download");
                expect(res.body.max_num_download).to.equal(3);
                expect(res.body).to.have.property("num_download");
                expect(res.body.num_download).to.equal(1);
                done();
            });
    });

    // checking logs by org1Admin searched by contract_org_id = contract owner id
    it('Should return logs searched by contract_org_id = contract owner id', function (done) {
        chai.request(Config.server).get('omr/api/v1/history?contract_org_id=' + org1.id + '&latest_only=false&maxNum=20') 
            .set('Accept',  'application/json')
            .set('token', org1AdminToken)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.body).to.be.an('array');
                expect(res.body).to.have.lengthOf(8);
                expect(res.body[0].type).to.equal('CreateContract');
                expect(res.body[0].contract).to.equal(contractId);
                expect(JSON.stringify(res.body[0].data)).to.equal(JSON.stringify({}));
                expect(res.body[1].type).to.equal('AddContractDetailrequested');
                expect(JSON.stringify(res.body[1].data)).to.equal(JSON.stringify({}));
                expect(res.body[2].type).to.equal('AddContractDetailcontractReady');
                expect(JSON.stringify(res.body[2].data)).to.equal(JSON.stringify({}));
                expect(res.body[3].type).to.equal('AddContractDetailcontractSigned');
                expect(JSON.stringify(res.body[3].data)).to.equal(JSON.stringify({}));
                expect(res.body[4].type).to.equal('AddContractDetailpaymentDone');
                expect(JSON.stringify(res.body[4].data)).to.equal(JSON.stringify({}));
                expect(res.body[5].type).to.equal('AddContractDetailpaymentVerified');
                expect(JSON.stringify(res.body[5].data)).to.equal(JSON.stringify({}));
                expect(res.body[6].type).to.equal('GivePermissionByContract');
                expect(JSON.stringify(res.body[6].data)).to.equal(JSON.stringify({}));
                expect(res.body[7].type).to.equal('DownloadOwnerDataAsRequester');
                expect(JSON.stringify(res.body[7].data)).to.equal(JSON.stringify({}));
                done();
            });
    });

    // checking logs by org1Admin searched by contract_org_id = contract requester id
    it('Should return logs searched by contract_org_id = contract requester id', function (done) {
        chai.request(Config.server).get('omr/api/v1/history?contract_org_id=' + org2.id + '&latest_only=false&maxNum=20') 
            .set('Accept',  'application/json')
            .set('token', org2AdminToken)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.body).to.be.an('array');
                expect(res.body).to.have.lengthOf(8);
                expect(res.body[0].type).to.equal('CreateContract');
                expect(res.body[0].contract).to.equal(contractId);
                expect(JSON.stringify(res.body[0].data)).to.equal(JSON.stringify({}));
                expect(res.body[1].type).to.equal('AddContractDetailrequested');
                expect(JSON.stringify(res.body[1].data)).to.equal(JSON.stringify({}));
                expect(res.body[2].type).to.equal('AddContractDetailcontractReady');
                expect(JSON.stringify(res.body[2].data)).to.equal(JSON.stringify({}));
                expect(res.body[3].type).to.equal('AddContractDetailcontractSigned');
                expect(JSON.stringify(res.body[3].data)).to.equal(JSON.stringify({}));
                expect(res.body[4].type).to.equal('AddContractDetailpaymentDone');
                expect(JSON.stringify(res.body[4].data)).to.equal(JSON.stringify({}));
                expect(res.body[5].type).to.equal('AddContractDetailpaymentVerified');
                expect(JSON.stringify(res.body[5].data)).to.equal(JSON.stringify({}));
                expect(res.body[6].type).to.equal('GivePermissionByContract');
                expect(JSON.stringify(res.body[6].data)).to.equal(JSON.stringify({}));
                expect(res.body[7].type).to.equal('DownloadOwnerDataAsRequester');
                expect(JSON.stringify(res.body[7].data)).to.equal(JSON.stringify({}));
                done();
            });
    });

    // checking logs by org1Admin searched by contractId
    it('Should return logs searched by contractId', function (done) {
        chai.request(Config.server).get('omr/api/v1/history?contract_id=' + contractId + '&latest_only=false&maxNum=20') 
            .set('Accept',  'application/json')
            .set('token', org2AdminToken)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.body).to.be.an('array');
                expect(res.body).to.have.lengthOf(8);
                expect(res.body[0].type).to.equal('CreateContract');
                expect(res.body[0].contract).to.equal(contractId);
                expect(JSON.stringify(res.body[0].data)).to.equal(JSON.stringify({}));
                expect(res.body[1].type).to.equal('AddContractDetailrequested');
                expect(JSON.stringify(res.body[1].data)).to.equal(JSON.stringify({}));
                expect(res.body[2].type).to.equal('AddContractDetailcontractReady');
                expect(JSON.stringify(res.body[2].data)).to.equal(JSON.stringify({}));
                expect(res.body[3].type).to.equal('AddContractDetailcontractSigned');
                expect(JSON.stringify(res.body[3].data)).to.equal(JSON.stringify({}));
                expect(res.body[4].type).to.equal('AddContractDetailpaymentDone');
                expect(JSON.stringify(res.body[4].data)).to.equal(JSON.stringify({}));
                expect(res.body[5].type).to.equal('AddContractDetailpaymentVerified');
                expect(JSON.stringify(res.body[5].data)).to.equal(JSON.stringify({}));
                expect(res.body[6].type).to.equal('GivePermissionByContract');
                expect(JSON.stringify(res.body[6].data)).to.equal(JSON.stringify({}));
                expect(res.body[7].type).to.equal('DownloadOwnerDataAsRequester');
                expect(JSON.stringify(res.body[7].data)).to.equal(JSON.stringify({}));
                done();
            });
    });

    // checking logs by org1service1 admin searched by contractId
    it('Should return logs searched by contractId', function (done) {
        chai.request(Config.server).get('omr/api/v1/history?contract_id=' + contractId + '&latest_only=false&maxNum=20') 
            .set('Accept',  'application/json')
            .set('token', defaultOrg1Service1AdminToken)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.body).to.be.an('array');
                expect(res.body).to.have.lengthOf(8);
                expect(res.body[0].type).to.equal('CreateContract');
                expect(res.body[0].contract).to.equal(contractId);
                expect(JSON.stringify(res.body[0].data)).to.equal(JSON.stringify({}));
                expect(res.body[1].type).to.equal('AddContractDetailrequested');
                expect(JSON.stringify(res.body[1].data)).to.equal(JSON.stringify({}));
                expect(res.body[2].type).to.equal('AddContractDetailcontractReady');
                expect(JSON.stringify(res.body[2].data)).to.equal(JSON.stringify({}));
                expect(res.body[3].type).to.equal('AddContractDetailcontractSigned');
                expect(JSON.stringify(res.body[3].data)).to.equal(JSON.stringify({}));
                expect(res.body[4].type).to.equal('AddContractDetailpaymentDone');
                expect(JSON.stringify(res.body[4].data)).to.equal(JSON.stringify({}));
                expect(res.body[5].type).to.equal('AddContractDetailpaymentVerified');
                expect(JSON.stringify(res.body[5].data)).to.equal(JSON.stringify({}));
                expect(res.body[6].type).to.equal('GivePermissionByContract');
                expect(JSON.stringify(res.body[6].data)).to.equal(JSON.stringify({}));
                expect(res.body[7].type).to.equal('DownloadOwnerDataAsRequester');
                expect(JSON.stringify(res.body[7].data)).to.equal(JSON.stringify({}));
                done();
            });
    });

    // checking logs by org3Admin searched by contractId
    it('Should return logs searched by contractId', function (done) {
        chai.request(Config.server).get('omr/api/v1/history?contract_id=' + contractId + '&latest_only=false&maxNum=20') 
            .set('Accept',  'application/json')
            .set('token', org3AdminToken)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.body).to.be.an('array');
                expect(res.body).to.have.lengthOf(0);
                done();
            });
    });

    // checking logs by default org3 admin searched by contractId
    it('Should return logs searched by contractId', function (done) {
        chai.request(Config.server).get('omr/api/v1/history?contract_id=' + contractId + '&latest_only=false&maxNum=20') 
            .set('Accept',  'application/json')
            .set('token', orgAdminToken3)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.body).to.be.an('array');
                expect(res.body).to.have.lengthOf(0);
                done();
            });
    });

    // checking logs by org3Service1 admin searched by contractId
    it('Should return logs searched by contractId', function (done) {
        chai.request(Config.server).get('omr/api/v1/history?contract_id=' + contractId + '&latest_only=false&maxNum=20') 
            .set('Accept',  'application/json')
            .set('token', defaultOrg3Service1AdminToken)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.body).to.be.an('array');
                expect(res.body).to.have.lengthOf(0);
                done();
            });
    });

    // checking logs by org3Admin searched by contract_org_id = contract requester id
    it('Should return logs searched by contract_org_id = contract requester id', function (done) {
        chai.request(Config.server).get('omr/api/v1/history?contract_org_id=' + org2.id + '&latest_only=false&maxNum=20') 
            .set('Accept',  'application/json')
            .set('token', org3AdminToken)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.body).to.be.an('array');
                expect(res.body).to.have.lengthOf(0);
                done();
            });
    });

    // checking logs by default org3 admin searched by contract_org_id = contract owner id
    it('Should return logs searched by contract_org_id = contract owner id', function (done) {
        chai.request(Config.server).get('omr/api/v1/history?contract_org_id=' + org1.id + '&latest_only=false&maxNum=20') 
            .set('Accept',  'application/json')
            .set('token', orgAdminToken3)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.body).to.be.an('array');
                expect(res.body).to.have.lengthOf(0);
                done();
            });
    });

    // checking logs by default org3 admin searched by contract_org_id = contract owner id
    it('Should return logs searched by contract_org_id = contract owner id', function (done) {
        chai.request(Config.server).get('omr/api/v1/history?contract_org_id=' + org1.id + '&latest_only=false&maxNum=20') 
            .set('Accept',  'application/json')
            .set('token', defaultOrg3Service1AdminToken)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.body).to.be.an('array');
                expect(res.body).to.have.lengthOf(0);
                done();
            });
    });
});