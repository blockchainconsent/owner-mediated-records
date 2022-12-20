const chai = require('chai');
const { expect } = chai;
const chaiHttp = require('chai-http');
chai.use(chaiHttp);
const chaiSubset = require('chai-subset');
chai.use(chaiSubset);

const Config = require('../../ConfigFile');
require('../../Organizations/OrgSetup/RegisterOrg1');
require('../../Organizations/OrgSetup/RegisterOrg2');
require("../RegisterUsers");
require("../../Login/User1Login");

//Test getOrg
describe('Getting Org Detail as Role User', function() {
    it('Should return a 200 test response & get Org1 public details', function (done) {
        chai.request(Config.server).get('omr/api/v1/orgs/' + Config.org1.id)
                .set('Accept', 'application/json')
                .set('token', Config.user1Token)
                .set('org_id', Config.org1.id)
                .end(function(err, res){                
                    expect(err).to.be.null
                    expect(res.status).to.equal(200);
                    expect(res.body.id).to.equal(Config.org1.id);
                    expect(res.body.name).to.equal(Config.org1.name);
                    expect(res.body.solution_private_data).to.be.null;
                    done();
                });
    });
});