package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
)

type Push struct {
	Token   string `json:"token"`
	User    string `json:"user"`
	Message string `json:"message"`
}

const PushOverAPI = "https://api.pushover.net/1/messages.json"

var PushClient = &http.Client{}

// send message to PushOver API
// TODO: store settings in a config file
func PushOver(msg string) {

	var push Push

	push.Token = ""
	push.User = ""
	push.Message = msg

	if push.Token == "" || push.User == "" {
		return
	}

	data, err := json.Marshal(&push)
	if err != nil {
		log.Println("PushOver:", err)
		return
	}

	_, err = PushClient.Post(PushOverAPI, "application/json", bytes.NewBuffer(data))
	if err != nil {
		log.Println("PushOver:", err)
		return
	}
}
