import pytest

from home.utils import n_tries


@pytest.mark.asyncio
async def test_n_tries_enough():
    iterations = []

    @n_tries(3)
    async def test():
        iterations.append(1)
        if len(iterations) < 2:
            raise ValueError()

    await test()
    assert len(iterations) == 2


@pytest.mark.asyncio
async def test_n_tries_fail():
    iterations = []

    class E(Exception):
        pass

    @n_tries(3)
    async def test():
        iterations.append(1)
        raise E()

    with pytest.raises(E):
        await test()
    assert len(iterations) == 3
