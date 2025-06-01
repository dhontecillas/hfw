# **HFW**: Handy FrameWork
![Coverage](https://img.shields.io/badge/Coverage-13.9%25-red)

[![Coverage Status](https://coveralls.io/repos/github/dhontecillas/hfw/badge.svg?branch=main)](https://coveralls.io/github/dhontecillas/hfw?branch=main)

Go framework for API and Web development.

## About

The HFW is a collection of libraries that helps into
building Web and API applications that rely on some
common services, and existing libraries:

* [PostgreSQL](https://www.postgresql.org/)
* [Redis](https://redis.io/)
* [Gin-Gonic](https://gin-gonic.com/)

## How to start

# Examples

## Web Example

[Documentation for the Web example](./examples/web_example/README.md)
In `examples/web_example` there is an exaple of how to set a
basic email authentication flow

## Observability Example

[Docmentation for the observability example](./examples/obs_example/README.md)


### Why Postgres?

Mostly a personal preference, but also because it has
native support to UUIDs, and we can use [ULIDs](https://github.com/oklog/ulid)
as IDs, that can be directly converted to UUID, and
that provides the benefits of an UUID with being incremental
so have better behaviour when indexing them, or to use
them as a pagination cursor.

### Why Redis?

Mostly because can do whatever a Memcached service provides,
but also provides other interesting data structures that
allow for efficient ratelimits, and can provide Pub/Sub channels
to distribute workload among different instantces.

### Why Gin Gonic ?

Mostly because is widely adopted and easy to use. Not benchmarking
nor comparisons have been made.

# Packages

## General usage

### `ids`

Just a wrapper over existing [OK log's ULIDs library](https://github.com/oklog/ulid)
with extra functions for formatting it as an UUID, or a Shuffled version to not
make so obvious that are sequential IDs. It also implements a **Scan** function for
database reading.


### `bundler`

The bundler package provides functions to collect assets that will be placed in
a single place in order to be deployed. Also, can collect and run migrations.

[Bundler Docs](./doc/bundler.md)


#### `AppUpdater`

The "app updater" of the bundler keeps track of modifications made to source files
and resource files, in order to create a new binary to run when something has changed
and to update the directory of resources.

### `config`

The `config` package reads the configuration, to prepare the required dependencies
for the app.

#### `CONFVARIANT` environment var

By default it will try to load the configuration from a file named
`config.yaml` under the `config` dir, but if the `CONFVARIANT` is set,
it will try to load instead the `config.[CONFVARIANT].yaml` file.

If no config found under that `config` dir, the `bundle/config` will be tried.

#### ExternalServices

`config.BuildExternalServices`, reads the configuration and selects the
implementation for the maile, the SQLDB and the notification services, and
a function to obtain an Insighter instance to log and record metrics.

With a call to `ExtServices()`, a structure with the same info that external
services is created but with a clone of the insighter.

#### Redis

For some functionality, like session handling, a redis server is required.

- `db.redis.master.host`
- `db.redis.master.port`

#### DB (Postgresql)

- `db.sql.master.name`
- `db.sql.master.host`
- `db.sql.master.port`
- `db.sql.master.user`
- `db.sql.master.pass`

#### Insights

- `prometheus.enabled`
- `prometheus.port`
- `prometheus.path`
- `prometheus.prefix`
- `sentry.enabled`
- `sentry.dsn`
- `graylog.enabled`
- `graylog.port`
- `graylog.host`
- `graylog.prefix`

#### Mailer

By using the `mailer.preferred` configuration setting the mailer to
be used can be selected.

Valid values are:

- sendgrid
- mailtrap
- roundcube
- console
- nop

##### Sendgrid

The required config to use sendgris is:

- `sendgrid.key`

##### Mailtrap

The required config to use mailtrap are:

- `mailtrap.user`
- `mailtrap.password`

##### Roundcube

##### Console

##### Nop

#### Notifications:

##### `notifications.templates.dir`

Will select the directory where the templates can be found.


### `mailer`

An interface definition to send simple emails.

There are different implementations:

- **Mailgun**: to use mailgun mailer
- **Mailtrap**: an smpt implementation to send mails to
    the mailtrap service (to check how emails are
    displayed and make visual tests).
- **Console**: to output the email to stadout (useful
    for development environments)
- **Mock**: useful for testing, as it stores all mails
    "sent" in memory.
- **Nop**: useful for testing when we don't care about.
    emails, but some code needs to send some email


It also contains a mailer wrapper: **`LoggerMailer`**,
that logs errors occured in another mailer instance.


### `i18n`

Package to hold functions related to i18n and localization.

#### `i18n/langs`

It contains the definitions for languages entities.

Usually you might want to use only these two:

- **LangCodesAndNames**: with the names of the language and
    its variation, along with the ISO codes for the language
    and its variation.

- **LangCodes**: contains only the ISO codes for the a given
    language and its variation.

But you can also get a structure with all the language definitions
using the `GetLangs` function that will return an array of `Lang`,
and its children `LangVariant`.


### `pkg/notifications`

A package to store templates to send email notifications. This package
is "work in progress". The idea would be to extend it with more ways
to send notifications (like websockets). Currently there is only an
email sender, and a "filesystem composer" that can read Go templates
from files.


## Use Cases

These packages provide some basic functionality that is usually needed
in a lot of applications, like use manages, api tokens..


### `usecases/users`

This package implements the usual email registration workflow. For emails,
it works using the `pkg/notifier` package, using templates for:

- Requests Registration (template `users_requestregistration`)
- Request Password Reset (template `users_requestpasswordreset`)

#### Gin

Under the `pkg/ginfw/auth` package, there is the key to store a an
authenticated user ID.

For managing web user sessions there is the `pkg/ginfw/web/session/session.go`


### `tokenapi`

A simple entity definition for letting users create their own API keys,
and use them to perform actions using an exposed API.

### `consterr`

A basic definition of a an error that will be a string. (Might disappear later on)

# Web App Structure


## URLs

### Main pages

Public pages


### Web App API

The web API is meant to be used by the frontend code. As such, it works by using
a CSRF token, cnd CORS restrictions to get requests only from the same domain.

These wep apis are served under `https://maindomain/wapp/` path

### Generic API

These are the stable APIs for mobile apps or third party integrations.


# Error handling

- Errors are logged at the usecase layer whenever is possible.
- Errors coming from other libraries that aren't know


# Alternatives 

These are much more elaborated frameworks, production ready that you might want to have a look at:

- [Iris](https://github.com/kataras/iris)
