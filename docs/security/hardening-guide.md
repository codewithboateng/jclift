# Hardening Guide

## Build & Binary

- Build with `-trimpath -ldflags="-s -w"`
- Prefer running as a dedicated OS user

## Network

- Bind to loopback or private interface
- Place behind reverse proxy that enforces TLS
- Set strict CORS (`--cors-allow` domains list)

## Auth

- Create minimal admins; prefer `viewer` for most users
- Rotate admin credentials; store securely
- Enforce long passwords; consider OS-level PAM in v1+

## Storage

- Put `jclift.db` on restricted path
- Back up with filesystem ACLs
- Optionally use full-disk encryption

## Logging

- Keep logs free of secrets and SYSIN bodies by default
- Rotate logs via service manager

## Air-Gapped

- Use `make pkg-airgap` bundle
- Verify checksums before install
- Disable outbound egress

## Runtime

- Systemd unit with `NoNewPrivileges=yes`, `ProtectSystem=full`
- Resource limits (CPU/Memory) if shared host
