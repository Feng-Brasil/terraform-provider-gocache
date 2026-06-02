// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package gocache

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const (
	generalRulesPath  = "/rules/settings"
	redirectRulesPath = "/rules/rewrite"
)

// GeneralSmartRule is the normalized representation of a general Smart Rule
// returned by the GoCache API.
type GeneralSmartRule struct {
	ID       string
	Metadata SmartRuleMetadata
	Action   GeneralSmartRuleAction
}

// RedirectSmartRule is the normalized representation of a redirect Smart Rule
// returned by the GoCache API.
type RedirectSmartRule struct {
	ID       string
	Metadata SmartRuleMetadata
	Action   RedirectSmartRuleAction
}

func (c *Client) doSmartRuleForm(method, path string, form url.Values) ([]byte, error) {
	req, err := http.NewRequest(method, baseURL+path, strings.NewReader(form.Encode()))

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

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	return body, nil
}

func appendMatchForm(form url.Values, match SmartRuleMatch) {
	setStr := func(key string, value *string) {
		if value != nil {
			form.Set("match["+key+"]", *value)
		}
	}

	addList := func(key string, values []string) {
		for _, v := range values {
			form.Add("match["+key+"]", v)
		}
	}

	setStr("host", match.Host)
	setStr("scheme", match.Scheme)
	setStr("request_uri", match.RequestURI)
	setStr("request", match.Request)
	setStr("http_user_agent", match.HTTPUserAgent)
	setStr("cookie", match.Cookie)
	setStr("cookie_content", match.CookieContent)
	setStr("http_referer", match.HTTPReferer)
	setStr("remote_address", match.RemoteAddress)
	setStr("header", match.Header)
	setStr("origin_country", match.OriginCountry)

	addList("http_version", match.HTTPVersion)
	addList("device_type", match.DeviceType)
	addList("request_method", match.RequestMethod)
	addList("origin_continent", match.OriginContinent)

	if match.OriginASN != nil {
		form.Set("match[origin_asn]", strconv.FormatInt(*match.OriginASN, 10))
	}
}

func appendMetadataForm(form url.Values, name, notes *string, status *bool) {
	if name != nil {
		form.Set("metadata[name]", *name)
	}

	if notes != nil {
		form.Set("metadata[notes]", *notes)
	}

	if status != nil {
		form.Set("metadata[status]", strconv.FormatBool(*status))
	}
}

func appendGeneralActionForm(form url.Values, action GeneralSmartRuleAction) {
	setStr := func(key string, value *string) {
		if value != nil {
			form.Set("action["+key+"]", *value)
		}
	}

	setInt := func(key string, value *int64) {
		if value != nil {
			form.Set("action["+key+"]", strconv.FormatInt(*value, 10))
		}
	}

	setBool := func(key string, value *bool) {
		if value != nil {
			form.Set("action["+key+"]", strconv.FormatBool(*value))
		}
	}

	setStr("set_host", action.SetHost)
	setStr("backend", action.Backend)
	setStr("set_uri", action.SetURI)
	setStr("custom_cache_key", action.CustomCacheKey)
	setInt("expires_ttl", action.ExpiresTTL)
	setStr("caching_behavior", action.CachingBehavior)
	setStr("cache_mode", action.CacheMode)
	setInt("cache_ttl", action.CacheTTL)
	setStr("cors", action.CORS)
	setStr("ssl_mode", action.SSLMode)
	setStr("connection", action.Connection)
	setStr("hide", action.Hide)
	setStr("set", action.Set)
	setStr("signed_url_key", action.SignedURLKey)
	setStr("signed_url_type", action.SignedURLType)
	setBool("cache_301", action.Cache301)
	setBool("cache_302", action.Cache302)
	setBool("cache_404", action.Cache404)
	setBool("gzip_status", action.GzipStatus)
	setBool("ignore_cache_control", action.IgnoreCacheControl)
	setBool("ignore_expires", action.IgnoreExpires)
	setBool("ignore_vary", action.IgnoreVary)
	setBool("waf_status", action.WAFStatus)
	setStr("waf_mode", action.WAFMode)
	setStr("waf_level", action.WAFLevel)
	setBool("rate_limit_status", action.RateLimitStatus)
	setBool("image_optimize", action.ImageOptimize)
	setBool("image_optimize_webp", action.ImageOptimizeWebp)
	setBool("image_optimize_progressive", action.ImageOptimizeProgressive)
	setBool("image_optimize_metadata", action.ImageOptimizeMetadata)
	setInt("image_optimize_level", action.ImageOptimizeLevel)

	for i, header := range action.SetReqHeader {
		form.Set("action[set_req_header]["+strconv.Itoa(i)+"]", header)
	}
}

// CreateGeneralSmartRule creates a general Smart Rule for a domain and returns
// the generated rule ID.
func (c *Client) CreateGeneralSmartRule(domain string, rule GeneralSmartRuleRequest) (string, error) {
	form := url.Values{}

	appendMetadataForm(form, rule.Name, rule.Notes, rule.Status)
	appendMatchForm(form, rule.Match)
	appendGeneralActionForm(form, rule.Action)

	body, err := c.doSmartRuleForm("POST", generalRulesPath+"/"+domain, form)

	if err != nil {
		return "", fmt.Errorf("failed to create general smart rule: %w", err)
	}

	var response smartRuleCreateResponse

	if err := json.Unmarshal(body, &response); err != nil {
		return "", err
	}

	return response.Response.ID, nil
}

// UpdateGeneralSmartRule updates an existing general Smart Rule.
func (c *Client) UpdateGeneralSmartRule(domain, id string, rule GeneralSmartRuleRequest) error {
	form := url.Values{}

	appendMetadataForm(form, rule.Name, rule.Notes, rule.Status)
	appendMatchForm(form, rule.Match)
	appendGeneralActionForm(form, rule.Action)

	_, err := c.doSmartRuleForm("PUT", generalRulesPath+"/"+domain+"/"+id, form)

	if err != nil {
		return fmt.Errorf("failed to update general smart rule: %w", err)
	}

	return nil
}

// ListGeneralSmartRules lists all general Smart Rules for a domain.
func (c *Client) ListGeneralSmartRules(domain string) ([]GeneralSmartRule, error) {
	body, err := c.doSmartRuleForm("GET", generalRulesPath+"/"+domain, url.Values{})

	if err != nil {
		return nil, fmt.Errorf("failed listing general smart rules: %w", err)
	}

	var response smartRulesListResponse[generalSmartRuleAPI]

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}

	rules := make([]GeneralSmartRule, 0, len(response.Response.Rules))

	for _, raw := range response.Response.Rules {
		rules = append(rules, raw.normalize())
	}

	return rules, nil
}

// DeleteGeneralSmartRule removes a general Smart Rule.
func (c *Client) DeleteGeneralSmartRule(domain, id string) error {
	_, err := c.doSmartRuleForm("DELETE", generalRulesPath+"/"+domain+"/"+id, url.Values{})

	if err != nil {
		return fmt.Errorf("failed deleting general smart rule: %w", err)
	}

	return nil
}

func appendRedirectActionForm(form url.Values, action RedirectSmartRuleAction) {
	if action.RedirectType != nil {
		form.Set("action[redirect_type]", strconv.FormatInt(*action.RedirectType, 10))
	}

	if action.RedirectTo != nil {
		form.Set("action[redirect_to]", *action.RedirectTo)
	}
}

// CreateRedirectSmartRule creates a redirect Smart Rule for a domain and
// returns the generated rule ID.
func (c *Client) CreateRedirectSmartRule(domain string, rule RedirectSmartRuleRequest) (string, error) {
	form := url.Values{}

	appendMetadataForm(form, rule.Name, rule.Notes, rule.Status)
	appendMatchForm(form, rule.Match)
	appendRedirectActionForm(form, rule.Action)

	body, err := c.doSmartRuleForm("POST", redirectRulesPath+"/"+domain, form)

	if err != nil {
		return "", fmt.Errorf("failed to create redirect smart rule: %w", err)
	}

	var response smartRuleCreateResponse

	if err := json.Unmarshal(body, &response); err != nil {
		return "", err
	}

	return response.Response.ID, nil
}

// UpdateRedirectSmartRule updates an existing redirect Smart Rule.
func (c *Client) UpdateRedirectSmartRule(domain, id string, rule RedirectSmartRuleRequest) error {
	form := url.Values{}

	appendMetadataForm(form, rule.Name, rule.Notes, rule.Status)
	appendMatchForm(form, rule.Match)
	appendRedirectActionForm(form, rule.Action)

	_, err := c.doSmartRuleForm("PUT", redirectRulesPath+"/"+domain+"/"+id, form)

	if err != nil {
		return fmt.Errorf("failed to update redirect smart rule: %w", err)
	}

	return nil
}

// ListRedirectSmartRules lists all redirect Smart Rules for a domain.
func (c *Client) ListRedirectSmartRules(domain string) ([]RedirectSmartRule, error) {
	body, err := c.doSmartRuleForm("GET", redirectRulesPath+"/"+domain, url.Values{})

	if err != nil {
		return nil, fmt.Errorf("failed listing redirect smart rules: %w", err)
	}

	var response smartRulesListResponse[redirectSmartRuleAPI]

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}

	rules := make([]RedirectSmartRule, 0, len(response.Response.Rules))

	for _, raw := range response.Response.Rules {
		rules = append(rules, raw.normalize())
	}

	return rules, nil
}

// DeleteRedirectSmartRule removes a redirect Smart Rule.
func (c *Client) DeleteRedirectSmartRule(domain, id string) error {
	_, err := c.doSmartRuleForm("DELETE", redirectRulesPath+"/"+domain+"/"+id, url.Values{})

	if err != nil {
		return fmt.Errorf("failed deleting redirect smart rule: %w", err)
	}

	return nil
}
