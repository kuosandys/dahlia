# anthology

Small CLI utility to get the latest articles from RSS/Atom feeds and compile them into an ebook format optimized for Kobo devices (`.kepub`). Optionally uploads the generated ebook to Dropbox so that it can be downloaded on a Kobo device. Designed to be run as a cron job with GitHub Actions.

Intended for personal use - ymmv.

## Usage

```
anthology [flags]

FLAGS:
  --configFileName      (string: "config")    name of config file
  --configFilePath      (string: ".")         path to config file
  --upload              (bool:   false)       upload to Dropbox
  --dropboxAppKey       (string: "")          Dropbox app key
  --dropboxAppSecret    (string: "")          Dropbox app secret
  --dropboxRefreshToken (string: "")          Dropbox refresh token
```

## Configuration

The application can be configured via `config.yml`:

- `lastHours`: time frame to check for new articles; defaults to the last 168 hours (1 week)
- `urls`: URLs of RSS/Atom feeds to fetch articles from
- `dropboxKoboFolder`: where to upload the generated ebook; defaults to Kobo's default application folder at `/Apps/Rakuten Kobo/`

## Setup

In order to upload the generated ebook to Dropbox, the Dropbox-specific flag values must be provided.

Dropbox app key and app secret can be found in Dropbox's app console after [creating an app](https://www.dropbox.com/developers).

The Dropbox refresh token is used by the application to obtain a short-lived OAuth access token for Dropbox's API. A utility script has been included to facilitate getting the refresh token and can be invoked from the project's root directory with `go run ./cmd/token --dropboxAppKey=<APP_KEY> --dropboxAppSecret=<APP_SECRET>`.

## Automating

The workflow as defined in `.github/workflows/run.yml` is responsible for running the application as a cron job with GitHub actions. The `TARGET_BRANCH` environmental variable can be used to configure the branch where the configuration file can be found, and defaults to the `build` branch.
