#!/bin/bash

cd /home/scpadm/scp-project
git pull

echo "Atualizando Front End"

rm -rf /var/www/html/*
cp build10.zip /var/www/html
cd /var/www/html
unzip build10.zip
mv build10/* .
chmod -R a+r *

echo "Restartando Orquestrador"
systemctl restart scp_orch

echo "Restartando Back End"
systemctl restart scp_back

echo "Restardando Master"
echo "Verifique se a biofabrica esta pausada"

sleep 30
systemctl restart scp_master
