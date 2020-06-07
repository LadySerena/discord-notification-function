package notify

import (
	"io/ioutil"
	"net/http"
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

func PostToDiscord(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	bodyBytes, readErr := ioutil.ReadAll(r.Body)
	if readErr != nil {
		
	}
}
