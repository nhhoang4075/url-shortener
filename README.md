# url-shortener (Go)

Dịch vụ rút gọn đường link, viết bằng Go, lưu PostgreSQL. Bản cơ bản có hai việc:

- `POST /shorten`: nhận URL dài, trả về mã ngắn.
- `GET /{mã}`: chuyển hướng (HTTP 302) tới URL dài, đồng thời đếm lượt click.

## Các phần (mỗi phần chạy được rồi mới sang phần sau)

| Phần | Việc | Kiến thức mới |
|---|---|---|
| 1 | Server trả `ok` ở `GET /health` | `net/http`, handler, `ListenAndServe` |
| 2 | `POST /shorten`, lưu tạm trong bộ nhớ | đọc/ghi JSON, `map`, sinh mã ngẫu nhiên, mã trạng thái |
| 3 | `GET /{mã}` chuyển hướng, 404 nếu không có | `PathValue`, `http.Redirect`, `Mutex` |
| 4 | Chuyển lưu trữ sang PostgreSQL | `database/sql`, driver `pgx`, bảng `links` |
| 5 | Đếm click, kiểm tra URL hợp lệ, cấu hình bằng biến môi trường | `UPDATE ... RETURNING`, `net/url`, `os.Getenv` |
| 6 | Test bằng `httptest`, tắt êm khi Ctrl+C | `testing`, `context`, `signal.NotifyContext` |

## Chạy

```bash
cd ~/Personal/url-shortener
go run .
```

Kiểm tra từ terminal khác:

```bash
curl -i http://localhost:8080/health
```

## Công cụ kiểm tra

```bash
gofmt -l .
go vet ./...
go build ./...
```
