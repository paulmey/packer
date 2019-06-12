package client

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"testing"

	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/Azure/go-autorest/autorest/azure/auth"
	"gopkg.in/yaml.v2"
)

var _ http.RoundTripper = &recorder{}

type recorder struct {
	Interactions []Interaction
}

type Interaction struct {
	Request
	Response
}

type InteractionModifier func(Interaction) Interaction

func (m InteractionModifier) Apply(i Interaction) Interaction {
	if m == nil {
		return i
	}
	return m(i)
}

type Request struct {
	Method string
	Url    string
	Header http.Header `yaml:",omitempty"`
	Body   string      `yaml:",omitempty"`
}

type Response struct {
	Header     http.Header `yaml:",omitempty"`
	Body       string      `yaml:",omitempty"`
	StatusCode int
}

func (r *recorder) RoundTrip(req *http.Request) (*http.Response, error) {

	request, err := newRequest(req)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	responseBody := ""
	if resp.Body != nil {
		d, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		responseBody = string(d)
		resp.Body.Close()
		resp.Body = ioutil.NopCloser(bytes.NewBuffer(d))
	}

	response := Response{
		StatusCode: resp.StatusCode,
		Header:     resp.Header,
		Body:       responseBody,
	}

	i := Interaction{
		request,
		response,
	}
	r.Interactions = append(r.Interactions, i)

	return resp, nil
}

func newRequest(req *http.Request) (Request, error) {
	requestBody := ""
	if req.Body != nil {
		d, err := ioutil.ReadAll(req.Body)
		if err != nil {
			return Request{}, err
		}
		requestBody = string(d)
		req.Body.Close()
		req.Body = ioutil.NopCloser(bytes.NewBuffer(d))
	}

	return Request{
		Method: req.Method,
		Url:    req.URL.String(),
		Header: req.Header,
		Body:   requestBody,
	}, nil
}

const MockSubscriptionID = "00000000-0000-1234-0000-000000000000"

func GetTestClientSet(t *testing.T, modifiers ...InteractionModifier) (AzureClientSet, func(), error) {

	cli := azureClientSet{}

	if os.Getenv("AZURE_RECORD") == "" {
		t.Log("Test uses existing recordings.")
		replayer, err := newReplayer(t.Name(), t, modifiers)
		if err != nil {
			return nil, nil, fmt.Errorf("Error initializing replayer: %v", err)

		}
		cli.sender = &http.Client{Transport: replayer}
		cli.subscriptionID = MockSubscriptionID
		return cli, func() {}, nil
	} else {
		if os.Getenv("AZURE_CLIENT_ID") == "" ||
			os.Getenv("AZURE_CLIENT_SECRET") == "" ||
			os.Getenv("AZURE_SUBSCRIPTION_ID") == "" ||
			os.Getenv("AZURE_TENANT_ID") == "" {
			t.Fatalf("AZURE_RECORD en var set, but one of AZURE_CLIENT_ID, AZURE_CLIENT_SECRET, AZURE_SUBSCRIPTION_ID, AZURE_TENANT_ID missing")
		}
		recorder := &recorder{}
		cli.sender = &http.Client{
			Transport: recorder,
		}
		cli.PollingDelay = 0

		a, err := auth.NewAuthorizerFromEnvironment()
		if err == nil {
			cli.authorizer = a
			cli.subscriptionID = os.Getenv("AZURE_SUBSCRIPTION_ID")
		} else {
			return nil, nil, fmt.Errorf("Error creating Azure authorizer: %v", err)
		}
		return cli, func() {
			var interactions []Interaction
			for _, i := range recorder.Interactions {
				for _, m := range modifiers {
					i = m.Apply(i)
				}
				interactions = append(interactions, i)
			}
			err := save(interactions, t.Name())
			if err != nil {
				panic(err)
			}
		}, nil
	}
}

var _ http.RoundTripper = &replayer{}

type replayer struct {
	*testing.T
	interactions     []Interaction
	ipointer         int
	requestMatcher   func(actualRequest, savedRequest Request) bool
	RequestModifiers []InteractionModifier
}

func defaultRequestMatcher(a, s Request) bool {
	return a.Url == s.Url && a.Method == s.Method && a.Body == s.Body
}

func (r *replayer) RoundTrip(req *http.Request) (*http.Response, error) {
	if r.ipointer >= len(r.interactions) {
		err := fmt.Errorf("replayer: ran out of interactions to replay")
		r.Fatal(err)
		return nil, err
	}

	savedInteraction := r.interactions[r.ipointer]

	actual, err := newRequest(req)
	if err != nil {
		err := fmt.Errorf("replayer: error reading request: %v", err)
		r.Fatal(err)
		return nil, err
	}

	// modify request so that it looks like the saved one
	i := Interaction{Request: actual}
	for _, m := range r.RequestModifiers {
		i = m.Apply(i)
	}
	actual = i.Request

	matcher := r.requestMatcher
	if matcher == nil {
		matcher = defaultRequestMatcher
	}
	if !matcher(actual, savedInteraction.Request) {
		err := fmt.Errorf("replayer: request %d did not match:\nactual: %+v\nexpected: %+v",
			r.ipointer,
			actual, savedInteraction.Request)
		r.Fatal(err)
		return nil, err
	}

	response := http.Response{
		Request:       req,
		StatusCode:    savedInteraction.StatusCode,
		Body:          ioutil.NopCloser(bytes.NewBufferString(savedInteraction.Response.Body)),
		ContentLength: int64(len(savedInteraction.Response.Body)),
		Header:        savedInteraction.Response.Header,
	}
	fmt.Printf("%d %s %d %d\n%s\n", r.ipointer, req.Method, response.StatusCode, savedInteraction.StatusCode,
		savedInteraction.Response.Body)
	r.ipointer = r.ipointer + 1

	return &response, nil
}

func newReplayer(name string, t *testing.T, modifiers []InteractionModifier) (*replayer, error) {

	d, err := ioutil.ReadFile(fmt.Sprintf("fixtures/%s.yml", name))
	if err != nil {
		return nil, err
	}

	var interactions []Interaction
	err = yaml.Unmarshal(d, &interactions)
	if err != nil {
		return nil, err
	}

	return &replayer{
		interactions:     interactions,
		T:                t,
		requestMatcher:   defaultRequestMatcher,
		RequestModifiers: modifiers,
	}, nil
}

func save(interactions []Interaction, name string) error {
	d, err := yaml.Marshal(interactions)
	if err != nil {
		return err
	}
	err = os.MkdirAll("fixtures", 0777)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(fmt.Sprintf("fixtures/%s.yml", name), d, 0666)
}

func filterRequestHeaders(interactions []Interaction) []Interaction {
	return interactions
}

func FindAzureErrorService(err error) *azure.ServiceError {
	switch e := err.(type) {
	case autorest.DetailedError:
		if e.Original != nil {
			return FindAzureErrorService(e.Original)
		}
		return nil
	case azure.RequestError:
		if e.Original != nil {
			return FindAzureErrorService(e.Original)
		}
		if e.ServiceError != nil {
			return e.ServiceError
		}
		return nil
	default:
		return nil
	}
}
