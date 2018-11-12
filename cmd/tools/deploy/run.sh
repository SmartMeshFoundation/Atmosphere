#!/bin/sh
go build
./newtestenv --keystore-path ../../../testdata/mykeystore --eth-rpc-endpoint ws://127.0.0.1:8546

