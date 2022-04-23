#! /bin/bash
printf "%s\n\n" "----> Entered IP is $1"
printf "%s\n\n" "----> Modifying the cert-conf.txt file"
sed -ie "s/CN = .*/CN = $1/g" cert-conf.txt
sed -ie "s/IP.1 = .*/IP.1 = $1/g" cert-conf.txt
printf "%s\n\n" "----> Generating the server.key and server.cert file for IP $1"
openssl req -new -nodes -x509 -days 730 -keyout server.key -out server.cert -config cert-conf.txt

printf "%s\n\n" "----> Below is the CN and IP address in the generated cert"

openssl x509 -in server.cert -noout -text | grep 'CN\|IP'
