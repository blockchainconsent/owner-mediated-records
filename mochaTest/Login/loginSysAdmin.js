var Config = require('../ConfigFile.js');
const { start } = require('repl');
const chai = require('chai');
const chaiHttp = require('chai-http');
const { expect, config } = require('chai');
const { addConfigFile } = require('fabric-client');
chai.use(chaiHttp);

//Test to get the login credentials of system Admin
//Expected: Token should generate for the given valid login
describe('Sys Admin Login', function () {
    before(done => {
        // wait for the server to start before running mocha tests
        Config.app.on('server_started', () => done());
    });

    it('Should return a 200 test response', function (done) {
        chai.request(Config.server).get('common/api/v1/login')
            .set('Accept', 'application/json')
            .set('user-id', Config.sysAdminId)
            .set('password', Config.sysAdminPass)
            .set('login-org', Config.caOrg)
            .set('login-channel', Config.channel)
            .set('signature', "")
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.not.deep.equal({})
                expect(res.body.token).to.not.equal("")
                Config.sysAdminToken = res.body.token;
                done();
            });
    });
    
})
