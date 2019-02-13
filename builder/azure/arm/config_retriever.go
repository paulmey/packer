package arm

// Method to resolve information about the user so that a client can be
// constructed to communicated with Azure.
//
// The following data are resolved.
//
// 1. TenantID

import (
	"log"

	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/hashicorp/packer/builder/azure/common"
)

type configRetriever struct {
	// test seams
	findTenantID func(azure.Environment, string) (string, error)
}

func newConfigRetriever() configRetriever {
	return configRetriever{
		common.FindTenantID,
	}
}

func (cr configRetriever) FillParameters(c *Config) error {
	if c.TenantID == "" && c.SubscriptionID != "" {
		log.Print("Getting tenant ID from ARM")
		tenantID, err := cr.findTenantID(*c.cloudEnvironment, c.SubscriptionID)
		if err != nil {
			return err
		}
		c.TenantID = tenantID
		log.Printf("Tenant ID: %s", c.TenantID)
	}

	return nil
}
