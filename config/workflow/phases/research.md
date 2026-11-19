---
name: work-research
description: Research and compare solution directions using current primary evidence, explicit criteria, and tradeoffs.
version: 0.3.0
---

# /work-research - Compare Solution Directions

When this invocation authors or revises document prose, read the installed
`readable-output` skill's `references/document-wording.md` and apply it before finalization.
Before requesting a human response to an actual document, use that skill's
**Document-bound human checkpoints** contract. Ordinary clarification does not
require a document, structural adaptation, or publication.

Routing gate (lazy): load `work-routing` only when `$ARGUMENTS` has a plausible explicit route, an active session route exists, or the exact `.maisternia/work-routing.json` or `${XDG_CONFIG_HOME:-~/.config}/maisternia/work-routing.json` exists. Otherwise continue locally without loading it. After loading, continue only with its cleaned task.

Research viable solution families before detailed planning.

When `work-routing` resolves several harnesses, run independent research lanes,
verify factual conflicts against primary evidence, and let the current harness
synthesize the recommendation without treating consensus as proof.

Input:

`$ARGUMENTS`

Inspect existing repository patterns first. Use current primary documentation
when external behavior matters. Compare two to four real options when meaningful
choice exists.

For public documentation, inspect local interfaces and installed dependency versions
first. Use native web search, then read version-matched primary sources such as
official documentation, tagged source, declarations, release notes, and maintainer
issues. Treat search snippets as leads and cite the original evidence used.
Keep code and private documents in authorized local or project search; send only
public queries to external services. If native search is unavailable, use an
already authorized integration or fetch known official URLs. Use an authorized
browser for pages that need rendering, and report unavailable capabilities.
Do not silently install services or accept account costs. Exa and Tavily remain
optional experiments unless equal verified quality with lower total tokens is
demonstrated, including source reads, synthesis, retries, and fallback work.

Evaluate:

- scope and complexity;
- correctness and testability;
- security;
- migration and deployment impact;
- maintainability;
- user impact;
- reversibility.

Do not edit files.

Return options, recommendation, rejected directions, risks, user decisions,
architectural constraints and cross-system implications for `/work-direction`
when the gate applies, and remaining evidence or human constraints. In guided
standard work or idea shaping, return findings to the coordinator for
`/work-collaborate research` before direction, option comparison, or planning,
even when the agent has no questions. The coordinator presents the updated
discovery brief and offers discussion, more research, and an optional human
draft. When direction is not required, proceed to `/work-plan` only after the
human explicitly continues. This research phase remains read-only.
