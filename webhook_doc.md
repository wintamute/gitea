# User System Webhooks

This document describes the system-level webhooks for user lifecycle events in Gitea.

## Overview

System webhooks allow administrators to receive notifications when user accounts are created, updated, deleted, or have their login status changed. These webhooks fire at the system level (not tied to any specific repository).

## Supported Events

| Event Type | Description |
|------------|-------------|
| `user_create` | Triggered when a new user account is created |
| `user_delete` | Triggered when a user account is deleted |
| `user_update` | Triggered when a user's profile is updated |
| `user_prohibit_login` | Triggered when a user's login status is changed |

## Payload Structure

All user events use the `UserPayload` structure:

```json
{
  "action": "created|deleted|updated|prohibited|allowed",
  "user": {
    "id": 123,
    "login": "username",
    "login_name": "",
    "source_id": 0,
    "full_name": "Full Name",
    "email": "user@example.com",
    "avatar_url": "https://gitea.example.com/avatars/abc123",
    "html_url": "https://gitea.example.com/username",
    "language": "en-US",
    "is_admin": false,
    "last_login": "2024-01-15T10:30:00Z",
    "created": "2024-01-01T00:00:00Z",
    "restricted": false,
    "active": true,
    "prohibit_login": false,
    "location": "",
    "pronouns": "",
    "website": "",
    "description": "",
    "visibility": "public",
    "followers_count": 0,
    "following_count": 0,
    "starred_repos_count": 0
  },
  "sender": {
    "id": 1,
    "login": "admin",
    "full_name": "Admin User",
    "email": "admin@example.com",
    "avatar_url": "https://gitea.example.com/avatars/xyz789"
  }
}
```

### Action Values

| Action | Event Type | Description |
|--------|------------|-------------|
| `created` | `user_create` | User account was created |
| `deleted` | `user_delete` | User account was deleted |
| `updated` | `user_update` | User profile was modified |
| `prohibited` | `user_prohibit_login` | User login was prohibited |
| `allowed` | `user_prohibit_login` | User login was allowed again |

### Fields

- **action**: The action that triggered the webhook
- **user**: The user object that was acted upon
- **sender**: The user who performed the action (admin or self for self-registration)

## API Reference

### Create a System Webhook

```http
POST /api/v1/admin/hooks
```

**Request Body:**

```json
{
  "type": "gitea",
  "config": {
    "url": "https://your-webhook-endpoint.com/webhook",
    "content_type": "json",
    "secret": "your-webhook-secret"
  },
  "events": [
    "user_create",
    "user_delete",
    "user_update",
    "user_prohibit_login"
  ],
  "active": true
}
```

**Example with curl:**

```bash
curl -X POST "https://gitea.example.com/api/v1/admin/hooks" \
  -H "Authorization: token YOUR_ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "type": "gitea",
    "config": {
      "url": "https://your-webhook-endpoint.com/webhook",
      "content_type": "json"
    },
    "events": ["user_create", "user_delete"],
    "active": true
  }'
```

**Response (201 Created):**

```json
{
  "id": 1,
  "type": "gitea",
  "config": {
    "content_type": "json",
    "url": "https://your-webhook-endpoint.com/webhook"
  },
  "events": ["user_create", "user_delete"],
  "active": true,
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:30:00Z"
}
```

### List System Webhooks

```http
GET /api/v1/admin/hooks
```

**Query Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `page` | integer | Page number (1-based) |
| `limit` | integer | Page size |
| `type` | string | Filter by type: `system`, `default`, or `all` |

**Example:**

```bash
curl -X GET "https://gitea.example.com/api/v1/admin/hooks" \
  -H "Authorization: token YOUR_ADMIN_TOKEN"
```

### Get a System Webhook

```http
GET /api/v1/admin/hooks/{id}
```

**Example:**

```bash
curl -X GET "https://gitea.example.com/api/v1/admin/hooks/1" \
  -H "Authorization: token YOUR_ADMIN_TOKEN"
```

### Update a System Webhook

```http
PATCH /api/v1/admin/hooks/{id}
```

**Request Body:**

```json
{
  "events": ["user_create", "user_delete", "user_update"],
  "active": true
}
```

**Example:**

```bash
curl -X PATCH "https://gitea.example.com/api/v1/admin/hooks/1" \
  -H "Authorization: token YOUR_ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "events": ["user_create", "user_delete", "user_update", "user_prohibit_login"],
    "active": true
  }'
```

### Delete a System Webhook

```http
DELETE /api/v1/admin/hooks/{id}
```

**Example:**

```bash
curl -X DELETE "https://gitea.example.com/api/v1/admin/hooks/1" \
  -H "Authorization: token YOUR_ADMIN_TOKEN"
```

**Response:** `204 No Content`

## User Management API (Triggers Webhooks)

### Create User (triggers `user_create`)

```http
POST /api/v1/admin/users
```

**Request Body:**

```json
{
  "username": "newuser",
  "email": "newuser@example.com",
  "password": "securePassword123!",
  "must_change_password": false,
  "send_notify": false
}
```

**Example:**

```bash
curl -X POST "https://gitea.example.com/api/v1/admin/users" \
  -H "Authorization: token YOUR_ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "newuser",
    "email": "newuser@example.com",
    "password": "securePassword123!",
    "must_change_password": false
  }'
```

### Update User (triggers `user_update` and/or `user_prohibit_login`)

```http
PATCH /api/v1/admin/users/{username}
```

**Request Body (profile update - triggers `user_update`):**

```json
{
  "full_name": "New Full Name",
  "email": "newemail@example.com",
  "website": "https://example.com",
  "location": "New York",
  "login_name": "newuser"
}
```

**Request Body (prohibit login - triggers `user_prohibit_login`):**

```json
{
  "prohibit_login": true,
  "login_name": "newuser"
}
```

**Example:**

```bash
# Update profile
curl -X PATCH "https://gitea.example.com/api/v1/admin/users/newuser" \
  -H "Authorization: token YOUR_ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "full_name": "Updated Name",
    "login_name": "newuser"
  }'

# Prohibit login
curl -X PATCH "https://gitea.example.com/api/v1/admin/users/newuser" \
  -H "Authorization: token YOUR_ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "prohibit_login": true,
    "login_name": "newuser"
  }'
```

### Delete User (triggers `user_delete`)

```http
DELETE /api/v1/admin/users/{username}
```

**Query Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `purge` | boolean | Completely purge the user from the system |

**Example:**

```bash
curl -X DELETE "https://gitea.example.com/api/v1/admin/users/newuser?purge=true" \
  -H "Authorization: token YOUR_ADMIN_TOKEN"
```

## Webhook Headers

When a webhook is triggered, the following headers are included:

| Header | Description |
|--------|-------------|
| `X-Gitea-Event` | The event type (e.g., `user_create`) |
| `X-Gitea-Delivery` | Unique delivery UUID |
| `X-Gitea-Signature` | HMAC signature if secret is configured |
| `X-Gitea-Signature-256` | HMAC-SHA256 signature if secret is configured |
| `Content-Type` | `application/json` |

## Example Payloads

### User Created

```json
{
  "action": "created",
  "user": {
    "id": 5,
    "login": "newuser",
    "full_name": "",
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

### User Deleted

```json
{
  "action": "deleted",
  "user": {
    "id": 5,
    "login": "deleteduser",
    "full_name": "Deleted User",
    "email": "deleted@example.com"
  },
  "sender": {
    "id": 1,
    "login": "admin",
    "full_name": "Admin User"
  }
}
```

### User Updated

```json
{
  "action": "updated",
  "user": {
    "id": 5,
    "login": "existinguser",
    "full_name": "New Full Name",
    "email": "existinguser@example.com",
    "website": "https://newwebsite.com",
    "location": "Updated Location"
  },
  "sender": {
    "id": 1,
    "login": "admin"
  }
}
```

### User Login Prohibited

```json
{
  "action": "prohibited",
  "user": {
    "id": 5,
    "login": "restricteduser",
    "full_name": "Restricted User",
    "email": "restricted@example.com",
    "prohibit_login": true
  },
  "sender": {
    "id": 1,
    "login": "admin"
  }
}
```

### User Login Allowed

```json
{
  "action": "allowed",
  "user": {
    "id": 5,
    "login": "unrestricteduser",
    "full_name": "Unrestricted User",
    "email": "unrestricted@example.com",
    "prohibit_login": false
  },
  "sender": {
    "id": 1,
    "login": "admin"
  }
}
```

## Trigger Locations

User webhooks are triggered from the following locations:

| Trigger | Event(s) | Description |
|---------|----------|-------------|
| Admin API: Create User | `user_create` | `POST /api/v1/admin/users` |
| Admin Web: Create User | `user_create` | Admin panel user creation form |
| Self-Registration | `user_create` | User registers themselves |
| Admin API: Edit User | `user_update`, `user_prohibit_login` | `PATCH /api/v1/admin/users/{username}` |
| Admin Web: Edit User | `user_update`, `user_prohibit_login` | Admin panel user edit form |
| Admin API: Delete User | `user_delete` | `DELETE /api/v1/admin/users/{username}` |
| Admin Web: Delete User | `user_delete` | Admin panel user deletion |

## Web UI Configuration

System webhooks can also be configured through the Gitea web interface:

1. Navigate to **Site Administration** > **Webhooks**
2. Click **Add Webhook** and select webhook type
3. Configure the webhook URL and settings
4. Under **Trigger On**, select the desired events in the **System Events** section:
   - User Created
   - User Deleted
   - User Updated
   - User Login Prohibited

## Notes

- System webhooks require admin privileges to create and manage
- The `sender` field identifies who performed the action
- For self-registration, `sender` equals `user` (the newly created user)
- User deletion notifications are sent **before** the user is actually deleted to ensure user data is available in the payload
- All webhook deliveries are logged and can be reviewed in the webhook settings
