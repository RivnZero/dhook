# Security Policy

## Supported Versions

| Version | Supported          |
|---------|--------------------|
| 1.x     | Yes                |
| < 1.0   | No                 |

## Reporting a Vulnerability

If you discover a security vulnerability in dhook, please report it responsibly.

**Do NOT open a public GitHub issue for security vulnerabilities.**

### How to Report

Send an email to the maintainer with:

- A description of the vulnerability
- Steps to reproduce the issue
- The potential impact
- A suggested fix (if you have one)

### What to Expect

- **Acknowledgment**: Within 48 hours of your report
- **Assessment**: Within one week, you will receive an initial assessment
- **Resolution**: Timeline depends on severity; critical issues are prioritized

### Safe Harbor

We support responsible disclosure and will not take legal action against security researchers who:

- Make a good faith effort to avoid privacy violations
- Only interact with accounts they own or have explicit permission to test
- Do not exploit a vulnerability beyond what is necessary to confirm its existence
- Report vulnerabilities promptly

## Scope

This security policy applies to:

- The dhook library source code
- Official example code in this repository
- The build and CI pipeline

## Best Practices for Users

- Always use environment variables or secure secret management for webhook URLs
- Never commit webhook URLs to version control
- Use the principle of least privilege when configuring webhooks
- Monitor your webhook usage for unexpected activity
