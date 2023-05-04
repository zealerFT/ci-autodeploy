# 证书配置，避免私有仓库的访问问题
set -e
cd "$(dirname "$0")"
##
## Run ssh-agent (inside the build environment)
##
eval $(ssh-agent -s)
##
## Add the SSH key stored in SSH_PRIVATE_KEY variable to the agent store
## We're using tr to fix line endings which makes ed25519 keys work
## without extra base64 encoding.
## https://gitlab.com/gitlab-examples/ssh-private-key/issues/1#note_48526556
##
ssh-add <(echo "$RSA_PRIVATE_KEY")
mkdir -p /root/.ssh
# 不检查证书有效性
echo -e "Host *\n\tStrictHostKeyChecking no\n\n" > /root/.ssh/config

make $1