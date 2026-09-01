# Development log archive

Each Markdown file in this directory is generated from Git history. Daily snapshots are grouped under `YYYY-MM/` directories. The generator records commit subjects and changed repository areas without inventing requirements or implementation intent.

Daily snapshots use this layout:

```text
docs/devlogs/YYYY-MM/YYYY-MM-DD.md
```

Generate a new snapshot from the repository root:

```bash
make devlog
```

To enable the commit-time validation hook:

```bash
make install-hooks
```

Before committing, the pre-commit hook verifies that today's devlog is staged and exactly matches a fresh temporary generation. It never overwrites the working tree and removes its temporary files when it exits.

The recent snapshot index is maintained in the generated Development log
section at the bottom of the repository [README](../../README.md).

## Commit policy

Never create a commit whose only purpose is to update a devlog. Run `make devlog`
while preparing a meaningful code or documentation change, then stage the
generated devlog together with that change in the same commit. If there is no
meaningful repository change, do not create a devlog-only commit.

Recommended workflow:

```bash
make devlog
git add .
git commit -m "your message"
git push origin main
```
