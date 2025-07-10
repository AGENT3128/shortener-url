#!/bin/bash

# Generate self-signed certificate for HTTPS server for local development
# Usage: ./generate_selfsign-cert.sh

# Generate private key with stronger encryption (4096 bit)
openssl genrsa -out shortener.key 4096

# Create config file for certificate extensions
cat > cert.conf <<EOF
[req]
default_bits = 4096
prompt = no
distinguished_name = req_distinguished_name
req_extensions = v3_req

[req_distinguished_name]
C=RU
ST=Moscow
L=Moscow
O=local
OU=local
CN=shortener.local

[v3_req]
keyUsage = digitalSignature, keyEncipherment
extendedKeyUsage = serverAuth
subjectAltName = @alt_names

[alt_names]
DNS.1 = shortener.local
DNS.2 = localhost
IP.1 = 127.0.0.1
EOF

# Generate certificate signing request with extensions
openssl req -new -key shortener.key -out shortener.csr -config cert.conf

# Generate self-signed certificate with proper extensions and longer validity
openssl x509 -req -days 3650 -in shortener.csr -signkey shortener.key -out shortener.crt -extensions v3_req -extfile cert.conf

# Set proper permissions for production security
chmod 600 shortener.key
chmod 644 shortener.crt

# Clean up
rm shortener.csr

echo "Certificate generated successfully!"