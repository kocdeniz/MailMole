# Installation Guide

This guide covers how to install and run MailMole on your system.

## Requirements

- **Go 1.24+** - Required to build from source
- **IMAP Access** - Both source and destination mail servers
- **Network Access** - Access to IMAP ports (usually 993 for TLS)

## Run from Source

The easiest way to run MailMole is directly from source:

```bash
git clone https://github.com/kocdeniz/MailMole.git
cd MailMole
go run .
```

## Build Static Binary

For a portable binary that works on any Linux system:

```bash
# Clone the repository
git clone https://github.com/kocdeniz/MailMole.git
cd MailMole

# Build static binary (no C dependencies)
CGO_ENABLED=0 go build -o mailmole .

# Make executable
chmod +x mailmole
```

## Run the Binary

```bash
./mailmole
```

## Verifying Installation

When you run MailMole, you should see the intro screen:

```
   /\_____/\
  /  o   o  \
 ( ==  ^  == )
  )         (
 (           )
( (  )   (  )  )
 (             )
  ((     )  ))
   ((    ))
    (( ))
     ( )
      v1.x.x

Press any key to continue...
```

## Next Steps

- [Getting Started Guide](getting-started.md) - Learn how to use MailMole
- [Manual Mode](usage/manual-mode.md) - Migrate single accounts
- [Bulk Mode](usage/bulk-mode.md) - Migrate multiple accounts
