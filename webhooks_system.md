Webhook Architecture

Layered Design

The webhook system follows a clean layered architecture:

Event Source → Notify Service → Webhook Notifier → Prepare → Queue → Deliver

1. Event Type Definitions (modules/webhook/type.go)

All event types are defined as HookEventType string constants:
- Repo events: create, delete, fork, push, repository, release, wiki, status
- Issue events: issues, issue_assign, issue_label, issue_milestone, issue_comment
- PR events: pull_request, pull_request_assign, pull_request_label, pull_request_review, pull_request_sync, etc.
- Workflow events: workflow_run, workflow_job
- System events (added on this branch): user_create, user_delete, user_update, user_prohibit_login

2. Three-Level Webhook Scope
   ┌────────────┬─────────────────────────────────────────────────┬───────────────────────────────────────────┐
   │   Scope    │                 Identification                  │                 Use Case                  │
   ├────────────┼─────────────────────────────────────────────────┼───────────────────────────────────────────┤
   │ Repository │ RepoID != 0                                     │ Events from a specific repo               │
   ├────────────┼─────────────────────────────────────────────────┼───────────────────────────────────────────┤
   │ Owner      │ OwnerID != 0, RepoID = 0                        │ Events from a user/org's resources        │
   ├────────────┼─────────────────────────────────────────────────┼───────────────────────────────────────────┤
   │ System     │ RepoID = 0, OwnerID = 0, IsSystemWebhook = true │ Global events (user lifecycle, all repos) │
   └────────────┴─────────────────────────────────────────────────┴───────────────────────────────────────────┘
3. Notification System (services/notify/)

An interface-based notification dispatcher:
- notifier.go defines the Notifier interface with methods for every event type
- notify.go iterates all registered notifiers and calls them
- The webhook notifier (services/webhook/notifier.go) implements this interface and is registered at init

4. Event Flow (User Creation Example)

POST /api/v1/admin/users  (routers/api/v1/admin/user.go)
→ user_service.CreateUser()           [DB record created]
→ notify_service.CreateUser()         [broadcast to notifiers]
→ webhookNotifier.CreateUser()      [services/webhook/notifier.go]
→ PrepareSystemWebhooks()         [finds system webhooks]
→ PrepareWebhook()              [checks event enabled, creates HookTask]
→ enqueueHookTask()           [adds to async worker queue]
→ Deliver()                 [HTTP POST with HMAC signatures]

For repo events, PrepareWebhooks() aggregates webhooks from all three levels (repo + owner + system).

5. Critical Function: updateHookEvents() (routers/api/v1/utils/hook.go)

This function maps event name strings to the HookEvents map. Every new event type must be added here (as noted in CLAUDE.md). The system events branch added:
hookEvents[webhook_module.HookEventUserCreate] = util.SliceContainsString(events, "user_create", true)
hookEvents[webhook_module.HookEventUserDelete] = util.SliceContainsString(events, "user_delete", true)
hookEvents[webhook_module.HookEventUserUpdate] = util.SliceContainsString(events, "user_update", true)
hookEvents[webhook_module.HookEventUserProhibitLogin] = util.SliceContainsString(events, "user_prohibit_login", true)

6. Delivery (services/webhook/deliver.go)

- Async queue-based delivery via background workers
- Adds triple-compatibility headers: X-Gitea-*, X-GitHub-*, X-Gogs-*
- HMAC signatures (SHA1 + SHA256) using the webhook secret
- Supports multiple platforms: Gitea native, Slack, Discord, Telegram, Matrix, etc.

7. System Events Branch Additions

Commit 43a48031a2 added user lifecycle webhooks:
- New event types: user_create, user_delete, user_update, user_prohibit_login
- Payload: UserPayload struct with Action, User, and Sender fields
- Triggers: API admin endpoints, web admin forms, and self-registration
- Routing: Uses PrepareSystemWebhooks() (system webhooks only, not repo/owner)
- Tests: tests/integration/user_webhook_test.go with 292 lines of coverage
