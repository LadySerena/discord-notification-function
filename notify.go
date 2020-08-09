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
	Status string      `json:"status"`
	LogURL string      `json:"logUrl"`
	Source BuildSource `json:"source"`
}

type BuildSource struct {
	RepoSource RepoSource `json:"repoSource"`
}

type RepoSource struct {
	ProjectID  string `json:"projectId"`
	RepoName   string `json:"repoName"`
	BranchName string `json:"branchName"`
}

var discordURL = os.Getenv("DISCORD_URL")

var filterError = errors.New("filtering out WORKING or QUEUED")

var filterSet = map[string]bool{
	"WORKING": true,
	"QUEUED":  true,
}

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
	log.Printf("%+v\n", pubMessage)
	discordOutput, generateErr := generateDiscordMessage(pubMessage)
	if generateErr != nil {
		if generateErr == filterError {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	discordErr := sendToDiscord(*discordOutput)
	if discordErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(discordErr.Error())
		return
	}
	w.WriteHeader(http.StatusAccepted)
	return
}

func generateDiscordMessage(pubMessage PubSubMessage) (*DiscordMessage, error) {
	internalMessage, decodeErr := base64.StdEncoding.DecodeString(pubMessage.Message.Data)
	if decodeErr != nil {
		log.Printf("could not decode string due to: %s\n", decodeErr.Error())
		return nil, decodeErr
	}
	buildData := BuildMessage{}
	unmarshalErr := json.Unmarshal([]byte(internalMessage), &buildData)
	if unmarshalErr != nil {
		log.Printf("could not unmarshal json due to: %s\n", unmarshalErr.Error())
		return nil, unmarshalErr
	}
	//if the status is in our set we return an error to ignore the message
	if _, filterOut := filterSet[buildData.Status]; filterOut {
		return nil, filterError
	}
	msg := fmt.Sprintf("build status is: %s for repo %s branch  %s view logs at: %s", buildData.Status,
		buildData.Source.RepoSource.RepoName, buildData.Source.RepoSource.BranchName, buildData.LogURL)
	return &DiscordMessage{Content: msg}, nil
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
