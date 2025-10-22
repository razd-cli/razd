# Security Policy

## Supported Versions

We support the latest released version of razd with security updates.

| Version | Supported          |
| ------- | ------------------ |
| Latest  | :white_check_mark: |
| < Latest| :x:                |

## Reporting a Vulnerability

If you discover a security vulnerability within razd, please send an e-mail to the maintainers. All security vulnerabilities will be promptly addressed.

**Please do not report security vulnerabilities through public GitHub issues.**

### What to include

Please include the following information in your report:

- Type of issue (e.g. buffer overflow, SQL injection, cross-site scripting, etc.)
- Full paths of source file(s) related to the manifestation of the issue
- The location of the affected source code (tag/branch/commit or direct URL)
- Any special configuration required to reproduce the issue
- Step-by-step instructions to reproduce the issue
- Proof-of-concept or exploit code (if possible)
- Impact of the issue, including how an attacker might exploit the issue

### Response Process

1. **Acknowledgment**: We will acknowledge receipt of your vulnerability report within 48 hours.
2. **Assessment**: We will assess the vulnerability and determine its impact and severity.
3. **Fix Development**: We will develop a fix for confirmed vulnerabilities.
4. **Release**: We will release a security update and publicly disclose the vulnerability.

### Security Update Distribution

Security updates will be distributed through:

- GitHub Releases with security advisory
- Release notes highlighting security fixes
- Dependency scanning alerts (if applicable)

## Security Measures

razd implements the following security measures:

- **Dependency Scanning**: Automated vulnerability scanning of dependencies via cargo-audit
- **Code Analysis**: Static analysis via clippy with security-focused lints
- **Secure Builds**: Release binaries built with security hardening flags
- **Checksum Verification**: SHA256 checksums provided for all release binaries
- **Minimal Dependencies**: Limited dependency tree to reduce attack surface