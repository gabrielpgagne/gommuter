# gommuter

Fetch via Google Maps API the commute time for an itinerary.

- [ ] Cron jobs
- [ ] Data Dashboard
- [ ] Dockerize

## Compile for Rpi

`GOOS=linux GOARCH=arm64 go build`

## Transfer to Rpi

`scp commute-time rpi.local@Documents/commute_time`