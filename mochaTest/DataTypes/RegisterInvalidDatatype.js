var Config = require('../ConfigFile.js');
const chai = require('chai');
const { expect } = chai;
const chaiHttp = require('chai-http');
const { Logger } = require('log4js/lib/logger');
chai.use(chaiHttp);

require('../Login/loginSysAdmin');
require('./RegisterDatatypesSysAdmin.js');

describe('Registering Datatype with Same ID', function(){
    it('Should return a 400 test response', function(done) {
        let sameID = {
            "id": Config.datatype1.id,
            "description": "random"};

        chai.request(Config.server).post('omr/api/v1/datatypes')
        .set('Accept',  'application/json')
        .set('token', Config.sysAdminToken)
        .send(sameID)
        .end(function(err, res) {
            expect(err).to.be.null;
            expect(res.body).to.not.be.null;
            expect(res.body).to.have.property("msg");
            expect(res.body.msg).to.equal("Existing datatype with same id found");
            expect(res.status).to.equal(400);
            done();
        });
    })
})

//Test to Register Datatypes w/ Missing Data
//Expected: Test should pass and errors should be handled
describe('Registering Invalid Datatypes', function () {
    describe('Registering Datatypes Missing ID', function(){
        it('Should return a 400 test response', function(done) {
            let missingID = {
                "description" : "random"
            }
            
            chai.request(Config.server).post('omr/api/v1/datatypes')
                .set('Accept',  'application/json')
                .set('token', Config.sysAdminToken)
                .send(missingID)
                .end(function(err, res) {
                    expect(err).to.be.null;
                    expect(res).to.have.property("text");
                    expect(JSON.parse(res.text)).to.have.property("msg");
                    expect(JSON.parse(res.text).msg).to.equal("Invalid datatype:id missing");
                    expect(res.status).to.equal(400);
                    done();
                });
            
        })
    })

    describe('Registering Datatype with Missing Description', function(){
        it('Should return a 400 test response', function(done) {
            let noDescription = {
                "id": "random",
            };
            chai.request(Config.server).post('omr/api/v1/datatypes')
                .set('Accept',  'application/json')
                .set('token', Config.sysAdminToken)
                .send(noDescription)
                .end(function(err, res) {
                    expect(err).to.be.null
                    expect(res).to.have.property("text");
                    expect(JSON.parse(res.text)).to.have.property("msg");
                    expect(JSON.parse(res.text).msg).to.equal("Invalid datatype:description missing");
                    expect(res.status).to.equal(400);
                    done();
                });
            
        })
    })
})