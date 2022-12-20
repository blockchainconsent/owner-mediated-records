var Config = require('../ConfigFile.js');
const chai = require('chai');
const { expect } = chai;
const chaiHttp = require('chai-http')
chai.use(chaiHttp);

require("../Organizations/OrgSetup/RegisterOrg1")
require("../Login/loginSysAdmin")
require("../Login/OrgLogin/Org1Login")

describe('Register User as Sys Admin', function () {
    it('User1: should return a 200 test response', function (done) {  
        chai.request(Config.server).post('omr/api/v1/users') 
            .set('Accept',  'application/json')
            .set('token', Config.sysAdminToken)
            .send(Config.user1)
            .end(function(err) {
                expect(err).to.be.null;
                done();
            });
        
    });
});

describe('Register Org User as Org Admin', function () {
    it('User2: should return a 200 test response', function (done) {  
        chai.request(Config.server).post('omr/api/v1/users') 
            .set('Accept',  'application/json')
            .set('token', Config.orgAdminToken1)
            .send(Config.user2)
            .end(function(err) {
                expect(err).to.be.null;
                done();
            });
    });
});
