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

Until authenticated Sessions are implemented, the Windows Agent defaults to
the safe mock adapter. Native pointer control must be enabled explicitly:

```powershell
$env:REMOTEHAND_INPUT_MODE = "native"
$env:REMOTEHAND_POINTER_SENSITIVITY = "1.0"
.\dist\remotehand.exe
```

Sensitivity accepts values from `0.1` to `3.0`. Only enable native input on a
trusted private network during development.

Open `http://<windows-ip>:8080/` on the phone, tap the text area to use the
phone's system keyboard, then send the text to the currently focused Windows
application. The Controller reconnects automatically if the connection drops.

