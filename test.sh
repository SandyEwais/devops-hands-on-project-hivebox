#!/bin/bash
echo "Inside script:"
which go
go version
name="Test"
expected="Hello, Test!"

actual=$(go run main.go "$name")

if [ "$actual" == "$expected" ]; then
    echo "✅ Test passed!"
    exit 0
else
    echo "❌ Test failed! Expected '$expected' but got '$actual'"
    exit 1
fi