FROM golang:1.14-alpine

RUN mkdir -p /go/src/app
WORKDIR /go/src/app
COPY . /go/src/app
RUN set -ex \
	&& apk add --no-cache --virtual .build-deps \
	  gcc \
	  musl-dev \
	  git \
	  glide \
	&& go get github.com/sgotti/glide-vc \
	&& glide install \
	&& glide vc --use-lock-file --only-code --no-tests \
	&& go-wrapper install \
	&& go clean -i github.com/sgotti/glide-vc... \
	&& apk del .build-deps

RUN mkdir -p /etc/nginx/certs
ENV LETSENCRYPT_PATH /etc/nginx/certs
CMD ["go-wrapper", "run"]
