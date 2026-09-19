# Project-scoped code and Markdown search

This guide activates GitNexus for code and QMD for local Markdown in one
repository. Context7 remains the hosted source for current public library
documentation. This setup does not send private repository documents to
Context7.

## Prepare the tools

Install Maisternia from this revision, then review and install the two pinned
environment packs:

```sh
maisternia environment plan developer-context
maisternia environment plan project-docs-qmd
maisternia environment install --yes developer-context
maisternia environment install --yes project-docs-qmd
gitnexus --version  # 1.6.12 or newer
qmd --version       # 2.8.3 or newer
```

GitNexus has a PolyForm Noncommercial license; check its terms before using it
for work. QMD's first semantic query can download local models. Full-text
search works without that step.

## Remove stale global registrations

A project server does not hide a global server. Inspect each provider's global
MCP configuration and remove any old `gitnexus` or `qmd` registration before
activating the project files. For Codex, inspect `~/.codex/config.toml` or use
`codex mcp list`; `codex mcp remove gitnexus` removes a registered global
server. For Claude Code, use `claude mcp list` and remove the old user-scope
registration with `claude mcp remove -s user gitnexus`. For Antigravity, inspect
`~/.gemini/config/mcp_config.json`, the older
`~/.gemini/antigravity/mcp_config.json`, and its MCP manager. Review each
change to these mixed settings files; Maisternia never edits them here.

The activation command checks these common global files, plus `CODEX_HOME`
when set, for old registrations and stops if it finds them. This is a
diagnostic, not a complete inventory of active MCP servers: Codex CLI
overrides and Claude plugin or managed scopes can also load servers. Inspect
the provider's active MCP list and remove any other unscoped GitNexus entry.
The scan may also stop on an unrelated mention of either name; review the
reported file and rerun the plan. Repeat the check if a provider later changes
its MCP settings.

## Activate one project

Run from the repository root. `--docs` names an existing Markdown directory
inside the repository and defaults to `docs`. Use `--docs .` when Markdown is
spread across the repository.

```sh
maisternia developer-context plan --project "$PWD" --target all --docs docs
maisternia developer-context apply --project "$PWD" --target all --docs docs --yes
qmd update
gitnexus analyze .
```

`apply` is opt-in and creates these local files for Codex, Claude Code, and
Antigravity: `.codex/config.toml`, `.mcp.json`, `.agents/mcp_config.json`, and
`.qmd/index.yml`. It also creates `.qmd/.gitignore` for the config and SQLite
index. Existing files cause a conflict; review and merge them manually. The
command never overwrites them. For one provider, select `--target codex`,
`claude`, or `antigravity`. `--target hermes` only displays the manual recipe.
The `developer-context` preset separately stages Context7 MCP fragments and
Claude permissions; review its plan if you use that hosted documentation layer.
This command does not change an existing Context7 registration.

The generated paths are machine-specific. Keep these files out of commits,
including the active provider configs. Check `git status` before committing.
The QMD index may contain copies of private document text. Refresh it with
`qmd update` from this repository whenever Markdown changes; `qmd embed` is
optional for semantic retrieval. The launcher checks the exact QMD collection
configuration before starting, so edited collections need manual review.

The GitNexus MCP process starts with this repository as its working directory,
`GITNEXUS_MCP_ALLOWED_REPOS` set to its canonical path, and
`GITNEXUS_MCP_READ_ONLY=1`. The QMD MCP process also starts here and reads
only the generated project index. These controls limit the tools exposed by
these servers; they do not sandbox the agent's other tools or prevent a
separately configured global MCP server from running.

Trust the project in Codex and approve the project MCP servers in Claude Code
when those providers prompt. After restarting each provider, inspect its
active MCP servers and run a
search for a symbol or Markdown heading that exists only in this repository.
Confirm that GitNexus does not list another repository and QMD results come
from this project's `docs` directory. Check the provider's own MCP logs if a
server does not start; the launcher rejects an older tool version.
Make sure `maisternia`, `gitnexus`, and `qmd` are on the provider process's
`PATH`, especially when launching a GUI app outside a terminal.

## Hermes: select a dedicated profile manually

Hermes does not automatically load a project MCP file. Create one profile per
repository and select it explicitly when starting Hermes. A blank profile
avoids copying an old global GitNexus registration:

```sh
hermes profile create my-project
hermes -p my-project config set terminal.cwd /absolute/path/to/project
```

Configure the profile's model and credentials using Hermes's normal setup.
Then add the following entries to
`~/.hermes/profiles/my-project/config.yaml`, replacing the example absolute
path and merging with any existing `mcp_servers` map:

```yaml
mcp_servers:
  gitnexus:
    command: maisternia
    args: [developer-context, serve, gitnexus, --project, /absolute/path/to/project]
    tools:
      include: [query, context, impact, trace, detect_changes, check, route_map, tool_map, shape_check, api_impact, explain, pdg_query]
      resources: false
      prompts: false
  qmd:
    command: maisternia
    args: [developer-context, serve, qmd, --project, /absolute/path/to/project, --docs, docs]
    tools:
      include: [query, get, multi_get, status]
      resources: false
      prompts: false
```

Launch with `hermes -p my-project chat`. Do not make this profile the sticky
default when switching among repositories. The `serve` wrapper enforces the
project root and GitNexus allowlist even if Hermes starts from another working
directory. Keep each profile's MCP list limited to the repository it serves.

Provider references: [Codex project configuration](https://developers.openai.com/codex/config-basic),
[Claude Code MCP scopes](https://code.claude.com/docs/en/mcp),
[Antigravity workspace MCP](https://antigravity.google/docs/mcp),
[QMD project indexes](https://github.com/tobi/qmd/blob/main/README.md), and
[Hermes profiles](https://github.com/NousResearch/hermes-agent/blob/main/website/docs/user-guide/profiles.md).
