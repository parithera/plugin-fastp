# syntax=docker/dockerfile:1

ARG KIND=plugins
ARG USERNAME=codeclarity
ARG PLUGINNAME=plugin

# DEV IMAGE
FROM golang:1.24.0 AS plugin-dev
ARG KIND
ARG PLUGINNAME

WORKDIR /
RUN wget http://opengene.org/fastp/fastp
RUN chmod a+x ./fastp
RUN cp ./fastp /usr/local/bin/fastp

WORKDIR /codeclarity/${KIND}/${PLUGINNAME}
ADD .cloud/docker/config/* /root

RUN go install github.com/air-verse/air@latest
CMD air -c .air.toml

# DEBUG IMAGE
FROM golang:1.24.0 AS plugin-debug
ARG KIND
ARG PLUGINNAME
ENV KIND ${KIND}
ENV PLUGINNAME ${PLUGINNAME}

WORKDIR /
RUN wget http://opengene.org/fastp/fastp
RUN chmod a+x ./fastp
RUN cp ./fastp /usr/local/bin/fastp

WORKDIR /codeclarity/${KIND}/${PLUGINNAME}
ADD .cloud/docker/config/* /root

RUN go install github.com/go-delve/delve/cmd/dlv@latest
ADD .cloud/docker/config/* /root
WORKDIR /codeclarity/${KIND}/${PLUGINNAME}
CMD /go/bin/dlv debug . --headless --listen=:40000 --accept-multiclient
# CMD tail -f /dev/null


# BUILD IMAGE
FROM golang:1.24.0 AS plugin-build
ARG KIND
ARG PLUGINNAME
COPY . /codeclarity/${KIND}/${PLUGINNAME}
WORKDIR /codeclarity/${KIND}/${PLUGINNAME}
ADD .cloud/docker/config/* /root
RUN rm go.work
RUN apt update && apt upgrade -y
RUN apt install -y git
RUN [ "go", "build", "-o", "plugin" ]

# PRODUCTION IMAGE
FROM debian:12.9-slim AS plugin
ARG KIND
ARG USERNAME
ARG PLUGINNAME

RUN apt update && apt upgrade -y
RUN apt install -y wget
RUN wget http://opengene.org/fastp/fastp
RUN chmod a+x ./fastp
RUN cp ./fastp /usr/local/bin/fastp

# Clean cache
RUN apt-get clean
RUN rm -rf /var/lib/apt/lists/*

RUN groupadd -g 1001 ${USERNAME}
RUN adduser --uid 1001 -gid 1001 ${USERNAME}
USER ${USERNAME}

WORKDIR /home/${USERNAME}/${PLUGINNAME}
COPY --from=plugin-build --chown=codeclarity:codeclarity /codeclarity/${KIND}/${PLUGINNAME}/plugin .
COPY --from=plugin-build --chown=codeclarity:codeclarity /codeclarity/${KIND}/${PLUGINNAME}/config.json .

RUN mkdir /home/${USERNAME}/private

ENTRYPOINT [ "./plugin" ]
