#!/bin/bash

# Exit ngay khi có lỗi
set -e

# In ra lệnh đang chạy (tùy chọn)
set -x

# Chạy lệnh build
go build cmd/server/main.go
