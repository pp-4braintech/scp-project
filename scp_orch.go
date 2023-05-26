package main

import (
	"fmt"
	"log"
	"net"
	"strings"
	"time"
)

const scp_escape = '\x1b'
const scp_ack = "ACK"
const scp_cmd = "CMD"
const scp_err = "ERR"
const scp_join = "JOIN"
const scp_ping = "PING"
const scp_destroy = "DESTROY"
const scp_state_JOIN0 = 10
const scp_state_JOIN1 = 11
const scp_state_TCP0 = 20
const scp_state_TCPFAIL = 29
const scp_max_len = 512
const scp_keepalive_time = 10
const scp_timeout_ms = 2000
const scp_buff_size = 512
tmp = append(tmp, y)

type scp_slave_map struct {
	slave_udp_addr  string
	slave_tcp_addr  string
	slave_scp_addr  string
	slave_scp_state uint8
	go_chan         chan string
}

var scp_slaves map[string]scp_slave_map

func checkErr(err error) {
	if err != nil {
		log.Println("[SCP ERROR]", err)
	}
}

func scp_splitparam(param string) []string {
	scp_data := strings.Split(param, "/")
	if len(scp_data) < 1 {
		return nil
	}
	return scp_data
}

func scp_sendtcp(scp_con net.Conn, scp_message string, wait_ack bool) (string, error) {
	if len(scp_message) > scp_buff_size {
		scp_message = scp_message[0 : scp_buff_size-2]
	}
	msg := scp_message + string(scp_escape)
	_, err := scp_con.Write([]byte(msg))
	checkErr(err)
	if err != nil {
		return scp_err, err
	}
	if wait_ack {
		err = scp_con.SetReadDeadline(time.Now().Add(time.Duration(scp_timeout_ms) * time.Millisecond))
		checkErr(err)
		var ret = make([]byte, scp_max_len)
		_, err := scp_con.Read(ret)
		checkErr(err)
		//fmt.Println("tcp recebido", string(ret))
		return string(ret), err
	}
	return scp_ack, err
}

func scp_master_tcp_client(scp_slave *scp_slave_map) {

	slave_data := *scp_slave
	slave_addr := slave_data.slave_tcp_addr
	fmt.Println("TCP con ", slave_addr)
	slave_tcp_con, err := net.Dial("tcp", slave_addr)
	checkErr(err)

	if err == nil {
		fmt.Println("Conexao TCP estabelecida com slave")
		slave_data.slave_scp_state = scp_state_TCP0
		*scp_slave = slave_data
		slave_data.go_chan <- scp_ack
	} else {
		fmt.Println("ERRO Conexao TCP com slave")
		checkErr(err)
		slave_data.slave_scp_state = scp_state_TCPFAIL
		*scp_slave = slave_data
		slave_data.go_chan <- scp_err
		return
	}

	defer slave_tcp_con.Close()

	begin_time := time.Now().Unix()
	for {
		select {
		case chan_msg := <-slave_data.go_chan:
			if chan_msg == scp_destroy {
				fmt.Println("TCP destroy recebido")
				slave_data.go_chan <- scp_ack
				return
			}
			fmt.Println("TCP Enviando", chan_msg, "para", slave_data.slave_scp_addr)
			ret, err := scp_sendtcp(slave_tcp_con, chan_msg, true)
			if err == nil {
				slave_data.go_chan <- ret
			} else {
				slave_data.go_chan <- scp_err
			}
			begin_time = time.Now().Unix()
		default:
			current_time := time.Now().Unix()
			elapsed_seconds := current_time - begin_time
			if elapsed_seconds > scp_keepalive_time {
				fmt.Println("Enviando PING para", scp_slave.slave_scp_addr)
				ret, err := scp_sendtcp(slave_tcp_con, scp_ping, true)
				fmt.Println("ret =", ret)
				if err != nil {
					fmt.Println("ERR ao tratar PING")
				}
				begin_time = current_time
			}
			time.Sleep(50 * time.Millisecond)
		}
	}
}

func scp_process_udp(msg []byte, net_addr net.Addr) {

	params := scp_splitparam(string(msg[0:p_size]))
	scp_command := params[0]
	fmt.Println("SCP comm", scp_command, " / SCP data", params[1:])

	switch scp_command {
	case scp_join:
		scp_msg_data := params[1]
		fmt.Println("JOIN recebido de", scp_msg_data, len(scp_msg_data))
		slave_data, ok := scp_slaves[scp_msg_data]
		if ok {
			fmt.Println("JOIN recebido de slave já existente", slave_data)
			_, err = con.WriteTo([]byte(scp_err), net_addr)
			checkErr(err)
			fmt.Println("Destruindo SCP TCP")
			slave_data.go_chan <- scp_destroy
			fmt.Println("...saindo do channel")
			time.Sleep(60 * time.Millisecond)
			select {
			case ret := <-slave_data.go_chan:
				if ret == scp_ack {
					fmt.Println("fechando chain")
					close(slave_data.go_chan)
					fmt.Println("deletando dados na tabela")
					delete(scp_slaves, scp_msg_data)
				} else {
					fmt.Println("Falha ao destruir SCP TCP")
				}
			default:
				fmt.Println("SCP TCP nao respondeu")
			}

		} else {
			udp_addr := net_addr.String()
			tmp_addr := strings.Split(udp_addr, ":")
			if len(params) >= 3 {
				tcp_addr := tmp_addr[0] + ":" + params[2]
				process_chan := make(chan string)
				scp_slaves[scp_msg_data] = scp_slave_map{udp_addr, tcp_addr, scp_msg_data, scp_state_JOIN0, process_chan}
				fmt.Println("Slave inserido na tabela =", scp_slaves[scp_msg_data])
				_, err = con.WriteTo([]byte(scp_ack), net_addr)
				checkErr(err)
			} else {
				fmt.Println("JOIN com parametros incorretos", params)
			}
		}

	case scp_ack:
		scp_msg_data := params[1]
		fmt.Println("ACK recebido de", scp_msg_data, len(scp_msg_data))
		slave_data, ok := scp_slaves[scp_msg_data]
		if !ok {
			fmt.Println("Slave fora da tabela", scp_msg_data, slave_data, ok)
			_, err = con.WriteTo([]byte(scp_err), net_addr)
			checkErr(err)
		} else {
			if slave_data.slave_scp_state == scp_state_JOIN0 {
				fmt.Println("JOIN confirmado")
				slave_data.slave_scp_state = scp_state_JOIN1
				scp_slaves[scp_msg_data] = slave_data
				_, err = con.WriteTo([]byte(scp_ack), net_addr)
				checkErr(err)
				go scp_master_tcp_client(&slave_data)
				ret := <-slave_data.go_chan
				if ret == scp_err {
					fmt.Println("ERRO ao criar conexao TCP com cliente")
					slave_data.slave_scp_state = scp_state_JOIN0
				}
				scp_slaves[scp_msg_data] = slave_data
				//slave_data.go_chan <- "PUT/5/1/END"
				//slave_data.go_chan <- "GET/1/END"
			}
		}

	case scp_cmd:
		scp_msg_slaveaddr := params[1]
		slave_data, ok := scp_slaves[scp_msg_slaveaddr]
		if !ok {
			fmt.Println("CMD para Slave fora da tabela", scp_msg_slaveaddr, slave_data, ok)
			_, err = con.WriteTo([]byte(scp_err), net_addr)
			checkErr(err)
		} else {
			cmd := params[2]
			tam := len(cmd)
			for _, v := range params[3:] {
				cmd += "/" + v
				tam += 1 + len(v)
				//fmt.Println(i, v, len(v))
			}
			scp_msg_slavecmd := cmd[0:tam]
			fmt.Println("CMD para", scp_msg_slaveaddr, scp_msg_slavecmd, "len", len(scp_msg_slavecmd))
			slave_data.go_chan <- scp_msg_slavecmd
			ret := <-slave_data.go_chan
			fmt.Println("CMD ret=", ret)
			_, err = con.WriteTo([]byte(ret), net_addr)
			checkErr(err)
		}

	}
}

func scp_master_udp() {

	con, err := net.ListenPacket("udp", ":7007")
	checkErr(err)
	defer con.Close()

	for {
		reply := make([]byte, 1024)
		p_size, net_addr, err := con.ReadFrom(reply)
		checkErr(err)

		fmt.Println("msg recebida:", string(reply))
		fmt.Println("origem:", net_addr)

		go scp_process_udp(reply, net_addr)
		
		// params := scp_splitparam(string(reply[0:p_size]))
		// scp_command := params[0]
		// fmt.Println("SCP comm", scp_command, " / SCP data", params[1:])

		// switch scp_command {
		// case scp_join:
		// 	scp_msg_data := params[1]
		// 	fmt.Println("JOIN recebido de", scp_msg_data, len(scp_msg_data))
		// 	slave_data, ok := scp_slaves[scp_msg_data]
		// 	if ok {
		// 		fmt.Println("JOIN recebido de slave já existente", slave_data)
		// 		_, err = con.WriteTo([]byte(scp_err), net_addr)
		// 		checkErr(err)
		// 		fmt.Println("Destruindo SCP TCP")
		// 		slave_data.go_chan <- scp_destroy
		// 		fmt.Println("...saindo do channel")
		// 		time.Sleep(60 * time.Millisecond)
		// 		select {
		// 		case ret := <-slave_data.go_chan:
		// 			if ret == scp_ack {
		// 				fmt.Println("fechando chain")
		// 				close(slave_data.go_chan)
		// 				fmt.Println("deletando dados na tabela")
		// 				delete(scp_slaves, scp_msg_data)
		// 			} else {
		// 				fmt.Println("Falha ao destruir SCP TCP")
		// 			}
		// 		default:
		// 			fmt.Println("SCP TCP nao respondeu")
		// 		}

		// 	} else {
		// 		udp_addr := net_addr.String()
		// 		tmp_addr := strings.Split(udp_addr, ":")
		// 		if len(params) >= 3 {
		// 			tcp_addr := tmp_addr[0] + ":" + params[2]
		// 			process_chan := make(chan string)
		// 			scp_slaves[scp_msg_data] = scp_slave_map{udp_addr, tcp_addr, scp_msg_data, scp_state_JOIN0, process_chan}
		// 			fmt.Println("Slave inserido na tabela =", scp_slaves[scp_msg_data])
		// 			_, err = con.WriteTo([]byte(scp_ack), net_addr)
		// 			checkErr(err)
		// 		} else {
		// 			fmt.Println("JOIN com parametros incorretos", params)
		// 		}
		// 	}

		// case scp_ack:
		// 	scp_msg_data := params[1]
		// 	fmt.Println("ACK recebido de", scp_msg_data, len(scp_msg_data))
		// 	slave_data, ok := scp_slaves[scp_msg_data]
		// 	if !ok {
		// 		fmt.Println("Slave fora da tabela", scp_msg_data, slave_data, ok)
		// 		_, err = con.WriteTo([]byte(scp_err), net_addr)
		// 		checkErr(err)
		// 	} else {
		// 		if slave_data.slave_scp_state == scp_state_JOIN0 {
		// 			fmt.Println("JOIN confirmado")
		// 			slave_data.slave_scp_state = scp_state_JOIN1
		// 			scp_slaves[scp_msg_data] = slave_data
		// 			_, err = con.WriteTo([]byte(scp_ack), net_addr)
		// 			checkErr(err)
		// 			go scp_master_tcp_client(&slave_data)
		// 			ret := <-slave_data.go_chan
		// 			if ret == scp_err {
		// 				fmt.Println("ERRO ao criar conexao TCP com cliente")
		// 				slave_data.slave_scp_state = scp_state_JOIN0
		// 			}
		// 			scp_slaves[scp_msg_data] = slave_data
		// 			//slave_data.go_chan <- "PUT/5/1/END"
		// 			//slave_data.go_chan <- "GET/1/END"
		// 		}
		// 	}

		// case scp_cmd:
		// 	scp_msg_slaveaddr := params[1]
		// 	slave_data, ok := scp_slaves[scp_msg_slaveaddr]
		// 	if !ok {
		// 		fmt.Println("CMD para Slave fora da tabela", scp_msg_slaveaddr, slave_data, ok)
		// 		_, err = con.WriteTo([]byte(scp_err), net_addr)
		// 		checkErr(err)
		// 	} else {
		// 		cmd := params[2]
		// 		tam := len(cmd)
		// 		for _, v := range params[3:] {
		// 			cmd += "/" + v
		// 			tam += 1 + len(v)
		// 			//fmt.Println(i, v, len(v))
		// 		}
		// 		scp_msg_slavecmd := cmd[0:tam]
		// 		fmt.Println("CMD para", scp_msg_slaveaddr, scp_msg_slavecmd, "len", len(scp_msg_slavecmd))
		// 		slave_data.go_chan <- scp_msg_slavecmd
		// 		ret := <-slave_data.go_chan
		// 		fmt.Println("CMD ret=", ret)
		// 		_, err = con.WriteTo([]byte(ret), net_addr)
		// 		checkErr(err)
		// 	}

		// }
		fmt.Println()
		fmt.Println("scp slave", scp_slaves)
		fmt.Println()
	}
}

func main() {
	fmt.Println("SCP Orchestrator iniciando")
	scp_slaves = make(map[string]scp_slave_map)
	scp_master_udp()

}
