package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const scp_err = "ERR"
const scp_nonexist = "NONEXIST"
const scp_config = "CONFIG"
const scp_biofabrica = "BIOFABRICA"
const scp_par_loadbfdata = "LOADBFDATA"

const max_buf = 8192

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

func scp_sendmsg_master(cmd string) string {

	ipc, err := net.Dial("unix", "/tmp/scp_master.sock")
	if err != nil {
		checkErr(err)
		return scp_err
	}
	defer ipc.Close()

	_, err = ipc.Write([]byte(cmd))
	if err != nil {
		checkErr(err)
		return scp_err
	}

	buf := make([]byte, max_buf)
	n, errf := ipc.Read(buf)
	if errf != nil {
		checkErr(err)
	}
	//fmt.Printf("recebido: %s\n", buf[:n])
	return string(buf[:n])
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

func save_bf_data(filename string) int {
	filecsv, err := os.Create(filename)
	if err != nil {
		checkErr(err)
		return -1
	}
	defer filecsv.Close()
	n := 0
	csvwriter := csv.NewWriter(filecsv)
	thisbf := mybf
	s := fmt.Sprintf("%s,%s,%s,%s,%s,%s,%s,%f,%f,%s,%s", thisbf.BFId, thisbf.BFName, thisbf.Status, thisbf.CustomerId,
		thisbf.CustomerName, thisbf.Address, thisbf.SWVersion, thisbf.LatLong[0], thisbf.LatLong[1],
		thisbf.LastUpdate, thisbf.BFIP)
	csvstr := scp_splitparam(s, ",")
	// fmt.Println("DEBUG SAVE", csvstr)
	err = csvwriter.Write(csvstr)
	if err != nil {
		checkErr(err)
	} else {
		n++
	}
	csvwriter.Flush()
	return n
}

func get_uuid() string {
	out, err := exec.Command("uuidgen").Output()
	if err != nil {
		checkErr(err)
		return ""
	}
	fmt.Printf("DEBUG GET UUID: Gerando ID unico = %s", out)
	uuid := out[:36]
	return string(uuid)
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
	myID := ""
	for {
		n_bf := load_bf_data("/etc/scpd/bf_data.csv")
		if n_bf < 1 {
			mybf.BFId = get_uuid()
			myID = mybf.BFId
			fmt.Println("ERROR SCP AGENT: Arquivo contendo dados da Biofabrica nao encontrado. Usando config padrao e criando arquivo", mybf)
			save_bf_data("/etc/scpd/bf_data.csv")
		} else if strings.Contains(mybf.BFId, "BFIP") && len(myID) == 0 {
			mybf.BFId = get_uuid()
			myID = mybf.BFId
			fmt.Println("ERROR SCP AGENT: Nome da Biofabrica obsoleto. Criando ID aleatorio e regravando configuracoes", mybf)
			// save_bf_data("/etc/scpd/bf_data.csv")
			save_bf_data("/etc/scpd/bf_data_new.csv")
			cmd := scp_config + "/" + scp_biofabrica + "/" + scp_par_loadbfdata + "/END"
			scp_sendmsg_master(cmd)
		}
		mybf.BFIP = get_tun_ip()
		scp_update_network()
		time.Sleep(1 * time.Minute)
	}
}
