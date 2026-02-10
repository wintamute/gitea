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

func Pointer[T any](d T) *T {
	return &d
}

func testAPICreateSystemWebhook(t *testing.T, session *TestSession, url, event string) int64 {
	token := getTokenForLoggedInUser(t, session, auth_model.AccessTokenScopeAll)
	req := NewRequestWithJSON(t, "POST", "/api/v1/admin/hooks", api.CreateHookOption{
		Type: "gitea",
		Config: api.CreateHookOptionConfig{
			"content_type": "json",
			"url":          url,
		},
		Events: []string{event},
		Active: true,
	}).AddTokenAuth(token)
	resp := MakeRequest(t, req, http.StatusCreated)

	var hook api.Hook
	err := json.Unmarshal(resp.Body.Bytes(), &hook)
	require.NoError(t, err)
	return hook.ID
}

func testAPIDeleteSystemWebhook(t *testing.T, session *TestSession, hookID int64) {
	token := getTokenForLoggedInUser(t, session, auth_model.AccessTokenScopeAll)
	req := NewRequest(t, "DELETE", fmt.Sprintf("/api/v1/admin/hooks/%d", hookID)).AddTokenAuth(token)
	MakeRequest(t, req, http.StatusNoContent)
}

func testAPICreateUser(t *testing.T, session *TestSession, username, email, password string) *api.User {
	token := getTokenForLoggedInUser(t, session, auth_model.AccessTokenScopeAll)
	req := NewRequestWithJSON(t, "POST", "/api/v1/admin/users", api.CreateUserOption{
		Username:           username,
		Email:              email,
		Password:           password,
		MustChangePassword: Pointer(false),
	}).AddTokenAuth(token)
	resp := MakeRequest(t, req, http.StatusCreated)

	var user api.User
	err := json.Unmarshal(resp.Body.Bytes(), &user)
	require.NoError(t, err)
	return &user
}

func testAPIDeleteUser(t *testing.T, session *TestSession, username string) {
	token := getTokenForLoggedInUser(t, session, auth_model.AccessTokenScopeAll)
	req := NewRequest(t, "DELETE", "/api/v1/admin/users/"+url.PathEscape(username)+"?purge=true").AddTokenAuth(token)
	MakeRequest(t, req, http.StatusNoContent)
}

func testAPIEditUser(t *testing.T, session *TestSession, username string, opts api.EditUserOption) {
	token := getTokenForLoggedInUser(t, session, auth_model.AccessTokenScopeAll)
	req := NewRequestWithJSON(t, "PATCH", "/api/v1/admin/users/"+url.PathEscape(username), opts).AddTokenAuth(token)
	MakeRequest(t, req, http.StatusOK)
}

func Test_WebhookUserCreate(t *testing.T) {
	onGiteaRun(t, func(t *testing.T, giteaURL *url.URL) {
		var payloads []api.UserPayload
		var triggeredEvent string

		provider := newMockWebhookProvider(func(r *http.Request) {
			content, _ := io.ReadAll(r.Body)
			var payload api.UserPayload
			err := json.Unmarshal(content, &payload)
			assert.NoError(t, err)
			payloads = append(payloads, payload)
			triggeredEvent = r.Header.Get("X-Gitea-Event")
		}, http.StatusOK)
		defer provider.Close()

		// 1. Create a system webhook for user_create event
		session := loginUser(t, "user1") // user1 is admin
		hookID := testAPICreateSystemWebhook(t, session, provider.URL(), "user_create")
		defer testAPIDeleteSystemWebhook(t, session, hookID)

		// 2. Trigger the webhook by creating a user
		testAPICreateUser(t, session, "webhooktestuser", "webhooktestuser@example.com", "password123!")

		// Clean up the created user
		defer testAPIDeleteUser(t, session, "webhooktestuser")

		// 3. Validate the webhook is triggered
		assert.Len(t, payloads, 1)
		assert.Equal(t, string(webhook_module.HookEventUserCreate), triggeredEvent)
		assert.Equal(t, api.HookUserCreated, payloads[0].Action)
		assert.Equal(t, "webhooktestuser", payloads[0].User.UserName)
		assert.Equal(t, "webhooktestuser@example.com", payloads[0].User.Email)
		assert.NotNil(t, payloads[0].Sender)
		assert.Equal(t, "user1", payloads[0].Sender.UserName)
	})
}

func Test_WebhookUserDelete(t *testing.T) {
	onGiteaRun(t, func(t *testing.T, giteaURL *url.URL) {
		var payloads []api.UserPayload
		var triggeredEvent string

		provider := newMockWebhookProvider(func(r *http.Request) {
			content, _ := io.ReadAll(r.Body)
			var payload api.UserPayload
			err := json.Unmarshal(content, &payload)
			assert.NoError(t, err)
			payloads = append(payloads, payload)
			triggeredEvent = r.Header.Get("X-Gitea-Event")
		}, http.StatusOK)
		defer provider.Close()

		// 1. Create a system webhook for user_delete event
		session := loginUser(t, "user1") // user1 is admin
		hookID := testAPICreateSystemWebhook(t, session, provider.URL(), "user_delete")
		defer testAPIDeleteSystemWebhook(t, session, hookID)

		// 2. First create a user that we'll delete
		testAPICreateUser(t, session, "webhookdeleteuser", "webhookdeleteuser@example.com", "password123!")

		// 3. Trigger the webhook by deleting the user
		testAPIDeleteUser(t, session, "webhookdeleteuser")

		// 4. Validate the webhook is triggered
		assert.Len(t, payloads, 1)
		assert.Equal(t, string(webhook_module.HookEventUserDelete), triggeredEvent)
		assert.Equal(t, api.HookUserDeleted, payloads[0].Action)
		assert.Equal(t, "webhookdeleteuser", payloads[0].User.UserName)
		assert.NotNil(t, payloads[0].Sender)
		assert.Equal(t, "user1", payloads[0].Sender.UserName)
	})
}

func Test_WebhookUserCreateAndDelete(t *testing.T) {
	onGiteaRun(t, func(t *testing.T, giteaURL *url.URL) {
		var createPayloads []api.UserPayload
		var deletePayloads []api.UserPayload

		createProvider := newMockWebhookProvider(func(r *http.Request) {
			content, _ := io.ReadAll(r.Body)
			var payload api.UserPayload
			err := json.Unmarshal(content, &payload)
			assert.NoError(t, err)
			createPayloads = append(createPayloads, payload)
		}, http.StatusOK)
		defer createProvider.Close()

		deleteProvider := newMockWebhookProvider(func(r *http.Request) {
			content, _ := io.ReadAll(r.Body)
			var payload api.UserPayload
			err := json.Unmarshal(content, &payload)
			assert.NoError(t, err)
			deletePayloads = append(deletePayloads, payload)
		}, http.StatusOK)
		defer deleteProvider.Close()

		// 1. Create system webhooks for both events
		session := loginUser(t, "user1")
		createHookID := testAPICreateSystemWebhook(t, session, createProvider.URL(), "user_create")
		defer testAPIDeleteSystemWebhook(t, session, createHookID)

		deleteHookID := testAPICreateSystemWebhook(t, session, deleteProvider.URL(), "user_delete")
		defer testAPIDeleteSystemWebhook(t, session, deleteHookID)

		// 2. Create and then delete a user
		testAPICreateUser(t, session, "webhookbothuser", "webhookbothuser@example.com", "password123!")
		testAPIDeleteUser(t, session, "webhookbothuser")

		// 3. Validate both webhooks are triggered
		assert.Len(t, createPayloads, 1)
		assert.Equal(t, api.HookUserCreated, createPayloads[0].Action)
		assert.Equal(t, "webhookbothuser", createPayloads[0].User.UserName)

		assert.Len(t, deletePayloads, 1)
		assert.Equal(t, api.HookUserDeleted, deletePayloads[0].Action)
		assert.Equal(t, "webhookbothuser", deletePayloads[0].User.UserName)
	})
}

func Test_WebhookUserUpdate(t *testing.T) {
	onGiteaRun(t, func(t *testing.T, giteaURL *url.URL) {
		var payloads []api.UserPayload
		var triggeredEvent string

		provider := newMockWebhookProvider(func(r *http.Request) {
			content, _ := io.ReadAll(r.Body)
			var payload api.UserPayload
			err := json.Unmarshal(content, &payload)
			assert.NoError(t, err)
			payloads = append(payloads, payload)
			triggeredEvent = r.Header.Get("X-Gitea-Event")
		}, http.StatusOK)
		defer provider.Close()

		// 1. Create a system webhook for user_update event
		session := loginUser(t, "user1") // user1 is admin
		hookID := testAPICreateSystemWebhook(t, session, provider.URL(), "user_update")
		defer testAPIDeleteSystemWebhook(t, session, hookID)

		// 2. First create a user that we'll update
		testAPICreateUser(t, session, "webhookupdateuser", "webhookupdateuser@example.com", "password123!")
		defer testAPIDeleteUser(t, session, "webhookupdateuser")

		// 3. Trigger the webhook by updating the user
		testAPIEditUser(t, session, "webhookupdateuser", api.EditUserOption{
			FullName:  Pointer("Updated Full Name"),
			LoginName: "webhookupdateuser",
		})

		// 4. Validate the webhook is triggered
		assert.Len(t, payloads, 1)
		assert.Equal(t, string(webhook_module.HookEventUserUpdate), triggeredEvent)
		assert.Equal(t, api.HookUserUpdated, payloads[0].Action)
		assert.Equal(t, "webhookupdateuser", payloads[0].User.UserName)
		assert.Equal(t, "Updated Full Name", payloads[0].User.FullName)
		assert.NotNil(t, payloads[0].Sender)
		assert.Equal(t, "user1", payloads[0].Sender.UserName)
	})
}

func Test_WebhookUserProhibitLogin(t *testing.T) {
	onGiteaRun(t, func(t *testing.T, giteaURL *url.URL) {
		var payloads []api.UserPayload
		var triggeredEvents []string

		provider := newMockWebhookProvider(func(r *http.Request) {
			content, _ := io.ReadAll(r.Body)
			var payload api.UserPayload
			err := json.Unmarshal(content, &payload)
			assert.NoError(t, err)
			payloads = append(payloads, payload)
			triggeredEvents = append(triggeredEvents, r.Header.Get("X-Gitea-Event"))
		}, http.StatusOK)
		defer provider.Close()

		// 1. Create a system webhook for user_prohibit_login event
		session := loginUser(t, "user1") // user1 is admin
		hookID := testAPICreateSystemWebhook(t, session, provider.URL(), "user_prohibit_login")
		defer testAPIDeleteSystemWebhook(t, session, hookID)

		// 2. First create a user that we'll prohibit/allow login
		testAPICreateUser(t, session, "webhookprohibituser", "webhookprohibituser@example.com", "password123!")
		defer testAPIDeleteUser(t, session, "webhookprohibituser")

		// 3. Trigger the webhook by prohibiting login
		testAPIEditUser(t, session, "webhookprohibituser", api.EditUserOption{
			ProhibitLogin: Pointer(true),
			LoginName:     "webhookprohibituser",
		})

		// 4. Validate the prohibit webhook is triggered
		assert.Len(t, payloads, 1)
		assert.Equal(t, string(webhook_module.HookEventUserProhibitLogin), triggeredEvents[0])
		assert.Equal(t, api.HookUserProhibited, payloads[0].Action)
		assert.Equal(t, "webhookprohibituser", payloads[0].User.UserName)

		// 5. Trigger the webhook by allowing login again
		testAPIEditUser(t, session, "webhookprohibituser", api.EditUserOption{
			ProhibitLogin: Pointer(false),
			LoginName:     "webhookprohibituser",
		})

		// 6. Validate the allow webhook is triggered
		assert.Len(t, payloads, 2)
		assert.Equal(t, string(webhook_module.HookEventUserProhibitLogin), triggeredEvents[1])
		assert.Equal(t, api.HookUserAllowed, payloads[1].Action)
		assert.Equal(t, "webhookprohibituser", payloads[1].User.UserName)
		assert.NotNil(t, payloads[1].Sender)
		assert.Equal(t, "user1", payloads[1].Sender.UserName)
	})
}

// newMockWebhookProvider is defined in repo_webhook_test.go but we need access to it
// The function is already exported from there, so we don't need to redefine it here
