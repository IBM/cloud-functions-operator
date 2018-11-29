/*
 * Copyright 2017-2018 IBM Corporation
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package common

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"k8s.io/api/core/v1"

	context "github.com/ibm/cloud-operators/pkg/context"
)

// CompositionClient represents a client to the composition service
type CompositionClient struct {
	APIKey string
}

// NewCompositionClient creates a new composition service client
func NewCompositionClient(ctx context.Context, owctx *v1.SecretEnvSource) (*CompositionClient, error) {
	secretName := "seed-defaults-owprops"
	if owctx != nil && owctx.Name != "" {
		secretName = owctx.Name
	}
	config, err := GetWskPropertiesFromSecret(ctx, secretName)
	if err != nil {
		return nil, fmt.Errorf("error getting Cloud Function API Key: %v", err)
	}
	return &CompositionClient{
		APIKey: config.WskCliAuthKey,
	}, nil
}

// Get returns a composition
func (client CompositionClient) Get(name string) (map[string]interface{}, *http.Response, error) {
	url, err := compositionURL(name)
	if err != nil {
		return nil, nil, err
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil { // in case the URL is malformed
		return nil, nil, err
	}

	return client.authDo(req, "application/json")
}

// Update updates a composition
func (client CompositionClient) Update(name string, composition string, contentType string) (map[string]interface{}, *http.Response, error) {
	return client.put(name, composition, contentType, true)
}

func (client CompositionClient) put(name string, composition string, contentType string, overwrite bool) (map[string]interface{}, *http.Response, error) {
	url, err := compositionURL(name)
	if err != nil {
		return nil, nil, err
	}
	if overwrite {
		url += "?overwrite=true"
	}

	req, err := http.NewRequest("PUT", url, strings.NewReader(composition))
	if err != nil { // in case the URL is malformed
		return nil, nil, err
	}

	return client.authDo(req, contentType)
}

// Invoke invokes a composition
func (client CompositionClient) Invoke(name string, params interface{}) (map[string]interface{}, *http.Response, error) {
	url, err := compositionURL(name)
	if err != nil {
		return nil, nil, err
	}
	url += "?blocking=true"

	var buf io.ReadWriter
	buf = new(bytes.Buffer)
	if params != nil {
		encoder := json.NewEncoder(buf)
		encoder.SetEscapeHTML(false)
		err := encoder.Encode(params)

		if err != nil {
			return nil, nil, err
		}
	} else {
		buf.Write([]byte("{}"))
	}

	req, err := http.NewRequest("POST", url, buf)
	if err != nil { // in case the URL is malformed
		return nil, nil, err
	}
	return client.authDo(req, "application/json")
}

// Delete deletes a composition
func (client CompositionClient) Delete(name string) (map[string]interface{}, *http.Response, error) {
	url, err := compositionURL(name)
	if err != nil {
		return nil, nil, err
	}
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil { // in case the URL is malformed
		return nil, nil, err
	}
	return client.authDo(req, "application/json")
}

func (client CompositionClient) authDo(req *http.Request, contentType string) (map[string]interface{}, *http.Response, error) {
	req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(client.APIKey)))
	req.Header.Set("Content-Type", contentType)

	httpClient := http.Client{
		Timeout: time.Second * 30,
	}
	response, err := httpClient.Do(req)
	if err != nil {
		return nil, response, err
	}
	defer response.Body.Close()

	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, response, err
	}

	var parsed interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		return nil, response, err
	}

	if response.StatusCode >= 400 {
		return parsed.(map[string]interface{}), response, fmt.Errorf("HTTP failure: %s", string(data))
	}

	return parsed.(map[string]interface{}), nil, nil
}

func compositionURL(fullname string) (string, error) {
	parts := strings.Split(fullname, "/")
	l := len(parts)
	if l != 1 && l != 2 && l != 4 {
		return "", fmt.Errorf("Malformed composition name %s", fullname)
	}
	name := parts[l-1]
	pkgName := "_"
	if l == 2 {
		// pkgName/name
		pkgName = parts[0]
	} else if l == 4 {
		// /namespace/pkgName/name
		pkgName = parts[2]
	}

	// TODO: custom sherpa endpoint
	return "https://sherpa-stage.wdpdist.com/v1/namespaces/_/packages/" + url.PathEscape(pkgName) + "/compositions/" + url.PathEscape(name), nil
}
