FROM alpine:latest

# Install dependencies
RUN apk add --no-cache \
    python3 \
    py3-pip \
    curl \
    time \
    bash \
    ca-certificates

# Install AWS CLI v2
RUN curl "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "awscliv2.zip" && \
    unzip awscliv2.zip && \
    ./aws/install && \
    rm -rf awscliv2.zip aws/

# Install s5cmd
RUN curl -L "https://github.com/peak/s5cmd/releases/latest/download/s5cmd_$(uname -s | tr '[:upper:]' '[:lower:]')_$(uname -m | sed 's/x86_64/amd64/').tar.gz" -o s5cmd.tar.gz && \
    tar -xzf s5cmd.tar.gz && \
    mv s5cmd /usr/local/bin/ && \
    chmod +x /usr/local/bin/s5cmd && \
    rm s5cmd.tar.gz

# Verify installations
RUN aws --version && \
    s5cmd version && \
    time --version

# Set working directory
WORKDIR /app

# Default shell
CMD ["/bin/bash"]
