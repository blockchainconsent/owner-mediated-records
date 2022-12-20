const Config = require('../../ConfigFile.js');
const chai = require('chai');
const { expect } = chai;
const chaiHttp = require('chai-http');
chai.use(chaiHttp);
const idGenerator = require('../../Utils/helper');

require("../../Login/loginSysAdmin");
require("../../Organizations/OrgSetup/RegisterOrg1");
require("../../Login/OrgLogin/Org1Login");

describe('Get User', function () {
    var newUser = {
        "id": idGenerator(),
        "org": Config.org1.id,
        "secret": "pass0",
        "name": "newUser",
        "role": "user",
        "email":  "user@example.com",
        "ca_org": Config.caOrg,
        "data": {}
    };
   
    it('Register user should return a 200 response', function (done) {  
        chai.request(Config.server).post('omr/api/v1/users') 
            .set('Accept',  'application/json')
            .set('token', Config.orgAdminToken1)
            .send(newUser)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body.msg).to.include("success");
                done();
            });
    });

    it('Get user should return a 200 response, as org1 admin', function (done) {  
        chai.request(Config.server).get('omr/api/v1/users/' + newUser.id) 
            .set('Accept',  'application/json')
            .set('token', Config.orgAdminToken1)
            .send(newUser)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body.id).to.equal(newUser.id);
                expect(res.body.org).to.equal(newUser.org);
                expect(res.body.name).to.equal(newUser.name);
                expect(res.body.secret).to.equal(newUser.secret);
                expect(res.body.email).to.equal(newUser.email);
                done();
            });
    });

    it('Get user should return a 200 response, as system admin', function (done) {  
        chai.request(Config.server).get('omr/api/v1/users/' + newUser.id) 
            .set('Accept',  'application/json')
            .set('token', Config.sysAdminToken)
            .send(newUser)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body.id).to.equal(newUser.id);
                expect(res.body.org).to.equal(newUser.org);
                expect(res.body.name).to.equal(newUser.name);
                expect(res.body.secret).to.equal("");
                expect(res.body.email).to.equal("");
                done();
            });
    });
});