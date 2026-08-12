# /work-adapt-for-reader - Adapt Text to Its Reader

Transform supplied text, referenced files, or current conversation context for
the reader and use described in:

`$ARGUMENTS`

Use the installed `adapt-for-reader` skill and its mode, preference, and design
principle references. Preserve meaning, evidence, uncertainty, constraints, and
source provenance.

Resolve preferences by the skill's precedence rules; the current request wins.
If the reader or intended use is missing and plausible choices
would materially change the output, ask one focused question:

> Who will use this text, and what should they be able to do after reading it?

Do not ask when active instructions, a matching situation preference, or the
request already resolves the reader contract. For low-stakes reversible output,
infer a reasonable mode and state the assumption only when useful.

Select `scan`, `decide`, `learn`, `operate`, `reference`, or `narrative`. Apply
plain-language, accessibility, density, evidence, and visual preferences as
modifiers. Use tables or diagrams only when they materially reduce comparison
or relationship-building work.

Return the adapted text or artifact first. Do not edit a source file unless the
user requested an edit. For long output, preserve the complete Markdown and use
the configured readable-output renderer when available.
