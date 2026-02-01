-- Modify "payments" table
ALTER TABLE "payments" ALTER COLUMN "stripe_subscription_status" DROP NOT NULL, ALTER COLUMN "stripe_subscription_status" DROP DEFAULT;
