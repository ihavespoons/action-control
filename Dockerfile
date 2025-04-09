FROM alpine:3.21
ARG TARGETARCH

RUN apk --no-cache add ca-certificates

WORKDIR /app
COPY ./action-control-linux-${TARGETARCH} /app/action-control

# Create entrypoint script to handle boolean flags
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh
RUN chmod +x /action-control

ENTRYPOINT ["/entrypoint.sh"]