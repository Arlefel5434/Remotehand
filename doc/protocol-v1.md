# Remotehand protocol v1

## Transport

- Endpoint: `GET /ws` upgraded to WebSocket.
- Messages are UTF-8 JSON text frames, at most 8 KiB.
- Compression is disabled.
- Only one Controller Connection is accepted; a second receives HTTP 409.
- Same-origin browser connections and clients without `Origin` are accepted.
- The Connection is long-lived and has no idle timeout. Agent sends a ping every
  20 seconds and waits 5 seconds for pong.
- Delivery is best-effort and ordered within one Connection. Commands are not
  replayed after reconnect.

## Command envelope

```json
{
  "version": 1,
  "type": "move",
  "id": "optional-id-up-to-64-characters",
  "payload": {"dx": 12, "dy": -4}
}
```

Unknown envelope fields are rejected. `id` is optional and echoed in an error.
Successful Commands are not acknowledged.

## Commands

On connection and after a successful volume or mute command the Agent sends
`{"version":1,"type":"volume_state","payload":{"level":50,"muted":false}}`.
The level is a percentage read from the current default playback endpoint.

| Type | Payload | Limits |
|---|---|---|
| `move` | `{"dx":12,"dy":-4}` | integers from -200 to 200 |
| `scroll` | `{"steps":-1}` | vertical wheel steps, −8 to 8 excluding zero. Positive scrolls up. |
| `click` | `{"button":"left"}` | `left` or `right` |
| `text` | `{"value":"Привет"}` | valid UTF-8, at most 2048 bytes |
| `key` | `{"key":"enter"}` | `enter`, `backspace`, `escape`, `restore_tab`, `new_tab`, `next_tab`, `find` |
| `volume` | `{"delta":5}` | non-zero integer from -10 to 10 |
| `mute` | `{"muted":true}` | boolean |

## Error envelope

```json
{
  "version": 1,
  "type": "error",
  "id": "optional-command-id",
  "payload": {
    "code": "invalid_payload",
    "message": "dx and dy must be between -200 and 200"
  }
}
```

Unknown Command types and invalid payloads return an error without closing the
Connection. Malformed envelopes close with 1007, unsupported versions with
1002, oversized messages with 1009, and Agent shutdown with 1001.
