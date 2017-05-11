FROM golang:1.8-alpine

RUN mkdir -p /go/src/app
WORKDIR /go/src/app
COPY . /go/src/app
RUN set -ex \
		&& apk add --no-cache --virtual .build-deps \
			gcc \
			musl-dev \
			git \
		&& go-wrapper download \
    && go-wrapper install \
		&& apk del .build-deps

RUN mkdir -p /etc/nginx/certs
WORKDIR /etc/nginx/certs
CMD ["go-wrapper", "run"]
