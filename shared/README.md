# Shared

Cross-runtime helpers live here, beside `runtimes/` and `contracts/`.

- `lib/` is portable and must not import Notegic code.
- `util/` contains reusable application-facing utilities such as EditableBlock
  flattening and shared response/exception writers.
- `cookies/`, `exceptions/`, and `tokens/` remain semantic shared boundaries
  rather than generic utility packages.
- Other packages may provide shared application utilities when more than one runtime needs them.
- Domain business logic remains with its owning service.
