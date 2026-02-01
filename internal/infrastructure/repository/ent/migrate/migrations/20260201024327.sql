-- Modify "payments" table
ALTER TABLE "payments" ALTER COLUMN "stripe_interval" DROP NOT NULL, ALTER COLUMN "stripe_interval" DROP DEFAULT, ALTER COLUMN "wechat_platform" DROP NOT NULL, ALTER COLUMN "wechat_platform" DROP DEFAULT, ALTER COLUMN "stripe_subscription_status" DROP NOT NULL, ALTER COLUMN "stripe_subscription_status" DROP DEFAULT;
