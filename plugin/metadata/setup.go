package metadata

import (
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

	m := &Metadata{}

	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		m.Next = next
		return m
	})

	c.OnStartup(func() error {
		plugins := dnsserver.GetConfig(c).Handlers()
		// Collect all plugins which implement Metadater interface
		for _, p := range plugins {
			if met, ok := p.(Metadater); ok {
				m.Metadaters = append(m.Metadaters, met)
			}
		}
		return nil
	})

	return nil
}
