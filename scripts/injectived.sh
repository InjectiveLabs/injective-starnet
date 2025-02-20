#!/bin/sh

echo "############### injectived service started ################"
#!/bin/bash

sudo cp $HOME/.injectived/libwasmvm.x86_64.so /usr/lib

ulimit -n 120000
yes 12345678 | $HOME/.injectived/injectived \
--log-level "info" \
--rpc.laddr "tcp://0.0.0.0:26657" \
--home $HOME/.injectived \
start