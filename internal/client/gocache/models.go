// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package gocache

import "encoding/json"

type DNSRecord struct {
	ID      json.Number `json:"record_id"`
	Type    string      `json:"type"`
	Name    string      `json:"name"`
	Content string      `json:"content"`
	TTL     string      `json:"ttl"`
	Cloud   string      `json:"cloud"`
}

type DNSRecordRequest struct {
	Type    string `json:"type"`
	Name    string `json:"name"`
	Content string `json:"content"`
	TTL     int64  `json:"ttl"`
	Cloud   int64  `json:"cloud"`
}

type DNSRecordsResponse struct {
	StatusCode int `json:"status_code"`
	Response   struct {
		Records []DNSRecord `json:"records"`
	} `json:"response"`
}

type DNSCreateResponse struct {
	StatusCode int `json:"status_code"`
	Response   struct {
		Records []DNSRecord `json:"records"`
	} `json:"response"`
}

type APIResponse struct {
	StatusCode int `json:"status_code"`
}
