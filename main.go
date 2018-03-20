package main

import (
	"log"
	"net"
	"os"
)

type IPDiscoverer interface {
	Find() (net.IP, error)
}

type IPUpdater interface {
	Update(ip net.IP) error
}

func main() {
	disc := ZyxelLookup(os.Getenv("ZYX_MODEM_URL"), os.Getenv("ZYX_USERNAME"), os.Getenv("ZYX_PASSWORD"))
	ipv4, err := disc.Find()
	if err != nil {
		log.Fatal(err)
	}
	upd := GoogleDomainsUpdater(os.Getenv("GD_HOSTNAME"), os.Getenv("GD_USERNAME"), os.Getenv("GD_PASSWORD"))
	err = upd.Update(ipv4)
	if err != nil {
		log.Fatal(err)
	}
}
