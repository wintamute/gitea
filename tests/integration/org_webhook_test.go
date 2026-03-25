// Copyright 2024 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package integration

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"testing"

	auth_model "code.gitea.io/gitea/models/auth"
	"code.gitea.io/gitea/modules/json"
	api "code.gitea.io/gitea/modules/structs"
	webhook_module "code.gitea.io/gitea/modules/webhook"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testAPICreateOrg(t *testing.T, session *TestSession, orgName string) *api.Organization {
	token := getTokenForLoggedInUser(t, session, auth_model.AccessTokenScopeAll)
	req := NewRequestWithJSON(t, "POST", "/api/v1/orgs", api.CreateOrgOption{
		UserName:   orgName,
		Visibility: "public",
	}).AddTokenAuth(token)
	resp := MakeRequest(t, req, http.StatusCreated)

	var org api.Organization
	err := json.Unmarshal(resp.Body.Bytes(), &org)
	require.NoError(t, err)
	return &org
}

func testAPIDeleteOrg(t *testing.T, session *TestSession, orgName string) {
	token := getTokenForLoggedInUser(t, session, auth_model.AccessTokenScopeAll)
	req := NewRequest(t, "DELETE", fmt.Sprintf("/api/v1/orgs/%s", url.PathEscape(orgName))).AddTokenAuth(token)
	MakeRequest(t, req, http.StatusNoContent)
}

func Test_WebhookOrgCreate(t *testing.T) {
	onGiteaRun(t, func(t *testing.T, giteaURL *url.URL) {
		var payloads []api.OrganizationPayload
		var triggeredEvent string

		provider := newMockWebhookProvider(func(r *http.Request) {
			content, _ := io.ReadAll(r.Body)
			var payload api.OrganizationPayload
			err := json.Unmarshal(content, &payload)
			assert.NoError(t, err)
			payloads = append(payloads, payload)
			triggeredEvent = r.Header.Get("X-Gitea-Event")
		}, http.StatusOK)
		defer provider.Close()

		// 1. Create a system webhook for org_create event
		session := loginUser(t, "user1") // user1 is admin
		hookID := testAPICreateSystemWebhook(t, session, provider.URL(), "org_create")
		defer testAPIDeleteSystemWebhook(t, session, hookID)

		// 2. Trigger the webhook by creating an organization
		testAPICreateOrg(t, session, "webhooktestorg")
		defer testAPIDeleteOrg(t, session, "webhooktestorg")

		// 3. Validate the webhook is triggered
		assert.Len(t, payloads, 1)
		assert.Equal(t, string(webhook_module.HookEventOrgCreate), triggeredEvent)
		assert.Equal(t, api.HookOrgCreated, payloads[0].Action)
		assert.Equal(t, "webhooktestorg", payloads[0].Organization.Name)
		assert.NotNil(t, payloads[0].Sender)
		assert.Equal(t, "user1", payloads[0].Sender.UserName)
	})
}

func Test_WebhookOrgDelete(t *testing.T) {
	onGiteaRun(t, func(t *testing.T, giteaURL *url.URL) {
		var payloads []api.OrganizationPayload
		var triggeredEvent string

		provider := newMockWebhookProvider(func(r *http.Request) {
			content, _ := io.ReadAll(r.Body)
			var payload api.OrganizationPayload
			err := json.Unmarshal(content, &payload)
			assert.NoError(t, err)
			payloads = append(payloads, payload)
			triggeredEvent = r.Header.Get("X-Gitea-Event")
		}, http.StatusOK)
		defer provider.Close()

		// 1. Create a system webhook for org_delete event
		session := loginUser(t, "user1") // user1 is admin
		hookID := testAPICreateSystemWebhook(t, session, provider.URL(), "org_delete")
		defer testAPIDeleteSystemWebhook(t, session, hookID)

		// 2. First create an organization that we'll delete
		testAPICreateOrg(t, session, "webhookdeleteorg")

		// 3. Trigger the webhook by deleting the organization
		testAPIDeleteOrg(t, session, "webhookdeleteorg")

		// 4. Validate the webhook is triggered
		assert.Len(t, payloads, 1)
		assert.Equal(t, string(webhook_module.HookEventOrgDelete), triggeredEvent)
		assert.Equal(t, api.HookOrgDeleted, payloads[0].Action)
		assert.Equal(t, "webhookdeleteorg", payloads[0].Organization.Name)
		assert.NotNil(t, payloads[0].Sender)
		assert.Equal(t, "user1", payloads[0].Sender.UserName)
	})
}
