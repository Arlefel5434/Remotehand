# Web client

The Go binary embeds the Controller, manifest, PNG icons and service worker.
On iPhone, open the Controller in Safari, tap Share, then Add to Home Screen.
If the standalone window requests pairing, copy the link from the laptop QR
using the camera and paste it into the pairing field. Close the Safari tab
connected to the Agent first to release the single Controller slot.

The Agent's LAN HTTP address supports the live Controller. Service worker
registration requires HTTPS or localhost. On a secure origin it provides a
retry page when the Agent is unreachable. Commands and credentials are never
queued for offline use.
