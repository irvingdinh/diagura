---
name: diagura-go-best-practices
description: Opinionated usage patterns for diagura packages. This skill should be used when writing, reviewing, or refactoring code that uses diagura's core packages (config, container, etc.). Triggers on tasks involving diagura package usage, module registration, config management, or application lifecycle.
license: MIT
metadata:
  author: Irving Dinh <irving.dinh@gmail.com>
  version: "1.0.0"
---

# Diagura Go Best Practices

Opinionated usage patterns for diagura's core packages. These are not general Go best practices (see `go-best-practices` for that) — these are "the diagura way" of building applications with this framework.

## When to Apply

Reference these guidelines when:
- Using the config package (`localhost/app/core/config`)
- Registering defaults and validation rules for config keys
- Writing module constructors that depend on configuration
- Testing code that uses the config package
- Adding new config keys to the application

## Rule Categories

| Priority | Category | Prefix |
|----------|----------|--------|
| 1 | Core Packages | `core-` |

## Quick Reference

### 1. Core Packages

- `core-config` - Config package lifecycle, defaults, validation, getters, testing

## How to Use

Read individual rule files for detailed explanations and code examples:

```
rules/core-config.md
```

Each rule file contains:
- Prescriptive guidance on how to use the package
- Incorrect/correct code examples where helpful
- The rationale behind each convention
