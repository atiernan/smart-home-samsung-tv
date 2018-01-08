package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"log"
	"net"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
)

var portNumber = 55000
var webSocketPort = 8001

// SamsungTV contains the information to connect to a Samsung TV
type SamsungTV struct {
	Host            string
	ApplicationID   string
	ApplicationName string
	legacySocket    net.Conn
	webSocket       *websocket.Conn
	connectedLegacy bool
	connectedWS     bool
}

// Connect to a Samsung TV at the given IP address
func (tv *SamsungTV) Connect() bool {
	tv.connectedLegacy = false
	tv.connectedWS = false

	u := url.URL{
		Scheme:  "ws",
		Host:    fmt.Sprintf("%s:%d", tv.Host, webSocketPort),
		Path:    "/api/v2/channels/samsung.remote.control",
		RawPath: "name=" + tv.ApplicationID,
	}

	var err error
	dialer := websocket.Dialer{}
	dialer.HandshakeTimeout = time.Second
	tv.webSocket, _, err = dialer.Dial(u.String(), nil)
	if err != nil {
		tv.legacySocket, err = net.DialTimeout("tcp", fmt.Sprintf("%s:%d", tv.Host, portNumber), time.Second)
		if err == nil {
			tv.connectedLegacy = true
		} else {
			fmt.Println("Failed to connect to TV")
		}
	} else {
		_, _, err := tv.webSocket.ReadMessage()
		if err != nil {
			log.Println("Error reading initial websocket message:", err)
		}
		tv.connectedWS = true
	}

	return (tv.connectedLegacy || tv.connectedWS)
}

// Close the connection to the TV
func (tv *SamsungTV) Close() {
	if tv.connectedLegacy {
		tv.legacySocket.Close()
		tv.connectedLegacy = false
	} else if tv.connectedWS {
		tv.webSocket.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		tv.webSocket.Close()
		tv.connectedWS = false
	}
}

func (tv SamsungTV) sendLegacyCommand(command string) bool {
	header := bytes.Buffer{}
	header.WriteByte(0x64)
	header.WriteByte(0x00)
	appendBase64(tv.Host, &header)
	appendBase64(tv.ApplicationID, &header)
	appendBase64(tv.ApplicationName, &header)

	tv.legacySocket.Write(wrapMessage(&header))

	cmd := bytes.Buffer{}
	cmd.Write([]byte{0x00, 0x00, 0x00})
	appendBase64(command, &cmd)

	tv.legacySocket.Write(wrapMessage(&cmd))

	response := make([]byte, 64)
	length, err := tv.legacySocket.Read(response)
	if err != nil {
		return false
	}
	if length > 0 {
		return true
	}

	return true
}

// sendWSCommand sends a command to a TV using a websocket
func (tv SamsungTV) sendWSCommand(command string) bool {
	type paramsStruct struct {
		Cmd          string
		DataOfCmd    string
		Option       bool
		TypeOfRemote string
	}

	type payLoad struct {
		Method string       `json:"method"`
		Params paramsStruct `json:"params"`
	}

	var msg = &payLoad{
		Method: "ms.remote.control",
		Params: paramsStruct{
			Cmd:          "Click",
			DataOfCmd:    command,
			Option:       false,
			TypeOfRemote: "SendRemoteKey",
		},
	}

	err := tv.webSocket.WriteJSON(msg)
	return (err == nil)
}

// SendCommand sends a command to the TV
// Requires Connect to be called first
func (tv SamsungTV) SendCommand(command string) bool {
	if tv.connectedLegacy {
		return tv.sendLegacyCommand(command)
	} else if tv.connectedWS {
		return tv.sendWSCommand(command)
	} else {
		fmt.Println("Not connected to TV")
	}
	return false
}

// SendSingleCommand connects to the TV, sends the command, and closes the connection
func (tv SamsungTV) SendSingleCommand(command string) bool {
	var result = false
	if tv.Connect() {
		if tv.SendCommand(command) {
			result = true
		}
		tv.Close()
	}
	return result
}

func wrapMessage(msg *bytes.Buffer) []byte {
	wrapped := bytes.Buffer{}
	appName := "iphone..iapp.samsung"

	wrapped.WriteByte(0x00)
	wrapped.WriteByte(uint8(len(appName)))
	wrapped.WriteByte(0x00)
	wrapped.Write([]byte(appName))
	wrapped.WriteByte(uint8(msg.Len()))
	wrapped.WriteByte(0x00)
	wrapped.Write(msg.Bytes())
	return wrapped.Bytes()
}

func appendBase64(value string, msg *bytes.Buffer) {
	encodedValue := []byte(base64.StdEncoding.EncodeToString([]byte(value)))
	msg.WriteByte(uint8(len(encodedValue)))
	msg.WriteByte(0x00)
	msg.Write(encodedValue)
}
