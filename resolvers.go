package main

import (
	"net"
	"strings"

	"github.com/miekg/dns"
)

type Resolver interface {
	Resolve(q *dns.Msg) (*dns.Msg, error)
}

type BasicResolver struct {
	Server string
}

func (b BasicResolver) Resolve(q *dns.Msg) (*dns.Msg, error) {
	return dns.Exchange(q, b.Server)
}

type DoTResolver struct {
	Server string
}

func (t DoTResolver) Resolve(q *dns.Msg) (*dns.Msg, error) {
	c := new(dns.Client)
	c.Net = "tcp-tls"
	response, _, err := c.Exchange(q, t.Server)
	return response, err
}

type StaticResolver struct {
	DomainsToIPs map[string]string
	Base         string
}

func (s StaticResolver) Resolve(q *dns.Msg) (*dns.Msg, error) {
	msg := dns.Msg{}
	msg.SetReply(q)
	switch q.Question[0].Qtype {
	case dns.TypeA:
		msg.Authoritative = true
		domain := strings.ReplaceAll(q.Question[0].Name, s.Base+".", "")
		address, ok := s.DomainsToIPs[domain+"."+s.Base]
		if ok {
			msg.Answer = append(msg.Answer, &dns.A{
				Hdr: dns.RR_Header{Name: domain, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 60},
				A:   net.ParseIP(address),
			})
		}
	case dns.TypePTR:
		msg.Authoritative = true
		queryIp := msg.Question[0].Name
		queryIp = strings.ReplaceAll(queryIp, ".in-addr.arpa.", "")
		pieces := reverse(strings.Split(queryIp, "."))
		queryIp = strings.Join(pieces[:], ".")
		for domain, ip := range s.DomainsToIPs {
			if queryIp == ip {
				msg.Answer = append(msg.Answer, &dns.PTR{
					Hdr: dns.RR_Header{Name: msg.Question[0].Name, Rrtype: dns.TypePTR, Class: dns.ClassINET, Ttl: 60},
					Ptr: domain + "." + s.Base + ".",
				})
				break
			}
		}
	}
	return &msg, nil
}

type ResolverMapping struct {
	Resolver Resolver
	Domain   string
	Network  string
}
