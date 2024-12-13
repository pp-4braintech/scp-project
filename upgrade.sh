#!/bin/bash

echo "Executando atualizacao do Software"

# ping 8.8.4.4 -c 5
# if [ $? -ne 0 ] 
#     then
#         echo "Nao foi possivel fazer a atualizacao" 
#         exit 1
# fi

cd /home/scpadm/scp-project

cp scp_* /tmp/
rm -f scp_master scp_orch scp_back scp_agent
git config --global --add safe.directory /home/scpadm/scp-project
git stash
git pull
if [ $? -ne 0 ] 
    then
        echo "Nao foi possivel fazer a atualizacao"
        cp /tmp/scp_* . 
        exit 1
fi

cp initd/_bashrc /home/scpadm/.bashrc

DIRSYS=/etc/systemd/system/
FILEAGENT=scp_agent.service

if [ -e "$DIRSYS$FILEAGENT" ] 
    then
        echo "SCP AGENT OK"
    else
        echo "Criando SCP AGENT"
        cp /home/scpadm/scp-project/initd/scp_agent.service /etc/systemd/system/scp_agent.service
        systemctl enable scp_agent
        systemctl start scp_agent
fi

DIRETC=/etc/scpd/
FILESTDA=scp_standalone.flag
FILESERV=scp_slaveusb.service

if [ -e "$DIRETC$FILESTDA" ] 
    then
	echo "STANDALONE COPIANDO START BROWSE"
	cp /home/scpadm/scp-project/initd/start_browser.sh /home/scpadm/scp-project/

        if [ -e "$DIRSYS$FILESERV" ] 
            then
                echo "STANDALONE SCP SLAVEUSB OK"
            else
                echo "STANDALONE Criando SCP SLAVEUSB"
                cp /home/scpadm/scp-project/initd/scp_slaveusb.service /etc/systemd/system/scp_slaveusb.service
                systemctl enable scp_slaveusb
                systemctl start scp_slaveusb
        fi
    else 
        echo "MODO BIOFABRICA"
fi

echo "Atualizando Front End"

rm -rf /var/www/html/*
if [ -e "$DIRETC$FILESTDA" ] 
    then 
        echo "COPIANDO BUILD 127 - StandAlone"
        cp build127.zip /var/www/html/build10.zip
    else
        echo "COPIANDO BUILD 10 - Biofabrica"
        cp build10.zip /var/www/html
fi

cd /var/www/html
unzip build10.zip
mv build10/* .
chmod -R a+r *

cd /home/scpadm/scp-project

echo "Parando Master"
systemctl stop scp_master

echo "Restartando Orquestrador"
systemctl restart scp_orch

if [ -e "$DIRETC$FILESTDA" ] 
    then
        echo "Restartando SlaveUSB"
        systemctl restart scp_slaveusb
    else 
        echo "OK"
fi

echo "Restartando Back End"
systemctl restart scp_back

echo "Restartando Back Agent"
systemctl restart scp_agent

echo "Restardando Master"
sleep 60
systemctl start scp_master
