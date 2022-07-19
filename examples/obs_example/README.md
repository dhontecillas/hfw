# Obs Example

Basic example of how to use the observability libraries.

## Environment setup

In order to run the test and display metrics and logs,
there is a docker compose file under the compse dir:
[`examples/obs_example/compose`](./compose/docker-compose.yml).

It expects to have a local `tmp` dir to store the logs file,
so before running docker compose, just `mkdir tmp` under
the `compose` dir.

After that run:

```
docker-compose up -d
```

And you can point your browser to `http://localhost:3000/login` where
you can login with the default user (`admin`) and password (`admin`),
and you will be prompted to change it.

After that you can configure the sources for the metrics **Prometheus**,
by going to the "cog wheel" menu `Configuration -> Data Sources`, and
add it with `http://prometheus:9090` and it will be configured.

Then again back to `Configuration -> Data Sources` and now add a **Loki**
data source `http://loki:3100`.

Now you are ready to run the executable and see the fake metrics and logs.


## Run the executable

From the hfw root directory type `make obs_example` and the executable
will be created. Run `/.obs_example` to start generating fake metrics.

You can explore the Prometheus metrics, for example, take a look at
the rate of the homer endpoint:

```
rate(requests{app='obs_example', path='/v1/homer'}[1m])
```

Or for example, look at the Loki data source to see the logs there:

```
{app="obs_example"}
```

## Troubleshooting

The IP to access the host machine for scrapping is hardcoded in
the `compose/conf/prometheus.yml` file as `172.17.0.1`. Depending
on you host machine you might need to change it.
