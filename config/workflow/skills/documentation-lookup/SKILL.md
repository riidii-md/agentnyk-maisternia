---
name: documentation-lookup
description: Look up public library and tool documentation through local interfaces, native web search, and version-matched primary sources.
version: 0.1.0
---

# Documentation Lookup

Inspect local interfaces, lockfiles, and installed dependency versions before
searching. Use native web search to find public documentation, then read the
primary sources: original documentation, tagged source, declarations, release
notes, or maintainer issues. Prefer sources that match the installed version.
Search snippets and generated summaries are leads; read the original before
relying on a claim. Cite the URL and version or commit used. Distinguish an
unreleased fix or open issue from behavior in a released dependency.

Keep repository code and private documents in local or explicitly authorized
project search. Send only public queries to external services; do not include
private code, document text, credentials, or internal identifiers.

If native search is unavailable, use an already authorized search integration
or read known official URLs through an available fetch or browser tool. If a
page needs browser rendering, use the authorized browser. Report unavailable
capabilities or unreadable sources; do not imply verification succeeded. Ask
before connecting a service or accepting account costs when not authorized.

Exa and Tavily are optional experiments, not required dependencies or automatic
fallbacks. Prefer an alternative only when a controlled comparison shows equal
verified correctness, completeness, version accuracy, and source coverage with
lower total tokens, including source reads, synthesis, retries, and fallbacks.
Label missing usage measurements; response length alone does not prove savings.
