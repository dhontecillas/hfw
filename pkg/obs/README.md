# `obs` for **Observability**

## `Insighter` to aggregate observability tools

_Logging_, _Metrics_ and _Tracing_, are the tools that we have
for observability (or telemetry). The three of them
share some functionality: all of them allow to set tags
with values.

The package has be called `obs` (short form for Observability),
to avoid long hard to type

# Interfaces definitions

There are ongoing efforts to standarize the way to use
traces, metrics and logs with the [OpenTelemetry.io](https://opentelemetry.io).

Creating our interfaces for what we need allows us to use
the current libraries but easily switch to the opentelemetry
library to be "vendor independent".


## `Insighter` interface

The Insights object groups all these observability methods. In
order to be easy to use, the "subsystems" should have very
short name.

- **`T`**: for Traces interface
- **`M`**: for Metrics interface
- **`L`**: for Logs interface

A part from containing the other interfaces, the
insights objects could provide also methods for setting tags
(`(key, value)` pairs), that will be set automatically
for all the subsystems.

### Tag Definitions

Adding tags to metrics or traces can cause extra costs,
and performance issues when indexing them (specially with
tags that can have lots of different values).

Providing **Tag Definitions** when creating an `Insighter`
object allows us to discard tags that are not defined at startup
time. This restriction is only applied at the `Insighter` level,
so if you want to bypass it you can use the `ins.L` / `ins.M` / `ins.T`
to set the tags.

```go
type TagDefinition struct {
	Name    string
	TagType int
	ToL     bool
	ToM     bool
	ToT     bool
}
```

It also allows to send those tags to each of the "subsystems" (logs, metrics, traces).

Adding tags to a log message is usually not an issue as long as those
tags are not automatically used for indexing (usually log collectors
have their configuration to define on what to index).


A recommendation is to define the tag names with constants, and not use
something like: `ins.L.Str("foo", "bar")` but `ind.L.Str(kTagFoo, "bar")`.
(This way we can later add some tool to process


## Using a "builder" pattern.

In order to encapsulate the configuration for the different
implementations of each subsystem, we use obtain a "builder function"
at startup time per each implementation that we want to have, for
each of the subsystems (Logger / Meter / Tracer).

For example, at start up time we would call:

```go

var promMeter MeterBuilderFn

func main() {
    var err error
    var conf *PrometheusConfig = ReadPrometheusConfigFromSomewhere()
    promMeter, err = NewPrometheusMeterBuilder(log, conf)
    if err != nil {
        panic("")
    }
}
```

So, we could have several builder functions availabe for when we
receive a new request (yep, we need a global place to have that function :/).

Usually, after that we would also create / load all the TagDefinitions
(those tags that can be shared as are read only)

```go
var (
    logrusLoggerBuilder logs.LoggerBuilderFn

    promMeterBuilder    meter.MeterBuilderFn
    ddogMeterBuilder    meter.MeterBuilderFn
    multiMeterBuileer   meter.MeterBuilderFn

    nopTracerBuilder    tracer.TracerBuilderFn
    jaegerTracerBuilder tracer.TracerBuilderFn

    insighterTagDefinitions  *obs.InsighterTagTargets
)
```


### Steps to instantiate a new `Insighter` instance

- Call one of the `LoggerBuilderFn` that will return the
    Logger interface to be used
- Call the selected `TracerBuilderFn` and `MeterBuilderFn`
    (passing the instantiated builder).

- Call the `NewInsighter` constructor passing the created
    subsystem instances and the tag targets.

- Set any tags know at the moment of creation


## For API Requests

We want to create a new `Insigher` instance for each request,
never reuse a global one. The builder function approach
allows us to select the implementation depending
on configuration flags / environment.

### Access to the Insights instance

#### `FromContext`

We need to have access to an `Insights` instance as soon as
we receive a request, so it needs to be created at the outer
layer of the middleware stack, and be attached to the context
that is passed to other middlewares.

##### Provide a NOP Insigts instance when not set

The `obs.FromContext` implementation should never return a null
instance. If no `Insights` object has been set, a No-Op one
will be returned, to avoid nil checks in "client" code, and
also to have to create extra dependencies to be added in mock objects.


#### Extract it from the context at the controller layer

In order to reduce the amount of stuff that we retrieve from the
context, we should extract the `Insights` instance from the context
in the Controller, and pass it explictly to the use case.

A UseCase can:
- have an explicit `Insights` param in their function calls to be able
    to log, report metrics, etc...
- store the `Insights` instace at construction time (with `.New(... , obs Insights)`, so
    all internal functions do not need to have the extra param).




## Logs

Log interface styles:

- **logrus** style: for each message we set the additional `(key, values)` with want: this
    approach is more "free form", any logged output can have any additional key value
    pairs of any type.


- **zerolog** style: for each additional `(key, value)` we call a method that accepts
    an strong typed value (`Str(key, val string)`, `I64(key string, val int64)` ...)

- **struct** type approach: we do not provide "free form" `(key, value)` to be attached
    to the log. Instead we must define a struct type for each log that we want to
    output.


### Zerolog style

The zerolog style approach allows for giving more meaning to each value (as it has the type),
and allows to use the optimized zerolog library ;).


### `struct` style

This is the more strict one, and less pragmatic to write logs. The good part of this interface
would be that each log output is like a `Message` (like the ones that could be sent
through a rabbitmq), only that is written to the output, and that we could extract the type
definitions from the code and have a JSON Schema. This one is a little bit overkill, but ...
wouldn't be nice to have a catalog of all possible log outputs!? :)

In order to 'inherit' message attributes, we can just use embedded structs.


## Metrics

The most common metrics usage are:

- counters (monotonic / not)
- rates (also called gauges, in statsd style)

Currently, I haven't seen any reference to the 'histogram' or 'distribution' in the
OpenTelemetry spec. However, those are 'client side' aggregations, and perhaps is
more a concern of how the underlying counter is implemented. So, at `Insights`
object creation time, we could define some keys, that for rates, should be treated
as an histogram aggregation (abstracting the client code about how that metric
will be displayed).

Like in the case of logs, we can have stronger definitions (struct) of the metrics
that can be used.


## Traces

At some point we should start to use traces (and its `Span`'s) to keep trace
of the logic. And make sure that when we make a request to another system
we include the trace in some way (a header, message field, etc..).


# Structs style

**About struct approach**: One the problem of having structs is that prevents us from
having shortcuts to set tags in all subsystems. That can be kind of solve by creating
an 'InsightContext' struct with typed data, whose fields are always set for any of the
the metrics, logs and traces.


# Concurrency

In case of concurrency (spawning several goroutines), we should clone the Insights instance
(that would cost copying an small amount of dicts ?), and pass those to the goroutines.

