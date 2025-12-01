// SPDX-FileCopyrightText: 2025 SUSE LLC
// SPDX-FileContributor: Cédric Bosdonnat
//
// SPDX-License-Identifier: Apache-2.0

package lib

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
)

// Login authenticates on happy-compta with given credentials.
func (c *Client) Login(email string, password string) error {
	token, err := c.getToken(url_base + "/auth/login")
	if err != nil {
		return err
	}

	values := url.Values{}
	values.Set("_token", token)
	values.Set("lastRequestUrl", "")
	values.Set("email", email)
	values.Set("password", password)
	values.Set("type", "0")
	values.Set("submit", "Connexion")

	resp, err := c.client.PostForm(url_base+"/auth/login", values)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return errors.New("failed to login")
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)
	if bytes.Contains(data, []byte("Connectez-vous")) {
		return errors.New("failed to login")
	}
	return nil
}

func (c *Client) getToken(url string) (token string, err error) {
	resp, err := c.client.Get(url)
	if err != nil {
		err = fmt.Errorf("failed to get the token: %s", err)
		return
	}

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("failed to get the token, HTTP err: %d", resp.StatusCode)
		return
	}

	body, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		err = fmt.Errorf("failed to read the token request body: %s", err)
		return
	}

	re := regexp.MustCompile(`<input name="_token" type="hidden" value="([^"]+)"`)
	matches := re.FindSubmatch(body)
	if len(matches) != 2 {
		err = errors.New("failed to find the token")
		return
	}
	token = string(matches[1])
	return
}
