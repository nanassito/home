import pytest

from home.prometheus import prom_query_labels, prom_query_one


@pytest.mark.asyncio
async def test_prom_query_one():
    promql = "mqtt_config_advanced_channel"
    assert await prom_query_one(promql) == 11


@pytest.mark.asyncio
async def test_prom_query_labels():
    promql = '{topic="zigbee2mqtt_valve_backyard", __name__=~"mqtt_state_l[0-9]"}'
    labels = await prom_query_labels(promql)
    assert len(labels) == 4
    assert {
        "__name__": "mqtt_state_l4",
        "instance": "172.17.0.1:9000",
        "job": "zigbee",
        "topic": "zigbee2mqtt_valve_backyard",
    } in labels
