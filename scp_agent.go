package main

const 	scp_nonexist = "NONEXIST"

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"path/filepath"
	"strings"
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

var mybf = Biofabrica_data{"bf008", "Modelo", "ERRO", "HA", "Hubio Agro", "", "1.2.15", [2]float64{-18.9236672, -48.1827026}, "", "192.168.0.23"}

var myid = "bf001"

func checkErr(err error) {
	if err != nil {
		log.Println("[SCP ERROR]", err)
	}
}

func scp_splitparam(param string, separator string) []string {
	scp_data := strings.Split(param, separator)
	if len(scp_data) < 1 {
		return nil
	}
	return scp_data
}

func get_tun_ip() string {
	tun_ip := ""
	cmdpath, _ := filepath.Abs("/sbin/ifconfig")
	cmd := exec.Command(cmdpath, "tun0") // "| grep 'inet ' | awk '{ print $2}'")
	// cmd := exec.Command(cmdpath)
	// cmd.Dir = "/sbin/"
	output, err := cmd.CombinedOutput()
	if err != nil {
		checkErr(err)
	} else {
		out_str := string(output)
		p := strings.Index(out_str, "inet")
		if p >= 0 {
			ret := scp_splitparam(out_str[p:], " ")
			if len(ret) > 1 {
				tun_ip = ret[1]
			}
		}

	}
	return tun_ip
}

func scp_update_network() {
	body, err := json.Marshal(mybf)
	if err != nil {
		checkErr(err)
	}
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
	if strings.Contains(ret.Body, scp_nonexist) {
		fmt.Println("DEBUG SCP UPDATE NETWORK: Biofabrica nao existe, criando entrada")
		net_url = "http://network.hubioagro.com.br/bf_new"
		ret, err = http.Post(net_url, "application/json", payload)
		if err != nil {
			checkErr(err)
			return
		}
	}
}

func main() {
	for {
		mybf.BFIP = get_tun_ip()
		scp_update_network()
		time.Sleep(1 * time.Minute)
	}

}
