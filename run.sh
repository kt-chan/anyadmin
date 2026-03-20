#!/bin/bash

# --- refresh anyadmin system setup ---
rm -rf data.json
touch data.json
chown demo:demo data.json

# --- Rebuild and restart Docker containers ---
cd docker
echo "Rebuilding and restarting containers..."
docker compose up --force-recreate --remove-orphans -d
docker compose logs -f anyadmin
