# Node/TypeScript Conventions

- Use the project's package manager (`npm`, `pnpm`, or `yarn`) consistently — don't mix.
- `npm run dev` / `pnpm dev` — start the dev server
- `npm test` / `pnpm test` — run tests
- `npm run lint` / `pnpm lint` — run linter

When debugging library internals in Node, find the source in node_modules:
```bash
find node_modules -name "*.js" -path "*<package>*" | head -20
```
