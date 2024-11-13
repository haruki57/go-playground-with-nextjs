# ベースイメージ
FROM golang:1.23.3

# ワーキングディレクトリの設定
WORKDIR /app

# Goモジュールを有効にし、依存パッケージをコピー
COPY go.mod ./
RUN go mod download

# ソースコードをコピーしてビルド
COPY . .
RUN go build -o main .

# ポートを指定
EXPOSE 8080

# 実行
CMD ["./main"]
