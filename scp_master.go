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
const scp_ibc = "IBC"
const scp_orch_addr = ":7007"
const scp_ipc_name = "/tmp/scp_master.sock"
const scp_refreshwait = 5000

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
const bio_max_valves = 8

type Bioreact struct {
	BioreactorID string
	Deviceaddr   string
	Screenaddr   string
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
	Withdraw     uint32
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

type Bioreact_cfg struct {
	IBCID      string
	Deviceaddr string
	Screenaddr string
	Maxvolume  uint32
	Pump_dev   string
	Aero_dev   string
	Peris_dev  [5]string
	Valv_devs  [8]string
	Vol_devs   [2]string
	PH_dev     string
	Temp_dev   string
	Levelhigh  string
	Levellow   string
	Emergency  string
}
type IBC_cfg struct {
	IBCID      string
	Deviceaddr string
	Screenaddr string
	Maxvolume  uint32
	Pump_dev   string
	Valv_devs  [4]string
	Vol_devs   [2]string
}

var ibc_cfg map[string]IBC_cfg
var bio_cfg map[string]Bioreact_cfg

var bio = []Bioreact{
	{"BIOR001", "55:3A7D80", "66:FA12F4", bio_producting, "Bacillus Subtilis", 100, 10, false, true, [8]int{1, 1, 0, 0, 0, 0, 0, 0}, 28, 7, [2]int{2, 5}, [2]int{25, 17}, [2]int{48, 0}, 0},
	{"BIOR002", "2F:A2CFF4", "66:FA12F4", bio_cip, "Bacillus Megaterium", 200, 5, true, false, [8]int{0, 0, 1, 0, 0, 1, 0, 1}, 26, 7, [2]int{1, 1}, [2]int{0, 5}, [2]int{0, 30}, 0},
	{"BIOR003", "42:A8AB4", "66:FA12F4", bio_loading, "Bacillus Amyloliquefaciens", 1000, 3, false, false, [8]int{0, 0, 0, 1, 0, 0, 1, 0}, 28, 7, [2]int{1, 1}, [2]int{0, 10}, [2]int{0, 30}, 0},
	{"BIOR004", "8D:A8AB4", "66:FA12F4", bio_unloading, "Azospirilum brasiliense", 500, 5, true, false, [8]int{0, 0, 0, 0, 1, 1, 0, 0}, 25, 7, [2]int{1, 1}, [2]int{0, 5}, [2]int{0, 15}, 0},
	{"BIOR005", "55:3A7D80", "66:FA12F4", bio_done, "Tricoderma harzianum", 0, 10, false, false, [8]int{2, 0, 0, 0, 0, 0, 0, 0}, 28, 7, [2]int{5, 5}, [2]int{0, 0}, [2]int{72, 0}, 0},
	{"BIOR006", "42:A8AB4", "66:FA12F4", bio_nonexist, "", 0, 0, false, false, [8]int{0, 0, 0, 0, 0, 0, 0, 0}, 0, 0, [2]int{0, 0}, [2]int{0, 0}, [2]int{0, 0}, 0},
}

var ibc = []IBC{
	{"IBC01", bio_storing, "Bacillus Subtilis", 100, 1, false, [4]int{0, 0, 0, 0}, [2]int{24, 15}, 0},
	{"IBC02", bio_storing, "Bacillus Megaterium", 4000, 10, false, [4]int{0, 0, 0, 0}, [2]int{12, 5}, 0},
	{"IBC03", bio_loading, "Bacillus Amyloliquefaciens", 1000, 3, false, [4]int{0, 0, 1, 0}, [2]int{0, 30}, 0},
	{"IBC04", bio_unloading, "Azospirilum brasiliense", 500, 2, false, [4]int{0, 0, 0, 1}, [2]int{4, 50}, 0},
	{"IBC05", bio_storing, "Tricoderma harzianum", 1000, 3, false, [4]int{0, 0, 0, 0}, [2]int{13, 17}, 0},
	{"IBC06", bio_cip, "Tricoderma harzianum", 250, 1, true, [4]int{0, 1, 0, 0}, [2]int{0, 5}, 0},
	{"IBC07", bio_empty, "", 0, 0, false, [4]int{0, 0, 0, 0}, [2]int{0, 0}, 0},
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
		ibc_cfg[id] = IBC_cfg{id, dev_addr, screen_addr, uint32(voltot), pumpdev,
			[4]string{vdev1, vdev2, vdev3, vdev4}, [2]string{voldev1, voldev2}}
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
		if !strings.Contains(r[0], "#") && len(r) == 26 {
			id := r[0]
			dev_addr := r[1]
			screen_addr := r[2]
			voltot, _ := strconv.Atoi(strings.Replace(r[3], " ", "", -1))
			pumpdev := r[4]
			aerodev := r[5]
			perdev1 := r[6]
			perdev2 := r[7]
			perdev3 := r[8]
			perdev4 := r[9]
			perdev5 := r[10]
			vdev1 := r[11]
			vdev2 := r[12]
			vdev3 := r[13]
			vdev4 := r[14]
			vdev5 := r[15]
			vdev6 := r[16]
			vdev7 := r[17]
			vdev8 := r[18]
			voldev1 := r[19]
			voldev2 := r[20]
			phdev := r[21]
			tempdev := r[22]
			lhigh := r[23]
			llow := r[24]
			emerg := r[25]

			bio_cfg[id] = Bioreact_cfg{id, dev_addr, screen_addr, uint32(voltot), pumpdev, aerodev,
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

func scp_sendmsg_orch(cmd string) string {

	if demo {
		return scp_ack
	}
	fmt.Println("TO ORCH:", cmd)
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
	fmt.Println("Recebido:", string(ret))
	return string(ret)
}

func scp_get_alldata() {
	if demo {
		return
	}
	for {
		for k, b := range bio {
			if len(bio_cfg[b.BioreactorID].Deviceaddr) > 0 {

				bioaddr := bio_cfg[b.BioreactorID].Deviceaddr
				tempdev := bio_cfg[b.BioreactorID].Temp_dev
				phdev := bio_cfg[b.BioreactorID].PH_dev

				cmd1 := "CMD/" + bioaddr + "/GET/" + tempdev + "/END"
				ret1 := scp_sendmsg_orch(cmd1)
				params := scp_splitparam(ret1, "/")
				if params[0] == scp_ack {
					tempint, _ := strconv.Atoi(params[1])
					bio[k].Temperature = float32(tempint)
				}
				cmd2 := "CMD/" + bioaddr + "/GET/" + phdev + "/END"
				ret2 := scp_sendmsg_orch(cmd2)
				params = scp_splitparam(ret2, "/")
				if params[0] == scp_ack {
					phint, _ := strconv.Atoi(params[1])
					bio[k].PH = float32(phint)
				}
			}
		}
		time.Sleep(scp_refreshwait * time.Millisecond)
	}
}

func scp_process_conn(conn net.Conn) {
	buf := make([]byte, 512)
	n, err := conn.Read(buf)
	if err != nil {
		checkErr(err)
		return
	}
	fmt.Printf("msg: %s\n", buf[:n])
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

		default:
			conn.Write([]byte(scp_err))
		}
	case scp_put:
		scp_object := params[1]
		switch scp_object {
		case scp_bioreactor:
			fmt.Println("obj=", scp_object)
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
					var cmd1, cmd2 string
					value, err := strconv.ParseBool(subparams[1])
					checkErr(err)
					biodev := bio_cfg[bioid].Deviceaddr
					bioscr := bio_cfg[bioid].Screenaddr
					pumpdev := bio_cfg[bioid].Pump_dev
					bio[ind].Pumpstatus = value
					if value {
						cmd1 = "CMD/" + biodev + "/PUT/" + pumpdev + ",1/END"
						cmd2 = "CMD/" + bioscr + "/PUT/S270,1/END"
					} else {
						cmd1 = "CMD/" + biodev + "/PUT/" + pumpdev + ",0/END"
						cmd2 = "CMD/" + bioscr + "/PUT/S270,0/END"
					}
					ret1 := scp_sendmsg_orch(cmd1)
					fmt.Println("RET CMD1 =", ret1)
					ret2 := scp_sendmsg_orch(cmd2)
					fmt.Println("RET CMD2 =", ret2)
					conn.Write([]byte(scp_ack))

				case scp_dev_aero:
					var cmd1, cmd2, cmd3 string
					value, err := strconv.ParseBool(subparams[1])
					checkErr(err)
					bio[ind].Aerator = value
					if value {
						cmd1 = "CMD/" + bio[ind].Deviceaddr + "/PUT/D6,1/END"
						cmd2 = "CMD/" + bio[ind].Screenaddr + "/PUT/S271,1/END"
						cmd3 = "CMD/" + bio[ind].Deviceaddr + "/PUT/A7,127/END"

					} else {
						cmd1 = "CMD/" + bio[ind].Deviceaddr + "/PUT/D6,0/END"
						cmd2 = "CMD/" + bio[ind].Screenaddr + "/PUT/S271,0/END"
						cmd3 = "CMD/" + bio[ind].Deviceaddr + "/PUT/A7,0/END"
					}
					ret1 := scp_sendmsg_orch(cmd1)
					fmt.Println("RET CMD1 =", ret1)
					ret2 := scp_sendmsg_orch(cmd2)
					fmt.Println("RET CMD2 =", ret2)
					ret3 := scp_sendmsg_orch(cmd3)
					fmt.Println("RET CMD3 =", ret3)
					conn.Write([]byte(scp_ack))

				case scp_dev_valve:
					var cmd1, cmd2, cmd3 string
					value_valve, err := strconv.Atoi(subparams[1])
					checkErr(err)
					value_status, err := strconv.Atoi(subparams[2])
					checkErr(err)
					//fmt.Println(value_valve, value_status)
					if (value_valve >= 0) && (value_valve < bio_max_valves) {
						bio[ind].Valvs[value_valve] = value_status
						conn.Write([]byte(scp_ack))
						// var valve_str1, valve_str2 string
						// if value_valve < 7 {
						// 	valve_str1 = fmt.Sprintf("%d", value_valve+7)
						// } else {
						// 	valve_str1 = "16"
						// }
						valve_str2 := fmt.Sprintf("%d", value_valve+201)
						biodev := bio_cfg[bioid].Deviceaddr
						bioscr := bio_cfg[bioid].Screenaddr
						valvaddr := bio_cfg[bioid].Valv_devs[value_valve]
						if value_status > 0 {
							cmd1 = "CMD/" + biodev + "/MOD/" + valvaddr[1:] + ",3/END"
							cmd2 = "CMD/" + biodev + "/PUT/" + valvaddr + ",1/END"
							cmd3 = "CMD/" + bioscr + "/PUT/S" + valve_str2 + ",1/END"
						} else {
							cmd1 = "CMD/" + biodev + "/MOD/" + valvaddr[1:] + ",3/END"
							cmd2 = "CMD/" + biodev + "/PUT/" + valvaddr + ",0/END"
							cmd3 = "CMD/" + bioscr + "/PUT/S" + valve_str2 + ",0/END"
						}
						ret1 := scp_sendmsg_orch(cmd1)
						fmt.Println("RET CMD1 =", ret1)
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
			ind := get_ibc_index(params[2])
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
					value, err := strconv.ParseBool(subparams[1])
					checkErr(err)
					ibc[ind].Pumpstatus = value
					conn.Write([]byte(scp_ack))

				case scp_dev_valve:
					value_valve, err := strconv.Atoi(subparams[1])
					checkErr(err)
					value_status, err := strconv.Atoi(subparams[2])
					checkErr(err)
					//fmt.Println(value_valve, value_status)
					if (value_valve >= 0) && (value_valve < bio_max_valves) {
						ibc[ind].Valvs[value_valve] = value_status
						conn.Write([]byte(scp_ack))
					}
				default:
					conn.Write([]byte(scp_err))
				}
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
	fmt.Println("BIO cfg", bio_cfg)
	go scp_get_alldata()
	scp_master_ipc()
}
