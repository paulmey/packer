package arm

import (
	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/go-autorest/autorest/azure"
)

// for clientID/secret auth
type secretOAuthTokenProvider struct {
	env                              azure.Environment
	clientID, clientSecret, tenantID string
}

func NewSecretOAuthTokenProvider(env azure.Environment, clientID, clientSecret, tenantID string) oAuthTokenProvider {
	return &secretOAuthTokenProvider{env, clientID, clientSecret, tenantID}
}

func (a *secretOAuthTokenProvider) getServicePrincipalToken() (*adal.ServicePrincipalToken, error) {
	return a.getServicePrincipalTokenWithResource(a.env.ResourceManagerEndpoint)
}

func (a *secretOAuthTokenProvider) getServicePrincipalTokenWithResource(resource string) (*adal.ServicePrincipalToken, error) {
	oauthConfig, err := adal.NewOAuthConfig(a.env.ActiveDirectoryEndpoint, a.tenantID)
	if err != nil {
		return nil, err
	}

	spt, err := adal.NewServicePrincipalToken(
		*oauthConfig,
		a.clientID,
		a.clientSecret,
		resource)

	return spt, err
}
