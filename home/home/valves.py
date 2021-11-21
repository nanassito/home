import json
import logging
from asyncio import Lock
from datetime import timedelta
from enum import Enum
from aioprometheus import Gauge

from home.model import Actionable, Entity
from home.mqtt import mqtt_send
from home.prometheus import prom_query_labels, prom_query_one
from home.time import TimeZone, now, today_at
from home.utils import n_tries
from home import facts

log = logging.getLogger(__name__)




class State(Enum):
    On = "ON"
    Off = "OFF"


_VALVE_DESIRED_STATE_GAUGE = Gauge(
    "valve_desired_state", "Desired state of a valve relay"
)
_VALVE_CURRENT_STATE_GAUGE = Gauge(
    "valve_current_state", "Observed state of a valve relay"
)


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
        self.counter_labels: dict[str, str] = {
            "topic": self.entity.prom_topic,
            "line": str(self.line),
            "area": self.area,
        }

    async def get_desired_state(self: "Valve") -> State:
        async def _get_desired_state() -> State:
            if now() < today_at(*self.not_earlier_than):
                return State.Off

            if await facts.has_other_water_users(self.area):
                return State.Off

            promql = f'{{ {self.prom_labels}, __name__=~"mqtt_state_l[0-9]", __name__!="{self.prom_metric}" }} == 1'
            other_running = await prom_query_labels(promql)
            if other_running:
                self.log.warning(f"Found other running valves: {other_running}")
                return State.Off

            promql = f"sum_over_time({self.prom_metric}{{{self.prom_labels}}}[70h])"
            run_time = timedelta(minutes=await prom_query_one(promql))
            return {False: State.Off, True: State.On}[run_time < self.duration]

        state = await _get_desired_state()
        _VALVE_DESIRED_STATE_GAUGE.set(
            self.counter_labels, {State.Off: 0, State.On: 1}[state]
        )
        return state

    @n_tries(3)
    async def get_current_state(self: "Valve") -> State:
        promql = f"{self.prom_metric}{{{self.prom_labels}}}"
        state = await prom_query_one(promql)
        _VALVE_CURRENT_STATE_GAUGE.set(self.counter_labels, int(state))
        return {0.0: State.Off, 1.0: State.On}[state]

    async def apply_state(self: "Valve", state: State) -> None:
        if state == State.On:
            await facts.start_using_water(self.area)
        else:
            await facts.stop_using_water(self.area)

        await mqtt_send(
            self.entity.mqtt, json.dumps({f"state_l{self.line}": state.value})
        )
        self.log.info(f"Switched {state.value.lower()}.")


_NET = (18, 00, TimeZone.PT)
VALVE_BACKYARD_SIDE = Valve(
    "side", line=1, duration=timedelta(minutes=10), not_earlier_than=_NET
)
VALVE_BACKYARD_HOUSE = Valve(
    "house", line=2, duration=timedelta(minutes=15), not_earlier_than=_NET
)
VALVE_BACKYARD_SCHOOL = Valve(
    "school", line=3, duration=timedelta(minutes=15), not_earlier_than=_NET
)
VALVE_BACKYARD_DECK = Valve(
    "deck", line=4, duration=timedelta(minutes=15), not_earlier_than=_NET
)

ALL_VALVES = [
    VALVE_BACKYARD_SIDE,
    VALVE_BACKYARD_HOUSE,
    VALVE_BACKYARD_SCHOOL,
    VALVE_BACKYARD_DECK,
]
