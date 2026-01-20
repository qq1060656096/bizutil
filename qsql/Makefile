.PHONY: test build clean install examples lint fmt

# 测试
cover:
	go test -cover ./...

# 基准测试
test:
	go test -count=1 -v -race -bench=. -benchmem ./...