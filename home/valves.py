import asyncio
import json
from datetime import timedelta
from enum import Enum

from home.mqtt import mqtt_send

import logging


log = logging.getLogger(__name__)


class State(Enum):
    On = "ON"
    Off = "OFF"


class Valve:
    def __init__(self: "Valve", name: str) -> None:
        self._name = name

    @property
    def mqtt_topic(self: "Valve") -> str:
        return f"zigbee2mqtt/valve_{self._name}/set"

    async def _switch(self: "Valve", state: State) -> None:
        await mqtt_send(
            "zigbee2mqtt/valve_backyard_side/set",
            json.dumps({"state": state.value}),
        )
        log.info(f"Valve {self._name} switched {state.value.lower()}.")

    async def switch_on(self: "Valve") -> None:
        await self._switch(State.On)

    async def switch_off(self: "Valve") -> None:
        await self._switch(State.Off)

    async def run_for(self: "Valve", duration: timedelta) -> None:
        await self.switch_on()
        # TODO: Should schedule the shutoff instead of waiting
        await asyncio.sleep(duration.total_seconds())
        await self.switch_off()
