-- 000002_create_channels_and_contacts.up.sql

-- WhatsApp Channels
CREATE TABLE IF NOT EXISTS whatsapp_channels (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    phone_number VARCHAR(50),
    instance_name VARCHAR(255) NOT NULL UNIQUE,
    instance_token VARCHAR(255) NOT NULL UNIQUE,
    status VARCHAR(50) NOT NULL DEFAULT 'Disconnected',
    qrcode_base64 TEXT,
    settings JSONB NOT NULL DEFAULT '{}'::jsonb,
    last_connected_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_channels_org_status ON whatsapp_channels(organization_id, status);
CREATE INDEX IF NOT EXISTS idx_channels_instance_token ON whatsapp_channels(instance_token);
CREATE INDEX IF NOT EXISTS idx_channels_instance_name ON whatsapp_channels(instance_name);

-- Contacts (Deduplicated by normalized E.164 phone number per organization)
CREATE TABLE IF NOT EXISTS contacts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    phone_number VARCHAR(50) NOT NULL,
    full_name VARCHAR(255),
    email VARCHAR(255),
    avatar_url TEXT,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_contact_org_phone UNIQUE (organization_id, phone_number)
);

CREATE INDEX IF NOT EXISTS idx_contacts_org_phone ON contacts(organization_id, phone_number);
CREATE INDEX IF NOT EXISTS idx_contacts_org_name ON contacts(organization_id, full_name);
CREATE INDEX IF NOT EXISTS idx_contacts_org_created ON contacts(organization_id, created_at DESC);

