-- Modify "payments" table
ALTER TABLE "payments" ADD COLUMN "stripe_invoice_id" character varying NULL;
-- Modify "stripe_events" table
ALTER TABLE "stripe_events" ADD COLUMN "subscription_id" character varying NULL;
