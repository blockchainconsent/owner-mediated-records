const Config = require('../../ConfigFile.js');
const chai = require('chai');
const { expect } = chai;
const chaiHttp = require('chai-http');
chai.use(chaiHttp);
chai.use(require('chai-like'));
const chaiThings = require('chai-things');
chai.use(chaiThings);
const idGenerator = require('../../Utils/helper');

require('../../Login/loginSysAdmin');
require("../../Organizations/OrgSetup/RegisterOrg1");
require("../../Login/OrgLogin/Org1Login");

describe('Check of adding and revoking Service Admin Permission', function() {

    // register datatype1
    before((done) => {
        chai.request(Config.server).post('omr/api/v1/datatypes') 
            .set('Accept',  'application/json')
            .set('token', Config.orgAdminToken1)
            .send({id: idGenerator(), description: Config.datatype1.description})
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                Config.datatype1.id = res.body.id;
                done();
            });
    })

    // register datatype2
    before((done) => {
        chai.request(Config.server).post('omr/api/v1/datatypes') 
            .set('Accept',  'application/json')
            .set('token', Config.orgAdminToken1)
            .send({id: idGenerator(), description: Config.datatype2.description})
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                Config.datatype2.id = res.body.id;
                done();
            });
    })

    // register datatype3
    before((done) => {
        chai.request(Config.server).post('omr/api/v1/datatypes') 
            .set('Accept',  'application/json')
            .set('token', Config.orgAdminToken1)
            .send({id: idGenerator(), description: Config.datatype3.description})
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                Config.datatype3.id = res.body.id;
                done();
            });
    })

    // register datatype4
    before((done) => {
        chai.request(Config.server).post('omr/api/v1/datatypes') 
            .set('Accept',  'application/json')
            .set('token', Config.orgAdminToken1)
            .send({id: idGenerator(), description: Config.datatype4.description})
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                Config.datatype4.id = res.body.id;
                done();
            });
    })

    // register newService1
    let newService1 = {};    
    before((done) => {
        newService1 = {
            "id": idGenerator(),
            "name": "newService",
            "secret": "newservicepass",
            "ca_org": Config.caOrg,
            "email": "newserviceemail@example.com",
            "org_id": Config.org1.id,
            "summary": "New service under org 1. Has one datatype",
            "terms": {
                "term1" : "example term",
                "term2" : "example term",
            },
            "payment_required": "yes",
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
                "sample data 1": "service 3 sample data 1",
                "sample data 2": "service 3 sample data 2"
            }
        };

        chai.request(Config.server).post('omr/api/v1/services') 
            .set('Accept',  'application/json')
            .set('token', Config.orgAdminToken1)
            .send(newService1)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("success");
                done();
                });
    });

    // register newService2
    let newService2 = {};    
    before((done) => {
        newService2 = {
            "id": idGenerator(),
            "name": "newService2",
            "secret": "newservicepass",
            "ca_org": Config.caOrg,
            "email": "newservice2email@example.com",
            "org_id": Config.org1.id,
            "summary": "New service under org 1. Has two datatypes",
            "terms": {
                "term1" : "example term",
                "term2" : "example term",
            },
            "payment_required": "yes",
            "datatypes": [
                {
                    "datatype_id": Config.datatype2.id,
                    "access":[
                        "write",
                        "read"
                    ]
                },
                {
                    "datatype_id": Config.datatype3.id,
                    "access":[
                        "write"
                    ]
                }
            ],
            "solution_private_data": {
                "sample data 1": "service 3 sample data 1",
                "sample data 2": "service 3 sample data 2"
            }
        };

        chai.request(Config.server).post('omr/api/v1/services') 
            .set('Accept',  'application/json')
            .set('token', Config.orgAdminToken1)
            .send(newService2)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("success");
                done();
                });
    });

    // create newUser1 which doesn't belong to any org
    const newUser1 = {
        "id": idGenerator(),
        "secret": "newuser1",
        "name": "newUser1",
        "role": "user",
        "org" : "",
        "email":  "newuser1email@example.com",
        "ca_org": Config.caOrg,
        "data": {
          "address": "2 User St"
        }
    };

    // create newUser2 which belongs to org1
    const newUser2 = {
        "id": idGenerator(),
        "secret": "newuser2",
        "name": "newUser2",
        "role": "user",
        "org" : Config.org1.id,
        "email":  "newuser2email@example.com",
        "ca_org": Config.caOrg,
        "data": {
          "address": "2 User St"
        }
    };

    // create newUser3 which doesn't belong to any org
    const newUser3 = {
        "id": idGenerator(),
        "secret": "newuser3",
        "name": "newUser3",
        "role": "user",
        "org" : "",
        "email":  "newuser3email@example.com",
        "ca_org": Config.caOrg,
        "data": {
          "address": "2 User St"
        }
    };

    // create newUser4 which belongs to org1
    const newUser4 = {
        "id": idGenerator(),
        "secret": "newuser4",
        "name": "newUser4",
        "role": "user",
        "org" : Config.org1.id,
        "email":  "newuser4email@example.com",
        "ca_org": Config.caOrg,
        "data": {
          "address": "2 User St"
        }
    };
    
    // register new user1 (doesn't belong to any Org, with role user)
    it('Should register newUser1 with role user successfully', function (done) {
        chai.request(Config.server).post('omr/api/v1/users') 
            .set('Accept',  'application/json')
            .set('token', Config.sysAdminToken)
            .send(newUser1)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("success");
                done();
            });
    });

    // register new user2 (belongs to Org1, with role user)
    it('Should register newUser2 with role user successfully', function (done) {
        chai.request(Config.server).post('omr/api/v1/users') 
            .set('Accept',  'application/json')
            .set('token', Config.orgAdminToken1)
            .send(newUser2)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("success");
                done();
            });
    });

    // register new user3 (doesn't belong to any Org, with role user)
    it('Should register newUser3 with role user successfully', function (done) {
        chai.request(Config.server).post('omr/api/v1/users') 
            .set('Accept',  'application/json')
            .set('token', Config.sysAdminToken)
            .send(newUser3)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("success");
                done();
            });
    });

    // register new user4 (belongs to Org1, with role user)
    it('Should register newUser4 with role user successfully', function (done) {
        chai.request(Config.server).post('omr/api/v1/users') 
            .set('Accept',  'application/json')
            .set('token', Config.orgAdminToken1)
            .send(newUser4)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("success");
                done();
            });
    });

    // enroll user 1 to service1 as org admin
    it('Should successfully be enrolled to service 1', function (done) {
        const bodyRequest = {
            "user": newUser1.id,
            "service": newService1.id,
            "status": "active"
        }
        chai.request(Config.server).post('omr/api/v1/services/' + newService1.id + '/user/enroll')
            .set('Accept', 'application/json')
            .set('token', Config.orgAdminToken1)
            .send(bodyRequest)
            .end(function(err, res){           
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("success");
                done();
            });
    });

    // grant service admin permission to newUser4
    it('Should successfully grant service admin permission to user', function (done) {
        chai.request(Config.server).put('omr/api/v1/users/' + newUser4.id + '/permissions/services/' + newService1.id) 
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

    // revoke service admin permission from newUser4
    it('Should successfully revoke service admin permission from user', function (done) {
        chai.request(Config.server).delete('omr/api/v1/users/' + newUser4.id + '/permissions/services/' + newService1.id) 
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

    // grant service admin permission to newUser4 one more time
    it('Should successfully grant service admin permission to user one more time', function (done) {
        chai.request(Config.server).put('omr/api/v1/users/' + newUser4.id + '/permissions/services/' + newService1.id) 
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

    // login as service admin
    let serviceAdminToken = '';
    it('Should return a login token for service admin', function (done) {
        chai.request(Config.server).get('common/api/v1/login')
            .set('Accept', 'application/json')
            .set('user-id', newUser4.id)
            .set('password', newUser4.secret)
            .set('login-org', newUser4.ca_org)
            .set('login-channel', Config.channel)
            .set('signature', "")
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("token");
                expect(res.body.token).to.not.equal("");
                serviceAdminToken = res.body.token;
                done();
            });
    });

    // get the list of users enrolled to service1
    it('Should return a list of users enrolled to service1', function (done) {       
        chai.request(Config.server).get('omr/api/v1/services/' + newService1.id + '/enrollments')
            .set('Accept', 'application/json')
            .set('token', serviceAdminToken)
            .end(function(err, res){
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.be.an('array');
                expect(res.body).to.have.lengthOf(1);
                done();
            });
    });

    // revoke service admin permission from newUser4 one more time
    it('Should successfully revoke service admin permission from user one more time', function (done) {
        chai.request(Config.server).delete('omr/api/v1/users/' + newUser4.id + '/permissions/services/' + newService1.id) 
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

    // get the list of users enrolled to service1 without service admin permission
    it('Should return a blank list of users enrolled to service1 since newUser4 is not an service admin anymore', function (done) {
        chai.request(Config.server).get('omr/api/v1/services/' + newService1.id + '/enrollments')
            .set('Accept', 'application/json')
            .set('token', serviceAdminToken)
            .end(function(err, res){
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.be.an('array');
                expect(res.body).to.have.lengthOf(0);
                done();
            });
    });

});


describe('Check Service Admin Permission - add datatype to service/remove datatype from service', function() {

    // register datatype1
    before((done) => {
        chai.request(Config.server).post('omr/api/v1/datatypes') 
            .set('Accept',  'application/json')
            .set('token', Config.orgAdminToken1)
            .send({id: idGenerator(), description: Config.datatype1.description})
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                Config.datatype1.id = res.body.id;
                done();
            });
    })

    // register datatype2
    before((done) => {
        chai.request(Config.server).post('omr/api/v1/datatypes') 
            .set('Accept',  'application/json')
            .set('token', Config.orgAdminToken1)
            .send({id: idGenerator(), description: Config.datatype2.description})
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                Config.datatype2.id = res.body.id;
                done();
            });
    })

    // register datatype3
    before((done) => {
        chai.request(Config.server).post('omr/api/v1/datatypes') 
            .set('Accept',  'application/json')
            .set('token', Config.orgAdminToken1)
            .send({id: idGenerator(), description: Config.datatype3.description})
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                Config.datatype3.id = res.body.id;
                done();
            });
    })

    // register datatype4
    before((done) => {
        chai.request(Config.server).post('omr/api/v1/datatypes') 
            .set('Accept',  'application/json')
            .set('token', Config.orgAdminToken1)
            .send({id: idGenerator(), description: Config.datatype4.description})
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                Config.datatype4.id = res.body.id;
                done();
            });
    })

    // register newService1
    let newService1 = {};    
    before((done) => {
        newService1 = {
            "id": idGenerator(),
            "name": "newService",
            "secret": "newservicepass",
            "ca_org": Config.caOrg,
            "email": "newserviceemail@example.com",
            "org_id": Config.org1.id,
            "summary": "New service under org 1. Has one datatype",
            "terms": {
                "term1" : "example term",
                "term2" : "example term",
            },
            "payment_required": "yes",
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
                "sample data 1": "service 3 sample data 1",
                "sample data 2": "service 3 sample data 2"
            }
        };

        chai.request(Config.server).post('omr/api/v1/services') 
            .set('Accept',  'application/json')
            .set('token', Config.orgAdminToken1)
            .send(newService1)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("success");
                done();
                });
    });

    // register newService2
    let newService2 = {};    
    before((done) => {
        newService2 = {
            "id": idGenerator(),
            "name": "newService2",
            "secret": "newservicepass",
            "ca_org": Config.caOrg,
            "email": "newservice2email@example.com",
            "org_id": Config.org1.id,
            "summary": "New service under org 1. Has two datatypes",
            "terms": {
                "term1" : "example term",
                "term2" : "example term",
            },
            "payment_required": "yes",
            "datatypes": [
                {
                    "datatype_id": Config.datatype2.id,
                    "access":[
                        "write",
                        "read"
                    ]
                },
                {
                    "datatype_id": Config.datatype3.id,
                    "access":[
                        "write"
                    ]
                }
            ],
            "solution_private_data": {
                "sample data 1": "service 3 sample data 1",
                "sample data 2": "service 3 sample data 2"
            }
        };

        chai.request(Config.server).post('omr/api/v1/services') 
            .set('Accept',  'application/json')
            .set('token', Config.orgAdminToken1)
            .send(newService2)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("success");
                done();
                });
    });

    // create new user which belong to org1
    const newUser4 = {
        "id": idGenerator(),
        "secret": "newuser4",
        "name": "newUser4",
        "role": "user",
        "org" : Config.org1.id,
        "email":  "newuser4email@example.com",
        "ca_org": Config.caOrg,
        "data": {
          "address": "2 User St"
        }
    };
    
    // register newUser4
    it('Should register newUser4 with role user successfully', function (done) {
        chai.request(Config.server).post('omr/api/v1/users') 
            .set('Accept',  'application/json')
            .set('token', Config.orgAdminToken1)
            .send(newUser4)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("success");
                done();
            });
    });

    // grant service admin permission of newService1 to newUser4
    it('Should successfully grant service admin permission to user', function (done) {
        chai.request(Config.server).put('omr/api/v1/users/' + newUser4.id + '/permissions/services/' + newService1.id) 
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

    // login as service admin
    let serviceAdminToken = '';
    it('Should return a login token for service admin', function (done) {
        chai.request(Config.server).get('common/api/v1/login')
            .set('Accept', 'application/json')
            .set('user-id', newUser4.id)
            .set('password', newUser4.secret)
            .set('login-org', newUser4.ca_org)
            .set('login-channel', Config.channel)
            .set('signature', "")
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("token");
                expect(res.body.token).to.not.equal("");
                serviceAdminToken = res.body.token;
                done();
            });
    });

    // get existing service1 as service admin
    it('Should return an initial information about service1', function (done) {
        chai.request(Config.server).get('omr/api/v1/services/' + newService1.id)
            .set('Accept', 'application/json')
            .set('token', serviceAdminToken)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body.datatypes).to.be.an('array');
                expect(res.body.datatypes).to.have.lengthOf(1);
                expect(res.body.datatypes[0].datatype_id).to.equal(newService1.datatypes[0].datatype_id);
                done();
            });
    });

    // add datatype to the service1 as service admin
    it('Should add new datatype to the service1', function (done) {
        chai.request(Config.server).post('omr/api/v1/services/' + newService1.id + '/addDatatype/' + Config.datatype4.id) 
            .set('Accept',  'application/json')
            .set('token', serviceAdminToken)
            .send({
                "datatype_id": Config.datatype4.id,
                "access": [
                  "write"
                ]
            })
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body.msg).to.include("success");
                done();
            });
    });

    // get existing service1 as service admin
    it('Should return an information about service1', function (done) {
        chai.request(Config.server).get('omr/api/v1/services/' + newService1.id)
            .set('Accept', 'application/json')
            .set('token', serviceAdminToken)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body.datatypes).to.be.an('array');
                expect(res.body.datatypes).to.have.lengthOf(2);
                expect(res.body.datatypes[0].datatype_id).to.equal(newService1.datatypes[0].datatype_id);
                expect(res.body.datatypes[1].datatype_id).to.equal(Config.datatype4.id);
                done();
            });
    });

    // remove datatype from the service1 as service admin
    it('Should remove an appropriate datatype from the service1', function (done) {
        chai.request(Config.server).delete('omr/api/v1/services/' + newService1.id + '/removeDatatype/' + Config.datatype1.id) 
            .set('Accept',  'application/json')
            .set('token', serviceAdminToken)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("success");
                done();
            });
    });

    // get existing service1 as service admin
    it('Should return an updated information about service1', function (done) {
        chai.request(Config.server).get('omr/api/v1/services/' + newService1.id)
            .set('Accept', 'application/json')
            .set('token', serviceAdminToken)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body.datatypes).to.be.an('array');
                expect(res.body.datatypes).to.have.lengthOf(1);
                expect(res.body.datatypes[0].datatype_id).to.equal(Config.datatype4.id);
                done();
            });
    });
});

describe('Check Service Admin Permission - update service', function() {

    // register datatype1
    before((done) => {
        chai.request(Config.server).post('omr/api/v1/datatypes') 
            .set('Accept',  'application/json')
            .set('token', Config.orgAdminToken1)
            .send({id: idGenerator(), description: Config.datatype1.description})
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                Config.datatype1.id = res.body.id;
                done();
            });
    })

    // register datatype2
    before((done) => {
        chai.request(Config.server).post('omr/api/v1/datatypes') 
            .set('Accept',  'application/json')
            .set('token', Config.orgAdminToken1)
            .send({id: idGenerator(), description: Config.datatype2.description})
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                Config.datatype2.id = res.body.id;
                done();
            });
    })

    // register datatype3
    before((done) => {
        chai.request(Config.server).post('omr/api/v1/datatypes') 
            .set('Accept',  'application/json')
            .set('token', Config.orgAdminToken1)
            .send({id: idGenerator(), description: Config.datatype3.description})
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                Config.datatype3.id = res.body.id;
                done();
            });
    })

    // register datatype4
    before((done) => {
        chai.request(Config.server).post('omr/api/v1/datatypes') 
            .set('Accept',  'application/json')
            .set('token', Config.orgAdminToken1)
            .send({id: idGenerator(), description: Config.datatype4.description})
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                Config.datatype4.id = res.body.id;
                done();
            });
    })

    // register newService1
    let newService1 = {};    
    before((done) => {
        newService1 = {
            "id": idGenerator(),
            "name": "newService1",
            "secret": "newservicepass",
            "ca_org": Config.caOrg,
            "email": "newservice1email@example.com",
            "org_id": Config.org1.id,
            "summary": "New service under org 1. Has two datatypes",
            "terms": {
                "term1" : "example term",
                "term2" : "example term",
            },
            "payment_required": "yes",
            "datatypes": [
                {
                    "datatype_id": Config.datatype2.id,
                    "access":[
                        "write",
                        "read"
                    ]
                },
                {
                    "datatype_id": Config.datatype3.id,
                    "access":[
                        "write"
                    ]
                }
            ],
            "solution_private_data": {
                "sample data 1": "some data",
                "sample data 2": "some sata 2"
            }
        };

        chai.request(Config.server).post('omr/api/v1/services') 
            .set('Accept',  'application/json')
            .set('token', Config.orgAdminToken1)
            .send(newService1)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("success");
                done();
                });
    });

    // create new service admin user
    const newUser4 = {
        "id": idGenerator(),
        "secret": "newuser4",
        "name": "newUser4",
        "role": "user",
        "org" : Config.org1.id,
        "email":  "newuser4email@example.com",
        "ca_org": Config.caOrg,
        "data": {
          "address": "2 User St"
        }
    };
    
    // register newUser4
    it('Should register newUser4 with role user successfully', function (done) {
        chai.request(Config.server).post('omr/api/v1/users') 
            .set('Accept',  'application/json')
            .set('token', Config.orgAdminToken1)
            .send(newUser4)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("success");
                done();
            });
    });

    // grant service admin permission of newService1 to newUser4
    it('Should successfully grant service admin permission to user', function (done) {
        chai.request(Config.server).put('omr/api/v1/users/' + newUser4.id + '/permissions/services/' + newService1.id) 
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

    // login as service admin
    let serviceAdminToken = '';
    it('Should return a login token for service admin', function (done) {
        chai.request(Config.server).get('common/api/v1/login')
            .set('Accept', 'application/json')
            .set('user-id', newUser4.id)
            .set('password', newUser4.secret)
            .set('login-org', newUser4.ca_org)
            .set('login-channel', Config.channel)
            .set('signature', "")
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("token");
                expect(res.body.token).to.not.equal("");
                serviceAdminToken = res.body.token;
                done();
            });
    });

    // get existing service1 as service admin
    it('Should return an initial information about service1', function (done) {
        chai.request(Config.server).get('omr/api/v1/services/' + newService1.id)
            .set('Accept', 'application/json')
            .set('token', serviceAdminToken)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body.datatypes).to.be.an('array');
                expect(res.body.datatypes).to.have.lengthOf(2);
                expect(res.body.datatypes[0].datatype_id).to.equal(Config.datatype2.id);
                expect(res.body.datatypes[1].datatype_id).to.equal(Config.datatype3.id);
                done();
            });
    });

    // update datatypes (access rights) for service1 as service admin
    it('Should update newService1 as service admin', function (done) {
        chai.request(Config.server).put('omr/api/v1/services/' + newService1.id) 
            .set('Accept',  'application/json')
            .set('token', serviceAdminToken)
            .send({
                "id": newService1.id,
                "name": "newService1",
                "secret": "newservicepass",
                "ca_org": Config.caOrg,
                "email": "newservice2email@example.com",
                "org_id": Config.org1.id,
                "summary": "New service under org 1. Has two datatypes",
                "terms": {
                    "term1" : "example term",
                    "term2" : "example term",
                },
                "payment_required": "yes",
                "datatypes": [
                    {
                        "datatype_id": Config.datatype4.id,
                        "access":[
                            "write"
                        ]
                    }
                ],
                "solution_private_data": {
                    "sample data 1": "service 3 sample data 1",
                    "sample data 2": "service 3 sample data 2"
                },
                "status": "active"
            })
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.include("success");
                done();
            });
    });

    // get updated service1 as service admin
    it('Should return an updated information about service1', function (done) {
        chai.request(Config.server).get('omr/api/v1/services/' + newService1.id)
            .set('Accept', 'application/json')
            .set('token', serviceAdminToken)
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(200);
                expect(res.body.datatypes).to.be.an('array');
                expect(res.body.datatypes).to.have.lengthOf(1);
                expect(res.body.datatypes[0].datatype_id).to.equal(Config.datatype4.id);
                done();
            });
    });

    // revoke service admin permission from newUser4
    it('Should successfully revoke service admin permission from user', function (done) {
        chai.request(Config.server).delete('omr/api/v1/users/' + newUser4.id + '/permissions/services/' + newService1.id) 
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

    // user should not be able to update service1 as service admin anymore
    it('Should not update the service1 anymore', function (done) {
        chai.request(Config.server).put('omr/api/v1/services/' + newService1.id) 
            .set('Accept',  'application/json')
            .set('token', serviceAdminToken)
            .send({
                "id": newService1.id,
                "name": "newService1",
                "secret": "newservicepass",
                "ca_org": Config.caOrg,
                "email": "newservice2email@example.com",
                "org_id": Config.org1.id,
                "summary": "New service under org 1. Has two datatypes",
                "terms": {
                    "term1" : "example term",
                    "term2" : "example term",
                },
                "payment_required": "yes",
                "datatypes": [
                    {
                        "datatype_id": Config.datatype1.id,
                        "access":[
                            "write"
                        ]
                    }
                ],
                "solution_private_data": {
                    "sample data 1": "service 3 sample data 1",
                    "sample data 2": "service 3 sample data 2"
                },
                "status": "active"
            })
            .end(function(err, res) {
                expect(err).to.be.null;
                expect(res.status).to.equal(500);
                expect(res.body).to.have.property("msg");
                expect(res.body.msg).to.equal("Service is registered to CA, but failed to update service (CC):Failed to update service");
                done();
            });
    });
});
