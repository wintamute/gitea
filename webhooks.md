Webhook Architecture Overview

Key Directories
┌─────────────────────────────┬────────────────────────────────────────────────┐
│            Path             │                    Purpose                     │
├─────────────────────────────┼────────────────────────────────────────────────┤
│ modules/webhook/            │ Type definitions & event configuration         │
├─────────────────────────────┼────────────────────────────────────────────────┤
│ models/webhook/             │ Database models (Webhook, HookTask)            │
├─────────────────────────────┼────────────────────────────────────────────────┤
│ services/webhook/           │ Event triggering, payload formatting, delivery │
├─────────────────────────────┼────────────────────────────────────────────────┤
│ routers/api/v1/repo/hook.go │ API endpoints                                  │
├─────────────────────────────┼────────────────────────────────────────────────┤
│ modules/structs/hook.go     │ Payload struct definitions                     │
└─────────────────────────────┴────────────────────────────────────────────────┘
Event Flow

Event occurs (push, issue, PR, etc.)
↓
notify_service dispatches to all notifiers
↓
webhookNotifier (services/webhook/notifier.go)
↓
PrepareWebhooks() - finds matching webhooks, creates HookTask records
↓
enqueueHookTask() - pushes to delivery queue
↓
handler() processes queue - calls Deliver()
↓
HTTP request sent to webhook URL

Core Components

1. Event Types (modules/webhook/type.go): 28+ events including push, issues, pull_request, release, wiki, package, etc.
2. Webhook Types: GITEA, GOGS, SLACK, DISCORD, DINGTALK, TELEGRAM, MSTEAMS, FEISHU, MATRIX, WECHATWORK, PACKAGIST
3. Notifier Pattern (services/webhook/notifier.go): Implements notify.Notifier interface, registered at init. Each event method calls PrepareWebhooks() with the appropriate payload.
4. Delivery System (services/webhook/deliver.go): Uses a WorkerPoolQueue for async delivery with configurable HTTP client, timeouts, and TLS settings.
5. Security: HMAC-SHA256 signatures (X-Gitea-Signature), encrypted secrets, host allowlists, and atomic delivery to prevent duplicates.

Webhook Scope Levels

- Repository: RepoID > 0
- User/Organization: OwnerID > 0
- System-wide: IsSystemWebhook = true
