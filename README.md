# gommuter

Fetch via Google Maps API the commute time for an itinerary.

## Setup

Build the Go executable: `go build . && mv gommutetime cron`.

Create an environment file `/path/to/gommuter/cron/cron.env` with the `GOOGLE_MAPS_API_KEY` variable. Then, add itineraries in `cron/crontab`.

## Usage

The project uses two docker images:

1. A "cron" container which calls the Google Maps API at the times configured in `cron/crontab`
2. A "web" container which reads the data from "cron" and creates a dashboard from it at `0.0.0.0:8050/tcp`

To launch the containers:

`docker compose up -d`

You should be able to access the dashboard on port `8050/tcp` on any interface, for example by typing "http://127.0.0.1:8050" in a browser.
