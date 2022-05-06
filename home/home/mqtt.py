import asyncio
import json
import logging
import re
from typing import Any, Callable

from aioprometheus import Counter
from asyncio_mqtt import Client as Mqtt
from asyncio_mqtt import MqttError
from paho.mqtt.client import MQTTMessage

from home.utils import n_tries
from home.web import WEB

log = logging.getLogger(__name__)

_PROM_ZIGBEE_LOG = Counter(
    "zigbee_logs",
    "Counter of each logs published by zigbee2mqtt",
)


@n_tries(3)
async def mqtt_send(topic: str, message: Any) -> None:
    if isinstance(message, (dict, list)):
        message = json.dumps(message)
    if not isinstance(message, str):
        message = str(message)
    async with Mqtt("192.168.1.1") as mqtt:
        await mqtt.publish(topic, payload=message.encode())


async def watch_mqtt_topic(topic: str, callback: Callable[[MQTTMessage], None]):
    async def _watch_mqtt_topic():
        async with Mqtt("192.168.1.1") as mqtt:
            async with mqtt.filtered_messages(topic) as messages:
                await mqtt.subscribe(topic)
                async for message in messages:
                    await callback(message)

    while True:
        try:
            await _watch_mqtt_topic()
        except MqttError as err:
            log.warning(f"Got an issue with mqtt: {err}")


PROM_LABEL_RX = re.compile(r"[^\w ]")


async def handle_zigbee_error(msg: MQTTMessage) -> None:
    msg = json.loads(msg.payload).get("message")
    _PROM_ZIGBEE_LOG.inc({"msg": PROM_LABEL_RX.sub("", msg)})
    log.error(msg)


def init() -> None:
    @WEB.on_event("startup")
    def _():
        asyncio.create_task(
            watch_mqtt_topic("zigbee2mqtt/bridge/log", handle_zigbee_error)
        )
