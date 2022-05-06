import asyncio
from dataclasses import dataclass
from enum import Enum
import logging

from fastapi import Request
from fastapi.responses import HTMLResponse
from home.mqtt import watch_mqtt_topic, MQTTMessage

from home.prometheus import prom_query_one
from home.web import TEMPLATES, WEB

log = logging.getLogger(__name__)


class Mode(Enum):
    OFF = "off"
    AUTO = "heat_cool"
    COOL = "cool"
    HEAT = "heat"
    FAN = "fan_only"
    DRY = "dry"


class Fan(Enum):
    AUTO = "auto"
    LOW = "low"
    MEDIUM = "medium"
    HIGH = "high"


class Swing(Enum):
    OFF = "off"
    BOTH = "both"
    VERTICAL = "vertical"


@dataclass
class Hvac:
    esp_name: str
    current_temp: float = -1
    target_temp: float = -1
    mode: Mode = Mode.OFF
    fan: Fan = Fan.AUTO
    swing: Swing = Swing.BOTH

    async def on_mqtt(self: "Hvac", msg: MQTTMessage):
        changed = True
        match msg.topic.rsplit("/", 1)[-1]:
            case "mode_state":
                self.mode = Mode(msg.payload.decode())
            case "current_temperature_state":
                self.current_temp = float(msg.payload)
            case "target_temperature_low_state":
                self.target_temp = float(msg.payload)
            case "fan_mode_state":
                self.fan = Fan(msg.payload.decode())
            case "swing_mode_state":
                self.swing = Swing(msg.payload.decode())
            case _:
                changed = False
        if changed:
            log.debug(self)


@dataclass
class Room:
    name: str
    sensor_topic: str
    hvacs: list[Hvac]
    min_temp: int = 19
    max_temp: int = 28


ALL_ROOMS = [
    Room("Zaya", "zigbee2mqtt_air_zaya", [Hvac("zaya")]),
    Room("Parent", "zigbee2mqtt_air_parent", [Hvac("parent")]),
    Room("Salon", "zigbee2mqtt_air_livingroom", [Hvac("living"), Hvac("kitchen")]),
    Room("Office", "zigbee2mqtt_air_office", [Hvac("office")]),
    Room("Outside", "zigbee2mqtt_air_outside", []),
]


async def hvac_controller():
    pass


def init():
    @WEB.on_event("startup")
    def _():
        for room in ALL_ROOMS:
            for hvac in room.hvacs:
                asyncio.create_task(
                    watch_mqtt_topic(f"esphome/{hvac.esp_name}/+", hvac.on_mqtt)
                )

    @WEB.get("/temperature", response_class=HTMLResponse)
    async def get_soaker(request: Request):
        async def get_temp(promql: str) -> float | str:
            try:
                return round(await prom_query_one(promql), 1)
            except Exception as err:
                log.error(err)
                return "--.-"

        return TEMPLATES.TemplateResponse(
            "temperature.html.jinja",
            {
                "request": request,
                "page": "Temperature",
                "rooms": [
                    {
                        "name": room,
                        "current": await get_temp(
                            f'mqtt_temperature{{topic="{topic}"}}'
                        ),
                        "min_1d": await get_temp(
                            f'min_over_time(mqtt_temperature{{topic="{topic}"}}[1d])'
                        ),
                        "max_1d": await get_temp(
                            f'max_over_time(mqtt_temperature{{topic="{topic}"}}[1d])'
                        ),
                        "link": f'https://prometheus.epa.jaminais.fr/graph?g0.expr=mqtt_temperature{{topic%3D"{topic}"}}&g0.tab=0&g0.range_input=1d',
                    }
                    for room, topic in {
                        "Zaya": "zigbee2mqtt_air_zaya",
                        "Parent": "zigbee2mqtt_air_parent",
                        "Salon": "zigbee2mqtt_air_livingroom",
                        "Office": "zigbee2mqtt_air_office",
                        "Outside": "zigbee2mqtt_air_outside",
                    }.items()
                ],
            },
        )
