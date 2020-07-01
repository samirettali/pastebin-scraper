# Pastebin scraper

This is a simple pastebin scraper.

It uses the paid [scraping APIs](https://pastebin.com/doc_scraping_api) to get
the pastes and by default it saves them in MongoDB. It uses
[Healthchecks](https://healthchecks.io/) to monitor it's status.

These environmental variables need to exist in the `.env` file and they will be
used by `docker-compose`:
* `MONGO_URI`
* `MONGO_DB`
* `MONGO_COL`
* `HEALTHCHECK`
