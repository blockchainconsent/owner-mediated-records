# Owner Mediated Records

This repo contains the Owner-Mediated-Records (OMR) solution for securely exchanging health records via a blockchain.

## Prerequisites

1. Install Go 1.12.12

Download Go for your platform:
- [Mac](https://go.dev/dl/go1.12.12.darwin-amd64.pkg)
- [Windows](https://go.dev/dl/go1.12.12.windows-amd64.msi)
- [Linux](https://go.dev/dl/go1.12.12.linux-amd64.tar.gz)

These follow [these instructions](https://go.dev/doc/install) to complete the installation.

2. Clone Hyperledger Fabric 1.4.4

```
mkdir -p $GOPATH/src/github.com/hyperledger
cd $GOPATH/src/github.com/hyperledger
git clone --branch release-1.4 https://github.com/hyperledger/fabric.git
```

3. Install NodeJS v8.x or v10.x and npm v6.x or greater

NodeJS for various platforms can be found [here](https://nodejs.org/en/download/).

Windows only:
```
npm --add-python-to-path=true install --global windows-build-tools
npm install --global grpc
```

Restart your command prompt after executing.

4. Install Docker & Docker Compose

Docker for various platforms can be found [here](https://hub.docker.com/search?type=edition&offering=community&q=).

On Mac & Windows, that installation includes Docker Compose. On Linux, follow [these instructions](https://docs.docker.com/compose/install/) to install Compose separately.

## Development Setup

### Git Submodules

OMR is built on top of the HCLS SDK, which must be pulled using git submodules.
This git repo has two submodules called `common` and `utils`, as specified in `.gitmodules`.
For more information on git submodules, see [here](https://git-scm.com/book/en/v2/Git-Tools-Submodules).

#### Cloning

Run the following commands to clone the submodules. Note that if cloning over HTTPS, you must enable a [Github access token](https://help.github.com/en/github/authenticating-to-github/creating-a-personal-access-token-for-the-command-line).

```
$ git clone https://github.com/blockchain-hcls/owner-mediated-records.git --recurse-submodules
$ cd owner-mediated-records
$ git submodule update --init --recursive
```

**Note:** It is assumed from here on that all paths are relative to the root `owner-mediated-records` directory.

#### Pulling

Run the following command to pull the required submodules. Note that these also need to be pulled over HTTPS.

```
$ git submodule update --recursive
```

**Note:** This will not pull the latest version of the submodules, but rather will pull the commit that is checked into
`owner-mediated-records`. This is the intended behavior.

#### Upgrading common Submodule
If you would like to upgrade to the latest version of `common`, run this:
1. `$ cd chaincodes/src/solution_chaincode/vendor/common`
2. `$ git pull origin master`


#### Upgrading utils Submodule
If you would like to upgrade to the latest version of `utils`, run this:
1. `$ cd common/utils`
2. `$ git pull origin master`

Then commit & push the update to `owner-mediated-records` once you fix any compilation/test errors.

#### Using Forked Repository
If you cloned from your forked repository, use the following to add upstream in your forked repo:
1. `$ git remote add upstream https://github.com/blockchain-hcls/owner-mediated-records.git`
2. `$ git fetch upstream --recurse-submodules`

Then update your local repo from the original repo and update your forked repo:
1. `$ git pull upstream master --recurse-submodules`
2. `$ git push`

### Go Chaincode

#### Building

Run the following commands from `owner-mediated-records` to add the Go code to your `GOPATH` and build the code:

```
$ export GOPATH=$PWD/chaincodes:$GOPATH
$ cd chaincodes/src/
$ go build --tags nopkcs11 ./...
```

#### Testing

Run the following command to run the Go tests:

```
$ cd chaincodes/src/
$ go test --tags nopkcs11 ./...
```

#### Formatting

Run the following command to format the Go code:

```
$ cd chaincodes/src/
$ go fmt ./...
```

**Note: Always format the Go code before submitting a Merge Request!**

### Fabric Network

#### Packaging Chaincode

Run the following command to package the chaincode for deployment.
**Note:** first make sure that the chaincode is able to build without errors.

```
$ ./scripts/deploy_chaincode.sh
```

This will create a zip file containing the chaincode at `chaincodes/deploy/chaincode.zip`.

#### Creating Network

Run the following commands to spin up the fabric network:

```
$ cd ./network_local_HF/
$ ./byfn.sh up -m owner_mediated_records -f docker-compose-e2e.yaml -y
```

If you are not using the restart script and are manually starting the server, run the following commands from the root `owner-mediated-records` directory to copy the contents of `server/config/local_HF` into `server/config`:

```
$ cd server/config
$ cp ./local_HF/* .
```

### Node Server

#### Building NodeJS Code

Run the following command to install npm dependencies and link the submodules (from the root `owner-mediated-records` directory):

```
$ npm run setup
```

## Running the App

With the Fabric Network running and the Node Server set up, run the following commands to start the app (from the root `owner-mediated-records` directory):

```
$ cd server
$ node server.js
```

The app should be accessible at `localhost:3000`, with Swagger running at `localhost:3000/api-docs`.

## Convenience Scripts

Use the cleanup script to restart Docker & the temporary directories:

```
$ ./scripts/cleanup.sh
```

Use the restart script to quickly restart the server (takes care of cleanup and restarting the Fabric Network and Node Server):

```
$ ./scripts/restart.sh
```
