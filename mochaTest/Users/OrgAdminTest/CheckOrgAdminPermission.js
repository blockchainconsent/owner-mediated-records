const chai = require('chai');
const { expect } = chai;
const chaiHttp = require('chai-http')
chai.use(chaiHttp);
var chaiSubset = require('chai-subset');
chai.use(chaiSubset);
const idGenerator = require('../../Utils/helper');
const Config = require('../../ConfigFile');

require('../../Organizations/OrgSetup/RegisterOrg1');
require('../../Organizations/OrgSetup/RegisterOrg2');
require("../../Login/OrgLogin/Org1Login");

// Test getOrg
describe('Check Org Admin Permission', function() {

    const newUser1 = {
        "id": idGenerator(),
        "secret": "newuser1",
        "name": "newUser1",
        "role": "user",
        "org" : Config.org1.id,
        "email":  "newuser1email@example.com",
        "ca_org": Config.caOrg,
        "data": {
          "address": "2 User St"
        }
    };

    // register newUser1
    it('Should register newUser1 with role user successfully', function (done) {
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

    let newUser1Token = '';
    it('Should return a login token for newUser1', function (done) {
        chai.request(Config.server).get('common/api/v1/login')
            .set('Accept', 'application/json')
            .set('user-id', newUser1.id)
            .set('password', newUser1.secret)
            .set('login-org', newUser1.ca_org)
            .set('login-channel', Config.channel)
            .set('signature', "")
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("token");
                expect(res.body.token).to.not.equal("");
                newUser1Token = res.body.token;
                done();
            });
    });

    it('newUser1 is_org_admin should be false', function (done) {
        chai.request(Config.server).get('omr/api/v1/users/' + newUser1.id)
            .set('Accept', 'application/json')
            .set('token', newUser1Token)
            .end(function(err, res){
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body.id).to.equal(newUser1.id);
                expect(res.body.solution_info.is_org_admin).to.be.false;
                done();
            });
    });


    it('Should fail to register datatype, newUser1 is not org admin', function (done) {  
        chai.request(Config.server).post('omr/api/v1/datatypes') 
            .set('Accept',  'application/json')
            .set('token', newUser1Token)
            .send({id: idGenerator(), description: 'description'})
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(500);
                done();
            });
    });

    it('Giving newUser1 Org Admin Permission', function (done) {  
        chai.request(Config.server).put('omr/api/v1/users/' + newUser1.id + '/permissions/admin/' + Config.org1.id) 
            .set('Accept',  'application/json')
            .set('token', Config.orgAdminToken1)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("success");
                done();
            });
    });

    it('newUser1 is_org_admin should be true', function (done) {
        chai.request(Config.server).get('omr/api/v1/users/' + newUser1.id)
            .set('Accept', 'application/json')
            .set('token', newUser1Token)
            .end(function(err, res){                
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body.id).to.equal(newUser1.id);
                expect(res.body.solution_info.is_org_admin).be.true;
                done();
            });
    });

    it('Should successfully register datatype, newUser1 is org admin', function (done) {  
        chai.request(Config.server).post('omr/api/v1/datatypes') 
            .set('Accept',  'application/json')
            .set('token', newUser1Token)
            .send({id: idGenerator(), description: 'description'})
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                done();
            });
    });

    it('Remove newUser1 Org Admin Permission', function (done) {  
        chai.request(Config.server).delete('omr/api/v1/users/' + newUser1.id + '/permissions/admin/' + Config.org1.id) 
            .set('Accept',  'application/json')
            .set('token', Config.orgAdminToken1)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("success");
                done();
            });
    });

    it('newUser1 is_org_admin should be false', function (done) {
        chai.request(Config.server).get('omr/api/v1/users/' + newUser1.id)
            .set('Accept', 'application/json')
            .set('token', newUser1Token)
            .end(function(err, res){                
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body.id).to.equal(newUser1.id);
                expect(res.body.solution_info.is_org_admin).to.be.false;
                done();
            });
    });

    it('Should fail to register datatype, newUser1 is no longer org admin', function (done) {  
        chai.request(Config.server).post('omr/api/v1/datatypes') 
            .set('Accept',  'application/json')
            .set('token', newUser1Token)
            .send({id: idGenerator(), description: 'description'})
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(500);
                done();
            });
    });
});