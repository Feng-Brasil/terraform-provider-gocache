# Contributing Guidelines

Thank you for your interest in contributing to our project. Whether it's a bug report, new feature, correction, or additional
documentation, we greatly value feedback and contributions from our community.

Please read through this document before submitting any issues or merge requests to ensure we have all the necessary
information to effectively respond to your bug report or contribution.

## Requirements

To use this repository, you'll want to make sure you have the tools listed in [mise.toml](./mise.toml):

To simplify the process of installing these tools, you can install [mise](https://mise.jdx.dev/), then run the following to concurrently install all the tools you need, pinned to the versions they were tested with (as tracked in the [mise.toml](./mise.toml) file):

```bash
mise install
```

**Strongly recommended:** [activate mise](https://mise.jdx.dev/cli/activate.html) in your shell so that tool versions are managed automatically when you enter the project directory. Add one of the following to your shell's rc file:

```bash
# bash (~/.bashrc)
echo "eval \"$(mise activate bash)\""

# zsh (~/.zshrc)
eval "$(mise activate zsh)"

# fish (~/.config/fish/config.fish)
mise activate fish | source
```

### Recommended extensions

To improve your development experience, we recommend installing these extensions:

- YAML (`redhat.vscode-yaml`) - install from Cursor/Vscode Extensions panel by searching the extension ID, or via [Open VSX page](https://marketplace.cursorapi.com/items/?itemName=redhat.vscode-yaml)
- HashiCorp Terraform (`hashicorp.terraform`) - install from Cursor/Vscode Extensions panel by searching the extension ID, or via [Open VSX page](https://marketplace.cursorapi.com/items/?itemName=hashicorp.terraform)

## Local Development Setup

Run these steps before making changes to the code. They will help automate the commands to lint the code or run tests. These steps help to ensure high code quality and reduce the likelihood that the changes inadvertently break something.

Before making infrastructure changes, take time to read the project documentation in [docs/](./docs/) to understand the architecture and conventions. Also read [AGENTS.md](./AGENTS.md), which documents important repository patterns and standards.

### Install pre-commit hooks

> This will ensure that the commands we want to execute before each commit are executed automatically.

```shell
pre-commit install
```

### Execute pre-commit hooks manually on all files

```shell
pre-commit run --all-files
```

### Update pre-commit hooks

```shell
pre-commit autoupdate
```
