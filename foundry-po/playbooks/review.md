# Review session

An agent has completed a task. Review their work.

## On startup

1. Read the session context for task_id, agent_role, risk_level, branch
2. Read `projects/<name>/plan.md` to understand the task in context
3. Read `projects/<name>/execution_spec.md` for current state of the spec
4. Check the diff on the agent's branch

## Review by risk level

**Low risk:** Skim the diff. Check that tests pass and linter is clean. If acceptable, mark the task as complete in `plan.md`.

**Medium risk:** Review each file change against the spec. Check that the implementation matches the task description. Verify tests cover the core behavior. Flag anything that deviates from the spec — update `execution_spec.md` if the deviation is justified, or send the agent back to fix it.

**High risk:** Line-by-line review. Check security implications, error handling, edge cases. Verify comprehensive test coverage. If the work is acceptable, create a PR for human review. Write review notes to `artifacts/reviews/<task_id>.md`.

## If the spec needs to change

1. Log the mutation to `mutations.jsonl`:
   ```json
   {
     "timestamp": "<ISO 8601>",
     "field_changed": "<section.subsection>",
     "description": "<what changed>",
     "reason": "<the constraint or discovery that forced this change>",
     "diff": "<before and after>"
   }
   ```
2. Update `execution_spec.md`
3. Notify affected agents via task updates

## Outputs

- Update `plan.md` task status (done, or back to in_progress with notes)
- Write review notes to `artifacts/reviews/` for medium and high risk
- Update `execution_spec.md` if the plan diverged
- Append to `mutations.jsonl` if spec changed
