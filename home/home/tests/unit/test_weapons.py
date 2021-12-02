import json
from datetime import timedelta
from unittest.mock import AsyncMock, patch

import pytest
from home.valves import Valve
from home.weapons import Soaker


class MockValve(Valve):
    def __init__(self, state):
        super().__init__("mock", 0)
        self.state = self.should_be_running = state
        self.change_history = []

    async def is_running(self):
        return self.state

    async def switch_off(self):
        self.state = self.should_be_running = False
        self.change_history.append(False)

    async def switch_on(self):
        self.state = self.should_be_running = True
        self.change_history.append(True)


@pytest.mark.asyncio
@pytest.mark.parametrize(
    ("message", "change_history", "was_running"),
    [
        ("Should remain open", [True], True),
        ("Should be closed", [True, False], False),
    ]
)
async def test_soaker_reverts_valves(message, change_history, was_running):
    valve = MockValve(was_running)
    soaker = Soaker(valve)
    soaker.duration = timedelta(seconds=0)
    await soaker.soak('{"occupancy": true}')
    assert valve.change_history == change_history, message


@pytest.mark.asyncio
@pytest.mark.parametrize(
    ("message", "runs", "occupancy", "mower_running"),
    [
        ("Don't run because no occupancy", False, False, False),
        ("Don't run because no occupancy and mower running", False, False, True),
        ("Run because occupancy and mower not running", True, True, False),
        ("Don't run because mower running", False, True, True),
    ],
)
async def test_soaker_runs(message, runs, occupancy, mower_running):
    valve = MockValve(False)
    soaker = Soaker(valve)
    soaker.duration = timedelta(seconds=0)
    with patch("home.weapons.facts.is_mower_running", new_callable=AsyncMock) as is_mower_running:
        is_mower_running.return_value = mower_running
        await soaker.soak(json.dumps({"occupancy": occupancy}))
    assert bool(valve.change_history) == runs, message
