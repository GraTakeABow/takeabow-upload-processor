FROM alpine:3.5
RUN apk add --update \
    tzdata \
    ffmpeg \
    ca-certificates \
    youtube-dl \
    && rm -rf /var/cache/apk/* > /dev/null
ENV ZONEINFO /usr/share/zoneinfo/
COPY takeabow-upload-processor /
ENTRYPOINT ["/takeabow-upload-processor"]