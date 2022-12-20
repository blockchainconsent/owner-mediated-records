var Config = require('../../ConfigFile.js');
const chai = require('chai');
const chaiHttp = require('chai-http');
const { expect, config } = require('chai');
chai.use(chaiHttp);

//Test to get the login credentials of org1Admin
//Expected: Token should generate for the given valid login
describe('Org1 Login', function () {

    it('Should return a 200 test response', function (done) {
        chai.request(Config.server).get('common/api/v1/login')
            .set('Accept', 'application/json')
            .set('user-id', Config.org1.id)
            .set('password', Config.org1.secret)
            .set('login-org', Config.org1.ca_org)
            .set('login-channel', Config.channel)
            .set('signature', "")
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.not.deep.equal({})
                expect(res.body.token).to.not.equal("")
                Config.orgAdminToken1 = res.body.token;
                done();
            });
    });
    
})
