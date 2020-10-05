
CREATE TABLE IF NOT EXISTS "USER" (
	"id" BIGSERIAL,
	"created_at"	timestamp,
	"updated_at"	timestamp,
	"deleted_at"	timestamp,
	"username" TEXT,
	"telegram_id" INTEGER,
	PRIMARY KEY ("id")
);

CREATE TABLE IF NOT EXISTS "ALERT" (
	"id"	BIGSERIAL,
	"created_at"	timestamp,
	"updated_at"	timestamp,
	"deleted_at"	timestamp,
	"name"	text,
	"description"	text,
	"owner_id"	integer,
	"token"	text,
	PRIMARY KEY("id"),
	CONSTRAINT "fk_ALERT_owner" FOREIGN KEY("owner_id") REFERENCES "USER"("id")
);

CREATE TABLE IF NOT EXISTS "ALERT_RECEIVER" (
	"alert_id"	INTEGER,
	"user_id"	INTEGER,
    "created_at"	timestamp,
	"updated_at"	timestamp,
	"deleted_at"	timestamp,
	PRIMARY KEY("alert_id", "user_id")
);

CREATE TABLE IF NOT EXISTS "INVITE" (
	"id"	BIGSERIAL,
	"created_at"	timestamp,
	"updated_at"	timestamp,
	"deleted_at"	timestamp,
	"alert_id"	integer,
	"token"	text,
	"one_time"	BOOLEAN,
	"expiration"	timestamp,
	CONSTRAINT "fk_INVITE_alert" FOREIGN KEY("alert_id") REFERENCES "ALERT"("id"),
	PRIMARY KEY("id")
);