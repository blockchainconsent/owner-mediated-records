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

describe('Download owner data as requester with given consent', function() {

    let contractId;
    let consentTokenForRequester;
    let consentTokenForOwner;

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

    // upload owner data to o1d1 by o1s1 admin
    before((done) => {
        const bodyRequest = {
            "test owner data 2": "2 owner data o1s1d1"
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

    // upload owner data to o1d1 by o1s1 admin
    before((done) => {
        const bodyRequest = {
            "test owner data 3": "3 owner data o1s1d1"
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

    // upload owner data to o1d1 by o1s1 admin
    before((done) => {
        const bodyRequest = {
            "test owner data 4": "4 owner data o1s1d1"
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

    it('Should successfully create consent for org2Service1', function (done) {
        const bodyRequest = {
            "service_id": org1Service1.id,
            "datatype_id": org1Service1.datatypes[0].datatype_id,
            "target_id": org2Service1.id,
            "option": [
                "write",
                "read"
            ],
            "expiration": 0
        }
        chai.request(Config.server).post('omr/api/v1/consents/ownerData')
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

    // get validate consent token as requester
    it('Should successfully get validate consent token as requester', function (done) {
        chai.request(Config.server).get('omr/api/v1/services/' + org2Service1.id + 
        '/owner_service/' + org1Service1.id + '/datatype/' + org1Service1.datatypes[0].datatype_id + 
         '/validation/write/ownerdata')
            .set('Accept', 'application/json')
            .set('token', defaultOrg2Service1AdminToken)
            .end(function(err, res){
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("token");
                consentTokenForRequester = res.body.token;
                done();
            });
    });

    // get validate consent token as owner
    it('Should successfully get validate consent token as owner', function (done) {
        chai.request(Config.server).get('omr/api/v1/services/' + org2Service1.id + 
        '/owner_service/' + org1Service1.id + '/datatype/' + org1Service1.datatypes[0].datatype_id + 
         '/validation/write/ownerdata')
            .set('Accept', 'application/json')
            .set('token', defaultOrg1Service1AdminToken)
            .end(function(err, res){
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("token");
                consentTokenForOwner = res.body.token;
                done();
            });
    });


    // download owner data by requester with validate consent token
    it('Should successfully download owner data by requester with validate consent token', function (done) {
        chai.request(Config.server).get('omr/api/v1/ownerData/download/' + consentTokenForRequester)
            .set('Accept', 'application/json')
            .set('token', defaultOrg2Service1AdminToken)
            .end(function(err, res){
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.be.an('array');
                expect(res.body).to.have.lengthOf(4);
                expect(JSON.stringify(res.body[0].data)).to.equal(JSON.stringify({
                    "test owner data":"owner data o1s1d1"
                }));
                expect(JSON.stringify(res.body[1].data)).to.equal(JSON.stringify({
                    "test owner data 2":"2 owner data o1s1d1"
                }));
                expect(JSON.stringify(res.body[2].data)).to.equal(JSON.stringify({
                    "test owner data 3":"3 owner data o1s1d1"
                }));
                expect(JSON.stringify(res.body[3].data)).to.equal(JSON.stringify({
                    "test owner data 4":"4 owner data o1s1d1"
                }));
                done();
            });
    });

    // download owner data by owner with validate consent token
    it('Should successfully download owner data by owner with validate consent token', function (done) {
        chai.request(Config.server).get('omr/api/v1/ownerData/download/' + consentTokenForOwner)
            .set('Accept', 'application/json')
            .set('token', defaultOrg1Service1AdminToken)
            .end(function(err, res){
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.be.an('array');
                expect(res.body).to.have.lengthOf(4);
                expect(JSON.stringify(res.body[0].data)).to.equal(JSON.stringify({
                    "test owner data":"owner data o1s1d1"
                }));
                expect(JSON.stringify(res.body[1].data)).to.equal(JSON.stringify({
                    "test owner data 2":"2 owner data o1s1d1"
                }));
                expect(JSON.stringify(res.body[2].data)).to.equal(JSON.stringify({
                    "test owner data 3":"3 owner data o1s1d1"
                }));
                expect(JSON.stringify(res.body[3].data)).to.equal(JSON.stringify({
                    "test owner data 4":"4 owner data o1s1d1"
                }));
                done();
            });
    });
});