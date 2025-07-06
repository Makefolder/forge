// Forge - Automated Docker container deployment tool for VPS environments.
// Monitors Git repositories and redeploys containers on new commits.
// Copyright (C) 2025 Artemii Fedotov
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.
package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

type HttpClient struct {
	httpClient *http.Client
}

func New(timeout time.Duration) *HttpClient {
	return &HttpClient{
		httpClient: &http.Client{Timeout: timeout},
	}
}

func (c *HttpClient) request(ctx context.Context, method string, url *url.URL, headers map[string]string, body any) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequestWithContext(ctx, method, url.String(), bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	return c.httpClient.Do(req)
}

func (c *HttpClient) Get(ctx context.Context, url *url.URL, headers map[string]string) (*http.Response, error) {
	return c.request(ctx, http.MethodGet, url, headers, nil)
}

func (c *HttpClient) Post(ctx context.Context, url *url.URL, headers map[string]string, body any) (*http.Response, error) {
	return c.request(ctx, http.MethodPost, url, headers, body)
}

func (c *HttpClient) Put(ctx context.Context, url *url.URL, headers map[string]string, body any) (*http.Response, error) {
	return c.request(ctx, http.MethodPut, url, headers, body)
}

func (c *HttpClient) Delete(ctx context.Context, url *url.URL, headers map[string]string) (*http.Response, error) {
	return c.request(ctx, http.MethodDelete, url, headers, nil)
}
