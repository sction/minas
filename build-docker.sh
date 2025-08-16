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
export DOCKER_BUILDKIT=1
# docker login -u sorc
# docker login
# docker run --privileged --rm tonistiigi/binfmt --install all
# docker buildx create --use --name mybuilder
# docker buildx ls
docker buildx build \
  --platform linux/arm64,linux/amd64 \
  --build-arg VER=${ver} \
  --build-arg BUILD_DATE=${build_date} \
  --push \
  --tag sorc/${name}:${ver}_${build_date} \
  --tag sorc/${name}:${ver} \
  --tag sorc/${name}:latest .

  # --build-arg HTTP_PROXY=http://10.10.10.41:1082 \
  # --build-arg HTTPS_PROXY=http://10.10.10.41:1082 \
  # --build-arg NO_PROXY=localhost,127.0.0.1,10.10.10.41 \