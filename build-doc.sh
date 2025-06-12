#!/bin/bash

# Exit ngay khi có lỗi
set -e

# In ra lệnh đang chạy (tùy chọn)
set -x

# Chạy lệnh swag
swag init -g cmd/server/main.go --parseDependency --parseInternal
