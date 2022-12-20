var Config = require('../ConfigFile.js');
const chai = require('chai');
const { expect } = chai;
const chaiHttp = require('chai-http')
chai.use(chaiHttp);

const { v4: uuidv4 } = require('uuid');

require("./RegisterServiceOrg1Admin.js")
require("../Login/OrgLogin/Org1Login");

//Removes Datatype1 from Service 2 Org1
describe('Remove Datatype From Service as Org1 Admin', function () {
    const datatypeID = uuidv4();
    const datatypeToAdd = {
        "datatype_id": datatypeID,
        "access": [
          "read",
          "write"
        ]
    };
    const datatypeToRegister = {
        "id": datatypeID,
        "description": "description"
    };

    before((done) => {   
        // register datatype
        chai.request(Config.server).post('omr/api/v1/datatypes') 
            .set('Accept',  'application/json')
            .set('token', Config.sysAdminToken)
            .send(datatypeToRegister)
            .end(function(err, res) {
                expect(err).to.be.null
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("success");
                done();
            });
    });

    before((done) => {
        // add datatype to service
        chai.request(Config.server).post('omr/api/v1/services/' + Config.service2.id + '/addDatatype/' + datatypeID) 
            .set('Accept',  'application/json')
            .set('token', Config.orgAdminToken1)
            .send(datatypeToAdd)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("success");
                done();
            }); 
    });
    
    it('Should return a 200 test response', function (done) { 
        chai.request(Config.server)
            .delete('omr/api/v1/services/' + Config.service2.id + '/removeDatatype/' + datatypeID) 
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
