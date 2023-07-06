package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

const demo = true
const testmode = true

const scp_on = 1
const scp_off = 0

const scp_ack = "ACK"
const scp_err = "ERR"
const scp_get = "GET"
const scp_put = "PUT"
const scp_run = "RUN"
const scp_die = "DIE"
const scp_sched = "SCHED"
const scp_fail = "FAIL"

const scp_dev_pump = "PUMP"
const scp_dev_aero = "AERO"
const scp_dev_valve = "VALVE"
const scp_dev_water = "WATER"
const scp_dev_sprayball = "SPRAYBALL"

const scp_par_withdraw = "WITHDRAW"
const scp_par_out = "OUT"
const scp_par_time = "TIME"
const scp_par_volume = "VOLUME"
const scp_par_grow = "GROW"
const scp_par_cip = "CIP"
const scp_par_status = "STATUS"
const scp_par_step = "STEP"

const scp_job_org = "ORG"
const scp_job_on = "ON"
const scp_job_set = "SET"
const scp_job_wait = "WAIT"
const scp_job_ask = "ASK"
const scp_job_off = "OFF"
const scp_job_run = "RUN"
const scp_job_stop = "STOP"
const scp_job_done = "DONE"

const scp_msg_cloro = "CLORO"
const scp_msg_meio = "MEIO"
const scp_msg_inoculo = "INOCULO"
const scp_msg_meio_inoculo = "MEIO-INOCULO"

const scp_bioreactor = "BIOREACTOR"
const scp_biofabrica = "BIOFABRICA"
const scp_totem = "TOTEM"
const scp_ibc = "IBC"
const scp_out = "OUT"
const scp_drop = "DROP"
const scp_clean = "CLEAN"
const scp_donothing = "NOTHING"
const scp_orch_addr = ":7007"
const scp_ipc_name = "/tmp/scp_master.sock"
const scp_refreshwait = 100
const scp_refreshsleep = 1000
const scp_timeout_ms = 5500
const scp_schedwait = 500

const scp_timewaitvalvs = 15000
const scp_maxtimewithdraw = 30

const bio_diametro = 1430  // em mm
const bio_v1_zero = 1483.0 // em mm
const bio_v2_zero = 1502.0 // em mm
const ibc_v1_zero = 2652.0 // em mm   2647

// const scp_join = "JOIN"
const bio_data_filename = "dumpdata"

const bio_nonexist = "NULL"
const bio_die = "DIE"
const bio_cip = "CIP"
const bio_pause = "PAUSADO"
const bio_wait = "AGUARDANDO"
const bio_starting = "INICIANDO"
const bio_loading = "CARREGANDO"
const bio_unloading = "DESENVASE"
const bio_producting = "PRODUZINDO"
const bio_empty = "VAZIO"
const bio_done = "CONCLUIDO"
const bio_storing = "ARMAZENANDO"
const bio_error = "ERRO"
const bio_ready = "PRONTO"
const bio_water = "AGUA"
const bio_max_valves = 8
const bio_max_msg = 50

const TEMPMAX = 120

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

type Bioreact struct {
	BioreactorID string
	Status       string
	OrgCode      string
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
	Withdraw     uint32
	OutID        string
	Queue        []string
}

type IBC struct {
	IBCID      string
	Status     string
	Organism   string
	Volume     uint32
	Level      uint8
	Pumpstatus bool
	Valvs      [4]int
	Timetotal  [2]int
	Withdraw   uint32
	OutID      string
}

type Totem struct {
	TotemID    string
	Status     string
	Pumpstatus bool
	Valvs      [2]int
	Perist     [4]int
}

type Biofabrica struct {
	BiofabricaID string
	Valvs        [9]int
	Pumpwithdraw bool
	Messages     []string
}

type Path struct {
	FromID    string
	ToID      string
	Cleantime int
	Path      string
}

type Bioreact_cfg struct {
	BioreactorID string
	Deviceaddr   string
	Screenaddr   string
	Maxvolume    uint32
	Pump_dev     string
	Aero_dev     string
	Aero_rele    string
	Peris_dev    [5]string
	Valv_devs    [8]string
	Vol_devs     [2]string
	PH_dev       string
	Temp_dev     string
	Levelhigh    string
	Levellow     string
	Emergency    string
	Heater       string
}
type IBC_cfg struct {
	IBCID      string
	Deviceaddr string
	Screenaddr string
	Maxvolume  uint32
	Pump_dev   string
	Valv_devs  [4]string
	Vol_devs   [2]string
	Levellow   string
}

type Totem_cfg struct {
	TotemID    string
	Deviceaddr string
	Pumpdev    string
	Peris_dev  [4]string
	Valv_devs  [2]string
}

type Biofabrica_cfg struct {
	DeviceID   string
	Deviceaddr string
	Deviceport string
}

type Scheditem struct {
	Bioid   string
	Seq     int
	OrgCode string
}

var finishedsetup = false
var schedrunning = false
var devsrunning = false
var autowithdraw = true

var ibc_cfg map[string]IBC_cfg
var bio_cfg map[string]Bioreact_cfg
var totem_cfg map[string]Totem_cfg
var biofabrica_cfg map[string]Biofabrica_cfg
var paths map[string]Path
var valvs map[string]int
var organs map[string]Organism
var schedule []Scheditem
var recipe []string
var cipbio []string
var cipibc []string

var bio = []Bioreact{
	{"BIOR01", bio_empty, "", "", 0, 0, false, false, [8]int{0, 0, 0, 0, 0, 0, 0, 0}, 0, 0, [2]int{2, 5}, [2]int{25, 17}, [2]int{48, 0}, 0, "OUT", []string{}},
	{"BIOR02", bio_empty, "", "", 0, 0, false, false, [8]int{0, 0, 0, 0, 0, 0, 0, 0}, 0, 0, [2]int{1, 1}, [2]int{0, 5}, [2]int{0, 30}, 0, "OUT", []string{}},
	{"BIOR03", bio_empty, "", "", 0, 0, false, false, [8]int{0, 0, 0, 0, 0, 0, 0, 0}, 0, 0, [2]int{1, 1}, [2]int{0, 10}, [2]int{0, 30}, 0, "OUT", []string{}},
	{"BIOR04", bio_empty, "", "", 0, 0, false, false, [8]int{0, 0, 0, 0, 0, 0, 0, 0}, 0, 0, [2]int{1, 1}, [2]int{0, 5}, [2]int{0, 15}, 0, "OUT", []string{}},
	{"BIOR05", bio_empty, "", "", 0, 0, false, false, [8]int{0, 0, 0, 0, 0, 0, 0, 0}, 0, 0, [2]int{5, 5}, [2]int{0, 0}, [2]int{72, 0}, 0, "OUT", []string{}},
	{"BIOR06", bio_ready, "", "", 1000, 5, false, false, [8]int{0, 0, 0, 0, 0, 0, 0, 0}, 0, 0, [2]int{0, 0}, [2]int{0, 0}, [2]int{0, 0}, 0, "OUT", []string{}},
}

var ibc = []IBC{
	{"IBC01", bio_ready, "", 0, 0, false, [4]int{0, 0, 0, 0}, [2]int{0, 0}, 0, "OUT"},
	{"IBC02", bio_ready, "", 0, 0, false, [4]int{0, 0, 0, 0}, [2]int{0, 0}, 0, "OUT"},
	{"IBC03", bio_ready, "Bacillus Amyloliquefaciens", 1000, 2, false, [4]int{0, 0, 0, 0}, [2]int{0, 30}, 0, "OUT"},
	{"IBC04", bio_ready, "Azospirilum brasiliense", 100, 1, false, [4]int{0, 0, 0, 0}, [2]int{4, 50}, 0, "OUT"},
	{"IBC05", bio_ready, "Tricoderma harzianum", 100, 1, false, [4]int{0, 0, 0, 0}, [2]int{13, 17}, 0, "OUT"},
	{"IBC06", bio_ready, "Tricoderma harzianum", 100, 1, false, [4]int{0, 0, 0, 0}, [2]int{0, 5}, 0, "OUT"},
	{"IBC07", bio_ready, "", 100, 1, false, [4]int{0, 0, 0, 0}, [2]int{0, 0}, 0, "OUT"},
}

var totem = []Totem{
	{"TOTEM01", bio_ready, false, [2]int{0, 0}, [4]int{0, 0, 0, 0}},
	{"TOTEM02", bio_ready, false, [2]int{0, 0}, [4]int{0, 0, 0, 0}},
}

var biofabrica = Biofabrica{
	"BIOFABRICA001", [9]int{0, 0, 0, 0, 0, 0, 0, 0, 0}, false, []string{},
}

func checkErr(err error) {
	if err != nil {
		log.Println("[SCP ERROR]", err)
	}
}

func load_tasks_conf(filename string) []string {
	tasks := []string{}
	file, err := os.Open(filename)
	if err != nil {
		checkErr(err)
		return nil
	}
	defer file.Close()
	csvr := csv.NewReader(file)
	paths = make(map[string]Path, 0)
	for {
		r, err := csvr.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			if perr, ok := err.(*csv.ParseError); ok && perr.Err != csv.ErrFieldCount {
				checkErr(err)
				return nil
			}
		}
		// fmt.Println(r)
		if r[0][0] != '#' {
			str := ""
			for _, s := range r {
				str += s + ","
			}
			str += "END"
			tasks = append(tasks, str)
		}
	}
	return tasks
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
	organs = make(map[string]Organism, len(records))
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
		organs[code] = org
		totalrecords = k
	}
	return totalrecords
}

func load_ibcs_conf(filename string) int {
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
	ibc_cfg = make(map[string]IBC_cfg, len(records))
	for k, r := range records {
		id := r[0]
		dev_addr := r[1]
		screen_addr := r[2]
		voltot, _ := strconv.Atoi(strings.Replace(r[3], " ", "", -1))
		pumpdev := r[4]
		vdev1 := r[5]
		vdev2 := r[6]
		vdev3 := r[7]
		vdev4 := r[8]
		voldev1 := r[9]
		voldev2 := r[10]
		llow := r[11]
		ibc_cfg[id] = IBC_cfg{id, dev_addr, screen_addr, uint32(voltot), pumpdev,
			[4]string{vdev1, vdev2, vdev3, vdev4}, [2]string{voldev1, voldev2}, llow}
		totalrecords = k
	}
	return totalrecords
}

func load_bios_conf(filename string) int {
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
	bio_cfg = make(map[string]Bioreact_cfg, len(records))
	totalrecords = 0
	for _, r := range records {
		if !strings.Contains(r[0], "#") && len(r) == 28 {
			id := r[0]
			dev_addr := r[1]
			screen_addr := r[2]
			voltot, _ := strconv.Atoi(strings.Replace(r[3], " ", "", -1))
			pumpdev := r[4]
			aerodev := r[5]
			aerorele := r[6]
			perdev1 := r[7]
			perdev2 := r[8]
			perdev3 := r[9]
			perdev4 := r[10]
			perdev5 := r[11]
			vdev1 := r[12]
			vdev2 := r[13]
			vdev3 := r[14]
			vdev4 := r[15]
			vdev5 := r[16]
			vdev6 := r[17]
			vdev7 := r[18]
			vdev8 := r[19]
			voldev1 := r[20]
			voldev2 := r[21]
			phdev := r[22]
			tempdev := r[23]
			lhigh := r[24]
			llow := r[25]
			emerg := r[26]
			heater := r[27]

			bio_cfg[id] = Bioreact_cfg{id, dev_addr, screen_addr, uint32(voltot), pumpdev, aerodev, aerorele,
				[5]string{perdev1, perdev2, perdev3, perdev4, perdev5},
				[8]string{vdev1, vdev2, vdev3, vdev4, vdev5, vdev6, vdev7, vdev8},
				[2]string{voldev1, voldev2}, phdev, tempdev, lhigh, llow, emerg, heater}
			totalrecords += 1
		} else if len(r) != 26 {
			fmt.Println("ERROR BIO CFG: numero de parametros invalido", r)
		}
	}
	return totalrecords
}

func load_totems_conf(filename string) int {
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
	totem_cfg = make(map[string]Totem_cfg, len(records))
	for k, r := range records {
		id := r[0]
		dev_addr := r[1]
		pumpdev := r[2]
		perdev1 := r[3]
		perdev2 := r[4]
		perdev3 := r[5]
		perdev4 := r[6]
		vdev1 := r[7]
		vdev2 := r[8]
		totem_cfg[id] = Totem_cfg{id, dev_addr, pumpdev,
			[4]string{perdev1, perdev2, perdev3, perdev4},
			[2]string{vdev1, vdev2}}
		totalrecords = k
	}
	return totalrecords
}

func load_biofabrica_conf(filename string) int {
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
	biofabrica_cfg = make(map[string]Biofabrica_cfg, len(records))
	for k, r := range records {
		dev_id := r[0]
		dev_addr := r[1]
		dev_port := r[2]
		biofabrica_cfg[dev_id] = Biofabrica_cfg{dev_id, dev_addr, dev_port}
		totalrecords = k
	}
	return totalrecords
}

func load_paths_conf(filename string) int {
	var totalrecords int
	file, err := os.Open(filename)
	if err != nil {
		checkErr(err)
		return -1
	}
	defer file.Close()
	csvr := csv.NewReader(file)
	paths = make(map[string]Path, 0)
	totalrecords = 0
	for {
		r, err := csvr.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			if perr, ok := err.(*csv.ParseError); ok && perr.Err != csv.ErrFieldCount {
				checkErr(err)
				return -1
			}
		}
		// fmt.Println(r)
		if r[0][0] != '#' {
			from_id := r[0]
			to_id := r[1]
			clean_time, _ := strconv.Atoi(r[2])
			path_id := from_id + "-" + to_id
			pathstr := ""
			for i := 3; i < len(r); i++ {
				pathstr += r[i] + ","
			}
			pathstr += "END"
			paths[path_id] = Path{from_id, to_id, clean_time, pathstr}
			totalrecords++
		}
	}
	return totalrecords
}

func set_valv_status(devtype string, devid string, valvid string, value int) bool {
	var devaddr, scraddr, valvaddr, valve_scrstr string
	id := devid + "/" + valvid
	valvs[id] = value
	switch devtype {
	case scp_donothing:
		return true
	case scp_bioreactor:
		ind := get_bio_index(devid)
		devaddr = bio_cfg[devid].Deviceaddr
		scraddr = bio_cfg[devid].Screenaddr
		if ind >= 0 {
			v, err := strconv.Atoi(valvid[1:])
			if err == nil {
				bio[ind].Valvs[v-1] = value
				valvaddr = bio_cfg[devid].Valv_devs[v-1]
				valve_scrstr = fmt.Sprintf("S%d", v+200)
			} else {
				fmt.Println("ERRO SET VAL: id da valvula nao inteiro", valvid)
				return false
			}
		} else {
			fmt.Println("ERRO SET VAL: BIORREATOR nao encontrado", devid)
			return false
		}
	case scp_ibc:
		ind := get_ibc_index(devid)
		devaddr = ibc_cfg[devid].Deviceaddr
		scraddr = ""
		if ind >= 0 {
			v, err := strconv.Atoi(valvid[1:])
			if err == nil {
				ibc[ind].Valvs[v-1] = value
				valvaddr = ibc_cfg[devid].Valv_devs[v-1]
			} else {
				fmt.Println("ERRO SET VAL: id da valvula nao inteiro", valvid)
				return false
			}
		} else {
			fmt.Println("ERRO SET VAL: IBC nao encontrado", devid)
			return false
		}
	case scp_totem:
		ind := get_totem_index(devid)
		devaddr = totem_cfg[devid].Deviceaddr
		scraddr = ""
		if ind >= 0 {
			v, err := strconv.Atoi(valvid[1:])
			if err == nil {
				totem[ind].Valvs[v-1] = value
				valvaddr = totem_cfg[devid].Valv_devs[v-1]
			} else {
				fmt.Println("ERRO SET VAL: id da valvula nao inteiro", valvid)
				return false
			}
		} else {
			fmt.Println("ERRO SET VAL: TOTEM nao encontrado", devid)
			return false
		}
	case scp_biofabrica:
		devaddr = biofabrica_cfg[valvid].Deviceaddr
		scraddr = ""
		valvaddr = biofabrica_cfg[valvid].Deviceport
		v, err := strconv.Atoi(valvid[3:])
		if err == nil {
			biofabrica.Valvs[v-1] = value
		} else {
			fmt.Println("ERRO SET VAL: BIOFABRICA - id da valvula nao inteiro", valvid)
			return false
		}
	}
	cmd1 := fmt.Sprintf("CMD/%s/PUT/%s,%d/END", devaddr, valvaddr, value)
	fmt.Println(cmd1)
	ret1 := scp_sendmsg_orch(cmd1)
	fmt.Println("RET CMD1 =", ret1)
	if !strings.Contains(ret1, scp_ack) && !testmode {
		fmt.Println("ERRO SET VAL: SEND MSG ORCH falhou", ret1)
		return false
	}
	if len(scraddr) > 0 {
		cmd2 := fmt.Sprintf("CMD/%s/PUT/%s,%d/END", scraddr, valve_scrstr, value)
		fmt.Println(cmd2)
		ret2 := scp_sendmsg_orch(cmd2)
		fmt.Println("RET CMD2 =", ret2)
	}
	return true
}

func get_valv_status(devid string, valvid string) int {
	id := devid + "/" + valvid
	value, ok := valvs[id]
	if ok {
		return value
	}
	return -1
}

func set_allvalvs_status() {
	for _, b := range bio {
		for n, v := range b.Valvs {
			valvid := fmt.Sprintf("V%d", n+1)
			set_valv_status(scp_donothing, b.BioreactorID, valvid, v)
		}
	}
	for _, i := range ibc {
		for n, v := range i.Valvs {
			valvid := fmt.Sprintf("V%d", n+1)
			set_valv_status(scp_donothing, i.IBCID, valvid, v)
		}
	}
	for _, t := range totem {
		for n, v := range t.Valvs {
			valvid := fmt.Sprintf("V%d", n+1)
			set_valv_status(scp_donothing, t.TotemID, valvid, v)
		}
	}
	// Biofrabrica
	for n, v := range biofabrica.Valvs {
		valvid := fmt.Sprintf("VBF%02d", n+1)
		set_valv_status(scp_donothing, "BIOFABRICA", valvid, v)
	}
	fmt.Println(valvs)
}

func save_all_data(filename string) int {
	buf1, _ := json.Marshal(bio)
	err1 := os.WriteFile(filename+"_bio.json", []byte(buf1), 0644)
	checkErr(err1)
	buf2, _ := json.Marshal(ibc)
	err2 := os.WriteFile(filename+"_ibc.json", []byte(buf2), 0644)
	checkErr(err2)
	buf3, _ := json.Marshal(totem)
	err3 := os.WriteFile(filename+"_totem.json", []byte(buf3), 0644)
	checkErr(err3)
	buf4, _ := json.Marshal(biofabrica)
	err4 := os.WriteFile(filename+"_biofabrica.json", []byte(buf4), 0644)
	checkErr(err4)
	return 0
}

func load_all_data(filename string) int {
	dat1, err1 := os.ReadFile(filename + "_bio.json")
	checkErr(err1)
	if err1 == nil {
		json.Unmarshal([]byte(dat1), &bio)
		fmt.Println("-- bio data = ", bio)
	}

	dat2, err2 := os.ReadFile(filename + "_ibc.json")
	checkErr(err2)
	if err2 == nil {
		json.Unmarshal([]byte(dat2), &ibc)
		fmt.Println("-- ibc data = ", ibc)
	}

	dat3, err3 := os.ReadFile(filename + "_totem.json")
	checkErr(err3)
	if err3 == nil {
		json.Unmarshal([]byte(dat3), &totem)
		fmt.Println("-- totem data = ", totem)
	}

	dat4, err4 := os.ReadFile(filename + "_biofabrica.json")
	checkErr(err4)
	if err4 == nil {
		json.Unmarshal([]byte(dat4), &biofabrica)
		fmt.Println("-- biofabrica data = ", biofabrica)
	}
	set_allvalvs_status()
	return 0
}

func scp_splitparam(param string, separator string) []string {
	scp_data := strings.Split(param, separator)
	if len(scp_data) < 1 {
		return nil
	}
	return scp_data
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

func get_ibc_index(ibc_id string) int {
	if len(ibc_id) > 0 {
		for i, v := range ibc {
			if v.IBCID == ibc_id {
				return i
			}
		}
	}
	return -1
}

func get_totem_index(totem_id string) int {
	if len(totem_id) > 0 {
		for i, v := range totem {
			if v.TotemID == totem_id {
				return i
			}
		}
	}
	return -1
}

func get_scp_type(dev_id string) string {
	if strings.Contains(dev_id, "BIOFABRICA") {
		return scp_biofabrica
	} else if strings.Contains(dev_id, "BIOR") {
		return scp_bioreactor
	} else if strings.Contains(dev_id, "IBC") {
		return scp_ibc
	} else if strings.Contains(dev_id, "TOTEM") {
		return scp_totem
	} else if strings.Contains(dev_id, "OUT") {
		return scp_out
	} else if strings.Contains(dev_id, "DROP") {
		return scp_drop
	} else if strings.Contains(dev_id, "CLEAN") {
		return scp_clean
	}
	return scp_donothing
}

func scp_sendmsg_orch(cmd string) string {

	if demo {
		return scp_ack
	}
	//fmt.Println("TO ORCH:", cmd)
	con, err := net.Dial("udp", scp_orch_addr)
	if err != nil {
		checkErr(err)
		return scp_err
	}
	defer con.Close()

	_, err = con.Write([]byte(cmd))
	if err != nil {
		checkErr(err)
		return scp_err
	}
	//fmt.Println("Enviado:", cmd, len(cmd))

	err = con.SetReadDeadline(time.Now().Add(time.Duration(scp_timeout_ms) * time.Millisecond))
	checkErr(err)
	ret := make([]byte, 1024)
	_, err = con.Read(ret)
	if err != nil {
		checkErr(err)
		return scp_err
	}
	//fmt.Println("Recebido:", string(ret))
	return string(ret)
}

func board_add_message(m string) {
	n := len(biofabrica.Messages)
	stime := time.Now().Format("15:04")
	msg := fmt.Sprintf("%c[%s] %s", m[0], stime, m[1:])
	if n < bio_max_msg {
		biofabrica.Messages = append(biofabrica.Messages, msg)
	} else {
		biofabrica.Messages = append(biofabrica.Messages[2:], msg)
	}
}

func scp_setup_devices() {
	if demo {
		return
	}
	fmt.Println("CONFIGURANDO DISPOSITIVOS")

	fmt.Println("\n\nCONFIGURANDO BIORREATORES")
	for _, b := range bio_cfg {
		if len(b.Deviceaddr) > 0 {
			fmt.Println("device:", b.BioreactorID, "-", b.Deviceaddr)
			var cmd []string
			bioaddr := b.Deviceaddr
			cmd = make([]string, 0)
			cmd = append(cmd, "CMD/"+bioaddr+"/MOD/"+b.Pump_dev[1:]+",3/END")
			cmd = append(cmd, "CMD/"+bioaddr+"/MOD/"+b.Aero_dev[1:]+",3/END")
			cmd = append(cmd, "CMD/"+bioaddr+"/MOD/"+b.Aero_rele[1:]+",3/END")
			for i := 0; i < len(b.Peris_dev); i++ {
				cmd = append(cmd, "CMD/"+bioaddr+"/MOD/"+b.Peris_dev[i][1:]+",3/END")
			}
			for i := 0; i < len(b.Valv_devs); i++ {
				cmd = append(cmd, "CMD/"+bioaddr+"/MOD/"+b.Valv_devs[i][1:]+",3/END")
			}
			cmd = append(cmd, "CMD/"+bioaddr+"/MOD/"+b.Levelhigh[1:]+",1/END")
			cmd = append(cmd, "CMD/"+bioaddr+"/MOD/"+b.Levellow[1:]+",1/END")
			cmd = append(cmd, "CMD/"+bioaddr+"/MOD/"+b.Emergency[1:]+",1/END")
			cmd = append(cmd, "CMD/"+b.Screenaddr+"/PUT/S200,1/END")
			nerr := 0
			for k, c := range cmd {
				fmt.Print(k, "  ", c, " ")
				ret := scp_sendmsg_orch(c)
				if !strings.Contains(ret, scp_ack) {
					nerr++
				}
				fmt.Println(ret, nerr)
				if strings.Contains(ret, scp_die) {
					fmt.Println("ERROR SETUP DEVICES: BIORREATOR DIE", b.BioreactorID)
					break
				}
				time.Sleep(scp_refreshwait / 2 * time.Millisecond)
			}
			i := get_bio_index(b.BioreactorID)
			if i >= 0 {
				if nerr > 1 && !testmode {
					bio[i].Status = bio_nonexist
					fmt.Println("ERROR SETUP DEVICES: BIORREATOR com erros", b.BioreactorID)
				} else if bio[i].Status == bio_nonexist {
					bio[i].Status = bio_ready
				}
			} else {
				fmt.Println("ERROR SETUP DEVICES: Biorreator nao encontrado na tabela", b.BioreactorID)
			}
		}
	}

	fmt.Println("\n\nCONFIGURANDO IBCS")
	for _, ib := range ibc_cfg {
		if len(ib.Deviceaddr) > 0 {
			fmt.Println("device:", ib.IBCID, "-", ib.Deviceaddr)
			var cmd []string
			ibcaddr := ib.Deviceaddr
			cmd = make([]string, 0)
			cmd = append(cmd, "CMD/"+ibcaddr+"/MOD/"+ib.Pump_dev[1:]+",3/END")
			for i := 0; i < len(ib.Valv_devs); i++ {
				cmd = append(cmd, "CMD/"+ibcaddr+"/MOD/"+ib.Valv_devs[i][1:]+",3/END")
			}
			// cmd = append(cmd, "CMD/"+ibcaddr+"/MOD/"+b.Levelhigh[1:]+",1/END")
			// cmd = append(cmd, "CMD/"+ibcaddr+"/MOD/"+b.Levellow[1:]+",1/END")
			// cmd = append(cmd, "CMD/"+ibcaddr+"/MOD/"+b.Emergency[1:]+",1/END")

			nerr := 0
			for k, c := range cmd {
				fmt.Print(k, "  ", c, " ")
				ret := scp_sendmsg_orch(c)
				if !strings.Contains(ret, scp_ack) {
					nerr++
				}
				fmt.Println(ret, nerr)
				if strings.Contains(ret, scp_die) {
					fmt.Println("ERROR SETUP DEVICES: IBC DIE", ib.IBCID)
					break
				}
				time.Sleep(scp_refreshwait / 2 * time.Millisecond)
			}
			i := get_ibc_index(ib.IBCID)
			if i >= 0 {
				if nerr > 0 && !testmode {
					ibc[i].Status = bio_nonexist
					fmt.Println("ERROR SETUP DEVICES: IBC com erros", ib.IBCID)
				} else if ibc[i].Status == bio_nonexist {
					ibc[i].Status = bio_ready
				}
			} else {
				fmt.Println("ERROR SETUP DEVICES: IBC nao encontrado na tabela", ib.IBCID)
			}
		}
	}

	fmt.Println("\n\nCONFIGURANDO TOTEMS")
	for _, tot := range totem_cfg {
		if len(tot.Deviceaddr) > 0 {
			fmt.Println("device:", tot.TotemID, "-", tot.Deviceaddr)
			var cmd []string
			totemaddr := tot.Deviceaddr
			cmd = make([]string, 0)
			cmd = append(cmd, "CMD/"+totemaddr+"/MOD/"+tot.Pumpdev[1:]+",3/END")
			for i := 0; i < len(tot.Valv_devs); i++ {
				cmd = append(cmd, "CMD/"+totemaddr+"/MOD/"+tot.Valv_devs[i][1:]+",3/END")
			}
			for i := 0; i < len(tot.Peris_dev); i++ {
				cmd = append(cmd, "CMD/"+totemaddr+"/MOD/"+tot.Peris_dev[i][1:]+",3/END")
			}

			nerr := 0
			for k, c := range cmd {
				fmt.Print(k, "  ", c, " ")
				ret := scp_sendmsg_orch(c)
				if !strings.Contains(ret, scp_ack) {
					nerr++
				}
				fmt.Println(ret, nerr)
				if strings.Contains(ret, scp_die) {
					fmt.Println("ERROR SETUP DEVICES: TOTEM DIE", tot.TotemID)
					break
				}
				time.Sleep(scp_refreshwait / 2 * time.Millisecond)
			}
			i := get_totem_index(tot.TotemID)
			if i >= 0 {
				if nerr > 0 && !testmode {
					totem[i].Status = bio_nonexist
					fmt.Println("ERROR SETUP DEVICES: TOTEM com erros", tot.TotemID)
				}
			} else {
				fmt.Println("ERROR SETUP DEVICES: TOTEM nao encontrado na tabela", tot.TotemID)
			}
		}
	}

	fmt.Println("\n\nCONFIGURANDO BIOFABRICA")
	for _, bf := range biofabrica_cfg {
		if len(bf.Deviceaddr) > 0 {
			fmt.Println("device:", bf.DeviceID, "-", bf.Deviceaddr)
			var cmd []string
			bfaddr := bf.Deviceaddr
			cmd = make([]string, 0)
			cmd = append(cmd, "CMD/"+bfaddr+"/MOD/"+bf.Deviceport[1:]+",3/END")

			for k, c := range cmd {
				fmt.Print(k, "  ", c, " ")
				ret := scp_sendmsg_orch(c)
				fmt.Println(ret)
				if ret[0:2] == "DIE" {
					fmt.Println("SLAVE ERROR - DIE")
					break
				}
				time.Sleep(scp_refreshwait / 2 * time.Millisecond)
			}
		}
	}
	finishedsetup = true
}

func scp_get_alldata() {
	if demo {
		return
	}
	countsave := 0
	for {
		if finishedsetup {
			for k, b := range bio {
				if len(bio_cfg[b.BioreactorID].Deviceaddr) > 0 {
					i := get_bio_index(b.BioreactorID)
					if i >= 0 && (bio[i].Status != bio_nonexist && bio[i].Status != bio_error) {
						bioaddr := bio_cfg[b.BioreactorID].Deviceaddr
						tempdev := bio_cfg[b.BioreactorID].Temp_dev
						phdev := bio_cfg[b.BioreactorID].PH_dev
						v1dev := bio_cfg[b.BioreactorID].Vol_devs[0]
						//v2dev := bio_cfg[b.BioreactorID].Vol_devs[1]

						cmd1 := "CMD/" + bioaddr + "/GET/" + tempdev + "/END"
						ret1 := scp_sendmsg_orch(cmd1)
						params := scp_splitparam(ret1, "/")
						if params[0] == scp_ack {
							tempint, _ := strconv.Atoi(params[1])
							tempfloat := float32(tempint) / 100.0
							if (tempfloat >= 0) && (tempfloat <= TEMPMAX) {
								bio[k].Temperature = tempfloat
							}
						}
						cmd2 := "CMD/" + bioaddr + "/GET/" + phdev + "/END"
						ret2 := scp_sendmsg_orch(cmd2)
						params = scp_splitparam(ret2, "/")
						if params[0] == scp_ack {
							phint, _ := strconv.Atoi(params[1])
							phfloat := float32(phint) / 100.0
							if (phfloat >= 0) && (phfloat <= 14) {
								bio[k].PH = phfloat
							}
						}
						cmd3 := "CMD/" + bioaddr + "/GET/" + v1dev + "/END"
						ret3 := scp_sendmsg_orch(cmd3)
						params = scp_splitparam(ret3, "/")
						if params[0] == scp_ack {
							dint, _ := strconv.Atoi(params[1])
							area := math.Pi * math.Pow(bio_diametro/2000.0, 2)
							dfloat := float64(bio_v1_zero) - float64(dint)
							vol1 := area * dfloat
							fmt.Println("DEBUG Volume ", b.BioreactorID, dint, area, dfloat, vol1)
							if (vol1 >= 0) && (vol1 <= float64(bio_cfg[b.BioreactorID].Maxvolume)*1.2) {
								bio[k].Volume = uint32(vol1)
								level := (vol1 / float64(bio_cfg[b.BioreactorID].Maxvolume)) * 10
								level_int := uint8(level)
								if level_int != bio[k].Level {
									bio[k].Level = level_int
									levels := fmt.Sprintf("%d", level_int)
									cmd := "CMD/" + bio_cfg[b.BioreactorID].Screenaddr + "/PUT/S231," + levels + "/END"
									ret := scp_sendmsg_orch(cmd)
									fmt.Println("SCREEN:", cmd, level, levels, ret)
								}
								if vol1 == 0 {
									bio[k].Status = bio_empty
								}
							}
						}
					}

				}
				time.Sleep(scp_refreshwait * time.Millisecond)
			}
			for k, b := range ibc {
				if len(ibc_cfg[b.IBCID].Deviceaddr) > 0 {
					i := get_ibc_index(b.IBCID)
					if i >= 0 && (ibc[i].Status != bio_nonexist && ibc[i].Status != bio_error) {
						ibcaddr := ibc_cfg[b.IBCID].Deviceaddr
						v1dev := ibc_cfg[b.IBCID].Vol_devs[0]
						//v2dev := bio_cfg[b.BioreactorID].Vol_devs[1]

						cmd1 := "CMD/" + ibcaddr + "/GET/" + v1dev + "/END"
						ret1 := scp_sendmsg_orch(cmd1)
						var vol1 float64
						vol1 = -1
						params := scp_splitparam(ret1, "/")
						if params[0] == scp_ack {
							dint, _ := strconv.Atoi(params[1])
							area := math.Pi * math.Pow(bio_diametro/2000.0, 2)
							dfloat := float64(ibc_v1_zero) - float64(dint)
							vol1 = area * dfloat
							fmt.Println("DEBUG Volume USOM", b.IBCID, ibc_cfg[b.IBCID].Deviceaddr, dint, area, dfloat, vol1, ret1)
						} else {
							fmt.Println("DEBUG ERRO USOM", b.IBCID, ret1, params)
						}

						v2dev := ibc_cfg[b.IBCID].Vol_devs[1]
						cmd2 := "CMD/" + ibcaddr + "/GET/" + v2dev + "/END"
						ret2 := scp_sendmsg_orch(cmd2)
						params = scp_splitparam(ret2, "/")
						var vol2 float64
						vol2 = -1
						if params[0] == scp_ack {
							dint, _ := strconv.Atoi(params[1])
							area := math.Pi * math.Pow(bio_diametro/2000.0, 2)
							dfloat := float64(ibc_v1_zero) - float64(dint)
							vol2 = area * dfloat
							fmt.Println("DEBUG Volume LASER", b.IBCID, ibc_cfg[b.IBCID].Deviceaddr, dint, area, dfloat, vol2, ret2)
						} else {
							fmt.Println("DEBUG ERRO LASER", b.IBCID, ret2, params)
						}
						var volc float64
						if vol1 == -1 && vol2 > 0 {
							volc = vol2
						} else if vol2 == -1 && vol1 > 0 {
							volc = vol1
						} else if vol1 == -1 && vol2 == -1 {
							volc = -1
						} else if vol1 < vol2 {
							volc = vol1
						} else {
							volc = vol2
						}
						if (volc >= 0) && (volc <= float64(ibc_cfg[b.IBCID].Maxvolume)*1.2) {
							ibc[k].Volume = uint32(volc)
							level := (volc / float64(bio_cfg[b.IBCID].Maxvolume)) * 10
							level_int := uint8(level)
							if level_int != ibc[k].Level {
								ibc[k].Level = level_int
								// levels := fmt.Sprintf("%d", level_int)
								// cmd := "CMD/" + ibc_cfg[b.IBCID].Screenaddr + "/PUT/S231," + levels + "/END"
								// ret := scp_sendmsg_orch(cmd)
								// fmt.Println("SCREEN:", cmd, level, levels, ret)
							}
							if volc == 0 {
								ibc[k].Status = bio_empty
							}
						}
					}

				}
				time.Sleep(scp_refreshwait * time.Millisecond)
			}

			countsave++
			if countsave == 5 {
				save_all_data(bio_data_filename)
				countsave = 0
			}

			time.Sleep(scp_refreshsleep * time.Millisecond)

		}
	}
}

func set_valvs_value(vlist []string, value int, abort_on_error bool) int {
	tot := 0
	for _, p := range vlist {
		if p != "END" {
			val, ok := valvs[p]
			if ok {
				sub := scp_splitparam(p, "/")
				dtype := get_scp_type(sub[0])
				if val == (1 - value) {
					if !set_valv_status(dtype, sub[0], sub[1], value) {
						fmt.Println("ERRO SET VALVS VALUE: nao foi possivel setar valvula", p)
						if abort_on_error {
							return -1
						}
					}
					tot++
				} else if val == 1 {
					fmt.Println("ERRO SET VALVS VALUE: nao foi possivel setar valvula", p)
					if abort_on_error {
						return -1
					}
				} else {
					fmt.Println("ERRO SET VALVS VALUE: valvula com erro", p)
					if abort_on_error {
						return -1
					}
				}
			} else {
				fmt.Println("ERRO SET VALVS VALUE: valvula nao existe", p)
				if abort_on_error {
					return -1
				}
			}
		}
	}
	return tot
}

func test_path(vpath []string, value int) bool {
	ret := true
	for _, p := range vpath {
		if p == "END" {
			break
		}
		val, ok := valvs[p]
		// fmt.Println("step", p, "ret=", ret, "val=", val, "ok=", ok)
		ret = ret && (val == value) && ok
		// fmt.Println("ret final=", ret)
	}
	return ret
}

func scp_run_withdraw(devtype string, devid string) int {
	switch devtype {
	case scp_bioreactor:
		ind := get_bio_index(devid)
		pathid := devid + "-" + bio[ind].OutID
		pathstr := paths[pathid].Path
		if len(pathstr) == 0 {
			fmt.Println("ERRO RUN WITHDRAW 01: path nao existe", pathid)
			return -1
		}
		vpath := scp_splitparam(pathstr, ",")
		if !test_path(vpath, 0) {
			fmt.Println("ERRO RUN WITHDRAW 02: falha de valvula no path", pathid)
			return -1
		}
		board_add_message("CDesenvase " + devid + " para " + bio[ind].OutID)
		var pilha []string = make([]string, 0)
		for k, p := range vpath {
			fmt.Println("step", k, p)
			if p == "END" {
				break
			}
			val, ok := valvs[p]
			if ok {
				sub := scp_splitparam(p, "/")
				dtype := get_scp_type(sub[0])
				if val == 0 {
					if !set_valv_status(dtype, sub[0], sub[1], 1) {
						fmt.Println("ERRO RUN WITHDRAW 03: nao foi possivel setar valvula", p)
						set_valvs_value(pilha, 0, false) // undo
						return -1
					}
				} else if val == 1 {
					fmt.Println("ERRO RUN WITHDRAW 04: valvula ja aberta", p)
					set_valvs_value(pilha, 0, false) // undo
					return -1
				} else {
					fmt.Println("ERRO RUN WITHDRAW 05: valvula com erro", p)
					set_valvs_value(pilha, 0, false) // undo
					return -1
				}
			} else {
				fmt.Println("ERRO RUN WITHDRAW 06: valvula nao existe", p)
				set_valvs_value(pilha, 0, false) // undo
				return -1
			}
			pilha = append([]string{p}, pilha...)
		}
		fmt.Println(pilha)
		vol_ini := bio[ind].Volume
		bio[ind].Status = bio_unloading
		time.Sleep(scp_timewaitvalvs * time.Millisecond)
		fmt.Println("WARN RUN WITHDRAW 07: Ligando bomba", devid)
		biodev := bio_cfg[devid].Deviceaddr
		bioscr := bio_cfg[devid].Screenaddr
		pumpdev := bio_cfg[devid].Pump_dev
		bio[ind].Pumpstatus = true
		cmd1 := "CMD/" + biodev + "/PUT/" + pumpdev + ",1/END"
		cmd2 := "CMD/" + bioscr + "/PUT/S270,1/END"
		ret1 := scp_sendmsg_orch(cmd1)
		fmt.Println("DEBUG RUN WITHDRAW 08: CMD1 =", cmd1, " RET=", ret1)
		ret2 := scp_sendmsg_orch(cmd2)
		fmt.Println("DEBUG RUN WITHDRAW 09: CMD2 =", cmd2, " RET=", ret2)
		if !strings.Contains(ret1, scp_ack) && !testmode {
			fmt.Println("ERRO RUN WITHDRAW 10: BIORREATOR falha ao ligar bomba")
			cmd2 := "CMD/" + bioscr + "/PUT/S270,0/END"
			scp_sendmsg_orch(cmd2)
			set_valvs_value(pilha, 0, false)
			return -1
		}
		t_start := time.Now()
		for {
			vol_now := bio[ind].Volume
			// t_now := time.Now()
			t_elapsed := time.Since(t_start).Seconds()
			vol_out := vol_ini - vol_now
			if bio[ind].Withdraw == 0 {
				break
			}
			if vol_now < vol_ini && vol_out >= bio[ind].Withdraw {
				fmt.Println("DEBUG RUN WITHDRAW 11: Volume de desenvase atingido", vol_ini, vol_now, bio[ind].Withdraw)
				break
			}
			if t_elapsed > scp_maxtimewithdraw {
				fmt.Println("DEBUG RUN WITHDRAW 12: Tempo maximo de withdraw esgotado", t_elapsed, scp_maxtimewithdraw)
				break
			}
			time.Sleep(scp_refreshwait * time.Millisecond)
		}
		bio[ind].Withdraw = 0
		board_add_message("IDesenvase concluido")
		fmt.Println("WARN RUN WITHDRAW 13: Desligando bomba", devid)
		bio[ind].Pumpstatus = false
		cmd1 = "CMD/" + biodev + "/PUT/" + pumpdev + ",0/END"
		cmd2 = "CMD/" + bioscr + "/PUT/S270,0/END"
		ret1 = scp_sendmsg_orch(cmd1)
		fmt.Println("DEBUG RUN WITHDRAW 14: CMD1 =", cmd1, " RET=", ret1)
		ret2 = scp_sendmsg_orch(cmd2)
		fmt.Println("DEBUG RUN WITHDRAW 15: CMD2 =", cmd2, " RET=", ret2)
		set_valvs_value(pilha, 0, false)
		bio[ind].Status = bio_ready
		time.Sleep(scp_timewaitvalvs * time.Millisecond)
		var pathclean string = ""
		dest_type := get_scp_type(bio[ind].OutID)
		if dest_type == scp_out || dest_type == scp_drop {
			pathclean = "TOTEM02-CLEAN4"
			board_add_message("IEnxague LINHAS 2/4")
		} else if dest_type == scp_ibc {
			pathclean = "TOTEM02-CLEAN3"
			board_add_message("IEnxague LINHAS 2/3")
		} else {
			fmt.Println("ERRO RUN WITHDRAW 16: destino para clean desconhecido", dest_type)
			return -1
		}
		pathstr = paths[pathclean].Path
		if len(pathstr) == 0 {
			fmt.Println("ERRO RUN WITHDRAW 17: path CLEAN linha nao existe", pathclean)
			return -1
		}
		var time_to_clean int64 = int64(paths[pathclean].Cleantime) * 1000
		vpath = scp_splitparam(pathstr, ",")
		if !test_path(vpath, 0) {
			fmt.Println("ERRO RUN WITHDRAW 18: falha de valvula no path", pathstr)
			return -1
		}
		if set_valvs_value(vpath, 1, true) < 1 {
			fmt.Println("ERROR RUN WITHDRAW 19: Falha ao abrir valvulas CLEAN linha", pathstr)
			set_valvs_value(vpath, 0, false)
		}
		time.Sleep(scp_timewaitvalvs * time.Millisecond)
		fmt.Println("WARN RUN WITHDRAW 20: Ligando bomba TOTEM02", devid)
		tind := get_totem_index("TOTEM02")
		if tind < 0 {
			fmt.Println("WARN RUN WITHDRAW 21: TOTEM02 nao encontrado", totem)
		}
		totemdev := totem_cfg["TOTEM02"].Deviceaddr
		pumpdev = totem_cfg["TOTEM02"].Pumpdev
		totem[tind].Pumpstatus = true
		cmd1 = "CMD/" + totemdev + "/PUT/" + pumpdev + ",1/END"
		ret1 = scp_sendmsg_orch(cmd1)
		fmt.Println("DEBUG RUN WITHDRAW 22: CMD1 =", cmd1, " RET=", ret1)
		if !strings.Contains(ret1, scp_ack) && !testmode {
			fmt.Println("ERRO RUN WITHDRAW 23: BIORREATOR falha ao ligar bomba TOTEM02")
			totem[tind].Pumpstatus = false
			set_valvs_value(vpath, 0, false)
			return -1
		}
		time.Sleep(time.Duration(time_to_clean) * time.Millisecond)
		fmt.Println("WARN RUN WITHDRAW 24: Desligando bomba TOTEM02", devid)
		totem[tind].Pumpstatus = false
		cmd1 = "CMD/" + totemdev + "/PUT/" + pumpdev + ",0/END"
		ret1 = scp_sendmsg_orch(cmd1)
		fmt.Println("DEBUG RUN WITHDRAW 25: CMD1 =", cmd1, " RET=", ret1)
		if !strings.Contains(ret1, scp_ack) && !testmode {
			fmt.Println("ERRO RUN WITHDRAW 26: BIORREATOR falha ao ligar bomba TOTEM02")
			totem[tind].Pumpstatus = false
			set_valvs_value(vpath, 0, false)
			return -1
		}
		set_valvs_value(vpath, 0, false)
		board_add_message("IEnxague concluído")

	case scp_ibc:
		ind := get_ibc_index(devid)
		pathid := devid + "-" + ibc[ind].OutID
		pathstr := paths[pathid].Path
		if len(pathstr) == 0 {
			fmt.Println("ERRO RUN WITHDRAW 27: path nao existe", pathid)
			return -1
		}
		vpath := scp_splitparam(pathstr, ",")
		if !test_path(vpath, 0) {
			fmt.Println("ERRO RUN WITHDRAW 28: falha de valvula no path", pathid)
			return -1
		}
		board_add_message("CDesenvase " + devid + " para " + ibc[ind].OutID)
		var pilha []string = make([]string, 0)
		for k, p := range vpath {
			fmt.Println("step", k, p)
			if p == "END" {
				break
			}
			val, ok := valvs[p]
			if ok {
				sub := scp_splitparam(p, "/")
				dtype := get_scp_type(sub[0])
				if val == 0 {
					if !set_valv_status(dtype, sub[0], sub[1], 1) {
						fmt.Println("ERRO RUN WITHDRAW 29: nao foi possivel setar valvula", p)
						return -1
					}
				} else if val == 1 {
					fmt.Println("ERRO RUN WITHDRAW 30: valvula ja aberta", p)
					return -1
				} else {
					fmt.Println("ERRO RUN WITHDRAW 31: valvula com erro", p)
					return -1
				}
			} else {
				fmt.Println("ERRO RUN WITHDRAW 32: valvula nao existe", p)
				return -1
			}
			pilha = append([]string{p}, pilha...)
		}
		vol_ini := ibc[ind].Volume
		ibc[ind].Status = bio_unloading
		time.Sleep(scp_timewaitvalvs * time.Millisecond)
		fmt.Println("WARN RUN WITHDRAW 33: Ligando bomba", devid)
		pumpdev := biofabrica_cfg["PBF01"].Deviceaddr
		pumpport := biofabrica_cfg["PBF01"].Deviceport
		biofabrica.Pumpwithdraw = true
		cmd1 := "CMD/" + pumpdev + "/PUT/" + pumpport + ",1/END"
		ret1 := scp_sendmsg_orch(cmd1)
		fmt.Println("DEBUG RUN WITHDRAW 34: CMD1 =", cmd1, " RET=", ret1)
		if !strings.Contains(ret1, scp_ack) && !testmode {
			fmt.Println("ERRO RUN WITHDRAW 35: IBC falha ao ligar bomba desenvase")
			return -1
		}
		t_start := time.Now()
		for {
			vol_now := ibc[ind].Volume
			// t_now := time.Now()
			t_elapsed := time.Since(t_start).Seconds()
			vol_out := vol_ini - vol_now
			if ibc[ind].Withdraw == 0 {
				break
			}
			if vol_now < vol_ini && vol_out >= ibc[ind].Withdraw {
				fmt.Println("DEBUG RUN WITHDRAW 36: STOP Volume de desenvase atingido", vol_ini, vol_now, ibc[ind].Withdraw)
				break
			}
			if t_elapsed > scp_maxtimewithdraw {
				fmt.Println("DEBUG RUN WITHDRAW 37: STOP Tempo maximo de withdraw esgotado", t_elapsed, scp_maxtimewithdraw)
				break
			}
			time.Sleep(scp_refreshwait * time.Millisecond)
		}
		ibc[ind].Withdraw = 0
		board_add_message("IDesenvase concluido")

		fmt.Println("WARN RUN WITHDRAW 38: Desligando bomba biofabrica", pumpdev)
		biofabrica.Pumpwithdraw = false
		cmd1 = "CMD/" + pumpdev + "/PUT/" + pumpport + ",0/END"
		ret1 = scp_sendmsg_orch(cmd1)
		fmt.Println("DEBUG RUN WITHDRAW 39: CMD1 =", cmd1, " RET=", ret1)
		set_valvs_value(pilha, 0, false)
		time.Sleep(scp_timewaitvalvs * time.Millisecond)
		ibc[ind].Status = bio_ready
		var pathclean string = ""
		dest_type := get_scp_type(ibc[ind].OutID)
		pathclean = "TOTEM02-CLEAN4"
		pathstr = paths[pathclean].Path
		if len(pathstr) == 0 {
			fmt.Println("ERRO RUN WITHDRAW 40: path CLEAN linha nao existe", pathclean)
			return -1
		}
		var time_to_clean int64 = int64(paths[pathclean].Cleantime) * 1000
		vpath = scp_splitparam(pathstr, ",")
		if !test_path(vpath, 0) {
			fmt.Println("ERRO RUN WITHDRAW 41: falha de valvula no path", pathstr)
			return -1
		}
		board_add_message("ILimpando LINHA 4")
		if set_valvs_value(vpath, 1, true) < 1 {
			fmt.Println("ERROR RUN WITHDRAW 42: Falha ao abrir valvulas CLEAN linha", pathstr)
			set_valvs_value(vpath, 0, false)
		}
		time.Sleep(scp_timewaitvalvs * time.Millisecond)
		fmt.Println("WARN RUN WITHDRAW 43: Ligando bomba TOTEM02", devid)
		tind := get_totem_index("TOTEM02")
		if tind < 0 {
			fmt.Println("WARN RUN WITHDRAW 44: TOTEM02 nao encontrado", totem)
		}
		totemdev := totem_cfg["TOTEM02"].Deviceaddr
		pumpdev = totem_cfg["TOTEM02"].Pumpdev
		totem[tind].Pumpstatus = true
		cmd1 = "CMD/" + totemdev + "/PUT/" + pumpdev + ",1/END"
		ret1 = scp_sendmsg_orch(cmd1)
		fmt.Println("DEBUG RUN WITHDRAW 45: CMD1 =", cmd1, " RET=", ret1)
		if !strings.Contains(ret1, scp_ack) && !testmode {
			fmt.Println("ERRO RUN WITHDRAW 46: BIORREATOR falha ao ligar bomba TOTEM02")
			totem[tind].Pumpstatus = false
			set_valvs_value(vpath, 0, false)
			return -1
		}
		time.Sleep(time.Duration(time_to_clean/2) * time.Millisecond)
		fmt.Println("WARN RUN WITHDRAW 47: Desligando bomba TOTEM02", devid)
		totem[tind].Pumpstatus = false
		cmd1 = "CMD/" + totemdev + "/PUT/" + pumpdev + ",0/END"
		ret1 = scp_sendmsg_orch(cmd1)
		fmt.Println("DEBUG RUN WITHDRAW 48: CMD1 =", cmd1, " RET=", ret1)
		if !strings.Contains(ret1, scp_ack) && !testmode {
			fmt.Println("ERRO RUN WITHDRAW 49: BIORREATOR falha ao ligar bomba TOTEM02")
			set_valvs_value(vpath, 0, false)
			return -1
		}
		set_valvs_value(vpath, 0, false)
		time.Sleep(scp_timewaitvalvs * time.Millisecond)
		board_add_message("IEnxague concluído")
		if dest_type == scp_ibc {
			pathclean = "TOTEM02-CLEAN3"
			pathstr = paths[pathclean].Path
			if len(pathstr) == 0 {
				fmt.Println("ERRO RUN WITHDRAW 50: path CLEAN linha nao existe", pathclean)
				return -1
			}
			var time_to_clean int64 = int64(paths[pathclean].Cleantime) * 1000
			vpath = scp_splitparam(pathstr, ",")
			if !test_path(vpath, 0) {
				fmt.Println("ERRO RUN WITHDRAW 51: falha de valvula no path", pathstr)
				return -1
			}
			board_add_message("ILimpando LINHA 3")
			if set_valvs_value(vpath, 1, true) < 1 {
				fmt.Println("ERROR RUN WITHDRAW 52: Falha ao abrir valvulas CLEAN linha", pathstr)
				set_valvs_value(vpath, 0, false)
			}
			time.Sleep(scp_timewaitvalvs * time.Millisecond)
			fmt.Println("WARN RUN WITHDRAW 53: Ligando bomba TOTEM02", devid)
			tind := get_totem_index("TOTEM02")
			if tind < 0 {
				fmt.Println("WARN RUN WITHDRAW 54: TOTEM02 nao encontrado", totem)
			}
			totemdev := totem_cfg["TOTEM02"].Deviceaddr
			pumpdev = totem_cfg["TOTEM02"].Pumpdev
			totem[tind].Pumpstatus = true
			cmd1 = "CMD/" + totemdev + "/PUT/" + pumpdev + ",1/END"
			ret1 = scp_sendmsg_orch(cmd1)
			fmt.Println("DEBUG RUN WITHDRAW 55: CMD1 =", cmd1, " RET=", ret1)
			if !strings.Contains(ret1, scp_ack) && !testmode {
				fmt.Println("ERRO RUN WITHDRAW 56: BIORREATOR falha ao ligar bomba TOTEM02")
				totem[tind].Pumpstatus = false
				set_valvs_value(vpath, 0, false)
				return -1
			}
			time.Sleep(time.Duration(time_to_clean/2) * time.Millisecond)
			fmt.Println("WARN RUN WITHDRAW 57: Desligando bomba TOTEM02", devid)
			totem[tind].Pumpstatus = false
			cmd1 = "CMD/" + totemdev + "/PUT/" + pumpdev + ",0/END"
			ret1 = scp_sendmsg_orch(cmd1)
			fmt.Println("DEBUG RUN WITHDRAW 58: CMD1 =", cmd1, " RET=", ret1)
			if !strings.Contains(ret1, scp_ack) && !testmode {
				fmt.Println("ERRO RUN WITHDRAW 59: BIORREATOR falha ao ligar bomba TOTEM02")
				set_valvs_value(vpath, 0, false)
				return -1
			}
			set_valvs_value(vpath, 0, false)
			board_add_message("IEnxague concluído")
		}
	}
	return 0
}

func scp_turn_aero(bioid string, changevalvs bool, value int, percent int) bool {
	ind := get_bio_index(bioid)
	if ind < 0 {
		fmt.Println("ERROR SCP TURN: Biorreator nao existe", bioid)
		return false
	}
	devaddr := bio_cfg[bioid].Deviceaddr
	scraddr := bio_cfg[bioid].Screenaddr
	aerorele := bio_cfg[bioid].Aero_rele
	aerodev := bio_cfg[bioid].Aero_dev
	dev_valvs := []string{bioid + "/V1", bioid + "/V2"}

	if value == scp_off {
		cmd0 := fmt.Sprintf("CMD/%s/PUT/%s,%d/END", devaddr, aerorele, value)
		ret0 := scp_sendmsg_orch(cmd0)
		fmt.Println("DEBUG SCP TURN AERO: CMD =", cmd0, "\tRET =", ret0)
		if !strings.Contains(ret0, scp_ack) && !testmode {
			fmt.Println("ERROR SCP TURN AERO:", bioid, " erro ao definir valor[", value, "] rele aerador ", ret0)
			if changevalvs {
				set_valvs_value(dev_valvs, 1-value, false)
			}
			return false
		}
		bio[ind].Aerator = false
		cmds := fmt.Sprintf("CMD/%s/PUT/S271,%d/END", scraddr, value)
		rets := scp_sendmsg_orch(cmds)
		fmt.Println("DEBUG SCP TURN AERO: CMD =", cmds, "\tRET =", rets)
		if !strings.Contains(rets, scp_ack) && !testmode {
			fmt.Println("ERROR SCP TURN AERO:", bioid, " erro ao mudar aerador na screen ", scraddr, rets)
		}

	}

	if changevalvs {
		if test_path(dev_valvs, 1-value) {
			if set_valvs_value(dev_valvs, value, true) < 0 {
				fmt.Println("ERROR SCP TURN AERO: erro ao definir valor [", value, "] das valvulas", dev_valvs)
				return false
			}
		} else {
			fmt.Println("ERROR SCP TURN AERO: erro nas valvulas", dev_valvs)
			return false
		}
	}
	aerovalue := int(255.0 * (float32(percent) / 100.0))
	cmd1 := fmt.Sprintf("CMD/%s/PUT/%s,%d/END", devaddr, aerodev, aerovalue)
	ret1 := scp_sendmsg_orch(cmd1)
	fmt.Println("DEBUG SCP TURN AERO: CMD =", cmd1, "\tRET =", ret1)
	if !strings.Contains(ret1, scp_ack) && !testmode {
		fmt.Println("ERROR SCP TURN AERO:", bioid, " erro ao definir ", percent, "% aerador", ret1)
		if changevalvs {
			set_valvs_value(dev_valvs, 1-value, false)
		}
		return false
	}

	time.Sleep(scp_timewaitvalvs * time.Millisecond)
	if value == scp_on {
		cmd2 := fmt.Sprintf("CMD/%s/PUT/%s,%d/END", devaddr, aerorele, value)
		ret2 := scp_sendmsg_orch(cmd2)
		fmt.Println("DEBUG SCP TURN AERO: CMD =", cmd2, "\tRET =", ret2)
		if !strings.Contains(ret2, scp_ack) && !testmode {
			fmt.Println("ERROR SCP TURN ERO:", bioid, " erro ao definir valor[", value, "] rele aerador ", ret2)
			if changevalvs {
				set_valvs_value(dev_valvs, 1-value, false)
			}
			ret1 = scp_sendmsg_orch(cmd1)
			return false
		}
		bio[ind].Aerator = true
		cmds := fmt.Sprintf("CMD/%s/PUT/S271,%d/END", scraddr, value)
		rets := scp_sendmsg_orch(cmds)
		fmt.Println("DEBUG SCP TURN AERO: CMD =", cmds, "\tRET =", rets)
		if !strings.Contains(rets, scp_ack) && !testmode {
			fmt.Println("ERROR SCP TURN AERO:", bioid, " erro ao mudar aerador na screen ", scraddr, rets)
		}
	}

	return true
}

func scp_turn_pump(devtype string, main_id string, valvs []string, value int) bool {
	var devaddr, pumpdev string
	var ind int
	scraddr := ""
	switch devtype {
	case scp_bioreactor:
		ind = get_bio_index(main_id)
		if ind < 0 {
			fmt.Println("ERROR SCP TURN PUMP: Biorreator nao existe", main_id)
			return false
		}
		devaddr = bio_cfg[main_id].Deviceaddr
		pumpdev = bio_cfg[main_id].Pump_dev
		scraddr = bio_cfg[main_id].Screenaddr

	case scp_ibc:
		ind = get_ibc_index(main_id)
		if ind < 0 {
			fmt.Println("ERROR SCP TURN PUMP: IBC nao existe", main_id)
			return false
		}
		devaddr = ibc_cfg[main_id].Deviceaddr
		pumpdev = ibc_cfg[main_id].Pump_dev

	case scp_totem:
		ind = get_totem_index(main_id)
		if ind < 0 {
			fmt.Println("ERROR SCP TURN PUMP: IBC nao existe", main_id)
			return false
		}
		devaddr = totem_cfg[main_id].Deviceaddr
		pumpdev = totem_cfg[main_id].Pumpdev

	default:
		fmt.Println("ERROR SCP TURN PUMP: Dispositivo nao suportado", devtype, main_id)
	}

	if value == scp_off {
		cmd := fmt.Sprintf("CMD/%s/PUT/%s,%d/END", devaddr, pumpdev, value)
		ret := scp_sendmsg_orch(cmd)
		fmt.Println("DEBUG SCP TURN PUMP: CMD =", cmd, "\tRET =", ret)
		if !strings.Contains(ret, scp_ack) && !testmode {
			fmt.Println("ERROR SCP TURN PUMP:", main_id, " erro ao definir ", value, " bomba", ret)
			if len(valvs) > 0 {
				set_valvs_value(valvs, 1-value, false)
				time.Sleep(scp_timewaitvalvs * time.Millisecond)
			}
			return false
		}
		switch devtype {
		case scp_bioreactor:
			bio[ind].Pumpstatus = false
		case scp_ibc:
			ibc[ind].Pumpstatus = false
		case scp_totem:
			totem[ind].Pumpstatus = false
		}
		if len(scraddr) > 0 {
			cmds := fmt.Sprintf("CMD/%s/PUT/S270,%d/END", scraddr, value)
			rets := scp_sendmsg_orch(cmds)
			fmt.Println("DEBUG SCP TURN AERO: CMD =", cmds, "\tRET =", rets)
			if !strings.Contains(rets, scp_ack) && !testmode {
				fmt.Println("ERROR SCP TURN AERO: erro ao mudar bomba na screen ", scraddr, rets)
			}
		}
	}

	if test_path(valvs, 1-value) {
		if set_valvs_value(valvs, value, true) < 0 {
			fmt.Println("ERROR SCP TURN PUMP:", devtype, " erro ao definir valor [", value, "] das valvulas", valvs)
			return false
		}
	} else {
		fmt.Println("ERROR SCP TURN PUMP:", devtype, " erro nas valvulas", valvs)
		return false
	}
	time.Sleep(scp_timewaitvalvs * time.Millisecond)

	if value == scp_on {
		cmd := fmt.Sprintf("CMD/%s/PUT/%s,%d/END", devaddr, pumpdev, value)
		ret := scp_sendmsg_orch(cmd)
		fmt.Println("DEBUG SCP TURN PUMP: CMD =", cmd, "\tRET =", ret)
		if !strings.Contains(ret, scp_ack) && !testmode {
			fmt.Println("ERROR SCP TURN PUMP:", main_id, " erro ao definir ", value, " bomba", ret)
			if len(valvs) > 0 {
				set_valvs_value(valvs, 1-value, false)
				time.Sleep(scp_timewaitvalvs * time.Millisecond)
			}
			return false
		}
		switch devtype {
		case scp_bioreactor:
			bio[ind].Pumpstatus = true
		case scp_ibc:
			ibc[ind].Pumpstatus = true
		case scp_totem:
			totem[ind].Pumpstatus = true
		}
		if len(scraddr) > 0 {
			cmds := fmt.Sprintf("CMD/%s/PUT/S270,%d/END", scraddr, value)
			rets := scp_sendmsg_orch(cmds)
			fmt.Println("DEBUG SCP TURN AERO: CMD =", cmds, "\tRET =", rets)
			if !strings.Contains(rets, scp_ack) && !testmode {
				fmt.Println("ERROR SCP TURN AERO: erro ao mudar bomba na screen ", scraddr, rets)
			}
		}
	}
	return true
}

func pop_first_sched(bioid string, remove bool) Scheditem {
	var ret Scheditem
	for k, s := range schedule {
		if s.Bioid == bioid {
			ret = Scheditem{s.Bioid, s.Seq, s.OrgCode}
			if remove {
				if k > 0 {
					if k < len(schedule)-1 {
						schedule = append(schedule[:k], schedule[k+1:]...)
					} else {
						schedule = schedule[:k]
					}
				} else {
					schedule = schedule[k+1:]
				}
			}
			return ret
		}
	}
	ret = Scheditem{}
	return ret
}

func pop_first_job(bioid string, remove bool) string {
	ind := get_bio_index(bioid)
	if ind < 0 {
		fmt.Println("ERROR POP FIRST WORK: Biorreator nao existe", bioid)
		return ""
	}
	n := len(bio[ind].Queue)
	ret := ""
	if n > 0 {
		ret = bio[ind].Queue[0]
		if remove {
			bio[ind].Queue = bio[ind].Queue[1:]
		}
	}
	return ret
}

func scp_run_job(bioid string, job string) bool {
	fmt.Println("\n\nSIMULANDO EXECUCAO", bioid, job)
	ind := get_bio_index(bioid)
	if ind < 0 {
		fmt.Println("ERROR SCP RUN JOB: Biorreator nao existe", bioid)
		return false
	}
	params := scp_splitparam(job, "/")
	subpars := []string{}
	if len(params) > 1 {
		subpars = scp_splitparam(params[1], ",")
	}
	switch params[0] {
	case scp_job_org:
		var orgcode string
		if len(subpars) > 0 {
			orgcode = subpars[0]
			if len(organs[orgcode].Orgname) > 0 {
				bio[ind].OrgCode = subpars[0]
				bio[ind].Organism = organs[orgcode].Orgname
				bio[ind].Timetotal = [2]int{organs[orgcode].Timetotal, 0}
				bio[ind].Timeleft = [2]int{organs[orgcode].Timetotal, 0}
			} else {
				fmt.Println("ERROR SCP RUN JOB: Organismo nao existe", params)
				return false
			}
		} else {
			fmt.Println("ERROR SCP RUN JOB: Falta parametros em", scp_job_org, params)
			return false
		}
		board_add_message("IIniciando Cultivo " + organs[orgcode].Orgname + " no " + bioid)

	case scp_job_set:
		if len(subpars) > 1 {
			flag := subpars[0]
			switch flag {
			case scp_par_status:
				biostatus := subpars[1]
				bio[ind].Status = biostatus
			case scp_par_step:
				biostep_str := subpars[1]
				biostep, _ := strconv.Atoi(biostep_str)
				bio[ind].Step[0] = biostep
			default:
				fmt.Println("ERROR SCP RUN JOB: Parametro invalido em", scp_job_set, flag, params)
				return false
			}
		} else {
			fmt.Println("ERROR SCP RUN JOB: Falta parametros em", scp_job_set, params)
			return false
		}

	case scp_job_run:
		if len(subpars) > 0 {
			cmd := subpars[0]
			switch cmd {
			case scp_par_grow:
				fmt.Println("running GROW")

			case scp_par_cip:
				qini := []string{bio[ind].Queue[0]}
				qini = append(qini, cipbio...)
				bio[ind].Queue = append(qini, bio[ind].Queue[1:]...)
				fmt.Println("\n\nTRUQUE CIP:", bio[ind].Queue)
				board_add_message("ICIP Automático no biorreator " + bioid)
				return true

			case scp_par_withdraw:
				bio[ind].Withdraw = bio[ind].Volume
				bio[ind].OutID = strings.Replace(bioid, "BIOR", "IBC", -1)
				board_add_message("IDesenvase Automático do biorreator " + bioid + " para " + bio[ind].OutID)
				if scp_run_withdraw(scp_bioreactor, bioid) < 0 {
					return false
				}

			}
		} else {
			fmt.Println("ERROR SCP RUN JOB: Falta parametros em", scp_job_run, params)
			return false
		}

	case scp_job_ask:
		if len(subpars) > 0 {
			msg := subpars[0]
			scraddr := bio_cfg[bioid].Screenaddr
			var cmd1 string = ""
			var msgask string = ""
			switch msg {
			case scp_msg_cloro:
				cmd1 = fmt.Sprintf("CMD/%s/PUT/S400,2/END", scraddr)
				msgask = "CLORO"
			case scp_msg_meio_inoculo:
				cmd1 = fmt.Sprintf("CMD/%s/PUT/S400,1/END", scraddr)
				msgask = "MEIO e INOCULO"
			default:
				fmt.Println("ERROR SCP RUN JOB:", bioid, " ASK invalido", subpars)
				return false
			}
			ret1 := scp_sendmsg_orch(cmd1)
			fmt.Println("DEBUG SCP RUN JOB:: CMD =", cmd1, "\tRET =", ret1)
			if !strings.Contains(ret1, scp_ack) && !testmode {
				fmt.Println("ERROR SCP RUN JOB:", bioid, " erro ao enviar PUT screen", scraddr, ret1)
				return false
			}
			cmd2 := fmt.Sprintf("CMD/%s/GET/S451/END", scraddr)
			board_add_message("ABiorreator " + bioid + " aguardando " + msgask)
			t_start := time.Now()
			for {
				ret2 := scp_sendmsg_orch(cmd2)
				// fmt.Println("DEBUG SCP RUN JOB:: CMD =", cmd2, "\tRET =", ret2)
				if !strings.Contains(ret2, scp_ack) && !testmode {
					fmt.Println("ERROR SCP RUN JOB:", bioid, " erro ao envirar GET screen", scraddr, ret2)
					return false
				}
				data := scp_splitparam(ret2, "/")
				if len(data) > 1 {
					if data[1] == "1" {
						break
					}
				}
				t_elapsed := time.Since(t_start).Seconds()
				if t_elapsed > scp_maxtimewithdraw {
					fmt.Println("DEBUG SCP RUN JOB: Tempo maximo de ASK esgotado", bioid, t_elapsed, scp_maxtimewithdraw)
					if !testmode {
						return false
					}
					break
				}
				time.Sleep(scp_refreshwait * time.Millisecond)
			}
		} else {
			fmt.Println("ERROR SCP RUN JOB: Falta parametros em", scp_job_org, params)
			return false
		}

	case scp_job_done:
		bio[ind].Status = bio_ready
		board_add_message("CCultivo concluído no " + bioid + " - Pronto para Desenvase")
		return true

	case scp_job_wait:
		var time_int uint64
		var err error
		if len(subpars) > 1 {
			switch subpars[0] {
			case scp_par_time:
				time_str := subpars[1]
				time_int, err = strconv.ParseUint(time_str, 10, 32)
				if err != nil {
					fmt.Println("ERROR SCP RUN JOB: WAIT TIME invalido", time_str, params)
					return false
				}
				time_dur := time.Duration(time_int)
				fmt.Println("DEBUG SCP RUN JOB: WAIT de", time_dur.Seconds(), "segundos")
				time.Sleep(time_dur * time.Second)
			case scp_par_volume:
				var vol_max uint64
				var err error
				vol_str := subpars[1]
				vol_max, err = strconv.ParseUint(vol_str, 10, 32)
				if err != nil {
					fmt.Println("ERROR SCP RUN JOB: WAIT VOLUME invalido", vol_str, params)
					return false
				}
				if vol_max > uint64(bio_cfg[bioid].Maxvolume) {
					fmt.Println("ERROR SCP RUN JOB: WAIT VOLUME maior do que maximo do Biorreator", vol_max, bioid, bio_cfg[bioid].Maxvolume)
					return false
				}
				t_start := time.Now()
				for {
					vol_now := uint64(bio[ind].Volume)
					t_elapsed := time.Since(t_start).Seconds()
					if vol_now >= vol_max {
						break
					}
					if t_elapsed > scp_maxtimewithdraw {
						fmt.Println("DEBUG SCP RUN JOB: Tempo maximo de withdraw esgotado", t_elapsed, scp_maxtimewithdraw)
						if !testmode {
							return false
						}
						break
					}
					time.Sleep(scp_refreshwait * time.Millisecond)
				}
			}
		} else {
			fmt.Println("ERROR SCP RUN JOB: Falta parametros em", scp_job_org, params)
			return false
		}

	case scp_job_on:
		if len(subpars) > 1 {
			device := subpars[0]
			switch device {
			case scp_dev_aero:
				perc_str := subpars[1]
				perc_int, err := strconv.Atoi(perc_str)
				if err != nil {
					checkErr(err)
					fmt.Println("ERROR SCP RUN JOB: Parametros invalido em", scp_job_on, params)
					return false
				}
				if !scp_turn_aero(bioid, true, 1, perc_int) {
					fmt.Println("ERROR SCP RUN JOB: Erro ao ligar aerador em", bioid)
					return false
				}
			case scp_dev_pump:
				valvs := []string{}
				for k := 1; k < len(subpars) && subpars[k] != "END"; k++ {
					v := bioid + "/" + subpars[k]
					valvs = append(valvs, v)
				}
				if !scp_turn_pump(scp_bioreactor, bioid, valvs, 1) {
					fmt.Println("ERROR SCP RUN JOB: Erro ao ligar bomba em", bioid, valvs)
					return false
				}
			case scp_dev_water:
				totem := subpars[1]
				totem_ind := get_totem_index(totem)
				if totem_ind < 0 {
					fmt.Println("ERRO SCP RUN JOB: Totem nao existe", totem)
					return false
				}
				use_spball := false
				if len(subpars) > 2 && subpars[2] != "END" {
					if subpars[2] == scp_dev_sprayball {
						use_spball = true
					} else {
						fmt.Println("ERROR SCP RUN JOB: ON Parametro invalido", bioid, subpars)
						return false
					}
				}
				pathid := totem + "-" + bioid
				pathstr := paths[pathid].Path
				if len(pathstr) == 0 {
					fmt.Println("ERRO SCP RUN JOB: path nao existe", pathid)
					return false
				}
				var npath string
				if use_spball {
					npath = strings.Replace(pathstr, "/V4", "/V8", -1)
					spball_valv := bioid + "/V3"
					npath = spball_valv + "," + npath
				} else {
					npath = pathstr
				}
				fmt.Println("npath=", npath)
				vpath := scp_splitparam(npath, ",")
				watervalv := totem + "/V1"
				n := len(vpath)
				vpath = append(vpath[:n-1], watervalv)
				vpath = append(vpath, "END")
				fmt.Println("DEBUG", vpath)
				if !scp_turn_pump(scp_totem, totem, vpath, 1) {
					fmt.Println("ERROR SCP RUN JOB: Erro ao ligar bomba em", bioid, valvs)
					return false
				}
			}
		} else {
			fmt.Println("ERROR SCP RUN JOB: Falta parametros em", scp_job_on, params)
			return false
		}
	case scp_job_off:
		if len(subpars) > 0 {
			device := subpars[0]
			switch device {
			case scp_dev_aero:
				if !scp_turn_aero(bioid, true, 0, 0) {
					fmt.Println("ERROR SCP RUN JOB: Erro ao desligar aerador em", bioid)
					return false
				}
			case scp_dev_pump:
				valvs := []string{}
				for k := 1; k < len(subpars) && subpars[k] != "END"; k++ {
					v := bioid + "/" + subpars[k]
					valvs = append(valvs, v)
				}
				if !scp_turn_pump(scp_bioreactor, bioid, valvs, 0) {
					fmt.Println("ERROR SCP RUN JOB: Erro ao desligar bomba em", bioid, valvs)
					return false
				}
			case scp_dev_water:
				totem := subpars[1]
				totem_ind := get_totem_index(totem)
				if totem_ind < 0 {
					fmt.Println("ERRO SCP RUN JOB: Totem nao existe", totem)
					return false
				}
				use_spball := false
				if len(subpars) > 2 && subpars[2] != "END" {
					if subpars[2] == scp_dev_sprayball {
						use_spball = true
					} else {
						fmt.Println("ERROR SCP RUN JOB: ON Parametro invalido", bioid, subpars)
						return false
					}
				}
				pathid := totem + "-" + bioid
				pathstr := paths[pathid].Path
				if len(pathstr) == 0 {
					fmt.Println("ERRO SCP RUN JOB: path nao existe", pathid)
					return false
				}
				var npath string
				if use_spball {
					npath = strings.Replace(pathstr, "/V4", "/V8", -1)
					spball_valv := bioid + "/V3"
					npath = spball_valv + "," + npath
				} else {
					npath = pathstr
				}
				fmt.Println("npath=", npath)
				vpath := scp_splitparam(npath, ",")
				watervalv := totem + "/V1"
				n := len(vpath)
				vpath = append(vpath[:n-1], watervalv)
				vpath = append(vpath, "END")
				if !scp_turn_pump(scp_totem, totem, vpath, 0) {
					fmt.Println("ERROR SCP RUN JOB: Erro ao ligar bomba em", bioid, valvs)
					return false
				}
			}
		} else {
			fmt.Println("ERROR SCP RUN JOB: Falta parametros em", scp_job_off, params)
			return false
		}
	default:
		fmt.Println("ERROR SCP RUN JOB: JOB invalido", bioid, job, params)
	}
	time.Sleep(1000 * time.Millisecond)
	return true
}

func scp_run_bio(bioid string) {
	fmt.Println("STARTANDO RUN", bioid)
	ind := get_bio_index(bioid)
	if ind < 0 {
		fmt.Println("ERROR SCP RUN BIO: Biorreator nao existe", bioid)
		return
	}
	for bio[ind].Status != bio_die {
		if len(bio[ind].Queue) > 0 && bio[ind].Status != bio_nonexist && bio[ind].Status != bio_pause && bio[ind].Status != bio_error {
			var ret bool = false
			job := pop_first_job(bioid, false)
			if len(job) > 0 {
				ret = scp_run_job(bioid, job)
			}
			if !ret {
				fmt.Println("ERROR SCP RUN BIO: Nao foi possivel executar job", bioid, job)
			} else {
				pop_first_job(bioid, true)
			}
		}
		time.Sleep(scp_schedwait * time.Millisecond)
	}
}

func scp_run_devs() {
	for _, b := range bio {
		go scp_run_bio(b.BioreactorID)
	}
}

func scp_scheduler() {
	if !devsrunning {
		scp_run_devs()
	}
	schedrunning = true
	for schedrunning == true {
		for k, b := range bio {
			// fmt.Println(k, " bio =", b)
			r := pop_first_sched(b.BioreactorID, false)
			if len(r.Bioid) > 0 {
				if b.Status == bio_empty && len(b.Queue) == 0 { // && b.Volume == 0
					fmt.Println("\n", k, " Schedule inicial", schedule, "//", len(schedule), "POP de ", b.BioreactorID)
					s := pop_first_sched(b.BioreactorID, true)
					fmt.Println("Schedule depois do POP", schedule, "//", len(schedule), "\n\n")
					if len(s.Bioid) > 0 {
						orginfo := []string{"ORG/" + s.OrgCode + ",END"}
						bio[k].Queue = append(orginfo, recipe...)
						if autowithdraw {
							wdraw := []string{"STATUS/DESENVASE,END", "RUN/WITHDRAW,END", "RUN/CIP/END"}
							bio[k].Queue = append(bio[k].Queue, wdraw...)
						}
						bio[k].Status = bio_starting
						fmt.Println("DEBUG SCP SCHEDULER: Biorreator", b.BioreactorID, " ira produzir", s.OrgCode, "-", bio[k].Organism)
					}
				}
			}
		}
		time.Sleep(scp_schedwait * time.Millisecond)
	}
}

func create_sched(lista []string) int {
	tot := 0
	for _, i := range lista {
		if i == "END" {
			break
		}
		item := scp_splitparam(i, ",")
		bioid := item[0]
		bioseq, _ := strconv.Atoi(item[1])
		orgcode := item[2]
		ind := get_bio_index(bioid)
		if ind < 0 {
			fmt.Println("ERROR CREATE SCHED: Biorreator nao existe", bioid)
		} else {
			schedule = append(schedule, Scheditem{bioid, bioseq, orgcode})
			tot++
		}
	}
	fmt.Println(schedule)
	return tot
}

func scp_process_conn(conn net.Conn) {
	buf := make([]byte, 512)
	n, err := conn.Read(buf)
	if err != nil {
		checkErr(err)
		return
	}
	//fmt.Printf("msg: %s\n", buf[:n])
	params := scp_splitparam(string(buf[:n]), "/")
	// fmt.Println(params)
	scp_command := params[0]
	switch scp_command {
	case scp_sched:
		scp_object := params[1]
		switch scp_object {
		case scp_biofabrica:
			lista := params[2:]
			n := create_sched(lista)
			if n > 0 && !schedrunning {
				go scp_scheduler()
			}
		}
	case scp_get:
		scp_object := params[1]
		switch scp_object {
		case scp_bioreactor:
			if params[2] == "END" {
				buf, err := json.Marshal(bio)
				checkErr(err)
				conn.Write([]byte(buf))
			} else {
				ind := get_bio_index(params[2])
				if ind >= 0 {
					buf, err := json.Marshal(bio[ind])
					checkErr(err)
					conn.Write([]byte(buf))
				} else {
					conn.Write([]byte(scp_err))
				}
			}

		case scp_ibc:
			if params[2] == "END" {
				buf, err := json.Marshal(ibc)
				checkErr(err)
				conn.Write([]byte(buf))
			} else {
				ind := get_ibc_index(params[2])
				if ind >= 0 {
					buf, err := json.Marshal(ibc[ind])
					checkErr(err)
					conn.Write([]byte(buf))
				} else {
					conn.Write([]byte(scp_err))
				}
			}

		case scp_biofabrica:
			// fmt.Println(params)
			if params[2] == "END" {
				buf, err := json.Marshal(biofabrica)
				checkErr(err)
				conn.Write([]byte(buf))
			} else {
				conn.Write([]byte(scp_err))
			}

		case scp_totem:
			// fmt.Println(params)
			if params[2] == "END" {
				buf, err := json.Marshal(totem)
				checkErr(err)
				conn.Write([]byte(buf))
			} else {
				ind := get_totem_index(params[2])
				if ind >= 0 {
					buf, err := json.Marshal(totem[ind])
					checkErr(err)
					conn.Write([]byte(buf))
				} else {
					conn.Write([]byte(scp_err))
				}
			}

		default:
			conn.Write([]byte(scp_err))
		}
	case scp_put:
		scp_object := params[1]
		switch scp_object {
		case scp_bioreactor:
			//fmt.Println("obj=", scp_object)
			bioid := params[2]
			ind := get_bio_index(bioid)
			if ind < 0 {
				conn.Write([]byte(scp_err))
			} else {
				subparams := scp_splitparam(params[3], ",")
				scp_device := subparams[0]
				// fmt.Println("subparams=", subparams)
				switch scp_device {
				case scp_par_out:
					outid := subparams[1]
					if outid == "Descarte" {
						outid = "DROP"
					} else if outid == "Externo" {
						outid = "OUT"
					}
					tmp := bioid + "-" + outid
					_, ok := paths[tmp]
					if ok {
						bio[ind].OutID = outid
						conn.Write([]byte(scp_ack))
					} else {
						fmt.Println("ID de saida", outid, " nao existe")
						conn.Write([]byte(scp_err))
					}

				case scp_par_withdraw:
					vol, err := strconv.Atoi(subparams[1])
					checkErr(err)
					if err == nil {
						bio[ind].Withdraw = uint32(vol)
					}
					go scp_run_withdraw(scp_bioreactor, bioid)
					conn.Write([]byte(scp_ack))
				case scp_dev_pump:
					var cmd2, cmd3 string
					value, err := strconv.ParseBool(subparams[1])
					checkErr(err)
					biodev := bio_cfg[bioid].Deviceaddr
					bioscr := bio_cfg[bioid].Screenaddr
					pumpdev := bio_cfg[bioid].Pump_dev
					bio[ind].Pumpstatus = value
					if value {
						cmd2 = "CMD/" + biodev + "/PUT/" + pumpdev + ",1/END"
						cmd3 = "CMD/" + bioscr + "/PUT/S270,1/END"
					} else {
						cmd2 = "CMD/" + biodev + "/PUT/" + pumpdev + ",0/END"
						cmd3 = "CMD/" + bioscr + "/PUT/S270,0/END"
					}
					ret2 := scp_sendmsg_orch(cmd2)
					fmt.Println("RET CMD2 =", ret2)
					ret3 := scp_sendmsg_orch(cmd3)
					fmt.Println("RET CMD3 =", ret3)
					conn.Write([]byte(scp_ack))

				case scp_dev_aero:
					var cmd1, cmd2, cmd3 string
					value, err := strconv.ParseBool(subparams[1])
					checkErr(err)
					bio[ind].Aerator = value
					biodev := bio_cfg[bioid].Deviceaddr
					bioscr := bio_cfg[bioid].Screenaddr
					aerodev := bio_cfg[bioid].Aero_dev
					if value {
						cmd1 = "CMD/" + biodev + "/PUT/D27,1/END"
						cmd2 = "CMD/" + biodev + "/PUT/" + aerodev + ",255/END"
						cmd3 = "CMD/" + bioscr + "/PUT/S271,1/END"

					} else {
						cmd1 = "CMD/" + biodev + "/PUT/D27,0/END"
						cmd2 = "CMD/" + biodev + "/PUT/" + aerodev + ",0/END"
						cmd3 = "CMD/" + bioscr + "/PUT/S271,0/END"
					}
					ret1 := scp_sendmsg_orch(cmd1)
					fmt.Println("RET CMD1 =", ret1)
					ret2 := scp_sendmsg_orch(cmd2)
					fmt.Println("RET CMD2 =", ret2)
					ret3 := scp_sendmsg_orch(cmd3)
					fmt.Println("RET CMD3 =", ret3)
					conn.Write([]byte(scp_ack))

				case scp_dev_valve:
					// var cmd2, cmd3 string
					value_valve, err := strconv.Atoi(subparams[1])
					checkErr(err)
					value_status, err := strconv.Atoi(subparams[2])
					checkErr(err)
					//fmt.Println(value_valve, value_status)
					if (value_valve >= 0) && (value_valve < bio_max_valves) {
						valvid := fmt.Sprintf("V%d", value_valve+1)
						set_valv_status(scp_bioreactor, bioid, valvid, value_status)
						// bio[ind].Valvs[value_valve] = value_status
						conn.Write([]byte(scp_ack))
						// valve_str2 := fmt.Sprintf("%d", value_valve+201)
						// biodev := bio_cfg[bioid].Deviceaddr
						// bioscr := bio_cfg[bioid].Screenaddr
						// valvaddr := bio_cfg[bioid].Valv_devs[value_valve]
						// if value_status > 0 {
						// 	cmd2 = "CMD/" + biodev + "/PUT/" + valvaddr + ",1/END"
						// 	cmd3 = "CMD/" + bioscr + "/PUT/S" + valve_str2 + ",1/END"
						// } else {
						// 	cmd2 = "CMD/" + biodev + "/PUT/" + valvaddr + ",0/END"
						// 	cmd3 = "CMD/" + bioscr + "/PUT/S" + valve_str2 + ",0/END"
						// }
						// ret2 := scp_sendmsg_orch(cmd2)
						// fmt.Println("RET CMD2 =", ret2)
						// ret3 := scp_sendmsg_orch(cmd3)
						// fmt.Println("RET CMD3 =", ret3)
						// conn.Write([]byte(scp_ack))
					}
				default:
					conn.Write([]byte(scp_err))
				}
			}

		case scp_ibc:
			ibcid := params[2]
			ind := get_ibc_index(ibcid)
			if ind < 0 {
				conn.Write([]byte(scp_err))
			} else {
				subparams := scp_splitparam(params[3], ",")
				scp_device := subparams[0]
				switch scp_device {
				case scp_par_out:
					outid := subparams[1]
					if outid == "Descarte" {
						outid = "DROP"
					} else if outid == "Externo" {
						outid = "OUT"
					}
					tmp := ibcid + "-" + outid
					_, ok := paths[tmp]
					if ok {
						ibc[ind].OutID = outid
						conn.Write([]byte(scp_ack))
					} else {
						fmt.Println("ID de saida", outid, " nao existe")
						conn.Write([]byte(scp_err))
					}
				case scp_par_withdraw:
					vol, err := strconv.Atoi(subparams[1])
					checkErr(err)
					if err == nil {
						ibc[ind].Withdraw = uint32(vol)
					}
					go scp_run_withdraw(scp_ibc, ibcid)
					conn.Write([]byte(scp_ack))
				case scp_dev_pump:
					var cmd2 string
					value, err := strconv.ParseBool(subparams[1])
					checkErr(err)
					ibc[ind].Pumpstatus = value
					ibcdev := ibc_cfg[ibcid].Deviceaddr
					pumpdev := ibc_cfg[ibcid].Pump_dev
					//ibcscr := bio_cfg[ibcid].Screenaddr
					if value {
						cmd2 = "CMD/" + ibcdev + "/PUT/" + pumpdev + ",1/END"
					} else {
						cmd2 = "CMD/" + ibcdev + "/PUT/" + pumpdev + ",0/END"
					}

					ret2 := scp_sendmsg_orch(cmd2)
					fmt.Println("RET CMD2 =", ret2)

					conn.Write([]byte(scp_ack))

				case scp_dev_valve:
					// var cmd2 string
					value_valve, err := strconv.Atoi(subparams[1])
					checkErr(err)
					value_status, err := strconv.Atoi(subparams[2])
					checkErr(err)
					//fmt.Println(value_valve, value_status)
					if (value_valve >= 0) && (value_valve < bio_max_valves) {
						valvid := fmt.Sprintf("V%d", value_valve+1)
						set_valv_status(scp_ibc, ibcid, valvid, value_status)
						// ibc[ind].Valvs[value_valve] = value_status

						// ibcdev := ibc_cfg[ibcid].Deviceaddr
						// valvaddr := ibc_cfg[ibcid].Valv_devs[value_valve]
						// //ibcscr := bio_cfg[ibcid].Screenaddr
						// if value_status == 1 {
						// 	cmd2 = "CMD/" + ibcdev + "/PUT/" + valvaddr + ",1/END"
						// } else if value_status == 0 {
						// 	cmd2 = "CMD/" + ibcdev + "/PUT/" + valvaddr + ",0/END"
						// }

						// ret2 := scp_sendmsg_orch(cmd2)
						// fmt.Println("RET CMD2 =", ret2)
						conn.Write([]byte(scp_ack))
					}
				default:
					conn.Write([]byte(scp_err))
				}
			}

		case scp_totem:
			totemid := params[2]
			ind := get_totem_index(totemid)
			if ind < 0 {
				conn.Write([]byte(scp_err))
			} else {
				subparams := scp_splitparam(params[3], ",")
				scp_device := subparams[0]
				switch scp_device {
				case scp_dev_pump:
					var cmd2 string
					value, err := strconv.ParseBool(subparams[1])
					checkErr(err)
					totem[ind].Pumpstatus = value
					totemdev := totem_cfg[totemid].Deviceaddr
					pumpdev := totem_cfg[totemid].Pumpdev
					if value {
						cmd2 = "CMD/" + totemdev + "/PUT/" + pumpdev + ",1/END"
					} else {
						cmd2 = "CMD/" + totemdev + "/PUT/" + pumpdev + ",0/END"
					}
					ret2 := scp_sendmsg_orch(cmd2)
					fmt.Println("RET CMD2 =", ret2)
					conn.Write([]byte(scp_ack))

				case scp_dev_valve:
					// var cmd2 string
					value_valve, err := strconv.Atoi(subparams[1])
					checkErr(err)
					value_status, err := strconv.Atoi(subparams[2])
					checkErr(err)
					//fmt.Println(value_valve, value_status)
					if (value_valve >= 0) && (value_valve < bio_max_valves) {
						valvid := fmt.Sprintf("V%d", value_valve+1)
						set_valv_status(scp_totem, totemid, valvid, value_status)
						// totem[ind].Valvs[value_valve] = value_status

						// totemdev := totem_cfg[totemid].Deviceaddr
						// valvaddr := totem_cfg[totemid].Valv_devs[value_valve]
						// if value_status == 1 {
						// 	cmd2 = "CMD/" + totemdev + "/PUT/" + valvaddr + ",1/END"
						// } else if value_status == 0 {
						// 	cmd2 = "CMD/" + totemdev + "/PUT/" + valvaddr + ",0/END"
						// }

						// ret2 := scp_sendmsg_orch(cmd2)
						// fmt.Println("RET CMD2 =", ret2)
						conn.Write([]byte(scp_ack))
					}
				default:
					conn.Write([]byte(scp_err))
				}
			}

		case scp_biofabrica:
			subparams := scp_splitparam(params[2], ",")
			//fmt.Println("---subparams=", subparams)
			scp_device := subparams[0]
			switch scp_device {
			case scp_dev_pump:
				var cmd2 string
				value, err := strconv.ParseBool(subparams[1])
				checkErr(err)
				biofabrica.Pumpwithdraw = value
				devaddr := biofabrica_cfg["PBF01"].Deviceaddr
				devport := biofabrica_cfg["PBF01"].Deviceport
				if value {
					cmd2 = "CMD/" + devaddr + "/PUT/" + devport + ",1/END"
				} else {
					cmd2 = "CMD/" + devaddr + "/PUT/" + devport + ",0/END"
				}
				ret2 := scp_sendmsg_orch(cmd2)
				fmt.Println("RET CMD2 =", ret2)

				conn.Write([]byte(scp_ack))

			case scp_dev_valve:
				// var cmd2 string
				value_valve, err := strconv.Atoi(subparams[1])
				checkErr(err)
				value_status, err := strconv.Atoi(subparams[2])
				checkErr(err)
				//fmt.Println(value_valve, value_status)
				if (value_valve >= 0) && (value_valve < 9) {
					// biofabrica.Valvs[value_valve] = value_status
					devid := fmt.Sprintf("VBF%02d", value_valve+1)
					set_valv_status(scp_biofabrica, "BIOFABRICA", devid, value_status)
					// devaddr := biofabrica_cfg[devid].Deviceaddr
					// devport := biofabrica_cfg[devid].Deviceport
					// if value_status == 1 {
					// 	cmd2 = "CMD/" + devaddr + "/PUT/" + devport + ",1/END"
					// } else if value_status == 0 {
					// 	cmd2 = "CMD/" + devaddr + "/PUT/" + devport + ",0/END"
					// }
					// fmt.Println("biofabrica valvula", cmd2)
					// ret2 := scp_sendmsg_orch(cmd2)
					// fmt.Println("RET CMD2 =", ret2)
					conn.Write([]byte(scp_ack))
				}
			default:
				conn.Write([]byte(scp_err))
			}

		default:
			conn.Write([]byte(scp_err))
		}
		// fmt.Println(valvs)
	default:
		conn.Write([]byte(scp_err))
	}
	// scp_sendmsg_orch(string(buf[:n]))
	// conn.Write([]byte(scp_ack))
	conn.Close()
}

func scp_master_ipc() {
	_, f_exist := os.Stat(scp_ipc_name)
	if f_exist == nil {
		f_delete := os.Remove(scp_ipc_name)
		if f_delete != nil {
			checkErr(f_delete)
			return
		}
	}
	ipc, err := net.Listen("unix", scp_ipc_name)
	if err != nil {
		checkErr(err)
		return
	}
	defer ipc.Close()

	for {
		conn, err := ipc.Accept()
		if err != nil {
			checkErr(err)
		}

		go scp_process_conn(conn)

	}
}

func main() {
	if testmode {
		fmt.Println("WARN:  EXECUTANDO EM TESTMODE\n\n\n")
	}
	norgs := load_organisms("organismos_conf.csv")
	if norgs < 0 {
		log.Fatal("Não foi possivel ler o arquivo de organismos")
	}
	recipe = load_tasks_conf("receita_conf.csv")
	if recipe == nil {
		log.Fatal("Não foi possivel ler o arquivo contendo a receita de producao")
	}
	cipbio = load_tasks_conf("cip_bio_conf.csv")
	if recipe == nil {
		log.Fatal("Não foi possivel ler o arquivo contendo ciclo de CIP")
	}
	fmt.Println("receita=", recipe)
	fmt.Println("cip=", cipbio)
	nibccfg := load_ibcs_conf("ibc_conf.csv")
	if nibccfg < 1 {
		log.Fatal("FATAL: Arquivo de configuracao dos IBCs nao encontrado")
	}
	nbiocfg := load_bios_conf("bio_conf.csv")
	if nbiocfg < 1 {
		log.Fatal("FATAL: Arquivo de configuracao dos Bioreatores nao encontrado")
	}
	ntotemcfg := load_totems_conf("totem_conf.csv")
	if ntotemcfg < 1 {
		log.Fatal("FATAL: Arquivo de configuracao dos Totems nao encontrado")
	}
	nbiofabricacfg := load_biofabrica_conf("biofabrica_conf.csv")
	if nbiofabricacfg < 1 {
		log.Fatal("FATAL: Arquivo de configuracao da Biofabrica nao encontrado")
	}
	npaths := load_paths_conf("paths_conf.csv")
	if npaths < 1 {
		log.Fatal("FATAL: Arquivo de configuracao de PATHs invalido")
	}
	// fmt.Println("BIO cfg", bio_cfg)
	// fmt.Println("IBC cfg", ibc_cfg)
	// fmt.Println("TOTEM cfg", totem_cfg)
	// fmt.Println("Biofabrica cfg", biofabrica_cfg)
	// fmt.Println("PATHs ", paths)
	// fmt.Println("BIO ", bio)
	// fmt.Println("IBC ", ibc)
	// fmt.Println("TOTEM ", totem)
	// fmt.Println("Biofabrica ", biofabrica)
	valvs = make(map[string]int, 0)
	load_all_data(bio_data_filename)
	go scp_setup_devices()

	go scp_get_alldata()
	scp_master_ipc()
}
