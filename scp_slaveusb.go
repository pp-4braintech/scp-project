package main

import (
	"encoding/hex"
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
	"strings"
)

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
	for _, b := range machex {
		fmt.Printf("%x ", b)
	}
	fmt.Println(scpaddr)
	return string(scpaddr)
}

func scp_setup_slave() {
	mymac := get_eth_mac()
	fmt.Println("My mac =", mymac)
	mac2scpaddr(mymac)
}

func main() {
	scp_setup_slave()
}
