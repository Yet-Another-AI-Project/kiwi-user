-- Rename a column from "company_intro" to "brand_short_name"
ALTER TABLE "organization_applications" RENAME COLUMN "company_intro" TO "brand_short_name";
-- Rename a column from "industry" to "primary_business"
ALTER TABLE "organization_applications" RENAME COLUMN "industry" TO "primary_business";
-- Modify "organization_applications" table
ALTER TABLE "organization_applications" ADD COLUMN "usage_scenario" character varying NOT NULL, ADD COLUMN "referrer_name" character varying NULL, ADD COLUMN "discovery_way" character varying NULL;
