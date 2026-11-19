# Public documentation search

Maisternia defaults to native web search followed by primary-source reading.
Inspect local interfaces and installed dependency versions first. Prefer the
original documentation, tagged source, declarations, release notes, or maintainer
issues for that version. Cite the evidence used. Search snippets and generated
summaries do not establish correctness by themselves.

The `developer-context` preset installs this policy as `documentation-lookup`
for Codex and Claude Code. The `work-research` workflow carries it for Codex,
Claude Code, Antigravity, and Hermes. GitNexus and QMD serve separate local code
and Markdown search needs. Do not send private code, document text, credentials,
or internal identifiers in public search queries.

When a harness has no native search, use an already authorized integration or
read known official URLs with an available fetch tool. Use an authorized browser
if the page requires rendering. Report missing capabilities and unreadable
sources. Connecting a service or accepting account costs requires authority for
that action; installing this policy does not grant it.

## Existing installations

Context7 has been removed from the catalog. The developer-context MCP fragments
now contain only GitNexus, and the Claude permission source contains only exact
GitNexus tool names. The old Codex Resource Lab MCP resource has been deleted.
No preset installs Exa or Tavily in its place.

With the updated Maisternia binary or an explicit `--repo` pointing at this
revision, preview each installed preset:

```sh
maisternia preset plan --scope user --target all developer-context
maisternia preset plan --scope user --target codex codex-resource-lab
```

After reviewing conflicts, use the matching `preset apply` command with `--yes`.
An upgrade updates the owned developer-context fragments. The lab upgrade
removes its previously owned `mcp-context7.toml` and backs it up. Locally edited
or symlinked targets still stop the operation. Use the original installation
scope; project installations require `--scope project --project <directory>`.
Render into a fresh output directory when inspecting the new catalog: rendering
does not uninstall files from an earlier output directory or provider home.

Two boundaries require separate review:

- Claude permissions use additive JSON-array union. Applying the new source
  does **not** subtract `mcp__context7__resolve-library-id` or
  `mcp__context7__query-docs` from existing mixed settings. Back up that file and
  remove only those exact values after reviewing their ownership. Preserve all
  other permissions and settings. Uninstalling the preset also preserves merged
  values; it is not a cleanup shortcut.
- The review fragments never automatically registered an active server in a
  mixed provider configuration. Inspect the provider's active MCP list and each
  applicable scope. Use the provider's supported removal path for a confirmed
  retired registration. A Context7 name alone does not prove Maisternia owns it.
  Review separately installed documentation skills too. The installer reports
  a conflict instead of replacing an independently installed skill silently.

The initial read-only inspection for this task found three owned fragments: the Codex and
Claude developer-context fragments and the Codex lab fragment. Both
developer-context fragments match their recorded installed checksums, and the
lab upgrade plans a removal. The two obsolete permission values are present in
Claude's mixed settings. The developer-context upgrade also reports independently
installed `documentation-lookup` skills in both Codex and Claude as conflicts.
Active Context7 entries were also present in Codex, Claude user settings, and
Hermes; their registration provenance is not tracked by Maisternia. This records
the starting state. Applying the catalog does not remove those independently
registered servers; their cleanup requires the separately authorized provider
operations described above.

## Exa and Tavily trials

The [trial report](experiments/documentation-search/README.md) records six fixed
public questions, keyless MCP setup, observed retrieval failures, timings, and
measurement limits. Both integrations connected without subscriptions or API
tokens. Free-tier caps prevented a complete controlled comparison. Requests for
payment, account tokens, or bonus-credit submissions were not followed.

Native search remains the default. Neither alternative demonstrated an
end-to-end token benefit in this run. Equal verified correctness, completeness,
version accuracy, and source coverage with lower **total** token use would be a
successful improvement. Total usage must include tool definitions where
measurable, all search and source reads, synthesis, retries, and fallbacks.
A shorter response that omits required facts is not an efficiency improvement.

Trial endpoints receive public queries and requested URLs. They add external
availability, quota, account-cost, and maintenance dependencies. Remote server
versions can change without a local package update. Do not add a permanent
fallback until a metered comparison demonstrates its value for that harness and
task type. Roll back an isolated trial by ending its session and discarding its
explicit configuration; no global registration needs removal.
