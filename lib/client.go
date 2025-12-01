// SPDX-FileCopyrightText: 2025 SUSE LLC
// SPDX-FileContributor: Cédric Bosdonnat
//
// SPDX-License-Identifier: Apache-2.0

package lib

import (
	"net/http"
	"net/http/cookiejar"

	"golang.org/x/net/publicsuffix"
)

const (
	url_base = "https://app.happy-compta.fr"
)

type Client struct {
	client *http.Client
}

// NemClient sets up a new happy-compta client.
func NewClient() (client *Client, err error) {
	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		return
	}
	client = &Client{
		client: &http.Client{Jar: jar},
	}
	return
}

func (c *Client) followRedirects(follow bool) {
	if follow {
		c.client.CheckRedirect = nil
	} else {
		c.client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}
}
