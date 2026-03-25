# Phase transition session

A phase in the execution plan has all tasks marked complete. Evaluate readiness for the next phase.

## On startup

1. Read `projects/<name>/plan.md`
2. Read `projects/<name>/execution_spec.md`
3. Review artifacts produced during the completed phase

## Your job

1. Verify all tasks in the phase are genuinely complete (not just marked done — check that outputs exist)
2. Check that phase deliverables match the spec
3. Identify any gaps or issues that need resolution before moving on
4. If ready, update the next phase's status to in_progress and unblock its tasks

## Outputs

- Update `plan.md` phase statuses
- Write phase completion summary to `artifacts/phases/<phase>.md`
- Update `execution_spec.md` if anything shifted
- Flag issues to the user if the phase isn't truly ready
