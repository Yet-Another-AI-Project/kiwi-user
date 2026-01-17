-- Create index "organizationrequest_organization_id" to table: "organization_requests"
CREATE INDEX "organizationrequest_organization_id" ON "organization_requests" ("organization_id");
-- Create index "organizationuser_user_id_organization_id" to table: "organization_users"
CREATE UNIQUE INDEX "organizationuser_user_id_organization_id" ON "organization_users" ("user_id", "organization_id");
