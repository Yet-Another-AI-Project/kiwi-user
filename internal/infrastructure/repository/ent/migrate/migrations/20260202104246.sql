-- Modify "payments" table
ALTER TABLE "payments" ADD COLUMN "stripe_cancel_at_period_end" boolean NOT NULL DEFAULT false;
