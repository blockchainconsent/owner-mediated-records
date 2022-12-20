var Config = require('../ConfigFile.js');
const chai = require('chai');
const { expect } = chai;
const chaiHttp = require('chai-http')
chai.use(chaiHttp);

require("./RegisterServiceOrg2Admin.js")
require("../Login/OrgLogin/Org2Login");


describe('Add Datatype where path does not match data', function () {

    var invalidRequest = {
        "datatype_id": "invalid",
        "access": [
          "read",
          "write"
        ]
    }

    it('Should return a 400 test response', function (done) { 
        chai.request(Config.server).post('omr/api/v1/services/' + Config.service3.id + '/addDatatype/' + 'path') 
            .set('Accept',  'application/json')
            .set('token', Config.orgAdminToken2)
            .send(invalidRequest)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(400);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.equal("Invalid request: datatype_id mismatch");
                done();
            });
        
    });
});

describe('Add Datatype with Missing Access', function () {

    var invalidRequest = {
        "datatype_id" : "invalid"
    }

    it('Should return a 400 test response', function (done) { 
        chai.request(Config.server).post('omr/api/v1/services/' + Config.service3.id + '/addDatatype/' + invalidRequest.datatype_id) 
            .set('Accept',  'application/json')
            .set('token', Config.orgAdminToken2)
            .send(invalidRequest)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(400);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.equal("Invalid data:access missing");
                done();
            });
        
    });
});

describe('Add Datatype with Empty Access Body', function () {

    var invalidRequest = {
        "datatype_id" : "invalid",
        "access" : []
    }

    it('Should return a 400 test response', function (done) { 
        chai.request(Config.server).post('omr/api/v1/services/' + Config.service3.id + '/addDatatype/' + invalidRequest.datatype_id) 
            .set('Accept',  'application/json')
            .set('token', Config.orgAdminToken2)
            .send(invalidRequest)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(400);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.equal("Invalid data:must include at least one access option");
                done();
            });
        
    });
});

describe('Add Datatype to Unregistered Service', function () {
    var invalidRequest, unregService;

    before(function() {
        invalidRequest = {
            "datatype_id" : Config.datatype1,
            "access" : [
                "read",
                "write"
            ]
        };
        unregService = "unregistered";
    });

    it('Should return a 400 test response', function (done) { 
        chai.request(Config.server).post('omr/api/v1/services/' + unregService+ '/addDatatype/' + invalidRequest.datatype_id) 
            .set('Accept',  'application/json')
            .set('token', Config.orgAdminToken2)
            .send(invalidRequest)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(400);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.equal("Invalid id: Service with id " + unregService + " does not exist");
                done();
            });
        
    });
});

describe('Add Unregistered Datatype to Service', function () {

    var invalidRequest = {
        "datatype_id" : "unregistered",
        "access" : [
            "read",
            "write"
        ]
    }

    it('Should return a 404 test response', function (done) { 
        chai.request(Config.server).post('omr/api/v1/services/' + Config.service3.id + '/addDatatype/' + invalidRequest.datatype_id) 
            .set('Accept',  'application/json')
            .set('token', Config.orgAdminToken2)
            .send(invalidRequest)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(404);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.equal("Datatype not found");
                done();
            });
        
    });
});
