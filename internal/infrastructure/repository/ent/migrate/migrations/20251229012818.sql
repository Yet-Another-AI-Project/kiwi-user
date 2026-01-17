-- Create "qy_wechat_user_ids" table
CREATE TABLE "qy_wechat_user_ids" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "deleted_at" timestamptz NULL,
  "qy_wechat_user_id" character varying NULL,
  "open_id" character varying NULL,
  "user_id" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "qy_wechat_user_ids_users_qy_wechat_user_ids" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "qywechatuserid_qy_wechat_user_id" to table: "qy_wechat_user_ids"
CREATE UNIQUE INDEX "qywechatuserid_qy_wechat_user_id" ON "qy_wechat_user_ids" ("qy_wechat_user_id");
-- Create index "qywechatuserid_user_id" to table: "qy_wechat_user_ids"
CREATE UNIQUE INDEX "qywechatuserid_user_id" ON "qy_wechat_user_ids" ("user_id");
