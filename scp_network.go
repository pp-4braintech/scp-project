package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/rs/cors"
)

var net192 = false

const (
	scp_ack = "ACK"
	scp_err = "ERR"
)

var execpath string
var bf_default string = "bf001"

type Biofabrica_data struct {
	BFId         string
	CustomerId   string
	CustomerName string
	SWVersion    string
	LatLong      [2]float64
	LastUpdate   string
	BFIP         string
}

var bfs = []Biofabrica_data{{"bf001", "Unigeo", "Unigeo", "1.2.15", [2]float64{0, 0}, "", "10.10.0.3"}}

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

func test_file(filename string) bool {
	mf, err := os.Stat(filename)
	if err != nil {
		if !os.IsNotExist(err) {
			checkErr(err)
		}
		return false
	}
	fmt.Println("DEBUG: Arquivo encontrado", mf.Name())
	return true
}

func get_bf_index(bf_id string) int {
	if len(bf_id) > 0 {
		for i, v := range bfs {
			if v.BFId == bf_id {
				return i
			}
		}
	}
	return -1
}

func scp_proxy(bfid string, endpoint string) string {
	ind := get_bf_index(bfid)
	if ind < 0 {
		fmt.Println("ERROR SCP PROXY: Biofabrica nao encontrada", bfid)
		return ""
	}
	bf_url := fmt.Sprintf("http://%s/%s", bfs[ind].BFIP, endpoint)
	res, err := http.Get(bf_url)
	fmt.Println("DEBUG RES=", res)
	if err != nil {
		checkErr(err)
		return ""
	}
	// fmt.Println(res)
	rdata, err := ioutil.ReadAll(res.Body)
	if err != nil {
		checkErr(err)
		return ""
	}
	fmt.Println("rdata=", rdata)
	return string(rdata)
	// // fmt.Println(string(rdata))
	// json.Unmarshal(rdata, &last_biofabrica)
	// // fmt.Println(last_biofabrica)
	// fmt.Println("DEBUG CHECK LASTVERSION: Ajustando ultima versão para:", last_biofabrica.Version)
	// biofabrica.LastVersion = last_biofabrica.Version
}

func main_network(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	fmt.Printf("Req: %s %s\n", r.Host, r.URL.Path)

	switch r.Method {
	case "GET":
		var jsonStr []byte
		// var err error
		endpoint := r.URL.Path
		if endpoint == "/" {
			fmt.Println("acessando raiz")
			jsonStr, _ = json.Marshal(bfs)
			os.Stdout.Write(jsonStr)
			w.Write([]byte(jsonStr))
		} else {
			params := scp_splitparam(endpoint, "/")
			fmt.Println("end=", params, len(params))
			if len(params) > 2 {
				cmd := params[1]
				if cmd == "bf_default" {
					ind := get_bf_index(bf_default)
					if ind < 0 {
						w.Write([]byte(scp_err))
					} else {
						jsonStr, _ = json.Marshal(bfs)
						// os.Stdout.Write(jsonStr)
						w.Write([]byte(jsonStr))
					}
				} else {
					ret := scp_proxy(bf_default, endpoint)
					w.Write([]byte(ret))
				}

			}
		}
		// totem_id := r.URL.Query().Get("Id")
		//fmt.Println("bio_id =", bio_id)
		//fmt.Println()
		// if len(totem_id) > 0 {
		// 	cmd := "GET/" + scp_totem + "/" + totem_id + "/END"
		// 	jsonStr = []byte(scp_sendmsg_master(cmd))
		// } else {
		// 	cmd := "GET/" + scp_totem + "/END"
		// 	jsonStr = []byte(scp_sendmsg_master(cmd))
		// }
		//os.Stdout.Write(jsonStr)
		//jsonStr = []byte(scp_sendmsg_master(cmd))
		// os.Stdout.Write(jsonStr)
		// w.Write([]byte(jsonStr))

	}
	// case "PUT":
	// 	err := r.ParseForm()
	// 	if err != nil {
	// 		fmt.Println("ParseForm() err: ", err)
	// 		return
	// 	}
	// 	// fmt.Println("Post from website! r.PostFrom = ", r.PostForm)
	// 	// fmt.Println("Post Data", r.Form)
	// 	totem_id := r.FormValue("Id")
	// 	if len(totem_id) >= 0 {
	// 		pump := r.FormValue("Pump")
	// 		peris := r.FormValue("Perist")
	// 		valve := r.FormValue("Valve")
	// 		value_status := r.FormValue("Status")
	// 		// fmt.Println("Pump = ", pump)
	// 		// fmt.Println("Valve = ", valve)
	// 		// fmt.Println("Status = ", valve_status)
	// 		if pump != "" {
	// 			cmd := "PUT/" + scp_totem + "/" + totem_id + "/" + scp_dev_pump + "," + pump + "/END"
	// 			jsonStr := []byte(scp_sendmsg_master(cmd))
	// 			// os.Stdout.Write(jsonStr)
	// 			w.Write([]byte(jsonStr))
	// 		}
	// 		if valve != "" {
	// 			cmd := "PUT/" + scp_totem + "/" + totem_id + "/" + scp_dev_valve + "," + valve + "," + value_status + "/END"
	// 			jsonStr := []byte(scp_sendmsg_master(cmd))
	// 			// os.Stdout.Write(jsonStr)
	// 			w.Write([]byte(jsonStr))
	// 		}
	// 		if peris != "" {
	// 			cmd := "PUT/" + scp_totem + "/" + totem_id + "/" + scp_dev_peris + "," + peris + "," + value_status + "/END"
	// 			jsonStr := []byte(scp_sendmsg_master(cmd))
	// 			// os.Stdout.Write(jsonStr)
	// 			w.Write([]byte(jsonStr))
	// 		}
	// 	}

	// default:

	// }
}

func main() {

	net192 = test_file("/etc/scpd/scp_net192.flag")
	if net192 {
		fmt.Println("WARN:  EXECUTANDO EM NET192\n\n\n")
		execpath = "/home/paulo/scp-project/"
	} else {
		execpath = "/home/scpadm/scp-project/"
	}

	mux := http.NewServeMux()
	cors := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{
			http.MethodPut,
			http.MethodPost,
			http.MethodGet,
		},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: false,
	})

	mux.HandleFunc("/", main_network)

	// mux.HandleFunc("/ibc_view", ibc_view)

	// mux.HandleFunc("/totem_view", totem_view)

	// mux.HandleFunc("/biofabrica_view", biofabrica_view)

	// mux.HandleFunc("/simulator", biofactory_sim)

	// mux.HandleFunc("/config", set_config)

	// mux.HandleFunc("/wdpanel", withdraw_panel)

	handler := cors.Handler(mux)

	http.ListenAndServe(":7077", handler)
}
