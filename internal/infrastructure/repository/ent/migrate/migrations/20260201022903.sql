-- Drop index "payment_customer_id" from table: "payments"
DROP INDEX "payment_customer_id";
-- Drop index "payment_subscription_id" from table: "payments"
DROP INDEX "payment_subscription_id";
-- Rename a column from "transaction_id" to "wechat_open_id"
ALTER TABLE "payments" RENAME COLUMN "transaction_id" TO "wechat_open_id";
-- Rename a column from "open_id" to "wechat_transaction_id"
ALTER TABLE "payments" RENAME COLUMN "open_id" TO "wechat_transaction_id";
-- Rename a column from "subscription_id" to "stripe_subscription_id"
ALTER TABLE "payments" RENAME COLUMN "subscription_id" TO "stripe_subscription_id";
-- Rename a column from "subscription_status" to "stripe_interval"
ALTER TABLE "payments" RENAME COLUMN "subscription_status" TO "stripe_interval";
-- Rename a column from "interval" to "stripe_customer_id"
ALTER TABLE "payments" RENAME COLUMN "interval" TO "stripe_customer_id";
-- Rename a column from "current_period_start" to "stripe_current_period_start"
ALTER TABLE "payments" RENAME COLUMN "current_period_start" TO "stripe_current_period_start";
-- Rename a column from "current_period_end" to "stripe_current_period_end"
ALTER TABLE "payments" RENAME COLUMN "current_period_end" TO "stripe_current_period_end";
-- Rename a column from "customer_id" to "stripe_customer_email"
ALTER TABLE "payments" RENAME COLUMN "customer_id" TO "stripe_customer_email";
-- Rename a column from "customer_email" to "stripe_checkout_session_id"
ALTER TABLE "payments" RENAME COLUMN "customer_email" TO "stripe_checkout_session_id";
-- Modify "payments" table
ALTER TABLE "payments" DROP COLUMN "checkout_session_id", ADD COLUMN "stripe_subscription_status" character varying NOT NULL DEFAULT 'unpaid';
-- Create index "payment_stripe_checkout_session_id" to table: "payments"
CREATE INDEX "payment_stripe_checkout_session_id" ON "payments" ("stripe_checkout_session_id");
-- Create index "payment_stripe_customer_id" to table: "payments"
CREATE INDEX "payment_stripe_customer_id" ON "payments" ("stripe_customer_id");
-- Create index "payment_stripe_subscription_id" to table: "payments"
CREATE INDEX "payment_stripe_subscription_id" ON "payments" ("stripe_subscription_id");
