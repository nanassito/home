import json
import yaml
from pathlib import Path


REPO =  Path(__file__).resolve().parent.parent


with (REPO / "configs" / "devices.json").open() as fd:
    data = json.load(fd)


# Update `devices` in the Zigbee2mqtt configurations
ZIGBEE2MQTT = (REPO / "zigbee2mqtt" / "configuration.yaml")
with ZIGBEE2MQTT.open() as fd:
    cfg = yaml.load(fd, yaml.Loader)
cfg["devices"] = {
    xid: {"friendly_name": name}
    for name, xid, in data["zigbee2mqtt"]["server"].items()
    if xid != ""
}
with ZIGBEE2MQTT.open("w") as fd:
    yaml.dump(cfg, fd)


# Update switches configurations
cfg = {
    name: {
        "Mqtt": {
            "SetTopic": f"zigbee2mqtt/{network}/{'/'.join(location)}/valve/set",
            "MsgActive": json.dumps({line: "ON" if default_open else "OFF"}),
            "MsgRest": json.dumps({line: "OFF" if default_open else "ON"}),
        },
        "Prometheus": {
            "Metric": f"mqtt_{line}",
            "Labels": {
                "location": "_".join(location),
                "type": dtype,
            },
            "ValueActive": 1 if default_open else 0,
            "ValueRest": 0 if default_open else 1,
        }
    }
    for name, (network, location, dtype, line, default_open) in {
        "switch_adhoc": ("server", ("adhoc", ), "switch", "state", False),
        "switch_washer": ("raspi", ("garage", "washer"), "power", "state", False),
        "valve_backyard_side": ("server", ("backyard", ), "valve", "state_l1", True),
        "valve_backyard_house": ("server", ("backyard", ), "valve", "state_l2", True),
        "valve_backyard_school": ("server", ("backyard", ), "valve", "state_l3", True),
        "valve_backyard_deck": ("server", ("backyard", ), "valve", "state_l4", True),
        "valve_frontyard_street": ("raspi", ("frontyard", ), "valve", "state_l1", True),
        "valve_frontyard_driveway": ("raspi", ("frontyard", ), "valve", "state_l2", True),
        "valve_frontyard_neighbor": ("raspi", ("frontyard", ), "valve", "state_l3", True),
        "valve_frontyard_planter": ("raspi", ("frontyard", ), "valve", "state_l4", True),
    }.items()
}
with (REPO / "configs" / "switches.json").open("w") as fd:
    json.dump(cfg, fd, sort_keys=True, indent=4)