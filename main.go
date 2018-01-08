package main

import (
	"encoding/json"
	"flag"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/atiernan/smartHomeSamsungTVCommon"
	"github.com/ghthor/gowol"
)

type tv struct {
	MACAddress string
	ID         string
	Host       string
}
type configData struct {
	TVs       []tv
	ServerURL string
}

type tvEndpointURI struct {
	URI string `json:"uri"`
}
type tvEndpoints struct {
	V1 tvEndpointURI `json:"v1"`
	V2 tvEndpointURI `json:"v2"`
}
type tvSimpleInformation struct {
	Data   tvEndpoints `json:"data"`
	Remote string      `json:"remote"`
	ID     string      `json:"sid"`
	TTL    uint16      `json:"ttl"`
	Type   string      `json:"type"`
}
type tvDetailedData struct {
	FrameTVSupport    bool   `json:"FrameTVSupport"`
	GamePadSupport    bool   `json:"GamePadSupport"`
	ImeSyncedSupport  bool   `json:"ImeSyncedSupport"`
	OS                string `json:"OS"`
	VoiceSupport      bool   `json:"VoiceSupport"`
	CountryCode       string `json:"countryCode"`
	Description       string `json:"description"`
	DeveloperIP       string `json:"developerIP"`
	DeveloperMode     string `json:"developerMode"`
	DUID              string `json:"duid"`
	FirmwareVersion   string `json:"firmwareVersion"`
	ID                string `json:"id"`
	IP                string `json:"ip"`
	Model             string `json:"model"`
	ModelName         string `json:"modelName"`
	Name              string `json:"name"`
	NetworkType       string `json:"networkType"`
	Resolution        string `json:"resolution"`
	SmartHubAgreement string `json:"smartHubAgreement"`
	SSID              string `json:"ssid"`
	Type              string `json:"type"`
	UDN               string `json:"udn"`
	WifiMac           string `json:"wifiMac"`
}
type tvDetailedInformation struct {
	Device    tvDetailedData `json:"device"`
	ID        string         `json:"id"`
	IsSupport string         `json:"isSupport"`
	Name      string         `json:"name"`
	Remote    string         `json:"remote"`
	Type      string         `json:"type"`
	URI       string         `json:"uri"`
	Version   string         `json:"version"`
}

func readConfig(path string) configData {
	data, _ := ioutil.ReadFile(path)
	config := configData{}
	json.Unmarshal(data, &config)
	return config
}

func listenForTVs(callback func(tvSimpleInformation)) {
	// Set up listening socket
	readBufferSize := 8192
	addr, err := net.ResolveUDPAddr("udp", "224.0.0.7:8001")
	if err != nil {
		log.Fatal(err)
	}
	conn, err := net.ListenMulticastUDP("udp", nil, addr)
	conn.SetReadBuffer(readBufferSize)

	b := make([]byte, readBufferSize)
	for {
		n, _, err := conn.ReadFromUDP(b)
		if err != nil {
			log.Fatal(err)
		} else {
			tvInformation := tvSimpleInformation{}
			err := json.Unmarshal(b[:n], &tvInformation)
			if err != nil {
				log.Fatal(err)
			} else {
				callback(tvInformation)
			}
		}
	}
}

func tvController(tvData tv, IPChannel chan string, serverURL string) {
	currentIP := "0.0.0.0"

	for {
		time.Sleep(1 * time.Second)

		select {
		case ip := <-IPChannel:
			if ip != currentIP {
				log.Printf("%s is at %s", tvData.ID, ip)
			}
			currentIP = ip
		default:
		}

		response, err := http.Get(serverURL + "device/endpoint?deviceID=" + tvData.ID)
		if err != nil {
			log.Println("Failed to fetch latest status from \"" + serverURL + "device/endpoint\"")
			continue
		}

		decoder := json.NewDecoder(response.Body)
		var message smartHomeSamsungTVCommon.DeviceEndpointResponse
		if err := decoder.Decode(&message); err == io.EOF {
			break
		}

		if err != nil {
			log.Fatal(err)

		} else {
			tv := SamsungTV{
				Host:            currentIP,
				ApplicationID:   "google-home-samsung",
				ApplicationName: "Google Samsung Remote",
			}

			if message.TVSwitchedOn {
				if currentIP != "0.0.0.0" {
					if !tv.SendSingleCommand("KEY_POWER") {
						wol.MagicWake(tvData.MACAddress, "255.255.255.255")
					}
				} else {
					wol.MagicWake(tvData.MACAddress, "255.255.255.255")
				}
			}

			if currentIP != "0.0.0.0" {
				if message.TVSwitchedOff {
					tv.SendSingleCommand("KEY_POWER")
				}
				if message.VolumeUp > 0 {
					if tv.Connect() {
						for i := 0; i < message.VolumeUp; i++ {
							tv.SendCommand("KEY_VOLUP")
							time.Sleep(500 * time.Millisecond)
						}
						tv.Close()
					}
				}
				if message.VolumeDown > 0 {
					if tv.Connect() {
						for i := 0; i < message.VolumeDown; i++ {
							tv.SendCommand("KEY_VOLDOWN")
							time.Sleep(500 * time.Millisecond)
						}
						tv.Close()
					}
				}
				if message.VolumeMute {
					tv.SendSingleCommand("KEY_MUTE")
				}
				if message.Pause {
					tv.SendSingleCommand("KEY_PAUSE")
				}
				if message.Play {
					tv.SendSingleCommand("KEY_PLAY")
				}
				if message.OK {
					tv.SendSingleCommand("KEY_OK")
				}
			}
		}
		response.Body.Close()
	}
	log.Fatal("TV Controller finished, controllers are supposed to run indefinetley")
}

func searchTVs() {
	log.Println("Searching for TVs")
	log.Println("Please make sure the TV you are searching for is switched on...")

	var tvIDtoMac map[string]string
	tvIDtoMac = make(map[string]string)
	callback := func(tv tvSimpleInformation) {
		if _, ok := tvIDtoMac[tv.ID]; !ok {
			resp, _ := http.Get(tv.Data.V2.URI)
			msg := tvDetailedInformation{}
			decoder := json.NewDecoder(resp.Body)
			decoder.Decode(&msg)
			resp.Body.Close()

			log.Printf("Found \"%s\"\r\n", msg.Name)
			log.Printf(" - ID: %s", msg.ID)
			log.Printf(" - MAC: %s", msg.Device.WifiMac)
			tvIDtoMac[tv.ID] = msg.Device.WifiMac
		}
	}
	listenForTVs(callback)
}

func main() {
	if len(os.Args) > 1 {
		if os.Args[1] == "search" {
			searchTVs()
		} else {
			configFilePath := flag.String("config", "/dev/null", "The config file to use")
			flag.Parse()
			config := readConfig(*configFilePath)

			var tvIDtoIPMap map[string]chan string
			tvIDtoIPMap = make(map[string]chan string)

			for _, tv := range config.TVs {
				log.Printf("Creating goroutine for %s", tv.ID)
				tvIDtoIPMap[tv.ID] = make(chan string)
				go tvController(tv, tvIDtoIPMap[tv.ID], config.ServerURL)
			}

			callback := func(tvInformation tvSimpleInformation) {
				u, _ := url.Parse(tvInformation.Data.V2.URI)
				tvIDtoIPMap[tvInformation.ID] <- u.Hostname()
			}
			go listenForTVs(callback)

			for {
			}
		}
	}
}
