package main

import (
	"errors"
	"log"
	"net"
	"strconv"
	"strings"

	"github.com/miekg/dns"
)

type handler struct{}

var resolvers []ResolverMapping

func reverse(numbers []string) []string {
	for i := 0; i < len(numbers)/2; i++ {
		j := len(numbers) - i - 1
		numbers[i], numbers[j] = numbers[j], numbers[i]
	}
	return numbers
}

func GetDNSServer(query string) (Resolver, error) {
	if strings.Contains(query, ".in-addr.arpa.") {
		query = strings.ReplaceAll(query, ".in-addr.arpa.", "")
		pieces := reverse(strings.Split(query, "."))
		query = strings.Join(pieces[:], ".")
	}
	ip := net.ParseIP(query)
	for _, resolver := range resolvers {
		if ip != nil {
			_, ipNetwork, _ := net.ParseCIDR(resolver.Network)
			if ipNetwork.Contains(ip) {
				return resolver.Resolver, nil
			}
		} else {
			if resolver.Domain != "" && strings.Contains(query, resolver.Domain) {
				return resolver.Resolver, nil
			}
		}
	}
	return nil, errors.New("zone not found")
}

func (h *handler) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	domain := r.Question[0].Name
	server, err := GetDNSServer(domain)
	if err == nil {
		in, err := server.Resolve(r)
		if err != nil {
			println(err.Error())
		} else {
			err := w.WriteMsg(in)
			if err != nil {
				println(err.Error())
			}
		}
	} else {
		response := emptyDnsResponse(r)
		err := w.WriteMsg(&response)
		if err != nil {
			println(err.Error())
		}
	}
}

func emptyDnsResponse(r *dns.Msg) dns.Msg {
	return dns.Msg{
		MsgHdr: dns.MsgHdr{
			Id:                 r.Id,
			Response:           true,
			Opcode:             r.Opcode,
			Authoritative:      false,
			Truncated:          false,
			RecursionDesired:   r.RecursionDesired,
			RecursionAvailable: false,
			Zero:               false,
			AuthenticatedData:  false,
			CheckingDisabled:   false,
			Rcode:              2,
		},
		Compress: false,
		Question: r.Question,
		Answer:   []dns.RR{},
		Ns:       nil,
		Extra:    nil,
	}
}

func main() {
	conf, err := LoadConfiguration("config.json")
	if err != nil {
		log.Fatal(err)
	}
	resolvers = conf
	srv := &dns.Server{Addr: "0.0.0.0:" + strconv.Itoa(5354), Net: "udp"}
	srv.Handler = &handler{}
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("Failed to set udp listener %s\n", err.Error())
	}
}
