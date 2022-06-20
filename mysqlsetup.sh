#!/bin/bash
set -e
service mysql start
mysql < "CREATE DATABASE evm_data;"

service mysql stop
