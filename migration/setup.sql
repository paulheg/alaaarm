CREATE TABLE "VERSION" (
    "id" INTEGER,
	"updated_at" DATETIME,
	"version" integer,
	PRIMARY KEY("id")
);

CREATE TABLE "ALERT" (
	"id"	INTEGER,
	"created_at"	DATETIME,
	"updated_at"	DATETIME,
	"deleted_at"	DATETIME,
	"name"	text,
	"description"	text,
	"owner_id"	integer,
	"token"	text,
	PRIMARY KEY("id"),
	CONSTRAINT "fk_ALERT_owner" FOREIGN KEY("owner_id") REFERENCES "USER"("id")
);

CREATE TABLE "ALERT_RECEIVER" (
	"alert_id"	INTEGER,
	"user_id"	INTEGER,
    "created_at"	DATETIME,
	"updated_at"	DATETIME,
	"deleted_at"	DATETIME,
	PRIMARY KEY("alert_id", "user_id")
);

CREATE TABLE "INVITE" (
	"id"	integer,
	"created_at"	datetime,
	"updated_at"	datetime,
	"deleted_at"	datetime,
	"alert_id"	integer,
	"token"	text,
	"one_time"	numeric,
	"expiration"	datetime,
	CONSTRAINT "fk_INVITE_alert" FOREIGN KEY("alert_id") REFERENCES "ALERT"("id"),
	PRIMARY KEY("id")
);

CREATE TABLE "USER" (
	"id" INTEGER,
	"created_at"	datetime,
	"updated_at"	datetime,
	"deleted_at"	datetime,
	"username" TEXT,
	"telegram_id" INTEGER,
	PRIMARY KEY ("id")
)