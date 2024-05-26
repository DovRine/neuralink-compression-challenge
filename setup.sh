#!/usr/bin/env bash

echo "compile encode: start"
go build -o ../../bin/encode src/encode/main.go
echo "compile encode: complete"

echo "compile decode: start"
go build -o ../../bin/decode src/decode/main.go
echo "compile decode: complete"
