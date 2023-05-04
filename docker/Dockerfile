FROM alpine:3.14

LABEL name="autodeploy" \
description="autodeploy service by fermi" \
owner="example@gmail.com"

WORKDIR /bin

ADD bin/autodeploy /bin/autodeploy

COPY docker/entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
