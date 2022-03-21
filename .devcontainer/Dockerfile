# [Choice] Go version (use -bullseye variants on local arm64/Apple Silicon): 1, 1.16, 1.17, 1-bullseye, 1.16-bullseye, 1.17-bullseye, 1-buster, 1.16-buster, 1.17-buster
ARG VARIANT=1.18-bullseye
FROM mcr.microsoft.com/vscode/devcontainers/go:0-${VARIANT}

# Install system tools and configure PostgreSQL
RUN sh -c 'echo "deb http://apt.postgresql.org/pub/repos/apt/ `lsb_release -cs`-pgdg main" >> /etc/apt/sources.list.d/pgdg.list' && \
    wget -q https://www.postgresql.org/media/keys/ACCC4CF8.asc -O - | apt-key add -  && \
    apt-get update && export DEBIAN_FRONTEND=noninteractive \
    && apt-get -y install --no-install-recommends docker.io \
                                                  git \
                                                  httpie \
                                                  postgresql \
                                                  postgresql-contrib \
                                                  libpq-dev \
                                                  shellcheck \
                                                  netcat \
    && rm -rf /var/lib/apt/lists/* \
    && service postgresql start \
    # Setting up user/password and database for 'koko'
    && su -c "psql -c \"CREATE USER koko WITH SUPERUSER PASSWORD 'koko';\"" - postgres \
    && su -c "psql -c \"CREATE DATABASE koko;\"" - postgres

# Force Docker storage driver to vfs
RUN echo '{\n\
  "storage-driver": "vfs"\n\
}'\n >> //etc/docker/daemon.json

# Modify the 'vscode' user to be 'koko'
RUN usermod -l koko vscode \
    && usermod -d /home/vscode -m koko \
    && groupmod -n koko vscode \
    && echo koko ALL=\(root\) NOPASSWD:ALL > /etc/sudoers.d/koko \
    && chmod 0440 /etc/sudoers.d/koko


# Configure Docker engine for use with koko user
RUN usermod -aG docker koko

USER koko
