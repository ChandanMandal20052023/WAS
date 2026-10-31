-- 000004_create_notes_tags_assignments.up.sql

-- Tags
CREATE TABLE IF NOT EXISTS tags (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    color_hex VARCHAR(7) NOT NULL DEFAULT '#128C7E',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_tag_org_name UNIQUE (organization_id, name)
);

CREATE INDEX IF NOT EXISTS idx_tags_org ON tags(organization_id);

-- Contact Tags
CREATE TABLE IF NOT EXISTS contact_tags (
    contact_id UUID NOT NULL REFERENCES contacts(id) ON DELETE CASCADE,
    tag_id UUID NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (contact_id, tag_id)
);

CREATE INDEX IF NOT EXISTS idx_contact_tags_tag ON contact_tags(tag_id);

-- Conversation Tags
CREATE TABLE IF NOT EXISTS conversation_tags (
    conversation_id UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    tag_id UUID NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (conversation_id, tag_id)
);

CREATE INDEX IF NOT EXISTS idx_conv_tags_tag ON conversation_tags(tag_id);

-- Conversation Assignments (Audit trail of agent handoffs)
CREATE TABLE IF NOT EXISTS conversation_assignments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    conversation_id UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    assigned_to_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    assigned_by_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    reason TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_conv_assign_conv ON conversation_assignments(conversation_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_conv_assign_agent ON conversation_assignments(organization_id, assigned_to_user_id);

-- Internal Notes (Private to team, strictly never sent to WhatsApp)
CREATE TABLE IF NOT EXISTS internal_notes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    conversation_id UUID REFERENCES conversations(id) ON DELETE CASCADE,
    contact_id UUID REFERENCES contacts(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_internal_notes_conv ON internal_notes(conversation_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_internal_notes_contact ON internal_notes(contact_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_internal_notes_user ON internal_notes(user_id);
