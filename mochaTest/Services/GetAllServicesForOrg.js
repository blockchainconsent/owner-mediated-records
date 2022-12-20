const Config = require('../ConfigFile.js');
const chai = require('chai');
const { expect } = chai;
const chaiHttp = require('chai-http');
chai.use(chaiHttp);
chai.use(require('chai-like'));
const chaiThings = require('chai-things');
chai.use(chaiThings);

require("../Organizations/OrgSetup/RegisterOrg2");
require("./RegisterServiceOrg1Admin");
require("../Login/OrgLogin/Org1Login");
require("../Login/loginSysAdmin");
require("../Login/OrgLogin/Org2Login");

let service1, service2;

describe('Get Org1 Services as Org1 Admin', function () {
    before(function() {
        service1 = {
            "service_id": Config.service1.id,
            "service_name": Config.service1.name,
            "org_id": Config.org1.id,
            "summary": Config.service1.summary,
            "terms": Config.service1.terms,
            "payment_required": Config.service1.payment_required
        };
        service2 = {
            "service_id": Config.service2.id,
            "service_name": Config.service2.name,
            "org_id": Config.org1.id,
            "summary": Config.service2.summary,
            "terms": Config.service2.terms,
            "payment_required": Config.service2.payment_required
        };
    });

    it('Should return a 200 response and show private data', function (done) {  
        chai.request(Config.server).get('omr/api/v1/orgs/' + Config.org1.id + '/services') 
            .set('Accept',  'application/json')
            .set('token', Config.orgAdminToken1)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.be.an('array');
                expect(res.body).to.contain.something.like(service1);
                expect(res.body).to.contain.something.like(service2);
                const email = res.body[0].email;
                const isExpectedEmail = (email == Config.service1.email) || (email == Config.service2.email);
                expect(isExpectedEmail);
                const privateData = JSON.stringify(res.body[0].solution_private_data);
                const isExpectedPrivateData = (privateData == JSON.stringify(Config.service1.solution_private_data))
                || (privateData == JSON.stringify(Config.service2.solution_private_data));
                expect(isExpectedPrivateData);
                done();
            });
    });
});

require("../Login/OrgLogin/Org2Login");
describe('Get Org1 Services as Org2 Admin', function () {

    it('Should return a 200 response and hide private data', function (done) {  
        chai.request(Config.server).get('omr/api/v1/orgs/' + Config.org1.id + '/services') 
            .set('Accept',  'application/json')
            .set('token', Config.orgAdminToken2)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.be.an('array');
                expect(res.body).to.contain.something.like(service1);
                expect(res.body).to.contain.something.like(service2);
                expect(res.body[0].email).to.equal("");
                expect(res.body[0].solution_private_data).to.be.null;
                done();
            });
    });
});

require("../Login/loginSysAdmin");
describe('Get Org Services as SysAdmin', function () {

    it('Should return a 200 response and hide private data', function (done) {  
        chai.request(Config.server).get('omr/api/v1/orgs/' + Config.org1.id + '/services') 
            .set('Accept',  'application/json')
            .set('token', Config.sysAdminToken)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.be.an('array');
                expect(res.body).to.contain.something.like(service1);
                expect(res.body).to.contain.something.like(service2);
                expect(res.body[0].email).to.equal("");
                expect(res.body[0].solution_private_data).to.be.null;
                done();
            }); 
    });
});