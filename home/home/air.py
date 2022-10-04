import asyncio
import logging
from dataclasses import dataclass, field
from datetime import datetime, timedelta
from enum import Enum
from math import inf

from aioprometheus.collectors import Gauge
from fastapi import HTTPException, Request
from fastapi.responses import HTMLResponse
from pydantic import BaseModel

from home.facts import get_outside_temp
from home.mqtt import MQTTMessage, mqtt_send, watch_mqtt_topic
from home.prometheus import prom_query_one
from home.time import now
from home.web import TEMPLATES, WEB

log = logging.getLogger(__name__)


class Mode(Enum):
    INVALID = None
    OFF = "off"
    AUTO = "heat_cool"
    COOL = "cool"
    HEAT = "heat"
    FAN = "fan_only"
    DRY = "dry"


MODE_TO_INT = {k: v for (v, k) in enumerate(Mode)}
PROM_GAUGE_HVAC_STATE = Gauge(
    "hvac_reported_state", f"Reported state of the HVAC unit. {MODE_TO_INT}"
)
PROM_GAUGE_HVAC_FAN = Gauge("hvac_reported_fan", "reported fan of each HVAC unit.")
PROM_GAUGE_HVAC_MODE = Gauge("hvac_reported_mode", "reported mode of each HVAC unit.")


class Fan(Enum):
    INVALID = None
    AUTO = "auto"
    LOW = "low"
    MEDIUM = "medium"
    HIGH = "high"


# class Swing(Enum):
#     OFF = "off"
#     BOTH = "both"
#     VERTICAL = "vertical"


@dataclass
class HvacState:
    target_temp: int = -1
    mode: Mode = Mode.INVALID
    fan: Fan = Fan.INVALID


class HvacControl(Enum):
    AUTO = "auto"
    APP = "app"
    REMOTE = "remote"


@dataclass
class Hvac:
    esp_name: str
    reported_state: HvacState = field(default_factory=HvacState)
    desired_state: HvacState = field(default_factory=HvacState)
    last_command: datetime = field(default=now(), repr=False)
    log: logging.Logger = field(init=False, repr=False)
    control: HvacControl = HvacControl.AUTO

    def __post_init__(self: "Hvac") -> None:
        self.log = log.getChild("Hvac").getChild(self.esp_name)

    @property
    def esp_topic(self: "Room") -> str:
        return f"esphome_{self.esp_name}"

    async def get_current_temp(self: "Room") -> float:
        return await prom_query_one(
            f'mqtt_current_temperature_state{{topic="{self.esp_topic}"}}'
        )

    async def on_mqtt(self: "Hvac", msg: MQTTMessage):
        changed = True
        match msg.topic.rsplit("/", 1)[-1]:
            case "mode_state":
                mode = self.reported_state.mode = Mode(msg.payload.decode())
                PROM_GAUGE_HVAC_STATE.set(
                    {"unit": self.esp_name}, MODE_TO_INT.get(mode, -1)
                )
                for m in Mode:
                    PROM_GAUGE_HVAC_MODE.set(
                        {"unit": self.esp_name, "mode": m},
                        bool(mode == m),
                    )
            case "target_temperature_low_state":
                self.reported_state.target_temp = int(float(msg.payload))
            case "fan_mode_state":
                fan = self.reported_state.fan = Fan(msg.payload.decode())
                for f in Fan:
                    PROM_GAUGE_HVAC_FAN.set(
                        {"unit": self.esp_name, "fan": f},
                        bool(fan == f),
                    )
            case _:
                changed = False
        if changed:
            self.log.debug(self)

    async def protect_against_uart(self: "Hvac") -> None:
        delay = self.last_command + timedelta(milliseconds=500) - now()
        await asyncio.sleep(delay.total_seconds())
        self.last_command = now()

    async def enforce_mode(self: "Hvac") -> None:
        mode = self.desired_state.mode
        if mode != self.reported_state.mode and mode is not Mode.INVALID:
            await mqtt_send(f"esphome/{self.esp_name}/mode_command", mode.value)
            await self.protect_against_uart()

    async def enforce_fan(self: "Hvac") -> None:
        fan = self.desired_state.fan
        if fan != self.reported_state.fan and fan is not Fan.INVALID:
            await mqtt_send(f"esphome/{self.esp_name}/fan_mode_command", fan.value)
            await self.protect_against_uart()

    async def enforce_temp(self: "Hvac") -> None:
        temp = self.desired_state.target_temp
        if temp != self.reported_state.target_temp and temp != -1:
            await mqtt_send(f"esphome/{self.esp_name}/target_temperature_command", temp)
            await self.protect_against_uart()

    async def control_loop(self: "Hvac") -> None:
        while True:
            await asyncio.sleep(1)
            if self.control is HvacControl.REMOTE:
                continue
            await self.enforce_mode()
            await self.enforce_temp()
            await self.enforce_fan()


@dataclass
class Room:
    name: str
    sensor_topic: str
    hvacs: list[Hvac]
    min_temp: int = 19
    max_temp: int = 33

    async def get_current_temp(self: "Room") -> float:
        return await prom_query_one(f'mqtt_temperature{{topic="{self.sensor_topic}"}}')


ALL_ROOMS = [
    Room("Zaya", "zigbee2mqtt_air_zaya", [Hvac("zaya")]),
    Room("Parent", "zigbee2mqtt_air_parent", [Hvac("parent")]),
    Room("Salon", "zigbee2mqtt_air_livingroom", [Hvac("living"), Hvac("kitchen")]),
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
        if curr >= max(room.max_temp - 5, room.min_temp):
            desired_temp_delta -= curr - room.max_temp + 3
    if desired_temp_delta > 0:
        return Mode.HEAT
    elif desired_temp_delta < 0:
        return Mode.COOL
    else:
        return Mode.OFF


class HvacController:
    async def run():
        while True:
            await asyncio.sleep(60)
            outside_temp = await get_outside_temp()
            mode = await infer_general_mode()
            for room in ALL_ROOMS:
                for hvac in room.hvacs:
                    hvac_curr = await hvac.get_current_temp()
                    if hvac.control is HvacControl.AUTO:
                        # Set the running mode
                        curr = await room.get_current_temp()
                        if mode == mode.HEAT and room.min_temp + 3 <= curr:
                            hvac.desired_state.mode = Mode.OFF  # Room is warm enough
                        elif mode == mode.COOL and room.max_temp - 5 >= min(
                            curr, hvac_curr
                        ):
                            hvac.desired_state.mode = Mode.OFF  # Room is cold enough
                        else:
                            hvac.desired_state.mode = (
                                mode  # Apply whatever the majority needs
                            )

                        # Set the temperature target
                        delta_temp = abs(curr - hvac_curr)
                        if mode is Mode.HEAT:
                            hvac.desired_state.target_temp = room.min_temp
                            if delta_temp > 3:
                                hvac.desired_state.fan = Fan.HIGH
                            elif delta_temp > 1.5:
                                hvac.desired_state.fan = Fan.MEDIUM
                            else:
                                hvac.desired_state.fan = Fan.AUTO
                        if mode is Mode.COOL:
                            if outside_temp >= room.max_temp + 5:
                                offset = 2
                                hvac.desired_state.fan = Fan.MEDIUM
                            else:
                                offset = 0
                                hvac.desired_state.fan = Fan.AUTO
                            hvac.desired_state.target_temp = room.max_temp + offset


class _HttpRoomRequest(BaseModel):
    room: str
    min_temp: int
    max_temp: int


class _HttpHvacRemoteRequest(BaseModel):
    hvac: str


class _HttpHvacAppRequest(BaseModel):
    hvac: str
    mode: Mode
    fan: Fan
    target_temp: int


def init():
    @WEB.on_event("startup")
    def _():
        for room in ALL_ROOMS:
            for hvac in room.hvacs:
                asyncio.create_task(
                    watch_mqtt_topic(f"esphome/{hvac.esp_name}/+", hvac.on_mqtt)
                )
                asyncio.create_task(hvac.control_loop())
        asyncio.create_task(HvacController.run())

    @WEB.get("/temperature", response_class=HTMLResponse)
    async def get_temperature(request: Request):
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
                        "name": room.name,
                        "current": await get_temp(
                            f'mqtt_temperature{{topic="{room.sensor_topic}"}}'
                        ),
                        "min_1d": await get_temp(
                            f'min(min_over_time(mqtt_temperature{{topic="{room.sensor_topic}"}}[1d]))'
                        ),
                        "max_1d": await get_temp(
                            f'max(max_over_time(mqtt_temperature{{topic="{room.sensor_topic}"}}[1d]))'
                        ),
                        "min_temp": room.min_temp,
                        "max_temp": room.max_temp,
                        "link": f"https://prometheus.epa.jaminais.fr/graph?"
                        + "&".join(
                            [
                                f"g{idx}.expr=sum by (metric)("
                                f'label_replace(mqtt_temperature{{topic%3D"{room.sensor_topic}"}},"metric","Actual","","")'
                                " or "
                                f'label_replace(mqtt_current_temperature_state{{topic%3D"{hvac.esp_topic}"}},"metric","Reported","","")'
                                " or "
                                f'label_replace(mqtt_target_temperature_low_state{{topic%3D"{hvac.esp_topic}"}},"metric","Target","","")'
                                f")&g{idx}.tab=0&g{idx}.range_input=6h"
                                for idx, hvac in enumerate(room.hvacs)
                            ]
                        ),
                        "hvacs": room.hvacs,
                    }
                    for room in ALL_ROOMS
                ],
            },
        )

    @WEB.post("/api/room")
    async def http_room(rq: _HttpRoomRequest):
        for room in ALL_ROOMS:
            if room.name == rq.room:
                for hvac in room.hvacs:
                    hvac.control = HvacControl.AUTO
                room.min_temp = rq.min_temp
                room.max_temp = rq.max_temp
                return
        return HTTPException(400, f"No room named {rq.room}.")

    @WEB.post("/api/hvac/app")
    async def http_hvac_app(rq: _HttpHvacAppRequest):
        for room in ALL_ROOMS:
            for hvac in room.hvacs:
                if hvac.esp_name == rq.hvac:
                    hvac.control = HvacControl.APP
                    hvac.desired_state.mode = rq.mode
                    hvac.desired_state.fan = rq.fan
                    hvac.desired_state.target_temp = rq.target_temp
                    return
        return HTTPException(400, f"No hvac named {rq.hvac}.")

    @WEB.post("/api/hvac/remote")
    async def http_hvac_remote(rq: _HttpHvacRemoteRequest):
        for room in ALL_ROOMS:
            for hvac in room.hvacs:
                if hvac.esp_name == rq.hvac:
                    hvac.control = HvacControl.REMOTE
                    return
        return HTTPException(400, f"No hvac named {rq.hvac}.")
