const Config = require('../ConfigFile.js');
const chai = require('chai');
const { expect } = chai;
const chaiHttp = require('chai-http')
chai.use(chaiHttp);
const idGenerator = require('../Utils/helper');

require("../Login/loginSysAdmin")

//Test to Register New User as SysAdmin
describe('Register New User as SysAdmin', function () {
    //generate new uuid with only alphanumberic characters
    var newUser = {
        "id": idGenerator(),
        "secret": "newuser",
        "name": "newUser",
        "role": "user",
        "email":  "useremail@example.com",
        "ca_org": Config.caOrg,
        "data": {}
    };
   
    it('New user should return a 200 test response', function (done) {  
        chai.request(Config.server).post('omr/api/v1/users') 
            .set('Accept',  'application/json')
            .set('token', Config.sysAdminToken)
            .send(newUser)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("success");
                done();
            });
        
    });
});