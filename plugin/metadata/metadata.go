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
	// Metadata returns a TODO
	Metadata(context.Context, dns.ResponseWriter, *dns.Msg) (context.Context, error)
}

type Metadata struct {
	Metadaters []Metadater
	Next       plugin.Handler
}

func (m *Metadata) Name() string { return "metadata" }

// ServeDNS implements the plugin.Handler interface.
func (m *Metadata) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {

	// Go through all metadaters and collect metadata
	for _, metadater := range m.Metadaters {
		if c, err := metadater.Metadata(ctx, w, r); err == nil {
			ctx = c
		}
	}

	// context.WithValue(parent, key, val)
	rcode, err := plugin.NextOrFailure(m.Name(), m.Next, ctx, w, r)

	return rcode, err
}
