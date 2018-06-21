package metadata

import (
	"context"

	"github.com/coredns/coredns/plugin"
	"github.com/miekg/dns"
)

// Metadater TODO interface needs to be implemented by each plugin willing to provide
// healthhceck information to the health plugin. Note this method should return
// quickly, i.e. just checking a boolean status, as it is called every second
// from the health plugin.
type Metadater interface {
	// Metadata returns metadata value by variable name.
	// Must provide the values for all variables returned by MetadataVarNames().
	Metadata(string, context.Context, dns.ResponseWriter, *dns.Msg) (interface{}, error)
	// Returns list of metadata variables which this Metadater provides
	MetadataVarsAvailable() []string
}

type Metadata struct {
	Metadaters map[string]Metadater
	Next       plugin.Handler
}

func (m *Metadata) Name() string { return "metadata" }

func (m *Metadata) GetVal(varName string, ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (interface{}, error) {
	metadater := m.Metadaters[varName]
	return metadater.Metadata(varName, ctx, w, r)
}

// ServeDNS implements the plugin.Handler interface.
func (m *Metadata) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {

	ctx = context.WithValue(ctx, "metadata", m.GetVal)

	// context.WithValue(parent, key, val)
	rcode, err := plugin.NextOrFailure(m.Name(), m.Next, ctx, w, r)

	return rcode, err
}
