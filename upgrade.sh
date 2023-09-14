#!/bin/bash

cd /home/scpadm/scp-project

rm scp_master scp_orch scp_back
git stash
git pull

echo "Atualizando Front End"

rm -rf /var/www/html/*
cp build10.zip /var/www/html
cd /var/www/html
unzip build10.zip
mv build10/* .
chmod -R a+r *

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
