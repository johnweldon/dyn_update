package main

import (
	"fmt"
	"log"
	"net"
	"os"
)

// IPDiscoverer will find the apparent public IP
type IPDiscoverer interface {
	Find() (net.IP, error)
}

// IPUpdater will update a DNS record
type IPUpdater interface {
	Update(ip net.IP) error
}

type nilUpdater struct{}

func (*nilUpdater) Update(ip net.IP) error {
	return fmt.Errorf("current ip: %s. update not implemented", ip)
}

func main() {
	disc := OpenDNS()
	ipv4, err := disc.Find()
	if err != nil {
		log.Fatal(err)
	}

	var upd IPUpdater = &nilUpdater{}
	host := ""

	gdRecord := os.Getenv("GD_HOSTNAME")
	gdUser := os.Getenv("GD_USERNAME")
	gdPass := os.Getenv("GD_PASSWORD")
	if len(gdRecord) != 0 && len(gdUser) != 0 && len(gdPass) != 0 {
		upd = NewGoogleDomainsUpdater(gdRecord, gdUser, gdPass)
		host = gdRecord
		err = upd.Update(ipv4)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Updated IP address of %s on Google Domains: %s", host, ipv4)
	}

	cfRecord := os.Getenv("CF_HOSTNAME")
	cfToken := os.Getenv("CF_TOKEN")
	cfZoneID := os.Getenv("CF_ZONE_ID")
	if len(cfRecord) != 0 && len(cfToken) != 0 && len(cfZoneID) != 0 {
		upd = NewCloudFlareUpdater(cfToken, cfZoneID, cfRecord, "A")
		host = cfRecord
		err = upd.Update(ipv4)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Updated IP address of %s on CloudFlare: %s", host, ipv4)
	}

}
