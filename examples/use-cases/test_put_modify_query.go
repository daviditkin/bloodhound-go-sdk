// Copyright 2023 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	context "context"
	"fmt"
	"log"
	. "oapi-client/sdk"
)

func main() {
	// httpClient that handles localhost with subdomains (bloodhound.localhost)
	var customHttpClient, rerr = GetLocalhostWithSubdomainHttpClient()
	if rerr != nil {
		log.Fatal("Ooof cant make bloodhound.localhost resolving http.Client", rerr)
	}

	// API token
	var token = "GOc513qwNfLGAFWXlHX4o/kGa22mSogaFYm8hSH5aMbSC0oGRUmELA==" // Your API token
	var token_id = "54315a5a-13ff-470c-a5d6-3f18e6fbf41c"                  // Your API token id

	// HMAC Security Provider
	var hmacTokenProvider, serr = NewSecurityProviderHMACCredentials(token, token_id)

	if serr != nil {
		log.Fatal("Error creating hmac token middleware", serr)
	}
	client, crerr := NewClientWithResponses(
		"http://bloodhound.localhost/",
		WithRequestEditorFn(hmacTokenProvider.Intercept),
		WithBaseURL("http://bloodhound.localhost/"),
		WithHTTPClient(customHttpClient))
	if crerr != nil {
		log.Fatal("Error creating client", crerr)
	}

	cypherName := "foo"
	cypherQueryTxt := "MATCH p=shortestPath((u:AZUser)-[*1..]->(d:AZGroup {name: 'Admin Tier Zero'})) RETURN p"
	cypher := CreateSavedQueryJSONRequestBody{
		Name:  &cypherName,
		Query: &cypherQueryTxt,
	}

	response, err := client.ListSavedQueriesWithResponse(context.Background(), nil)
	if err != nil {
		log.Fatal("Error listing saved queries", response, err)
	}
	for _, v := range *response.JSON200.Data {
		if *v.Name == cypherName {
			// Remove it first?
			r, err := client.DeleteSavedQueryWithResponse(context.Background(), int32(*v.Id), nil)
			if err != nil || r.StatusCode() != 204 {
				log.Fatal("Error deleting saved query", r, err)
			}
		}
	}

	// Lets create it
	createQueryResp, err := client.CreateSavedQueryWithResponse(context.Background(), &CreateSavedQueryParams{}, cypher)
	if err != nil {
		return
	}

	if createQueryResp.StatusCode() != 200 {
		log.Fatal("Error creating query", createQueryResp.StatusCode(), createQueryResp.Status())
	}

	queryId := int32(*createQueryResp.JSON201.Data.Id)

	newName := "new name"
	var update = UpdateSavedQueryJSONRequestBody{
		Name: &newName,
	}

	client.UpdateSavedQueryWithResponse(context.Background(), queryId, nil, update)

	client.DeleteSavedQueryWithResponse(context.Background(), queryId, nil)
	// Get the API Version from the server
	var params = &GetApiVersionParams{}
	version, err := client.GetApiVersionWithResponse(context.Background(), params)
	if err != nil {
		log.Print("Error while getting api version", err)
		return
	}
	if version.StatusCode() == 200 {
		fmt.Printf("Version: %s\n", *version.JSON200.Data.ServerVersion)
	} else {
		log.Fatal("Error getting api version", version.StatusCode())
	}
}