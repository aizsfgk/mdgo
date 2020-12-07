#!/usr/bin/env bash
#
###

FILE=$1
FILE_OUT=${FILE%.*}

echo "go build -o $FILE_OUT $FILE"
go build -o $FILE_OUT $FILE