# Action Control

A CLI tool to report and enforce GitHub Actions usage policies.

## Features

- **Report**: Generate reports of GitHub Actions usage across repositories
- **Enforce**: Enforce policies on which actions are allowed or denied in repositories
- **Export**: Generate policy files based on currently used actions
- Supports organization-wide scans or single repository analysis
- Multiple output formats (markdown, JSON)
- Repository-specific policy overrides
- Flexible policy modes: allow-list or deny-list

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

Create a `policy.yaml` file to define allowed or denied actions:

```yaml
# Choose policy mode: "allow" or "deny"
policy_mode: "allow"  # Default if omitted

# In "allow" mode: Only these actions are allowed
allowed_actions:
  - "actions/checkout"
  - "actions/setup-node"
  # Add more allowed actions...

# In "deny" mode: These actions are explicitly forbidden
denied_actions:
  - "unauthorized/action"
  - "security-risk/action"
  # Add more denied actions...

# Repositories excluded from policy enforcement
excluded_repos:
  - "your-org/sandbox-repo"

# Custom rules for specific repositories
custom_rules:
  "your-org/special-repo":
    policy_mode: "deny"  # Can override the global mode
    denied_actions:
      - "custom/special-action-to-deny"
```

### Policy Modes

Action Control supports two policy modes:

1. **Allow Mode (Default)**: Only actions explicitly listed in `allowed_actions` are permitted. All others are denied.
2. **Deny Mode**: All actions are permitted except those explicitly listed in `denied_actions`.

You can set the mode globally with `policy_mode`, or override it for specific repositories in `custom_rules`.

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
# Create an allow-mode policy from all actions used in an organization
action-control export --org your-organization

# Create a deny-mode policy
action-control export --org your-organization --policy-mode deny

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
- `--policy-mode`: Select the policy mode (allow or deny, default: allow)
- `--include-versions`: Include version tags in action references
- `--include-custom`: Generate repository-specific custom rules
- `--org`: Specify the organization to scan
- `--repo`: Specify a single repository to scan (format: owner/repo)

## Integrating with CI/CD

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

```yaml
name: GitHub Actions Policy Enforcement

on:
  schedule:
    - cron: '0 0 * * 1'  # Run every Monday
  workflow_dispatch:  # Allow manual trigger
  pull_request:
    paths:
      - '.github/workflows/**'  # Run when workflow files change

jobs:
  enforce:
    runs-on: ubuntu-latest
    steps:
      - name: Enforce GitHub Actions Policy
        uses: ihavespoons/action-control@main
        with:
          github_token: ${{ secrets.GITHUB_TOKEN }}
          policy_content: |
            policy_mode: allow
            allowed_actions:
              - actions/checkout@v3
              - actions/setup-node@v4
              - actions/cache@v3
            excluded_repos:
              - your-org/sandbox-repo
```

## GitHub Action Configuration

When using the GitHub Action, it automatically analyzes the current repository. The policy configuration is provided through the `policy_content` input parameter, which allows for dynamic policy configuration without needing to commit policy files to your repository.

**Note**: When running as a GitHub Action, local policy files are ignored, and only the policy content provided through the `policy_content` input is used. This ensures consistent enforcement across environments.

### Using Policy Content from Variables

You can also store your policy in GitHub variables or secrets:

```yaml
name: Enforce GitHub Actions Policy

on:
  schedule:
    - cron: '0 0 * * 1'  # Run every Monday
  workflow_dispatch:  # Allow manual trigger

jobs:
  enforce-policy:
    runs-on: ubuntu-latest
    steps:
      - name: Enforce GitHub Actions Policy
        uses: ihavespoons/action-control@main
        with:
          github_token: ${{ secrets.GITHUB_TOKEN }}
          policy_content: ${{ vars.ACTION_CONTROL_POLICY_CONTENT }}
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