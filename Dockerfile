FROM golang:1.11.2-alpine3.8
ENV SRC_DIR /go/src/app/
WORKDIR ${SRC_DIR}
RUN apk add build-base && \
    apk add git
COPY app/ ${SRC_DIR}
ENTRYPOINT [ "/bin/sh","-c" ]
CMD [ "/go/bin/app" ]
RUN cd ${SRC_DIR} && \
    go get -t ./... && \
    go install -v