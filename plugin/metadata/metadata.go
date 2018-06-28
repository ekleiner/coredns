package metadata

import (
	"context"

	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/plugin/pkg/variables"
	"github.com/coredns/coredns/request"

	"github.com/miekg/dns"
)

// Metadata implements collecting metadata information from all plugins that
// implement the Metadataer interface.
type Metadata struct {
	Zones       []string
	Metadataers []Metadataer
	Next        plugin.Handler
}

// Name implements the Handler interface.
func (m *Metadata) Name() string { return "metadata" }

// ServeDNS implements the plugin.Handler interface.
func (m *Metadata) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {

	md, ctx := newMD(ctx)

	state := request.Request{W: w, Req: r}
	if plugin.Zones(m.Zones).Matches(state.Name()) != "" {
		// Go through all Metadataers and collect metadata
		for _, metadataer := range m.Metadataers {
			for _, varName := range metadataer.MetadataVarNames() {
				if val, ok := metadataer.Metadata(ctx, w, r, varName); ok {
					md.setValue(varName, val)
				}
			}
		}
	}

	rcode, err := plugin.NextOrFailure(m.Name(), m.Next, ctx, w, r)

	return rcode, err
}

// MetadataVarNames implements the plugin.Metadataer interface.
func (m *Metadata) MetadataVarNames() []string { return variables.All }

// Metadata implements the plugin.Metadataer interface.
func (m *Metadata) Metadata(ctx context.Context, w dns.ResponseWriter, r *dns.Msg, varName string) (interface{}, bool) {
	if val, err := variables.GetValue(varName, w, r); err == nil {
		return val, true
	}
	return nil, false
}
