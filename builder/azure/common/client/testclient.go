package client

import (
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/Azure/go-autorest/autorest/azure/auth"
	"github.com/seborama/govcr"
	"os"
	"testing"
)

func GetTestClientSet(t *testing.T) (AzureClientSet, error) {
	replayOnly :=
		os.Getenv("AZURE_CLIENT_ID") == "" ||
			os.Getenv("AZURE_CLIENT_SECRET") == "" ||
			os.Getenv("AZURE_SUBSCRIPTION_ID") == "" ||
			os.Getenv("AZURE_TENANT_ID") == ""

	cfg := &govcr.VCRConfig{
		DisableRecording: true,
		RequestFilters: govcr.RequestFilters{
			govcr.RequestDeleteHeaderKeys("Authorization", "User-Agent"),
			govcr.RequestFilter(func(req govcr.Request) govcr.Request {
				req.URL.Path = subscriptionPathRegex.ReplaceAllLiteralString(req.URL.Path,
					"/subscriptions/00000000-0000-1234-0000-000000000000")
				return req
			}).OnPath(`/management.azure.com/`),
		},
		ResponseFilters: govcr.ResponseFilters{
			govcr.ResponseDeleteHeaderKeys(
				"Date", "Server", "Cache-Control", "Expires", "Pragma", "Strict-Transport-Security", "Vary",
				"X-Content-Type-Options",
				"X-Ms-Correlation-Request-Id",
				"X-Ms-Ratelimit-Remaining-Resource",
				"X-Ms-Ratelimit-Remaining-Subscription-Reads",
				"X-Ms-Request-Id",
				"X-Ms-Routing-Request-Id"),
		},
	}
	cli := azureClientSet{}

	if replayOnly {
		t.Log("Azure credentials not available, will use existing recordings.")
		cli.subscriptionID = "00000000-0000-1234-0000-000000000000"
	} else {
		if os.Getenv("AZURE_RECORD") != "" {
			cfg.DisableRecording = false
			cfg.Logging = true
		}
		a, err := auth.NewAuthorizerFromEnvironment()
		if err == nil {
			cli.authorizer = a
			cli.subscriptionID = os.Getenv("AZURE_SUBSCRIPTION_ID")
			cli.PollingDelay = 0
		}
	}

	vcr := govcr.NewVCR(t.Name(), cfg)
	cli.sender = vcr.Client
	return cli, nil
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
