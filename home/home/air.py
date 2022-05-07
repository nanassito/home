import asyncio
from dataclasses import dataclass, field
from datetime import datetime, timedelta
from enum import Enum
import logging
from math import inf

from fastapi import Request
from fastapi.responses import HTMLResponse
from home.mqtt import mqtt_send, watch_mqtt_topic, MQTTMessage

from home.prometheus import prom_query_one
from home.time import now
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


# class Swing(Enum):
#     OFF = "off"
#     BOTH = "both"
#     VERTICAL = "vertical"


@dataclass
class Hvac:
    esp_name: str
    current_temp: float = -1
    target_temp: int = -1
    mode: Mode = Mode.OFF
    fan: Fan = Fan.AUTO
    last_command: datetime = field(default=now(), repr=False)
    log: logging.Logger = field(init=False, repr=False)

    def __post_init__(self: "Hvac") -> None:
        self.log = log.getChild("Hvac").getChild(self.esp_name)

    async def on_mqtt(self: "Hvac", msg: MQTTMessage):
        changed = True
        match msg.topic.rsplit("/", 1)[-1]:
            case "mode_state":
                self.mode = Mode(msg.payload.decode())
            case "current_temperature_state":
                self.current_temp = float(msg.payload)
            case "target_temperature_low_state":
                self.target_temp = int(float(msg.payload))
            case "fan_mode_state":
                self.fan = Fan(msg.payload.decode())
            case _:
                changed = False
        if changed:
            self.log.debug(self)
    
    async def protect_against_uart(self: "Hvac") -> None:
        delay = self.last_command + timedelta(milliseconds=500) - now()
        await asyncio.sleep(delay.total_seconds())
        self.last_command = now()
    
    async def set_mode(self: "Hvac", mode: Mode) -> None:
        if self.mode != mode:
            self.log.info(f"Mode {self.mode} > {mode}")
            await mqtt_send(f"esphome/{self.esp_name}/mode_command", mode.value)
            await self.protect_against_uart()
    
    async def set_fan(self: "Hvac", fan: Fan) -> None:
        if self.fan != fan:
            self.log.info(f"Fan {self.fan} > {fan}")
            await mqtt_send(f"esphome/{self.esp_name}/fan_mode_command", fan.value)
            await self.protect_against_uart()
    
    async def set_temp(self: "Hvac", temp: int) -> None:
        if self.target_temp != temp:
            self.log.info(f"Target temperature {self.target_temp} > {temp}")
            await mqtt_send(f"esphome/{self.esp_name}/target_temperature_command", temp)
            await self.protect_against_uart()


@dataclass
class Room:
    name: str
    sensor_topic: str
    hvacs: list[Hvac]
    min_temp: int = 19
    max_temp: int = 28

    async def get_current_temp(self: "Room") -> float:
        return await prom_query_one(f'mqtt_temperature{{topic="{self.sensor_topic}"}}')


ALL_ROOMS = [
    Room("Zaya", "zigbee2mqtt_air_zaya", [Hvac("zaya")]),
    Room("Parent", "zigbee2mqtt_air_parent", [Hvac("parent")]),
    Room("Salon", "zigbee2mqtt_air_livingroom", [Hvac("living"), Hvac("kitchen")], min_temp=21),
    Room("Office", "zigbee2mqtt_air_office", [Hvac("office")]),
    Room("Outside", "zigbee2mqtt_air_outside", []),
]


async def infer_general_mode():
    desired_temp_delta = 0
    for room in ALL_ROOMS:
        if not room.hvacs:
            continue
        curr = await room.get_current_temp()
        if curr < room.min_temp:
            desired_temp_delta += inf
        if curr <= min(room.min_temp + 3, room.max_temp):
            desired_temp_delta += room.min_temp + 3 - curr
        if curr > room.max_temp:
            desired_temp_delta -= inf
        if curr >= max(room.max_temp - 3, room.min_temp):
            desired_temp_delta += curr - room.max_temp + 3
    if desired_temp_delta > 0:
        return Mode.HEAT
    elif desired_temp_delta < 0:
        return Mode.COOL
    else:
        return Mode.OFF


async def hvac_controller():
    # TODO: Add support for disabling
    # TODO: Add support for disabling a specific unit
    while True:
        await asyncio.sleep(60)
        mode = await infer_general_mode()
        for room in ALL_ROOMS:
            for hvac in room.hvacs:
                # Set the running mode
                curr = await room.get_current_temp()
                if mode == mode.HEAT and room.min_temp + 3 <= curr:
                    await hvac.set_mode(Mode.OFF)  # Room is warm enough
                elif mode == mode.COOL and room.max_temp - 3 >= curr:
                    await hvac.set_mode(Mode.OFF)  # Room is cold enough
                else:
                    await hvac.set_mode(mode)  # Apply whatever the majority needs

                # Set the temperature target
                if mode is Mode.HEAT:
                    await hvac.set_temp(room.min_temp)
                if mode is Mode.COOL:
                    await hvac.set_temp(room.max_temp)
                # Set the fan speed
                delta_temp = abs(await room.get_current_temp() - hvac.current_temp)
                if delta_temp > 4:
                    await hvac.set_fan(Fan.HIGH)
                elif delta_temp > 2:
                    await hvac.set_fan(Fan.MEDIUM)
                else:
                    await hvac.set_fan(Fan.AUTO)


def init():
    @WEB.on_event("startup")
    def _():
        for room in ALL_ROOMS:
            for hvac in room.hvacs:
                asyncio.create_task(
                    watch_mqtt_topic(f"esphome/{hvac.esp_name}/+", hvac.on_mqtt)
                )
        asyncio.create_task(hvac_controller())

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
