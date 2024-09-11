#!/bin/env bash

rm -rf certs
mkdir certs
pushd certs

openssl genrsa \
    -out ca.key 2048

openssl req \
    -x509 \
    -new \
    -nodes \
    -key ca.key \
    -sha256 \
    -days 1825 \
    -out ca.crt \
    -subj "/C=DE/ST=NRW/L=Aachen/O=nulll/CN=Certificate Authority"

# Server
openssl genrsa \
    -out server.key 2048

openssl req \
    -new \
    -subj "/C=DE/ST=NRW/L=Aachen/O=nulll/CN=localhost" \
    -key server.key \
    -out server.csr

cat > server.ext << EOF
authorityKeyIdentifier=keyid, issuer
basicConstraints=CA:FALSE
keyUsage = critical, digitalSignature, nonRepudiation, keyEncipherment, dataEncipherment
extendedKeyUsage=serverAuth
subjectAltName = DNS:localhost"
EOF

openssl x509 \
    -req \
    -in server.csr \
    -CA ca.crt \
    -CAkey ca.key \
    -CAcreateserial \
    -out server.crt \
    -days 825 \
    -sha256 \
    -extfile server.ext

# Clients
for CLIENT in client1 client2; do
    openssl genrsa \
        -out ${CLIENT}.key 2048

    openssl req \
        -new \
        -subj "/C=DE/ST=NRW/L=Aachen/O=nulll/CN=${CLIENT}" \
        -key ${CLIENT}.key \
        -out ${CLIENT}.csr

    cat > ${CLIENT}.ext << EOF
authorityKeyIdentifier=keyid, issuer
basicConstraints=CA:FALSE
keyUsage = critical, nonRepudiation, digitalSignature, keyEncipherment
extendedKeyUsage=clientAuth
EOF

    openssl x509 \
        -req \
        -in ${CLIENT}.csr \
        -CA ca.crt \
        -CAkey ca.key \
        -CAcreateserial \
        -out ${CLIENT}.crt \
        -days 825 \
        -sha256 \
        -extfile ${CLIENT}.ext
done
