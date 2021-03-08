package discord_notification_function

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	secretmanagerpb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1"

	"google.golang.org/genproto/googleapis/devtools/cloudbuild/v1"
	"google.golang.org/protobuf/encoding/protojson"
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

var (
	filterError = errors.New("filtering out WORKING or QUEUED")
	filterSet   = map[string]bool{
		"WORKING": true,
		"QUEUED":  true,
	}
	webhookSecret = os.Getenv("WEBHOOK_SECRET_NAME")
)

func init() {
	ctx := context.Background()
	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		log.Fatalf("failed to setup client: %v", err)
	}
	req := &secretmanagerpb.GetSecretVersionRequest{
		Name: webhookSecret,
	}
	resp, getSecretErr := client.GetSecret(ctx, req)
	if getSecretErr != nil{
		log.Fatalf("could not get secret due to: %s", getSecretErr.Error())
	}
	resp.get
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
	log.Printf("%+v", pubMessage)
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
}

func generateDiscordMessage(pubMessage PubSubMessage) (*DiscordMessage, error) {
	internalMessage, decodeErr := base64.StdEncoding.DecodeString(pubMessage.Message.Data)
	if decodeErr != nil {
		log.Printf("could not decode string due to: %s\n", decodeErr.Error())
		return nil, decodeErr
	}
	log.Printf("%s", string(internalMessage))
	// logic lifted from https://github.com/GoogleCloudPlatform/cloud-build-notifiers/blob/df590d6e6838bacd8c340c0e3ce4b6409c837e5f/lib/notifiers/notifiers.go#L425-L445
	buildData := new(cloudbuild.Build)
	uo := protojson.UnmarshalOptions{
		AllowPartial:   true,
		DiscardUnknown: true,
	}
	if err := uo.Unmarshal(internalMessage, buildData); err != nil {
		log.Printf("could not unmarhsal build due to: %s\n", err.Error())
	}
	// if the status is in our set we return an error to ignore the message
	if _, filterOut := filterSet[buildData.Status.String()]; filterOut {
		return nil, filterError
	}
	msg := fmt.Sprintf("build status is: %s for repo %s branch  %s view logs at: %s", buildData.Status,
		buildData.Source.GetRepoSource().RepoName, buildData.Source.GetRepoSource().GetBranchName(), buildData.LogUrl)
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
