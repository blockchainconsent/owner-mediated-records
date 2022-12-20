const Config = require('../../ConfigFile.js');
const chai = require('chai');
const { expect } = chai;
const chaiHttp = require('chai-http');
chai.use(chaiHttp);
chai.use(require('chai-like'));
const chaiThings = require('chai-things');
chai.use(chaiThings);
const idGenerator = require('../../Utils/helper');

require('../../Login/loginSysAdmin');
require("../../Organizations/OrgSetup/RegisterOrg1");
require("../../DataTypes/RegisterDatatypesSysAdmin");
require("../../Services/RegisterServiceOrg1Admin");
require("../../Login/OrgLogin/Org1Login");

describe('Check Audit Admin Permission as org admin', function() {

    // create 3 new random users
    const newUser1 = {
        "id": idGenerator(),
        "secret": "newuser1",
        "name": "newUser1",
        "role": "user",
        "org" : "",
        "email":  "newuser1email@example.com",
        "ca_org": Config.caOrg,
        "data": {
          "address": "2 User St"
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
          "address": "2 User St"
        }
    };

    const newAuditor = {
        "id": idGenerator(),
        "secret": "newuser3",
        "name": "newAuditor",
        "role": "audit",
        "org" : "",
        "email":  "newuser3email@example.com",
        "ca_org": Config.caOrg,
        "data": {
          "address": "2 User St"
        }
    };
    
    // register new user1 (doesn't belong to any Org, with role user)
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

    // register new user2 (belongs to Org1, with role user)
    it('Should register newUser2 with role user successfully', function (done) {
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

    // register new user3 (doesn't belong to any Org, with role audit)
    it('Should register newAuditor with role audit successfully', function (done) {
        chai.request(Config.server).post('omr/api/v1/users') 
            .set('Accept',  'application/json')
            .set('token', Config.sysAdminToken)
            .send(newAuditor)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("success");
                done();
            });
    });

    // grant auditor admin permission to newAuditor as org admin
    it('Should successfully grant auditor permission to user as org admin', function (done) {
        chai.request(Config.server).put('omr/api/v1/users/' + newAuditor.id + '/permissions/audit/' + Config.service1.id) 
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

    it('Should not successfully grant auditor permission to user with role is not equal to audit', function (done) {
        chai.request(Config.server).put('omr/api/v1/users/' + newUser1.id + '/permissions/audit/' + Config.service1.id) 
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

    // revoke auditor admin permission from newAuditor as org admin
    it('Should successfully revoke auditor permission from user as org admin', function (done) {
        const msg = 'removing auditor permission completed successfully';
        chai.request(Config.server).delete('omr/api/v1/users/' + newAuditor.id + '/permissions/audit/' + Config.service1.id) 
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

    it('Should successfully grant auditor permission to user as org admin ONE MORE TIME', function (done) {
        chai.request(Config.server).put('omr/api/v1/users/' + newAuditor.id + '/permissions/audit/' + Config.service1.id) 
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

    it('Should successfully revoke auditor permission from user as org admin ONE MORE TIME', function (done) {
        chai.request(Config.server).delete('omr/api/v1/users/' + newAuditor.id + '/permissions/audit/' + Config.service1.id) 
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
});


describe('Check Audit Admin Permission as service admin', function() {

    // create 4 new random users
    const newUser1 = {
        "id": idGenerator(),
        "secret": "newuser1",
        "name": "newUser1",
        "role": "user",
        "org" : "",
        "email":  "newuser1email@example.com",
        "ca_org": Config.caOrg,
        "data": {
          "address": "2 User St"
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
          "address": "2 User St"
        }
    };

    const newAuditor = {
        "id": idGenerator(),
        "secret": "newuser3",
        "name": "newAuditor",
        "role": "audit",
        "org" : "",
        "email":  "newuser3email@example.com",
        "ca_org": Config.caOrg,
        "data": {
          "address": "2 User St"
        }
    };

    const newServiceAdmin = {
        "id": idGenerator(),
        "secret": "newuser4",
        "name": "newServiceAdmin",
        "role": "user",
        "org" : Config.org1.id,
        "email":  "newuser4email@example.com",
        "ca_org": Config.caOrg,
        "data": {
          "address": "2 User St"
        }
    };
    
    // register new user1 (doesn't belong to any Org, with role user)
    it('Should register newUser1 with role user successfully', function (done) {
        chai.request(Config.server).post('omr/api/v1/users') 
            .set('Accept',  'application/json')
            .set('token', Config.sysAdminToken)
            .send(newUser1)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("success");
                done();
            });
    });

    // register new user2 (belongs to Org1, with role user)
    it('Should register newUser2 with role user successfully', function (done) {
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

    // register new user3 (doesn't belong to any Org, with role audit)
    it('Should register newAuditor with role audit successfully', function (done) {
        chai.request(Config.server).post('omr/api/v1/users') 
            .set('Accept',  'application/json')
            .set('token', Config.sysAdminToken)
            .send(newAuditor)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("success");
                done();
            });
    });

    // register new user4 (belongs to Org1, with role user)
    it('Should register newnewServiceAdmin with role user successfully', function (done) {
        chai.request(Config.server).post('omr/api/v1/users') 
            .set('Accept',  'application/json')
            .set('token', Config.orgAdminToken1)
            .send(newServiceAdmin)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("success");
                done();
            });
    });

    // grant service admin permission to newServiceAdmin
    it('Should successfully grant service admin permission to user', function (done) {
        chai.request(Config.server).put('omr/api/v1/users/' + newServiceAdmin.id + '/permissions/services/' + Config.service1.id) 
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

    // login as service admin
    let serviceAdminToken = '';
    it('Should return a login token for service admin', function (done) {
        chai.request(Config.server).get('common/api/v1/login')
            .set('Accept', 'application/json')
            .set('user-id', newServiceAdmin.id)
            .set('password', newServiceAdmin.secret)
            .set('login-org', newServiceAdmin.ca_org)
            .set('login-channel', Config.channel)
            .set('signature', "")
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("token");
                expect(res.body.token).to.not.equal("");
                serviceAdminToken = res.body.token;
                done();
            });
    });

    // grant auditor admin permission to newAuditor as org admin
    it('Should successfully grant auditor permission to user as org admin', function (done) {
        chai.request(Config.server).put('omr/api/v1/users/' + newAuditor.id + '/permissions/audit/' + Config.service1.id) 
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

    // revoke auditor admin permission from newAuditor as org admin
    it('Should successfully revoke auditor permission from user as org admin', function (done) {
        chai.request(Config.server).delete('omr/api/v1/users/' + newAuditor.id + '/permissions/audit/' + Config.service1.id) 
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

    // grant auditor admin permission to newAuditor as serviceadmin
    it('Should successfully grant auditor permission to user as service admin', function (done) {
        chai.request(Config.server).put('omr/api/v1/users/' + newAuditor.id + '/permissions/audit/' + Config.service1.id) 
            .set('Accept',  'application/json')
            .set('token', serviceAdminToken)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("success");
                done();
            });
    });

    // grant auditor admin permission to newAuditor as serviceadmin
    it('Should successfully grant auditor permission to user as service admin ONE MORE TIME', function (done) {
        chai.request(Config.server).put('omr/api/v1/users/' + newAuditor.id + '/permissions/audit/' + Config.service1.id) 
            .set('Accept',  'application/json')
            .set('token', serviceAdminToken)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("success");
                done();
            });
    });

    // grant auditor permission to user with role user
    it('Should not successfully grant auditor permission to user with role is not equal to audit', function (done) {
        chai.request(Config.server).put('omr/api/v1/users/' + newUser1.id + '/permissions/audit/' + Config.service1.id) 
            .set('Accept',  'application/json')
            .set('token', serviceAdminToken)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(500);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("error");
                done();
            });
    });

    // revoke auditor admin permission from newUser3 as serviceadmin
    it('Should successfully revoke auditor permission from user as service admin', function (done) {
        chai.request(Config.server).delete('omr/api/v1/users/' + newAuditor.id + '/permissions/audit/' + Config.service1.id) 
            .set('Accept',  'application/json')
            .set('token', serviceAdminToken)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("success");
                done();
            });
    });

    // revoke auditor permission as org admin one more time
    it('Should successfully revoke auditor permission from user as org admin ONE MORE TIME', function (done) {
        chai.request(Config.server).delete('omr/api/v1/users/' + newAuditor.id + '/permissions/audit/' + Config.service1.id) 
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
});

describe('Check that org user that used to be a service admin/org admin is not able to remove auditor permission', function() {

    // create new org admin user
    const newOrgAdminUser = {
        "id": idGenerator(),
        "secret": "newuser2",
        "name": "newOrgAdminUser",
        "role": "user",
        "org" : Config.org1.id,
        "email":  "newuser2email@example.com",
        "ca_org": Config.caOrg,
        "data": {
          "address": "2 User St"
        }
    };

    // create new auditor user
    const newAuditor = {
        "id": idGenerator(),
        "secret": "newuser3",
        "name": "newAuditor",
        "role": "audit",
        "org" : "",
        "email":  "newuser3email@example.com",
        "ca_org": Config.caOrg,
        "data": {
          "address": "2 User St"
        }
    };

    // create new service admin user
    const newServiceAdmin = {
        "id": idGenerator(),
        "secret": "newuser4",
        "name": "newServiceAdmin",
        "role": "user",
        "org" : Config.org1.id,
        "email":  "newuser4email@example.com",
        "ca_org": Config.caOrg,
        "data": {
          "address": "2 User St"
        }
    };
    
    // register newOrgAdminUser (belongs to Org1, with role user)
    it('Should register newOrgAdminUser with role user successfully', function (done) {
        chai.request(Config.server).post('omr/api/v1/users') 
            .set('Accept',  'application/json')
            .set('token', Config.orgAdminToken1)
            .send(newOrgAdminUser)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("success");
                done();
            });
    });

    // register newAuditor (doesn't belong to any Org, with role audit)
    it('Should register newAuditor with role audit successfully', function (done) {
        chai.request(Config.server).post('omr/api/v1/users') 
            .set('Accept',  'application/json')
            .set('token', Config.sysAdminToken)
            .send(newAuditor)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("success");
                done();
            });
    });

    // register newServiceAdmin (belongs to Org1, with role user)
    it('Should register newnewServiceAdmin with role user successfully', function (done) {
        chai.request(Config.server).post('omr/api/v1/users') 
            .set('Accept',  'application/json')
            .set('token', Config.orgAdminToken1)
            .send(newServiceAdmin)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("success");
                done();
            });
    });

    // grant service admin permission to newServiceAdmin
    it('Should successfully grant service admin permission to user', function (done) {
        chai.request(Config.server).put('omr/api/v1/users/' + newServiceAdmin.id + '/permissions/services/' + Config.service1.id) 
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

    // grant org admin permission to newOrgAdminUser
    it('Should successfully grant org admin permission to user', function (done) {
        chai.request(Config.server).put('omr/api/v1/users/' + newOrgAdminUser.id + '/permissions/admin/' + Config.org1.id) 
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

    // revoke org admin permission from newOrgAdminUser
    it('Should successfully revoke org admin permission from user', function (done) {
        chai.request(Config.server).delete('omr/api/v1/users/' + newOrgAdminUser.id + '/permissions/admin/' + Config.org1.id) 
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

    // revoke service admin permission from newServiceAdmin
    it('Should successfully revoke service admin permission from user', function (done) {
        chai.request(Config.server).delete('omr/api/v1/users/' + newServiceAdmin.id + '/permissions/services/' + Config.service1.id) 
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

    // login as new org admin
    let newOrgAdminUserToken = '';
    it('Should return a login token for new org admin', function (done) {
        chai.request(Config.server).get('common/api/v1/login')
            .set('Accept', 'application/json')
            .set('user-id', newOrgAdminUser.id)
            .set('password', newOrgAdminUser.secret)
            .set('login-org', newOrgAdminUser.ca_org)
            .set('login-channel', Config.channel)
            .set('signature', "")
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("token");
                expect(res.body.token).to.not.equal("");
                newOrgAdminUserToken = res.body.token;
                done();
            });
    });

    // login as new service admin
    let serviceAdminToken = '';
    it('Should return a login token for service admin', function (done) {
        chai.request(Config.server).get('common/api/v1/login')
            .set('Accept', 'application/json')
            .set('user-id', newServiceAdmin.id)
            .set('password', newServiceAdmin.secret)
            .set('login-org', newServiceAdmin.ca_org)
            .set('login-channel', Config.channel)
            .set('signature', "")
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("token");
                expect(res.body.token).to.not.equal("");
                serviceAdminToken = res.body.token;
                done();
            });
    });

    // grant auditor admin permission to newAuditor as org admin
    it('Should successfully grant auditor permission to user as org admin', function (done) {
        chai.request(Config.server).put('omr/api/v1/users/' + newAuditor.id + '/permissions/audit/' + Config.service1.id) 
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

    // revoke auditor admin permission from newAuditor as new org admin
    it('Should return 500 error when org user that used to be an org admin is trying to revoke auditor permission from user', function (done) {
        chai.request(Config.server).delete('omr/api/v1/users/' + newAuditor.id + '/permissions/audit/' + Config.service1.id) 
            .set('Accept',  'application/json')
            .set('token', newOrgAdminUserToken)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(500);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("error");
                done();
            });
    });

    // revoke auditor admin permission from user as newServiceAdmin
    it('Should return 500 error when org user that used to be an service admin is trying to revoke auditor permission from user', function (done) {
        chai.request(Config.server).delete('omr/api/v1/users/' + newAuditor.id + '/permissions/audit/' + Config.service1.id) 
            .set('Accept',  'application/json')
            .set('token', serviceAdminToken)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(500);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("error");
                done();
            });
    });
});