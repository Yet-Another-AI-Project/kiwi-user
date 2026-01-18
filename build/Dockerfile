FROM golang:1.25 as builder

ARG github_user
ARG github_access_token

RUN go env -w GOPRIVATE=github.com
RUN git config --global url."https://${github_user}:${github_access_token}@github.com".insteadOf "https://github.com"

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN GOOS=linux go build -o /apiserver cmd/server/apiserver/main.go

FROM golang:1.25

RUN curl -sSf https://atlasgo.sh | sh
RUN apt update && apt install -y jq

COPY --from=builder /apiserver /apiserver
COPY --from=builder /app/internal/infrastructure/repository/ent/migrate/migrations /internal/infrastructure/repository/ent/migrate/migrations
COPY --from=builder /app/Makefile /Makefile

WORKDIR /

CMD [ "/apiserver" ]