# Social Media Backend Server

Một backend server hoàn chỉnh cho ứng dụng mạng xã hội được xây dựng bằng Golang, sử dụng PostgreSQL và WebRTC cho video calling.

## 🚀 Tính năng

### Đã triển khai
- ✅ **Authentication & Authorization**
  - JWT với RSA asymmetric keys
  - Đăng ký, đăng nhập, logout
  - Refresh token
  - Đổi mật khẩu
  - Rate limiting cho auth endpoints

- ✅ **User Management**
  - Profile management
  - Tìm kiếm user với text search
  - Upload avatar
  - Cập nhật thông tin cá nhân

- ✅ **Friend System**
  - Gửi/chấp nhận/từ chối lời mời kết bạn
  - Xóa bạn bè
  - Block/unblock users
  - Danh sách bạn bè với cursor-based pagination

- ✅ **Infrastructure**
  - PostgreSQL với indexing tối ưu
  - Clean Architecture pattern
  - Middleware system (Auth, CORS, Logging, Rate Limiting)
  - Configuration management
  - Graceful shutdown

### Đang phát triển
- 🔄 **Posts & Feed System**
  - Tạo, chỉnh sửa, xóa posts
  - News feed với cursor pagination
  - Likes, comments, shares
  - Media upload (images, videos)
  - Privacy settings

- 🔄 **WebRTC Video Calling**
  - WebSocket signaling server
  - Audio/video calls
  - Room management
  - Call history

- 🔄 **Real-time Features**
  - WebSocket connections
  - Live notifications
  - Online status tracking

## 🛠 Tech Stack

- **Language:** Go 1.21+
- **Web Framework:** Gin
- **Database:** PostgreSQL
- **Authentication:** JWT với RSA keys
- **WebRTC:** Pion WebRTC
- **WebSocket:** Gorilla WebSocket
- **Containerization:** Docker (planned)

## 📋 Yêu cầu hệ thống

- Go 1.21 hoặc cao hơn
- PostgreSQL 12+
- Redis (tùy chọn, cho session management)

## 🔧 Cài đặt

### 1. Clone repository
```bash
git clone <your-repo-url>
cd social_server
```

### 2. Cài đặt dependencies
```bash
go mod download
```

### 3. Setup PostgreSQL
```bash
# Cài đặt PostgreSQL locally
# Ubuntu/Debian: sudo apt-get install postgresql postgresql-contrib
# macOS: brew install postgresql
# Tạo database: social_media
createdb social_media
```

### 4. Tạo RSA keys
```bash
# Keys đã được tạo sẵn trong thư mục keys/
# Nếu muốn tạo mới:
openssl genrsa -out keys/private.pem 2048
openssl rsa -in keys/private.pem -pubout -out keys/public.pem
```

### 5. Cấu hình environment
```bash
cp .env.example .env
# Chỉnh sửa các giá trị trong .env theo môi trường của bạn
```

### 6. Chạy ứng dụng
```bash
# Development
go run cmd/server/main.go

# Build và chạy
go build -o bin/server cmd/server/main.go
./bin/server
```

## ⚙️ Cấu hình

### Environment Variables

| Variable | Mô tả | Mặc định |
|----------|--------|----------|
| `SERVER_HOST` | Host của server | `localhost` |
| `SERVER_PORT` | Port của server | `8080` |
| `POSTGRES_HOST` | PostgreSQL host | `localhost` |
| `POSTGRES_PORT` | PostgreSQL port | `5432` |
| `POSTGRES_USER` | PostgreSQL user | `postgres` |
| `POSTGRES_PASSWORD` | PostgreSQL password | `your_password` |
| `POSTGRES_DATABASE` | Tên database | `social_media` |
| `JWT_SECRET` | Secret key cho JWT | `your-secret-key-change-in-production` |
| `RSA_PRIVATE_KEY_PATH` | Đường dẫn private key | `keys/private.pem` |
| `RSA_PUBLIC_KEY_PATH` | Đường dẫn public key | `keys/public.pem` |

Xem `.env.example` để biết danh sách đầy đủ các biến môi trường.

## 📚 API Documentation

### Base URL
```
http://localhost:8080/api/v1
```

### Authentication Endpoints

#### Đăng ký
```http
POST /auth/register
Content-Type: application/json

{
  "username": "johndoe",
  "email": "john@example.com",
  "password": "password123",
  "first_name": "John",
  "last_name": "Doe"
}
```

#### Đăng nhập
```http
POST /auth/login
Content-Type: application/json

{
  "email_or_username": "john@example.com",
  "password": "password123"
}
```

#### Refresh Token
```http
POST /auth/refresh
Content-Type: application/json

{
  "refresh_token": "your_refresh_token"
}
```

### User Endpoints

#### Lấy profile
```http
GET /users/me
Authorization: Bearer <access_token>
```

#### Cập nhật profile
```http
PUT /users/me
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "display_name": "John Doe Updated",
  "bio": "Software Developer"
}
```

#### Tìm kiếm users
```http
GET /users/search?q=john&limit=20&cursor=<cursor_id>
```

#### Gửi lời mời kết bạn
```http
POST /users/{user_id}/friend-request
Authorization: Bearer <access_token>
```

### Response Format

#### Success Response
```json
{
  "data": {...},
  "message": "Success message"
}
```

#### Error Response
```json
{
  "error": "error_code",
  "message": "Error description"
}
```

#### Paginated Response
```json
{
  "data": [...],
  "next_cursor": "cursor_id",
  "has_more": true
}
```

## 📁 Cấu trúc project

```
social_server/
├── cmd/
│   └── server/
│       └── main.go              # Entry point
├── internal/
│   ├── config/
│   │   └── config.go            # Configuration management
│   ├── handlers/
│   │   └── user_handler.go      # HTTP handlers
│   ├── middleware/
│   │   ├── auth.go              # Authentication middleware
│   │   ├── cors.go              # CORS middleware
│   │   ├── logging.go           # Logging middleware
│   │   └── rate_limit.go        # Rate limiting
│   ├── models/
│   │   ├── user.go              # User models
│   │   ├── post.go              # Post models
│   │   ├── webrtc.go            # WebRTC models
│   │   └── auth.go              # Auth models
│   ├── repositories/
│   │   ├── interfaces.go        # Repository interfaces
│   │   └── postgres/
│   │       ├── user_repository.go
│   │       ├── chat_repository.go
│   │       └── post_repository.go
│   ├── routes/
│   │   └── routes.go            # Route definitions
│   └── services/
│       └── auth.go              # Business logic
├── keys/
│   ├── private.pem              # RSA private key
│   └── public.pem               # RSA public key
├── memory_bank/
│   └── wf_golang_social_backend # Workflow tracking
├── .env.example                 # Environment template
├── go.mod                       # Go modules
├── go.sum                       # Go modules checksum
└── README.md                    # This file
```

## 🔒 Security Features

- **JWT với RSA Keys:** Sử dụng asymmetric encryption cho tokens
- **Rate Limiting:** Bảo vệ khỏi spam và brute force attacks
- **Input Validation:** Validate tất cả input từ client
- **CORS Protection:** Cấu hình CORS phù hợp
- **Password Hashing:** Sử dụng bcrypt để hash passwords
- **SQL Injection Protection:** Sử dụng parameterized queries với GORM

## 🚀 Deployment

### Development
```bash
go run cmd/server/main.go
```

### Production Build
```bash
CGO_ENABLED=0 GOOS=linux go build -o bin/server cmd/server/main.go
```

### Docker (Coming Soon)
```bash
docker build -t social-server .
docker run -p 8080:8080 social-server
```

## 📊 Monitoring & Health Check

### Health Check
```http
GET /health
```

Response:
```json
{
  "status": "ok",
  "service": "social_server",
  "version": "1.0.0"
}
```

## 🧪 Testing

```bash
# Chạy tests
go test ./...

# Test với coverage
go test -cover ./...

# Test specific package
go test ./internal/services/...
```

## 📈 Performance Considerations

- **Database Indexing:** Tạo indexes cho các trường thường xuyên query
- **Cursor Pagination:** Hiệu quả hơn offset pagination cho large datasets
- **Connection Pooling:** PostgreSQL connection pool được cấu hình tối ưu
- **Rate Limiting:** Ngăn chặn abuse và giảm tải server
- **Graceful Shutdown:** Đảm bảo requests đang xử lý được hoàn thành

## 🤝 Contributing

1. Fork repository
2. Tạo feature branch (`git checkout -b feature/amazing-feature`)
3. Commit changes (`git commit -m 'Add amazing feature'`)
4. Push to branch (`git push origin feature/amazing-feature`)
5. Tạo Pull Request

## 📝 Changelog

### v1.0.0 (Current)
- ✅ Basic authentication system
- ✅ User management
- ✅ Friend system
- ✅ Cursor-based pagination
- ✅ Rate limiting
- ✅ PostgreSQL integration với GORM ORM

### v1.1.0 (Planned)
- 🔄 Posts & Feed system
- 🔄 WebRTC video calling
- 🔄 Real-time notifications

## 📞 Support

Nếu gặp vấn đề hoặc có câu hỏi, vui lòng:
1. Kiểm tra [Issues](../../issues)
2. Tạo issue mới nếu chưa có
3. 6

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

Made with ❤️ using Go and PostgreSQL
