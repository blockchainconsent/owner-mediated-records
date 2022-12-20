#!/bin/bash
echo -------------------------
echo deploy chaincode
# ./deploy_chaincode.sh -i true
while echo $1 | grep -q ^-; do
    eval $(echo $1 | sed 's/^-//')=$2
    shift
    shift
done

interactive=${i}

PROJECT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SRC_DIR="$PROJECT_DIR/chaincodes/src"
DEPLOY_DIR="$PROJECT_DIR/chaincodes/deploy"
echo PROJECT_DIR=$PROJECT_DIR
echo SRC_DIR=$SRC_DIR
echo DEPLOY_DIR=$DEPLOY_DIR
echo -------------------------

# delete and create deploy directory
echo deleting $DEPLOY_DIR
rm -rf $DEPLOY_DIR
echo creating $DEPLOY_DIR
mkdir -p $DEPLOY_DIR

# For HUN deployment, we don't need Common sdk & Test files
cd $SRC_DIR
zip -r $DEPLOY_DIR/chaincode.zip ./solution_chaincode -x "./solution_chaincode/vendor/common/*" -x "*/testdata/*" -x "*/*_test.go"
# For dev testing use: zip -r $DEPLOY_DIR/chaincode.zip ./solution_chaincode

echo finished packaging chaincode
if [ "$interactive" == "true" ]; then
    read -p "Press any key to close... " -n1 -s
fi
