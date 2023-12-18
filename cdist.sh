#!/bin/bash

echo "Criando pacote Distribuicao do Software"

ping 8.8.4.4 -c 5
if [ $? -ne 0 ] 
    then
        echo "Falha na rede, abortando" 
        exit 1
fi

cd /home/paulo/work/iot/scp-source

rm -f scp_master scp_orch scp_back scp_agent
git config --global --add safe.directory /home/paulo/work/iot/scp-source
git stash
git pull
if [ $? -ne 0 ] 
    then
        echo "Nao foi possivel fazer a atualizacao"
        cp /tmp/scp_*.go . 
        exit 1
fi

echo "Atualizando Front End"

cp build10.zip /home/paulo/work/iot/scp-project


cd /home/paulo/work/iot/scp-source

echo "Compilando e Copiando Orquestrador"
go build scp_orch.go
if [ $? -ne 0 ] 
    then
        echo "Falha na compilacao do Orquestrador"
        exit 1
fi
cp scp_orch /home/paulo/work/iot/scp-project/


echo "Compilando e Copiando Back End"
go build scp_back.go
if [ $? -ne 0 ] 
    then
        echo "Falha na compilacao do Back End"
        exit 1
fi
cp scp_back /home/paulo/work/iot/scp-project/


echo "Compilando e Copiando Agent"
go build scp_agent.go
if [ $? -ne 0 ] 
    then
        echo "Falha na compilacao do Agent"
        exit 1
fi
cp scp_agent /home/paulo/work/iot/scp-project/


echo "Compilando e Copiando Master"
go build scp_master.go
if [ $? -ne 0 ] 
    then
        echo "Falha na compilacao do Master"
        exit 1
fi
cp scp_master /home/paulo/work/iot/scp-project/


echo "Compilando e Copiando Network"
go build scp_network.go
if [ $? -ne 0 ] 
    then
        echo "Falha na compilacao do Network"
        exit 1
fi
cp scp_network /home/paulo/work/iot/scp-project/

cd /home/paulo/work/iot/scp-project

git push --force-with-lease 


