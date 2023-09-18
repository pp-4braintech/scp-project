#!/bin/bash

echo "Executando atualizacao do Software"

ping 8.8.4.4 -c 5
if [ $? -ne 0 ] 
    then
        echo "Nao foi possivel fazer a atualizacao" 
        exit 1
fi

cd /home/scpadm/scp-project

cp scp_*.go /tmp/
rm scp_master scp_orch scp_back
git config --global --add safe.directory /home/scpadm/scp-project
git stash
git pull
if [ $? -ne 0 ] 
    then
        echo "Nao foi possivel fazer a atualizacao"
        cp /tmp/scp_*.go . 
        exit 1
fi

echo "Atualizando Front End"

rm -rf /var/www/html/*
cp build10.zip /var/www/html
cd /var/www/html
unzip build10.zip
mv build10/* .
chmod -R a+r *

cd /home/scpadm/scp-project

echo "Restartando Orquestrador"
go build scp_orch.go
systemctl restart scp_orch

echo "Restartando Back End"
go build scp_back.go
systemctl restart scp_back

echo "Restardando Master"
go build scp_master.go
echo "Verifique se a biofabrica esta pausada"

sleep 30
systemctl restart scp_master
