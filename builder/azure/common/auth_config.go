package common

import (
	"context"
	"errors"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2016-06-01/subscriptions"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/to"
	"log"
)

type ClientConfig struct {
	// optional: if not set, walk down the list to see which one if filled out or
	// otherwise try device login
	AuthenticationType string `mapstructure:"authentication_type"`

	SubscriptionID string `mapstructure:"subscription_id"`

	ServicePrincipalConfig `mapstructure:",squash"`
	OAuthToken             string `mapstructure:"oauth_token"`

	ManagementURI string `mapstructure:"management_uri"`

	client AzureClient
}

const (
	AuthenticationTypeDeviceLogin      string = "devicelogin"
	AuthenticationTypeServicePrincipal string = "sp"
	AuthenticationTypeToken            string = "token"
)

type ServicePrincipalConfig struct {
	// Authentication via OAUTH
	ClientID     string `mapstructure:"client_id"`
	ClientSecret string `mapstructure:"client_secret"`
	ObjectID     string `mapstructure:"object_id"`
	TenantID     string `mapstructure:"tenant_id"`
}

func (c *ClientConfig) ResolveClient() error {
	if c.SubscriptionID == "" {
		errors.New("subscription_id is required.")
	}

	cli := azureClient{
		autorest.NewClientWithUserAgent("Packer.io"),
		c.ManagementURI,
	}

	if cli.BaseUri == "" {
		cli.BaseUri = subscriptions.DefaultBaseURI
	}

	cli.Sender = Sender
	c.client = cli

	iter, err := c.client.SubscriptionsClient().ListComplete(context.Background())
	if err != nil {
		return fmt.Errorf("Error retrieving subscriptions: %v", err)
	}

	found := false
	for ; iter.NotDone(); iter.Next() {
		sub := iter.Value()
		if to.String(sub.SubscriptionID) == c.SubscriptionID {

			log.Printf("azure: found subscription %q (%s)",
				to.String(sub.DisplayName),
				to.String(sub.SubscriptionID))
			found = true
			break
		}
	}
	if !found {
		log.Print("azure: subscription not found")
		return fmt.Errorf("Subscription %s not available using these credentials",
			c.SubscriptionID)
	}

	return nil
}

type AzureClient interface {
	SubscriptionsClient() subscriptions.Client
}

type azureClient struct {
	autorest.Client
	BaseUri string
}

func (c azureClient) SubscriptionsClient() subscriptions.Client {
	cli := subscriptions.NewClientWithBaseURI(c.BaseUri)
	cli.Client = c.Client
	cli.Client.UserAgent = subscriptions.UserAgent()
	return cli
}
