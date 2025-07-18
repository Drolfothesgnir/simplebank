DROP TABLE IF EXISTS "verification_emails" CASCADE;

ALTER TABLE "users" DROP COLUMN IF EXISTS is_email_verified;

ALTER TABLE IF EXISTS "verification_emails" DROP CONSTRAINT IF EXISTS "verification_emails_username_fkey";