package metadata

import (
	"testing"
)

const typicalResponse = `
{ 
	"compute":{
    "location":"westus2",
		"name":"azure-vm",
		"offer":"",
		"osType":"Linux",
		"placementGroupId":"",
		"platformFaultDomain":"0",
		"platformUpdateDomain":"0",
		"publisher":"",
		"resourceGroupName":"azure-resource-group",
		"sku":"",
		"subscriptionId":"11111111-2222-3333-4444-555555555555",
		"tags":"",
		"version":"",
		"vmId":"66666666-7777-8888-9999-000000000000",
		"vmSize":"Standard_D8s_v3"},
		"network":{
			"interface":[
			{
				"ipv4":{
					"ipAddress":[{
						"privateIpAddress":"1.2.3.4",
						"publicIpAddress":""
					}],
					"subnet":[{
						"address":"5.6.7.8",
						"prefix":"27"
					}]
				},
				"ipv6":{
					"ipAddress":[]
				},
				"macAddress":"ABCDEFABCDEF"
			}]
		}
	}
`

func Test_GetSubscriptionID(t *testing.T) {
	fetcher = func() ([]byte, error) {
		return []byte(typicalResponse), nil
	}

	md, err := Get()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "11111111-2222-3333-4444-555555555555"
	if md.SubscriptionID != expected {
		t.Fatalf("Expected subscription id to be %q, but got %q", expected, md.SubscriptionID)
	}
}
