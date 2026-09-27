# mdmaid.desk Space Routing

Use this contract whenever a workflow registers or imports a document into
mdmaid.desk. A Space is a dynamic presentation filter derived from repository
identity and document tags. It is not a publication destination, filesystem
authorization boundary, attention state, or workflow approval.

mdmaid.desk is the sole source of truth for Space definitions and membership.
Do not persist a second matcher registry in Maisternia or provider settings.
Do not read or edit SQLite. Use only the public mdmaid.desk CLI or API.

## Resolve routing inputs

Follow `project-naming.md` first. Before Space routing, settle all facts that
will be sent to registration:

- the resolved workspace ID;
- its canonical credential-free repository key; and
- the exact normalized tags planned for this document.

Evaluate routing for every publication. Re-read current Space definitions and
repository inventory with `mdmaid-desk space list --json` and
`mdmaid-desk repository list --json`. Cache only stable discovery facts such as
the executable, daemon, resolved workspace, and repository key. Never cache a
classification answer across documents because tags and Space definitions can
change.

The persisted matcher kinds are `repository`, `repository-namespace`, and
`tag`. Match them with OR semantics:

- repository values match the canonical repository key exactly;
- a repository-namespace matches a repository key only when it begins with
  that value plus `/`, never equality or an arbitrary string prefix; and
- tag values match one of the exact normalized tags planned for the document.

The public `--workspace` option is input convenience: mdmaid.desk resolves the
workspace to its repository key and persists a repository matcher. It is not an exact-workspace matcher and may include sibling workspaces for the same
repository.

Membership may be zero, one, or many Spaces. When at least one current Space
matches, do not ask which Space to use. Retain every matched Space for the
post-delivery check.

## Ask once for unmatched content

When no current Space matches, ask the human to choose an existing Space or
name a new one and provide at least one matcher. Offer the stable scopes that
mdmaid.desk supports: this repository, a repository namespace, or a document
tag. Explain the repository-wide effect of the `--workspace` convenience.

An explicit `All for this session` response permits this publication without a
durable matcher. Cache that response only for the current session and exact
repository/tag input. mdmaid.desk has no durable negative or ignore matcher, so
do not invent one in Maisternia.

## Reconcile positive choices safely

Mutate Space state only after an explicit positive human choice.

1. Inspect before create or update.
2. Reuse an already-compatible definition.
3. When creating a missing Space, re-read after a create conflict and accept
   only a compatible resulting definition.
4. Reject empty matcher sets and normalize/deduplicate the proposed values.
5. Remember that `mdmaid-desk space matchers set` replaces the complete matcher
   collection. Preserve the full observed collection when proposing an update.

Full-set replacement requires established exclusive single-writer authority
for that Space during reconciliation. A read/re-read cannot detect a concurrent
write because mdmaid.desk currently has no compare-and-swap token. If exclusive
single-writer authority cannot be established, do not invoke
`space matchers set`; show the preserved, deduplicated full replacement for
manual execution or defer until mdmaid.desk provides an atomic add/CAS API.
Never overwrite a conflicting definition silently.

## Capability and failure dispositions

Package presence, package version, daemon compatibility, and Space capability
are separate facts. Use the published Spaces-capable minimum recorded by the
managed environment pack when one exists; never guess an unpublished version.
A healthy older daemon requires upgrade and restart guidance. Never bypass it
through direct catalog access.

Handle failures as follows:

- When Space commands or capabilities are absent but baseline registration is
  independently verified, passive delivery may proceed with
  `classification deferred`. Do not claim a Space. An exact-revision decision
  may continue only when its existing decision capability independently passes.
- For a healthy older daemon, do not mutate and do not fall back to SQLite.
  Permit baseline delivery only when the CLI/daemon pair independently proves
  it, and report the required upgrade/restart.
- For malformed Space JSON or a public-contract mismatch, stop routing,
  mutation, and desk delivery for this attempt. Preserve the artifact and
  report the exact retry.
- For schema incompatibility or native runtime/module failure, preserve the
  artifact and stop desk delivery. Report that distinct repair; do not open the
  catalog directly.
- For workspace or repository ambiguity, stop for human resolution before
  mutation or delivery.
- For a conflicting existing definition, preserve it and pause for explicit
  human resolution. An explicit session-only All choice permits baseline
  delivery; an exact-revision decision publication may also proceed only when
  its existing decision capability independently passes. All changes only
  Space routing: retain the existing `--expect` request and revision-bound wait.

## Verify delivery before a Space claim

Registration and import remain workspace operations; there is no `--space`
document field. After successful delivery, obtain the exact document ID from
the normal receipt. For each Space that will be named in the result, require
that exact document ID in the first field of `mdmaid-desk list --space <id>` or
use the equivalent authenticated API result.

If the scoped postcondition fails, the document remains delivered but its Space
classification is unproven. Retain the delivery receipt, report the failed
classification, and make no “published to Space” claim. An already-created
exact-revision decision request remains governed by its existing wait; Space
membership does not grant approval and does not invalidate a valid decision
transport.

Extend the existing publication receipt with the inspected Space IDs, normalized
matchers when a user-authorized mutation occurred, and the scoped result. This
is task evidence only and must never drive later classification.
