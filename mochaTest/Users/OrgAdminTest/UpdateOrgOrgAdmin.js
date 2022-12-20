const chai = require('chai');
const { expect } = chai;
const chaiHttp = require('chai-http')
chai.use(chaiHttp);
chai.use(require('chai-like'));
const chaiThings = require('chai-things')
chai.use(chaiThings)

const Config = require('../../ConfigFile');
require('../../Organizations/OrgSetup/RegisterOrg1');
require('../../Organizations/OrgSetup/RegisterOrg2');
require("../RegisterUsers");
require("../../Login/User2Login");
require("../../Login/OrgLogin/Org1Login");

describe('Updating Org as Org Admin:', function() {
    it('Giving User2 Org Admin Permission', function (done) {  
        chai.request(Config.server).put('omr/api/v1/users/' + Config.user2.id + '/permissions/admin/' + Config.org1.id) 
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

    var oldName = Config.org1.name;
    var oldEmail = Config.org1.email;
    var oldData = Config.org1.data;

    var newOrg1 = {
        "id": Config.org1.id,
        "name": "newOrg1Name",
        "ca_org": Config.org1.ca_org,
        "status": "active",
        "email" : "newOrgEmail@org.com"
    };

    it('Updating Org1: Should return a 200 test response', function(done) {
        chai.request(Config.server).put('omr/api/v1/orgs/' + Config.org1.id)
                .set('Accept', 'application/json')
                .set('token', Config.user2Token)
                .send(newOrg1)
                .end(function(err, res){             
                    expect(err).to.be.null;
                    expect(res.status).to.equal(200);
                    expect(res.body).to.have.property("msg");
                    expect(res.body.msg).to.include("success");
                    done();
                });
    })

    it('Getting Org1 Details: Should return updated data', function (done) {
        chai.request(Config.server).get('omr/api/v1/orgs/' + Config.org1.id)
                .set('Accept', 'application/json')
                .set('token', Config.orgAdminToken1)
                .set('org_id', Config.org1.id)
                .end(function(err, res){                
                    expect(err).to.be.null
                    expect(res.status).to.equal(200);
                    expect(res.body.name).to.equal(newOrg1.name);
                    expect(res.body.email).to.equal(newOrg1.email);
                    done();
                });
    })

    it('Remove User2 Org Admin Permission', function (done) {  
        chai.request(Config.server).delete('omr/api/v1/users/' + Config.user2.id + '/permissions/admin/' + Config.org1.id) 
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

    it('Updating Org1: Should return a 500 test response', function(done) {
        chai.request(Config.server).put('omr/api/v1/orgs/' + Config.org1.id)
                .set('Accept', 'application/json')
                .set('token', Config.user2Token)
                .send(newOrg1)
                .end(function(err, res){             
                    expect(err).to.be.null;
                    expect(res.status).to.equal(500);
                    expect(res.body).to.have.property("msg");
                    expect(res.body.msg).to.equal("Org is registered to CA, but failed to register org in Blockchain:Failed to register org");
                    done();
                });
    })


    //reset org data by updating as original org admin 
    after(function(done) {
        var resetOrg1 = {
            "id": Config.org1.id,
            "name": oldName,
            "ca_org": Config.org1.ca_org,
            "data": oldData,
            "status": "active",
            "email" : oldEmail
        }
    
        chai.request(Config.server).put('omr/api/v1/orgs/' + Config.org1.id)
                .set('Accept', 'application/json')
                .set('org_id', Config.org1.id)
                .set('token', Config.orgAdminToken1)
                .send(resetOrg1)
                .end(function(err, res){                
                    expect(err).to.be.null;
                    expect(res.status).to.equal(200);
                    expect(res.body).to.have.property("msg");
                    expect(res.body.msg).to.include("success");
                    done();
                    
                });
    })

});

