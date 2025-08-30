-- +goose Up
-- +goose StatementBegin

-- Seed admin user if none exists
-- This migration creates a default admin user only if no admin users exist
-- The actual values will be inserted by the companion Go program

DO $$
DECLARE
    admin_exists INTEGER;
BEGIN
    -- Check if any admin user (access_level >= 3) already exists
    SELECT COUNT(*) INTO admin_exists 
    FROM users 
    WHERE access_level >= 3;
    
    -- Only create admin user if none exists
    IF admin_exists = 0 THEN
        -- Insert default admin user (will be updated by Go program with env vars)
        INSERT INTO users (
            first_name, 
            last_name, 
            email, 
            password, 
            access_level, 
            created_at, 
            updated_at
        ) VALUES (
            'Admin',
            'User', 
            'admin@milosresidence.com',
            '$2a$12$3XdLBlXMphdX2Y8KDiM3quIxTwOlOG8ahdZhFmFCmER1TXHBaaFu.', -- bcrypt hash for 'admin123'
            3, -- Administrator access level
            CURRENT_TIMESTAMP,
            CURRENT_TIMESTAMP
        );
        
        RAISE NOTICE 'Created default admin user: admin@milosresidence.com';
        RAISE NOTICE 'Default password is: admin123 - CHANGE THIS IMMEDIATELY!';
        RAISE NOTICE 'Run: make seed-admin-env to customize with environment variables';
    ELSE
        RAISE NOTICE 'Admin user already exists, skipping creation';
    END IF;
END $$;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Remove any admin users that were created by this migration
-- We use a conservative approach to only remove the specific default admin
DELETE FROM users 
WHERE email = 'admin@milosresidence.com'
  AND access_level = 3
  AND first_name = 'Admin'
  AND last_name = 'User';

-- +goose StatementEnd