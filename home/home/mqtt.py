from asyncio_mqtt import Client as Mqtt

from home.utils import n_tries


@n_tries(3)
async def mqtt_send(topic: str, message: str) -> None:
    async with Mqtt("192.168.1.1") as mqtt:
        await mqtt.publish(topic, payload=message.encode())