#!/bin/bash

go build -buildmode=plugin -o api1/api.so api1/api.go
go build -buildmode=plugin -o api2/api.so api2/api.go

go build

rm -rf modules
mkdir -p modules

cp api1/api.so modules/api1.so
cp api2/api.so modules/api2.so