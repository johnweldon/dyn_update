package main

import (
	"context"
	"net"
)

const (
	opendnsResolver = "208.67.220.220"
)

// OpenDNSResolver is used to lookup public ip with OpenDNS
type OpenDNSResolver struct{}

// OpenDNS will return an OpenDNSIPResolver
func OpenDNS() OpenDNSResolver { return OpenDNSResolver{} }

// Find returns the apparent public IP
func (OpenDNSResolver) Find() (net.IP, error) {
	resolver := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{}
			return d.DialContext(ctx, "udp", net.JoinHostPort(opendnsResolver, "53"))
		},
	}
	ips, err := resolver.LookupHost(context.Background(), "myip.opendns.com")
	if err != nil {
		return net.IP{}, err
	}
	return net.ParseIP(ips[0]), nil
}
