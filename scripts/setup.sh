#!/bin/bash

echo Building Root App NodeJS Code
npm install

cd common/utils
npm link

echo Running npm link on chain-utils package
cd ../..
npm link common-utils

echo Done
