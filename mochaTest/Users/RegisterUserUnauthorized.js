var Config = require('../ConfigFile.js');
const chai = require('chai');
const { expect } = chai;
const chaiHttp = require('chai-http')
chai.use(chaiHttp);

require("../Login/loginSysAdmin")

//Test to Register New User in Org1 as Sys Admin
//Should fail to register user
describe('Register User in Org1 as SysAdmin', function () {
    var newID = Math.random().toString().slice(2,9)

    var newUser = {
        "id": newID,
        "secret": "newuser",
        "name": "New User: " + newID,
        "role": "user",
        "org" : Config.org1.id,
        "email":  "useremail@example.com",
        "ca_org": Config.caOrg,
        "data": {}
    };
   
    it('New user should return a 500 test response', function (done) {  
        chai.request(Config.server).post('omr/api/v1/users') 
            .set('Accept',  'application/json')
            .set('token', Config.sysAdminToken)
            .send(newUser)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(500);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.equal("User is registered to CA, but failed to register user in Blockchain:Failed to register user");
                done();
            });
        
    });
});