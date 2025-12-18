# Response Package

统一的响应格式和错误处理包，提供标准化的 HTTP 响应结构、错误类型定义和中间件支持。

## 特性

- **统一响应格式**: 标准化的成功和错误响应结构
- **丰富的错误类型**: 预定义的业务错误和系统错误
- **错误处理中间件**: 全局错误捕获和处理
- **请求追踪**: 请求ID生成和传播
- **结构化日志**: 详细的请求和错误日志记录
- **类型安全**: 强类型的错误检查和处理

## 快速开始

### 1. 基础响应

```go
import (
    "github.com/gin-gonic/gin"
    "backend-go/pkg/response"
)

func GetUser(c *gin.Context) {
    user := &User{ID: 1, Name: "John"}
    
    // 成功响应
    response.Success(c, user)
    
    // 带消息的成功响应
    response.SuccessWithMessage(c, "用户获取成功", user)
    
    // 创建成功响应
    response.Created(c, user)
}
```

### 2. 错误响应

```go
func CreateUser(c *gin.Context) {
    var user User
    if err := c.ShouldBindJSON(&user); err != nil {
        // 验证错误
        response.ValidationError(c, "参数验证失败", err.Error())
        return
    }
    
    if userExists(user.Email) {
        // 业务错误
        response.Conflict(c, "用户已存在", map[string]string{"email": user.Email})
        return
    }
    
    if err := saveUser(&user); err != nil {
        // 内部错误
        response.InternalError(c, "保存用户失败")
        return
    }
    
    response.Created(c, user)
}
```

### 3. 分页响应

```go
func ListUsers(c *gin.Context) {
    users, total := getUserList(page, pageSize)
    
    pagination := &response.PaginationInfo{
        Page:       page,
        PageSize:   pageSize,
        Total:      total,
        TotalPages: (total + pageSize - 1) / pageSize,
        HasNext:    page < totalPages,
        HasPrev:    page > 1,
    }
    
    response.Paginated(c, users, pagination)
}
```

## 响应格式

### 成功响应

```json
{
  "success": true,
  "code": 200,
  "message": "操作成功",
  "data": {
    "id": 1,
    "name": "John"
  },
  "timestamp": 1640995200,
  "request_id": "req_123456"
}
```

### 错误响应

```json
{
  "success": false,
  "code": 400,
  "message": "参数验证失败",
  "error": {
    "type": "validation_error",
    "code": "VALIDATION_FAILED",
    "message": "参数验证失败",
    "details": {
      "field": "email",
      "reason": "invalid format"
    }
  },
  "timestamp": 1640995200,
  "request_id": "req_123456"
}
```

### 分页响应

```json
{
  "success": true,
  "code": 200,
  "message": "查询成功",
  "data": [
    {"id": 1, "name": "John"},
    {"id": 2, "name": "Jane"}
  ],
  "pagination": {
    "page": 1,
    "page_size": 10,
    "total": 100,
    "total_pages": 10,
    "has_next": true,
    "has_prev": false
  },
  "timestamp": 1640995200,
  "request_id": "req_123456"
}
```

## 错误类型

### 预定义错误类型

```go
const (
    ErrorTypeValidation    = "validation_error"      // 验证错误
    ErrorTypeBusiness      = "business_error"        // 业务错误
    ErrorTypeAuthentication = "authentication_error" // 认证错误
    ErrorTypeAuthorization = "authorization_error"   // 授权错误
    ErrorTypeNotFound      = "not_found_error"       // 未找到错误
    ErrorTypeConflict      = "conflict_error"        // 冲突错误
    ErrorTypeInternal      = "internal_error"        // 内部错误
    ErrorTypeExternal      = "external_error"        // 外部服务错误
    ErrorTypeDatabase      = "database_error"        // 数据库错误
    ErrorTypeCache         = "cache_error"           // 缓存错误
    ErrorTypeTimeout       = "timeout_error"         // 超时错误
    ErrorTypeRateLimit     = "rate_limit_error"      // 限流错误
)
```

### 业务错误创建

```go
// 用户相关错误
userNotFoundErr := response.NewUserNotFoundError(123)
userExistsErr := response.NewUserExistsError("john@example.com")
invalidCredentialsErr := response.NewInvalidCredentialsError()

// 比赛相关错误
matchNotFoundErr := response.NewMatchNotFoundError(456)
matchStartedErr := response.NewMatchStartedError(456)

// 预测相关错误
predictionExistsErr := response.NewPredictionExistsError(123, 456)
voteExistsErr := response.NewVoteExistsError(123, 789)

// 系统错误
dbErr := response.NewDatabaseError("select users", originalErr)
cacheErr := response.NewCacheError("get user", originalErr)
timeoutErr := response.NewTimeoutError("database query")
```

### 自定义错误

```go
// 创建自定义应用错误
customErr := &response.AppError{
    Type:       "custom_error",
    Code:       "CUSTOM_ERROR_CODE",
    Message:    "自定义错误消息",
    Details:    map[string]interface{}{"key": "value"},
    StatusCode: 422,
}

// 添加堆栈信息
customErr.WithStack()

// 添加原因错误
customErr.WithCause(originalErr)

// 包装现有错误
wrappedErr := response.WrapError(originalErr, "business_error", "OPERATION_FAILED", "操作失败", 500)
```

## 错误检查

### 错误类型检查

```go
if response.IsAppError(err) {
    appErr := err.(*response.AppError)
    // 处理应用错误
}

if response.IsValidationError(err) {
    // 处理验证错误
}

if response.IsBusinessError(err) {
    // 处理业务错误
}

if response.IsNotFoundError(err) {
    // 处理未找到错误
}

if response.IsDatabaseError(err) {
    // 处理数据库错误
}
```

### 错误代码检查

```go
if response.IsErrorCode(err, response.CodeUserNotFound) {
    // 处理用户不存在错误
}

if response.IsErrorType(err, response.ErrorTypeAuthentication) {
    // 处理认证错误
}
```

## 中间件

### 错误处理中间件

```go
import "backend-go/pkg/middleware"

// 全局错误处理
r.Use(middleware.GlobalErrorHandler())

// 自定义错误处理
errorHandler := middleware.NewErrorHandler(
    middleware.WithLogger(logger),
    middleware.WithStackTrace(true),
    middleware.WithRecover(true),
)
r.Use(errorHandler.ErrorHandlerMiddleware())

// 恢复中间件
r.Use(middleware.RecoveryMiddleware(logger))
```

### 请求ID中间件

```go
// 默认请求ID中间件（UUID格式）
r.Use(middleware.DefaultRequestIDMiddleware())

// 短请求ID中间件
r.Use(middleware.ShortRequestIDMiddleware())

// 自定义请求ID生成器
generator := &middleware.TimestampIDGenerator{}
r.Use(middleware.RequestIDMiddleware(generator))

// 获取请求ID
requestID := middleware.GetRequestID(c)
```

### 日志中间件

```go
// 访问日志中间件
r.Use(middleware.AccessLogMiddleware(logger))

// 结构化日志中间件
r.Use(middleware.StructuredLoggingMiddleware(logger))

// 错误日志中间件
r.Use(middleware.ErrorLoggingMiddleware(logger))

// 慢请求日志中间件
r.Use(middleware.SlowRequestLoggingMiddleware(logger, 1*time.Second))
```

## 响应写入器

### 使用响应写入器接口

```go
func HandleRequest(c *gin.Context) {
    writer := response.NewGinResponseWriter(c)
    
    user, err := getUserByID(123)
    if err != nil {
        if response.IsNotFoundError(err) {
            writer.NotFound("用户不存在")
        } else {
            writer.InternalError("获取用户失败")
        }
        return
    }
    
    writer.Success(user)
}
```

### 中断响应

```go
func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        token := c.GetHeader("Authorization")
        if token == "" {
            response.AbortWithUnauthorized(c, "缺少认证令牌")
            return
        }
        
        if !validateToken(token) {
            response.AbortWithUnauthorized(c, "无效的认证令牌")
            return
        }
        
        c.Next()
    }
}
```

## 健康检查和指标

### 健康检查

```go
func HealthCheck(c *gin.Context) {
    details := map[string]interface{}{
        "database": checkDatabase(),
        "redis":    checkRedis(),
        "version":  "1.0.0",
    }
    
    status := "healthy"
    if !allServicesHealthy(details) {
        status = "unhealthy"
    }
    
    response.HealthCheck(c, status, details)
}
```

### 指标响应

```go
func Metrics(c *gin.Context) {
    metrics := map[string]interface{}{
        "requests_total":   getRequestCount(),
        "error_rate":       getErrorRate(),
        "response_time":    getAvgResponseTime(),
    }
    
    response.Metrics(c, metrics)
}
```

## 配置示例

### 完整的中间件配置

```go
func setupMiddleware(r *gin.Engine, logger *logrus.Logger) {
    // 请求ID
    r.Use(middleware.DefaultRequestIDMiddleware())
    
    // 日志记录
    r.Use(middleware.AccessLogMiddleware(logger))
    
    // 错误处理和恢复
    r.Use(middleware.RecoveryMiddleware(logger))
    r.Use(middleware.GlobalErrorHandler())
    
    // 错误报告
    logReporter := middleware.NewLogErrorReporter(logger)
    r.Use(middleware.ErrorReportingMiddleware(logReporter))
    
    // 验证错误处理
    r.Use(middleware.ValidationErrorHandler())
    
    // 404和405处理
    r.NoRoute(middleware.NotFoundHandler())
    r.NoMethod(middleware.MethodNotAllowedHandler())
}
```

### 开发环境配置

```go
func setupDevelopment(r *gin.Engine) {
    logger := logrus.New()
    logger.SetLevel(logrus.DebugLevel)
    logger.SetFormatter(&logrus.TextFormatter{})
    
    errorHandler := middleware.NewErrorHandler(
        middleware.WithLogger(logger),
        middleware.WithStackTrace(true),
        middleware.WithRecover(true),
    )
    
    r.Use(middleware.DefaultRequestIDMiddleware())
    r.Use(middleware.AccessLogMiddleware(logger))
    r.Use(errorHandler.ErrorHandlerMiddleware())
}
```

### 生产环境配置

```go
func setupProduction(r *gin.Engine) {
    logger := logrus.New()
    logger.SetLevel(logrus.InfoLevel)
    logger.SetFormatter(&logrus.JSONFormatter{})
    
    errorHandler := middleware.NewErrorHandler(
        middleware.WithLogger(logger),
        middleware.WithStackTrace(false), // 生产环境不显示堆栈
        middleware.WithRecover(true),
        middleware.WithSkipPaths([]string{"/health", "/metrics"}),
    )
    
    r.Use(middleware.DefaultRequestIDMiddleware())
    r.Use(middleware.StructuredLoggingMiddleware(logger))
    r.Use(errorHandler.ErrorHandlerMiddleware())
}
```

## 最佳实践

### 1. 错误处理

```go
// ✅ 好的做法
func GetUser(c *gin.Context) {
    id, err := strconv.Atoi(c.Param("id"))
    if err != nil {
        response.BadRequest(c, "无效的用户ID", map[string]string{"id": c.Param("id")})
        return
    }
    
    user, err := userService.GetUser(id)
    if err != nil {
        switch {
        case response.IsNotFoundError(err):
            response.NotFound(c, err.Error())
        case response.IsDatabaseError(err):
            response.InternalError(c, "数据库查询失败")
        default:
            response.InternalError(c, "获取用户失败")
        }
        return
    }
    
    response.Success(c, user)
}

// ❌ 不好的做法
func GetUserBad(c *gin.Context) {
    user, err := userService.GetUser(123)
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()}) // 不统一的响应格式
        return
    }
    
    c.JSON(200, user) // 缺少统一的响应结构
}
```

### 2. 业务错误处理

```go
// ✅ 使用预定义的业务错误
func CreateUser(c *gin.Context) {
    var user User
    if err := c.ShouldBindJSON(&user); err != nil {
        response.ValidationError(c, "参数验证失败", err.Error())
        return
    }
    
    if exists, _ := userService.UserExists(user.Email); exists {
        response.Conflict(c, "用户已存在", map[string]string{"email": user.Email})
        return
    }
    
    createdUser, err := userService.CreateUser(&user)
    if err != nil {
        response.InternalError(c, "创建用户失败")
        return
    }
    
    response.Created(c, createdUser)
}
```

### 3. 中间件使用

```go
// ✅ 正确的中间件顺序
func setupRouter() *gin.Engine {
    r := gin.New()
    
    // 1. 请求ID（最先）
    r.Use(middleware.DefaultRequestIDMiddleware())
    
    // 2. 日志记录
    r.Use(middleware.AccessLogMiddleware(logger))
    
    // 3. 恢复和错误处理
    r.Use(middleware.RecoveryMiddleware(logger))
    r.Use(middleware.GlobalErrorHandler())
    
    // 4. 业务中间件
    r.Use(authMiddleware())
    
    return r
}
```

### 4. 错误日志记录

```go
// ✅ 结构化错误日志
func HandleDatabaseError(err error, operation string) error {
    logger.WithFields(logrus.Fields{
        "error":     err.Error(),
        "operation": operation,
        "timestamp": time.Now(),
    }).Error("Database operation failed")
    
    return response.NewDatabaseError(operation, err)
}
```

## 测试

### 单元测试

```go
func TestSuccessResponse(t *testing.T) {
    w := httptest.NewRecorder()
    c, _ := gin.CreateTestContext(w)
    
    testData := map[string]interface{}{"id": 1, "name": "test"}
    response.Success(c, testData)
    
    assert.Equal(t, http.StatusOK, w.Code)
    
    var resp response.Response
    err := json.Unmarshal(w.Body.Bytes(), &resp)
    require.NoError(t, err)
    
    assert.True(t, resp.Success)
    assert.Equal(t, http.StatusOK, resp.Code)
    assert.NotNil(t, resp.Data)
}
```

### 集成测试

```go
func TestUserAPI(t *testing.T) {
    router := setupTestRouter()
    
    // 测试创建用户
    user := User{Username: "test", Email: "test@example.com"}
    body, _ := json.Marshal(user)
    
    w := httptest.NewRecorder()
    req, _ := http.NewRequest("POST", "/api/users", bytes.NewBuffer(body))
    req.Header.Set("Content-Type", "application/json")
    
    router.ServeHTTP(w, req)
    
    assert.Equal(t, http.StatusCreated, w.Code)
    
    var resp response.Response
    json.Unmarshal(w.Body.Bytes(), &resp)
    assert.True(t, resp.Success)
}
```

## 故障排除

### 常见问题

1. **响应格式不一致**
   - 确保所有API都使用response包的函数
   - 避免直接使用gin的JSON方法

2. **错误信息泄露**
   - 生产环境不要显示详细的错误堆栈
   - 使用适当的错误消息

3. **日志过多**
   - 配置适当的日志级别
   - 使用跳过路径避免记录健康检查等请求

4. **性能问题**
   - 避免在热路径上记录详细日志
   - 使用异步日志记录

### 调试技巧

1. **启用详细日志**
   ```go
   logger.SetLevel(logrus.DebugLevel)
   ```

2. **启用堆栈跟踪**
   ```go
   errorHandler := middleware.NewErrorHandler(
       middleware.WithStackTrace(true),
   )
   ```

3. **检查请求ID**
   ```go
   requestID := middleware.GetRequestID(c)
   logger.WithField("request_id", requestID).Info("Processing request")
   ```

## 依赖

- [Gin](https://github.com/gin-gonic/gin) - HTTP Web 框架
- [Logrus](https://github.com/sirupsen/logrus) - 结构化日志库