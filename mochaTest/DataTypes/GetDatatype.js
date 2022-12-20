var Config = require('../ConfigFile.js');
const chai = require('chai');
const { expect } = chai;
const chaiHttp = require('chai-http')
chai.use(chaiHttp);

require('../Login/loginSysAdmin');
require('./RegisterDatatypesSysAdmin')

//Checks the datatype returned matches the ID and description of what was registered
describe('Getting Registered Datatypes as SysAdmin', function() {
    it('Should contain datatype1', function (done) {
        chai.request(Config.server).get('omr/api/v1/datatypes/' + Config.datatype1.id)
                .set('Accept', 'application/json')
                .set('token', Config.sysAdminToken)
                .end(function(err, res){                
                    expect(err).to.be.null;
                    expect(res.status).to.equal(200);
                    expect(res.body).to.have.property("datatype_id");
                    expect(res.body.datatype_id).to.equal(Config.datatype1.id);
                    expect(res.body).to.have.property("description");
                    expect(res.body.description).to.equal(Config.datatype1.description);
                    done();
                });
    })

    it('Should contain datatype2', function (done) {
        chai.request(Config.server).get('omr/api/v1/datatypes/' + Config.datatype2.id)
                .set('Accept', 'application/json')
                .set('token', Config.sysAdminToken)
                .end(function(err, res){                
                    expect(err).to.be.null;
                    expect(res.status).to.equal(200);
                    expect(res.body).to.have.property("datatype_id");
                    expect(res.body.datatype_id).to.equal(Config.datatype2.id);
                    expect(res.body).to.have.property("description");
                    expect(res.body.description).to.equal(Config.datatype2.description);
                    done();
                });
    })

});

require("../Organizations/OrgSetup/RegisterOrg1")
require("../Login/OrgLogin/Org1Login")
describe('Getting Registered Datatypes as OrgAdmin', function() {
    it('Should contain datatype1', function (done) {
        chai.request(Config.server).get('omr/api/v1/datatypes/' + Config.datatype1.id)
                .set('Accept', 'application/json')
                .set('token', Config.orgAdminToken1)
                .end(function(err, res){                
                    expect(err).to.be.null;
                    expect(res.status).to.equal(200);
                    expect(res.body).to.have.property("datatype_id");
                    expect(res.body.datatype_id).to.equal(Config.datatype1.id);
                    expect(res.body).to.have.property("description");
                    expect(res.body.description).to.equal(Config.datatype1.description);
                    done();
                });
    })

    it('Should contain datatype2', function (done) {
        chai.request(Config.server).get('omr/api/v1/datatypes/' + Config.datatype2.id)
                .set('Accept', 'application/json')
                .set('token', Config.orgAdminToken1)
                .end(function(err, res){                
                    expect(err).to.be.null;
                    expect(res.status).to.equal(200);
                    expect(res.body).to.have.property("datatype_id");
                    expect(res.body.datatype_id).to.equal(Config.datatype2.id);
                    expect(res.body).to.have.property("description");
                    expect(res.body.description).to.equal(Config.datatype2.description);
                    done();
                });
    })

});

describe('Getting Datatypes that is Unregistered', function() {
    var unregistered = "unregistered";

    it('Should return an empty body', function (done) {
        chai.request(Config.server).get('omr/api/v1/datatypes/' + unregistered)
                .set('Accept', 'application/json')
                .set('token', Config.sysAdminToken)
                .end(function(err, res){                
                    expect(err).to.be.null;
                    expect(res.status).to.equal(200);
                    expect(res.body).to.have.property("datatype_id");
                    expect(res.body.datatype_id).to.equal("");
                    expect(res.body).to.have.property("description");
                    expect(res.body.description).to.equal("");
                    done();
                });
    })

});