# Remotehand

Remotehand provides a local control session between a phone and a personal
Windows computer.

## Language

**Agent**:
The native Remotehand process running in the interactive Windows session.
_Avoid_: Server, desktop app, host

**Controller**:
The phone web interface used to send control commands to an Agent.
_Avoid_: Client, remote, phone app

**Input Adapter**:
The boundary that applies Controller commands either to Windows or to a safe
test double.
_Avoid_: Driver, input service

**Session**:
One authorized connection from a Controller to an Agent.
_Avoid_: Pairing, socket, connection
