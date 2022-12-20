const chai = require('chai');
const { expect } = chai;
const chaiHttp = require('chai-http');
chai.use(chaiHttp);
const chaiSubset = require('chai-subset');
chai.use(chaiSubset);
chai.use(require('chai-like'));
const chaiThings = require('chai-things');
chai.use(chaiThings);

const Config = require('../ConfigFile');
require('../Organizations/OrgSetup/RegisterOrg1');
require('../Login/OrgLogin/Org1Login');
require('../Organizations/OrgSetup/RegisterOrg2');
require('../Login/OrgLogin/Org2Login');

//Test getOrg
//Checks the correct org details are returned given the org_id
//returns private and public details
describe('Getting Org1 Detail as Org1 Admin', function() {

    it('Should return a 200 test response & get Org1 private details', function (done) {
        chai.request(Config.server).get('omr/api/v1/orgs/' + Config.org1.id)
                .set('Accept', 'application/json')
                .set('token', Config.orgAdminToken1)
                .set('org_id', Config.org1.id)
                .end(function(err, res){                
                    expect(err).to.be.null
                    expect(res.status).to.equal(200);
                    expect(res).to.not.deep.equal({});
                    expect(res.body.id).to.equal(Config.org1.id);
                    expect(res.body.name).to.equal(Config.org1.name);
                    expect(res.body.email).to.equal(Config.org1.email);
                    expect(res.body.secret).to.equal(Config.org1.secret);
                    done();
                });
    });
});

//Test getOrg
//Should only return public details
describe('Getting Org1 Detail as Org2 Admin', function() {

    it('Should return a 200 test response & get Org1 public details', function (done) {
        chai.request(Config.server).get('omr/api/v1/orgs/' + Config.org1.id)
                .set('Accept', 'application/json')
                .set('token', Config.orgAdminToken2)
                .set('org_id', Config.org1.id)
                .end(function(err, res){  
                    expect(err).to.be.null
                    expect(res.status).to.equal(200);
                    expect(res).to.not.deep.equal({});
                    expect(res.body.id).to.equal(Config.org1.id);
                    expect(res.body.name).to.equal(Config.org1.name);
                    expect(res.body.email).to.equal("");
                    expect(res.body.secret).to.equal("");
                    expect(res.body.solution_private_data).to.be.null;
                    done();
                });
    });
});