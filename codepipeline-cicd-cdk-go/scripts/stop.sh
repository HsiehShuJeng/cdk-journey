#!/usr/bin/env bash
set -x
APP_HOME="/home/ec2-user/server2"
ifAppRunning=`pgrep java`
if [[ -n  $ifAppRunning ]]; then
   sudo killall java 
fi
echo "listing files..."
ls -l $APP_HOME/
