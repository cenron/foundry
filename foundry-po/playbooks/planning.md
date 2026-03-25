# Planning session

You're brainstorming with the user to produce a spec for their project.

## On startup

1. Read `projects/<name>/spec.md` if it exists (you're continuing)
2. Read `projects/<name>/memory/` for any prior context
3. Read `projects/<name>/decisions/` for decisions already made

## If spec.md doesn't exist

This is a new project. Start by understanding what the user wants to build. Ask questions one at a time. Focus on:
- What problem does this solve?
- Who uses it?
- What's the tech stack? (confirm against project.yaml)
- What are the boundaries — what's in scope, what's not?

Write initial findings to `spec.md` as you go. Don't wait until the end — the spec is a living document during planning.

## If spec.md exists

Read it, summarize where things stand, and ask the user what they want to focus on this session. Pick up where the last session left off.

## Outputs

- Update `spec.md` with everything discussed
- Write any significant decisions to `decisions/<topic>.md`
- Write anything worth remembering across sessions to `memory/`

## When the user says the spec is ready

Before approval, run a grill session using the `/grill-me` skill. This is a structured adversarial review — not a casual "any concerns?" but a rigorous stress test.

Tell the user:

> "Before we lock this in, I want to run /grill-me on this spec. It's a structured adversarial review — I'll challenge assumptions, probe edge cases, and run a pre-mortem to find gaps that would be expensive to discover during execution.
>
> You can skip this, but know that any gaps I'd catch here will surface during execution — when agents may have already built on top of wrong assumptions and rework costs multiply.
>
> Want to run the grill session?"

If they agree:
1. Invoke `/grill-me` targeting the spec
2. After the grill session completes, update `spec.md` with any changes
3. Write the grill findings to `decisions/spec-review.md`

If they skip:
1. Note it in `memory/`: "User skipped /grill-me spec review on <date>. Spec was not adversarially reviewed before approval."
2. This flag changes your behavior during execution — assume there are hidden gaps. Review more conservatively. Be more willing to escalate risk levels.

Then tell the user to approve in the UI.
