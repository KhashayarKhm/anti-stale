# Anti-stale

A CLI tool to automatically find and revive GitHub issues/pull requests that have been marked as stale by stale bots, preventing them from being automatically closed.

## What it does

Anti-stale checks GitHub repositories for issues and pull requests tagged with stale labels and can automatically comment on them to remove the stale status, keeping important issues active. It traverses a configuration of owners and their repositories to check specified issues and PRs.

## Why Stale Bots Are Harmful

Stale bots can prematurely close important issues, stifling collaboration and losing valuable context. Here’s why they’re considered problematic:
- [GitHub Stale Bots - Blog by Ben Winding](https://blog.benwinding.com/github-stale-bots/)
- [Reddit: GitHub Stale Bots: A False Economy](https://www.reddit.com/r/programming/comments/kzvryq/github_stale_bots_a_false_economy/)
- [Reddit: GitHub Stale Bot Considered Harmful](https://www.reddit.com/r/programming/comments/sh6a1t/github_stale_bot_considered_harmful/)
- [Hacker News: Discussion on Stale Bots](https://news.ycombinator.com/item?id=28998374)

## Installation

```bash
# Install from source
go install github.com/yourusername/anti-stale@latest

# Or clone and build
git clone https://github.com/yourusername/anti-stale.git
cd anti-stale
go build -o anti-stale .
```

## Prerequisites

- Go 1.20 or higher
- GitHub Personal Access Token with `repo` permissions
- Access to the repositories you want to check

## Configuration

Create a configuration file (default: `anti-stale.json`):

```json
{
  "token": "your_github_token_here",
  "userAgent": "your_github_username",
  "owners": {
    "owner-of-the-project": {
      "project": {
        "issues": [1, 2],
        "prs": []
      },
      "project2": {
        "issues": [1]
      }
    }
  }
}
```

### Configuration Fields
- `token`: Your GitHub Personal Access Token (required).
- `userAgent`: A string identifying your client (e.g., your GitHub username, required).
- `owners`: A nested object mapping owners to repositories and their issues/PRs:
  - Keys are GitHub usernames/organization names.
  - Values are objects mapping repository names to:
    - `issues`: Array of issue numbers to check.
    - `prs`: Array of pull request numbers to check.

### GitHub Token Setup
1. Go to GitHub Settings → Developer settings → Personal access tokens.
2. Generate a new token with `repo` scope.
3. Add the token to `anti-stale.json`

## Usage

### Basic Commands

```bash
# Check for stale issues (dry run)
anti-stale check

# Check and automatically comment on stale issues
anti-stale check --reply

# Interactive mode - decide for each issue
anti-stale check --interactive

# Use custom configuration file
anti-stale --config /path/to/config.json check

# Use custom stale label
anti-stale check --label "needs-attention"
```

### Examples

```bash
# Find all stale issues without taking action
anti-stale check

# Automatically comment on all stale issues
anti-stale check --reply --msg "This issue is still important"

# Interactive mode with custom message
anti-stale check -i --msg "Still working on this, please keep open"

# Debug mode with verbose logging
anti-stale --log-level 0 check --reply
```

## Global Options

| Flag | Description | Default |
|------|-------------|---------|
| `--config`, `-c` | Path to configuration file | `./config.json` |
| `--log-level` | Logging level (0=Debug, 1=Info, 2=Warn, 3=Error) | `1` |

## Check Command Options

| Flag | Description | Default |
|------|-------------|---------|
| `--reply` | Automatically comment on stale issues/PRs | `false` |
| `--interactive`, `-i` | Prompt for confirmation on each issue/PR | `false` |
| `--msg` | Custom message to post as comment | `"This issue is still relevant"` |
| `--label`, `-l` | Name of the stale label to look for | `"stale"` |

## How it Works

1. **Discovery**: Traverses the `owners` tree in `config.json` to identify repositories and their issues/PRs.
2. **Filtering**: Checks for open issues/PRs with stale labels using GitHub’s GraphQL API.
3. **Action**: Reports findings or comments to remove stale status.
4. **Rate Limiting**: Respects GitHub API limits with configurable delays.

## Best Practices

- **Test first**: Run without `--reply` to preview affected issues/PRs.
- **Use meaningful messages**: Customize comments to explain why issues should remain open.
- **Monitor rate limits**: Use appropriate delays to avoid `429` errors.
- **Be selective**: Use interactive mode (`--interactive`) for sensitive repositories.

## Contributing

1. Fork the repository.
2. Create a feature branch (`git checkout -b feat/amazing-feature`).
3. Commit your changes (`git commit -m 'feat: Add amazing feature'`).
4. Push to the branch (`git push origin feat/amazing-feature`).
5. Open a Pull Request.

**Make sure your pull request respects [branch](https://conventional-branch.github.io/) and [commit message](https://www.conventionalcommits.org/en/v1.0.0) conventions.**

## Roadmap

- [ ] Use string builder for GraphQL query construction.
- [ ] Custom messages per issue in interactive mode.
- [ ] Display last update time and stale countdown.
- [ ] Auto-reopen closed issues when appropriate.
- [ ] Support for multiple stale labels.
- [ ] Webhook integration for real-time monitoring.
- [ ] Dashboard for monitoring stale issue trends.
- [ ] Slack/Discord notifications.
