package metadata

import (
	"fmt"

	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
	"github.com/mholt/caddy"
)

func init() {
	caddy.RegisterPlugin("metadata", caddy.Plugin{
		ServerType: "dns",
		Action:     setup,
	})
}

func setup(c *caddy.Controller) error {
	c.Next()
	if c.NextArg() {
		return plugin.Error("metadata", c.ArgErr())
	}

	h := &Metadata{}

	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		h.Next = next
		return h
	})

	c.OnStartup(func() error {
		plugins := dnsserver.GetConfig(c).Handlers()
		for _, p := range plugins {
			if m, ok := p.(Metadater); ok {
				varNames := m.MetadataVarsAvailable()
				for _, name := range varNames {
					h.Metadaters[name] = m
				}
			}
		}
		fmt.Println(h.Metadaters)
		return nil
	})

	return nil
}
