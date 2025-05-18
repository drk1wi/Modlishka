FROM golang:alpine

MAINTAINER Krzysztof Bednarczyk <admin@bordeux.net>

ENV MODLISHKA_REPOSITORY="github.com/drk1wi/Modlishka"
ENV INSTALL_PACKAGES="git make gcc musl-dev"
ENV PROJECT_DIR="${GOPATH}/src/${MODLISHKA_REPOSITORY}"
ENV MODLISHKA_BIN="/bin/proxy"

COPY ./run-server.sh /bin/run-server.sh
ADD . ${PROJECT_DIR}

RUN set -ex \
		&& chmod +x /bin/run-server.sh \
		&& apk add --no-cache ${INSTALL_PACKAGES}\
        && cd ${PROJECT_DIR}/ && make \
		&& cp ${PROJECT_DIR}/dist/proxy ${MODLISHKA_BIN} \
		&& apk del ${INSTALL_PACKAGES} && rm -rf /var/cache/apk/* && rm -rf ${GOPATH}/src/*
		
VOLUME [ "/data" ]
WORKDIR /data
CMD ["run-server.sh"]

EXPOSE 80 443