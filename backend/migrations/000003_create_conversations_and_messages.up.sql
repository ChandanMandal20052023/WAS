-- 000003_create_conversations_and_messages.up.sql

-- Conversations
CREATE TABLE IF NOT EXISTS conversations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    channel_id UUID NOT NULL REFERENCES whatsapp_channels(id) ON DELETE CASCADE,
    contact_id UUID NOT NULL REFERENCES contacts(id) ON DELETE CASCADE,
    assigned_agent_id UUID REFERENCES users(id) ON DELETE SET NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'open' CHECK (status IN ('open', 'pending', 'resolved')),
    last_message_preview TEXT,
    last_message_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    unread_count INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_conversation_org_channel_contact UNIQUE (organization_id, channel_id, contact_id)
);

CREATE INDEX IF NOT EXISTS idx_conversations_org_status ON conversations(organization_id, status);
CREATE INDEX IF NOT EXISTS idx_conversations_org_assigned ON conversations(organization_id, assigned_agent_id);
CREATE INDEX IF NOT EXISTS idx_conversations_org_last_msg ON conversations(organization_id, last_message_time DESC);
CREATE INDEX IF NOT EXISTS idx_conversations_channel ON conversations(channel_id);
CREATE INDEX IF NOT EXISTS idx_conversations_contact ON conversations(contact_id);

-- Messages
CREATE TABLE IF NOT EXISTS messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    conversation_id UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    whatsapp_message_id VARCHAR(255) UNIQUE,
    sender_type VARCHAR(50) NOT NULL CHECK (sender_type IN ('customer', 'agent', 'system')),
    sender_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    message_type VARCHAR(50) NOT NULL DEFAULT 'text' CHECK (message_type IN ('text', 'image', 'audio', 'video', 'document', 'location', 'sticker', 'contact')),
    body TEXT,
    media_url TEXT,
    media_mime_type VARCHAR(100),
    delivery_status VARCHAR(50) NOT NULL DEFAULT 'pending' CHECK (delivery_status IN ('pending', 'sent', 'delivered', 'read', 'failed')),
    raw_payload JSONB,
    sent_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_messages_conv_sent ON messages(conversation_id, sent_at ASC);
CREATE INDEX IF NOT EXISTS idx_messages_org_created ON messages(organization_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_messages_whatsapp_id ON messages(whatsapp_message_id);
CREATE INDEX IF NOT EXISTS idx_messages_delivery_status ON messages(delivery_status);
