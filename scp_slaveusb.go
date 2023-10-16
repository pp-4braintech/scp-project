package main

import (
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

func scp_setup_slave() {
	fmt.Println("My mac =", get_eth_mac())
}

func main() {
	scp_setup_slave()
}
