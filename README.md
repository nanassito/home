# Home
Don't waste your time on this repo, this is very much useless to everyone but me and my family. This is the code and architecture that runs the House.

# TODOs
* System
  * Send logs/metrics to Honeycomb
  * Protect MQTT
  * Read zigbee2mqtt logs from mqtt://zigbee2mqtt/bridge/log
* Features
  * Move state to a db
  * For heating, force the fan speed if outdoor temperature is <[5~7]°C
  * No need to water if `avg_over_time(mqtt_humidity{topic="zigbee2mqtt_air_outside"}[7d]) > 85`
  * Maybe double the watering time if `avg_over_time(mqtt_humidity{topic="zigbee2mqtt_air_outside"}[7d]) < 25`
  * Integrate rainfall data
  * Integrate weather forecast?
* UI
  * Irrigation page should show the active sprinkler (and reload upon start)
  * Speedup irrigation page
