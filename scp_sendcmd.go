package main

import (
	"fmt"
	"log"
	"net"
	"os"
)

// const scp_ack = "ACK"
// const scp_err = "ERR"
// const scp_join = "JOIN"

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func scp_client_udp(cmd string) {

	con, err := net.Dial("udp", ":7007")
	checkErr(err)
	defer con.Close()

	_, err = con.Write([]byte(cmd))
	checkErr(err)
	fmt.Println("Enviado:", cmd, len(cmd))

	ret := make([]byte, 1024)
	_, err = con.Read(ret)
	checkErr(err)
	fmt.Println("Recebido:", string(ret))

}

func main() {

	if len(os.Args) != 2 {
		fmt.Println("Use: scp_sendcmd message")
		os.Exit(1)
	}
	scp_client_udp(os.Args[1])
}
