#!/bin/bash

echo
echo " ____    _____      _      ____    _____ "
echo "/ ___|  |_   _|    / \    |  _ \  |_   _|"
echo "\___ \    | |     / _ \   | |_) |   | |  "
echo " ___) |   | |    / ___ \  |  _ <    | |  "
echo "|____/    |_|   /_/   \_\ |_| \_\   |_|  "
echo
echo "Build your first network (BYFN) end-to-end local network deploy"
echo
CHANNEL_NAME="$1"
DELAY="$2"
LANGUAGE="$3"
TIMEOUT="$4"
VERBOSE="$5"
NO_CHAINCODE="$6"
CCNAME="$7"
VERSION="$8"
INIT_ARGS="$9"
: ${CHANNEL_NAME:="mychannel"}
: ${DELAY:="3"}
: ${LANGUAGE:="golang"}
: ${TIMEOUT:="10"}
: ${VERBOSE:="false"}
: ${NO_CHAINCODE:="false"}
: ${CCNAME:="owner_mediated_records"}
: ${VERSION:="v1.0"}
: ${INIT_ARGS:="_local"}
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

echo "Channel name : "$CHANNEL_NAME
echo "Init Args : "$INIT_ARGS
# import utils
. scripts/utils.sh

createChannel() {
	setGlobals 0 1

	if [ -z "$CORE_PEER_TLS_ENABLED" -o "$CORE_PEER_TLS_ENABLED" = "false" ]; then
                set -x
		peer channel create -o orderer.example.com:7050 -c $CHANNEL_NAME -f ./channel-artifacts/channel.tx >&log.txt
		res=$?
                set +x
	else
				set -x
		peer channel create -o orderer.example.com:7050 -c $CHANNEL_NAME -f ./channel-artifacts/channel.tx --tls $CORE_PEER_TLS_ENABLED --cafile $ORDERER_CA >&log.txt
		res=$?
				set +x
	fi
	cat log.txt
	verifyResult $res "Channel creation failed"
	echo "===================== Channel '$CHANNEL_NAME' created ===================== "
	echo
}

joinChannel () {
	for org in 1 2; do
	    for peer in 0 1; do
		joinChannelWithRetry $peer $org
		echo "===================== peer${peer}.org${org} joined channel '$CHANNEL_NAME' ===================== "
		sleep $DELAY
		echo
	    done
	done
}

## Create channel
echo "Creating channel..."
createChannel

## Join all the peers to the channel
echo "Having all peers join the channel..."
joinChannel

## Set the anchor peers for each org in the channel
echo "Updating anchor peers for org1..."
updateAnchorPeers 0 1
echo "Updating anchor peers for org2..."
updateAnchorPeers 0 2

if [ "${NO_CHAINCODE}" != "true" ]; then
    #creage package
    #echo "Packing chaincode deployment spec..."
    #packageChaincode ${CCNAME} ${VERSION}

	# Install chaincode
    echo "Installing chaincode on peer0.org1..."
    installChaincode 0 1 ${CCNAME} ${VERSION}
    echo "Installing chaincode on peer1.org1..."
    installChaincode 1 1 ${CCNAME} ${VERSION}
    echo "Installing chaincode on peer0.org2..."
    installChaincode 0 2 ${CCNAME} ${VERSION}
    echo "Installing chaincode on peer1.org2..."
    installChaincode 1 2 ${CCNAME} ${VERSION}

	# Instantiate chaincode on peer0.org2
	echo "Instantiating chaincode on peer0.org1..."
    instantiateChaincode 0 1 ${CCNAME} ${VERSION} ${INIT_ARGS}
    echo "Instantiating chaincode on peer0.org2..."
    instantiateChaincode 0 2 ${CCNAME} ${VERSION} ${INIT_ARGS}
    #instantiateChaincode 1 1 ${CCNAME} ${VERSION} ${INIT_ARGS}
    #instantiateChaincode 1 2 ${CCNAME} ${VERSION} ${INIT_ARGS}

fi

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
