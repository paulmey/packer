package common

import (
	"github.com/Azure/go-autorest/autorest"
	"net/http"
)

var Sender autorest.Sender = autorest.SenderFunc(http.DefaultClient.Do)
