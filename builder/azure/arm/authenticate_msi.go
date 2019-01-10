package arm

import (
	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/go-autorest/autorest/azure"
)

// for managed identity auth
type msiOAuthTokenProvider struct {
	env azure.Environment
}

func NewMSIOAuthTokenProvider(env azure.Environment) oAuthTokenProvider {
	return &msiOAuthTokenProvider{env}
}

func (a *msiOAuthTokenProvider) getServicePrincipalToken() (*adal.ServicePrincipalToken, error) {
	return a.getServicePrincipalTokenWithResource(a.env.ResourceManagerEndpoint)
}

func (a *msiOAuthTokenProvider) getServicePrincipalTokenWithResource(resource string) (*adal.ServicePrincipalToken, error) {
	return adal.NewServicePrincipalTokenFromMSI("http://169.254.169.254/metadata/identity/oauth2/token", resource)
}
