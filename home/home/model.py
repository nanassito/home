from abc import ABC, abstractmethod
from typing import Generic, TypeVar


class Entity:
    def __init__(self: "Entity", name: str) -> None:
        self.name = name

    @property
    def mqtt(self: "Entity") -> str:
        return f"zigbee2mqtt/{self.name}/set"

    @property
    def prom_topic(self: "Entity") -> str:
        return f"zigbee2mqtt_{self.name}"


# State must be comparable
_State = TypeVar("_State")


class Actionable(ABC, Generic[_State]):
    """Base interface every actionable object must follow.

    Note the getters are expected to be run approximately once per minute so be
    careful with their run time."""

    @abstractmethod
    async def get_desired_state(self: "Actionable") -> _State:
        """Get the desired state at the time of call.

        This can be cached or serve results from a background task if needed."""
        ...

    @abstractmethod
    async def get_current_state(self: "Actionable") -> _State:
        """Get the actual state at the time of call.

        This should NOT be cached in order to return the most accurate result possible."""
        ...

    @abstractmethod
    async def apply_state(self: "Actionable", state: _State) -> None:
        ...
