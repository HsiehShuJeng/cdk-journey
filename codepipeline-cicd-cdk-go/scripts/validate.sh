#!/bin/bash
set -x
echo "Waiting for 5 seconds before checking health.."
sleep 15

status_code=$(curl --write-out %{http_code} --silent --output /dev/null http://localhost:8080)

if [[ "$status_code" -ne 200 ]] ; then
  echo "FAIL: App is not healthy - $status_code"
  exit 1
else
  echo "PASS: App is responding with $status_code"
  exit 0
fi

