# Go Conventions

- Use `go-chi/chi` for HTTP routing
- Use `jmoiron/sqlx` for database access — hand-written SQL, no ORM
- Error wrapping: `fmt.Errorf("doing X: %w", err)`
- Constructor injection: `NewFoo(deps...) *Foo`
- Table-driven tests as the default pattern
- Small interfaces — one or two methods, defined at the consumer
