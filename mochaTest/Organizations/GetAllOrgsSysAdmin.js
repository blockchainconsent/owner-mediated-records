const Config = require('../ConfigFile.js');
const chai = require('chai');
const { expect } = chai;
const chaiHttp = require('chai-http');
chai.use(chaiHttp);
chai.use(require('chai-like'));
const chaiThings = require('chai-things');
chai.use(chaiThings);

require("../Login/loginSysAdmin");

//Checks that the array of orgs returned includes all registered orgs
describe('Getting All Registered Orgs as SysAdmin', function() {
      require('./OrgSetup/RegisterOrg1');
        it('Should contain org1', function (done) {

            const org1 = {
                "id" : Config.org1.id,
                "name" : Config.org1.name
            };

            chai.request(Config.server).get('omr/api/v1/orgs')
                    .set('Accept', 'application/json')
                    .set('token', Config.sysAdminToken)
                    .end(function(err, res){                
                        expect(err).to.be.null;
                        expect(res.status).to.equal(200);
                        expect(res.body).to.be.an('array');
                        expect(res.body).to.contain.something.like(org1);
                        done();
                    });
        });

        require('./OrgSetup/RegisterOrg2')
        it('Should contain org2', function(done) {
            const org2 = {
                "id" : Config.org2.id,
                "name" : Config.org2.name
            };

            chai.request(Config.server).get('omr/api/v1/orgs')
                    .set('Accept', 'application/json')
                    .set('token', Config.sysAdminToken)
                    .end(function(err, res){                
                        expect(err).to.be.null;
                        expect(res.status).to.equal(200);
                        expect(res.body).to.contain.something.like(org2);
                        done();
                    });
        });
});