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

describe('Checking of permission of org users for Contract flow', function() {

    let contractId;

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
                "sample data 1": "service sample data 3",
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

    // register of org2Service2
    let org2Service2 = {};    
    before((done) => {
        org2Service2 = {
            "id": idGenerator(),
            "name": "org2Service2",
            "secret": "org2Service2pass",
            "ca_org": Config.caOrg,
            "email": "org2Service2email@example.com",
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
            .send(org2Service2)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("success");
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

    // create new user which belong to org1
    const newUser4 = {
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
    
    // register newUser4
    before((done) => {
        chai.request(Config.server).post('omr/api/v1/users') 
            .set('Accept',  'application/json')
            .set('token', Config.orgAdminToken1)
            .send(newUser4)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("success");
                done();
            });
    });

    // grant service admin permission of org1Service2 to newUser4
    before((done) => {
        chai.request(Config.server).put('omr/api/v1/users/' + newUser4.id + '/permissions/services/' + org1Service2.id) 
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
    let o1s2AdminToken = '';
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
                o1s2AdminToken = res.body.token;
                done();
            });
    });

    // create new user which belong to org2
    const newUser5 = {
        "id": idGenerator(),
        "secret": "newuser5",
        "name": "New User: " + this.id,
        "role": "user",
        "org" : Config.org2.id,
        "email":  "newuser5email@example.com",
        "ca_org": Config.caOrg,
        "data": {
          "address": "2 User St"
        }
    };
    
    // register newUser5
    before((done) => {
        chai.request(Config.server).post('omr/api/v1/users') 
            .set('Accept',  'application/json')
            .set('token', Config.orgAdminToken2)
            .send(newUser5)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("success");
                done();
            });
    });

    // grant service admin permission of org2Service2 to newUser5
    before((done) => {
        chai.request(Config.server).put('omr/api/v1/users/' + newUser5.id + '/permissions/services/' + org2Service2.id) 
            .set('Accept',  'application/json')
            .set('token', Config.orgAdminToken2)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("success");
                done();
            });
    });

    // login as service admin
    let o2s2AdminToken = '';
    before((done) => {
        chai.request(Config.server).get('common/api/v1/login')
            .set('Accept', 'application/json')
            .set('user-id', newUser5.id)
            .set('password', newUser5.secret)
            .set('login-org', newUser5.ca_org)
            .set('login-channel', Config.channel)
            .set('signature', "")
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("token");
                expect(res.body.token).to.not.equal("");
                o2s2AdminToken = res.body.token;
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

    it('Should successfully create contract by data owner', function (done) {
        const bodyRequest = {
            "owner_org_id": Config.org1.id,
            "owner_service_id": org1Service1.id,
            "requester_org_id": Config.org2.id,
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

    it('Should NOT successfully change contract terms by service admin of another service', function (done) {
        const newTermsByOwner = {
            "contract_id": contractId,
            "contract_terms": {"contract terms":"updated terms by owner"}
          }
        chai.request(Config.server).post('omr/api/v1/contracts/' + contractId + '/changeTerms')
            .set('Accept', 'application/json')
            .set('token', o1s2AdminToken)
            .send(newTermsByOwner)
            .end(function(err, res){          
                expect(err).to.be.null;
                expect(res.status).to.equal(500);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("error");
                done();
            });
    });

    it('Should NOT successfully change contract terms by service admin of another service', function (done) {
        const newTermsByOwner = {
            "contract_id": contractId,
            "contract_terms": {"contract terms":"updated terms by owner"}
          }
        chai.request(Config.server).post('omr/api/v1/contracts/' + contractId + '/changeTerms')
            .set('Accept', 'application/json')
            .set('token', o2s2AdminToken)
            .send(newTermsByOwner)
            .end(function(err, res){          
                expect(err).to.be.null;
                expect(res.status).to.equal(500);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("error");
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

    it('Should NOT successfully sign contract by service admin of another service', function (done) {
        let bodyRequest = {
            "contract_id": contractId,
            "contract_terms": {},
            "signed_by": org2Service2.id
          }
        chai.request(Config.server).post('omr/api/v1/contracts/' + contractId + '/sign')
            .set('Accept', 'application/json')
            .set('token', o2s2AdminToken)
            .send(bodyRequest)
            .end(function(err, res){
                expect(err).to.be.null;
                expect(res.status).to.equal(500);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("error");
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

    // download 1
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

    it('Should NOT successfully get contract data by service admin of another service', function (done) {
        chai.request(Config.server).get('omr/api/v1/contracts/' + contractId)
            .set('Accept', 'application/json')
            .set('token', o2s2AdminToken)
            .end(function(err, res){
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body.contract_id).to.equal("");
                expect(res.body.owner_org_id).to.equal("");
                expect(res.body.owner_service_id).to.equal("");
                expect(res.body.requester_org_id).to.equal("");
                expect(res.body.contract_terms).to.be.null;
                expect(res.body.contract_details).to.be.null;
                done();
            });
    });

    it('Should NOT successfully get contract data by service admin of another service', function (done) {
        chai.request(Config.server).get('omr/api/v1/contracts/' + contractId)
            .set('Accept', 'application/json')
            .set('token', o1s2AdminToken)
            .end(function(err, res){
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body.contract_id).to.equal("");
                expect(res.body.owner_org_id).to.equal("");
                expect(res.body.owner_service_id).to.equal("");
                expect(res.body.requester_org_id).to.equal("");
                expect(res.body.contract_terms).to.be.null;
                expect(res.body.contract_details).to.be.null;
                done();
            });
    });

    it('Should NOT successfully get contracts by service admin of another service filtered by status of contract', function (done) {
        chai.request(Config.server).get('omr/api/v1/contracts/owner/' + org1Service2.id + '/status/*')
            .set('Accept', 'application/json')
            .set('token', o1s2AdminToken)
            .end(function(err, res){
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.be.an('array');
                expect(res.body).to.have.lengthOf(0);
                done();
            });
    });

    it('Should NOT successfully get contracts by service admin of another service filtered by status of contract', function (done) {
        chai.request(Config.server).get('omr/api/v1/contracts/requester/' + org2Service2.id + '/status/*')
            .set('Accept', 'application/json')
            .set('token', o2s2AdminToken)
            .end(function(err, res){
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.be.an('array');
                expect(res.body).to.have.lengthOf(0);
                done();
            });
    });

    it('Should NOT successfully terminate contract by service admin of another service', function (done) {
        let bodyRequest = {
            "contract_id": contractId,
            "terminated_by": org1Service2.id
          }
        chai.request(Config.server).post('omr/api/v1/contracts/' + contractId + '/terminate')
            .set('Accept', 'application/json')
            .set('token', o1s2AdminToken)
            .send(bodyRequest)
            .end(function(err, res){
                expect(err).to.be.null;
                expect(res.status).to.equal(500);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("error");
                done();
            });
    });

    it('Should NOT successfully terminate contract by service admin of another service', function (done) {
        let bodyRequest = {
            "contract_id": contractId,
            "terminated_by": org2Service2.id
          }
        chai.request(Config.server).post('omr/api/v1/contracts/' + contractId + '/terminate')
            .set('Accept', 'application/json')
            .set('token', o2s2AdminToken)
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