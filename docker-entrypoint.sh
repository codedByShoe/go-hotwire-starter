#!/bin/sh
set -e

# Without a replica URL, run the server directly (no backups).
if [ -z "$LITESTREAM_REPLICA_URL" ]; then
    exec /app/server
fi

# On a fresh volume, restore the latest backup before starting.
litestream restore -config /etc/litestream.yml -if-db-not-exists -if-replica-exists "$DATABASE_PATH"

# Litestream runs the server as a child process and keeps replicating
# until it exits.
exec litestream replicate -config /etc/litestream.yml -exec /app/server
