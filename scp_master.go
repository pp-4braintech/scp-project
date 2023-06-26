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

const demo = false
const testmode = true

const scp_ack = "ACK"
const scp_err = "ERR"
const scp_get = "GET"
const scp_put = "PUT"
const scp_fail = "FAIL"
const scp_dev_pump = "PUMP"
const scp_dev_aero = "AERO"
const scp_dev_valve = "VALVE"
const scp_par_withdraw = "WITHDRAW"
const scp_par_out = "OUT"
const scp_bioreactor = "BIOREACTOR"
const scp_biofabrica = "BIOFABRICA"
const scp_totem = "TOTEM"
const scp_ibc = "IBC"
const scp_donothing = "NOTHING"
const scp_orch_addr = ":7007"
const scp_ipc_name = "/tmp/scp_master.sock"
const scp_refreshwait = 500
const scp_refreshsleep = 2500

const scp_timewaitvalvs = 12000
const scp_maxtimewithdraw time.Duration = 600000

const bio_diametro = 1430  // em mm
const bio_v1_zero = 1483.0 // em mm
const bio_v2_zero = 1502.0 // em mm
const ibc_v1_zero = 2652.0 // em mm   2647

// const scp_join = "JOIN"
const bio_data_filename = "dumpdata"

const bio_nonexist = "NULL"
const bio_cip = "CIP"
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

const TEMPMAX = 120

type Bioreact struct {
	BioreactorID string
	// Deviceaddr   string
	// Screenaddr   string
	Status      string
	Organism    string
	Volume      uint32
	Level       uint8
	Pumpstatus  bool
	Aerator     bool
	Valvs       [8]int
	Temperature float32
	PH          float32
	Step        [2]int
	Timeleft    [2]int
	Timetotal   [2]int
	Withdraw    uint32
	OutID       string
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
}

type Path struct {
	FromID string
	ToID   string
	Path   string
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

var finishedsetup = false

var ibc_cfg map[string]IBC_cfg
var bio_cfg map[string]Bioreact_cfg
var totem_cfg map[string]Totem_cfg
var biofabrica_cfg map[string]Biofabrica_cfg
var paths map[string]Path
var valvs map[string]int

var bio = []Bioreact{
	{"BIOR01", bio_nonexist, "", 2000, 10, false, false, [8]int{0, 0, 0, 0, 0, 0, 0, 0}, 0, 0, [2]int{2, 5}, [2]int{25, 17}, [2]int{48, 0}, 0, "OUT"},
	{"BIOR02", bio_nonexist, "", 0, 0, false, false, [8]int{0, 0, 0, 0, 0, 0, 0, 0}, 0, 0, [2]int{1, 1}, [2]int{0, 5}, [2]int{0, 30}, 0, "OUT"},
	{"BIOR03", bio_nonexist, "", 0, 0, false, false, [8]int{0, 0, 0, 0, 0, 0, 0, 0}, 0, 0, [2]int{1, 1}, [2]int{0, 10}, [2]int{0, 30}, 0, "OUT"},
	{"BIOR04", bio_nonexist, "", 0, 0, false, false, [8]int{0, 0, 0, 0, 0, 0, 0, 0}, 0, 0, [2]int{1, 1}, [2]int{0, 5}, [2]int{0, 15}, 0, "OUT"},
	{"BIOR05", bio_nonexist, "Tricoderma harzianum", 0, 0, false, false, [8]int{0, 0, 0, 0, 0, 0, 0, 0}, 0, 0, [2]int{5, 5}, [2]int{0, 0}, [2]int{72, 0}, 0, "OUT"},
	{"BIOR06", bio_nonexist, "", 0, 0, false, false, [8]int{0, 0, 0, 0, 0, 0, 0, 0}, 0, 0, [2]int{0, 0}, [2]int{0, 0}, [2]int{0, 0}, 0, "OUT"},
}

var ibc = []IBC{
	{"IBC01", bio_nonexist, "Bacillus Subtilis", 1000, 2, false, [4]int{0, 0, 0, 0}, [2]int{24, 15}, 0, "OUT"},
	{"IBC02", bio_nonexist, "Bacillus Megaterium", 0, 0, false, [4]int{0, 0, 0, 0}, [2]int{12, 5}, 0, "OUT"},
	{"IBC03", bio_nonexist, "Bacillus Amyloliquefaciens", 0, 0, false, [4]int{0, 0, 0, 0}, [2]int{0, 30}, 0, "OUT"},
	{"IBC04", bio_nonexist, "Azospirilum brasiliense", 0, 0, false, [4]int{0, 0, 0, 0}, [2]int{4, 50}, 0, "OUT"},
	{"IBC05", bio_nonexist, "Tricoderma harzianum", 0, 0, false, [4]int{0, 0, 0, 0}, [2]int{13, 17}, 0, "OUT"},
	{"IBC06", bio_nonexist, "Tricoderma harzianum", 0, 0, false, [4]int{0, 0, 0, 0}, [2]int{0, 5}, 0, "OUT"},
	{"IBC07", bio_nonexist, "", 0, 0, false, [4]int{0, 0, 0, 0}, [2]int{0, 0}, 0, "OUT"},
}

var totem = []Totem{
	{"TOTEM01", bio_nonexist, false, [2]int{0, 0}, [4]int{0, 0, 0, 0}},
	{"TOTEM02", bio_nonexist, false, [2]int{0, 0}, [4]int{0, 0, 0, 0}},
}

var biofabrica = Biofabrica{
	"BIOFABRICA001", [9]int{0, 0, 0, 0, 0, 0, 0, 0, 0}, false,
}

func checkErr(err error) {
	if err != nil {
		log.Println("[SCP ERROR]", err)
	}
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
		if !strings.Contains(r[0], "#") && len(r) == 27 {
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

			bio_cfg[id] = Bioreact_cfg{id, dev_addr, screen_addr, uint32(voltot), pumpdev, aerodev, aerorele,
				[5]string{perdev1, perdev2, perdev3, perdev4, perdev5},
				[8]string{vdev1, vdev2, vdev3, vdev4, vdev5, vdev6, vdev7, vdev8},
				[2]string{voldev1, voldev2}, phdev, tempdev, lhigh, llow, emerg}
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
			path_id := from_id + "-" + to_id
			pathstr := ""
			for i := 2; i < len(r); i++ {
				pathstr += r[i] + ","
			}
			pathstr += "END"
			paths[path_id] = Path{from_id, to_id, pathstr}
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
				valve_scrstr = fmt.Sprintf("%d", v+200)
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
	if !strings.Contains(ret1[:2], scp_ack) && !testmode {
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
	} else {
		return -1
	}

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

	ret := make([]byte, 1024)
	_, err = con.Read(ret)
	if err != nil {
		checkErr(err)
		return scp_err
	}
	//fmt.Println("Recebido:", string(ret))
	return string(ret)
}

func scp_setup_devices() {
	if demo {
		return
	}
	fmt.Println("CONFIGURANDO DISPOSITIVOS")
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
				if ret == scp_err {
					nerr++
				}
				fmt.Println(ret)
				if ret[0:2] == "DIE" {
					fmt.Println("SLAVE ERROR - DIE")
					break
				}
				time.Sleep(scp_refreshwait / 2 * time.Millisecond)
			}
			i := get_bio_index(b.BioreactorID)
			if i >= 0 {
				if nerr == 0 {
					bio[i].Status = bio_empty
				} else {
					bio[i].Status = bio_error
				}
			}
		}
	}
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
				fmt.Println(ret)
				if ret[0:2] == "DIE" {
					fmt.Println("SLAVE ERROR - DIE")
					break
				}
				time.Sleep(scp_refreshwait / 2 * time.Millisecond)
			}
			i := get_ibc_index(ib.IBCID)
			if i >= 0 {
				if nerr == 0 {
					ibc[i].Status = bio_empty
				} else {
					ibc[i].Status = bio_error
				}
			}
		}
	}

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
				fmt.Println(ret)
				if ret[0:2] == "DIE" {
					fmt.Println("SLAVE ERROR - DIE")
					break
				}
				time.Sleep(scp_refreshwait / 2 * time.Millisecond)
			}
			i := get_totem_index(tot.TotemID)
			if i >= 0 {
				if nerr == 0 {
					totem[i].Status = bio_ready
				} else {
					totem[i].Status = bio_error
				}
			}
		}
	}

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
								} else {
									bio[k].Status = bio_ready
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
								bio[k].Level = level_int
								// levels := fmt.Sprintf("%d", level_int)
								// cmd := "CMD/" + ibc_cfg[b.IBCID].Screenaddr + "/PUT/S231," + levels + "/END"
								// ret := scp_sendmsg_orch(cmd)
								// fmt.Println("SCREEN:", cmd, level, levels, ret)
							}
							if volc == 0 {
								ibc[k].Status = bio_empty
							} else {
								ibc[k].Status = bio_ready
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

func scp_run_withdraw(devtype string, devid string) int {
	switch devtype {
	case scp_bioreactor:
		ind := get_bio_index(devid)
		pathid := devid + "-" + bio[ind].OutID
		pathstr := paths[pathid].Path
		if len(pathstr) == 0 {
			fmt.Println("ERRO RUN WITHDRAW: path nao existe", pathid)
			return -1
		}
		vpath := scp_splitparam(pathstr, ",")
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
						fmt.Println("ERRO RUN WITHDRAW: nao foi possivel setar valvula", p)
						return -1
					}
				} else if val == 1 {
					fmt.Println("ERRO RUN WITHDRAW: valvula ja aberta", p)
					return -1
				} else {
					fmt.Println("ERRO RUN WITHDRAW: valvula com erro", p)
					return -1
				}
			} else {
				fmt.Println("ERRO RUN WITHDRAW: valvula nao existe", p)
				return -1
			}
		}
		vol_ini := bio[ind].Volume
		time.Sleep(scp_timewaitvalvs * time.Millisecond)
		fmt.Println("WARN RUN WITHDRAW: Ligando bomba", devid)
		biodev := bio_cfg[devid].Deviceaddr
		bioscr := bio_cfg[devid].Screenaddr
		pumpdev := bio_cfg[devid].Pump_dev
		bio[ind].Pumpstatus = true
		cmd1 := "CMD/" + biodev + "/PUT/" + pumpdev + ",1/END"
		cmd2 := "CMD/" + bioscr + "/PUT/S270,1/END"
		ret1 := scp_sendmsg_orch(cmd1)
		fmt.Println("DEBUG RUN WITHDRAW: CMD1 =", cmd1, " RET=", ret1)
		ret2 := scp_sendmsg_orch(cmd2)
		fmt.Println("DEBUG RUN WITHDRAW: CMD2 =", cmd2, " RET=", ret2)
		t_start := time.Now()
		for {
			vol_now := bio[ind].Volume
			t_now := time.Now()
			if vol_ini-vol_now >= bio[ind].Withdraw {
				fmt.Println("DEBUG RUN WITHDRAW: Volume de desenvase atingido", vol_ini, vol_now)
				break
			}
			if t_now.Sub(t_start) > scp_maxtimewithdraw {
				fmt.Println("DEBUG RUN WITHDRAW: Tempo maixo de withdraw esgota", t_now.Sub(t_start), scp_maxtimewithdraw)
				break
			}
			time.Sleep(scp_refreshwait * time.Millisecond)
		}
		bio[ind].Withdraw = 0
		fmt.Println("WARN RUN WITHDRAW: Desligando bomba", devid)
		bio[ind].Pumpstatus = true
		cmd1 = "CMD/" + biodev + "/PUT/" + pumpdev + ",0/END"
		cmd2 = "CMD/" + bioscr + "/PUT/S270,1/END"
		ret1 = scp_sendmsg_orch(cmd1)
		fmt.Println("DEBUG RUN WITHDRAW: CMD1 =", cmd1, " RET=", ret1)
		ret2 = scp_sendmsg_orch(cmd2)
		fmt.Println("DEBUG RUN WITHDRAW: CMD2 =", cmd2, " RET=", ret2)
	case scp_ibc:
		ind := get_ibc_index(devid)
		pathid := devid + "-" + ibc[ind].OutID
		pathstr := paths[pathid].Path
		if len(pathstr) == 0 {
			fmt.Println("ERRO RUN WITHDRAW: path nao existe", pathid)
			return -1
		}
		vpath := scp_splitparam(pathstr, ",")
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
						fmt.Println("ERRO RUN WITHDRAW: nao foi possivel setar valvula", p)
						return -1
					}
				} else if val == 1 {
					fmt.Println("ERRO RUN WITHDRAW: valvula ja aberta", p)
					return -1
				} else {
					fmt.Println("ERRO RUN WITHDRAW: valvula com erro", p)
					return -1
				}
			} else {
				fmt.Println("ERRO RUN WITHDRAW: valvula nao existe", p)
				return -1
			}
		}
		time.Sleep(scp_timewaitvalvs * time.Millisecond)
		fmt.Println("WARN RUN WITHDRAW: Ligando bomba", devid)
		pumpdev := biofabrica_cfg["PBF01"].Deviceaddr
		pumpport := biofabrica_cfg["PBF01"].Deviceport
		biofabrica.Pumpwithdraw = true
		cmd1 := "CMD/" + pumpdev + "/PUT/" + pumpport + ",1/END"
		ret1 := scp_sendmsg_orch(cmd1)
		fmt.Println("DEBUG RUN WITHDRAW: CMD1 =", cmd1, " RET=", ret1)
	}
	return 0
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
