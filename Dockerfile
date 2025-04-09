FROM alpine:3.21
ARG TARGETARCH

RUN apk --no-cache add ca-certificates

WORKDIR /app
COPY ./action-control-linux-${TARGETARCH} /app/action-control

COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh
RUN chmod +x /app/action-control

ENTRYPOINT ["/entrypoint.sh"]