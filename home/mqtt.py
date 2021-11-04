from asyncio_mqtt import Client as Mqtt


async def mqtt_send(topic: str, message: str) -> None:
    async with Mqtt("192.168.1.1") as mqtt:
        message = "10%"
        await mqtt.publish(topic, payload=message.encode())
