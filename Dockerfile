FROM --platform=${TARGETPLATFORM} docker.io/library/alpine:3.20.2@sha256:0a4eaa0eecf5f8c050e5bba433f58c052be7587ee8af3e8b3910ef9ab5fbe9f5

ARG TARGETPLATFORM
ARG CMD_NAME
ENV COMMAND_NAME=${CMD_NAME}

COPY ${TARGETPLATFORM}/${CMD_NAME} /usr/local/bin/

CMD ["/bin/sh", "-c", "${COMMAND_NAME}"]
