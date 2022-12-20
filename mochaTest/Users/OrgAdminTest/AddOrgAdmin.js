var Config = require('../../ConfigFile.js');
const chai = require('chai');
const { expect } = chai;
const chaiHttp = require('chai-http')
chai.use(chaiHttp);

require("../../Organizations/OrgSetup/RegisterOrg1")
require("../../Login/OrgLogin/Org1Login")
require("../RegisterUsers");

// Test to Add OrgAdmin Permission
describe('Add Org Admin Permission to User2', function () {
   
    it('Should return a 200 test response', function (done) {  
        chai.request(Config.server).put('omr/api/v1/users/' + Config.user2.id + '/permissions/admin/' + Config.org1.id) 
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
});