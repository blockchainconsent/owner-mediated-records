var Config = require('../ConfigFile.js');
const chai = require('chai');
const { expect } = chai;
const chaiHttp = require('chai-http');
const { getConfigSetting } = require('fabric-client');
chai.use(chaiHttp);

require('./OrgSetup/RegisterOrg1');
require('../Login/OrgLogin/Org1Login');

describe('Tests that should FAIL to update Org', function() {
    var missingID= {
            "name": "newOrg1Name",
            "ca_org": Config.org1.ca_org,
            "data": {},
            "status": "active",
            "email": "org1@email.com",
    };

    it('Updating Org with Missing ID', function(done) {
        chai.request(Config.server).put('omr/api/v1/orgs/' + Config.org1.id)
                .set('Accept', 'application/json')
                .set('token', Config.orgAdminToken1)
                .send(missingID)
                .end(function(err, res){             
                    expect(err).to.be.null;
                    expect(res).to.have.property("text");
                    expect(JSON.parse(res.text)).to.have.property("msg");
                    expect(JSON.parse(res.text).msg).to.equal("Invalid data:id missing");
                    expect(res.status).to.equal(400);
                    done();
                });
    })

    var diffPath= {
        "id": "diffPath",
        "name": "newOrg1Name",
        "ca_org": Config.org1.ca_org,
        "data": {},
        "status": "active",
        "email": "org1@email.com",
    };

    it('Updating Org with Diff Path From Data Body', function(done) {
        chai.request(Config.server).put('omr/api/v1/orgs/' + Config.org1.id)
                .set('Accept', 'application/json')
                .set('token', Config.orgAdminToken1)
                .send(diffPath)
                .end(function(err, res){             
                    expect(err).to.be.null;
                    expect(res).to.have.property("text");
                    expect(JSON.parse(res.text)).to.have.property("msg");
                    expect(JSON.parse(res.text).msg).to.equal("ID in path and data body does not match");
                    expect(res.status).to.equal(400);
                    done();
                });
    })

    var missingName= {
        "id": Config.org1.id,
        "ca_org": Config.org1.ca_org,
        "data": {},
        "status": "active",
        "email": "org1@email.com",
    };

    it('Updating Org with Missing Name', function(done) {
        chai.request(Config.server).put('omr/api/v1/orgs/' + Config.org1.id)
                .set('Accept', 'application/json')
                .set('token', Config.orgAdminToken1)
                .send(missingName)
                .end(function(err, res){             
                    expect(err).to.be.null;
                    expect(res).to.have.property("text");
                    expect(JSON.parse(res.text)).to.have.property("msg");
                    expect(JSON.parse(res.text).msg).to.equal("Invalid data:name missing");
                    expect(res.status).to.equal(400);
                    done();
                });
    })

    var missingCa_org= {
        "id": Config.org1.id,
        "name": Config.org1.name,
        "data": {},
        "status": "active",
        "email": "org1@email.com",
    };

    it('Updating Org with Missing CA_Org', function(done) {
        chai.request(Config.server).put('omr/api/v1/orgs/' + Config.org1.id)
                .set('Accept', 'application/json')
                .set('token', Config.orgAdminToken1)
                .send(missingCa_org)
                .end(function(err, res){             
                    expect(err).to.be.null;
                    expect(res).to.have.property("text");
                    expect(JSON.parse(res.text)).to.have.property("msg");
                    expect(JSON.parse(res.text).msg).to.equal("Invalid data:ca_org missing");
                    expect(res.status).to.equal(400);
                    done();
                });
    })

    var inactiveOrg= {
        "id": Config.org1.id,
        "name": Config.org1.name,
        "ca_org": Config.org1.ca_org,
        "data": {},
        "status": "inactive",
        "email": "org1@email.com",
    };

    it('Updating Org to be Inactive', function(done) {
        chai.request(Config.server).put('omr/api/v1/orgs/' + Config.org1.id)
                .set('Accept', 'application/json')
                .set('token', Config.orgAdminToken1)
                .send(inactiveOrg)
                .end(function(err, res){             
                    expect(err).to.be.null;
                    expect(res).to.have.property("text");
                    expect(JSON.parse(res.text)).to.have.property("msg");
                    expect(JSON.parse(res.text).msg).to.equal("Invalid data:status has to be \"active\"");
                    expect(res.status).to.equal(400);
                    done();
                });
    })

    var missingEmail = {
        "id": Config.org1.id,
        "name": Config.org1.name,
        "ca_org": Config.org1.ca_org,
        "data": {},
        "status": "active",
    };

    it('Updating Org with Missing Email', function(done) {
        chai.request(Config.server).put('omr/api/v1/orgs/' + Config.org1.id)
                .set('Accept', 'application/json')
                .set('token', Config.orgAdminToken1)
                .send(missingEmail)
                .end(function(err, res){             
                    expect(err).to.be.null;
                    expect(res).to.have.property("text");
                    expect(JSON.parse(res.text)).to.have.property("msg");
                    expect(JSON.parse(res.text).msg).to.equal("Invalid data:email missing");
                    expect(res.status).to.equal(400);
                    done();
                });
    })


    var newSecret =  {
        "id" : Config.org1.id,
        "name": Config.org1.name,
        "ca_org": Config.org1.ca_org,
        "data": {},
        "status": "active",
        "secret": "newsecret",
        "email": "org1@email.com",
    };

    it('Trying to update secret', function(done) {
        chai.request(Config.server).put('omr/api/v1/orgs/' + Config.org1.id)
                .set('Accept', 'application/json')
                .set('token', Config.orgAdminToken1)
                .send(newSecret)
                .end(function(err, res){             
                    expect(err).to.be.null;
                    expect(res).to.have.property("text");
                    expect(JSON.parse(res.text)).to.have.property("msg");
                    expect(JSON.parse(res.text).msg).to.equal("Org admin's secret cannot be changed");
                    expect(res.status).to.equal(400);
                    done();
                });
    })

    var unregisteredId  = "notAnOrg"
    var newOrg1 = {
        "id": unregisteredId,
        "name": Config.org1.name,
        "ca_org": Config.org1.ca_org,
        "data": {},
        "status": "active",
        "email": "org1@email.com",
    };

    it('Updating nonexistent Org', function(done) {
        chai.request(Config.server).put('omr/api/v1/orgs/' + unregisteredId)
                .set('Accept', 'application/json')
                .set('token', Config.orgAdminToken1)
                .send(newOrg1)
                .end(function(err, res){             
                    expect(err).to.be.null;
                    expect(res).to.have.property("text");
                    expect(JSON.parse(res.text)).to.have.property("msg");
                    expect(JSON.parse(res.text).msg).to.equal("Org not found");
                    expect(res.status).to.equal(404);
                    done();
                });
    })

});
