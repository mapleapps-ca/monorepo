
# MapleFile CLI

A secure, end-to-end encrypted file storage and collaboration platform with a powerful command-line interface.

[![Version](https://img.shields.io/badge/version-1.0.0--alpha-orange.svg)]()
[![Go Version](https://img.shields.io/badge/go-1.24.3-blue.svg)](https://golang.org/doc/go1.24)
[![License](https://img.shields.io/badge/license-Open%20Source-green.svg)]()

## 🔐 Features

- **End-to-End Encryption (E2EE)**: All files are encrypted before leaving your device
- **Secure Collections**: Organize files in encrypted collections with granular sharing
- **Hybrid Storage**: Keep files local-only, cloud-only, or synchronized across devices
- **Collaboration**: Share collections securely with other users using E2EE
- **Account Recovery**: Secure account recovery using cryptographic recovery keys
- **Cross-Platform**: Works on Windows, macOS, and Linux
- **Flexible Sync**: Granular control over synchronization between local and cloud storage

## 📋 Table of Contents

- [Installation](#installation)
- [Quick Start](#quick-start)
- [Authentication Commands](#authentication--user-management)
- [Collections Management](#collections-management)
- [Files Management](#files-management)
- [Synchronization](#synchronization)
- [Configuration](#configuration--utilities)
- [Advanced Usage](#advanced-usage)
- [Security Model](#security-model)
- [Troubleshooting](#troubleshooting)

## 🚀 Installation

### Pre-built Binaries

Coming soon ...

### Build from Source

```bash
git clone https://github.com/mapleapps-ca/maplefile-cli.git
cd maplefile-cli
go build -o maplefile-cli .
```

### Package Managers

Coming soon ...

## ⚡ Quick Start

1. **Register a new account:**
   ```bash
   maplefile-cli register \
     --email john@example.com \
     --password "your-secure-password" \
     --firstname "John" \
     --lastname "Doe" \
     --agree-terms
   ```

2. **Verify your email:**
   ```bash
   maplefile-cli verify-email --code 123456
   ```

3. **Log in:**
   ```bash
   maplefile-cli login --email john@example.com
   ```

4. **Create your first collection:**
   ```bash
   maplefile-cli collections create "My Documents" --password your-password
   ```

5. **Add a file:**
   ```bash
   maplefile-cli files add "/path/to/document.pdf" \
     --collection COLLECTION_ID \
     --password your-password
   ```

6. **Sync with cloud:**
   ```bash
   maplefile-cli sync --password your-password
   ```

## 🔑 Authentication & User Management

### Registration

Register a new MapleFile account:

```bash
# Basic registration
maplefile-cli register \
  --email user@example.com \
  --password "secure-password" \
  --firstname "First" \
  --lastname "Last" \
  --agree-terms

# Full registration with optional fields
maplefile-cli register \
  --email user@example.com \
  --password "secure-password" \
  --firstname "First" \
  --lastname "Last" \
  --phone "+1234567890" \
  --country "Canada" \
  --timezone "America/Toronto" \
  --agree-terms \
  --agree-promotions \
  --agree-tracking
```

### Email Verification

Verify your email address after registration:

```bash
# Verify with code from email
maplefile-cli verify-email --code 123456

# Or pass code as argument
maplefile-cli verify-email 123456
```

### Login

#### Unified Login (Recommended)
```bash
# Interactive login - walks through all steps
maplefile-cli login --email user@example.com

# Non-interactive login (if you have all details)
maplefile-cli login \
  --email user@example.com \
  --ott 123456 \
  --password your-password
```

#### Manual Login Steps
```bash
# Step 1: Request one-time token
maplefile-cli request-login-token --email user@example.com

# Step 2: Verify token from email
maplefile-cli verify-login-token --email user@example.com --ott 123456

# Step 3: Complete login with password
maplefile-cli complete-login --email user@example.com
```

### User Profile Management

```bash
# View your profile
maplefile-cli me get

# View detailed profile
maplefile-cli me get --verbose

# Update profile
maplefile-cli me update \
  --email user@example.com \
  --first-name "John" \
  --last-name "Doe" \
  --phone "+1234567890" \
  --country "Canada" \
  --timezone "America/Toronto"
```

### Account Recovery

```bash
# Show your recovery key (save this securely!)
maplefile-cli show-recovery-key

# Start account recovery
maplefile-cli recovery start --email user@example.com --recovery-key-file ~/recovery.key

# Complete recovery with new password
maplefile-cli recovery complete

# Check recovery status
maplefile-cli recovery status
```

### Session Management

```bash
# Refresh authentication tokens
maplefile-cli refreshtoken

# Log out
maplefile-cli logout
```

## 📁 Collections Management

Collections are secure containers for organizing your encrypted files.

### Create Collections

```bash
# Create root collection
maplefile-cli collections create "My Documents" --password your-password

# Create album collection
maplefile-cli collections create "Vacation Photos" \
  --type album \
  --password your-password

# Create sub-collection
maplefile-cli collections create "Project Files" \
  --parent PARENT_COLLECTION_ID \
  --password your-password
```

### List Collections

```bash
# List all root collections
maplefile-cli collections list

# List sub-collections of a parent
maplefile-cli collections list --parent COLLECTION_ID

# List with detailed information
maplefile-cli collections list --verbose

# List by state
maplefile-cli collections list --state active
maplefile-cli collections list --state deleted
maplefile-cli collections list --state archived

# List locally modified collections
maplefile-cli collections list --modified
```

### Delete and Restore Collections

```bash
# Soft delete (can be restored)
maplefile-cli collections delete COLLECTION_ID

# Archive collection
maplefile-cli collections delete COLLECTION_ID --archive

# Delete collection and all children
maplefile-cli collections delete COLLECTION_ID --with-children

# Skip confirmation
maplefile-cli collections delete COLLECTION_ID --force

# Restore deleted collection
maplefile-cli collections restore COLLECTION_ID
```

### Share Collections

```bash
# Share with read-only access
maplefile-cli collections share \
  --id COLLECTION_ID \
  --email recipient@example.com \
  --permission read_only \
  --password your-password

# Share with read-write access including sub-collections
maplefile-cli collections share \
  --id COLLECTION_ID \
  --email recipient@example.com \
  --permission read_write \
  --descendants \
  --password your-password

# Share with admin access
maplefile-cli collections share \
  --id COLLECTION_ID \
  --email recipient@example.com \
  --permission admin \
  --password your-password

# Remove user access
maplefile-cli collections unshare \
  --id COLLECTION_ID \
  --email recipient@example.com

# List collection members
maplefile-cli collections members --id COLLECTION_ID

# List collections shared with you
maplefile-cli collections list-shared
```

## 📄 Files Management

### Add Files

```bash
# Add file with auto-upload (recommended)
maplefile-cli files add "/path/to/document.pdf" \
  --collection COLLECTION_ID \
  --password your-password

# Add file locally only (upload later)
maplefile-cli files add "/path/to/photo.jpg" \
  --collection COLLECTION_ID \
  --local-only \
  --password your-password

# Add with custom name
maplefile-cli files add "/path/to/file.txt" \
  --collection COLLECTION_ID \
  --name "My Document" \
  --password your-password

# Add with encrypted-only storage (most secure)
maplefile-cli files add "/path/to/secret.pdf" \
  --collection COLLECTION_ID \
  --storage-mode encrypted_only \
  --password your-password
```

### List Files

```bash
# List files in specific collection
maplefile-cli files list --collection COLLECTION_ID

# List with detailed information
maplefile-cli files list --collection COLLECTION_ID --verbose
```

### Download Files

```bash
# Download file to current directory
maplefile-cli files get FILE_ID --password your-password

# Download to specific directory
maplefile-cli files get FILE_ID \
  --output ~/Downloads/ \
  --password your-password

# Download with custom filename
maplefile-cli files get FILE_ID \
  --output ~/Documents/my-file.pdf \
  --password your-password

# Overwrite existing files
maplefile-cli files get FILE_ID \
  --output existing-file.txt \
  --force \
  --password your-password
```

### Delete Files

```bash
# Delete file completely (both local and cloud)
maplefile-cli files delete FILE_ID --password your-password

# Delete only local copy (keep in cloud)
maplefile-cli files delete FILE_ID --local-only

# Delete only cloud copy (keep local)
maplefile-cli files delete FILE_ID --cloud-only --password your-password

# Skip confirmation
maplefile-cli files delete FILE_ID --force --password your-password
```

### File Synchronization

Control where your files are stored:

```bash
# Move file to cloud-only storage (free up local space)
maplefile-cli files filesync offload \
  --file-id FILE_ID \
  --password your-password

# Download cloud-only file to local storage
maplefile-cli files filesync onload \
  --file-id FILE_ID \
  --password your-password

# Delete file from cloud only
maplefile-cli files filesync cloud-only-delete \
  --file-id FILE_ID \
  --password your-password
```

### File Security Operations

```bash
# Lock file to encrypted-only mode (maximum security)
maplefile-cli files misc lock \
  --file-id FILE_ID \
  --password your-password

# Unlock file to access decrypted content
maplefile-cli files misc unlock \
  --file-id FILE_ID \
  --password your-password \
  --mode hybrid

# Debug E2EE key chain issues
maplefile-cli files misc debug-e2ee \
  --file-id FILE_ID \
  --password your-password
```

## 🔄 Synchronization

### Main Sync Commands

```bash
# Sync everything (collections + files) - recommended
maplefile-cli sync --password your-password

# Sync only collections
maplefile-cli sync --collections --password your-password

# Sync only file metadata
maplefile-cli sync --files --password your-password

# Custom batch sizes for large datasets
maplefile-cli sync \
  --collection-batch-size 25 \
  --file-batch-size 30 \
  --password your-password
```

### Debug Sync Issues

```bash
# Run full diagnostic
maplefile-cli sync debug --password your-password

# Check specific components
maplefile-cli sync debug --auth --password your-password
maplefile-cli sync debug --network
maplefile-cli sync debug --sync-state --password your-password
```

## ⚙️ Configuration & Utilities

### Configuration Management

```bash
# Get current cloud provider address
maplefile-cli config get

# Set cloud provider address
maplefile-cli config set https://api.maplefile.com
```

### Cloud Operations

```bash
# Look up public user information
maplefile-cli cloud public-user-lookup --email user@example.com
```

### Health Check

```bash
# Check server connectivity
maplefile-cli healthcheck
```

### Version Information

```bash
# Show version
maplefile-cli version
```

## 🔧 Advanced Usage

### Storage Modes

MapleFile supports three storage modes for files:

- **`encrypted_only`**: Only encrypted version stored locally (most secure, requires decryption for access)
- **`hybrid`**: Both encrypted and decrypted versions stored (convenient, default)
- **`decrypted_only`**: Only decrypted version stored (not recommended for sensitive files)

### Sync Strategies

When sharing collections, you can control how local state is updated:

- **`immediate`** (default): Update local collection immediately after cloud sharing
- **`cloud-pull`**: Pull fresh collection data from cloud after sharing
- **`none`**: Don't update local state (original behavior)

### Batch Processing

For large datasets, use batch processing options:

```bash
# Sync with custom batch sizes
maplefile-cli sync \
  --collection-batch-size 50 \
  --file-batch-size 100 \
  --max-batches 200 \
  --password your-password
```

## 🔒 Security Model

### End-to-End Encryption

MapleFile uses a complete E2EE key chain:

1. **Password** → Key Encryption Key (KEK)
2. **KEK** → Master Key
3. **Master Key** → Collection Keys
4. **Collection Keys** → File Keys
5. **File Keys** → Encrypted File Content

### Key Management

- **Recovery Keys**: Cryptographic keys for account recovery (store securely!)
- **Collection Keys**: Unique encryption keys per collection
- **File Keys**: Individual encryption keys per file
- **Zero-Knowledge**: Server never sees your passwords or decrypted data

### Best Practices

1. **Use strong passwords** (8+ characters with mixed case, numbers, symbols)
2. **Save your recovery key** in a secure location (password manager, safe)
3. **Use `encrypted_only` storage mode** for highly sensitive files
4. **Regularly sync** to keep data backed up
5. **Verify recipients** before sharing collections

## 🐛 Troubleshooting

### Common Issues

#### Authentication Problems
```bash
# Check if logged in
maplefile-cli me get

# Refresh expired tokens
maplefile-cli refreshtoken

# Debug authentication
maplefile-cli sync debug --auth --password your-password
```

#### Sync Issues
```bash
# Run sync diagnostics
maplefile-cli sync debug --password your-password

# Check network connectivity
maplefile-cli sync debug --network

# Check sync state consistency
maplefile-cli sync debug --sync-state --password your-password
```

#### File Access Problems
```bash
# Debug E2EE key chain
maplefile-cli files misc debug-e2ee --file-id FILE_ID --password your-password

# Check file status
maplefile-cli files list --collection COLLECTION_ID --verbose
```

#### Server Connectivity
```bash
# Test server connection
maplefile-cli healthcheck

# Check configuration
maplefile-cli config get
```

### Error Messages

| Error | Solution |
|-------|----------|
| "Password is required for E2EE operations" | Add `--password your-password` flag |
| "Incorrect password" | Verify your password is correct |
| "Collection not found" | Check collection ID with `collections list` |
| "File not found" | Verify file exists with `files list` |
| "Permission denied" | Check you have access to the resource |
| "Network error" | Check internet connection and server status |

### Getting Help

```bash
# Show help for any command
maplefile-cli --help
maplefile-cli COMMAND --help
maplefile-cli COMMAND SUBCOMMAND --help

# Examples:
maplefile-cli collections --help
maplefile-cli files add --help
maplefile-cli sync debug --help
```

## 📞 Support

- **Issues**: [GitHub Issues](https://github.com/mapleapps-ca/maplefile-cli/issues)

## 📄 License

This application is licensed under the [**GNU Affero General Public License v3.0**](https://opensource.org/license/agpl-v3). See [LICENSE](LICENSE) for more information.

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

---

**MapleFile CLI** - Secure, encrypted file storage that puts your privacy first.
