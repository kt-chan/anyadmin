#!/bin/bash

# ==============================================
# Force refresh Git branch and restart containers
# ==============================================

# --- Configuration ---
BRANCH="dev-20260123-ktchan"          # change to your branch
REMOTE="origin"                       # remote name
DOCKER_COMPOSE_DIR="docker"           # directory containing docker-compose.yaml

# --- Check if we are inside a git repo ---
if ! git rev-parse --git-dir > /dev/null 2>&1; then
    echo "Error: Not inside a Git repository. Exiting."
    exit 1
fi

# --- Fetch latest from remote ---
echo "Fetching latest from $REMOTE..."
git fetch $REMOTE

# --- Ensure we are on the target branch ---
current_branch=$(git branch --show-current)
if [ "$current_branch" != "$BRANCH" ]; then
    echo "Switching to branch $BRANCH..."
    git checkout $BRANCH || { echo "Failed to checkout $BRANCH"; exit 1; }
fi

# --- Discard all local changes (tracked files) and reset to remote ---
echo "Discarding local changes and resetting to $REMOTE/$BRANCH..."
git reset --hard $REMOTE/$BRANCH

# --- Remove untracked files and directories (optional) ---
echo "Removing untracked files and directories..."
git clean -fd

# --- Optional: show status ---
echo "Current git status:"
git status --short

# --- Rebuild and restart Docker containers ---
echo "Rebuilding and restarting containers..."
cd "$DOCKER_COMPOSE_DIR" || { echo "Directory $DOCKER_COMPOSE_DIR not found"; exit 1; }
docker compose up --build --force-recreate --remove-orphans -d

# --- Return to original directory ---
cd - > /dev/null

echo "Done! Containers are running with the latest code."