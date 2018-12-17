#!/bin/bash
cd $SRC_DIR
# compile the app
go get
go build -o stream
cp stream /app/
cp conf.json /app/conf.json
# run the app
cd /app
./stream