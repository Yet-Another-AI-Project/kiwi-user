-- Modify "payments" table
ALTER TABLE "payments" DROP COLUMN "platform", ADD COLUMN "wechat_platform" character varying NULL;
