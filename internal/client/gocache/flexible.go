// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package gocache

import (
	"bytes"
	"encoding/json"
	"strconv"
)

// flexibleString unmarshals a JSON value that the GoCache API may return as a
// string, a number or a boolean into a Go string. This is needed because the
// Smart Rules API serializes most values as strings, but is not always
// consistent across fields.
type flexibleString struct {
	Value string
	Set   bool
}

func (f *flexibleString) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, []byte("null")) {
		return nil
	}

	var s string

	if err := json.Unmarshal(data, &s); err == nil {
		f.Value = s
		f.Set = true

		return nil
	}

	var b bool

	if err := json.Unmarshal(data, &b); err == nil {
		f.Value = strconv.FormatBool(b)
		f.Set = true

		return nil
	}

	var n json.Number

	if err := json.Unmarshal(data, &n); err == nil {
		f.Value = n.String()
		f.Set = true

		return nil
	}

	return nil
}

// flexibleStringList unmarshals a JSON value that the GoCache API may return as
// either a single string or an array of strings into a []string.
type flexibleStringList struct {
	Values []string
	Set    bool
}

func (f *flexibleStringList) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, []byte("null")) {
		return nil
	}

	var list []string

	if err := json.Unmarshal(data, &list); err == nil {
		f.Values = list
		f.Set = true

		return nil
	}

	var single string

	if err := json.Unmarshal(data, &single); err == nil {
		f.Values = []string{single}
		f.Set = true

		return nil
	}

	return nil
}
