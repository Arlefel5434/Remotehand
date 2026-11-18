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
toolchain. At startup the Windows console displays a QR code and URL for each
listening LAN address. Scan the Wi-Fi or Ethernet code with the phone camera.
For automatic address selection set `REMOTEHAND_LISTEN_ADDR=:18080` (or leave
it unset for port 8080). Restart after changing networks to print fresh codes.
An explicit interface address limits the Agent to that address.

Docker is used for the
toolchain and mock server; the Agent itself must run natively in Windows to
access the interactive desktop session.

The Windows Agent defaults to the mock adapter. Enable native input explicitly:

```powershell
$env:REMOTEHAND_INPUT_MODE = "native"
$env:REMOTEHAND_POINTER_SENSITIVITY = "1.0"
.\dist\remotehand.exe
```

Sensitivity accepts values from `0.1` to `3.0`. Only enable native input on a
trusted private network during development.

Scan the QR on the laptop, tap the text area to use the
phone's system keyboard, then send the text to the currently focused Windows
application. The Controller reconnects automatically if the connection drops.

The QR carries a random credential valid for the current Agent run. Keep it
private. Reconnects in the same browser tab reuse it. After a restart, scan
the new QR. Enter `disconnect` in the Windows console to stop control and
invalidate the credential. HTTP/WebSocket traffic is unencrypted. Use only
a trusted private LAN. For Docker mock testing, the local operator can copy
`/run/pair/token` from the server container and open the Controller with
`#pair=<token>` appended to its URL.

