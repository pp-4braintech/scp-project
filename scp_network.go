package main

import (
	"fmt"
	"net/http"

	"github.com/rs/cors"
)

func main() {

	net192 = test_file("/etc/scpd/scp_net192.flag")
	if net192 {
		fmt.Println("WARN:  EXECUTANDO EM NET192\n\n\n")
		execpath = "/home/paulo/scp-project/"
	} else {
		execpath = "/home/scpadm/scp-project/"
	}

	mux := http.NewServeMux()
	cors := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{
			http.MethodPut,
			http.MethodPost,
			http.MethodGet,
		},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: false,
	})

	mux.HandleFunc("/", main_handler)

	mux.HandleFunc("/ibc_view", ibc_view)

	mux.HandleFunc("/totem_view", totem_view)

	mux.HandleFunc("/biofabrica_view", biofabrica_view)

	mux.HandleFunc("/simulator", biofactory_sim)

	mux.HandleFunc("/config", set_config)

	mux.HandleFunc("/wdpanel", withdraw_panel)

	handler := cors.Handler(mux)

	http.ListenAndServe(":7000", handler)
}
