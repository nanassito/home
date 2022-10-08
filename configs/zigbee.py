import json
import yaml
from pathlib import Path


REPO =  Path(__file__).resolve().parent.parent
ZIGBEE2MQTT = (REPO / "zigbee2mqtt" / "configuration.yaml")


with (REPO / "configs" / "devices.json").open() as fd:
    data = json.load(fd)


with ZIGBEE2MQTT.open() as fd:
    cfg = yaml.load(fd, yaml.Loader)


cfg["devices"] = {
    xid: {"friendly_name": name}
    for name, xid, in data["zigbee2mqtt"]["server"].items()
    if xid != ""
}

with ZIGBEE2MQTT.open("w") as fd:
    yaml.dump(cfg, fd)
