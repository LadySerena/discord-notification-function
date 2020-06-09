package discord_notification_function

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

type PubSubMessage struct {
	Message      MessageBlock `json:"message"`
	Subscription string       `json:"subscription"`
}

type MessageBlock struct {
	Attributes AttributesBlock `json:"attributes"`
	Data       string          `json:"data"`
	MessageID  string          `json:"message_id"`
}

type AttributesBlock struct {
	BuildID string `json:"buildID"`
	Status  string `json:"status"`
}

type DiscordMessage struct {
	Content string `json:"content"`
}

type BuildMessage struct {
	Status string `json:"status"`
	LogURL string `json:"logUrl"`
}

var discordURL = os.Getenv("DISCORD_URL")

func GetBuildMessage(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	bodyBytes, readErr := ioutil.ReadAll(r.Body)
	if readErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	pubMessage := PubSubMessage{}
	unmarshalErr := json.Unmarshal(bodyBytes, &pubMessage)
	if unmarshalErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("could not unmarshal json due to: %s\n", unmarshalErr.Error())
		return
	}
	discordOutput := generateDiscordMessage(pubMessage)
	if discordOutput == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	discordErr := sendToDiscord(*discordOutput)
	if discordErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(discordErr.Error())
		return
	}
	w.WriteHeader(http.StatusOK)
	return
}

func generateDiscordMessage(pubMessage PubSubMessage) *DiscordMessage {
	internalMessage, decodeErr := base64.StdEncoding.DecodeString(pubMessage.Message.Data)
	if decodeErr != nil {
		log.Printf("could not decode string due to: %s\n", decodeErr.Error())
		return nil
	}
	buildData := BuildMessage{}
	unmarshalErr := json.Unmarshal([]byte(internalMessage), &buildData)
	if unmarshalErr != nil {
		log.Printf("could not unmarshal json due to: %s\n", unmarshalErr.Error())
		return nil
	}
	msg := fmt.Sprintf("build status is: %s, view logs at: %s", buildData.Status, buildData.LogURL)
	return &DiscordMessage{Content: msg}
}

func sendToDiscord(data DiscordMessage) error {

	messageContent, marshalErr := json.Marshal(data)
	if marshalErr != nil {
		return marshalErr
	}
	log.Println(string(messageContent))
	response, postErr := http.DefaultClient.Post(discordURL, "application/json", bytes.NewBuffer(messageContent))
	if postErr != nil {
		return postErr
	}
	if response.StatusCode != http.StatusNoContent {
		errorMessage := fmt.Sprintf("api didn't return expected status, received %d, but expected %d", response.StatusCode, http.StatusNoContent)
		return errors.New(errorMessage)
	}
	return nil
}


