FROM python:3.10

WORKDIR /app

ADD home /app/home
ADD poetry.lock pyproject.toml /app/

RUN pip install poetry && pwd && ls && poetry install --no-dev

CMD poetry run python -m home