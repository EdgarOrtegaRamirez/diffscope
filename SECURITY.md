# Security Policy

## Supported Versions
| Version | Supported          |
| ------- | ------------------ |
| 1.x     | :white_check_mark: |

## Reporting a Vulnerability
Please report security vulnerabilities to [GitHub Security Advisories](https://github.com/EdgarOrtegaRamirez/diffscope/security/advisories/new).

## Security Design
- DiffScope only reads local git data — no network access
- Configuration files are validated against a strict schema before use
- No dynamic code evaluation or execution of parsed code
- All file paths are validated and resolved relative to the working directory
