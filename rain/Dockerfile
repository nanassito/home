FROM python:3.10

WORKDIR /app

ADD poetry.lock pyproject.toml /app/
RUN pip install poetry && poetry install --no-dev --no-root

ADD rain /app/rain
RUN poetry install --no-dev

CMD poetry run python -m rain