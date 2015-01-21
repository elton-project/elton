#!/bin/sh

go install 2>&1 | awk -F'"|"' '{print $2}' | xargs go get
go install
