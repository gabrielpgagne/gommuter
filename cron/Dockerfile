FROM ubuntu:22.04

RUN apt-get update && \
    apt-get install -y \
    cron \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Create app directory
WORKDIR /app

# Copy script and crontab
COPY . .
COPY crontab /etc/cron.d/app-cron

RUN chmod 0644 /etc/cron.d/app-cron
RUN crontab /etc/cron.d/app-cron
RUN touch /var/log/app-cron.log

# Create data directory
RUN mkdir -p /app/data

# https://stackoverflow.com/a/66280277
CMD printenv > /etc/environment && cron && tail -f /var/log/app-cron.log
