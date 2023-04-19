FROM alpine:latest

RUN apk update && \
    apk add --no-cache git github-cli yq docker

COPY entrypoint.sh /home/entrypoint.sh
RUN chmod +x /home/entrypoint.sh

ENTRYPOINT ["/home/entrypoint.sh"]