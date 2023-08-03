package provider

import (
	"fmt"
	"strings"
)

// SetString sets provider's attributes.
func (provider *Provider) SetString(providerStr string) error {
	attrs := strings.Split(providerStr, "/")
	if len(attrs) < 4 {
		return fmt.Errorf("wrong provider format: %s", providerStr)
	}

	provider.URL = providerStr
	provider.Host = attrs[0]
	provider.Namespace = attrs[1]
	provider.Name = attrs[2]
	provider.Version = attrs[3]

	return nil
}
