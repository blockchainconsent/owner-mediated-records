const Config = require('../../ConfigFile.js');
const chai = require('chai');
const { expect } = chai;
const chaiHttp = require('chai-http');
chai.use(chaiHttp);
const { v4: uuidv4 } = require('uuid');

require("../../Login/loginSysAdmin");
exports.orgDetails;

//Test to Register organisation through system Admin
//Generates Random orgID:
//Expected: Test should pass, orgnisation registration should be successful
//Copy of RegisterOrg1: allows user to register another org object
describe('Register Random New Org with SysAdmin', function () {
    let newOrg = {
        "id": uuidv4(),
        "secret": "random",
        "name": "Random Org",
        "ca_org": Config.caOrg,
        "email": "random@email.com",
        "status": "active"
    };
    
    exports.orgDetails = newOrg;
    
        it('Org should be registered', function (done) {  
        chai.request(Config.server).post('omr/api/v1/orgs') 
            .set('Accept',  'application/json')
            .set('token', Config.sysAdminToken)
            .send(newOrg)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("success");
                done();
            });
    });
});