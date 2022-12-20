var Config = require('../ConfigFile.js');
const chai = require('chai');
const { expect } = chai;
const chaiHttp = require('chai-http')
chai.use(chaiHttp);

require("../Organizations/OrgSetup/RegisterOrg1");
require("../Login/OrgLogin/Org1Login");

//Test to Register datatype through org Admin
//Generates Random datatype ID:
describe('Register Datatype as Org Admin', function () {
    let newDatatype = {
        "id": Math.random().toString().slice(2,9),
        "description": "a random description"
    };

     it('Datatype should be registered & give successful message', function (done) {  
        chai.request(Config.server).post('omr/api/v1/datatypes') 
            .set('Accept',  'application/json')
            .set('token', Config.orgAdminToken1)
            .send(newDatatype)
            .end(function(err, res) {
                expect(err).to.be.null
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("success");
                done();
            });
        
    });
    
});