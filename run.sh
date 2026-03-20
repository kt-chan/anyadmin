#!/bin/bash

# --- Rebuild and restart Docker containers ---
echo "Rebuilding and restarting containers..."
docker compose -f "docker/docker-compose.yaml" up --force-recreate --remove-orphans -d
docker compose logs -f anyadmin
