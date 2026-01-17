-- Create "mail_vertify_codes" table
CREATE TABLE "mail_vertify_codes" ("id" uuid NOT NULL, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NOT NULL, "deleted_at" timestamptz NULL, "email" character varying NOT NULL, "code" character varying NOT NULL, "expires_at" timestamptz NOT NULL, PRIMARY KEY ("id"));
-- Create index "mailvertifycode_email" to table: "mail_vertify_codes"
CREATE INDEX "mailvertifycode_email" ON "mail_vertify_codes" ("email");
