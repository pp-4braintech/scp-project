package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

const scp_ack = "ACK"
const scp_err = "ERR"
const scp_get = "GET"
const scp_put = "PUT"
const scp_dev_pump = "PUMP"
const scp_dev_aero = "AERO"
const scp_dev_valve = "VALVE"
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
}

var bio = []Bioreact{
	{"BIOR001", "55:3A7D80", "66:FA12F4", bio_producting, "Bacillus Subtilis", 100, 10, false, true, [8]int{1, 1, 0, 0, 0, 0, 0, 0}, 28, 7, [2]int{2, 5}, [2]int{25, 17}, [2]int{48, 0}},
	{"BIOR002", "2F:A2CFF4", "66:FA12F4", bio_cip, "Bacillus Megaterium", 200, 5, true, false, [8]int{0, 0, 1, 0, 0, 1, 0, 1}, 26, 7, [2]int{1, 1}, [2]int{0, 5}, [2]int{0, 30}},
	{"BIOR003", "42:A8AB4", "66:FA12F4", bio_loading, "Bacillus Amyloliquefaciens", 1000, 3, false, false, [8]int{0, 0, 0, 1, 0, 0, 1, 0}, 28, 7, [2]int{1, 1}, [2]int{0, 10}, [2]int{0, 30}},
	{"BIOR004", "8D:A8AB4", "66:FA12F4", bio_unloading, "Azospirilum brasiliense", 500, 5, true, false, [8]int{0, 0, 0, 0, 1, 1, 0, 0}, 25, 7, [2]int{1, 1}, [2]int{0, 5}, [2]int{0, 15}},
	{"BIOR005", "55:3A7D80", "66:FA12F4", bio_done, "Tricoderma harzianum", 0, 10, false, false, [8]int{2, 0, 0, 0, 0, 0, 0, 0}, 28, 7, [2]int{5, 5}, [2]int{0, 0}, [2]int{72, 0}},
	{"BIOR006", "42:A8AB4", "66:FA12F4", bio_nonexist, "", 0, 0, false, false, [8]int{0, 0, 0, 0, 0, 0, 0, 0}, 0, 0, [2]int{0, 0}, [2]int{0, 0}, [2]int{0, 0}},
}

var ibc = []IBC{
	{"IBC01", bio_storing, "Bacillus Subtilis", 100, 1, false, [4]int{0, 0, 0, 0}, [2]int{24, 15}},
	{"IBC02", bio_storing, "Bacillus Megaterium", 200, 1, false, [4]int{0, 0, 0, 0}, [2]int{12, 5}},
	{"IBC03", bio_loading, "Bacillus Amyloliquefaciens", 1000, 3, false, [4]int{0, 0, 1, 0}, [2]int{0, 30}},
	{"IBC04", bio_unloading, "Azospirilum brasiliense", 500, 2, false, [4]int{0, 0, 0, 1}, [2]int{4, 50}},
	{"IBC05", bio_storing, "Tricoderma harzianum", 1000, 3, false, [4]int{0, 0, 0, 0}, [2]int{13, 17}},
	{"IBC06", bio_cip, "Tricoderma harzianum", 2000, 5, true, [4]int{0, 1, 0, 0}, [2]int{0, 5}},
	{"IBC07", bio_empty, "", 0, 0, false, [4]int{0, 0, 0, 0}, [2]int{0, 0}},
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
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
	for {
		for k, b := range bio {
			cmd1 := "CMD/" + b.Deviceaddr + "/GET/D14/END"
			ret1 := scp_sendmsg_orch(cmd1)
			params := scp_splitparam(ret1, "/")
			if params[0] == scp_ack {
				if params[1] == "0" {
					bio[k].Pumpstatus = false
				} else {
					bio[k].Pumpstatus = true
				}
			}
			cmd2 := "CMD/" + b.Deviceaddr + "/GET/A7/END"
			ret2 := scp_sendmsg_orch(cmd2)
			params = scp_splitparam(ret2, "/")
			if params[0] == scp_ack {
				if params[1] == "0" {
					bio[k].Aerator = false
				} else {
					bio[k].Aerator = true
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
			ind := get_bio_index(params[2])
			if ind < 0 {
				conn.Write([]byte(scp_err))
			} else {
				subparams := scp_splitparam(params[3], ",")
				scp_device := subparams[0]
				fmt.Println("subparams=", subparams)
				switch scp_device {
				case scp_dev_pump:
					var cmd1, cmd2 string
					value, err := strconv.ParseBool(subparams[1])
					checkErr(err)
					bio[ind].Pumpstatus = value
					if value {
						cmd1 = "CMD/" + bio[ind].Deviceaddr + "/PUT/D14,1/END"
						cmd2 = "CMD/" + bio[ind].Screenaddr + "/PUT/S270,1/END"
					} else {
						cmd1 = "CMD/" + bio[ind].Deviceaddr + "/PUT/D14,0/END"
						cmd2 = "CMD/" + bio[ind].Screenaddr + "/PUT/S270,0/END"
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
					var cmd1, cmd2 string
					value_valve, err := strconv.Atoi(subparams[1])
					checkErr(err)
					value_status, err := strconv.Atoi(subparams[2])
					checkErr(err)
					//fmt.Println(value_valve, value_status)
					if (value_valve >= 0) && (value_valve < bio_max_valves) {
						bio[ind].Valvs[value_valve] = value_status
						conn.Write([]byte(scp_ack))
						var valve_str1, valve_str2 string
						if value_valve < 7 {
							valve_str1 = fmt.Sprintf("%d", value_valve+7)
						} else {
							valve_str1 = "16"
						}
						valve_str2 = fmt.Sprintf("%d", value_valve+201)
						if value_status > 0 {
							cmd1 = "CMD/" + bio[ind].Deviceaddr + "/PUT/D" + valve_str1 + ",1/END"
							cmd2 = "CMD/" + bio[ind].Screenaddr + "/PUT/S" + valve_str2 + ",1/END"
						} else {
							cmd1 = "CMD/" + bio[ind].Deviceaddr + "/PUT/D" + valve_str1 + ",0/END"
							cmd2 = "CMD/" + bio[ind].Screenaddr + "/PUT/S" + valve_str2 + ",0/END"
						}
						ret1 := scp_sendmsg_orch(cmd1)
						fmt.Println("RET CMD1 =", ret1)
						ret2 := scp_sendmsg_orch(cmd2)
						fmt.Println("RET CMD2 =", ret2)
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
	go scp_get_alldata()
	scp_master_ipc()
}
