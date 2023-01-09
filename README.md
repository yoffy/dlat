# dlat

Measure display latency.

## Hardware Requirements

* Arduino
* Resistor
* CdS

## Sortware Requirements

* Windows (x64)
* go 19.x
* Arduino 1.8.x

## Build

```sh
$ make
```

## Run (on Arduino)

Run `dlat.ino`.

## Run (on Windows)

1. Check your serial port for the Arduino before running
1. Set the display brightness to maximum
1. The CdS contact with the display
1. Run `dlat.exe`

Command example when the serial port is `COM5`: 

```sh
$ ./dlat.exe com5
```
