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

func testAPICreateTeam(t *testing.T, session *TestSession, orgName, teamName string) *api.Team {
	token := getTokenForLoggedInUser(t, session, auth_model.AccessTokenScopeAll)
	req := NewRequestWithJSON(t, "POST", fmt.Sprintf("/api/v1/orgs/%s/teams", url.PathEscape(orgName)), api.CreateTeamOption{
		Name:       teamName,
		Permission: "read",
		UnitsMap:   map[string]string{"repo.code": "read"},
	}).AddTokenAuth(token)
	resp := MakeRequest(t, req, http.StatusCreated)

	var team api.Team
	err := json.Unmarshal(resp.Body.Bytes(), &team)
	require.NoError(t, err)
	return &team
}

func testAPIDeleteTeam(t *testing.T, session *TestSession, teamID int64) {
	token := getTokenForLoggedInUser(t, session, auth_model.AccessTokenScopeAll)
	req := NewRequest(t, "DELETE", fmt.Sprintf("/api/v1/teams/%d", teamID)).AddTokenAuth(token)
	MakeRequest(t, req, http.StatusNoContent)
}

func Test_WebhookTeamCreate(t *testing.T) {
	onGiteaRun(t, func(t *testing.T, giteaURL *url.URL) {
		var payloads []api.TeamPayload
		var triggeredEvent string

		provider := newMockWebhookProvider(func(r *http.Request) {
			content, _ := io.ReadAll(r.Body)
			var payload api.TeamPayload
			err := json.Unmarshal(content, &payload)
			assert.NoError(t, err)
			payloads = append(payloads, payload)
			triggeredEvent = r.Header.Get("X-Gitea-Event")
		}, http.StatusOK)
		defer provider.Close()

		// 1. Create a system webhook for team_create event
		session := loginUser(t, "user1") // user1 is admin
		hookID := testAPICreateSystemWebhook(t, session, provider.URL(), "team_create")
		defer testAPIDeleteSystemWebhook(t, session, hookID)

		// 2. Create an organization first
		testAPICreateOrg(t, session, "webhookteamorg")
		defer testAPIDeleteOrg(t, session, "webhookteamorg")

		// 3. Trigger the webhook by creating a team
		team := testAPICreateTeam(t, session, "webhookteamorg", "webhooktestteam")
		defer testAPIDeleteTeam(t, session, team.ID)

		// 4. Validate the webhook is triggered
		assert.Len(t, payloads, 1)
		assert.Equal(t, string(webhook_module.HookEventTeamCreate), triggeredEvent)
		assert.Equal(t, api.HookTeamCreated, payloads[0].Action)
		assert.Equal(t, "webhooktestteam", payloads[0].Team.Name)
		assert.NotNil(t, payloads[0].Organization)
		assert.Equal(t, "webhookteamorg", payloads[0].Organization.Name)
		assert.NotNil(t, payloads[0].Sender)
		assert.Equal(t, "user1", payloads[0].Sender.UserName)
	})
}

func Test_WebhookTeamDelete(t *testing.T) {
	onGiteaRun(t, func(t *testing.T, giteaURL *url.URL) {
		var payloads []api.TeamPayload
		var triggeredEvent string

		provider := newMockWebhookProvider(func(r *http.Request) {
			content, _ := io.ReadAll(r.Body)
			var payload api.TeamPayload
			err := json.Unmarshal(content, &payload)
			assert.NoError(t, err)
			payloads = append(payloads, payload)
			triggeredEvent = r.Header.Get("X-Gitea-Event")
		}, http.StatusOK)
		defer provider.Close()

		// 1. Create a system webhook for team_delete event
		session := loginUser(t, "user1") // user1 is admin
		hookID := testAPICreateSystemWebhook(t, session, provider.URL(), "team_delete")
		defer testAPIDeleteSystemWebhook(t, session, hookID)

		// 2. Create an organization and team
		testAPICreateOrg(t, session, "webhookteamdelorg")
		defer testAPIDeleteOrg(t, session, "webhookteamdelorg")

		team := testAPICreateTeam(t, session, "webhookteamdelorg", "webhookdelteam")

		// 3. Trigger the webhook by deleting the team
		testAPIDeleteTeam(t, session, team.ID)

		// 4. Validate the webhook is triggered
		assert.Len(t, payloads, 1)
		assert.Equal(t, string(webhook_module.HookEventTeamDelete), triggeredEvent)
		assert.Equal(t, api.HookTeamDeleted, payloads[0].Action)
		assert.Equal(t, "webhookdelteam", payloads[0].Team.Name)
		assert.NotNil(t, payloads[0].Organization)
		assert.Equal(t, "webhookteamdelorg", payloads[0].Organization.Name)
		assert.NotNil(t, payloads[0].Sender)
		assert.Equal(t, "user1", payloads[0].Sender.UserName)
	})
}
