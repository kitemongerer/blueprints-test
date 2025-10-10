FROM buildpack-deps:bookworm

ENV LANG=C.UTF-8 \
    LC_ALL=C.UTF-8 \
    TERM=xterm-256color \
    DEBIAN_FRONTEND=noninteractive

# Install dependencies
RUN apt-get -qq update \
  && apt-get -qq install -y --no-install-recommends \
    apt-utils \
    ca-certificates \
    apt-transport-https \
    curl \
    dnsutils \
    lsb-release \
    gnupg2 \
    jq \
    unzip \
    git \
    pigz \
    zip \
    time \
  > /dev/null \
  && apt-get -qq clean \
  && rm -rf \
    /var/lib/apt/lists/* \
    /tmp/* \
    /var/tmp/* \
  && :

# Install AWS CLI v2
RUN curl "https://awscli.amazonaws.com/awscli-exe-linux-x86_64-2.18.13.zip" -o "awscliv2.zip" \
  && unzip awscliv2.zip \
  && ./aws/install \
  && rm awscliv2.zip \
  && rm -rf aws

# Copy s5cmd from the official s5cmd image
COPY --from=peakcom/s5cmd:latest s5cmd /usr/local/bin/s5cmd

# Verify installations
RUN aws --version && \
    s5cmd version && \
    which time

# Set working directory
WORKDIR /app

# Default shell
CMD ["/bin/bash"]