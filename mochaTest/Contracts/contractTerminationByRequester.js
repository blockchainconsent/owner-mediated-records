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

describe('Checking of ability of termination of Contract by Data Requester', function() {

    let contractId1;
    let contractId2;
    let contractId3;
    let contractId4;
    let contractId5;
    let contractId6;
    let contractId7;

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
                contractId1 = res.body.contract_id;
                done();
            });
    });

    it('Should successfully terminate contract by data requester', function (done) {
        let bodyRequest = {
            "contract_id": contractId1,
            "terminated_by": org2Service1.id
          }
        chai.request(Config.server).post('omr/api/v1/contracts/' + contractId1 + '/terminate')
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
                contractId2 = res.body.contract_id;
                done();
            });
    });
    
    it('Should successfully change contract terms by data requester', function (done) {
        const newTermsByRequester = {
            "contract_id": contractId2,
            "contract_terms": {"contract terms":"updated terms by requester"}
          }
        chai.request(Config.server).post('omr/api/v1/contracts/' + contractId2 + '/changeTerms')
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

    it('Should successfully terminate contract by data requester', function (done) {
        let bodyRequest = {
            "contract_id": contractId2,
            "terminated_by": org2Service1.id
          }
        chai.request(Config.server).post('omr/api/v1/contracts/' + contractId2 + '/terminate')
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
                contractId3 = res.body.contract_id;
                done();
            });
    });

    it('Should successfully change contract terms by data owner', function (done) {
        const newTermsByOwner = {
            "contract_id": contractId3,
            "contract_terms": {"contract terms":"updated terms by owner"}
          }
        chai.request(Config.server).post('omr/api/v1/contracts/' + contractId3 + '/changeTerms')
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
            "contract_id": contractId3,
            "contract_terms": {},
            "signed_by": org2Service1.id
          }
        chai.request(Config.server).post('omr/api/v1/contracts/' + contractId3 + '/sign')
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

    it('Should successfully terminate contract by data requester', function (done) {
        let bodyRequest = {
            "contract_id": contractId3,
            "terminated_by": org2Service1.id
          }
        chai.request(Config.server).post('omr/api/v1/contracts/' + contractId3 + '/terminate')
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
                contractId4 = res.body.contract_id;
                done();
            });
    });

    it('Should successfully sign contract by requester', function (done) {
        let bodyRequest = {
            "contract_id": contractId4,
            "contract_terms": {},
            "signed_by": org2Service1.id
          }
        chai.request(Config.server).post('omr/api/v1/contracts/' + contractId4 + '/sign')
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
            "contract_id": contractId4,
            "contract_terms": {}
          }
        chai.request(Config.server).post('omr/api/v1/contracts/' + contractId4 + '/payment')
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

    it('Should successfully terminate contract by data requester', function (done) {
        let bodyRequest = {
            "contract_id": contractId4,
            "terminated_by": org2Service1.id
          }
        chai.request(Config.server).post('omr/api/v1/contracts/' + contractId4 + '/terminate')
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
                contractId5 = res.body.contract_id;
                done();
            });
    });

    it('Should successfully sign contract by requester', function (done) {
        let bodyRequest = {
            "contract_id": contractId5,
            "contract_terms": {},
            "signed_by": org2Service1.id
          }
        chai.request(Config.server).post('omr/api/v1/contracts/' + contractId5 + '/sign')
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
            "contract_id": contractId5,
            "contract_terms": {}
          }
        chai.request(Config.server).post('omr/api/v1/contracts/' + contractId5 + '/payment')
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
            "contract_id": contractId5,
            "contract_terms": {}
          }
        chai.request(Config.server).post('omr/api/v1/contracts/' + contractId5 + '/verify')
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

    it('Should successfully terminate contract by data requester', function (done) {
        let bodyRequest = {
            "contract_id": contractId5,
            "terminated_by": org2Service1.id
          }
        chai.request(Config.server).post('omr/api/v1/contracts/' + contractId5 + '/terminate')
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
                contractId6 = res.body.contract_id;
                done();
            });
    });

    it('Should successfully sign contract by requester', function (done) {
        let bodyRequest = {
            "contract_id": contractId6,
            "contract_terms": {},
            "signed_by": org2Service1.id
          }
        chai.request(Config.server).post('omr/api/v1/contracts/' + contractId6 + '/sign')
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
            "contract_id": contractId6,
            "contract_terms": {}
          }
        chai.request(Config.server).post('omr/api/v1/contracts/' + contractId6 + '/payment')
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
            "contract_id": contractId6,
            "contract_terms": {}
          }
        chai.request(Config.server).post('omr/api/v1/contracts/' + contractId6 + '/verify')
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
        chai.request(Config.server).post('omr/api/v1/contracts/' + contractId6 +
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

    it('Should successfully terminate contract by data requester', function (done) {
        let bodyRequest = {
            "contract_id": contractId6,
            "terminated_by": org2Service1.id
          }
        chai.request(Config.server).post('omr/api/v1/contracts/' + contractId6 + '/terminate')
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
                contractId7 = res.body.contract_id;
                done();
            });
    });

    it('Should successfully sign contract by requester', function (done) {
        let bodyRequest = {
            "contract_id": contractId7,
            "contract_terms": {},
            "signed_by": org2Service1.id
          }
        chai.request(Config.server).post('omr/api/v1/contracts/' + contractId7 + '/sign')
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
            "contract_id": contractId7,
            "contract_terms": {}
          }
        chai.request(Config.server).post('omr/api/v1/contracts/' + contractId7 + '/payment')
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
            "contract_id": contractId7,
            "contract_terms": {}
          }
        chai.request(Config.server).post('omr/api/v1/contracts/' + contractId7 + '/verify')
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
        chai.request(Config.server).post('omr/api/v1/contracts/' + contractId7 +
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
        chai.request(Config.server).get('omr/api/v1/contracts/' + contractId7 + 
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

    it('Should successfully terminate contract by data requester', function (done) {
        let bodyRequest = {
            "contract_id": contractId7,
            "terminated_by": org2Service1.id
          }
        chai.request(Config.server).post('omr/api/v1/contracts/' + contractId7 + '/terminate')
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

    it('Should successfully get contracts by data owner filtered by status of contract', function (done) {
        chai.request(Config.server).get('omr/api/v1/contracts/owner/' + org1Service1.id + '/status/*')
            .set('Accept', 'application/json')
            .set('token', defaultOrg1Service1AdminToken)
            .end(function(err, res){
                expect(err).to.be.null;
                expect(res.body).to.be.an('array');
                expect(res.body).to.have.lengthOf(7);
                expect(res.body[0].state).to.equal("terminated");
                expect(res.body[1].state).to.equal("terminated");
                expect(res.body[2].state).to.equal("terminated");
                expect(res.body[3].state).to.equal("terminated");
                expect(res.body[4].state).to.equal("terminated");
                expect(res.body[5].state).to.equal("terminated");
                expect(res.body[6].state).to.equal("terminated");
                done();
            });
    });
});