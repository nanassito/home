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
async def test_soaker_reverts_valves():
    was_running = MockValve(True)
    was_not_running = MockValve(False)

    class MockSoaker(Soaker):
        VALVES = [was_running, was_not_running]
        DURATION = timedelta(seconds=0)

    await MockSoaker.soak('{"occupancy": true}')
    assert was_running.change_history == [True], "This valve should remain open"
    assert was_not_running.change_history == [
        True,
        False,
    ], "This valve should have been closed"


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

    class MockSoaker(Soaker):
        VALVES = [valve]
        DURATION = timedelta(seconds=0)

    with patch("home.weapons.facts.is_mower_running", new_callable=AsyncMock) as is_mower_running:
        is_mower_running.return_value = mower_running
        await MockSoaker.soak(json.dumps({"occupancy": occupancy}))
    assert bool(valve.change_history) == runs, message
