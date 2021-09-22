FROM python:3.9-slim  AS build-env

WORKDIR /
ADD assert.py ./

RUN pip3 install --upgrade pip && \
    pip install requests

FROM gcr.io/distroless/python3-debian10
COPY --from=build-env assert.py ./
COPY --from=build-env /usr/local/lib/python3.9/site-packages /usr/local/lib/python3.9/site-packages

WORKDIR /
ENV PYTHONPATH=/usr/local/lib/python3.9/site-packages
CMD ["assert.py"]

