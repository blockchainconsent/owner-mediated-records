var Config = require('../ConfigFile.js');
const chai = require('chai');
const { expect } = chai;
const chaiHttp = require('chai-http')
chai.use(chaiHttp);

require("./RegisterServiceOrg1Admin.js")
require("../Login/OrgLogin/Org1Login");

describe('Remove Datatype From Unregistered Service', function () {

    var unregService = "unregistered";

    it('Should return a 400 test response', function (done) { 
        chai.request(Config.server).delete('omr/api/v1/services/' + unregService + 
                '/removeDatatype/' + Config.datatype1.id) 
            .set('Accept',  'application/json')
            .set('token', Config.orgAdminToken1)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(400);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.equal("Invalid id: Service with id " + unregService + " does not exist");
                done();
            });
        
    });
});

describe('Remove Unregistered Datatype from Service', function () {

    it('Should return a 400 test response', function (done) { 
        chai.request(Config.server).delete('omr/api/v1/services/' + Config.service1.id + '/removeDatatype/' + "unregistered") 
            .set('Accept',  'application/json')
            .set('token', Config.orgAdminToken1)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(404);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.equal("Datatype not found");
                done();
            });
        
    });
});

describe('Remove Datatype not attached to Service from Service', function () {

    it('Should return a 400 test response', function (done) { 
        chai.request(Config.server).delete('omr/api/v1/services/' + Config.service1.id + '/removeDatatype/' + Config.datatype3.id) 
            .set('Accept',  'application/json')
            .set('token', Config.orgAdminToken1)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(500);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("error");
                done();
            });
        
    });
});


