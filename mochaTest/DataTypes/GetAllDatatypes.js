const Config = require('../ConfigFile.js');
const chai = require('chai');
const { expect } = chai;
const chaiHttp = require('chai-http')
chai.use(chaiHttp);
chai.use(require('chai-like'));
const chaiThings = require('chai-things')
chai.use(chaiThings)

let datatype1, datatype2, datatype3, datatype4;

require("../Login/loginSysAdmin");
require('./RegisterDatatypesSysAdmin');
describe('Getting All Registered Datatypes As SysAdmin', function() {
    before(function() {
        datatype1 = {
            "datatype_id": Config.datatype1.id,
            "description": Config.datatype1.description
        };
        datatype2 = {
            "datatype_id": Config.datatype2.id,
            "description": Config.datatype2.description
        };
        datatype3 = {
            "datatype_id": Config.datatype3.id,
            "description": Config.datatype3.description
        };
        datatype4 = {
            "datatype_id": Config.datatype4.id,
            "description": Config.datatype4.description
        };
    });

    it('Should contain Registered Datatypes', function (done) {
        chai.request(Config.server).get('omr/api/v1/datatypes')
                .set('Accept', 'application/json')
                .set('token', Config.sysAdminToken)
                .end(function(err, res){              
                    expect(err).to.be,null;
                    expect(res.status).to.equal(200);
                    expect(res.body).to.be.an('array');
                    expect(res.body).to.contain.something.like(datatype1);
                    expect(res.body).to.contain.something.like(datatype2);
                    expect(res.body).to.contain.something.like(datatype3);
                    expect(res.body).to.contain.something.like(datatype4);
                    done();
                });
    })
});

require("../Organizations/OrgSetup/RegisterOrg1")
require("../Login/OrgLogin/Org1Login")
describe('Getting All Registered Datatypes As OrgAdmin', function() {
    it('Should contain Registered Datatypes', function (done) {
        chai.request(Config.server).get('omr/api/v1/datatypes')
                .set('Accept', 'application/json')
                .set('token', Config.orgAdminToken1)
                .end(function(err, res){              
                    expect(err).to.be,null;
                    expect(res.status).to.equal(200);
                    expect(res.body).to.be.an('array');
                    expect(res.body).to.contain.something.like(datatype1);
                    expect(res.body).to.contain.something.like(datatype2);
                    expect(res.body).to.contain.something.like(datatype3);
                    expect(res.body).to.contain.something.like(datatype4);
                    done();
                });
    })
});