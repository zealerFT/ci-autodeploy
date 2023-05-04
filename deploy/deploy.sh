#!/usr/bin/env bash

ROOT="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PAYLOAD=$ROOT/$1

sed "s,<token>,${ACCESS_TOKEN},g; s,<refer>,$CI_JOB_URL,g; s,<author>,$GITLAB_USER_EMAIL,g; s,<message>,$CI_COMMIT_TITLE,g; s,<env>,$ENV,g; s,<image>,$DEPLOY_IMAGE,g; s,<version>,$CLUSTER_VERSION,g;" "$PAYLOAD" > "$PAYLOAD.tmp" && mv "$PAYLOAD.tmp" "$PAYLOAD"

cat $PAYLOAD
http_code=$(curl -X POST -H "Content-Type:application/json " \
            -H "Authorization: Bearer ${AUTO_DEPLOY}" \
            --data-binary @$PAYLOAD \
            -s -o out.json \
            -w '%{http_code}' \
            http://autodeploy:80/api/v1/autodeploy/yaml/image/update) # 这里调用的就是自动部署的服务

echo 'HTTP Response' $http_code
cat out.json

if [[ $http_code -eq 200 ]]; then
    exit 0
fi
exit 1
