package main

import (
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
