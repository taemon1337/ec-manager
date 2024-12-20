FROM hashicorp/packer:latest

# Install AWS CLI
RUN apk add --no-cache \
    python3 \
    py3-pip \
    && python3 -m pip install --break-system-packages --no-cache-dir awscli

# Set working directory
WORKDIR /workspace

# Copy Packer configurations
COPY . .

# Initialize Packer plugins during build
RUN packer init rhel9/rhel9.pkr.hcl && \
    packer init ubuntu/ubuntu.pkr.hcl

# Default command
ENTRYPOINT ["packer"]
