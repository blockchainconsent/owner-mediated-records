var Config = require('../ConfigFile.js');
const chai = require('chai');
const { expect } = chai;
const chaiHttp = require('chai-http')
chai.use(chaiHttp);
chai.use(require('chai-like'));
const chaiThings = require('chai-things')
chai.use(chaiThings)

require("../Organizations/OrgSetup/RegisterOrg2")
require("./RegisterServiceOrg1Admin")
require("../Login/OrgLogin/Org1Login");
require("../Login/OrgLogin/Org2Login");
require("../Login/loginSysAdmin");

describe('Get Service1 by ID as Org1 Admin', function () {
    it('Should return a 200 response and show private data', function (done) {  
        chai.request(Config.server).get('omr/api/v1/services/' + Config.service1.id) 
            .set('Accept',  'application/json')
            .set('token', Config.orgAdminToken1)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body.service_id).to.equal(Config.service1.id);
                expect(res.body.org_id).to.equal(Config.service1.org_id);
                expect(res.body.summary).to.equal(Config.service1.summary);
                expect(JSON.stringify(res.body.terms)).to.equal(JSON.stringify(Config.service1.terms));
                expect(res.body.email).to.equal(Config.service1.email)
                expect(JSON.stringify(res.body.solution_private_data)).to.equal(JSON.stringify(Config.service1.solution_private_data));
                done();
            });
        
    });

});

describe('Get Service1 by ID as Org2 Admin', function () {
    it('Should return a 200 response and hide private data', function (done) {  
        chai.request(Config.server).get('omr/api/v1/services/' + Config.service1.id) 
            .set('Accept',  'application/json')
            .set('token', Config.orgAdminToken2)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body.service_id).to.equal(Config.service1.id);
                expect(res.body.org_id).to.equal(Config.service1.org_id);
                expect(res.body.summary).to.equal(Config.service1.summary);
                expect(JSON.stringify(res.body.terms)).to.equal(JSON.stringify(Config.service1.terms));
                expect(res.body.email).to.equal("")
                expect(res.body.solution_private_data).to.be.null;
                done();
            });
        
    });

});

describe('Get Service1 by ID as Sys Admin', function () {
    it('Should return a 200 response and hide private data', function (done) {  
        chai.request(Config.server).get('omr/api/v1/services/' + Config.service1.id) 
            .set('Accept',  'application/json')
            .set('token', Config.sysAdminToken)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body.service_id).to.equal(Config.service1.id);
                expect(res.body.org_id).to.equal(Config.service1.org_id);
                expect(res.body.summary).to.equal(Config.service1.summary);
                expect(JSON.stringify(res.body.terms)).to.equal(JSON.stringify(Config.service1.terms));
                expect(res.body.email).to.equal("")
                expect(res.body.solution_private_data).to.be.null;
                done();
            });
        
    });

});