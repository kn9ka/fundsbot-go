# syntax=docker/dockerfile:1

FROM golang:1.20

# Set destination for COPY
WORKDIR /app

# Download Go modules
COPY go.mod go.sum ./
RUN go mod download
COPY ./ ./

## Build
RUN CGO_ENABLED=0 GOOS=linux go build -o ./fundsbot ./cmd/app/main.go

## Run
CMD [ "./fundsbot" ]