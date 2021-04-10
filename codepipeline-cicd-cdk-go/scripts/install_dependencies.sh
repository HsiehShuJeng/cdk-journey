#!/usr/bin/env bash
set -x
sudo yum -y remove java-1.6* java-1.7* > /var/tmp/app-deploy.log 2>&1
sudo yum -y install java-1.8* >> /var/tmp/app-deploy.log 2>&1
