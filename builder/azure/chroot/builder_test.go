package chroot

import (
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/Azure/go-autorest/autorest"
	azcommon "github.com/hashicorp/packer/builder/azure/common"

	chk "gopkg.in/check.v1"

	"github.com/dnaeon/go-vcr/recorder"
)

// func TestBuilder_ImplementsBuilder(t *testing.T) {
// 	var raw interface{}
// 	raw = &Builder{}
// 	if _, ok := raw.(packer.Builder); !ok {
// 		t.Fatal("Builder should be a builder")
// 	}
// }

func testConfig() map[string]interface{} {
	return map[string]interface{}{}
}

type ConfigSourceSuite struct{}

var _ = chk.Suite(&ConfigSourceSuite{})

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { chk.TestingT(t) }
func (s *ConfigSourceSuite) SetUpTest(c *chk.C) {
	setRecorder(c)
}

func (s *ConfigSourceSuite) TestWhenNoSourceThenFail(c *chk.C) {
	config := testConfig()
	config["source"] = ""

	b := Builder{}

	warn, err := b.Prepare(config)

	if len(warn) != 0 {
		c.Log("Warnings: ", warn)
	}
	if err == nil {
		c.Error("Expected Prepare to fail with empty source")
	}
}

func (s *ConfigSourceSuite) WhenSourceUrnNotExistsThenFail(t *testing.T) {
	image := "NotExists:UbuntuServer:16.04-LTS:LaTest"
	config := testConfig()
	config["source"] = image

	b := Builder{}

	warn, err := b.Prepare(config)

	if len(warn) != 0 {
		t.Log("Warnings: ", warn)
	}
	if err == nil ||
		!strings.Contains(err.Error(), "Config: Image URN not found") ||
		!strings.Contains(err.Error(), image) {
		t.Errorf("Expected 'Config: Image URN not found' but got %q", err)
	}
}

// func (s *ConfigSourceSuite) MetadataShouldBeComplete(t *testing.T) {
// 	for key := range vmMetadata {
// 		tc := vmMetadata
// 		delete(tc, key)
// 		t.Run(fmt.Sprintf("'%s' empty", key),
// 			test_MetadataIncomplete(tc))
// 	}
// }

// func (s *ConfigSourceSuite) MetadataIncomplete(md interface{}) func(*testing.T) {
// 	return func(t *testing.T) {
// 		oldsender := azcommon.Sender
// 		defer func() { azcommon.Sender = oldsender }()
// 		azcommon.Sender = &httpmock.Sender{
// 			[]httpmock.Mock{
// 				httpmock.Get("http://169\\.254\\.169\\.254/metadata/instance/compute[^/]*", md),
// 			}}

// 		b := Builder{}
// 		warn, err := b.Prepare(testConfig())

// 		if len(warn) != 0 {
// 			t.Log("Warnings: ", warn)
// 		}
// 		if err == nil || strings.Contains(err.Error(), "VM metadata not complete") {
// 			t.Errorf("Expected 'VM metadata not complete', but got %q", err)
// 		}
// 	}
// }

func (s *ConfigSourceSuite) WarnsSubscriptionIDOverride(t *testing.T) {
	c := testConfig()
	c["subscription_id"] = "not-the-vm-metadata-sub-id"
	b := Builder{}

	warns, err := b.Prepare(c)
	if err != nil {
		t.Logf("err: %+v", err)
	}
	for _, w := range warns {
		matched, err := regexp.MatchString("subscription_id \\([^)]*\\) is overridden", w)
		if err != nil {
			t.Fatal(err)
		}
		if matched {
			return
		}
	}
	t.Errorf("Expected warning about subscription_id being overridden: %v", warns)
}

var vmMetadata = map[string]string{
	"location":          "westmock",
	"subscriptionId":    "125",
	"resourceGroupName": "mockedResourceGroup",
	"name":              "mockedVMName",
}

const (
	recordingsFolder = "recordings"
)

func setRecorder(c *chk.C) *recorder.Recorder {
	tests := strings.Split(c.TestName(), ".")
	path := filepath.Join(recordingsFolder, tests[0], tests[1])

	rec, err := recorder.New(path)
	c.Assert(err, chk.IsNil)
	azcommon.Sender = autorest.SenderFunc(rec.RoundTrip)

	// rec.SetMatcher(func(r *http.Request, i cassette.Request) bool {
	// 	return compareMethods(r, i) &&
	// 		compareURLs(r, i) &&
	// 		compareHeaders(r, i) &&
	// 		compareBodies(r, i)
	// })
	return rec
}

// func TestMain(m *testing.M) {
// 	azcommon.Sender = &httpmock.Sender{
// 		[]httpmock.Mock{
// 			httpmock.Get("^http://169\\.254\\.169\\.254/metadata/instance/compute\\?",
// 				vmMetadata),
// 			httpmock.Get("^https://management\\.azure\\.com/subscriptions\\?",
// 				map[string]interface{}{
// 					"value": []interface{}{
// 						map[string]string{
// 							"subscriptionId": "125",
// 							"displayName":    "Mocked Subscription",
// 						}}}),
// 			httpmock.GetNotFound("^https://management.azure.com/subscriptions/[^/]+/providers/Microsoft.Compute/locations/[^/]+/publishers/NotExists/artifacttypes/vmimage/offers/[^/]+/skus/[^/]+/versions\\?",
// 				response_image_not_found_error),
// 		}}

// 	os.Exit(m.Run())
// }

// const (
// 	response_image_not_found_error string = `{
// 		"error": {
// 			"code": "NotFound",
// 			"message": "Artifact: VMImage was not found."
// 		}
// 	}`
// 	response_list_platform_image_version string = `[
// 	{
// 		"location": "westus",
// 		"name": "4.0.20160617",
// 		"id": "/Subscriptions/{subscription-id}/Providers/Microsoft.Compute/Locations/westus/Publishers/PublisherA/ArtifactTypes/VMImage/Offers/OfferA/Skus/SkuA/Versions/4.0.20160617"
// 	},
// 	{
// 		"location": "westus",
// 		"name": "4.0.20160721",
// 		"id": "/Subscriptions/{subscription-id}/Providers/Microsoft.Compute/Locations/westus/Publishers/PublisherA/ArtifactTypes/VMImage/Offers/OfferA/Skus/SkuA/Versions/4.0.20160721"
// 	},
// 	{
// 		"location": "westus",
// 		"name": "4.0.20160812",
// 		"id": "/Subscriptions/{subscription-id}/Providers/Microsoft.Compute/Locations/westus/Publishers/PublisherA/ArtifactTypes/VMImage/Offers/OfferA/Skus/SkuA/Versions/4.0.20160812"
// 	}]`
// )
