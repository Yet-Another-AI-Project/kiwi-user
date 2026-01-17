-- Drop index "mailvertifycode_email" from table: "mail_vertify_codes"
DROP INDEX "mailvertifycode_email";
-- Modify "mail_vertify_codes" table
ALTER TABLE "mail_vertify_codes" ADD COLUMN "type" character varying NOT NULL;
-- Create index "mailvertifycode_type_email" to table: "mail_vertify_codes"
CREATE UNIQUE INDEX "mailvertifycode_type_email" ON "mail_vertify_codes" ("type", "email");
