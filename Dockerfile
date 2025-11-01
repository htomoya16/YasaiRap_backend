####################
# ビルドステージ
####################
FROM golang:1.25.1-alpine AS builder

WORKDIR /app

# パッケージのインストール（--no-cacheでキャッシュ削除）
RUN apk add --no-cache \
            git \
            curl

# Go モジュールファイルをコピー
COPY go.mod go.sum ./

# 依存関係をダウンロード
RUN go mod download

# ソースコードをコピー
COPY . .

# ビルド（静的リンク）
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/server

####################
# 実行ステージ
####################
FROM alpine:3.20

WORKDIR /app

# パッケージのインストール（--no-cacheでキャッシュ削除）
RUN apk add --no-cache \
            ca-certificates \
            tzdata \
            curl

# ビルドステージで作成したバイナリをコピー
COPY --from=builder /app/main .

# 非rootユーザ
RUN adduser -D -u 65532 user
USER user

# ポートを動的に設定
ENV PORT=8080
EXPOSE 8080

# 実行
CMD ["./main"]
