package chroot

import (
	"errors"
	"log"
	"net/http"
	"os"
	"testing"

	"github.com/Azure/go-autorest/autorest"
	azcommon "github.com/hashicorp/packer/builder/azure/common"

	"github.com/hashicorp/packer/packer"
)

func TestBuilder_ImplementsBuilder(t *testing.T) {
	var raw interface{}
	raw = &Builder{}
	if _, ok := raw.(packer.Builder); !ok {
		t.Fatal("Builder should be a builder")
	}
}

func testConfig() map[string]interface{} {
	return map[string]interface{}{}
}

func TestBuilderPrepare_WhenSourceThenFail(t *testing.T) {
	config := testConfig()
	config["source"] = ""

	b := Builder{}

	warn, err := b.Prepare(config)

	if len(warn) != 0 {
		t.Log("Warnings: ", warn)
	}
	if err == nil {
		t.Error("Expected Prepare to fail with empty source")
	}
}

func TestBuilderPrepare_WhenSourceUrnNotExistsThenFail(t *testing.T) {
	config := testConfig()
	config["source"] = "Canonical:UbuntuServer:16.04-LTS:LaTest"

	b := Builder{}

	warn, err := b.Prepare(config)

	if len(warn) != 0 {
		t.Log("Warnings: ", warn)
	}
	if err == nil {
		t.Error("Expected Prepare to fail when source is a non-exisiting YRN")
	}
}

func TestMain(m *testing.M) {
	azcommon.Sender = autorest.SenderFunc(func(req *http.Request) (*http.Response, error) {
		log.Fatalf("UNHANDLED HTTP TRAFFIC: %s %s", req.Method, req.URL)
		return nil, errors.New("HTTP traffic not allowed in tests")
	})

	os.Exit(m.Run())
}
