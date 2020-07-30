FROM golang:1.14-alpine AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd

RUN CGO_ENABLED=0 go build -ldflags '-extldflags "-static"' -o service cmd/smartdumpsterservice/*.go

FROM scratch

COPY --from=build /app/service /smartdumpsterservice
COPY signkey /signkey
COPY dashboard /dashboard

EXPOSE 8080
CMD [ "/smartdumpsterservice" ]