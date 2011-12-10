#!/bin/sh
#
# Use OpenSSL to generate a self-signed certificate, valid for a year.
# Pass in the filename.
echo "A temporary passphrase will be requested multiple times during this process."
echo "Enter the same one each time. The output file will not need a passphrase."
openssl genrsa -des3 -out $1.key.tmp 1024
openssl rsa -in $1.key.tmp -out $1.key  # Remove passphrase.
openssl req -new -key $1.key -out $1.csr
openssl x509 -req -days 365 -in $1.csr -signkey $1.key -out $1.crt
rm $1.key.tmp
rm $1.csr
