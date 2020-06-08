package discord_notification_function

import (
	"bytes"
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

var discordURL = os.Getenv("DISCORD_URL")

func GetBuildMessage(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	bodyBytes, readErr := ioutil.ReadAll(r.Body)
	if readErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	discordErr := sendToDiscord(bodyBytes)
	if discordErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(discordErr.Error())
		return
	}
	w.WriteHeader(http.StatusOK)
	return
}

func sendToDiscord(data []byte) error {
	message := DiscordMessage{
		Content: string(data),
	}
	messageContent, marshalErr := json.Marshal(message)
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


