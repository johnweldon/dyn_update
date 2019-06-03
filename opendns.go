package main

import "net"

type openDNSLookup struct{}

func OpenDNS() openDNSLookup { return openDNSLookup{} }

func (openDNSLookup) Find() (net.IP, error) {
	ips, err := net.LookupHost("myip.opendns.com")
	if err != nil {
		return net.IP{}, err
	}
	return net.ParseIP(ips[0]), nil
}
