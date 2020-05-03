FROM golang AS build

WORKDIR /go/src/smartdumpster_service
COPY src .
RUN go get -v ./...
RUN CGO_ENABLED=0 go build -ldflags '-extldflags "-static"' main.go

FROM scratch

COPY --from=build /go/src/smartdumpster_service/main /smartdumpster_service

EXPOSE 80
CMD [ "/smartdumpster_service" ]