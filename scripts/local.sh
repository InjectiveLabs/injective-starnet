#!/bin/bash

# Define chain-stresser config params
INSTANCES=1
VALIDATORS=2
SENTRIES=1
EVM=false
PROD=true

INJECTIVE_CORE_PATH="../injective-core"
CHAIN_STRESSER_PATH="../chain-stresser"

ROOT_DIR=$PWD
# Build injectived binary
cd $INJECTIVE_CORE_PATH
#make install

# Build chain-stresser binary
cd $CHAIN_STRESSER_PATH
#make install

cd $ROOT_DIR

# Generate nodes configs
chain-stresser generate --instances $INSTANCES --validators $VALIDATORS --sentries $SENTRIES --evm $EVM --prod $PROD

chmod -R 777 $CHAIN_STRESSER_PATH

# copy files to validators
# Loop over validators and copy injectived and libwasmvm.x86_64.so
cd $GOPATH/bin
WASMVM_SO=$(ldd injectived | grep libwasmvm.x86_64.so | awk '{ print $3 }')
for i in $(seq 0 $(($VALIDATORS - 1))); do
    cp injectived $WASMVM_SO $ROOT_DIR/chain-stresser-deploy/validators/$i/
done

# copy files to sentries
for i in $(seq 0 $(($SENTRIES - 1))); do
    cp injectived $WASMVM_SO $ROOT_DIR/chain-stresser-deploy/sentry-nodes/$i/
done
