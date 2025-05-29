#!/bin/bash

# Define chain-stresser config params
INSTANCES=1
VALIDATORS=60
SENTRIES=1
EVM=false
PROD=true
# Generate nodes configs
chain-stresser generate --instances $INSTANCES --validators $VALIDATORS --sentries $SENTRIES --evm $EVM --prod $PROD
