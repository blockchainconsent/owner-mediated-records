var Config = require('../../ConfigFile.js');
const chai = require('chai');
const { expect } = chai;
const chaiHttp = require('chai-http');
chai.use(chaiHttp);

require('../../Organizations/OrgSetup/RegisterOrg1');
require("../RegisterUsers");
require("../../Login/User1Login");

describe('Updating Org as Role User', function() {
    var newOrg1 = {
        "id": Config.org1.id,
        "name": Config.org1.name,
        "ca_org": Config.org1.ca_org,
        "data": {},
        "status": "active",
        "email": "org1@email.com",
    };

    it('Should Return a 401 response', function(done) {
        chai.request(Config.server).put('omr/api/v1/orgs/' + Config.org1.id)
                .set('Accept', 'application/json')
                .set('token', Config.user1Token)
                .send(newOrg1)
                .end(function(err, res){             
                    expect(err).to.be.null;
                    expect(res).to.have.property("text");
                    expect(res.body).to.have.property("msg");
                    expect(res.body.msg).to.equal("Unauthorized to update the org");
                    expect(res.status).to.equal(401);
                    done();
                });
    })

});
