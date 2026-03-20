#!/bin/bash

# ==============================================
# Refresh script for anyadmin project
# Place this script in the parent directory of the anyadmin repo.
# It will clone the repo if missing, then update and restart containers.
# ==============================================

# --- Configuration ---
REPO_URL="https://github.com/kt-chan/anyadmin.git"
BRANCH="dev-20260123-ktchan"
REPO_DIR="anyadmin"                              # name of the directory after clone
DOCKER_COMPOSE_PATH="./docker/docker-compose.yaml"
SERVICE_NAME="anyadmin"                          # service to tail logs (optional)

# --- Determine where the script is located ---
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR" || exit 1

echo "Working directory: $(pwd)"

# --- Step 1: Ensure the repository exists locally ---
if [ ! -d "$REPO_DIR" ]; then
    echo "Repository directory '$REPO_DIR' not found. Cloning..."
    git clone --branch "$BRANCH" "$REPO_URL" "$REPO_DIR"
    if [ $? -ne 0 ]; then
        echo "Failed to clone repository. Exiting."
        exit 1
    fi
    echo "Repository cloned."
elif [ ! -d "$REPO_DIR/.git" ]; then
    echo "Directory '$REPO_DIR' exists but is not a Git repository. Please remove or rename it and re-run."
    exit 1
fi

# --- Step 2: Enter the repository ---
cd "$REPO_DIR" || exit 1

# --- Step 3: Fetch latest from remote ---
echo "Fetching latest from remote..."
git fetch origin

# --- Step 4: Ensure we are on the correct branch ---
current_branch=$(git branch --show-current)
if [ "$current_branch" != "$BRANCH" ]; then
    echo "Switching to branch $BRANCH..."
    git checkout "$BRANCH" || { echo "Failed to checkout $BRANCH"; exit 1; }
fi

# --- Step 5: Discard all local changes and reset to remote ---
echo "Discarding local changes and resetting to origin/$BRANCH..."
git reset --hard "origin/$BRANCH"

# --- Step 6: Remove untracked files and directories ---
echo "Removing untracked files and directories..."
git clean -fd

# --- Step 7: Verify working tree is clean ---
echo "Current git status:"
git status --short

# --- Step 8: Rebuild and restart Docker containers ---
echo "Rebuilding and restarting containers..."
docker compose -f "$DOCKER_COMPOSE_PATH" up --build --force-recreate --remove-orphans -d

# --- Step 9: Optionally tail logs ---
echo "Tailing logs for service '$SERVICE_NAME' (Ctrl+C to exit)..."
docker compose -f "$DOCKER_COMPOSE_PATH" logs -f "$SERVICE_NAME"
