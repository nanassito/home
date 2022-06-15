set -eux

USER=$1
openssl genrsa -des3 -out ${USER}.key 4096
openssl req -new -key ${USER}.key -out ${USER}.csr
openssl x509 -req -days 365 -in ${USER}.csr -CA ca.crt -CAkey ca.key -set_serial 01 -out ${USER}.crt
openssl pkcs12 -export -out ${USER}.pfx -inkey ${USER}.key -in ${USER}.crt -certfile ca.crt