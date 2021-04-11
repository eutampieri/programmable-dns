package main

import (
	"log"
	"net"
	"strconv"
	"strings"

	"github.com/miekg/dns"
)

type handler struct{}

var defaultServer = DoTResolver{Server: "[2606:4700:4700::1111]:853"}

var resolvers []ResolverMapping

func reverse(numbers []string) []string {
	for i := 0; i < len(numbers)/2; i++ {
		j := len(numbers) - i - 1
		numbers[i], numbers[j] = numbers[j], numbers[i]
	}
	return numbers
}

func GetDNSServer(query string) Resolver {
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
				return resolver.Resolver
			}
		} else {
			if strings.Contains(query, resolver.Domain) {
				return resolver.Resolver
			}
		}
	}
	return defaultServer
}

func (h *handler) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	domain := r.Question[0].Name
	server := GetDNSServer(domain)
	in, err := server.Resolve(r)
	if err != nil {
		println(err.Error())
	} else {
		w.WriteMsg(in)
	}
}

func main() {
	srv := &dns.Server{Addr: "127.0.0.1:" + strconv.Itoa(5354), Net: "udp"}
	srv.Handler = &handler{}
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("Failed to set udp listener %s\n", err.Error())
	}
}
