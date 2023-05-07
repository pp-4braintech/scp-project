package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/rs/cors"
)

const bio_nonexist = "NULL"
const bio_cip = "CIP"
const bio_loading = "CARREGANDO"
const bio_unloading = "ESVAZIANDO"
const bio_producting = "PRODUZINDO"
const bio_empty = "VAZIO"
const bio_done = "CONCLUIDO"
const bio_error = "ERRO"

type Bioreact struct {
	BioreactorID string
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
	Pumpstatus bool
}

func checkErr(err error) {
	if err != nil {
		log.Println("[SCP ERROR]", err)
	}
}

func ibc_view(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	ibc := []IBC{
		{"IBC01", bio_done, "Bacillus Subtilis", 100, false},
		{"IBC02", bio_done, "Bacillus Megaterium", 200, false},
		{"IBC03", bio_loading, "Bacillus Amyloliquefaciens", 1000, false},
		{"IBC04", bio_unloading, "Azospirilum brasiliense", 500, false},
		{"IBC05", bio_done, "Tricoderma harzianum", 1000, false},
		{"IBC06", bio_cip, "Tricoderma harzianum", 2000, true},
		{"IBC07", bio_empty, "", 0, false},
	}
	// fmt.Println("bio", ibc)
	fmt.Println("Request", r)
	fmt.Println()
	jsonStr, err := json.Marshal(ibc)
	os.Stdout.Write(jsonStr)
	checkErr(err)
	w.Write([]byte(jsonStr))
	fmt.Println()
	fmt.Println()
}

func bioreactor_view(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	bio := []Bioreact{
		{"BIOR001", bio_producting, "Bacillus Subtilis", 100, 10, false, true, [8]int{1, 1, 0, 0, 0, 0, 0, 0}, 28, 7, [2]int{2, 5}, [2]int{25, 17}, [2]int{48, 0}},
		{"BIOR002", bio_cip, "Bacillus Megaterium", 200, 5, false, false, [8]int{0, 0, 1, 0, 0, 1, 0, 1}, 26, 7, [2]int{1, 1}, [2]int{0, 5}, [2]int{0, 30}},
		{"BIOR003", bio_loading, "Bacillus Amyloliquefaciens", 1000, 3, false, false, [8]int{0, 0, 0, 1, 0, 0, 1, 0}, 28, 7, [2]int{1, 1}, [2]int{0, 10}, [2]int{0, 30}},
		{"BIOR004", bio_unloading, "Azospirilum brasiliense", 500, 5, true, false, [8]int{0, 0, 0, 0, 1, 1, 0, 0}, 25, 7, [2]int{1, 1}, [2]int{0, 5}, [2]int{0, 15}},
		{"BIOR005", bio_done, "Tricoderma harzianum", 0, 10, false, false, [8]int{0, 0, 0, 0, 0, 0, 0, 0}, 28, 7, [2]int{5, 5}, [2]int{0, 0}, [2]int{72, 0}},
		{"BIOR006", bio_nonexist, "", 0, 0, false, false, [8]int{0, 0, 0, 0, 0, 0, 0, 0}, 0, 0, [2]int{0, 0}, [2]int{0, 0}, [2]int{0, 0}},
	}
	// fmt.Println("bio", bio)
	fmt.Println("Request", r)
	bio_id := r.URL.Query().Get("Id")
	fmt.Println("bio_id =", bio_id)
	fmt.Println()
	jsonStr, err := json.Marshal(bio)
	bio_found := false
	if len(bio_id) > 0 {
		for _, v := range bio {
			if v.BioreactorID == bio_id {
				jsonStr, err = json.Marshal(v)
				bio_found = true
				break
			}
		}
		if !bio_found {
			jsonStr, err = json.Marshal("ERR Id not found")
		}
	}
	os.Stdout.Write(jsonStr)
	checkErr(err)
	w.Write([]byte(jsonStr))
	fmt.Println()
	fmt.Println()
}

func main() {

	mux := http.NewServeMux()
	cors := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{
			http.MethodPost,
			http.MethodGet,
		},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: false,
	})

	mux.HandleFunc("/bioreactor_view", bioreactor_view)

	mux.HandleFunc("/ibc_view", ibc_view)

	handler := cors.Handler(mux)

	http.ListenAndServe(":5000", handler)
}
