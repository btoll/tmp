FROM ubuntu

ENV DEBIAN_FRONTEND noninteractive
RUN apt-get update \
    && apt-get install build-essential dnsutils -y

ENV HOME /home/test
RUN useradd --create-home --home-dir $HOME test

COPY gotest /usr/bin/

USER test

ENTRYPOINT ["gotest"]

