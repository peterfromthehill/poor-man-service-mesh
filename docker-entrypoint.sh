#!/bin/sh

curl https://www.1qay.net/ca.crt -o /ca/tls.crt
set 
sleep 10
exec "$@"

