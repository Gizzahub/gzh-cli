# Context Documentation - gzh-cli

This directory contains detailed context documentation extracted from CLAUDE.md for LLM optimization.

## Purpose

Keep CLAUDE.md under 300 lines while maintaining comprehensive guidance through linked context documents.

## Files

| File | Purpose | When to Read |
|------|---------|--------------|
| [architecture-guide.md](architecture-guide.md) | Integration pattern, extensions, lifecycle | Before major changes |
| [testing-guide.md](testing-guide.md) | Test organization, mocking, coverage | Writing tests |
| [build-guide.md](build-guide.md) | Build workflow, troubleshooting | Build issues |
| [common-tasks.md](common-tasks.md) | Adding commands, modifying wrappers | Daily development |

## Quick Access

**New to the project?** Start here:
1. Read CLAUDE.md (quick overview)
2. Read cmd/AGENTS_COMMON.md (project conventions)
3. Read architecture-guide.md (understand integration pattern)
4. Read common-tasks.md (see how to work)

**Adding a command?**
- Check common-tasks.md for workflow
- Read relevant cmd/{module}/AGENTS.md

**Modifying integration library?**
- Check architecture-guide.md for wrapper vs library decision
- Read common-tasks.md for local development setup

**Build problems?**
- Check build-guide.md for troubleshooting

**Writing tests?**
- Check testing-guide.md for organization and helpers
