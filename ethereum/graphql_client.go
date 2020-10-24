// Copyright 2020 Coinbase, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ethereum

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"time"
)

const (
	graphQLIdleConnectionTimeout = 30 * time.Second
	graphQLMaxIdle               = 100
	graphQLHTTPTimeout           = 15 * time.Second
	graphQLPath                  = "graphql"
)

// GraphQLClient is a client used to make graphQL
// queries to geth's graphql endpoint.
type GraphQLClient struct {
	client *http.Client
	url    string
}

// Query makes a query to the graphQL endpoint.
func (g *GraphQLClient) Query(ctx context.Context, input string) (string, error) {
	query := map[string]string{
		"query": input,
	}

	jsonValue, err := json.Marshal(query)
	if err != nil {
		return "", err
	}

	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		g.url,
		bytes.NewBuffer(jsonValue),
	)
	if err != nil {
		return "", err
	}

	response, err := g.client.Do(request)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func newGraphQLClient(baseURL string) (*GraphQLClient, error) {
	// Compute GraphQL Endpoint
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}
	u.Path = path.Join(u.Path, graphQLPath)

	// Setup HTTP Client
	client := &http.Client{
		Timeout: graphQLHTTPTimeout,
	}
	// Override transport idle connection settings
	//
	// See this conversation around why `.Clone()` is used here:
	// https://github.com/golang/go/issues/26013
	customTransport := http.DefaultTransport.(*http.Transport).Clone()
	customTransport.IdleConnTimeout = graphQLIdleConnectionTimeout
	customTransport.MaxIdleConns = graphQLMaxIdle
	customTransport.MaxIdleConnsPerHost = graphQLMaxIdle
	client.Transport = customTransport

	return &GraphQLClient{
		client: client,
		url:    u.String(),
	}, nil
}
