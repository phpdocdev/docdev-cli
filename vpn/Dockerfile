FROM nubacuk/docker-openvpn:aarch64

RUN apk update && apk add socat

COPY . /

RUN chmod +x /docker-entrypoint.sh

ENTRYPOINT ["/docker-entrypoint.sh"]