package main

import (
	"log"
	"net"
	"strconv"
	"strings"

	"github.com/miekg/dns"
)

type handler struct{}

//var defaultServer = DoTResolver{Server: "[2606:4700:4700::1112]:53"}
var defaultServer = DoTResolver{Server: "1.1.1.1:853"}

var ipsToSrvs = map[string]string
var hostnamesToSrvs = map[string]string
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
	println(query)
	ip := net.ParseIP(query)
	if ip != nil {
		for ipRange, server := range ipsToSrvs {
			_, ipNetwork, _ := net.ParseCIDR(ipRange)
			if ipNetwork.Contains(ip) {
				return server
			}
		}
		return defaultServer
	} else {
		for hostname, server := range hostnamesToSrvs {
			if strings.Contains(query, hostname) {
				return server
			}
		}
		if strings.Count(query, ".") > 1 {
			return defaultServer
		} else {
			// Here we should check where the request came from, but we use the default resolver anyway
			return defaultServer
		}
	}
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
	srv := &dns.Server{Addr: ":" + strconv.Itoa(53), Net: "udp"}
	srv.Handler = &handler{}
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("Failed to set udp listener %s\n", err.Error())
	}
}
