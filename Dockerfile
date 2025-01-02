FROM golang:1.23 AS builder

WORKDIR /app

RUN apt-get update && apt-get install ansible sshpass -y

COPY . .

RUN make build

FROM builder

COPY --from=builder /app/cmd/onboarder/onboarder /usr/local/bin/onboarder
COPY --from=builder /app/cmd/agent/dcfagent /usr/local/bin/dcfagent


