# ============================================================================
# DEVELOPERS NOTE:
# THE PURPOSE OF THIS DOCKERFILE IS TO BUILD THE CLOUD-SERVICES BACKEND
# EXECUTABLE IN A CONTAINER FOR DEVELOPMENT PURPOSES ON YOUR
# MACHINE. DO NOT RUN THIS IN PRODUCTION ENVIRONMENT.
# ============================================================================

# Start with the official Golang image
# This provides us with a complete Go development environment
FROM golang:1.24.3

# ============================================================================
# SETUP PROJECT DIRECTORY STRUCTURE
# ============================================================================
# Copy all project files to the working directory inside the container
# This creates a directory structure that matches your local environment
COPY . /go/src/github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend
# Set the working directory for subsequent commands
# All commands after this will run from this directory
WORKDIR /go/src/github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend

# ============================================================================
# DEPENDENCY MANAGEMENT
# ============================================================================
# Copy dependency files separately to take advantage of Docker layer caching
# If only dependencies change, we don't need to re-copy all source code
COPY go.mod ./
COPY go.sum ./
# Download all dependencies specified in go.mod
# This ensures all required libraries are available in the container
RUN go mod download

# Copy Go source files
# If you want all files (not just .go files), you would use: COPY . .
COPY *.go .

# ============================================================================
# INSTALL DEVELOPMENT TOOLS
# ============================================================================
# 1. Install CompileDaemon - A tool that watches your files and automatically
#    rebuilds and restarts your application when changes are detected
RUN ["go", "get", "github.com/githubnemo/CompileDaemon"]
RUN ["go", "install", "github.com/githubnemo/CompileDaemon"]

# 2. Install goimports - Does two important things:
#    a) Formats your code according to Go standards (like gofmt)
#    b) Automatically manages your import statements (adds missing ones,
#       removes unused ones, and sorts them)
RUN ["go", "install", "golang.org/x/tools/cmd/goimports@latest"]

# 3. Install golint - Checks your code for style mistakes based on Go
#    community recommendations (not just syntax errors, but stylistic issues)
RUN ["go", "install", "golang.org/x/lint/golint@latest"]

# ============================================================================
# CREATE BUILD SCRIPT
# ============================================================================
# Create a shell script that will run all our code quality tools before building
# This script will:
#   1. Format code and manage imports using goimports
#   2. Check for style issues using golint
#   3. Build the application
RUN echo '#!/bin/sh\n\
    \n\
    # Print a divider line for readability\n\
    echo "============================================================"\n\
    echo "BEGINNING CODE QUALITY CHECKS AND BUILD PROCESS"\n\
    echo "============================================================"\n\
    \n\
    # Step 1: Run goimports to format code and manage imports\n\
    echo "\n[1/3] Formatting code and updating imports..."\n\
    goimports -w .\n\
    if [ $? -ne 0 ]; then\n\
    echo "Error during formatting. See above for details."\n\
    exit 1\n\
    fi\n\
    \n\
    # Step 2: Run golint to check for style issues\n\
    echo "\n[2/3] Checking code style with golint..."\n\
    golint ./...\n\
    # Note: We dont exit on golint errors since they are suggestions\n\
    # If you want to enforce these rules, add: if [ $? -ne 0 ]; then exit 1; fi\n\
    \n\
    # Step 3: Build the application\n\
    echo "\n[3/3] Building application..."\n\
    go build .\n\
    if [ $? -ne 0 ]; then\n\
    echo "Build failed. See above for details."\n\
    exit 1\n\
    fi\n\
    \n\
    echo "\nBuild completed successfully!"\n\
    ' > /go/bin/quality-build.sh

# Make the script executable
RUN chmod +x /go/bin/quality-build.sh

# ============================================================================
# SET UP CONTINUOUS DEVELOPMENT ENVIRONMENT
# ============================================================================
# Configure CompileDaemon to:
#   1. Watch the directory for file changes (-directory="./")
#   2. Poll for changes rather than using file system events (-polling=true)
#     Note: Polling is often more reliable in containerized environments
#   3. Run our quality check script when files change (-build="/go/bin/quality-build.sh")
#   4. Run the compiled binary after successful build (-command="./backend daemon")
#   5. Suppress log prefix for cleaner output (-log-prefix=false)
ENTRYPOINT CompileDaemon -polling=true -log-prefix=false -build="/go/bin/quality-build.sh" -command="./mapleapps-backend daemon" -directory="./"

# ============================================================================
# BUILD INSTRUCTIONS (COMMENTED FOR REFERENCE)
# ============================================================================
# To build this Docker image, run:
# docker build --rm -t mapleapps-backend -f dev.Dockerfile .

# ============================================================================
# EXECUTION INSTRUCTIONS (COMMENTED FOR REFERENCE)
# ============================================================================
# To run the container, execute:
# docker run -d -p 8000:8000 mapleapps-backend
