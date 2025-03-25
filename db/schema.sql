CREATE TABLE IF NOT EXISTS "sessions" (
    "id" TEXT PRIMARY KEY NOT NULL,
    "user_id" INTEGER NOT NULL,
    "expires_at" TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE TABLE IF NOT EXISTS "users" (
    "id" SERIAL PRIMARY KEY NOT NULL,
    "google_id" TEXT NOT NULL UNIQUE,
    "email" VARCHAR NOT NULL UNIQUE,
    "name" TEXT NOT NULL,
    "picture" TEXT NOT NULL
);

DO $$
BEGIN
    ALTER TABLE "sessions"
    ADD CONSTRAINT "sessions_user_id_users_id_fk"
    FOREIGN KEY ("user_id")
    REFERENCES "public"."users"("id")
    ON DELETE NO ACTION
    ON UPDATE NO ACTION;
EXCEPTION
    WHEN duplicate_object THEN NULL;
END $$;
