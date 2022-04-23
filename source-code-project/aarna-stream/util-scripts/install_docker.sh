#!/bin/bash
sudo apt-get update
sudo apt-get install ca-certificates curl gnupg lsb-release -y
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg --yes
sudo echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/ubuntu \
  $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
sudo apt-get update
sudo apt-get install docker-ce=18.06.2~ce~3-0~ubuntu -y
sudo chmod 777 /etc/apt/sources.list.d/docker.list
sudo echo "deb https://download.docker.com/linux/ubuntu bionic stable" > /etc/apt/sources.list.d/docker.list
if [[ ! -e /etc/docker/daemon.json ]]; then
    sudo touch /etc/docker/daemon.json
    sudo chmod 777 /etc/docker/daemon.json
    echo '{"storage-driver": "aufs"}' | sudo tee /etc/docker/daemon.json
    echo "Done Creating and updating docker daemon.json file "
fi
sudo systemctl restart docker
