var Config = require('../ConfigFile.js');
const chai = require('chai');
const { expect } = chai;
const chaiHttp = require('chai-http');
const { getConfigSetting } = require('fabric-client');
chai.use(chaiHttp);

require('./OrgSetup/RegisterOrg1');
require('../Login/loginSysAdmin')

describe('Updating Org as unauthorized Users', function() {
    var newOrg1 = {
        "id": Config.org1.id,
        "name": Config.org1.name,
        "ca_org": Config.org1.ca_org,
        "data": {},
        "status": "active",
        "email": "org1@email.com",
    };

    it('Updating as Sys Admin: Should Return a 401 response', function(done) {
        chai.request(Config.server).put('omr/api/v1/orgs/' + newOrg1.id)
                .set('Accept', 'application/json')
                .set('token', Config.sysAdminToken)
                .send(newOrg1)
                .end(function(err, res){             
                    expect(err).to.be.null;
                    expect(res).to.have.property("text");
                    expect(JSON.parse(res.text)).to.have.property("msg");
                    expect(JSON.parse(res.text).msg).to.equal("Unauthorized to update the org");
                    expect(res.status).to.equal(401);
                    done();
                });
    })
    

});
