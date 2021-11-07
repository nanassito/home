import pytest
from home.prometheus import prom_query_one


@pytest.mark.asyncio
async def test_fetch_something():
    promql = "mqtt_config_advanced_channel"
    assert await prom_query_one(promql) == 11
