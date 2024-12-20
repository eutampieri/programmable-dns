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
	Server string `json:"server"`
}

func (b BasicResolver) Resolve(q *dns.Msg) (*dns.Msg, error) {
	return dns.Exchange(q, b.Server)
}

type DoTResolver struct {
	Server string
	Client *dns.Client
}

func (t DoTResolver) Resolve(q *dns.Msg) (*dns.Msg, error) {
	response, _, err := t.Client.Exchange(q, t.Server)
	return response, err
}

func MakeDoTResolver(server string) DoTResolver {
	c := new(dns.Client)
	c.Net = "tcp-tls"
	c.SingleInflight = true
	return DoTResolver{Server: server, Client: c}
}

type StaticResolver struct {
	DomainsToIPs map[string]string `json:"domainsToIPs"`
	Base         string            `json:"base"`
}

func (s StaticResolver) Resolve(q *dns.Msg) (*dns.Msg, error) {
	msg := dns.Msg{}
	msg.SetReply(q)
	switch q.Question[0].Qtype {
	case dns.TypeA:
		msg.Authoritative = true
		domain := strings.ReplaceAll(q.Question[0].Name, "."+s.Base+".", "")
		address, ok := s.DomainsToIPs[domain]
		if ok {
			msg.Answer = append(msg.Answer, &dns.A{
				Hdr: dns.RR_Header{Name: q.Question[0].Name, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 60},
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

type SuffixResolver struct {
	Server    string `json:"server"`
	NewSuffix string `json:"newSuffix"`
	OldSuffix string `json:"oldSuffix"`
}

func (s SuffixResolver) Resolve(q *dns.Msg) (*dns.Msg, error) {
	for i := range q.Question {
		q.Question[i].Name = strings.ReplaceAll(q.Question[i].Name, s.NewSuffix, s.OldSuffix)
	}
	ans, err := dns.Exchange(q, s.Server)
	if err != nil {
		return nil, err
	}
	if q.Question[0].Qtype == dns.TypePTR {
		for i := range ans.Answer {
			if t, ok := ans.Answer[0].(*dns.PTR); ok {
				oldA := ans.Answer[i]
				ans.Answer[i] = &dns.PTR{
					Hdr: *oldA.Header(),
					Ptr: strings.ReplaceAll(t.Ptr, s.OldSuffix, s.NewSuffix),
				}
			}
		}
	} else {
		for i := range ans.Question {
			ans.Question[i].Name = strings.ReplaceAll(ans.Question[i].Name, s.OldSuffix, s.NewSuffix)
		}
		for i := range ans.Answer {
			ans.Answer[i].Header().Name = strings.ReplaceAll(ans.Answer[i].Header().Name, s.OldSuffix, s.NewSuffix)
		}
	}
	return ans, nil
}

type MergeResolver struct {
	Resolvers []Resolver
}

func (s MergeResolver) Resolve(q *dns.Msg) (*dns.Msg, error) {
	for _, r := range s.Resolvers {
		ans, err := r.Resolve(q)
		if err == nil && len(ans.Answer) > 0 {
			return ans, nil
		}
	}
	response := emptyDnsResponse(q)
	return &response, nil
}
