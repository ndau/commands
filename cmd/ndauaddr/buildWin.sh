#! /bin/bash

echo "This builds the executable for windows."
GOOS=windows GOARCH=386 go build -o ndauAddr.exe main.go
