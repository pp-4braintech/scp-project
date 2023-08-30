package main

import (
	"encoding/csv"
	"encoding/json"

	// "filepath"
	"fmt"
	"io"
	"log"
	"math"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"gonum.org/v1/gonum/stat"
)

// antiga placa desenvase 05:C2DDBC

var demo = false
var devmode = false
var net192 = false
var testmode = false
var autowithdraw = false

const control_ph = false
const control_temp = false
const control_foam = true

const (
	scp_on  = 1
	scp_off = 0

	bio_escala = 1

	scp_ack     = "ACK"
	scp_err     = "ERR"
	scp_get     = "GET"
	scp_put     = "PUT"
	scp_run     = "RUN"
	scp_die     = "DIE"
	scp_sched   = "SCHED"
	scp_start   = "START"
	scp_status  = "STATUS"
	scp_stop    = "STOP"
	scp_pause   = "PAUSE"
	scp_fail    = "FAIL"
	scp_netfail = "NETFAIL"
	scp_ready   = "READY"

	scp_state_JOIN0   = 10
	scp_state_JOIN1   = 11
	scp_state_TCP0    = 20
	scp_state_TCPFAIL = 29
)

const scp_dev_pump = "PUMP"
const scp_dev_aero = "AERO"
const scp_dev_valve = "VALVE"
const scp_dev_water = "WATER"
const scp_dev_sprayball = "SPRAYBALL"
const scp_dev_peris = "PERIS"
const scp_dev_vol0 = "VOL0"
const scp_dev_volusom = "VOLUSOM"
const scp_dev_vollaser = "VOLLASER"
const scp_dev_volfluxo_out = "VOLFLUXO_OUT"
const scp_dev_volfluxo_in1 = "VOLFLUXO_IN1"

const scp_par_withdraw = "WITHDRAW"
const scp_par_out = "OUT"
const scp_par_time = "TIME"
const scp_par_volume = "VOLUME"
const scp_par_grow = "GROW"
const scp_par_cip = "CIP"
const scp_par_status = "STATUS"
const scp_par_step = "STEP"
const scp_par_maxstep = "MAXSTEP"
const scp_par_heater = "HEATER"
const scp_par_slaves = "SLAVES"
const scp_par_select = "SELECT"
const scp_par_inc = "INC"
const scp_par_dec = "DEC"
const scp_par_start = "START"
const scp_par_stop = "STOP"
const scp_par_ph4 = "PH4"
const scp_par_ph7 = "PH7"
const scp_par_ph10 = "PH10"
const scp_par_calibrate = "CALIBRATE"
const scp_par_save = "SAVE"
const scp_par_restart = "RESTART"
const scp_par_testmode = "TESTMODE"
const scp_par_getconfig = "GETCONFIG"
const scp_par_deviceaddr = "DEVICEADDR"
const scp_par_screenaddr = "SCREENADDR"
const scp_par_linewash = "LINEWASH"
const scp_par_linecip = "LINECIP"
const scp_par_circulate = "CIRCULATE"
const scp_par_totaltime = "TOTALTIME"
const scp_par_manydraw = "MANYDRAW"
const scp_par_manyout = "MANYOUT"

const scp_job_org = "ORG"
const scp_job_on = "ON"
const scp_job_set = "SET"
const scp_job_wait = "WAIT"
const scp_job_ask = "ASK"
const scp_job_off = "OFF"
const scp_job_run = "RUN"
const scp_job_stop = "STOP"
const scp_job_done = "DONE"
const scp_job_commit = "COMMIT"

const scp_msg_cloro = "CLORO"
const scp_msg_meio = "MEIO"
const scp_msg_inoculo = "INOCULO"
const scp_msg_meio_inoculo = "MEIO-INOCULO"

const scp_bioreactor = "BIOREACTOR"
const scp_biofabrica = "BIOFABRICA"
const scp_totem = "TOTEM"
const scp_ibc = "IBC"
const scp_wdpanel = "WDPANEL"
const scp_config = "CONFIG"
const scp_out = "OUT"
const scp_drop = "DROP"
const scp_clean = "CLEAN"
const scp_donothing = "NOTHING"
const scp_orch_addr = ":7007"
const scp_ipc_name = "/tmp/scp_master.sock"

const scp_refreshwait = 50
const scp_refresstatus = 15
const scp_refresscreens = 10
const scp_refreshsleep = 100
const scp_timeout_ms = 2500
const scp_schedwait = 500
const scp_clockwait = 60000
const scp_timetosave = 45
const scp_checksetup = 60
const scp_mustupdate_bio = 30
const scp_mustupdate_ibc = 45

const scp_timewaitvalvs = 15000
const scp_timephwait = 10000 // Tempo que o ajuste de PH e aplicado durante o cultivo
const scp_timetempwait = 3000
const scp_timewaitbeforeph = 10000
const scp_timegrowwait = 30000
const scp_maxtimewithdraw = 1800 // separar nas funcoes do JOB
const scp_timelinecip = 20       // em segundos
const scp_timeoutdefault = 60

const bio_deltatemp = 0.1 // variacao de temperatura maximo em percentual
const bio_deltaph = 0.3   // variacao de ph maximo em valor absoluto

const bio_withdrawstep = 50

const bio_diametro = 1530  // em mm   era 1430
const bio_v1_zero = 1483.0 // em mm
const bio_v2_zero = 1502.0 // em mm
const ibc_v1_zero = 2652.0 // em mm   2647
const ibc_v2_zero = 2652.0 // em mm

const flow_ratio = 0.03445

const bio_emptying_rate = 55.0 / 100.0

// const scp_join = "JOIN"
const data_filename = "dumpdata"

const pingmax = 3
const timetocheck = 30

const bio_nonexist = "NULL"
const bio_die = "DIE"
const bio_cip = "CIP"
const bio_pause = "PAUSADO"
const bio_wait = "AGUARDANDO"
const bio_starting = "INICIANDO"
const bio_loading = "CARREGANDO"
const bio_unloading = "DESENVASE"
const bio_producting = "PRODUZINDO"
const bio_empty = "VAZIO"
const bio_done = "CONCLUIDO"
const bio_storing = "ARMAZENANDO"
const bio_error = "ERRO"
const bio_ready = "PRONTO"
const bio_water = "AGUA"
const bio_update = "ATUALIZANDO"
const bio_circulate = "CIRCULANDO"
const bio_max_valves = 8
const bio_max_msg = 50
const bioreactor_max_msg = 7
const bio_max_foam = 4

const line_13 = "1_3"
const line_14 = "1_4"
const line_23 = "2_3"
const line_24 = "2_4"

const TEMPMAX = 120

type Organism struct {
	Code       string
	Orgname    string
	Lifetime   int
	Prodvol    int
	Cultmedium string
	Timetotal  int
	Aero       [3]int
	PH         [3]string
}

type Bioreact struct {
	BioreactorID string
	Status       string
	OrgCode      string
	Organism     string
	Vol0         int
	Vol1         int32
	Vol2         int32
	VolInOut     float64
	Volume       uint32
	Level        uint8
	Pumpstatus   bool
	Aerator      bool
	AeroRatio    int
	Valvs        [8]int
	Perist       [5]int
	Heater       bool
	Temperature  float32
	TempMax      float32
	PH           float32
	Step         [2]int
	Timeleft     [2]int
	Timetotal    [2]int
	Withdraw     uint32
	OutID        string
	Queue        []string
	UndoQueue    []string
	RedoQueue    []string
	MustOffQueue []string
	Vol_zero     [2]float32
	LastStatus   string
	MustStop     bool
	MustPause    bool
	ShowVol      bool
	Messages     []string
	PHref        [3]float64
	RegresPH     [2]float64
}

type Bioreact_ETL struct {
	BioreactorID string
	Status       string
	OrgCode      string
	Organism     string
	Vol0         int
	Vol1         int32
	Vol2         int32
	Volume       uint32
	Level        uint8
	Pumpstatus   bool
	Aerator      bool
	Valvs        [8]int
	Perist       [5]int
	Heater       bool
	Temperature  float32
	TempMax      float32
	PH           float32
	Step         [2]int
	Timeleft     [2]int
	Timetotal    [2]int
	Withdraw     uint32
	OutID        string
	Vol_zero     [2]float32
	LastStatus   string
	MustStop     bool
	MustPause    bool
	ShowVol      bool
	Messages     []string
	PHref        [3]float64
	RegresPH     [2]float64
}

type IBC struct {
	IBCID        string
	Status       string
	OrgCode      string
	Organism     string
	Vol0         int
	Vol1         int32
	Vol2         int32
	Volume       uint32
	Level        uint8
	Pumpstatus   bool
	Valvs        [4]int
	Step         [2]int
	Timetotal    [2]int
	Withdraw     uint32
	OutID        string
	Vol_zero     [2]float32
	MustStop     bool
	MustPause    bool
	Selected     bool
	Queue        []string
	UndoQueue    []string
	RedoQueue    []string
	MustOffQueue []string
	LastStatus   string
	ShowVol      bool
	VolumeOut    uint32
}

type IBC_ETL struct {
	IBCID      string
	Status     string
	OrgCode    string
	Organism   string
	Vol0       int
	Vol1       int32
	Vol2       int32
	Volume     uint32
	Level      uint8
	Pumpstatus bool
	Valvs      [4]int
	Step       [2]int
	Timetotal  [2]int
	Withdraw   uint32
	OutID      string
	Vol_zero   [2]float32
	MustStop   bool
	MustPause  bool
	Selected   bool
	LastStatus string
	ShowVol    bool
	VolumeOut  uint32
}

type Totem struct {
	TotemID    string
	Status     string
	Pumpstatus bool
	Valvs      [2]int
	Perist     [4]int
}

type Biofabrica struct {
	BiofabricaID string
	Valvs        [9]int
	Pumpwithdraw bool
	Messages     []string
	Status       string
	VolumeOut    float64
	VolOutPart   float64
	LastCountOut uint32
	VolumeIn1    float64
	VolIn1Part   float64
	LastCountIn1 uint32
	TestMode     bool
	TechMode     bool
	Useflowin    bool
}

type Path struct {
	FromID    string
	ToID      string
	Cleantime int
	Path      string
}

type Bioreact_cfg struct {
	BioreactorID string
	Deviceaddr   string
	Screenaddr   string
	Maxvolume    uint32
	Pump_dev     string
	Aero_dev     string
	Aero_rele    string
	Peris_dev    [5]string
	Valv_devs    [8]string
	Vol_devs     [2]string
	PH_dev       string
	Temp_dev     string
	Levelhigh    string
	Levellow     string
	Emergency    string
	Heater       string
}

type IBC_cfg struct {
	IBCID      string
	Deviceaddr string
	Screenaddr string
	Maxvolume  uint32
	Pump_dev   string
	Valv_devs  [4]string
	Vol_devs   [2]string
	Levellow   string
}

type Totem_cfg struct {
	TotemID    string
	Deviceaddr string
	Pumpdev    string
	Peris_dev  [4]string
	Valv_devs  [2]string
}

type Biofabrica_cfg struct {
	DeviceID   string
	Deviceaddr string
	Deviceport string
}

type Scheditem struct {
	Bioid   string
	Seq     int
	OrgCode string
}

type DevAddrData struct {
	DevAddr string
	DevType string
	DevID   string
}

var execpath string
var localconfig_path string
var mainrouter string

var finishedsetup = false
var schedrunning = false
var devsrunning = false

var ibc_cfg map[string]IBC_cfg
var bio_cfg map[string]Bioreact_cfg
var totem_cfg map[string]Totem_cfg
var biofabrica_cfg map[string]Biofabrica_cfg
var paths map[string]Path
var valvs map[string]int
var organs map[string]Organism
var addrs_type map[string]DevAddrData
var schedule []Scheditem
var recipe []string
var cipbio []string
var cipibc []string

var mainmutex sync.Mutex
var withdrawmutex sync.Mutex

var withdrawrunning = false

var bio = []Bioreact{
	{"BIOR01", bio_update, "", "", 0, 0, 0, 0, 1000, 0, false, false, 0, [8]int{0, 0, 0, 0, 0, 0, 0, 0}, [5]int{0, 0, 0, 0, 0}, false, 0, 0, 0, [2]int{2, 5}, [2]int{25, 17}, [2]int{48, 0}, 0, "OUT", []string{}, []string{}, []string{}, []string{}, [2]float32{0, 0}, "", false, false, true, []string{}, [3]float64{0, 0, 0}, [2]float64{0, 0}},
	{"BIOR02", bio_update, "", "", 0, 0, 0, 0, 0, 0, false, false, 0, [8]int{0, 0, 0, 0, 0, 0, 0, 0}, [5]int{0, 0, 0, 0, 0}, false, 0, 0, 0, [2]int{1, 1}, [2]int{0, 5}, [2]int{0, 30}, 0, "OUT", []string{}, []string{}, []string{}, []string{}, [2]float32{0, 0}, "", false, false, true, []string{}, [3]float64{0, 0, 0}, [2]float64{0, 0}},
	{"BIOR03", bio_update, "", "", 0, 0, 0, 0, 0, 0, false, false, 0, [8]int{0, 0, 0, 0, 0, 0, 0, 0}, [5]int{0, 0, 0, 0, 0}, false, 0, 0, 0, [2]int{1, 1}, [2]int{0, 10}, [2]int{0, 30}, 0, "OUT", []string{}, []string{}, []string{}, []string{}, [2]float32{0, 0}, "", false, false, true, []string{}, [3]float64{0, 0, 0}, [2]float64{0, 0}},
	{"BIOR04", bio_update, "", "", 0, 0, 0, 0, 0, 0, false, false, 0, [8]int{0, 0, 0, 0, 0, 0, 0, 0}, [5]int{0, 0, 0, 0, 0}, false, 0, 0, 0, [2]int{1, 1}, [2]int{0, 5}, [2]int{0, 15}, 0, "OUT", []string{}, []string{}, []string{}, []string{}, [2]float32{0, 0}, "", false, false, true, []string{}, [3]float64{0, 0, 0}, [2]float64{0, 0}},
	{"BIOR05", bio_update, "", "", 0, 0, 0, 0, 0, 0, false, false, 0, [8]int{0, 0, 0, 0, 0, 0, 0, 0}, [5]int{0, 0, 0, 0, 0}, false, 0, 0, 0, [2]int{5, 5}, [2]int{0, 0}, [2]int{72, 0}, 0, "OUT", []string{}, []string{}, []string{}, []string{}, [2]float32{0, 0}, "", false, false, true, []string{}, [3]float64{0, 0, 0}, [2]float64{0, 0}},
	{"BIOR06", bio_update, "PA", "Priestia Aryabhattai", 0, 0, 0, 0, 1000, 5, false, false, 0, [8]int{0, 0, 0, 0, 0, 0, 0, 0}, [5]int{0, 0, 0, 0, 0}, false, 0, 0, 0, [2]int{0, 0}, [2]int{0, 0}, [2]int{0, 0}, 0, "OUT", []string{}, []string{}, []string{}, []string{}, [2]float32{0, 0}, "", false, false, true, []string{}, [3]float64{0, 0, 0}, [2]float64{0, 0}},
}

var ibc = []IBC{
	{"IBC01", bio_update, "", "", 0, 0, 0, 0, 0, false, [4]int{0, 0, 0, 0}, [2]int{0, 0}, [2]int{0, 0}, 0, "OUT", [2]float32{0, 0}, false, false, false, []string{}, []string{}, []string{}, []string{}, "", true, 0},
	{"IBC02", bio_update, "", "", 0, 0, 0, 0, 0, false, [4]int{0, 0, 0, 0}, [2]int{0, 0}, [2]int{0, 0}, 0, "OUT", [2]float32{0, 0}, false, false, false, []string{}, []string{}, []string{}, []string{}, "", true, 0},
	{"IBC03", bio_update, "", "Bacillus Amyloliquefaciens", 0, 0, 0, 1000, 2, false, [4]int{0, 0, 0, 0}, [2]int{1, 5}, [2]int{0, 0}, 0, "OUT", [2]float32{0, 0}, false, false, false, []string{}, []string{}, []string{}, []string{}, "", true, 0},
	{"IBC04", bio_update, "", "Azospirilum brasiliense", 0, 0, 0, 100, 1, false, [4]int{0, 0, 0, 0}, [2]int{4, 50}, [2]int{0, 0}, 0, "OUT", [2]float32{0, 0}, false, false, false, []string{}, []string{}, []string{}, []string{}, "", true, 0},
	{"IBC05", bio_update, "", "Tricoderma harzianum", 0, 0, 0, 100, 1, false, [4]int{0, 0, 0, 0}, [2]int{13, 17}, [2]int{0, 0}, 0, "OUT", [2]float32{0, 0}, false, false, false, []string{}, []string{}, []string{}, []string{}, "", true, 0},
	{"IBC06", bio_update, "", "Tricoderma harzianum", 0, 0, 0, 100, 1, false, [4]int{0, 0, 0, 0}, [2]int{0, 5}, [2]int{0, 0}, 0, "OUT", [2]float32{0, 0}, false, false, false, []string{}, []string{}, []string{}, []string{}, "", true, 0},
	{"IBC07", bio_update, "", "", 0, 0, 0, 100, 1, false, [4]int{0, 0, 0, 0}, [2]int{0, 0}, [2]int{0, 0}, 0, "OUT", [2]float32{0, 0}, false, false, false, []string{}, []string{}, []string{}, []string{}, "", true, 0},
}

var totem = []Totem{
	{"TOTEM01", bio_ready, false, [2]int{0, 0}, [4]int{0, 0, 0, 0}},
	{"TOTEM02", bio_ready, false, [2]int{0, 0}, [4]int{0, 0, 0, 0}},
}

var biofabrica = Biofabrica{
	"BIOFABRICA001", [9]int{0, 0, 0, 0, 0, 0, 0, 0, 0}, false, []string{}, scp_ready, 0, 0, 0, 0, 0, 0, false, false, true,
}

var biobak = bio // Salva status atual

func sumXYandXX(arrayX []float64, arrayY []float64, meanX float64, meanY float64) (float64, float64) {
	resultXX := 0.0
	resultXY := 0.0
	for x := 0; x < len(arrayX); x++ {
		for y := 0; y < len(arrayY); y++ {
			if x == y {
				resultXY += (arrayX[x] - meanX) * (arrayY[y] - meanY)
			}
		}
		resultXX += (arrayX[x] - meanX) * (arrayX[x] - meanX)
	}
	return resultXY, resultXX
}

func estimateB0B1(x []float64, y []float64) (float64, float64) {
	var meanX float64
	var meanY float64
	var sumXY float64
	var sumXX float64
	meanX = stat.Mean(x, nil) //mean of x
	meanY = stat.Mean(y, nil) //mean pf y
	sumXY, sumXX = sumXYandXX(x, y, meanX, meanY)
	// regression coefficients
	b1 := sumXY / sumXX    // b1 or slope
	b0 := meanY - b1*meanX // b0 or intercept
	return b0, b1
}

func calc_PH(x float64, b0 float64, b1 float64) float64 {
	ph := b0 + b1*x
	fmt.Println("DEBUG CALC PH: Equacao = ", b0, " + ", b1, "*", x, " = ", ph)
	return ph
}

func calc_mediana(x []float64) float64 {
	var mediana float64
	sort.Float64s(x)
	n := len(x)
	c := int(n / 2)
	if n >= 5 {
		s := x[c-1] + x[c] + x[c+1]
		mediana = s / 3.0
	} else {
		mediana = x[c]
	}
	return mediana
}

func checkErr(err error) {
	if err != nil {
		log.Println("[SCP ERROR]", err)
	}
}

func isin(list []string, element string) bool {
	for _, s := range list {
		if element == s {
			return true
		}
	}
	return false
}

func load_tasks_conf(filename string) []string {
	tasks := []string{}
	file, err := os.Open(filename)
	if err != nil {
		checkErr(err)
		return nil
	}
	defer file.Close()
	csvr := csv.NewReader(file)
	paths = make(map[string]Path, 0)
	for {
		r, err := csvr.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			if perr, ok := err.(*csv.ParseError); ok && perr.Err != csv.ErrFieldCount {
				checkErr(err)
				return nil
			}
		}
		// fmt.Println(r)
		if r[0][0] != '#' {
			str := ""
			for _, s := range r {
				str += s + ","
			}
			str += "END"
			tasks = append(tasks, str)
		}
	}
	return tasks
}

func load_organisms(filename string) int {
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
	organs = make(map[string]Organism, len(records))
	for k, r := range records {
		code := r[0]
		name := r[1]
		lifetime, _ := strconv.Atoi(strings.Replace(r[2], " ", "", -1))
		volume, _ := strconv.Atoi(strings.Replace(r[3], " ", "", -1))
		medium := strings.Replace(r[4], " ", "", -1)
		tottime, _ := strconv.Atoi(strings.Replace(r[5], " ", "", -1))
		aero1, _ := strconv.Atoi(strings.Replace(r[6], " ", "", -1))
		aero2, _ := strconv.Atoi(strings.Replace(r[7], " ", "", -1))
		aero3, _ := strconv.Atoi(strings.Replace(r[8], " ", "", -1))
		ph1 := strings.Replace(r[9], " ", "", -1)
		ph2 := strings.Replace(r[10], " ", "", -1)
		ph3 := strings.Replace(r[11], " ", "", -1)
		org := Organism{code, name, lifetime, volume, medium, tottime, [3]int{aero1, aero2, aero3}, [3]string{ph1, ph2, ph3}}
		organs[code] = org
		totalrecords = k
	}
	return totalrecords
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
		llow := r[11]
		ibc_cfg[id] = IBC_cfg{id, dev_addr, screen_addr, uint32(voltot), pumpdev,
			[4]string{vdev1, vdev2, vdev3, vdev4}, [2]string{voldev1, voldev2}, llow}
		totalrecords = k

		old, ok := addrs_type[dev_addr]
		if !ok {
			addrs_type[dev_addr] = DevAddrData{dev_addr, scp_ibc, id}
		} else {
			fmt.Println("ERROR LOAD IBCS CONF: ADDR", dev_addr, " já cadastrado na tabela de devices com tipo", old)
		}
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
		if !strings.Contains(r[0], "#") && len(r) == 28 {
			id := r[0]
			dev_addr := r[1]
			screen_addr := r[2]
			voltot, _ := strconv.Atoi(strings.Replace(r[3], " ", "", -1))
			pumpdev := r[4]
			aerodev := r[5]
			aerorele := r[6]
			perdev1 := r[7]
			perdev2 := r[8]
			perdev3 := r[9]
			perdev4 := r[10]
			perdev5 := r[11]
			vdev1 := r[12]
			vdev2 := r[13]
			vdev3 := r[14]
			vdev4 := r[15]
			vdev5 := r[16]
			vdev6 := r[17]
			vdev7 := r[18]
			vdev8 := r[19]
			voldev1 := r[20]
			voldev2 := r[21]
			phdev := r[22]
			tempdev := r[23]
			lhigh := r[24]
			llow := r[25]
			emerg := r[26]
			heater := r[27]

			bio_cfg[id] = Bioreact_cfg{id, dev_addr, screen_addr, uint32(voltot), pumpdev, aerodev, aerorele,
				[5]string{perdev1, perdev2, perdev3, perdev4, perdev5},
				[8]string{vdev1, vdev2, vdev3, vdev4, vdev5, vdev6, vdev7, vdev8},
				[2]string{voldev1, voldev2}, phdev, tempdev, lhigh, llow, emerg, heater}

			old, ok := addrs_type[dev_addr]
			if !ok {
				addrs_type[dev_addr] = DevAddrData{dev_addr, scp_bioreactor, id}
			} else {
				fmt.Println("ERROR LOAD BIOS CONF: ADDR", dev_addr, " já cadastrado na tabela de devices com tipo", old)
			}
			totalrecords += 1
		} else if len(r) != 26 {
			fmt.Println("ERROR BIO CFG: numero de parametros invalido", r)
		}
	}
	return totalrecords
}

func save_bios_conf(filename string) int {
	filecsv, err := os.Create(filename)
	if err != nil {
		checkErr(err)
		return -1
	}
	defer filecsv.Close()
	n := 0
	csvwriter := csv.NewWriter(filecsv)
	for _, b := range bio_cfg {
		s := fmt.Sprintf("%s,%s,%s,%d,%s,%s,%s,", b.BioreactorID, b.Deviceaddr, b.Screenaddr, b.Maxvolume, b.Pump_dev, b.Aero_dev, b.Aero_rele)
		for _, p := range b.Peris_dev {
			s += p + ","
		}
		for _, v := range b.Valv_devs {
			s += v + ","
		}
		s += fmt.Sprintf("%s,%s,%s,%s,%s,%s,%s,%s", b.Vol_devs[0], b.Vol_devs[1], b.PH_dev, b.Temp_dev, b.Levelhigh, b.Levellow, b.Emergency, b.Heater)
		csvstr := scp_splitparam(s, ",")
		// fmt.Println("DEBUG SAVE", csvstr)
		err = csvwriter.Write(csvstr)
		if err != nil {
			checkErr(err)
		} else {
			n++
		}
	}
	csvwriter.Flush()
	return n
}

func load_totems_conf(filename string) int {
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
	totem_cfg = make(map[string]Totem_cfg, len(records))
	for k, r := range records {
		id := r[0]
		dev_addr := r[1]
		pumpdev := r[2]
		perdev1 := r[3]
		perdev2 := r[4]
		perdev3 := r[5]
		perdev4 := r[6]
		vdev1 := r[7]
		vdev2 := r[8]
		totem_cfg[id] = Totem_cfg{id, dev_addr, pumpdev,
			[4]string{perdev1, perdev2, perdev3, perdev4},
			[2]string{vdev1, vdev2}}
		totalrecords = k
		old, ok := addrs_type[dev_addr]
		if !ok {
			addrs_type[dev_addr] = DevAddrData{dev_addr, scp_totem, id}
		} else {
			fmt.Println("ERROR LOAD TOTEMS CONF: ADDR", dev_addr, " já cadastrado na tabela de devices com tipo", old)
		}
	}
	return totalrecords
}

func load_biofabrica_conf(filename string) int {
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
	biofabrica_cfg = make(map[string]Biofabrica_cfg, len(records))
	for k, r := range records {
		dev_id := r[0]
		dev_addr := r[1]
		dev_port := r[2]
		biofabrica_cfg[dev_id] = Biofabrica_cfg{dev_id, dev_addr, dev_port}
		totalrecords = k
		old, ok := addrs_type[dev_addr]
		if !ok {
			addrs_type[dev_addr] = DevAddrData{dev_addr, scp_biofabrica, dev_id}
		} else {
			fmt.Println("WARN LOAD BIOFABRICA CONF: ADDR", dev_addr, " já cadastrado na tabela de devices com tipo", old)
		}
	}
	return totalrecords
}

func load_paths_conf(filename string) int {
	var totalrecords int
	file, err := os.Open(filename)
	if err != nil {
		checkErr(err)
		return -1
	}
	defer file.Close()
	csvr := csv.NewReader(file)
	paths = make(map[string]Path, 0)
	totalrecords = 0
	for {
		r, err := csvr.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			if perr, ok := err.(*csv.ParseError); ok && perr.Err != csv.ErrFieldCount {
				checkErr(err)
				return -1
			}
		}
		// fmt.Println(r)
		if r[0][0] != '#' {
			from_id := r[0]
			to_id := r[1]
			clean_time, _ := strconv.Atoi(r[2])
			path_id := from_id + "-" + to_id
			pathstr := ""
			for i := 3; i < len(r); i++ {
				pathstr += r[i] + ","
			}
			pathstr += "END"
			paths[path_id] = Path{from_id, to_id, clean_time, pathstr}
			totalrecords++
		}
	}
	return totalrecords
}

func get_devtype_byaddr(dev_addr string) string {
	t, ok := addrs_type[dev_addr]
	if ok {
		return t.DevType
	}
	fmt.Println("ERROR GET DEVTYPE BYADDR: ADDR não EXISTE", dev_addr)
	return ""
}

func get_devid_byaddr(dev_addr string) string {
	t, ok := addrs_type[dev_addr]
	if ok {
		return t.DevID
	}
	fmt.Println("ERROR GET DEVTYPE BYADDR: ADDR não EXISTE", dev_addr)
	return ""
}

func set_valv_status(devtype string, devid string, valvid string, value int) bool {
	var devaddr, scraddr, valvaddr, valve_scrstr string
	id := devid + "/" + valvid
	valvs[id] = value
	switch devtype {
	case scp_donothing:
		return true
	case scp_bioreactor:
		ind := get_bio_index(devid)
		devaddr = bio_cfg[devid].Deviceaddr
		scraddr = bio_cfg[devid].Screenaddr
		if ind >= 0 {
			v, err := strconv.Atoi(valvid[1:])
			if err == nil {
				bio[ind].Valvs[v-1] = value
				valvaddr = bio_cfg[devid].Valv_devs[v-1]
				valve_scrstr = fmt.Sprintf("S%d", v+200)
			} else {
				fmt.Println("ERROR SET VAL: id da valvula nao inteiro", valvid)
				return false
			}
		} else {
			fmt.Println("ERROR SET VAL: BIORREATOR nao encontrado", devid)
			return false
		}
	case scp_ibc:
		ind := get_ibc_index(devid)
		devaddr = ibc_cfg[devid].Deviceaddr
		scraddr = ""
		if ind >= 0 {
			v, err := strconv.Atoi(valvid[1:])
			if err == nil {
				ibc[ind].Valvs[v-1] = value
				valvaddr = ibc_cfg[devid].Valv_devs[v-1]
			} else {
				fmt.Println("ERROR SET VAL: id da valvula nao inteiro", valvid)
				return false
			}
		} else {
			fmt.Println("ERROR SET VAL: IBC nao encontrado", devid)
			return false
		}
	case scp_totem:
		ind := get_totem_index(devid)
		devaddr = totem_cfg[devid].Deviceaddr
		scraddr = ""
		if ind >= 0 {
			v, err := strconv.Atoi(valvid[1:])
			if err == nil {
				totem[ind].Valvs[v-1] = value
				valvaddr = totem_cfg[devid].Valv_devs[v-1]
			} else {
				fmt.Println("ERROR SET VAL: id da valvula nao inteiro", valvid)
				return false
			}
		} else {
			fmt.Println("ERROR SET VAL: TOTEM nao encontrado", devid)
			return false
		}
	case scp_biofabrica:
		devaddr = biofabrica_cfg[valvid].Deviceaddr
		scraddr = ""
		valvaddr = biofabrica_cfg[valvid].Deviceport
		v, err := strconv.Atoi(valvid[3:])
		if err == nil {
			biofabrica.Valvs[v-1] = value
		} else {
			fmt.Println("ERROR SET VAL: BIOFABRICA - id da valvula nao inteiro", valvid)
			return false
		}
	}
	cmd1 := fmt.Sprintf("CMD/%s/PUT/%s,%d/END", devaddr, valvaddr, value)
	fmt.Println(cmd1)
	ret1 := scp_sendmsg_orch(cmd1)
	fmt.Println("RET CMD1 =", ret1)
	if !strings.Contains(ret1, scp_ack) && !devmode {
		fmt.Println("ERROR SET VAL: SEND MSG ORCH falhou", ret1)
		return false
	}
	if len(scraddr) > 0 {
		cmd2 := fmt.Sprintf("CMD/%s/PUT/%s,%d/END", scraddr, valve_scrstr, value)
		fmt.Println(cmd2)
		ret2 := scp_sendmsg_orch(cmd2)
		fmt.Println("RET CMD2 =", ret2)
	}
	return true
}

func get_valv_status(devid string, valvid string) int {
	id := devid + "/" + valvid
	value, ok := valvs[id]
	if ok {
		return value
	}
	return -1
}

func set_allvalvs_status() {
	for _, b := range bio {
		for n, v := range b.Valvs {
			valvid := fmt.Sprintf("V%d", n+1)
			set_valv_status(scp_donothing, b.BioreactorID, valvid, v)
		}
	}
	for _, i := range ibc {
		for n, v := range i.Valvs {
			valvid := fmt.Sprintf("V%d", n+1)
			set_valv_status(scp_donothing, i.IBCID, valvid, v)
		}
	}
	for _, t := range totem {
		for n, v := range t.Valvs {
			valvid := fmt.Sprintf("V%d", n+1)
			set_valv_status(scp_donothing, t.TotemID, valvid, v)
		}
	}
	// Biofrabrica
	for n, v := range biofabrica.Valvs {
		valvid := fmt.Sprintf("VBF%02d", n+1)
		set_valv_status(scp_donothing, "BIOFABRICA", valvid, v)
	}
	fmt.Println(valvs)
}

func save_all_data(filename string) int {
	buf1, _ := json.Marshal(bio)
	err1 := os.WriteFile(localconfig_path+filename+"_bio.json", []byte(buf1), 0644)
	checkErr(err1)
	buf2, _ := json.Marshal(ibc)
	err2 := os.WriteFile(localconfig_path+filename+"_ibc.json", []byte(buf2), 0644)
	checkErr(err2)
	buf3, _ := json.Marshal(totem)
	err3 := os.WriteFile(localconfig_path+filename+"_totem.json", []byte(buf3), 0644)
	checkErr(err3)
	buf4, _ := json.Marshal(biofabrica)
	err4 := os.WriteFile(localconfig_path+filename+"_biofabrica.json", []byte(buf4), 0644)
	checkErr(err4)
	buf5, _ := json.Marshal(schedule)
	err5 := os.WriteFile(localconfig_path+filename+"_schedule.json", []byte(buf5), 0644)
	checkErr(err5)

	return 0
}

func load_all_data(filename string) int {
	dat1, err1 := os.ReadFile(localconfig_path + filename + "_bio.json")
	checkErr(err1)
	if err1 == nil {
		json.Unmarshal([]byte(dat1), &bio)
		fmt.Println("-- bio data = ", bio)
	}

	dat2, err2 := os.ReadFile(localconfig_path + filename + "_ibc.json")
	checkErr(err2)
	if err2 == nil {
		json.Unmarshal([]byte(dat2), &ibc)
		fmt.Println("-- ibc data = ", ibc)
	}

	dat3, err3 := os.ReadFile(localconfig_path + filename + "_totem.json")
	checkErr(err3)
	if err3 == nil {
		json.Unmarshal([]byte(dat3), &totem)
		fmt.Println("-- totem data = ", totem)
	}

	dat4, err4 := os.ReadFile(localconfig_path + filename + "_biofabrica.json")
	checkErr(err4)
	if err4 == nil {
		json.Unmarshal([]byte(dat4), &biofabrica)
		fmt.Println("-- biofabrica data = ", biofabrica)
	}
	set_allvalvs_status()

	dat5, err5 := os.ReadFile(localconfig_path + filename + "_schedule.json")
	checkErr(err5)
	if err5 == nil {
		json.Unmarshal([]byte(dat5), &schedule)
		fmt.Println("-- schedule data = ", schedule)
	}

	return 0
}

func tcp_host_isalive(host string, tcpport string, timemax time.Duration) bool {
	timeout := time.Duration(timemax * time.Second)
	_, err := net.DialTimeout("tcp", host+":"+tcpport, timeout)
	if err != nil {
		checkErr(err)
		return false

	}
	return true
}

func scp_run_recovery() {
	fmt.Println("\n\nWARN RUN RECOVERY: Executando RECOVERY da Biofabrica")
	board_add_message("ERETORNANDO de EMERGENCIA")
	board_add_message("ANecessário aguardar 5 minutos até reestabelecimento dos equipamentos")
	time.Sleep(300 * time.Second)
	scp_setup_devices(true)
	for _, b := range ibc {
		if b.Status == bio_nonexist || b.Status == bio_error {
			board_add_message("AFavor checar IBC " + b.IBCID)
		}
	}
	for _, b := range bio {
		if b.Status != bio_nonexist && b.Status != bio_error {
			pause_device(scp_bioreactor, b.BioreactorID, false)
		} else {
			board_add_message("AFavor checar Biorreator " + b.BioreactorID)
		}
	}
	if !schedrunning {
		go scp_scheduler()
	}
}

func scp_emergency_pause() {
	fmt.Println("\n\nCRITICAL EMERGENCY PAUSE: Executando EMERGENCY PAUSE da Biofabrica")
	board_add_message("EPARADA de EMERGENCIA")
	for _, b := range bio {
		pause_device(scp_bioreactor, b.BioreactorID, true)
	}
}

func scp_check_network() {
	for {
		fmt.Println("DEBUG CHECK NETWORK: Testando comunicacao com MAINROUTER", mainrouter, pingmax)
		if !tcp_host_isalive(mainrouter, "80", pingmax) {
			if biofabrica.Status != scp_netfail {
				fmt.Println("FATAL CHECK NETWORK: Sem comunicacao com MAINROUTER", mainrouter)
				biofabrica.Status = scp_netfail
				save_all_data(data_filename)
				scp_emergency_pause()
				save_all_data(data_filename)
			}
		} else {
			fmt.Println("DEBUG CHECK NETWORK: OK comunicacao com MAINROUTER", mainrouter)
			if biofabrica.Status == scp_netfail {
				if finishedsetup {
					biofabrica.Status = scp_ready
					scp_run_recovery()
				}
			}
		}
		time.Sleep(timetocheck * time.Second)
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

func get_biobak_index(bio_id string) int {
	if len(bio_id) > 0 {
		for i, v := range biobak {
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

func get_totem_index(totem_id string) int {
	if len(totem_id) > 0 {
		for i, v := range totem {
			if v.TotemID == totem_id {
				return i
			}
		}
	}
	return -1
}

func get_scp_type(dev_id string) string {
	if strings.Contains(dev_id, "BIOFABRICA") {
		return scp_biofabrica
	} else if strings.Contains(dev_id, "BIOR") {
		return scp_bioreactor
	} else if strings.Contains(dev_id, "IBC") {
		return scp_ibc
	} else if strings.Contains(dev_id, "TOTEM") {
		return scp_totem
	} else if strings.Contains(dev_id, "OUT") {
		return scp_out
	} else if strings.Contains(dev_id, "DROP") {
		return scp_drop
	} else if strings.Contains(dev_id, "CLEAN") {
		return scp_clean
	}
	return scp_donothing
}

func scp_sendmsg_orch(cmd string) string {

	if demo || strings.Contains(cmd, "FF:FFFFFF") {
		return scp_ack
	}
	mainmutex.Lock()
	//fmt.Println("TO ORCH:", cmd)
	con, err := net.Dial("udp", scp_orch_addr)
	if err != nil {
		checkErr(err)
		mainmutex.Unlock()
		return scp_err
	}
	defer con.Close()
	defer mainmutex.Unlock()

	_, err = con.Write([]byte(cmd))
	if err != nil {
		checkErr(err)
		return scp_err
	}
	//fmt.Println("Enviado:", cmd, len(cmd))

	err = con.SetReadDeadline(time.Now().Add(time.Duration(scp_timeout_ms) * time.Millisecond))
	checkErr(err)
	ret := make([]byte, 1024)
	_, err = con.Read(ret)
	if err != nil {
		fmt.Println("ERROR SCP SENDMSG ORCH: Timeout ao receber resposta do orchestrator, requisicao feita:", cmd)
		checkErr(err)
		return scp_err
	}
	//fmt.Println("Recebido:", string(ret))
	return string(ret)
}

func board_add_message(m string) {
	n := len(biofabrica.Messages)
	stime := time.Now().Format("15:04")
	msg := fmt.Sprintf("%c%s [%s]", m[0], m[1:], stime)
	if n < bio_max_msg {
		biofabrica.Messages = append(biofabrica.Messages, msg)
	} else {
		biofabrica.Messages = append(biofabrica.Messages[2:], msg)
	}
}

func bio_add_message(bioid string, m string) {
	ind := get_bio_index(bioid)
	if ind < 0 {
		fmt.Println("ERROR BIO ADD MESSAGE: Biorreator nao existe", bioid)
		return
	}
	n := len(bio[ind].Messages)
	stime := time.Now().Format("15:04")
	msg := fmt.Sprintf("%c%s [%s]", m[0], m[1:], stime)
	if n < bioreactor_max_msg {
		bio[ind].Messages = append(bio[ind].Messages, msg)
	} else {
		bio[ind].Messages = append(bio[ind].Messages[2:], msg)
	}
}

func scp_setup_devices(mustall bool) {
	if demo {
		return
	}
	fmt.Println("\n\nDEBUG SETUP DEVICES: Configurando BIORREATORES")
	for _, b := range bio_cfg {
		ind := get_bio_index(b.BioreactorID)
		if len(b.Deviceaddr) > 0 && ind >= 0 {
			if mustall || bio[ind].Status == bio_nonexist || bio[ind].Status == bio_error {
				fmt.Println("DEBUG SETUP DEVICES: Device:", b.BioreactorID, "-", b.Deviceaddr)
				var cmd []string
				bioaddr := b.Deviceaddr
				cmd = make([]string, 0)
				cmd = append(cmd, "CMD/"+bioaddr+"/MOD/"+b.Pump_dev[1:]+",3/END")
				cmd = append(cmd, "CMD/"+bioaddr+"/MOD/"+b.Aero_dev[1:]+",3/END")
				cmd = append(cmd, "CMD/"+bioaddr+"/MOD/"+b.Aero_rele[1:]+",3/END")
				for i := 0; i < len(b.Peris_dev); i++ {
					cmd = append(cmd, "CMD/"+bioaddr+"/MOD/"+b.Peris_dev[i][1:]+",3/END")
				}
				for i := 0; i < len(b.Valv_devs); i++ {
					cmd = append(cmd, "CMD/"+bioaddr+"/MOD/"+b.Valv_devs[i][1:]+",3/END")
				}
				cmd = append(cmd, "CMD/"+bioaddr+"/MOD/"+b.Levelhigh[1:]+",0/END")
				cmd = append(cmd, "CMD/"+bioaddr+"/MOD/"+b.Levellow[1:]+",0/END")
				cmd = append(cmd, "CMD/"+bioaddr+"/MOD/"+b.Emergency[1:]+",0/END")
				cmd = append(cmd, "CMD/"+bioaddr+"/MOD/"+b.Heater[1:]+",3/END")
				cmd = append(cmd, "CMD/"+b.Screenaddr+"/PUT/S200,1/END")
				nerr := 0
				for k, c := range cmd {
					fmt.Print(k, "  ", c, " ")
					ret := scp_sendmsg_orch(c)
					if !strings.Contains(ret, scp_ack) {
						nerr++
					}
					fmt.Println(ret, nerr)
					if strings.Contains(ret, scp_die) {
						fmt.Println("ERROR SETUP DEVICES: BIORREATOR DIE", b.BioreactorID)
						break
					}
					time.Sleep(scp_refreshwait / 2 * time.Millisecond)
				}

				if bio[ind].Vol_zero[0] == 0 {
					bio[ind].Vol_zero[0] = bio_v1_zero
				}
				if bio[ind].Vol_zero[1] == 0 {
					bio[ind].Vol_zero[1] = bio_v2_zero
				}
				if nerr > 1 && !devmode {
					bio[ind].Status = bio_nonexist
					fmt.Println("ERROR SETUP DEVICES: BIORREATOR com erros", b.BioreactorID)
				} else if bio[ind].Status == bio_nonexist || bio[ind].Status == bio_error {
					if bio[ind].Volume == 0 {
						bio[ind].Status = bio_empty
					} else {
						bio[ind].Status = bio_ready
					}
				}
			}
		}
	}

	fmt.Println("\n\nDEBUG SETUP DEVICES: Configurando IBCs")
	for _, ib := range ibc_cfg {
		ind := get_ibc_index(ib.IBCID)
		if len(ib.Deviceaddr) > 0 && ind >= 0 {
			if mustall || ibc[ind].Status == bio_nonexist || ibc[ind].Status == bio_error {
				fmt.Println("DEBUG SETUP DEVICES: Device:", ib.IBCID, "-", ib.Deviceaddr)
				var cmd []string
				ibcaddr := ib.Deviceaddr
				cmd = make([]string, 0)
				cmd = append(cmd, "CMD/"+ibcaddr+"/MOD/"+ib.Pump_dev[1:]+",3/END")
				for i := 0; i < len(ib.Valv_devs); i++ {
					cmd = append(cmd, "CMD/"+ibcaddr+"/MOD/"+ib.Valv_devs[i][1:]+",3/END")
				}
				// cmd = append(cmd, "CMD/"+ibcaddr+"/MOD/"+b.Levelhigh[1:]+",1/END")
				cmd = append(cmd, "CMD/"+ibcaddr+"/MOD/"+ib.Levellow[1:]+",0/END")
				// cmd = append(cmd, "CMD/"+ibcaddr+"/MOD/"+b.Emergency[1:]+",1/END")

				nerr := 0
				for k, c := range cmd {
					fmt.Print(k, "  ", c, " ")
					ret := scp_sendmsg_orch(c)
					if !strings.Contains(ret, scp_ack) {
						nerr++
					}
					fmt.Println(ret, nerr)
					if strings.Contains(ret, scp_die) {
						fmt.Println("ERROR SETUP DEVICES: IBC DIE", ib.IBCID)
						break
					}
					time.Sleep(scp_refreshwait / 2 * time.Millisecond)
				}

				if ibc[ind].Vol_zero[0] == 0 {
					ibc[ind].Vol_zero[0] = ibc_v1_zero
				}
				if ibc[ind].Vol_zero[1] == 0 {
					ibc[ind].Vol_zero[1] = ibc_v2_zero
				}
				if nerr > 0 && !devmode {
					ibc[ind].Status = bio_nonexist
					fmt.Println("ERROR SETUP DEVICES: IBC com erros", ib.IBCID)
				} else if ibc[ind].Status == bio_nonexist || ibc[ind].Status == bio_error {
					if ibc[ind].Volume == 0 {
						ibc[ind].Status = bio_empty
					} else {
						ibc[ind].Status = bio_ready
					}
				}
			}
		}
	}

	fmt.Println("\n\nDEBUG SETUP DEVICES: Configurando TOTEMs")
	for _, tot := range totem_cfg {
		ind := get_totem_index(tot.TotemID)
		if len(tot.Deviceaddr) > 0 && ind >= 0 {
			if mustall || totem[ind].Status == bio_nonexist || totem[ind].Status == bio_error {
				fmt.Println("DEBUG SETUP DEVICES: Device:", tot.TotemID, "-", tot.Deviceaddr)
				var cmd []string
				totemaddr := tot.Deviceaddr
				cmd = make([]string, 0)
				cmd = append(cmd, "CMD/"+totemaddr+"/MOD/"+tot.Pumpdev[1:]+",3/END")
				for i := 0; i < len(tot.Valv_devs); i++ {
					cmd = append(cmd, "CMD/"+totemaddr+"/MOD/"+tot.Valv_devs[i][1:]+",3/END")
				}
				for i := 0; i < len(tot.Peris_dev); i++ {
					cmd = append(cmd, "CMD/"+totemaddr+"/MOD/"+tot.Peris_dev[i][1:]+",3/END")
				}

				nerr := 0
				for k, c := range cmd {
					fmt.Print(k, "  ", c, " ")
					ret := scp_sendmsg_orch(c)
					if !strings.Contains(ret, scp_ack) {
						nerr++
					}
					fmt.Println(ret, nerr)
					if strings.Contains(ret, scp_die) {
						fmt.Println("ERROR SETUP DEVICES: TOTEM DIE", tot.TotemID)
						break
					}
					time.Sleep(scp_refreshwait / 2 * time.Millisecond)
				}
				if nerr > 0 && !devmode {
					totem[ind].Status = bio_nonexist
					fmt.Println("ERROR SETUP DEVICES: TOTEM com erros", tot.TotemID)
				} else if nerr == 0 {
					totem[ind].Status = bio_ready
				}
			}
		}
	}

	fmt.Println("\n\nDEBUG SETUP DEVICES: Configurando BIOFABRICA")
	if mustall || biofabrica.Status == scp_fail {
		for _, bf := range biofabrica_cfg {
			if len(bf.Deviceaddr) > 0 {
				fmt.Println("DEBUG SETUP DEVICES: Device:", bf.DeviceID, "-", bf.Deviceaddr)
				var cmd []string
				bfaddr := bf.Deviceaddr
				cmd = make([]string, 0)
				if bf.Deviceport[0] != 'C' {
					cmd = append(cmd, "CMD/"+bfaddr+"/MOD/"+bf.Deviceport[1:]+",3/END")
				}

				nerr := 0
				for k, c := range cmd {
					fmt.Print()
					ret := scp_sendmsg_orch(c)
					fmt.Println("DEBUG SETUP DEVICES: ", k, "  ", c, " ", ret)
					if !strings.Contains(ret, scp_ack) {
						nerr++
					}
					if ret[0:2] == "DIE" {
						fmt.Println("SLAVE ERROR - DIE")
						nerr++
						break
					}
					time.Sleep(scp_refreshwait / 2 * time.Millisecond)
				}
				if nerr > 0 && !devmode && biofabrica.Status != scp_fail {
					biofabrica.Status = scp_fail
					fmt.Println("CRITICAL SETUP DEVICES: BIOFABRICA com erros")
					board_add_message("EFALHA CRITICA EM VALVULAS DA BIOFABRICA")
				} else if nerr == 0 {
					biofabrica.Status = scp_ready
				}
			}
		}
	}

	finishedsetup = true
}

func scp_get_ph_voltage(bioid string) float64 {
	ind := get_bio_index(bioid)
	if ind < 0 {
		fmt.Println("ERROR SCP GET PH VOLTAGE: Biorreator NAO ENCONTRO", bioid)
		return -1
	}
	bioaddr := bio_cfg[bioid].Deviceaddr
	phdev := bio_cfg[bioid].PH_dev
	aerostatus := bio[ind].Aerator
	aeroratio := bio[ind].AeroRatio

	if len(bioaddr) > 0 {
		if aerostatus {
			if bio[ind].Status != bio_producting {
				return -1
			}
			if !scp_turn_aero(bioid, false, 0, 0) {
				fmt.Println("ERROR SCP GET PH VOLTAGE: Erro ao desligar Aerador do Biorreator", bioid)
				return -1
			}
			time.Sleep(scp_timewaitbeforeph * time.Millisecond)
		}
		cmd_ph := "CMD/" + bioaddr + "/GET/" + phdev + "/END"
		ret_ph := scp_sendmsg_orch(cmd_ph)
		fmt.Println("DEBUG SCP GET PH VOLTAGE: Lendo Voltagem PH do Biorreator", bioid, cmd_ph, ret_ph)
		if aerostatus && !scp_turn_aero(bioid, false, 1, aeroratio) {
			fmt.Println("ERROR SCP GET PH VOLTAGE: Erro ao religar Aerador do Biorreator", bioid)
		}
		params := scp_splitparam(ret_ph, "/")
		if params[0] == scp_ack {
			phint, _ := strconv.Atoi(params[1])
			phfloat := float64(phint) / 100.0
			fmt.Println("DEBUG SCP GET PH VOLTAGE: Voltagem PH", bioid, phint, phfloat)
			return phfloat
		}
	} else {
		fmt.Println("ERROR SCP GET PH VOLTAGE: ADDR Biorreator nao existe", bioid)
	}
	return -1
}

func scp_get_ph(bioid string) float64 {
	ind := get_bio_index(bioid)
	if ind >= 0 {
		phvolt := scp_get_ph_voltage(bioid)
		if phvolt >= 2 && phvolt <= 5 {
			if bio[ind].RegresPH[0] == 0 && bio[ind].RegresPH[1] == 0 {
				fmt.Println("ERROR SCP GET PH: Biorreator com PH NAO CALIBRADO", bioid)
				return -1
			}
			b0 := bio[ind].RegresPH[0]
			b1 := bio[ind].RegresPH[1]
			ph := calc_PH(phvolt, b0, b1)
			if (ph >= 0) && (ph <= 14) {
				fmt.Println("DEBUG SCP GET PH: Biorreator", bioid, " PH=", ph, "PHVolt=", phvolt)
				return ph
			} else {
				fmt.Println("ERROR SCP GET PH: Valor INVALIDO de PH no Biorreator", bioid, " PH=", ph, "PHVolt=", phvolt)
			}
		} else {
			fmt.Println("ERROR SCP GET PH: Valor INVALIDO de PHVolt no Biorreator", bioid, "PHVolt=", phvolt)
		}
	} else {
		fmt.Println("ERROR SCP GET PH: Biorreator nao existe", bioid)
	}
	return -1
}

func scp_update_ph(bioid string) {
	ind := get_bio_index(bioid)
	if ind >= 0 {
		phtmp := scp_get_ph(bioid)
		if phtmp >= 0 {
			bio[ind].PH = float32(math.Trunc(phtmp*10) / 10.0)
		}
	}
}

func scp_update_screen(bioid string, all bool) {
	ind := get_bio_index(bioid)
	if ind < 0 {
		return
	}
	bioscr := bio_cfg[bioid].Screenaddr
	vol_str := fmt.Sprintf("%d", bio[ind].Volume)
	cmd := "CMD/" + bioscr + "/PUT/S232," + vol_str + "/END"
	ret := scp_sendmsg_orch(cmd)
	fmt.Println("DEBUG SCP UPDATE SCREEN: cmd=", cmd, "ret=", ret)
	if !strings.Contains(ret, "ACK") {
		return
	}
	ph_str := fmt.Sprintf("%d", int(bio[ind].PH))
	cmd = "CMD/" + bioscr + "/PUT/S243," + ph_str + "/END"
	ret = scp_sendmsg_orch(cmd)
	if !strings.Contains(ret, "ACK") {
		return
	}
	temp_str := fmt.Sprintf("%d", int(bio[ind].Temperature))
	cmd = "CMD/" + bioscr + "/PUT/S241," + temp_str + "/END"
	ret = scp_sendmsg_orch(cmd)
	if !strings.Contains(ret, "ACK") {
		return
	}
	if all {
		scp_update_screen_times(bioid)
	}
}

func scp_update_screen_times(bioid string) {
	ind := get_bio_index(bioid)
	if ind < 0 {
		return
	}
	bioscr := bio_cfg[bioid].Screenaddr

	var cmd, ret string

	if bio[ind].Step[0] != 0 {
		step_str := fmt.Sprintf("%d", int(bio[ind].Step[0]))
		cmd = "CMD/" + bioscr + "/PUT/S281," + step_str + "/END"
		ret = scp_sendmsg_orch(cmd)
		if !strings.Contains(ret, "ACK") {
			return
		}
	}
	if bio[ind].Step[1] != 0 {
		step_str := fmt.Sprintf("%d", int(bio[ind].Step[1]))
		cmd = "CMD/" + bioscr + "/PUT/S282," + step_str + "/END"
		ret = scp_sendmsg_orch(cmd)
		if !strings.Contains(ret, "ACK") {
			return
		}
	}

	total_str := fmt.Sprintf("%d", int(bio[ind].Timetotal[0]))
	cmd = "CMD/" + bioscr + "/PUT/S283," + total_str + "/END"
	ret = scp_sendmsg_orch(cmd)
	if !strings.Contains(ret, "ACK") {
		return
	}
	total_str = fmt.Sprintf("%d", int(bio[ind].Timetotal[1]))
	cmd = "CMD/" + bioscr + "/PUT/S289," + total_str + "/END"
	ret = scp_sendmsg_orch(cmd)
	if !strings.Contains(ret, "ACK") {
		return
	}

	left_str := fmt.Sprintf("%d", int(bio[ind].Timeleft[0]))
	cmd = "CMD/" + bioscr + "/PUT/S284," + left_str + "/END"
	ret = scp_sendmsg_orch(cmd)
	if !strings.Contains(ret, "ACK") {
		return
	}
	left_str = fmt.Sprintf("%d", int(bio[ind].Timeleft[1]))
	cmd = "CMD/" + bioscr + "/PUT/S285," + left_str + "/END"
	ret = scp_sendmsg_orch(cmd)
	if !strings.Contains(ret, "ACK") {
		return
	}
}

func scp_get_temperature(bioid string) float64 {
	bioaddr := bio_cfg[bioid].Deviceaddr
	tempdev := bio_cfg[bioid].Temp_dev
	if len(bioaddr) > 0 {
		cmd_temp := "CMD/" + bioaddr + "/GET/" + tempdev + "/END"
		ret_temp := scp_sendmsg_orch(cmd_temp)
		params := scp_splitparam(ret_temp, "/")
		fmt.Println("DEBUG SCP GET TEMPERATURE: Lendo Temperatura do Biorreator", bioid, "cmd=", cmd_temp, "ret=", ret_temp, "len=", len(ret_temp), "params=", params)
		if params[0] == scp_ack {
			tempint, err := strconv.Atoi(params[1])
			if err == nil {
				tempfloat := float64(tempint) / 10.0
				return tempfloat
			} else {
				fmt.Println("ERROR SCP GET TEMPERATURE: Retorno invalido, nao numerico", bioid, ret_temp)
			}
		}
	} else {
		fmt.Println("ERROR SCP GET TEMPERATURE: ADDR Biorreator nao existe", bioid)
	}
	return -1
}

func scp_get_volume(main_id string, dev_type string, vol_type string) (int, float64) {
	var dev_addr, vol_dev string
	var ind int
	switch dev_type {
	case scp_bioreactor:
		ind = get_bio_index(main_id)
		if ind >= 0 {
			dev_addr = bio_cfg[main_id].Deviceaddr
			switch vol_type {
			case scp_dev_vol0:
				vol_dev = bio_cfg[main_id].Levellow
			case scp_dev_volusom:
				vol_dev = bio_cfg[main_id].Vol_devs[0]
			case scp_dev_vollaser:
				vol_dev = bio_cfg[main_id].Vol_devs[1]
			default:
				fmt.Println("ERROR SCP GET VOLUME: Disposito de volume invalido para o biorreator", main_id, vol_type)
				return -1, -1
			}
		} else {
			fmt.Println("ERROR SCP GET VOLUME: Biorreator inexistente", main_id)
			return -1, -1
		}
	case scp_ibc:
		ind = get_ibc_index(main_id)
		if ind >= 0 {
			dev_addr = ibc_cfg[main_id].Deviceaddr
			switch vol_type {
			case scp_dev_vol0:
				vol_dev = ibc_cfg[main_id].Levellow
			case scp_dev_volusom:
				vol_dev = ibc_cfg[main_id].Vol_devs[0]
			case scp_dev_vollaser:
				vol_dev = ibc_cfg[main_id].Vol_devs[1]
			default:
				fmt.Println("ERROR SCP GET VOLUME: Disposito de volume invalido para o IBC", main_id, vol_type)
				return -1, -1
			}
		} else {
			fmt.Println("ERROR SCP GET VOLUME: IBC inexistente", main_id)
			return -1, -1
		}
	case scp_biofabrica:
		if vol_type == scp_dev_volfluxo_out {
			dev_addr = biofabrica_cfg["FBF01"].Deviceaddr
			vol_dev = biofabrica_cfg["FBF01"].Deviceport
		} else if vol_type == scp_dev_volfluxo_in1 {
			dev_addr = biofabrica_cfg["FBF02"].Deviceaddr
			vol_dev = biofabrica_cfg["FBF02"].Deviceport
		} else {
			fmt.Println("ERROR SCP GET VOLUME: VOL_TYPE invalido para biofabrica", main_id, vol_type)
			return -1, -1
		}
	default:
		fmt.Println("ERROR SCP GET VOLUME: Tipo de dispositivo invalido", dev_type, main_id)
		return -1, -1
	}
	cmd := "CMD/" + dev_addr + "/GET/" + vol_dev + "/END"
	ret := scp_sendmsg_orch(cmd)
	params := scp_splitparam(ret, "/")
	fmt.Println("DEBUG SCP GET VOLUME: ", main_id, dev_type, vol_type, " == CMD=", cmd, "  RET=", ret)
	var volume float64
	var dint int64
	volume = -1
	if params[0] == scp_ack {
		dint, _ = strconv.ParseInt(params[1], 10, 32)
		if vol_type == scp_dev_vol0 {
			if dint != 0 && dint != 1 {
				fmt.Println("ERROR SCP GET VOLUME: Retorno do ORCH para VOLUME0 com ERRO", main_id, dint, ret)
				return -1, -1
			}
			return int(dint), float64(dint)
		}
	} else {
		fmt.Println("ERROR SCP GET VOLUME: Retorno do ORCH com ERRO", main_id, ret)
		return -1, -1
	}
	if dint == 0 && (vol_type != scp_dev_volfluxo_out || vol_type != scp_dev_volfluxo_in1) {
		fmt.Println("ERROR SCP GET VOLUME: LEITURA INVALIDA do SENSOR", main_id, vol_type, ret)
		return -1, -1
	}
	if vol_type == scp_dev_volusom && dint == 250 {
		fmt.Println("ERROR SCP GET VOLUME: ULTRASSOM em DEADZONE", main_id, dint, ret)
		return -1, -1
	}
	var area, dfloat float64
	area = math.Pi * math.Pow(bio_diametro/2000.0, 2)
	if vol_type != scp_dev_volfluxo_out && vol_type != scp_dev_volfluxo_in1 {
		switch vol_type {
		case scp_dev_volusom:
			switch dev_type {
			case scp_bioreactor:
				dfloat = float64(bio[ind].Vol_zero[0]) - float64(dint)
			case scp_ibc:
				dfloat = float64(ibc[ind].Vol_zero[0]) - float64(dint)
			}
		case scp_dev_vollaser:
			switch dev_type {
			case scp_bioreactor:
				dfloat = float64(bio[ind].Vol_zero[1]) - float64(dint)
			case scp_ibc:
				dfloat = float64(ibc[ind].Vol_zero[1]) - float64(dint)
			}
		}
		volume = area * dfloat
	} else if vol_type == scp_dev_volfluxo_out {
		if dint < int64(biofabrica.LastCountOut) {
			biofabrica.VolOutPart += float64(math.MaxUint16) * flow_ratio
		}
		volume = (float64(dint) * flow_ratio) + biofabrica.VolOutPart
		biofabrica.LastCountOut = uint32(dint)
	} else if vol_type == scp_dev_volfluxo_in1 {
		if dint < int64(biofabrica.LastCountIn1) {
			biofabrica.VolIn1Part += float64(math.MaxUint16) * flow_ratio
		}
		volume = (float64(dint) * flow_ratio) + biofabrica.VolIn1Part
		biofabrica.LastCountIn1 = uint32(dint)
	}

	if volume < 0 {
		fmt.Println("ERROR SCP GET VOLUME: VOLUME NEGATIVO encontrado", main_id, vol_type, dint, volume)
		volume = 0
	}
	return int(dint), volume
}

func scp_refresh_status() {
	var err error
	nslaves := 0
	nslavesok := 0
	nslavesnok := 0
	cmdstatus := "STATUS/END"
	status := scp_sendmsg_orch(cmdstatus)
	pars := scp_splitparam(status, "/")
	if len(pars) > 1 && pars[0] == scp_status {
		if strings.Contains(pars[1], scp_par_slaves) {
			for i, p := range pars[1:] {
				// fmt.Println("DEBUG SCP REFRESH STATUS: ORCH", i, p)
				data := scp_splitparam(p, ",")
				if len(data) > 0 {
					if i == 0 {
						nslaves, err = strconv.Atoi(data[1])
						if err != nil {
							fmt.Println("ERROR SCP REFRESH STATUS: Numero de slaves incorreto vindo do ORCH", pars)
						}
					} else if len(data) > 2 {
						dev_addr := data[0]
						ipaddr := data[1]
						devstatus, _ := strconv.Atoi(data[2])
						fmt.Println("DEBUG SCP REFRESH STATUS: SLAVE DATA:", dev_addr, ipaddr, devstatus)
						if devstatus == scp_state_TCP0 { // Dispositivo OK
							nslavesok++
						} else { // Dispositivo NAO OK
							nslavesnok++
							dev_type := get_devtype_byaddr(dev_addr)
							dev_id := get_devid_byaddr(dev_addr)
							switch dev_type {
							case scp_bioreactor:
								ind := get_bio_index(dev_id)
								if ind >= 0 {
									fmt.Println("DEBUG SCP REFRESH STATUS: FALHA no Biorreator", dev_id)
									bio[ind].Status = bio_error
								} else {
									fmt.Println("ERROR SCP REFRESH STATUS: Biorreator não existe na tabela", dev_id)
								}

							case scp_ibc:
								ind := get_ibc_index(dev_id)
								if ind >= 0 {
									fmt.Println("DEBUG SCP REFRESH STATUS: FALHA no IBC", dev_id)
									ibc[ind].Status = bio_error
								} else {
									fmt.Println("ERROR SCP REFRESH STATUS: IBC não existe na tabela", dev_id)
								}

							case scp_totem:
								ind := get_totem_index(dev_id)
								if ind >= 0 {
									fmt.Println("DEBUG SCP REFRESH STATUS: FALHA no TOTEM", dev_id)
									totem[ind].Status = bio_error
								} else {
									fmt.Println("ERROR SCP REFRESH STATUS: TOMEM não existe na tabela", dev_id)
								}

							case scp_biofabrica:
								fmt.Println("DEBUG SCP REFRESH STATUS: FALHA na Biofabrica", dev_id)
								biofabrica.Status = scp_fail

							default:
								fmt.Println("ERROR SCP REFRESH STATUS: TIPO INVALIDO na tabela", dev_addr)
							}
						}
					}
				}
			}
		}
	}
	fmt.Println("\nDEBUG SCP REFRESH STATUS: ORCH STATUS  Devices:", nslaves, "\tOK:", nslavesok, "\tNOK:", nslavesnok)
}

func scp_sync_functions() {
	t_start_save := time.Now()
	t_start_status := time.Now()
	t_start_screens := time.Now()
	n_bio := 0
	for {
		if finishedsetup {
			t_elapsed_save := uint32(time.Since(t_start_save).Seconds())
			if t_elapsed_save >= scp_timetosave {
				save_all_data(data_filename)
				t_start_save = time.Now()
			}

			t_elapsed_status := uint32(time.Since(t_start_status).Seconds())
			if t_elapsed_status >= scp_refresstatus {
				if finishedsetup {
					scp_refresh_status()
				}
				t_start_status = time.Now()
			}

			t_elapsed_screens := uint32(time.Since(t_start_screens).Seconds())
			if t_elapsed_screens >= scp_refresscreens {
				go scp_update_screen(bio[n_bio].BioreactorID, false)
				n_bio++
				if n_bio >= len(bio) {
					n_bio = 0
				}
				t_start_screens = time.Now()
			}

		}
		time.Sleep(scp_refreshsleep * time.Millisecond)
	}
}

func scp_update_biolevel(bioid string) {
	ind := get_bio_index(bioid)
	if ind < 0 {
		return
	}
	level := (float64(bio[ind].Volume) / float64(bio_cfg[bioid].Maxvolume)) * 10.0
	if level > 10 {
		level = 10
	}
	level_int := uint8(level)
	if level_int != bio[ind].Level {
		bio[ind].Level = level_int
		levels := fmt.Sprintf("%d", level_int)
		cmd := "CMD/" + bio_cfg[bioid].Screenaddr + "/PUT/S231," + levels + "/END"
		ret := scp_sendmsg_orch(cmd)
		fmt.Println("SCREEN:", cmd, level, levels, ret)
	}
}

func scp_get_alldata() {
	if demo {
		return
	}
	t_start_bio := time.Now()
	t_start_ibc := time.Now()
	// t_start_save := time.Now()
	// t_start_status := time.Now()
	t_start_setup := time.Now()
	lastvolin := float64(-1)
	lastvolout := float64(-1)
	hasupdatevolin := false
	hasupdatevolout := false
	firsttime := true
	bio_seq := 0
	ibc_seq := 0
	for {
		if finishedsetup {
			needtorunsetup := false
			t_elapsed_bio := uint32(time.Since(t_start_bio).Seconds())
			mustupdate_bio := false
			if t_elapsed_bio >= scp_mustupdate_bio || firsttime {
				mustupdate_bio = true
				t_start_bio = time.Now()
			}

			hasupdatevolin = false
			hasupdatevolout = false

			for _, b := range bio {
				if len(bio_cfg[b.BioreactorID].Deviceaddr) > 0 && (b.Status != bio_nonexist && b.Status != bio_error) {
					ind := get_bio_index(b.BioreactorID)
					mustupdate_this := (mustupdate_bio && (bio_seq == ind)) || firsttime || bio[ind].Status == bio_update

					if mustupdate_this || b.Status == bio_producting || b.Status == bio_cip || b.Valvs[2] == 1 {
						if t_elapsed_bio%5 == 0 {
							t_tmp := scp_get_temperature(b.BioreactorID)
							if (t_tmp >= 0) && (t_tmp <= TEMPMAX) {
								bio[ind].Temperature = float32(t_tmp)
								if bio[ind].Heater && float32(t_tmp) >= bio[ind].TempMax {
									scp_turn_heater(b.BioreactorID, 0, false)
								}
							}
						}
					}

					if mustupdate_this || b.Status == bio_producting || b.Valvs[1] == 1 {
						if t_elapsed_bio%5 == 0 {
							go scp_update_ph(b.BioreactorID)
						}
					}

					if mustupdate_this || b.Valvs[6] == 1 || b.Valvs[4] == 1 {

						if biofabrica.Useflowin {
							if b.Valvs[6] == 1 && (b.Valvs[3] == 1 || (b.Valvs[2] == 1 && b.Valvs[7] == 1) || (b.Valvs[1] == 1 && b.Valvs[7] == 1)) {
								count, vol_tmp := scp_get_volume(scp_biofabrica, scp_biofabrica, scp_dev_volfluxo_in1)
								if count >= 0 {
									fmt.Println("DEBUG SCP GET ALL DATA: Biorreator", b.BioreactorID, " usando volume vindo do Totem01 =", vol_tmp, lastvolin)
									biofabrica.VolumeIn1 = bio_escala * (math.Trunc(vol_tmp / bio_escala))
									if vol_tmp > lastvolin && lastvolin > 0 {
										biovolin := vol_tmp - lastvolin
										bio[ind].VolInOut += biovolin
										bio[ind].Volume = uint32(bio[ind].VolInOut)
										scp_update_biolevel(b.BioreactorID)
									}
									lastvolin = vol_tmp
									hasupdatevolin = true
								} else {
									fmt.Println("ERROR SCP GET ALL DATA: Valor invalido ao ler Volume INFLUXO", count, vol_tmp)
								}
							} else if b.Valvs[4] == 1 && b.Valvs[5] == 1 && b.Pumpstatus && (biofabrica.Valvs[7] == 1 || biofabrica.Valvs[8] == 1) {
								count, vol_tmp := scp_get_volume(scp_biofabrica, scp_biofabrica, scp_dev_volfluxo_out)
								if count >= 0 {
									fmt.Println("DEBUG SCP GET ALL DATA: Biorreator", b.BioreactorID, " usando volume vindo do FLUXOUT =", vol_tmp, lastvolout)
									biofabrica.VolumeOut = bio_escala * (math.Trunc(vol_tmp / bio_escala))
									if vol_tmp > lastvolout && lastvolout > 0 {
										biovolout := vol_tmp - lastvolout
										bio[ind].VolInOut -= biovolout
										if bio[ind].VolInOut < 0 {
											bio[ind].VolInOut = 0
										}
										bio[ind].Volume = uint32(bio[ind].VolInOut)
										scp_update_biolevel(b.BioreactorID)
									}
									lastvolout = vol_tmp
									hasupdatevolout = true
								} else {
									fmt.Println("ERROR SCP GET ALL DATA: Valor invalido ao ler Volume INFLUXO", count, vol_tmp)
								}
							}

						} else {
							var vol0 float64 = -1
							var dint int
							dint, _ = scp_get_volume(b.BioreactorID, scp_bioreactor, scp_dev_vol0)
							if dint >= 0 {
								bio[ind].Vol0 = dint
							}
							if bio[ind].Status == bio_cip && (bio[ind].Valvs[1] == 1 || bio[ind].Valvs[2] == 1 || bio[ind].Valvs[3] == 1) {
								fmt.Println("DEBUG GET ALLDATA: CIP EXECUTANDO - IGNORANDO VOLUME ZERO", b.BioreactorID)
							} else {
								vol0 = float64(dint)
								fmt.Println("DEBUG GET ALLDATA: Volume ZERO", b.BioreactorID, bio_cfg[b.BioreactorID].Deviceaddr, dint, vol0)
							}

							var vol1, vol1_pre float64
							vol1 = -1
							dint, vol1_pre = scp_get_volume(b.BioreactorID, scp_bioreactor, scp_dev_volusom)
							if dint == 1 || dint == 0 { // WORKARROUND para quando retornar valor do V0 no ULTRASSOM
								dint, vol1_pre = scp_get_volume(b.BioreactorID, scp_bioreactor, scp_dev_volusom)
							}

							if vol0 == 0 {
								if dint > 0 && float32(dint) >= (bio_v1_zero*0.7) && float32(dint) <= (bio_v1_zero*1.3) {
									fmt.Println("DEBUG GET ALLDATA: Volume ZERO atingido, mudando Vol0 USOM", b.BioreactorID, dint)
									//
									// ATENCAO, DELISGUEI ATE IMPLEMENTAR O ZERO MANUAL
									//
									// bio[ind].Vol_zero[0] = float32(dint)
								} else {
									fmt.Println("ERROR GET ALLDATA: Volume ZERO atingido, mas ULTRASSOM fora da faixa", b.BioreactorID, dint)
								}
							}
							if dint > 0 && vol1_pre >= 0 {
								vol1 = 10 * (math.Trunc(vol1_pre / 10))
							}
							bio[ind].Vol1 = int32(vol1_pre)
							fmt.Println("DEBUG GET ALLDATA: Volume USOM", b.BioreactorID, bio_cfg[b.BioreactorID].Deviceaddr, dint, vol1_pre)

							var vol2, vol2_pre float64
							vol2 = -1
							dint, vol2_pre = scp_get_volume(b.BioreactorID, scp_bioreactor, scp_dev_vollaser)
							if vol0 == 0 {
								if dint > 0 && float32(dint) >= (bio_v2_zero*0.7) && float32(dint) <= (bio_v2_zero*1.3) {
									fmt.Println("DEBUG GET ALLDATA: Volume ZERO atingido, mudando Vol0 LASER ", b.BioreactorID, dint)
									//
									// ATENCAO, DELISGUEI ATE IMPLEMENTAR O ZERO MANUAL
									//
									//bio[ind].Vol_zero[1] = float32(dint)
								} else {
									fmt.Println("ERROR GET ALLDATA: Volume ZERO atingido, mas ULTRASSOM fora da faixa", b.BioreactorID, dint)
								}
							}
							if dint > 0 && vol2_pre >= 0 {
								vol2 = 10 * (math.Trunc(vol2_pre / 10))
							}
							bio[ind].Vol2 = int32(vol2_pre)
							fmt.Println("DEBUG GET ALLDATA: Volume LASER", b.BioreactorID, bio_cfg[b.BioreactorID].Deviceaddr, dint, vol2_pre)

							var volc float64
							volc = float64(bio[ind].Volume)
							if vol0 == 0 {
								if vol1 < 100 && vol2 < 100 {
									fmt.Println("DEBUG GET ALLDATA: Volume ZERO DETECTADO", b.BioreactorID)
									volc = 0
								} else {
									fmt.Println("ERROR GET ALLDATA: Volume ZERO DETECTADO e Vol1/Vol2 divergem", b.BioreactorID)
									bio_add_message(b.BioreactorID, "AFavor verificar SENSOR de nivel 0")
									volc = -1
								}
							}
							if vol0 != 0 || volc == -1 {
								if vol1 < 0 && vol2 < 0 {
									fmt.Println("ERROR GET ALLDATA: IGNORANDO VOLUMES INVALIDOS", b.BioreactorID, vol1, vol2)
								} else {
									if bio[ind].Valvs[4] == 1 { // Desenvase
										if vol1 < 0 {
											if vol2 >= 0 && vol2 < float64(bio[ind].Volume) {
												volc = vol2
											} else if vol1 >= 0 {
												volc = vol1
											}
										} else {
											volc = vol1
										}

									} else if bio[ind].Valvs[6] == 1 { // Carregando Agua
										if bio[ind].Valvs[2] == 0 { // Sprayball desligado
											if vol1 < 0 {
												if vol2 >= 0 && vol2 > float64(bio[ind].Volume) {
													volc = vol2
												} else if vol1 >= 0 {
													volc = vol1
												}
											} else {
												volc = vol1
											}
										}

									} else {
										if bio[ind].Status == bio_producting {
											if vol1 >= 0 {
												volc = vol1
											} else if vol2 >= 0 {
												volc = vol2
											}
										} else if bio[ind].Status == bio_cip && bio[ind].Valvs[2] == 1 { //  Se for CIP e Sptrayball ligado, ignorar
											fmt.Println("DEBUG GET ALLDATA: CIP+SPRAYBALL - IGNORANDO VOLUMES ", b.BioreactorID, vol1, vol2)
										} else {
											if vol1 >= 0 {
												volc = vol1
											} else if vol2 >= 0 {
												volc = vol2
											}
										}
									}
								}
							}

							if b.Status == bio_update {
								if vol1 != -1 || vol2 != -1 {
									bio[ind].Status = bio_ready
								} else if !devmode {
									bio_add_message(b.BioreactorID, "AVerifique sensores de Volume")
									bio[ind].Status = bio_error
								}
							}

							if volc >= 0 {
								bio[ind].Volume = uint32(volc)
								scp_update_biolevel(b.BioreactorID)
								if volc == 0 && vol0 == 0 && bio[ind].Status != bio_producting && bio[ind].Status != bio_loading && bio[ind].Status != bio_cip {
									bio[ind].Status = bio_empty
								}
							}
						}
					}
					if devmode && bio[ind].Status == bio_update {
						if bio[ind].Volume == 0 {
							bio[ind].Status = bio_empty
						} else {
							bio[ind].Status = bio_ready
						}
					}
				} else if b.Status == bio_nonexist || b.Status == bio_error {
					needtorunsetup = true
				}
				time.Sleep(scp_refreshwait * time.Millisecond)
			}
			if mustupdate_bio {
				bio_seq++
				if bio_seq >= len(bio) {
					bio_seq = 0
				}
			}
			t_elapsed_ibc := uint32(time.Since(t_start_ibc).Seconds())
			mustupdate_ibc := false
			if t_elapsed_ibc >= scp_mustupdate_ibc || firsttime {
				mustupdate_ibc = true
				t_start_ibc = time.Now()
			}
			for _, b := range ibc {
				if len(ibc_cfg[b.IBCID].Deviceaddr) > 0 && (b.Status != bio_nonexist && b.Status != bio_error) {
					ind := get_ibc_index(b.IBCID)
					mustupdate_this := (mustupdate_ibc && (ibc_seq == ind)) || firsttime || ibc[ind].Status == bio_update
					if ind >= 0 && (mustupdate_this || b.Valvs[3] == 1 || b.Valvs[2] == 1) {
						if devmode {
							fmt.Println("DEBUG GET ALLDATA: Lendo dados do IBC", b.IBCID)
						}

						var vol0 float64 = -1
						var dint int
						dint, _ = scp_get_volume(b.IBCID, scp_ibc, scp_dev_vol0)
						if dint <= 0 || dint > 1 {
							dint, _ = scp_get_volume(b.IBCID, scp_ibc, scp_dev_vol0)
						}
						if dint >= 0 {
							ibc[ind].Vol0 = dint
						}
						if ibc[ind].Status == bio_cip && (ibc[ind].Valvs[0] == 1 || ibc[ind].Valvs[1] == 1) {
							fmt.Println("DEBUG GET ALLDATA: CIP EXECUTANDO - IGNORANDO VOLUME ZERO", b.IBCID)
						} else {
							vol0 = float64(dint)
							fmt.Println("DEBUG GET ALLDATA: Volume ZERO", b.IBCID, ibc_cfg[b.IBCID].Deviceaddr, dint, vol0)
						}

						var vol1, vol1_pre float64
						vol1 = -1
						dint, vol1_pre = scp_get_volume(b.IBCID, scp_ibc, scp_dev_volusom)
						if dint == 1 { // WORKARROUND para quando retornar valor do V0 no ULTRASSOM
							dint, vol1_pre = scp_get_volume(b.IBCID, scp_ibc, scp_dev_volusom)
						}

						if vol0 == 0 {
							if dint > 0 && float32(dint) >= (ibc_v1_zero*0.8) && float32(dint) <= (ibc_v1_zero*1.2) {
								fmt.Println("DEBUG GET ALLDATA: Volume ZERO atingido, mudando Vol0 USOM", b.IBCID, dint)
								//
								// ATENCAO, DELISGUEI ATE IMPLEMENTAR O ZERO MANUAL
								//
								// ibc[ind].Vol_zero[0] = float32(dint)
							} else {
								fmt.Println("ERROR GET ALLDATA: Volume ZERO atingido, mas ULTRASSOM fora da faixa", b.IBCID, dint)
							}
						}
						if dint > 0 && vol1_pre >= 0 {
							vol1 = 10 * (math.Trunc(vol1_pre / 10))
						}
						ibc[ind].Vol1 = int32(vol1_pre)
						fmt.Println("DEBUG GET ALLDATA: Volume USOM", b.IBCID, bio_cfg[b.IBCID].Deviceaddr, dint, vol1_pre)

						var vol2, vol2_pre float64
						vol2 = -1
						dint, vol2_pre = scp_get_volume(b.IBCID, scp_ibc, scp_dev_vollaser)
						if vol0 == 0 {
							if dint > 0 && float32(dint) >= (ibc_v2_zero*0.8) && float32(dint) <= (ibc_v2_zero*1.2) {
								fmt.Println("DEBUG GET ALLDATA: Volume ZERO atingido, mudando Vol0 LASER", b.IBCID, dint)
								//
								// ATENCAO, DELISGUEI ATE IMPLEMENTAR O ZERO MANUAL
								//
								// ibc[ind].Vol_zero[1] = float32(dint)
							} else {
								fmt.Println("ERROR GET ALLDATA: Volume ZERO atingido, mas ULTRASSOM fora da faixa", b.IBCID, dint)
							}
						}
						if dint > 0 && vol2_pre >= 0 {
							vol2 = 10 * (math.Trunc(vol2_pre / 10))
						}
						ibc[ind].Vol2 = int32(vol2_pre)
						fmt.Println("DEBUG GET ALLDATA: Volume LASER", b.IBCID, bio_cfg[b.IBCID].Deviceaddr, dint, vol2_pre)

						// cmd2 := "CMD/" + ibcaddr + "/GET/" + v2dev + "/END"
						// ret2 := scp_sendmsg_orch(cmd2)
						// params = scp_splitparam(ret2, "/")
						// vol2 = -1
						// if params[0] == scp_ack {
						// 	dint, _ := strconv.Atoi(params[1])
						// 	area = 0
						// 	dfloat = 0
						// 	if vol0 == 0 && (dint > 0 && float32(dint) >= (ibc_v2_zero*0.7) && float32(dint) <= (ibc_v2_zero*1.3)) {
						// 		fmt.Println("DEBUG GET ALLDATA: Volume ZERO atingido, mudango Vol0", b.IBCID, dint)
						// 		b.Vol_zero[1] = float32(dint)
						// 	}
						// 	area = math.Pi * math.Pow(bio_diametro/2000.0, 2)
						// 	dfloat = float64(b.Vol_zero[1]) - float64(dint)
						// 	vol2_pre := area * dfloat
						// 	if dint > 0 {
						// 		if vol2_pre >= 0 {
						// 			vol2 = 10 * (math.Trunc(vol2_pre / 10))
						// 		} else {
						// 			vol2 = 0
						// 		}
						// 		ibc[ind].Vol2 = int32(vol2_pre)
						// 	} else {
						// 		ibc[ind].Vol2 = -1
						// 	}

						// 	fmt.Println("DEBUG GET ALLDATA: Volume LASER", b.IBCID, ibc_cfg[b.IBCID].Deviceaddr, dint, area, dfloat, vol2, ret2)
						// } else {
						// 	fmt.Println("ERROR GET ALLDATA: LASER", b.IBCID, ret2, params)
						// }

						var volc float64
						volc = float64(ibc[ind].Volume)
						if vol0 == 0 {
							fmt.Println("DEBUG GET ALLDATA: Volume ZERO DETECTADO", b.IBCID)
							volc = 0
						} else if vol1 < 0 && vol2 < 0 {
							fmt.Println("ERROR GET ALLDATA: IGNORANDO VOLUMES INVALIDOS", b.IBCID, vol1, vol2)
						} else {
							if ibc[ind].Valvs[3] == 1 { // Desenvase
								if vol1 < 0 {
									if vol2 >= 0 && vol2 < float64(ibc[ind].Volume) {
										volc = vol2
									}
								} else {
									volc = vol1
								}
							} else if ibc[ind].Valvs[2] == 1 { // Carregando
								if ibc[ind].Valvs[1] == 0 { // Sprayball desligado
									if vol1 < 0 {
										if vol2 >= 0 && vol2 > float64(ibc[ind].Volume) {
											volc = vol2
										}
									} else {
										volc = vol1
									}
								}
							} else {
								if ibc[ind].Status == bio_cip && ibc[ind].Valvs[1] == 1 { //  Se for CIP e Sptrayball ligado, ignorar
									fmt.Println("DEBUG GET ALLDATA: CIP+SPRAYBALL - IGNORANDO VOLUMES ", b.IBCID, vol1, vol2)
								} else {
									if vol1 >= 0 {
										volc = vol1
									} else if vol2 >= 0 {
										volc = vol2
									}
								}
							}
						}

						if volc >= 0 {
							ibc[ind].Volume = uint32(volc)
							level := (volc / float64(ibc_cfg[b.IBCID].Maxvolume)) * 10.0
							level_int := uint8(level)
							if level_int != ibc[ind].Level {
								ibc[ind].Level = level_int
							}
							if volc == 0 && vol0 == 0 && ibc[ind].Status != bio_loading && ibc[ind].Status != bio_cip {
								ibc[ind].Status = bio_empty
							}
						}
						if devmode && ibc[ind].Status == bio_update {
							if ibc[ind].Volume == 0 {
								ibc[ind].Status = bio_empty
							} else {
								ibc[ind].Status = bio_ready
							}
						}

					}
				} else if b.Status == bio_nonexist || b.Status == bio_error {
					needtorunsetup = true
				}
				time.Sleep(scp_refreshwait * time.Millisecond)
			}
			if mustupdate_ibc {
				ibc_seq++
				if ibc_seq >= len(ibc) {
					ibc_seq = 0
				}
			}

			if !hasupdatevolout && (mustupdate_ibc || biofabrica.Valvs[7] == 1 || biofabrica.Valvs[8] == 1) {
				count, vol_tmp := scp_get_volume(scp_biofabrica, scp_biofabrica, scp_dev_volfluxo_out)
				if count >= 0 {
					fmt.Println("DEBUG SCP GET ALL DATA: Volume lido no desenvase =", vol_tmp)
					biofabrica.VolumeOut = bio_escala * (math.Trunc(vol_tmp / bio_escala))
					lastvolout = vol_tmp
				} else {
					fmt.Println("ERROR SCP GET ALL DATA: Valor invalido ao ler Volume OUTFLUXO", count, vol_tmp)
				}
			}

			if !hasupdatevolin && (mustupdate_bio || biofabrica.Valvs[1] == 1) {
				count, vol_tmp := scp_get_volume(scp_biofabrica, scp_biofabrica, scp_dev_volfluxo_in1)
				if count >= 0 {
					fmt.Println("DEBUG SCP GET ALL DATA: Volume lido na entrada vindo do Totem01 =", vol_tmp)
					biofabrica.VolumeIn1 = bio_escala * (math.Trunc(vol_tmp / bio_escala))
					lastvolin = vol_tmp

				} else {
					fmt.Println("ERROR SCP GET ALL DATA: Valor invalido ao ler Volume INFLUXO", count, vol_tmp)
				}
			}

			// t_elapsed_save := uint32(time.Since(t_start_save).Seconds())
			// if t_elapsed_save >= scp_timetosave {
			// 	save_all_data(data_filename)
			// 	t_start_save = time.Now()
			// }

			firsttime = false

			for _, t := range totem {
				if t.Status == bio_error || t.Status == bio_nonexist {
					needtorunsetup = true
				}
			}

			needtorunsetup = needtorunsetup || (biofabrica.Status == scp_fail)

			t_elapsed_setup := uint32(time.Since(t_start_setup).Seconds())
			if t_elapsed_setup >= scp_checksetup {
				if finishedsetup && needtorunsetup {
					go scp_setup_devices(false)
				}
				t_start_setup = time.Now()
			}

			// t_elapsed_status := uint32(time.Since(t_start_status).Seconds())
			// if t_elapsed_status >= scp_refresstatus {
			// 	if finishedsetup {
			// 		scp_refresh_status()
			// 	}
			// 	t_start_status = time.Now()
			// }

		}
		time.Sleep(scp_refreshsleep * time.Millisecond)
	}
}

func test_file(filename string) bool {
	mf, err := os.Stat(filename)
	if err != nil {
		if !os.IsNotExist(err) {
			checkErr(err)
		}
		return false
	}
	fmt.Println("DEBUG: Arquivo encontrado", mf.Name())
	return true
}

func set_valvs_value(vlist []string, value int, abort_on_error bool) int {
	tot := 0
	for _, p := range vlist {
		if p != "END" {
			val, ok := valvs[p]
			if ok {
				sub := scp_splitparam(p, "/")
				dtype := get_scp_type(sub[0])
				if val == (1 - value) {
					if !set_valv_status(dtype, sub[0], sub[1], value) {
						fmt.Println("ERROR SET VALVS VALUE: nao foi possivel setar valvula", p)
						if abort_on_error {
							return -1
						}
					}
					tot++
				} else if val == 1 {
					fmt.Println("ERROR SET VALVS VALUE: nao foi possivel setar valvula", p)
					if abort_on_error {
						return -1
					}
				} else {
					fmt.Println("ERROR SET VALVS VALUE: valvula com erro", p)
					if abort_on_error {
						return -1
					}
				}
			} else {
				fmt.Println("ERROR SET VALVS VALUE: valvula nao existe", p)
				if abort_on_error {
					return -1
				}
			}
		}
	}
	return tot
}

func test_path(vpath []string, value int) bool {
	ret := true
	for _, p := range vpath {
		if p == "END" {
			break
		}
		val, ok := valvs[p]
		// fmt.Println("step", p, "ret=", ret, "val=", val, "ok=", ok)
		ret = ret && (val == value) && ok
		// fmt.Println("ret final=", ret)
	}
	return ret
}

func turn_withdraw_var(value bool) {
	withdrawrunning = value
}

func scp_run_linewash(lines string) bool {
	fmt.Println("DEBUG SCP RUN LINEWASH: Executando enxague das linhas", lines)
	var pathclean string = ""
	totem := ""
	switch lines {
	case line_13:
		pathclean = "TOTEM01-CLEAN3"
		totem = "TOTEM01"
	case line_14:
		pathclean = "TOTEM01-CLEAN4"
		totem = "TOTEM01"
	case line_23:
		pathclean = "TOTEM02-CLEAN3"
		totem = "TOTEM02"
	case line_24:
		pathclean = "TOTEM02-CLEAN4"
		totem = "TOTEM02"
	default:
		fmt.Println("ERROR SCP RUN LINEWASH: Linhas invalidas", lines)
		return false
	}

	pathstr := paths[pathclean].Path
	if len(pathstr) == 0 {
		fmt.Println("ERROR SCP RUN LINEWASH:: path WASH linha nao existe", pathclean)
		return false
	}
	var time_to_clean int64 = int64(paths[pathclean].Cleantime) * 1000
	vpath := scp_splitparam(pathstr, ",")
	if !scp_turn_pump(scp_totem, totem, vpath, 1) {
		fmt.Println("ERROR SCP RUN LINEWASH: Falha ao abrir valvulvas e ligar bomba do totem", totem, vpath)
		return false
	}

	time.Sleep(time.Duration(time_to_clean) * time.Millisecond)

	if !scp_turn_pump(scp_totem, totem, vpath, 0) {
		fmt.Println("ERROR SCP RUN LINEWASH: Falha ao fechar valvulvas e desligar bomba do totem", totem, vpath)
		return false
	}

	board_add_message("IEnxague concluído Linhas " + lines)
	return true
}

func scp_run_linecip(lines string) bool {
	fmt.Println("DEBUG SCP RUN LINEWASH: Executando CIP das linhas", lines)

	var pathclean string = ""
	totem_str := ""
	switch lines {
	case line_13:
		pathclean = "TOTEM01-CLEAN3"
		totem_str = "TOTEM01"
	case line_14:
		pathclean = "TOTEM01-CLEAN4"
		totem_str = "TOTEM01"
	case line_23:
		pathclean = "TOTEM02-CLEAN3"
		totem_str = "TOTEM02"
	case line_24:
		pathclean = "TOTEM02-CLEAN4"
		totem_str = "TOTEM02"
	default:
		fmt.Println("ERROR SCP RUN LINEWASH: Linhas invalidas", lines)
		return false
	}

	pathstr := paths[pathclean].Path
	if len(pathstr) == 0 {
		fmt.Println("ERROR SCP RUN LINEWASH:: path WASH linha nao existe", pathclean)
		return false
	}
	var time_to_clean int64 = int64(paths[pathclean].Cleantime) * 1000
	vpath := scp_splitparam(pathstr, ",")
	vpath_peris := scp_splitparam(pathstr, ",")
	perisvalv := totem_str + "/V2"
	n := len(vpath)
	vpath_peris = append(vpath_peris[1:n-1], perisvalv)
	vpath_peris = append(vpath_peris, "END")
	fmt.Println("DEBUG SCP RUN LINEWASH: vpath ", vpath)
	fmt.Println("DEBUG SCP RUN LINEWASH: vpath peris ", vpath_peris)

	all_peris := [2]string{"P1", "P2"}
	tmax := scp_timewaitvalvs * 10
	if devmode || testmode {
		tmax = scp_timeoutdefault / 100
	}

	for _, peris_str := range all_peris {

		if test_path(vpath_peris, 0) {
			if set_valvs_value(vpath_peris, 1, true) < 0 {
				fmt.Println("ERROR SCP RUN LINEWASH: ERRO ao abrir valvulas no path ", vpath_peris)
				return false
			}
		} else {
			fmt.Println("ERROR SCP RUN LINEWASH: ERRO nas valvulas no path ", vpath_peris)
			return false
		}

		for i := 0; i < tmax; i++ {
			time.Sleep(100 * time.Millisecond)
		}

		if !scp_turn_peris(scp_totem, totem_str, peris_str, 1) {
			fmt.Println("ERROR SCP RUN LINEWASH: ERROR ao ligar peristaltica em", totem_str, peris_str)
			return false
		}

		time.Sleep(scp_timelinecip * time.Second) // VALIDAR TEMPO DE BLEND na LINHA

		if !scp_turn_peris(scp_totem, totem_str, peris_str, 0) {
			fmt.Println("ERROR SCP RUN LINEWASH: ERROR ao desligar peristaltica em", totem_str, peris_str)
			return false
		}

		if test_path(vpath_peris, 1) {
			if set_valvs_value(vpath_peris, 0, true) < 0 {
				fmt.Println("ERROR SCP RUN LINEWASH: ERRO ao fechar valvulas no path ", vpath_peris)
				return false
			}
		} else {
			fmt.Println("ERROR SCP RUN LINEWASH: ERRO nas valvulas no path ", vpath_peris)
			return false
		}

		for i := 0; i < tmax; i++ {
			time.Sleep(100 * time.Millisecond)
		}

		if !scp_turn_pump(scp_totem, totem_str, vpath, 1) {
			fmt.Println("ERROR SCP RUN LINEWASH: Falha ao abrir valvulvas e ligar bomba do totem", totem, vpath)
			return false
		}

		time.Sleep(time.Duration(time_to_clean) * time.Millisecond)

		if !scp_turn_pump(scp_totem, totem_str, vpath, 0) {
			fmt.Println("ERROR SCP RUN LINEWASH: Falha ao fechar valvulvas e desligar bomba do totem", totem, vpath)
			return false
		}
	}

	board_add_message("ICIP concluído Linhas " + lines)
	return true
}

func scp_run_withdraw(devtype string, devid string, linewash bool, untilempty bool) int {
	withdrawmutex.Lock()
	turn_withdraw_var(true)
	defer withdrawmutex.Unlock()
	defer turn_withdraw_var(false)

	switch devtype {
	case scp_bioreactor:
		ind := get_bio_index(devid)
		if ind < 0 {
			fmt.Println("ERROR RUN WITHDRAW 01: Biorreator nao existe", devid)
			return -1
		}
		prev_status := bio[ind].Status
		pathid := devid + "-" + bio[ind].OutID
		pathstr := paths[pathid].Path
		if len(pathstr) == 0 {
			fmt.Println("ERROR RUN WITHDRAW 01: path nao existe", pathid)
			return -1
		}
		vpath := scp_splitparam(pathstr, ",")
		if !test_path(vpath, 0) {
			fmt.Println("ERROR RUN WITHDRAW 02: falha de valvula no path", pathid)
			return -1
		}
		board_add_message("CDesenvase " + devid + " para " + bio[ind].OutID)
		var pilha []string = make([]string, 0)
		for k, p := range vpath {
			fmt.Println("step", k, p)
			if p == "END" {
				break
			}
			val, ok := valvs[p]
			if ok {
				sub := scp_splitparam(p, "/")
				dtype := get_scp_type(sub[0])
				if val == 0 {
					if !set_valv_status(dtype, sub[0], sub[1], 1) {
						fmt.Println("ERROR RUN WITHDRAW 03: nao foi possivel setar valvula", p)
						set_valvs_value(pilha, 0, false) // undo
						return -1
					}
				} else if val == 1 {
					fmt.Println("ERROR RUN WITHDRAW 04: valvula ja aberta", p)
					set_valvs_value(pilha, 0, false) // undo
					return -1
				} else {
					fmt.Println("ERROR RUN WITHDRAW 05: valvula com erro", p)
					set_valvs_value(pilha, 0, false) // undo
					return -1
				}
			} else {
				fmt.Println("ERROR RUN WITHDRAW 06: valvula nao existe", p)
				set_valvs_value(pilha, 0, false) // undo
				return -1
			}
			pilha = append([]string{p}, pilha...)
		}
		fmt.Println(pilha)
		vol_ini := bio[ind].Volume
		bio[ind].Status = bio_unloading
		time.Sleep(scp_timewaitvalvs * time.Millisecond)
		fmt.Println("WARN RUN WITHDRAW 07: Ligando bomba", devid)
		biodev := bio_cfg[devid].Deviceaddr
		bioscr := bio_cfg[devid].Screenaddr
		pumpdev := bio_cfg[devid].Pump_dev
		bio[ind].Pumpstatus = true
		cmd1 := "CMD/" + biodev + "/PUT/" + pumpdev + ",1/END"
		cmd2 := "CMD/" + bioscr + "/PUT/S270,1/END"
		ret1 := scp_sendmsg_orch(cmd1)
		fmt.Println("DEBUG RUN WITHDRAW 08: CMD1 =", cmd1, " RET=", ret1)
		ret2 := scp_sendmsg_orch(cmd2)
		fmt.Println("DEBUG RUN WITHDRAW 09: CMD2 =", cmd2, " RET=", ret2)
		if !strings.Contains(ret1, scp_ack) && !devmode {
			fmt.Println("ERROR RUN WITHDRAW 10: BIORREATOR falha ao ligar bomba")
			cmd2 := "CMD/" + bioscr + "/PUT/S270,0/END"
			scp_sendmsg_orch(cmd2)
			set_valvs_value(pilha, 0, false)
			return -1
		}
		var vol_out int64
		var vol_bio_out_start float64
		vol_bio_init := bio[ind].Volume
		if bio[ind].OutID == scp_out || bio[ind].OutID == scp_drop {
			_, vol_bio_out_start = scp_get_volume(scp_biofabrica, scp_biofabrica, scp_dev_volfluxo_out)
		} else {
			vol_bio_out_start = -1
		}
		t_start := time.Now()
		var maxtime float64
		if devmode || testmode {
			maxtime = 60
		} else {
			maxtime = scp_maxtimewithdraw
		}
		t_elapsed := float64(0)
		mustwaittime := false
		waittime := float64(0)
		if untilempty && biofabrica.Useflowin && get_scp_type(bio[ind].OutID) == scp_ibc {
			// bio[ind].ShowVol = false
			mustwaittime = true
			waittime = float64(bio[ind].Volume)*bio_emptying_rate + 20
		}
		for {
			vol_now := bio[ind].Volume
			// t_now := time.Now()
			t_elapsed = time.Since(t_start).Seconds()
			vol_out = int64(vol_ini - vol_now)
			vol_bio_out_now := biofabrica.VolumeOut - vol_bio_out_start
			if bio[ind].Withdraw == 0 {
				break
			}
			if untilempty {
				if mustwaittime {
					if t_elapsed >= waittime {
						break
					}
				} else if vol_now == 0 {
					break
				}
			} else {
				if vol_now == 0 || (vol_now < vol_ini && vol_out >= int64(bio[ind].Withdraw)) {
					if vol_now == 0 || vol_bio_out_start < 0 || vol_bio_out_now >= float64(bio[ind].Withdraw) {
						fmt.Println("DEBUG RUN WITHDRAW 11: Volume de desenvase atingido", vol_ini, vol_now, bio[ind].Withdraw)
						break
					}
				}
			}
			if t_elapsed > maxtime {
				fmt.Println("DEBUG RUN WITHDRAW 12: Tempo maximo de withdraw esgotado", t_elapsed, maxtime)
				break
			}
			if biofabrica.Useflowin && mustwaittime && int32(t_elapsed)%5 == 0 {
				volout := t_elapsed / bio_emptying_rate
				vol_tmp := float64(vol_bio_init) - volout
				if vol_tmp < 0 {
					vol_tmp = 0
				}
				bio[ind].Volume = uint32(vol_tmp)
				// go scp_update_biolevel(bio[ind].BioreactorID)
				go scp_update_screen(bio[ind].BioreactorID, false)
			}
			time.Sleep(scp_refreshwait * time.Millisecond)
		}
		if bio[ind].Volume == 0 && bio[ind].Vol0 != 0 {
			for i := 0; i < 200 && bio[ind].Withdraw != 0; i++ {
				if bio[ind].Vol0 == 0 {
					break
				}
				time.Sleep(100 * time.Millisecond)
			}
		}
		if biofabrica.Useflowin && untilempty {
			if bio[ind].Withdraw != 0 {
				bio[ind].Volume = 0
				bio[ind].VolInOut = 0
			} else if mustwaittime {
				volout := t_elapsed * bio_emptying_rate
				vol_tmp := float64(vol_bio_init) - volout
				if vol_tmp < 0 {
					vol_tmp = 0
				}
				bio[ind].Volume = uint32(vol_tmp)
				bio[ind].VolInOut = vol_tmp
			}
		}
		if bio[ind].Volume == 0 {
			bio[ind].Status = bio_empty
			bio[ind].Level = 0
			bio[ind].Step = [2]int{0, 0}
			bio[ind].Timetotal = [2]int{0, 0}
			bio[ind].Timeleft = [2]int{0, 0}
		}

		// go scp_update_biolevel(bio[ind].BioreactorID)
		go scp_update_screen(bio[ind].BioreactorID, false)

		bio[ind].Withdraw = 0

		board_add_message("IDesenvase concluido")
		fmt.Println("WARN RUN WITHDRAW 13: Desligando bomba", devid)

		set_valvs_value(pilha, 0, false)
		if untilempty {
			time.Sleep((scp_timewaitvalvs / 3) * 2 * time.Millisecond)
		}

		bio[ind].Pumpstatus = false
		cmd1 = "CMD/" + biodev + "/PUT/" + pumpdev + ",0/END"
		cmd2 = "CMD/" + bioscr + "/PUT/S270,0/END"
		ret1 = scp_sendmsg_orch(cmd1)
		fmt.Println("DEBUG RUN WITHDRAW 14: CMD1 =", cmd1, " RET=", ret1)
		ret2 = scp_sendmsg_orch(cmd2)
		fmt.Println("DEBUG RUN WITHDRAW 15: CMD2 =", cmd2, " RET=", ret2)

		if untilempty {
			time.Sleep(scp_timewaitvalvs / 3 * time.Millisecond)
		} else {
			time.Sleep(scp_timewaitvalvs * time.Millisecond)
		}

		// bio[ind].Status = bio_ready
		dest_type := get_scp_type(bio[ind].OutID)
		if dest_type == scp_ibc {
			ind_ibc := get_ibc_index(bio[ind].OutID)
			ibc[ind_ibc].Organism = bio[ind].Organism
			ibc[ind_ibc].OrgCode = bio[ind].OrgCode
		}
		if linewash {
			var pathclean string = ""
			if dest_type == scp_out || dest_type == scp_drop {
				pathclean = "TOTEM02-CLEAN4"
				board_add_message("IEnxague LINHAS 2/4")
			} else if dest_type == scp_ibc {
				pathclean = "TOTEM02-CLEAN3"
				board_add_message("IEnxague LINHAS 2/3")
			} else {
				fmt.Println("ERROR RUN WITHDRAW 16: destino para clean desconhecido", dest_type)
				return -1
			}
			pathstr = paths[pathclean].Path
			if len(pathstr) == 0 {
				fmt.Println("ERROR RUN WITHDRAW 17: path WASH linha nao existe", pathclean)
				return -1
			}
			var time_to_clean int64 = int64(paths[pathclean].Cleantime) * 1000
			vpath = scp_splitparam(pathstr, ",")
			if !test_path(vpath, 0) {
				fmt.Println("ERROR RUN WITHDRAW 18: falha de valvula no path", pathstr)
				return -1
			}
			if set_valvs_value(vpath, 1, true) < 1 {
				fmt.Println("ERROR RUN WITHDRAW 19: Falha ao abrir valvulas CLEAN linha", pathstr)
				set_valvs_value(vpath, 0, false)
			}
			time.Sleep(scp_timewaitvalvs * time.Millisecond)
			fmt.Println("WARN RUN WITHDRAW 20: Ligando bomba TOTEM02", devid)
			tind := get_totem_index("TOTEM02")
			if tind < 0 {
				fmt.Println("WARN RUN WITHDRAW 21: TOTEM02 nao encontrado", totem)
			}
			totemdev := totem_cfg["TOTEM02"].Deviceaddr
			pumpdev = totem_cfg["TOTEM02"].Pumpdev
			totem[tind].Pumpstatus = true
			cmd1 = "CMD/" + totemdev + "/PUT/" + pumpdev + ",1/END"
			ret1 = scp_sendmsg_orch(cmd1)
			fmt.Println("DEBUG RUN WITHDRAW 22: CMD1 =", cmd1, " RET=", ret1)
			if !strings.Contains(ret1, scp_ack) && !devmode {
				fmt.Println("ERROR RUN WITHDRAW 23: BIORREATOR falha ao ligar bomba TOTEM02")
				totem[tind].Pumpstatus = false
				set_valvs_value(vpath, 0, false)
				return -1
			}
			time.Sleep(time.Duration(time_to_clean) * time.Millisecond)
			fmt.Println("WARN RUN WITHDRAW 24: Desligando bomba TOTEM02", devid)
			totem[tind].Pumpstatus = false
			cmd1 = "CMD/" + totemdev + "/PUT/" + pumpdev + ",0/END"
			ret1 = scp_sendmsg_orch(cmd1)
			fmt.Println("DEBUG RUN WITHDRAW 25: CMD1 =", cmd1, " RET=", ret1)
			if !strings.Contains(ret1, scp_ack) && !devmode {
				fmt.Println("ERROR RUN WITHDRAW 26: BIORREATOR falha ao ligar bomba TOTEM02")
				totem[tind].Pumpstatus = false
				set_valvs_value(vpath, 0, false)
				return -1
			}
			set_valvs_value(vpath, 0, false)
			board_add_message("IEnxague concluído")
		}
		bio[ind].Status = prev_status

	case scp_ibc:
		ind := get_ibc_index(devid)
		if ind < 0 {
			fmt.Println("ERROR RUN WITHDRAW 01: IBC nao existe", devid)
			return -1
		}
		prev_status := bio[ind].Status
		pathid := devid + "-" + ibc[ind].OutID
		pathstr := paths[pathid].Path
		if len(pathstr) == 0 {
			fmt.Println("ERROR RUN WITHDRAW 27: path nao existe", pathid)
			return -1
		}
		vpath := scp_splitparam(pathstr, ",")
		if !test_path(vpath, 0) {
			fmt.Println("ERROR RUN WITHDRAW 28: falha de valvula no path", pathid)
			return -1
		}
		board_add_message("CDesenvase " + devid + " para " + ibc[ind].OutID)
		var pilha []string = make([]string, 0)
		for k, p := range vpath {
			fmt.Println("step", k, p)
			if p == "END" {
				break
			}
			val, ok := valvs[p]
			if ok {
				sub := scp_splitparam(p, "/")
				dtype := get_scp_type(sub[0])
				if val == 0 {
					if !set_valv_status(dtype, sub[0], sub[1], 1) {
						fmt.Println("ERROR RUN WITHDRAW 29: nao foi possivel setar valvula", p)
						return -1
					}
				} else if val == 1 {
					fmt.Println("ERROR RUN WITHDRAW 30: valvula ja aberta", p)
					return -1
				} else {
					fmt.Println("ERROR RUN WITHDRAW 31: valvula com erro", p)
					return -1
				}
			} else {
				fmt.Println("ERROR RUN WITHDRAW 32: valvula nao existe", p)
				return -1
			}
			pilha = append([]string{p}, pilha...)
		}
		vol_ini := ibc[ind].Volume
		ibc[ind].VolumeOut = ibc[ind].Withdraw
		ibc[ind].Status = bio_unloading
		time.Sleep(scp_timewaitvalvs * time.Millisecond)
		fmt.Println("WARN RUN WITHDRAW 33: Ligando bomba", devid)
		pumpdev := biofabrica_cfg["PBF01"].Deviceaddr
		pumpport := biofabrica_cfg["PBF01"].Deviceport
		biofabrica.Pumpwithdraw = true
		cmd1 := "CMD/" + pumpdev + "/PUT/" + pumpport + ",1/END"
		ret1 := scp_sendmsg_orch(cmd1)
		fmt.Println("DEBUG RUN WITHDRAW 34: CMD1 =", cmd1, " RET=", ret1)
		if !strings.Contains(ret1, scp_ack) && !devmode {
			fmt.Println("ERROR RUN WITHDRAW 35: IBC falha ao ligar bomba desenvase")
			ibc[ind].Status = prev_status
			return -1
		}
		var vol_out int64
		var vol_bio_out_start float64
		use_volfluxo := false
		if ibc[ind].OutID == scp_out || ibc[ind].OutID == scp_drop {
			use_volfluxo = true
			var count int
			count, vol_bio_out_start = scp_get_volume(scp_biofabrica, scp_biofabrica, scp_dev_volfluxo_out)
			if count < 0 {
				vol_bio_out_start = biofabrica.VolumeOut
			}
		} else {
			vol_bio_out_start = -1
		}
		var maxtime float64
		if devmode || testmode {
			maxtime = 60
		} else {
			maxtime = scp_maxtimewithdraw
		}
		t_start := time.Now()
		for {
			vol_now := ibc[ind].Volume
			t_elapsed := time.Since(t_start).Seconds()
			vol_out = int64(vol_ini - vol_now)
			vol_bio_out_now := biofabrica.VolumeOut - vol_bio_out_start
			var vout float64
			if use_volfluxo && vol_bio_out_start >= 0 {
				vout = float64(ibc[ind].Withdraw) - vol_bio_out_now
				// fmt.Println("-------------    ibc=", float64(ibc[ind].Withdraw), "  flowout=", vol_bio_out_now, "start=", vol_bio_out_start)
			} else {
				vout = float64(ibc[ind].Withdraw) - float64(vol_out)
			}
			if vout < 0 {
				vout = 0
			}
			ibc[ind].VolumeOut = uint32(vout)
			fmt.Println("vout=", vout, ibc[ind].VolumeOut)
			if ibc[ind].Withdraw == 0 {
				break
			}
			if vol_now == 0 || (vol_now < vol_ini && vol_out >= int64(ibc[ind].Withdraw)) {
				if vol_now == 0 || vol_bio_out_start < 0 || vol_bio_out_now >= float64(ibc[ind].Withdraw) {
					fmt.Println("DEBUG RUN WITHDRAW 11: STOP Volume de desenvase atingido", vol_ini, vol_now, ibc[ind].Withdraw)
					break
				}
			}
			if t_elapsed > maxtime {
				fmt.Println("DEBUG RUN WITHDRAW 37: STOP Tempo maximo de withdraw esgotado", t_elapsed, scp_maxtimewithdraw)
				break
			}
			time.Sleep(scp_refreshwait * time.Millisecond)
		}
		if ibc[ind].Volume == 0 && ibc[ind].Vol0 != 0 {

			for i := 0; i < 300 && ibc[ind].Withdraw != 0; i++ { // 25 seg além do ZERO
				if ibc[ind].Vol0 == 0 {
					break
				}
				time.Sleep(100 * time.Millisecond)
			}
		}
		ibc[ind].Withdraw = 0
		board_add_message("IDesenvase IBC " + devid + " concluido")
		fmt.Println("WARN RUN WITHDRAW 38: Desligando bomba biofabrica", pumpdev)
		biofabrica.Pumpwithdraw = false
		cmd1 = "CMD/" + pumpdev + "/PUT/" + pumpport + ",0/END"
		ret1 = scp_sendmsg_orch(cmd1)
		fmt.Println("DEBUG RUN WITHDRAW 39: CMD1 =", cmd1, " RET=", ret1)
		set_valvs_value(pilha, 0, false)
		time.Sleep(scp_timewaitvalvs * time.Millisecond)
		// ibc[ind].Status = bio_ready
		if linewash {
			var pathclean string = ""
			dest_type := get_scp_type(ibc[ind].OutID)
			if dest_type == scp_out {
				pathclean = "TOTEM02-CLEAN9"
			} else {
				pathclean = "TOTEM02-CLEAN4"
			}
			pathstr = paths[pathclean].Path
			if len(pathstr) == 0 {
				fmt.Println("ERROR RUN WITHDRAW 40: path CLEAN linha nao existe", pathclean)
				ibc[ind].Status = prev_status
				return -1
			}
			var time_to_clean int64 = int64(paths[pathclean].Cleantime) * 1000
			vpath = scp_splitparam(pathstr, ",")
			if !test_path(vpath, 0) {
				fmt.Println("ERROR RUN WITHDRAW 41: falha de valvula no path", pathstr)
				ibc[ind].Status = prev_status
				return -1
			}
			board_add_message("ILimpando LINHA 4")
			if set_valvs_value(vpath, 1, true) < 1 {
				fmt.Println("ERROR RUN WITHDRAW 42: Falha ao abrir valvulas CLEAN linha", pathstr)
				set_valvs_value(vpath, 0, false)
			}
			time.Sleep(scp_timewaitvalvs * time.Millisecond)
			fmt.Println("WARN RUN WITHDRAW 43: Ligando bomba TOTEM02", devid)
			tind := get_totem_index("TOTEM02")
			if tind < 0 {
				fmt.Println("WARN RUN WITHDRAW 44: TOTEM02 nao encontrado", totem)
			}
			totemdev := totem_cfg["TOTEM02"].Deviceaddr
			pumpdev = totem_cfg["TOTEM02"].Pumpdev
			totem[tind].Pumpstatus = true
			cmd1 = "CMD/" + totemdev + "/PUT/" + pumpdev + ",1/END"
			ret1 = scp_sendmsg_orch(cmd1)
			fmt.Println("DEBUG RUN WITHDRAW 45: CMD1 =", cmd1, " RET=", ret1)
			if !strings.Contains(ret1, scp_ack) && !devmode {
				fmt.Println("ERROR RUN WITHDRAW 46: BIORREATOR falha ao ligar bomba TOTEM02")
				totem[tind].Pumpstatus = false
				set_valvs_value(vpath, 0, false)
				ibc[ind].Status = prev_status
				return -1
			}
			time.Sleep(time.Duration(time_to_clean/2) * time.Millisecond)
			fmt.Println("WARN RUN WITHDRAW 47: Desligando bomba TOTEM02", devid)
			totem[tind].Pumpstatus = false
			cmd1 = "CMD/" + totemdev + "/PUT/" + pumpdev + ",0/END"
			ret1 = scp_sendmsg_orch(cmd1)
			fmt.Println("DEBUG RUN WITHDRAW 48: CMD1 =", cmd1, " RET=", ret1)
			if !strings.Contains(ret1, scp_ack) && !devmode {
				fmt.Println("ERROR RUN WITHDRAW 49: BIORREATOR falha ao ligar bomba TOTEM02")
				set_valvs_value(vpath, 0, false)
				ibc[ind].Status = prev_status
				return -1
			}
			set_valvs_value(vpath, 0, false)
			time.Sleep(scp_timewaitvalvs * time.Millisecond)
			board_add_message("IEnxague concluído")
			if dest_type == scp_ibc {
				pathclean = "TOTEM02-CLEAN3"
				pathstr = paths[pathclean].Path
				if len(pathstr) == 0 {
					fmt.Println("ERROR RUN WITHDRAW 50: path CLEAN linha nao existe", pathclean)
					ibc[ind].Status = prev_status
					return -1
				}
				var time_to_clean int64 = int64(paths[pathclean].Cleantime) * 1000
				vpath = scp_splitparam(pathstr, ",")
				if !test_path(vpath, 0) {
					fmt.Println("ERROR RUN WITHDRAW 51: falha de valvula no path", pathstr)
					ibc[ind].Status = prev_status
					return -1
				}
				board_add_message("ILimpando LINHA 3")
				if set_valvs_value(vpath, 1, true) < 1 {
					fmt.Println("ERROR RUN WITHDRAW 52: Falha ao abrir valvulas CLEAN linha", pathstr)
					set_valvs_value(vpath, 0, false)
				}
				time.Sleep(scp_timewaitvalvs * time.Millisecond)
				fmt.Println("WARN RUN WITHDRAW 53: Ligando bomba TOTEM02", devid)
				tind := get_totem_index("TOTEM02")
				if tind < 0 {
					fmt.Println("WARN RUN WITHDRAW 54: TOTEM02 nao encontrado", totem)
				}
				totemdev := totem_cfg["TOTEM02"].Deviceaddr
				pumpdev = totem_cfg["TOTEM02"].Pumpdev
				totem[tind].Pumpstatus = true
				cmd1 = "CMD/" + totemdev + "/PUT/" + pumpdev + ",1/END"
				ret1 = scp_sendmsg_orch(cmd1)
				fmt.Println("DEBUG RUN WITHDRAW 55: CMD1 =", cmd1, " RET=", ret1)
				if !strings.Contains(ret1, scp_ack) && !devmode {
					fmt.Println("ERROR RUN WITHDRAW 56: BIORREATOR falha ao ligar bomba TOTEM02")
					totem[tind].Pumpstatus = false
					set_valvs_value(vpath, 0, false)
					ibc[ind].Status = prev_status
					return -1
				}
				time.Sleep(time.Duration(time_to_clean/2) * time.Millisecond)
				fmt.Println("WARN RUN WITHDRAW 57: Desligando bomba TOTEM02", devid)
				totem[tind].Pumpstatus = false
				cmd1 = "CMD/" + totemdev + "/PUT/" + pumpdev + ",0/END"
				ret1 = scp_sendmsg_orch(cmd1)
				fmt.Println("DEBUG RUN WITHDRAW 58: CMD1 =", cmd1, " RET=", ret1)
				if !strings.Contains(ret1, scp_ack) && !devmode {
					fmt.Println("ERROR RUN WITHDRAW 59: BIORREATOR falha ao ligar bomba TOTEM02")
					set_valvs_value(vpath, 0, false)
					ibc[ind].Status = prev_status
					return -1
				}
				set_valvs_value(vpath, 0, false)
				board_add_message("IEnxague concluído")
			}
		}
		ibc[ind].Status = prev_status
	default:
		fmt.Println("DEBUG RUN WITHDRAW 58: Devtype invalido", devtype, devid)
	}
	return 0
}

func scp_turn_aero(bioid string, changevalvs bool, value int, percent int) bool {
	ind := get_bio_index(bioid)
	if ind < 0 {
		fmt.Println("ERROR SCP TURN: Biorreator nao existe", bioid)
		return false
	}
	devaddr := bio_cfg[bioid].Deviceaddr
	scraddr := bio_cfg[bioid].Screenaddr
	aerorele := bio_cfg[bioid].Aero_rele
	aerodev := bio_cfg[bioid].Aero_dev
	dev_valvs := []string{bioid + "/V1", bioid + "/V2"}

	if value == scp_off {
		cmd0 := fmt.Sprintf("CMD/%s/PUT/%s,%d/END", devaddr, aerorele, value)
		ret0 := scp_sendmsg_orch(cmd0)
		fmt.Println("DEBUG SCP TURN AERO: CMD =", cmd0, "\tRET =", ret0)
		if !strings.Contains(ret0, scp_ack) && !devmode {
			fmt.Println("ERROR SCP TURN AERO:", bioid, " ERROR ao definir valor[", value, "] rele aerador ", ret0)
			if changevalvs {
				set_valvs_value(dev_valvs, 1-value, false)
			}
			return false
		}
		bio[ind].Aerator = false
		cmds := fmt.Sprintf("CMD/%s/PUT/S271,%d/END", scraddr, value)
		rets := scp_sendmsg_orch(cmds)
		fmt.Println("DEBUG SCP TURN AERO: CMD =", cmds, "\tRET =", rets)
		if !strings.Contains(rets, scp_ack) && !devmode {
			fmt.Println("ERROR SCP TURN AERO:", bioid, " ERROR ao mudar aerador na screen ", scraddr, rets)
		}

	}

	if changevalvs {
		musttest := value == 1
		if test_path(dev_valvs, 1-value) || !musttest {
			if set_valvs_value(dev_valvs, value, musttest) < 0 {
				fmt.Println("ERROR SCP TURN AERO: ERROR ao definir valor [", value, "] das valvulas", dev_valvs)
				return false
			}
		} else {
			fmt.Println("ERROR SCP TURN AERO: ERROR nas valvulas", dev_valvs)
			return false
		}
	}
	aerovalue := int(255.0 * (float32(percent) / 100.0))
	cmd1 := fmt.Sprintf("CMD/%s/PUT/%s,%d/END", devaddr, aerodev, aerovalue)
	ret1 := scp_sendmsg_orch(cmd1)
	fmt.Println("DEBUG SCP TURN AERO: CMD =", cmd1, "\tRET =", ret1)
	if !strings.Contains(ret1, scp_ack) && !devmode {
		fmt.Println("ERROR SCP TURN AERO:", bioid, " ERROR ao definir ", percent, "% aerador", ret1)
		if changevalvs {
			set_valvs_value(dev_valvs, 1-value, false)
		}
		return false
	}

	if changevalvs {
		tmax := scp_timewaitvalvs / 1000
		for i := 0; i < tmax; i++ {
			if bio[ind].MustPause || bio[ind].MustPause {
				break
			}
			time.Sleep(1000 * time.Millisecond)
		}
	}

	if value == scp_on {
		cmd2 := fmt.Sprintf("CMD/%s/PUT/%s,%d/END", devaddr, aerorele, value)
		ret2 := scp_sendmsg_orch(cmd2)
		fmt.Println("DEBUG SCP TURN AERO: CMD =", cmd2, "\tRET =", ret2)
		if !strings.Contains(ret2, scp_ack) && !devmode {
			fmt.Println("ERROR SCP TURN AERO:", bioid, " ERROR ao definir valor[", value, "] rele aerador ", ret2)
			if changevalvs {
				set_valvs_value(dev_valvs, 1-value, false)
			}
			cmdoff := fmt.Sprintf("CMD/%s/PUT/%s,%d/END", devaddr, aerodev, 0)
			ret1 = scp_sendmsg_orch(cmdoff)
			return false
		}
		bio[ind].Aerator = true
		cmds := fmt.Sprintf("CMD/%s/PUT/S271,%d/END", scraddr, value)
		rets := scp_sendmsg_orch(cmds)
		fmt.Println("DEBUG SCP TURN AERO: CMD =", cmds, "\tRET =", rets)
		if !strings.Contains(rets, scp_ack) && !devmode {
			fmt.Println("ERROR SCP TURN AERO:", bioid, " ERROR ao mudar aerador na screen ", scraddr, rets)
		}
	}
	bio[ind].AeroRatio = percent

	return true
}

func scp_turn_peris(devtype string, bioid string, perisid string, value int) bool {
	var ind, max int
	switch devtype {
	case scp_bioreactor:
		ind = get_bio_index(bioid)
		max = 5
	case scp_totem:
		ind = get_totem_index(bioid)
		max = 4
	default:
		fmt.Println("ERROR SCP TURN PERIS: Dispositivo invalido", devtype, bioid)
		return false
	}
	if ind < 0 {
		fmt.Println("ERROR SCP TURN PERIS: Dispositivo nao existe", devtype, bioid)
		return false
	}
	peris_int, err := strconv.Atoi(perisid[1:])
	if err != nil || peris_int > max {
		checkErr(err)
		fmt.Println("ERROR SCP TURN PERIS: Peristaltica invalida", bioid, perisid)
		return false
	}

	peris_dev := ""
	scrdev := ""
	devaddr := ""
	switch devtype {
	case scp_bioreactor:
		peris_dev = bio_cfg[bioid].Peris_dev[peris_int-1]
		devaddr = bio_cfg[bioid].Deviceaddr
		scrdev = bio_cfg[bioid].Screenaddr
	case scp_totem:
		peris_dev = totem_cfg[bioid].Peris_dev[peris_int-1]
		devaddr = totem_cfg[bioid].Deviceaddr
	}
	// if devtype == scp_totem && value == 1 {
	// 	if !set_valv_status(scp_totem, bioid, "V2", value) && !devmode {
	// 		fmt.Println("ERROR SCP TURN PERIS: ERRO ao abrir valvula V2 do TOTEM ", bioid)
	// 		return false
	// 	}
	// }
	cmd0 := fmt.Sprintf("CMD/%s/PUT/%s,%d/END", devaddr, peris_dev, value)
	ret0 := scp_sendmsg_orch(cmd0)
	fmt.Println("DEBUG SCP TURN PERIS: CMD =", cmd0, "\tRET =", ret0)
	if !strings.Contains(ret0, scp_ack) && !devmode {
		fmt.Println("ERROR SCP TURN PERIS:", bioid, " ERROR ao definir valor[", value, "] peristaltica ", ret0)
		return false
	}
	fmt.Println("DEBUG SCP TURN PERIS: Screen", scrdev)
	switch devtype {
	case scp_bioreactor:
		bio[ind].Perist[peris_int-1] = value
	case scp_totem:
		totem[ind].Perist[peris_int-1] = value
	}
	// if devtype == scp_totem && value == 0 {
	// 	if !set_valv_status(scp_totem, bioid, "V2", value) && !devmode {
	// 		fmt.Println("ERROR SCP TURN PERIS: ERRO ao fechar valvula V2 do TOTEM ", bioid)
	// 		return false
	// 	}
	// }
	return true
}

func scp_turn_heater(bioid string, maxtemp float32, value bool) bool {
	var ind int
	ind = get_bio_index(bioid)
	if ind < 0 {
		fmt.Println("ERROR SCP TURN HEATER: Biorreator nao existe", bioid)
		return false
	}
	devaddr := bio_cfg[bioid].Deviceaddr
	heater_dev := bio_cfg[bioid].Heater
	value_str := "0"
	if value {
		value_str = "1"
	}
	cmd0 := fmt.Sprintf("CMD/%s/PUT/%s,%s/END", devaddr, heater_dev, value_str)
	ret0 := scp_sendmsg_orch(cmd0)
	fmt.Println("DEBUG SCP TURN HEATER: CMD =", cmd0, "\tRET =", ret0)
	if !strings.Contains(ret0, scp_ack) && !devmode {
		fmt.Println("ERROR SCP TURN HEATER:", bioid, " ERROR ao definir valor[", value, "] aquecedor ", ret0)
		return false
	}
	bio[ind].Heater = value
	bio[ind].TempMax = maxtemp
	return true
}

func scp_turn_pump(devtype string, main_id string, valvs []string, value int) bool {
	var devaddr, pumpdev string
	var ind int
	scraddr := ""
	switch devtype {
	case scp_bioreactor:
		ind = get_bio_index(main_id)
		if ind < 0 {
			fmt.Println("ERROR SCP TURN PUMP: Biorreator nao existe", main_id)
			return false
		}
		devaddr = bio_cfg[main_id].Deviceaddr
		pumpdev = bio_cfg[main_id].Pump_dev
		scraddr = bio_cfg[main_id].Screenaddr

	case scp_ibc:
		ind = get_ibc_index(main_id)
		if ind < 0 {
			fmt.Println("ERROR SCP TURN PUMP: IBC nao existe", main_id)
			return false
		}
		devaddr = ibc_cfg[main_id].Deviceaddr
		pumpdev = ibc_cfg[main_id].Pump_dev

	case scp_totem:
		ind = get_totem_index(main_id)
		if ind < 0 {
			fmt.Println("ERROR SCP TURN PUMP: IBC nao existe", main_id)
			return false
		}
		devaddr = totem_cfg[main_id].Deviceaddr
		pumpdev = totem_cfg[main_id].Pumpdev

	default:
		fmt.Println("ERROR SCP TURN PUMP: Dispositivo nao suportado", devtype, main_id)
	}

	if value == scp_off {
		cmd := fmt.Sprintf("CMD/%s/PUT/%s,%d/END", devaddr, pumpdev, value)
		ret := scp_sendmsg_orch(cmd)
		fmt.Println("DEBUG SCP TURN PUMP: CMD =", cmd, "\tRET =", ret)
		if !strings.Contains(ret, scp_ack) && !devmode {
			fmt.Println("ERROR SCP TURN PUMP:", main_id, " ERROR ao definir ", value, " bomba", ret)
			if len(valvs) > 0 {
				set_valvs_value(valvs, 1-value, false)
				time.Sleep(scp_timewaitvalvs * time.Millisecond)
			}
			return false
		}
		switch devtype {
		case scp_bioreactor:
			bio[ind].Pumpstatus = false
		case scp_ibc:
			ibc[ind].Pumpstatus = false
		case scp_totem:
			totem[ind].Pumpstatus = false
		}
		if len(scraddr) > 0 {
			cmds := fmt.Sprintf("CMD/%s/PUT/S270,%d/END", scraddr, value)
			rets := scp_sendmsg_orch(cmds)
			fmt.Println("DEBUG SCP TURN AERO: CMD =", cmds, "\tRET =", rets)
			if !strings.Contains(rets, scp_ack) && !devmode {
				fmt.Println("ERROR SCP TURN AERO: ERROR ao mudar bomba na screen ", scraddr, rets)
			}
		}
	}

	musttest := value == 1
	if test_path(valvs, 1-value) || !musttest {
		if set_valvs_value(valvs, value, musttest) < 0 {
			fmt.Println("ERROR SCP TURN PUMP:", devtype, " ERROR ao definir valor [", value, "] das valvulas", valvs)
			return false
		}
	} else {
		fmt.Println("ERROR SCP TURN PUMP:", devtype, " ERROR nas valvulas", valvs)
		return false
	}

	tmax := scp_timewaitvalvs / 1000
	for i := 0; i < tmax; i++ {
		switch devtype {
		case scp_bioreactor:
			if bio[ind].MustPause || bio[ind].MustPause {
				i = tmax
			}
		}
		time.Sleep(1000 * time.Millisecond)
	}

	if value == scp_on {
		cmd := fmt.Sprintf("CMD/%s/PUT/%s,%d/END", devaddr, pumpdev, value)
		ret := scp_sendmsg_orch(cmd)
		fmt.Println("DEBUG SCP TURN PUMP: CMD =", cmd, "\tRET =", ret)
		if !strings.Contains(ret, scp_ack) && !devmode {
			fmt.Println("ERROR SCP TURN PUMP:", main_id, " ERROR ao definir ", value, " bomba", ret)
			if len(valvs) > 0 {
				set_valvs_value(valvs, 1-value, false)
				time.Sleep(scp_timewaitvalvs * time.Millisecond)
			}
			return false
		}
		switch devtype {
		case scp_bioreactor:
			bio[ind].Pumpstatus = true
		case scp_ibc:
			ibc[ind].Pumpstatus = true
		case scp_totem:
			totem[ind].Pumpstatus = true
		}
		if len(scraddr) > 0 {
			cmds := fmt.Sprintf("CMD/%s/PUT/S270,%d/END", scraddr, value)
			rets := scp_sendmsg_orch(cmds)
			fmt.Println("DEBUG SCP TURN AERO: CMD =", cmds, "\tRET =", rets)
			if !strings.Contains(rets, scp_ack) && !devmode {
				fmt.Println("ERROR SCP TURN AERO: ERROR ao mudar bomba na screen ", scraddr, rets)
			}
		}
	}
	return true
}

func pop_first_sched(bioid string, remove bool) Scheditem {
	var ret Scheditem
	for k, s := range schedule {
		if s.Bioid == bioid {
			ret = Scheditem{s.Bioid, s.Seq, s.OrgCode}
			if remove {
				if k > 0 {
					if k < len(schedule)-1 {
						schedule = append(schedule[:k], schedule[k+1:]...)
					} else {
						schedule = schedule[:k]
					}
				} else {
					schedule = schedule[k+1:]
				}
			}
			return ret
		}
	}
	ret = Scheditem{}
	return ret
}

func pop_first_job(devtype string, main_id string, remove bool) string {
	switch devtype {
	case scp_bioreactor:
		ind := get_bio_index(main_id)
		if ind < 0 {
			fmt.Println("ERROR POP FIRST WORK: Biorreator nao existe", main_id)
			return ""
		}
		n := len(bio[ind].Queue)
		ret := ""
		if n > 0 {
			ret = bio[ind].Queue[0]
			if remove {
				bio[ind].Queue = bio[ind].Queue[1:]
			}
		}
		return ret
	case scp_ibc:
		ind := get_ibc_index(main_id)
		if ind < 0 {
			fmt.Println("ERROR POP FIRST WORK: IBC nao existe", main_id)
			return ""
		}
		n := len(ibc[ind].Queue)
		ret := ""
		if n > 0 {
			ret = ibc[ind].Queue[0]
			if remove {
				ibc[ind].Queue = ibc[ind].Queue[1:]
			}
		}
		return ret
	}
	return ""
}

func pop_first_undojob(devtype string, main_id string, remove bool) string {
	switch devtype {
	case scp_bioreactor:
		ind := get_bio_index(main_id)
		if ind < 0 {
			fmt.Println("ERROR POP FIRST WORK: Biorreator nao existe", main_id)
			return ""
		}
		n := len(bio[ind].UndoQueue)
		ret := ""
		if n > 0 {
			ret = bio[ind].UndoQueue[0]
			if remove {
				bio[ind].UndoQueue = bio[ind].UndoQueue[1:]
			}
		}
		return ret
	case scp_ibc:
		ind := get_ibc_index(main_id)
		if ind < 0 {
			fmt.Println("ERROR POP FIRST WORK: IBC nao existe", main_id)
			return ""
		}
		n := len(ibc[ind].UndoQueue)
		ret := ""
		if n > 0 {
			ret = ibc[ind].UndoQueue[0]
			if remove {
				ibc[ind].UndoQueue = ibc[ind].UndoQueue[1:]
			}
		}
		return ret
	}
	return ""
}

func scp_adjust_ph(bioid string, ph float32) { //  ATENCAO - MUDAR PH
	ind := get_bio_index(bioid)
	fmt.Println("DEBUG SCP ADJUST PH: Ajustando PH", bioid, bio[ind].PH, ph)
	valvs := []string{bioid + "/V4", bioid + "/V6"}
	if !scp_turn_pump(scp_bioreactor, bioid, valvs, 1) {
		fmt.Println("ERROR SCP ADJUST PH: Falha ao abrir valvulas e ligar bomba", bioid, valvs)
		return
	}
	if bio[ind].MustPause || bio[ind].MustStop {
		return
	}

	if bio[ind].PH > ph {
		if !scp_turn_peris(scp_bioreactor, bioid, "P1", 1) {
			fmt.Println("ERROR SCP ADJUST PH: Falha ao ligar Peristaltica P1", bioid)
		} else {
			time.Sleep(scp_timephwait * time.Millisecond)
			if !scp_turn_peris(scp_bioreactor, bioid, "P1", 0) {
				fmt.Println("ERROR SCP ADJUST PH: Falha ao desligar Peristaltica P1", bioid)
			}
		}
	} else {
		if !scp_turn_peris(scp_bioreactor, bioid, "P2", 1) {
			fmt.Println("ERROR SCP ADJUST PH: Falha ao ligar Peristaltica P2", bioid)
		} else {
			time.Sleep(scp_timephwait * time.Millisecond)
			if !scp_turn_peris(scp_bioreactor, bioid, "P2", 0) {
				fmt.Println("ERROR SCP ADJUST PH: Falha ao desligar Peristaltica P2", bioid)
			}
		}
	}
	time.Sleep(20 * time.Second)

	if !scp_turn_pump(scp_bioreactor, bioid, valvs, 0) {
		fmt.Println("ERROR SCP ADJUST PH: Falha ao fechar valvulas e desligar bomba", bioid, valvs)
		return
	}
}

func scp_adjust_temperature(bioid string, temp float32) {
	ind := get_bio_index(bioid)
	fmt.Println("DEBUG SCP ADJUST TEMP: Ajustando Temperatura", bioid, bio[ind].Temperature, temp)
	valvs := []string{bioid + "/V4", bioid + "/V6"}
	if !scp_turn_pump(scp_bioreactor, bioid, valvs, 1) {
		fmt.Println("ERROR SCP ADJUST TEMP: Falha ao abrir valvulas e ligar bomba", bioid, valvs)
		return
	}
	for n := 0; n < 3; n++ {
		if bio[ind].MustPause || bio[ind].MustStop {
			break
		}
		if temp*0.95 <= bio[ind].Temperature && bio[ind].Temperature <= temp*1.05 {
			break
		} else if bio[ind].Temperature < temp {
			if !scp_turn_heater(bioid, temp, true) {
				fmt.Println("ERROR SCP ADJUST TEMP: Falha ao ligar aquecedor", bioid)
			} else {
				time.Sleep(scp_timetempwait * time.Millisecond)
				if !scp_turn_heater(bioid, temp, false) {
					fmt.Println("ERROR SCP ADJUST TEMP: Falha ao desligar aquecedor", bioid)
				}
			}
		} else {
			fmt.Println("WARN SCP ADJUST TEMP: Temperatura acima do limite em", bioid, bio[ind].Temperature, "/", temp)
			break
		}
		time.Sleep(scp_timetempwait * time.Millisecond)
	}
	if !scp_turn_pump(scp_bioreactor, bioid, valvs, 0) {
		fmt.Println("ERROR SCP ADJUST TEMP: Falha ao fechar valvulas e desligar bomba", bioid, valvs)
		return
	}
}

func scp_adjust_foam(bioid string) {
	ind := get_bio_index(bioid)
	fmt.Println("DEBUG SCP ADJUST FOAM: Ajustando Esmpuma do Biorreator", bioid)
	if bio[ind].MustPause || bio[ind].MustStop {
		return
	}
	if !scp_turn_peris(scp_bioreactor, bioid, "P5", 1) {
		fmt.Println("ERROR SCP ADJUST FOAM: Falha ao LIGAR peristaltica 5", bioid)
	} else {
		for i := 0; i < 70; i++ {
			if bio[ind].MustPause || bio[ind].MustStop {
				break
			}
			time.Sleep(100 * time.Millisecond)
		}
		if !scp_turn_peris(scp_bioreactor, bioid, "P5", 0) {
			fmt.Println("ERROR SCP ADJUST FOAM: Falha ao DESLIGAR peristaltica 5", bioid)
		}
	}
}

func scp_adjust_aero(bioid string, aero int) bool {
	fmt.Println("DEBUG SCP ADJUST AERO: Ajustando Aerador do Biorreator", bioid, " para", aero)
	ret := scp_turn_aero(bioid, false, scp_on, aero)
	return ret
}

func scp_grow_bio(bioid string) bool {
	ind := get_bio_index(bioid)
	if ind < 0 {
		fmt.Println("ERROR SCP GROW BIO: Biorreator nao encontrado", bioid)
		return false
	}
	orgid := bio[ind].OrgCode
	org, ok := organs[orgid]
	if !ok {
		fmt.Println("ERROR SCP GROW BIO: Organismo nao encontrado", orgid)
		return false
	}
	fmt.Println("DEBUG SCP GROW BIO: Iniciando cultivo de", org.Orgname, " no Biorreator", bioid, " tempo=", org.Timetotal)
	ttotal := float64(org.Timetotal * 60)
	if devmode || testmode {
		ttotal = scp_timeoutdefault / 60
	}
	time.Sleep(5 * time.Second)
	vol_start := bio[ind].Volume
	pday := -1
	var minph, maxph, worktemp float64
	var aero int
	aero_prev := -1
	t_start := time.Now()
	t_start_ph := time.Now()

	if control_foam {
		scp_adjust_foam(bioid)
	}

	ncontrol_foam := 1
	for {
		t_elapsed := time.Since(t_start).Minutes()
		if t_elapsed >= ttotal {
			break
		}
		t_day := int(t_elapsed / (60 * 24))
		if t_day != pday {
			var err error
			if t_day > len(org.PH) {
				fmt.Println("ERROR SCP GROW BIO: Dia de cultivo invalido", t_day, org)
			} else {
				vals := scp_splitparam(org.PH[t_day], "-")
				minph, err = strconv.ParseFloat(vals[0], 32)
				if err != nil {
					checkErr(err)
					fmt.Println("ERROR SCP GROW BIO: Valor de PH invalido", vals, org)
				}
				maxph, err = strconv.ParseFloat(vals[1], 32)
				if err != nil {
					checkErr(err)
					fmt.Println("ERROR SCP GROW BIO: Valor de PH invalido", vals, org)
				}
				aero = org.Aero[t_day]
				fmt.Println("\n\nDEBUG SCP GROW BIO: Day", t_day, " - Parametros de PH", minph, maxph)
				worktemp = 28
				pday = t_day
			}
			if control_foam {
				scp_adjust_foam(bioid)
				ncontrol_foam++
			}
		}
		if bio[ind].MustPause || bio[ind].MustStop {
			break
		}
		if aero != aero_prev {
			if !scp_adjust_aero(bioid, aero) {
				fmt.Println("ERROR SCP GROW BIO: Falha ao ajustar aeracao", bioid, aero)
			}
			aero_prev = aero
		}
		if control_foam && ncontrol_foam < bio_max_foam && float64(bio[ind].Volume) > float64(vol_start)*1.05 {
			scp_adjust_foam(bioid)
			ncontrol_foam++
		}
		if bio[ind].MustPause || bio[ind].MustStop {
			break
		}
		t_elapsed_ph := time.Since(t_start_ph).Minutes()
		if control_ph && t_elapsed_ph >= 10 {
			ph_tmp := scp_get_ph(bioid)
			if ph_tmp > 0 {
				bio[ind].PH = float32(ph_tmp)
				if bio[ind].PH < float32(minph-bio_deltaph) {
					scp_adjust_ph(bioid, float32(minph))
				} else if bio[ind].PH > float32(maxph+bio_deltaph) {
					scp_adjust_ph(bioid, float32(maxph))
				}
				t_start_ph = time.Now()
			}
		}
		if bio[ind].MustPause || bio[ind].MustStop {
			break
		}
		if control_temp && bio[ind].Temperature < float32(worktemp*(1-bio_deltatemp)) {
			fmt.Println("WARN SCP GROW BIO: Ajustando temperatura", bioid, bio[ind].Temperature)
			scp_adjust_temperature(bioid, float32(worktemp))
		}
		if bio[ind].MustPause || bio[ind].MustStop {
			break
		}
		time.Sleep(scp_timegrowwait * time.Millisecond)
	}
	return true
}

func scp_circulate(devtype string, main_id string, period int) {
	var valvs []string
	var ind int
	switch devtype {
	case scp_bioreactor:
		ind = get_bio_index(main_id)
		if ind < 0 {
			fmt.Println("ERROR SCP CIRCULATE: Biorreator nao encontrado", main_id)
			return
		}
		valvs = []string{main_id + "/V4", main_id + "/V6"}
	case scp_ibc:
		ind = get_ibc_index(main_id)
		if ind < 0 {
			fmt.Println("ERROR SCP CIRCULATE: IBC nao encontrado", main_id)
			return
		}
		valvs = []string{main_id + "/V1"}
	default:
		fmt.Println("ERROR SCP CIRCULATE: Tipo de dispositivo invalido", devtype, main_id)
		return
	}
	if !scp_turn_pump(devtype, main_id, valvs, 1) {
		fmt.Println("ERROR SCP CIRCULATE: Nao foi possivel ligar circulacao em ", main_id)
		return
	}
	n := 0
	stop := false
	for !stop {
		time.Sleep(1 * time.Second)
		n++
		if period == 0 {
			switch devtype {
			case scp_bioreactor:
				if bio[ind].Status != bio_circulate {
					stop = true
				}
			case scp_ibc:
				if ibc[ind].Status != bio_circulate {
					stop = true
				}
			}
		} else {
			if n >= period*60 {
				stop = true
			}
		}
	}
	if !scp_turn_pump(devtype, main_id, valvs, 0) {
		fmt.Println("ERROR SCP CIRCULATE: Nao foi possivel desligar circulacao em ", main_id)
	}
	switch devtype {
	case scp_bioreactor:
		bio[ind].Status = bio[ind].LastStatus
	case scp_ibc:
		ibc[ind].Status = ibc[ind].LastStatus
	}
}

func scp_run_job_bio(bioid string, job string) bool {
	if devmode {
		fmt.Println("\n\nSCP RUN JOB SIMULANDO EXECUCAO", bioid, job)
	} else {
		fmt.Println("\n\nSCP RUN JOB EXECUTANDO", bioid, job)
	}
	bio_add_message(bioid, "C"+job)
	ind := get_bio_index(bioid)
	if ind < 0 {
		fmt.Println("ERROR SCP RUN JOB: Biorreator nao existe", bioid)
		return false
	}
	params := scp_splitparam(job, "/")
	subpars := []string{}
	if len(params) > 1 {
		subpars = scp_splitparam(params[1], ",")
	}
	switch params[0] {
	case scp_job_org:
		var orgcode string
		if len(subpars) > 0 {
			orgcode = subpars[0]
			if len(organs[orgcode].Orgname) > 0 {
				bio[ind].OrgCode = subpars[0]
				bio[ind].Organism = organs[orgcode].Orgname
				bio[ind].Timetotal = [2]int{organs[orgcode].Timetotal, 0}
				bio[ind].Timeleft = [2]int{organs[orgcode].Timetotal, 0}
				go scp_update_screen_times(bioid)
			} else {
				fmt.Println("ERROR SCP RUN JOB: Organismo nao existe", params)
				return false
			}
		} else {
			fmt.Println("ERROR SCP RUN JOB: Falta parametros em", scp_job_org, params)
			return false
		}
		board_add_message("IIniciando Cultivo " + organs[orgcode].Orgname + " no " + bioid)

	case scp_job_set:
		if len(subpars) > 1 {
			flag := subpars[0]
			switch flag {
			case scp_par_status:
				biostatus := subpars[1]
				bio[ind].Status = biostatus
			case scp_par_step:
				biostep_str := subpars[1]
				biostep, _ := strconv.Atoi(biostep_str)
				bio[ind].Step[0] = biostep
				go scp_update_screen_times(bioid)
			case scp_par_maxstep:
				biomaxstep_str := subpars[1]
				biomaxstep, _ := strconv.Atoi(biomaxstep_str)
				bio[ind].Step[1] = biomaxstep
			case scp_par_totaltime:
				biototaltime_str := subpars[1]
				if biototaltime_str == "DEFAULT" {
					orgtime := organs[bio[ind].OrgCode].Timetotal * 60
					fmt.Println("DEBUG SCP RUN JOB: Definindo Tempo DEFAULT ", bioid, bio[ind].OrgCode, orgtime)
					// if orgtime > 0 {
					// 	bio[ind].Timetotal[0] = int(orgtime / 60)
					// 	bio[ind].Timeleft[0] = int(orgtime / 60)
					// 	bio[ind].Timetotal[1] = int(orgtime % 60)
					// 	bio[ind].Timeleft[1] = int(orgtime % 60)
					// } else {
					// 	bio[ind].Timetotal[0] = 0
					// 	bio[ind].Timetotal[1] = 0
					// 	bio[ind].Timeleft[0] = 0
					// 	bio[ind].Timeleft[1] = 0
					// 	fmt.Println("ERROR SCP RUN JOB: Tempo DEFAULT invalido", flag, params, bio[ind].OrgCode, orgtime)
					// }
				} else {
					biototaltime, err := strconv.Atoi(biototaltime_str)
					if err == nil {
						bio[ind].Timetotal[0] = int(biototaltime / 60)
						bio[ind].Timeleft[0] = int(biototaltime / 60)
						bio[ind].Timetotal[1] = int(biototaltime % 60)
						bio[ind].Timeleft[1] = int(biototaltime % 60)
					} else {
						bio[ind].Timetotal[0] = 0
						bio[ind].Timetotal[1] = 0
						bio[ind].Timeleft[0] = 0
						bio[ind].Timeleft[1] = 0
						fmt.Println("ERROR SCP RUN JOB: Tempo invalido", flag, params, biototaltime_str)
						checkErr(err)
					}
				}
			default:
				fmt.Println("ERROR SCP RUN JOB: Parametro invalido em", scp_job_set, flag, params)
				return false
			}
		} else {
			fmt.Println("ERROR SCP RUN JOB: Falta parametros em", scp_job_set, params)
			return false
		}

	case scp_job_commit:
		if len(subpars) > 1 {
			obj := subpars[0]
			if obj == "ALL" {
				bio[ind].MustOffQueue = []string{}
				fmt.Println("WARN SCP RUN JOB: COMMIT ALL executado", subpars)
			}
		}
		bio[ind].UndoQueue = []string{}
		bio[ind].RedoQueue = []string{}

	case scp_job_run:
		if len(subpars) > 0 {
			cmd := subpars[0]
			switch cmd {
			case scp_par_grow:
				scp_grow_bio(bioid)

			case scp_par_cip:
				bio[ind].ShowVol = false
				qini := []string{bio[ind].Queue[0]}
				qini = append(qini, cipbio...)
				bio[ind].Queue = append(qini, bio[ind].Queue[1:]...)
				fmt.Println("\n\nTRUQUE CIP:", bio[ind].Queue)
				board_add_message("IExecutando CIP no biorreator " + bioid)
				return true

			case scp_par_withdraw:
				bio[ind].Withdraw = bio[ind].Volume
				if len(subpars) > 2 {
					outid := subpars[1]
					bio[ind].OutID = outid
				}
				board_add_message("IDesenvase Automático do biorreator " + bioid + " para " + bio[ind].OutID)
				if scp_run_withdraw(scp_bioreactor, bioid, false, true) < 0 {
					return false
				} else {
					bio[ind].Organism = ""
					bio[ind].Withdraw = 0
					bio[ind].Volume = 0 // verificar
				}

			}
		} else {
			fmt.Println("ERROR SCP RUN JOB: Falta parametros em", scp_job_run, params)
			return false
		}

	case scp_job_ask:
		if len(subpars) > 0 {
			msg := subpars[0]
			scraddr := bio_cfg[bioid].Screenaddr
			var cmd1 string = ""
			var msgask string = ""
			scrmain := fmt.Sprintf("CMD/%s/PUT/S200,1/END", scraddr)
			switch msg {
			case scp_msg_cloro:
				cmd1 = fmt.Sprintf("CMD/%s/PUT/S400,2/END", scraddr)
				msgask = "CLORO"
			case scp_msg_meio:
				cmd1 = fmt.Sprintf("CMD/%s/PUT/S400,3/END", scraddr)
				msgask = "MEIO"
			case scp_msg_inoculo:
				cmd1 = fmt.Sprintf("CMD/%s/PUT/S400,1/END", scraddr)
				msgask = "INOCULO"
			default:
				fmt.Println("ERROR SCP RUN JOB:", bioid, " ASK invalido", subpars)
				return false
			}
			ret1 := scp_sendmsg_orch(cmd1)
			fmt.Println("DEBUG SCP RUN JOB:: CMD =", cmd1, "\tRET =", ret1)
			if !strings.Contains(ret1, scp_ack) && !devmode {
				fmt.Println("ERROR SCP RUN JOB:", bioid, " ERROR ao enviar PUT screen", scraddr, ret1)
				return false
			}
			cmd2 := fmt.Sprintf("CMD/%s/GET/S451/END", scraddr)
			board_add_message("ABiorreator " + bioid + " aguardando " + msgask)
			t_start := time.Now()
			for {
				ret2 := scp_sendmsg_orch(cmd2)
				// fmt.Println("DEBUG SCP RUN JOB:: CMD =", cmd2, "\tRET =", ret2)
				if !strings.Contains(ret2, scp_ack) && !devmode {
					fmt.Println("ERROR SCP RUN JOB:", bioid, " ERRO ao envirar GET screen", scraddr, ret2)
					scp_sendmsg_orch(scrmain)
					return false
				}
				data := scp_splitparam(ret2, "/")
				if len(data) > 1 {
					if data[1] == "1" {
						break
					}
				}
				if bio[ind].MustPause || bio[ind].MustStop {
					return false
				}
				t_elapsed := time.Since(t_start).Seconds()
				if t_elapsed > scp_timeoutdefault {
					fmt.Println("DEBUG SCP RUN JOB: Tempo maximo de ASK esgotado", bioid, t_elapsed, scp_maxtimewithdraw)
					if !devmode {
						scp_sendmsg_orch(scrmain)
						return testmode
					}
					break
				}
				time.Sleep(scp_refreshwait * time.Millisecond)
			}
			scp_sendmsg_orch(scrmain)
		} else {
			fmt.Println("ERROR SCP RUN JOB: Falta parametros em", scp_job_org, params)
			return false
		}

	case scp_job_done:
		if bio[ind].Volume == 0 {
			bio[ind].Status = bio_empty
		} else {
			bio[ind].Status = bio_ready
		}
		bio[ind].ShowVol = true
		board_add_message("CProcesso concluído no " + bioid)
		bio[ind].UndoQueue = []string{}
		bio[ind].RedoQueue = []string{}
		bio[ind].MustOffQueue = []string{}
		bio[ind].Step = [2]int{0, 0}
		bio[ind].Timetotal = [2]int{0, 0}
		bio[ind].Timeleft = [2]int{0, 0}
		go scp_update_screen_times(bioid)
		return true

	case scp_job_wait:
		var time_int uint64
		var err error
		if len(subpars) > 1 {
			switch subpars[0] {
			case scp_par_time:
				time_str := subpars[1]
				time_int, err = strconv.ParseUint(time_str, 10, 32)
				if devmode || testmode {
					if time_int > uint64(scp_timeoutdefault) {
						time_int = uint64(scp_timeoutdefault)
					}
				}
				if err != nil {
					fmt.Println("ERROR SCP RUN JOB: WAIT TIME invalido", time_str, params)
					return false
				}
				var time_dur time.Duration
				if !testmode {
					time_dur = time.Duration(time_int)
				} else {
					time_dur = time.Duration(60)
				}
				fmt.Println("DEBUG SCP RUN JOB: WAIT de", time_dur.Seconds(), "segundos")
				var n time.Duration
				for n = 0; n < time_dur; n++ {
					if bio[ind].MustPause || bio[ind].MustStop {
						return false
					}
					time.Sleep(1000 * time.Millisecond)
				}

			case scp_par_volume:
				var vol_max uint64
				var err error
				vol_str := subpars[1]
				vol_max, err = strconv.ParseUint(vol_str, 10, 32)
				if err != nil {
					fmt.Println("ERROR SCP RUN JOB: WAIT VOLUME invalido", vol_str, params)
					return false
				}
				if vol_max > uint64(bio_cfg[bioid].Maxvolume) {
					fmt.Println("ERROR SCP RUN JOB: WAIT VOLUME maior do que maximo do Biorreator", vol_max, bioid, bio_cfg[bioid].Maxvolume)
					return false
				}

				time_max := scp_timeoutdefault
				time_min := 0
				par_time := false
				if len(subpars) > 3 {
					time_max, err = strconv.Atoi(subpars[2])
					if err != nil {
						time_max = scp_timeoutdefault
						checkErr(err)
						fmt.Println("ERROR SCP RUN JOB: WAIT VOLUME,TEMPO MAX invalido", vol_str, params)
					} else {
						fmt.Println("DEBUG SCP RUN JOB: WAIT VOLUME e/ou TEMPO MAX", vol_str, time_max)
						par_time = true
					}
				}
				if len(subpars) > 4 {
					time_min, err = strconv.Atoi(subpars[3])
					if err != nil {
						time_min = 0
						checkErr(err)
						fmt.Println("ERROR SCP RUN JOB: WAIT VOLUME,TEMPO MIN invalido", vol_str, params)
					} else {
						fmt.Println("DEBUG SCP RUN JOB: WAIT VOLUME e TEMPO MIN", vol_str, time_min)
						par_time = true
					}
				}
				t_start := time.Now()
				for {
					vol_now := uint64(bio[ind].Volume)
					t_elapsed := time.Since(t_start).Seconds()
					if vol_now >= vol_max && t_elapsed >= float64(time_min) {
						break
					}
					if bio[ind].MustPause || bio[ind].MustStop {
						return false
					}
					if t_elapsed > float64(time_max) {
						fmt.Println("DEBUG SCP RUN JOB: Tempo maximo de withdraw esgotado", t_elapsed, scp_maxtimewithdraw)
						if !devmode && !par_time {
							return testmode
						}
						break
					}
					time.Sleep(scp_refreshwait * time.Millisecond)
				}
			}
		} else {
			fmt.Println("ERROR SCP RUN JOB: Falta parametros em", scp_job_org, params)
			return false
		}

	case scp_job_on:
		if len(subpars) > 1 {
			device := subpars[0]
			switch device {
			case scp_dev_aero:
				perc_str := subpars[1]
				perc_int, err := strconv.Atoi(perc_str)
				if err != nil {
					checkErr(err)
					fmt.Println("ERROR SCP RUN JOB: Parametros invalido em", scp_job_on, params)
					return false
				}
				if !scp_turn_aero(bioid, true, 1, perc_int) {
					fmt.Println("ERROR SCP RUN JOB: ERROR ao ligar aerador em", bioid)
					return false
				}
			case scp_dev_pump:
				valvs := []string{}
				for k := 1; k < len(subpars) && subpars[k] != "END"; k++ {
					v := bioid + "/" + subpars[k]
					valvs = append(valvs, v)
				}
				if !scp_turn_pump(scp_bioreactor, bioid, valvs, 1) {
					fmt.Println("ERROR SCP RUN JOB: ERROR ao ligar bomba em", bioid, valvs)
					return false
				}
			case scp_dev_peris:
				peris_str := subpars[1]
				if !scp_turn_peris(scp_bioreactor, bioid, peris_str, 1) {
					fmt.Println("ERROR SCP RUN JOB: ERROR ao ligar peristaltica em", bioid, peris_str)
					return false
				}
			case scp_par_heater:
				temp_str := subpars[1]
				temp_int, err := strconv.Atoi(temp_str)
				if err != nil {
					checkErr(err)
					fmt.Println("ERROR SCP RUN JOB: Parametro de temperatura invalido", bioid, temp_str, subpars)
					return false
				}
				if !scp_turn_heater(bioid, float32(temp_int), true) {
					fmt.Println("ERROR SCP RUN JOB: ERROR ao ligar aquecedor em", bioid)
				}

			case scp_dev_water:
				totem := subpars[1]
				totem_ind := get_totem_index(totem)
				if totem_ind < 0 {
					fmt.Println("ERROR SCP RUN JOB: Totem nao existe", totem)
					return false
				}
				use_spball := false
				if len(subpars) > 2 && subpars[2] != "END" {
					if subpars[2] == scp_dev_sprayball {
						use_spball = true
					} else {
						fmt.Println("ERROR SCP RUN JOB: ON Parametro invalido", bioid, subpars)
						return false
					}
				}
				pathid := totem + "-" + bioid
				pathstr := paths[pathid].Path
				if len(pathstr) == 0 {
					fmt.Println("ERROR SCP RUN JOB: path nao existe", pathid)
					return false
				}
				var npath string
				if use_spball {
					npath = strings.Replace(pathstr, "/V4", "/V8", -1)
					spball_valv := bioid + "/V3"
					npath = spball_valv + "," + npath
				} else {
					npath = pathstr
				}
				fmt.Println("npath=", npath)
				vpath := scp_splitparam(npath, ",")
				watervalv := totem + "/V1"
				n := len(vpath)
				vpath = append(vpath[:n-1], watervalv)
				vpath = append(vpath, "END")
				fmt.Println("DEBUG", vpath)
				if !scp_turn_pump(scp_totem, totem, vpath, 1) {
					fmt.Println("ERROR SCP RUN JOB: ERROR ao ligar bomba em", bioid, valvs)
					return false
				}
			}
		} else {
			fmt.Println("ERROR SCP RUN JOB: Falta parametros em", scp_job_on, params)
			return false
		}
	case scp_job_off:
		if len(subpars) > 0 {
			device := subpars[0]
			switch device {
			case scp_dev_aero:
				if !scp_turn_aero(bioid, true, 0, 0) {
					fmt.Println("ERROR SCP RUN JOB: ERROR ao desligar aerador em", bioid)
					return false
				}
			case scp_dev_pump:
				valvs := []string{}
				for k := 1; k < len(subpars) && subpars[k] != "END"; k++ {
					v := bioid + "/" + subpars[k]
					valvs = append(valvs, v)
				}
				if !scp_turn_pump(scp_bioreactor, bioid, valvs, 0) {
					fmt.Println("ERROR SCP RUN JOB: ERROR ao desligar bomba em", bioid, valvs)
					return false
				}
			case scp_dev_peris:
				peris_str := subpars[1]
				if !scp_turn_peris(scp_bioreactor, bioid, peris_str, 0) {
					fmt.Println("ERROR SCP RUN JOB: ERROR ao ligar peristaltica em", bioid, peris_str)
					return false
				}
			case scp_par_heater:
				if !scp_turn_heater(bioid, float32(0), false) {
					fmt.Println("ERROR SCP RUN JOB: ERROR ao ligar aquecedor em", bioid)
				}
			case scp_dev_water:
				totem := subpars[1]
				totem_ind := get_totem_index(totem)
				if totem_ind < 0 {
					fmt.Println("ERROR SCP RUN JOB: Totem nao existe", totem)
					return false
				}
				use_spball := false
				if len(subpars) > 2 && subpars[2] != "END" {
					if subpars[2] == scp_dev_sprayball {
						use_spball = true
					} else {
						fmt.Println("ERROR SCP RUN JOB: ON Parametro invalido", bioid, subpars)
						return false
					}
				}
				pathid := totem + "-" + bioid
				pathstr := paths[pathid].Path
				if len(pathstr) == 0 {
					fmt.Println("ERROR SCP RUN JOB: path nao existe", pathid)
					return false
				}
				var npath string
				if use_spball {
					npath = strings.Replace(pathstr, "/V4", "/V8", -1)
					spball_valv := bioid + "/V3"
					npath = spball_valv + "," + npath
				} else {
					npath = pathstr
				}
				fmt.Println("npath=", npath)
				vpath := scp_splitparam(npath, ",")
				watervalv := totem + "/V1"
				n := len(vpath)
				vpath = append(vpath[:n-1], watervalv)
				vpath = append(vpath, "END")
				if !scp_turn_pump(scp_totem, totem, vpath, 0) {
					fmt.Println("ERROR SCP RUN JOB: ERROR ao ligar bomba em", bioid, valvs)
					return false
				}
			}
		} else {
			fmt.Println("ERROR SCP RUN JOB: Falta parametros em", scp_job_off, params)
			return false
		}
	default:
		fmt.Println("ERROR SCP RUN JOB: JOB invalido", bioid, job, params)
	}
	time.Sleep(500 * time.Millisecond)
	return true
}

func scp_run_job_ibc(ibcid string, job string) bool {
	if devmode {
		fmt.Println("\n\nSCP RUN JOB SIMULANDO EXECUCAO", ibcid, job)
	} else {
		fmt.Println("\n\nSCP RUN JOB EXECUTANDO", ibcid, job)
	}
	ind := get_ibc_index(ibcid)
	if ind < 0 {
		fmt.Println("ERROR SCP RUN JOB: IBC nao existe", ibcid)
		return false
	}
	params := scp_splitparam(job, "/")
	subpars := []string{}
	if len(params) > 1 {
		subpars = scp_splitparam(params[1], ",")
	}
	switch params[0] {
	case scp_job_org:
		var orgcode string
		if len(subpars) > 0 {
			orgcode = subpars[0]
			if len(organs[orgcode].Orgname) > 0 {
				ibc[ind].OrgCode = subpars[0]
				ibc[ind].Organism = organs[orgcode].Orgname
				ibc[ind].Timetotal = [2]int{0, 0}
			} else {
				fmt.Println("ERROR SCP RUN JOB: Organismo nao existe", params)
				return false
			}
		} else {
			fmt.Println("ERROR SCP RUN JOB: Falta parametros em", scp_job_org, params)
			return false
		}
		// board_add_message("IIniciando Cultivo " + organs[orgcode].Orgname + " no " + ibcid)

	case scp_job_set:
		if len(subpars) > 1 {
			flag := subpars[0]
			switch flag {
			case scp_par_status:
				biostatus := subpars[1]
				ibc[ind].Status = biostatus
			case scp_par_step:
				biostep_str := subpars[1]
				biostep, _ := strconv.Atoi(biostep_str)
				ibc[ind].Step[0] = biostep
			case scp_par_maxstep:
				biomaxstep_str := subpars[1]
				biomaxstep, _ := strconv.Atoi(biomaxstep_str)
				ibc[ind].Step[1] = biomaxstep
			case scp_par_totaltime:
				biototaltime_str := subpars[1]
				if biototaltime_str == "DEFAULT" {
					fmt.Println("ERROR SCP RUN JOB: Tempo DEFAULT nao suportado", flag, params)
				} else {
					biototaltime, err := strconv.Atoi(biototaltime_str)
					if err == nil {
						ibc[ind].Timetotal[0] = int(biototaltime / 60)
						ibc[ind].Timetotal[1] = int(biototaltime % 60)
					} else {
						ibc[ind].Timetotal[0] = 0
						ibc[ind].Timetotal[1] = 0
						fmt.Println("ERROR SCP RUN JOB: Tempo invalido", flag, params, biototaltime_str)
						checkErr(err)
					}
				}
			default:
				fmt.Println("ERROR SCP RUN JOB: Parametro invalido em", scp_job_set, flag, params)
				return false
			}
		} else {
			fmt.Println("ERROR SCP RUN JOB: Falta parametros em", scp_job_set, params)
			return false
		}

	case scp_job_commit:
		if len(subpars) > 1 {
			obj := subpars[0]
			if obj == "ALL" {
				ibc[ind].MustOffQueue = []string{}
				fmt.Println("WARN SCP RUN JOB: COMMIT ALL executado", subpars)
			}
		}
		ibc[ind].UndoQueue = []string{}
		ibc[ind].RedoQueue = []string{}

	case scp_job_run:
		if len(subpars) > 0 {
			cmd := subpars[0]
			switch cmd {

			case scp_par_cip:
				ibc[ind].ShowVol = false
				qini := []string{ibc[ind].Queue[0]}
				qini = append(qini, cipibc...)
				ibc[ind].Queue = append(qini, ibc[ind].Queue[1:]...)
				fmt.Println("\n\nTRUQUE CIP:", ibc[ind].Queue)
				board_add_message("IExecutando CIP no IBC " + ibcid)
				return true

			case scp_par_withdraw:
				ibc[ind].Withdraw = ibc[ind].Volume
				if len(subpars) > 2 {
					outid := subpars[1]
					ibc[ind].OutID = outid
				}
				board_add_message("IDesenvase Automático do IBC " + ibcid + " para " + ibc[ind].OutID)
				if scp_run_withdraw(scp_ibc, ibcid, false, true) < 0 {
					fmt.Println("ERROR SCP RUN JOB IBC: Falha ao fazer o desenvase do IBC", ibc[ind].IBCID)
					return false
				} else {
					ibc[ind].Organism = ""
					ibc[ind].Withdraw = 0
					// ibc[ind].Volume = 0 // verificar
				}

			}
		} else {
			fmt.Println("ERROR SCP RUN JOB: Falta parametros em", scp_job_run, params)
			return false
		}

	case scp_job_ask:
		fmt.Println("ERROR SCP RUN JOB: ASK nao disponivel para IBC")
		// if len(subpars) > 0 {
		// 	msg := subpars[0]
		// 	scraddr := bio_cfg[bioid].Screenaddr
		// 	var cmd1 string = ""
		// 	var msgask string = ""
		// 	scrmain := fmt.Sprintf("CMD/%s/PUT/S200,1/END", scraddr)
		// 	switch msg {
		// 	case scp_msg_cloro:
		// 		cmd1 = fmt.Sprintf("CMD/%s/PUT/S400,2/END", scraddr)
		// 		msgask = "CLORO"
		// 	case scp_msg_meio_inoculo:
		// 		cmd1 = fmt.Sprintf("CMD/%s/PUT/S400,1/END", scraddr)
		// 		msgask = "MEIO e INOCULO"
		// 	default:
		// 		fmt.Println("ERROR SCP RUN JOB:", bioid, " ASK invalido", subpars)
		// 		return false
		// 	}
		// 	ret1 := scp_sendmsg_orch(cmd1)
		// 	fmt.Println("DEBUG SCP RUN JOB:: CMD =", cmd1, "\tRET =", ret1)
		// 	if !strings.Contains(ret1, scp_ack) && !devmode {
		// 		fmt.Println("ERROR SCP RUN JOB:", bioid, " ERROR ao enviar PUT screen", scraddr, ret1)
		// 		return false
		// 	}
		// 	cmd2 := fmt.Sprintf("CMD/%s/GET/S451/END", scraddr)
		// 	board_add_message("ABiorreator " + bioid + " aguardando " + msgask)
		// 	t_start := time.Now()
		// 	for {
		// 		ret2 := scp_sendmsg_orch(cmd2)
		// 		// fmt.Println("DEBUG SCP RUN JOB:: CMD =", cmd2, "\tRET =", ret2)
		// 		if !strings.Contains(ret2, scp_ack) && !devmode {
		// 			fmt.Println("ERROR SCP RUN JOB:", bioid, " ERROR ao envirar GET screen", scraddr, ret2)
		// 			scp_sendmsg_orch(scrmain)
		// 			return false
		// 		}
		// 		data := scp_splitparam(ret2, "/")
		// 		if len(data) > 1 {
		// 			if data[1] == "1" {
		// 				break
		// 			}
		// 		}
		// 		if bio[ind].MustPause || bio[ind].MustStop {
		// 			return false
		// 		}
		// 		t_elapsed := time.Since(t_start).Seconds()
		// 		if t_elapsed > scp_timeoutdefault {
		// 			fmt.Println("DEBUG SCP RUN JOB: Tempo maximo de ASK esgotado", bioid, t_elapsed, scp_maxtimewithdraw)
		// 			if !devmode {
		// 				scp_sendmsg_orch(scrmain)
		// 				return testmode
		// 			}
		// 			break
		// 		}
		// 		time.Sleep(scp_refreshwait * time.Millisecond)
		// 	}
		// 	scp_sendmsg_orch(scrmain)
		// } else {
		// 	fmt.Println("ERROR SCP RUN JOB: Falta parametros em", scp_job_org, params)
		// 	return false
		// }

	case scp_job_done:
		ibc[ind].Status = bio_ready
		board_add_message("CProcesso concluído no " + ibcid)
		ibc[ind].UndoQueue = []string{}
		ibc[ind].RedoQueue = []string{}
		ibc[ind].MustOffQueue = []string{}
		ibc[ind].Step = [2]int{0, 0}
		ibc[ind].ShowVol = true
		return true

	case scp_job_wait:
		var time_int uint64
		var err error
		if len(subpars) > 1 {
			switch subpars[0] {
			case scp_par_time:
				time_str := subpars[1]
				time_int, err = strconv.ParseUint(time_str, 10, 32)
				if devmode || testmode {
					if time_int > uint64(scp_timeoutdefault) {
						time_int = uint64(scp_timeoutdefault)
					}
				}
				if err != nil {
					fmt.Println("ERROR SCP RUN JOB: WAIT TIME invalido", time_str, params)
					return false
				}
				time_dur := time.Duration(time_int)
				fmt.Println("DEBUG SCP RUN JOB: WAIT de", time_dur.Seconds(), "segundos")
				var n time.Duration
				for n = 0; n < time_dur; n++ {
					if ibc[ind].MustPause || ibc[ind].MustStop {
						return false
					}
					time.Sleep(1000 * time.Millisecond)
				}

			case scp_par_volume:
				var vol_max uint64
				var err error
				vol_str := subpars[1]
				vol_max, err = strconv.ParseUint(vol_str, 10, 32)
				if err != nil {
					fmt.Println("ERROR SCP RUN JOB: WAIT VOLUME invalido", vol_str, params)
					return false
				}
				if vol_max > uint64(ibc_cfg[ibcid].Maxvolume) {
					fmt.Println("ERROR SCP RUN JOB: WAIT VOLUME maior do que maximo do Biorreator", vol_max, ibcid, ibc_cfg[ibcid].Maxvolume)
					return false
				}

				time_max := scp_timeoutdefault
				time_min := 0
				par_time := false
				if len(subpars) > 3 {
					time_max, err = strconv.Atoi(subpars[2])
					if err != nil {
						time_max = scp_timeoutdefault
						checkErr(err)
						fmt.Println("ERROR SCP RUN JOB: WAIT VOLUME,TEMPO MAX invalido", vol_str, params)
					} else {
						fmt.Println("DEBUG SCP RUN JOB: WAIT VOLUME e/ou TEMPO MAX", vol_str, time_max)
						par_time = true
					}
				}
				if len(subpars) > 4 {
					time_min, err = strconv.Atoi(subpars[3])
					if err != nil {
						time_min = 0
						checkErr(err)
						fmt.Println("ERROR SCP RUN JOB: WAIT VOLUME,TEMPO MIN invalido", vol_str, params)
					} else {
						fmt.Println("DEBUG SCP RUN JOB: WAIT VOLUME e TEMPO MIN", vol_str, time_min)
						par_time = true
					}
				}
				t_start := time.Now()
				for {
					vol_now := uint64(ibc[ind].Volume)
					t_elapsed := time.Since(t_start).Seconds()
					if vol_now >= vol_max && t_elapsed >= float64(time_min) {
						break
					}
					if ibc[ind].MustPause || ibc[ind].MustStop {
						return false
					}
					if t_elapsed > float64(time_max) {
						fmt.Println("DEBUG SCP RUN JOB: Tempo maximo de withdraw esgotado", t_elapsed, scp_maxtimewithdraw)
						if !devmode && !par_time {
							return testmode
						}
						break
					}
					time.Sleep(scp_refreshwait * time.Millisecond)
				}
			}
		} else {
			fmt.Println("ERROR SCP RUN JOB: Falta parametros em", scp_job_org, params)
			return false
		}

	case scp_job_on:
		if len(subpars) > 1 {
			device := subpars[0]
			switch device {
			case scp_dev_pump:
				valvs := []string{}
				for k := 1; k < len(subpars) && subpars[k] != "END"; k++ {
					v := ibcid + "/" + subpars[k]
					valvs = append(valvs, v)
				}
				if !scp_turn_pump(scp_ibc, ibcid, valvs, 1) {
					fmt.Println("ERROR SCP RUN JOB: ERROR ao ligar bomba em", ibcid, valvs)
					return false
				}

			case scp_dev_peris:
				if len(subpars) > 3 {
					peris_str := subpars[1]
					totem_str := subpars[2]
					pathid := totem_str + "-" + ibcid
					pathstr := paths[pathid].Path
					if len(pathstr) == 0 {
						fmt.Println("ERROR SCP RUN JOB: path nao existe", pathid)
						return false
					}
					fmt.Println("npath=", pathstr)
					vpath := scp_splitparam(pathstr, ",")
					perisvalv := totem_str + "/V2"
					n := len(vpath)
					vpath = append(vpath[:n-1], perisvalv)
					vpath = append(vpath, "END")
					fmt.Println("DEBUG", vpath)
					if test_path(vpath, 0) {
						if set_valvs_value(vpath, 1, true) < 0 {
							fmt.Println("ERROR SCP RUN JOB: ERRO ao abrir valvulas no path ", vpath)
							return false
						}
					} else {
						fmt.Println("ERROR SCP RUN JOB: ERRO nas valvulas no path ", vpath)
						return false
					}
					tmax := scp_timewaitvalvs / 100
					for i := 0; i < tmax; i++ {
						// switch devtype {
						// case scp_bioreactor:
						// 	if bio[ind].MustPause || bio[ind].MustPause {
						// 		i = tmax
						// 	}
						// }
						time.Sleep(100 * time.Millisecond)
					}
					if !scp_turn_peris(scp_totem, totem_str, peris_str, 1) {
						fmt.Println("ERROR SCP RUN JOB: ERROR ao ligar peristaltica em", totem_str, peris_str)
						return false
					}
				}

			case scp_dev_water:
				totem := subpars[1]
				totem_ind := get_totem_index(totem)
				if totem_ind < 0 {
					fmt.Println("ERROR SCP RUN JOB: Totem nao existe", totem)
					return false
				}
				if len(subpars) > 2 && subpars[2] != "END" {
					if subpars[2] == scp_dev_sprayball {
						fmt.Println("ERROR SCP RUN JOB: Nao e possivel entrar agua pelo sprayball em IBC", ibcid, params)
						break
					}
				}
				pathid := totem + "-" + ibcid
				pathstr := paths[pathid].Path
				if len(pathstr) == 0 {
					fmt.Println("ERROR SCP RUN JOB: path nao existe", pathid)
					return false
				}
				var npath string
				npath = pathstr
				fmt.Println("npath=", npath)
				vpath := scp_splitparam(npath, ",")
				watervalv := totem + "/V1"
				n := len(vpath)
				vpath = append(vpath[:n-1], watervalv)
				vpath = append(vpath, "END")
				fmt.Println("DEBUG", vpath)
				if !scp_turn_pump(scp_totem, totem, vpath, 1) {
					fmt.Println("ERROR SCP RUN JOB: ERROR ao ligar bomba em", ibcid, valvs)
					return false
				}
			}
		} else {
			fmt.Println("ERROR SCP RUN JOB: Falta parametros em", scp_job_on, params)
			return false
		}
	case scp_job_off:
		if len(subpars) > 0 {
			device := subpars[0]
			switch device {
			case scp_dev_pump:
				valvs := []string{}
				for k := 1; k < len(subpars) && subpars[k] != "END"; k++ {
					v := ibcid + "/" + subpars[k]
					valvs = append(valvs, v)
				}
				if !scp_turn_pump(scp_ibc, ibcid, valvs, 0) {
					fmt.Println("ERROR SCP RUN JOB: ERROR ao desligar bomba em", ibcid, valvs)
					return false
				}

			case scp_dev_peris:
				if len(subpars) > 3 {
					peris_str := subpars[1]
					totem_str := subpars[2]
					pathid := totem_str + "-" + ibcid
					pathstr := paths[pathid].Path
					if len(pathstr) == 0 {
						fmt.Println("ERROR SCP RUN JOB: path nao existe", pathid)
						return false
					}
					if !scp_turn_peris(scp_totem, totem_str, peris_str, 0) {
						fmt.Println("ERROR SCP RUN JOB: ERROR ao desligar peristaltica em", totem_str, peris_str)
						return false
					}
					fmt.Println("npath=", pathstr)
					vpath := scp_splitparam(pathstr, ",")
					perisvalv := totem_str + "/V2"
					n := len(vpath)
					vpath = append(vpath[:n-1], perisvalv)
					vpath = append(vpath, "END")
					fmt.Println("DEBUG", vpath)
					if test_path(vpath, 1) {
						if set_valvs_value(vpath, 0, true) < 0 {
							fmt.Println("ERROR SCP RUN JOB: ERRO ao fechar valvulas no path ", vpath)
							return false
						}
					} else {
						fmt.Println("ERROR SCP RUN JOB: ERRO nas valvulas no path ", vpath)
						return false
					}
					tmax := scp_timewaitvalvs / 100
					for i := 0; i < tmax; i++ {
						// switch devtype {
						// case scp_bioreactor:
						// 	if bio[ind].MustPause || bio[ind].MustPause {
						// 		i = tmax
						// 	}
						// }
						time.Sleep(100 * time.Millisecond)
					}
				}

			case scp_dev_water:
				totem := subpars[1]
				totem_ind := get_totem_index(totem)
				if totem_ind < 0 {
					fmt.Println("ERROR SCP RUN JOB: Totem nao existe", totem)
					return false
				}

				if len(subpars) > 2 && subpars[2] != "END" {
					if subpars[2] == scp_dev_sprayball {
						fmt.Println("ERROR SCP RUN JOB: Nao e possivel entrar agua pelo sprayball em IBC", ibcid, params)
					}
				}
				pathid := totem + "-" + ibcid
				pathstr := paths[pathid].Path
				if len(pathstr) == 0 {
					fmt.Println("ERROR SCP RUN JOB: path nao existe", pathid)
					return false
				}
				var npath string
				npath = pathstr
				fmt.Println("npath=", npath)
				vpath := scp_splitparam(npath, ",")
				watervalv := totem + "/V1"
				n := len(vpath)
				vpath = append(vpath[:n-1], watervalv)
				vpath = append(vpath, "END")
				if !scp_turn_pump(scp_totem, totem, vpath, 0) {
					fmt.Println("ERROR SCP RUN JOB: ERROR ao ligar bomba em", ibcid, valvs)
					return false
				}
			}
		} else {
			fmt.Println("ERROR SCP RUN JOB: Falta parametros em", scp_job_off, params)
			return false
		}
	default:
		fmt.Println("ERROR SCP RUN JOB: JOB invalido", ibcid, job, params)
	}
	time.Sleep(500 * time.Millisecond)
	return true
}

func scp_invert_onoff(job string) string {
	inv := ""
	if strings.Contains(job, "ON/") {
		inv = strings.Replace(job, "ON", "OFF", -1)
	} else if strings.Contains(job, "OFF/") {
		inv = strings.Replace(job, "OFF", "ON", -1)
	}
	return inv
}

func scp_run_bio(bioid string) {
	fmt.Println("STARTANDO RUN", bioid)
	ind := get_bio_index(bioid)
	if ind < 0 {
		fmt.Println("ERROR SCP RUN BIO: Biorreator nao existe", bioid)
		return
	}
	for bio[ind].Status != bio_die {
		if len(bio[ind].Queue) > 0 {
			fmt.Println("\n\nBIO", bioid, " status", bio[ind].Status)
			fmt.Println("\nQUEUE:", bio[ind].Queue)
			fmt.Println("\nUNDOQUEUE:", bio[ind].UndoQueue)
			fmt.Println("\nREDOQUEUE:", bio[ind].RedoQueue)
			fmt.Println("\nMUSTOFFQUEUE:", bio[ind].MustOffQueue)
		}

		if bio[ind].Status != bio_nonexist && bio[ind].Status != bio_error {
			if len(bio[ind].Queue) > 0 && bio[ind].Status != bio_pause && !bio[ind].MustPause && !bio[ind].MustStop {
				var ret bool = false
				job := pop_first_job(scp_bioreactor, bioid, false)
				if len(job) > 0 {
					ret = scp_run_job_bio(bioid, job)
				}
				if !ret {
					fmt.Println("ERROR SCP RUN BIO: Nao foi possivel executar JOB", bioid, job)
				} else {
					onoff := scp_invert_onoff(job)
					if len(onoff) > 0 {
						if !strings.Contains(onoff, "MUSTOFF") {
							bio[ind].UndoQueue = append(bio[ind].UndoQueue, onoff)
						} else {
							onoff_must := strings.Replace(onoff, ",MUSTOFF,", ",", -1)
							bio[ind].MustOffQueue = append(bio[ind].MustOffQueue, onoff_must)
						}
					}
					pop_first_job(scp_bioreactor, bioid, true)
				}
			} else if len(bio[ind].UndoQueue) > 0 && (bio[ind].MustPause || bio[ind].MustStop) {
				var ret bool = false
				job := pop_first_undojob(scp_bioreactor, bioid, false)
				if len(job) > 0 {
					ret = scp_run_job_bio(bioid, job)
				}
				if !ret {
					fmt.Println("ERROR SCP RUN BIO: Nao foi possivel executar UNDO JOB", bioid, job)
				} else {
					onoff := scp_invert_onoff(job)
					if len(onoff) > 0 {
						bio[ind].RedoQueue = append(bio[ind].RedoQueue, onoff)
					}
					pop_first_undojob(scp_bioreactor, bioid, true)
				}
			}

		}
		time.Sleep(scp_schedwait * time.Millisecond)
	}
}

func scp_run_ibc(ibcid string) {
	fmt.Println("STARTANDO RUN", ibcid)
	ind := get_ibc_index(ibcid)
	if ind < 0 {
		fmt.Println("ERROR SCP RUN BIO: Biorreator nao existe", ibcid)
		return
	}
	for ibc[ind].Status != bio_die {
		if len(ibc[ind].Queue) > 0 {
			fmt.Println("\n\nIBC", ibcid, " status", ibc[ind].Status)
			fmt.Println("\nQUEUE:", ibc[ind].Queue)
			fmt.Println("\nUNDOQUEUE:", ibc[ind].UndoQueue)
			fmt.Println("\nREDOQUEUE:", ibc[ind].RedoQueue)
			fmt.Println("\nMUSTOFFQUEUE:", ibc[ind].MustOffQueue)
		}

		if ibc[ind].Status != bio_nonexist && ibc[ind].Status != bio_error {
			if len(ibc[ind].Queue) > 0 && ibc[ind].Status != bio_pause && !ibc[ind].MustPause && !ibc[ind].MustStop {
				var ret bool = false
				job := pop_first_job(scp_ibc, ibcid, false)
				if len(job) > 0 {
					ret = scp_run_job_ibc(ibcid, job)
				}
				if !ret {
					fmt.Println("ERROR SCP RUN BIO: Nao foi possivel executar JOB", ibcid, job)
				} else {
					onoff := scp_invert_onoff(job)
					if len(onoff) > 0 {
						if !strings.Contains(onoff, "MUSTOFF") {
							ibc[ind].UndoQueue = append(ibc[ind].UndoQueue, onoff)
						} else {
							onoff_must := strings.Replace(onoff, ",MUSTOFF,", ",", -1)
							ibc[ind].MustOffQueue = append(ibc[ind].MustOffQueue, onoff_must)
						}
					}
					pop_first_job(scp_ibc, ibcid, true)
				}
			} else if len(ibc[ind].UndoQueue) > 0 && (ibc[ind].MustPause || ibc[ind].MustStop) {
				var ret bool = false
				job := pop_first_undojob(scp_ibc, ibcid, false)
				if len(job) > 0 {
					ret = scp_run_job_ibc(ibcid, job)
				}
				if !ret {
					fmt.Println("ERROR SCP RUN BIO: Nao foi possivel executar UNDO JOB", ibcid, job)
				} else {
					onoff := scp_invert_onoff(job)
					if len(onoff) > 0 {
						ibc[ind].RedoQueue = append(ibc[ind].RedoQueue, onoff)
					}
					pop_first_undojob(scp_ibc, ibcid, true)
				}
			}

		}
		time.Sleep(scp_schedwait * time.Millisecond)
	}
}

func scp_clock() {
	t_start := time.Now()
	for {
		for _, b := range bio {
			// fmt.Println("CLOCK CHECANDo Biorreator", b.BioreactorID)
			if b.Status != bio_pause && b.Status != bio_error && b.Status != bio_nonexist && b.Status != bio_ready && b.Status != bio_empty {
				ind := get_bio_index(b.BioreactorID)
				t_elapsed := time.Since(t_start).Minutes()
				totalleft := bio[ind].Timeleft[0]*60 + bio[ind].Timeleft[1]
				fmt.Println("CLOCK TOTAL LEFT", b.BioreactorID, totalleft)
				if totalleft > 0 {
					totalleft -= int(t_elapsed)
					bio[ind].Timeleft[0] = int(totalleft / 60)
					bio[ind].Timeleft[1] = int(totalleft % 60)
				}
			}
		}

		for _, b := range ibc {
			if b.Status == bio_cip {
				ind := get_ibc_index(b.IBCID)
				t_elapsed := time.Since(t_start).Minutes()
				totalleft := ibc[ind].Timetotal[0]*60 + ibc[ind].Timetotal[1]
				if totalleft > 0 {
					totalleft -= int(t_elapsed)
					ibc[ind].Timetotal[0] = int(totalleft / 60)
					ibc[ind].Timetotal[1] = int(totalleft % 60)
				}
			}
		}

		t_start = time.Now()
		time.Sleep(scp_clockwait * time.Millisecond)

	}
}

func scp_run_devs() {
	for _, b := range bio {
		go scp_run_bio(b.BioreactorID)
	}
	for _, b := range ibc {
		go scp_run_ibc(b.IBCID)
	}
}

func scp_scheduler() {
	schedrunning = true
	if !devsrunning {
		scp_run_devs()
		go scp_clock()
	}
	for schedrunning == true {
		for k, b := range bio {
			// fmt.Println(k, " bio =", b)
			r := pop_first_sched(b.BioreactorID, false)
			if len(r.Bioid) > 0 {
				if b.Status == bio_empty && len(b.Queue) == 0 { // && b.Volume == 0
					fmt.Println("\n", k, " Schedule inicial", schedule, "//", len(schedule), "POP de ", b.BioreactorID)
					s := pop_first_sched(b.BioreactorID, true)
					fmt.Println("Schedule depois do POP", schedule, "//", len(schedule), "\n\n")
					if len(s.Bioid) > 0 {
						if s.OrgCode == scp_par_cip {
							bio[k].Queue = []string{"RUN/CIP/END"}
						} else {
							orginfo := []string{"ORG/" + s.OrgCode + ",END"}
							bio[k].Queue = append(orginfo, recipe...)
							if autowithdraw {
								outjob := "RUN/WITHDRAW," + strings.Replace(b.BioreactorID, "BIOR", "IBC", -1) + ",END"
								wdraw := []string{"SET/STATUS,DESENVASE,END", outjob, "RUN/CIP/END"}
								bio[k].Queue = append(bio[k].Queue, wdraw...)
							}
							bio[k].Status = bio_starting
							fmt.Println("DEBUG SCP SCHEDULER: Biorreator", b.BioreactorID, " ira produzir", s.OrgCode, "-", bio[k].Organism)
						}
					}
				}
			}
		}
		for k, b := range ibc {
			r := pop_first_sched(b.IBCID, false)
			if len(r.Bioid) > 0 {
				if b.Status == bio_empty && len(b.Queue) == 0 { // && b.Volume == 0
					fmt.Println("\n", k, " Schedule inicial", schedule, "//", len(schedule), "POP de ", b.IBCID)
					s := pop_first_sched(b.IBCID, true)
					fmt.Println("Schedule depois do POP", schedule, "//", len(schedule), "\n\n")
					if len(s.Bioid) > 0 {
						if s.OrgCode == scp_par_cip {
							ibc[k].Queue = []string{"RUN/CIP/END"}
						} else {
							orginfo := []string{"ORG/" + s.OrgCode + ",END"}
							ibc[k].Queue = append(orginfo, recipe...)
							// if autowithdraw {
							// 	outjob := "RUN/WITHDRAW," + strings.Replace(b.IBCID, "BIOR", "IBC", -1) + ",END"
							// 	wdraw := []string{"SET/STATUS,DESENVASE,END", outjob, "RUN/CIP/END"}
							// 	bio[k].Queue = append(bio[k].Queue, wdraw...)
							// }
							ibc[k].Status = bio_starting
							fmt.Println("DEBUG SCP SCHEDULER: Biorreator", b.IBCID, " ira produzir", s.OrgCode, "-", ibc[k].Organism)
						}
					}
				}
			}
		}
		time.Sleep(scp_schedwait * time.Millisecond)
	}
}

func create_sched(lista []string) int {
	tot := 0
	for _, i := range lista {
		if i == "END" {
			break
		}
		item := scp_splitparam(i, ",")
		main_id := item[0]
		bioseq, _ := strconv.Atoi(item[1])
		orgcode := item[2]
		ind_bio := get_bio_index(main_id)
		ind_ibc := get_ibc_index(main_id)
		if ind_bio >= 0 || ind_ibc >= 0 {
			schedule = append(schedule, Scheditem{main_id, bioseq, orgcode})
			tot++
		} else {
			fmt.Println("ERROR CREATE SCHED: DISPOSITIVO nao existe", main_id)
		}
	}
	fmt.Println(schedule)
	return tot
}

func pause_device(devtype string, main_id string, pause bool) bool {
	switch devtype {
	case scp_bioreactor:
		ind := get_bio_index(main_id)
		indbak := get_biobak_index(main_id)
		if ind < 0 {
			fmt.Println("ERROR PAUSE DEVICE: Biorreator nao existe", main_id)
			break
		}
		fmt.Println("\n\nDEBUG PAUSE DEVICE: Biorreator", main_id, " em", pause)
		if pause && bio[ind].Status != bio_pause {
			fmt.Println("DEBUG PAUSE DEVICE: Pausando Biorreator", main_id)
			biobak[indbak] = bio[ind]
			bio[ind].LastStatus = bio[ind].Status
			for _, j := range bio[ind].MustOffQueue {
				if !isin(bio[ind].UndoQueue, j) {
					bio[ind].UndoQueue = append([]string{j}, bio[ind].UndoQueue...)
				}
			}
			// bio[ind].UndoQueue = append(bio[ind].MustOffQueue, bio[ind].UndoQueue...)
			bio[ind].MustPause = true
			bio[ind].Status = bio_pause
			if !bio[ind].MustStop {
				board_add_message("ABiorreator " + main_id + " pausado")
			}

		} else if !pause {
			fmt.Println("DEBUG PAUSE DEVICE: Retomando Biorreator", main_id)
			bio[ind].Queue = append(bio[ind].RedoQueue, bio[ind].Queue...)
			// fmt.Println("****** LAST STATUS no PAUSE", bio[ind].LastStatus)
			if bio[ind].LastStatus == bio_pause {
				if bio[ind].Volume == 0 {
					bio[ind].Status = bio_empty
				} else {
					bio[ind].Status = bio_ready
				}
			} else {
				bio[ind].Status = bio[ind].LastStatus
			}
			bio[ind].UndoQueue = []string{}
			bio[ind].RedoQueue = []string{}
			bio[ind].MustPause = false
			bio[ind].MustStop = false
			bio[ind].LastStatus = bio_pause
			if !bio[ind].MustStop {
				board_add_message("APausa no Biorreator " + main_id + " liberada")
			}
			if !schedrunning {
				go scp_scheduler()
			}
		}

	case scp_ibc:
		ind := get_ibc_index(main_id)
		// indbak := get_biobak_index(main_id)
		if ind < 0 {
			fmt.Println("ERROR PAUSE DEVICE: IBC nao existe", main_id)
			break
		}
		fmt.Println("\n\nDEBUG PAUSE DEVICE: IBC", main_id, " em", pause)
		if pause && ibc[ind].Status != bio_pause {
			fmt.Println("DEBUG PAUSE DEVICE: Pausando IBC", main_id)
			// biobak[indbak] = bio[ind]
			ibc[ind].LastStatus = ibc[ind].Status
			for _, j := range ibc[ind].MustOffQueue {
				if !isin(ibc[ind].UndoQueue, j) {
					ibc[ind].UndoQueue = append([]string{j}, ibc[ind].UndoQueue...)
				}
			}
			// bio[ind].UndoQueue = append(bio[ind].MustOffQueue, bio[ind].UndoQueue...)
			ibc[ind].MustPause = true
			ibc[ind].Status = bio_pause
			if !ibc[ind].MustStop {
				board_add_message("AIBC " + main_id + " pausado")
			}

		} else if !pause {
			fmt.Println("DEBUG PAUSE DEVICE: Retomando IBC", main_id)
			ibc[ind].Queue = append(ibc[ind].RedoQueue, ibc[ind].Queue...)
			// fmt.Println("****** LAST STATUS no PAUSE", bio[ind].LastStatus)
			if ibc[ind].LastStatus == bio_pause {
				if ibc[ind].Volume == 0 {
					ibc[ind].Status = bio_empty
				} else {
					ibc[ind].Status = bio_ready
				}
			} else {
				ibc[ind].Status = ibc[ind].LastStatus
			}
			ibc[ind].UndoQueue = []string{}
			ibc[ind].RedoQueue = []string{}
			ibc[ind].MustPause = false
			ibc[ind].MustStop = false
			ibc[ind].LastStatus = bio_pause
			if !ibc[ind].MustStop {
				board_add_message("APausa no IBC " + main_id + " liberada")
			}
			if !schedrunning {
				go scp_scheduler()
			}
		}
	default:
		fmt.Println("ERROR PAUSE DEVICE: Tipo de dispositivo invalido", devtype, main_id)
		return false
	}
	return true
}

func stop_device(devtype string, main_id string) bool {
	switch devtype {
	case scp_bioreactor:
		ind := get_bio_index(main_id)
		if ind < 0 {
			fmt.Println("ERROR STOP: Biorreator nao existe", main_id)
			return false
		}
		fmt.Println("\n\nDEBUG STOP: Executando STOP para", main_id)
		bio[ind].Withdraw = 0
		bio_add_message(main_id, "ABiorreator Interrompido")
		if bio[ind].Status != bio_empty || true { // corrigir
			bio[ind].MustStop = true
			pause_device(devtype, main_id, true)
			for {
				time.Sleep(5000 * time.Millisecond)
				if len(bio[ind].UndoQueue) == 0 {
					break
				}
			} //
			bio[ind].Queue = []string{}
			bio[ind].RedoQueue = []string{}
			bio[ind].MustOffQueue = []string{}
			bio[ind].MustStop = false
			bio[ind].Timetotal[0] = 0
			bio[ind].Timetotal[1] = 0
			bio[ind].Timeleft[0] = 0
			bio[ind].Timeleft[1] = 0
			bio[ind].Step[0] = 0
			bio[ind].Step[1] = 0
			q := pop_first_sched(bio[ind].BioreactorID, false)

			if len(q.Bioid) == 0 { // Verificar depois
				if bio[ind].Volume == 0 {
					bio[ind].Status = bio_empty
				} else {
					bio[ind].Status = bio_ready
				}
				bio[ind].MustPause = false
			}
			bio[ind].ShowVol = true
		}

	case scp_ibc:
		ind := get_ibc_index(main_id)
		if ind < 0 {
			fmt.Println("ERROR STOP: IBC nao existe", main_id)
			return false
		}
		fmt.Println("\n\nDEBUG STOP: Executando STOP para", main_id)
		ibc[ind].Withdraw = 0
		board_add_message("AIBC " + main_id + " interrompido")
		if ibc[ind].Status != bio_empty || true { // corrigir
			ibc[ind].MustStop = true
			pause_device(devtype, main_id, true)
			for {
				time.Sleep(5000 * time.Millisecond)
				if len(ibc[ind].UndoQueue) == 0 {
					break
				}
			} //
			ibc[ind].Queue = []string{}
			ibc[ind].RedoQueue = []string{}
			ibc[ind].MustOffQueue = []string{}
			ibc[ind].MustStop = false
			ibc[ind].Timetotal[0] = 0
			ibc[ind].Timetotal[1] = 0
			ibc[ind].Step[0] = 0
			ibc[ind].Step[1] = 0
			q := pop_first_sched(ibc[ind].IBCID, false)

			if len(q.Bioid) == 0 {
				if ibc[ind].Volume == 0 {
					ibc[ind].Status = bio_empty
				} else {
					ibc[ind].Status = bio_ready
				}
				ibc[ind].MustPause = false
			}
			ibc[ind].ShowVol = true
		}
	}
	save_all_data(data_filename)
	return true
}

func scp_restart_services() {
	// fmt.Println("Reestartando Servico ORCH")
	cmdpath, _ := filepath.Abs("/usr/bin/systemctl")
	cmd := exec.Command(cmdpath, "restart", "scp_orch")
	cmd.Dir = "/usr/bin"
	output, err := cmd.CombinedOutput()
	if len(output) > 0 {
		fmt.Println("OUPUT", string(output))
	}
	if err != nil {
		checkErr(err)
		fmt.Println("Falha ao Restartar ORCH")
		return
	}
	time.Sleep(10 * time.Second)
	fmt.Println("Reestartando Servico BACKEND")
	cmd = exec.Command(cmdpath, "restart", "scp_back")
	cmd.Dir = "/usr/bin"
	output, err = cmd.CombinedOutput()
	if len(output) > 0 {
		fmt.Println("OUPUT", string(output))
	}
	if err != nil {
		checkErr(err)
		fmt.Println("Falha ao Restartar BACK")
		return
	}
	time.Sleep(10 * time.Second)
	fmt.Println("Reestartando Servico MASTER")
	cmd = exec.Command(cmdpath, "restart", "scp_master")
	cmd.Dir = "/usr/bin"
	output, err = cmd.CombinedOutput()
	if len(output) > 0 {
		fmt.Println("OUPUT", string(output))
	}
	if err != nil {
		checkErr(err)
		fmt.Println("Falha ao Restartar MASTER")
		return
	}
}

func scp_run_manydraw_out(data string, dest string) {
	ibc_par := scp_splitparam(data, ",")
	for _, b := range ibc_par {
		d := scp_splitparam(b, "=")
		i := get_ibc_index(d[0])
		if i >= 0 && len(d) >= 2 {
			vol, err := strconv.Atoi(d[1])
			if err == nil {
				ibc[i].OutID = dest
				ibc[i].Withdraw = uint32(vol)
				fmt.Println("DEBUG SCP RUN MANYDRAW OUT: Desenvase de", d[0], " para", dest, " Volume", vol)
				scp_run_withdraw(scp_ibc, d[0], false, false)
			} else {
				checkErr(err)
			}
		}
	}
}

func scp_process_conn(conn net.Conn) {
	buf := make([]byte, 512)
	n, err := conn.Read(buf)
	if err != nil {
		checkErr(err)
		return
	}
	//fmt.Printf("msg: %s\n", buf[:n])
	params := scp_splitparam(string(buf[:n]), "/")
	// fmt.Println(params)
	scp_command := params[0]
	switch scp_command {
	case scp_sched:
		scp_object := params[1]
		switch scp_object {
		case scp_biofabrica:
			lista := params[2:]
			// fmt.Println("LISTA:", lista)
			n := create_sched(lista)
			if n > 0 && !schedrunning {
				go scp_scheduler()
			}
		}

	case scp_config:
		scp_object := params[1]
		switch scp_object {
		case scp_bioreactor:
			if len(params) > 4 {
				bioid := params[2]
				ind := get_bio_index(bioid)
				if ind >= 0 {
					switch params[3] {
					case scp_par_getconfig:
						biocfg, ok := bio_cfg[bioid]
						if ok {
							buf, err := json.Marshal(biocfg)
							checkErr(err)
							conn.Write([]byte(buf))
						}
					case scp_par_deviceaddr:
						if len(params) > 4 {
							devaddr := params[4]
							biocfg := bio_cfg[bioid]
							biocfg.Deviceaddr = devaddr
							bio_cfg[bioid] = biocfg
							fmt.Println("DEBUG SCP PROCESS CON: Mudanca endereco do Biorreator", bioid, " para", devaddr, " = ", bio_cfg[bioid])
							conn.Write([]byte(scp_ack))
						}

					case scp_par_screenaddr:
						if len(params) > 4 {
							devaddr := params[4]
							biocfg := bio_cfg[bioid]
							biocfg.Screenaddr = devaddr
							bio_cfg[bioid] = biocfg
							fmt.Println("DEBUG SCP PROCESS CON: Mudanca endereco da tela do Biorreator", bioid, " para", devaddr, " = ", bio_cfg[bioid])
							conn.Write([]byte(scp_ack))
						}

					case scp_par_ph4:
						fmt.Println("DEBUG CONFIG: Ajustando PH 4")
						n := 0
						var data []float64
						for i := 0; i <= 7; i++ {
							tmp := scp_get_ph_voltage(bioid)
							if tmp >= 2 && tmp <= 5 {
								data = append(data, tmp)
								n++
							}
						}
						mediana := calc_mediana(data)
						if mediana > 0 {
							bio[ind].PHref[0] = mediana
							fmt.Println("DEBUG CONFIG: Mediana Voltagem PH 4", bio[ind].PHref[0], " amostras =", n)
						} else {
							fmt.Println("ERROR CONFIG: Valores INVALIDOS de PH 4")
						}

					case scp_par_ph7:
						fmt.Println("DEBUG CONFIG: Ajustando PH 7")
						n := 0
						var data []float64
						for i := 0; i <= 7; i++ {
							tmp := scp_get_ph_voltage(bioid)
							if tmp >= 2 && tmp <= 5 {
								data = append(data, tmp)
								n++
							}
						}
						mediana := calc_mediana(data)
						if mediana > 0 {
							bio[ind].PHref[1] = mediana
							fmt.Println("DEBUG CONFIG: Mediana Voltagem PH 7", bio[ind].PHref[1], " amostras =", n)
						} else {
							fmt.Println("ERROR CONFIG: Valores INVALIDOS de PH 7")
						}

					case scp_par_ph10:
						fmt.Println("DEBUG CONFIG: Ajustando PH 10")
						n := 0
						var data []float64
						for i := 0; i <= 7; i++ {
							tmp := scp_get_ph_voltage(bioid)
							if tmp >= 2 && tmp <= 5 {
								data = append(data, tmp)
								n++
							}
						}
						mediana := calc_mediana(data)
						if mediana > 0 {
							bio[ind].PHref[2] = mediana
							fmt.Println("DEBUG CONFIG: Mediana Voltagem PH 10", bio[ind].PHref[2], " amostras =", n)
						} else {
							fmt.Println("ERROR CONFIG: Valores INVALIDOS de PH 10")
						}

					case scp_par_calibrate:
						fmt.Println("DEBUG CONFIG: Calculando regressao linear para o PH")
						if bio[ind].PHref[0] > 0 && bio[ind].PHref[1] > 0 && bio[ind].PHref[2] > 0 {
							X_data := []float64{bio[ind].PHref[0], bio[ind].PHref[1], bio[ind].PHref[2]}
							y_data := []float64{4, 7, 10}
							// Executa a regressao linear
							b0, b1 := estimateB0B1(X_data, y_data)
							bio[ind].RegresPH[0] = b0
							bio[ind].RegresPH[1] = b1
							fmt.Println("DEBUG CONFIG: Coeficientes da Regressao Linear: b0=", b0, " b1=", b1)
						} else {
							fmt.Println("ERROR CONFIG: Nao e possivel fazer regressao linear, valores invalidos", bio[ind].PHref)
						}
					}
				} else {
					fmt.Println("ERROR CONFIG: Biorreator nao existe", bioid)
				}
			} else {
				fmt.Println("ERROR CONFIG: BIORREATOR - Numero de parametros invalido", params)
			}
		case scp_biofabrica:
			if len(params) > 3 {
				cmd := params[2]
				switch cmd {
				case scp_par_save:
					fmt.Println("DEBUG CONFIG: Salvando configuracoes")
					save_all_data(data_filename)
					save_bios_conf(localconfig_path + "bio_conf.csv")
					conn.Write([]byte(scp_ack))

				case scp_par_restart:
					fmt.Println("DEBUG CONFIG: Restartando Service")
					scp_restart_services()

				case scp_par_testmode:
					if len(params) > 4 {
						flag_str := params[3]
						fmt.Println("DEBUG CONFIG: Mudando TESTMODE para", flag_str)
						flag, err := strconv.ParseBool(flag_str)
						if err != nil {
							checkErr(err)
							conn.Write([]byte(scp_err))
						} else {
							biofabrica.TestMode = flag
							testmode = flag
							conn.Write([]byte(scp_ack))
						}
					} else {
						fmt.Println("ERROR CONFIG: BIOFABRICA TESTMODE - Numero de parametros invalido", params)
					}
				}

			} else {
				fmt.Println("ERROR CONFIG: BIOFABRICA - Numero de parametros invalido", params)
			}
		}

	case scp_start:
		scp_object := params[1]
		switch scp_object {
		case scp_bioreactor:
			bioid := params[2]
			orgcode := params[3]
			fmt.Println("START", bioid, orgcode, params)
			ind := get_bio_index(bioid)
			if ind < 0 {
				fmt.Println("ERROR START: Biorreator nao existe", bioid)
				break
			}
			if orgcode == scp_par_cip || len(organs[orgcode].Orgname) > 0 {
				fmt.Println("START", orgcode)
				biotask := []string{bioid + ",0," + orgcode}
				n := create_sched(biotask)
				if n > 0 && !schedrunning {
					go scp_scheduler()
				}
			} else {
				fmt.Println("ORG INVALIDO")
			}

		case scp_ibc:
			ibcid := params[2]
			orgcode := params[3]
			fmt.Println("START", ibcid, orgcode, params)
			ind := get_ibc_index(ibcid)
			if ind < 0 {
				fmt.Println("ERROR START: IBC nao existe", ibcid)
				break
			}
			if orgcode == scp_par_cip || len(organs[orgcode].Orgname) > 0 {
				fmt.Println("START", orgcode)
				biotask := []string{ibcid + ",0," + orgcode}
				n := create_sched(biotask)
				if n > 0 && !schedrunning {
					go scp_scheduler()
				}
			} else {
				fmt.Println("ORG INVALIDO")
			}

		case scp_biofabrica:
			cmdpar := params[2]
			switch cmdpar {
			case scp_par_linewash:
				scp_run_linewash(params[3])

			case scp_par_linecip:
				scp_run_linecip(params[3])

			default:
				fmt.Println("ERROR START: Parametro de Biofabrica invalido", params)
			}
		}

	case scp_stop:
		devtype := params[1]
		id := params[2]
		if !stop_device(devtype, id) {
			fmt.Println("ERROR STOP: Nao foi possivel parar dispositivo", devtype, id)
			break
		}

	case scp_pause:
		fmt.Println("PAUSE")
		devtype := params[1]
		id := params[2]
		pauseflag := params[3]
		if len(pauseflag) > 0 {
			pause, err := strconv.ParseBool(pauseflag)
			if err != nil {
				checkErr(err)
				break
			}
			if !pause_device(devtype, id, pause) {
				fmt.Println("ERROR PAUSE: Nao foi possivel pausar dispositivo", devtype, id)
				break
			}
		} else {
			fmt.Println("ERROR PAUSE: Parametros invalidos", devtype, id, params)
		}

	case scp_wdpanel:
		fmt.Println("DEBUG SCP PROCESS CONN:", params)
		if len(params) > 2 {
			subpars := scp_splitparam(params[1], ",")
			subcmd := subpars[0]
			ibc_id := subpars[1]
			ind := get_ibc_index(ibc_id)
			if ind < 0 {
				fmt.Println("ERROR WDPANEL: IBC invalido", ibc_id, params)
				conn.Write([]byte(scp_err))
				return
			}
			switch subcmd {
			case scp_par_select:
				for _, b := range ibc {
					i := get_ibc_index(b.IBCID)
					ibc[i].Selected = false
				}
				ibc[ind].Selected = true
				conn.Write([]byte(scp_ack))
				fmt.Println("DEBUG WDPANEL: IBC Selecionado", ibc_id)

			case scp_par_inc:
				if ibc[ind].Withdraw < ibc[ind].Volume {
					ibc[ind].Withdraw += bio_withdrawstep
					if ibc[ind].Withdraw > ibc[ind].Volume {
						ibc[ind].Withdraw = ibc[ind].Volume
					}
					fmt.Println("DEBUG WDPANEL: Withdraw IBC", ibc_id, ibc[ind].Withdraw)
				} else {
					fmt.Println("ERROR WDPANEL: Volume Máximo para Desenvase atingido", ibc_id)
					conn.Write([]byte(scp_err))
				}

			case scp_par_dec:
				if ibc[ind].Volume > 0 {
					if int32(ibc[ind].Withdraw)-bio_withdrawstep > 0 {
						ibc[ind].Withdraw -= bio_withdrawstep
					} else {
						ibc[ind].Withdraw = 0
					}
					if ibc[ind].Withdraw < 0 {
						ibc[ind].Withdraw = 0
					}
					fmt.Println("DEBUG WDPANEL: Withdraw IBC", ibc_id, ibc[ind].Withdraw)
				} else {
					fmt.Println("ERROR WDPANEL: Volume Mínimo para Desenvase atingido", ibc_id)
					conn.Write([]byte(scp_err))
				}

			case scp_par_start:
				ibc[ind].OutID = "OUT"
				fmt.Println(scp_ibc, ibc_id)
				if !withdrawrunning {
					go scp_run_withdraw(scp_ibc, ibc_id, true, false)
					fmt.Println("DEBUG WDPANEL: Executando Desenvase do", ibc_id, " volume", ibc[ind].Withdraw)
					conn.Write([]byte(scp_ack))
				} else {
					fmt.Println("ERROR WDPANEL: Desenvase ja em andamento", ibc_id)
					conn.Write([]byte(scp_err))
				}

			case scp_par_stop:
				ibc[ind].Withdraw = 0
				fmt.Println("DEBUG WDPANEL: Parando Desenvase do IBC", ibc_id)
				conn.Write([]byte(scp_ack))

			}
		} else {
			fmt.Println("ERROR WDPANEL: Parametros invalidos", params)
			conn.Write([]byte(scp_err))
			return
		}

	case scp_get:
		scp_object := params[1]
		switch scp_object {
		case scp_bioreactor:
			if params[2] == "END" {
				buf, err := json.Marshal(bio)
				checkErr(err)
				bio_etl := make([]Bioreact_ETL, 0)
				json.Unmarshal(buf, &bio_etl)
				buf2, err2 := json.Marshal(bio_etl)
				checkErr(err2)
				conn.Write([]byte(buf2))
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
				ibc_etl := make([]IBC_ETL, 0)
				json.Unmarshal(buf, &ibc_etl)
				buf2, err2 := json.Marshal(ibc_etl)
				checkErr(err2)
				conn.Write([]byte(buf2))
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

		case scp_biofabrica:
			// fmt.Println(params)
			if params[2] == "END" {
				buf, err := json.Marshal(biofabrica)
				checkErr(err)
				conn.Write([]byte(buf))
			} else {
				conn.Write([]byte(scp_err))
			}

		case scp_totem:
			// fmt.Println(params)
			if params[2] == "END" {
				buf, err := json.Marshal(totem)
				checkErr(err)
				conn.Write([]byte(buf))
			} else {
				ind := get_totem_index(params[2])
				if ind >= 0 {
					buf, err := json.Marshal(totem[ind])
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
			//fmt.Println("obj=", scp_object)
			bioid := params[2]
			ind := get_bio_index(bioid)
			if bioid != "ALL" && ind < 0 {
				conn.Write([]byte(scp_err))
			} else {
				subparams := scp_splitparam(params[3], ",")
				scp_device := subparams[0]
				// fmt.Println("subparams=", subparams)
				switch scp_device {
				case scp_par_out:
					if ind >= 0 {
						outid := subparams[1]
						if outid == "Descarte" {
							outid = "DROP"
						} else if outid == "Externo" {
							outid = "OUT"
						}
						tmp := bioid + "-" + outid
						_, ok := paths[tmp]
						if ok {
							bio[ind].OutID = outid
							conn.Write([]byte(scp_ack))
						} else {
							fmt.Println("ID de saida", outid, " nao existe")
							conn.Write([]byte(scp_err))
						}
					}

				case scp_par_circulate:
					fmt.Println("DEBUG SCP PROCESS CONN: PAR CIRCULATE", params, subparams)
					if len(subparams) >= 2 {
						status, err := strconv.ParseBool(subparams[1])
						checkErr(err)
						if err == nil {
							if !status {
								if bioid == "ALL" {
									for _, b := range bio {
										i := get_bio_index(b.BioreactorID)
										if bio[i].Status == bio_circulate {
											bio[i].Status = bio[i].LastStatus
										}
									}
								} else {
									if bio[ind].Status == bio_circulate {
										bio[ind].Status = bio[ind].LastStatus
									}
								}
							} else {
								rec_time := 5
								if len(subparams) >= 3 {
									rec_time, err = strconv.Atoi(subparams[2])
									if err != nil {
										checkErr(err)
										rec_time = 5
									}
								}
								if bioid == "ALL" {
									for _, b := range bio {
										i := get_bio_index(b.BioreactorID)
										if bio[i].Status == bio_ready && bio[i].Volume > 0 {
											bio[i].LastStatus = bio[i].Status
											bio[i].Status = bio_circulate
											go scp_circulate(scp_bioreactor, b.BioreactorID, rec_time)
										}
									}
								} else if bio[ind].Status == bio_ready && bio[ind].Volume > 0 {
									bio[ind].LastStatus = bio[ind].Status
									bio[ind].Status = bio_circulate
									go scp_circulate(scp_bioreactor, bioid, rec_time)
								}
							}
							conn.Write([]byte(scp_ack))
						} else {
							conn.Write([]byte(scp_err))
						}
					} else {
						fmt.Println("ERROR PUT BIORREACTOR: Falta parametros em circulate", subparams)
					}

				case scp_par_withdraw:
					vol, err := strconv.Atoi(subparams[1])
					checkErr(err)
					if err == nil {
						bio[ind].Withdraw = uint32(vol)
						if bio[ind].Withdraw > 0 {
							if get_scp_type(bio[ind].OutID) == scp_ibc {
								go scp_run_withdraw(scp_bioreactor, bioid, true, true)
							} else {
								go scp_run_withdraw(scp_bioreactor, bioid, true, false)
							}
						}
						conn.Write([]byte(scp_ack))
					} else {
						conn.Write([]byte(scp_err))
					}

				case scp_dev_pump:
					var cmd2, cmd3 string
					value, err := strconv.ParseBool(subparams[1])
					checkErr(err)
					biodev := bio_cfg[bioid].Deviceaddr
					bioscr := bio_cfg[bioid].Screenaddr
					pumpdev := bio_cfg[bioid].Pump_dev
					bio[ind].Pumpstatus = value
					if value {
						cmd2 = "CMD/" + biodev + "/PUT/" + pumpdev + ",1/END"
						cmd3 = "CMD/" + bioscr + "/PUT/S270,1/END"
					} else {
						cmd2 = "CMD/" + biodev + "/PUT/" + pumpdev + ",0/END"
						cmd3 = "CMD/" + bioscr + "/PUT/S270,0/END"
					}
					ret2 := scp_sendmsg_orch(cmd2)
					fmt.Println("RET CMD2 =", ret2)
					ret3 := scp_sendmsg_orch(cmd3)
					fmt.Println("RET CMD3 =", ret3)
					conn.Write([]byte(scp_ack))

				case scp_dev_aero:
					var cmd1, cmd2, cmd3 string
					value, err := strconv.ParseBool(subparams[1])
					checkErr(err)
					bio[ind].Aerator = value
					biodev := bio_cfg[bioid].Deviceaddr
					bioscr := bio_cfg[bioid].Screenaddr
					aerodev := bio_cfg[bioid].Aero_dev
					if value {
						cmd1 = "CMD/" + biodev + "/PUT/D27,1/END"
						cmd2 = "CMD/" + biodev + "/PUT/" + aerodev + ",255/END"
						cmd3 = "CMD/" + bioscr + "/PUT/S271,1/END"

					} else {
						cmd1 = "CMD/" + biodev + "/PUT/D27,0/END"
						cmd2 = "CMD/" + biodev + "/PUT/" + aerodev + ",0/END"
						cmd3 = "CMD/" + bioscr + "/PUT/S271,0/END"
					}
					ret1 := scp_sendmsg_orch(cmd1)
					fmt.Println("RET CMD1 =", ret1)
					ret2 := scp_sendmsg_orch(cmd2)
					fmt.Println("RET CMD2 =", ret2)
					ret3 := scp_sendmsg_orch(cmd3)
					fmt.Println("RET CMD3 =", ret3)
					conn.Write([]byte(scp_ack))

				case scp_dev_valve:
					value_valve, err := strconv.Atoi(subparams[1])
					checkErr(err)
					value_status, err := strconv.Atoi(subparams[2])
					checkErr(err)
					if (value_valve >= 0) && (value_valve < bio_max_valves) {
						valvid := fmt.Sprintf("V%d", value_valve+1)
						if set_valv_status(scp_bioreactor, bioid, valvid, value_status) {
							conn.Write([]byte(scp_ack))
						} else {
							conn.Write([]byte(scp_err))
						}
					}

				case scp_dev_peris:
					value_peris, err := strconv.Atoi(subparams[1])
					checkErr(err)
					value_status, err := strconv.Atoi(subparams[2])
					checkErr(err)
					if (value_peris >= 0) && (value_peris < 5) {
						perisid := fmt.Sprintf("V%d", value_peris+1)
						if scp_turn_peris(scp_bioreactor, bioid, perisid, value_status) {
							conn.Write([]byte(scp_ack))
						} else {
							conn.Write([]byte(scp_err))
						}
					}

				default:
					conn.Write([]byte(scp_err))
				}
			}

		case scp_ibc:
			ibcid := params[2]
			ind := get_ibc_index(ibcid)
			if ibcid != "ALL" && ind < 0 {
				conn.Write([]byte(scp_err))
			} else {
				subparams := scp_splitparam(params[3], ",")
				scp_device := subparams[0]
				switch scp_device {
				case scp_par_out:
					outid := subparams[1]
					if outid == "Descarte" {
						outid = "DROP"
					} else if outid == "Externo" {
						outid = "OUT"
					}
					tmp := ibcid + "-" + outid
					_, ok := paths[tmp]
					if ok {
						ibc[ind].OutID = outid
						conn.Write([]byte(scp_ack))
					} else {
						fmt.Println("ID de saida", outid, " nao existe")
						conn.Write([]byte(scp_err))
					}

				case scp_par_manydraw:
					fmt.Println("DEBUG SCP PROCESS CONN: PAR MANYDRAW", params, subparams)
					if len(params) > 4 {
						go scp_run_manydraw_out(params[4], "IBC07")
					}

				case scp_par_manyout:
					fmt.Println("DEBUG SCP PROCESS CONN: PAR MANYOUT", params, subparams)
					if len(params) > 4 {
						go scp_run_manydraw_out(params[4], "OUT")
					}

				case scp_par_circulate:
					fmt.Println("DEBUG SCP PROCESS CONN: PAR CIRCULATE", params, subparams)
					if len(subparams) >= 2 {
						status, err := strconv.ParseBool(subparams[1])
						checkErr(err)
						if err == nil {
							if !status {
								if ibcid == "ALL" {
									for _, b := range ibc {
										i := get_ibc_index(b.IBCID)
										if ibc[i].Status == bio_circulate {
											ibc[i].Status = ibc[i].LastStatus
										}
									}
								} else {
									if ibc[ind].Status == bio_circulate {
										ibc[ind].Status = ibc[ind].LastStatus
									}
								}
							} else {
								rec_time := 5
								if len(subparams) >= 3 {
									rec_time, err = strconv.Atoi(subparams[2])
									if err != nil {
										checkErr(err)
										rec_time = 5
									}
								}
								if ibcid == "ALL" {
									for _, b := range ibc {
										i := get_ibc_index(b.IBCID)
										if ibc[i].Status == bio_ready && ibc[i].Volume > 0 {
											ibc[i].LastStatus = ibc[i].Status
											ibc[i].Status = bio_circulate
											go scp_circulate(scp_ibc, b.IBCID, rec_time)
										}
									}
								} else if ibc[ind].Status == bio_ready && ibc[ind].Volume > 0 {
									ibc[ind].LastStatus = ibc[ind].Status
									ibc[ind].Status = bio_circulate
									go scp_circulate(scp_ibc, ibcid, rec_time)
								}
							}

							conn.Write([]byte(scp_ack))
						} else {
							conn.Write([]byte(scp_err))
						}
					}

				case scp_par_withdraw:
					vol, err := strconv.Atoi(subparams[1])
					checkErr(err)
					if err == nil {
						ibc[ind].Withdraw = uint32(vol)
						if ibc[ind].Withdraw > 0 {
							go scp_run_withdraw(scp_ibc, ibcid, true, false)
						}
						conn.Write([]byte(scp_ack))
					}
					conn.Write([]byte(scp_err))

				case scp_dev_pump:
					var cmd2 string
					value, err := strconv.ParseBool(subparams[1])
					checkErr(err)
					ibc[ind].Pumpstatus = value
					ibcdev := ibc_cfg[ibcid].Deviceaddr
					pumpdev := ibc_cfg[ibcid].Pump_dev
					//ibcscr := bio_cfg[ibcid].Screenaddr
					if value {
						cmd2 = "CMD/" + ibcdev + "/PUT/" + pumpdev + ",1/END"
					} else {
						cmd2 = "CMD/" + ibcdev + "/PUT/" + pumpdev + ",0/END"
					}

					ret2 := scp_sendmsg_orch(cmd2)
					fmt.Println("RET CMD2 =", ret2)

					conn.Write([]byte(scp_ack))

				case scp_dev_valve:
					// var cmd2 string
					value_valve, err := strconv.Atoi(subparams[1])
					checkErr(err)
					value_status, err := strconv.Atoi(subparams[2])
					checkErr(err)
					//fmt.Println(value_valve, value_status)
					if (value_valve >= 0) && (value_valve < bio_max_valves) {
						valvid := fmt.Sprintf("V%d", value_valve+1)
						set_valv_status(scp_ibc, ibcid, valvid, value_status)
						conn.Write([]byte(scp_ack))
					}
				default:
					conn.Write([]byte(scp_err))
				}
			}

		case scp_totem:
			totemid := params[2]
			ind := get_totem_index(totemid)
			if ind < 0 {
				conn.Write([]byte(scp_err))
			} else {
				subparams := scp_splitparam(params[3], ",")
				scp_device := subparams[0]
				switch scp_device {
				case scp_dev_pump:
					var cmd2 string
					value, err := strconv.ParseBool(subparams[1])
					checkErr(err)
					totem[ind].Pumpstatus = value
					totemdev := totem_cfg[totemid].Deviceaddr
					pumpdev := totem_cfg[totemid].Pumpdev
					if value {
						cmd2 = "CMD/" + totemdev + "/PUT/" + pumpdev + ",1/END"
					} else {
						cmd2 = "CMD/" + totemdev + "/PUT/" + pumpdev + ",0/END"
					}
					ret2 := scp_sendmsg_orch(cmd2)
					fmt.Println("RET CMD2 =", ret2)
					conn.Write([]byte(scp_ack))

				case scp_dev_valve:
					// var cmd2 string
					value_valve, err := strconv.Atoi(subparams[1])
					checkErr(err)
					value_status, err := strconv.Atoi(subparams[2])
					checkErr(err)
					//fmt.Println(value_valve, value_status)
					if (value_valve >= 0) && (value_valve < bio_max_valves) {
						valvid := fmt.Sprintf("V%d", value_valve+1)
						set_valv_status(scp_totem, totemid, valvid, value_status)
						conn.Write([]byte(scp_ack))
					}

				case scp_dev_peris:
					// var cmd2 string
					value_peris, err := strconv.Atoi(subparams[1])
					checkErr(err)
					value_status, err := strconv.Atoi(subparams[2])
					checkErr(err)
					//fmt.Println(value_valve, value_status)
					if (value_peris >= 0) && (value_peris < 4) {
						perisid := fmt.Sprintf("V%d", value_peris+1)
						if scp_turn_peris(scp_totem, totemid, perisid, value_status) {
							conn.Write([]byte(scp_ack))
						} else {
							conn.Write([]byte(scp_err))
						}
					}

				default:
					conn.Write([]byte(scp_err))
				}
			}

		case scp_biofabrica:
			subparams := scp_splitparam(params[2], ",")
			//fmt.Println("---subparams=", subparams)
			scp_device := subparams[0]
			switch scp_device {
			case scp_dev_pump:
				var cmd2 string
				value, err := strconv.ParseBool(subparams[1])
				checkErr(err)
				biofabrica.Pumpwithdraw = value
				devaddr := biofabrica_cfg["PBF01"].Deviceaddr
				devport := biofabrica_cfg["PBF01"].Deviceport
				if value {
					cmd2 = "CMD/" + devaddr + "/PUT/" + devport + ",1/END"
				} else {
					cmd2 = "CMD/" + devaddr + "/PUT/" + devport + ",0/END"
				}
				ret2 := scp_sendmsg_orch(cmd2)
				fmt.Println("RET CMD2 =", ret2)

				conn.Write([]byte(scp_ack))

			case scp_dev_valve:
				// var cmd2 string
				value_valve, err := strconv.Atoi(subparams[1])
				checkErr(err)
				value_status, err := strconv.Atoi(subparams[2])
				checkErr(err)
				//fmt.Println(value_valve, value_status)
				if (value_valve >= 0) && (value_valve < 9) {
					// biofabrica.Valvs[value_valve] = value_status
					devid := fmt.Sprintf("VBF%02d", value_valve+1)
					set_valv_status(scp_biofabrica, "BIOFABRICA", devid, value_status)
					// devaddr := biofabrica_cfg[devid].Deviceaddr
					// devport := biofabrica_cfg[devid].Deviceport
					// if value_status == 1 {
					// 	cmd2 = "CMD/" + devaddr + "/PUT/" + devport + ",1/END"
					// } else if value_status == 0 {
					// 	cmd2 = "CMD/" + devaddr + "/PUT/" + devport + ",0/END"
					// }
					// fmt.Println("biofabrica valvula", cmd2)
					// ret2 := scp_sendmsg_orch(cmd2)
					// fmt.Println("RET CMD2 =", ret2)
					conn.Write([]byte(scp_ack))
				}
			default:
				conn.Write([]byte(scp_err))
			}

		default:
			conn.Write([]byte(scp_err))
		}
		// fmt.Println(valvs)
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

		} else {
			go scp_process_conn(conn)
		}
	}
}

func main() {

	go scp_check_network()

	localconfig_path = "/etc/scpd/"
	addrs_type = make(map[string]DevAddrData, 0)
	net192 = test_file("/etc/scpd/scp_net192.flag")
	if net192 {
		fmt.Println("WARN:  EXECUTANDO EM NET192\n\n\n")
		execpath = "/home/paulo/work/iot/scp-project/"
		mainrouter = "192.168.0.1"
	} else {
		execpath = "/home/scpadm/scp-project/"
		mainrouter = "10.0.0.1"
	}
	devmode = test_file("/etc/scpd/scp_devmode.flag")
	if devmode {
		fmt.Println("WARN:  EXECUTANDO EM DEVMODE\n\n\n")
	}
	testmode = test_file("/etc/scpd/scp_testmode.flag")
	if testmode {
		biofabrica.TestMode = true
		fmt.Println("WARN:  EXECUTANDO EM TESTMODE\n\n\n")
	}

	norgs := load_organisms(execpath + "organismos_conf.csv")
	if norgs < 0 {
		log.Fatal("Não foi possivel ler o arquivo de organismos")
	}
	recipe = load_tasks_conf(execpath + "receita_conf.csv")
	if recipe == nil {
		log.Fatal("Não foi possivel ler o arquivo contendo a receita de producao")
	}
	cipbio = load_tasks_conf(execpath + "cip_bio_conf.csv")
	if recipe == nil {
		log.Fatal("Não foi possivel ler o arquivo contendo ciclo de CIP de Biorreator")
	}
	cipibc = load_tasks_conf(execpath + "cip_ibc_conf.csv")
	if recipe == nil {
		log.Fatal("Não foi possivel ler o arquivo contendo ciclo de CIP de IBC")
	}
	nibccfg := load_ibcs_conf(localconfig_path + "ibc_conf.csv")
	if nibccfg < 1 {
		log.Fatal("FATAL: Arquivo de configuracao dos IBCs nao encontrado")
	}
	nbiocfg := load_bios_conf(localconfig_path + "bio_conf.csv")
	if nbiocfg < 1 {
		nbiocfg_bak := load_bios_conf(localconfig_path + "bio_conf.csv.bak")
		if nbiocfg_bak < 1 {
			log.Fatal("FATAL: Arquivo de configuracao dos Bioreatores nao encontrado")
		}
	}
	ntotemcfg := load_totems_conf(localconfig_path + "totem_conf.csv")
	if ntotemcfg < 1 {
		log.Fatal("FATAL: Arquivo de configuracao dos Totems nao encontrado")
	}
	nbiofabricacfg := load_biofabrica_conf(localconfig_path + "biofabrica_conf.csv")
	if nbiofabricacfg < 1 {
		log.Fatal("FATAL: Arquivo de configuracao da Biofabrica nao encontrado")
	}
	npaths := load_paths_conf(localconfig_path + "paths_conf.csv")
	if npaths < 1 {
		log.Fatal("FATAL: Arquivo de configuracao de PATHs invalido")
	}

	valvs = make(map[string]int, 0)
	load_all_data(data_filename)

	biofabrica.TechMode = test_file("/etc/scpd/scp_techmode.flag")

	go scp_setup_devices(true)
	go scp_get_alldata()
	go scp_sync_functions()

	scp_master_ipc()
	time.Sleep(10 * time.Second)
	// if !schedrunning {
	// 	go scp_scheduler()
	// }
}
