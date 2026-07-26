# /codex-plan - Run the Plan Phase with Codex

This is a permanent explicit-runner alias for `/work plan`.

Claude must invoke Codex in read-only mode using the configured reasoning model.
Pass the accepted definition, decision, proof inputs, repository context,
conversation handoff, and `$ARGUMENTS`.

Codex must follow the canonical plan phase and discover repository rules before
assuming conventions. Claude summarizes the plan and unresolved decisions.
