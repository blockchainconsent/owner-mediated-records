var Config = require('../../ConfigFile.js');
const chai = require('chai');
const { expect } = chai;
const chaiHttp = require('chai-http')
chai.use(chaiHttp);

require("../../DataTypes/RegisterDatatypesSysAdmin")
require("../RegisterUsers");
require("../../Login/User1Login");

describe('Update Datatype as Role User', function() {
    let newData;

    before(function() {
        newData = {
            "id" : Config.datatype1.id,
            "description" : "new description"
        };
    });
    
    it('Should fail and return a 500 test response', function(done) {
        chai.request(Config.server).put('omr/api/v1/datatypes/' + Config.datatype1.id) 
            .set('Accept',  'application/json')
            .set('token', Config.user1Token)
            .send(newData)
            .end(function(err, res) {
                expect(err).to.be.null
                expect(res.status).to.equal(500);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("error");
                done();
            });
    });

    it('Getting Datatype should have original description', function(done) {
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
    });

    after(function(done) {
        require("../../Login/loginSysAdmin")
        chai.request(Config.server).put('omr/api/v1/datatypes/' + Config.datatype1.id) 
            .set('Accept',  'application/json')
            .set('token', Config.sysAdminToken)
            .send(Config.datatype1)
            .end(function(err, res) {
                expect(err).to.be.null
                expect(res.status).to.equal(200);
                done();
            });
    })

})