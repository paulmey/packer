package httpmock

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"

	"github.com/Azure/go-autorest/autorest"
)

type Sender struct {
	Mocks []Mock
}

type Mock struct {
	MethodRegex string
	URIRegex    string
	Sender      autorest.Sender
}

func Get(uriRegex string, response interface{}) Mock {
	return Mock{
		http.MethodGet,
		uriRegex,
		NewMockSender(response)}
}

func (m Mock) Matches(req *http.Request) bool {
	matched, err := regexp.MatchString(m.MethodRegex, req.Method)
	if err != nil {
		log.Fatalf("Error matching mock method: %v", m)
	}
	if !matched {
		return false
	}

	matched, err = regexp.MatchString(m.URIRegex, req.URL.String())
	if err != nil {
		log.Fatalf("Error matching mock uri: %v", m)
	}
	return matched
}

func (s *Sender) Do(req *http.Request) (*http.Response, error) {
	if s != nil {
		for _, mock := range s.Mocks {
			if mock.Matches(req) {
				return mock.Sender.Do(req)
			}
		}
	}
	log.Fatalf("UNHANDLED HTTP TRAFFIC: %s %s", req.Method, req.URL)
	return nil, errors.New("HTTP traffic not allowed in tests")
}

func NewMockSender(response interface{}) autorest.Sender {
	var data []byte
	if d, ok := response.([]byte); ok {
		data = d
	} else if s, ok := response.(string); ok {
		data = []byte(s)
	} else {
		d, err := json.Marshal(response)
		if err != nil {
			log.Fatalf("Error setting up mock: %v", err)
		}
		data = d
	}

	return autorest.SenderFunc(
		func(req *http.Request) (*http.Response, error) {
			resp := &http.Response{}
			resp.StatusCode = 200
			resp.Status = "200 OK"
			resp.Body = ioutil.NopCloser(bytes.NewReader(data))
			resp.ContentLength = int64(len(data))
			return resp, nil
		})
}
