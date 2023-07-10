#!/bin/bash

# Create the CA Key and Certificate for signing Client Certs
openssl req -new -nodes -x509 -days 365 -keyout ca.key -out ca.crt

# Create the Server Key and CSR (certifiate signing request)
openssl req -new -newkey rsa:2048 -nodes -keyout server.key -out server.csr
# Sugn the Server CSR with shated CA cert
openssl x509 -req -days 365 -in server.csr -CA ca.crt -CAkey ca.key -out server.crt -extfile req.cnf

# Create the Client Key and CSR (certifiate signing request)
openssl req -new -newkey rsa:2048 -nodes -keyout client.key -out client.csr
# Sugn the Client CSR with shated CA cert
openssl x509 -req -days 365 -in client.csr -CA ca.crt -CAkey ca.key -out client.crt

# validate a cert
# openssl x509 -text -noout -in server.crt

# validate a csr
# openssl req -text -noout -verify -in server.csr
