# Use Docker for development and a native Windows Agent

Remotehand uses Docker Compose for development, tests, the mock server, and
cross-compilation, but runs the resulting Agent natively in the interactive
Windows session. A Linux container cannot access the host session required by
Windows `SendInput` and Core Audio, so a fully containerized Agent would not
provide the core product behavior.

## Consequences

The development toolchain does not depend on Go installed in WSL, while manual
acceptance tests must execute `remotehand.exe` on Windows. Windows integration
must remain CGO-free so the Linux container can cross-compile it reliably.
