# [Choice] Go version (use -bullseye variants on local arm64/Apple Silicon): 1, 1.16, 1.17, 1-bullseye, 1.16-bullseye, 1.17-bullseye, 1-buster, 1.16-buster, 1.17-buster
ARG VARIANT=1-bullseye
FROM mcr.microsoft.com/vscode/devcontainers/go:0-${VARIANT}

# [Choice] Node.js version: none, lts/*, 16, 14, 12, 10
ARG NODE_VERSION="none"
RUN if [ "${NODE_VERSION}" != "none" ]; then su vscode -c "umask 0002 && . /usr/local/share/nvm/nvm.sh && nvm install ${NODE_VERSION} 2>&1"; fi

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

# Modify the 'vscode' user to be 'koko'
RUN usermod -l koko vscode \
    && usermod -d /home/vscode -m koko \
    && groupmod -n koko vscode \
    && echo koko ALL=\(root\) NOPASSWD:ALL > /etc/sudoers.d/koko \
    && chmod 0440 /etc/sudoers.d/koko


# Configure Docker engine for use with koko user
RUN usermod -aG docker koko

USER koko
ENV GO111MODULE=on
RUN go get -x github.com/spf13/cobra/cobra

# [Optional] Uncomment this line to install global node packages.
# RUN su koko -c "source /usr/local/share/nvm/nvm.sh && npm install -g <your-package-here>" 2>&1
