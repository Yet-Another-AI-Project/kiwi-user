-- Modify "payments" table
ALTER TABLE "payments" ALTER COLUMN "stripe_interval" SET NOT NULL, ALTER COLUMN "stripe_interval" SET DEFAULT 'monthly', ALTER COLUMN "wechat_platform" SET NOT NULL, ALTER COLUMN "wechat_platform" SET DEFAULT 'unknown', ALTER COLUMN "stripe_subscription_status" SET NOT NULL, ALTER COLUMN "stripe_subscription_status" SET DEFAULT 'unpaid';
