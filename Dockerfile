FROM iron/base
LABEL maintainer "carlos.panato <ctadeu@gmail.com>"
LABEL version="1.3"

WORKDIR /app

# copy binary into image
COPY main /app


EXPOSE 8087
ENTRYPOINT ["./main"]