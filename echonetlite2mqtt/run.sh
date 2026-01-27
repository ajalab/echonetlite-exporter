#!/bin/bash

docker run \
  --name echonetlite2mqtt \
  --network host \
  -e MQTT_BROKER=mqtt://localhost \
  banban525/echonetlite2mqtt
