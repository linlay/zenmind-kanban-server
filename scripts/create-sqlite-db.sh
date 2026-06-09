#!/bin/sh
set -eu

usage() {
	cat <<'EOF'
Usage:
  scripts/create-sqlite-db.sh --db ./data/kanban.db [--demo] [--force]

Options:
  --db PATH   SQLite database path to create.
  --demo      Also load demo projects and issues.
  --force     Replace an existing database and its WAL/SHM sidecar files.
  --help      Show this help.
EOF
}

DB_PATH=""
WITH_DEMO=0
FORCE=0

while [ "$#" -gt 0 ]; do
	case "$1" in
		--db)
			if [ "$#" -lt 2 ]; then
				echo "missing value for --db" >&2
				exit 2
			fi
			DB_PATH="$2"
			shift 2
			;;
		--demo)
			WITH_DEMO=1
			shift
			;;
		--force)
			FORCE=1
			shift
			;;
		--help|-h)
			usage
			exit 0
			;;
		*)
			echo "unknown option: $1" >&2
			usage >&2
			exit 2
			;;
	esac
done

if [ -z "$DB_PATH" ]; then
	echo "--db is required" >&2
	usage >&2
	exit 2
fi

if ! command -v sqlite3 >/dev/null 2>&1; then
	echo "sqlite3 is required but was not found in PATH" >&2
	exit 1
fi

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd -P)
ROOT_DIR=$(CDPATH= cd -- "$SCRIPT_DIR/.." && pwd -P)
STORE_DIR="$ROOT_DIR/internal/store"
DEMO_DIR="$ROOT_DIR/cmd/demo"

for required in "$STORE_DIR/schema.sql" "$STORE_DIR/workflow_seed.sql" "$STORE_DIR/seed_defaults.sql"; do
	if [ ! -f "$required" ]; then
		echo "required SQL file not found: $required" >&2
		exit 1
	fi
done

if [ "$WITH_DEMO" -eq 1 ]; then
	for required in "$DEMO_DIR/01_projects.sql" "$DEMO_DIR/02_issues.sql"; do
		if [ ! -f "$required" ]; then
			echo "required demo SQL file not found: $required" >&2
			exit 1
		fi
	done
fi

if [ -e "$DB_PATH" ] || [ -e "$DB_PATH-wal" ] || [ -e "$DB_PATH-shm" ]; then
	if [ "$FORCE" -ne 1 ]; then
		echo "database already exists: $DB_PATH" >&2
		echo "pass --force to replace it" >&2
		exit 1
	fi
	rm -f "$DB_PATH" "$DB_PATH-wal" "$DB_PATH-shm"
fi

DB_DIR=$(dirname -- "$DB_PATH")
mkdir -p "$DB_DIR"

TMP_SQL=$(mktemp)
trap 'rm -f "$TMP_SQL"' EXIT

NOW=$(date -u '+%Y-%m-%dT%H:%M:%SZ')

append_sql() {
	file="$1"
	sed "s|__NOW__|$NOW|g" "$file" >> "$TMP_SQL"
	printf '\n' >> "$TMP_SQL"
}

append_sql "$STORE_DIR/schema.sql"
append_sql "$STORE_DIR/workflow_seed.sql"
append_sql "$STORE_DIR/seed_defaults.sql"

if [ "$WITH_DEMO" -eq 1 ]; then
	append_sql "$DEMO_DIR/01_projects.sql"
	append_sql "$DEMO_DIR/02_issues.sql"
fi

sqlite3 "$DB_PATH" < "$TMP_SQL" >/dev/null

FK_ERRORS=$(sqlite3 "$DB_PATH" "PRAGMA foreign_key_check;")
if [ -n "$FK_ERRORS" ]; then
	echo "foreign key check failed:" >&2
	echo "$FK_ERRORS" >&2
	exit 1
fi

INVALID_STATUS_COUNT=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM workflow_status WHERE COLUMN_KEY_ NOT IN ('backlog','todo','in_progress','in_review','completed');")
if [ "$INVALID_STATUS_COUNT" != "0" ]; then
	echo "workflow_status contains non-standard COLUMN_KEY_ values" >&2
	exit 1
fi

MISSING_ISSUE_STATUS_COUNT=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM issue i LEFT JOIN workflow_status wst ON wst.ID_ = i.STATUS_ID_ WHERE wst.ID_ IS NULL OR wst.KEY_ = '' OR wst.NAME_ = '';")
if [ "$MISSING_ISSUE_STATUS_COUNT" != "0" ]; then
	echo "some issues cannot resolve statusKey/statusName from workflow_status" >&2
	exit 1
fi

WORKFLOW_COUNT=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM workflow WHERE DELETED_AT_ IS NULL;")
PROJECT_COUNT=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM project WHERE DELETED_AT_ IS NULL;")
ISSUE_COUNT=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM issue WHERE DELETED_AT_ IS NULL;")

echo "SQLite database created: $DB_PATH"
echo "workflows=$WORKFLOW_COUNT projects=$PROJECT_COUNT issues=$ISSUE_COUNT demo=$WITH_DEMO"
