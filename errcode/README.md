# errcode

结构化业务错误码实现（Google API 风格）。

## 特性

- **标准化错误码格式**：10位数字编码，包含模块、HTTP状态码和顺序号
- **HTTP状态码自动解析**：从错误码直接提取对应的HTTP状态码
- **语义错误码支持**：可选的reason字段，用于API语义错误
- **完全兼容Go错误处理**：支持`errors.Is`、`errors.As`和`errors.Unwrap`
- **错误包装**：支持包装已有错误，保留原始错误信息

## 错误码格式

```
1位占位符 + 2位模块 + 3位HTTP状态码 + 4位顺序

示例：
1 01 404 0001
│ │  │   └── 顺序号
│ │  └────── HTTP状态码
│ └──────── 模块
└────────── 占位符（仅保证可转 int，1-9）
```

## 安装

```bash
go get github.com/your-repo/bizutil/errcode
```

## 使用方法

### 创建基本错误

```go
package main

import (
    "fmt"
    "github.com/your-repo/bizutil/errcode"
)

func main() {
    // 9位错误码（自动补占位符）
    err := errcode.New(014040001, "用户不存在")
    fmt.Println(err) // 输出: [1014040001] 用户不存在
    
    // 10位完整错误码
    err = errcode.New(1014040001, "用户不存在")
    fmt.Println(err) // 输出: [1014040001] 用户不存在
}
```

### 创建带语义码的错误（推荐用于API）

```go
err := errcode.NewWithReason(1014040001, "USER_NOT_FOUND", "用户不存在")
fmt.Println(err) // 输出: [USER_NOT_FOUND:1014040001] 用户不存在
```

### 错误包装

```go
originalErr := fmt.Errorf("database connection failed")
err := errcode.Wrap(105500001, originalErr, "服务内部错误")
fmt.Println(err) // 输出: [105500001] 服务内部错误: database connection failed
```

### 带语义码的错误包装

```go
originalErr := fmt.Errorf("database connection failed")
err := errcode.WrapWithReason(105500001, "INTERNAL_ERROR", originalErr, "服务内部错误")
fmt.Println(err) // 输出: [INTERNAL_ERROR:105500001] 服务内部错误: database connection failed
```

## 错误处理

### 获取错误信息

```go
err := errcode.NewWithReason(1014040001, "USER_NOT_FOUND", "用户不存在")
if e, ok := err.(*errcode.Error); ok {
    fmt.Println("错误码:", e.Code())        // 1014040001
    fmt.Println("语义码:", e.Reason())      // USER_NOT_FOUND
    fmt.Println("错误消息:", e.Message())   // 用户不存在
    fmt.Println("HTTP状态:", e.HTTPStatus()) // 404
    fmt.Println("数字码:", e.CodeInt())     // 1014040001
}
```

### 错误比较

```go
err1 := errcode.New(1014040001, "用户不存在")
err2 := errcode.New(1014040001, "另一个消息")

// 使用 errors.Is
if errors.Is(err1, err2) {
    fmt.Println("错误码相同")
}

// 直接比较
if e, ok := err1.(*errcode.Error); ok {
    target := &errcode.Error{code: "1014040001"}
    if e.Is(target) {
        fmt.Println("错误码匹配")
    }
}
```

### 错误解包

```go
originalErr := fmt.Errorf("原始错误")
err := errcode.Wrap(105500001, originalErr, "包装错误")

// 获取原始错误
if errors.Unwrap(err) == originalErr {
    fmt.Println("原始错误匹配")
}
```

## 错误码设计原则

1. **第一位仅占位**：无业务含义，仅保证可转换为int（1-9）
2. **模块标识**：第2-3位表示业务模块
3. **HTTP状态码**：第4-6位直接对应HTTP状态码
4. **顺序号**：第7-10位用于区分同类型错误
5. **reason可选**：用于API语义错误，便于客户端处理

## 常见错误码示例

```go
const (
    // 用户模块 (01)
    UserNotFound     = 1014040001 // 用户不存在
    UserExists       = 1014090001 // 用户已存在
    UserInvalid      = 1014000001 // 用户数据无效
    
    // 订单模块 (02)  
    OrderNotFound    = 1024040001 // 订单不存在
    OrderExpired     = 1024100001 // 订单已过期
    OrderPaid        = 1024090001 // 订单已支付
    
    // 系统模块 (99)
    InternalError    = 1995000001 // 内部错误
    ServiceUnavailable = 1995030001 // 服务不可用
    RateLimitExceeded = 1994290001 // 请求频率超限
)
```

## API响应示例

```go
func handleUserNotFound(w http.ResponseWriter, r *http.Request) {
    err := errcode.NewWithReason(1014040001, "USER_NOT_FOUND", "用户不存在")
    
    response := map[string]interface{}{
        "error": map[string]interface{}{
            "code":    err.(*errcode.Error).Code(),
            "reason":  err.(*errcode.Error).Reason(),
            "message": err.(*errcode.Error).Message(),
        },
        "status": err.(*errcode.Error).HTTPStatus(),
    }
    
    w.WriteHeader(err.(*errcode.Error).HTTPStatus())
    json.NewEncoder(w).Encode(response)
}
```

## Gin 框架集成

### 错误处理中间件

```go
package middleware

import (
    "net/http"
    
    "github.com/gin-gonic/gin"
    "github.com/your-repo/bizutil/errcode"
)

// ErrorHandler 统一错误处理中间件
func ErrorHandler() gin.HandlerFunc {
    return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
        var err error
        
        switch x := recovered.(type) {
        case string:
            err = errcode.New(1995000001, x)
        case error:
            err = x
        default:
            err = errcode.New(1995000001, "未知错误")
        }
        
        HandleError(c, err)
    })
}

// HandleError 处理错误响应
func HandleError(c *gin.Context, err error) {
    if ec, ok := err.(*errcode.Error); ok {
        c.JSON(ec.HTTPStatus(), gin.H{
            "error": gin.H{
                "code":    ec.Code(),
                "reason":  ec.Reason(),
                "message": ec.Message(),
            },
            "status": ec.HTTPStatus(),
        })
        return
    }
    
    // 处理非 errcode 错误
    c.JSON(http.StatusInternalServerError, gin.H{
        "error": gin.H{
            "code":    "1995000001",
            "reason":  "INTERNAL_ERROR",
            "message": err.Error(),
        },
        "status": http.StatusInternalServerError,
    })
}
```

### Controller 使用示例

```go
package controller

import (
    "github.com/gin-gonic/gin"
    "github.com/your-repo/bizutil/errcode"
)

// UserController 用户控制器
type UserController struct{}

// GetUser 获取用户信息
func (ctrl *UserController) GetUser(c *gin.Context) {
    userID := c.Param("id")
    
    // 模拟业务逻辑
    if userID == "0" {
        err := errcode.NewWithReason(1014040001, "USER_NOT_FOUND", "用户不存在")
        HandleError(c, err)
        return
    }
    
    if userID == "invalid" {
        err := errcode.NewWithReason(1014000001, "USER_INVALID", "用户ID格式错误")
        HandleError(c, err)
        return
    }
    
    // 成功响应
    c.JSON(http.StatusOK, gin.H{
        "data": gin.H{
            "id":   userID,
            "name": "张三",
        },
    })
}

// CreateUser 创建用户
func (ctrl *UserController) CreateUser(c *gin.Context) {
    var req struct {
        Name  string `json:"name" binding:"required"`
        Email string `json:"email" binding:"required,email"`
    }
    
    if err := c.ShouldBindJSON(&req); err != nil {
        ec := errcode.NewWithReason(1014000001, "USER_INVALID", "请求参数错误")
        HandleError(c, ec)
        return
    }
    
    // 模拟用户已存在
    if req.Email == "exists@example.com" {
        ec := errcode.NewWithReason(1014090001, "USER_EXISTS", "用户已存在")
        HandleError(c, ec)
        return
    }
    
    // 创建用户逻辑...
    c.JSON(http.StatusCreated, gin.H{
        "data": gin.H{
            "id":    "12345",
            "name":  req.Name,
            "email": req.Email,
        },
    })
}
```

### 路由配置

```go
package main

import (
    "github.com/gin-gonic/gin"
    "github.com/your-repo/bizutil/errcode"
    "github.com/your-repo/middleware"
    "github.com/your-repo/controller"
)

func main() {
    r := gin.Default()
    
    // 注册错误处理中间件
    r.Use(middleware.ErrorHandler())
    
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

## GORM 集成

### 错误转换函数

```go
package database

import (
    "errors"
    "gorm.io/gorm"
    "github.com/your-repo/bizutil/errcode"
)

// ConvertGormError 将 GORM 错误转换为 errcode
func ConvertGormError(err error, operation string) error {
    if err == nil {
        return nil
    }
    
    if errors.Is(err, gorm.ErrRecordNotFound) {
        return errcode.NewWithReason(1014040001, "RECORD_NOT_FOUND", 
            operation+"失败：记录不存在")
    }
    
    if errors.Is(err, gorm.ErrInvalidTransaction) {
        return errcode.NewWithReason(1995000001, "INVALID_TRANSACTION", 
            operation+"失败：无效的事务")
    }
    
    if errors.Is(err, gorm.ErrDuplicatedKey) {
        return errcode.NewWithReason(1014090001, "DUPLICATE_KEY", 
            operation+"失败：数据已存在")
    }
    
    // 数据库连接错误
    if errors.Is(err, gorm.ErrDBNotConfigured) {
        return errcode.NewWithReason(1995030001, "DB_NOT_CONFIGURED", 
            "数据库未配置")
    }
    
    // 包装其他数据库错误
    return errcode.WrapWithReason(1995000001, "DATABASE_ERROR", 
        err, operation+"失败：数据库错误")
}
```

### Repository 使用示例

```go
package repository

import (
    "gorm.io/gorm"
    "github.com/your-repo/bizutil/errcode"
    "github.com/your-repo/database"
)

// User 用户模型
type User struct {
    ID    uint   `gorm:"primaryKey"`
    Name  string `gorm:"size:100;not null"`
    Email string `gorm:"size:100;uniqueIndex;not null"`
}

// UserRepository 用户仓库
type UserRepository struct {
    db *gorm.DB
}

// NewUserRepository 创建用户仓库
func NewUserRepository(db *gorm.DB) *UserRepository {
    return &UserRepository{db: db}
}

// GetByID 根据ID获取用户
func (r *UserRepository) GetByID(id uint) (*User, error) {
    var user User
    err := r.db.First(&user, id).Error
    if err != nil {
        return nil, database.ConvertGormError(err, "查询用户")
    }
    return &user, nil
}

// GetByEmail 根据邮箱获取用户
func (r *UserRepository) GetByEmail(email string) (*User, error) {
    var user User
    err := r.db.Where("email = ?", email).First(&user).Error
    if err != nil {
        return nil, database.ConvertGormError(err, "查询用户")
    }
    return &user, nil
}

// Create 创建用户
func (r *UserRepository) Create(user *User) error {
    err := r.db.Create(user).Error
    if err != nil {
        return database.ConvertGormError(err, "创建用户")
    }
    return nil
}

// Update 更新用户
func (r *UserRepository) Update(user *User) error {
    result := r.db.Save(user)
    if result.Error != nil {
        return database.ConvertGormError(result.Error, "更新用户")
    }
    
    if result.RowsAffected == 0 {
        return errcode.NewWithReason(1014040001, "USER_NOT_FOUND", 
            "更新失败：用户不存在")
    }
    
    return nil
}

// Delete 删除用户
func (r *UserRepository) Delete(id uint) error {
    result := r.db.Delete(&User{}, id)
    if result.Error != nil {
        return database.ConvertGormError(result.Error, "删除用户")
    }
    
    if result.RowsAffected == 0 {
        return errcode.NewWithReason(1014040001, "USER_NOT_FOUND", 
            "删除失败：用户不存在")
    }
    
    return nil
}

// List 分页获取用户列表
func (r *UserRepository) List(page, pageSize int) ([]*User, int64, error) {
    var users []*User
    var total int64
    
    // 计算总数
    if err := r.db.Model(&User{}).Count(&total).Error; err != nil {
        return nil, 0, database.ConvertGormError(err, "统计用户数量")
    }
    
    // 分页查询
    offset := (page - 1) * pageSize
    err := r.db.Offset(offset).Limit(pageSize).Find(&users).Error
    if err != nil {
        return nil, 0, database.ConvertGormError(err, "查询用户列表")
    }
    
    return users, total, nil
}
```

### Service 层集成示例

```go
package service

import (
    "github.com/your-repo/bizutil/errcode"
    "github.com/your-repo/repository"
)

// UserService 用户服务
type UserService struct {
    userRepo *repository.UserRepository
}

// NewUserService 创建用户服务
func NewUserService(userRepo *repository.UserRepository) *UserService {
    return &UserService{userRepo: userRepo}
}

// GetUser 获取用户
func (s *UserService) GetUser(id uint) (*repository.User, error) {
    user, err := s.userRepo.GetByID(id)
    if err != nil {
        // 可以在这里添加业务逻辑的错误转换
        if ec, ok := err.(*errcode.Error); ok && ec.Reason() == "RECORD_NOT_FOUND" {
            return nil, errcode.NewWithReason(1014040001, "USER_NOT_FOUND", "用户不存在")
        }
        return nil, err
    }
    return user, nil
}

// CreateUser 创建用户
func (s *UserService) CreateUser(name, email string) (*repository.User, error) {
    // 检查邮箱是否已存在
    existing, err := s.userRepo.GetByEmail(email)
    if err == nil && existing != nil {
        return nil, errcode.NewWithReason(1014090001, "USER_EXISTS", "用户已存在")
    }
    
    // 创建用户
    user := &repository.User{
        Name:  name,
        Email: email,
    }
    
    err = s.userRepo.Create(user)
    if err != nil {
        return nil, err
    }
    
    return user, nil
}
```

## 注意事项

1. 错误码必须是9位或10位数字
2. HTTP状态码部分必须在100-599范围内
3. 包装错误时，如果传入的error为nil，将返回nil
4. 错误码第一位自动补1（当使用9位错误码时）
5. 与标准库`error`接口完全兼容，可无缝集成现有代码
6. 在Gin中使用时，建议配合错误处理中间件实现统一错误响应
7. 在GORM中使用时，建议封装错误转换函数，统一处理数据库错误
