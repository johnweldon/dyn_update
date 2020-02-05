package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"time"
)

// GoogleDomainsDNSUpdater implements a DNS updater for Google Domains
type GoogleDomainsDNSUpdater struct {
	hostname string
	username string
	password string
}

// NewGoogleDomainsUpdater returns an updater for Google Domains
func NewGoogleDomainsUpdater(hostname string, username string, password string) *GoogleDomainsDNSUpdater {
	return &GoogleDomainsDNSUpdater{hostname: hostname, username: username, password: password}
}

// Update will update the DNS record
func (g *GoogleDomainsDNSUpdater) Update(ip net.IP) error {
	if !g.needsUpdate(ip) {
		log.Printf("IP %q is current", ip.String())
		return nil
	}

	client := http.Client{Timeout: 30 * time.Second}
	ipurl, err := url.Parse("https://domains.google.com/nic/update")
	if err != nil {
		return err
	}
	q := ipurl.Query()
	q.Set("hostname", g.hostname)
	q.Set("myip", ip.String())
	ipurl.RawQuery = q.Encode()

	req, err := http.NewRequest("GET", ipurl.String(), nil)
	if err != nil {
		return err
	}
	req.SetBasicAuth(g.username, g.password)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	switch resp.StatusCode {
	case http.StatusOK:
		switch string(data) {
		case "badauth":
			return fmt.Errorf("missing auth")
		case "notfqdn":
			return fmt.Errorf("missing hostname to update")
		default:
			log.Printf("updated: %q", string(data))
			return nil
		}
	default:
		return fmt.Errorf("failed to update:\n%s", string(data))
	}
}

func (g *GoogleDomainsDNSUpdater) needsUpdate(ip net.IP) bool {
	if ips, err := net.LookupIP(g.hostname); err == nil {
		if len(ips) > 0 {
			for _, i := range ips {
				if i.Equal(ip) {
					return false
				}
			}
		}
	}
	return true
}
