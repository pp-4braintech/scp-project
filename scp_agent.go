package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Biofabrica_data struct {
	BFId         string
	BFName       string
	Status       string
	CustomerId   string
	CustomerName string
	Address      string
	SWVersion    string
	LatLong      [2]float64
	LastUpdate   string
	BFIP         string
}

var mybf = Biofabrica_data{"bf000", "Modelo", "ERRO", "HA", "Hubio Agro", "", "1.2.15", [2]float64{-18.9236672, -48.1827026}, "", "192.168.0.23"}

var myid = "bf001"

func scp_update_network() {
	body, _ := json.Marshal(mybf)
	payload := bytes.NewBuffer(body)

	fmt.Println("DEBUG UPDATE NETWORK: Atualizando Netowrk com dados locais")
	net_url := "http://network.hubioagro.com.br/bf_update"
	ret, err := http.Post(net_url, "application/json", payload)
	// fmt.Println("RES=", res)
	defer ret.Body.Close()
	if err != nil {
		checkErr(err)
		return
	}
	fmt.Println("DEBUG SCP UPDATE NETWORK: Retorno da net=", ret)
}

func main() {
	for {
		scp_update_network()
		time.Sleep(1 * time.Minute)
	}

}
