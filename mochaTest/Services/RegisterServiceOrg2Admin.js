var Config = require('../ConfigFile.js');
const chai = require('chai');
const { expect } = chai;
const chaiHttp = require('chai-http')
chai.use(chaiHttp);

require("../Organizations/OrgSetup/RegisterOrg2")
require("../DataTypes/RegisterDatatypesSysAdmin")
require("../Login/OrgLogin/Org2Login");

// Test to Register Services through OrgAdmin
// Registers service3 from Config to Org1
describe('Register Service with Org2 Admin', function () {   
    it('service 3 should be registered', function (done) {  
        Config.service3.datatypes[0].datatype_id = Config.datatype1.id;
        Config.service3.datatypes[1].datatype_id = Config.datatype2.id;

        chai.request(Config.server).post('omr/api/v1/services') 
            .set('Accept',  'application/json')
            .set('token', Config.orgAdminToken2)
            .send(Config.service3)
            .end(function(err, res) {
                expect(err).to.be.null;
                done();
            });
    });
});