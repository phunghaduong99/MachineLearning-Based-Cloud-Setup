#SaaS web app backend image build

To control the version/tag of the docker image, edit the version in the image_version.txt

    The server.key and server.cert are created for localhost as the CN, so can be used directly on the development machines

# To generate the self signed cert for the portal

    ./cert-gen.sh <public IP of the machine>

## Example

    🕙 14:18:34 ❯ ./cert-gen.sh 192.168.101.38
    ----> Entered IP is 192.168.101.38

    ----> Modifying the cert-conf.txt file

    ----> Generating the server.key and server.cert file for IP 192.168.101.38

    Generating a RSA private key
    ...............................................................................................................................................++++

    ...........................................++++
    writing new private key to 'server.key'
    -----
    ----> Below is the CN and IP address in the generated cert

          Issuer: C = IN, ST = KA, L = SomeCity, O = aarna, OU = saas, CN = 192.168.101.38
            Subject: C = IN, ST = KA, L = SomeCity, O = aarna, OU = saas, CN = 192.168.101.38
                    IP Address:192.168.101.38

#Run the app in local dev environment
To set up a local dev environment run the below commands
    
    1. install the dependencies: npm install.
    2. update startup.sh for port, mongodb address etc (make sure mongo is up and running).
    3. Copy the UI bundle to be served in 'build' folder at root. 
    4. run 'startup.sh'
