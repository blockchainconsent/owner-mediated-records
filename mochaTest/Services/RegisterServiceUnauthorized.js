var Config = require('../ConfigFile.js');
const chai = require('chai');
const { expect } = chai;
const chaiHttp = require('chai-http')
chai.use(chaiHttp);

require("../Organizations/OrgSetup/RegisterOrg1")
require("../Organizations/OrgSetup/RegisterOrg2")
require("../DataTypes/RegisterDatatypesSysAdmin")
require("../Login/loginSysAdmin");
require("../Login/OrgLogin/Org2Login");


var newService = {
  "id": "newService",
  "name": "newService",
  "secret": "service1pass",
  "ca_org": Config.caOrg,
  "email": "service1email@example.com",
  "org_id": Config.org1.id,
  "summary": "New Service under org 1. Has one datatype",
  "terms": {},
  "payment_required": "yes",
  "datatypes": [
      {
          "datatype_id": Config.datatype1.id,
          "access":[
              "write",
              "read"
          ]
      }
  ]
};

//Test to Register Service through system Admin
describe('Register Service with SysAdmin', function () {

    it('Should Return a 500 status code', function (done) {  
        chai.request(Config.server).post('omr/api/v1/services') 
            .set('Accept',  'application/json')
            .set('token', Config.sysAdminToken)
            .send(newService)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(500);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.equal("Service is registered to CA, but failed to register service in Blockchain:Failed to register service")
                done();
            });
        
    });

    it('Service Should not be registered', function(done) {
        chai.request(Config.server).get('omr/api/v1/services/' + newService.id) 
            .set('Accept',  'application/json')
            .set('token', Config.sysAdminToken)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.deep.equal({});
                done();
            });
    })
});

//Test to Register Service through system Admin
//Should fail
describe('Register Service to Org1 as Org2 Admin', function () {

    it('Should Return a 500 status code', function (done) {  
        chai.request(Config.server).post('omr/api/v1/services') 
            .set('Accept',  'application/json')
            .set('token', Config.orgAdminToken2)
            .send(newService)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(500);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.equal("Service is registered to CA, but failed to register service in Blockchain:Failed to register service")
                done();
            });
        
    });

    it('Service Should not be registered', function(done) {
        chai.request(Config.server).get('omr/api/v1/services/' + newService.id) 
            .set('Accept',  'application/json')
            .set('token', Config.sysAdminToken)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.deep.equal({});
                done();
            });
    })
});