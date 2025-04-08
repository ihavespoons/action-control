FROM alpine:3.21
ARG TARGETARCH

RUN apk --no-cache add ca-certificates

WORKDIR /app
COPY /home/runner/work/action-control/action-control/artifacts/action-control-linux-${TARGETARCH} /app/action-control

# Create entrypoint script to handle boolean flags
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]