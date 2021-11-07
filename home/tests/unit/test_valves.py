from datetime import timedelta

import pytest
from home.time import now
from home.valves import State, Valve as _Valve


class Valve(_Valve):
    def __init__(self, already_ran_today, after):
        super().__init__("area", 42, timedelta(seconds=5), after)
        self._already_ran_today = already_ran_today

    async def already_ran_today(self: "Valve") -> bool:
        return self._already_ran_today


@pytest.mark.parametrize(
    ("after", "already_ran_today", "expected"),
    [
        (now() + timedelta(seconds=5), False, State.Off),  # too soon
        (now() - timedelta(seconds=1), True, State.Off),  # already ran
        (now() - timedelta(seconds=1), False, State.On),  # need to run
        (now() + timedelta(seconds=6), False, State.Off),  # completed
    ],
)
@pytest.mark.asyncio
async def test_get_desired_state(after, already_ran_today, expected):
    valve = Valve(already_ran_today, after)
    await valve.get_desired_state() == expected
