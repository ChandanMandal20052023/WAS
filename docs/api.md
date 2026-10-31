# WAS — REST API & WebSocket Protocol Specification

---

## 1. Global Conventions

- **Base Path:** `/api/v1`
- **Request Format:** `application/json`
- **Response Format:** Standard JSON payload:
  ```json
  {
    "success": true,
    "data": { ... },
    "error": null,
    "meta": { "page": 1, "total": 100 }
  }
  ```
- **Error Format:**
  ```json
  {
    "success": false,
    "data": null,
    "error": {
      "code": "RESOURCE_NOT_FOUND",
      "message": "The requested conversation was not found"
    }
  }
  ```
- **Authentication:** Bearer token in the `Authorization` header:
  `Authorization: Bearer <jwt_access_token>`

---

## 2. API Endpoints

### Authentication
- `POST /api/v1/auth/register` — Create a new organization and owner account.
- `POST /api/v1/auth/login` — Authenticate with email and password.
- `POST /api/v1/auth/refresh` — Exchange refresh token for new access token.
- `POST /api/v1/auth/logout` — Invalidate current session and tokens.
- `GET /api/v1/auth/me` — Retrieve current user profile and active memberships.
- `POST /api/v1/auth/forgot-password` — Request password reset email.
- `POST /api/v1/auth/reset-password` — Complete password reset with token.

### Organizations & Team
- `GET /api/v1/organizations/current` — Get active organization details and settings.
- `PUT /api/v1/organizations/current` — Update organization configuration.
- `GET /api/v1/team/members` — List all members, roles, and conversation workloads.
- `PUT /api/v1/team/members/:id/role` — Update a member's role (`Admin`, `Manager`, `Agent`).
- `PUT /api/v1/team/members/:id/status` — Activate or deactivate a member.
- `POST /api/v1/team/invites` — Send email invitation to an agent or manager.
- `POST /api/v1/team/invites/:token/accept` — Accept invitation and join organization.

### WhatsApp Channels
- `GET /api/v1/channels` — List all organization channels and live connection statuses.
- `POST /api/v1/channels` — Create a new channel instance in Evolution Go.
- `GET /api/v1/channels/:id/qr` — Get real-time pairing QR code base64 image.
- `POST /api/v1/channels/:id/connect` — Initiate connection to WhatsApp.
- `POST /api/v1/channels/:id/pair` — Request 8-digit phone pairing code.
- `POST /api/v1/channels/:id/disconnect` — Disconnect WhatsApp session.
- `POST /api/v1/channels/:id/reconnect` — Trigger session reconnection.
- `DELETE /api/v1/channels/:id` — Delete channel and remove instance from engine.
- `GET /api/v1/channels/:id/status` — Check live status from Evolution Go.

### Contacts
- `GET /api/v1/contacts` — Paginated contact search and filtering.
- `POST /api/v1/contacts` — Create new contact (enforces E.164 phone uniqueness).
- `GET /api/v1/contacts/:id` — Contact details, tags, and conversation history.
- `PUT /api/v1/contacts/:id` — Update contact information.
- `DELETE /api/v1/contacts/:id` — Delete contact record.
- `POST /api/v1/contacts/:id/tags` — Attach or detach tags from contact.

### Conversations & Messages
- `GET /api/v1/conversations` — Filter conversations by status (`open`, `pending`, `resolved`), channel, or assigned agent.
- `GET /api/v1/conversations/:id` — Full conversation thread metadata.
- `GET /api/v1/conversations/:id/messages` — Paginated message history.
- `POST /api/v1/conversations/:id/messages/text` — Send outgoing WhatsApp text message.
  - Body: `{"text": "Hello customer"}`
- `POST /api/v1/conversations/:id/messages/media` — Send outgoing WhatsApp media attachment.
  - Body: `{"media_url": "...", "caption": "...", "mime_type": "image/jpeg"}`
- `POST /api/v1/conversations/:id/assign` — Assign conversation to a specific agent.
  - Body: `{"agent_id": "...", "reason": "..."}`
- `PUT /api/v1/conversations/:id/status` — Change conversation status.
  - Body: `{"status": "resolved"}`
- `POST /api/v1/conversations/:id/read` — Mark all unread messages as read.
- `GET /api/v1/conversations/:id/notes` — Fetch internal private team notes.
- `POST /api/v1/conversations/:id/notes` — Add internal note (strictly never sent to WhatsApp).
  - Body: `{"content": "Customer asked for an enterprise quote."}`

### Webhooks (Inbound from Evolution Go)
- `POST /api/v1/webhooks/evolution/:token` — Receives real-time events from Evolution Go:
  - Events: `Message`, `SendMessage`, `Connected`, `Disconnected`, `QRCode`, `Receipt`.

### Dashboard & Analytics
- `GET /api/v1/dashboard/stats` — Real-time metrics (total contacts, active channels, open/pending/resolved conversations, unread messages).
- `GET /api/v1/dashboard/activity` — Timeline of recent customer and team events.

---

## 3. WebSocket Protocol (`/ws`)

### Connection URL
`ws://localhost:8080/ws?token=<jwt_access_token>`

### Message Envelope
```json
{
  "event": "new_message | message_status | conversation_updated | channel_status",
  "payload": { ... },
  "timestamp": "2026-10-31T18:00:00Z"
}
```

### Supported Events
1. `new_message`: Pushed to all agents in the organization when a customer or agent sends a message.
2. `message_status`: Updates message delivery status (`sent` ➔ `delivered` ➔ `read`).
3. `conversation_assigned`: Pushed when a conversation is assigned to an agent.
4. `conversation_status`: Pushed when a conversation is marked open, pending, or resolved.
5. `channel_status`: Pushed when a WhatsApp channel connects, disconnects, or changes state.

