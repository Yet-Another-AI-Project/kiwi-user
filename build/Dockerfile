FROM golang:1.24 as builder

ARG gitlab_user
ARG gitlab_access_token

RUN go env -w GOPRIVATE=git.futurx.cn
RUN git config --global url."https://${gitlab_user}:${gitlab_access_token}@git.futurx.cn".insteadOf "https://git.futurx.cn"

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN GOOS=linux go build -o /apiserver cmd/server/apiserver/main.go

FROM golang:1.24

RUN curl -sSf https://atlasgo.sh | sh
RUN apt update && apt install -y jq

COPY --from=builder /apiserver /apiserver
COPY --from=builder /app/internal/infrastructure/repository/ent/migrate/migrations /internal/infrastructure/repository/ent/migrate/migrations
COPY --from=builder /app/Makefile /Makefile

WORKDIR /

CMD [ "/apiserver" ]