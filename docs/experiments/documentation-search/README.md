# Documentation search retrieval pilot — 2026-11-19

Keep native web search plus primary-source reading as the default. Exa and
Tavily connected through free keyless MCP, but neither established lower total
token use at equivalent verified quality. The user's constraint was to skip
access that needs a subscription or API token. Quota failures were therefore
recorded without signing up, paying, or submitting data for bonus credits.

This is a retrieval pilot. It is **not** a completed controlled answer-quality
or end-to-end token benchmark. All token totals in [results.json](results.json)
are `null`, meaning unavailable. No savings percentage is inferred from text
length, payload bytes, request latency, or an incomplete answer.

## Inputs and setup

[cases.json](cases.json) fixes six questions, version targets, search queries,
required facts, and primary URLs. Cases cover a Go declaration, a TypeScript
migration, an obscure CLI, a Node API, Go release notes, and Excalidraw types.
Bubble Tea, Go, and mdmaid versions come from this repository's pins. Node
24.18.0 was available locally. Excalidraw 0.18.0 is an explicit evaluation target.
The Zod question targets the major-version migration, not an exact patch.

| Lane | Integration used | Version / scope |
|---|---|---|
| Native | Codex `web.run` search and open | Current connected tool; service version unavailable |
| Exa | Hosted Streamable HTTP MCP | `exa-search-server` 3.5.1 |
| Tavily | Hosted Streamable HTTP MCP, keyless header | `tavily-mcp` 4.1.5 |

[mcp.json](mcp.json) is an inert experiment configuration. The endpoints were
initialized and their tool schemas inspected with a temporary HTTP MCP client
using protocol `2025-03-26`. No package or persistent agent configuration was
installed. This exercises the supported remote integration without registering
it in the current agent's active tool list. Remote versions are observations,
not immutable package pins. No API key, subscription, or account was used.

Each lane received the identical query string and primary URL. Each case made
one search and one source read, including a read when the search failed. The
primary URL was fixed in advance, so this tests assisted retrieval rather than
unassisted source discovery. The requests were:

- Native: `search_query: [{q: query}]`, then `open: [{ref_id: primary_url}]`,
  with `response_length: "short"` for each. Result count is controlled by the
  service, not exposed as a matching five-result limit.
- Exa: `web_search_exa` with `query`, `numResults: 5`, and its required
  `objective` set to `Find primary-source evidence for: <question> Target
  version: <version>`. Then `web_fetch_exa` with one URL and
  `maxCharacters: 12000`.
- Tavily: `tavily_search` with `query`, `max_results: 5`,
  `search_depth: "basic"`, `include_raw_content: false`, and
  `include_images: false`. Then `tavily_extract` with the URL,
  `extract_depth: "basic"`, `query: question`, and `include_usage: true`.

The source-read controls differ because the supported schemas differ. Exa's
objective is also extra task context that the other search interfaces do not
require. These are material limits on comparability. Native calls were made
sequentially in one coordinator session. Each remote lane used its own MCP
session, but no lane used a fresh model context. Exact model identifier,
reasoning setting, tokenizer, and per-lane model usage were unavailable.

## Observed retrieval results

Times below are one observed search plus one source read in milliseconds,
including failures. They exclude setup, human approval waits, synthesis, audit
follow-ups, and preliminary integration research. They are not end-to-end
latencies or statistical estimates. All remote requests returned HTTP 200;
application-level failures still count as failures.

| Case | Native outcome / ms | Exa outcome / ms | Tavily outcome / ms |
|---|---|---|---|
| Bubble Tea Key | Content / 3,015 | Content / 891 | Both calls capped / 1,151 |
| Zod migration | Content / 3,045 | Search capped; read content / 696 | Search content; read capped / 1,892 |
| mdmaid 0.1.17 | Search content; read cache miss / 3,360 | Read reports 0.1.18, not requested 0.1.17 / 3,140 | Both calls capped / 1,053 |
| Node abort | Content / 3,516 | Read `CRAWL_TIMEOUT` / 8,405 | Both calls capped / 1,017 |
| Go synctest | Content / 3,054 | Content / 852 | Both calls capped / 1,046 |
| Excalidraw | Current docs; pinned types still needed / 3,691 | Current docs; pinned types still needed / 7,909 | Search capped; current-doc read content / 1,117 |

Each lane made 12 measured retrieval calls. Their summed retrieval times were
19,681 ms native, 21,893 ms Exa, and 7,276 ms Tavily. Tavily returned quota
messages for ten calls, so its lower time is not a successful speed comparison.
Exa returned a free-limit message once and a crawl timeout once. Native had one
cache-miss read failure. No quota recovery, paid access, or timeout retry was
attempted after inspecting the results.

Tavily quota responses included `monthly_cap_reached_bonus_eligible` despite
`isError: false`. A correct consumer must inspect returned content rather than
treating HTTP 200 or that flag as success. Successful Tavily calls were the Zod
search and Excalidraw extraction. The latter reported `usage.credits: 0`.

The final recorded initialization and tool-list round trips took 401 + 296 ms
for Exa and 569 + 575 ms for Tavily. The case-run initialization added 277 and
566 ms respectively. Earlier connectivity diagnostics included sandbox DNS
failures and an Exa Python HTTP-client failure; a curl transport connected
successfully. These diagnostics and initial smoke handshakes are outside the
12-call figures and are another reason those figures are not end-to-end usage.

## Primary-source correctness checks

These checks establish reference facts. They were shared audit work after
retrieval, not independent fresh-context answers or lane scores. Their tool
calls and local checks are not included in the latency table or a token total.

- [Bubble Tea v1.3.10 source](https://github.com/charmbracelet/bubbletea/blob/v1.3.10/key.go)
  and local `go doc github.com/charmbracelet/bubbletea Key` confirm `Type`,
  `Runes`, `Alt`, and `Paste`, with types `KeyType`, `[]rune`, `bool`, and `bool`.
  Native and Exa retrieved this version's complete field list. A v2 declaration
  would answer a different question.
- [Zod's migration guide](https://zod.dev/v4/changelog#defaults-applied-within-optional-fields)
  confirms that the example omits the key in Zod 3 and supplies the default in
  Zod 4. A shorter response that omits the key-existence impact is incomplete.
  The guide is mutable; a future patch-specific trial must pin source versions.
- The installed `mdmaid --version` reports 0.1.17. Its `--help` confirms the
  validation command, exits 0/1/2 for valid/invalid/runtime-unavailable, and the
  optional Puppeteer requirement for `--render`. The versioned npm page was
  unavailable through native fetch, while Exa returned metadata for 0.1.18.
  Current [upstream source](https://github.com/riidii-md/mdmaid) alone therefore
  did not establish the pinned version. The local CLI check resolved that gap;
  validation-runtime failure was not induced in this pilot.
- [Node v24.18.0 docs](https://nodejs.org/download/release/v24.18.0/docs/api/globals.html#static-method-abortsignalanysignals)
  list introduction in v20.3.0/v18.17.0. A local `node:test` check confirms that
  the first aborted input supplies the composed signal's reason and later
  aborts do not replace it. An initial one-line assertion failed; diagnostic
  output and the structured test then passed. That discrepancy was not resolved
  into a runtime finding. Exa search also returned nightly docs, which need
  version review before use.
- [Go 1.25 release notes](https://go.dev/doc/go1.25#testing/synctest), local
  `go doc testing/synctest.Test`, and the
  [maintainer migration issue](https://github.com/golang/go/issues/74012)
  distinguish the generally available `Test` API from experimental `Run`.
  Source evidence about an old experiment must not be presented as the current
  toolchain requirement.
- [Excalidraw v0.18.0 declarations](https://github.com/excalidraw/excalidraw/blob/v0.18.0/packages/excalidraw/types.ts)
  permit an object, a promise, or an initializer function. The
  [imported-data declaration](https://github.com/excalidraw/excalidraw/blob/v0.18.0/packages/excalidraw/data/types.ts)
  covers scene elements, app state, files, scrolling, and library data. The
  public initialData docs are useful but insufficient to establish the pinned
  type contract by themselves. No TypeScript compile check was run.

## Tokens, costs, and limits

| Measurement | Native | Exa | Tavily |
|---|---|---|---|
| Retrieved-context tokens | Unavailable | Unavailable | Unavailable |
| Tool-definition tokens | Unavailable | Unavailable | Unavailable |
| Total tokens including synthesis/retries/fallback | Unavailable | Unavailable | Unavailable |
| Account billing ledger | Unavailable | No account used | No account used |
| Measured tool-list transport bytes | Unavailable | 2,488 | 12,889 |

Payload bytes in the JSON record describe different transports and wrappers.
They are not tokenizer counts or billed usage. Tavily listed five tools even
though only Search and Extract were exercised. This pilot did not expose either
schema to a separate model lane, so schema overhead was not measured in context.

Official [Exa MCP documentation](https://exa.ai/docs/get-started/exa-mcp)
permits free rate-limited search and fetch. Its
[pricing page](https://exa.ai/pricing) lists starter credit allowances and paid
search/content rates; an authenticated account changes the cost and quota
conditions. [Tavily keyless documentation](https://docs.tavily.com/documentation/keyless)
describes free capped Search and Extract. Its
[credit pricing](https://docs.tavily.com/documentation/api-credits) lists 1,000
monthly free account credits, basic search at one credit, and paid usage at
$0.008 per credit. No account allowance was claimed, and no charge was inferred
from response size. Account-specific quota and billing totals remain unknown.

Both services receive queries and requested URLs. Browser rendering, source
availability, remote-server changes, and keyless quotas add operational work.
Native search also depends on the harness's own access and limits. Discard the
isolated MCP configuration to roll back; no global registration was created.

## Decision and remaining proof

Exa sometimes returned useful primary-source content, including the complete
Bubble Tea declaration and Go migration evidence. Tavily's free allowance did
not support a complete lane in this run. Neither result establishes a reason
to replace native search or install a permanent fallback in any harness.

A promotion trial must use the same exact model, reasoning settings, tool
visibility, versions, case rubric, and comparable fresh contexts, with repeated
and counterbalanced runs. Record actual model usage for prompts, available tool
definitions, every retrieved response, synthesis, retries, source verification,
and fallback work. Preserve cached-token accounting separately. Record latency,
calls, failures, quota, and money separately from tokens. Require all facts and
version-matched primary evidence before comparing cost. Equal verified quality
with lower total tokens is a win; missing required facts invalidate that claim.

The full metered comparison remains incomplete under this session's access and
measurement constraints. Do not describe this pilot as an efficiency benchmark
that Exa, Tavily, or native search won.

## Catalog verification

The repository change passed formatting, vet, unit tests, race detection,
coverage generation, build, catalog doctor, and rendering for all targets.
The new regression check rejects retired server resources and instructions in
rendered output and checks the source-reading policy across all four providers.
The generated output contains no Context7 reference.

The first full test run inherited `CODEX_HOME` and incorrectly inspected live
settings despite the tests' temporary home directories. Coverage also encountered
a Go 1.25.1/1.25.8 tool mismatch. Successful unit, race, and coverage runs used
`env -u CODEX_HOME GOTOOLCHAIN=go1.25.8` for the child process. No production
path, preflight, symlink, conflict, or confirmation check was weakened.

The initial live upgrade plan updates the two managed developer-context fragments and
removes the lab fragment. It also reports conflicts for two independently
installed documentation skills. Existing mixed Claude permissions need their
separate, exact-value cleanup. These observations record the pre-upgrade plan;
they do not imply that applying the catalog removes independently registered
servers or subtracts permissions from mixed settings.
