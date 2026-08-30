# Notification contracts

This package owns the versioned contracts emitted by the Notification runtime.
Core publishes `notification.requested.v1` using the Core event contract. After
the Notification database transaction commits, Notification publishes
`notification.created.v1` for RealtimeGateway. User identity is an external
`user_public_id`; Notification does not create a cross-database foreign key.

Typed notification payloads and their `validate` tags live under `types/`.
Custom validator functions are registered by the Notification runtime under
`runtimes/notification/validations/`; the runtime does not duplicate payload
structs outside this contract package.
