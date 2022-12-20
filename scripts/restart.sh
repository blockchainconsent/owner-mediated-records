#!/bin/bash

restartType="$1"
PROJECT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )/.." && pwd )"

case $restartType in
    server-clean)
    (cd $PROJECT_DIR/scripts; ./cleanup.sh)

    (cd $PROJECT_DIR/scripts; ./deploy_chaincode.sh)

    (cd $PROJECT_DIR/network_local_HF; ./byfn.sh up -m owner_mediated_records -f docker-compose-e2e.yaml -y)

    (cd $PROJECT_DIR/server; node server.js)
    ;;
    server)
    (cd $PROJECT_DIR/scripts; ./deploy_chaincode.sh)

    (cd $PROJECT_DIR/network_local_HF; ./byfn.sh up -m owner_mediated_records -f docker-compose-e2e.yaml -y)

    (cd $PROJECT_DIR/server; node server.js)
    exit 0
    ;;
    *)
    (cd $PROJECT_DIR/scripts; ./cleanup.sh)

    (cd $PROJECT_DIR/chaincodes/src; go build --tags nopkcs11 ./...)

    (cd $PROJECT_DIR/scripts; ./deploy_chaincode.sh)

    (cd $PROJECT_DIR/network_local_HF; ./byfn.sh up -m owner_mediated_records -f docker-compose-e2e.yaml -y)

    (cd $PROJECT_DIR/server/config; cp ./local_HF/* .)

    (cd $PROJECT_DIR; npm run setup)

    #(cd $PROJECT_DIR; npm run build)

    (cd $PROJECT_DIR/server; node server.js)
    ;;
esac
