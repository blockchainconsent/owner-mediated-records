const Config = require('../../ConfigFile.js');
const chai = require('chai');
const { expect } = chai;
const chaiHttp = require('chai-http');
chai.use(chaiHttp);
chai.use(require('chai-like'));
const chaiThings = require('chai-things');
chai.use(chaiThings);

require('../../Organizations/OrgSetup/RegisterOrg1');
require('../../Organizations/OrgSetup/RegisterOrg2');
require("../RegisterUsers");
require("../../Login/User1Login");

//Checks that the array of orgs returned includes all registered orgs
describe('Getting All Registered Orgs as Role User', function() {
    it('Should contain org1 and org2', function (done) {

        const org1 = {
            "id" : Config.org1.id,
            "name" : Config.org1.name,
            "solution_private_data": null
        }

        const org2 = {
            "id" : Config.org2.id,
            "name" : Config.org2.name,
            "solution_private_data": null
        }

        chai.request(Config.server).get('omr/api/v1/orgs')
                .set('Accept', 'application/json')
                .set('token', Config.user1Token)
                .end(function(err, res){                
                    expect(err).to.be.null;
                    expect(res.status).to.equal(200);
                    expect(res.body).to.be.an('array');
                    expect(res.body).to.contain.something.like(org1);
                    expect(res.body).to.contain.something.like(org2);
                    done();
                });
    })

});