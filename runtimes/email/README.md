# Email

Email sends Notegic's transactional messages, including welcome, validation,
and security-alert emails.

## Feature contribution

| Capability | Email responsibility |
| --- | --- |
| Onboarding | Delivers welcome and account-validation messages. |
| Account safety | Delivers security alerts. |
| Reliable delivery | Consumes email requests and processes them through a bounded local worker queue. |

## Technical design

This Go runtime consumes Core email requests, renders runtime-owned HTML
templates, and hands delivery to the configured SMTP sender. It exposes only
health endpoints; Core owns the decision to send a message and the contracts
used to request it. Keep email templates and delivery behavior here rather
than in Core or a gateway.

Run the complete local environment with `make compose-up` from the repository
root. SMTP and Kafka configuration come from the SOPS-managed environment and
must not be committed in plaintext.

## Commands

Run these from this directory:

```sh
make test
make test-race
make test-e2e
```

## Structure

```text
commands/           Runtime command entry point
configs/            SMTP, consumer, and renderer configuration
renderers/          Template rendering
senders/            Delivery and queue-facing sender code
templates/          Runtime-owned HTML email templates and assets
transports/core/    Core email-request consumer
transports/status/  Health and startup endpoints
types/              Email-specific types
test/               Runtime tests
```

See [Kafka event contracts](../../docs/system-design/kafka-event-contracts.md)
for the Core-to-Email request boundary.
