package metadata

import (
	"context"
	"testing"

	"github.com/coredns/coredns/plugin/test"
	"github.com/miekg/dns"
)

// testMetadataer implements fake Metadataers. Plugins which inmplement Metadataer interface
type testMetadataer struct {
	key   string
	value interface{}
}

func (m testMetadataer) Metadata(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) map[string]interface{} {
	return map[string]interface{}{m.key: m.value}
}

// testHandler implements plugin.Handler
type testHandler struct{ ctx context.Context }

func (m *testHandler) Name() string { return "testHandler" }

func (m *testHandler) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	m.ctx = ctx
	return 0, nil
}

func TestMetadataServDns(t *testing.T) {
	expectedMetadata := []testMetadataer{
		{"testkey1", "testvalue"},
		{"testkey2", 795},
	}
	// Create fake Metadataers based on expectedMetadata
	Metadataers := []Metadataer{}
	for _, e := range expectedMetadata {
		Metadataers = append(Metadataers, e)
	}
	// Fake handler which stores the resulting context
	next := &testHandler{}

	metadata := Metadata{
		Zones:       []string{"."},
		Metadataers: Metadataers,
		Next:        next,
	}
	metadata.ServeDNS(context.TODO(), &test.ResponseWriter{}, new(dns.Msg))

	// Verify that next plugin can find metadata in context from all Metadataers
	for _, expected := range expectedMetadata {
		md, ok := FromContext(next.ctx)
		if !ok {
			t.Fatalf("Metadata is expected but not present inside the context")
		}
		metadataVal, ok := md.Get(expected.key)
		if !ok {
			t.Fatalf("Value by key %v can't be retrieved", expected.key)
		}
		if metadataVal != expected.value {
			t.Errorf("Expected value %v, but got %v", expected.value, metadataVal)
		}
		wrongKey := "wrong_key"
		metadataVal, ok = md.Get(wrongKey)
		if ok {
			t.Fatalf("Value by key %v is not expected to be recieved, but got: %v", wrongKey, metadataVal)
		}
	}
}
