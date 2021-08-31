#!/bin/bash

####################################################################################
##### Specify software and dependencies that are required for this project     #####
#####                                                                          #####
##### Note:                                                                    ##### 
##### (1) A log file is auto-created when this file runs. If you want to write #####
##### to it, the relative path is ./.brev/logs/setup.log                       #####
#####                                                                          #####
##### (2) The working directory is /home/brev/<PROJECT_FOLDER_NAME>. Execution #####
##### of this file happens at this level.                                      #####
####################################################################################

##### Yarn #####
# echo "##### Yarn #####" >> ./.brev/logs/setup.log
# curl -sS https://dl.yarnpkg.com/debian/pubkey.gpg | sudo apt-key add
# echo "deb https://dl.yarnpkg.com/debian/ stable main" | sudo tee /etc/apt/sources.list.d/yarn.list
# sudo apt update
# sudo apt install -y yarn

##### Node v14.x + npm #####
# echo "##### Node v14.x + npm #####" >> ./.brev/logs/setup.log
# curl -fsSL https://deb.nodesource.com/setup_14.x | sudo -E bash -
# sudo apt-get install -y nodejs

##### Python + Pip + Poetry #####
# echo "##### Python + Pip + Poetry #####" >> ./.brev/logs/setup.log
# sudo apt-get install -y python3-distutils
# sudo apt-get install -y python3-apt
# curl -sSL https://raw.githubusercontent.com/python-poetry/poetry/master/get-poetry.py | python3 -
# curl https://bootstrap.pypa.io/get-pip.py -o get-pip.py
# python3 get-pip.py
# rm get-pip.py
# source $HOME/.poetry/env

##### Golang v16x #####
echo "##### Golang v16x #####"
wget https://golang.org/dl/go1.16.7.linux-amd64.tar.gz
sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.16.7.linux-amd64.tar.gz
echo "" | sudo tee -a ~/.zshrc
echo "export PATH=\$PATH:/home/brev/lib/go/bin" | sudo tee -a ~/.zshrc
echo "export GOPATH=/home/brev/lib/go" | sudo tee -a ~/.zshrc
echo "export TMPDIR=/home/brev/tmp" | sudo tee -a ~/.zshrc
source ~/.zshrc
rm go1.16.7.linux-amd64.tar.gz