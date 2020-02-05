ARG DOCKER_BUILD_IMAGE=golang:1.13

FROM ${DOCKER_BUILD_IMAGE} AS build
WORKDIR /app
COPY . /app/
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build .

FROM iron/base
LABEL maintainer "carlos.panato <ctadeu@gmail.com>"
LABEL version="1.4"

WORKDIR /app

# copy binary into image
COPY --from=build /app/mattermost-away-reminder /app

EXPOSE 8087
ENTRYPOINT /app/mattermost-away-reminder
