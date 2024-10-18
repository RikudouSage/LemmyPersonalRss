# Lemmy Personal RSS plugin

An app that creates personal RSS feeds based on stuff people saved. Meant to be installed in the same docker
as Lemmy and then available on the same domain through a reverse proxy. It needs to be set up like that because
for the initialization it uses the Lemmy cookie with JWT (**the JWT is stored in a local db, treat the db like a password**).

## How does it work?

The app uses a local database, either SQLite (if configured) or a default in-memory (only useful for testing, after
every restart data will disappear) to store data of every user that initialized the module by visiting `/rss/init`.

The init process creates a unique hash for each user that cannot be guessed and is long enough to avoid accidental
discovery.

## How to use it?

First run it either as a binary or as a docker image. The binary is fully static, single file and has no dependencies.
To build it download go and run `go build -o rss` which will create a binary file `rss`. Then just run it using
`./rss`.

If you want to use docker, simply use the automatically built image, `ghcr.io/rikudousage/lemmy-personal-rss`.
The `latest` version is always the latest stable, `dev` is latest unstable. You can also use individual versions.

## Configuration

Configuration is done using env variables, whether you use docker or a static binary is irrelevant.

| Environment variable   | Description                                                                                                                                                       | Default                                              |
|------------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------|------------------------------------------------------|
| `INSTANCE`             | The instance the app is running on. With no protocol and should be exactly the domain (or subdomain) Lemmy is running on. Defaults to current http host if empty. | `<current http host>`                                |
| `DATABASE_PATH`        | The filesystem path to the SQLite database. If empty, in-memory database is used instead. Always fill this out for production.                                    | `<empty string>`                                     |
| `PORT`                 | The port the app listens on. For docker it might make more sense to use the default and instead rebind it to the OS.                                              | `8080`                                               |
| `CACHE_DURATION`       | How long should each feed be cached for to not hit the api constantly, in seconds.                                                                                | `300`                                                |
| `LOGGING`              | Whether logging is enabled or disabled, should be a string saying `true` or `false`. Note that errors are always logged, regardless of this setting.              | `true`                                               |
| `SINGLE_INSTANCE_MODE` | Whether users from any instance can register, or only from your chosen one. Must be a string `true` or `false`.                                                   | `true` if `INSTANCE` is specified, `false` otherwise |

## Docker compose

Here's an example of a docker compose file that I use in production.

```yaml
services:
  personal_rss:
    image: ghcr.io/rikudousage/lemmy-personal-rss:latest
    restart: always
    environment:
      - INSTANCE=lemmings.world
      - DATABASE_PATH=/opt/database/lemmy-rss.sqlite
    volumes:
      - ./volumes/lemmy-rss:/opt/database
    ports:
      - 3005:8080
```

## Reverse proxy

As mentioned earlier, you need to expose the app under `https://[your-instance]/rss/*`. For example, for lemmings.world
you can try it at https://lemmings.world/rss/init.

To achieve that you need to configure your webserver. I can't list every webserver there is
(but feel free to make a PR to this README with your favourite one!), but here's the one I use for Caddy:

```
lemmings.world {
        // other configuration values
        reverse_proxy /rss/* localhost:3005
        // other configuration values
}
```
