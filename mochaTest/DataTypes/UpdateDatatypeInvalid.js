var Config = require('../ConfigFile.js');
const chai = require('chai');
const { expect } = chai;
const chaiHttp = require('chai-http')
chai.use(chaiHttp);

require("../ConfigFile")
require("../DataTypes/RegisterDatatypesSysAdmin")
require("../Login/loginSysAdmin");

describe('Update Datatype with Missing ID', function() {
    var noID = {
        "description" : "new description"
    }
    it('Should return a 400 test response and appropriate message', function(done) {
        chai.request(Config.server).put('omr/api/v1/datatypes/' + Config.datatype1.id) 
            .set('Accept',  'application/json')
            .set('token', Config.sysAdminToken)
            .send(noID)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(400);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.equal("Invalid datatype:id missing");
                done();
            });
    });

});

describe('Update Datatype with Missing Description', function() {
    let noId;
    
    before(function() {
        noID = {
            "id" : Config.datatype1.id
        };
    });

    it('Should return a 400 test response and appropriate message', function(done) {
        chai.request(Config.server).put('omr/api/v1/datatypes/' + Config.datatype1.id) 
            .set('Accept',  'application/json')
            .set('token', Config.sysAdminToken)
            .send(noID)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(400);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.equal("Invalid datatype:description missing");
                done();
            });
    });

});

describe('Update Datatype with Path diff from Body', function() {
    let newData;
    
    before(function() {
        newData = {
            "id" : Config.datatype1.id,
            "description" : "random"
        };
    });

    it('Should return a 400 test response and appropriate message', function(done) {
        chai.request(Config.server).put('omr/api/v1/datatypes/' + Config.datatype2.id) 
            .set('Accept',  'application/json')
            .set('token', Config.sysAdminToken)
            .send(newData)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(400);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.equal("Invalid datatype: id mismatch");
                done();
            });
    });

});

describe('Update Unregistered Datatype', function() {
    let randomID, newData;
    
    before(function() {
        randomID = "notRegistered";
        newData = {
            "id" : randomID,
            "description" : "random"
        };
    });
    

    it('Should return a 400 test response and appropriate message', function(done) {
        chai.request(Config.server).put('omr/api/v1/datatypes/' + randomID) 
            .set('Accept',  'application/json')
            .set('token', Config.sysAdminToken)
            .send(newData)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(400);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.equal("Existing datatype with this id not found");
                done();
            });
    });

});
