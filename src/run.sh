#!/bin/bash

NUMGR=`cat config | grep "NUM_GOROUTINE" | awk -F '=' '{print $2}'`
BLKSTABLE=`cat config | grep "BLOCK_STABLE_COUNT" | awk -F '=' '{print $2}'`
BLKSTART=`cat config | grep "BLOCK_START" | awk -F '=' '{print $2}'`
AVGBTIME=`cat config | grep "AVG_BLOCK_TIME" | awk -F '=' '{print $2}'`
RPCURL=`cat config | grep "RPC_URL" | awk -F '=' '{print $2}'`

cd /go/src
nohup go run main.go dbstruct.go -g $NUMGR -s $BLKSTABLE -b $BLKSTART -t $AVGBTIME -u $RPCURL > /dev/null &
nohup go run srvmain.go dbstruct.go > srvmain.log 2>&1 & 
