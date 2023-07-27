package main

import (
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
	"time"
)

const scp_escape = '\x1b'
const scp_ack = "ACK"
const scp_cmd = "CMD"
const scp_err = "ERR"
const scp_die = "DIE"
const scp_timeout = "TOUT"
const scp_join = "JOIN"
const scp_ping = "PING"
const scp_pong = "PONG"
const scp_destroy = "DESTROY"
const scp_status = "STATUS"
const scp_state_JOIN0 = 10
const scp_state_JOIN1 = 11
const scp_state_TCP0 = 20
const scp_state_TCPFAIL = 29
const scp_max_len = 512
const scp_keepalive_time = 6
const scp_timeout_ms = 3000
const scp_buff_size = 512
const scp_max_err = 7

type scp_slave_map struct {
	slave_udp_addr  string
	slave_tcp_addr  string
	slave_scp_addr  string
	slave_scp_state uint8
	go_chan         chan string
	slave_errors    uint8
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

func get_status() string {
	stat := fmt.Sprintf("STATUS/SLAVES,%d/", len(scp_slaves))
	for _, s := range scp_slaves {
		stat += fmt.Sprintf("%s,%s,%d/", s.slave_scp_addr, s.slave_tcp_addr, s.slave_scp_state)
	}
	stat += "END"
	return stat
}

func print_scp_slave() {
	fmt.Println("--- SCP Slaves Map")
	for k, s := range scp_slaves {
		fmt.Println(k, s)
	}
	fmt.Println()
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

func scp_send_ping(scp_slave *scp_slave_map, slave_con net.Conn) {
	slave_data := *scp_slave
	slave_addr := slave_data.slave_tcp_addr
	fmt.Println("Enviando PING para", slave_addr)
	ret, err := scp_sendtcp(slave_con, scp_ping, true)
	// if err != nil || !strings.Contains(ret, scp_pong) {
	if err != nil {
		scp_slave.slave_errors++
		fmt.Println(scp_slave.slave_scp_addr, "--->>>  ERR ao tratar PING", ret)
	} else {
		fmt.Println(scp_slave.slave_scp_addr, " PING ret =", ret)
		scp_slave.slave_errors = 0
	}
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
			fmt.Println("TCP CLIENT CHANNEL: ", chan_msg)
			if chan_msg == scp_destroy {
				fmt.Println("TCP destroy recebido")
				//slave_data.go_chan <- scp_ack
				return
			}
			if scp_slave.slave_errors > scp_max_err {
				fmt.Println("----->>>> TCP CLIENT com excesso de erros")
				slave_data.go_chan <- scp_die
			} else {
				fmt.Println("TCP Enviando", chan_msg, "para", slave_data.slave_scp_addr)
				ret, err := scp_sendtcp(slave_tcp_con, chan_msg, true)
				// if len(slave_data.go_chan) == 0 {
				// 	fmt.Println("*** ERRO NO CHANNEL")
				// 	return
				// }
				if err == nil {
					slave_data.go_chan <- ret
					scp_slave.slave_errors = 0
				} else {
					scp_slave.slave_errors++
					slave_data.go_chan <- scp_err
				}
				begin_time = time.Now().Unix()
			}
		default:
			current_time := time.Now().Unix()
			elapsed_seconds := current_time - begin_time
			if (elapsed_seconds > scp_keepalive_time) && (scp_slave.slave_errors < scp_max_err) {
				scp_send_ping(scp_slave, slave_tcp_con)
				begin_time = current_time
				// fmt.Println(scp_slave)
			}
		}
		time.Sleep(50 * time.Millisecond)
	}
}

func scp_change_status(scp_slave *scp_slave_map, new_status uint8) {
	scp_slave.slave_scp_state = new_status
}

func scp_process_udp(con net.PacketConn, msg []byte, p_size int, net_addr net.Addr, m *sync.Mutex) {
	var err error

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
			// _, err = con.WriteTo([]byte(scp_err), net_addr)             MUDEI POR CONTA DAS TELAS ENTRAR EM LOOP TENTANDO FAZER JOIN
			// checkErr(err)
			fmt.Println("Destruindo SCP TCP")

			go func() {
				slave_data.go_chan <- scp_destroy
				fmt.Println("destroy enviado com sucesso")
			}()
			time.Sleep(50 * time.Millisecond)
			// fmt.Println("fechando chain")
			// close(slave_data.go_chan)

			// fmt.Println("deletando dados na tabela")
			// delete(scp_slaves, scp_msg_data)

			// fmt.Println("...saindo do channel")
			// time.Sleep(100 * time.Millisecond)
			// select {
			// case ret := <-slave_data.go_chan:
			// 	if ret == scp_ack {
			// 		fmt.Println("SCP TCP destroy ACK")
			// 	} else {
			// 		fmt.Println("Falha ao destruir SCP TCP =", ret)
			// 	}
			// default:
			// 	fmt.Println("Não houve resposta do chan")
			// }
			// fmt.Println("Deletando  dados na tabela")
			//delete(scp_slaves, scp_msg_data)

		}

		udp_addr := net_addr.String()
		tmp_addr := strings.Split(udp_addr, ":")
		if len(params) >= 3 {
			tcp_addr := tmp_addr[0] + ":" + params[2]
			process_chan := make(chan string)
			m.Lock()
			scp_slaves[scp_msg_data] = scp_slave_map{slave_udp_addr: udp_addr, slave_tcp_addr: tcp_addr, slave_scp_addr: scp_msg_data,
				slave_scp_state: scp_state_JOIN0, go_chan: process_chan, slave_errors: 0}
			m.Unlock()
			fmt.Println("Slave inserido na tabela =", scp_slaves[scp_msg_data])
			_, err = con.WriteTo([]byte(scp_ack), net_addr)
			checkErr(err)
		} else {
			fmt.Println("JOIN com parametros incorretos", params)
			_, err = con.WriteTo([]byte(scp_err), net_addr)
			checkErr(err)
		}

	case scp_ack:
		scp_msg_data := params[1]
		fmt.Println("ACK recebido de", scp_msg_data, len(scp_msg_data))
		slave_data, ok := scp_slaves[scp_msg_data]
		if !ok {
			fmt.Println("ACK de Slave fora da tabela", scp_msg_data, slave_data, ok)
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
			} else {
				fmt.Println("\n\nACK UDP de dispositivo que não estava em JOIN0")
			}
		}

	case scp_status:
		scp_msg_data := params[1]
		fmt.Println("STATUS recebido de", scp_msg_data, len(scp_msg_data))
		stat := get_status()
		_, err = con.WriteTo([]byte(stat), net_addr)
		checkErr(err)

	case scp_cmd:
		scp_msg_slaveaddr := params[1]
		slave_data, ok := scp_slaves[scp_msg_slaveaddr]
		if !ok {
			fmt.Println("CMD para Slave fora da tabela", scp_msg_slaveaddr, slave_data, ok)
			_, err = con.WriteTo([]byte(scp_err), net_addr)
			checkErr(err)
		} else if slave_data.slave_scp_state == scp_state_TCPFAIL {
			fmt.Println("ERRO Cliente TCP")
			_, err = con.WriteTo([]byte(scp_err), net_addr)
			checkErr(err)
		} else if slave_data.slave_scp_state != scp_state_TCP0 {
			fmt.Println("ERRO Cliente NÃO COMPLETOU JOIN", slave_data)
			_, err = con.WriteTo([]byte(scp_err), net_addr)
			checkErr(err)
		} else if slave_data.slave_errors < scp_max_err {
			cmd := params[2]
			tam := len(cmd)
			for _, v := range params[3:] {
				cmd += "/" + v
				tam += 1 + len(v)
				//fmt.Println(i, v, len(v))
			}
			scp_msg_slavecmd := cmd[0:tam]
			fmt.Println("CMD para", scp_msg_slaveaddr, scp_msg_slavecmd, "len", len(scp_msg_slavecmd))
			//go func () {
			slave_data.go_chan <- scp_msg_slavecmd
			fmt.Println("CMD enviado para o CHANNEL")
			ret := <-slave_data.go_chan
			fmt.Println("CMD ret=", ret)
			_, err = con.WriteTo([]byte(ret), net_addr)
			checkErr(err)
			if ret == scp_die {
				slave_data.slave_scp_state = scp_state_TCPFAIL
				scp_slaves[scp_msg_slaveaddr] = slave_data
			}
			return

		}

		// select {
		// case slave_data.go_chan <- scp_msg_slavecmd:
		// 	fmt.Println("CMD enviado para o CHANNEL")
		// 	ret := <-slave_data.go_chan
		// 	fmt.Println("CMD ret=", ret)
		// 	_, err = con.WriteTo([]byte(ret), net_addr)
		// 	checkErr(err)
		// default:
		// 	fmt.Println("FALHA no envio de CMD no CHANNEL")
		// 	_, err = con.WriteTo([]byte(scp_err), net_addr)
		// 	checkErr(err)
		// }
		//}

	}
}

func scp_master_udp() {

	var m sync.Mutex
	con, err := net.ListenPacket("udp", ":7007")
	checkErr(err)
	defer con.Close()
	nslaves := 0

	for {
		reply := make([]byte, 1024)
		p_size, net_addr, err := con.ReadFrom(reply)
		checkErr(err)

		fmt.Println("msg recebida:", string(reply))
		fmt.Println("origem:", net_addr)

		go scp_process_udp(con, reply, p_size, net_addr, &m)

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
		if len(scp_slaves) != nslaves {
			print_scp_slave()
			nslaves = len(scp_slaves)
		}

	}
}

func main() {
	fmt.Println("SCP Orchestrator iniciando")
	scp_slaves = make(map[string]scp_slave_map)
	scp_master_udp()

}
