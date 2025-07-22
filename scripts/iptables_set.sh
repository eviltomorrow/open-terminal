#!/bin/bash

set -eo pipefail

ip=""
port=""

iptables -I DOCKER-USER  !  -s ${ip}  -p tcp --dport ${port} -j DROP
iptables -I DOCKER-USER     -s ${ip}  -p tcp --dport ${port} -j ACCEPT