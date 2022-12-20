#!/bin/bash
echo Cleanup:

PROJECT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )/.." && pwd )"

echo PROJECT_DIR=$PROJECT_DIR

echo Take down and remove fabric containers...

cd $PROJECT_DIR/network_local_HF; ./byfn.sh down -f docker-compose-e2e.yaml -y

docker rm $(docker ps -aq)

echo Done.

echo Cleaning data artifacts created by previous runs...

rm -rf $PROJECT_DIR/tmp $PROJECT_DIR/tmp_org*

echo Done.

echo Cleanup complete.
