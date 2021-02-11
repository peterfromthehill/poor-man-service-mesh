#!/bin/sh

# FIXME
curl https://www.1qay.net/ca.crt -o /ca/tls.crt
exec "$@"

