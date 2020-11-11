FROM golang:1.15-alpine

WORKDIR /src

EXPOSE 80
COPY . .

WORKDIR ./notification_service
RUN go get github.com/silenceper/gowatch
RUN go mod vendor
RUN go build -mod=vendor

RUN ls -lath
CMD ["gowatch"]