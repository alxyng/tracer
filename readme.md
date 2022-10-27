# Tracer

Tracer is a small service written in Go that enables MQTT for the EPEVER Tracer MPPT Solar Charge Controller.

![Grafana Dashboard](img/dashboard.png?raw=true "Grafana Dashboard")

The goal of the project is to **use minimal hardware** and **be extensible**. The repository contains the following binaries, each decoupled using MQTT:

| Binary | Description | Role |
| --- | --- | --- |
| `tracer-controller` | Exposes Tracer functionality over MQTT | Required
| `tracer-api` | Exposes Tracer functionality over HTTP | Optional
| `tracer-app` | (To do) Web UI for the API | Optional
| `tracer-writer` | PostgreSQL writer | Optional

## Prerequisites

### Hardware

RJ45 | Top Down | Hat
:---:|:---:|:---:
![RJ45](img/rj45.jpg?raw=true "RJ45") | ![Top Down](img/topdown.jpg?raw=true "Top Down") | ![Hat](img/hat.jpg?raw=true "Hat")

I'm using the [RS485 Serial HAT](https://thepihut.com/products/rs422-rs485-serial-hat) to communicate with the Tracer MPPT Controller. For instructions on configuring the HAT and setting up the Raspberry PI to use the it, including how to wire RS485 devices, [click here](https://www.instructables.com/How-to-Use-Modbus-With-Raspberry-Pi/).

### MQTT

All services communicate using MQTT. At a minimum, this is required to run `tracer-controller`. [Click here](https://pimylifeup.com/raspberry-pi-mosquitto-mqtt-server/) for instructions on installing an MQTT server.

### PostgreSQL and Grafana

`tracer-writer` consumes the topic `tracer/reading` and writes to PostgreSQL when a message is received. The schema for the database is attached at the bottom of this readme. For instructions on installing PostresSQL on Raspberry Pi, [click here](https://pimylifeup.com/raspberry-pi-postgresql/).

The Grafana Dashboard pictured above is located in the root of the repo as [dashboard.json](dashboard.json). For instructions on installing Frafana on Raspberry Pi, [click here](https://grafana.com/tutorials/install-grafana-on-raspberry-pi/).

## Installing

This assumes you are running 64-bit Raspberry Pi OS. If you are using the 32-bit OS, change `GOARCH` from `arm64` to `arm` in the Makefile. Make sure to also change the `RPI_ADDR` to the address of your Raspberry Pi on the network.

```bash
make # build binaries into ./build
make deploy # copy binaries to /home/pi/tracer/ on you Raspberry Pi
```

After the files have been copied, systemd unit files can be created:

Create `/etc/systemd/system/tracer.api.service`:

```ini
[Unit]
After=network.target
StartLimitIntervalSec=30
StartLimitBurst=5

[Service]
ExecStart=/home/pi/tracer/api
Restart=on-failure

[Install]
WantedBy=multi-user.target
```

Create `/etc/systemd/system/tracer.controller.service`:

```ini
[Unit]
After=network.target
StartLimitIntervalSec=30
StartLimitBurst=5

[Service]
ExecStart=/home/pi/tracer/controller
Restart=on-failure

[Install]
WantedBy=multi-user.target
```

Create `/etc/systemd/system/tracer.writer.service`, making sure to change the value to `TRACER_DATABASE_DSN` to your database DSN:

```ini
[Unit]
After=network.target postgresql.service
StartLimitIntervalSec=30
StartLimitBurst=5

[Service]
Environment=TRACER_DATABASE_DSN=postgres://username:password@host:5432/database
ExecStart=/home/pi/tracer/writer
Restart=on-failure

[Install]
WantedBy=multi-user.target
```

Then enable and start the services:

```bash
make enable # enable systemd services
make start # start systemd services
```

## Schema

To use `tracer-writer`, the following table will need to be created:

```sql
CREATE TABLE readings(
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  over_temperature bool NOT NULL,
  day bool NOT NULL,
  solar_voltage NUMERIC(5, 2) NOT NULL,
  solar_current NUMERIC(5, 2) NOT NULL,
  solar_power NUMERIC(8, 2) NOT NULL,
  load_voltage NUMERIC(5, 2) NOT NULL,
  load_current NUMERIC(5, 2) NOT NULL,
  load_power NUMERIC(8, 2) NOT NULL,
  battery_temperature NUMERIC(5, 2) NOT NULL,
  device_temperature NUMERIC(5, 2) NOT NULL,
  battery_soc INTEGER NOT NULL,
  battery_rated_voltage INTEGER NOT NULL,
  maximum_battery_voltage_today NUMERIC(5, 2) NOT NULL,
  minimum_battery_voltage_today NUMERIC(5, 2) NOT NULL,
  consumed_energy_today NUMERIC(8, 2) NOT NULL,
  consumed_energy_month NUMERIC(8, 2) NOT NULL,
  consumed_energy_year NUMERIC(8, 2) NOT NULL,
  consumed_energy_total NUMERIC(8, 2) NOT NULL,
  generated_energy_today NUMERIC(8, 2) NOT NULL,
  generated_energy_month NUMERIC(8, 2) NOT NULL,
  generated_energy_year NUMERIC(8, 2) NOT NULL,
  generated_energy_total NUMERIC(8, 2) NOT NULL,
  battery_voltage NUMERIC(5, 2) NOT NULL,
  battery_current NUMERIC(5, 2) NOT NULL,
  read_duration interval NOT NULL,
  time timestamp NOT NULL
);
CREATE INDEX readings_time_idx ON readings (time);

CREATE TABLE daily_energy(
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  generated_energy NUMERIC(32, 16) NOT NULL,
  consumed_energy NUMERIC(32, 16) NOT NULL,
  time timestamp NOT NULL
);
CREATE UNIQUE INDEX daily_energy_time_idx ON daily_energy (time);
```

