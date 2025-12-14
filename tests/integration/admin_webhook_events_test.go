// Copyright 2025 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package integration

import (
	"net/http"
	"net/http/httptest"
	"testing"

	auth_model "code.gitea.io/gitea/models/auth"
	"code.gitea.io/gitea/models/unittest"
	"code.gitea.io/gitea/models/webhook"
	api "code.gitea.io/gitea/modules/structs"
	webhook_module "code.gitea.io/gitea/modules/webhook"
	"code.gitea.io/gitea/tests"
)

func TestAdminWebhookUserEvents(t *testing.T) {
	defer tests.PrepareTestEnv(t)()

	// Spin up a dummy receiver
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	// Create a system webhook that listens to all events
	w := &webhook.Webhook{
		RepoID:          0,
		OwnerID:         0,
		IsSystemWebhook: true,
		URL:             srv.URL,
		HTTPMethod:      http.MethodPost,
		ContentType:     webhook.ContentTypeJSON,
		IsActive:        true,
		Type:            webhook_module.GITEA,
		HookEvent: &webhook_module.HookEvent{
			SendEverything: true,
		},
	}
	if err := w.UpdateEvent(); err != nil {
		t.Fatalf("UpdateEvent: %v", err)
	}
	var err error
	w.ID, err = webhook.CreateWebhook(t.Context(), w)
	if err != nil {
		t.Fatalf("CreateWebhook: %v", err)
	}

	// Login as admin and get token
	session := loginUser(t, "user1")
	token := getTokenForLoggedInUser(t, session, auth_model.AccessTokenScopeAll)

	// 1) Create a user via admin API -> expect admin_user_create
	username := "admwh_user"
	req := NewRequestWithJSON(t, "POST", "/api/v1/admin/users", api.CreateUserOption{
		Username: username,
		Email:    "admwh_user@example.com",
		Password: "StrongPassw0rd!",
	}).AddTokenAuth(token)
	MakeRequest(t, req, http.StatusCreated)

	// Check hook task created for create
	created := assertHasAdminHookTask(t, w, webhook_module.HookEventAdminUserCreate, "created")
	if !created {
		t.Fatalf("expected admin_user_create webhook task with action=created")
	}

	// 2) Update user full name -> expect admin_user_update
	req = NewRequestWithJSON(t, "PATCH", "/api/v1/admin/users/"+username, api.EditUserOption{
		FullName: unittest.ToPtr("New Name"),
	}).AddTokenAuth(token)
	MakeRequest(t, req, http.StatusOK)
	updated := assertHasAdminHookTask(t, w, webhook_module.HookEventAdminUserUpdate, "updated")
	if !updated {
		t.Fatalf("expected admin_user_update webhook task with action=updated")
	}

	// 3) Suspend user -> expect admin_user_suspend
	req = NewRequestWithJSON(t, "PATCH", "/api/v1/admin/users/"+username, api.EditUserOption{
		ProhibitLogin: unittest.ToPtr(true),
	}).AddTokenAuth(token)
	MakeRequest(t, req, http.StatusOK)
	suspended := assertHasAdminHookTask(t, w, webhook_module.HookEventAdminUserSuspend, "suspended")
	if !suspended {
		t.Fatalf("expected admin_user_suspend webhook task with action=suspended")
	}

	// 4) Delete user -> expect admin_user_delete
	req = NewRequest(t, "DELETE", "/api/v1/admin/users/"+username).AddTokenAuth(token)
	MakeRequest(t, req, http.StatusNoContent)
	deleted := assertHasAdminHookTask(t, w, webhook_module.HookEventAdminUserDelete, "deleted")
	if !deleted {
		t.Fatalf("expected admin_user_delete webhook task with action=deleted")
	}
}

// assertHasAdminHookTask checks if a hook task exists for the given webhook with the matching event type
// and if its payload action field equals expAction.
func assertHasAdminHookTask(t *testing.T, w *webhook.Webhook, evt webhook_module.HookEventType, expAction string) bool {
	t.Helper()
	tasks, err := webhook.HookTasks(t.Context(), w.ID, 1)
	if err != nil {
		t.Fatalf("HookTasks: %v", err)
	}
	for _, task := range tasks {
		if task.EventType != evt {
			continue
		}
		// Parse payload
		var p api.AdminUserPayload
		if err := jsonUnmarshal([]byte(task.PayloadContent), &p); err == nil {
			if p.Action == expAction {
				return true
			}
		}
	}
	return false
}

// jsonUnmarshal is a tiny wrapper to use the same JSON module as the app
func jsonUnmarshal(b []byte, v any) error { //nolint:unparam
	return json.Unmarshal(b, v)
}
