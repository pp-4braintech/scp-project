package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/rs/cors"
)

const scp_err = "ERR"
const scp_dev_pump = "PUMP"
const scp_dev_aero = "AERO"
const scp_dev_valve = "VALVE"

const bio_nonexist = "NULL"
const bio_cip = "CIP"
const bio_loading = "CARREGANDO"
const bio_unloading = "ESVAZIANDO"
const bio_producting = "PRODUZINDO"
const bio_empty = "VAZIO"
const bio_done = "CONCLUIDO"
const bio_error = "ERRO"
const bio_max_valves = 8
const max_buf = 8192

type Bioreact struct {
	BioreactorID string
	Status       string
	Organism     string
	Volume       uint32
	Level        uint8
	Pumpstatus   bool
	Aerator      bool
	Valvs        [8]int
	Temperature  float32
	PH           float32
	Step         [2]int
	Timeleft     [2]int
	Timetotal    [2]int
}

type IBC struct {
	IBCID      string
	Status     string
	Organism   string
	Volume     uint32
	Level      uint8
	Pumpstatus bool
	Valvs      [4]int
}

var bio = []Bioreact{
	{"BIOR001", bio_producting, "Bacillus Subtilis", 100, 10, false, true, [8]int{1, 1, 0, 0, 0, 0, 0, 0}, 28, 7, [2]int{2, 5}, [2]int{25, 17}, [2]int{48, 0}},
	{"BIOR002", bio_cip, "Bacillus Megaterium", 200, 5, false, false, [8]int{0, 0, 1, 0, 0, 1, 0, 1}, 26, 7, [2]int{1, 1}, [2]int{0, 5}, [2]int{0, 30}},
	{"BIOR003", bio_loading, "Bacillus Amyloliquefaciens", 1000, 3, false, false, [8]int{0, 0, 0, 1, 0, 0, 1, 0}, 28, 7, [2]int{1, 1}, [2]int{0, 10}, [2]int{0, 30}},
	{"BIOR004", bio_unloading, "Azospirilum brasiliense", 500, 5, true, false, [8]int{0, 0, 0, 0, 1, 1, 0, 0}, 25, 7, [2]int{1, 1}, [2]int{0, 5}, [2]int{0, 15}},
	{"BIOR005", bio_done, "Tricoderma harzianum", 0, 10, false, false, [8]int{2, 0, 0, 0, 0, 0, 0, 0}, 28, 7, [2]int{5, 5}, [2]int{0, 0}, [2]int{72, 0}},
	{"BIOR006", bio_nonexist, "", 0, 0, false, false, [8]int{0, 0, 0, 0, 0, 0, 0, 0}, 0, 0, [2]int{0, 0}, [2]int{0, 0}, [2]int{0, 0}},
}

var ibc = []IBC{
	{"IBC01", bio_done, "Bacillus Subtilis", 100, 1, false, [4]int{0, 0, 0, 0}},
	{"IBC02", bio_done, "Bacillus Megaterium", 200, 1, false, [4]int{0, 0, 0, 0}},
	{"IBC03", bio_loading, "Bacillus Amyloliquefaciens", 1000, 3, false, [4]int{0, 0, 0, 0}},
	{"IBC04", bio_unloading, "Azospirilum brasiliense", 500, 2, false, [4]int{0, 0, 0, 0}},
	{"IBC05", bio_done, "Tricoderma harzianum", 1000, 3, false, [4]int{0, 0, 0, 0}},
	{"IBC06", bio_cip, "Tricoderma harzianum", 2000, 5, true, [4]int{0, 0, 0, 0}},
	{"IBC07", bio_empty, "", 0, 0, false, [4]int{0, 0, 0, 0}},
}

func checkErr(err error) {
	if err != nil {
		log.Println("[SCP ERROR]", err)
	}
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
	fmt.Printf("recebido: %s\n", buf[:n])
	return string(buf[:n])
}

func get_bio_index(bio_id string) int {
	if len(bio_id) > 0 {
		for i, v := range bio {
			if v.BioreactorID == bio_id {
				return i
			}
		}
	}
	return -1
}

func ibc_view(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// fmt.Println("bio", ibc)
	fmt.Println("Request", r)
	switch r.Method {
	case "GET":
		var jsonStr []byte
		ibc_id := r.URL.Query().Get("Id")
		fmt.Println("ibc_id =", ibc_id)
		fmt.Println()
		if len(ibc_id) > 0 {
			cmd := "GET/IBC/" + ibc_id + "/END"
			jsonStr = []byte(scp_sendmsg_master(cmd))
		} else {
			cmd := "GET/IBC/END"
			jsonStr = []byte(scp_sendmsg_master(cmd))
		}
		os.Stdout.Write(jsonStr)
		w.Write([]byte(jsonStr))
	default:

	}
}

func bioreactor_view(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// fmt.Println("bio", bio)
	fmt.Println("Request", r)
	switch r.Method {
	case "GET":
		var jsonStr []byte
		bio_id := r.URL.Query().Get("Id")
		fmt.Println("bio_id =", bio_id)
		fmt.Println()
		if len(bio_id) > 0 {
			cmd := "GET/BIOREACTOR/" + bio_id + "/END"
			jsonStr = []byte(scp_sendmsg_master(cmd))
		} else {
			cmd := "GET/BIOREACTOR/END"
			jsonStr = []byte(scp_sendmsg_master(cmd))
		}
		os.Stdout.Write(jsonStr)
		w.Write([]byte(jsonStr))
	case "PUT":
		err := r.ParseForm()
		if err != nil {
			fmt.Println("ParseForm() err: ", err)
			return
		}
		fmt.Println("Post from website! r.PostFrom = ", r.PostForm)
		fmt.Println("Post Data", r.Form)
		// for i, d := range r.Form {
		// 	fmt.Println(i, d)
		// }
		bio_id := r.FormValue("Id")
		ind := get_bio_index(bio_id)
		if ind >= 0 {
			pump := r.FormValue("Pumpstatus")
			aero := r.FormValue("Aerator")
			valve := r.FormValue("Valve")
			valve_status := r.FormValue("Status")
			fmt.Println("Bio_id = ", bio_id)
			fmt.Println("Pump = ", pump)
			fmt.Println("Aero = ", aero)
			fmt.Println("Valve = ", valve)
			fmt.Println("Status = ", valve_status)
			if pump != "" {
				cmd := "PUT/BIOREACTOR/" + bio_id + "/" + scp_dev_pump + "," + pump + "/END"
				jsonStr := []byte(scp_sendmsg_master(cmd))
				w.Write([]byte(jsonStr))
			}
			if aero != "" {
				cmd := "PUT/BIOREACTOR/" + bio_id + "/" + scp_dev_aero + "," + aero + "/END"
				jsonStr := []byte(scp_sendmsg_master(cmd))
				w.Write([]byte(jsonStr))
			}
			if valve != "" {
				cmd := "PUT/BIOREACTOR/" + bio_id + "/" + scp_dev_valve + "," + valve + "," + valve_status + "/END"
				jsonStr := []byte(scp_sendmsg_master(cmd))
				w.Write([]byte(jsonStr))

			}
		}

	default:
		fmt.Fprintf(w, "Sorry, only GET and POST methods are supported.")
	}
	fmt.Println()
	fmt.Println()
}

func scp_bio_init() {
	fmt.Println("Iniciando MOD")
	for i := 2; i < 11; i++ {
		scp_msg := fmt.Sprintf("CMD/55:3A7D80/MOD/%d,1/END", i)
		fmt.Println("CMD=", scp_msg)
		scp_sendmsg_master(scp_msg)
	}

}

func main() {

	scp_bio_init()
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

	mux.HandleFunc("/bioreactor_view", bioreactor_view)

	mux.HandleFunc("/ibc_view", ibc_view)

	handler := cors.Handler(mux)

	http.ListenAndServe(":5000", handler)
}
