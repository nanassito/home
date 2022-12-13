from collections import defaultdict
import json
import yaml
from pathlib import Path


REPO =  Path(__file__).resolve().parent.parent


with (REPO / "configs" / "inputs" / "zigbee.json").open() as fd:
    zigbee = json.load(fd)


def mqtt(device):
    return f"device/{'/'.join(device['prometheus']['location'])}/{device['prometheus']['type']}"


def promLabels(device):
    return {
        "location": "_".join(device["prometheus"]["location"]),
        "type": device["prometheus"]["type"],
    }


print("Update `devices` in the Zigbee2mqtt configurations")
ZIGBEE2MQTT = (REPO / "zigbee2mqtt" / "configuration.yaml")
with ZIGBEE2MQTT.open() as fd:
    cfg = yaml.load(fd, yaml.Loader)
cfg["devices"] = {
    address: {
        "friendly_name": mqtt(device)
    }
    for address, device in zigbee["server"].items()
}
with ZIGBEE2MQTT.open("w") as fd:
    yaml.dump(cfg, fd)


print("Update switches configurations")
cfg = {
    nickname: {
        "Mqtt": {
            "GetTopic": f"zigbee2mqtt/{network}/{mqtt(device)}",
            "GetRegex": f'.*"{switch["line"]}": ?"(?P<State>(?P<Active>ON)?(?P<AtRest>OFF)?)".*',
            "SetTopic": f"zigbee2mqtt/{network}/{mqtt(device)}/set",
            "MsgActive": json.dumps({switch["line"]: "ON" if switch["default_open"] else "OFF"}),
            "MsgRest": json.dumps({switch["line"]: "OFF" if switch["default_open"] else "ON"}),
        },
        "Prometheus": {
            "Metric": f"mqtt_{switch['line']}",
            "Labels": promLabels(device),
            "ValueActive": 1 if switch["default_open"] else 0,
            "ValueRest": 0 if switch["default_open"] else 1,
        }
    }
    for network, devices in zigbee.items()
    for device in devices.values()
    for nickname, switch in device.get("switches", {}).items()
}
with (REPO / "configs" / "switches.json").open("w") as fd:
    json.dump(cfg, fd, sort_keys=True, indent=4)


print("Update air/HVAC configuration")
sensors = {
    device["prometheus"]["location"][0]: {
        "mqttTopic": f"zigbee2mqtt/{network}/{mqtt(device)}",
        "prometheusLabels": promLabels(device),
    }
    for network, devices in zigbee.items()
    for device in devices.values()
    if device["prometheus"]["type"] == "air"
}
cfg = {
    "sensors": {},
    "hvacs": {},
    "outside": sensors["backyard"],
}
with open(REPO / "configs" / "inputs" / "rooms.json") as fd:
    specs = json.load(fd)
for room, spec in specs.items():
    if room != "livingroom":
        continue  # Until air is ready for prime time.
    hvacs = set()
    for hvac_file in spec["hvacs"]:
        with (REPO / hvac_file).open() as fd:
            raw = [l for l in fd.readlines() if "!secret" not in l]
        hvac = yaml.load("".join(raw), yaml.Loader)
        climate = hvac["climate"]
        cfg["hvacs"][climate["name"]]= {
            "room": room,
            "setModeMqttTopic": climate["mode_command_topic"],
            "reportModeMqttTopic": climate["mode_state_topic"],
            "setFanMqttTopic": climate["fan_mode_command_topic"],
            "reportFanMqttTopic": climate["fan_mode_state_topic"],
            "setTemperatureMqttTopic": climate["target_temperature_command_topic"],
            "reportTemperatureMqttTopic": climate["target_temperature_low_state_topic"],
            "prometheusLabels": {"type": "hvac", "location": climate["name"]},
        }
        hvacs.add(climate["name"])
    cfg["sensors"][room] = sensors[room]
cfg["schedules"] = {}
with open(REPO / "configs" / "inputs" / "schedule.json") as fd:
    for scheduleName, schedule in json.load(fd).items():
        if scheduleName not in cfg["sensors"]:
            continue
        schedule["roomName"] = scheduleName
        cfg["schedules"][scheduleName] = schedule
with (REPO / "configs" / "air.json").open("w") as fd:
    json.dump(cfg, fd, sort_keys=True, indent=4)