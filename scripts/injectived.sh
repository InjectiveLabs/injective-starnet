#!/bin/sh

echo "############### injectived service started ################"
#!/bin/bash

ulimit -n 120000
yes 12345678 | injectived \
--log-level "info" \
--rpc.laddr "tcp://0.0.0.0:26657" \
--home $HOME/.injectived \
start