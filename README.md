# url-shortener (Go)

Dịch vụ rút gọn đường link, viết bằng Go, lưu trữ bằng PostgreSQL. Cơ bản có hai việc:

- `POST /shorten`: nhận URL dài, trả về mã ngắn.
- `GET /{code}`: chuyển hướng (HTTP 302) tới URL gốc, đồng thời đếm lượt click.

## Chuẩn bị database

```bash
createdb url_shortener
psql url_shortener
```

Trong `psql`, tạo bảng:

```sql
CREATE TABLE links (
    code       text PRIMARY KEY,
    url        text NOT NULL,
    clicks     bigint NOT NULL DEFAULT 0,
    created_at timestamptz NOT NULL DEFAULT now()
);
```

## Cấu hình

Cấu hình đọc từ biến môi trường, ví dụ như file `.env`.

| Biến | Bắt buộc | Ý nghĩa |
|---|---|---|---|
| `DATABASE_URL` | có | Chuỗi kết nối PostgreSQL, ví dụ `postgres://user@localhost:5432/url_shortener` |
| `ADDR` | không | Địa chỉ và cổng server lắng nghe |

## Chạy

```bash
cd ~/Personal/url-shortener
go run .
```

Bấm `Ctrl+C` để tắt: server chờ các request đang xử lý (tối đa 15 giây) rồi mới thoát.

Chạy trong terminal khác:

```bash
curl -i http://localhost:8080/health
curl -i -X POST http://localhost:8080/shorten -d '{"url":"https://example.com"}'
curl -i http://localhost:8080/<code>
```

## Công cụ kiểm tra

```bash
gofmt -l .
go vet ./...
go build ./...
```
