import logging
from typing import Callable, Tuple, TypeVar

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
    return retry(n, lambda _err, state: (state > 1, state - 1))
