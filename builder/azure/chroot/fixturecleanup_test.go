package chroot

import (
	"encoding/json"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/profiles/latest/compute/mgmt/compute"
	"github.com/hashicorp/packer/builder/azure/common/client"
	"log"
	"net/http"
	"regexp"
	"strings"
)

func cleanGetVMResponses(i client.Interaction) client.Interaction {
	// clean VM response objects
	if i.Response.Body != "" && i.Method == http.MethodGet &&
		regexp.MustCompile(`^https://management.azure.com/subscriptions/([^/]+)/resourceGroups/([^/]+)/providers/Microsoft.Compute/virtualMachines/([^/]+)$`).MatchString(i.Url) {

		var vm compute.VirtualMachine
		err := json.Unmarshal([]byte(i.Response.Body), &vm)
		if err == nil {
			replacement := compute.VirtualMachine{
				VirtualMachineProperties: &compute.VirtualMachineProperties{
					ProvisioningState: vm.ProvisioningState,
					StorageProfile:    vm.StorageProfile,
				},
			}
			replacement.StorageProfile.OsDisk = nil
			d, err := json.Marshal(replacement)
			if err == nil {
				i.Response.Body = string(d)
			} else {
				log.Println("failed to marshal VM response", err)
			}
		} else {
			log.Println("failed to unmarshal VM response:", err)
		}
		i.Response.Header.Set("Content-Length", fmt.Sprintf("%d", len(i.Response.Body)))
	}
	return i
}

func cleanMetaDataResponses(computeInfo **client.ComputeInfo) func(client.Interaction) client.Interaction {
	return func(i client.Interaction) client.Interaction {
		// clean metadata responses
		if i.Response.Body != "" && i.Method == http.MethodGet &&
			regexp.MustCompile(`^http://169.254.169.254/metadata/instance?`).MatchString(i.Url) {

			var info struct {
				client.ComputeInfo `json:"compute"`
			}
			err := json.Unmarshal([]byte(i.Response.Body), &info)
			if err == nil {
				if computeInfo != nil && *computeInfo == nil {
					*computeInfo = &info.ComputeInfo
				}

				d, err := json.Marshal(info)
				if err == nil {
					i.Response.Body = string(d)
				} else {
					log.Println("failed to marshal metadata response", err)
				}
			} else {
				log.Println("failed to unmarshal metadata response:", err)
			}
			i.Response.Header.Set("Content-Length", fmt.Sprintf("%d", len(i.Response.Body)))
		}
		return i
	}
}

// if info is not nil, replaces the identifiers from info with mock values
// useful for cleaning before writing
func replaceComputeInfo(info **client.ComputeInfo) func(client.Interaction) client.Interaction {
	return func(i client.Interaction) client.Interaction {
		const MockResourceGroupName = "testrg"
		const MockVMName = "testvm"
		if info != nil && *info != nil {
			replacer := strings.NewReplacer(
				(*info).SubscriptionID, client.MockSubscriptionID,
				(*info).Name, MockVMName,
				(*info).ResourceGroupName, MockResourceGroupName,
			)
			// replace subscription, vm amd rg id's
			i.Request.Url = replacer.Replace(i.Request.Url)
			i.Request.Body = replacer.Replace(i.Request.Body)
			i.Response.Body = replacer.Replace(i.Response.Body)

			for _, k := range []string{
				"Azure-Asyncoperation",
				"Location",
			} {
				if v := i.Response.Header.Get(k); v != "" {
					i.Response.Header.Set(k, replacer.Replace((v)))
				}
			}
			i.Response.Header.Set("Content-Length", fmt.Sprintf("%d", len(i.Response.Body)))
		}
		return i
	}
}

func cleanResponseHeaders(i client.Interaction) client.Interaction {
	// remove response headers
	for _, h := range []string{
		"Cache-Control",
		"Date",
		"Expires",
		"Pragma",
		"Server",
		"Strict-Transport-Security",
		"Vary",
		"X-Content-Type-Options",
		"X-Ms-Correlation-Request-Id",
		"X-Ms-Failure-Cause",
		"X-Ms-Request-Id",
		"X-Ms-Routing-Request-Id",
		"X-Ms-Ratelimit-Remaining-Resource",
		"X-Ms-Ratelimit-Remaining-Subscription-Writes",
		"X-Ms-Ratelimit-Remaining-Subscription-Reads",
		"X-Ms-Served-By",
	} {
		i.Response.Header.Del(h)
	}
	return i
}

func cleanRequestHeaders(i client.Interaction) client.Interaction {
	// remove request headers
	i.Request.Header.Del("Authorization")
	i.Request.Header.Del("User-Agent")

	return i
}
