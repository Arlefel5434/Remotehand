// Package web provides the embedded phone Controller.
package web

import "embed"

// Files contains the Controller assets served by the Agent.
//
//go:embed index.html app.js styles.css manifest.webmanifest sw.js icons/*.png
var Files embed.FS
