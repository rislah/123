FROM golang:latest as build
WORKDIR /go/src/fakes
COPY . .
RUN make build

FROM alpine:3.11
COPY --from=build /go/src/fakes/main .
COPY --from=build /go/src/fakes/GeoLite2-Country.mmdb .
RUN ls
ENTRYPOINT ["./main"] 
