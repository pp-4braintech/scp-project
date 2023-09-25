package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const scp_nonexist = "NONEXIST"

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

var mybf = Biofabrica_data{"bf000", "Nao Configurado", "ERRO", "HA", "Hubio Agro", "", "1.2.15", [2]float64{-15.9236672, -53.1827026}, "", "192.168.0.23"}

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
	if mybf.BFId == "bf000" {
		tun_ip = "192.168.0.23"
		return tun_ip
	}
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

func load_bf_data(filename string) int {
	mybf_new := Biofabrica_data{}
	file, err := os.Open(filename)
	if err != nil {
		checkErr(err)
		return 0
	}
	defer file.Close()
	csvr := csv.NewReader(file)
	n := 0
	for {
		r, err := csvr.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			if perr, ok := err.(*csv.ParseError); ok && perr.Err != csv.ErrFieldCount {
				checkErr(err)
				break
			}
		}
		// fmt.Println(r)
		if r[0][0] != '#' {
			mybf_new.BFId = r[0]
			mybf_new.BFName = r[1]
			mybf_new.Status = r[2]
			mybf_new.CustomerId = r[3]
			mybf_new.CustomerName = r[4]
			mybf_new.Address = r[5]
			mybf_new.SWVersion = r[6]
			lat_str := r[7]
			long_str := r[8]
			lat_f, err_lat := strconv.ParseFloat(lat_str, 64)
			if err_lat != nil {
				checkErr(err_lat)
			}
			long_f, err_long := strconv.ParseFloat(long_str, 64)
			if err_long != nil {
				checkErr(err_long)
			}
			if err_lat == nil && err_long == nil {
				mybf_new.LatLong = [2]float64{lat_f, long_f}
			} else {
				mybf_new.LatLong = [2]float64{0, 0}
			}
			mybf_new.LastUpdate = r[9]
			mybf_new.BFIP = r[10]
			n++
		}
		if n > 0 {
			break
		}
	}
	if n > 0 {
		mybf = mybf_new
	}
	return n
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
	ret_str, _ := ioutil.ReadAll(ret.Body)
	fmt.Println(string(ret_str))
	if strings.Contains(string(ret_str), scp_nonexist) {
		fmt.Println("DEBUG SCP UPDATE NETWORK: Biofabrica nao existe, criando entrada")
		net_url = "http://network.hubioagro.com.br/bf_new"
		payload_new := bytes.NewBuffer(body)
		ret_new, err := http.Post(net_url, "application/json", payload_new)
		if err != nil {
			checkErr(err)
			return
		}
		defer ret_new.Body.Close()
	}
}

func main() {
	// n_bf := load_bf_data("/etc/scpd/bf_data.csv")
	// if n_bf < 1 {
	// 	mybf.BFId = "BFIP-" + get_tun_ip()
	// 	fmt.Println("ERROR SCP AGENT: Arquivo contendo dados da Biofabrica nao encontrado. Usando config padrao", mybf)
	// }
	for {
		n_bf := load_bf_data("/etc/scpd/bf_data.csv")
		if n_bf < 1 {
			mybf.BFId = "BFIP-" + get_tun_ip()
			fmt.Println("ERROR SCP AGENT: Arquivo contendo dados da Biofabrica nao encontrado. Usando config padrao", mybf)
		}
		mybf.BFIP = get_tun_ip()
		scp_update_network()
		time.Sleep(1 * time.Minute)
	}
}
