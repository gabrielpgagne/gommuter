# gommuter

Fetch via Google Maps API the commute time for an itinerary.

## Setup

Create an environment file `/path/to/gommuter/cron<ID>.env` with the variables:

- `GOOGLE_MAPS_API_KEY`
- `FROM`: the starting point
- `TO`: the destination
- `ID`: ID (e.g., 1)

For example:

```conf
# cron1.env
GOOGLE_MAPS_API_KEY="foobarbaz"
FROM="Apple Park"
TO="Googleplex"
ID="1"
```

You can have multiple itineraries by creating multiple env files with different IDs. A tab is created for each itinerary in the web dashboard.

## Usage

The project uses two docker images:

1. A "cron" container which calls the Google Maps API at the times configured in `cron/crontab`
2. A "web" container which reads the data from "cron" and creates a dashboard from it at `0.0.0.0:8050/tcp`

To launch the containers:

`docker compose up -d`

You should be able to access the dashboard on port `8050/tcp` on any interface, for example by typing "http://127.0.0.1:8050" in a browser.
