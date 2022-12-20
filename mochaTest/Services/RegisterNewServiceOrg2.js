var Config = require('../ConfigFile.js');
const chai = require('chai');
const { expect } = chai;
const chaiHttp = require('chai-http')
chai.use(chaiHttp);
const { v4: uuidv4 } = require('uuid')

require("../Organizations/OrgSetup/RegisterOrg2")
require("../DataTypes/RegisterDatatypesSysAdmin")
require("../Login/OrgLogin/Org2Login");

//Test to Register Service through OrgAdmin
//Registers a new service to Org2
describe('Register Service as OrgAdmin', function () {        
    it('New service should return a 200 test response', function (done) {  
        var newID = uuidv4();
        var newService = {
            "id": newID,
            "name": newID,
            "secret": "servicepass",
            "ca_org": Config.caOrg,
            "email": "serviceemail@example.com",
            "org_id": Config.org2.id,
            "summary": "service summary",
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
        
        chai.request(Config.server).post('omr/api/v1/services') 
            .set('Accept',  'application/json')
            .set('token', Config.orgAdminToken2)
            .send(newService)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.equal("service registration completed successfully");
                done();
            });
        
    });
});