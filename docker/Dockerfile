FROM alpine:3.7

ARG IMAGE_BUILD_TIME
ARG MIGRATE_VERSION

LABEL IMAGE_BUILD_TIME=${IMAGE_BUILD_TIME} \
    MIGRATE_VERSION=${MIGRATE_VERSION}

RUN cd /usr/local/bin \
    && wget -qO- https://github.com/pasztorpisti/migrate/releases/download/${MIGRATE_VERSION}/migrate-linux-amd64.zip | unzip -q - \
    && chmod +x migrate

ENTRYPOINT ["migrate"]
