-- Modify "organizations" table
ALTER TABLE "organizations" ADD COLUMN "permission_code" character varying NOT NULL DEFAULT 'init_permission_code';
