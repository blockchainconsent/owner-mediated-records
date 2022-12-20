const Config = require('../../ConfigFile.js');
const chai = require('chai');
const { expect } = chai;
const chaiHttp = require('chai-http');
chai.use(chaiHttp);

require('../../Login/loginSysAdmin');

describe('Reregistering existing org', function(){
    require('./RegisterOrg1');
    let invalidOrg = {
        "id": Config.org1.id,
        "secret": "invalidpass",
        "name": "invalidname",
        "ca_org": "caOrg",
        "email": "org@email.com",
        "status": "active"
    };

    it('Registering Org with Same ID: Should return a 200 test response', function(done) {
        chai.request(Config.server).post('omr/api/v1/orgs')
        .set('Accept',  'application/json')
        .set('token', Config.sysAdminToken)
        .send(invalidOrg)
        .end(function(err, res) {
            expect(err).to.be.null
            expect(res).to.have.property("text");
            expect(res.text).to.equal('Org was already registered');

            //TODO: confirm if this status code is correct
            //expect(res.status).to.equal(200);
            
            done();
        });
    });

    it('Reregistering Existing Org: Should return a 200 test response', function(done) {
        chai.request(Config.server).post('omr/api/v1/orgs')
        .set('Accept',  'application/json')
        .set('token', Config.sysAdminToken)
        .send(Config.org1)
        .end(function(err, res) {
            expect(err).to.be.null
            expect(res).to.have.property("text");
            expect(res.text).to.equal('Org was already registered');
            expect(res.status).to.equal(200);
            done();
        });
    });
});

//Test to Register Orgs w/ Missing Data
//Expected: Test should pass and errors should be handled
describe('Registering Invalid Orgs', function () {
    describe('Registering Org Missing ID', function(){
        it('Should return a 400 test response', function(done) {
            let missingID = {
                "secret": "invalidpass",
                "name": "invalidname",
                "ca_org": "caOrg",
                "email": "org@email.com",
                "status": "active"
            };
            
                chai.request(Config.server).post('omr/api/v1/orgs')
                .set('Accept',  'application/json')
                .set('token', Config.sysAdminToken)
                .send(missingID)
                .end(function(err, res) {
                    expect(err).to.be.null
                    expect(res).to.have.property("text");
                    expect(JSON.parse(res.text)).to.have.property("msg");
                    expect(JSON.parse(res.text).msg).to.equal("Invalid data:id missing");
                    expect(res.status).to.equal(400);
                    done();
                });
        });
    });

    describe('Registering Org Missing CA org', function(){
        it('Should return a 400 test response', function(done) {
            let noCAOrg = {
                "id": "id",
                "secret": "invalidpass",
                "name": "invalidname",
                "email": "org@email.com",
                "status": "active"
            };
            
                chai.request(Config.server).post('omr/api/v1/orgs')
                .set('Accept',  'application/json')
                .set('token', Config.sysAdminToken)
                .send(noCAOrg)
                .end(function(err, res) {
                    expect(err).to.be.null
                    expect(res).to.have.property("text");
                    expect(JSON.parse(res.text)).to.have.property("msg");
                    expect(JSON.parse(res.text).msg).to.equal("Invalid data:ca_org missing");
                    expect(res.status).to.equal(400);
                    done();
                });
        });
    });

    describe('Registering Org Missing Email', function(){
        it('Should return a 400 test response', function(done) {
            let noEmail = {
                "id": "id",
                "secret": "invalidpass",
                "name": "invalidname",
                "ca_org": "caOrg",
                "status": "active"
            };
            
                chai.request(Config.server).post('omr/api/v1/orgs')
                .set('Accept',  'application/json')
                .set('token', Config.sysAdminToken)
                .send(noEmail)
                .end(function(err, res) {
                    expect(err).to.be.null
                    expect(res).to.have.property("text");
                    expect(JSON.parse(res.text)).to.have.property("msg");
                    expect(JSON.parse(res.text).msg).to.equal("Invalid data:email missing");
                    expect(res.status).to.equal(400);
                    done();
                });  
        });
    });
});
