package proxy

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/coredns/coredns/plugin/pkg/dnstest"
	"github.com/coredns/coredns/plugin/test"
	"github.com/mholt/caddy/caddyfile"
	"github.com/miekg/dns"
)

func TestStop(t *testing.T) {
	config := "proxy . %s {\n health_check /healthcheck:%s %dms \n}"
	tests := []struct {
		intervalInMilliseconds  int
		numHealthcheckIntervals int
	}{
		{5, 1},
		{5, 2},
		{5, 3},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("Test %d", i), func(t *testing.T) {

			// Set up proxy.
			var counter int64
			backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				r.Body.Close()
				atomic.AddInt64(&counter, 1)
			}))

			defer backend.Close()

			port := backend.URL[17:] // Remove all crap up to the port
			back := backend.URL[7:]  // Remove http://
			c := caddyfile.NewDispenser("Testfile", strings.NewReader(fmt.Sprintf(config, back, port, test.intervalInMilliseconds)))
			upstreams, err := NewStaticUpstreams(&c)
			if err != nil {
				t.Errorf("Test %d, expected no error. Got: %s", i, err)
			}

			// Give some time for healthchecks to hit the server.
			time.Sleep(time.Duration(test.intervalInMilliseconds*test.numHealthcheckIntervals) * time.Millisecond)

			for _, upstream := range upstreams {
				if err := upstream.Stop(); err != nil {
					t.Errorf("Test %d, expected no error stopping upstream, got: %s", i, err)
				}
			}

			counterAfterShutdown := atomic.LoadInt64(&counter)

			// Give some time to see if healthchecks are still hitting the server.
			time.Sleep(time.Duration(test.intervalInMilliseconds*test.numHealthcheckIntervals) * time.Millisecond)

			if counterAfterShutdown == 0 {
				t.Errorf("Test %d, Expected healthchecks to hit test server, got none", i)
			}

			// health checks are in a go routine now, so one may well occur after we shutdown,
			// but we only ever expect one more
			counterAfterWaiting := atomic.LoadInt64(&counter)
			if counterAfterWaiting > (counterAfterShutdown + 1) {
				t.Errorf("Test %d, expected no more healthchecks after shutdown. got: %d healthchecks after shutdown", i, counterAfterWaiting-counterAfterShutdown)
			}
		})
	}
}

func TestProxySequentialPolicy(t *testing.T) {
	// Set up valid DNS server.
	s := dnstest.NewServer(func(w dns.ResponseWriter, r *dns.Msg) {
		ret := new(dns.Msg)
		ret.SetReply(r)
		ret.Answer = append(ret.Answer, test.A("example.org. IN A 127.0.0.1"))
		w.WriteMsg(ret)
	})
	defer s.Close()

	// Set up port which will not respond. Emulates unreachable server (i/o timeout)
	u, _ := net.Listen("tcp", ":0")
	defer u.Close()

	// Set up proxy with 2 endpoints: 1st not responding, 2nd valid.
	// "force_tcp" is for simplification of unreachable server setup
	proxyStanza := fmt.Sprintf("proxy . %v %v {\n policy sequential\n protocol dns force_tcp\n}",
		u.Addr().String(), s.Addr)
	c := caddyfile.NewDispenser("Testfile", strings.NewReader(proxyStanza))
	upstreams, err := NewStaticUpstreams(&c)
	if err != nil {
		t.Errorf("Expected no error. Got: %s", err)
	}
	proxy := Proxy{Upstreams: &upstreams}

	// Call ServeDNS and make sure that proxy successfully switched to valid 2nd endpoind and got expected respponce
	req := new(dns.Msg)
	req.SetQuestion("example.org.", dns.TypeA)
	rrw := dnstest.NewRecorder(&test.ResponseWriter{})
	if _, err := proxy.ServeDNS(context.TODO(), rrw, req); err != nil {
		t.Fatalf("Expected no error. Error: %v", err.Error())
	}
	if len(rrw.Msg.Answer) != 1 {
		t.Fatalf("Expected exactly one RR in the answer section, got: %d", len(rrw.Msg.Answer))
	}
	if rrw.Msg.Answer[0].Header().Rrtype != dns.TypeA {
		t.Errorf("Expected RR to A, got: %d", rrw.Msg.Answer[0].Header().Rrtype)
	}
	if rrw.Msg.Answer[0].(*dns.A).A.String() != "127.0.0.1" {
		t.Errorf("Expected 127.0.0.1, got: %s", rrw.Msg.Answer[0].(*dns.A).A.String())
	}
}
