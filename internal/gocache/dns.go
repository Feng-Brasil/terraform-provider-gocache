package gocache

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

func (c *Client) CreateDNSRecord(domain string, record DNSRecordRequest) (*DNSRecord, error) {
	form := url.Values{}

	form.Set("name", record.Name)
	form.Set("type", record.Type)
	form.Set("content", record.Content)
	form.Set("ttl", strconv.FormatInt(record.TTL, 10))
	form.Set("cloud", strconv.FormatInt(record.Cloud, 10))

	req, err := http.NewRequest(
		"POST",
		baseURL+"/dns/"+domain,
		strings.NewReader(form.Encode()),
	)

	if err != nil {
		return nil, err
	}

	req.Header.Set("GoCache-Token", c.Token)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.HTTPClient.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("failed to create DNS record, status code: %d", resp.StatusCode)
	}

	var response DNSCreateResponse

	err = json.NewDecoder(resp.Body).Decode(&response)

	if err != nil {
		return nil, err
	}

	if len(response.Response.Records) == 0 {
		return nil, fmt.Errorf(
			"DNS record was created but API did not return record data",
		)
	}

	return &response.Response.Records[0], nil
}

func (c *Client) ListDNSRecords(domain string) ([]DNSRecord, error) {
	req, err := http.NewRequest(
		"GET",
		baseURL+"/dns/"+domain,
		nil,
	)

	if err != nil {
		return nil, err
	}

	req.Header.Set("GoCache-Token", c.Token)

	resp, err := c.HTTPClient.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("failed listing DNS records, status code: %d", resp.StatusCode)
	}

	var response DNSRecordsResponse

	err = json.NewDecoder(resp.Body).Decode(&response)

	if err != nil {
		return nil, err
	}

	return response.Response.Records, nil
}

func (c *Client) UpdateDNSRecord(domain string, recordID string, record DNSRecordRequest) error {
	form := url.Values{}

	form.Set("record_id", recordID)
	form.Set("name", record.Name)
	form.Set("type", record.Type)
	form.Set("content", record.Content)
	form.Set("ttl", strconv.FormatInt(record.TTL, 10))
	form.Set("cloud", strconv.FormatInt(record.Cloud, 10))

	req, err := http.NewRequest(
		"PUT",
		baseURL+"/dns/"+domain,
		strings.NewReader(form.Encode()),
	)

	if err != nil {
		return err
	}

	req.Header.Set("GoCache-Token", c.Token)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.HTTPClient.Do(req)

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("failed updating DNS record, status code: %d", resp.StatusCode)
	}

	return nil
}

func (c *Client) DeleteDNSRecord(domain string, recordID string) error {
	form := url.Values{}

	form.Set("record_id", recordID)

	req, err := http.NewRequest(
		"DELETE",
		baseURL+"/dns/"+domain,
		strings.NewReader(form.Encode()),
	)

	if err != nil {
		return err
	}

	req.Header.Set("GoCache-Token", c.Token)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.HTTPClient.Do(req)

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("failed deleting DNS record, status code: %d", resp.StatusCode)
	}

	return nil
}
