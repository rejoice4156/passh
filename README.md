# Passh - SSH Key-Backed Password Manager

A simple terminal password manager backed by SSH keys. Passh allows you to securely store and retrieve passwords using your existing SSH keys, avoiding the complexity of GPG key management.

## Features

- Uses SSH keys for encryption/decryption
- Simple CLI interface
- Password generation
- Hierarchical password organization
- Minimal and easy to use

## Installation

Install using go:

```bash
go install github.com/rejoice4156/passh/cmd/passh@latest
```

Or build from source:

```bash
git clone https://github.com/rejoice4156/passh.git
cd passh
go build -o passh ./cmd/passh
```

## Usage

Passh provides a simple CLI interface for managing your passwords.

### Global Options

These options can be used with any command:

```bash
--store string       Password store directory (default: ~/.passh)
--public-key string  SSH public key path (default: ~/.ssh/id_rsa.pub or ~/.ssh/id_ed25519.pub)
--private-key string SSH private key path (default: ~/.ssh/id_rsa or ~/.ssh/id_ed25519)
--help, -h           Display help for the command
```

### Basic Commands

#### Adding Passwords

Store a new password:

```bash
# You will be prompted to enter the password
passh add github/personal
passh add email/work
passh add servers/production/db1
```

#### Generating Passwords

Generate and store a random password:

```bash
# Generate a 16-character password (default)
passh generate github/work

# Generate a longer password
passh generate bank/online --length 24

# Generate a password without symbols
passh generate wifi/home --length 12 --no-symbols
```

#### Retrieving Passwords

Get a stored password:

```bash
# Print password to stdout
passh get github/personal

# Copy to clipboard (pipe to clipboard utility)
passh get email/work | pbcopy  # macOS
passh get email/work | xclip -selection clipboard  # Linux
```

#### Listing Passwords

List all stored passwords:

```bash
# List all passwords
passh list

# You can use grep to filter results
passh list | grep github
```

#### Deleting Passwords

Delete a password:

```bash
passh delete github/personal
```

#### Organization

Passh organizes passwords in a hierarchical structure. Use forward slashes to create directories:

```bash
category/subcategory/entry
```

Examples:

- email/personal
- email/work
- servers/production/db1
- servers/staging/db1

This hierarchy is reflected in the filesystem structure under your password store directory.

#### Using Different SSH Keys

By default, Passh uses your SSH keys from ~/.ssh/, but you can specify different keys:

```bash
passh --public-key ~/.ssh/custom_key.pub --private-key ~/.ssh/custom_key get github/personal
```

#### Using a Different Store

You can specify a different location for your password store:

```bash
passh --store /path/to/custom/store add newentry
```

### Storage

By default, passwords are stored in ~/.passh/. You can change this with the --store flag.

### Security

- Passwords are encrypted using SSH keys
- Each password is stored in its own file
- Files are created with restricted permissions (0600)

### Help
For more information on a specific command, use the `--help` flag:

```bash
passh add --help
passh get --help
passh list --help
passh delete --help
passh generate --help
```
