# Pastebin scraper

This is a simple pastebin scraper.

It uses the paid [scraping APIs](https://pastebin.com/doc_scraping_api) to get
the pastes and can use MongoDB or Postgres as databases for storage. It also
uses [Healthchecks](https://healthchecks.io/) to monitor it's status.

These environmental variables need to exist in the `.env` file and they will be
used by `docker-compose`:
* `STORAGE_TYPE`: `postgres` or `mongo`

If you choose Mongo you have to set these environment variables:
* `MONGO_URI`
* `MONGO_DB`
* `MONGO_COL`

Alternatively, if you choose Postgres:
* `POSTGRES_HOST`
* `POSTGRES_PORT`
* `POSTGRES_USER`
* `POSTGRES_PASSWORD`
* `POSTGRES_DBNAME`

# Running
You can pull a docker image built for `x86` from docker hub:
```
$ docker pull samirettali/pastebin-scraper
```

Or if you are running it on another architecture (`arm` for example):
```
$ git clone github.com/samirettali/pastebin-scraper
$ cd pastebin-scraper
```

Open `Dockerfile` and change the `ENV` variable according to your architecture,
and then:
```
$ docker build -t pastebin-scraper .
```
