#!/bin/bash

NUMGR=`cat config | grep "num_gr" | awk -F '=' '{print $2}'`
BLKSTABLE=`cat config | grep "b_stable" | awk -F '=' '{print $2}'`
BLKSTART=`cat config | grep "b_start" | awk -F '=' '{print $2}'`

cd /go/src
go run main.go dbstruct.go -g $NUMGR -s $BLKSTABLE -b $BLKSTART &> /dev/null &
go run srvmain.go dbstruct.go &> /dev/null &
