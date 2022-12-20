var Config = require('../../ConfigFile.js');
const chai = require('chai');
const { expect } = chai;
const chaiHttp = require('chai-http')
chai.use(chaiHttp);
const idGenerator = require('../../Utils/helper');

require("../../Organizations/OrgSetup/RegisterOrg1")
require("../../Login/loginSysAdmin")
require("../../Login/OrgLogin/Org1Login")

describe('Add Org Admin Permission to org user as Sys Admin', function () {

    const newUser1 = {
        "id": idGenerator(),
        "secret": "newuser1",
        "name": "newUser1",
        "role": "user",
        "org" : "",
        "email":  "newuser1email@example.com",
        "ca_org": Config.caOrg,
        "data": {
          "address": "2 User St 1"
        }
    };

    const newUser2 = {
        "id": idGenerator(),
        "secret": "newuser2",
        "name": "newUser2",
        "role": "user",
        "org" : Config.org1.id,
        "email":  "newuser2email@example.com",
        "ca_org": Config.caOrg,
        "data": {
          "address": "2 User St 2"
        }
    };

    // register newUser1
    it('Should register newUser1 successfully', function (done) {
        chai.request(Config.server).post('omr/api/v1/users') 
            .set('Accept',  'application/json')
            .set('token', Config.orgAdminToken1)
            .send(newUser1)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("success");
                done();
            });
    });

    // register newUser2
    it('Should register newUser2 successfully', function (done) {
        chai.request(Config.server).post('omr/api/v1/users') 
            .set('Accept',  'application/json')
            .set('token', Config.orgAdminToken1)
            .send(newUser2)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("success");
                done();
            });
    });

    // Add Org Admin Permission to newUser2 by unauthorized user - sys admin
    it('Should not add org admin permission to user by unauthorized user', function (done) {  
        chai.request(Config.server).put('omr/api/v1/users/' + newUser2.id + '/permissions/admin/' + Config.org1.id) 
            .set('Accept',  'application/json')
            .set('token', Config.sysAdminToken)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(500);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("error");
                done();
            });
    });

    // Add Org Admin Permission to newUser1 which doesn't belong to Org
    it('Should not add org admin permission to user when they do not belong to any org', function (done) {  
        chai.request(Config.server).put('omr/api/v1/users/' + newUser1.id + '/permissions/admin/' + Config.org1.id) 
            .set('Accept',  'application/json')
            .set('token', Config.orgAdminToken1)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(500);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("error");
                done();
            });
    });
});