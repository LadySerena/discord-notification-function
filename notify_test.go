package notify

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)
const testBuildMessage = `{
  "message": {
    "attributes": {
      "buildId": "abcd-efgh...",
      "status": "SUCCESS"
    },
    "data": "SGVsbG8gQ2xvdWQgUHViL1N1YiEgSGVyZSBpcyBteSBtZXNzYWdlIQ==",
    "message_id": "136969346945"
  },
  "subscription": "projects/myproject/subscriptions/mysubscription"
}`

func TestGetBuildMessage(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(testBuildMessage))
	request.Header.Add("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	GetBuildMessage(rr, request)

}