package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

const demo = false

const scp_ack = "ACK"
const scp_err = "ERR"
const scp_get = "GET"
const scp_put = "PUT"
const scp_dev_pump = "PUMP"
const scp_dev_aero = "AERO"
const scp_dev_valve = "VALVE"
const scp_par_withdraw = "WITHDRAW"
const scp_bioreactor = "BIOREACTOR"
const scp_biofabrica = "BIOFABRICA"
const scp_totem = "TOTEM"
const scp_ibc = "IBC"
const scp_orch_addr = ":7007"
const scp_ipc_name = "/tmp/scp_master.sock"
const scp_refreshwait = 1000
const scp_refreshsleep = 10000

// const scp_join = "JOIN"

const bio_nonexist = "NULL"
const bio_cip = "CIP"
const bio_loading = "CARREGANDO"
const bio_unloading = "ESVAZIANDO"
const bio_producting = "PRODUZINDO"
const bio_empty = "VAZIO"
const bio_done = "CONCLUIDO"
const bio_storing = "ARMAZENANDO"
const bio_error = "ERRO"
const bio_ready = "PRONTO"
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

var bio = []Bioreact{
	{"BIOR01", bio_nonexist, "", 2000, 10, false, false, [8]int{0, 0, 0, 0, 0, 0, 0, 0}, 0, 0, [2]int{2, 5}, [2]int{25, 17}, [2]int{48, 0}, 0},
	{"BIOR02", bio_nonexist, "", 0, 0, false, false, [8]int{0, 0, 0, 0, 0, 0, 0, 0}, 0, 0, [2]int{1, 1}, [2]int{0, 5}, [2]int{0, 30}, 0},
	{"BIOR03", bio_nonexist, "", 0, 0, false, false, [8]int{0, 0, 0, 0, 0, 0, 0, 0}, 0, 0, [2]int{1, 1}, [2]int{0, 10}, [2]int{0, 30}, 0},
	{"BIOR04", bio_nonexist, "", 0, 0, false, false, [8]int{0, 0, 0, 0, 0, 0, 0, 0}, 0, 0, [2]int{1, 1}, [2]int{0, 5}, [2]int{0, 15}, 0},
	{"BIOR05", bio_nonexist, "Tricoderma harzianum", 0, 0, false, false, [8]int{0, 0, 0, 0, 0, 0, 0, 0}, 0, 0, [2]int{5, 5}, [2]int{0, 0}, [2]int{72, 0}, 0},
	{"BIOR06", bio_nonexist, "", 0, 0, false, false, [8]int{0, 0, 0, 0, 0, 0, 0, 0}, 0, 0, [2]int{0, 0}, [2]int{0, 0}, [2]int{0, 0}, 0},
}

var ibc = []IBC{
	{"IBC01", bio_nonexist, "Bacillus Subtilis", 1000, 2, false, [4]int{0, 0, 0, 0}, [2]int{24, 15}, 0},
	{"IBC02", bio_nonexist, "Bacillus Megaterium", 0, 0, false, [4]int{0, 0, 0, 0}, [2]int{12, 5}, 0},
	{"IBC03", bio_nonexist, "Bacillus Amyloliquefaciens", 0, 0, false, [4]int{0, 0, 0, 0}, [2]int{0, 30}, 0},
	{"IBC04", bio_nonexist, "Azospirilum brasiliense", 0, 0, false, [4]int{0, 0, 0, 0}, [2]int{4, 50}, 0},
	{"IBC05", bio_nonexist, "Tricoderma harzianum", 0, 0, false, [4]int{0, 0, 0, 0}, [2]int{13, 17}, 0},
	{"IBC06", bio_nonexist, "Tricoderma harzianum", 0, 0, false, [4]int{0, 0, 0, 0}, [2]int{0, 5}, 0},
	{"IBC07", bio_nonexist, "", 0, 0, false, [4]int{0, 0, 0, 0}, [2]int{0, 0}, 0},
}

var totem = []Totem{
	{"TOTEM01", bio_ready, false, [2]int{0, 0}, [4]int{0, 0, 0, 0}},
	{"TOTEM02", bio_ready, false, [2]int{0, 0}, [4]int{0, 0, 0, 0}},
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

			nerr := 0
			for k, c := range cmd {
				fmt.Print(k, "  ", c, " ")
				ret := scp_sendmsg_orch(c)
				if ret == scp_err {
					nerr++
				}
				fmt.Println(ret)
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

			// cmd = append(cmd, "CMD/"+ibcaddr+"/MOD/"+b.Levelhigh[1:]+",1/END")
			// cmd = append(cmd, "CMD/"+ibcaddr+"/MOD/"+b.Levellow[1:]+",1/END")
			// cmd = append(cmd, "CMD/"+ibcaddr+"/MOD/"+b.Emergency[1:]+",1/END")

			for k, c := range cmd {
				fmt.Print(k, "  ", c, " ")
				ret := scp_sendmsg_orch(c)
				fmt.Println(ret)
				time.Sleep(scp_refreshwait / 2 * time.Millisecond)
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
	for {
		if finishedsetup {
			for k, b := range bio {
				if len(bio_cfg[b.BioreactorID].Deviceaddr) > 0 {
					i := get_bio_index(b.BioreactorID)
					if i >= 0 && (bio[i].Status != bio_nonexist && bio[i].Status != bio_error) {
						bioaddr := bio_cfg[b.BioreactorID].Deviceaddr
						tempdev := bio_cfg[b.BioreactorID].Temp_dev
						phdev := bio_cfg[b.BioreactorID].PH_dev

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
					}

				}
				time.Sleep(scp_refreshwait * time.Millisecond)
			}
			time.Sleep(scp_refreshsleep * time.Millisecond)

		}
	}
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
	fmt.Println(params)
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
			fmt.Println(params)
			if params[2] == "END" {
				buf, err := json.Marshal(biofabrica)
				checkErr(err)
				conn.Write([]byte(buf))
			} else {
				conn.Write([]byte(scp_err))
			}

		case scp_totem:
			fmt.Println(params)
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
				fmt.Println("subparams=", subparams)
				switch scp_device {
				case scp_par_withdraw:
					vol, err := strconv.Atoi(subparams[1])
					checkErr(err)
					if err == nil {
						bio[ind].Withdraw = uint32(vol)
					}
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
						//cmd1 = "CMD/" + biodev + "/MOD/" + pumpdev[1:] + ",3/END"
						cmd2 = "CMD/" + biodev + "/PUT/" + pumpdev + ",1/END"
						cmd3 = "CMD/" + bioscr + "/PUT/S270,1/END"
					} else {
						//cmd1 = "CMD/" + biodev + "/MOD/" + pumpdev[1:] + ",3/END"
						cmd2 = "CMD/" + biodev + "/PUT/" + pumpdev + ",0/END"
						cmd3 = "CMD/" + bioscr + "/PUT/S270,0/END"
					}
					// ret1 := scp_sendmsg_orch(cmd1)
					// fmt.Println("RET CMD1 =", ret1)
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
						// cmd0 = "CMD/" + biodev + "/MOD/27,3/END"
						cmd1 = "CMD/" + biodev + "/PUT/D27,1/END"
						cmd2 = "CMD/" + biodev + "/PUT/" + aerodev + ",255/END"
						cmd3 = "CMD/" + bioscr + "/PUT/S271,1/END"

					} else {
						// cmd0 = "CMD/" + biodev + "/MOD/27,3/END"
						cmd1 = "CMD/" + biodev + "/PUT/D27,0/END"
						cmd2 = "CMD/" + biodev + "/PUT/" + aerodev + ",0/END"
						cmd3 = "CMD/" + bioscr + "/PUT/S271,0/END"
					}
					// ret0 := scp_sendmsg_orch(cmd0)
					// fmt.Println("RET CMD0 =", ret0)
					ret1 := scp_sendmsg_orch(cmd1)
					fmt.Println("RET CMD1 =", ret1)
					ret2 := scp_sendmsg_orch(cmd2)
					fmt.Println("RET CMD2 =", ret2)
					ret3 := scp_sendmsg_orch(cmd3)
					fmt.Println("RET CMD3 =", ret3)
					conn.Write([]byte(scp_ack))

				case scp_dev_valve:
					var cmd2, cmd3 string
					value_valve, err := strconv.Atoi(subparams[1])
					checkErr(err)
					value_status, err := strconv.Atoi(subparams[2])
					checkErr(err)
					//fmt.Println(value_valve, value_status)
					if (value_valve >= 0) && (value_valve < bio_max_valves) {
						bio[ind].Valvs[value_valve] = value_status
						conn.Write([]byte(scp_ack))
						valve_str2 := fmt.Sprintf("%d", value_valve+201)
						biodev := bio_cfg[bioid].Deviceaddr
						bioscr := bio_cfg[bioid].Screenaddr
						valvaddr := bio_cfg[bioid].Valv_devs[value_valve]
						if value_status > 0 {
							// cmd1 = "CMD/" + biodev + "/MOD/" + valvaddr[1:] + ",3/END"
							cmd2 = "CMD/" + biodev + "/PUT/" + valvaddr + ",1/END"
							cmd3 = "CMD/" + bioscr + "/PUT/S" + valve_str2 + ",1/END"
						} else {
							// cmd1 = "CMD/" + biodev + "/MOD/" + valvaddr[1:] + ",3/END"
							cmd2 = "CMD/" + biodev + "/PUT/" + valvaddr + ",0/END"
							cmd3 = "CMD/" + bioscr + "/PUT/S" + valve_str2 + ",0/END"
						}
						// ret1 := scp_sendmsg_orch(cmd1)
						// fmt.Println("RET CMD1 =", ret1)
						ret2 := scp_sendmsg_orch(cmd2)
						fmt.Println("RET CMD2 =", ret2)
						ret3 := scp_sendmsg_orch(cmd3)
						fmt.Println("RET CMD3 =", ret3)
						conn.Write([]byte(scp_ack))
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
				case scp_par_withdraw:
					vol, err := strconv.Atoi(subparams[1])
					checkErr(err)
					if err == nil {
						ibc[ind].Withdraw = uint32(vol)
					}
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
						//cmd1 = "CMD/" + biodev + "/MOD/" + pumpdev[1:] + ",3/END"
						cmd2 = "CMD/" + ibcdev + "/PUT/" + pumpdev + ",1/END"
						//cmd3 = "CMD/" + bioscr + "/PUT/S270,1/END"
					} else {
						//cmd1 = "CMD/" + biodev + "/MOD/" + pumpdev[1:] + ",3/END"
						cmd2 = "CMD/" + ibcdev + "/PUT/" + pumpdev + ",0/END"
						//cmd3 = "CMD/" + bioscr + "/PUT/S270,0/END"
					}
					// ret1 := scp_sendmsg_orch(cmd1)
					// fmt.Println("RET CMD1 =", ret1)
					ret2 := scp_sendmsg_orch(cmd2)
					fmt.Println("RET CMD2 =", ret2)
					// ret3 := scp_sendmsg_orch(cmd3)
					// fmt.Println("RET CMD3 =", ret3)
					conn.Write([]byte(scp_ack))

				case scp_dev_valve:
					var cmd2 string
					value_valve, err := strconv.Atoi(subparams[1])
					checkErr(err)
					value_status, err := strconv.Atoi(subparams[2])
					checkErr(err)
					//fmt.Println(value_valve, value_status)
					if (value_valve >= 0) && (value_valve < bio_max_valves) {
						ibc[ind].Valvs[value_valve] = value_status

						ibcdev := ibc_cfg[ibcid].Deviceaddr
						valvaddr := ibc_cfg[ibcid].Valv_devs[value_valve]
						//ibcscr := bio_cfg[ibcid].Screenaddr
						if value_status == 1 {
							//cmd1 = "CMD/" + biodev + "/MOD/" + pumpdev[1:] + ",3/END"
							cmd2 = "CMD/" + ibcdev + "/PUT/" + valvaddr + ",1/END"
							//cmd3 = "CMD/" + bioscr + "/PUT/S270,1/END"
						} else if value_status == 0 {
							//cmd1 = "CMD/" + biodev + "/MOD/" + pumpdev[1:] + ",3/END"
							cmd2 = "CMD/" + ibcdev + "/PUT/" + valvaddr + ",0/END"
							//cmd3 = "CMD/" + bioscr + "/PUT/S270,0/END"
						}
						// ret1 := scp_sendmsg_orch(cmd1)
						// fmt.Println("RET CMD1 =", ret1)
						ret2 := scp_sendmsg_orch(cmd2)
						fmt.Println("RET CMD2 =", ret2)
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
					//ibcscr := bio_cfg[ibcid].Screenaddr
					if value {
						//cmd1 = "CMD/" + biodev + "/MOD/" + pumpdev[1:] + ",3/END"
						cmd2 = "CMD/" + totemdev + "/PUT/" + pumpdev + ",1/END"
						//cmd3 = "CMD/" + bioscr + "/PUT/S270,1/END"
					} else {
						//cmd1 = "CMD/" + biodev + "/MOD/" + pumpdev[1:] + ",3/END"
						cmd2 = "CMD/" + totemdev + "/PUT/" + pumpdev + ",0/END"
						//cmd3 = "CMD/" + bioscr + "/PUT/S270,0/END"
					}
					// ret1 := scp_sendmsg_orch(cmd1)
					// fmt.Println("RET CMD1 =", ret1)
					ret2 := scp_sendmsg_orch(cmd2)
					fmt.Println("RET CMD2 =", ret2)
					// ret3 := scp_sendmsg_orch(cmd3)
					// fmt.Println("RET CMD3 =", ret3)
					conn.Write([]byte(scp_ack))

				case scp_dev_valve:
					var cmd2 string
					value_valve, err := strconv.Atoi(subparams[1])
					checkErr(err)
					value_status, err := strconv.Atoi(subparams[2])
					checkErr(err)
					//fmt.Println(value_valve, value_status)
					if (value_valve >= 0) && (value_valve < bio_max_valves) {
						totem[ind].Valvs[value_valve] = value_status

						totemdev := totem_cfg[totemid].Deviceaddr
						valvaddr := totem_cfg[totemid].Valv_devs[value_valve]
						//ibcscr := bio_cfg[ibcid].Screenaddr
						if value_status == 1 {
							//cmd1 = "CMD/" + biodev + "/MOD/" + pumpdev[1:] + ",3/END"
							cmd2 = "CMD/" + totemdev + "/PUT/" + valvaddr + ",1/END"
							//cmd3 = "CMD/" + bioscr + "/PUT/S270,1/END"
						} else if value_status == 0 {
							//cmd1 = "CMD/" + biodev + "/MOD/" + pumpdev[1:] + ",3/END"
							cmd2 = "CMD/" + totemdev + "/PUT/" + valvaddr + ",0/END"
							//cmd3 = "CMD/" + bioscr + "/PUT/S270,0/END"
						}
						// ret1 := scp_sendmsg_orch(cmd1)
						// fmt.Println("RET CMD1 =", ret1)
						ret2 := scp_sendmsg_orch(cmd2)
						fmt.Println("RET CMD2 =", ret2)
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
				//ibcscr := bio_cfg[ibcid].Screenaddr
				if value {
					//cmd1 = "CMD/" + biodev + "/MOD/" + pumpdev[1:] + ",3/END"
					cmd2 = "CMD/" + devaddr + "/PUT/" + devport + ",1/END"
					//cmd3 = "CMD/" + bioscr + "/PUT/S270,1/END"
				} else {
					//cmd1 = "CMD/" + biodev + "/MOD/" + pumpdev[1:] + ",3/END"
					cmd2 = "CMD/" + devaddr + "/PUT/" + devport + ",0/END"
					//cmd3 = "CMD/" + bioscr + "/PUT/S270,0/END"
				}
				// ret1 := scp_sendmsg_orch(cmd1)
				// fmt.Println("RET CMD1 =", ret1)
				ret2 := scp_sendmsg_orch(cmd2)
				fmt.Println("RET CMD2 =", ret2)
				// ret3 := scp_sendmsg_orch(cmd3)
				// fmt.Println("RET CMD3 =", ret3)
				conn.Write([]byte(scp_ack))

			case scp_dev_valve:
				var cmd2 string
				value_valve, err := strconv.Atoi(subparams[1])
				checkErr(err)
				value_status, err := strconv.Atoi(subparams[2])
				checkErr(err)
				//fmt.Println(value_valve, value_status)
				if (value_valve >= 0) && (value_valve < 9) {
					biofabrica.Valvs[value_valve] = value_status
					devid := fmt.Sprintf("VBF%02d", value_valve+1)
					devaddr := biofabrica_cfg[devid].Deviceaddr
					devport := biofabrica_cfg[devid].Deviceport
					//ibcscr := bio_cfg[ibcid].Screenaddr
					if value_status == 1 {
						//cmd1 = "CMD/" + biodev + "/MOD/" + pumpdev[1:] + ",3/END"
						cmd2 = "CMD/" + devaddr + "/PUT/" + devport + ",1/END"
						//cmd3 = "CMD/" + bioscr + "/PUT/S270,1/END"
					} else if value_status == 0 {
						//cmd1 = "CMD/" + biodev + "/MOD/" + pumpdev[1:] + ",3/END"
						cmd2 = "CMD/" + devaddr + "/PUT/" + devport + ",0/END"
						//cmd3 = "CMD/" + bioscr + "/PUT/S270,0/END"
					}
					// ret1 := scp_sendmsg_orch(cmd1)
					// fmt.Println("RET CMD1 =", ret1)
					fmt.Println("biofabrica valvula", cmd2)
					ret2 := scp_sendmsg_orch(cmd2)
					fmt.Println("RET CMD2 =", ret2)
					conn.Write([]byte(scp_ack))
				}
			default:
				conn.Write([]byte(scp_err))
			}

		default:
			conn.Write([]byte(scp_err))
		}

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
	fmt.Println("IBC cfg", ibc_cfg)
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
	fmt.Println("BIO cfg", bio_cfg)
	fmt.Println("IBC cfg", ibc_cfg)
	fmt.Println("TOTEM cfg", totem_cfg)
	fmt.Println("Biofabrica cfg", biofabrica_cfg)
	go scp_setup_devices()

	go scp_get_alldata()
	scp_master_ipc()
}
