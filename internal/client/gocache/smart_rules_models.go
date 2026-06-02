// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package gocache

import "strconv"

// SmartRuleMatch holds the matching criteria shared by GoCache Smart Rules.
//
// Fields are pointers so that the client can distinguish between an attribute
// that was not configured (nil) and one that was explicitly set to its zero
// value. Only non-nil fields are sent to the API.
type SmartRuleMatch struct {
	Host            *string
	Scheme          *string
	RequestURI      *string
	Request         *string
	HTTPVersion     []string
	DeviceType      []string
	HTTPUserAgent   *string
	Cookie          *string
	CookieContent   *string
	RequestMethod   []string
	HTTPReferer     *string
	RemoteAddress   *string
	Header          *string
	OriginCountry   *string
	OriginContinent []string
	OriginASN       *int64
}

// GeneralSmartRuleAction holds the action parameters for a general Smart Rule.
type GeneralSmartRuleAction struct {
	SetHost                  *string
	Backend                  *string
	SetURI                   *string
	CustomCacheKey           *string
	ExpiresTTL               *int64
	CachingBehavior          *string
	CacheMode                *string
	CacheTTL                 *int64
	CORS                     *string
	SSLMode                  *string
	Connection               *string
	Hide                     *string
	Set                      *string
	SetReqHeader             []string
	SignedURLKey             *string
	SignedURLType            *string
	Cache301                 *bool
	Cache302                 *bool
	Cache404                 *bool
	GzipStatus               *bool
	IgnoreCacheControl       *bool
	IgnoreExpires            *bool
	IgnoreVary               *bool
	WAFStatus                *bool
	WAFMode                  *string
	WAFLevel                 *string
	RateLimitStatus          *bool
	ImageOptimize            *bool
	ImageOptimizeWebp        *bool
	ImageOptimizeProgressive *bool
	ImageOptimizeMetadata    *bool
	ImageOptimizeLevel       *int64
}

// RedirectSmartRuleAction holds the action parameters for a redirect Smart Rule.
type RedirectSmartRuleAction struct {
	RedirectType *int64
	RedirectTo   *string
}

// GeneralSmartRuleRequest is the payload used to create/update a general Smart Rule.
type GeneralSmartRuleRequest struct {
	Name   *string
	Status *bool
	Notes  *string
	Match  SmartRuleMatch
	Action GeneralSmartRuleAction
}

// RedirectSmartRuleRequest is the payload used to create/update a redirect Smart Rule.
type RedirectSmartRuleRequest struct {
	Name   *string
	Status *bool
	Notes  *string
	Match  SmartRuleMatch
	Action RedirectSmartRuleAction
}

// SmartRuleMetadata represents the metadata block returned by the API. The
// GoCache API serializes these values as strings (e.g. "true"/"false").
type SmartRuleMetadata struct {
	Name      string `json:"name,omitempty"`
	Status    string `json:"status,omitempty"`
	Notes     string `json:"notes,omitempty"`
	UpdatedOn string `json:"updated_on,omitempty"`
}

// generalSmartRuleAPI is the raw representation of a general Smart Rule as
// returned by the listing endpoint. Match/action values may come back either as
// a string or as an array of strings, so flexibleString/flexibleStringList are
// used to normalize them.
type generalSmartRuleAPI struct {
	ID       string            `json:"id"`
	Metadata SmartRuleMetadata `json:"metadata"`
	Match    struct {
		Host            flexibleString     `json:"host"`
		Scheme          flexibleString     `json:"scheme"`
		RequestURI      flexibleString     `json:"request_uri"`
		HTTPVersion     flexibleStringList `json:"http_version"`
		DeviceType      flexibleStringList `json:"device_type"`
		HTTPUserAgent   flexibleString     `json:"http_user_agent"`
		Cookie          flexibleString     `json:"cookie"`
		CookieContent   flexibleString     `json:"cookie_content"`
		RequestMethod   flexibleStringList `json:"request_method"`
		HTTPReferer     flexibleString     `json:"http_referer"`
		RemoteAddress   flexibleString     `json:"remote_address"`
		Header          flexibleString     `json:"header"`
		OriginCountry   flexibleString     `json:"origin_country"`
		OriginContinent flexibleStringList `json:"origin_continent"`
		OriginASN       flexibleString     `json:"origin_asn"`
	} `json:"match"`
	Action struct {
		SetHost                  flexibleString     `json:"set_host"`
		Backend                  flexibleString     `json:"backend"`
		SetURI                   flexibleString     `json:"set_uri"`
		CustomCacheKey           flexibleString     `json:"custom_cache_key"`
		ExpiresTTL               flexibleString     `json:"expires_ttl"`
		CachingBehavior          flexibleString     `json:"caching_behavior"`
		CacheMode                flexibleString     `json:"cache_mode"`
		CacheTTL                 flexibleString     `json:"cache_ttl"`
		CORS                     flexibleString     `json:"cors"`
		SSLMode                  flexibleString     `json:"ssl_mode"`
		Connection               flexibleString     `json:"connection"`
		Hide                     flexibleString     `json:"hide"`
		Set                      flexibleString     `json:"set"`
		SetReqHeader             flexibleStringList `json:"set_req_header"`
		SignedURLKey             flexibleString     `json:"signed_url_key"`
		SignedURLType            flexibleString     `json:"signed_url_type"`
		Cache301                 flexibleString     `json:"cache_301"`
		Cache302                 flexibleString     `json:"cache_302"`
		Cache404                 flexibleString     `json:"cache_404"`
		GzipStatus               flexibleString     `json:"gzip_status"`
		IgnoreCacheControl       flexibleString     `json:"ignore_cache_control"`
		IgnoreExpires            flexibleString     `json:"ignore_expires"`
		IgnoreVary               flexibleString     `json:"ignore_vary"`
		WAFStatus                flexibleString     `json:"waf_status"`
		WAFMode                  flexibleString     `json:"waf_mode"`
		WAFLevel                 flexibleString     `json:"waf_level"`
		RateLimitStatus          flexibleString     `json:"rate_limit_status"`
		ImageOptimize            flexibleString     `json:"image_optimize"`
		ImageOptimizeWebp        flexibleString     `json:"image_optimize_webp"`
		ImageOptimizeProgressive flexibleString     `json:"image_optimize_progressive"`
		ImageOptimizeMetadata    flexibleString     `json:"image_optimize_metadata"`
		ImageOptimizeLevel       flexibleString     `json:"image_optimize_level"`
	} `json:"action"`
}

// redirectSmartRuleAPI is the raw representation of a redirect Smart Rule as
// returned by the listing endpoint.
type redirectSmartRuleAPI struct {
	ID       string            `json:"id"`
	Metadata SmartRuleMetadata `json:"metadata"`
	Match    struct {
		Request         flexibleString     `json:"request"`
		HTTPVersion     flexibleStringList `json:"http_version"`
		DeviceType      flexibleStringList `json:"device_type"`
		HTTPUserAgent   flexibleString     `json:"http_user_agent"`
		Cookie          flexibleString     `json:"cookie"`
		CookieContent   flexibleString     `json:"cookie_content"`
		RequestMethod   flexibleStringList `json:"request_method"`
		HTTPReferer     flexibleString     `json:"http_referer"`
		RemoteAddress   flexibleString     `json:"remote_address"`
		Header          flexibleString     `json:"header"`
		OriginCountry   flexibleString     `json:"origin_country"`
		OriginContinent flexibleStringList `json:"origin_continent"`
	} `json:"match"`
	Action struct {
		RedirectType flexibleString `json:"redirect_type"`
		RedirectTo   flexibleString `json:"redirect_to"`
	} `json:"action"`
}

type smartRulesListResponse[T any] struct {
	StatusCode int `json:"status_code"`
	Response   struct {
		Rules []T `json:"rules"`
	} `json:"response"`
}

type smartRuleCreateResponse struct {
	StatusCode int `json:"status_code"`
	Response   struct {
		ID string `json:"id"`
	} `json:"response"`
}

func strPtr(f flexibleString) *string {
	if !f.Set {
		return nil
	}

	value := f.Value

	return &value
}

func listPtr(f flexibleStringList) []string {
	if !f.Set {
		return nil
	}

	return f.Values
}

func boolPtr(f flexibleString) *bool {
	if !f.Set {
		return nil
	}

	b, err := strconv.ParseBool(f.Value)

	if err != nil {
		return nil
	}

	return &b
}

func int64Ptr(f flexibleString) *int64 {
	if !f.Set {
		return nil
	}

	n, err := strconv.ParseInt(f.Value, 10, 64)

	if err != nil {
		return nil
	}

	return &n
}

func (r generalSmartRuleAPI) normalize() GeneralSmartRule {
	return GeneralSmartRule{
		ID:       r.ID,
		Metadata: r.Metadata,
		Action: GeneralSmartRuleAction{
			SetHost:                  strPtr(r.Action.SetHost),
			Backend:                  strPtr(r.Action.Backend),
			SetURI:                   strPtr(r.Action.SetURI),
			CustomCacheKey:           strPtr(r.Action.CustomCacheKey),
			ExpiresTTL:               int64Ptr(r.Action.ExpiresTTL),
			CachingBehavior:          strPtr(r.Action.CachingBehavior),
			CacheMode:                strPtr(r.Action.CacheMode),
			CacheTTL:                 int64Ptr(r.Action.CacheTTL),
			CORS:                     strPtr(r.Action.CORS),
			SSLMode:                  strPtr(r.Action.SSLMode),
			Connection:               strPtr(r.Action.Connection),
			Hide:                     strPtr(r.Action.Hide),
			Set:                      strPtr(r.Action.Set),
			SetReqHeader:             listPtr(r.Action.SetReqHeader),
			SignedURLKey:             strPtr(r.Action.SignedURLKey),
			SignedURLType:            strPtr(r.Action.SignedURLType),
			Cache301:                 boolPtr(r.Action.Cache301),
			Cache302:                 boolPtr(r.Action.Cache302),
			Cache404:                 boolPtr(r.Action.Cache404),
			GzipStatus:               boolPtr(r.Action.GzipStatus),
			IgnoreCacheControl:       boolPtr(r.Action.IgnoreCacheControl),
			IgnoreExpires:            boolPtr(r.Action.IgnoreExpires),
			IgnoreVary:               boolPtr(r.Action.IgnoreVary),
			WAFStatus:                boolPtr(r.Action.WAFStatus),
			WAFMode:                  strPtr(r.Action.WAFMode),
			WAFLevel:                 strPtr(r.Action.WAFLevel),
			RateLimitStatus:          boolPtr(r.Action.RateLimitStatus),
			ImageOptimize:            boolPtr(r.Action.ImageOptimize),
			ImageOptimizeWebp:        boolPtr(r.Action.ImageOptimizeWebp),
			ImageOptimizeProgressive: boolPtr(r.Action.ImageOptimizeProgressive),
			ImageOptimizeMetadata:    boolPtr(r.Action.ImageOptimizeMetadata),
			ImageOptimizeLevel:       int64Ptr(r.Action.ImageOptimizeLevel),
		},
	}
}

func (r redirectSmartRuleAPI) normalize() RedirectSmartRule {
	return RedirectSmartRule{
		ID:       r.ID,
		Metadata: r.Metadata,
		Action: RedirectSmartRuleAction{
			RedirectType: int64Ptr(r.Action.RedirectType),
			RedirectTo:   strPtr(r.Action.RedirectTo),
		},
	}
}
