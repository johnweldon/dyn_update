package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strings"
	"time"

	"golang.org/x/net/publicsuffix"
)

// ZyxelResolver implements an IP resolver
type ZyxelResolver struct {
	username string
	password string
	baseurl  string
}

// NewZyxelResolver returns a ZyxelResolver
func NewZyxelResolver(baseURL string, username string, password string) *ZyxelResolver {
	return &ZyxelResolver{baseurl: baseURL, username: username, password: password}
}

// Find returns the public IP address according to the modem
func (z *ZyxelResolver) Find() (net.IP, error) {
	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		return net.IP{}, err
	}
	client := http.Client{Timeout: 30 * time.Second, Jar: jar}

	zyxURL, err := url.Parse(z.baseurl + "/modemstatus_connectionstatus.html")
	if err != nil {
		return net.IP{}, err
	}

	loginData := url.Values{"admin_username": {z.username}, "admin_password": {z.password}}
	resp, err := client.PostForm(z.baseurl+"/login.cgi", loginData)
	if err != nil {
		return net.IP{}, err
	}
	if len(jar.Cookies(zyxURL)) == 0 {
		return net.IP{}, fmt.Errorf("login to modem failed")
	}

	resp, err = client.Get(zyxURL.String())
	if err != nil {
		return net.IP{}, err
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return net.IP{}, err
	}

	re := regexp.MustCompile(`(?m:^var allStatus = "(.*)";\r?\n)`)
	matches := re.FindStringSubmatch(string(data))
	if len(matches) < 2 {
		return net.IP{}, fmt.Errorf("unable to find ip address, no match")
	}
	vars := strings.Split(matches[1], "||")
	if len(vars) < 9 {
		return net.IP{}, fmt.Errorf("unable to find ip address, unexpected vars")
	}

	ipv4 := net.ParseIP(vars[7])
	ipv6 := net.ParseIP(vars[8])
	log.Printf("IP addresses: %q, %q\n", ipv4, ipv6)
	return ipv4, nil
}
