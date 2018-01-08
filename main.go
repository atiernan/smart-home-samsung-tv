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

type tvConnectionDataURI struct {
	URI string `json:"uri"`
}
type tvConnectionData struct {
	V1 tvConnectionDataURI `json:"v1"`
	V2 tvConnectionDataURI `json:"v2"`
}
type tvConnectionInformation struct {
	Data   tvConnectionData `json:"data"`
	Remote string           `json:"remote"`
	ID     string           `json:"sid"`
	TTL    uint16           `json:"ttl"`
	Type   string           `json:"type"`
}

func readConfig(path string) configData {
	data, _ := ioutil.ReadFile(path)
	config := configData{}
	json.Unmarshal(data, &config)
	return config
}

func listenForTVs(channels map[string]chan string) {
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
			tvInformation := tvConnectionInformation{}
			err := json.Unmarshal(b[:n], &tvInformation)
			if err != nil {
				log.Fatal(err)
			} else {
				u, _ := url.Parse(tvInformation.Data.V2.URI)
				channels[tvInformation.ID] <- u.Hostname()
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

func main() {
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

	go listenForTVs(tvIDtoIPMap)

	for {
	}
}
