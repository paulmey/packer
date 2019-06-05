package client

import (
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/Azure/go-autorest/autorest/azure/auth"
	"github.com/dnaeon/go-vcr/cassette"
	"github.com/dnaeon/go-vcr/recorder"
)

func GetTestClientSet(t *testing.T) (AzureClientSet, func() error, error) {
	replayOnly :=
		os.Getenv("AZURE_CLIENT_ID") == "" ||
			os.Getenv("AZURE_CLIENT_SECRET") == "" ||
			os.Getenv("AZURE_SUBSCRIPTION_ID") == "" ||
			os.Getenv("AZURE_TENANT_ID") == ""

	const mockSubscriptionID = "00000000-0000-1234-0000-000000000000"

	r, err := recorder.New(fmt.Sprintf("fixtures/%s", t.Name()))
	if err != nil {
		return nil, nil, fmt.Errorf("Error setting up VCR: %v", err)
	}
	r.AddFilter(func(i *cassette.Interaction) error {
		delete(i.Request.Headers, "Authorization")
		delete(i.Request.Headers, "User-Agent")

		delete(i.Response.Headers, "Cache-Control")
		delete(i.Response.Headers, "Date")
		delete(i.Response.Headers, "Expires")
		delete(i.Response.Headers, "Pragma")
		delete(i.Response.Headers, "Server")
		delete(i.Response.Headers, "Strict-Transport-Security")
		delete(i.Response.Headers, "Vary")
		delete(i.Response.Headers, "X-Content-Type-Options")
		delete(i.Response.Headers, "X-Ms-Correlation-Request-Id")
		delete(i.Response.Headers, "X-Ms-Ratelimit-Remaining-Resource")
		delete(i.Response.Headers, "X-Ms-Ratelimit-Remaining-Subscription-Reads")
		delete(i.Response.Headers, "X-Ms-Request-Id")
		delete(i.Response.Headers, "X-Ms-Routing-Request-Id")
		delete(i.Response.Headers, "X-Ms-Served-By")

		return nil
	})

	//cfg := &govcr.VCRConfig{
	//	DisableRecording: true,
	//	RequestFilters: govcr.RequestFilters{
	//		govcr.RequestDeleteHeaderKeys("Authorization", "User-Agent"),
	//		govcr.RequestFilter(func(req govcr.Request) govcr.Request {
	//			req.URL.Path = subscriptionPathRegex.ReplaceAllLiteralString(req.URL.Path,
	//				"/subscriptions/00000000-0000-1234-0000-000000000000")
	//			return req
	//		}).OnPath(`/management.azure.com/`),
	//	},
	//	ResponseFilters: govcr.ResponseFilters{
	//		govcr.ResponseDeleteHeaderKeys(
	//			"Date", "Server", "Cache-Control", "Expires", "Pragma", "Strict-Transport-Security", "Vary",
	//			"X-Content-Type-Options",
	//			"X-Ms-Correlation-Request-Id",
	//			"X-Ms-Ratelimit-Remaining-Resource",
	//			"X-Ms-Ratelimit-Remaining-Subscription-Reads",
	//			"X-Ms-Request-Id",
	//			"X-Ms-Routing-Request-Id"),
	//	},
	//}
	cli := azureClientSet{}

	if replayOnly {
		t.Log("Azure credentials not available, will use existing recordings.")
		cli.subscriptionID = mockSubscriptionID
	} else {
		if os.Getenv("AZURE_RECORD") == "" {
			r.SetTransport(nil)
		}
		a, err := auth.NewAuthorizerFromEnvironment()
		if err == nil {
			cli.authorizer = a
			cli.subscriptionID = os.Getenv("AZURE_SUBSCRIPTION_ID")
			cli.PollingDelay = 0
		} else {
			r.Stop()
			return nil, nil, fmt.Errorf("Error creating Azure authorizer: %v", err)
		}
	}

	cli.sender = &http.Client{Transport: r}

	return cli, r.Stop, nil
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
