var Config = require('../../ConfigFile.js');
const chai = require('chai');
const { expect } = chai;
const chaiHttp = require('chai-http')
chai.use(chaiHttp);


require("../RegisterUsers");
require("../../Login/User1Login");

//Test to Register datatype as a User1
//should fail
describe('Register Datatype as Role User', function () {
    let newDatatype = {
        "id": Math.random().toString().slice(2,9),
        "description": "a random description"
    };

     it('Should fail to register datatype', function (done) {  
        chai.request(Config.server).post('omr/api/v1/datatypes') 
            .set('Accept',  'application/json')
            .set('token', Config.user1Token)
            .send(newDatatype)
            .end(function(err, res) {
                expect(err).to.be.null
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.equal("registerDatatype error");
                expect(res.status).to.equal(500);
                done();
            });
        
    });
    
});