-- v12 (compatible with v8+): Add a 'namespace' column to allow clustering devices by logical namespaces.
ALTER TABLE whatsmeow_device ADD COLUMN namespace TEXT DEFAULT '';

-- v13 (compatible with v8+): Adds an 'external_id' column to link devices with external systems using a stable external identifier.
ALTER TABLE whatsmeow_device ADD COLUMN external_id TEXT DEFAULT '';