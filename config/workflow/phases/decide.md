# /work-decide - Record the Chosen Direction

Routing gate (lazy): load `work-routing` only when `$ARGUMENTS` has a plausible explicit route, an active session route exists, or the exact `.maisternia/work-routing.json` or `${XDG_CONFIG_HOME:-~/.config}/maisternia/work-routing.json` exists. Otherwise continue locally without loading it. After loading, continue only with its cleaned task.

Create a durable decision record that prevents plan and implementation drift.

Input:

`$ARGUMENTS`

Return:

- Decision title
- Chosen approach
- Why it was chosen
- Rejected options and why
- Constraints accepted
- Assumptions
- Risks accepted
- Open questions
- Ready-for-planning status

Do not treat an unconfirmed recommendation as an approved decision.
