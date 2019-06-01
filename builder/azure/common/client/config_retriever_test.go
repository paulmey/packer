package client

// todo: re-enable tests
//import (
//	"errors"
//	"github.com/hashicorp/packer/builder/azure/arm"
//	"testing"
//
//	"github.com/Azure/go-autorest/autorest/azure"
//)
//
//func TestConfigRetrieverFillsTenantIDWhenEmpty(t *testing.T) {
//	c, _, _ := arm.newConfig(arm.getArmBuilderConfiguration(), arm.getPackerConfiguration())
//	if expected := ""; c.TenantID != expected {
//		t.Errorf("Expected TenantID to be %q but got %q", expected, c.TenantID)
//	}
//
//	sut := newTestConfigRetriever()
//	retrievedTid := "my-tenant-id"
//	findTenantID = func(azure.Environment, string) (string, error) { return retrievedTid, nil }
//	if err := FillParameters(c); err != nil {
//		t.Errorf("Unexpected error when calling sut.FillParameters: %v", err)
//	}
//
//	if expected := retrievedTid; c.TenantID != expected {
//		t.Errorf("Expected TenantID to be %q but got %q", expected, c.TenantID)
//	}
//}
//
//func TestConfigRetrieverLeavesTenantIDWhenNotEmpty(t *testing.T) {
//	c, _, _ := arm.newConfig(arm.getArmBuilderConfiguration(), arm.getPackerConfiguration())
//	userSpecifiedTid := "not-empty"
//	c.TenantID = userSpecifiedTid
//
//	sut := newTestConfigRetriever()
//	findTenantID = nil // assert that this not even called
//	if err := FillParameters(c); err != nil {
//		t.Errorf("Unexpected error when calling sut.FillParameters: %v", err)
//	}
//
//	if expected := userSpecifiedTid; c.TenantID != expected {
//		t.Errorf("Expected TenantID to be %q but got %q", expected, c.TenantID)
//	}
//}
//
//func TestConfigRetrieverReturnsErrorWhenTenantIDEmptyAndRetrievalFails(t *testing.T) {
//	c, _, _ := arm.newConfig(arm.getArmBuilderConfiguration(), arm.getPackerConfiguration())
//	if expected := ""; c.TenantID != expected {
//		t.Errorf("Expected TenantID to be %q but got %q", expected, c.TenantID)
//	}
//
//	sut := newTestConfigRetriever()
//	errorString := "sorry, I failed"
//	findTenantID = func(azure.Environment, string) (string, error) { return "", errors.New(errorString) }
//	if err := FillParameters(c); err != nil && err.Error() != errorString {
//		t.Errorf("Unexpected error when calling sut.FillParameters: %v", err)
//	}
//}
//
//func newTestConfigRetriever() configRetriever {
//	return configRetriever{
//		findTenantID: func(azure.Environment, string) (string, error) { return "findTenantID is mocked", nil },
//	}
//}
