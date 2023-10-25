package main

import (
	"encoding/csv"
	"encoding/json"

	// "filepath"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"gonum.org/v1/gonum/stat"
)

var demo = false
var devmode = false
var net192 = false
var testmode = false
var autowithdraw = false

const scp_onlyread_sensoribc = true

const control_ph = true
const control_temp = true
const control_foam = true

const (
	scp_version = "1.2.35" // 2023-10-25

	scp_on  = 1
	scp_off = 0

	bio_escala = 1

	last_ibc = "IBC07"

	scp_ack      = "ACK"
	scp_err      = "ERR"
	scp_get      = "GET"
	scp_put      = "PUT"
	scp_run      = "RUN"
	scp_die      = "DIE"
	scp_null     = "NULL"
	scp_sched    = "SCHED"
	scp_start    = "START"
	scp_status   = "STATUS"
	scp_stop     = "STOP"
	scp_pause    = "PAUSE"
	scp_fail     = "FAIL"
	scp_reboot   = "BOOT"
	scp_netfail  = "NETFAIL"
	scp_ready    = "READY"
	scp_sysstop  = "SYSSTOP"
	scp_stopall  = "STOPALL"
	scp_recovery = "RECOVERY"

	scp_state_JOIN0   = 10
	scp_state_JOIN1   = 11
	scp_state_TCP0    = 20
	scp_state_TCPFAIL = 29

	mainstatus_cip   = "CIP"
	mainstatus_grow  = "CULTIVO"
	mainstatus_org   = "ORGANISMO"
	mainstatus_empty = "VAZIO"

	bio_noaddr = "FF:FFFFFF"

	scp_magicport  = "D77"
	scp_magicvalue = "1"
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
const scp_dev_heater = "HEATER"
const scp_dev_all = "ALL"

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
const scp_par_techmode = "TECHMODE"
const scp_par_getconfig = "GETCONFIG"
const scp_par_getph = "GETPH"
const scp_par_deviceaddr = "DEVICEADDR"
const scp_par_screenaddr = "SCREENADDR"
const scp_par_linewash = "LINEWASH"
const scp_par_linecip = "LINECIP"
const scp_par_circulate = "CIRCULATE"
const scp_par_totaltime = "TOTALTIME"
const scp_par_manydraw = "MANYDRAW"
const scp_par_manyout = "MANYOUT"
const scp_par_continue = "CONTINUE"
const scp_par_reconfigdev = "RECONFIGDEV"
const scp_par_resetdata = "RESETDATA"
const scp_par_stopall = "STOPALL"
const scp_par_upgrade = "SYSUPGRADE"
const scp_par_lock = "LOCK"
const scp_par_unlock = "UNLOCK"
const scp_par_bfdata = "BFDATA"
const scp_par_loadbfdata = "LOADBFDATA"
const scp_par_restore = "RESTORE"
const scp_par_clenaperis = "CLEANPERIS"
const scp_par_setvolume = "SETVOLUME"

// const scp_par_version = "SYSVERSION"

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
const scp_job_msg = "MSG"

const scp_msg_cloro = "CLORO"
const scp_msg_meio = "MEIO"
const scp_msg_inoculo = "INOCULO"
const scp_msg_meio_inoculo = "MEIO-INOCULO"

const scp_bioreactor = "BIOREACTOR"
const scp_biofabrica = "BIOFABRICA"
const scp_totem = "TOTEM"
const scp_ibc = "IBC"
const scp_wdpanel = "WDPANEL"
const scp_intpanel = "INTPANEL"
const scp_screen = "SCREEN"
const scp_config = "CONFIG"
const scp_out = "OUT"
const scp_drop = "DROP"
const scp_clean = "CLEAN"
const scp_donothing = "NOTHING"
const scp_orch_addr = ":7007"
const scp_ipc_name = "/tmp/scp_master.sock"

const scp_refreshwait = 50
const scp_refresstatus = 15
const scp_refresscreens = 10 // em segundossss
const scp_refreshsleep = 100 // em ms
const scp_refreshsync = 5    // em segundos
const scp_timetosetup = 6    // intervalo no qual o setup_devices é executado em horas
const scp_timeout_ms = 2500
const scp_schedwait = 500
const scp_clockwait = 60 // em segundos
const scp_timetosave = 90
const scp_checksetup = 60
const scp_mustupdate_bio = 30
const scp_mustupdate_ibc = 45

const scp_timetocheckversion = 5 // em minutos
const scp_timewaitvalvs = 15000
const scp_timephwait_down = 10000 // Tempo que o ajuste de PH- e aplicado durante o cultivo
const scp_timephwait_up = 5000    // Tempo que o ajuste de PH+ e aplicado durante o cultivo
const scp_timetempwait = 10000
const scp_timewaitbeforeph = 10000
const scp_timegrowwait = 10000
const scp_maxtimewithdraw = 1800 // separar nas funcoes do JOB
const scp_timelinecip = 20       // em segundos
const time_cipline_blend = 30    // em segundos
const time_cipline_clean = 30    // em segundos
const scp_timeoutdefault = 60
const scp_maxwaitvolume = 30 // em minutos

const bio_deltatemp = 1.0 // variacao de temperatura maximo em percentual
const bio_deltaph = 0.0   // variacao de ph maximo em valor absoluto  -  ERA 0.1

const bio_withdrawstep = 50

const bio_ibctransftol = 50 // Na transferenciapara IBC, este é o volume acima do máximo permitido no IBC
const bio_deltavolzero = 33 // No withdraw, se nao variar em 25 segundos e for abaixo deste valor, zera o volume
const ibc_deltavolzero = 50 // idem para o IBC

const bio_diametro = 1530  // em mm   era 1430
const bio_v1_zero = 1483.0 // em mm
const bio_v2_zero = 1502.0 // em mm
const ibc_v1_zero = 2652.0 // em mm   2647
const ibc_v2_zero = 2652.0 // em mm

const flow_corfactor_out = 1.1
const flow_corfactor_in1 = 1.1
const flow_ratio = 0.03445 * flow_corfactor_out
const flow_ratio_in1 = 0.036525556 * flow_corfactor_in1

const bio_emptying_rate = 50.0 / 100.0

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

var status_codes = map[string]int{bio_cip: 1, bio_pause: 2, bio_wait: 3, bio_starting: 4, bio_loading: 5, bio_unloading: 6, bio_producting: 7,
	bio_empty: 8, bio_done: 9, bio_storing: 10, bio_error: 11, bio_ready: 12, bio_water: 13, bio_update: 14, bio_circulate: 15}

const bio_max_valves = 8
const bio_max_msg = 50
const bioreactor_max_msg = 7
const bio_max_foam = 4

const line_13 = "1_3"
const line_14 = "1_4"
const line_23 = "2_3"
const line_24 = "2_4"

const TEMPMAX = 80

type MsgReturn struct {
	Status  string
	Message string
}

type Biofabrica_data struct {
	BFId         string
	BFName       string
	Status       string
	CustomerId   string
	CustomerName string
	Address      string
	SWVersion    string
	LatLong      [2]float64
	LastUpdate   string
	BFIP         string
}

type Organism struct {
	Index      string
	Code       string
	Orgname    string
	Orgtype    string
	Lifetime   int
	Prodvol    int
	Cultmedium string
	Timetotal  int
	Temprange  string
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
	Temprunning  bool
	Emergpress   bool
	Continue     bool
	MainStatus   string
	UndoStatus   string
}

type Bioreact_ETL struct {
	BioreactorID string
	Status       string
	OrgCode      string
	Organism     string
	// Vol0         int
	// Vol1         int32
	// Vol2         int32
	VolInOut    float64
	Volume      uint32
	Level       uint8
	Pumpstatus  bool
	Aerator     bool
	Valvs       [8]int
	Perist      [5]int
	Heater      bool
	Temperature float32
	TempMax     float32
	PH          float32
	Step        [2]int
	Timeleft    [2]int
	Timetotal   [2]int
	Withdraw    uint32
	OutID       string
	Vol_zero    [2]float32
	LastStatus  string
	MustStop    bool
	MustPause   bool
	ShowVol     bool
	Messages    []string
	PHref       [3]float64
	RegresPH    [2]float64
	MainStatus  string
}

type IBC struct {
	IBCID        string
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
	MainStatus   string
	UndoStatus   string
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
	MainStatus string
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
	WaitList     []string
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
	PIntStatus   string
	POutStatus   string
	Critical     string
	Version      string
	LastVersion  string
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
	Volume  int
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
var runningsync = false
var runninggetall = false

var ibc_cfg map[string]IBC_cfg
var bio_cfg map[string]Bioreact_cfg
var totem_cfg map[string]Totem_cfg
var biofabrica_cfg map[string]Biofabrica_cfg
var paths map[string]Path
var valvs map[string]int
var organs map[string]Organism
var addrs_type map[string]DevAddrData
var schedule []Scheditem
var recipe_2000 []string
var recipe_1000 []string
var cipbio []string
var cipibc []string

var mainmutex sync.Mutex
var withdrawmutex sync.Mutex
var boardmutex sync.Mutex
var biomutex sync.Mutex
var waitlistmutex sync.Mutex
var upgrademutex sync.Mutex

var withdrawrunning = false

var mybf = Biofabrica_data{"bf999", "Nao Configurado", "ERRO", "HA", "Hubio Agro", "", "1.2.27", [2]float64{-15.9236672, -53.1827026}, "", "192.168.0.23"}

var bio = []Bioreact{
	{"BIOR01", bio_update, "", "", 0, 0, 0, 0, 0, 0, false, false, 0, [8]int{0, 0, 0, 0, 0, 0, 0, 0}, [5]int{0, 0, 0, 0, 0}, false, 0, 0, 0, [2]int{0, 0}, [2]int{0, 0}, [2]int{0, 0}, 0, "OUT", []string{}, []string{}, []string{}, []string{}, [2]float32{0, 0}, "", false, false, true, []string{}, [3]float64{0, 0, 0}, [2]float64{0, 0}, false, false, false, "", ""},
	{"BIOR02", bio_update, "", "", 0, 0, 0, 0, 0, 0, false, false, 0, [8]int{0, 0, 0, 0, 0, 0, 0, 0}, [5]int{0, 0, 0, 0, 0}, false, 0, 0, 0, [2]int{0, 0}, [2]int{0, 0}, [2]int{0, 0}, 0, "OUT", []string{}, []string{}, []string{}, []string{}, [2]float32{0, 0}, "", false, false, true, []string{}, [3]float64{0, 0, 0}, [2]float64{0, 0}, false, false, false, "", ""},
	{"BIOR03", bio_update, "", "", 0, 0, 0, 0, 0, 0, false, false, 0, [8]int{0, 0, 0, 0, 0, 0, 0, 0}, [5]int{0, 0, 0, 0, 0}, false, 0, 0, 0, [2]int{0, 0}, [2]int{0, 0}, [2]int{0, 0}, 0, "OUT", []string{}, []string{}, []string{}, []string{}, [2]float32{0, 0}, "", false, false, true, []string{}, [3]float64{0, 0, 0}, [2]float64{0, 0}, false, false, false, "", ""},
	{"BIOR04", bio_update, "", "", 0, 0, 0, 0, 0, 0, false, false, 0, [8]int{0, 0, 0, 0, 0, 0, 0, 0}, [5]int{0, 0, 0, 0, 0}, false, 0, 0, 0, [2]int{0, 0}, [2]int{0, 0}, [2]int{0, 0}, 0, "OUT", []string{}, []string{}, []string{}, []string{}, [2]float32{0, 0}, "", false, false, true, []string{}, [3]float64{0, 0, 0}, [2]float64{0, 0}, false, false, false, "", ""},
	{"BIOR05", bio_update, "", "", 0, 0, 0, 0, 0, 0, false, false, 0, [8]int{0, 0, 0, 0, 0, 0, 0, 0}, [5]int{0, 0, 0, 0, 0}, false, 0, 0, 0, [2]int{0, 0}, [2]int{0, 0}, [2]int{0, 0}, 0, "OUT", []string{}, []string{}, []string{}, []string{}, [2]float32{0, 0}, "", false, false, true, []string{}, [3]float64{0, 0, 0}, [2]float64{0, 0}, false, false, false, "", ""},
	{"BIOR06", bio_update, "", "", 0, 0, 0, 0, 0, 0, false, false, 0, [8]int{0, 0, 0, 0, 0, 0, 0, 0}, [5]int{0, 0, 0, 0, 0}, false, 0, 0, 0, [2]int{0, 0}, [2]int{0, 0}, [2]int{0, 0}, 0, "OUT", []string{}, []string{}, []string{}, []string{}, [2]float32{0, 0}, "", false, false, true, []string{}, [3]float64{0, 0, 0}, [2]float64{0, 0}, false, false, false, "", ""},
}

var ibc = []IBC{
	{"IBC01", bio_update, "", "", 0, 0, 0, 0, 0, 0, false, [4]int{0, 0, 0, 0}, [2]int{0, 0}, [2]int{0, 0}, 0, "OUT", [2]float32{0, 0}, false, false, false, []string{}, []string{}, []string{}, []string{}, "", true, 0, "", ""},
	{"IBC02", bio_update, "", "", 0, 0, 0, 0, 0, 0, false, [4]int{0, 0, 0, 0}, [2]int{0, 0}, [2]int{0, 0}, 0, "OUT", [2]float32{0, 0}, false, false, false, []string{}, []string{}, []string{}, []string{}, "", true, 0, "", ""},
	{"IBC03", bio_update, "", "", 0, 0, 0, 0, 0, 0, false, [4]int{0, 0, 0, 0}, [2]int{0, 0}, [2]int{0, 0}, 0, "OUT", [2]float32{0, 0}, false, false, false, []string{}, []string{}, []string{}, []string{}, "", true, 0, "", ""},
	{"IBC04", bio_update, "", "", 0, 0, 0, 0, 0, 0, false, [4]int{0, 0, 0, 0}, [2]int{0, 0}, [2]int{0, 0}, 0, "OUT", [2]float32{0, 0}, false, false, false, []string{}, []string{}, []string{}, []string{}, "", true, 0, "", ""},
	{"IBC05", bio_update, "", "", 0, 0, 0, 0, 0, 0, false, [4]int{0, 0, 0, 0}, [2]int{0, 0}, [2]int{0, 0}, 0, "OUT", [2]float32{0, 0}, false, false, false, []string{}, []string{}, []string{}, []string{}, "", true, 0, "", ""},
	{"IBC06", bio_update, "", "", 0, 0, 0, 0, 0, 0, false, [4]int{0, 0, 0, 0}, [2]int{0, 0}, [2]int{0, 0}, 0, "OUT", [2]float32{0, 0}, false, false, false, []string{}, []string{}, []string{}, []string{}, "", true, 0, "", ""},
	{"IBC07", bio_update, "", "", 0, 0, 0, 0, 0, 0, false, [4]int{0, 0, 0, 0}, [2]int{0, 0}, [2]int{0, 0}, 0, "OUT", [2]float32{0, 0}, false, false, false, []string{}, []string{}, []string{}, []string{}, "", true, 0, "", ""},
}

var totem = []Totem{
	{"TOTEM01", bio_ready, false, [2]int{0, 0}, [4]int{0, 0, 0, 0}},
	{"TOTEM02", bio_ready, false, [2]int{0, 0}, [4]int{0, 0, 0, 0}},
}

var biofabrica = Biofabrica{
	"BIOFABRICA001", [9]int{0, 0, 0, 0, 0, 0, 0, 0, 0}, false, []string{}, []string{}, scp_ready, 0, 0, 0, 0, 0, 0, false, false, true, "", "", "", "", "",
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
	if sumXX <= 0 {
		fmt.Println("ERROR estimateB0B1: Valor invalido de sumXX", sumXX)
		return 0, 0
	}
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
	if n == 0 {
		return 0
	}
	c := int(n / 2)
	if n >= 5 {
		s := x[c-1] + x[c] + x[c+1]
		mediana = s / 3.0
	} else if n > 0 {
		mediana = x[c]
	} else {
		fmt.Println("ERROR CALC MEDIANA: Numero invalido de amostras para calculo de mediada", n, x)
		return 0
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

func get_tun_ip() string {
	tun_ip := ""
	cmdpath, _ := filepath.Abs("/sbin/ifconfig")
	cmd := exec.Command(cmdpath, "tun0") // "| grep 'inet ' | awk '{ print $2}'")
	// cmd := exec.Command(cmdpath)
	// cmd.Dir = "/sbin/"
	output, err := cmd.CombinedOutput()
	if err != nil {
		checkErr(err)
	} else {
		out_str := string(output)
		p := strings.Index(out_str, "inet")
		if p >= 0 {
			ret := scp_splitparam(out_str[p:], " ")
			if len(ret) > 1 {
				tun_ip = ret[1]
			}
		}

	}
	return tun_ip
}

func get_bf_status() string {
	ok := true
	for _, b := range bio {
		if b.Status == bio_error {
			ok = false
		}
	}
	for _, b := range ibc {
		if b.Status == bio_error {
			ok = false
		}
	}
	for _, t := range totem {
		if t.Status == bio_error {
			ok = false
		}
	}
	if biofabrica.Status != scp_ready {
		ok = false
	}
	if ok {
		return bio_ready
	}
	return bio_error
}

func load_bf_data(filename string) int {
	mybf_new := Biofabrica_data{}
	file, err := os.Open(filename)
	if err != nil {
		checkErr(err)
		return 0
	}
	defer file.Close()
	csvr := csv.NewReader(file)
	n := 0
	for {
		r, err := csvr.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			if perr, ok := err.(*csv.ParseError); ok && perr.Err != csv.ErrFieldCount {
				checkErr(err)
				break
			}
		}
		// fmt.Println(r)
		if r[0][0] != '#' {
			mybf_new.BFId = r[0]
			mybf_new.BFName = r[1]
			mybf_new.Status = r[2]
			mybf_new.CustomerId = r[3]
			mybf_new.CustomerName = r[4]
			mybf_new.Address = r[5]
			mybf_new.SWVersion = r[6]
			lat_str := r[7]
			long_str := r[8]
			lat_f, err_lat := strconv.ParseFloat(lat_str, 64)
			if err_lat != nil {
				checkErr(err_lat)
			}
			long_f, err_long := strconv.ParseFloat(long_str, 64)
			if err_long != nil {
				checkErr(err_long)
			}
			if err_lat == nil && err_long == nil {
				mybf_new.LatLong = [2]float64{lat_f, long_f}
			} else {
				mybf_new.LatLong = [2]float64{0, 0}
			}
			mybf_new.LastUpdate = r[9]
			mybf_new.BFIP = r[10]
			n++
		}
		if n > 0 {
			break
		}
	}
	if n > 0 {
		mybf = mybf_new
		mybf.SWVersion = biofabrica.Version
		mybf.Status = get_bf_status()
	}
	return n
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
		ind := r[0]
		code := r[1]
		name := r[2]
		otype := r[3]
		lifetime, _ := strconv.Atoi(strings.Replace(r[4], " ", "", -1))
		volume, _ := strconv.Atoi(strings.Replace(r[5], " ", "", -1))
		medium := strings.Replace(r[6], " ", "", -1)
		tottime, _ := strconv.Atoi(strings.Replace(r[7], " ", "", -1))
		temprange := strings.Replace(r[8], " ", "", -1)
		aero1, _ := strconv.Atoi(strings.Replace(r[9], " ", "", -1))
		aero2, _ := strconv.Atoi(strings.Replace(r[10], " ", "", -1))
		aero3, _ := strconv.Atoi(strings.Replace(r[11], " ", "", -1))
		ph1 := strings.Replace(r[12], " ", "", -1)
		ph2 := strings.Replace(r[13], " ", "", -1)
		ph3 := strings.Replace(r[14], " ", "", -1)
		org := Organism{ind, code, name, otype, lifetime, volume, medium, tottime, temprange, [3]int{aero1, aero2, aero3}, [3]string{ph1, ph2, ph3}}
		organs[code] = org
		totalrecords = k
	}
	fmt.Println("DEBUG LOAD ORGANISMS: Organismos lidos:", organs)
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
			old, ok = addrs_type[screen_addr]
			if !ok {
				addrs_type[screen_addr] = DevAddrData{screen_addr, scp_screen, id}
			} else {
				fmt.Println("ERROR LOAD BIOS CONF: ADDR", screen_addr, " já cadastrado na tabela de devices com tipo", old)
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

func save_ibcs_conf(filename string) int {
	filecsv, err := os.Create(filename)
	if err != nil {
		checkErr(err)
		return -1
	}
	defer filecsv.Close()
	n := 0
	csvwriter := csv.NewWriter(filecsv)
	for _, b := range ibc_cfg {
		s := fmt.Sprintf("%s,%s,%s,%d,%s,", b.IBCID, b.Deviceaddr, b.Screenaddr, b.Maxvolume, b.Pump_dev)
		for _, p := range b.Valv_devs {
			s += p + ","
		}
		s += fmt.Sprintf("%s,%s,%s", b.Vol_devs[0], b.Vol_devs[1], b.Levellow)
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

func save_totems_conf(filename string) int {
	filecsv, err := os.Create(filename)
	if err != nil {
		checkErr(err)
		return -1
	}
	defer filecsv.Close()
	n := 0
	csvwriter := csv.NewWriter(filecsv)
	for _, b := range totem_cfg {
		s := fmt.Sprintf("%s,%s,%s,", b.TotemID, b.Deviceaddr, b.Pumpdev)
		for _, p := range b.Peris_dev {
			s += p + ","
		}
		s += fmt.Sprintf("%s,%s", b.Valv_devs[0], b.Valv_devs[1])
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

func save_bf_conf(filename string) int {
	filecsv, err := os.Create(filename)
	if err != nil {
		checkErr(err)
		return -1
	}
	defer filecsv.Close()
	n := 0
	csvwriter := csv.NewWriter(filecsv)
	for _, b := range biofabrica_cfg {
		s := fmt.Sprintf("%s,%s,%s", b.DeviceID, b.Deviceaddr, b.Deviceport)
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
	_, ok := biofabrica_cfg["FBF02"]
	if !ok {
		vbf01, okt := biofabrica_cfg["VBF01"]
		if okt {
			board_add_message("AFluxometro de entrada não presente nas configurações. Corrigindo problema automaticamente. Favor verificar configurações da Biofábrica", "")
			biofabrica_cfg["FBF02"] = Biofabrica_cfg{"FBF02", vbf01.Deviceaddr, "C7"}
		} else {
			board_add_message("EATENÇÃO: Fluxometro de entrada e Válvula 01 não presentes nas configurações. Acionar time de suporte", "")
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
			// fmt.Println("PATH id=", path_id, "path=", paths[path_id])
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
	var devaddr, valvaddr string
	var ind int
	id := devid + "/" + valvid
	valvs[id] = value
	switch devtype {
	case scp_donothing:
		return true
	case scp_bioreactor:
		ind = get_bio_index(devid)
		devaddr = bio_cfg[devid].Deviceaddr
		// scraddr = bio_cfg[devid].Screenaddr
		if ind >= 0 {
			v, err := strconv.Atoi(valvid[1:])
			if err == nil {
				// bio[ind].Valvs[v-1] = value
				valvaddr = bio_cfg[devid].Valv_devs[v-1]
				// valve_scrstr = fmt.Sprintf("S%d", v+200)
			} else {
				fmt.Println("ERROR SET VAL: id da valvula nao inteiro", valvid)
				return false
			}
		} else {
			fmt.Println("ERROR SET VAL: BIORREATOR nao encontrado", devid)
			return false
		}
	case scp_ibc:
		ind = get_ibc_index(devid)
		devaddr = ibc_cfg[devid].Deviceaddr
		// scraddr = ""
		if ind >= 0 {
			v, err := strconv.Atoi(valvid[1:])
			if err == nil {
				// ibc[ind].Valvs[v-1] = value
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
		ind = get_totem_index(devid)
		devaddr = totem_cfg[devid].Deviceaddr
		// scraddr = ""
		if ind >= 0 {
			v, err := strconv.Atoi(valvid[1:])
			if err == nil {
				// totem[ind].Valvs[v-1] = value
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
		// scraddr = ""
		valvaddr = biofabrica_cfg[valvid].Deviceport
		v, err := strconv.Atoi(valvid[3:])
		if err == nil {
			biofabrica.Valvs[v-1] = value
		} else {
			fmt.Println("ERROR SET VAL: BIOFABRICA - id da valvula nao inteiro", valvid)
			return false
		}
	}
	setok := true
	if value == 0 || value == 1 {
		cmd1 := fmt.Sprintf("CMD/%s/PUT/%s,%d/END", devaddr, valvaddr, value)
		// fmt.Println(cmd1)
		ret1 := scp_sendmsg_orch(cmd1)
		fmt.Println("DEBUG SET VALV STATUS: cmd=", cmd1, " ret=", ret1)
		if !strings.Contains(ret1, scp_ack) && !devmode && biofabrica.Critical != scp_netfail {
			// Mudança para testar falha no equipamento
			if (devtype == scp_bioreactor && bio[ind].Status != bio_error) || (devtype == scp_ibc && ibc[ind].Status != bio_error) || (devtype == scp_totem && totem[ind].Status != bio_error) ||
				(devtype == scp_biofabrica && biofabrica.Status != scp_fail) {
				fmt.Println("DEBUG SET VALV STATUS: Mudando setok para false")
				setok = false
			}
			// fmt.Println("ERROR SET VALV STATUS: SEND MSG ORCH falhou", ret1)
		}
	}
	if setok || true {
		valvs[id] = value
		switch devtype {
		case scp_bioreactor:
			if ind >= 0 {
				v, _ := strconv.Atoi(valvid[1:])
				bio[ind].Valvs[v-1] = value
			}
		case scp_ibc:
			if ind >= 0 {
				v, _ := strconv.Atoi(valvid[1:])
				ibc[ind].Valvs[v-1] = value
			}
		case scp_totem:
			if ind >= 0 {
				v, _ := strconv.Atoi(valvid[1:])
				totem[ind].Valvs[v-1] = value
			}
		case scp_biofabrica:
			v, err := strconv.Atoi(valvid[3:])
			if err == nil {
				biofabrica.Valvs[v-1] = value
			}
		}
	} else {
		fmt.Println("DEBUG SET VALV STATUS: SETOK false para", valvid)
	}
	// if len(scraddr) > 0 {
	// 	cmd2 := fmt.Sprintf("CMD/%s/PUT/%s,%d/END", scraddr, valve_scrstr, value)
	// 	fmt.Println(cmd2)
	// 	ret2 := scp_sendmsg_orch(cmd2)
	// 	fmt.Println("RET CMD2 =", ret2)
	// }
	return setok
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

func save_bf_data(filename string) int {
	mybf.Status = get_bf_status()
	mybf.SWVersion = biofabrica.Version
	filecsv, err := os.Create(filename)
	if err != nil {
		checkErr(err)
		return -1
	}
	defer filecsv.Close()
	n := 0
	csvwriter := csv.NewWriter(filecsv)
	thisbf := mybf
	s := fmt.Sprintf("%s,%s,%s,%s,%s,%s,%s,%f,%f,%s,%s", thisbf.BFId, thisbf.BFName, thisbf.Status, thisbf.CustomerId,
		thisbf.CustomerName, thisbf.Address, thisbf.SWVersion, thisbf.LatLong[0], thisbf.LatLong[1],
		thisbf.LastUpdate, thisbf.BFIP)
	csvstr := scp_splitparam(s, ",")
	// fmt.Println("DEBUG SAVE", csvstr)
	err = csvwriter.Write(csvstr)
	if err != nil {
		checkErr(err)
	} else {
		n++
	}
	csvwriter.Flush()
	return n
}

func save_one_data(filename string, datastr []byte) bool {
	var err error
	if test_file(filename) {
		finfo, _ := os.Stat(filename)
		if finfo.Size() > 0 {
			if test_file(filename + ".bak") {
				err = os.Remove(filename + ".bak")
				checkErr(err)
			}
			err = os.Rename(filename, filename+".bak")
			checkErr(err)
		} else {
			fmt.Println("ERROR SAVE ON DATA: Arquivo original IGNORADO, tamanho ZERO", filename)
		}
	}
	err = os.WriteFile(filename, []byte(datastr), 0644)
	if err != nil {
		checkErr(err)
		return false
	}
	return true
}

func save_all_data(filename string) int {
	var buf []byte

	fmt.Println("DEBUG SAVE ALL: Salvando todos os dados")

	ok := true
	buf, _ = json.Marshal(bio)
	if !save_one_data(localconfig_path+filename+"_bio.json", buf) {
		ok = false
	}

	buf, _ = json.Marshal(ibc)
	if !save_one_data(localconfig_path+filename+"_ibc.json", buf) {
		ok = false
	}

	buf, _ = json.Marshal(totem)
	if !save_one_data(localconfig_path+filename+"_totem.json", buf) {
		ok = false
	}

	buf, _ = json.Marshal(biofabrica)
	if !save_one_data(localconfig_path+filename+"_biofabrica.json", buf) {
		ok = false
	}

	buf, _ = json.Marshal(schedule)
	if !save_one_data(localconfig_path+filename+"_schedule.json", buf) {
		ok = false
	}

	save_bf_data(localconfig_path + "bf_data.csv")

	if !ok {
		board_add_message("EATENÇÃO: Falha ao gravar arquivos de segurança. Favor contactar SAC", "FAILSAVEALL")
	} else {
		board_del_message("FAILSAVEALL")
	}

	fmt.Println("DEBUG SAVE ALL: Executando SYNC do Sistema Operacional")

	cmdpath, _ := filepath.Abs("/usr/bin/sync")
	// cmd := exec.Command(cmdpath, "restart", "scp_orch")
	cmd := exec.Command(cmdpath)
	cmd.Dir = execpath
	output, err := cmd.CombinedOutput()
	if len(output) > 0 {
		fmt.Println("DEBUG SAVE ALL: SYNC OUPUT", string(output))
	}
	if err != nil {
		checkErr(err)
		fmt.Println("ERROR SAVE ALL: Falha ao Executar SYNC")
	}

	return 0
}

func load_all_data(filename string) int {
	var dat []byte
	var err error

	ok := true
	usebak := false
	if test_file(localconfig_path + filename + "_bio.json") {
		dat, err = os.ReadFile(localconfig_path + filename + "_bio.json")
	} else {
		dat, err = os.ReadFile(localconfig_path + filename + "_bio.json.bak")
		usebak = true
	}
	if err == nil {
		err = json.Unmarshal([]byte(dat), &bio)
		if err == nil {
			fmt.Println("DEBUG LOAD ALL DATA: Sucesso ao recuperar dumpdata de Biorreatores.")
		} else {
			fmt.Println("ERROR LOAD ALL DATA: Falha ao recuperar dumpdata de Biorreatores. Dados corrompidos.")
			ok = false
		}
		// fmt.Println("-- bio data = ", bio)
	} else {
		fmt.Println("ERROR LOAD ALL DATA: Nao foi possivel ler dumpdata de Biorreatores! Usando valores dummy.")
		checkErr(err)
		ok = false
	}

	if test_file(localconfig_path + filename + "_ibc.json") {
		dat, err = os.ReadFile(localconfig_path + filename + "_ibc.json")
	} else {
		dat, err = os.ReadFile(localconfig_path + filename + "_ibc.json.bak")
		usebak = true
	}
	if err == nil {
		err = json.Unmarshal([]byte(dat), &ibc)
		if err == nil {
			fmt.Println("DEBUG LOAD ALL DATA: Sucesso ao recuperar dumpdata de IBCs.")
		} else {
			fmt.Println("ERROR LOAD ALL DATA: Falha ao recuperar dumpdata de IBCs. Dados corrompidos.")
			ok = false
		}
		// fmt.Println("-- ibc data = ", ibc)
	} else {
		fmt.Println("ERROR LOAD ALL DATA: Nao foi possivel ler dumpdata de IBCs! Usando valores dummy.")
		checkErr(err)
		ok = false
	}

	if test_file(localconfig_path + filename + "_totem.json") {
		dat, err = os.ReadFile(localconfig_path + filename + "_totem.json")
	} else {
		dat, err = os.ReadFile(localconfig_path + filename + "_totem.json.bak")
		usebak = true
	}
	if err == nil {
		err = json.Unmarshal([]byte(dat), &totem)
		// fmt.Println("-- totem data = ", totem)
		if err == nil {
			fmt.Println("DEBUG LOAD ALL DATA: Sucesso ao recuperar dumpdata de Totems.")
		} else {
			fmt.Println("ERROR LOAD ALL DATA: Falha ao recuperar dumpdata de Totems. Dados corrompidos.")
			ok = false
		}
	} else {
		fmt.Println("ERROR LOAD ALL DATA: Nao foi possivel ler dumpdata de Totems! Usando valores dummy.")
		checkErr(err)
		ok = false
	}

	if test_file(localconfig_path + filename + "_biofabrica.json") {
		dat, err = os.ReadFile(localconfig_path + filename + "_biofabrica.json")
	} else {
		dat, err = os.ReadFile(localconfig_path + filename + "_biofabrica.json.bak")
		usebak = true
	}
	if err == nil {
		err = json.Unmarshal([]byte(dat), &biofabrica)
		// fmt.Println("-- biofabrica data = ", biofabrica)
		if err == nil {
			fmt.Println("DEBUG LOAD ALL DATA: Sucesso ao recuperar dumpdata da Biofabrica.")
		} else {
			fmt.Println("ERROR LOAD ALL DATA: Falha ao recuperar dumpdata da Biofabrica. Dados corrompidos.")
			ok = false
		}
	} else {
		fmt.Println("ERROR LOAD ALL DATA: Nao foi possivel ler dumpdata da Biofabrica! Usando valores dummy.")
		checkErr(err)
		ok = false
	}

	set_allvalvs_status()

	if test_file(localconfig_path + filename + "_schedule.json") {
		dat, err = os.ReadFile(localconfig_path + filename + "_schedule.json")
	} else {
		dat, err = os.ReadFile(localconfig_path + filename + "_schedule.json.bak")
		usebak = true
	}
	if err == nil {
		err = json.Unmarshal([]byte(dat), &schedule)
		// fmt.Println("-- schedule data = ", schedule)
		if err == nil {
			fmt.Println("DEBUG LOAD ALL DATA: Sucesso ao recuperar dumpdata de Schedule.")
		} else {
			fmt.Println("ERROR LOAD ALL DATA: Falha ao recuperar dumpdata de Schedule. Dados corrompidos.")
			ok = false
		}
	} else {
		fmt.Println("ERROR LOAD ALL DATA: Nao foi possivel ler dumpdata de Schedule! Usando VAZIO.")
		checkErr(err)
		ok = false
	}

	if !ok {
		board_add_message("EATENÇÃO: Falha ao ler arquivos de recuperação. Favor contactar SAC", "FAILLOADALL")
	} else {
		board_del_message("FAILLOADALL")
	}

	if usebak {
		board_add_message("AATENÇÃO: Arquivos de recuperação não encontrados. Utilizando cópias de segurança. Favor contactar SAC", "BAKLOADALL")
	} else {
		board_del_message("BACKLOADALL")
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
	board_add_message("ERETORNANDO de PARADA TOTAL", "")
	if biofabrica.Critical != scp_sysstop {
		board_add_message("ANecessário aguardar até 10 minutos até reestabelecimento dos equipamentos", "")
		if !devmode {
			time.Sleep(300 * time.Second)
		}
	} else {
		if !devmode {
			time.Sleep(120 * time.Second)
		}
	}
	if biofabrica.Critical == scp_netfail {
		biofabrica.Critical = scp_recovery
	}
	scp_setup_devices(true)
	for _, b := range bio {
		ind := get_bio_index(b.BioreactorID)
		if bio[ind].Status != bio_nonexist && bio[ind].Status != bio_error {
			pause_device(scp_bioreactor, bio[ind].BioreactorID, false)
		} else if bio[ind].Status == bio_error {
			board_add_message("AFavor checar Biorreator "+b.BioreactorID, "")
		}
	}
	for _, b := range ibc {
		ind := get_ibc_index(b.IBCID)
		if ibc[ind].Status != bio_nonexist && ibc[ind].Status != bio_error {
			pause_device(scp_ibc, ibc[ind].IBCID, false)
		} else if ibc[ind].Status == bio_error {
			board_add_message("AFavor checar IBC "+ibc[ind].IBCID, "")
		}
	}
	if !runningsync {
		go scp_sync_functions()
	}
	if !runninggetall {
		go scp_get_alldata()
	}
	// if !schedrunning {
	// 	time.Sleep(60 * time.Second)
	// 	go scp_scheduler()
	// }
}

func scp_emergency_pause() {
	fmt.Println("\n\nCRITICAL EMERGENCY PAUSE: Executando EMERGENCY PAUSE da Biofabrica")
	if biofabrica.Critical == scp_sysstop {
		board_add_message("EPARADA de TOTAL para MANUTENÇÃO do SISTEMA em PROGRESSO", "")
	} else {
		board_add_message("EPARADA de TOTAL em PROGRESSO", "")
	}
	for _, b := range bio {
		ind := get_bio_index(b.BioreactorID)
		if ind >= 0 {
			if bio[ind].Status != bio_nonexist {
				pause_device(scp_bioreactor, bio[ind].BioreactorID, true)
				bio[ind].Withdraw = 0
			}

		}

	}
	for _, b := range ibc {
		ind := get_ibc_index(b.IBCID)
		if ind >= 0 {
			if ibc[ind].Status != bio_nonexist {
				pause_device(scp_ibc, ibc[ind].IBCID, true)
				ibc[ind].Withdraw = 0
			}
		}

	}
}

func scp_check_lastversion() {
	var last_biofabrica Biofabrica

	fmt.Println("DEBUG CHECK LASTVERSION: Checando ultima versao do software")
	res, err := http.Get("http://biofabrica-main.hubioagro.com.br/biofabrica_view")
	// fmt.Println("RES=", res)
	if err != nil {
		checkErr(err)
		return
	}

	// fmt.Println(res)
	rdata, err := ioutil.ReadAll(res.Body)
	if err != nil {
		checkErr(err)
		return
	}
	// fmt.Println(string(rdata))
	json.Unmarshal(rdata, &last_biofabrica)
	// fmt.Println(last_biofabrica)
	fmt.Println("DEBUG CHECK LASTVERSION: Ajustando ultima versão para:", last_biofabrica.Version)
	biofabrica.LastVersion = last_biofabrica.Version
}

func scp_check_network() {
	fmt.Println("DEBUG SCP CHECK NETWORK: Iniciando CHECK, aguardando 60 segundos")
	time.Sleep(60 * time.Second)

	for {
		if biofabrica.Critical == scp_stopall {
			return
		}
		fmt.Println("DEBUG CHECK NETWORK: Testando comunicacao com MAINROUTER", mainrouter, pingmax)
		if !tcp_host_isalive(mainrouter, "80", pingmax) {
			if biofabrica.Critical != scp_netfail {
				fmt.Println("FATAL CHECK NETWORK: Sem comunicacao com MAINROUTER", mainrouter)
				biofabrica.Critical = scp_stopall
				save_all_data(data_filename)
				scp_emergency_pause()
				time.Sleep(30 * time.Second)
				biofabrica.Critical = scp_netfail
				save_all_data(data_filename)
			}
		} else {
			fmt.Println("DEBUG CHECK NETWORK: OK comunicacao com MAINROUTER", mainrouter)
			if biofabrica.Critical == scp_netfail || biofabrica.Critical == scp_sysstop {
				if finishedsetup {
					scp_run_recovery()
					biofabrica.Critical = scp_ready // estava depois do recovery
					// scp_setup_devices(true)
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

func board_has_message(id string) bool {
	boardmutex.Lock()
	defer boardmutex.Unlock()
	msg_id := fmt.Sprintf("{%s}", id)
	for _, m := range biofabrica.Messages {
		if strings.Contains(m, msg_id) {
			return true
		}
	}
	return false
}

func board_del_message(id string) bool {
	boardmutex.Lock()
	defer boardmutex.Unlock()
	msg_id := fmt.Sprintf("{%s}", id)
	// fmt.Println("DEBUG BOARD DEL MESSAGE: board inicial=", len(biofabrica.Messages), biofabrica.Messages)
	has_del := false
	for i := 0; i < len(biofabrica.Messages); i++ {
		if strings.Contains(biofabrica.Messages[i], msg_id) {
			has_del = true
			m1 := []string{}
			if i > 0 {
				m1 = biofabrica.Messages[:i]
			}
			m2 := []string{}
			if i < len(biofabrica.Messages)-1 {
				m2 = biofabrica.Messages[i+1:]
			}
			// fmt.Println("DEBUG BOARD DEL MESSAGE: m1=", len(m1), m1, " m2=", len(m2), m2)
			biofabrica.Messages = append(m1, m2...)
		}
	}
	// fmt.Println("DEBUG BOARD DEL MESSAGE: board final=", len(biofabrica.Messages), biofabrica.Messages)
	return has_del
}

func board_add_message(m string, id string) bool {
	if len(id) > 0 && board_has_message(id) {
		return false
	}
	boardmutex.Lock()
	defer boardmutex.Unlock()
	msg_id := id
	if len(id) == 0 {
		msg_id = "0"
	}
	n := len(biofabrica.Messages)
	stime := time.Now().Format("15:04 02/01")
	m_new := strings.Replace(m, "OUT", "Desenvase", -1)
	m_new2 := strings.Replace(m_new, "DROP", "Descarte", -1)
	msg := fmt.Sprintf("%c%s [%s]{%s}", m_new2[0], m_new2[1:], stime, msg_id)
	if n < bio_max_msg {
		biofabrica.Messages = append(biofabrica.Messages, msg)
	} else {
		biofabrica.Messages = append(biofabrica.Messages[2:], msg)
	}
	return true
}

func bio_has_message(bioid string, id string) bool {
	ind := get_bio_index(bioid)
	if ind < 0 {
		return false
	}
	biomutex.Lock()
	defer biomutex.Unlock()
	msg_id := fmt.Sprintf("{%s}", id)
	for _, m := range bio[ind].Messages {
		if strings.Contains(m, msg_id) {
			return true
		}
	}
	return false
}

func bio_del_message(bioid string, id string) bool {
	ind := get_bio_index(bioid)
	if ind < 0 {
		return false
	}
	biomutex.Lock()
	defer biomutex.Unlock()
	msg_id := fmt.Sprintf("{%s}", id)
	// fmt.Println("DEBUG BOARD DEL MESSAGE: board inicial=", len(bio[ind].Messages), bio[ind].Messages)
	has_del := false
	for i := 0; i < len(bio[ind].Messages); i++ {
		if strings.Contains(bio[ind].Messages[i], msg_id) {
			has_del = true
			m1 := []string{}
			if i > 0 {
				m1 = bio[ind].Messages[:i]
			}
			m2 := []string{}
			if i < len(bio[ind].Messages)-1 {
				m2 = bio[ind].Messages[i+1:]
			}
			// fmt.Println("DEBUG BOARD DEL MESSAGE: m1=", len(m1), m1, " m2=", len(m2), m2)
			bio[ind].Messages = append(m1, m2...)
		}
	}
	// fmt.Println("DEBUG BOARD DEL MESSAGE: board final=", len(bio[ind].Messages), bio[ind].Messages)
	return has_del
}

func bio_add_message(bioid string, m string, id string) bool {
	ind := get_bio_index(bioid)
	if ind < 0 {
		fmt.Println("ERROR BIO ADD MESSAGE: Biorreator nao existe", bioid)
		return false
	}
	if len(id) > 0 && bio_has_message(bioid, id) {
		return false
	}
	biomutex.Lock()
	defer biomutex.Unlock()
	msg_id := id
	if len(id) == 0 {
		msg_id = "0"
	}
	n := len(bio[ind].Messages)
	stime := time.Now().Format("15:04 02/01")
	m_new := strings.Replace(m, "OUT", "Desenvase", -1)
	m_new = strings.Replace(m, "DROP", "Descarte", -1)
	msg := fmt.Sprintf("%c%s [%s]{%s}", m_new[0], m_new[1:], stime, msg_id)
	if n < bioreactor_max_msg {
		bio[ind].Messages = append(bio[ind].Messages, msg)
	} else {
		bio[ind].Messages = append(bio[ind].Messages[1:], msg)
	}
	return true
}

func waitlist_has_message(id string) bool {
	waitlistmutex.Lock()
	defer waitlistmutex.Unlock()
	msg_id := fmt.Sprintf("{%s}", id)
	for _, m := range biofabrica.WaitList {
		if strings.Contains(m, msg_id) {
			return true
		}
	}
	return false
}

func waitlist_del_message(id string) {
	waitlistmutex.Lock()
	defer waitlistmutex.Unlock()
	msg_id := fmt.Sprintf("{%s", id)
	// fmt.Println("DEBUG BOARD DEL MESSAGE: board inicial=", len(biofabrica.Messages), biofabrica.Messages)
	for {
		has_del := false
		for i := 0; i < len(biofabrica.WaitList); i++ {
			if strings.Contains(biofabrica.WaitList[i], msg_id) {
				has_del = true
				m1 := []string{}
				if i > 0 {
					m1 = biofabrica.WaitList[:i]
				}
				m2 := []string{}
				if i < len(biofabrica.WaitList)-1 {
					m2 = biofabrica.WaitList[i+1:]
				}
				// fmt.Println("DEBUG BOARD DEL MESSAGE: m1=", len(m1), m1, " m2=", len(m2), m2)
				biofabrica.WaitList = append(m1, m2...)
				break
			}
		}
		if !has_del {
			break
		}
	}
	// fmt.Println("DEBUG BOARD DEL MESSAGE: board final=", len(biofabrica.Messages), biofabrica.Messages)
}

func waitlist_add_message(m string, id string) bool {
	if len(id) > 0 && waitlist_has_message(id) {
		return false
	}
	waitlistmutex.Lock()
	defer waitlistmutex.Unlock()
	msg_id := id
	if len(id) == 0 {
		msg_id = "0"
	}
	n := len(biofabrica.WaitList)
	stime := time.Now().Format("15:04")
	m_new := strings.Replace(m, "OUT", "Desenvase", -1)
	m_new2 := strings.Replace(m_new, "DROP", "Descarte", -1)
	msg := fmt.Sprintf("%c%s [%s]{%s}", m_new2[0], m_new2[1:], stime, msg_id)
	if n < bio_max_msg {
		biofabrica.WaitList = append(biofabrica.WaitList, msg)
	} else {
		biofabrica.WaitList = append(biofabrica.WaitList[2:], msg)
	}
	return true
}

func scp_setup_devices(mustall bool) {
	if demo {
		return
	}
	fmt.Println("\n\nDEBUG SETUP DEVICES: Configurando BIORREATORES")
	for _, b := range bio_cfg {
		// if biofabrica.Critical == scp_netfail {
		// 	break
		// }
		ind := get_bio_index(b.BioreactorID)
		bioexist := true
		if ind >= 0 {
			if bio_cfg[b.BioreactorID].Deviceaddr == bio_noaddr {
				bioexist = false
				bio[ind].Status = bio_nonexist
			}
		}
		if bioexist && len(b.Deviceaddr) > 0 && ind >= 0 {
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
				cmd = append(cmd, "CMD/"+bioaddr+"/MOD/"+scp_magicport+",3/END")
				cmd = append(cmd, "CMD/"+bioaddr+"/PUT/"+scp_magicport+","+scp_magicvalue+"/END")
				// cmd = append(cmd, "CMD/"+b.Screenaddr+"/PUT/S200,1/END")
				nerr := 0
				for k, c := range cmd {
					ret := scp_sendmsg_orch(c)
					if !strings.Contains(ret, scp_ack) {
						nerr++
					}
					fmt.Println(b.BioreactorID, "err=", nerr, k, c, " retOrch=", ret)
					if strings.Contains(ret, scp_die) {
						fmt.Println("ERROR SETUP DEVICES: BIORREATOR DIE", b.BioreactorID)
						break
					}
					if nerr > 3 {
						fmt.Println("ERROR SETUP DEVICES: BIORREATOR com EXCESSO de ERROS", b.BioreactorID)
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
				if biofabrica.Critical != scp_netfail {
					if nerr > 1 && !devmode && bio[ind].Status != bio_error {
						if bio[ind].Status == bio_pause {
							if bio[ind].LastStatus != bio_error {
								bio[ind].UndoStatus = bio[ind].LastStatus
							}
						}
						bio[ind].LastStatus = bio[ind].Status
						bio[ind].Status = bio_error
						fmt.Println("ERROR SETUP DEVICES: BIORREATOR com erros", b.BioreactorID)
					} else if nerr == 0 && (bio[ind].Status == bio_nonexist || bio[ind].Status == bio_error) {
						if bio[ind].Volume == 0 && len(bio[ind].Queue) == 0 {
							bio[ind].Status = bio_empty
						} else {
							bio[ind].Status = bio[ind].LastStatus
							if bio[ind].Status == bio_pause {
								if len(bio[ind].UndoStatus) > 0 {
									if bio[ind].UndoStatus != bio_error {
										bio[ind].LastStatus = bio[ind].UndoStatus
									}
								} else {
									bio[ind].LastStatus = bio_ready
								}
							}
						}
					}
				}

			}
		}
	}

	fmt.Println("\n\nDEBUG SETUP DEVICES: Configurando IBCs")
	for _, ib := range ibc_cfg {
		// if biofabrica.Critical == scp_netfail {
		// 	break
		// }
		ind := get_ibc_index(ib.IBCID)
		ibcexist := true
		if ind >= 0 {
			if ibc_cfg[ib.IBCID].Deviceaddr == bio_noaddr {
				ibcexist = false
				ibc[ind].Status = bio_nonexist
			}
		}
		if ibcexist && len(ib.Deviceaddr) > 0 && ind >= 0 {
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
				cmd = append(cmd, "CMD/"+ibcaddr+"/MOD/"+scp_magicport+",3/END")
				cmd = append(cmd, "CMD/"+ibcaddr+"/PUT/"+scp_magicport+","+scp_magicvalue+"/END")
				cmd = append(cmd, "CMD/"+ibcaddr+"/MOD/"+ib.Levellow[1:]+",0/END")
				// cmd = append(cmd, "CMD/"+ibcaddr+"/MOD/"+b.Emergency[1:]+",1/END")

				nerr := 0
				for k, c := range cmd {
					ret := scp_sendmsg_orch(c)
					if !strings.Contains(ret, scp_ack) {
						nerr++
					}
					fmt.Println(ib.IBCID, "err=", nerr, k, c, " retOrch=", ret)
					if strings.Contains(ret, scp_die) {
						fmt.Println("ERROR SETUP DEVICES: IBC DIE", ib.IBCID)
						break
					}
					if nerr > 3 {
						fmt.Println("ERROR SETUP DEVICES: IBC com EXCESSO de ERROS", ib.IBCID)
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
				if biofabrica.Critical != scp_netfail {
					if nerr > 1 && !devmode && ibc[ind].Status != bio_error {
						if ibc[ind].Status == bio_pause {
							if ibc[ind].LastStatus != bio_error {
								ibc[ind].UndoStatus = ibc[ind].LastStatus
							}
						}
						ibc[ind].LastStatus = ibc[ind].Status
						ibc[ind].Status = bio_error
						fmt.Println("ERROR SETUP DEVICES: IBC com erros", ib.IBCID)
					} else if nerr == 0 && (ibc[ind].Status == bio_nonexist || ibc[ind].Status == bio_error) {
						if ibc[ind].Volume == 0 && len(ibc[ind].Queue) == 0 {
							ibc[ind].Status = bio_empty
						} else {
							ibc[ind].Status = ibc[ind].LastStatus
							if len(ibc[ind].UndoStatus) > 0 {
								if ibc[ind].UndoStatus != bio_error {
									ibc[ind].LastStatus = ibc[ind].UndoStatus
								}
							} else {
								ibc[ind].LastStatus = bio_ready
							}
						}
					}
				}

			}
		}
	}

	fmt.Println("\n\nDEBUG SETUP DEVICES: Configurando TOTEMs")
	for _, tot := range totem_cfg {
		// if biofabrica.Critical == scp_netfail {
		// 	break
		// }
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
				cmd = append(cmd, "CMD/"+totemaddr+"/MOD/"+scp_magicport+",3/END")
				cmd = append(cmd, "CMD/"+totemaddr+"/PUT/"+scp_magicport+","+scp_magicvalue+"/END")
				nerr := 0
				for k, c := range cmd {
					ret := scp_sendmsg_orch(c)
					if !strings.Contains(ret, scp_ack) {
						nerr++
					}
					fmt.Println(tot.TotemID, "err=", nerr, k, c, " retOrch=", ret)
					if strings.Contains(ret, scp_die) {
						fmt.Println("ERROR SETUP DEVICES: TOTEM DIE", tot.TotemID)
						break
					}
					time.Sleep(scp_refreshwait / 2 * time.Millisecond)
				}
				if biofabrica.Critical != scp_netfail {
					if nerr > 0 && !devmode {
						totem[ind].Status = bio_error
						fmt.Println("ERROR SETUP DEVICES: TOTEM com erros", tot.TotemID, "#", nerr)
					} else if nerr == 0 {
						totem[ind].Status = bio_ready
					}
				}
			}
		}
	}

	fmt.Println("\n\nDEBUG SETUP DEVICES: Configurando BIOFABRICA")

	if mustall || biofabrica.Status == scp_fail || biofabrica.PIntStatus == bio_error || biofabrica.POutStatus == bio_error {
		biofabrica.POutStatus = bio_update
		biofabrica.PIntStatus = bio_update
		for _, bf := range biofabrica_cfg {
			if len(bf.Deviceaddr) > 0 {
				fmt.Println("DEBUG SETUP DEVICES: Device:", bf.DeviceID, "-", bf.Deviceaddr)
				var cmd []string
				bfaddr := bf.Deviceaddr
				cmd = make([]string, 0)
				if bf.Deviceport[0] != 'C' {
					cmd = append(cmd, "CMD/"+bfaddr+"/MOD/"+bf.Deviceport[1:]+",3/END")
					cmd = append(cmd, "CMD/"+bfaddr+"/MOD/"+scp_magicport+",3/END")
					cmd = append(cmd, "CMD/"+bfaddr+"/PUT/"+scp_magicport+","+scp_magicvalue+"/END")
				}

				nerr := 0
				err_local := ""
				for k, c := range cmd {
					fmt.Print()
					ret := scp_sendmsg_orch(c)
					fmt.Println("DEBUG SETUP DEVICES: ", bf.DeviceID, k, "  ", c, " ", ret)
					if !strings.Contains(ret, scp_ack) {
						nerr++
						params := scp_splitparam(c, "/")
						if len(params) > 2 {
							disp := get_devid_byaddr(params[1])
							if !devmode {
								if strings.Contains(disp, "VBF03") || strings.Contains(disp, "VBF04") || strings.Contains(disp, "VBF05") {
									err_local += "Painel Intermediário "
									biofabrica.PIntStatus = bio_error
								} else if strings.Contains(disp, "VBF06") || strings.Contains(disp, "VBF07") || strings.Contains(disp, "VBF08") || strings.Contains(disp, "VBF09") || strings.Contains(disp, "PBF01") || strings.Contains(disp, "FBF01") {
									err_local += "Painel Desenvase "
									biofabrica.POutStatus = bio_error
								} else if len(disp) > 0 {
									err_local += "Dispositivo " + disp
								}
							}
						}
					}
					if ret[0:2] == "DIE" {
						fmt.Println("SLAVE ERROR - DIE")
						nerr++
						break
					}
					time.Sleep(scp_refreshwait / 2 * time.Millisecond)
				}
				if biofabrica.POutStatus == bio_update {
					biofabrica.POutStatus = bio_ready
				}
				if biofabrica.PIntStatus == bio_update {
					biofabrica.PIntStatus = bio_ready
				}
				if nerr > 0 && !devmode && biofabrica.Status != scp_fail {
					biofabrica.Status = scp_fail
					fmt.Println("CRITICAL SETUP DEVICES: BIOFABRICA com erros", err_local)
					if len(err_local) > 0 {
						board_add_message("EFALHA CRITICA em "+err_local, "BFFAIL")
					}
				} else if nerr == 0 {
					biofabrica.Status = scp_ready
					biofabrica.POutStatus = bio_ready
					biofabrica.PIntStatus = bio_ready
					board_del_message("BFFAIL")
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
	if bio[ind].Status == bio_producting {
		aerostatus = true
	}
	aeroratio := bio[ind].AeroRatio

	if len(bioaddr) > 0 {
		if aerostatus {
			if bio[ind].Status != bio_producting {
				return -1
			}
			if !scp_turn_aero(bioid, true, 0, 0, false) {
				fmt.Println("ERROR SCP GET PH VOLTAGE: Erro ao desligar Aerador do Biorreator", bioid)
				scp_turn_aero(bioid, false, 1, aeroratio, false)
				return -1
			}
			time.Sleep(scp_timewaitbeforeph * time.Millisecond)
		}
		cmd_ph := "CMD/" + bioaddr + "/GET/" + phdev + "/END"

		n := 0
		var data []float64
		for i := 0; i <= 7; i++ {
			ret_ph := scp_sendmsg_orch(cmd_ph)
			params := scp_splitparam(ret_ph, "/")
			if len(params) > 1 {
				phint, err := strconv.Atoi(params[1])
				fmt.Println("DEBUG GET PH VOLTAGE: Valor retornado =", bioid, phint, "passo=", i)
				if err == nil && phint >= 300 && phint <= 500 {
					data = append(data, float64(phint))
					n++
				}
			} else {
				fmt.Println("DEBUG GET PH VOLTAGE: Valor retornado invalido =", bioid, ret_ph, "passo=", i)
			}
		}
		mediana := calc_mediana(data)
		phfloat := mediana / 100.0

		fmt.Println("DEBUG SCP GET PH VOLTAGE: Lendo Voltagem PH do Biorreator", bioid, cmd_ph, "- mediana =", mediana, " phfloat=", phfloat)
		if aerostatus {
			if !scp_turn_aero(bioid, true, 1, aeroratio, false) {
				fmt.Println("ERROR SCP GET PH VOLTAGE: Erro ao religar Aerador do Biorreator", bioid)
			}
		}
		if phfloat >= 3.0 && phfloat <= 5.0 {
			return phfloat
		} else {
			fmt.Println("ERROR SCP GET PH VOLTAGE: Valor invalido de PH Voltage", bioid, phfloat)
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

func scp_update_screen_vol(bioid string) {
	ind := get_bio_index(bioid)
	if ind < 0 {
		return
	}
	bioscr := bio_cfg[bioid].Screenaddr
	if bio[ind].ShowVol {
		vol_str := fmt.Sprintf("%d", bio[ind].Volume)
		cmd := "CMD/" + bioscr + "/PUT/S232," + vol_str + "/END"
		ret := scp_sendmsg_orch(cmd)
		fmt.Println("DEBUG SCP UPDATE SCREEN: cmd=", cmd, "ret=", ret)
		if !strings.Contains(ret, "ACK") {
			return
		}
	}
}

func scp_update_screen_phtemp(bioid string) {
	ind := get_bio_index(bioid)
	if ind < 0 {
		return
	}
	var cmd, ret string
	bioscr := bio_cfg[bioid].Screenaddr
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
}

func scp_update_screen_steps(bioid string) {
	ind := get_bio_index(bioid)
	if ind < 0 {
		return
	}
	bioscr := bio_cfg[bioid].Screenaddr

	step_str := fmt.Sprintf("%d", int(bio[ind].Step[1]))
	cmd := "CMD/" + bioscr + "/PUT/S282," + step_str + "/END"
	ret := scp_sendmsg_orch(cmd)
	fmt.Println("DEBUG SCP UPDATE SCREEN STEPS: cmd=", cmd, " ret=", ret)
	if !strings.Contains(ret, "ACK") {
		return
	}

	step_str = fmt.Sprintf("%d", int(bio[ind].Step[0]))
	cmd = "CMD/" + bioscr + "/PUT/S281," + step_str + "/END"
	ret = scp_sendmsg_orch(cmd)

	fmt.Println("DEBUG SCP UPDATE SCREEN STEPS: cmd=", cmd, " ret=", ret)
	if !strings.Contains(ret, "ACK") {
		return
	}

}

func scp_update_screen_times(bioid string) {
	ind := get_bio_index(bioid)
	if ind < 0 {
		return
	}
	bioscr := bio_cfg[bioid].Screenaddr

	var cmd, ret string

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

func scp_update_screen(bioid string) {
	return
	// ind := get_bio_index(bioid)
	// if ind < 0 {
	// 	return
	// }
	// bioscr := bio_cfg[bioid].Screenaddr

	// var cmd, ret string

	// status_code, ok := status_codes[bio[ind].Status]
	// if ok {
	// 	cmd = fmt.Sprintf("CMD/%s/PUT/S247,%d/END", bioscr, status_code)
	// 	ret = scp_sendmsg_orch(cmd)
	// 	if !strings.Contains(ret, "ACK") {
	// 		return
	// 	}
	// }
	// org_index := "0"
	// if bio[ind].Status != bio_cip && bio[ind].Status != bio_error && bio[ind].Status != bio_water && bio[ind].Status != bio_empty {
	// 	org_index = organs[bio[ind].OrgCode].Index
	// }
	// if len(org_index) > 0 {
	// 	cmd = "CMD/" + bioscr + "/PUT/S245," + org_index + "/END"
	// 	ret = scp_sendmsg_orch(cmd)
	// }
}

func scp_get_temperature(bioid string) float64 {
	bioaddr := bio_cfg[bioid].Deviceaddr
	tempdev := bio_cfg[bioid].Temp_dev
	if len(bioaddr) > 0 {
		cmd_temp := "CMD/" + bioaddr + "/GET/" + tempdev + "/END"

		n := 0
		var data []float64
		for i := 0; i <= 4; i++ {
			ret_temp := scp_sendmsg_orch(cmd_temp)
			params := scp_splitparam(ret_temp, "/")
			if len(params) > 1 {
				tempint, err := strconv.Atoi(params[1])
				fmt.Println("DEBUG GET TEMPERATURE: Valor retornado =", bioid, tempint, "passo=", i)
				if err == nil && tempint > 0 && tempint <= 1200 {
					data = append(data, float64(tempint))
					n++
				}
			} else {
				fmt.Println("ERROR GET TEMPERATURE: Valor retornado =", bioid, ret_temp, "passo=", i)
			}
		}
		mediana := calc_mediana(data)
		tempfloat := mediana / 10.0

		fmt.Println("DEBUG SCP GET TEMPERATURE: Lendo Temperatura do Biorreator", bioid, "cmd=", cmd_temp, "mediana=", mediana, "tempfloat=", tempfloat)
		if tempfloat > 0 && tempfloat < 120 {
			return tempfloat
		} else {
			fmt.Println("ERROR SCP GET TEMPERATURE: Retorno invalido, temperatura fora do range", bioid, tempfloat)
			return -1
		}
	} else {
		fmt.Println("ERROR SCP GET TEMPERATURE: ADDR Biorreator nao existe", bioid)
	}
	return -1
}

func scp_get_emerg(main_id string, dev_type string) int {
	var devaddr, emergdev string
	ind := -1
	switch dev_type {
	case scp_bioreactor:
		ind = get_bio_index(main_id)
		if ind < 0 {
			return -1
		}
		devaddr = bio_cfg[main_id].Deviceaddr
		emergdev = bio_cfg[main_id].Emergency

	case scp_ibc:
		return -1
		// ind = get_ibc_index(main_id)
		// if ind < 0 {
		// 	return -1
		// }
		// devaddr = ibc_cfg[main_id].Deviceaddr
		//emergdev = ibc_cfg[main_id].Emergency
	}
	cmd := "CMD/" + devaddr + "/GET/" + emergdev + "/END"
	ret := scp_sendmsg_orch(cmd)
	params := scp_splitparam(ret, "/")
	if len(params) > 1 && params[0] == scp_ack {
		emergint, err := strconv.Atoi(params[1])
		if err == nil {
			fmt.Println("DEBUG SCP GET EMERG: ", main_id, " Retorno do botao de emergencia =", emergint)
			return emergint
		} else {
			fmt.Println("ERROR SCP GET EMERG: Retorno invalido", main_id, ret, params)
		}
	} else {
		fmt.Println("ERROR SCP GET EMERG: Retorno faltando parametros", main_id, ret, params)
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
			_, ok := biofabrica_cfg["FBF01"]
			if ok {
				dev_addr = biofabrica_cfg["FBF01"].Deviceaddr
				vol_dev = biofabrica_cfg["FBF01"].Deviceport
			} else {
				fmt.Println("ERROR SCP GET VOLUME: Nao foi encontrados os dados do Fluxometro FBF01")
				board_add_message("EFBF01 Não configurado. Favor corrigir em Configurações / Biofábrica", "")
				return -1, -1
			}
		} else if vol_type == scp_dev_volfluxo_in1 {
			_, ok := biofabrica_cfg["FBF02"]
			if ok {
				dev_addr = biofabrica_cfg["FBF02"].Deviceaddr
				vol_dev = biofabrica_cfg["FBF02"].Deviceport
			} else {
				fmt.Println("ERROR SCP GET VOLUME: Nao foi encontrados os dados do Fluxometro FBF02")
				board_add_message("EFBF02 Não configurado. Favor corrigir em Configurações / Biofábrica", "")
				return -1, -1
			}
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
	var volume float64
	var dint int64
	var err error
	for i := 0; i < 3; i++ {
		fmt.Println("DEBUG SCP GET VOLUME: step", i, "main_id=", main_id, dev_type, vol_type, " == CMD=", cmd, "  RET=", ret)
		if params[0] == scp_ack {
			dint, err = strconv.ParseInt(params[1], 10, 32)
			if err == nil {
				if dint > 1 || vol_type == scp_dev_vol0 {
					break
				}
			}
		}
		ret = scp_sendmsg_orch(cmd)
		params = scp_splitparam(ret, "/")
	}
	// params := scp_splitparam(ret, "/")
	fmt.Println("DEBUG SCP GET VOLUME: ", main_id, dev_type, vol_type, " == CMD=", cmd, "  RET=", ret)
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
			biofabrica.VolIn1Part += float64(math.MaxUint16) * flow_ratio_in1
		}
		volume = (float64(dint) * flow_ratio_in1) + biofabrica.VolIn1Part
		biofabrica.LastCountIn1 = uint32(dint)
	}

	if volume < 0 {
		fmt.Println("ERROR SCP GET VOLUME: VOLUME NEGATIVO encontrado", main_id, vol_type, dint, volume)
		volume = 0
	}
	return int(dint), volume
}

func scp_test_boot(main_id string, dev_type string) string {
	var devaddr string
	switch dev_type {
	case scp_bioreactor:
		devaddr = bio_cfg[main_id].Deviceaddr
	case scp_ibc:
		devaddr = ibc_cfg[main_id].Deviceaddr
	case scp_totem:
		devaddr = totem_cfg[main_id].Deviceaddr
	case scp_wdpanel:
		devaddr = biofabrica_cfg["PBF01"].Deviceaddr
	case scp_intpanel:
		devaddr = biofabrica_cfg["VBF03"].Deviceaddr
	default:
		devaddr = ""
	}
	if devaddr == "" {
		fmt.Println("ERROR SCP TEST BOOT: Dispositivo invalido", main_id, dev_type)
		return scp_err
	}
	cmd := "CMD/" + devaddr + "/GET/" + scp_magicport + "/END"
	ret := scp_sendmsg_orch(cmd)
	params := scp_splitparam(ret, "/")
	ok := false
	for i := 0; i < 3; i++ {
		fmt.Println("DEBUG SCP TEST BOOT: Testando magicvalue no Dispositivo ", main_id, dev_type, "#", i, "cmd=", cmd, "ret=", ret)
		if params[0] == scp_ack && len(params) > 1 {
			if params[1] == scp_magicvalue || params[1] == "0" {
				ok = true
				break
			}
		}
		ret := scp_sendmsg_orch(cmd)
		params = scp_splitparam(ret, "/")
	}
	if ok {
		if params[1] != scp_magicvalue {
			fmt.Println("DEBUG SCP TEST BOOT: Dispositivo retornou valor 0 para magicvalue, indicio de reboot ", main_id, dev_type, params)
			return scp_reboot
		}
	} else {
		fmt.Println("ERROR SCP TEST BOOT: Dispositivo nao retornou valor valido para magicvalue ", main_id, dev_type, params)
		return scp_err
	}
	return scp_ack
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
						dev_id := get_devid_byaddr(dev_addr)
						dev_type := get_devtype_byaddr(dev_addr)
						fmt.Println("DEBUG SCP REFRESH STATUS: SLAVE DATA:", dev_id, dev_addr, ipaddr, devstatus, dev_type)
						if devstatus == scp_state_TCP0 { // Dispositivo OK
							nslavesok++
						} else { // Dispositivo NAO OK
							nslavesnok++
							switch dev_type {
							case scp_bioreactor:
								ind := get_bio_index(dev_id)
								if ind >= 0 {
									fmt.Println("DEBUG SCP REFRESH STATUS: FALHA no Biorreator", dev_id)
									if bio[ind].Status != bio_error && bio[ind].Status != bio_pause {
										if bio[ind].Status != bio_ready || bio[ind].Status != bio_empty {
											board_add_message("E"+bio[ind].BioreactorID+" com falha, favor verificar", "")
											bio_add_message(bio[ind].BioreactorID, "EEquipamento com falha, favor verificar", "")
											// go pause_device(scp_bioreactor, bio[ind].BioreactorID, true) // TESTANDO AUTOPAUSE
										}
									}
									if bio[ind].Status != bio_error {
										if bio[ind].Status == bio_pause {
											bio[ind].UndoStatus = bio[ind].LastStatus
										}
										bio[ind].LastStatus = bio[ind].Status
										bio[ind].Status = bio_error
									}
								} else {
									fmt.Println("ERROR SCP REFRESH STATUS: Biorreator não existe na tabela", dev_id)
								}

							case scp_ibc:
								ind := get_ibc_index(dev_id)
								if ind >= 0 {
									fmt.Println("DEBUG SCP REFRESH STATUS: FALHA no IBC", dev_id)
									if ibc[ind].Status != bio_error && ibc[ind].Status != bio_pause {
										if ibc[ind].Status != bio_ready || ibc[ind].Status != bio_empty {
											board_add_message("E"+ibc[ind].IBCID+" com falha, favor verificar", "")
											// pause_device(scp_ibc, ibc[ind].IBCID, true)
										}
									}
									if ibc[ind].Status != bio_error {
										ibc[ind].LastStatus = ibc[ind].Status
										ibc[ind].Status = bio_error
									}
								} else {
									fmt.Println("ERROR SCP REFRESH STATUS: IBC não existe na tabela", dev_id)
								}

							case scp_totem:
								ind := get_totem_index(dev_id)
								if ind >= 0 {
									fmt.Println("DEBUG SCP REFRESH STATUS: FALHA no TOTEM", dev_id)
									totem[ind].Status = bio_error
									if dev_id == "TOTEM01" {
										fmt.Println("DEBUG SCP REFRESH STATUS: FALHA no TOTEM01 e mudando Biofabrica para falha tambem")
										biofabrica.Status = scp_fail // novo
									}
								} else {
									fmt.Println("ERROR SCP REFRESH STATUS: TOMEM não existe na tabela", dev_id)
								}

							case scp_screen:
								fmt.Println("ERROR SCP REFRESH STATUS: FALHA na TELA do Biorreator", dev_id)

							case scp_biofabrica:
								fmt.Println("DEBUG SCP REFRESH STATUS: FALHA na Biofabrica", dev_id)
								biofabrica.Status = scp_fail
								if strings.Contains(dev_id, "VBF03") || strings.Contains(dev_id, "VBF04") || strings.Contains(dev_id, "VBF05") {
									biofabrica.PIntStatus = bio_error
								} else if strings.Contains(dev_id, "VBF06") || strings.Contains(dev_id, "VBF07") || strings.Contains(dev_id, "VBF08") || strings.Contains(dev_id, "VBF09") || strings.Contains(dev_id, "FBF01") || strings.Contains(dev_id, "PBF01") {
									biofabrica.POutStatus = bio_error
								} else if strings.Contains(dev_id, "VBF01") || strings.Contains(dev_id, "VBF02") {
									fmt.Println("DEBUG SCP REFRESH STATUS: FALHA no TOTEM01 por falha na válvula", dev_id)
									ind := get_totem_index("TOTEM01")
									totem[ind].Status = bio_error
								}

							default:
								fmt.Println("ERROR SCP REFRESH STATUS: TIPO INVALIDO na tabela", dev_addr, dev_type)
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
	runningsync = true
	t_start_save := time.Now()
	t_start_status := time.Now()
	t_start_screens := time.Now()
	t_start_screens_full := time.Now()
	t_start_version := time.Now()
	t_start_setup := time.Now()

	n_bio := 0
	for {
		if finishedsetup {

			if biofabrica.Critical == scp_stopall {
				runningsync = false
				return
			}

			t_elapsed_save := uint32(time.Since(t_start_save).Seconds())
			if t_elapsed_save >= scp_timetosave {
				save_all_data(data_filename)
				t_start_save = time.Now()
			}

			t_elapsed_setup := uint32(time.Since(t_start_setup).Hours())
			if t_elapsed_setup >= scp_timetosetup {
				scp_setup_devices(true)
				t_start_setup = time.Now()
			}

			t_elapsed_status := uint32(time.Since(t_start_status).Seconds())
			if t_elapsed_status >= scp_refresstatus {
				if finishedsetup && biofabrica.Critical != scp_netfail {
					scp_refresh_status()
				}
				t_start_status = time.Now()
			}

			t_elapsed_screens := uint32(time.Since(t_start_screens).Seconds())
			t_elapsed_screens_full := uint32(time.Since(t_start_screens_full).Seconds())
			if t_elapsed_screens >= scp_refresscreens {
				if t_elapsed_screens_full > scp_refresscreens*12 {
					scp_update_screen(bio[n_bio].BioreactorID)
					t_start_screens_full = time.Now()
				} else {
					scp_update_screen(bio[n_bio].BioreactorID)
				}
				n_bio++
				if n_bio >= len(bio) {
					n_bio = 0
				}
				t_start_screens = time.Now()
			}
			if !schedrunning {
				go scp_scheduler()
			}

		}
		t_elapsed_version := uint32(time.Since(t_start_version).Minutes())
		if t_elapsed_version > scp_timetocheckversion {
			go scp_check_lastversion()
			t_start_version = time.Now()
		}
		time.Sleep(scp_refreshsync * time.Second)
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
		// levels := fmt.Sprintf("%d", level_int)
		// cmd := "CMD/" + bio_cfg[bioid].Screenaddr + "/PUT/S231," + levels + "/END"
		// ret := scp_sendmsg_orch(cmd)
		// fmt.Println("SCREEN:", cmd, level, levels, ret)
	}
}

func scp_update_ibclevel(ibcid string) {
	ind := get_ibc_index(ibcid)
	if ind < 0 {
		return
	}
	level := (float64(ibc[ind].Volume) / float64(ibc_cfg[ibcid].Maxvolume)) * 10.0
	if level > 10 {
		level = 10
	}
	level_int := uint8(level)
	if level_int != ibc[ind].Level {
		ibc[ind].Level = level_int
		// levels := fmt.Sprintf("%d", level_int)
		// cmd := "CMD/" + bio_cfg[bioid].Screenaddr + "/PUT/S231," + levels + "/END"
		// ret := scp_sendmsg_orch(cmd)
		// fmt.Println("SCREEN:", cmd, level, levels, ret)
	}
}

func scp_get_alldata() {
	if demo {
		return
	}
	runninggetall = true
	t_start_bio := time.Now()
	t_start_ibc := time.Now()
	t_start_test_pipd := time.Now()
	t_start_test_totems := time.Now()
	t_start_setup := time.Now()
	lastvolin := float64(-1)
	lastvolout := float64(-1)
	hasupdatevolin := false
	hasupdatevolout := false
	firsttime := true
	bio_seq := 0
	ibc_seq := 0
	for {
		if biofabrica.Critical == scp_stopall {
			runninggetall = false
			return
		}
		if finishedsetup {

			t_elapsed_test_pipd := uint32(time.Since(t_start_test_pipd).Seconds())
			if t_elapsed_test_pipd > 25 {
				pi_addr := biofabrica_cfg["VBF03"].Deviceaddr
				pi_port := biofabrica_cfg["VBF03"].Deviceport
				if len(pi_addr) == 0 || len(pi_port) == 0 {
					fmt.Println("ERROR GET ALLDATA: Erro ao encontrar configuracao de VBF03 para teste do Painel Intermediario")
				} else {
					// cmd_pi := "CMD/" + pi_addr + "/GET/" + pi_port + "/END"
					// ret_pi := scp_sendmsg_orch(cmd_pi)
					// fmt.Println("DEBUG GET ALLDATA: Teste do Painel Intermediario  cmd=", cmd_pi, " ret=", ret_pi)
					testboot := scp_test_boot(scp_biofabrica, scp_intpanel)
					if testboot == scp_reboot {
						fmt.Println("ERROR GET ALLDATA: Teste do Painel Intermediario retornou que houve BOOT. Mudando status para ERROR")
						biofabrica.PIntStatus = bio_error
					} else if testboot == scp_err {
						fmt.Println("ERROR GET ALLDATA: Teste do Painel Intermediario retornou ERRO no test de boot", testboot)
					} else {
						fmt.Println("DEBUG GET ALLDATA: Teste do Painel Intermediario retornou ", testboot)
					}
				}

				pd_addr := biofabrica_cfg["VBF07"].Deviceaddr
				pd_port := biofabrica_cfg["VBF07"].Deviceport
				if len(pd_addr) == 0 || len(pd_port) == 0 {
					fmt.Println("ERROR GET ALL DATA: Erro ao encontrar configuracao de VBF07 para teste do Painel Desenvase")
				} else {
					// cmd_pd := "CMD/" + pd_addr + "/GET/" + pd_port + "/END"
					// ret_pd := scp_sendmsg_orch(cmd_pd)
					// fmt.Println("DEBUG GET ALL DATA: Teste do Painel Desenvase  cmd=", cmd_pd, " ret=", ret_pd)
					testboot := scp_test_boot(scp_biofabrica, scp_wdpanel)
					if testboot == scp_reboot {
						fmt.Println("ERROR GET ALLDATA: Teste do Painel de Desenvase retornou que houve BOOT. Mudando status para ERROR")
						biofabrica.POutStatus = bio_error
					} else if testboot == scp_err {
						fmt.Println("ERROR GET ALLDATA: Teste do Painel de Desenvase retornou ERRO no test de boot", testboot)
					} else {
						fmt.Println("DEBUG GET ALLDATA: Teste do Painel de Desenvase retornou ", testboot)
					}
				}
				t_start_test_pipd = time.Now()
			}

			t_elapsed_test_totems := uint32(time.Since(t_start_test_totems).Seconds())
			if t_elapsed_test_totems > 21 && rand.Intn(3) == 1 {
				t1_addr := biofabrica_cfg["VBF01"].Deviceaddr
				t1_port := biofabrica_cfg["VBF01"].Deviceport
				if len(t1_addr) == 0 || len(t1_port) == 0 {
					fmt.Println("ERROR GET ALL DATA: Erro ao encontrar configuracao de VBF01 para teste do TOTEM01")
				} else {
					// cmd_t1 := "CMD/" + t1_addr + "/GET/" + t1_port + "/END"
					// ret_t1 := scp_sendmsg_orch(cmd_t1)
					// fmt.Println("DEBUG GET ALL DATA: Teste da Válvulva VBF01 no TOTEM01 cmd=", cmd_t1, " ret=", ret_t1)
					testboot := scp_test_boot("TOTEM01", scp_totem)
					if testboot == scp_reboot {
						fmt.Println("ERROR GET ALLDATA: Teste do TOTEM01 retornou que houve BOOT. Mudando status para ERROR")
						ind := get_totem_index("TOTEM01")
						if ind >= 0 {
							totem[ind].Status = bio_error
							biofabrica.Status = scp_fail
						}
					} else if testboot == scp_err {
						fmt.Println("ERROR GET ALLDATA: Teste do TOTEM01 retornou ERRO no test de boot", testboot)
					} else {
						fmt.Println("DEBUG GET ALLDATA: Teste do TOTEM01 retornou ", testboot)
					}
				}

				t2_addr := totem_cfg["TOTEM02"].Deviceaddr
				t2_port := totem_cfg["TOTEM02"].Valv_devs[0]
				if len(t2_addr) == 0 || len(t2_port) == 0 {
					fmt.Println("ERROR GET ALL DATA: Erro ao encontrar configuracao de TOTEM02 para teste")
				} else {
					// cmd_t2 := "CMD/" + t2_addr + "/GET/" + t2_port + "/END"
					// ret_t2 := scp_sendmsg_orch(cmd_t2)
					// fmt.Println("DEBUG GET ALL DATA: Teste do TOTEM02 cmd=", cmd_t2, " ret=", ret_t2)
					testboot := scp_test_boot("TOTEM02", scp_totem)
					if testboot == scp_reboot {
						fmt.Println("ERROR GET ALLDATA: Teste do TOTEM02 retornou que houve BOOT. Mudando status para ERROR")
						ind := get_totem_index("TOTEM02")
						if ind >= 0 {
							totem[ind].Status = bio_error
						}
					} else if testboot == scp_err {
						fmt.Println("ERROR GET ALLDATA: Teste do TOTEM02 retornou ERRO no test de boot", testboot)
					} else {
						fmt.Println("DEBUG GET ALLDATA: Teste do TOTEM02 retornou ", testboot)
					}
				}

				t_start_test_totems = time.Now()
			}

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

					if mustupdate_this || rand.Intn(21) == 7 {
						testboot := scp_test_boot(b.BioreactorID, scp_bioreactor)
						if testboot == scp_reboot {
							fmt.Println("ERROR GET ALLDATA: Teste do Biorreator", b.BioreactorID, "retornou que houve BOOT. Mudando status para ERROR")
							bio[ind].Status = bio_error
						} else if testboot == scp_err {
							fmt.Println("ERROR GET ALLDATA: Teste do Biorreator", b.BioreactorID, "retornou ERRO no test de boot", testboot)
						} else {
							fmt.Println("DEBUG GET ALLDATA: Teste do Biorreator", b.BioreactorID, "retornou ", testboot)
						}
					}

					if mustupdate_this || b.Status == bio_producting || b.Status == bio_cip || b.Valvs[2] == 1 {

						if rand.Intn(21) == 7 {
							t_tmp := scp_get_temperature(b.BioreactorID)
							if (t_tmp >= 0) && (t_tmp <= TEMPMAX) {
								bio[ind].Temperature = float32(t_tmp)
								if bio[ind].Heater && float32(t_tmp) >= bio[ind].TempMax {
									fmt.Println("DEBUG SCP GET ALLDATA: Desligando resistencia por atingir temperatura maxima definida", b.BioreactorID, "tempnow=", t_tmp, "max=", bio[ind].TempMax)
									scp_turn_heater(b.BioreactorID, 0, false)
								} else if bio[ind].Heater && bio[ind].Pumpstatus && bio[ind].Volume > 0 && bio[ind].TempMax > 0 && float32(t_tmp) <= bio[ind].TempMax-5 {
									fmt.Println("DEBUG SCP GET ALLDATA: Ligando resistencia por temperatura ser inferior ao maximo - 5", b.BioreactorID, "tempnow=", t_tmp, "max-5=", bio[ind].TempMax-5)
									scp_turn_heater(b.BioreactorID, bio[ind].TempMax, true)
								}
							} else if t_tmp > TEMPMAX {
								fmt.Println("ERROR SCP GET ALLDATA: TEMPERATURA CRÍTICA - Desligando resistencia e alertando", b.BioreactorID, "tempnow=", t_tmp, "max=", bio[ind].TempMax, " TEMPMAX=", TEMPMAX)
								bio_add_message(b.BioreactorID, "EATENÇÃO: Temperatura do Biorreator Crítica!!! Verificar", "")
								scp_turn_heater(b.BioreactorID, bio[ind].TempMax, true)
							}
						}
					}

					if (mustupdate_this || b.Valvs[1] == 1) && (b.Status == bio_ready || b.Status == bio_empty) {
						if rand.Intn(15) == 7 {
							go scp_update_ph(b.BioreactorID)
						}
					}

					// if rand.Intn(7) == 3 && (b.Status == bio_producting || b.Status == bio_cip || b.Status == bio_circulate || b.Status == bio_loading || b.Status == bio_unloading) {
					// 	fmt.Println("DEBUG SCP GET ALL DATA: Verificando botao de emergencia do", b.BioreactorID)
					// 	emerg := scp_get_emerg(b.BioreactorID, scp_bioreactor)
					// 	if b.Emergpress {
					// 		if emerg == 0 {
					// 			emerg2 := scp_get_emerg(b.BioreactorID, scp_bioreactor)
					// 			if emerg2 == 0 {
					// 				bio_add_message(b.BioreactorID, "ABotão de emergência liberado")
					// 				b.Emergpress = false
					// 				pause_device(scp_bioreactor, b.BioreactorID, false)
					// 			}
					// 		}
					// 	} else {
					// 		if emerg == 1 && !b.MustPause {
					// 			emerg2 := scp_get_emerg(b.BioreactorID, scp_bioreactor)
					// 			if emerg2 == 1 {
					// 				bio_add_message(b.BioreactorID, "ABotão de emergência pressionado")
					// 				b.Emergpress = true
					// 				pause_device(scp_bioreactor, b.BioreactorID, true)
					// 			}
					// 		}
					// 	}
					// }

					if mustupdate_this || b.Valvs[6] == 1 || b.Valvs[4] == 1 {

						if rand.Intn(3) == 1 {
							scp_get_volume(b.BioreactorID, scp_bioreactor, scp_dev_vol0) // força tentar ler algo pra dar erro caso esteja off
						}

						if biofabrica.Useflowin {
							if b.Valvs[6] == 1 && (b.Valvs[3] == 1 || (b.Valvs[2] == 1 && b.Valvs[7] == 1) || (b.Valvs[1] == 1 && b.Valvs[7] == 1)) {
								count, vol_tmp := scp_get_volume(scp_biofabrica, scp_biofabrica, scp_dev_volfluxo_in1)
								if count >= 0 {
									vol_tmp = vol_tmp // * flow_corfactor_in1
									fmt.Println("DEBUG SCP GET ALL DATA: Biorreator", b.BioreactorID, " usando volume vindo do Totem01 =", vol_tmp, lastvolin)
									biofabrica.VolumeIn1 = bio_escala * (math.Trunc(vol_tmp / bio_escala))
									if vol_tmp > lastvolin && lastvolin > 0 {
										biovolin := vol_tmp - lastvolin
										bio[ind].VolInOut += biovolin
										bio[ind].Volume = uint32(bio[ind].VolInOut)
										fmt.Println("DEBUG SCP GET ALL DATA: Biorreator", b.BioreactorID, " NOVO volume =", bio[ind].VolInOut)
										scp_update_biolevel(b.BioreactorID)
									}
									lastvolin = vol_tmp
									hasupdatevolin = true
								} else {
									fmt.Println("ERROR SCP GET ALL DATA: Valor invalido ao ler Volume INFLUXO", count, vol_tmp)
								}
							} else if b.Valvs[4] == 1 && b.Valvs[5] == 1 && b.Pumpstatus && (biofabrica.Valvs[6] == 1 || biofabrica.Valvs[7] == 1 || biofabrica.Valvs[8] == 1) {
								count, vol_tmp := scp_get_volume(scp_biofabrica, scp_biofabrica, scp_dev_volfluxo_out)
								if count >= 0 {
									vol_tmp = vol_tmp // * flow_corfactor_out
									fmt.Println("DEBUG SCP GET ALL DATA: Biorreator", b.BioreactorID, " usando volume vindo do FLUXOUT =", vol_tmp, lastvolout)
									biofabrica.VolumeOut = bio_escala * (math.Trunc(vol_tmp / bio_escala))
									if vol_tmp > lastvolout && lastvolout > 0 {
										biovolout := vol_tmp - lastvolout
										bio[ind].VolInOut -= biovolout
										if bio[ind].VolInOut < 0 {
											bio[ind].VolInOut = 0
										}
										bio[ind].Volume = uint32(bio[ind].VolInOut)
										fmt.Println("DEBUG SCP GET ALL DATA: Biorreator", b.BioreactorID, " NOVO volume =", bio[ind].VolInOut)
										scp_update_biolevel(b.BioreactorID)
									}
									lastvolout = vol_tmp
									hasupdatevolout = true
								} else {
									fmt.Println("ERROR SCP GET ALL DATA: Valor invalido ao ler Volume INFLUXO", count, vol_tmp)
								}
							}

							// NOVO - Em 08/09/23
							if b.Status == bio_update {
								if bio[ind].Volume > 0 {
									bio[ind].Status = bio_ready
								} else {
									bio[ind].Status = bio_empty
								}
							}
							// if volc >= 0 {
							// 	bio[ind].Volume = uint32(volc)
							// 	scp_update_biolevel(b.BioreactorID)
							// 	if volc == 0 && vol0 == 0 && bio[ind].Status != bio_producting && bio[ind].Status != bio_loading && bio[ind].Status != bio_cip {
							// 		bio[ind].Status = bio_empty
							// 	}
							// }

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
									bio_add_message(b.BioreactorID, "AFavor verificar SENSOR de nivel 0", "")
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
									bio_add_message(b.BioreactorID, "AVerifique sensores de Volume", "")
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

						if devmode && bio[ind].Status == bio_update {
							if bio[ind].Volume == 0 {
								bio[ind].Status = bio_empty
							} else {
								bio[ind].Status = bio_ready
							}
						}

					}

				} else if b.Status == bio_error {
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

					if mustupdate_this || rand.Intn(21) == 7 {
						testboot := scp_test_boot(b.IBCID, scp_ibc)
						if testboot == scp_reboot {
							fmt.Println("ERROR GET ALLDATA: Teste do IBC", b.IBCID, "retornou que houve BOOT. Mudando status para ERROR")
							ibc[ind].Status = bio_error
						} else if testboot == scp_err {
							fmt.Println("ERROR GET ALLDATA: Teste do IBC", b.IBCID, "retornou ERRO no test de boot", testboot)
						} else {
							fmt.Println("DEBUG GET ALLDATA: Teste do IBC", b.IBCID, "retornou ", testboot)
						}
					}

					if ind >= 0 && (mustupdate_this || b.Valvs[3] == 1 || b.Valvs[2] == 1) {
						if devmode {
							fmt.Println("DEBUG GET ALLDATA: Lendo dados do IBC", b.IBCID)
						}

						if biofabrica.Useflowin {

							if rand.Intn(3) == 1 {
								scp_get_volume(b.IBCID, scp_ibc, scp_dev_vol0) // força a ler algo só pra atualizar se houver falha
							}

							if b.Valvs[2] == 1 && biofabrica.Valvs[1] == 1 {
								count, vol_tmp := scp_get_volume(scp_biofabrica, scp_biofabrica, scp_dev_volfluxo_in1)
								if count >= 0 {
									vol_tmp = vol_tmp // * flow_corfactor_in1
									fmt.Println("DEBUG SCP GET ALL DATA: IBC", b.IBCID, " usando volume vindo do Totem01 =", vol_tmp, lastvolin)
									biofabrica.VolumeIn1 = bio_escala * (math.Trunc(vol_tmp / bio_escala))
									if vol_tmp > lastvolin && lastvolin > 0 {
										ibcvolin := vol_tmp - lastvolin
										ibc[ind].VolInOut += ibcvolin
										ibc[ind].Volume = uint32(ibc[ind].VolInOut)
										fmt.Println("DEBUG SCP GET ALL DATA: IBC", b.IBCID, " NOVO volume =", ibc[ind].VolInOut)
										// scp_update_biolevel(b.BioreactorID)
									}
									lastvolin = vol_tmp
									hasupdatevolin = true
								} else {
									fmt.Println("ERROR SCP GET ALL DATA: Valor invalido ao ler Volume INFLUXO", count, vol_tmp)
								}
							} else if b.Valvs[3] == 1 && (biofabrica.Valvs[6] == 1 || biofabrica.Valvs[7] == 1 || biofabrica.Valvs[8] == 1) {
								count, vol_tmp := scp_get_volume(scp_biofabrica, scp_biofabrica, scp_dev_volfluxo_out)
								if count >= 0 {
									vol_tmp = vol_tmp // * flow_corfactor_out
									fmt.Println("DEBUG SCP GET ALL DATA: IBC", b.IBCID, " usando volume vindo do FLUXOUT =", vol_tmp, lastvolout)
									biofabrica.VolumeOut = bio_escala * (math.Trunc(vol_tmp / bio_escala))
									if vol_tmp > lastvolout && lastvolout > 0 {
										ibcvolout := vol_tmp - lastvolout
										if ibc[ind].VolInOut > ibcvolout {
											ibc[ind].VolInOut -= ibcvolout
										} else {
											ibc[ind].VolInOut = 0
										}
										if ibc[ind].VolInOut < 0 {
											ibc[ind].VolInOut = 0
										}
										ibc[ind].Volume = uint32(ibc[ind].VolInOut)
										fmt.Println("DEBUG SCP GET ALLDATA: IBC", b.IBCID, " NOVO volume =", ibc[ind].VolInOut)
										// scp_update_biolevel(b.BioreactorID)
									}
									lastvolout = vol_tmp
									hasupdatevolout = true
								} else {
									fmt.Println("ERROR SCP GET ALLDATA: Valor invalido ao ler Volume INFLUXO", count, vol_tmp)
								}
							}

							if ibc[ind].Status == bio_update {
								if ibc[ind].Volume == 0 {
									ibc[ind].Status = bio_empty
								} else {
									ibc[ind].Status = bio_ready
								}
							}

						}
						if scp_onlyread_sensoribc {

							var vol0 float64 = 1
							var dint int

							// dint, _ = scp_get_volume(b.IBCID, scp_ibc, scp_dev_vol0)
							// if dint <= 0 || dint > 1 {
							// 	dint, _ = scp_get_volume(b.IBCID, scp_ibc, scp_dev_vol0)
							// }
							// if dint >= 0 {
							// 	ibc[ind].Vol0 = dint
							// }
							// if ibc[ind].Status == bio_cip && (ibc[ind].Valvs[0] == 1 || ibc[ind].Valvs[1] == 1) {
							// 	fmt.Println("DEBUG GET ALLDATA: CIP EXECUTANDO - IGNORANDO VOLUME ZERO", b.IBCID)
							// } else {
							// 	vol0 = float64(dint)
							// 	fmt.Println("DEBUG GET ALLDATA: Volume ZERO", b.IBCID, ibc_cfg[b.IBCID].Deviceaddr, dint, vol0)
							// }

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
							fmt.Println("DEBUG GET ALLDATA: Volume USOM", b.IBCID, bio_cfg[b.IBCID].Deviceaddr, dint, vol1_pre, vol1)

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
							fmt.Println("DEBUG GET ALLDATA: Volume LASER", b.IBCID, bio_cfg[b.IBCID].Deviceaddr, dint, vol2_pre, vol2)

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

							// var volc float64
							// volc = float64(ibc[ind].Volume)
							// if vol0 == 0 {
							// 	fmt.Println("DEBUG GET ALLDATA: Volume ZERO DETECTADO", b.IBCID)
							// 	volc = 0
							// } else if vol1 < 0 && vol2 < 0 {
							// 	fmt.Println("ERROR GET ALLDATA: IGNORANDO VOLUMES INVALIDOS", b.IBCID, vol1, vol2)
							// } else {
							// 	if ibc[ind].Valvs[3] == 1 { // Desenvase
							// 		if vol1 < 0 {
							// 			if vol2 >= 0 && vol2 < float64(ibc[ind].Volume) {
							// 				volc = vol2
							// 			}
							// 		} else {
							// 			volc = vol1
							// 		}
							// 	} else if ibc[ind].Valvs[2] == 1 { // Carregando
							// 		if ibc[ind].Valvs[1] == 0 { // Sprayball desligado
							// 			if vol1 < 0 {
							// 				if vol2 >= 0 && vol2 > float64(ibc[ind].Volume) {
							// 					volc = vol2
							// 				}
							// 			} else {
							// 				volc = vol1
							// 			}
							// 		}
							// 	} else {
							// 		if ibc[ind].Status == bio_cip && ibc[ind].Valvs[1] == 1 { //  Se for CIP e Sptrayball ligado, ignorar
							// 			fmt.Println("DEBUG GET ALLDATA: CIP+SPRAYBALL - IGNORANDO VOLUMES ", b.IBCID, vol1, vol2)
							// 		} else {
							// 			if vol1 >= 0 {
							// 				volc = vol1
							// 			} else if vol2 >= 0 {
							// 				volc = vol2
							// 			}
							// 		}
							// 	}
							// }

							// if volc >= 0 {
							// 	ibc[ind].Volume = uint32(volc)
							// 	level := (volc / float64(ibc_cfg[b.IBCID].Maxvolume)) * 10.0
							// 	level_int := uint8(level)
							// 	if level_int != ibc[ind].Level {
							// 		ibc[ind].Level = level_int
							// 	}
							// 	if volc == 0 && vol0 == 0 && ibc[ind].Status != bio_loading && ibc[ind].Status != bio_cip {
							// 		ibc[ind].Status = bio_empty
							// 	}
							// }
							// if devmode && ibc[ind].Status == bio_update {
							// 	if ibc[ind].Volume == 0 {
							// 		ibc[ind].Status = bio_empty
							// 	} else {
							// 		ibc[ind].Status = bio_ready
							// 	}
							// }

						}
					}
				} else if b.Status == bio_error {
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

			if !hasupdatevolout && (mustupdate_ibc || biofabrica.Valvs[7] == 1 || biofabrica.Valvs[8] == 1 || biofabrica.Valvs[6] == 1) {
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
	_, err := os.Stat(filename)
	if err != nil {
		if !os.IsNotExist(err) {
			checkErr(err)
		}
		return false
	}
	// fmt.Println("DEBUG: Arquivo encontrado", mf.Name())
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
				if val == (1-value) || val == 3 || value == 3 { // ACRESCENTEI TIPO 3 = LOCK
					if !set_valv_status(dtype, sub[0], sub[1], value) {
						fmt.Println("ERROR SET VALVS VALUE: nao foi possivel setar valvula", p)
						if abort_on_error {
							return -1
						}
					}
					tot++
				} else if val == 1 {
					fmt.Println("ERROR SET VALVS VALUE: valor anterior invalido, nao foi possivel setar valvula", p)
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

func scp_run_linewash(lines string, washtime int) bool {
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
	fmt.Println("DEBUG SCP RUN LINEWASH:: Executando enxague - lines=", lines, "time=", washtime, "pathclean=", pathclean, "pathstr=", pathstr)
	if len(pathstr) == 0 {
		fmt.Println("ERROR SCP RUN LINEWASH:: path WASH linha nao existe", pathclean)
		return false
	}
	// var time_to_clean int64 = int64(paths[pathclean].Cleantime) * 1000
	time_to_clean := washtime * 10
	vpath := scp_splitparam(pathstr, ",")
	if !scp_turn_pump(scp_totem, totem, vpath, 1, true) {
		fmt.Println("ERROR SCP RUN LINEWASH: Falha ao abrir valvulvas e ligar bomba do totem", totem, vpath)
		return false
	}

	for i := 0; i < time_to_clean; i++ {
		if biofabrica.Critical == scp_stopall {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	if !scp_turn_pump(scp_totem, totem, vpath, 0, false) {
		fmt.Println("ERROR SCP RUN LINEWASH: Falha ao fechar valvulvas e desligar bomba do totem", totem, vpath)
		return false
	}
	sl := strings.Replace(lines, "_", "/", -1)
	board_add_message("IEnxague concluído LINHAS "+sl, "")
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
	// var time_to_clean int64 = int64(paths[pathclean].Cleantime) * 1000
	vpath := scp_splitparam(pathstr, ",")
	vpath_peris := scp_splitparam(pathstr, ",")
	perisvalv := totem_str + "/V2"
	n := len(vpath)
	vpath_peris = append(vpath_peris[1:n-1], perisvalv)
	vpath_peris = append(vpath_peris, "END")
	fmt.Println("DEBUG SCP RUN LINEWASH: vpath ", vpath)
	fmt.Println("DEBUG SCP RUN LINEWASH: vpath peris ", vpath_peris)

	all_peris := [2]string{"P1", "P2"}
	// tmax := scp_timewaitvalvs * 10
	// if devmode || testmode {
	// 	tmax = scp_timeoutdefault / 100
	// }

	for _, peris_str := range all_peris {

		fmt.Println("DEBUG SCP LINECIP: Ligando peristalticas", peris_str)
		if test_path(vpath_peris, 0) {
			if set_valvs_value(vpath_peris, 1, true) < 0 {
				fmt.Println("ERROR SCP RUN LINEWASH: ERRO ao abrir valvulas no path ", vpath_peris)
				return false
			}
		} else {
			fmt.Println("ERROR SCP RUN LINEWASH: ERRO nas valvulas no path - alguma valvula aberta ", vpath_peris)
			return false
		}

		// for i := 0; i < tmax; i++ {
		// 	time.Sleep(100 * time.Millisecond)
		// }

		time.Sleep((scp_timewaitvalvs * time.Millisecond))

		fmt.Println("DEBUG SCP LINECIP: Ligando peristalticas do totem", totem_str, peris_str)

		if !scp_turn_peris(scp_totem, totem_str, peris_str, 1) {
			fmt.Println("ERROR SCP RUN LINEWASH: ERROR ao ligar peristaltica em", totem_str, peris_str)
			return false
		}

		// time.Sleep(scp_timelinecip * time.Second) // VALIDAR TEMPO DE BLEND na LINHAv

		for i := 0; i < time_cipline_blend*10; i++ {
			if biofabrica.Critical == scp_stopall {
				break
			}
			time.Sleep(100 * time.Millisecond)
		}

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

		// for i := 0; i < tmax; i++ {
		// 	time.Sleep(100 * time.Millisecond)
		// }

		if biofabrica.Critical == scp_stopall {
			return false
		}

		time.Sleep((scp_timewaitvalvs * time.Microsecond))

		if !scp_turn_pump(scp_totem, totem_str, vpath, 1, true) {
			fmt.Println("ERROR SCP RUN LINEWASH: Falha ao abrir valvulvas e ligar bomba do totem", totem, vpath)
			return false
		}

		time.Sleep(time.Duration(time_cipline_clean) * time.Millisecond)

		if !scp_turn_pump(scp_totem, totem_str, vpath, 0, false) {
			fmt.Println("ERROR SCP RUN LINEWASH: Falha ao fechar valvulvas e desligar bomba do totem", totem, vpath)
			return false
		}

		time.Sleep(10000 * time.Millisecond)

	}

	board_add_message("ICIP concluído Linhas "+lines, "")
	return true
}

func MutexLocked(m *sync.Mutex) bool {
	const mutexLocked = 1
	state := reflect.ValueOf(m).Elem().FieldByName("state")
	return state.Int()&mutexLocked == mutexLocked
}

func scp_run_withdraw(devtype string, devid string, linewash bool, untilempty bool) int {
	if MutexLocked(&withdrawmutex) {
		waitlist_add_message("A"+devid+" aguardando termino de outro desenvase", devid+"WDBUSY")
		if devtype == scp_bioreactor {
			bio_add_message(devid, "ABiorreator aguardando termino de outro desenvase", "WDBUSY")
		}
	}
	withdrawmutex.Lock()
	turn_withdraw_var(true)
	defer withdrawmutex.Unlock()
	defer turn_withdraw_var(false)
	waitlist_del_message(devid + "WDBUSY")
	if devtype == scp_bioreactor {
		bio_del_message(devid, "WDBUSY")
	}
	if biofabrica.PIntStatus != bio_ready {
		fmt.Println("ERROR RUN WITHDRAW: Falha no Painel Intermediário, impossivel executar withdraw", devid)
		board_add_message("E"+devid+" não pode executar operação por falha no Painel Intermediário. Favor verificar", devid+"PIERROR")
		return -1
	}
	board_del_message(devid + "PIERROR")
	switch devtype {
	case scp_bioreactor:
		ind := get_bio_index(devid)
		if ind < 0 {
			fmt.Println("ERROR RUN WITHDRAW 01: Biorreator nao existe", devid)
			return -1
		}
		if bio[ind].MustPause || bio[ind].MustStop {
			return -1
		}
		tout := get_scp_type(bio[ind].OutID)
		if (tout == scp_out || tout == scp_drop) && biofabrica.POutStatus != bio_ready {
			fmt.Println("ERROR RUN WITHDRAW: Falha no Painel de Desenvase, impossivel executar withdraw", devid, tout, biofabrica.POutStatus)
			board_add_message("E"+devid+" não pode executar operação por falha no Painel de Desenvase. Favor verificar", devid+"BIOPOUTERROR")
			return -1
		}
		board_del_message(devid + "BIOPOUTERROR")
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
			waitlist_add_message("ADesenvase do Biorreator "+devid+" aguardando liberação da Linha", devid+"WITHDRAWBUSY")
			bio_add_message(devid, "ABiorreator aguardando liberação da Linha", "WITHDRAWBUSY")
			return -1
		}
		board_add_message("CDesenvase "+devid+" para "+bio[ind].OutID, "")
		fmt.Println("DEBUG RUN WITHDRAW: Desenvase do Biorreator", devid, " para", bio[ind].OutID, "volume=", bio[ind].Withdraw, " inicial=", bio[ind].Volume)
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
						waitlist_add_message("ADesenvase do Biorreator "+devid+" aguardando liberação da Linha", devid+"WITHDRAWBUSY")
						bio_add_message(devid, "ABiorreator aguardando liberação da Linha", "WITHDRAWBUSY")
						set_valvs_value(pilha, 0, false) // undo
						return -1
					}
				} else if val == 1 {
					fmt.Println("ERROR RUN WITHDRAW 04: valvula ja aberta", p)
					waitlist_add_message("ADesenvase do Biorreator "+devid+" aguardando liberação da Linha", devid+"WITHDRAWBUSY")
					bio_add_message(devid, "ABiorreator aguardando liberação da Linha", "WITHDRAWBUSY")
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
		waitlist_del_message(devid + "WITHDRAWBUSY")
		bio_del_message(devid, "WITHDRAWBUSY")
		// fmt.Println(pilha)
		vol_ini := bio[ind].Volume
		bio[ind].Status = bio_unloading
		time.Sleep(scp_timewaitvalvs * time.Millisecond)
		fmt.Println("WARN RUN WITHDRAW 07: Ligando bomba", devid)
		biodev := bio_cfg[devid].Deviceaddr
		// bioscr := bio_cfg[devid].Screenaddr
		pumpdev := bio_cfg[devid].Pump_dev
		bio[ind].Pumpstatus = true
		cmd1 := "CMD/" + biodev + "/PUT/" + pumpdev + ",1/END"
		// cmd2 := "CMD/" + bioscr + "/PUT/S270,1/END"
		ret1 := scp_sendmsg_orch(cmd1)
		fmt.Println("DEBUG RUN WITHDRAW 08: CMD1 =", cmd1, " RET=", ret1)
		// ret2 := scp_sendmsg_orch(cmd2)
		// fmt.Println("DEBUG RUN WITHDRAW 09: CMD2 =", cmd2, " RET=", ret2)
		if !strings.Contains(ret1, scp_ack) && !devmode {
			fmt.Println("ERROR RUN WITHDRAW 10: BIORREATOR falha ao ligar bomba")
			// cmd2 := "CMD/" + bioscr + "/PUT/S270,0/END"
			// scp_sendmsg_orch(cmd2)
			set_valvs_value(pilha, 0, false)
			return -1
		}
		var vol_out int64
		var vol_bio_out_start float64
		vol_bio_init := bio[ind].Volume
		vol_bio_last := bio[ind].Volume
		if bio[ind].OutID == scp_out || bio[ind].OutID == scp_drop {
			var count int
			count, vol_bio_out_start = scp_get_volume(scp_biofabrica, scp_biofabrica, scp_dev_volfluxo_out)
			if count < 0 {
				vol_bio_out_start = biofabrica.VolumeOut
			}
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
		vol_ibc_ini := float64(-1)
		mustwaittime := false
		waittime := float64(0)
		ibc_ind := -1
		if untilempty && biofabrica.Useflowin && get_scp_type(bio[ind].OutID) == scp_ibc {
			ibc_ind = get_ibc_index(bio[ind].OutID)
			if ibc_ind >= 0 {
				vol_ibc_ini = float64(ibc[ibc_ind].VolInOut)
				ibc[ibc_ind].MainStatus = mainstatus_org
			}
			// bio[ind].ShowVol = false
			mustwaittime = true
			waittime = float64(bio[ind].Volume)*bio_emptying_rate + 20
		}
		t_last_volchange := time.Now()
		abort_due_novolchange_mustzero := false
		for {
			vol_now := bio[ind].Volume
			// t_now := time.Now()
			t_elapsed = time.Since(t_start).Seconds()
			vol_out = int64(vol_ini - vol_now)
			vol_bio_out_now := biofabrica.VolumeOut - vol_bio_out_start
			if bio[ind].Withdraw == 0 {
				break
			}
			if vol_now == vol_bio_last {
				t_elapsed_volchage := time.Since(t_last_volchange).Seconds()
				if t_elapsed_volchage > 25 {
					if vol_now < vol_bio_init && vol_now <= bio_deltavolzero {
						abort_due_novolchange_mustzero = true
						fmt.Println("DEBUG RUN WITH DRAW: Desenvase abortado por volume nao variar em 25 seg e sendo ZERADO", devid)
					} else {
						board_add_message("A"+devid+" Desenvase/Transferência abortado por volume não variar em 25s. Favor verificar equipamentos", "")
						fmt.Println("ERROR RUN WITH DRAW: Desenvase abortado por volume nao variar em 25 seg", devid)
					}
					break
				}
			} else {
				vol_bio_last = vol_now
				t_last_volchange = time.Now()
			}
			if untilempty {
				if mustwaittime {
					if t_elapsed >= waittime {
						fmt.Println("DEBUG RUN WITH DRAW: Transferência UNTILEMPTY interrompida pois o tempo excedeu o tempo previsto", devid, "previsto =", waittime)
						break
					}
				} else if vol_now == 0 {
					fmt.Println("DEBUG RUN WITH DRAW: Transferência UNTILEMPTY interrompida pois volume atual ZERO", devid)
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
			if biofabrica.Useflowin && mustwaittime && rand.Intn(21) == 7 { // } int32(t_elapsed)%7 == 0 { // depois testar rand.Intn(21) == 7
				volout := t_elapsed / bio_emptying_rate
				vol_tmp := float64(vol_bio_init) - volout
				if vol_tmp < 0 {
					vol_tmp = 0
				}
				bio[ind].VolInOut = vol_tmp
				bio[ind].Volume = uint32(vol_tmp)
				fmt.Println("DEBUG RUN WITHDRAW: Desenvase de:", bio[ind].BioreactorID, "para:", bio[ind].OutID, "/", ibc_ind, "volout=", volout, "voltmp=", vol_tmp)
				if ibc_ind >= 0 && vol_ibc_ini >= 0 {
					ibc[ibc_ind].VolInOut = vol_ibc_ini + volout
					ibc[ibc_ind].Volume = uint32(ibc[ibc_ind].VolInOut)
					ibc[ibc_ind].OrgCode = bio[ind].OrgCode
					ibc[ibc_ind].Organism = bio[ind].Organism
					ibc[ibc_ind].MainStatus = mainstatus_org
					scp_update_ibclevel(ibc[ibc_ind].IBCID)
				}
				scp_update_biolevel(bio[ind].BioreactorID)
				// scp_update_screen_vol(bio[ind].BioreactorID)
			}
			time.Sleep(scp_refreshwait * time.Millisecond)
		}
		if abort_due_novolchange_mustzero {
			bio[ind].VolInOut = 0
			bio[ind].Volume = 0
		}
		if bio[ind].Volume == 0 {
			for i := 0; i < 250 && bio[ind].Withdraw != 0; i++ {
				// if bio[ind].Vol0 == 0 {
				// 	break
				// }
				time.Sleep(100 * time.Millisecond)
			}
		}
		if biofabrica.Useflowin && untilempty {
			if bio[ind].Withdraw != 0 {
				bio[ind].Volume = 0
				bio[ind].VolInOut = 0
				if ibc_ind >= 0 {
					ibc[ibc_ind].VolInOut = vol_ibc_ini + float64(vol_bio_init)
					ibc[ibc_ind].Volume = uint32(ibc[ibc_ind].VolInOut)
				}
			} else if mustwaittime {
				volout := t_elapsed / bio_emptying_rate
				vol_tmp := float64(vol_bio_init) - volout
				if vol_tmp < 0 {
					vol_tmp = 0
				}
				bio[ind].VolInOut = vol_tmp
				bio[ind].Volume = uint32(vol_tmp)
			}
		}
		if bio[ind].Volume == 0 {
			if !bio[ind].MustPause && !bio[ind].MustStop {
				bio[ind].Status = bio_empty
			}
			if prev_status == bio_ready {
				prev_status = bio_empty
				bio[ind].Step = [2]int{0, 0}
				bio[ind].Timetotal = [2]int{0, 0}
				bio[ind].Timeleft = [2]int{0, 0}
			}
			bio[ind].Level = 0
		}

		scp_update_biolevel(bio[ind].BioreactorID)
		// scp_update_screen_vol(bio[ind].BioreactorID)
		// scp_update_screen_steps(bio[ind].BioreactorID)
		// scp_update_screen_times(bio[ind].BioreactorID)

		bio[ind].Withdraw = 0

		// board_add_message("IDesenvase concluido", "")
		fmt.Println("WARN RUN WITHDRAW 13: Desligando bomba", devid)

		set_valvs_value(pilha, 0, false)
		if untilempty {
			time.Sleep((scp_timewaitvalvs / 3) * 2 * time.Millisecond)
		}

		bio[ind].Pumpstatus = false
		cmd1 = "CMD/" + biodev + "/PUT/" + pumpdev + ",0/END"
		// cmd2 = "CMD/" + bioscr + "/PUT/S270,0/END"
		ret1 = scp_sendmsg_orch(cmd1)
		fmt.Println("DEBUG RUN WITHDRAW 14: CMD1 =", cmd1, " RET=", ret1)
		// ret2 = scp_sendmsg_orch(cmd2)
		// fmt.Println("DEBUG RUN WITHDRAW 15: CMD2 =", cmd2, " RET=", ret2)

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
				board_add_message("IEnxague LINHAS 2/4", "")
			} else if dest_type == scp_ibc {
				pathclean = "TOTEM02-CLEAN3"
				board_add_message("IEnxague LINHAS 2/3", "")
			} else {
				fmt.Println("ERROR RUN WITHDRAW 16: destino para clean desconhecido", dest_type)
				return 0
			}
			if pathclean == "TOTEM02-CLEAN4" {
				scp_run_linewash(line_24, 30)
			} else if pathclean == "TOTEM02-CLEAN3" {
				scp_run_linewash(line_23, 30)
			}
			// pathstr = paths[pathclean].Path
			// if len(pathstr) == 0 {
			// 	fmt.Println("ERROR RUN WITHDRAW 17: path WASH linha nao existe", pathclean)
			// 	return 0
			// }
			// var time_to_clean int64 = int64(paths[pathclean].Cleantime) * 1000
			// vpath = scp_splitparam(pathstr, ",")
			// if !test_path(vpath, 0) {
			// 	fmt.Println("ERROR RUN WITHDRAW 18: falha de valvula no path", pathstr)
			// 	return 0
			// }
			// if set_valvs_value(vpath, 1, true) < 1 {
			// 	fmt.Println("ERROR RUN WITHDRAW 19: Falha ao abrir valvulas CLEAN linha", pathstr)
			// 	set_valvs_value(vpath, 0, false)
			// }
			// time.Sleep(scp_timewaitvalvs * time.Millisecond)
			// fmt.Println("WARN RUN WITHDRAW 20: Ligando bomba TOTEM02", devid)
			// tind := get_totem_index("TOTEM02")
			// if tind < 0 {
			// 	fmt.Println("WARN RUN WITHDRAW 21: TOTEM02 nao encontrado", totem)
			// }
			// totemdev := totem_cfg["TOTEM02"].Deviceaddr
			// pumpdev = totem_cfg["TOTEM02"].Pumpdev
			// totem[tind].Pumpstatus = true
			// cmd1 = "CMD/" + totemdev + "/PUT/" + pumpdev + ",1/END"
			// ret1 = scp_sendmsg_orch(cmd1)
			// fmt.Println("DEBUG RUN WITHDRAW 22: CMD1 =", cmd1, " RET=", ret1)
			// if !strings.Contains(ret1, scp_ack) && !devmode {
			// 	fmt.Println("ERROR RUN WITHDRAW 23: BIORREATOR falha ao ligar bomba TOTEM02")
			// 	totem[tind].Pumpstatus = false
			// 	set_valvs_value(vpath, 0, false)
			// 	return 0
			// }
			// time.Sleep(time.Duration(time_to_clean) * time.Millisecond)
			// fmt.Println("WARN RUN WITHDRAW 24: Desligando bomba TOTEM02", devid)
			// totem[tind].Pumpstatus = false
			// cmd1 = "CMD/" + totemdev + "/PUT/" + pumpdev + ",0/END"
			// ret1 = scp_sendmsg_orch(cmd1)
			// fmt.Println("DEBUG RUN WITHDRAW 25: CMD1 =", cmd1, " RET=", ret1)
			// if !strings.Contains(ret1, scp_ack) && !devmode {
			// 	fmt.Println("ERROR RUN WITHDRAW 26: BIORREATOR falha ao ligar bomba TOTEM02")
			// 	totem[tind].Pumpstatus = false
			// 	set_valvs_value(vpath, 0, false)
			// 	return 0
			// }
			// set_valvs_value(vpath, 0, false)
			// board_add_message("IEnxague concluído", "")
		}
		if !bio[ind].MustPause && !bio[ind].MustStop {
			bio[ind].Status = prev_status
		}

	case scp_ibc:
		ind := get_ibc_index(devid)
		if ind < 0 {
			fmt.Println("ERROR RUN WITHDRAW 01: IBC nao existe", devid)
			return -1
		}
		if ibc[ind].MustPause || ibc[ind].MustStop {
			return -1
		}
		if biofabrica.POutStatus != bio_ready {
			fmt.Println("ERROR RUN WITHDRAW: Falha no Painel de Desenvase, impossivel executar withdraw", devid, biofabrica.POutStatus)
			board_add_message("E"+devid+" não pode executar operação por falha no Painel de Desenvase. Favor verificar", devid+"IBCPOUTERROR")
			return -1
		}
		board_del_message(devid + "IBCPOUTERROR")
		prev_status := ibc[ind].Status
		pathid := devid + "-" + ibc[ind].OutID
		pathstr := paths[pathid].Path
		if len(pathstr) == 0 {
			fmt.Println("ERROR RUN WITHDRAW 27: path nao existe", pathid)
			return -1
		}
		vpath := scp_splitparam(pathstr, ",")
		if !test_path(vpath, 0) {
			fmt.Println("ERROR RUN WITHDRAW 28: falha de valvula no path", pathid)
			waitlist_add_message("ADesenvase do "+devid+" aguardando liberação da Linha", devid+"WITHDRAWBUSY")
			return -1
		}
		// board_add_message("CDesenvase iniciado "+devid+" para "+ibc[ind].OutID, "")
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
						waitlist_add_message("ADesenvase do "+devid+" aguardando liberação da Linha", devid+"WITHDRAWBUSY")
						return -1
					}
				} else if val == 1 {
					fmt.Println("ERROR RUN WITHDRAW 30: valvula ja aberta", p)
					waitlist_add_message("ADesenvase do "+devid+" aguardando liberação da Linha", devid+"WITHDRAWBUSY")
					// bio_add_message(devid, "ABiorreator aguardando liberação da Linha", "WITHDRAWBUSY")
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
		waitlist_del_message(devid + "WITHDRAWBUSY")
		// bio_del_message(devid, "WITHDRAWBUSY")
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
		var vol_out float64
		var vol_bio_out_start float64
		use_volfluxo := true
		// if ibc[ind].OutID == scp_out || ibc[ind].OutID == scp_drop {
		// use_volfluxo = true
		var count int
		count, vol_bio_out_start = scp_get_volume(scp_biofabrica, scp_biofabrica, scp_dev_volfluxo_out)
		if count < 0 {
			vol_bio_out_start = biofabrica.VolumeOut
		}
		// } else {
		// 	vol_bio_out_start = -1
		// }
		var maxtime float64
		if devmode || testmode {
			maxtime = 60
		} else {
			maxtime = scp_maxtimewithdraw
		}
		t_start := time.Now()
		vol_bio_last := ibc[ind].Volume
		t_last_volchange := time.Now()
		ibc7_ind := get_ibc_index(last_ibc)
		ibc7_vol_ini := float64(-1)
		if ibc7_ind >= 0 {
			ibc7_vol_ini = ibc[ibc7_ind].VolInOut
		}
		abort_due_novolchange_mustzero := false
		for {
			vol_now := ibc[ind].Volume
			t_elapsed := time.Since(t_start).Seconds()
			vol_out = float64(vol_ini) - float64(vol_now)
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
			if ibc[ind].OutID == last_ibc {
				if ibc7_ind >= 0 {
					ibc[ibc7_ind].VolInOut = ibc7_vol_ini + vol_bio_out_now
					if ibc[ibc7_ind].VolInOut < 0 {
						ibc[ibc7_ind].VolInOut = 0
					}
					ibc[ibc7_ind].Volume = uint32(ibc[ibc7_ind].VolInOut)
					scp_update_ibclevel(ibc[ibc7_ind].IBCID)
				} else {
					fmt.Println("ERROR RUN WITHDRAW: Ultimo IBC nao existe", last_ibc)
				}
			}
			// fmt.Println("vout=", vout, ibc[ind].VolumeOut)
			if ibc[ind].Withdraw == 0 {
				break
			}
			if vol_now == vol_bio_last {
				t_elapsed_volchage := time.Since(t_last_volchange).Seconds()
				if t_elapsed_volchage > 25 {
					if vol_now < vol_ini && vol_now < ibc_deltavolzero {
						abort_due_novolchange_mustzero = true
						fmt.Println("DEBUG RUN WITH DRAW: Desenvase abortado por volume nao variar em 25 seg e sendo ZERADO", devid)
					} else {
						board_add_message("A"+devid+" Desenvase/Transferência abortado por volume não variar em 25s. Favor verificar equipamentos", "")
						fmt.Println("ERROR RUN WITH DRAW: Desenvase abortado por volume nao variar em 25 seg", devid)
					}
					fmt.Println("DEBUG RUN WITH DRAW: Desenvase abortado por volume nao variar em 25 seg", devid)
					break
				}
			} else {
				vol_bio_last = vol_now
				t_last_volchange = time.Now()
			}
			if untilempty {
				if vol_now == 0 {
					break
				}
			} else if vol_now == 0 || (vol_now < vol_ini && vol_out >= float64(ibc[ind].Withdraw)) {
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
		if abort_due_novolchange_mustzero {
			ibc[ind].VolInOut = 0
			ibc[ind].Volume = 0
		}
		if ibc[ind].Volume == 0 {
			for i := 0; i < 250 && ibc[ind].Withdraw != 0; i++ { // 25 seg além do ZERO
				time.Sleep(100 * time.Millisecond)
			}
		}
		ibc[ind].Withdraw = 0
		// board_add_message("IDesenvase IBC "+devid+" concluido", "")
		// board_del_message(devid + "WITHDRAW")
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
			board_add_message("IEnxague LINHA 4", "")
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
			// board_add_message("IEnxague concluído", "")
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
				board_add_message("ILimpando LINHA 3", "")
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
				// board_add_message("IEnxague concluído", "")
			}
		}
		if ibc[ind].VolInOut < 0 {
			ibc[ind].VolInOut = 0
			ibc[ind].Volume = 0
		}
		if !ibc[ind].MustPause && !ibc[ind].MustStop {
			if ibc[ind].MainStatus != mainstatus_cip && ibc[ind].Volume == 0 {
				ibc[ind].Status = bio_empty
			} else {
				ibc[ind].Status = prev_status
			}
		}

	default:
		fmt.Println("DEBUG RUN WITHDRAW 58: Devtype invalido", devtype, devid)
	}
	return 0
}

func scp_turn_aero(bioid string, changevalvs bool, value int, percent int, musttest bool) bool {
	ind := get_bio_index(bioid)
	if ind < 0 {
		fmt.Println("ERROR SCP TURN: Biorreator nao existe", bioid)
		return false
	}
	devaddr := bio_cfg[bioid].Deviceaddr
	// scraddr := bio_cfg[bioid].Screenaddr
	aerorele := bio_cfg[bioid].Aero_rele
	aerodev := bio_cfg[bioid].Aero_dev
	dev_valvs := []string{bioid + "/V1", bioid + "/V2"}

	// if value == scp_off {
	// 	cmd0 := fmt.Sprintf("CMD/%s/PUT/%s,%d/END", devaddr, aerorele, value)
	// 	ret0 := scp_sendmsg_orch(cmd0)
	// 	fmt.Println("DEBUG SCP TURN AERO: CMD =", cmd0, "\tRET =", ret0)
	// 	if !strings.Contains(ret0, scp_ack) && !devmode {
	// 		fmt.Println("ERROR SCP TURN AERO:", bioid, " ERROR ao definir valor[", value, "] rele aerador ", ret0)
	// 		if changevalvs {
	// 			set_valvs_value(dev_valvs, 1-value, false)
	// 		}
	// 		return false
	// 	}
	// 	bio[ind].Aerator = false
	// 	cmds := fmt.Sprintf("CMD/%s/PUT/S271,%d/END", scraddr, value)
	// 	rets := scp_sendmsg_orch(cmds)
	// 	fmt.Println("DEBUG SCP TURN AERO: CMD =", cmds, "\tRET =", rets)
	// 	if !strings.Contains(rets, scp_ack) && !devmode {
	// 		fmt.Println("ERROR SCP TURN AERO:", bioid, " ERROR ao mudar aerador na screen ", scraddr, rets)
	// 	}

	// }

	if changevalvs {
		// musttest := value == 1
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

	if changevalvs {
		tmax := 10 // (2 * scp_timewaitvalvs / 3) / 1000
		for i := 0; i < tmax; i++ {
			// if bio[ind].MustPause || bio[ind].MustStop {
			// 	break
			// }
			time.Sleep(1000 * time.Millisecond)
		}
	}

	aerovalue := int(255.0 * (float32(percent) / 100.0))
	cmd1 := fmt.Sprintf("CMD/%s/PUT/%s,%d/END", devaddr, aerodev, aerovalue)
	ret1 := scp_sendmsg_orch(cmd1)
	fmt.Println("DEBUG SCP TURN AERO: CMD =", cmd1, "\tRET =", ret1)
	if !strings.Contains(ret1, scp_ack) && !devmode && biofabrica.Critical != scp_netfail && bio[ind].Status != bio_error {
		fmt.Println("ERROR SCP TURN AERO:", bioid, " ERROR ao definir ", percent, "% aerador", ret1)
		if changevalvs {
			set_valvs_value(dev_valvs, 1-value, false)
		}
		return false
	}

	if true {
		cmd2 := fmt.Sprintf("CMD/%s/PUT/%s,%d/END", devaddr, aerorele, value)
		ret2 := scp_sendmsg_orch(cmd2)
		fmt.Println("DEBUG SCP TURN AERO: CMD =", cmd2, "\tRET =", ret2)
		if !strings.Contains(ret2, scp_ack) && !devmode && biofabrica.Critical != scp_netfail && bio[ind].Status != bio_error {
			fmt.Println("ERROR SCP TURN AERO:", bioid, " ERROR ao definir valor[", value, "] rele aerador ", ret2)
			if value == 1 && changevalvs {
				set_valvs_value(dev_valvs, 0, false)
			}
			// if changevalvs {
			// 	set_valvs_value(dev_valvs, 1-value, false)
			// }
			// cmdoff := fmt.Sprintf("CMD/%s/PUT/%s,%d/END", devaddr, aerodev, 0)
			// ret1 = scp_sendmsg_orch(cmdoff)
			return false
		}
		if value == 1 {
			bio[ind].Aerator = true
		} else {
			bio[ind].Aerator = false
		}
		// cmds := fmt.Sprintf("CMD/%s/PUT/S271,%d/END", scraddr, value)
		// rets := scp_sendmsg_orch(cmds)
		// fmt.Println("DEBUG SCP TURN AERO: CMD =", cmds, "\tRET =", rets)
		// if !strings.Contains(rets, scp_ack) && !devmode {
		// 	fmt.Println("ERROR SCP TURN AERO:", bioid, " ERROR ao mudar aerador na screen ", scraddr, rets)
		// }
	}

	if changevalvs {
		tmax := 5 //(scp_timewaitvalvs / 3) / 1000
		for i := 0; i < tmax; i++ {
			// if bio[ind].MustPause || bio[ind].MustStop {
			// 	break
			// }
			time.Sleep(1000 * time.Millisecond)
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
	// scrdev := ""
	devaddr := ""
	switch devtype {
	case scp_bioreactor:
		peris_dev = bio_cfg[bioid].Peris_dev[peris_int-1]
		devaddr = bio_cfg[bioid].Deviceaddr
		// scrdev = bio_cfg[bioid].Screenaddr
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
	if !strings.Contains(ret0, scp_ack) && !devmode && biofabrica.Critical != scp_netfail {
		if devtype == scp_bioreactor && bio[ind].Status == bio_error {
			fmt.Println("DEBUG SCP TURN PERIS:", bioid, " com error, ignorando erro vindo do ORCH ", ret0)
		} else {
			fmt.Println("ERROR SCP TURN PERIS:", bioid, " ERROR ao definir valor[", value, "] peristaltica ", ret0)
			return false
		}

	}
	// fmt.Println("DEBUG SCP TURN PERIS: Screen", scrdev)
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
	if bio[ind].Temperature > TEMPMAX {
		fmt.Println("ERROR SCP TURN HEATER: Temperatura do Biorreator acima do limite", bioid, " temp=", bio[ind].Temperature, "  TEMPMAX=", TEMPMAX)
		return false
	}
	if bio[ind].Volume == 0 {
		fmt.Println("ERROR SCP TURN HEATER: Nao é possível ligar resistencia com Biorreator VAZIO", bioid, " volume=", bio[ind].Volume)
		return false
	}
	bio[ind].TempMax = maxtemp
	devaddr := bio_cfg[bioid].Deviceaddr
	heater_dev := bio_cfg[bioid].Heater
	value_str := "0"
	if value {
		value_str = "1"
		if maxtemp <= 0 {
			fmt.Println("ERROR SCP TURN HEATER: Não é permitido ligar heater sem definir temperatura máxima", bioid, maxtemp)
			return false
		}
	}
	cmd0 := fmt.Sprintf("CMD/%s/PUT/%s,%s/END", devaddr, heater_dev, value_str)
	ret0 := scp_sendmsg_orch(cmd0)
	fmt.Println("DEBUG SCP TURN HEATER: CMD =", cmd0, "\tRET =", ret0)
	if !strings.Contains(ret0, scp_ack) && !devmode && biofabrica.Critical != scp_netfail {
		fmt.Println("ERROR SCP TURN HEATER:", bioid, " ERROR ao definir valor[", value, "] aquecedor ", ret0)
		return false
	}
	bio[ind].Heater = value
	return true
}

func scp_turn_pump(devtype string, main_id string, valvs []string, value int, musttest bool) bool {
	var devaddr, pumpdev string
	var ind int
	// scraddr := ""
	switch devtype {
	case scp_bioreactor:
		ind = get_bio_index(main_id)
		if ind < 0 {
			fmt.Println("ERROR SCP TURN PUMP: Biorreator nao existe", main_id)
			return false
		}
		devaddr = bio_cfg[main_id].Deviceaddr
		pumpdev = bio_cfg[main_id].Pump_dev
		// scraddr = bio_cfg[main_id].Screenaddr

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
	case scp_biofabrica:
		devaddr = biofabrica_cfg["PBF01"].Deviceaddr
		if len(devaddr) == 0 {
			fmt.Println("ERROR SCP TURN PUMP: Valvula PBF01 nao existe", main_id)
			return false
		}
		pumpdev = biofabrica_cfg["PBF01"].Deviceport

	default:
		fmt.Println("ERROR SCP TURN PUMP: Dispositivo nao suportado", devtype, main_id)
	}

	// if value == scp_off {
	// 	cmd := fmt.Sprintf("CMD/%s/PUT/%s,%d/END", devaddr, pumpdev, value)
	// 	ret := scp_sendmsg_orch(cmd)
	// 	fmt.Println("DEBUG SCP TURN PUMP: CMD =", cmd, "\tRET =", ret)
	// 	if !strings.Contains(ret, scp_ack) && !devmode {
	// 		fmt.Println("ERROR SCP TURN PUMP:", main_id, " ERROR ao definir ", value, " bomba", ret)
	// 		if len(valvs) > 0 {
	// 			set_valvs_value(valvs, 1-value, false)
	// 			time.Sleep(scp_timewaitvalvs * time.Millisecond)
	// 		}
	// 		return false
	// 	}
	// 	switch devtype {
	// 	case scp_bioreactor:
	// 		bio[ind].Pumpstatus = false
	// 	case scp_ibc:
	// 		ibc[ind].Pumpstatus = false
	// 	case scp_totem:
	// 		totem[ind].Pumpstatus = false
	// 	case scp_biofabrica:
	// 		biofabrica.Pumpwithdraw = false
	// 	}
	// 	// if len(scraddr) > 0 {
	// 	// 	cmds := fmt.Sprintf("CMD/%s/PUT/S270,%d/END", scraddr, value)
	// 	// 	rets := scp_sendmsg_orch(cmds)
	// 	// 	fmt.Println("DEBUG SCP TURN AERO: CMD =", cmds, "\tRET =", rets)
	// 	// 	if !strings.Contains(rets, scp_ack) && !devmode {
	// 	// 		fmt.Println("ERROR SCP TURN AERO: ERROR ao mudar bomba na screen ", scraddr, rets)
	// 	// 	}
	// 	// }
	// }

	// musttest := value == 1
	if test_path(valvs, 1-value) || !musttest {
		if set_valvs_value(valvs, value, musttest) < 0 {
			fmt.Println("ERROR SCP TURN PUMP:", devtype, " ERROR ao definir valor [", value, "] das valvulas", valvs)
			return false
		}
	} else {
		fmt.Println("ERROR SCP TURN PUMP:", devtype, " ERROR nas valvulas", valvs, "1-value=", 1-value)
		return false
	}

	tmax := 10 // scp_timewaitvalvs / 1000
	for i := 0; i < tmax; i++ {
		// switch devtype {
		// case scp_bioreactor:
		// 	// if bio[ind].MustPause || bio[ind].MustPause {
		// 	// 	i = tmax
		// 	// }
		// }
		time.Sleep(1000 * time.Millisecond)
	}

	if true { // value == scp_on
		cmd := fmt.Sprintf("CMD/%s/PUT/%s,%d/END", devaddr, pumpdev, value)
		ret := scp_sendmsg_orch(cmd)
		fmt.Println("DEBUG SCP TURN PUMP: CMD =", cmd, "\tRET =", ret)
		if !strings.Contains(ret, scp_ack) && !devmode && biofabrica.Critical != scp_netfail {
			// Mudança para testar falha no equipamento
			if (devtype == scp_bioreactor && bio[ind].Status != bio_error) || (devtype == scp_ibc && ibc[ind].Status != bio_error) || (devtype == scp_totem && totem[ind].Status != bio_error) ||
				(devtype == scp_biofabrica && biofabrica.Status != scp_fail) {
				fmt.Println("ERROR SCP TURN PUMP:", main_id, " ERROR ao definir ", value, " bomba", ret)
				if len(valvs) > 0 && value == 1 {
					set_valvs_value(valvs, 0, false)
					time.Sleep(scp_timewaitvalvs * time.Millisecond)
				}
				return false
			}
		}
		if value == 0 {
			switch devtype {
			case scp_bioreactor:
				bio[ind].Pumpstatus = false
			case scp_ibc:
				ibc[ind].Pumpstatus = false
			case scp_totem:
				totem[ind].Pumpstatus = false
			case scp_biofabrica:
				biofabrica.Pumpwithdraw = false
			}
		} else {
			switch devtype {
			case scp_bioreactor:
				bio[ind].Pumpstatus = true
			case scp_ibc:
				ibc[ind].Pumpstatus = true
			case scp_totem:
				totem[ind].Pumpstatus = true
			case scp_biofabrica:
				biofabrica.Pumpwithdraw = true
			}
		}

		// if len(scraddr) > 0 {
		// 	cmds := fmt.Sprintf("CMD/%s/PUT/S270,%d/END", scraddr, value)
		// 	rets := scp_sendmsg_orch(cmds)
		// 	fmt.Println("DEBUG SCP TURN AERO: CMD =", cmds, "\tRET =", rets)
		// 	if !strings.Contains(rets, scp_ack) && !devmode {
		// 		fmt.Println("ERROR SCP TURN AERO: ERROR ao mudar bomba na screen ", scraddr, rets)
		// 	}
		// }
	}

	tmax = 5 // scp_timewaitvalvs / 1000
	for i := 0; i < tmax; i++ {
		// switch devtype {
		// case scp_bioreactor:
		// 	if bio[ind].MustPause || bio[ind].MustPause {
		// 		i = tmax
		// 	}
		// }
		time.Sleep(1000 * time.Millisecond)
	}

	return true
}

func pop_first_sched(bioid string, remove bool) Scheditem {
	var ret Scheditem
	for k, s := range schedule {
		if s.Bioid == bioid {
			ret = Scheditem{s.Bioid, s.Seq, s.OrgCode, s.Volume}
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

	if !bio[ind].Heater {
		if !scp_turn_pump(scp_bioreactor, bioid, valvs, 1, true) {
			fmt.Println("ERROR SCP ADJUST PH: Falha ao abrir valvulas e ligar bomba", bioid, valvs)
		}
	}

	if bio[ind].MustPause || bio[ind].MustStop {
		return
	}
	phstr := ""
	if bio[ind].PH > ph {
		if !scp_turn_peris(scp_bioreactor, bioid, "P1", 1) {
			fmt.Println("ERROR SCP ADJUST PH: Falha ao ligar Peristaltica P1", bioid)
		} else {
			time.Sleep(scp_timephwait_down * time.Millisecond)
			phstr = "PH-"
			if !scp_turn_peris(scp_bioreactor, bioid, "P1", 0) {
				fmt.Println("ERROR SCP ADJUST PH: Falha ao desligar Peristaltica P1", bioid)
			}
		}
	} else {
		if !scp_turn_peris(scp_bioreactor, bioid, "P2", 1) {
			fmt.Println("ERROR SCP ADJUST PH: Falha ao ligar Peristaltica P2", bioid)
		} else {
			time.Sleep(scp_timephwait_up * time.Millisecond)
			phstr = "PH+"
			if !scp_turn_peris(scp_bioreactor, bioid, "P2", 0) {
				fmt.Println("ERROR SCP ADJUST PH: Falha ao desligar Peristaltica P2", bioid)
			}
		}
	}

	if len(phstr) > 0 {
		bio_add_message(bioid, "IAplicando "+phstr+" para corrigir PH", "")
	}
	for n := 0; n < 20; n++ {
		if bio[ind].MustPause || bio[ind].MustStop {
			break
		}
		time.Sleep(1 * time.Second)
	}

	if !bio[ind].Heater {
		if !scp_turn_pump(scp_bioreactor, bioid, valvs, 0, false) {
			fmt.Println("ERROR SCP ADJUST PH: Falha ao fechar valvulas e desligar bomba", bioid, valvs)
		}
	}

}

func scp_adjust_temperature(bioid string, temp float32, maxtime float64) {
	ind := get_bio_index(bioid)
	if ind < 0 {
		return
	}
	bio[ind].Temprunning = true
	fmt.Println("DEBUG SCP ADJUST TEMP: Ajustando Temperatura", bioid, bio[ind].Temperature, temp)
	valvs := []string{bioid + "/V4", bioid + "/V6"}
	t_start := time.Now()
	if bio[ind].Temperature < temp {
		if !scp_turn_pump(scp_bioreactor, bioid, valvs, 1, true) {
			fmt.Println("ERROR SCP ADJUST TEMP: Falha ao abrir valvulas e ligar bomba", bioid, valvs)
			bio[ind].Temprunning = false
			return
		}
		if !scp_turn_heater(bioid, temp, true) {
			fmt.Println("ERROR SCP ADJUST TEMP: Falha ao ligar aquecedor", bioid)
			scp_turn_pump(scp_bioreactor, bioid, valvs, 0, false)
			bio[ind].Temprunning = false
			return
		}
	} else {
		bio[ind].Temprunning = false
		return
	}
	for {
		t_elapsed := time.Since(t_start).Minutes()
		if t_elapsed >= maxtime || !bio[ind].Temprunning {
			break
		}
		if bio[ind].MustPause || bio[ind].MustStop {
			break
		}
		if biofabrica.Critical != scp_ready {
			break
		}
		if bio[ind].Temperature >= temp {
			fmt.Println("WARN SCP ADJUST TEMP: Temperatura >= limite em", bioid, bio[ind].Temperature, "/", temp)
			break
		}
		time.Sleep(2 * time.Second)
	}
	if !scp_turn_heater(bioid, temp, false) {
		fmt.Println("ERROR SCP ADJUST TEMP: Falha GRAVE ao desligar aquecedor", bioid)
		bio_add_message(bioid, "EATENÇÃO: Falha ao desligar resistência do Biorreator. Favor entrar em contato com o SAC", "")
	}
	if !scp_turn_pump(scp_bioreactor, bioid, valvs, 0, false) {
		fmt.Println("ERROR SCP ADJUST TEMP: Falha ao fechar valvulas e desligar bomba", bioid, valvs)
	}
	time.Sleep(10 * time.Second)
	bio[ind].Temprunning = false
	return
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
	ret := scp_turn_aero(bioid, false, scp_on, aero, false)
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
	if bio[ind].Timeleft[0] != 0 || bio[ind].Timeleft[1] != 0 {
		ttotal = float64(bio[ind].Timeleft[0]*60 + bio[ind].Timeleft[1])
	}
	fmt.Println("DEBUG SCP GROW BIO: ", bioid, org.Orgname, " tempo total ajustado para", ttotal)
	if devmode || testmode {
		board_add_message("A"+bioid+" Iniciando Cultivo em modo teste, duração de 5 min", "")
		ttotal = 5
	}
	time.Sleep(5 * time.Second)
	vol_start := bio[ind].Volume
	pday := -1
	var minph, maxph, worktemp_min, worktemp_max float64
	var err error
	var aero int
	aero_prev := -1
	worktemp_min = 28 // Valor Padrão para a temperatura de cultivo
	worktemp_max = 28

	temps := scp_splitparam(org.Temprange, "-")
	worktemp_min, err = strconv.ParseFloat(temps[0], 32)
	if err != nil {
		fmt.Println("ERROR GROW BIO: Valor de tempetura minimo para o cultivo INVALIDO:", bioid, temps, " - Assumindo 28 graus para mínimo e máximo")
		worktemp_min = 28
	} else {
		worktemp_max, err = strconv.ParseFloat(temps[1], 32)
		if err != nil {
			fmt.Println("ERROR GROW BIO: Valor de tempetura maximo para o cultivo INVALIDO:", bioid, temps, " - Assumindo 28 graus para mínimo e máximo")
			worktemp_min = 28
			worktemp_max = 28
		}
	}

	t_start := time.Now()
	t_start_ph := time.Now()

	// if control_foam {
	// 	scp_adjust_foam(bioid)
	// }

	lastph := float32(0)
	ntries_ph := 0
	ncontrol_foam := 0
	for {
		t_elapsed := time.Since(t_start).Minutes()
		fmt.Println("DEBUG SCP GROW BIO: ", bioid, " t_elapsed=", t_elapsed, " ttotal=", ttotal, " tempmin=", worktemp_min, " tempmax=", worktemp_max)
		if t_elapsed >= ttotal {
			fmt.Println("DEBUG SCP GROW BIO: Biorreator terminando grow por tempo atingido", bioid)
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
				if control_foam {
					scp_adjust_foam(bioid)
					ncontrol_foam++
				}
			}
			pday = t_day
		}
		if bio[ind].MustPause || bio[ind].MustStop {
			return false
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
			return false
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
				if math.Abs(float64(lastph)-float64(bio[ind].PH)) < 0.1 {
					ntries_ph++
					if ntries_ph > 5 {
						bio_add_message(bioid, "EVárias tentativas de ajustar PH foram feitas e não houve variação. Verifique níveis de PH+ , PH- , magueiras e sensor de PH", "")
						ntries_ph = 0
					}
				} else {
					ntries_ph = 0
				}
				lastph = bio[ind].PH
			}
			t_start_ph = time.Now()
		}
		if bio[ind].MustPause || bio[ind].MustStop {
			return false
		}

		// Início da mudança para suportar ranges de temperatura e os mesmo diferentes por organismo
		fmt.Println("DEBUG SCP GROW BIO: dados de temp", bioid, control_temp, bio[ind].Temprunning, "tempnow=", bio[ind].Temperature, "min=", worktemp_min, "max=", worktemp_max)
		if control_temp && !bio[ind].Temprunning {
			if bio[ind].Temperature < float32(worktemp_min) {
				fmt.Println("WARN SCP GROW BIO: Temperatura abaixo do mínimo (", worktemp_min, "), ajustando temperatura", bioid, bio[ind].Temperature, " para:", worktemp_max)
				bio[ind].Temprunning = true
				go scp_adjust_temperature(bioid, float32(worktemp_max), ttotal)
			} else if bio[ind].Temperature > float32(worktemp_max) {
				fmt.Println("ERROR SCP GROW BIO: Temperatura acima do máximo no", bioid, bio[ind].Temperature, " máximo:", worktemp_max)
				bio_add_message(bioid, "AAVISO: tempetura está acima do máximo ideal para o cultivo, favor verificar", "")
			}
		}

		if bio[ind].MustPause || bio[ind].MustStop {
			return false
		}
		time.Sleep(scp_timegrowwait * time.Millisecond)
	}
	bio[ind].Temprunning = false
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
	if !scp_turn_pump(devtype, main_id, valvs, 1, true) {
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
	if !scp_turn_pump(devtype, main_id, valvs, 0, false) {
		fmt.Println("ERROR SCP CIRCULATE: Nao foi possivel desligar circulacao em ", main_id)
	}
	switch devtype {
	case scp_bioreactor:
		if bio[ind].Status != bio_pause {
			bio[ind].Status = bio[ind].LastStatus
		}
	case scp_ibc:
		if ibc[ind].Status != bio_pause {
			ibc[ind].Status = ibc[ind].LastStatus
		}
	}
}

func scp_fullstop_bio(bioid string, check_status bool, clearall bool) bool {
	ind := get_bio_index(bioid)
	if ind < 0 {
		fmt.Println("ERROR FULLSTOP BIO: Biorreator nao encontrado", bioid)
		return false
	}
	if check_status {
		if bio[ind].Status != bio_ready && bio[ind].Status != bio_empty {
			bio_add_message(bioid, "EEquipamento deve estar PRONTO ou VAZIO para Parada TOTAL", "")
			return false
		}
		bio_add_message(bioid, "AParada TOTAL solicitada", "")
	}
	ret := true
	ret_heater := scp_turn_heater(bioid, 0, false)
	if !ret_heater {
		fmt.Println("SCP FULLSTOP BIO: Falha ao desligar resistência do", bioid)
	}
	ret = ret && ret_heater
	ret_peris := scp_turn_peris(scp_bioreactor, bioid, "P1", 0)
	ret_peris = ret_peris && scp_turn_peris(scp_bioreactor, bioid, "P2", 0)
	ret_peris = ret_peris && scp_turn_peris(scp_bioreactor, bioid, "P3", 0)
	ret_peris = ret_peris && scp_turn_peris(scp_bioreactor, bioid, "P4", 0)
	ret_peris = ret_peris && scp_turn_peris(scp_bioreactor, bioid, "P5", 0)
	if !ret_peris {
		fmt.Println("SCP FULLSTOP BIO: Falha ao desligar peristalticas do", bioid)
	}
	ret = ret && ret_peris
	ret_aero := scp_turn_aero(bioid, true, 0, 0, false)
	if !ret_aero {
		fmt.Println("SCP FULLSTOP BIO: Falha ao desligar aerador do", bioid)
	}
	ret = ret && ret_aero
	valvs := []string{bioid + "/V3", bioid + "/V4", bioid + "/V5", bioid + "/V6", bioid + "/V7", bioid + "/V8"}
	ret_pump := scp_turn_pump(scp_bioreactor, bioid, valvs, 0, false)
	if !ret_pump {
		fmt.Println("SCP FULLSTOP BIO: Falha ao desligar bomba e valvulas do", bioid)
	}
	ret = ret && ret_pump
	if clearall {
		bio[ind].Queue = []string{}
		bio[ind].RedoQueue = []string{}
		bio[ind].MustOffQueue = []string{}
		bio[ind].MustPause = false
		bio[ind].MustStop = false
		bio[ind].Timetotal[0] = 0
		bio[ind].Timetotal[1] = 0
		bio[ind].Timeleft[0] = 0
		bio[ind].Timeleft[1] = 0
		bio[ind].Step[0] = 0
		bio[ind].Step[1] = 0
		if bio[ind].Volume == 0 {
			bio[ind].MainStatus = mainstatus_empty
		} else {
			bio[ind].MainStatus = mainstatus_org
		}
	}

	return ret
}

func scp_fullstop_ibc(ibcid string, check_status bool, clearall bool) bool {
	ind := get_ibc_index(ibcid)
	if ind < 0 {
		fmt.Println("ERROR FULLSTOP IBC: IBC nao encontrado", ibcid)
		return false
	}
	if check_status {
		if ibc[ind].Status != bio_ready && ibc[ind].Status != bio_empty {
			board_add_message("EEquipamento "+ibcid+" deve estar PRONTO ou VAZIO para Parada TOTAL", "")
			return false
		}
		board_add_message("AParada TOTAL solicitada para "+ibcid, "")
	}
	valvs := []string{ibcid + "/V1", ibcid + "/V2", ibcid + "/V3", ibcid + "/V4"}
	ret := scp_turn_pump(scp_ibc, ibcid, valvs, 0, false)
	if !ret {
		fmt.Println("SCP FULLSTOP IBC: Falha ao desligar bomba e valvulas do", ibcid)
	}
	if clearall {
		ibc[ind].Queue = []string{}
		ibc[ind].RedoQueue = []string{}
		ibc[ind].MustOffQueue = []string{}
		ibc[ind].MustPause = false
		ibc[ind].MustStop = false
		ibc[ind].Timetotal[0] = 0
		ibc[ind].Timetotal[1] = 0
		ibc[ind].Step[0] = 0
		ibc[ind].Step[1] = 0
		if ibc[ind].Volume == 0 {
			ibc[ind].MainStatus = mainstatus_empty
		} else {
			ibc[ind].MainStatus = mainstatus_org
		}
	}

	return ret
}

func scp_fullstop_totem(totemid string) bool {
	ind := get_totem_index(totemid)
	if ind < 0 {
		fmt.Println("ERROR FULLSTOP IBC: TOTEM nao encontrado", totemid)
		return false
	}
	board_add_message("AParada TOTAL solicitada para "+totemid, "")
	ret := true
	ret_peris := scp_turn_peris(scp_totem, totemid, "P1", 0)
	ret_peris = ret_peris && scp_turn_peris(scp_totem, totemid, "P2", 0)
	if !ret_peris {
		fmt.Println("SCP FULLSTOP TOTEM: Falha ao desligar peristalticas do", totemid)
	}
	ret = ret && ret_peris
	valvs := []string{totemid + "/V1", totemid + "/V2"}
	ret_pump := scp_turn_pump(scp_totem, totemid, valvs, 0, false)
	if !ret_pump {
		fmt.Println("SCP FULLSTOP TOTEM: Falha ao desligar bomba e valvulas do", totemid)
	}
	ret = ret && ret_pump
	return ret
}

func scp_fullstop_biofabrica() bool {
	board_add_message("AParada TOTAL solicitada para válvulas e bomba da Biofábrica", "")
	valvs := []string{}
	for _, e := range biofabrica_cfg {
		if strings.Contains(e.DeviceID, "VBF") {
			dev := "BIOFABRICA/" + e.DeviceID
			valvs = append(valvs, dev)
		}
	}
	ret_pump := scp_turn_pump(scp_biofabrica, "BF01", valvs, 0, false)
	if !ret_pump {
		fmt.Println("SCP FULLSTOP BIOFABRICA: Falha ao desligar bomba e valvulas da Biofabrica")
	}
	return ret_pump
}

func scp_fullstop_device(devid string, devtype string, check_status bool, clearall bool) bool {
	switch devtype {
	case scp_bioreactor:
		return scp_fullstop_bio(devid, check_status, clearall)
	case scp_ibc:
		return scp_fullstop_ibc(devid, check_status, clearall)
	case scp_totem:
		return scp_fullstop_totem(devid)
	case scp_biofabrica:
		return scp_fullstop_biofabrica()
	}
	fmt.Println("ERROR FULLSTOP DEVICE: Tipo de Dispositivo invalido", devtype, devid)
	return false
}

func scp_run_job_bio(bioid string, job string) bool {
	if devmode {
		fmt.Println("DEBUG SCP RUN JOB: SIMULANDO EXECUCAO", bioid, job)
	} else {
		fmt.Println("DEBUG SCP RUN JOB: EXECUTANDO", bioid, job)
	}
	if devmode || testmode {
		bio_add_message(bioid, "C"+job, "")
	}
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
	case scp_job_msg:
		if len(subpars) > 1 {
			msg := subpars[0]
			if len(msg) > 0 {
				if msg[0] == '!' {
					bio_add_message(bioid, "A"+msg[1:], "")
				} else {
					bio_add_message(bioid, "I"+msg, "")
				}
			}
		}
	case scp_job_org:
		var orgcode string
		if len(subpars) > 0 {
			orgcode = subpars[0]
			if len(organs[orgcode].Orgname) > 0 {
				bio[ind].OrgCode = subpars[0]
				bio[ind].Organism = organs[orgcode].Orgname
				bio[ind].Timetotal = [2]int{organs[orgcode].Timetotal, 0}
				bio[ind].Timeleft = [2]int{organs[orgcode].Timetotal, 0}
				bio[ind].MainStatus = mainstatus_grow
				scp_update_screen_times(bioid)
			} else {
				fmt.Println("ERROR SCP RUN JOB: Organismo nao existe", params)
				return false
			}
		} else {
			fmt.Println("ERROR SCP RUN JOB: Falta parametros em", scp_job_org, params)
			return false
		}
		board_add_message("CIniciando Cultivo "+organs[orgcode].Orgname+" no "+bioid, "")
		bio_add_message(bioid, "CIniciando Cultivo de "+organs[orgcode].Orgname, "")

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
				scp_update_screen_steps(bioid)
			case scp_par_maxstep:
				biomaxstep_str := subpars[1]
				biomaxstep, _ := strconv.Atoi(biomaxstep_str)
				bio[ind].Step[1] = biomaxstep
				scp_update_screen_steps(bioid)
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
				bio[ind].MainStatus = mainstatus_grow
				ret := scp_grow_bio(bioid)
				return ret

			case scp_par_cip:
				bio[ind].ShowVol = false
				qini := []string{bio[ind].Queue[0]}
				qini = append(qini, cipbio...)
				bio[ind].Queue = append(qini, bio[ind].Queue[1:]...)
				fmt.Println("\n\nTRUQUE CIP:", bio[ind].Queue)
				board_add_message("IExecutando CIP no "+bioid, "")
				bio_add_message(bioid, "IExecutando CIP", "")
				bio[ind].MainStatus = mainstatus_cip
				return true

			case scp_par_withdraw:
				bio[ind].Withdraw = bio[ind].Volume
				if len(subpars) > 2 {
					outid := subpars[1]
					bio[ind].OutID = outid
				}
				// board_add_message("IDesenvase Automático do biorreator " + bioid + " para " + bio[ind].OutID)
				if scp_run_withdraw(scp_bioreactor, bioid, false, true) < 0 {
					fmt.Println("ERROR SCP RUN JOB BIO: Falha ao fazer o desenvase do BIO", bio[ind].BioreactorID)
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
			// scraddr := bio_cfg[bioid].Screenaddr
			// var cmd1 string = ""
			var msgask string = ""
			// scrmain := fmt.Sprintf("CMD/%s/PUT/S200,1/END", scraddr)
			switch msg {
			case scp_msg_cloro:
				// cmd1 = fmt.Sprintf("CMD/%s/PUT/S400,2/END", scraddr)
				msgask = "CLORO"
			case scp_msg_meio:
				// cmd1 = fmt.Sprintf("CMD/%s/PUT/S400,3/END", scraddr)
				msgask = "MEIO"
			case scp_msg_inoculo:
				// cmd1 = fmt.Sprintf("CMD/%s/PUT/S400,1/END", scraddr)
				msgask = "INOCULO"
			default:
				fmt.Println("ERROR SCP RUN JOB:", bioid, " ASK invalido", subpars)
				return false
			}
			// ret1 := scp_sendmsg_orch(cmd1)
			// fmt.Println("DEBUG SCP RUN JOB:: CMD =", cmd1, "\tRET =", ret1)
			// if !strings.Contains(ret1, scp_ack) && !devmode {
			// 	fmt.Println("ERROR SCP RUN JOB:", bioid, " ERROR ao enviar PUT screen", scraddr, ret1)
			// 	return false
			// }
			// cmd2 := fmt.Sprintf("CMD/%s/GET/S451/END", scraddr)
			waitlist_add_message("ABiorreator "+bioid+" aguardando "+msgask, bioid+"ASKPROD")
			bio_add_message(bioid, "APor favor insira "+msgask+" e pressione PROSSEGUIR", "ASKPROD")
			bio[ind].Continue = false
			t_start := time.Now()
			for {
				if bio[ind].Continue == true {
					break
				}
				// ret2 := scp_sendmsg_orch(cmd2)
				// // fmt.Println("DEBUG SCP RUN JOB:: CMD =", cmd2, "\tRET =", ret2)
				// if !strings.Contains(ret2, scp_ack) && !devmode {
				// 	fmt.Println("ERROR SCP RUN JOB:", bioid, " ERRO ao envirar GET screen", scraddr, ret2)
				// 	scp_sendmsg_orch(scrmain)
				// 	return false
				// }
				// data := scp_splitparam(ret2, "/")
				// if len(data) > 1 {
				// 	if data[1] == "1" {
				// 		break
				// 	}
				// }
				if bio[ind].MustPause || bio[ind].MustStop {
					return false
				}
				t_elapsed := time.Since(t_start).Seconds()
				if t_elapsed > scp_timeoutdefault {
					fmt.Println("DEBUG SCP RUN JOB: Tempo maximo de ASK esgotado", bioid, t_elapsed, scp_maxtimewithdraw)
					if !devmode {
						// scp_sendmsg_orch(scrmain)
						return testmode
					}
					break
				}
				time.Sleep(1000 * time.Millisecond)
			}
			waitlist_del_message(bioid + "ASKPROD")
			bio_del_message(bioid, "ASKPROD")
			// scp_sendmsg_orch(scrmain)
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
		board_add_message("CProcesso concluído no "+bioid, "")
		bio[ind].UndoQueue = []string{}
		bio[ind].RedoQueue = []string{}
		bio[ind].MustOffQueue = []string{}
		bio[ind].Step = [2]int{0, 0}
		bio[ind].Timetotal = [2]int{0, 0}
		bio[ind].Timeleft = [2]int{0, 0}
		if bio[ind].MainStatus == mainstatus_cip {
			if bio[ind].Volume == 0 {
				bio[ind].MainStatus = mainstatus_empty
			} else {
				bio[ind].MainStatus = mainstatus_org
			}
		} else if bio[ind].MainStatus == mainstatus_grow {
			bio[ind].MainStatus = mainstatus_org
		} else {
			bio[ind].MainStatus = mainstatus_org
		}

		// scp_update_screen_times(bioid)
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

				time_max := scp_maxwaitvolume * 60
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
				if devmode || testmode {
					time_max = scp_timeoutdefault / 2
				}
				t_start := time.Now()
				t_last_volchange := time.Now()
				vol_bio_last := uint64(bio[ind].Volume)
				for {
					vol_now := uint64(bio[ind].Volume)
					t_elapsed := time.Since(t_start).Seconds()
					if vol_now >= uint64(bio_cfg[bio[ind].BioreactorID].Maxvolume) {
						break
					}
					if vol_now >= vol_max {
						if !par_time {
							break
						} else if t_elapsed >= float64(time_min) {
							break
						}
					}
					if bio[ind].MustPause || bio[ind].MustStop {
						return false
					}
					ind_totem := get_totem_index("TOTEM01")
					if ind_totem >= 0 {
						if totem[ind_totem].Status == bio_error || totem[ind_totem].Status == bio_nonexist {
							board_add_message("E"+bio[ind].BioreactorID+" volume não atingido por falha no TOTEM01", "")
							bio_add_message(bio[ind].BioreactorID, "EVolume não atingido por falha no TOTEM01", "")
							go pause_device(scp_bioreactor, bio[ind].BioreactorID, true)
							return false
						}
					}
					if vol_now == vol_bio_last {
						t_elapsed_volchage := time.Since(t_last_volchange).Seconds()
						if t_elapsed_volchage > 25 {
							// abort_due_novolchange = true
							fmt.Println("DEBUG SCP RUN JOB: WAIT VOLUME abortado por volume nao variar em 25 seg", bio[ind].BioreactorID)
							board_add_message("E"+bio[ind].BioreactorID+" volume não variou em 25s. Favor verificar equipamentos", "")
							bio_add_message(bio[ind].BioreactorID, "EVolume não variou em 25s. Favor verificar equipamentos", "")
							go pause_device(scp_bioreactor, bio[ind].BioreactorID, true)
							return false
						}
					} else {
						vol_bio_last = vol_now
						t_last_volchange = time.Now()
					}

					if t_elapsed > float64(time_max) {
						fmt.Println("DEBUG SCP RUN JOB: Tempo maximo de WAIT VOLUME esgotado", bioid, t_elapsed, scp_maxtimewithdraw)
						if !devmode && !par_time {
							return testmode
						}
						break
					}
					time.Sleep(scp_refreshwait * 2 * time.Millisecond)
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
				if !scp_turn_aero(bioid, true, 1, perc_int, false) {
					fmt.Println("ERROR SCP RUN JOB: ERROR ao ligar aerador em", bioid)
					return false
				}
			case scp_dev_pump:
				valvs := []string{}
				for k := 1; k < len(subpars) && subpars[k] != "END"; k++ {
					v := bioid + "/" + subpars[k]
					valvs = append(valvs, v)
				}
				if !scp_turn_pump(scp_bioreactor, bioid, valvs, 1, true) {
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
				if err != nil || temp_int <= 0 {
					checkErr(err)
					fmt.Println("ERROR SCP RUN JOB: Parametro de temperatura invalido", bioid, temp_str, subpars)
					return false
				}
				if bio[ind].Volume == 0 {
					fmt.Println("ERROR SCP RUN JOB: Nao é possivel ligar resistencia com Biorreator VAZIO", bioid, "Volume=", bio[ind].Volume, temp_str, subpars)
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
				// fmt.Println("DEBUG", vpath)
				if !scp_turn_pump(scp_totem, totem, vpath, 1, true) {
					waitlist_add_message("ABiorreator "+bioid+" aguardando liberação da Linha", bioid+"ONWATERBUSY")
					bio_add_message(bioid, "ABiorreator aguardando liberação da Linha", "ONWATERBUSY")
					fmt.Println("ERROR SCP RUN JOB: ERROR ao ligar bomba em", bioid, valvs)
					return false
				} else {
					waitlist_del_message(bioid + "ONWATERBUSY")
					bio_del_message(bioid, "ONWATERBUSY")
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
			case scp_dev_all:
				if !scp_fullstop_device(bioid, scp_bioreactor, false, false) {
					fmt.Println("ERROR SCP RUN JOB: ERROR ao desligar todos os dispositivos em", bioid)
					return true
				}
			case scp_dev_aero:
				if !scp_turn_aero(bioid, true, 0, 0, false) {
					fmt.Println("ERROR SCP RUN JOB: ERROR ao desligar aerador em", bioid)
					return false
				}
			case scp_dev_pump:
				valvs := []string{}
				for k := 1; k < len(subpars) && subpars[k] != "END"; k++ {
					v := bioid + "/" + subpars[k]
					valvs = append(valvs, v)
				}
				if !scp_turn_pump(scp_bioreactor, bioid, valvs, 0, false) {
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
				if !scp_turn_pump(scp_totem, totem, vpath, 0, false) {
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
	// if ibc[ind].MustPause || ibc[ind].MustPause {
	// 	return false
	// }
	params := scp_splitparam(job, "/")
	subpars := []string{}
	if len(params) > 1 {
		subpars = scp_splitparam(params[1], ",")
	}
	switch params[0] {
	case scp_job_msg:
		if len(subpars) > 1 {
			msg := subpars[0]
			if len(msg) > 0 {
				if msg[0] == '!' {
					bio_add_message(ibcid, "A"+msg[1:], "")
				} else {
					bio_add_message(ibcid, "I"+msg, "")
				}
			}
		}
	case scp_job_org:
		var orgcode string
		if len(subpars) > 0 {
			orgcode = subpars[0]
			if len(organs[orgcode].Orgname) > 0 {
				ibc[ind].OrgCode = subpars[0]
				ibc[ind].Organism = organs[orgcode].Orgname
				ibc[ind].Timetotal = [2]int{0, 0}
				ibc[ind].MainStatus = mainstatus_org
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
			case scp_par_volume:
				vol_str := subpars[1]
				vol_int, err := strconv.Atoi(vol_str)
				if err != nil {
					checkErr(err)
				} else {
					ibc[ind].VolInOut = float64(vol_int)
					ibc[ind].Volume = uint32(vol_int)
				}
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
				board_add_message("IExecutando CIP no "+ibcid, "")
				ibc[ind].MainStatus = mainstatus_cip
				return true

			case scp_par_withdraw:
				ibc[ind].Withdraw = ibc[ind].Volume
				if len(subpars) > 2 {
					outid := subpars[1]
					ibc[ind].OutID = outid
				}
				fmt.Println("DEBUG SCP RUN JOB IBC: Run Withdraw ", ibcid)
				// board_add_message("IDesenvase Automático do "+ibcid+" para "+ibc[ind].OutID, ibcid+"WDOUT")
				if scp_run_withdraw(scp_ibc, ibcid, false, true) < 0 {
					fmt.Println("ERROR SCP RUN JOB IBC: Falha ao fazer o desenvase do IBC", ibc[ind].IBCID)
					return false
				} else {
					fmt.Println("DEBUG SCP RUN JOB IBC: Run Withdraw SEM ERROS ", ibcid)
					ibc[ind].Organism = ""
					ibc[ind].Withdraw = 0
					// ibc[ind].Volume = 0 // verificar
				}
				// board_del_message(ibcid + "WDOUT")
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
		board_add_message("CProcesso concluído no "+ibcid, "")
		ibc[ind].UndoQueue = []string{}
		ibc[ind].RedoQueue = []string{}
		ibc[ind].MustOffQueue = []string{}
		ibc[ind].Step = [2]int{0, 0}
		ibc[ind].ShowVol = true
		if ibc[ind].MainStatus == mainstatus_cip {
			if ibc[ind].Volume == 0 {
				ibc[ind].MainStatus = mainstatus_empty
			} else {
				ibc[ind].MainStatus = mainstatus_org
			}
		} else {
			ibc[ind].MainStatus = mainstatus_org
		}

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
					if vol_now >= uint64(ibc_cfg[ibc[ind].IBCID].Maxvolume) {
						break
					}
					if vol_now >= vol_max && t_elapsed >= float64(time_min) {
						break
					}
					ind_totem := get_totem_index("TOTEM02")
					if ind_totem >= 0 {
						if totem[ind_totem].Status == bio_error || totem[ind_totem].Status == bio_nonexist {
							board_add_message("E"+ibc[ind].IBCID+" volume não atingido por falha no TOTEM02", "")
							go pause_device(scp_ibc, ibc[ind].IBCID, true)
							return false
						}
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
				if !scp_turn_pump(scp_ibc, ibcid, valvs, 1, true) {
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
					// fmt.Println("npath=", pathstr)
					vpath := scp_splitparam(pathstr, ",")
					perisvalv := totem_str + "/V2"
					n := len(vpath)
					vpath = append(vpath[:n-1], perisvalv)
					vpath = append(vpath, "END")
					fmt.Println("DEBUG SCP JOB IBC: ON DEV PERIS", ibcid, vpath)
					if test_path(vpath, 0) {
						if set_valvs_value(vpath, 1, true) < 0 {
							fmt.Println("ERROR SCP RUN JOB: ERRO ao abrir valvulas no path ", vpath)
							return false
						}
					} else {
						fmt.Println("ERROR SCP RUN JOB: ERRO nas valvulas no path ", vpath)
						return false
					}
					time.Sleep(scp_timewaitvalvs * time.Millisecond)
					// tmax := scp_timewaitvalvs / 100
					// for i := 0; i < tmax; i++ {
					// switch devtype {
					// case scp_bioreactor:
					// 	if bio[ind].MustPause || bio[ind].MustPause {
					// 		i = tmax
					// 	}
					// }
					// 	time.Sleep(100 * time.Millisecond)
					// }
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
				unlock_par := false
				if len(subpars) > 3 {
					unlock_par = subpars[2] == scp_par_unlock
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
				// fmt.Println("DEBUG", vpath)
				if !scp_turn_pump(scp_totem, totem, vpath, 1, !unlock_par) {
					waitlist_add_message("A"+ibcid+" aguardando liberação da Linha", ibcid+"ONWATERBUSY")
					fmt.Println("ERROR SCP RUN JOB: ERROR ao ligar bomba em", ibcid, valvs)
					return false
				} else {
					waitlist_del_message(ibcid + "ONWATERBUSY")
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
			case scp_dev_all:
				if !scp_fullstop_device(ibcid, scp_ibc, false, false) {
					fmt.Println("ERROR SCP RUN JOB: ERROR ao desligar todos os dispositivos em", ibcid)
					return true
				}
			case scp_dev_pump:
				valvs := []string{}
				for k := 1; k < len(subpars) && subpars[k] != "END"; k++ {
					v := ibcid + "/" + subpars[k]
					valvs = append(valvs, v)
				}
				if !scp_turn_pump(scp_ibc, ibcid, valvs, 0, false) {
					fmt.Println("ERROR SCP RUN JOB: ERROR ao desligar bomba em", ibcid, valvs)
					return false
				}

			case scp_dev_peris:
				if len(subpars) > 3 {
					peris_str := subpars[1]
					totem_str := subpars[2]
					lock_par := ""
					lock_valv := ""
					if len(subpars) > 3 {
						lock_par = subpars[3]
						if lock_par == scp_par_lock {
							if len(subpars) > 4 {
								lock_valv = "BIOFABRICA/" + subpars[4]
								// if lock_valv
							} else {
								lock_par = ""
								fmt.Println("ERROR SCP RUN JOB: OFF Faltou valvula no LOCK", ibcid, subpars)
							}
						} else {
							fmt.Println("ERROR SCP RUN JOB: OFF Parametro LOCK esperado", ibcid, subpars)
						}
					}
					if len(lock_valv) > 0 {
						fmt.Println("DEBUG SCP RUN JOB: Executando lock na valvula", ibcid, lock_valv)
						if set_valvs_value([]string{lock_valv}, 3, false) < 0 {
							fmt.Println("ERROR SCP RUN JOB: Não foi possivel dar lock na valvula", ibcid, lock_valv)
						}
					}
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
					pathstr_new := strings.Replace(pathstr, lock_valv+",", "", -1)
					fmt.Println("DEBUG SCP RUN JOB: OFF PERIS npath=", pathstr, " newpath=", pathstr_new)
					vpath := scp_splitparam(pathstr_new, ",")
					perisvalv := totem_str + "/V2"
					n := len(vpath)
					vpath = append(vpath[:n-1], perisvalv)
					vpath = append(vpath, "END")
					// fmt.Println("DEBUG", vpath)
					if test_path(vpath, 1) {
						if set_valvs_value(vpath, 0, true) < 0 {
							fmt.Println("ERROR SCP RUN JOB: ERRO ao fechar valvulas no path ", vpath)
							return false
						}
					} else {
						fmt.Println("ERROR SCP RUN JOB: ERRO nas valvulas no path ", vpath)
						return false
					}
					time.Sleep(scp_timewaitvalvs * time.Millisecond)
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
				if !scp_turn_pump(scp_totem, totem, vpath, 0, false) {
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
		if len(bio[ind].Queue) > 0 && (devmode || testmode) {
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
							if !strings.Contains(onoff, "IGNORE") {
								bio[ind].UndoQueue = append(bio[ind].UndoQueue, onoff)
							}
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
		time.Sleep(time.Duration(rand.Intn(7)) * scp_schedwait * time.Millisecond)
	}
	fmt.Println("DEBUG RUN BIO: Terminando para", bioid)
}

func scp_run_ibc(ibcid string) {
	fmt.Println("STARTANDO RUN", ibcid)
	ind := get_ibc_index(ibcid)
	if ind < 0 {
		fmt.Println("ERROR SCP RUN BIO: Biorreator nao existe", ibcid)
		return
	}
	for ibc[ind].Status != bio_die {
		if len(ibc[ind].Queue) > 0 && (devmode || testmode) {
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
							if !strings.Contains(onoff, "IGNORE") {
								ibc[ind].UndoQueue = append(ibc[ind].UndoQueue, onoff)
							}
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
		time.Sleep(time.Duration(rand.Intn(7)) * scp_schedwait * time.Millisecond)
	}
	fmt.Println("DEBUG RUN BIO: Terminando para", ibcid)
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
				fmt.Println("DEBUG SCP CLOCK: TOTAL LEFT BIO ", b.BioreactorID, totalleft)
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
				fmt.Println("DEBUG SCP CLOCK: TOTAL LEFT IBC ", b.IBCID, totalleft)
				if totalleft > 0 {
					totalleft -= int(t_elapsed)
					ibc[ind].Timetotal[0] = int(totalleft / 60)
					ibc[ind].Timetotal[1] = int(totalleft % 60)
				}
			}
		}

		t_start = time.Now()
		time.Sleep(scp_clockwait * time.Second)

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
	fmt.Println("DEBUG SCHEDULER: Iniciando Scheduler")
	if !devsrunning {
		scp_run_devs()
		go scp_clock()
		devsrunning = true
	}
	for schedrunning == true {
		for k, b := range bio {
			// fmt.Println(k, " bio =", b)
			r := pop_first_sched(b.BioreactorID, false)
			// fmt.Println("DEBUG SCHEDULER: Sched ")
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
							if s.Volume == 1000 {
								bio[k].Queue = append(orginfo, recipe_1000...)
							} else {
								bio[k].Queue = append(orginfo, recipe_2000...)
							}
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
							ibc[k].Queue = append(orginfo, recipe_2000...)
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
		volume := 0
		if len(item) >= 4 {
			volume, _ = strconv.Atoi(item[3])
		}
		ind_bio := get_bio_index(main_id)
		ind_ibc := get_ibc_index(main_id)
		if ind_bio >= 0 || ind_ibc >= 0 {
			schedule = append(schedule, Scheditem{main_id, bioseq, orgcode, volume})
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
			if bio[ind].Status == bio_error {
				bio[ind].UndoStatus = bio[ind].LastStatus
				bio[ind].LastStatus = bio_pause
			} else {
				bio[ind].LastStatus = bio[ind].Status
				bio[ind].Status = bio_pause
			}
			for _, j := range bio[ind].MustOffQueue {
				if !isin(bio[ind].UndoQueue, j) {
					bio[ind].UndoQueue = append([]string{j}, bio[ind].UndoQueue...)
				}
			}
			// bio[ind].UndoQueue = append(bio[ind].MustOffQueue, bio[ind].UndoQueue...)
			bio[ind].MustPause = true
			// bio[ind].Status = bio_pause    Modificado para o caso de biorreator com ERRO
			bio[ind].Withdraw = 0 // VALIDAR
			if bio[ind].Heater {
				scp_turn_heater(bio[ind].BioreactorID, 0, false)
			}
			if !bio[ind].MustStop {
				waitlist_add_message("ABiorreator "+main_id+" pausado", main_id+"PAUSE")
				bio_add_message(main_id, "ABiorreator pausado", "PAUSE")
			} else {
				bio_add_message(main_id, "ABiorreator sendo pausado para depois ser interrompido", "")
			}

		} else if !pause && (bio[ind].Status == bio_pause || bio[ind].MustPause) {
			fmt.Println("DEBUG PAUSE DEVICE: Retomando Biorreator", main_id)
			bio[ind].Queue = append(bio[ind].RedoQueue, bio[ind].Queue...)
			// fmt.Println("****** LAST STATUS no PAUSE", bio[ind].LastStatus)
			if bio[ind].LastStatus == bio_pause && bio[ind].Status == bio_pause { // CHECAR coloquei bio[ind].Status == bio_pause
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
				// board_add_message("APausa no Biorreator "+main_id+" liberada", "")
				// bio_add_message(main_id, "APausa no Biorreator liberada", "")
				waitlist_del_message(main_id + "PAUSE")
				bio_del_message(main_id, "PAUSE")
			}
			// if !schedrunning {
			// 	go scp_scheduler()
			// }
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
			ibc[ind].Withdraw = 0 // VALIDAR
			if !ibc[ind].MustStop {
				waitlist_add_message("AIBC "+main_id+" pausado", main_id+"PAUSE")
			}

		} else if !pause && (ibc[ind].Status == bio_pause || ibc[ind].MustPause) {
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
				// board_add_message("APausa no IBC "+main_id+" liberada", "")
				waitlist_del_message(main_id + "PAUSE")
			}
			// if !schedrunning {
			// 	go scp_scheduler()
			// }
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
		if bio[ind].MustStop {
			bio_add_message(main_id, "EBiorreator já sendo interrompido. Aguarde", "")
			return false
		}
		fmt.Println("\n\nDEBUG STOP: Executando STOP para", main_id)
		bio[ind].Withdraw = 0
		bio[ind].MustStop = true
		bio_add_message(main_id, "ABiorreator sendo Interrompido. Aguarde", "")
		if bio[ind].Status != bio_empty || true { // corrigir
			bio[ind].MustStop = true
			pause_device(devtype, main_id, true)
			// t_start =
			for {
				time.Sleep(3000 * time.Millisecond)
				if len(bio[ind].UndoQueue) == 0 {
					break
				}
			} //
			if bio[ind].Heater {
				scp_turn_heater(bio[ind].BioreactorID, 0, false)
			}
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
			if bio[ind].Volume == 0 {
				bio[ind].MainStatus = mainstatus_empty
			} else {
				bio[ind].MainStatus = mainstatus_org
			}
			for { // LIMPA FILA de TAREFAS --- MUDAR QUANDO FOR PERMITIR TAREFAS FUTURAS
				q := pop_first_sched(bio[ind].BioreactorID, true)
				if len(q.Bioid) == 0 {
					break
				}
			}

			// if len(q.Bioid) == 0 { // Verificar depois
			if bio[ind].Volume == 0 {
				bio[ind].Status = bio_empty
			} else {
				bio[ind].Status = bio_ready
			}
			bio[ind].MustPause = false
			// }
			bio[ind].ShowVol = true
			waitlist_del_message(bio[ind].BioreactorID)
		}

	case scp_ibc:
		ind := get_ibc_index(main_id)
		if ind < 0 {
			fmt.Println("ERROR STOP: IBC nao existe", main_id)
			return false
		}
		if ibc[ind].MustStop {
			board_add_message("E"+main_id+" já sendo interrompido, aguarde", "")
			return false
		}
		fmt.Println("\n\nDEBUG STOP: Executando STOP para", main_id)
		ibc[ind].Withdraw = 0
		board_add_message("A"+main_id+" interrompido", "")
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
			if ibc[ind].Volume == 0 {
				ibc[ind].MainStatus = mainstatus_empty
			} else {
				ibc[ind].MainStatus = mainstatus_org
			}

			for { // LIMPA FILA de TAREFAS --- MUDAR QUANDO FOR PERMITIR TAREFAS FUTURAS
				q := pop_first_sched(ibc[ind].IBCID, true)
				if len(q.Bioid) == 0 {
					break
				}
			}

			// q := pop_first_sched(ibc[ind].IBCID, false)
			// if len(q.Bioid) == 0 {
			if ibc[ind].Volume == 0 {
				ibc[ind].Status = bio_empty
			} else {
				ibc[ind].Status = bio_ready
			}
			ibc[ind].MustPause = false
			// }
			ibc[ind].ShowVol = true
			waitlist_del_message(ibc[ind].IBCID)
		}
	}
	save_all_data(data_filename)
	return true
}

func scp_restart_services() {
	// fmt.Println("Reestartando Servico ORCH")

	biofabrica.Critical = scp_stopall
	cmdpath, _ := filepath.Abs("/usr/bin/systemctl")
	// cmd := exec.Command(cmdpath, "restart", "scp_orch")
	// cmd.Dir = "/usr/bin"
	// output, err := cmd.CombinedOutput()
	// if len(output) > 0 {
	// 	fmt.Println("OUPUT", string(output))
	// }
	// if err != nil {
	// 	checkErr(err)
	// 	fmt.Println("Falha ao Restartar ORCH")
	// 	board_add_message("EFalha ao reiniciar Orquestrador", "")
	// 	return
	// }
	// board_add_message("EOrquestrador reiniciado", "")
	time.Sleep(30 * time.Second)
	fmt.Println("Reestartando Servico BACKEND")
	cmd := exec.Command(cmdpath, "restart", "scp_back")
	cmd.Dir = "/usr/bin"
	output, err := cmd.CombinedOutput()
	if len(output) > 0 {
		fmt.Println("OUPUT", string(output))
	}
	if err != nil {
		checkErr(err)
		fmt.Println("Falha ao Restartar BACK")
		board_add_message("EFalha ao reiniciar Backend", "")
		return
	}
	board_add_message("EBackend reiniciado", "")
	time.Sleep(30 * time.Second)
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
		board_add_message("EFalha ao reiniciar MASTER", "")
		return
	}
}

func scp_run_manydraw_out(data string, dest string) {
	var vol_ini map[string]float64

	vol_ini = make(map[string]float64)
	ibc_par := scp_splitparam(data, ",")

	for _, b := range ibc_par {
		d := scp_splitparam(b, "=")
		i := get_ibc_index(d[0])
		if i >= 0 && len(d) >= 2 {
			vol_ini[d[0]] = ibc[i].VolInOut
		} else {
			fmt.Println("ERROR RUN MANYDRAW OUT: Parametro invalido par=", b)
		}
	}

	n_tries := 0
	for {
		no_ok := 0
		for _, b := range ibc_par {
			d := scp_splitparam(b, "=")
			i := get_ibc_index(d[0])
			if i >= 0 && len(d) >= 2 {
				if !ibc[i].MustPause && !ibc[i].MustStop {
					vol_now := ibc[i].VolInOut
					fmt.Println("DEBUG RUN MANYDRAW OUT: par=", b, "i=", i, "d=", d, "vol_ini=", vol_ini[d[0]], "vol_now=", vol_now)
					vol, err := strconv.Atoi(d[1])
					if err == nil {
						if vol_now == 0 || (vol_now <= vol_ini[d[0]] && vol_now <= vol_ini[d[0]]-float64(vol)) {
							fmt.Println("DEBUG SCP RUN MANYDRAW OUT: NADA a FAZER, esenvase atingido para", d[0], "desenvase=", vol, "vol_ini=", vol_ini[d[0]], "vol_now=", vol_now)
						} else {
							ibc[i].OutID = dest
							ibc[i].Withdraw = uint32(vol)
							fmt.Println("DEBUG SCP RUN MANYDRAW OUT: Desenvase de", d[0], " para", dest, " Volume", vol)
							scp_run_withdraw(scp_ibc, d[0], false, false)
							if !ibc[i].MustPause && !ibc[i].MustStop && ibc[i].VolInOut > 0 && (ibc[i].VolInOut >= vol_ini[d[0]] || ibc[i].VolInOut >= vol_ini[d[0]]-float64(vol)) {
								no_ok++
								fmt.Println("ERROR SCP RUN MANYDRAW OUT: Volume de Desenvase NAO ATINGIDO em", d[0])
								if n_tries < 3 {
									board_add_message("AVolume de desenvase/transferência não atingido em "+d[0]+". Nova tentativa será feita em instantes", "")
								} else {
									board_add_message("AApós 3 tentativas não foi possível desenvasar/transferir volume esperado em "+d[0]+". Favor checar dispositivos", "")
								}
							}
						}
					} else {
						checkErr(err)
					}
				} else {
					fmt.Println("DEBUG SCP RUN MANYDRAW OUT: Desenvase nao executado pois mustpause/muststop acionado para", ibc[i].IBCID)
				}
			}
		}
		n_tries++
		if n_tries >= 3 || no_ok == 0 {
			break
		}
	}

}

func scp_process_conn(conn net.Conn) {
	defer conn.Close()
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

					case scp_par_getph:
						var msg MsgReturn
						if bio[ind].RegresPH[0] != 0 && bio[ind].RegresPH[1] != 0 {
							phtmp := scp_get_ph(bioid)
							if phtmp > 0 {
								msgstr := fmt.Sprintf("%2.1f", phtmp)
								msg = MsgReturn{scp_ack, msgstr}
							} else {
								msg = MsgReturn{scp_err, "Erro ao aferir PH. Sensor retornou valor inválido. Repita o processo e, se persistir o erro, entre em contato com o SAC"}
								bio_add_message(bioid, "E"+msg.Message, "")
							}
						} else {
							msg = MsgReturn{scp_err, "Não é possível fazer a Aferição pois o Sensor de PH não foi devidamente calibrado"}
							bio_add_message(bioid, "E"+msg.Message, "")
						}
						msgjson, _ := json.Marshal(msg)
						conn.Write([]byte(msgjson))

					case scp_par_stopall:
						if scp_fullstop_device(bioid, scp_bioreactor, true, true) {
							conn.Write([]byte(scp_ack))
						} else {
							conn.Write([]byte(scp_err))
						}

					case scp_par_restore:
						if bio[ind].Status != bio_ready && bio[ind].Status != bio_empty {
							board_add_message("E"+bioid+" deve estar PRONTO ou VAZIO para que o cultivo seja Restaurado", "")
							bio_add_message(bioid, "EBiorreator deve estar PRONTO ou VAZIO para que o cultivo seja Restaurado", "")
							conn.Write([]byte(scp_err))
						} else {
							if len(bio[ind].Queue) > 0 && bio[ind].Volume > 0 {
								fmt.Println("DEBUG PROCESS CONN: CONFIG: Restaurando cultivo que estava na Queue", bioid)
								bio[ind].Status = bio_producting
								board_add_message("A"+bioid+" Com cultivo Restaurado e retornando ao ponto anterior", "")
								bio_add_message(bioid, "ABiorreator com cultivo Restaurado e retornando ao ponto anterior", "")
							} else if bio[ind].Volume > 0 {
								orgcode := bio[ind].OrgCode
								if len(orgcode) > 0 && len(organs[orgcode].Orgname) > 0 {
									bio[ind].Status = bio_empty
									d1000 := math.Abs(float64(bio[ind].Volume) - 1000)
									d2000 := math.Abs(float64(bio[ind].Volume) - 2000)
									volume := "2000"
									if d1000 < d2000 {
										volume = "1000"
									}
									biotask := []string{bioid + ",0," + orgcode + "," + volume}
									n := create_sched(biotask)
									fmt.Println("DEBUG PROCESS CONN: CONFIG: Restore recriando task", bioid, biotask, n)
									board_add_message("A"+bioid+" Cultivo "+organs[orgcode].Orgname+" Restaurado a partir do Início. Quando for solicitado Cloro, Meio e Inóculo, basta pressionar PROSSEGUIR se os mesmos já tiverem sido inseridos", "")
									bio_add_message(bioid, "ACultivo Restaurado a partir do Início. Quando for solicitado Cloro, Meio e Inóculo, basta pressionar PROSSEGUIR se os mesmos já tiverem sido inseridos", "")
								} else {
									fmt.Println("ERROR PROCESS CONN: CONFIG: Microorganismo nao encontrado", bioid, orgcode)
									board_add_message("E"+bioid+" Não tem cultivo definido. Não é possível restaurar", "")
								}
							} else {
								board_add_message("E"+bioid+" Está com volume ZERO. Não é possível restaurar cultivo", "")
							}
						}

					case scp_par_resetdata:
						if bio[ind].Status == bio_empty || bio[ind].Status == bio_ready {
							bio_add_message(bioid, "ATodos os dados do Biorreator serão refinidos", "")
							board_add_message("ATodos os dados do "+bioid+" serão refinidos", "")
							bio[ind].OrgCode = "EMPTY"
							bio[ind].Organism = ""
							bio[ind].VolInOut = 0
							bio[ind].Volume = 0
							bio[ind].Level = 0
							bio[ind].Status = bio_empty
							bio[ind].Queue = []string{}
							bio[ind].RedoQueue = []string{}
							bio[ind].MustOffQueue = []string{}
							bio[ind].MustStop = false
							bio[ind].MustPause = false
							bio[ind].Timetotal[0] = 0
							bio[ind].Timetotal[1] = 0
							bio[ind].Timeleft[0] = 0
							bio[ind].Timeleft[1] = 0
							bio[ind].Step[0] = 0
							bio[ind].Step[1] = 0
							bio[ind].ShowVol = true
							bio[ind].MainStatus = mainstatus_empty
						} else {
							bio_add_message(bioid, "ESó é possível redefinir um Biorreator se ele estiver VAZIO ou PRONTO", "")
						}

					case scp_par_deviceaddr:
						if len(params) > 4 {
							devaddr := params[4]
							biocfg := bio_cfg[bioid]
							biocfg.Deviceaddr = strings.ToUpper(devaddr)
							bio_cfg[bioid] = biocfg
							fmt.Println("DEBUG SCP PROCESS CON: Mudanca endereco do Biorreator", bioid, " para", devaddr, " = ", bio_cfg[bioid])
							conn.Write([]byte(scp_ack))
							save_bios_conf(localconfig_path + "bio_conf.csv")
						}

					case scp_par_screenaddr:
						if len(params) > 4 {
							devaddr := params[4]
							biocfg := bio_cfg[bioid]
							biocfg.Screenaddr = strings.ToUpper(devaddr)
							bio_cfg[bioid] = biocfg
							fmt.Println("DEBUG SCP PROCESS CON: Mudanca endereco da tela do Biorreator", bioid, " para", devaddr, " = ", bio_cfg[bioid])
							conn.Write([]byte(scp_ack))
						}

					case scp_par_ph4:
						if !bio[ind].Aerator {
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
								fmt.Println("DEBUG CONFIG: ", bioid, "Mediana Voltagem PH 4", bio[ind].PHref[0], " amostras =", n)
								msg := MsgReturn{scp_ack, "Leitura do PH 4.0 feita com sucesso"}
								msgjson, _ := json.Marshal(msg)
								conn.Write([]byte(msgjson))

							} else {
								bio[ind].PHref[0] = 0
								fmt.Println("ERROR CONFIG: Valores INVALIDOS de PH 4")
								msg := MsgReturn{scp_err, "ERRO na calibração: Dados de PH 4 inválidos. Favor checar painel, cabos e sensor de PH"}
								bio_add_message(bioid, "E"+msg.Message, "")
								msgjson, _ := json.Marshal(msg)
								conn.Write([]byte(msgjson))
							}
						} else {
							fmt.Println("ERROR CONFIG: Tentativa de ajuste de PH 4 com aerador ligado")
							msg := MsgReturn{scp_err, "ERRO na calibração: Não é possível fazer a calibração com o Aerador ligado. Deslige-o e repita o procedimento"}
							bio_add_message(bioid, "E"+msg.Message, "")
							msgjson, _ := json.Marshal(msg)
							conn.Write([]byte(msgjson))

						}

					case scp_par_ph7:
						if !bio[ind].Aerator {
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
								if math.Abs(mediana-bio[ind].PHref[0]) >= 0.2 {
									bio[ind].PHref[1] = mediana
									fmt.Println("DEBUG CONFIG: ", bioid, "Mediana Voltagem PH 7", bio[ind].PHref[1], " amostras =", n)
									msg := MsgReturn{scp_ack, "Leitura do PH 7.0 feita com sucesso"}
									msgjson, _ := json.Marshal(msg)
									conn.Write([]byte(msgjson))

								} else {
									bio[ind].PHref[1] = 0
									msg := MsgReturn{scp_err, "ERRO na calibração: Dados de PH 7 muito próximos do PH 4. Favor checar solução de teste, painel, cabos e sensor de PH"}
									bio_add_message(bioid, "E"+msg.Message, "")
									msgjson, _ := json.Marshal(msg)
									conn.Write([]byte(msgjson))

								}
							} else {
								bio[ind].PHref[1] = 0
								fmt.Println("ERROR CONFIG: Valores INVALIDOS de PH 7")
								msg := MsgReturn{scp_err, "ERRO na calibração: Dados de PH 7 inválidos. Favor checar painel, cabos e sensor de PH"}
								bio_add_message(bioid, "E"+msg.Message, "")
								msgjson, _ := json.Marshal(msg)
								conn.Write([]byte(msgjson))
							}
						} else {
							fmt.Println("ERROR CONFIG: Tentativa de ajuste de PH 7 com aerador ligado")
							msg := MsgReturn{scp_err, "ERRO na calibração: Não é possível fazer a calibração com o Aerador ligado. Deslige-o e repita o procedimento"}
							bio_add_message(bioid, "E"+msg.Message, "")
							msgjson, _ := json.Marshal(msg)
							conn.Write([]byte(msgjson))

						}

					case scp_par_ph10:
						if !bio[ind].Aerator {
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
								if math.Abs(mediana-bio[ind].PHref[1]) >= 0.2 && math.Abs(mediana-bio[ind].PHref[0]) >= 0.4 {
									bio[ind].PHref[2] = mediana
									fmt.Println("DEBUG CONFIG: ", bioid, "Mediana Voltagem PH 10", bio[ind].PHref[2], " amostras =", n)
									msg := MsgReturn{scp_ack, "Leitura do PH 10.0 feita com sucesso"}
									msgjson, _ := json.Marshal(msg)
									conn.Write([]byte(msgjson))

								} else if math.Abs(mediana-bio[ind].PHref[0]) < 0.4 {
									bio[ind].PHref[2] = 0
									msg := MsgReturn{scp_err, "ERRO na calibração: Dados de PH 10 muito próximos do PH 4. Favor checar solução de teste, painel, cabos e sensor de PH"}
									bio_add_message(bioid, "E"+msg.Message, "")
									msgjson, _ := json.Marshal(msg)
									conn.Write([]byte(msgjson))

								} else {
									bio[ind].PHref[2] = 0
									msg := MsgReturn{scp_err, "ERRO na calibração: Dados de PH 10 muito próximos do PH 7. Favor checar solução de teste, painel, cabos e sensor de PH"}
									bio_add_message(bioid, "E"+msg.Message, "")
									msgjson, _ := json.Marshal(msg)
									conn.Write([]byte(msgjson))

								}
							} else {
								bio[ind].PHref[2] = 0
								fmt.Println("ERROR CONFIG: Valores INVALIDOS de PH 10")
								msg := MsgReturn{scp_err, "ERRO na calibração: Dados de PH 10 inválidos. Favor checar painel, cabos e sensor de PH"}
								bio_add_message(bioid, "E"+msg.Message, "")
								msgjson, _ := json.Marshal(msg)
								conn.Write([]byte(msgjson))

							}
						} else {
							fmt.Println("ERROR CONFIG: Tentativa de ajuste de PH 10 com aerador ligado")
							msg := MsgReturn{scp_err, "ERRO na calibração: Não é possível fazer a calibração com o Aerador ligado. Deslige-o e repita o procedimento"}
							bio_add_message(bioid, "E"+msg.Message, "")
							msgjson, _ := json.Marshal(msg)
							conn.Write([]byte(msgjson))

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
							if b0 == 0 && b1 == 0 {
								fmt.Println("ERROR CONFIG: Nao foi possivel efetuar Regressao Linear: b0=", b0, " b1=", b1)
								msg := MsgReturn{scp_err, "Não foi possível efetuar a Calibração do Sensor de PH. Verifique PHs 4, 7 e 10"}
								bio_add_message(bioid, "E"+msg.Message, "")
								msgjson, _ := json.Marshal(msg)
								conn.Write([]byte(msgjson))

							} else {
								fmt.Println("DEBUG CONFIG: Coeficientes da Regressao Linear: b0=", b0, " b1=", b1)
								msg := MsgReturn{scp_ack, "Calibração do Sensor de PH efetuada"}
								bio_add_message(bioid, "I"+msg.Message, "")
								msgjson, _ := json.Marshal(msg)
								conn.Write([]byte(msgjson))
							}
						} else {
							bio[ind].RegresPH[0] = 0
							bio[ind].RegresPH[1] = 0
							fmt.Println("ERROR CONFIG: Nao e possivel fazer regressao linear, valores invalidos", bio[ind].PHref)
							msg := MsgReturn{scp_err, "Não foi possivel efetuar a Calibração do Sensor de PH. Verifique PHs 4, 7 e 10"}
							bio_add_message(bioid, "E"+msg.Message, "")
							msgjson, _ := json.Marshal(msg)
							conn.Write([]byte(msgjson))
						}
					}
				} else {
					fmt.Println("ERROR CONFIG: Biorreator nao existe", bioid)
				}
			} else {
				fmt.Println("ERROR CONFIG: BIORREATOR - Numero de parametros invalido", params)
			}
		case scp_ibc:
			if len(params) > 3 {
				ibcid := params[2]
				fmt.Println("DEBUG CONFIG: IBC", ibcid, params)
				ind := get_ibc_index(ibcid)
				if ind >= 0 {
					switch params[3] {
					case scp_par_stopall:
						if scp_fullstop_device(ibcid, scp_ibc, true, true) {
							conn.Write([]byte(scp_ack))
						} else {
							conn.Write([]byte(scp_err))
						}
					case scp_par_getconfig:
						ibccfg, ok := ibc_cfg[ibcid]
						fmt.Println("->", ind, ibccfg, ok)
						if ok {
							buf, err := json.Marshal(ibccfg)
							checkErr(err)
							conn.Write([]byte(buf))
						}
					case scp_par_deviceaddr:
						if len(params) > 4 {
							devaddr := params[4]
							ibccfg := ibc_cfg[ibcid]
							ibccfg.Deviceaddr = strings.ToUpper(devaddr)
							ibc_cfg[ibcid] = ibccfg
							fmt.Println("DEBUG SCP PROCESS CON: Mudanca endereco do IBC", ibcid, " para", devaddr, " = ", ibc_cfg[ibcid])
							conn.Write([]byte(scp_ack))
							save_ibcs_conf(localconfig_path + "ibc_conf.csv")
						}

					case scp_par_setvolume:
						if len(params) > 4 {
							newvolume_str := params[4]
							newvolume, err := strconv.Atoi(newvolume_str)
							if err != nil {
								fmt.Println("ERROR CONFIG SETVOLUME: Volume inválido", ibcid, newvolume_str)
								conn.Write([]byte(scp_err))
							} else {
								ibc[ind].VolInOut = float64(newvolume)
								ibc[ind].Volume = uint32(newvolume)
								if newvolume == 0 {
									ibc[ind].Status = bio_empty
								} else if ibc[ind].Status == bio_empty {
									ibc[ind].Status = bio_ready
								}
								msg := fmt.Sprintf("AVolume do IBC %s redefinido para %d Litros", ibcid, newvolume)
								board_add_message(msg, "")
								fmt.Println("DEBUG CONFIG SETVOLUME: ", ibcid, " Volume redefinido para", newvolume)
								conn.Write([]byte(scp_ack))
							}
						}

					case scp_par_resetdata:
						if ibc[ind].Status == bio_empty || ibc[ind].Status == bio_ready {
							board_add_message("ATodos os dados do "+ibcid+" serão refinidos", "")
							ibc[ind].OrgCode = "EMPTY"
							ibc[ind].Organism = ""
							ibc[ind].VolInOut = 0
							ibc[ind].Volume = 0
							ibc[ind].Level = 0
							ibc[ind].Status = bio_empty
							ibc[ind].Queue = []string{}
							ibc[ind].RedoQueue = []string{}
							ibc[ind].MustOffQueue = []string{}
							ibc[ind].MustStop = false
							ibc[ind].Timetotal[0] = 0
							ibc[ind].Timetotal[1] = 0
							ibc[ind].Step[0] = 0
							ibc[ind].Step[1] = 0
							ibc[ind].ShowVol = true
							ibc[ind].MainStatus = mainstatus_empty
						} else {
							board_add_message("ANão é possível redifinir os dados do "+ibcid+" se não estiver VAZIO ou PRONTO", "")
						}

					}
				}
			}
		case scp_totem:
			if len(params) > 3 {
				fmt.Println("DEBUG CONFIG: Totem", params)
				totemid := params[2]
				ind := get_totem_index(totemid)
				if ind >= 0 {
					switch params[3] {
					case scp_par_stopall:
						if scp_fullstop_device(totemid, scp_totem, false, false) {
							conn.Write([]byte(scp_ack))
						} else {
							conn.Write([]byte(scp_err))
						}
					case scp_par_getconfig:
						totemcfg, ok := totem_cfg[totemid]
						if ok {
							buf, err := json.Marshal(totemcfg)
							checkErr(err)
							conn.Write([]byte(buf))
						}
					case scp_par_deviceaddr:
						if len(params) > 4 {
							devaddr := params[4]
							totemcfg := totem_cfg[totemid]
							totemcfg.Deviceaddr = strings.ToUpper(devaddr)
							totem_cfg[totemid] = totemcfg
							fmt.Println("DEBUG SCP PROCESS CON: Mudanca endereco do Totem", totemid, " para", devaddr, " = ", totem_cfg[totemid])
							conn.Write([]byte(scp_ack))
							save_totems_conf(localconfig_path + "totem_conf.csv")
						}
					}
				}
			}
		case scp_biofabrica:
			fmt.Println("DEBUG:", params)
			if len(params) > 3 {
				cmd := params[2]
				switch cmd {
				case scp_par_loadbfdata:
					fmt.Println("DEBUG CONFIG: RELOAD bf data")
					n_bf := load_bf_data(localconfig_path + "bf_data_new.csv")
					if n_bf < 1 {
						fmt.Println("ERROR CONFIG: Falha ao ler Arquivo contendo dados da Biofabrica nao encontrado")
						conn.Write([]byte(scp_err))
						return
					}
					buf, err := json.Marshal(mybf)
					checkErr(err)
					conn.Write([]byte(buf))

				case scp_par_bfdata:
					fmt.Println("DEBUG CONFIG: GET dados biofabrica", mybf)
					buf, err := json.Marshal(mybf)
					checkErr(err)
					conn.Write([]byte(buf))

				case scp_par_stopall:
					if scp_fullstop_device("ALL", scp_biofabrica, false, false) {
						conn.Write([]byte(scp_ack))
					} else {
						conn.Write([]byte(scp_err))
					}
				case scp_par_getconfig:
					fmt.Println("DEBUG CONFIG: GET configuracoes biofabrica", biofabrica_cfg)
					v := make([]Biofabrica_cfg, 0)
					for _, e := range biofabrica_cfg {
						v = append(v, e)
					}
					buf, err := json.Marshal(v)
					checkErr(err)
					conn.Write([]byte(buf))

				case scp_par_reconfigdev:
					board_add_message("ATodos os Equipamentos serão reconfigurados", "")
					scp_setup_devices(true)

				case scp_par_deviceaddr:
					if len(params) > 4 {
						devid := params[3]
						devaddr := params[4]
						devcfg, ok := biofabrica_cfg[devid]
						if ok {
							devcfg.Deviceaddr = strings.ToUpper(devaddr)
							biofabrica_cfg[devid] = devcfg
							fmt.Println("DEBUG SCP PROCESS CON: Mudanca endereco do Biofabrica", devid, " para", devaddr, " = ", biofabrica_cfg[devid])
							conn.Write([]byte(scp_ack))
						}
						save_bf_conf(localconfig_path + "biofabrica_conf.csv")
					}

				case scp_par_save:
					fmt.Println("DEBUG CONFIG: Salvando configuracoes")
					save_all_data(data_filename)
					save_bios_conf(localconfig_path + "bio_conf.csv")
					save_ibcs_conf(localconfig_path + "ibc_conf.csv")
					save_totems_conf(localconfig_path + "totem_conf.csv")
					save_bf_conf(localconfig_path + "biofabrica_conf.csv")
					conn.Write([]byte(scp_ack))

				case scp_par_restart:
					fmt.Println("DEBUG CONFIG: Restartando Service")
					save_all_data(data_filename)
					scp_restart_services()

				case scp_par_resetdata:
					fmt.Println("DEBUG CONFIG: Redefinindo Tarefas")
					schedule = []Scheditem{}
					waitlist_del_message("I")
					waitlist_del_message("B")
					set_allvalvs_status()

				case scp_par_upgrade:
					fmt.Println("DEBUG CONFIG: Upgrade em andamento")
					system_upgrade()

				// case scp_par_version:
				// 	buf, err := json.Marshal(biofabrica.Version)
				// 	checkErr(err)
				// 	conn.Write([]byte(buf))

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

				case scp_par_techmode:
					if len(params) > 4 {
						flag_str := params[3]
						fmt.Println("DEBUG CONFIG: Mudando TECHMODE para", flag_str)
						flag, err := strconv.ParseBool(flag_str)
						if err != nil {
							checkErr(err)
							conn.Write([]byte(scp_err))
						} else {
							biofabrica.TechMode = flag
							conn.Write([]byte(scp_ack))
						}
					} else {
						fmt.Println("ERROR CONFIG: BIOFABRICA TECHMODE - Numero de parametros invalido", params)
					}

				}

			} else {
				fmt.Println("ERROR CONFIG: BIOFABRICA - Numero de parametros invalido", params)
			}
		}

	case scp_start:
		if biofabrica.Critical != scp_ready {
			fmt.Println("ERROR PROCESS CONN: Nao permitido executar comando quando Critical ativo", params, biofabrica.Critical)
			board_add_message("ENão é possivel executar ação enquanto a Biofábrica estiver PARADA", "")
			conn.Write([]byte(scp_err))
			return
		}
		scp_object := params[1]
		switch scp_object {
		case scp_bioreactor:
			bioid := params[2]
			orgcode := params[3]
			volume := "2000"
			if len(params) > 4 {
				volume = params[4]
			}
			fmt.Println("DEBUG SCP START: Iniciando ", bioid, orgcode, params)
			ind := get_bio_index(bioid)
			if ind < 0 {
				fmt.Println("ERROR START: Biorreator nao existe", bioid)
				break
			}
			ind_totem := get_totem_index("TOTEM01")
			if ind_totem >= 0 {
				if totem[ind_totem].Status == bio_error || totem[ind_totem].Status == bio_nonexist {
					fmt.Println("ERROR START: TOTEM01 com falaha - Nao foi possivel inicial JOB no Biorreator ", bioid)
					board_add_message("E"+bioid+" não pode iniciar tarefa pois TOTEM01 com falha", "")
					bio_add_message(bioid, "ENão pode iniciar tarefa pois TOTEM01 com falha", "")
					return
				}
			}
			if orgcode != scp_par_cip && bio[ind].RegresPH[0] == 0 && bio[ind].RegresPH[1] == 0 && !devmode { //
				fmt.Println("ERROR START: Biorreator nao teve o PH Calibrado, impossivel iniciar cultivo", bioid)
				bio_add_message(bioid, "EImpossível iniciar cultivo, sensor de PH não calibrado", "")
				break
			}
			if orgcode == scp_par_cip || len(organs[orgcode].Orgname) > 0 {
				if orgcode == scp_par_cip {
					if (bio[ind].Status != bio_empty && bio[ind].Status != bio_ready) || bio[ind].Volume > 0 {
						fmt.Println("ERROR START: CIP invalido, biorreator nao esta vazio ou status invalido", bioid, bio[ind].Status, bio[ind].Volume)
						if bio[ind].Volume > 0 {
							bio_add_message(bioid, "ENão é possivel realizar CIP num biorretor que não esteja VAZIO", "")
						} else if bio[ind].Status == bio_cip {
							bio_add_message(bioid, "ECIP já iniciado no Biorreator. Aguarde", "")
						} else {
							bio_add_message(bioid, "EStatus "+bio[ind].Status+" não permite CIP", "")
						}
						return
					}
					// bio[ind].Status = bio_cip
				} else {
					if bio[ind].Volume > 2500 {
						bio_add_message(bioid, "ENão é possivel iniciar cultivo num biorretor que não esteja VAZIO", "")
						return
					} else if bio[ind].Status != bio_empty {
						bio_add_message(bioid, "EStatus "+bio[ind].Status+" não permite início de cultivo", "")
						return
					}
					// bio[ind].Status = bio_starting
				}
				bio_add_message(bioid, "ALembre-se de checar o nível dos produtos necessários ao Cultivo (PH+ , PH- e Antiespumante). Cheque também as mangueiras", "")
				biotask := []string{bioid + ",0," + orgcode + "," + volume}
				fmt.Println("DEBUG PROCESS CONN: START criando task:", biotask)
				n := create_sched(biotask)
				// if n > 0 && !schedrunning {
				// 	go scp_scheduler()
				// }
				fmt.Println("DEBUG SCP START: biotask=", biotask, "n=", n, "sched=", schedrunning)
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
			ind_totem := get_totem_index("TOTEM02")
			if ind_totem >= 0 {
				if totem[ind_totem].Status == bio_error || totem[ind_totem].Status == bio_nonexist {
					fmt.Println("ERROR START: TOTEM02 com falaha - Nao foi possivel inicial JOB no Biorreator ", ibcid)
					board_add_message("E"+ibcid+" não pode iniciar tarefa pois TOTEM02 com falha", "")
					break
				}
			}
			if orgcode == scp_par_cip || len(organs[orgcode].Orgname) > 0 {
				fmt.Println("START", orgcode)
				biotask := []string{ibcid + ",0," + orgcode}
				n := create_sched(biotask)
				// if n > 0 && !schedrunning {
				// 	go scp_scheduler()
				// }
				fmt.Println("DEBUG SCP START: ibctask=", biotask, "n=", n, "sched=", schedrunning)
			} else {
				fmt.Println("ORG INVALIDO")
			}

		case scp_biofabrica:
			cmdpar := params[2]
			switch cmdpar {
			case scp_par_linewash:
				time_str := params[4]
				time_int, err := strconv.Atoi(time_str)
				if err != nil {
					checkErr(err)
					time_int = 30
				}
				scp_run_linewash(params[3], time_int)

			case scp_par_linecip:
				scp_run_linecip(params[3])

			default:
				fmt.Println("ERROR START: Parametro de Biofabrica invalido", params)
			}
		}

	case scp_stop:
		if biofabrica.Critical != scp_ready {
			fmt.Println("ERROR PROCESS CONN: Nao permitido executar comando quando Critical ativo", params, biofabrica.Critical)
			board_add_message("ENão é possivel executar ação enquanto a Biofábrica estiver PARADA", "")
			conn.Write([]byte(scp_err))
			return
		}
		devtype := params[1]
		id := params[2]
		if !stop_device(devtype, id) {
			fmt.Println("ERROR STOP: Nao foi possivel parar dispositivo", devtype, id)
			break
		}

	case scp_pause:
		if biofabrica.Critical != scp_ready {
			fmt.Println("ERROR PROCESS CONN: Nao permitido executar comando quando Critical ativo", params, biofabrica.Critical)
			board_add_message("ENão é possivel executar ação enquanto a Biofábrica estiver PARADA", "")
			conn.Write([]byte(scp_err))
			return
		}
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
		if biofabrica.Critical != scp_ready {
			fmt.Println("ERROR PROCESS CONN: Nao permitido executar comando quando Critical ativo", params, biofabrica.Critical)
			board_add_message("ENão é possivel executar ação enquanto a Biofábrica estiver PARADA", "")
			conn.Write([]byte(scp_err))
			return
		}
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
				fmt.Println("DEBUG WDPANEL START: ", scp_ibc, ibc_id)
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
		if biofabrica.Critical != scp_ready {
			fmt.Println("ERROR PROCESS CONN: Nao permitido executar comando quando Critical ativo", params, biofabrica.Critical)
			board_add_message("ENão é possivel executar ação enquanto a Biofábrica estiver PARADA", "")
			conn.Write([]byte(scp_err))
			return
		}
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

				case scp_par_continue:
					fmt.Println("DEBUG SCP PROCESS CONN: Continue pressionado", bioid, params)
					if ind >= 0 {
						bio[ind].Continue = true
					} else {
						fmt.Println("ERROR SCP PROCESS CONN: Biorreator nao existe", bioid, params)
					}

				case scp_par_clenaperis:
					fmt.Println("DEBUG SCP PROCESS CONN: PAR CLEANPERIS", params, subparams)
					if bio[ind].Status != bio_empty {
						bio_add_message(bioid, "ENão é permitido Ligar Peristálticas se o Biorreator não estiver VAZIO", "")
						conn.Write([]byte(scp_err))
					} else {
						clean_time := 10
						if len(subparams) >= 3 {
							clean_time, err = strconv.Atoi(subparams[2])
							if err != nil {
								checkErr(err)
								clean_time = 10
							}
						}
						msg := fmt.Sprintf("ILigando Peristálticas para Limpeza por %d segundos", clean_time)
						bio_add_message(bioid, msg, "")
						for _, p := range []string{"P1", "P2", "P3", "P4", "P5"} {
							scp_turn_peris(scp_bioreactor, bioid, p, 1)
						}
						for i := 0; i < clean_time; i++ {
							if bio[ind].MustPause || bio[ind].MustStop || biofabrica.Critical == scp_stopall {
								break
							}
							time.Sleep(time.Second)
						}
						for _, p := range []string{"P1", "P2", "P3", "P4", "P5"} {
							scp_turn_peris(scp_bioreactor, bioid, p, 0)
						}
						conn.Write([]byte(scp_ack))
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
									if bio[ind].Status == bio_circulate { // CHECAR se PRECISA TESTAR MUSTSTOP/PAUSE
										if bio[ind].LastStatus != bio_circulate {
											bio[ind].Status = bio[ind].LastStatus
										} else {
											bio[ind].Status = bio_ready
										}
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
					if err != nil {
						fmt.Println("ERROR PUT BIORREACTOR: WITHDRAW: Volume passado para o desenvase invalido", bioid, subparams)
						checkErr(err)
					}
					if err == nil {
						fmt.Println("DEBUG PUT BIORREACTOR: WITHDRAW: Desenvase solicitado de=", bioid, " para=", bio[ind].OutID, "volume=", vol)
						bio[ind].Withdraw = uint32(vol)
						devout_type := get_scp_type(bio[ind].OutID)
						if devout_type == scp_ibc && vol != 0 {
							bio[ind].Withdraw = bio[ind].Volume
						}
						if bio[ind].Status != bio_ready && (bio[ind].Status == bio_empty && bio[ind].Volume == 0) {
							if bio[ind].Status == bio_unloading {
								bio_add_message(bioid, "EDesenvase em andamento no Biorreator. Aguarde", "")
								conn.Write([]byte(scp_err))
								return
							}
							bio_add_message(bioid, "ENão é possível fazer desenvase em um biorreator que não esteja PRONTO", "")
							conn.Write([]byte(scp_err))
							return
						}
						conn.Write([]byte(scp_ack))
						if bio[ind].Withdraw > 0 {
							if get_scp_type(bio[ind].OutID) == scp_ibc {
								ibc_ind := get_ibc_index(bio[ind].OutID)
								if ibc_ind >= 0 {
									if ibc[ibc_ind].Status == bio_empty || ibc[ibc_ind].OrgCode == bio[ind].OrgCode {
										if bio[ind].Volume+ibc[ibc_ind].Volume <= ibc_cfg[bio[ind].OutID].Maxvolume+bio_ibctransftol {
											// bio[ind].Status = bio_unloading
											bio[ind].MainStatus = mainstatus_org
											board_add_message("IEnxague de Linha para Transferência do "+bioid, "")
											bio_add_message(bioid, "IEnxague de Linha antes da Transferência", "")
											for {
												if bio[ind].MustPause || bio[ind].MustStop || biofabrica.Critical == scp_stopall {
													break
												}
												wtime := 180
												if devmode || testmode {
													wtime = 30
												}
												if scp_run_linewash(line_23, wtime) {
													break
												} else {
													waitlist_add_message("A"+bioid+" aguardando linha para enxague", bioid+"WAITLINEWASH")
													bio_add_message(bioid, "AAguardando linha para enxague", "WAITLINEWASH")
												}
												time.Sleep(2 * time.Second)
											}
											waitlist_del_message(bioid + "WAITLINEWASH")
											bio_del_message(bioid, "WAITLINEWASH")

											bio_add_message(bioid, "ITransferência iniciada", "")
											ibc[ibc_ind].Status = bio_loading
											for i := 0; i < 3; i++ {
												if bio[ind].MustPause || bio[ind].MustStop || biofabrica.Critical == scp_stopall {
													break
												}
												if scp_run_withdraw(scp_bioreactor, bioid, true, true) >= 0 {
													break
												}
												time.Sleep(2 * time.Second)
											}
											if ibc[ibc_ind].Volume > 0 {
												// board_add_message("APASSEI AQUI", "")
												ibc[ibc_ind].OrgCode = bio[ind].OrgCode
												ibc[ibc_ind].Organism = bio[ind].Organism
												ibc[ibc_ind].Status = bio_ready
											}

										} else {
											bio_add_message(bioid, "EAtual volume do "+bio[ind].OutID+" não suporte a transferência", "")
										}
									} else {
										bio_add_message(bioid, "ENão é permitido fazer transferência do Biorreator para um IBC que não esteja vazio ou com o mesmo microrganismo", "")
									}
								} else {
									fmt.Println("ERROR WITHDRAW: IBC destino não encontrado", bioid, bio[ind].OutID)
								}
							} else {
								bio_add_message(bioid, "IDesenvase iniciado", "")
								conn.Write([]byte(scp_ack))
								for {
									if bio[ind].MustPause || bio[ind].MustStop || biofabrica.Critical == scp_stopall {
										break
									}
									if scp_run_withdraw(scp_bioreactor, bioid, true, false) >= 0 {
										break
									}
									if bio[ind].Volume == 0 || bio[ind].Withdraw == 0 {
										break
									}
									time.Sleep(2 * time.Second)
								}
							}
						}
					} else {
						conn.Write([]byte(scp_err))
					}

				case scp_dev_pump:
					var cmd2 string
					value, err := strconv.ParseBool(subparams[1])
					checkErr(err)
					biodev := bio_cfg[bioid].Deviceaddr
					// bioscr := bio_cfg[bioid].Screenaddr
					pumpdev := bio_cfg[bioid].Pump_dev
					bio[ind].Pumpstatus = value
					if value {
						cmd2 = "CMD/" + biodev + "/PUT/" + pumpdev + ",1/END"
						// cmd3 = "CMD/" + bioscr + "/PUT/S270,1/END"
					} else {
						cmd2 = "CMD/" + biodev + "/PUT/" + pumpdev + ",0/END"
						// cmd3 = "CMD/" + bioscr + "/PUT/S270,0/END"
					}
					ret2 := scp_sendmsg_orch(cmd2)
					fmt.Println("RET CMD2 =", ret2)
					// ret3 := scp_sendmsg_orch(cmd3)
					// fmt.Println("RET CMD3 =", ret3)
					if !value && bio[ind].Heater {
						bio_add_message(bioid, "AATENÇÃO: Bomba foi DESLIGADA com resistência LIGADA! Efetuando desligamento automático", "")
						scp_turn_heater(bioid, 0, false)
					}
					conn.Write([]byte(scp_ack))

				case scp_dev_heater:
					value, err := strconv.ParseBool(subparams[1])
					if err != nil {
						checkErr(err)
						conn.Write([]byte(scp_err))
						return
					}
					if value {
						bio_add_message(bioid, "ENão é permitido ligar a resistência manualmente", "")
						conn.Write([]byte(scp_err))
						return
					}
					if !scp_turn_heater(bioid, 0, false) {
						fmt.Println("SCP PROCESS CONN: Erro ao delisgar resistência")
						bio_add_message(bioid, "EFALHA ao DESLIGAR RESISTÊNCIA", "")
						conn.Write([]byte(scp_err))
						return
					}
					conn.Write([]byte(scp_ack))

				case scp_dev_aero:
					var cmd1, cmd2 string
					value, err := strconv.ParseBool(subparams[1])
					checkErr(err)
					bio[ind].Aerator = value
					biodev := bio_cfg[bioid].Deviceaddr
					// bioscr := bio_cfg[bioid].Screenaddr
					aerodev := bio_cfg[bioid].Aero_dev
					if value {
						cmd1 = "CMD/" + biodev + "/PUT/D27,1/END"
						cmd2 = "CMD/" + biodev + "/PUT/" + aerodev + ",255/END"
						// cmd3 = "CMD/" + bioscr + "/PUT/S271,1/END"

					} else {
						cmd1 = "CMD/" + biodev + "/PUT/D27,0/END"
						cmd2 = "CMD/" + biodev + "/PUT/" + aerodev + ",0/END"
						// cmd3 = "CMD/" + bioscr + "/PUT/S271,0/END"
					}
					ret1 := scp_sendmsg_orch(cmd1)
					fmt.Println("RET CMD1 =", ret1)
					ret2 := scp_sendmsg_orch(cmd2)
					fmt.Println("RET CMD2 =", ret2)
					// ret3 := scp_sendmsg_orch(cmd3)
					// fmt.Println("RET CMD3 =", ret3)
					conn.Write([]byte(scp_ack))

				case scp_dev_valve:
					value_valve, err := strconv.Atoi(subparams[1])
					checkErr(err)
					value_status, err := strconv.Atoi(subparams[2])
					checkErr(err)
					if bio[ind].Heater && value_status == 0 {
						bio_add_message(bioid, "AATENÇÃO: Válvulas sendo fechadas manualmente com resistência ligada. Efetuando desligamento automático", "")
						scp_turn_heater(bioid, 0, false)
					}
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
							conn.Write([]byte(scp_ack))
							for i := 0; i < 3; i++ {
								if ibc[ind].MustPause || ibc[ind].MustStop || biofabrica.Critical == scp_stopall {
									break
								}
								if scp_run_withdraw(scp_ibc, ibcid, true, false) >= 0 {
									break
								}
								if ibc[ind].Volume == 0 || ibc[ind].Withdraw == 0 {
									break
								}
								time.Sleep(2 * time.Second)
							}
						}
					} else {
						conn.Write([]byte(scp_err))
					}

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

func system_upgrade() {
	upgrademutex.Lock()
	defer upgrademutex.Unlock()
	fmt.Println("WARN SCP MASTER UPGRADEstarted... Necessario aguardar cerca de 10 minutos...")
	board_add_message("EParando Biofábrica e Atualizando Software", "")
	biofabrica.Critical = scp_stopall
	scp_emergency_pause()
	time.Sleep(60 * time.Second)
	biofabrica.Critical = scp_sysstop
	save_all_data(data_filename)
	time.Sleep(5 * time.Second)
	cmdpath, _ := filepath.Abs(execpath + "upgrade_sys.sh")
	// cmd := exec.Command(cmdpath, "restart", "scp_orch")
	cmd := exec.Command(cmdpath)
	cmd.Dir = execpath
	output, err := cmd.CombinedOutput()
	if len(output) > 0 {
		fmt.Println("OUPUT", string(output))
	}
	if err != nil {
		checkErr(err)
		fmt.Println("Falha ao Executar shell de upgrade")
		board_add_message("EFalha ao Atualizar Software", "")
		return
	}
	os.Exit(0)
}

func master_shutdown(sigs chan os.Signal) {
	<-sigs
	fmt.Println("WARN SCP MASTER Shutdown started... Necessario aguardar cerca de 60 segundos...")
	board_add_message("EDESLIGAMENTO DO SISTEMA Solicitado", "")
	biofabrica.Critical = scp_stopall
	scp_emergency_pause()
	save_all_data(data_filename)
	// for _, b := range bio {
	// 	scp_fullstop_device(b.BioreactorID, scp_bioreactor, false)
	// }
	// for _, b := range ibc {
	// 	scp_fullstop_device(b.IBCID, scp_ibc, false)
	// }
	// for _, t := range totem {
	// 	scp_fullstop_device(t.TotemID, scp_totem, false)
	// }
	scp_fullstop_device("ALL", scp_biofabrica, false, false)
	time.Sleep(60 * time.Second)
	biofabrica.Critical = scp_sysstop
	save_all_data(data_filename)
	fmt.Println("DEBUG MASTER SHUTDOWN: concluido Dados da Biofabrica =", biofabrica)
	time.Sleep(5 * time.Second)
	os.Exit(0)
}

func main() {

	rand.Seed(time.Now().UnixNano())

	fmt.Println("DEBUG MAIN: MASTER-START iniciado")

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go master_shutdown(sigs)

	localconfig_path = "/etc/scpd/"
	addrs_type = make(map[string]DevAddrData, 0)
	net192 = test_file("/etc/scpd/scp_net192.flag")
	if net192 {
		fmt.Println("WARN:  EXECUTANDO EM NET192\n\n\n")
		execpath = "/home/paulo/scp-project/"
		mainrouter = "192.168.0.1"
	} else {
		execpath = "/home/scpadm/scp-project/"
		mainrouter = "10.0.0.1"
		// mainrouter = "192.168.0.1"
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

	n_bf := load_bf_data(localconfig_path + "bf_data.csv")
	if n_bf < 1 {
		mybf.BFId = "BFIP-" + get_tun_ip()
		fmt.Println("MASTER: Arquivo contendo dados da Biofabrica nao encontrado. Usando config padrao", mybf)
	}
	norgs := load_organisms(execpath + "organismos_conf.csv")
	if norgs < 0 {
		log.Fatal("Não foi possivel ler o arquivo de organismos")
	}
	recipe_2000 = load_tasks_conf(execpath + "receita_2000_conf.csv")
	if recipe_2000 == nil {
		log.Fatal("Não foi possivel ler o arquivo contendo a receita 2000L de producao")
	}
	recipe_1000 = load_tasks_conf(execpath + "receita_1000_conf.csv")
	if recipe_1000 == nil {
		log.Fatal("Não foi possivel ler o arquivo contendo a receita 1000L de producao")
	}

	cipbio = load_tasks_conf(execpath + "cip_bio_conf.csv")
	if cipbio == nil {
		log.Fatal("Não foi possivel ler o arquivo contendo ciclo de CIP de Biorreator")
	}
	cipibc = load_tasks_conf(execpath + "cip_ibc_conf.csv")
	if cipibc == nil {
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
	npaths := load_paths_conf(execpath + "paths_conf.csv")
	// fmt.Println(npaths, "PATHS === ", paths)
	if npaths < 1 {
		log.Fatal("FATAL: Arquivo de configuracao de PATHs invalido")
	}

	valvs = make(map[string]int, 0)
	load_all_data(data_filename)

	if biofabrica.Critical == scp_ready {
		fmt.Println("ERROR MAIN: Arquivo de recuperacao estava com status READY. Biofabrica nao teve shutddown correto.")
		board_add_message("EATENÇÃO: Biofábrica não teve Desligamento correto. Possível falha de energia e nobreak. Favor contactar SAC", "ERRSHUTDOWN")
	} else {
		board_del_message("ERRSHUTDOWN")
		if biofabrica.Critical == scp_stopall {
			fmt.Println("ERROR MAIN: Biofabrica retornando com Critical=scp_stopall")
			biofabrica.Critical = scp_sysstop
		} else {
			fmt.Println("DEBUG MAIN: Biofabrica retornando OK com Critical =", biofabrica.Critical)
		}
	}

	// biofabrica.TechMode = test_file("/etc/scpd/scp_techmode.flag")
	biofabrica.Version = scp_version
	biofabrica.LastVersion = scp_null
	biofabrica.Useflowin = true

	routerok := tcp_host_isalive(mainrouter, "80", pingmax)
	for {
		if !routerok {
			board_add_message("ERoteador PRINCIPAL OFFLINE. Aguardando para iniciar Biofábrica", "ERRROUTER")
		} else {
			break
		}
		time.Sleep(10 * time.Second)
	}
	board_del_message("ERRROUTER")

	go scp_check_network() // Estava no inicio

	go scp_setup_devices(true)
	go scp_get_alldata()
	go scp_sync_functions()
	go scp_check_lastversion()

	scp_master_ipc()
	time.Sleep(10 * time.Second)
	// if !schedrunning {
	// 	go scp_scheduler()
	// }
}
