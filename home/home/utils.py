import logging
from typing import Callable, Tuple, TypeVar

from aioprometheus.collectors import Gauge

log = logging.getLogger(__name__)


_T = TypeVar("_T")
_R = TypeVar("_R")


def retry(
    state: _T, decider: Callable[[Exception, _T], Tuple[bool, _T]]
) -> Callable[..., _R]:
    """Retry an coroutine if decider says so."""

    def wrapper(func):
        async def wrapped(*args, **kwargs):
            try_again = True
            _state = state
            while try_again:
                try:
                    return await func(*args, **kwargs)
                except Exception as err:
                    try_again, _state = decider(err, _state)
                    if try_again:
                        log.warning(f"Call to {func} failed with {err}. Retrying.")
                        continue
                    else:
                        log.warning(f"Call to {func} failed with {err}. Failing.")
                        raise

        return wrapped

    return wrapper


def n_tries(n: int):
    return retry(
        n, lambda _err, remaining_tries: (remaining_tries > 1, remaining_tries - 1)
    )


class FeatureFlag:
    _PROM = Gauge("feature_enabled", "0 if a feature is disabled, 1 if it is enabled.")

    def __init__(self: "FeatureFlag", name: str, enabled: bool = True) -> None:
        self.name = name
        self._enabled = enabled

    @property
    def enabled(self: "FeatureFlag") -> bool:
        return self._enabled

    @property
    def disabled(self: "FeatureFlag") -> bool:
        return not self._enabled

    def enable(self: "FeatureFlag") -> None:
        FeatureFlag._PROM.set({"feature": self.name}, 1)
        self._enabled = True

    def disable(self: "FeatureFlag") -> None:
        FeatureFlag._PROM.set({"feature": self.name}, 0)
        self._enabled = False
