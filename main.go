package main

import (
	it "blueblue/init"
	rabbitMQ "blueblue/rabbitMQ"
	"context"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/xiafei571/ble"
	"github.com/xiafei571/ble/linux"
)

var dur *time.Duration
var dir *string
var port *int
var logger *log.Logger
var stop bool = true
var mq = connect_MQ()
var sensorName = "sensor_01"

// Device represents a BLE device
type Device struct {
	Sensor        string    `json:"sensor"`
	Address       string    `json:"address"`
	Detected      time.Time `json:"detected"`
	Since         string    `json:"since"`
	Name          string    `json:"name"`
	RSSI          int       `json:"rssi"`
	Advertisement string    `json:"advertisement"`
	ScanResponse  string    `json:"scanresponse"`
	EventType     string    `json:"eventtype"`
}

var mutex sync.RWMutex
var devices map[string]Device

func init() {
	devices = make(map[string]Device)
	mutex = sync.RWMutex{}
	d, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal("Can't get running directory:", err)
	}
	dir = flag.String("dir", d, "directory where the public directory is in")
	dur = flag.Duration("d", 5*time.Second, "Scan duration")
	port = flag.Int("p", 23232, "the port where the server starts")
	flag.Parse()
}

func main() {
	f, err := os.OpenFile("blueblue.log",
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
	}
	defer f.Close()
	logger = log.New(f, "", log.LstdFlags)

	d, err := linux.NewDevice()
	if err != nil {
		logger.Fatal("Can't create new device:", err)
	}
	ble.SetDefaultDevice(d)
	// serve()
	scanBackend()
}

// Handle the advertisement scan
func adScanHandler(a ble.Advertisement) {
	mutex.Lock()
	str_adv := hex.EncodeToString(a.LEAdvertisingReportRaw())

	device := Device{
		Sensor:        sensorName,
		Address:       a.Addr().String(),
		Detected:      time.Now(),
		Name:          clean(a.LocalName()),
		RSSI:          a.RSSI(),
		Advertisement: formatHex(str_adv),
		ScanResponse:  formatHex(hex.EncodeToString(a.ScanResponseRaw())),
		EventType:     formatEventType(str_adv[4:6]),
	}
	devices[a.Addr().String()] = device

	logger.Println("device addr:" + a.Addr().String())
	mqPublish(device)

	mutex.Unlock()

}

func adScanAndPublishHandler(a ble.Advertisement) {
	mutex.Lock()
	str_adv := hex.EncodeToString(a.LEAdvertisingReportRaw())

	device := Device{
		Sensor:        sensorName,
		Address:       a.Addr().String(),
		Detected:      time.Now(),
		Name:          clean(a.LocalName()),
		RSSI:          a.RSSI(),
		Advertisement: formatHex(str_adv),
		ScanResponse:  formatHex(hex.EncodeToString(a.ScanResponseRaw())),
		EventType:     formatEventType(str_adv[4:6]),
	}

	logger.Println("device addr:" + a.Addr().String())
	mqPublish(device)

	mutex.Unlock()

}

// publish to mq
func mqPublish(device Device) {
	data, err := json.Marshal(device)
	if err != nil {
		logger.Println(err)
	}
	mq.PublishByte(data)
}

// start the web server
func serve() {
	mux := http.NewServeMux()
	mux.Handle("/public/", http.StripPrefix("/public/", http.FileServer(http.Dir(*dir+"/public"))))
	mux.HandleFunc("/", index)
	mux.HandleFunc("/stop", stopScan)
	mux.HandleFunc("/start", startScan)
	mux.HandleFunc("/devices", showDevices)
	server := &http.Server{
		Addr:    "127.0.0.1:" + strconv.Itoa(*port),
		Handler: mux,
	}
	fmt.Println("Started blueblue server at", server.Addr)
	server.ListenAndServe()
}

// index for web server
func index(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles(*dir + "/public/index.html")
	if err != nil {
		fmt.Println("error: ", err)
		fmt.Println("index for web server", *dir+"/public/index.html")
	}
	t.Execute(w, stop)
}

// handler to show list of devices
func showDevices(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles(*dir + "/public/devices.html")

	if err != nil {
		fmt.Println("error: ", err)
		fmt.Println("index for web server", *dir+"/public/devices.html")
	}
	// convert map to array, added detect since duration and
	// remove anything that's more than 60 seconds
	data := []Device{}
	for _, device := range devices {
		device.Since = strconv.Itoa(int(time.Since(device.Detected).Seconds()))
		tn := time.Now().Add(-1 * time.Duration(60) * time.Second)
		if tn.Before(device.Detected) {
			data = append(data, device)
		}
	}
	// sort by RSSI
	sort.SliceStable(data, func(i, j int) bool {
		return data[i].RSSI > data[j].RSSI
	})
	t.Execute(w, data)
}

// handler to start scanning
func startScan(w http.ResponseWriter, r *http.Request) {
	if !stop {
		w.WriteHeader(409)
	} else {
		go scan()
	}
}

// handler to stop scanning
func stopScan(w http.ResponseWriter, r *http.Request) {
	if stop {
		w.WriteHeader(409)
	} else {
		stop = true
	}
}

// scan goroutine
func scan() {
	stop = false
	logger.Println("Started scanning every", *dur)
	for !stop {
		ctx := ble.WithSigHandler(context.WithTimeout(context.Background(), *dur))
		ble.Scan(ctx, false, adScanHandler, nil)
		time.Sleep(30 * time.Second)
	}
	logger.Println("Stopped scanning.")
	stop = true
}

func scanBackend() {
	logger.Println("Started scanning every", *dur)
	logger.Println("Ctrl + c to stop")
	for true {
		ctx := ble.WithSigHandler(context.WithTimeout(context.Background(), *dur))
		ble.Scan(ctx, false, adScanAndPublishHandler, nil)
		time.Sleep(30 * time.Second)
	}
}

// reformat string for proper display of hex
func formatHex(instr string) (outstr string) {
	outstr = ""
	for i := range instr {
		if i%2 == 0 {
			outstr += instr[i:i+2] + " "
		}
	}
	return
}

func formatEventType(instr string) (outstr string) {
	if instr == "00" {
		outstr = "ADV_IND"
	} else if instr == "01" {
		outstr = "ADV_DIRECT_IND"
	} else if instr == "02" {
		outstr = "ADV_NONCONN_IND"
	} else if instr == "03" {
		outstr = "SCAN_REQ"
	} else if instr == "10" {
		outstr = "SCAN_RSP"
	} else if instr == "11" {
		outstr = "CONNECT_REQ"
	} else if instr == "12" {
		outstr = "ADV_SCAN_IND"
	} else {
		outstr = "UNKNOWN"
	}
	return
}

// clean up the non-ASCII characters
func clean(input string) string {
	return strings.TrimFunc(input, func(r rune) bool {
		return !unicode.IsGraphic(r)
	})
}

func connect_MQ() *rabbitMQ.RabbitMQ {
	config := it.InitConfigure()
	user := config.GetString("MQ.User")
	password := config.GetString("MQ.Password")
	ip := config.GetString("MQ.IP")
	vhost := config.GetString("MQ.Vhost")
	MQURL := "amqp://" + user + ":" + password + "@" + ip + "/" + vhost
	fmt.Println(MQURL)

	rabbitmq := rabbitMQ.NewRabbitMQSimple(MQURL, ""+"blue")
	return rabbitmq
}
