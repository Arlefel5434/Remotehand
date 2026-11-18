# Remotehand

Remotehand is a personal remote control for a Windows laptop. The MVP uses a
phone browser as a touchpad, keyboard input panel, and system volume control.

The implementation is intentionally split into independently verifiable stages.
Start with [doc/README.md](doc/README.md).

## Development with Docker

Requirements: Docker Desktop with WSL integration, Docker Compose, and GNU Make.
Go does not need to be installed in WSL.

```bash
make check
```

Run the mock HTTP server:

```bash
make server
curl http://localhost:8080/health
```

Build the native Windows Agent:

```bash
make build-windows
```

The executable is written to `dist/remotehand.exe`. Docker is used for the
toolchain and mock server; the Agent itself must run natively in Windows to
access the interactive desktop session.

