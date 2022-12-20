var Config = require('../../ConfigFile.js');
const chai = require('chai');
const { expect } = chai;
const chaiHttp = require('chai-http')
chai.use(chaiHttp);

require("../RegisterUsers");
require("../../Login/User1Login");

//Test to Register organisation as User1
//Generates Random orgID:
//Expected: Test should fail, user should not be able to register org
describe('Register New Org as Role User', function () {
    let newOrg = {
        "id": "org" + Math.random().toString().slice(2,9),
        "secret": "random",
        "name": "Random Org",
        "ca_org": Config.caOrg,
        "email": "random@email.com",
        "status": "active"
    };
    
    
        it('Should fail to register org and return 500 status code', function (done) {  
        chai.request(Config.server).post('omr/api/v1/orgs') 
            .set('Accept',  'application/json')
            .set('token', Config.user1Token)
            .send(newOrg)
            .end(function(err, res) {
                expect(err).to.be.null
                expect(res.status).to.equal(500);
                expect(res.body).to.not.be.null;
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.equal("Org is registered to CA, but failed to register org in Blockchain:Failed to register org")
                done();
            });
        
    });
    
});