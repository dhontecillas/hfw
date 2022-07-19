# **HFW***'s Bundler

The `bundler` package identifies certain paths as containing static
assets:

- `*/static` : files that should be served as they are, without going through the app
- `*/html_templates` : those are files that should be stored in the server, in an
    accessible place for the app to load them and use them to render the ouptut
- `*/notifications/templates` : those files are collected, and used to send notifications


## The "bundling" process

The bundling process consists of creating a folder structure that can easily be
deployed to a server.

It consists on the following directories:

- `app` : The place where the app executable will be put, with the files that the
    app must be able to acces:
    - `app/dbmigrations` : the collected migration files, that can be run before
        start the application (be careful when launching several instances of
        the same app, as concurrency has not been tested: in that case it might
        be better to run migrations as a separate initial process when doing a
        new deployment).
    - `app/data`: tempaltes to be used by the app
        - `app/data/html_templates`: html templates to render
        - `app/data/notifications/templates`: templates to be used for messages
- `static` : A directory containing all the collected static files, in a single place, so we can put them in a place to be directly served by an http server, or some CDN

### Gathering files

Be careful of not storing the bundle under a subdir that is in the path of the files
to collect or you would end with a recursive file structure.


### Compiling the executable

Currently the command to build the executable is hardcoded and executed using:

```go
exec.Command("go", "build", "-o", outExe, inPkg)
```

So no extra options can be passed or configured.


### Compressing the "bundle"

As in the compile step, this is a dirty hardcoded call to the `tar` tool using
`exec.Command`.


## Executing the bundle operations

There are several config options that allows us to tweak what we want the bundler to do.
Those can be found in `pkg/bundler/cmd.go`.

### Migrations

Migration are pairs of files `*.up.sql` / `*.down.sql` that contain the database
changes required to for a given feature.

#### Collecting the files

- `bundler.migrations.collect`: true / false to decide if migrations should be collected
- `bundler.migrations.dst`: the directory where the found migrations will be placed
- `bundler.migrations.scan`: a list of directories to scan for new migrations

There is also an option, not directly related to "Bundling", that is the
`bundler.migrations.migrate`, and that is used to actually run the migrations. The values can be `up` , `down` or a number for the migration.

When collecting the files, not only the name of the file is taken into account, but
also a hash of its content is computed, just to check that there has not been changes
to an existing file, or that there are files with the same name but different content
in different directories.

In the process of collection the migrations, a new migration number is assigned
incrementally. Is done that way, because if the destination directory already
have the migration applied, it won't collect it again.

#### Packing the migrations

- `bundler.pack.dst`: Destination of the collected migration files
- `bundler.pack.srcs`: Additional directories to scan for migrations files

## TODO

- Detect recursive collection of files.
- Allow to configure the build command for the executable.
- Perhaps running the migrations should not be part of the bundle package.
