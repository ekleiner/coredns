package metadata

import (
	"context"

	"github.com/coredns/coredns/plugin"
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
		for _, Metadataer := range m.Metadataers {
			metadata := Metadataer.Metadata(ctx, w, r)
			md.addValues(metadata)
		}
	}

	rcode, err := plugin.NextOrFailure(m.Name(), m.Next, ctx, w, r)

	return rcode, err
}
