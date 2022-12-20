var Config = require('../ConfigFile.js');
const chai = require('chai');
const { expect } = chai;
const chaiHttp = require('chai-http')
chai.use(chaiHttp);


require("../Organizations/OrgSetup/RegisterOrg1")
require("../Organizations/OrgSetup/RegisterOrg2")
require("./RegisterServiceOrg1Admin");
require("./RegisterServiceOrg2Admin");
require("../Login/OrgLogin/Org2Login");
require("../Login/OrgLogin/Org1Login");

var invalidService = {
    "id": "invalid",
    "name": "invalid service",
    "secret": "pass",
    "ca_org": Config.caOrg,
    "email": "serviceemail@example.com",
    "org_id": Config.org2.id,
    "summary": "invalid service",
    "terms": {},
    "payment_required": "no",
    "datatypes": [
        {
            "datatype_id": Config.datatype1.id,
            "access":[
                "write",
                "read"
            ]
        }
    ],
    "solution_private_data": {
        "sample data 1": "sample data 1",
        "sample data 2": "sample data 2"

    }
  };

describe('Register Service without ID', function () {
    var serviceID = invalidService.id;

    it('Should return a 400 test response', function (done) { 
        delete invalidService.id; 
        chai.request(Config.server).post('omr/api/v1/services') 
            .set('Accept',  'application/json')
            .set('token', Config.orgAdminToken2)
            .send(invalidService)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(400);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.equal("Invalid data:id missing");
                done();
            });
        
    });

    //reset the service object
    after(function() {
        invalidService.id = serviceID;
    })
});

describe('Register Service without Name', function () {
    var serviceName = invalidService.name;

    it('Should return a 400 test response', function (done) { 
        delete invalidService.name; 
        chai.request(Config.server).post('omr/api/v1/services') 
            .set('Accept',  'application/json')
            .set('token', Config.orgAdminToken2)
            .send(invalidService)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(400);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.equal("Invalid data:name missing");
                done();
            });
        
    });

    //reset the service object
    after(function() {
        invalidService.name = serviceName;
    })
});

describe('Register Service without Secret', function () {
    var serviceSecret = invalidService.secret;

    it('Should return a 400 test response', function (done) { 
        delete invalidService.secret; 
        chai.request(Config.server).post('omr/api/v1/services') 
            .set('Accept',  'application/json')
            .set('token', Config.orgAdminToken2)
            .send(invalidService)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(400);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.equal("Invalid data:secret missing");
                done();
            });
        
    });

    //reset the service object
    after(function() {
        invalidService.secret = serviceSecret;
    })
});

describe('Register Service without CA Org', function () {
    var serviceCA = invalidService.ca_org;

    it('Should return a 400 test response', function (done) {  
        delete invalidService.ca_org;
        chai.request(Config.server).post('omr/api/v1/services') 
            .set('Accept',  'application/json')
            .set('token', Config.orgAdminToken2)
            .send(invalidService)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(400);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.equal("Invalid data:ca_org missing");
                done();
            });
        
    });

    //reset the service object
    after(function() {
        invalidService.ca_org = serviceCA;
    })
});

describe('Register Service without Email', function () {
    var serviceEmail = invalidService.email;

    it('Should return a 400 test response', function (done) { 
        delete invalidService.email; 
        chai.request(Config.server).post('omr/api/v1/services') 
            .set('Accept',  'application/json')
            .set('token', Config.orgAdminToken2)
            .send(invalidService)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(400);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.equal("Invalid data:email missing");
                done();
            });
        
    });

    //reset the service object
    after(function() {
        invalidService.email = serviceEmail;
    })
});

describe('Register Service without OrgID', function () {
    var orgID = invalidService.org_id;

    it('Should return a 400 test response', function (done) { 
        delete invalidService.org_id; 
        chai.request(Config.server).post('omr/api/v1/services') 
            .set('Accept',  'application/json')
            .set('token', Config.orgAdminToken2)
            .send(invalidService)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(400);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.equal("Invalid data:org_id missing");
                done();
            });
        
    });

    //reset the service object
    after(function() {
        invalidService.org_id = orgID;
    })
});


describe('Register Service without Summary', function () {
    var serviceSummary = invalidService.summary;

    it('Should return a 400 test response', function (done) { 
        delete invalidService.summary; 
        chai.request(Config.server).post('omr/api/v1/services') 
            .set('Accept',  'application/json')
            .set('token', Config.orgAdminToken2)
            .send(invalidService)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(400);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.equal("Invalid data: summary missing");
                done();
            });
        
    });

    //reset the service object
    after(function() {
        invalidService.summary = serviceSummary;
    })
});

describe('Register Service without Payment', function () {
    var servicePayment = invalidService.payment_required;

    it('Should return a 400 test response', function (done) {  
        delete invalidService.payment_required;
        chai.request(Config.server).post('omr/api/v1/services') 
            .set('Accept',  'application/json')
            .set('token', Config.orgAdminToken2)
            .send(invalidService)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(400);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.equal("Invalid data: payment_required must be either 'yes' or 'no'");
                done();
            });
        
    });

    //reset the service object
    after(function() {
        invalidService.payment_required = servicePayment;
    })
});

describe('Register Service without Datatype Array', function () {
    var dataArray = JSON.stringify(invalidService.datatypes);

    it('Should return a 400 test response', function (done) {  
        delete invalidService.datatypes;
        chai.request(Config.server).post('omr/api/v1/services') 
            .set('Accept',  'application/json')
            .set('token', Config.orgAdminToken2)
            .send(invalidService)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(400);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.equal("Invalid data: must have a list of datatypes");
                done();
            });
        
    });

    //reset the service object
    after(function() {
        invalidService.datatypes = dataArray;
    })
});

describe('Register Service with Empty Datatype Array', function () {
    var dataArray = JSON.stringify(invalidService.datatypes);

    it('Should return a 400 test response', function (done) {  
        invalidService.datatypes = [];
        chai.request(Config.server).post('omr/api/v1/services') 
            .set('Accept',  'application/json')
            .set('token', Config.orgAdminToken2)
            .send(invalidService)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(400);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.equal("Invalid data: must include at least one datatype");
                done();
            });
        
    });

    //reset the service object
    after(function() {
        invalidService.datatypes = dataArray;
    })
});

describe('Register Service with Same ID as a Service in Another Org', function () {

    var sameID = {
        "id": Config.service1.id,
        "name": "service1copy",
        "secret": "pass",
        "ca_org": Config.caOrg,
        "email": "serviceemail@example.com",
        "org_id": Config.org2.id,
        "summary": "service with same ID as registered service 1",
        "terms": {},
        "payment_required": "no",
        "datatypes": [
            {
                "datatype_id": Config.datatype1.id,
                "access":[
                    "write",
                    "read"
                ]
            }
        ],
        "solution_private_data": {
        }
      };

    it('Should return a 400 test response', function (done) {  
        chai.request(Config.server).post('omr/api/v1/services') 
            .set('Accept',  'application/json')
            .set('token', Config.orgAdminToken2)
            .send(sameID)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(400);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.equal("Existing service with same id found");
                done();
            });
        
    });
});

describe('Register Service with Datatype Missing Read Write', function () {

    var sameID = {
        "id": "invalidOrg",
        "name": "invalid service",
        "secret": "invalid pass",
        "ca_org": Config.caOrg,
        "email": "serviceemail@example.com",
        "org_id": Config.org2.id,
        "summary": "invalid service",
        "terms": {},
        "payment_required": "no",
        "datatypes": [
            {
                "datatype_id": Config.datatype1.id,
            }
        ],
        "solution_private_data": {
        }
      };

    it('Should return a 500 test response', function (done) {  
        chai.request(Config.server).post('omr/api/v1/services') 
            .set('Accept',  'application/json')
            .set('token', Config.orgAdminToken2)
            .send(sameID)
            .end(function(err, res) {
                expect(err).to.be.null;

                expect(res.status).to.equal(500);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.equal("Service is registered to CA, but failed to register service in Blockchain:Failed to register service");
                done();
            });
        
    });
});
