# System Webhook Events

This document describes all system-level webhook events added to Gitea for administrative lifecycle management. These webhooks fire at the system level (not tied to any specific repository) and are managed through the Site Administration panel or the admin API.

## Table of Contents

- [Overview](#overview)
- [User Events](#user-events)
- [Organization Events](#organization-events)
- [Team Events](#team-events)
- [Team Member Events](#team-member-events)
- [API Reference](#api-reference)
- [Web UI Configuration](#web-ui-configuration)
- [Webhook Headers](#webhook-headers)
- [Architecture](#architecture)

---

## Overview

System webhooks allow Gitea administrators to receive HTTP notifications when administrative actions occur. Unlike repository or organization webhooks, system webhooks are global and fire regardless of which repository or organization is involved.

All system webhooks are identified by: `RepoID = 0`, `OwnerID = 0`, `IsSystemWebhook = true`.

### Supported Event Categories

| Category | Events | Description |
|----------|--------|-------------|
| **User** | `user_create`, `user_delete`, `user_update`, `user_prohibit_login` | User account lifecycle |
| **Organization** | `org_create`, `org_delete` | Organization lifecycle |
| **Team** | `team_create`, `team_delete` | Team lifecycle |
| **Team Member** | `team_add_member`, `team_remove_member` | Team membership changes |

---

## User Events

### Payload Structure (`UserPayload`)

```json
{
  "action": "created|deleted|updated|prohibited|allowed",
  "user": { /* User object */ },
  "sender": { /* User object - who performed the action */ }
}
```

### Events

#### `user_create`

Triggered when a new user account is created.

| Action Value | Trigger Sources |
|---|---|
| `created` | Admin API (`POST /api/v1/admin/users`), Admin Web UI, Self-registration |

**Note:** For self-registration, `sender` equals `user` (the newly created user).

#### `user_delete`

Triggered when a user account is deleted.

| Action Value | Trigger Sources |
|---|---|
| `deleted` | Admin API (`DELETE /api/v1/admin/users/{username}`), Admin Web UI |

**Note:** The notification is sent **before** the user is actually deleted, so the full user data is available in the payload.

#### `user_update`

Triggered when a user's profile is updated by an admin.

| Action Value | Trigger Sources |
|---|---|
| `updated` | Admin API (`PATCH /api/v1/admin/users/{username}`), Admin Web UI |

#### `user_prohibit_login`

Triggered when a user's login permission is changed.

| Action Value | Description |
|---|---|
| `prohibited` | User login was disabled |
| `allowed` | User login was re-enabled |

**Trigger Sources:** Admin API (`PATCH /api/v1/admin/users/{username}`), Admin Web UI

### Example Payload

```json
{
  "action": "created",
  "user": {
    "id": 5,
    "login": "newuser",
    "full_name": "New User",
    "email": "newuser@example.com",
    "avatar_url": "https://gitea.example.com/avatars/abc123",
    "html_url": "https://gitea.example.com/newuser",
    "is_admin": false,
    "active": true,
    "prohibit_login": false,
    "created": "2024-01-15T10:30:00Z"
  },
  "sender": {
    "id": 1,
    "login": "admin",
    "full_name": "Admin User",
    "email": "admin@example.com"
  }
}
```

---

## Organization Events

### Payload Structure (`OrganizationPayload`)

```json
{
  "action": "created|deleted",
  "organization": { /* Organization object */ },
  "sender": { /* User object - who performed the action */ }
}
```

### Events

#### `org_create`

Triggered when a new organization is created.

| Action Value | Trigger Sources |
|---|---|
| `created` | API (`POST /api/v1/orgs`), Admin API (`POST /api/v1/admin/orgs`), Web UI |

#### `org_delete`

Triggered when an organization is deleted.

| Action Value | Trigger Sources |
|---|---|
| `deleted` | API (`DELETE /api/v1/orgs/{org}`), Web UI (Organization Settings) |

### Example Payload

```json
{
  "action": "created",
  "organization": {
    "id": 10,
    "name": "my-org",
    "full_name": "My Organization",
    "avatar_url": "https://gitea.example.com/avatars/org123",
    "description": "An example organization",
    "website": "",
    "location": "",
    "visibility": "public"
  },
  "sender": {
    "id": 1,
    "login": "admin",
    "full_name": "Admin User"
  }
}
```

---

## Team Events

### Payload Structure (`TeamPayload`)

```json
{
  "action": "created|deleted",
  "team": { /* Team object (includes organization) */ },
  "organization": { /* Organization object */ },
  "sender": { /* User object - who performed the action */ }
}
```

### Events

#### `team_create`

Triggered when a new team is created in an organization.

| Action Value | Trigger Sources |
|---|---|
| `created` | API (`POST /api/v1/orgs/{org}/teams`), Web UI (Organization Teams page) |

#### `team_delete`

Triggered when a team is deleted from an organization.

| Action Value | Trigger Sources |
|---|---|
| `deleted` | API (`DELETE /api/v1/teams/{id}`), Web UI (Organization Teams page) |

### Example Payload

```json
{
  "action": "created",
  "team": {
    "id": 5,
    "name": "developers",
    "description": "Development team",
    "organization": {
      "id": 10,
      "name": "my-org"
    },
    "includes_all_repositories": false,
    "permission": "write",
    "units": ["repo.code", "repo.issues", "repo.pulls"],
    "units_map": {
      "repo.code": "write",
      "repo.issues": "write",
      "repo.pulls": "write"
    }
  },
  "organization": {
    "id": 10,
    "name": "my-org",
    "full_name": "My Organization"
  },
  "sender": {
    "id": 1,
    "login": "admin"
  }
}
```

---

## Team Member Events

### Payload Structure (`TeamPayload` with `member` field)

```json
{
  "action": "member_added|member_removed",
  "team": { /* Team object (includes organization) */ },
  "organization": { /* Organization object */ },
  "member": { /* User object - the member added/removed */ },
  "sender": { /* User object - who performed the action */ }
}
```

### Events

#### `team_add_member`

Triggered when a user is added to a team.

| Action Value | Trigger Sources |
|---|---|
| `member_added` | API (`PUT /api/v1/teams/{id}/members/{username}`), Web UI (add member form, join team, accept team invite) |

#### `team_remove_member`

Triggered when a user is removed from a team.

| Action Value | Trigger Sources |
|---|---|
| `member_removed` | API (`DELETE /api/v1/teams/{id}/members/{username}`), Web UI (remove member, leave team) |

### Example Payload (Member Added)

```json
{
  "action": "member_added",
  "team": {
    "id": 5,
    "name": "developers",
    "description": "Development team",
    "organization": {
      "id": 10,
      "name": "my-org"
    },
    "permission": "write"
  },
  "organization": {
    "id": 10,
    "name": "my-org",
    "full_name": "My Organization"
  },
  "member": {
    "id": 3,
    "login": "newmember",
    "full_name": "New Member",
    "email": "newmember@example.com"
  },
  "sender": {
    "id": 1,
    "login": "admin"
  }
}
```

### Example Payload (Member Removed)

```json
{
  "action": "member_removed",
  "team": {
    "id": 5,
    "name": "developers",
    "organization": {
      "id": 10,
      "name": "my-org"
    }
  },
  "organization": {
    "id": 10,
    "name": "my-org"
  },
  "member": {
    "id": 3,
    "login": "removedmember",
    "full_name": "Removed Member"
  },
  "sender": {
    "id": 1,
    "login": "admin"
  }
}
```

---

## API Reference

### Create a System Webhook

```http
POST /api/v1/admin/hooks
```

```json
{
  "type": "gitea",
  "config": {
    "url": "https://your-endpoint.com/webhook",
    "content_type": "json",
    "secret": "your-webhook-secret"
  },
  "events": [
    "user_create", "user_delete", "user_update", "user_prohibit_login",
    "org_create", "org_delete",
    "team_create", "team_delete", "team_add_member", "team_remove_member"
  ],
  "active": true
}
```

```bash
curl -X POST "https://gitea.example.com/api/v1/admin/hooks" \
  -H "Authorization: token YOUR_ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "type": "gitea",
    "config": {
      "url": "https://your-endpoint.com/webhook",
      "content_type": "json"
    },
    "events": ["user_create", "org_create", "team_create", "team_add_member"],
    "active": true
  }'
```

### List System Webhooks

```http
GET /api/v1/admin/hooks
```

### Update a System Webhook

```http
PATCH /api/v1/admin/hooks/{id}
```

### Delete a System Webhook

```http
DELETE /api/v1/admin/hooks/{id}
```

---

## Web UI Configuration

System webhooks can be configured through the Gitea admin panel:

1. Navigate to **Site Administration** > **Webhooks**
2. Click **Add Webhook** and select the webhook type
3. Configure the URL and settings
4. Under **Trigger On**, select events in the **System Events** section:

| Checkbox | Event |
|----------|-------|
| User Created | `user_create` |
| User Deleted | `user_delete` |
| User Updated | `user_update` |
| User Login Prohibited | `user_prohibit_login` |
| Organization Created | `org_create` |
| Organization Deleted | `org_delete` |
| Team Created | `team_create` |
| Team Deleted | `team_delete` |
| Team Member Added | `team_add_member` |
| Team Member Removed | `team_remove_member` |

---

## Webhook Headers

All webhook deliveries include the following HTTP headers:

| Header | Description |
|--------|-------------|
| `X-Gitea-Event` | The event type (e.g., `user_create`, `team_add_member`) |
| `X-Gitea-Delivery` | Unique delivery UUID |
| `X-Gitea-Signature` | HMAC-SHA1 signature (if secret is configured) |
| `X-Gitea-Signature-256` | HMAC-SHA256 signature (if secret is configured) |
| `Content-Type` | `application/json` |

---

## Architecture

### Event Flow

```
Action (API/Web UI)
  -> notify_service.EventName()         # Broadcast to all notifiers
    -> webhookNotifier.EventName()      # Webhook notifier implementation
      -> PrepareSystemWebhooks()        # Find matching system webhooks
        -> PrepareWebhook()             # Check event is enabled, create HookTask
          -> enqueueHookTask()          # Add to async worker queue
            -> Deliver()                # HTTP POST with signatures
```

### Key Files

| File | Purpose |
|------|---------|
| `modules/webhook/type.go` | Event type constants (`HookEventType`) |
| `modules/structs/hook.go` | Payload structs (`UserPayload`, `OrganizationPayload`, `TeamPayload`) |
| `services/notify/notifier.go` | Notifier interface definition |
| `services/notify/notify.go` | Dispatcher functions |
| `services/webhook/notifier.go` | Webhook notifier implementation |
| `services/webhook/payloader.go` | Payload convertor interface and routing |
| `services/webhook/general.go` | Text formatting helpers for platform convertors |
| `routers/api/v1/utils/hook.go` | `updateHookEvents()` - event registration |
| `services/forms/repo_form.go` | Web form field definitions |
| `routers/web/repo/setting/webhook.go` | Web form to event mapping |
| `templates/repo/settings/webhook/settings.tmpl` | Admin UI checkboxes |

### Trigger Points

| Event | API Trigger | Web UI Trigger |
|-------|-------------|----------------|
| `user_create` | `POST /api/v1/admin/users` | Admin panel, Self-registration |
| `user_delete` | `DELETE /api/v1/admin/users/{username}` | Admin panel |
| `user_update` | `PATCH /api/v1/admin/users/{username}` | Admin panel |
| `user_prohibit_login` | `PATCH /api/v1/admin/users/{username}` | Admin panel |
| `org_create` | `POST /api/v1/orgs`, `POST /api/v1/admin/orgs` | Create org form |
| `org_delete` | `DELETE /api/v1/orgs/{org}` | Org settings |
| `team_create` | `POST /api/v1/orgs/{org}/teams` | Org teams page |
| `team_delete` | `DELETE /api/v1/teams/{id}` | Org teams page |
| `team_add_member` | `PUT /api/v1/teams/{id}/members/{username}` | Add member form, Join team, Accept invite |
| `team_remove_member` | `DELETE /api/v1/teams/{id}/members/{username}` | Remove member, Leave team |

### Supported Webhook Platforms

All system events are supported across all webhook delivery platforms:

- Gitea (native JSON)
- Gogs
- Slack
- Discord
- DingTalk
- Telegram
- Microsoft Teams
- Feishu (Lark)
- Matrix
- WeCom (WeChat Work)
- Packagist
