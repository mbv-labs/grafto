FROM node:18.16.1 AS build-resources

WORKDIR /

COPY resources/ resources
COPY static static

RUN cd resources && npm ci
RUN cd resources && npm run build-css

FROM golang:1.22 AS build-go

ARG appRelease=0.0.1

ENV APP_RELEASE=$appRelease

WORKDIR /

RUN go install github.com/a-h/templ/cmd/templ@latest

COPY . .

RUN templ generate

COPY --from=build-resources static static
COPY --from=build-resources pkg/mail pkg/mail

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w -X 'main.$APP_RELEASE'" -mod=readonly -v -o app cmd/app/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w -X 'main.$APP_RELEASE'" -mod=readonly -v -o worker cmd/worker/main.go

FROM scratch

WORKDIR /

COPY --from=build-go /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build-go app app
COPY --from=build-go worker worker
COPY --from=build-go static static 
COPY --from=build-go resources/seo resources/seo

CMD ["app", "worker"]
