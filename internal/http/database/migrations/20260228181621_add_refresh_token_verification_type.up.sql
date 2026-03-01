DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_enum
        WHERE enumlabel = 'REFRESH_TOKEN'
        AND enumtypid = 'verification_token_type'::regtype
    ) THEN
        ALTER TYPE verification_token_type
        ADD VALUE 'REFRESH_TOKEN';
    END IF;
END$$;