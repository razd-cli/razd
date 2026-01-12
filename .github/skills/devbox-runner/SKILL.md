---
name: devbox-runner
description: Ensures all development commands (php, composer, pnpm, npm, node, npx, openspec) are executed within the Devbox environment. Use this skill when running any CLI tool in this project to guarantee correct language versions and PHP extensions are available.
compatibility: Requires Devbox (jetify-com/devbox) installed on the system. Works with any AI coding assistant that supports Agent Skills.
allowed-tools: Bash(devbox:*)
---

# Devbox Runner

This project uses [Devbox](https://www.jetify.com/devbox) to manage its development environment, including PHP, Node.js, Composer, and PNPM. Running tools directly without Devbox may use incompatible system-wide versions or miss required PHP extensions.

## Command Prefixing

**Always** prefix the following commands with `devbox run`:

| Direct Command | Devbox Command |
|----------------|----------------|
| `php`          | `devbox run php` |
| `composer`     | `devbox run composer` |
| `pnpm`         | `devbox run pnpm` |
| `npm`          | `devbox run npm` |
| `node`         | `devbox run node` |
| `npx`          | `devbox run npx` |
| `openspec`     | `devbox run openspec` |

### Examples

```bash
# Artisan commands
devbox run php artisan migrate
devbox run php artisan make:model User

# Composer
devbox run composer install
devbox run composer require laravel/sanctum

# PNPM / NPM
devbox run pnpm install
devbox run pnpm run build
```

## Persistent Shell (Optional)

If the terminal session persists, you can enter the Devbox shell once:

```bash
devbox shell
```

While inside the shell, you do **not** need the `devbox run` prefix:

```bash
php artisan serve
pnpm run dev
```

## Devbox Scripts

Prefer running scripts defined in `devbox.json` using:

```bash
devbox run <script-name>
```

## Important

- **Do not assume** the environment is already loaded.
- **Always use** `devbox run` explicitly to ensure correct versions and extensions.
