#!/bin/bash

echo
echo " ____    _____      _      ____    _____ "
echo "/ ___|  |_   _|    / \    |  _ \  |_   _|"
echo "\___ \    | |     / _ \   | |_) |   | |  "
echo " ___) |   | |    / ___ \  |  _ <    | |  "
echo "|____/    |_|   /_/   \_\ |_| \_\   |_|  "
echo
echo "Creat CDS package"
echo
CCNAME="$1"
VERSION="$2"
LANGUAGE="$3"
: ${CCNAME:="owner_mediated_records"}
: ${VERSION:="v1.0"}
: ${LANGUAGE:="golang"}
LANGUAGE=`echo "$LANGUAGE" | tr [:upper:] [:lower:]`
COUNTER=1
MAX_RETRY=10

CC_SRC_PATH="github.com/chaincodes/solution_chaincode/"
if [ "$LANGUAGE" = "node" ]; then
	CC_SRC_PATH="github.com/chaincodes/solution_chaincode/node/"
fi

if [ "$LANGUAGE" = "java" ]; then
	CC_SRC_PATH="github.com/chaincodes/solution_chaincode/java/"
fi

echo "Chincode name : "$CCNAME
echo "Chincode version : "$VERSION
echo "Chincode language : "$LANGUAGE
echo "Chincode path : /opt/gopath/src/"$CC_SRC_PATH

# import utils
. scripts/utils.sh

#creage package
echo "Packing chaincode deployment spec..."
packageChaincode ${CCNAME} ${VERSION}


echo
echo "========= All GOOD, BYFN execution completed =========== "
echo

echo
echo " _____   _   _   ____   "
echo "| ____| | \ | | |  _ \  "
echo "|  _|   |  \| | | | | | "
echo "| |___  | |\  | | |_| | "
echo "|_____| |_| \_| |____/  "
echo

exit 0
