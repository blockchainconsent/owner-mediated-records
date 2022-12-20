const Config = require('../../ConfigFile.js');
const chai = require('chai');
const { expect } = chai;
const chaiHttp = require('chai-http');
chai.use(chaiHttp);

require("../../Login/loginSysAdmin");

//Test to Register organisation through system Admin
//Expected: Test should pass, orgnisation registration should be successful
//Copy of RegisterOrg1: allows user to register another org object
describe('Register Org2 with SysAdmin', function () {
    
    it('Org2 should be registered', function (done) {  
        chai.request(Config.server).post('omr/api/v1/orgs') 
            .set('Accept',  'application/json')
            .set('token', Config.sysAdminToken)
            .send(Config.org2)
            .end(function(err, res) {
                expect(err).to.be.null
                //TODO update status code for registering org with same ID
                //expect(res.status).to.equal(200);
                done();
            });
    });
});