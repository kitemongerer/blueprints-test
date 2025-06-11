# agent-implementations/ProspectResearchAgent_v0.1.0/Dockerfile
# Dockerfile for running the ADK agent with a custom FastAPI server

# --- Stage 1: Build Stage (Install dependencies with Poetry) ---
FROM python:3.11-slim AS builder

# POETRY_HOME should be in PATH for subsequent RUN commands
ENV PYTHONDONTWRITEBYTECODE=1 \
PYTHONUNBUFFERED=1 \
POETRY_VERSION=1.7.1 \
POETRY_HOME="/opt/poetry" \
POETRY_VIRTUALENVS_IN_PROJECT=true \
PATH="${POETRY_HOME}/bin:${PATH}"

# Install system dependencies (curl for poetry installer, and sh for entrypoint)
RUN apt-get update && apt-get install -y --no-install-recommends \
curl \
&& apt-get clean \
&& rm -rf /var/lib/apt/lists/*

# Install Poetry
RUN curl -sSL https://install.python-poetry.org | python3 - --version ${POETRY_VERSION} --yes

# Set working directory for poetry project
WORKDIR /app_build

# Copy dependency files
COPY pyproject.toml poetry.lock* ./

# Install dependencies using Poetry. This will create a .venv directory.
# Explicitly call poetry using its full path to ensure it's found.
RUN "${POETRY_HOME}/bin/poetry" install --no-dev --no-interaction --no-ansi
RUN echo "Builder stage: Dependencies installed into .venv"

# --- Stage 2: Final Stage (Copy venv and application code) ---
FROM python:3.11-slim AS final

ENV PYTHONDONTWRITEBYTECODE=1 \
PYTHONUNBUFFERED=1 \
VENV_PATH="/app/.venv" \
PATH="/app/.venv/bin:${PATH}"

# Create a non-root user and group
RUN groupadd -r appuser && useradd --no-log-init -r -g appuser appuser

WORKDIR /app

# Copy the virtual environment from the builder stage
COPY --from=builder /app_build/.venv ${VENV_PATH}

# Copy the application code
COPY --chown=appuser:appuser ./prospectresearchagent_agent.py /app/prospectresearchagent_agent.py
# This is the A2A card definition
COPY --chown=appuser:appuser ./agent.json /app/agent.json
# Configuration for the server
COPY --chown=appuser:appuser ./agent.yaml /app/agent.yaml
# Custom FastAPI A2A server
# COPY --chown=appuser:appuser ./server.py /app/server.py
# Copy the entrypoint script (Temporarily comment out as we are testing a barebones ENTRYPOINT)
# COPY --chown=appuser:appuser ./entrypoint.sh /app/entrypoint.sh

# Make entrypoint script executable (Temporarily comment out)
# RUN chmod +x /app/entrypoint.sh

# Ensure correct ownership for all app files
RUN chown -R appuser:appuser /app

# Switch to non-root user
USER appuser

# Expose the port the FastAPI app runs on (will be set by $PORT from Render, defaults to 8000 in entrypoint)
# This EXPOSE is more for documentation; Render uses the PORT env var.
EXPOSE 8000

# Original CMD (commented out)
CMD ["sh", "-c", "echo '>>>> Docker CMD is executing! Hello from inside the container! Starting simple HTTP server...' && ls -la /app && python3 -m http.server ${PORT:-8000}"]
# Original ENTRYPOINT (commented out)
# ENTRYPOINT ["/app/entrypoint.sh"]

# TEMPORARY TEST: Absolute simplest possible ENTRYPOINT to see if *anything* prints
# ONLY THIS LINE SHOULD BE UNCOMMENTED FOR THIS TEST
# ENTRYPOINT ["/bin/bash", "-c", "echo 'RENDER TEST: Basic shell command is running!' && id && ls -la / && sleep 60"]
