# Action Control

A CLI tool to report and enforce GitHub Actions usage policies.

## Features

- **Report**: Generate reports of GitHub Actions usage across repositories
- **Enforce**: Enforce policies on which actions are allowed in repositories
- **Export**: Generate policy files based on currently used actions
- Supports organization-wide scans or single repository analysis
- Multiple output formats (markdown, JSON)
- Repository-specific policy overrides

## Installation

```bash
go install github.com/ihavespoons/action-control@latest
```

## Configuration

Create a `config.yaml` file in one of these locations (searched in this order):
1. Current directory
2. `$HOME/.config/action-control/`
3. Home directory (`$HOME`)

```yaml
github_token: "your-github-token"
organization: "your-org"
output_format: "markdown"
```

You can also set environment variables with the `ACTION_CONTROL_` prefix:
```bash
export ACTION_CONTROL_GITHUB_TOKEN="your-token"
export ACTION_CONTROL_ORGANIZATION="your-org"
```

## Policy Configuration

Create a `policy.yaml` file to define allowed actions:

```yaml
# Default allowed actions for all repositories
allowed_actions:
  - "actions/checkout"
  - "actions/setup-node"
  # Add more allowed actions...

# Repositories excluded from policy enforcement
excluded_repos:
  - "your-org/sandbox-repo"

# Custom rules for specific repositories
custom_rules:
  "your-org/special-repo":
    allowed_actions:
      - "actions/checkout"
      - "custom/special-action"
```

## Repository-specific Policy

Repositories can include their own policy file at `.github/action-control-policy.yaml`. This will be merged with the global policy.

## Usage

### Generating Reports

```bash
# Report on all repositories in an organization
action-control report --org your-organization

# Report on a specific repository
action-control report --repo owner/repo-name

# Output in JSON format
action-control report --org your-organization --output json
```

### Enforcing Policy

```bash
# Enforce policy across an organization
action-control enforce --org your-organization --policy path/to/policy.yaml

# Enforce policy on a specific repository
action-control enforce --repo owner/repo-name --policy path/to/policy.yaml
```

The command will exit with an error code if any violations are found.

### Exporting Policy

Generate a policy file based on currently used actions:

```bash
# Create a policy from all actions used in an organization
action-control export --org your-organization

# Create a policy with version information included
action-control export --org your-organization --include-versions

# Create a policy with repository-specific rules
action-control export --org your-organization --include-custom

# Export policy to a specific file
action-control export --org your-organization --file custom-policy.yaml

# Export policy based on a specific repository
action-control export --repo owner/repo-name
```

## Export Options

The export command supports the following options:

- `--file`: Specify the output file path (default: policy.yaml)
- `--include-versions`: Include version tags in action references
- `--include-custom`: Generate repository-specific custom rules
- `--org`: Specify the organization to scan
- `--repo`: Specify a single repository to scan (format: owner/repo)

## Integrating with CI/CD

You can integrate action-control into your CI/CD pipelines to enforce policies automatically:

```yaml
name: Enforce Action Policies

on:
  schedule:
    - cron: '0 0 * * *'  # Run daily
  workflow_dispatch:     # Allow manual trigger

jobs:
  enforce:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
      - name: Install action-control
        run: go install github.com/ihavespoons/action-control@latest
      - name: Check policy compliance
        run: action-control enforce --org your-organization
        env:
          ACTION_CONTROL_GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

## Use Cases

### Bootstrap a New Policy

```bash
# 1. First, export the current state to create a baseline
action-control export --org your-organization --file initial-policy.yaml

# 2. Edit the policy file to remove unwanted actions

# 3. Enforce the new policy
action-control enforce --org your-organization --policy initial-policy.yaml
```

### Periodic Compliance Checks

Run regularly in CI to ensure all repositories stay compliant with your action policies:

```bash
action-control enforce --org your-organization
```

### Repository-Specific Exceptions

1. Create a base organization policy
2. Allow repositories to define specific overrides in their `.github/action-control-policy.yaml` files
3. Run enforcement with the base policy; repository-specific policies will be automatically merged

## Using as a GitHub Action

You can use Action Control directly in your GitHub workflows:

### Example: Enforce policy

```yaml
name: GitHub Actions Policy Enforcement

on:
  schedule:
    - cron: '0 0 * * *'  # Daily
  pull_request:
    paths:
      - '.github/workflows/**'  # Run when workflow files change

jobs:
  enforce:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v3
        
      - name: Enforce Actions Policy
        uses: ihavespoons/action-control@main
        with:
          github_token: ${{ secrets.GITHUB_TOKEN }}
          organization: your-organization
          policy_file: .github/policies/actions-policy.yaml
```

## Development

### Testing

Run the test suite:

```bash
make test
```

Or run specific test components:

```bash
make test-unit      # Run unit tests only
make test-integration # Run integration tests
make coverage       # Generate test coverage report
```

### Building

Build the binary:

```bash
make build
```

The binary will be created in the `bin/` directory.