FROM tiltdev/restart-helper:2024-06-06 as restart-helper

FROM golang:1.22

COPY . .
RUN \
  CGO_ENABLED=0 \
  GOOS=linux \
  GOARCH=amd64 \
  go build \
  -tags netgo \
  -trimpath \
  -ldflags '-s -w' \
  -o app \
  ./main.go

RUN ["touch", "/tmp/.restart-proc"]
RUN ["chmod", "666", "/tmp/.restart-proc"]
COPY --from=restart-helper --chown=1000:1000 /tilt-restart-wrapper /home/render/
COPY --from=restart-helper --chown=1000:1000 /entr /home/render/

ENTRYPOINT /home/render/tilt-restart-wrapper --watch_file=/tmp/.restart-proc --entr_path /home/render/entr ./app
