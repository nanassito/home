import asyncio
from datetime import timedelta
from unittest.mock import AsyncMock, patch

import pytest

from home.time import TimeZone
from home.valves import State, Valve


@pytest.mark.parametrize(
    ("net", "run_time", "other_running", "expected"),
    [
        ((23, 59, TimeZone.PT), 3, [], State.Off),  # too soon
        ((00, 00, TimeZone.PT), 0, [], State.On),  # need to run
        ((00, 00, TimeZone.PT), 3, [], State.On),  # keep running
        ((00, 00, TimeZone.PT), 5, [], State.Off),  # completed
        ((00, 00, TimeZone.PT), 3, [{"id": "other"}], State.Off),  # Other running
    ],
)
@pytest.mark.asyncio
async def test_get_desired_state(net, run_time, other_running, expected):
    valve = Valve("area", 42, timedelta(minutes=5), net)
    # Can't use py3.10 parenthesized context manager due to https://github.com/psf/black/issues/1948
    with patch("home.valves.prom_query_one", new_callable=AsyncMock) as prom_query_one:
        with patch(
            "home.valves.prom_query_labels", new_callable=AsyncMock
        ) as prom_query_labels:
            prom_query_one.return_value = run_time
            prom_query_labels.return_value = other_running
            assert await valve.get_desired_state() == expected
