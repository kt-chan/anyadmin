#!/bin/bash

# --- refresh anyadmin system setup ---
rm -rf data.json
touch data.json
touch .env
mkdir -p logs keys tars
chown demo:demo data.json logs keys tars

# --- Rebuild and restart Docker containers ---
cd docker
echo "Rebuilding and restarting containers..."
docker compose up --force-recreate --remove-orphans -d
docker compose logs -f anyadmin
