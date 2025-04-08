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

## GitHub Action Configuration

When using the GitHub Action, the policy configuration is provided through the `policy_content` input parameter. This allows for dynamic policy configuration without needing to commit policy files to your repository.

**Note**: When running as a GitHub Action, local policy files are ignored, and only the policy content provided through the `policy_content` input is used. This ensures consistent enforcement across environments.

### Example: Using Policy Content in a GitHub Workflow

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
          organization: your-organization
          policy_content: |
            policy_mode: allow
            allowed_actions:
              - actions/checkout@v3
              - actions/setup-node@v4
              - actions/cache@v3
            excluded_repos:
              - your-org/sandbox-repo
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

## 5. Example of a Complete Workflow File

```yaml
name: Enforce GitHub Actions Policy

on:
  schedule:
    - cron: '0 0 * * 1'  # Run every Monday
  workflow_dispatch:  # Allow manual trigger
  push:
    paths:
      - '.github/workflows/**'  # Run when workflows change

jobs:
  enforce-policy:
    runs-on: ubuntu-latest
    steps:
      - name: Enforce GitHub Actions Policy
        uses: ihavespoons/action-control@main
        with:
          github_token: ${{ secrets.GITHUB_TOKEN }}
          organization: your-organization
          policy_content: ${{ vars.ACTION_CONTROL_POLICY_CONTENT }}
```

## Testing the Feature

You can test this feature by:

1. Setting the `ACTION_CONTROL_POLICY_CONTENT` environment variable
2. Running the CLI with the `--ignore-local-policy` flag
3. Verifying that it uses the policy content from the environment variable instead of any local policy files

This implementation allows the GitHub Action to exclusively use the policy provided through the environment variable, completely ignoring any local policy files, while still allowing the normal CLI usage to work with local policy files.