import json
import logging
from datetime import timedelta
from enum import Enum

from home.model import Actionable, Entity
from home.mqtt import mqtt_send
from home.prometheus import prom_query_labels, prom_query_one
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
        not_earlier_than: "tuple[int, int, TimeZone]",
    ) -> None:
        self.area = area
        self.line = line
        self.duration = duration
        self.not_earlier_than = not_earlier_than

        self.log = log.getChild(self.__class__.__name__).getChild(self.area)
        self.entity = Entity("valve_backyard")
        self.prom_metric = f"mqtt_state_l{self.line}"
        self.prom_labels = f'topic="{self.entity.prom_topic}"'

    async def get_desired_state(self: "Valve") -> State:
        if now() < today_at(*self.not_earlier_than):
            return State.Off

        promql = f'{{ {self.prom_labels}, __name__=~"mqtt_state_l[0-9]", __name__!="{self.prom_metric}" }} == 1'
        other_running = await prom_query_labels(promql)
        if other_running:
            self.log.warning(f"Found other running valves: {other_running}")
            return State.Off

        promql = f"sum_over_time({self.prom_metric}{{{self.prom_labels}}}[20h])"
        run_time = timedelta(minutes=await prom_query_one(promql))
        return {False: State.Off, True: State.On}[run_time < self.duration]

    @n_tries(3)
    async def get_current_state(self: "Valve") -> State:
        promql = f"{self.prom_metric}{{{self.prom_labels}}}"
        state = await prom_query_one(promql)
        return {0.0: State.Off, 1.0: State.On}[state]

    async def apply_state(self: "Valve", state: State) -> None:
        await mqtt_send(
            self.entity.mqtt, json.dumps({f"state_l{self.line}": state.value})
        )
        self.log.info(f"Switched {state.value.lower()}.")


VALVE_BACKYARD_SIDE = Valve(
    "side",
    line=1,
    duration=timedelta(minutes=10),
    not_earlier_than=(20, 00, TimeZone.PT),
)
VALVE_BACKYARD_HOUSE = Valve(
    "house",
    line=2,
    duration=timedelta(minutes=15),
    not_earlier_than=(20, 10, TimeZone.PT),
)
VALVE_BACKYARD_SCHOOL = Valve(
    "school",
    line=3,
    duration=timedelta(minutes=15),
    not_earlier_than=(20, 25, TimeZone.PT),
)
VALVE_BACKYARD_DECK = Valve(
    "deck",
    line=4,
    duration=timedelta(minutes=15),
    not_earlier_than=(20, 40, TimeZone.PT),
)

ALL_VALVES = [
    VALVE_BACKYARD_SIDE,
    VALVE_BACKYARD_HOUSE,
    VALVE_BACKYARD_SCHOOL,
    VALVE_BACKYARD_DECK,
]
