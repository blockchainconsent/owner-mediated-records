var Config = require('../ConfigFile.js');
const chai = require('chai');
const { expect } = chai;
const chaiHttp = require('chai-http')
chai.use(chaiHttp);
chai.use(require('chai-like'));
const chaiThings = require('chai-things')
chai.use(chaiThings)

const idGenerator = require('../Utils/helper');

require("./RegisterServiceOrg2Admin.js")
require("../Login/OrgLogin/Org2Login");

//Adds datatypes to service 3
describe('Attempt to add datatype to service more than once', function () {
    const datatypeID = idGenerator();
    const datatypeToAdd = {
        "datatype_id": datatypeID,
        "access": [
          "read",
          "write"
        ]
    };
    const datatypeToRegister = {
        "id": datatypeID,
        "description": "description"
    };

    before((done) => {
        chai.request(Config.server).post('omr/api/v1/datatypes') 
            .set('Accept',  'application/json')
            .set('token', Config.orgAdminToken2)
            .send(datatypeToRegister)
            .end(function(err, res) {
                expect(err).to.be.null
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("success");
                done();
            });
    });

    it('Should return a 200 test response', function (done) {                 
        chai.request(Config.server).post('omr/api/v1/services/' + Config.service3.id + '/addDatatype/' + datatypeID) 
            .set('Accept',  'application/json')
            .set('token', Config.orgAdminToken2)
            .send(datatypeToAdd)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("success");
                done();
            });
        
    });

    it('Service should include new datatype', function (done) { 
        const expectedDatatype = {
            "service_id": Config.service3.id,
            "datatype_id": datatypeID,
        }

        chai.request(Config.server).get('omr/api/v1/services/' + Config.service3.id) 
            .set('Accept',  'application/json')
            .set('token', Config.orgAdminToken2)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body.datatypes).to.contain.something.like(expectedDatatype);
                done();
            }); 
        
    });

    it('Should return a 500 test response, datatype was already added to service', function(done) {                
        // datatype has already been added to the service
        chai.request(Config.server).post('omr/api/v1/services/' + Config.service3.id + '/addDatatype/' + datatypeID) 
            .set('Accept',  'application/json')
            .set('token', Config.orgAdminToken2)
            .send(datatypeToAdd)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(500);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("error");
                done();
            });
    });

});