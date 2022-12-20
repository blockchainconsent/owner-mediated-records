var Config = require('../ConfigFile.js');
const chai = require('chai');
const { expect } = chai;
const chaiHttp = require('chai-http')
chai.use(chaiHttp);

let oldData, newData;

require("../DataTypes/RegisterDatatypesSysAdmin")
require("../Login/loginSysAdmin");
describe('Update Datatype as SysAdmin', function() {
    before(function() {
        oldData = Config.datatype1;
        newData = {
            "id" : Config.datatype1.id,
            "description" : "new description"
        };
    });

    it('Should return a 200 test response', function(done) {
        chai.request(Config.server).put('omr/api/v1/datatypes/' + Config.datatype1.id) 
            .set('Accept',  'application/json')
            .set('token', Config.sysAdminToken)
            .send(newData)
            .end(function(err, res) {
                expect(err).to.be.null
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("success");
                done();
            });
    });

    it('Should update datatype description', function(done) {
        chai.request(Config.server).get('omr/api/v1/datatypes/' + Config.datatype1.id) 
            .set('Accept',  'application/json')
            .set('token', Config.sysAdminToken)
            .send(newData)
            .end(function(err, res) {
                expect(err).to.be.null
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("datatype_id");
                expect(res.body.datatype_id).to.equal(newData.id);
                expect(res.body).to.have.property("description");
                expect(res.body.description).to.equal(newData.description);
                done();
            });
    });

    //to restore datatype back to original description
    after(function(done){
        chai.request(Config.server).put('omr/api/v1/datatypes/' + Config.datatype1.id) 
            .set('Accept',  'application/json')
            .set('token', Config.sysAdminToken)
            .send(oldData)
            .end(function(err, res) {
                expect(err).to.be.null
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("success");
                done();
            });
    });
});


require('../Organizations/OrgSetup/RegisterOrg1')
require('../Login/OrgLogin/Org1Login')
describe('Update Datatype as OrgAdmin', function() {
    it('Should return a 200 test response', function(done) {
        chai.request(Config.server).put('omr/api/v1/datatypes/' + Config.datatype1.id) 
            .set('Accept',  'application/json')
            .set('token', Config.orgAdminToken1)
            .send(newData)
            .end(function(err, res) {
                expect(err).to.be.null
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("success");
                done();
            });
    });

    it('Should update datatype description', function(done) {
        chai.request(Config.server).get('omr/api/v1/datatypes/' + Config.datatype1.id) 
            .set('Accept',  'application/json')
            .set('token', Config.orgAdminToken1)
            .send(newData)
            .end(function(err, res) {
                expect(err).to.be.null
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("datatype_id");
                expect(res.body.datatype_id).to.equal(newData.id);
                expect(res.body).to.have.property("description");
                expect(res.body.description).to.equal(newData.description);
                done();
            });
    });

    //to restore datatype back to original description  
    after(function(done){
        chai.request(Config.server).put('omr/api/v1/datatypes/' + Config.datatype1.id) 
            .set('Accept',  'application/json')
            .set('token', Config.orgAdminToken1)
            .send(oldData)
            .end(function(err, res) {
                expect(err).to.be.null
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("success");
                done();
            });
    });
});