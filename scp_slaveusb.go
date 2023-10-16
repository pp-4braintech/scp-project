package main

import (
	"encoding/hex"
	"fmt"
	"log"
	"net"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const firm_version = "SLAVEUSB_V1.0.0"

const scp_slave_udpport = "9009"
const scp_slave_tcpport = "9119"
const scp_master_port = "7007"
const scp_retries = 10
const scp_keepalive_time = 10

const scp_ack = "ACK"
const scp_join = "JOIN"
const scp_end = "END"
const scp_put = "PUT"
const scp_get = "GET"
const scp_mod = "MOD"
const scp_inv = "INV"
const scp_ierr = "IERR"
const scp_ping = "PING"

var scp_slave_addr string
var scp_connected = false

var scp_last_IP net.Addr

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

func get_eth_mac() string {
	eth_mac := ""
	cmdpath, _ := filepath.Abs("/sbin/ifconfig")
	cmd := exec.Command(cmdpath, "enp3s0") // "| grep 'inet ' | awk '{ print $2}'")
	// cmd := exec.Command(cmdpath)
	// cmd.Dir = "/sbin/"
	output, err := cmd.CombinedOutput()
	if err != nil {
		checkErr(err)
	} else {
		out_str := string(output)
		p := strings.Index(out_str, "ether")
		if p >= 0 {
			ret := scp_splitparam(out_str[p:], " ")
			if len(ret) > 1 {
				eth_mac = ret[1]
			}
		}

	}
	return eth_mac
}

func mac2scpaddr(mac string) string {
	macs := scp_splitparam(mac, ":")
	if len(macs) != 6 {
		fmt.Println("ERROR MAC2SCPADDR: Endereco mac invalido", mac)
		return ""
	}
	macstr := ""
	for _, m := range macs {
		macstr += m
	}
	machex, err := hex.DecodeString(macstr)
	if err != nil {
		fmt.Println("ERROR MAC2SCPADDR: Nao foi possivel converter mac para hex", mac, macstr)
		return ""
	}
	machash := machex[0] ^ machex[1] ^ machex[2]
	scpaddr := fmt.Sprintf("%02x:%02x%02x%02x", machash, machex[3], machex[4], machex[5])
	// for _, b := range machex {
	// 	fmt.Printf("%x ", b)
	// }
	// fmt.Println(scpaddr)
	return string(scpaddr)
}

func scp_sendudp(con net.PacketConn, scp_dest_addr net.Addr, scp_message []byte, scp_len int, wait_ack bool) bool {
	var buf [2048]byte

	has_ack := !wait_ack
	for ntries := 0; ntries < scp_retries; ntries++ {
		_, err := con.WriteTo(scp_message, scp_dest_addr)
		checkErr(err)
		for reps := 0; !has_ack && reps < scp_retries; reps++ {
			packetSize, addr, err := con.ReadFrom(buf[0:])
			if err != nil {
				checkErr(err)
			} else {
				scp_last_IP = addr
				bufstr := fmt.Sprintf("s", buf)
				fmt.Println("Remote addr =", scp_last_IP, " dados =", bufstr, " size =", packetSize)
				if strings.Contains(bufstr, scp_ack) {
					has_ack = true
				}
			}
		}
		if has_ack {
			break
		}
		time.Sleep(1 * time.Second)
	}
	return has_ack
}

func scp_join_server() bool {
	con, err := net.ListenPacket("udp4", ":"+scp_slave_udpport)
	if err != nil {
		fmt.Println("ERROR SCP JOIN SERVER: Nao foi possivel criar o socket UDP")
		checkErr(err)
		return false
	}
	defer con.Close()
	broadcast_addr, err := net.ResolveUDPAddr("udp4", "10.0.0.255:"+scp_master_port)
	if err != nil {
		fmt.Println("ERROR SCP JOIN SERVER: Nao foi possivel determinar endereco do Master")
		checkErr(err)
		return false
	}
	// hasjoin := false
	// ntries := 0
	join_msg := fmt.Sprintf("%s/%s/%u/%s/%s", scp_join, scp_slave_addr, scp_slave_tcpport, firm_version, scp_end)
	scp_sendudp(con, broadcast_addr, []byte(join_msg), len(join_msg), true)
	return true
}

func scp_setup_slave() {
	mymac := get_eth_mac()
	fmt.Println("DEBUG SCP SETUP SLAVE: My MAC =", mymac)
	tmpaddr := mac2scpaddr(mymac)
	if len(tmpaddr) > 0 {
		scp_slave_addr = tmpaddr
		fmt.Println("DEBUG SCP SETUP SLAVE: Endereco SCP =", scp_slave_addr)
	}
}

func main() {
	scp_setup_slave()
}
