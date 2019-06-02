package client

import (
	"net/http"
	"regexp"
	"time"

	"github.com/Azure/azure-sdk-for-go/profiles/latest/compute/mgmt/compute"
	"github.com/Azure/azure-sdk-for-go/profiles/latest/compute/mgmt/compute/computeapi"
	"github.com/Azure/go-autorest/autorest"
)

type AzureClientSet interface {
	MetadataClient() MetadataClientAPI

	DisksClient() computeapi.DisksClientAPI
	ImagesClient() computeapi.ImagesClientAPI
	VirtualMachinesClient() computeapi.VirtualMachinesClientAPI

	PollClient() autorest.Client
}

var subscriptionPathRegex = regexp.MustCompile(`/subscriptions/([[:xdigit:]]{8}(-[[:xdigit:]]{4}){3}-[[:xdigit:]]{12})`)

var _ AzureClientSet = &azureClientSet{}

type azureClientSet struct {
	sender         autorest.Sender
	authorizer     autorest.Authorizer
	subscriptionID string
	PollingDelay   time.Duration
}

func New(c Config, say func(string)) (AzureClientSet, error) {
	token, err := c.GetServicePrincipalToken(say, c.CloudEnvironment.ResourceManagerEndpoint)
	if err != nil {
		return nil, err
	}
	return &azureClientSet{
		authorizer:     autorest.NewBearerAuthorizer(token),
		subscriptionID: c.SubscriptionID,
		sender:         http.DefaultClient,
	}, nil
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
	return metadataClient{s.sender}
}

func (s azureClientSet) DisksClient() computeapi.DisksClientAPI {
	c := compute.NewDisksClient(s.subscriptionID)
	s.configureAutorestClient(&c.Client)
	c.PollingDelay = s.PollingDelay
	return c
}

func (s azureClientSet) ImagesClient() computeapi.ImagesClientAPI {
	c := compute.NewImagesClient(s.subscriptionID)
	s.configureAutorestClient(&c.Client)
	c.PollingDelay = s.PollingDelay
	return c
}

func (s azureClientSet) VirtualMachinesClient() computeapi.VirtualMachinesClientAPI {
	c := compute.NewVirtualMachinesClient(s.subscriptionID)
	s.configureAutorestClient(&c.Client)
	c.PollingDelay = s.PollingDelay
	return c
}

func (s azureClientSet) PollClient() autorest.Client {
	c := autorest.NewClientWithUserAgent("Packer-Azure-ClientSet")
	s.configureAutorestClient(&c)
	c.PollingDelay = time.Second / 3
	return c
}
