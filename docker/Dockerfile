# =================== 构建阶段 ===================
FROM alpine:latest AS builder

WORKDIR /app

ARG TARGETOS
ARG TARGETARCH

RUN mkdir -p /app/data

COPY --chmod=0755 backend-artifacts/ech0-${TARGETOS}-${TARGETARCH} /app/ech0

# =================== 最终镜像 ===================
FROM alpine:latest

WORKDIR /app
RUN apk add --no-cache tzdata

COPY --from=builder /app /app

EXPOSE 6277

ENTRYPOINT ["/app/ech0"]

CMD ["serve"]
