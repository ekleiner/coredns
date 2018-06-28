package metadata

import (
	"context"
	"testing"

	"github.com/coredns/coredns/plugin/test"

	"github.com/miekg/dns"
)

// testMetadataer implements fake Metadataers. Plugins which inmplement Metadataer interface
type testMetadataer map[string]interface{}

func (m testMetadataer) MetadataVarNames() []string {
	keys := []string{}
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func (m testMetadataer) Metadata(ctx context.Context, w dns.ResponseWriter, r *dns.Msg, key string) (val interface{}, ok bool) {
	value, ok := m[key]
	return value, ok
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
		testMetadataer{"testkey1": "testvalue1"},
		testMetadataer{"testkey2": 2, "testkey3": "testvalue3"},
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
		for expKey, expVal := range expected {
			metadataVal, valOk := md.Value(expKey)
			if !valOk {
				t.Fatalf("Value by key %v can't be retrieved", expKey)
			}
			if metadataVal != expVal {
				t.Errorf("Expected value %v, but got %v", expVal, metadataVal)
			}
		}
		wrongKey := "wrong_key"
		metadataVal, ok := md.Value(wrongKey)
		if ok {
			t.Fatalf("Value by key %v is not expected to be recieved, but got: %v", wrongKey, metadataVal)
		}
	}
}
