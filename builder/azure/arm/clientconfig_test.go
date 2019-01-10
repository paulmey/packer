package arm

import (
	"fmt"
	"os"
	"testing"

	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/hashicorp/packer/builder/azure/common/constants"
	"github.com/hashicorp/packer/packer"
)

func Test_newConfig_ClientConfig(t *testing.T) {
	baseConfig := map[string]string{
		"capture_name_prefix":    "ignore",
		"capture_container_name": "ignore",
		"location":               "ignore",
		"image_url":              "ignore",
		"storage_account":        "ignore",
		"resource_group_name":    "ignore",
		"os_type":                constants.Target_Linux,
	}

	tests := []struct {
		name    string
		args    []interface{}
		wantErr bool
	}{
		{
			name: "no client_id, client_secret or subscription_id should enable MSI auth",
			args: []interface{}{
				baseConfig,
				getPackerConfiguration(),
			},
			wantErr: false,
		},
		{
			name: "subscription_id is set will trigger device flow",
			args: []interface{}{
				baseConfig,
				map[string]string{
					"subscription_id": "error",
				},
				getPackerConfiguration(),
			},
			wantErr: false,
		},
		{
			name: "client_id without client_secret should error",
			args: []interface{}{
				baseConfig,
				map[string]string{
					"client_id": "error",
				},
				getPackerConfiguration(),
			},
			wantErr: true,
		},
		{
			name: "client_secret without client_id should error",
			args: []interface{}{
				baseConfig,
				map[string]string{
					"client_secret": "error",
				},
				getPackerConfiguration(),
			},
			wantErr: true,
		},
		{
			name: "missing subscription_id when using secret",
			args: []interface{}{
				baseConfig,
				map[string]string{
					"client_id":     "ok",
					"client_secret": "ok",
				},
				getPackerConfiguration(),
			},
			wantErr: true,
		},
		{
			name: "tenant_id alone should fail",
			args: []interface{}{
				baseConfig,
				map[string]string{
					"tenant_id": "ok",
				},
				getPackerConfiguration(),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := newConfig(tt.args...)
			if (err != nil) != tt.wantErr {
				t.Errorf("newConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func Test_ClientConfig_DeviceLogin(t *testing.T) {
	getEnvOrSkip(t, "AZURE_DEVICE_LOGIN")
	cfg := ClientConfig{
		SubscriptionID:   getEnvOrSkip(t, "AZURE_SUBSCRIPTION"),
		cloudEnvironment: getCloud(),
	}
	assertValid(t, cfg)

	spt, sptkv, err := cfg.getServicePrincipalTokens(
		func(s string) { fmt.Printf("SAY: %s\n", s) })
	if err != nil {
		t.Fatalf("Expected nil err, but got: %v", err)
	}
	token := spt.Token()
	if token.AccessToken == "" {
		t.Fatal("Expected management token to have non-nil access token")
	}
	if token.RefreshToken == "" {
		t.Fatal("Expected management token to have non-nil refresh token")
	}
	kvtoken := sptkv.Token()
	if kvtoken.AccessToken == "" {
		t.Fatal("Expected keyvault token to have non-nil access token")
	}
	if kvtoken.RefreshToken == "" {
		t.Fatal("Expected keyvault token to have non-nil refresh token")
	}
}

func Test_ClientConfig_ClientPassword(t *testing.T) {
	cfg := ClientConfig{
		SubscriptionID:   getEnvOrSkip(t, "AZURE_SUBSCRIPTION"),
		ClientID:         getEnvOrSkip(t, "AZURE_CLIENTID"),
		ClientSecret:     getEnvOrSkip(t, "AZURE_CLIENTSECRET"),
		TenantID:         getEnvOrSkip(t, "AZURE_TENANTID"),
		cloudEnvironment: getCloud(),
	}
	assertValid(t, cfg)

	spt, sptkv, err := cfg.getServicePrincipalTokens(func(s string) { fmt.Printf("SAY: %s\n", s) })
	if err != nil {
		t.Fatalf("Expected nil err, but got: %v", err)
	}
	token := spt.Token()
	if token.AccessToken == "" {
		t.Fatal("Expected management token to have non-nil access token")
	}
	if token.RefreshToken != "" {
		t.Fatal("Expected management token to have no refresh token")
	}
	kvtoken := sptkv.Token()
	if kvtoken.AccessToken == "" {
		t.Fatal("Expected keyvault token to have non-nil access token")
	}
	if kvtoken.RefreshToken != "" {
		t.Fatal("Expected keyvault token to have no refresh token")
	}
}

func getEnvOrSkip(t *testing.T, envVar string) string {
	v := os.Getenv(envVar)
	if v == "" {
		t.Skipf("%s is empty, skipping", envVar)
	}
	return v
}

func getCloud() *azure.Environment {
	cloudName := os.Getenv("AZURE_CLOUD")
	if cloudName == "" {
		cloudName = "AZUREPUBLICCLOUD"
	}
	c, _ := azure.EnvironmentFromName(cloudName)
	return &c
}

// tests for assertRequiredParametersSet

func Test_ClientConfig_CanUseDeviceCode(t *testing.T) {
	cfg := emptyClientConfig()
	cfg.SubscriptionID = "12345"
	// TenantID is optional

	assertValid(t, cfg)
}

func assertValid(t *testing.T, cfg ClientConfig) {
	errs := &packer.MultiError{}
	cfg.assertRequiredParametersSet(errs)
	if len(errs.Errors) != 0 {
		t.Fatal("Expected errs to be empty: ", errs)
	}
}

func assertInvalid(t *testing.T, cfg ClientConfig) {
	errs := &packer.MultiError{}
	cfg.assertRequiredParametersSet(errs)
	if len(errs.Errors) == 0 {
		t.Fatal("Expected errs to be non-empty")
	}
}

func Test_ClientConfig_CanUseClientSecret(t *testing.T) {
	cfg := emptyClientConfig()
	cfg.SubscriptionID = "12345"
	cfg.ClientID = "12345"
	cfg.ClientSecret = "12345"

	assertValid(t, cfg)
}

func Test_ClientConfig_CanUseClientSecretWithTenantID(t *testing.T) {
	cfg := emptyClientConfig()
	cfg.SubscriptionID = "12345"
	cfg.ClientID = "12345"
	cfg.ClientSecret = "12345"
	cfg.TenantID = "12345"

	assertValid(t, cfg)
}

func emptyClientConfig() ClientConfig {
	cfg := ClientConfig{}
	_ = cfg.setCloudEnvironment()
	return cfg
}
