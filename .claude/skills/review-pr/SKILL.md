---
name: review-pr
description: Pull GitHub PR comments, review each one, fix legitimate issues, add tests for edge cases, and reply confirming the fix. Use when the user says "review PR", "check PR comments", "address PR feedback", "fix PR comments", "review-pr", or mentions wanting to handle review feedback on a pull request. Also trigger when the user references a PR number and wants to act on its comments, or says things like "what did reviewers say", "handle the feedback", or "go through the PR comments". Works with both human and bot comments (Copilot, CodeRabbit, etc.).
---

# Review PR

You are a PR comment reviewer and fixer. Your job is to pull all comments from a GitHub PR, evaluate each one, fix the legitimate issues in code, add or update tests to close the edge cases, and reply on the PR confirming what you did.

## Step 1: Identify the PR

Determine which PR to review:

- If the user provides a PR number (e.g., `/review-pr 12` or "check comments on PR #12"), use that.
- If no number is given, find the PR for the current branch:
  ```bash
  gh pr view --json number,title,url,headRefName
  ```
- If no PR exists for the current branch, tell the user and stop.

## Step 2: Pull all comments

Fetch both review comments (inline on code) and conversation comments:

```bash
# Inline review comments (on specific lines of code)
gh api repos/{owner}/{repo}/pulls/{pr_number}/comments

# General conversation comments
gh api repos/{owner}/{repo}/issues/{pr_number}/comments
```

Parse out the relevant fields: `id`, `body`, `path` (for inline comments), `line`/`original_line`, `user.login`, `html_url`, and `in_reply_to_id` (to identify threads vs new comments).

Filter out:
- Your own previous replies (comments from the bot user that are replies to other comments)
- Comments that are already part of a resolved thread where a fix was confirmed

## Step 3: Triage each comment

For each comment, classify it:

| Category | Action |
|----------|--------|
| **Bug / logic error** | Fix the code, add a test |
| **Edge case not handled** | Fix the code, add a test |
| **Style / naming / formatting** | Fix the code, no test needed |
| **Documentation / typo** | Fix it, no test needed |
| **Question / discussion** | Skip — don't fix, don't reply |
| **Already addressed** | Skip — the code already handles it |
| **Disagrees with design** | Skip — flag to user for decision |

When in doubt about whether a comment is actionable, err toward fixing it. Bot comments (Copilot, CodeRabbit, etc.) should be treated the same as human comments — they often catch real issues.

Assign a severity level to each comment:

| Severity | Meaning | Examples |
|----------|---------|----------|
| 🔴 **Critical** | Will cause bugs, data corruption, or crashes in production | Logic errors, wrong identifiers, SQL returning wrong results, missing null checks on hot paths |
| 🟡 **Moderate** | Correctness issue in edge cases, or degrades reliability | Unhandled edge cases, missing validation, inefficient queries that break at scale |
| 🟢 **Low** | Cosmetic, cleanup, or minor improvement | Unused imports, typos, naming, style nits, minor optimizations |

Present the triage to the user before making changes:

```
## PR #12 Comment Review

### Will fix (3):
1. 🔴 @reviewer: "This doesn't handle empty input" (pipeline.py:45) → add guard clause + test
2. 🟡 @copilot: "Potential race condition in batch processing" (watcher.py:78) → add lock + test
3. 🟢 @reviewer: "Typo in error message" (handler.py:92) → fix typo

### Skipping (2):
4. @reviewer: "Why not use Redis instead of SQLite?" → design question, needs your input
5. @reviewer: "Looks good!" → no action needed

Proceed with fixes?
```

Wait for the user to confirm or adjust before proceeding.

## Step 4: Fix each issue

For each comment that needs fixing:

1. **Read the relevant file** to understand the context around the commented line
2. **Make the minimal fix** — don't refactor surrounding code or "improve" things the comment didn't mention
3. **If the fix involves logic or behavior**: write or update a test that specifically covers the edge case the comment identified. The test should fail without the fix and pass with it. Name the test descriptively so it's clear what edge case it covers.
4. **If the fix is cosmetic** (typo, naming, formatting): no test needed

Run the test suite after each fix to make sure nothing broke. If a fix causes test failures, investigate and resolve before moving on.

## Step 5: Reply on each fixed comment

After fixing an issue, post a reply on the PR comment confirming the fix:

```bash
# For inline review comments
gh api repos/{owner}/{repo}/pulls/{pr_number}/comments \
  -X POST \
  -f body="Fixed in $(git rev-parse --short HEAD) — [description of what was done]. Added test: \`test_name_here\`." \
  -F in_reply_to={comment_id}

# For conversation comments
gh api repos/{owner}/{repo}/issues/{pr_number}/comments \
  -X POST \
  -f body="Fixed in $(git rev-parse --short HEAD) — [description of what was done]."
```

Keep replies concise. Include:
- The short commit hash
- What was fixed (one sentence)
- The test name if a test was added

## Step 6: Commit and summarize

After all fixes are applied:

1. **Run the full test suite** to verify everything passes
2. **Stage only the files you changed** — never `git add .`
3. **Commit with a message** that references the PR:

```
Address PR #12 review feedback

- Fix empty input handling in pipeline.py (added guard clause)
- Fix race condition in watcher.py (added processing lock)
- Fix typo in handler.py error message
- Added 2 tests covering reported edge cases
```

4. **Push to the branch** so the PR updates
5. **Print a final summary** of what was done:

```
## Summary
- Fixed: 3 comments
- Skipped: 2 comments (1 design question, 1 approval)
- Tests added: 2
- All tests passing
```

## Step 7: Update lessons learned

After all fixes are committed, check if any of the bugs caught by reviewers reveal a **pattern worth remembering** — something that could prevent the same class of mistake in future work. Not every fix warrants an entry; only add one when:

- The bug was non-obvious and would be easy to repeat (e.g., identifier mismatch across layers, SQL JOIN type causing silent data loss)
- The root cause is a general pattern, not a one-off typo
- The takeaway would meaningfully change how you approach similar code next time

If any fixes qualify, append an entry to `.claude/lessons_learned.md`:

```markdown
## [YYYY-MM-DD] Short title
**What happened:** One-sentence description of the bug the reviewer caught.
**Takeaway:** The general rule or check to apply going forward.
```

Keep entries concise. One PR review might produce zero entries or one — rarely more. Don't log cosmetic fixes or style nits.

## Important guidelines

- **Don't fix what isn't broken.** If a comment suggests a change but the current code is correct, skip it and flag it for the user.
- **Match existing patterns.** Fixes should follow the codebase's established conventions — check how similar code is structured nearby before writing yours.
- **One fix at a time.** Apply fixes sequentially so each one can be verified independently. If a fix causes issues, it's easy to identify which one.
- **Tests prove the fix.** When adding a test for a bug fix, the test should target the specific edge case the reviewer identified. Don't write broad tests — write the test that would have caught the bug.
- **Don't reply to design discussions.** If a comment questions an architectural decision, flag it for the user rather than making a judgment call. These need human input.
