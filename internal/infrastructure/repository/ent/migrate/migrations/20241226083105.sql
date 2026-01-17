-- Modify "organizations" table
ALTER TABLE "organizations" ADD COLUMN "refresh_at" timestamptz NULL;
