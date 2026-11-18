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

**Connection**:
A long-lived, not yet authorized WebSocket transport between a Controller and
an Agent.
_Avoid_: Session, socket

**Session**:
One authorized connection from a Controller to an Agent.
_Avoid_: Pairing, socket, Connection

**Command**:
A one-way instruction sent by a Controller for an Agent to apply.
_Avoid_: Event, request, message
