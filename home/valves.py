import json
import logging
from datetime import timedelta
from enum import Enum

from home.model import Actionable, Entity
from home.mqtt import mqtt_send
from home.prometheus import prom_query_one
from home.time import TimeZone, now, today_at
from home.utils import n_tries

log = logging.getLogger(__name__)


class State(Enum):
    On = "ON"
    Off = "OFF"


class Valve(Actionable):
    def __init__(
        self: "Valve",
        area: str,
        line: int,
        duration: timedelta,
        after: "tuple[int, int, TimeZone]",
    ) -> None:
        self.entity = Entity("valve_backyard")
        self.area = area
        self.line = line
        self.duration = duration
        self.after = after

    async def already_ran_today(self: "Valve") -> bool:
        promql = f'max_over_time(mqtt_state_l1{{topic="{self.entity.prom_topic}"}}[12h]) - mqtt_state_l1'
        return await prom_query_one(promql) == 1

    async def get_desired_state(self: "Valve") -> State:
        if self.already_ran_today:
            return State.Off
        after = today_at(*self.after)
        if after <= now() <= after + self.duration:
            return State.On
        else:
            return State.Off

    @n_tries(3)
    async def get_current_state(self: "Valve") -> State:
        promql = f'mqtt_state_l{self.line}{{topic="{self.entity.prom_topic}"}}'
        state = await prom_query_one(promql)
        return {0.0: State.Off, 1.0: State.On}[state]

    async def apply_state(self: "Valve", state: State) -> None:
        await mqtt_send(
            self.entity.mqtt, json.dumps({f"state_l{self.line}": state.value})
        )
        log.info(
            f"Backyard valve near to the {self.area} switched {state.value.lower()}."
        )


VALVE_BACKYARD_SIDE = Valve(
    "side", line=1, duration=timedelta(minutes=10), after=(21, 00, TimeZone.PT)
)
VALVE_BACKYARD_HOUSE = Valve(
    "house", line=2, duration=timedelta(minutes=15), after=(21, 10, TimeZone.PT)
)
VALVE_BACKYARD_SCHOOL = Valve(
    "school", line=3, duration=timedelta(minutes=15), after=(21, 25, TimeZone.PT)
)
VALVE_BACKYARD_DECK = Valve(
    "deck", line=4, duration=timedelta(minutes=15), after=(21, 40, TimeZone.PT)
)

ALL_VALVES = [
    VALVE_BACKYARD_SIDE,
    VALVE_BACKYARD_HOUSE,
    VALVE_BACKYARD_SCHOOL,
    VALVE_BACKYARD_DECK,
]
