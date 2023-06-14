#!/usr/bin/env bash

CERT_NAME=rapidsir
openssl req -x509 -nodes -new -sha256 -days 1024 -newkey rsa:2048 -keyout ${CERT_NAME}.key -out ${CERT_NAME}.pem -subj "/C=US/CN=rapidsir.com"
openssl x509 -outform pem -in ${CERT_NAME}.pem -out ${CERT_NAME}.crt