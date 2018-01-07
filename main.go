package main

import (
	"encoding/json"
	"flag"
	"io"
	"io/ioutil"
	"log"
	"net/http"
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

func readConfig(path string) configData {
	data, _ := ioutil.ReadFile(path)
	config := configData{}
	json.Unmarshal(data, &config)
	return config
}

func main() {
	configFilePath := flag.String("config", "/dev/null", "The config file to use")
	flag.Parse()

	config := readConfig(*configFilePath)

	tv := SamsungTV{
		Host:            config.TVs[0].Host,
		ApplicationID:   "google-home-samsung",
		ApplicationName: "Google Samsung Remote",
	}

	for {
		time.Sleep(1 * time.Second)
		response, err := http.Get(config.ServerURL + "device/endpoint")
		if err != nil {
			log.Println("Failed to fetch latest status from \"" + config.ServerURL + "device/endpoint\"")
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
			if message.TVSwitchedOn {
				if tv.Connect() {
					if !tv.SendCommand("KEY_POWER") {
						log.Fatal("Failed to send command")
					}
					tv.Close()
				} else {
					log.Println("Failed to send power command, sending WoL packet")
					wol.MagicWake(config.TVs[0].MACAddress, "255.255.255.255")
				}
			}
			if message.TVSwitchedOff {
				if tv.Connect() {
					tv.SendCommand("KEY_POWER")
					tv.Close()
				}
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
				if tv.Connect() {
					tv.SendCommand("KEY_MUTE")
					tv.Close()
				}
			}
			if message.Pause {
				if tv.Connect() {
					tv.SendCommand("KEY_PAUSE")
					tv.Close()
				}
			}
			if message.Play {
				if tv.Connect() {
					tv.SendCommand("KEY_PLAY")
					tv.Close()
				}
			}
			if message.OK {
				if tv.Connect() {
					tv.SendCommand("KEY_OK")
					tv.Close()
				}
			}
		}
		response.Body.Close()
	}
}
