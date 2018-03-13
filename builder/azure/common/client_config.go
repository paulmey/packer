package common

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/oauth2/jws"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2017-03-30/compute"
	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2016-06-01/subscriptions"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/Azure/go-autorest/autorest/to"
)

type ClientConfig struct {
	// optional: if not set, walk down the list to see which one if filled out or
	// otherwise try device login
	AuthenticationType string `mapstructure:"authentication_type"`

	SubscriptionID string `mapstructure:"subscription_id"`

	ServicePrincipalConfig `mapstructure:",squash"`
	OAuthToken             string `mapstructure:"oauth_token"`

	ManagementURI string `mapstructure:"management_uri"`
}

const (
	AuthenticationTypeDeviceLogin      string = "interactive"
	AuthenticationTypeServicePrincipal string = "sp"
	AuthenticationTypeToken            string = "oauth"
)

func (c *ClientConfig) resolveAuthenticationType() {
	if c.AuthenticationType == "" {
		if c.ServicePrincipalConfig.isConfigured() {
			c.AuthenticationType = AuthenticationTypeServicePrincipal
		} else if c.OAuthToken != "" {
			c.AuthenticationType = AuthenticationTypeToken
		} else {
			c.AuthenticationType = AuthenticationTypeDeviceLogin
		}
	}
}

type ServicePrincipalConfig struct {
	// Authentication via OAUTH
	ClientID     string `mapstructure:"client_id"`
	ClientSecret string `mapstructure:"client_secret"`
	ObjectID     string `mapstructure:"object_id"`
	TenantID     string `mapstructure:"tenant_id"`
}

func (spc ServicePrincipalConfig) isConfigured() bool {
	return spc.ClientID != "" &&
		spc.ClientSecret != "" &&
		spc.ObjectID != "" &&
		spc.TenantID != ""
}

func (c *ClientConfig) GetClient() (AzureClient, error) {
	if c.SubscriptionID == "" {
		return nil, errors.New("subscription_id is required.")
	}

	cli := azureClient{
		autorest.NewClientWithUserAgent("Packer.io"),
		c.ManagementURI,
		c.SubscriptionID,
	}

	if cli.BaseUri == "" {
		cli.BaseUri = subscriptions.DefaultBaseURI
	}

	cli.Sender = Sender

	c.resolveAuthenticationType()

	authorizer, err := c.getAuthorizer()
	if err != nil {
		return nil, err
	}
	cli.Authorizer = authorizer

	iter, err := cli.SubscriptionsClient().ListComplete(context.Background())
	if err != nil {
		return azureClient{}, fmt.Errorf("Error retrieving subscriptions: %v", err)
	}

	found := false
	for ; iter.NotDone(); iter.Next() {
		sub := iter.Value()
		if to.String(sub.SubscriptionID) == cli.SubscriptionID {
			found = true
			break
		}
	}
	if !found {
		return azureClient{}, fmt.Errorf("Subscription %s not available using these credentials",
			c.SubscriptionID)
	}

	return cli, nil
}

func (c ClientConfig) getAuthorizer() (autorest.Authorizer, error) {
	if c.AuthenticationType == AuthenticationTypeDeviceLogin {
		return c.getDeviceLoginAuthorizer()
	} else if c.AuthenticationType == AuthenticationTypeServicePrincipal {
		return c.getServicePrincipalAuthorizer()
	} else if c.AuthenticationType == AuthenticationTypeToken {
		return c.getOAuthAuthorize()
	}
	return nil, fmt.Errorf("ClientConfig: Invalid value for authentication_type: %s (accepted: %s, %s, %s)",
		c.AuthenticationType,
		AuthenticationTypeDeviceLogin,
		AuthenticationTypeServicePrincipal,
		AuthenticationTypeToken)
}

func (c ClientConfig) getDeviceLoginAuthorizer() (autorest.Authorizer, error) {
	return nil, errors.New("getDeviceLoginAuthorizer: not implemented")
}

func (c ClientConfig) getServicePrincipalAuthorizer() (autorest.Authorizer, error) {
	return nil, errors.New("getServicePrincipalAuthorizer: not implemented")
}

func (c ClientConfig) getOAuthAuthorize() (autorest.Authorizer, error) {
	claims, err := jws.Decode(c.OAuthToken)
	if err != nil {
		return nil, err
	}

	// Iss: https://sts.windows.net/72f988bf-86f1-41af-91ab-2d7cd011db47/
	// https://docs.microsoft.com/en-us/azure/active-directory/develop/active-directory-token-and-claims

	issParts := strings.SplitN(claims.Iss, "/", 5)
	if len(issParts) < 5 {
		return nil, fmt.Errorf("getOAuthAuthorizer: Could not find tenant id in issuer claim: %s", claims.Iss)
	}

	token := NewAutoRefreshToken(adal.Token{
		AccessToken:  c.OAuthToken,
		RefreshToken: c.OAuthToken,
		ExpiresOn:    fmt.Sprintf("%d", claims.Exp),
		Resource:     claims.Aud,
	}, c.getOAuthConfig(issParts[3]))
	return autorest.NewBearerAuthorizer(&token), nil
}

func (c *ClientConfig) getOAuthConfig(tenantID string) adal.OAuthConfig {
	// todo: other clouds
	cfg, _ := adal.NewOAuthConfig(azure.PublicCloud.ActiveDirectoryEndpoint, tenantID)
	return *cfg
}

type AzureClient interface {
	GetComputeMetadata() (ComputeMetadata, error)
	SubscriptionsClient() subscriptions.Client
	PlatformImagesClient() compute.VirtualMachineImagesClient
	ManagedDisksClient() compute.DisksClient

	SetAuthorizer(autorest.Authorizer)
}

type azureClient struct {
	autorest.Client
	BaseUri        string
	SubscriptionID string
}

func (c azureClient) SubscriptionsClient() subscriptions.Client {
	cli := subscriptions.NewClientWithBaseURI(c.BaseUri)
	cli.Client = c.Client
	cli.Client.UserAgent = subscriptions.UserAgent()
	return cli
}

func (c azureClient) PlatformImagesClient() compute.VirtualMachineImagesClient {
	cli := compute.NewVirtualMachineImagesClientWithBaseURI(c.BaseUri, c.SubscriptionID)
	cli.Client = c.Client
	cli.Client.UserAgent = compute.UserAgent()
	return cli
}

func (c azureClient) ManagedDisksClient() compute.DisksClient {
	cli := compute.NewDisksClientWithBaseURI(c.BaseUri, c.SubscriptionID)
	cli.Client = c.Client
	cli.Client.UserAgent = compute.UserAgent()
	return cli
}

func (c azureClient) SetAuthorizer(authorizer autorest.Authorizer) {
	c.Client.Authorizer = authorizer
}
