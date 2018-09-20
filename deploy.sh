#!/bin/sh

if [ $# != 2 ]; then
  echo "please input \"create OR update\" \"username:password\""
  echo "ex) $0 create username:password"
  exit 1
fi

METHOD=POST
URL="http://marathon.example.com/v2/apps"
if [ $1 == "update" ]; then
    METHOD=PUT
    URL="${URL}/docker-builder"
fi
curl -v -X${METHOD} -H "Content-Type: application/json" ${URL} -d @marathon.json -u "$2"