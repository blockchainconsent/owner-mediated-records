#!/bin/sh
# build 164 hopefully maybe someday
# Check for "--full" flag
FULL=false
if [[ $* == *--full* ]] ; then
  printf "\n\nRunning FULL test including JMeter.\n\n"
  FULL=true
else
  printf "\n\nRunning SHORT test without JMeter.\n\n"
fi

printf "\n***** HostName $HOSTNAME *****\n"
export DOCKER_ENDPOINT=tcp://$HOSTNAME:2375
printf "\n***** DOCKER_ENDPOINT $DOCKER_ENDPOINT *****\n"
# fail if any command below fails
#set -e
function cleanup {
  printf "\n\n***** Cleaning up workspace.1 *****\n\n"
  printf "\n***** workspace=: $WORKSPACE *****\n" 
  cd $WORKSPACE
  docker-compose -f server/config/network_local_v1.1/artifacts/docker-compose.yaml down
  printf "\n***** Cleaning up workspace.2 *****\n"
  cd ..
  rm -rf gopath
  rm -rf go
  forever stopall
}
#trap 'cleanup ${LINENO} $? ' ERR
printf "\n\n***** Call cleanupfunction *****\n\n"
cleanup
cd $WORKSPACE
rm -rf tmp
rm -rf tmp_org*
cd ..
rm -rf gopath
rm -rf go

# Download and install golang
printf "\n\n***** Installing Go 1.11.1 *****\n\n"
wget https://dl.google.com/go/go1.11.1.linux-amd64.tar.gz -nv -nc
tar -xf go1.11.1.linux-amd64.tar.gz


# Set up golang env
export GOROOT=$PWD/go
export PATH=$PATH:$GOROOT/bin
mkdir -p gopath
export GOPATH=$PWD/gopath


# Download Fabric v1.1
printf "\n\n***** Downloading Hyperledger Fabric 1.1 *****\n\n"
mkdir -p $GOPATH/src/github.com/hyperledger
cd $GOPATH/src/github.com/hyperledger
git clone --branch release-1.1 https://github.com/hyperledger/fabric.git


# Deploy chaincode
printf "\n\n***** Deploying Go code *****\n\n"
cd $WORKSPACE/scripts
chmod 755 deploy_chaincode.sh
./deploy_chaincode.sh


## Build & run Go tests
# mkdir -p $WORKSPACE/server/deploy/chaincode
export GOPATH=$GOPATH:$WORKSPACE/server
cd $WORKSPACE/server/src/chaincode
printf "\n\n***** Building Go code *****\n\n"
go build --tags nopkcs11 ./...
printf "\n\n***** Running Go tests *****\n\n"
go test --tags nopkcs11 ./...


# Only start the app if "--full" flag was passed
if [ "$FULL" = true ] ; then

  # Start Fabric docker containers
  printf "\n\n***** Starting Fabric Docker Containers *****\n\n"
  cd $WORKSPACE/server/config/network_local_v1.1/artifacts
  docker pull hyperledger/fabric-ccenv:x86_64-1.1.0
  docker pull hyperledger/fabric-baseos:x86_64-0.4.6

  printf "\n***** HostName $HOSTNAME *****\n"
  export DOCKER_ENDPOINT=tcp://$HOSTNAME:2375
  printf "\n\n***** DockerEndpoint $DOCKER_ENDPOINT *****\n\n"
  docker-compose -f docker-compose.yaml up -d

  # Install NodeJS dependencies
  printf "\n\n***** Installing NodeJS Dependencies *****\n\n"
  cd $WORKSPACE
  cd $WORKSPACE
  npm config set loglevel error
  npm install forever -g
  npm install grpc -g
  npm install
  cd server/utils
  npm link
  cd ../..
  npm link chain-utils
  npm link jsonschema
  npm run setup

  # npm config set loglevel error
#   npm install forever -g
#   npm install grpc -g



# Build the React app
  printf "\n\n***** Building the React app *****\n\n"
  npm run build


  # Download JMeter updated jmeter version, do not forget to change it in the jenkins file
  printf "\n\n***** Downloading JMeter *****\n\n"
  cd $WORKSPACE
  wget http://mirror.reverse.net/pub/apache//jmeter/binaries/apache-jmeter-5.0.tgz -nv -nc
  tar -xf apache-jmeter-5.0.tgz

  touch server.log
  # Start the NodeJS app
  printf "\n\n***** Starting server.js *****\n\n"
  
  cd $WORKSPACE
  
  #forever -l $WORKSPACE/server.log -a start server.js -m 1
  nohup node server/server.js &> server.log &
  printf "\n\n***** what is here $(ls -alt) *****\n\n"
  # Wait for the server to start up
  export SERVEROUTA=$(tail -n 50 $WORKSPACE/server.log)
  printf "\n\n***** SERVEROUTA $SERVEROUTA *****\n\n"
  printf "\n\n***** Waiting for the server to start up *****\n\n"
  ( tail -F $WORKSPACE/server.log & ) | grep -m1 -E 'SERVER READY|SERVER FAILED TO START SUCCESSFULLY'
  printf "\n\n***** Server is up *****\n\n"

  # Save the container logs
  cd $WORKSPACE/server/config/network_local_v1.1/artifacts
  export WHOISHERE=$(ls -alt)
  export WHEREAMI=$(pwd)
  printf "\n\n***** Who is here $WHOISHERE *****\n\n"
  printf "\n\n***** Where am I $WHEREAMI *****\n\n"
  
  docker-compose logs >& $WORKSPACE/containers.log &
  docker logs -f $(docker ps --format '{{.Names}}' | grep dev) >& $WORKSPACE/chaincode.log &

fi