Improvements
------------

- Use `coveralls` to show unit test reports: https://coveralls.io/sign-up

- Simplify the current `pkg/notifications` package, because we do not
    have "carrrier"s right now.

- in `pkg/tokenapi/repo_sqlx.go` we have the token api with a field that
    is called `LastUsed` (that we are not using?) but also is not
    the best way to stre it, as the full row would be reused if we
    keep updating it on every use: it should hav its own column, or,
    just store it in some mem storage, and only serialize it to the
    database once every X time. Easy solution is to have its own table,
    for now.

- When trying to register an existing email, send a warning email to
the actual user, telling him, that he already has an account.

- In the logging interface add a way to format strings

- in `pkg/bundler/bundle.go`, it is suggested to use `WalkDir` instead
    of `Walk` because is more perfomant (in `collectWithSuffix` func)

- in `pkg/bundler/bundle.go`, `PrepareBundleDirs` might be better if
    it was private (?)

- check why `bundle_test.go` has some tests commented

Caveats
-------
- the `pkg/usecases/users/repo_sqlx.go` should use a code defined time
    instead of rely on sql's `NOW()` fuction (see the `ActivateUser` call).
    This way, the `consumed` in one table, will match the `created` in
    the other one (because "logically" is done at the same time).

- When creating a user, if it fails in the notification phase, it will have
a ready to activate user, but we don't know how to retry the notification.


What is the issue with the content part in the templates that does not
show up in a normal render.

Decide?

We can create several user registration requests for the same email,
without expiring any previous one.

Check that once is registered, all other activation links are actually
expirer (or consumed somehow).
