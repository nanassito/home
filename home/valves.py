import asyncio
import json
import logging
from abc import ABC, abstractmethod
from datetime import timedelta
from enum import Enum

from home.mqtt import mqtt_send

log = logging.getLogger(__name__)


class State(Enum):
    On = "ON"
    Off = "OFF"


class Valve(ABC):
    def __init__(self, base_duration: timedelta) -> None:
        self.base_duration = base_duration

    @abstractmethod
    async def switch(self: "Valve", state: State) -> None:
        ...

    async def switch_on(self: "Valve") -> None:
        await self.switch(State.On)

    async def switch_off(self: "Valve") -> None:
        await self.switch(State.Off)

    async def run(self: "Valve", duration: timedelta = None) -> None:
        duration = duration or self.base_duration
        await self.switch_on()
        # TODO: Should schedule the shutoff instead of waiting
        await asyncio.sleep(duration.total_seconds())
        await self.switch_off()


class BackyardValve(Valve):
    def __init__(self: "BackyardValve", area: str, line: int, base_duration: timedelta) -> None:
        super().__init__(base_duration)
        self._area = area
        self._line = line

    async def switch(self: "BackyardValve", state: State) -> None:
        await mqtt_send(
            "zigbee2mqtt/valve_backyard/set",
            json.dumps({f"state_l{self._line}": state.value}),
        )
        log.info(
            f"Backyard valve close to the {self._area} switched {state.value.lower()}."
        )


VALVE_BACKYARD_SIDE = BackyardValve("side", 1, timedelta(minutes=10))
VALVE_BACKYARD_HOUSE = BackyardValve("house", 2, timedelta(minutes=15))
VALVE_BACKYARD_SCHOOL = BackyardValve("school", 3, timedelta(minutes=15))
VALVE_BACKYARD_DECK = BackyardValve("deck", 4, timedelta(minutes=15))

ALL_VALVES = [
    VALVE_BACKYARD_SIDE,
    VALVE_BACKYARD_HOUSE,
    VALVE_BACKYARD_SCHOOL,
    # VALVE_BACKYARD_DECK,
]
