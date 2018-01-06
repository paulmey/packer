package common

import (
	"fmt"
	"strings"
)

func PlatformImageUrn(urn string) (publisher, offer, sku, version string, err error) {
	parts := strings.Split(urn, ":")
	if len(parts) != 4 {
		err = fmt.Errorf("exptected 4 parts in URN but found %d", len(parts))
		return
	}

	return parts[0], parts[1], parts[2], parts[3], nil
}
