FROM python:slim-bullseye

ENV HOME /home/test
RUN useradd --create-home --home-dir $HOME test

RUN apt-get update \
    && apt-get install dnsutils -y

WORKDIR $HOME

COPY assert.py ./
RUN chown -R test:test $HOME

USER test

RUN python -m pip install requests --no-warn-script-location

CMD ["python", "assert.py"]

