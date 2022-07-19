# Web Example

Basic example of how to use **HFW** to create a web API

## Requirements

You need to have these tools to run the example:

- [docker](https://docs.docker.com/get-docker)
- [docker-compose](https://docs.docker.com/compose/)

## Environment vars to set up:

**Copy the [sample_env.sh](./sample_env.sh) to `env.sh`** and modify
the values to your needs.

If you use the provided [docker-compose.yml](./docker-compose.yml)
file to launch your local environment, you won't need to touch those.

Run:
```
docker-compose up -d
```
to have the service dependecies running as docker containers
(check the docker file to see what is going to be set up).


And source the environment vars to bring the environment vars
to your current environment:
```
source ./env.sh
```

## Run the example
```
go run .
```

And you can open the page in the browser:

[http://localhost:8080/](http://localhost:8080/)

There you can use the link to go to a registration form, that once completed successfully,
should tell you to check your email. But, you should go to the terminal where you are
running the server to check for the log line that is printing the content of the email
(because the provided config, uses the console mailer):

```
[GIN] 2022/06/18 - 18:34:09 | 200 |      8.6649ms |       127.0.0.1 | GET      "/users/register"
Email SENT: To: "example@gmail.com" <example@gmail.com>, From: "No Reply" <noreply@example.com>, Subject: Activate your account
, Text: Activate your account at http://localhost:8080/users/activate/?token=46327889443c9b7e35cb4ed1bb464718a9f16914e434f0afb6f5e562725285d7
{"file":"/home/dhontecillas/prg/projects/hfw/pkg/mailer/log.go:34","level":"info","msg":"Mail SENT (time: 34.861µs): To: \"example@gmail.com\" \u003cexample@gmail.com\u003e, From: \"No Reply\" \u003cnoreply@example.com\u003e, Subject: Activate your account\n, Text: Activate your account at http://localhost:8080/users/activate/?token=46327889443c9b7e35cb4ed1bb464718a9f16914e434f0afb6f5e562725285d7","time":"2022-06-18T18:34:25+02:00"}
```

Text for the message you have the link:

```
http://localhost:8080/users/activate/?token=46327889443c9b7e35cb4ed1bb464718a9f16914e434f0afb6f5e562725285d7
```

## The config file

This app expects to have a `config.yaml` file to run. We could
simply ignore the result from the result of the call to
`InitConfig` and just use the environment vars (however
the framework expects to use the config file, even it
could be empty, and we could just use environment vars).

The file must be placed under `./config` or `./bundle/config`
dir, with a file name of `config.yaml`.


## Basic HTML templates

In order to render basic templates, there is a helper function
`NewMultiRenderEngineFromDirs`, that uses the
[multitemplate gin library](https://github.com/gin-contrib/multitemplate)
under the hood to search for directories called `html_templates`
to load its `.html` files (as well as files under a `templates_html/inc`
subdir that can be used a components for the main templates)


## Running a migration

In the `pkg/bundler/cmd.go` file there is `ExecuteBundlerOperations`
that based on config (file or env vars) can execute several operations:

- **Collect migrations**:
    - `bundler.migrations.collect`
    - `bundler.migrations.scan`: a list of directories to scan to
        find db migration files

- **Execute migrations**:
    - `bundler.migrations.migrate` : the value of this string can be:
        - `up` : to get up to the latest version
        - `down`: to got to a previous version
        - `[int64 value]`: to go to a concrete migration version
    - `bundler.migrations.dst` : the directory where the migration
        files are contained.

- **Pack files**: this operation allows to collect files to be
    deployed to a server.
    - `bundler.pack.dst`
    - `bundler.pack.srcs`
    - `bundler.pack.env`

In this case we are interested only in the migrations part (even we
could collect migrations from `pkg/ginfw/web/wusers`, but for
simplicity, the files have been included in the current example
folder)

So, in this case we can "bypass" the full bundler config, and just
run `ApplyMigrationsFromConfig` function (that will jsut execute migrations).
