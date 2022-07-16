#!/bin/zsh

# This is the bare minimum to run in development. For full list of flags,
# run ./vigilate -help

go build -o vigilate cmd/web/*.go && ./vigilate \
-dbuser='postgres' \
-dbpass='123456' \
-pusherHost='localhost' \
-pusherKey='123abc' \
-pusherSecret='abc123' \
-pusherApp="1"
#-pusherPort="4001"
#-pusherSecure=false