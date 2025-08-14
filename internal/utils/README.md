# JWT 工具包

这个工具包提供了完整的 JWT (JSON Web Token) 功能，包括 token 生成、解析、验证和中间件。

## 功能特性

- ✅ JWT Token 生成
- ✅ JWT Token 解析和验证
- ✅ Token 信息提取
- ✅ Token 刷新
- ✅ Gin 中间件支持
- ✅ 角色权限控制
- ✅ CORS 支持

## 文件结构

```
internal/utils/
├── jwt.go          # JWT 核心功能
├── middleware.go   # Gin 中间件
├── jwt_test.go     # 单元测试
├── example.go      # 使用示例
└── README.md       # 说明文档
```

## 快速开始

### 1. 创建 JWT 管理器

```go
import "video-crawler/internal/utils"

// 创建 JWT 管理器
secretKey := "your-secret-key-here"
tokenDuration := 24 * time.Hour
jwtManager := utils.NewJWTManager(secretKey, tokenDuration)
```

### 2. 生成 Token

```go
// 生成 JWT token
token, err := jwtManager.GenerateToken(123, "john_doe", "admin")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Token: %s\n", token)
```

### 3. 解析 Token

```go
// 解析 JWT token
claims, err := jwtManager.ParseToken(token)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("用户ID: %d\n", claims.UserID)
fmt.Printf("用户名: %s\n", claims.Username)
fmt.Printf("角色: %s\n", claims.Role)
```

### 4. 验证 Token

```go
// 验证 token 是否有效
isValid := jwtManager.ValidateToken(token)
if isValid {
    fmt.Println("Token 有效")
} else {
    fmt.Println("Token 无效")
}
```

### 5. 获取 Token 信息

```go
// 获取 token 中的所有信息
info, err := jwtManager.GetTokenInfo(token)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Token 信息: %+v\n", info)
```

## Gin 中间件使用

### 1. 必需认证中间件

```go
import (
    "github.com/gin-gonic/gin"
    "video-crawler/internal/utils"
)

func main() {
    r := gin.Default()
    
    // 创建 JWT 管理器
    jwtManager := utils.NewJWTManager("secret-key", 24*time.Hour)
    
    // 创建需要认证的路由组
    authorized := r.Group("/api")
    authorized.Use(utils.JWTAuthMiddleware(jwtManager))
    {
        authorized.GET("/profile", func(c *gin.Context) {
            userID := c.GetUint("user_id")
            username := c.GetString("username")
            role := c.GetString("role")
            
            c.JSON(200, gin.H{
                "user_id":  userID,
                "username": username,
                "role":     role,
            })
        })
    }
    
    r.Run(":8080")
}
```

### 2. 可选认证中间件

```go
// 可选认证 - 不强制要求 token
public := r.Group("/public")
public.Use(utils.OptionalJWTAuthMiddleware(jwtManager))
{
    public.GET("/info", func(c *gin.Context) {
        if userID, exists := c.Get("user_id"); exists {
            // 用户已登录
            c.JSON(200, gin.H{"message": "已登录用户", "user_id": userID})
        } else {
            // 匿名用户
            c.JSON(200, gin.H{"message": "匿名用户"})
        }
    })
}
```

### 3. 角色权限中间件

```go
// 需要管理员权限的路由
admin := r.Group("/admin")
admin.Use(utils.JWTAuthMiddleware(jwtManager))
admin.Use(utils.RoleAuthMiddleware("admin"))
{
    admin.GET("/dashboard", func(c *gin.Context) {
        c.JSON(200, gin.H{"message": "管理员面板"})
    })
}

// 需要多个角色之一的路由
moderator := r.Group("/moderate")
moderator.Use(utils.JWTAuthMiddleware(jwtManager))
moderator.Use(utils.RoleAuthMiddleware("admin", "moderator"))
{
    moderator.POST("/approve", func(c *gin.Context) {
        c.JSON(200, gin.H{"message": "内容审核"})
    })
}
```

### 4. CORS 中间件

```go
// 添加 CORS 支持
r.Use(utils.CORSMiddleware())
```

## API 参考

### JWTManager

#### 方法

- `NewJWTManager(secretKey string, tokenDuration time.Duration) *JWTManager`
  - 创建新的 JWT 管理器

- `GenerateToken(userID uint, username, role string) (string, error)`
  - 生成 JWT token

- `ParseToken(tokenString string) (*JWTClaims, error)`
  - 解析 JWT token

- `ValidateToken(tokenString string) bool`
  - 验证 token 是否有效

- `RefreshToken(tokenString string) (string, error)`
  - 刷新 token（仅在即将过期时）

- `GetTokenInfo(tokenString string) (map[string]interface{}, error)`
  - 获取 token 中的所有信息

### 中间件

- `JWTAuthMiddleware(jwtManager *JWTManager) gin.HandlerFunc`
  - 必需认证中间件

- `OptionalJWTAuthMiddleware(jwtManager *JWTManager) gin.HandlerFunc`
  - 可选认证中间件

- `RoleAuthMiddleware(requiredRoles ...string) gin.HandlerFunc`
  - 角色权限中间件

- `CORSMiddleware() gin.HandlerFunc`
  - CORS 中间件

### 工具函数

- `ExtractTokenFromHeader(authHeader string) (string, error)`
  - 从 Authorization header 中提取 token

## 测试

运行单元测试：

```bash
go test ./internal/utils -v
```

运行示例程序：

```bash
go run cmd/jwt-example/main.go
```

## 配置建议

### 1. 密钥管理

```go
// 从环境变量读取密钥
secretKey := os.Getenv("JWT_SECRET_KEY")
if secretKey == "" {
    secretKey = "default-secret-key" // 开发环境默认值
}
```

### 2. Token 过期时间

```go
// 根据环境设置不同的过期时间
var tokenDuration time.Duration
if os.Getenv("ENV") == "production" {
    tokenDuration = 1 * time.Hour // 生产环境 1 小时
} else {
    tokenDuration = 24 * time.Hour // 开发环境 24 小时
}
```

### 3. 错误处理

```go
// 自定义错误处理
func handleJWTAuthError(c *gin.Context, err error) {
    switch {
    case strings.Contains(err.Error(), "token is expired"):
        c.JSON(401, gin.H{"error": "Token 已过期"})
    case strings.Contains(err.Error(), "token is malformed"):
        c.JSON(401, gin.H{"error": "Token 格式错误"})
    default:
        c.JSON(401, gin.H{"error": "认证失败"})
    }
}
```

## 安全注意事项

1. **密钥安全**: 使用强密钥并妥善保管
2. **HTTPS**: 生产环境必须使用 HTTPS
3. **Token 存储**: 客户端安全存储 token
4. **过期时间**: 设置合理的 token 过期时间
5. **刷新机制**: 实现 token 刷新机制
6. **注销**: 实现 token 注销机制（黑名单）

## 示例项目

查看 `cmd/jwt-example/main.go` 获取完整的使用示例。
