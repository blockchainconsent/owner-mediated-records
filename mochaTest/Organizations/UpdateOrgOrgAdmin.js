let Config = require('../ConfigFile.js');
const chai = require('chai');
const { expect } = chai;
const chaiHttp = require('chai-http');
const like = require('chai-like');
const { getConfigSetting } = require('fabric-client');
chai.use(chaiHttp);
chai.use(like);

require('./OrgSetup/RegisterOrg1');
require('../Login/OrgLogin/Org1Login');

//Test updating org info
//Checks that org is successfully updated
describe('Updating An Existing Org Info', function() {

    let oldName = Config.org1.name;
    let oldEmail = Config.org1.email;
    let oldData = Config.org1.data;

    let newOrg1 = {
            "id": Config.org1.id,
            "name": "newOrg1Name",
            "ca_org": Config.org1.ca_org,
            "data": {
                "address" : "0 New Address St"
            },
            "status": "active",
            "email" : "newOrgEmail@org.com"
    };

    it('Updating Org1: Should return a 200 test response', function(done) {
        chai.request(Config.server).put('omr/api/v1/orgs/' + Config.org1.id)
                .set('Accept', 'application/json')
                .set('token', Config.orgAdminToken1)
                .send(newOrg1)
                .end(function(err, res){             
                    expect(err).to.be.null;
                    expect(res.status).to.equal(200);
                    expect(res.body).to.have.property("msg");
                    expect(res.body.msg).to.include("success");
                    done();
                });
    })

    it('Getting Org1 Details: Should return updated data', function (done) {
        chai.request(Config.server).get('omr/api/v1/orgs/' + Config.org1.id)
                .set('Accept', 'application/json')
                .set('token', Config.orgAdminToken1)
                .set('org_id', Config.org1.id)
                .end(function(err, res){                
                    expect(err).to.be.null
                    expect(res.status).to.equal(200);
                    expect(res.body.name).to.equal(newOrg1.name);
                    expect(res.body.email).to.equal(newOrg1.email);
                    expect(res.body.solution_private_data).to.be.like(newOrg1.data);
                    done();
                });
    })

    after(function(done) {
        let resetOrg1 = {
            "id": Config.org1.id,
            "name": oldName,
            "ca_org": Config.org1.ca_org,
            "data": oldData,
            "status": "active",
            "email" : oldEmail
        }

        chai.request(Config.server).put('omr/api/v1/orgs/' + Config.org1.id)
                .set('Accept', 'application/json')
                .set('org_id', Config.org1.id)
                .set('token', Config.orgAdminToken1)
                .send(resetOrg1)
                .end(function(err, res){                
                    expect(err).to.be.null;
                    expect(res.status).to.equal(200);
                    expect(res.body).to.have.property("msg");
                    expect(res.body.msg).to.include("success");
                    done(); 
                });
    })
});

