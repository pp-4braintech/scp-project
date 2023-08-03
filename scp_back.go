package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/rs/cors"
)

// var demo = false
var devmode = false

// var testmode = false

const scp_err = "ERR"
const scp_ack = "ACK"
const scp_dev_pump = "PUMP"
const scp_dev_aero = "AERO"
const scp_dev_valve = "VALVE"
const scp_dev_peris = "PERIS"
const scp_biofabrica = "BIOFABRICA"
const scp_totem = "TOTEM"
const scp_bioreactor = "BIOREACTOR"
const scp_wdpanel = "WDPANEL"
const scp_config = "CONFIG"

const scp_par_withdraw = "WITHDRAW"
const scp_par_out = "OUT"
const scp_par_ph4 = "PH4"
const scp_par_ph7 = "PH7"
const scp_par_ph10 = "PH10"
const scp_par_calibrate = "CALIBRATE"
const scp_par_save = "SAVE"
const scp_par_restart = "RESTART"

const scp_sched = "SCHED"
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

// const execpath = "./"

const vol_bioreactor = 2000
const vol_ibc = 4000
const overhead = 1.1
const max_bios = 36
const max_days = 60

// type Bioreact struct {
// 	BioreactorID string
// 	Status       string
// 	Organism     string
// 	Volume       uint32
// 	Level        uint8
// 	Pumpstatus   bool
// 	Aerator      bool
// 	Valvs        [8]int
// 	Temperature  float32
// 	PH           float32
// 	Step         [2]int
// 	Timeleft     [2]int
// 	Timetotal    [2]int
// }

// type IBC struct {
// 	IBCID      string
// 	Status     string
// 	Organism   string
// 	Volume     uint32
// 	Level      uint8const scp_biofabrica = "BIOFABRICA"
// 	Pumpstatus bool
// 	Valvs      [4]int
// }

type Organism struct {
	Code       string
	Orgname    string
	Lifetime   int
	Prodvol    int
	Cultmedium string
	Timetotal  int
	Aero       [3]int
	PH         [3]string
}

type BioList struct {
	OrganismName string
	Code         string
	Selected     bool
}

type Prodlist struct {
	Bioid  string
	Values []int
	Codes  []string
}

var orgs []Organism
var lastsched []Prodlist
var execpath string

func checkErr(err error) {
	if err != nil {
		log.Println("[SCP ERROR]", err)
	}
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

func bio_to_code(bioname string) string {
	n := len(bioname)
	if n < 1 {
		return ""
	}
	biosplit := strings.Split(bioname, " ")
	nick := ""
	for _, k := range biosplit {
		nick += string(k[0])
	}
	return nick
}

func get_first_bio_available(prod [max_bios][max_days]int, maxbio int, maxday int) (int, int) {
	nbio := -1
	nday := -1
	for i := 0; i < maxbio; i++ {
		for j := 0; j < maxday; j++ {
			if prod[i][j] < 0 {
				if nday < 0 || j < nday {
					nday = j
					nbio = i
				}
			}
		}
	}
	return nbio, nday
}

func load_organisms(filename string) int {
	var totalrecords int
	file, err := os.Open(filename)
	if err != nil {
		checkErr(err)
		return -1
	}
	defer file.Close()
	records, err := csv.NewReader(file).ReadAll()
	if err != nil {
		checkErr(err)
		return -1
	}
	orgs = make([]Organism, len(records))
	for k, r := range records {
		code := r[0]
		name := r[1]
		lifetime, _ := strconv.Atoi(strings.Replace(r[2], " ", "", -1))
		volume, _ := strconv.Atoi(strings.Replace(r[3], " ", "", -1))
		medium := strings.Replace(r[4], " ", "", -1)
		tottime, _ := strconv.Atoi(strings.Replace(r[5], " ", "", -1))
		aero1, _ := strconv.Atoi(strings.Replace(r[6], " ", "", -1))
		aero2, _ := strconv.Atoi(strings.Replace(r[7], " ", "", -1))
		aero3, _ := strconv.Atoi(strings.Replace(r[8], " ", "", -1))
		ph1 := strings.Replace(r[9], " ", "", -1)
		ph2 := strings.Replace(r[10], " ", "", -1)
		ph3 := strings.Replace(r[11], " ", "", -1)
		org := Organism{code, name, lifetime, volume, medium, tottime, [3]int{aero1, aero2, aero3}, [3]string{ph1, ph2, ph3}}
		orgs[k] = org
		totalrecords = k
	}
	return totalrecords
}

func min_bio_sim(farmarea int, dailyarea int, orglist []BioList) (int, int, int, int, []Prodlist) {
	var total int
	var totalorg, totaltime int32
	var op, ot map[int]int32
	var o []int
	total = 0
	totalorg = 0
	totaltime = 0
	o = []int{}
	op = make(map[int]int32)
	ot = make(map[int]int32)

	for k, r := range orglist {
		if r.Selected {
			o = append(o, k)
			op[k] = int32(orgs[k].Prodvol * farmarea)
			totalorg += op[k]
			ot[k] = op[k] * int32(orgs[k].Timetotal)
			totaltime += ot[k]
			fmt.Println(orgs[k].Orgname, op[k], ot[k])
		}
	}
	fmt.Println("Volume total =", totalorg)
	fmt.Println("Tempo total =", totaltime)
	ndias := int(farmarea / dailyarea)
	fmt.Println("Numero dias =", ndias)
	total = int(math.Ceil(float64((float64(totaltime) * overhead) / float64(ndias*24*vol_bioreactor))))
	fmt.Println("Numero bioreatores =", total)
	fmt.Println("Organismos:", o)
	fmt.Println("Producao :", op)

	if ndias > max_days || total > max_bios {
		fmt.Println("numero maximo de dias ou bio excedido")
		return ndias, total, 0, 0, nil
	}
	var prodm [max_bios][max_days]int

	for i := 0; i < max_bios; i++ {
		for j := 0; j < max_days; j++ {
			prodm[i][j] = -1
		}
	}
	//	prodm = make(map[int][int]int)
	// i := 0
	d := 0
	b := 0
	// d0 := 0
	n := 0
	fday := -1
	haschange := true
	for d < ndias && haschange {
		//fmt.Println(prodm)
		haschange = false

		// d = 0
		// for d0 = 0; d0 < ndias; d0++ {
		// 	if prodm[b][d0] == 0 {
		// 		d = d0
		// 		break
		// 	}
		// }
		b, d = get_first_bio_available(prodm, total, ndias)
		if b < 0 || d < 0 {
			fmt.Println("Nao ha slot de producao disponivel")
			break
		}
		//fmt.Println("bio=", b, "dia=", d, "org=", n, " fday=", fday)
		for {
			if op[o[n]] > 0 {
				for i := 0; i < int(orgs[o[n]].Timetotal/24); i++ {
					//fmt.Print("dia=", d, " org=", n, " time=", orgs[o[n]].Timetotal, " prod=", op[o[n]])
					prodm[b][d] = o[n]
					proday := int32(math.Ceil(float64(vol_bioreactor*24) / float64(orgs[o[n]].Timetotal)))
					//fmt.Println(" proday=", proday)
					op[o[n]] -= proday
					d++
					haschange = true
				}
			}
			n++
			if n == len(o) {
				if fday < 0 {
					fday = d
				}
				n = 0
				if !haschange {
					break
				}
			}
			if haschange {
				break
			}
		}
		if d >= ndias {
			break
		}

	}

	//fmt.Println(prodm)

	max := 0
	v := make([]Prodlist, 0)
	for k, x := range prodm {
		var tmpcode []string
		tmpcode = []string{}
		var tmpnum []int
		tmpnum = []int{}
		if k < total {
			fmt.Printf("Bio%02d  ", k+1)
			for j, y := range x {
				if y >= 0 {
					fmt.Printf("%2d ", y)
					tmpcode = append(tmpcode, orgs[y].Code) // bio_to_code(orgs[y].Orgname)
					tmpnum = append(tmpnum, y)
					if j > max {
						max = j
					}
				}
			}
			fmt.Println()
			bioid := fmt.Sprintf("BIOR%02d", k+1)
			v = append(v, Prodlist{bioid, tmpnum, tmpcode})
		}
	}
	prodias := max + 1
	fmt.Println("Dias de Producao =", prodias)
	fmt.Println("Primeiro dia =", fday)
	//var jsonStr []byte
	//jsonStr, err := json.Marshal(prodm)
	//checkErr(err)
	//fmt.Println(prodm)
	//fmt.Println(v)
	//fmt.Println(jsonStr)

	// jsonStr, err := json.Marshal(v)
	// checkErr(err)
	// os.Stdout.Write(jsonStr)
	return ndias, total, prodias, fday, v
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

func ibc_view(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// fmt.Println("bio", ibc)
	//fmt.Println("Request", r)
	switch r.Method {
	case "GET":
		var jsonStr []byte
		ibc_id := r.URL.Query().Get("Id")
		//fmt.Println("ibc_id =", ibc_id)
		//fmt.Println()
		if len(ibc_id) > 0 {
			cmd := "GET/IBC/" + ibc_id + "/END"
			jsonStr = []byte(scp_sendmsg_master(cmd))
		} else {
			cmd := "GET/IBC/END"
			jsonStr = []byte(scp_sendmsg_master(cmd))
		}
		// os.Stdout.Write(jsonStr)
		w.Write([]byte(jsonStr))

	case "PUT":
		err := r.ParseForm()
		if err != nil {
			fmt.Println("ParseForm() err: ", err)
			return
		}
		// fmt.Println("Post from website! r.PostFrom = ", r.PostForm)
		// fmt.Println("Post Data", r.Form)
		// for i, d := range r.Form {
		// 	fmt.Println(i, d)
		// }
		ibc_id := r.FormValue("Id")
		if len(ibc_id) >= 0 {
			pump := r.FormValue("Pumpstatus")
			valve := r.FormValue("Valve")
			valve_status := r.FormValue("Status")
			withdraw := r.FormValue("Withdraw")
			outdev := r.FormValue("Out")
			b_pause := r.FormValue("Pause")
			b_stop := r.FormValue("Stop")
			b_start := r.FormValue("Start")
			orgcode := r.FormValue("OrgCode")

			if pump != "" {
				cmd := "PUT/IBC/" + ibc_id + "/" + scp_dev_pump + "," + pump + "/END"
				jsonStr := []byte(scp_sendmsg_master(cmd))
				w.Write([]byte(jsonStr))
			}
			if valve != "" {
				cmd := "PUT/IBC/" + ibc_id + "/" + scp_dev_valve + "," + valve + "," + valve_status + "/END"
				jsonStr := []byte(scp_sendmsg_master(cmd))
				w.Write([]byte(jsonStr))
			}
			if outdev != "" {
				// fmt.Println("PAR OUT", outdev)
				cmd := "PUT/IBC/" + ibc_id + "/" + scp_par_out + "," + outdev + "/END"
				jsonStr := []byte(scp_sendmsg_master(cmd))
				w.Write([]byte(jsonStr))
			}
			if withdraw != "" {
				cmd := "PUT/IBC/" + ibc_id + "/" + scp_par_withdraw + "," + withdraw + "/END"
				jsonStr := []byte(scp_sendmsg_master(cmd))
				w.Write([]byte(jsonStr))
			}
			if b_start != "" {
				if orgcode != "" {
					cmd := "START/IBC/" + ibc_id + "/" + orgcode + "/END"
					jsonStr := []byte(scp_sendmsg_master(cmd))
					w.Write([]byte(jsonStr))
				} else {
					fmt.Println("ERROR IBC VIEW: Start faltando orgcode", r)
				}
			}
			if b_pause != "" {
				cmd := "PAUSE/IBC/" + ibc_id + "/" + b_pause + "/END"
				jsonStr := []byte(scp_sendmsg_master(cmd))
				w.Write([]byte(jsonStr))
			}
			if b_stop != "" {
				cmd := "STOP/IBC/" + ibc_id + "/END"
				jsonStr := []byte(scp_sendmsg_master(cmd))
				w.Write([]byte(jsonStr))
			}

		}

	default:

	}
}

func bioreactor_view(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// fmt.Println("bio", bio)
	//fmt.Println("Request", r)
	switch r.Method {
	case "GET":
		var jsonStr []byte
		bio_id := r.URL.Query().Get("Id")
		//fmt.Println("bio_id =", bio_id)
		//fmt.Println()
		if len(bio_id) > 0 {
			cmd := "GET/BIOREACTOR/" + bio_id + "/END"
			jsonStr = []byte(scp_sendmsg_master(cmd))
		} else {
			cmd := "GET/BIOREACTOR/END"
			jsonStr = []byte(scp_sendmsg_master(cmd))
		}
		//os.Stdout.Write(jsonStr)
		w.Write([]byte(jsonStr))
	case "PUT":
		err := r.ParseForm()
		if err != nil {
			fmt.Println("ParseForm() err: ", err)
			return
		}
		//fmt.Println("Put from website! r.PostFrom = ", r.PostForm)
		//fmt.Println("Put Data", r.Form)
		// for i, d := range r.Form {
		// 	fmt.Println(i, d)
		// }
		bio_id := r.FormValue("Id")
		if len(bio_id) >= 0 {
			pump := r.FormValue("Pumpstatus")
			aero := r.FormValue("Aerator")
			valve := r.FormValue("Valve")
			peris := r.FormValue("Perist")
			b_pause := r.FormValue("Pause")
			b_stop := r.FormValue("Stop")
			b_start := r.FormValue("Start")
			orgcode := r.FormValue("OrgCode")
			value_status := r.FormValue("Status")
			withdraw := r.FormValue("Withdraw")
			outdev := r.FormValue("Out")

			if b_start != "" {
				if orgcode != "" {
					cmd := "START/BIOREACTOR/" + bio_id + "/" + orgcode + "/END"
					jsonStr := []byte(scp_sendmsg_master(cmd))
					w.Write([]byte(jsonStr))
				} else {
					fmt.Println("ERROR BIOREACTOR VIEW: Start faltando orgcode", r)
				}
			}
			if b_pause != "" {
				cmd := "PAUSE/BIOREACTOR/" + bio_id + "/" + b_pause + "/END"
				jsonStr := []byte(scp_sendmsg_master(cmd))
				w.Write([]byte(jsonStr))
			}
			if b_stop != "" {
				cmd := "STOP/BIOREACTOR/" + bio_id + "/END"
				jsonStr := []byte(scp_sendmsg_master(cmd))
				w.Write([]byte(jsonStr))
			}

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
				cmd := "PUT/BIOREACTOR/" + bio_id + "/" + scp_dev_valve + "," + valve + "," + value_status + "/END"
				jsonStr := []byte(scp_sendmsg_master(cmd))
				w.Write([]byte(jsonStr))

			}
			if outdev != "" {
				fmt.Println("PAR OUT", outdev)
				cmd := "PUT/BIOREACTOR/" + bio_id + "/" + scp_par_out + "," + outdev + "/END"
				jsonStr := []byte(scp_sendmsg_master(cmd))
				w.Write([]byte(jsonStr))
			}
			if withdraw != "" {
				cmd := "PUT/BIOREACTOR/" + bio_id + "/" + scp_par_withdraw + "," + withdraw + "/END"
				jsonStr := []byte(scp_sendmsg_master(cmd))
				w.Write([]byte(jsonStr))
			}
			if peris != "" {
				cmd := "PUT/BIOREACTOR/" + bio_id + "/" + scp_dev_peris + "," + peris + "," + value_status + "/END"
				jsonStr := []byte(scp_sendmsg_master(cmd))
				// os.Stdout.Write(jsonStr)
				w.Write([]byte(jsonStr))
			}
		}

	default:
		fmt.Fprintf(w, "Sorry, only GET and POST methods are supported.")
	}
	// fmt.Println()
	// fmt.Println()
}

func totem_view(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case "GET":
		var jsonStr []byte
		totem_id := r.URL.Query().Get("Id")
		//fmt.Println("bio_id =", bio_id)
		//fmt.Println()
		if len(totem_id) > 0 {
			cmd := "GET/" + scp_totem + "/" + totem_id + "/END"
			jsonStr = []byte(scp_sendmsg_master(cmd))
		} else {
			cmd := "GET/" + scp_totem + "/END"
			jsonStr = []byte(scp_sendmsg_master(cmd))
		}
		//os.Stdout.Write(jsonStr)
		//jsonStr = []byte(scp_sendmsg_master(cmd))
		// os.Stdout.Write(jsonStr)
		w.Write([]byte(jsonStr))

	case "PUT":
		err := r.ParseForm()
		if err != nil {
			fmt.Println("ParseForm() err: ", err)
			return
		}
		// fmt.Println("Post from website! r.PostFrom = ", r.PostForm)
		// fmt.Println("Post Data", r.Form)
		totem_id := r.FormValue("Id")
		if len(totem_id) >= 0 {
			pump := r.FormValue("Pump")
			peris := r.FormValue("Perist")
			valve := r.FormValue("Valve")
			value_status := r.FormValue("Status")
			// fmt.Println("Pump = ", pump)
			// fmt.Println("Valve = ", valve)
			// fmt.Println("Status = ", valve_status)
			if pump != "" {
				cmd := "PUT/" + scp_totem + "/" + totem_id + "/" + scp_dev_pump + "," + pump + "/END"
				jsonStr := []byte(scp_sendmsg_master(cmd))
				// os.Stdout.Write(jsonStr)
				w.Write([]byte(jsonStr))
			}
			if valve != "" {
				cmd := "PUT/" + scp_totem + "/" + totem_id + "/" + scp_dev_valve + "," + valve + "," + value_status + "/END"
				jsonStr := []byte(scp_sendmsg_master(cmd))
				// os.Stdout.Write(jsonStr)
				w.Write([]byte(jsonStr))
			}
			if peris != "" {
				cmd := "PUT/" + scp_totem + "/" + totem_id + "/" + scp_dev_peris + "," + peris + "," + value_status + "/END"
				jsonStr := []byte(scp_sendmsg_master(cmd))
				// os.Stdout.Write(jsonStr)
				w.Write([]byte(jsonStr))
			}
		}

	default:

	}
}

func biofabrica_view(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case "GET":
		var jsonStr []byte
		cmd := "GET/BIOFABRICA/END"
		jsonStr = []byte(scp_sendmsg_master(cmd))
		// os.Stdout.Write(jsonStr)
		w.Write([]byte(jsonStr))

	case "PUT":
		err := r.ParseForm()
		if err != nil {
			fmt.Println("ParseForm() err: ", err)
			return
		}
		// fmt.Println("Post from website! r.PostFrom = ", r.PostForm)
		// fmt.Println("Post Data", r.Form)

		pump := r.FormValue("Pumpwithdraw")
		valve := r.FormValue("Valve")
		valve_status := r.FormValue("Status")
		// fmt.Println("Pumpwithdraw = ", pump)
		// fmt.Println("Valve = ", valve)
		// fmt.Println("Status = ", valve_status)
		if pump != "" {
			cmd := "PUT/" + scp_biofabrica + "/" + scp_dev_pump + "," + pump + "/END"
			jsonStr := []byte(scp_sendmsg_master(cmd))
			// os.Stdout.Write(jsonStr)
			w.Write([]byte(jsonStr))
		}
		if valve != "" {
			cmd := "PUT/" + scp_biofabrica + "/" + scp_dev_valve + "," + valve + "," + valve_status + "/END"
			jsonStr := []byte(scp_sendmsg_master(cmd))
			// os.Stdout.Write(jsonStr)
			w.Write([]byte(jsonStr))
		}

	default:

	}
}

func set_config(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case "GET":
		w.Write([]byte("NAO IMPLEMENTADO"))
		// var jsonStr []byte
		// cmd := "GET/BIOFABRICA/END"
		// jsonStr = []byte(scp_sendmsg_master(cmd))
		// // os.Stdout.Write(jsonStr)
		// w.Write([]byte(jsonStr))

	case "PUT":
		err := r.ParseForm()
		if err != nil {
			fmt.Println("ParseForm() err: ", err)
			return
		}
		bioid := r.FormValue("BioId")
		ph4 := r.FormValue("PH4")
		ph7 := r.FormValue("PH7")
		ph10 := r.FormValue("PH10")
		calibrate := r.FormValue("Calibrate")
		saveconfig := r.FormValue("SaveConfig")
		restart := r.FormValue("Restart")

		if len(bioid) > 0 {
			if ph4 != "" {
				cmd := scp_config + "/" + scp_bioreactor + "/" + bioid + "/" + scp_par_ph4 + "/END"
				jsonStr := []byte(scp_sendmsg_master(cmd))
				// os.Stdout.Write(jsonStr)
				w.Write([]byte(jsonStr))
			}

			if ph7 != "" {
				cmd := scp_config + "/" + scp_bioreactor + "/" + bioid + "/" + scp_par_ph7 + "/END"
				jsonStr := []byte(scp_sendmsg_master(cmd))
				// os.Stdout.Write(jsonStr)
				w.Write([]byte(jsonStr))
			}

			if ph10 != "" {
				cmd := scp_config + "/" + scp_bioreactor + "/" + bioid + "/" + scp_par_ph10 + "/END"
				jsonStr := []byte(scp_sendmsg_master(cmd))
				// os.Stdout.Write(jsonStr)
				w.Write([]byte(jsonStr))
			}

			if calibrate != "" {
				cmd := scp_config + "/" + scp_bioreactor + "/" + bioid + "/" + scp_par_calibrate + "/END"
				jsonStr := []byte(scp_sendmsg_master(cmd))
				// os.Stdout.Write(jsonStr)
				w.Write([]byte(jsonStr))
			}
		}

		if saveconfig != "" {
			cmd := scp_config + "/" + scp_biofabrica + "/" + scp_par_save + "/END"
			jsonStr := []byte(scp_sendmsg_master(cmd))
			// os.Stdout.Write(jsonStr)
			w.Write([]byte(jsonStr))
		}

		if restart != "" {
			cmd := scp_config + "/" + scp_biofabrica + "/" + scp_par_restart + "/END"
			jsonStr := []byte(scp_sendmsg_master(cmd))
			// os.Stdout.Write(jsonStr)
			w.Write([]byte(jsonStr))
		}

	default:

	}
}

func biofactory_sim(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	orgdata := make([]BioList, len(orgs))
	for k, r := range orgs {
		orgdata[k].OrganismName = r.Orgname
		orgdata[k].Code = r.Code
		orgdata[k].Selected = false
	}
	orgcip := BioList{"CIP", "CIP", false}
	orgdata = append(orgdata, orgcip)
	// fmt.Println("bio", bio)
	switch r.Method {
	case "GET":
		var jsonStr []byte
		jsonStr, err := json.Marshal(orgdata)
		checkErr(err)
		//os.Stdout.Write(jsonStr)
		w.Write([]byte(jsonStr))
	case "PUT":
		err := r.ParseForm()
		if err != nil {
			fmt.Println("ParseForm() err: ", err)
			return
		}
		exec_str := r.FormValue("Execute")
		exec_cmd, _ := strconv.ParseBool(exec_str)
		if exec_cmd {
			s_str := ""
			for _, s := range lastsched {
				for k, c := range s.Codes {
					seq := fmt.Sprintf("%d", k)
					s_str += s.Bioid + "," + seq + "," + c + "/"
				}
			}
			s_str += "END"
			cmd := scp_sched + "/" + scp_biofabrica + "/" + s_str
			fmt.Println("DEBUG SIM: to master", cmd)
			jsonStr := []byte(scp_sendmsg_master(cmd))
			os.Stdout.Write(jsonStr)
			w.Write([]byte(jsonStr))
		}
		farm_area_form := r.FormValue("Farmarea")
		farm_area, _ := strconv.Atoi(farm_area_form)
		daily_area_form := r.FormValue("Dailyarea")
		daily_area, _ := strconv.Atoi(daily_area_form)
		org_sel_form := r.FormValue("Organismsel")
		fmt.Println(farm_area, daily_area, org_sel_form)
		sels := []int{}
		err = json.Unmarshal([]byte(org_sel_form), &sels)
		//fmt.Println(sels)
		if (len(sels) >= 0) && (farm_area > daily_area) {
			for _, r := range sels {
				if r < len(orgdata) {
					orgdata[r].Selected = true
				} else {
					fmt.Println("Invalid Organism index")
				}
				//fmt.Println(i, r)
			}
			//fmt.Println("orgdata =", orgdata)
			var ndias, numbios, diasprod, primdia int
			ndias, numbios, diasprod, primdia, lastsched = min_bio_sim(farm_area, daily_area, orgdata)
			type Result struct {
				Totaldays        int
				Totalbioreactors int
				Totalprod        int
				Firstday         int
				Scheduler        []Prodlist
			}
			var ret = Result{ndias, numbios, diasprod, primdia, lastsched}
			jsonStr, err := json.Marshal(ret)
			checkErr(err)
			w.Write([]byte(jsonStr))
		}

	default:
		fmt.Fprintf(w, "Sorry, only GET and POST methods are supported.")
	}
	fmt.Println()
	fmt.Println()
}

func withdraw_panel(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// fmt.Println("bio", bio)
	switch r.Method {
	case "GET":
		// var jsonStr []byte
		// jsonStr, err := json.Marshal(orgdata)
		// checkErr(err)
		//os.Stdout.Write(jsonStr)
		fmt.Println("Metodo GET para WDPANEL nao suportado")
		w.Write([]byte(scp_err))
		return
	case "PUT":
		err := r.ParseForm()
		if err != nil {
			fmt.Println("ParseForm() err: ", err)
			w.Write([]byte(scp_err))
			return
		}
		id_str := r.FormValue("Id")
		value_str := r.FormValue("Value")
		vol_inc := r.FormValue("VolInc")
		vol_dec := r.FormValue("VolDec")
		start := r.FormValue("Start")
		stop := r.FormValue("Stop")

		if len(id_str) == 0 {
			w.Write([]byte(scp_err))
			return
		}

		id_int, err := strconv.Atoi(id_str)
		if err != nil {
			checkErr(err)
			w.Write([]byte(scp_err))
			return
		}

		if len(value_str) > 0 {
			value_int, err := strconv.Atoi(value_str)
			if err != nil {
				checkErr(err)
				w.Write([]byte(scp_err))
				return
			}
			if value_int == 1 {
				s_str := fmt.Sprintf("SELECT,IBC%02d/END", id_int+1)
				cmd := scp_wdpanel + "/" + s_str
				fmt.Println("DEBUG WDPANEL: to master", cmd)
				jsonStr := []byte(scp_sendmsg_master(cmd))
				os.Stdout.Write(jsonStr)
				w.Write([]byte(jsonStr))
			}
			return
		}

		if len(vol_inc) > 0 {
			vol_inc_int, err := strconv.Atoi(vol_inc)
			if err != nil {
				checkErr(err)
				w.Write([]byte(scp_err))
				return
			}
			if vol_inc_int == 1 {
				s_str := fmt.Sprintf("INC,IBC%02d/END", id_int+1)
				cmd := scp_wdpanel + "/" + s_str
				fmt.Println("DEBUG WDPANEL: to master", cmd)
				jsonStr := []byte(scp_sendmsg_master(cmd))
				os.Stdout.Write(jsonStr)
				w.Write([]byte(jsonStr))
			}
		}

		if len(vol_dec) > 0 {
			vol_dec_int, err := strconv.Atoi(vol_dec)
			if err != nil {
				checkErr(err)
				w.Write([]byte(scp_err))
			}
			if vol_dec_int == 1 {
				s_str := fmt.Sprintf("DEC,IBC%02d/END", id_int+1)
				cmd := scp_wdpanel + "/" + s_str
				fmt.Println("DEBUG WDPANEL: to master", cmd)
				jsonStr := []byte(scp_sendmsg_master(cmd))
				os.Stdout.Write(jsonStr)
				w.Write([]byte(jsonStr))
			}
		}

		if len(start) > 0 {
			start_int, err := strconv.Atoi(start)
			if err != nil {
				checkErr(err)
				w.Write([]byte(scp_err))
				return
			}
			if start_int == 1 {
				s_str := fmt.Sprintf("START,IBC%02d/END", id_int+1)
				cmd := scp_wdpanel + "/" + s_str
				fmt.Println("DEBUG WDPANEL: to master", cmd)
				jsonStr := []byte(scp_sendmsg_master(cmd))
				os.Stdout.Write(jsonStr)
				w.Write([]byte(jsonStr))
			}
		}

		if len(stop) > 0 {
			w.Write([]byte(scp_ack)) ////// tecla STOP desligada ate arrumar botao
			return

			stop_int, err := strconv.Atoi(stop)
			if err != nil {
				checkErr(err)
				w.Write([]byte(scp_err))
				return
			}
			if stop_int == 1 {
				s_str := fmt.Sprintf("STOP,IBC%02d/END", id_int+1)
				cmd := scp_wdpanel + "/" + s_str
				fmt.Println("DEBUG WDPANEL: to master", cmd)
				jsonStr := []byte(scp_sendmsg_master(cmd))
				os.Stdout.Write(jsonStr)
				w.Write([]byte(jsonStr))
			}
		}

	default:
		fmt.Fprintf(w, "Sorry, only GET and POST methods are supported.")
	}
	fmt.Println()
	fmt.Println()
	return
}

func main() {

	devmode = test_file("/etc/scpd/scp_devmode.flag")
	if devmode {
		fmt.Println("WARN:  EXECUTANDO EM DEVMODE\n\n\n")
		execpath = "/home/paulo/work/iot/scp-project/"
	} else {
		execpath = "/home/scpadm/scp-project/"
	}
	//scp_bio_init()
	if load_organisms(execpath+"organismos_conf.csv") < 0 {
		fmt.Println("Não foi possivel ler o arquivo de organismos")
		return
	}
	//fmt.Println(orgs)
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

	lastsched = make([]Prodlist, 0)

	mux.HandleFunc("/bioreactor_view", bioreactor_view)

	mux.HandleFunc("/ibc_view", ibc_view)

	mux.HandleFunc("/totem_view", totem_view)

	mux.HandleFunc("/biofabrica_view", biofabrica_view)

	mux.HandleFunc("/simulator", biofactory_sim)

	mux.HandleFunc("/config", set_config)

	mux.HandleFunc("/wdpanel", withdraw_panel)

	handler := cors.Handler(mux)

	http.ListenAndServe(":5000", handler)
}
