-- Drop index "wechatopenid_open_id_platform" from table: "wechat_open_ids"
DROP INDEX "wechatopenid_open_id_platform";
-- Create index "wechatopenid_user_id_platform" to table: "wechat_open_ids"
CREATE UNIQUE INDEX "wechatopenid_user_id_platform" ON "wechat_open_ids" ("user_id", "platform");
