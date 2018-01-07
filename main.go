package main

import (
	"encoding/json"
	"flag"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/atiernan/smartHomeSamsungTVCommon"
	"github.com/chbmuc/cec"
)

func main() {
	tvIP := flag.String("tv-ip", "127.0.0.1", "The IP address of the TV to control")
	serverURL := flag.String("server", "http://example.com", "The URL to the webserver")
	flag.Parse()

	if *tvIP == "127.0.0.1" {
		log.Println("A host IP must be specified")
		return
	}

	if *serverURL == "http://example.com" {
		log.Println("A server URL must be specified")
		return
	}

	c, err := cec.Open("", "cec.go")
	if err != nil {
		log.Println(err)
		return
	}

	tv := SamsungTV{
		Host:            *tvIP,
		ApplicationID:   "google-home-samsung",
		ApplicationName: "Google Samsung Remote",
	}

	for {
		time.Sleep(1 * time.Second)
		response, err := http.Get(*serverURL)
		if err != nil {
			log.Println("Failed to fetch latest status")
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
				c.PowerOn(0)
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
		}
		response.Body.Close()
	}
}
