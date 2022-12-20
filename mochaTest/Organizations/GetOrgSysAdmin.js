const chai = require('chai');
const { expect } = chai;
const chaiHttp = require('chai-http');
chai.use(chaiHttp);
const chaiSubset = require('chai-subset');
chai.use(chaiSubset);


const Config = require('../ConfigFile');
require('../Login/loginSysAdmin');

//Test getOrg
//Checks the correct org details are returned given the org_id
describe('Getting Org Detail as SysAdmin', function() {
    require('./OrgSetup/RegisterOrg1');
    require('./OrgSetup/RegisterOrg2');

    describe('Getting Org that is Not Registered', function() {
        it('Should return a 200 test response & empty response body', function(done) {
            const fakeID = "fakeOrg";
            chai.request(Config.server).get('omr/api/v1/orgs/' + fakeID)
                .set('Accept', 'application/json')
                .set('token', Config.sysAdminToken)
                .set('org_id', fakeID)
                .end(function(err, res) {
                    expect(err).to.be.null
                    expect(res.status).to.equal(200);
                    expect(res.body).to.be.empty;
                    done();
                });
        });
    });

    describe('Getting Registered Orgs', function() {
        it('Should return a 200 test response & get Org1 public details', function (done) {
            chai.request(Config.server).get('omr/api/v1/orgs/' + Config.org1.id)
                    .set('Accept', 'application/json')
                    .set('token', Config.sysAdminToken)
                    .set('org_id', Config.org1.id)
                    .end(function(err, res){                
                        expect(err).to.be.null
                        expect(res.status).to.equal(200);
                        expect(res.body.id).to.equal(Config.org1.id);
                        expect(res.body.name).to.equal(Config.org1.name);
                        expect(res.body.email).to.equal("");
                        expect(res.body.secret).to.equal("");
                        done();
                    });
        });

        it('Should return a 200 test response & get Org2 public details', function(done) {
            chai.request(Config.server).get('omr/api/v1/orgs/' + Config.org2.id)
                    .set('Accept', 'application/json')
                    .set('token', Config.sysAdminToken)
                    .set('org_id', Config.org2.id)
                    .end(function(err, res){ 
                        expect(err).to.be.null              
                        expect(res.status).to.equal(200);
                        expect(res.body.id).to.equal(Config.org2.id);
                        expect(res.body.name).to.equal(Config.org2.name);
                        expect(res.body.email).to.equal("");
                        expect(res.body.secret).to.equal("");
                        done();
                    });
        });
    });
});