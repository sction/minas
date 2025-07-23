#!/bin/bash
name=minas
ver=$1
# 在代码中搜索 my-version，修改版本号
build_date=$(date +"%Y%m%d")
if [ -z "${ver}" ]; then
  ver=1.3.3
fi
echo ${ver}_${build_date}
export DOCKER_CLI_EXPERIMENTAL=enabled
echo ${DOCKER_HUB_KEY} | docker login --username ${DOCKER_HUB_USER} --password-stdin
docker buildx build \
  --platform linux/arm64,linux/amd64 \
  --build-arg VER=${ver} \
  --build-arg BUILD_DATE=${build_date} \
  --build-arg HTTP_PROXY=http://10.10.10.41:20172 \
  --build-arg HTTPS_PROXY=http://10.10.10.41:20172 \
  --build-arg NO_PROXY=localhost,127.0.0.1,10.10.10.41 \
  --push \
  --tag sorc/${name}:${ver}_${build_date} \
  --tag sorc/${name}:${ver} \
  --tag sorc/${name}:latest .
