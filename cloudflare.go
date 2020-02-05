package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"
)

const (
	baseURL = "https://api.cloudflare.com/client/v4/"
)

// CloudFlareUpdater implements an updater for CloudFlare
type CloudFlareUpdater struct {
	authToken  string
	zoneID     string
	recordName string
	recordType string
	client     *http.Client
}

type cfPage struct {
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	Count      int `json:"count"`
	TotalCount int `json:"total_count"`
}

type cfError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type cfResponse struct {
	Result   interface{} `json:"result"`
	Success  bool        `json:"success"`
	Errors   []cfError   `json:"errors"`
	Messages []string    `json:"messages"`
	Info     cfPage      `json:"result_info"`
}

type cfRecord struct {
	Type    string `json:"type"`
	Name    string `json:"name"`
	Content string `json:"content"`
	TTL     int    `json:"ttl"`
}

func newResponse() *cfResponse {
	return &cfResponse{
		Result: []map[string]interface{}{},
	}
}

// NewCloudFlareUpdater returns an updater for CloudFlare
func NewCloudFlareUpdater(tok string, zoneID string, recordName string, recordType string) *CloudFlareUpdater {
	if len(recordType) == 0 {
		recordType = "A"
	}
	return &CloudFlareUpdater{
		authToken:  tok,
		zoneID:     zoneID,
		recordName: recordName,
		recordType: recordType,
		client: &http.Client{
			Transport: &http.Transport{
				IdleConnTimeout: 30 * time.Second,
			},
		},
	}
}

func (u *CloudFlareUpdater) newRequest(method string, URL string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, URL, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", u.authToken))
	return req, nil
}

func (u *CloudFlareUpdater) doRequest(method string, URL string, body io.Reader) (*cfResponse, error) {
	response := newResponse()
	req, err := u.newRequest(method, URL, body)
	if err != nil {
		return response, err
	}
	resp, err := u.client.Do(req)
	if err != nil {
		return response, err
	}
	defer resp.Body.Close()
	if err = json.NewDecoder(resp.Body).Decode(response); err != nil {
		return response, err
	}
	if !response.Success {
		for _, e := range response.Errors {
			if err != nil {
				err = fmt.Errorf("%d: %s %w", e.Code, e.Message, err)
			} else {
				err = fmt.Errorf("%d: %s", e.Code, e.Message)
			}
		}
		return response, err
	}
	return response, nil
}

func (u *CloudFlareUpdater) getRecord(name string, recordType string) (string, cfRecord, error) {
	rec := cfRecord{Name: name, Type: recordType, TTL: 1}
	id := ""
	dnsURL := fmt.Sprintf("%s/zones/%s/dns_records?name=%s&type=%s", baseURL, u.zoneID, name, recordType)
	response, err := u.doRequest("GET", dnsURL, nil)
	if err != nil {
		return id, rec, err
	}
	rm, ok := response.Result.([]interface{})
	if !ok {
		return id, rec, fmt.Errorf("unexpected response type: %T", response.Result)
	}
	if len(rm) == 0 {
		return id, rec, nil
	}
	first, ok := rm[0].(map[string]interface{})
	if !ok {
		return id, rec, fmt.Errorf("unexpected response type: %T", rm[0])
	}
	if n, ok := first["name"]; ok {
		rec.Name = n.(string)
	}
	if t, ok := first["type"]; ok {
		rec.Type = t.(string)
	}
	if c, ok := first["content"]; ok {
		rec.Content = c.(string)
	}
	if t, ok := first["ttl"]; ok {
		rec.TTL = int(t.(float64))
	}
	if i, ok := first["id"]; ok {
		id = i.(string)
	}
	return id, rec, nil
}

// Update updates the DNS record with the IP
func (u *CloudFlareUpdater) Update(ip net.IP) error {
	id, rec, err := u.getRecord(u.recordName, u.recordType)
	if err != nil {
		return err
	}
	if len(id) == 0 {
		return fmt.Errorf("missing record: %s", u.recordName)
	}
	rec.Content = ip.String()
	b, err := json.Marshal(rec)
	if err != nil {
		return err
	}
	body := bytes.NewBuffer(b)
	dnsURL := fmt.Sprintf("%s/zones/%s/dns_records/%s", baseURL, u.zoneID, id)
	response, err := u.doRequest("PUT", dnsURL, body)
	if err != nil {
		return err
	}
	if !response.Success {
		for _, e := range response.Errors {
			if err != nil {
				err = fmt.Errorf("%d: %s %w", e.Code, e.Message, err)
			} else {
				err = fmt.Errorf("%d: %s", e.Code, e.Message)
			}
		}
		return err
	}
	return nil
}
