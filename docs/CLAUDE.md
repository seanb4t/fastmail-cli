# docs/

Project documentation and Zensical documentation site.

## Directories

| Directory | Description |
|-----------|-------------|
| `reference/` | API and command reference docs |
| `site/` | Zensical documentation site (deployed to Cloudflare Pages) |

## Site Structure

The `site/` directory contains a Zensical static site:
- `.zensical.yaml` - Site configuration and navigation
- `index.md` - Home page
- `getting-started.md` - Quick start guide
- `reference/` - CLI and MCP documentation

## Development

```bash
# Preview docs locally
task docs:preview

# Build docs
task docs:build
```

## Conventions

- Markdown format for all docs
- Keep docs close to code (CLAUDE.md in packages)
- Reference docs generated where possible
- User guides in docs/, API docs in code
