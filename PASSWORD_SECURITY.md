# 密码安全机制说明

## 概述

为了在请求日志中保护用户密码安全，系统采用了前端MD5加密 + 后端Salt处理的双重安全机制。

## 安全流程

### 1. 前端密码处理
- 用户在登录/注册页面输入明文密码
- 前端使用 `crypto-js` 库对密码进行MD5加密
- 加密后的密码通过HTTP请求发送到后端
- **优势**: 请求日志中不会暴露明文密码

### 2. 后端密码处理
- 后端接收前端传来的MD5密码
- 生成随机Salt（盐值）
- 将MD5密码与Salt组合，再次进行MD5加密
- 存储格式: `MD5(MD5(明文密码) + ":" + Salt)`

### 3. 密码验证流程
- 用户登录时，前端对输入密码进行MD5加密
- 后端使用存储的Salt对MD5密码进行二次加密
- 比较加密结果与存储的密码哈希

## 技术实现

### 前端实现
```typescript
import MD5 from 'crypto-js/md5'

// 登录时
const encryptedPassword = MD5(password).toString()
const response = await fetch('/api/user/login', {
  method: 'POST',
  body: JSON.stringify({ username, password: encryptedPassword })
})

// 注册时
const encryptedPassword = MD5(password).toString()
const response = await fetch('/api/user/register', {
  method: 'POST',
  body: JSON.stringify({ username, nickname, password: encryptedPassword })
})
```

### 后端实现
```go
// 注册时
salt := uuid.New().String()
password = utils.SaltedMd5PasswordFromMd5(md5Password, salt)

// 登录验证时
if userEntity.Password != utils.SaltedMd5PasswordFromMd5(password, userEntity.Salt) {
    // 密码错误
}
```

## 安全优势

1. **日志安全**: 请求日志中只显示MD5哈希，不暴露明文密码
2. **双重加密**: 即使MD5被破解，还有Salt保护
3. **Salt随机性**: 每个用户的Salt都是唯一的
4. **彩虹表防护**: Salt机制有效防止彩虹表攻击

## 测试示例

### 测试密码
- 明文: `admin`
- MD5: `21232f297a57a5a743894a0e4a801fc3`

- 明文: `123456`
- MD5: `e10adc3949ba59abbe56e057f20f883e`

### API测试
```bash
# 注册测试
curl -X POST http://localhost:8080/api/user/register \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","nickname":"测试用户","password":"e10adc3949ba59abbe56e057f20f883e"}'

# 登录测试
curl -X POST http://localhost:8080/api/user/login \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","password":"e10adc3949ba59abbe56e057f20f883e"}'
```

## 注意事项

1. **兼容性**: 此机制仅适用于新注册的用户
2. **测试数据**: 当前为测试环境，不影响生产数据
3. **HTTPS建议**: 生产环境建议使用HTTPS传输
4. **密码强度**: 建议用户使用强密码，前端可添加密码强度检查

## 依赖库

### 前端依赖
```json
{
  "crypto-js": "^4.2.0",
  "@types/crypto-js": "^4.2.2"
}
```

### 安装命令
```bash
npm install crypto-js @types/crypto-js
```
