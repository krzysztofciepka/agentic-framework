FROM golang:1.26-alpine AS build

RUN apk add --no-cache nodejs npm gcc musl-dev

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN cd web && npm install --silent && npm run build

RUN CGO_ENABLED=1 go build -o /server ./cmd/server

FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata

COPY --from=build /server /server

EXPOSE 8080

ENTRYPOINT ["/server"]
