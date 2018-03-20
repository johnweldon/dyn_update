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

type gdUpdater struct {
	hostname string
	username string
	password string
}

func GoogleDomainsUpdater(hostname string, username string, password string) IPUpdater {
	return &gdUpdater{hostname: hostname, username: username, password: password}
}

func (g *gdUpdater) Update(ip net.IP) error {
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
		return fmt.Errorf("failed to update:\n%q\n", string(data))
	}
	return fmt.Errorf("failed to update")
}

func (g *gdUpdater) needsUpdate(ip net.IP) bool {
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
