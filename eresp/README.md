# eresp

企业级统一 API 响应结构（Google API 风格）。

## 特性

- **标准化响应格式**：统一的 JSON 响应结构，符合 Google API 设计规范
- **成功/错误分离**：`data` 字段仅用于成功返回，`details` 字段仅用于错误附加信息
- **语义错误标识**：`reason` 字段提供稳定的机器可读错误标识
- **链路追踪支持**：内置 `trace_id` 字段支持分布式链路追踪
- **错误自动转换**：支持从 `error` 类型自动转换为统一响应结构
- **泛型支持**：提供类型安全的泛型成功响应方法
- **不可变设计**：响应对象支持不可变操作，避免意外修改

## 安装

```bash
go get github.com/qq1060656096/bizutil/eresp
```

## 响应结构

```json
{
  "code": 0,                    // 业务码（0=成功，非0=错误）
  "reason": "USER_NOT_FOUND",   // 稳定错误标识（可选）
  "message": "用户不存在",       // 用户提示信息
  "data": {},                   // 成功返回数据（仅成功时存在）
  "details": {},                // 错误附加信息（仅错误时存在）
  "trace_id": "abc123"          // 链路追踪ID（可选）
}
```

## 使用方法

### 成功响应

#### 基本成功响应

```go
package main

import (
    "net/http"
    "github.com/gin-gonic/gin"
    "github.com/qq1060656096/bizutil/eresp"
)

func handleSuccess(c *gin.Context) {
    // 简单成功响应
    resp := eresp.OKResp("操作成功", "创建完成")
    c.JSON(http.StatusOK, resp)
    
    // 输出:
    // {
    //   "code": 0,
    //   "message": "创建完成",
    //   "data": "操作成功"
    // }
}
```

#### 泛型成功响应（推荐）

```go
func handleUser(c *gin.Context) {
    user := struct {
        ID   int    `json:"id"`
        Name string `json:"name"`
    }{
        ID:   1,
        Name: "张三",
    }
    
    // 类型安全的泛型响应
    resp := eresp.Ok(user, "获取用户信息成功")
    c.JSON(http.StatusOK, resp)
    
    // 输出:
    // {
    //   "code": 0,
    //   "message": "获取用户信息成功",
    //   "data": {
    //     "id": 1,
    //     "name": "张三"
    //   }
    // }
}
```

#### 无数据成功响应

```go
func handleDelete(c *gin.Context) {
    // 删除操作，无返回数据
    resp := eresp.OKResp(nil, "删除成功")
    c.JSON(http.StatusOK, resp)
    
    // 输出:
    // {
    //   "code": 0,
    //   "message": "删除成功"
    // }
}
```

### 错误响应

#### 基本错误响应

```go
func handleError(c *gin.Context) {
    resp := eresp.ErrorResp(
        400,                    // 业务码
        "BAD_REQUEST",         // 语义标识
        "请求参数错误",          // 用户消息
        map[string]string{      // 错误详情
            "field": "email",
            "error": "格式不正确",
        },
    )
    c.JSON(http.StatusBadRequest, resp)
    
    // 输出:
    // {
    //   "code": 400,
    //   "reason": "BAD_REQUEST",
    //   "message": "请求参数错误",
    //   "details": {
    //     "field": "email",
    //     "error": "格式不正确"
    //   }
    // }
}
```

### 链路追踪

```go
func handleWithTrace(c *gin.Context) {
    traceID := c.GetHeader("X-Trace-ID")
    
    resp := eresp.OKResp("data", "成功").
        WithTrace(traceID)
    
    c.JSON(http.StatusOK, resp)
    
    // 输出:
    // {
    //   "code": 0,
    //   "message": "成功",
    //   "data": "data",
    //   "trace_id": "abc123"
    // }
}
```

## 错误自动转换

`eresp.FromError` 方法支持从 `error` 类型自动转换为统一响应结构：

### 支持的错误类型

1. **errcode.Error** - 业务错误，自动提取 `CodeInt()`、`Reason()`、`Message()`
2. **HTTP Error** - 实现 `StatusCode()` 接口的 HTTP 错误
3. **普通错误** - 其他任何实现了 `error` 接口的类型

### 转换示例

```go
package main

import (
    "errors"
    "net/http"
    "github.com/gin-gonic/gin"
    "github.com/qq1060656096/bizutil/eresp"
    "github.com/qq1060656096/bizutil/errcode"
)

func handleErrorConversion(c *gin.Context) {
    var err error
    
    // 1. errcode.Error 自动转换
    err = errcode.NewWithReason(1014040001, "USER_NOT_FOUND", "用户不存在")
    resp := eresp.FromError(err, map[string]string{"user_id": "123"})
    // 输出: code=1014040001, reason="USER_NOT_FOUND", message="用户不存在"
    
    // 2. HTTP Error 自动转换
    type httpError struct {
        status int
        msg    string
    }
    func (e *httpError) Error() string { return e.msg }
    func (e *httpError) StatusCode() int { return e.status }
    
    err = &httpError{status: 404, msg: "Not Found"}
    resp = eresp.FromError(err, "request details")
    // 输出: code=404, reason="HTTP_ERROR", message="Not Found"
    
    // 3. 普通错误转换
    err = errors.New("something went wrong")
    resp = eresp.FromError(err, nil)
    // 输出: code=-1, reason="INTERNAL_ERROR", message="internal server error"
    
    // 4. nil error 转换为成功响应
    resp = eresp.FromError(nil, nil)
    // 输出: code=0, message="OK"
    
    c.JSON(http.StatusOK, resp)
}
```

## Gin 框架集成

### 统一响应中间件

```go
package middleware

import (
    "net/http"
    "github.com/gin-gonic/gin"
    "github.com/qq1060656096/bizutil/eresp"
    "github.com/qq1060656096/bizutil/errcode"
)

// ResponseMiddleware 统一响应处理中间件
func ResponseMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Next()
        
        // 如果已经有响应，则跳过
        if c.Writer.Written() {
            return
        }
        
        // 获取链路追踪ID
        traceID := c.GetHeader("X-Trace-ID")
        if traceID == "" {
            traceID = generateTraceID() // 自定义生成函数
        }
        
        // 处理错误
        if len(c.Errors) > 0 {
            err := c.Errors.Last().Err
            resp := eresp.FromError(err, nil).WithTrace(traceID)
            
            // 设置HTTP状态码
            if ec, ok := err.(*errcode.Error); ok {
                c.Status(ec.HTTPStatus())
            } else {
                c.Status(http.StatusInternalServerError)
            }
            
            c.JSON(resp.Code, resp)
            return
        }
        
        // 默认成功响应
        resp := eresp.OKResp(nil, "操作成功").WithTrace(traceID)
        c.JSON(http.StatusOK, resp)
    }
}
```

### Controller 使用示例

```go
package controller

import (
    "net/http"
    "github.com/gin-gonic/gin"
    "github.com/qq1060656096/bizutil/eresp"
    "github.com/qq1060656096/bizutil/errcode"
)

// UserController 用户控制器
type UserController struct{}

// GetUser 获取用户信息
func (ctrl *UserController) GetUser(c *gin.Context) {
    userID := c.Param("id")
    
    // 参数验证
    if userID == "" {
        c.Error(errcode.NewWithReason(1014000001, "INVALID_PARAM", "用户ID不能为空"))
        return
    }
    
    // 模拟业务逻辑
    if userID == "0" {
        c.Error(errcode.NewWithReason(1014040001, "USER_NOT_FOUND", "用户不存在"))
        return
    }
    
    // 成功响应
    user := map[string]interface{}{
        "id":   userID,
        "name": "张三",
        "age":  25,
    }
    
    resp := eresp.Ok(user, "获取用户信息成功")
    c.JSON(http.StatusOK, resp)
}

// CreateUser 创建用户
func (ctrl *UserController) CreateUser(c *gin.Context) {
    var req struct {
        Name  string `json:"name" binding:"required"`
        Email string `json:"email" binding:"required,email"`
    }
    
    if err := c.ShouldBindJSON(&req); err != nil {
        c.Error(errcode.NewWithReason(1014000001, "INVALID_PARAM", "请求参数错误"))
        return
    }
    
    // 模拟用户已存在
    if req.Email == "exists@example.com" {
        c.Error(errcode.NewWithReason(1014090001, "USER_EXISTS", "用户已存在"))
        return
    }
    
    // 创建用户成功
    user := map[string]interface{}{
        "id":    "12345",
        "name":  req.Name,
        "email": req.Email,
    }
    
    resp := eresp.Ok(user, "创建用户成功")
    c.JSON(http.StatusCreated, resp)
}
```

### 路由配置

```go
package main

import (
    "github.com/gin-gonic/gin"
    "github.com/qq1060656096/bizutil/middleware"
    "github.com/qq1060656096/bizutil/controller"
)

func main() {
    r := gin.Default()
    
    // 注册响应中间件
    r.Use(middleware.ResponseMiddleware())
    
    userCtrl := &controller.UserController{}
    
    // 用户路由
    api := r.Group("/api/v1")
    {
        api.GET("/users/:id", userCtrl.GetUser)
        api.POST("/users", userCtrl.CreateUser)
    }
    
    r.Run(":8080")
}
```

## 常量定义

```go
const (
    SuccessCode = 0   // 成功码
    UnknownCode = -1  // 未知错误码
)
```

## API 设计原则

1. **code=0 表示成功**：所有成功响应的业务码都必须为 0
2. **data 只用于成功返回**：成功响应包含 `data` 字段，错误响应不包含
3. **details 只用于错误附加信息**：错误响应包含 `details` 字段，成功响应不包含
4. **reason 为稳定机器可读标识**：用于客户端程序化处理错误
5. **message 为用户友好提示**：用于直接展示给用户的信息
6. **trace_id 支持链路追踪**：便于分布式系统中的问题排查

## 响应示例

### 成功响应示例

```json
{
  "code": 0,
  "message": "获取用户列表成功",
  "data": {
    "users": [
      {
        "id": 1,
        "name": "张三",
        "email": "zhangsan@example.com"
      },
      {
        "id": 2,
        "name": "李四",
        "email": "lisi@example.com"
      }
    ],
    "total": 2,
    "page": 1,
    "page_size": 10
  },
  "trace_id": "trace-123-456"
}
```

### 错误响应示例

```json
{
  "code": 1014040001,
  "reason": "USER_NOT_FOUND",
  "message": "用户不存在",
  "details": {
    "user_id": "123",
    "search_time": "2024-01-01T12:00:00Z"
  },
  "trace_id": "trace-123-456"
}
```

## 性能考虑

- **泛型方法**：`eresp.Ok[T]` 提供类型安全的同时保持高性能
- **不可变设计**：`WithTrace` 方法返回新对象，避免并发问题
- **零拷贝优化**：响应结构体设计紧凑，减少内存分配

## 测试

```bash
# 运行所有测试
go test ./...

# 运行性能测试
go test -bench=. ./...

# 生成测试覆盖率报告
go test -cover ./...
```

## 依赖关系

- `github.com/qq1060656096/bizutil/errcode` - 业务错误码包（可选）

## 注意事项

1. 响应对象是值类型，`WithTrace` 方法会返回新的响应对象
2. `FromError` 方法会自动处理 `nil` 错误，返回成功响应
3. 在并发环境中使用时，注意响应对象的不可变性
4. 建议配合 `errcode` 包使用，以获得完整的错误处理能力
5. `trace_id` 字段为可选，但在微服务架构中强烈建议使用

