package metadata

import (
	"context"

	"github.com/coredns/coredns/plugin/pkg/log"
	"github.com/miekg/dns"
)

// Metadataer interface needs to be implemented by each plugin willing to provide
// metadata information for other plugins.
// Note: this method should work quickly, because it is called for every request
// from the metadata plugin.
type Metadataer interface {
	// Metadata is expected to return map with metadata information which can be
	// later retrieved from context by any other plugin. It may return empty
	// map if no metadata needs to be published.
	Metadata(context.Context, dns.ResponseWriter, *dns.Msg) map[string]interface{}
}

// MD is metadata information storage
type MD map[string]interface{}

// metadataKey defines the type of key that is used to save metadata into the context
type metadataKey struct{}

// newMD initializes MD and attaches it to context
func newMD(ctx context.Context) (MD, context.Context) {
	m := MD{}
	return m, context.WithValue(ctx, metadataKey{}, m)
}

// FromContext retrieves MD struct from context
func FromContext(ctx context.Context) (md MD, ok bool) {
	if metadata := ctx.Value(metadataKey{}); metadata != nil {
		if md, ok := metadata.(MD); ok {
			return md, true
		}
	}
	return nil, false
}

// Get returns metadata value by key
func (m MD) Get(key string) (value interface{}, ok bool) {
	value, ok = m[key]
	return
}

// addValues adds metadata values.
// If variable with a new key already attached then new is not appllied, old is removed.
func (m MD) addValues(src map[string]interface{}) {
	duplicates := []string{}
	for k, v := range src {
		if _, ok := m[k]; !ok {
			m[k] = v
		} else {
			duplicates = append(duplicates, k)
			log.Errorf("Metadata variable '%v' has duplicates. None of them is used.", k)
		}
	}
	// Remove duplicated variables
	for _, k := range duplicates {
		delete(m, k)
	}
}
