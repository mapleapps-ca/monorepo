# ============================================================================
# DEVELOPERS NOTE:
# THE PURPOSE OF THIS DOCKERFILE IS TO BUILD THE CLOUD-SERVICES BACKEND
# EXECUTABLE IN A CONTAINER FOR DEVELOPMENT PURPOSES ON YOUR
# MACHINE. DO NOT RUN THIS IN PRODUCTION ENVIRONMENT.
# ============================================================================

# Start with the official Golang image
FROM golang:1.24.3

# ============================================================================
# SETUP PROJECT DIRECTORY STRUCTURE
# ============================================================================
# Set the working directory first
WORKDIR /go/src/github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend

# ============================================================================
# DEPENDENCY MANAGEMENT (DO THIS FIRST FOR BETTER CACHING)
# ============================================================================
# Copy dependency files first to take advantage of Docker layer caching
COPY go.mod go.sum ./
# Download all dependencies
RUN go mod download

# ============================================================================
# INSTALL DEVELOPMENT TOOLS
# ============================================================================
# Install CompileDaemon for hot reloading
RUN go install github.com/githubnemo/CompileDaemon@latest

# Install goimports for code formatting
RUN go install golang.org/x/tools/cmd/goimports@latest

# Install golint for style checking (note: golint is deprecated, using staticcheck instead)
RUN go install honnef.co/go/tools/cmd/staticcheck@latest

# ============================================================================
# CREATE SIMPLIFIED BUILD SCRIPT
# ============================================================================
RUN echo '#!/bin/sh\n\
    echo "============================================================"\n\
    echo "BEGINNING BUILD PROCESS"\n\
    echo "============================================================"\n\
    \n\
    echo "[1/2] Running goimports..."\n\
    goimports -w . || echo "Warning: goimports had issues"\n\
    \n\
    echo "[2/2] Building application..."\n\
    go build -o mapleapps-backend .\n\
    if [ $? -ne 0 ]; then\n\
    echo "Build failed!"\n\
    exit 1\n\
    fi\n\
    \n\
    echo "Build completed successfully!"\n\
    ' > /go/bin/build.sh && chmod +x /go/bin/build.sh

# ============================================================================
# COPY SOURCE CODE (AFTER DEPENDENCIES)
# ============================================================================
# Copy all source code
COPY . .

# ============================================================================
# SET UP CONTINUOUS DEVELOPMENT ENVIRONMENT
# ============================================================================
# Use CompileDaemon with simpler configuration
ENTRYPOINT ["CompileDaemon", "-polling=true", "-log-prefix=false", "-build=/go/bin/build.sh", "-command=./mapleapps-backend daemon", "-directory=./"]
