---
locked_sections:
  - "The Prime Directive: Never Guess"
  - "Code Philosophy"
  - "Architecture Principles"
  - "Testing"
  - "Verification Loop"
  - "Git & Workflow"
---

# Design Principles

Core engineering philosophy for this project. Locked sections are inherited from the workspace and must not be modified unless explicitly asked. The Tech Stack section evolves per-project.

## The Prime Directive: Never Guess

- **Never guess** about APIs, framework behavior, best practices, or library usage. If not 100% certain, look up the latest official documentation first.
- If docs are unavailable or ambiguous, **ask** before proceeding. Do not assume, improvise, or rely on potentially outdated knowledge.
- Applies to: API signatures, framework conventions, config options, CLI flags, library methods, and any behavior you cannot verify from the codebase or docs.
- Wrong guesses compound into hard-to-find bugs. **When in doubt, look it up or ask.**

### High-risk identifiers — always verify from docs:

- Model names / model IDs
- API endpoint paths and versions
- SDK method signatures and config object fields
- Auth mechanisms (API key vs service account vs OAuth)

### Debugging library internals:

When a bug involves library internals, **read the installed source code first** — find it in the project's dependency directory (`node_modules/`, `.venv/`, `vendor/`, Go module cache). Never state how a library works based on memory alone.

## Code Philosophy

### Simplicity Above All

- Prefer the simplest solution. No abstractions for hypothetical scenarios.
- Only change what's needed. Don't refactor, add comments, or "improve" surrounding code unless asked.
- Three similar lines of code is better than a premature abstraction.

### Consistency Is Non-Negotiable

- Once a pattern is established, propagate it to all similar code. No mixed generations of style.
- Read before writing. Understand existing patterns before changing anything.
- Match the project's existing naming, formatting, and conventions.
- Follow the language/framework's idiomatic best practices — they take priority over personal style.

### Code Reads Like a Narrative

- Orchestrator methods describe *what* happens; helper methods handle *how*.
- Extract even small operations into named methods when it makes the calling code read like an outline.
- Guard clauses at the top, happy path flows straight down at the lowest indentation.
- Use blank lines to group related logic into visual blocks — code reads as grouped chunks, not a wall of text.

### Small Files, Clear Boundaries

- One logical unit per file — extract early, not when it "gets big enough."
- If you can name the concept, it gets its own file.

### Delete, Don't Comment

- Delete dead code. Don't comment it out, rename to `_unused`, or add `// removed` markers.
- No backwards-compatibility hacks for removed code.

## Architecture Principles

Follow SOLID principles. Keep coupling low and cohesion high.

### Feature-Folder Structure

Unless the language/framework has a stronger established convention:

```
src/
  features/
    auth/           # Everything for auth: routes, services, controllers, tests, types
    billing/        # Everything for billing
  shared/           # Cross-feature utilities, components, middleware
  models/           # Relational entities / domain models
```

- Each feature folder is self-contained: routes, services, controllers, types, and tests live together.
- `shared/` holds code used by multiple features. Move code here only when reuse is proven, not speculative.
- Language/framework best practices override this structure when they conflict.

### Design Patterns

- **Dependency injection over globals.** Classes/structs receive dependencies via constructor. Don't pull from global state or import singletons.
- **Bundle config into dedicated objects.** Group related settings into immutable config objects and inject those — don't pass raw values or read env vars deep in the call stack.
- **Composition at the entry point.** All wiring and construction happens in one visible place. Use a factory function to construct the dependency graph from config.
- **Composition over inheritance.** Prefer composing objects from smaller parts. Use inheritance only for genuine is-a relationships.
- **Separate what's yours from what's theirs.** External services get interfaces and dedicated provider folders. The boundary: if switching the vendor requires rewriting the integration, it needs an abstraction. Provider-specific code never mixes into business logic.
- **Validate at boundaries only.** Validate at system boundaries (user input, external APIs). Trust internal code and framework guarantees.

## Testing

- Run existing tests after changes to verify nothing broke.
- Write tests for new functionality unless told otherwise.
- Match the existing test framework and patterns.
- **NEVER update a test just to make it pass.** If a test fails, the code is guilty until proven innocent. Verify the actual behavior is correct first:
  - Web apps: run in-browser validation before touching the test.
  - APIs: make real requests to confirm the response is correct.
  - CLI tools: run the command and inspect actual output.
- Only update a test when you have confirmed the new behavior is genuinely correct and the test expectation was wrong or outdated.

## Verification Loop

Before considering any task complete:

1. Typecheck (if applicable)
2. Lint
3. Run affected tests
4. Before PRs: run the full test suite
5. **After major refactors:** full consistency pass across all touched files:
   - Mixed patterns (old style vs new style living side by side)
   - Naming inconsistencies (e.g. camelCase in one file, snake_case in another)
   - Orphaned imports, dead code, or leftover references to removed code
   - Architectural violations (logic in the wrong layer, broken feature-folder boundaries)
   - No Frankenstein code — the codebase should read like one person wrote it

Don't just run the steps — prove it works. Ask: "Would a staff engineer approve this?"

## Git & Workflow

**Trunk-based development:**

- `main` is the trunk — always shippable, never broken.
- All work happens on short-lived feature branches cut from `main`.
- One branch per phase or logical unit of work.
- Open a PR only when work is complete and ready to merge.
- Branch naming: `phase/<name>`, `feat/<name>`, `fix/<name>`.
- Commit messages: concise, imperative mood, focused on "why" not "what."
- Stage specific files — never `git add .` or `git add -A`.

### Planning

- Enter plan mode for any non-trivial task (3+ steps or architectural decisions).
- If an approach goes sideways, stop and re-plan immediately — don't keep pushing a broken path.
- Write detailed specs upfront to reduce ambiguity.

### Bug Fixing

- When given a bug report: investigate and fix it. Don't ask for hand-holding.
- Read logs, trace errors, find failing tests — then resolve them.

### Lessons Learned

Maintain `.claude/lessons_learned.md` — read it at the start of every task, reference it before making decisions.

- **Update immediately** — after ANY correction from the user, after hitting an unexpected error, discovering a framework quirk, or finding a non-obvious solution. Write a rule that prevents the same mistake.
- Never let it go stale. If a lesson no longer applies, remove it.

## Tech Stack

### Go Conventions

- Use **air** for hot reloading during development (`air` watches for file changes and rebuilds automatically).
- Set up a `.air.toml` config at the project root to configure build commands, watched directories, and excluded paths.
- Use `go-chi/chi` for HTTP routing. Lightweight, `net/http` compatible, composable middleware.
- Use `jmoiron/sqlx` for database access. Hand-written SQL, no ORM. `sqlx.Get`/`sqlx.Select` for struct scanning, `db.NamedExec` for named parameters.
- No GORM, no Ent, no query builders. You control the queries.
- Interfaces are small — one or two methods. Define where consumed, not where implemented.
- Constructor injection: `NewFoo(deps...) *Foo`.
- Error wrapping: `fmt.Errorf("doing X: %w", err)`.
- Table-driven tests as the default pattern.
- For database tests, use a real test database — not mocks.

When debugging library internals in Go, find the source in the module cache:
```bash
find $(go env GOMODCACHE) -path "*<module>*" -name "*.go" | xargs grep -l "<symbol>"
```

### Node/TypeScript Conventions

- Use the project's package manager (`npm`, `pnpm`, or `yarn`) consistently — don't mix.
- `npm run dev` / `pnpm dev` — start the dev server
- `npm test` / `pnpm test` — run tests
- `npm run lint` / `pnpm lint` — run linter

When debugging library internals in Node, find the source in node_modules:
```bash
find node_modules -name "*.js" -path "*<package>*" | head -20
```

<!-- This section is mutable and evolves per-project. -->
