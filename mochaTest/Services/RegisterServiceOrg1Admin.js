var Config = require('../ConfigFile.js');
const chai = require('chai');
const { expect } = chai;
const chaiHttp = require('chai-http')
chai.use(chaiHttp);

require("../Organizations/OrgSetup/RegisterOrg1")
require("../DataTypes/RegisterDatatypesSysAdmin")
require("../Login/OrgLogin/Org1Login");

// Test Register Services as OrgAdmin
// Registers service1 and service2 from Config to Org1
describe('Register Service with Org1 Admin', function () {   
    it('service 1 should be registered', function (done) {  
        Config.service1.datatypes[0].datatype_id = Config.datatype1.id;
        
        chai.request(Config.server).post('omr/api/v1/services') 
            .set('Accept',  'application/json')
            .set('token', Config.orgAdminToken1)
            .send(Config.service1)
            .end(function(err, res) {
                expect(err).to.be.null;
                done();
            });
    });

    it('service 2 should be registered', function (done) {  
        Config.service2.datatypes[0].datatype_id = Config.datatype1.id;
        Config.service2.datatypes[1].datatype_id = Config.datatype2.id;
        
        chai.request(Config.server).post('omr/api/v1/services') 
            .set('Accept',  'application/json')
            .set('token', Config.orgAdminToken1)
            .send(Config.service2)
            .end(function(err, res) {
                expect(err).to.be.null;
                done();
            });
    });
});