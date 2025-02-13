# gommuter

Fetch via Google Maps API the commute time for an itinerary.

## Setting up

Create an environment file `/path/to/gommuter/.env` with the variables:

- `GOOGLE_MAPS_API_KEY`
- `FROM`: the starting point
- `TO`: the destination

For example:

```
# .env
GOOGLE_MAPS_API_KEY="foobarbaz"
FROM="Apple Park"
TO="Googleplex"
```

## Compile for Rpi

`GOOS=linux GOARCH=arm64 go build; cp gommutetime cron/`

## Transfer to Rpi

`scp commute-time rpi.local:Documents/commute_time`

## Launch Docker containers

The project uses two docker images: 

1. A "cron" container which calls the Google Maps API at the times configured in `cron/crontab`
2. A "web" container which reads the CSV data generated from "cron" and creates a dashboard from it at `0.0.0.0:8050/tcp`

To launch the containers:

`docker compose up -d`

You should be able to access the dashboard on port `8050/tcp` on any interface, for example by typing "http://127.0.0.1:8050" in a browser.
