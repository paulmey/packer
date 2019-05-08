package client

import (
	"fmt"
	"os"
	"regexp"
	"testing"
	"time"

	"github.com/Azure/azure-sdk-for-go/profiles/latest/compute/mgmt/compute"
	"github.com/Azure/azure-sdk-for-go/profiles/latest/compute/mgmt/compute/computeapi"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure/auth"
	"github.com/seborama/govcr"
)

type AzureClientSet interface {
	MetadataClient() MetadataClientAPI

	DisksClient() computeapi.DisksClientAPI
	VirtualMachinesClient() computeapi.VirtualMachinesClientAPI

	PollClient() autorest.Client
}

var subscriptionPathRegex = regexp.MustCompile(`/subscriptions/([[:xdigit:]]{8}(-[[:xdigit:]]{4}){3}-[[:xdigit:]]{12})`)

func GetTestClientSet(t *testing.T) (AzureClientSet, error) {
	replayOnly :=
		os.Getenv("AZURE_CLIENT_ID") == "" ||
			os.Getenv("AZURE_CLIENT_SECRET") == "" ||
			os.Getenv("AZURE_SUBSCRIPTION_ID") == "" ||
			os.Getenv("AZURE_TENANT_ID") == ""

	cfg := &govcr.VCRConfig{
		DisableRecording: true,
		RequestFilters: govcr.RequestFilters{
			govcr.RequestDeleteHeaderKeys("Authorization"),
			govcr.RequestFilter(func(req govcr.Request) govcr.Request {
				req.URL.Path = subscriptionPathRegex.ReplaceAllLiteralString(req.URL.Path,
					"/subscriptions/00000000-0000-1234-0000-000000000000")
				return req
			}).OnPath(`/management.azure.com/`),
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

type azureClientSet struct {
	sender         autorest.Sender
	authorizer     autorest.Authorizer
	subscriptionID string
}

func (s azureClientSet) configureAutorestClient(c *autorest.Client) {
	c.Authorizer = s.authorizer
	c.Sender = s.sender

	//l := log.New(os.Stdout,"AzureClient: ", log.LstdFlags)
	//li := autorest.LoggingInspector{Logger: l}
	//c.RequestInspector = li.WithInspection()
	//c.ResponseInspector = li.ByInspecting()
}

func (s azureClientSet) MetadataClient() MetadataClientAPI {
	return metadataClient{
		Sender: s.sender,
	}
}

func (s azureClientSet) DisksClient() computeapi.DisksClientAPI {
	c := compute.NewDisksClient(s.subscriptionID)
	s.configureAutorestClient(&c.Client)
	fmt.Sprintf("***** %+v", c.Client)
	return c
}

func (s azureClientSet) VirtualMachinesClient() computeapi.VirtualMachinesClientAPI {
	c := compute.NewVirtualMachinesClient(s.subscriptionID)
	s.configureAutorestClient(&c.Client)
	return c
}

func (s azureClientSet) PollClient() autorest.Client {
	c := autorest.NewClientWithUserAgent("Packer-Azure-ClientSet")
	s.configureAutorestClient(&c)
	c.PollingDelay = time.Second / 3
	return c
}
