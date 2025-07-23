# syntax=docker/dockerfile:1
FROM alpine:3.21
LABEL author=sorc@sction.org
ARG TARGETOS
ARG TARGETARCH
COPY ./dist/minas_linux_${TARGETARCH} /usr/bin/minas
COPY ./rclone/ca.crt /usr/local/share/ca-certificates/myrootca.crt
#RUN apk add ncat openssh-client
RUN apk add tzdata ca-certificates curl unzip
RUN curl -L \
    --fail \
    --show-error \
    --progress-bar \
    --location \
    --output /tmp/rclone.zip \
    "https://downloads.rclone.org/v1.70.3/rclone-v1.70.3-linux-${TARGETARCH}.zip" && \
    unzip /tmp/rclone.zip -d /tmp/ && \
    cp -R /tmp/rclone-v1.70.3-linux-${TARGETARCH}/rclone /usr/bin/rclone && \
    rm -rf /tmp/rclone*
RUN cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" >/etc/timezone && \
    apk del tzdata \
    && mkdir /app && \
    chmod +x /usr/bin/minas && \
    chmod +x /usr/bin/rclone && \
    update-ca-certificates
WORKDIR /app
CMD ["minas","server"]
