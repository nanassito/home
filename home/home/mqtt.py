import asyncio
import json
import logging
from typing import Callable

from asyncio_mqtt import Client as Mqtt
from asyncio_mqtt import MqttError
from home.home import web

from home.utils import n_tries

log = logging.getLogger(__name__)


@n_tries(3)
async def mqtt_send(topic: str, message: dict | str) -> None:
    if isinstance(message, dict):
        message = json.dumps(message)
    async with Mqtt("192.168.1.1") as mqtt:
        await mqtt.publish(topic, payload=message.encode())


async def watch_mqtt_topic(topic: str, callback: Callable[[bytes], None]):
    async def _watch_mqtt_topic():
        async with Mqtt("192.168.1.1") as mqtt:
            async with mqtt.filtered_messages(topic) as messages:
                await mqtt.subscribe(topic)
                async for message in messages:
                    await callback(message.payload)

    while True:
        try:
            await _watch_mqtt_topic()
        except MqttError as err:
            log.warning(f"Got an issue with mqtt: {err}")


async def handle_zigbee_error(message: str) -> None:
    log.error(message)


def init() -> None:
    @web.on_event("startup")
    def _():
        asyncio.create_task(
            watch_mqtt_topic("zigbee2mqtt/bridge/log", handle_zigbee_error)
        )
