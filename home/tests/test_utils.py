import asyncio

import pytest

from home.utils import n_tries


def test_n_tries_enough():
    iterations = []

    @n_tries(3)
    async def test():
        iterations.append(1)
        if len(iterations) < 2:
            raise ValueError()

    asyncio.run(test())
    assert len(iterations) == 2


def test_n_tries_fail():
    iterations = []

    class E(Exception):
        pass

    @n_tries(3)
    async def test():
        iterations.append(1)
        raise E()

    with pytest.raises(E):
        asyncio.run(test())
    assert len(iterations) == 3
