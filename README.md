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
it unset for port 8080). With automatic selection the Agent waits while offline
and checks Wi-Fi and Ethernet every two seconds. After connecting or changing
LAN address, scan the QR displayed in the same console window.
An explicit interface address limits the Agent to that address.

Docker is used for the
toolchain and mock server; the Agent itself must run natively in Windows to
access the interactive desktop session.

## Запуск на Windows

Распакуйте архив `dist/Remotehand-windows-amd64.zip` в постоянную папку, например
`C:\Program Files\Remotehand`, и откройте `remotehand.exe` обычным двойным щелчком.
По умолчанию агент управляет рабочим столом и слушает порт 8080. Для тестирования
можно выбрать `REMOTEHAND_INPUT_MODE=mock`. Контрольные суммы находятся в
`dist/SHA256SUMS`.

Для указанной папки и порта можно один раз выполнить в PowerShell от администратора:

```powershell
New-NetFirewallRule -Name RemotehandPrivate8080 -DisplayName 'Remotehand (Private LAN)' `
  -Direction Inbound -Action Allow -Profile Private -Protocol TCP -LocalPort 8080 `
  -RemoteAddress LocalSubnet -Program "$env:ProgramFiles\Remotehand\remotehand.exe"
```

При обновлении закройте агент и замените файл в той же папке. Если вы выбрали
другой путь или порт, укажите его и в правиле. Для работы используйте доверенную
локальную сеть с профилем Private. Запускайте сам агент с обычными правами.
Параметр `REMOTEHAND_POINTER_SENSITIVITY` принимает значения от 0.1 до 3.0.

При смене Wi-Fi или переходе на точку доступа отсканируйте новый QR-код.
Если старая страница открыта, в форму сопряжения можно вставить новую ссылку.
Если ярлык iPhone указывает на недоступный адрес, откройте новый QR в Safari
и добавьте страницу на домашний экран заново.

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

