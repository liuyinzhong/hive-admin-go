# utils 目录规则

本目录是项目公共工具。新增或修改本目录代码前必须先读本文件；在其它目录使用这里的短时 token 公共方法前，先读「短时 token 与临时凭证」。

## 短时 token 与临时凭证

凡是需要“已登录用户把资源临时暴露给非登录方”或“敏感操作需要短期可验证凭证”的场景，统一使用 `utils.SignShortLivedToken` 和 `utils.ParseShortLivedToken`，不要在每个 Service 里复制 `jwt.NewWithClaims` + `SignedString` 或 `jwt.ParseWithClaims` 的样板代码。

适用场景：

- 私有文件预览/下载授权给外部服务（如 kkFileView 取文件、PDF.js 渲染）。
- 邮箱验证、密码重置链接（payload 含 userId + 用途标记，有效期 15~30 分钟）。
- 跨服务调用临时凭证（避免传递长期密钥）。
- 回调 URL 签名（含时间戳防重放）。
- 敏感操作二次确认 token（如批量删除、转账）。

使用方式：

1. 调用方定义自己的 `Claims` 结构体，嵌入 `jwt.RegisteredClaims`，自行设置 `ExpiresAt`、`IssuedAt`、`Issuer` 和业务字段。
2. 签发：`signed, err := utils.SignShortLivedToken(claims)`，复用项目 JWT 密钥，无需额外配置。
3. 校验：`claims := &MyClaims{}; err := utils.ParseShortLivedToken(tokenString, claims)`，解析成功后由调用方做业务校验。
4. 公开取资源接口放在 `/api/public/` 下，不挂 `AuthMiddleware`；公开接口只验 token 签名和时效，不验证调用方身份。

安全边界（强制）：

- token 在有效期内可被任何持有者重放使用，**调用方必须**在 payload 中锁定资源标识（如 `taskId+userId`），并在取资源时复用对应的业务校验（如 `creator_id` 校验），确保 token 泄露后只能访问签发时锁定的那一个资源。
- 有效期按业务最小化原则设置：文件预览 1~5 分钟、邮件验证 15~30 分钟、回调签名 5 分钟内。
- 需要真正一次性失效时由调用方自行引入数据库存储（如 `sys_*_token` 表加 `used` 字段），公共方法不承担一次性语义。
- token payload 不得包含密码、密钥、完整 Authorization Header、连接串或文件绝对路径；只放资源标识和用途标记。
- 公开取资源接口不得返回与签发资源无关的字段，不得顺带返回用户 Token 或其它敏感信息。
- 业务错误统一返回 `models.NewErrorResponse()`，token 无效/过期统一映射为业务错误码，不暴露 JWT 内部错误细节。

现有实现：

- 下载中心预览：`services/download_task_service.go` 的 `GeneratePreviewURL` / `GetPreviewFile`，`DownloadPreviewClaims` 含 `taskId+userId`，有效期 5 分钟，公开路由 `/api/public/downloads/preview/:token`。详见 `business-docs/system/download-center.md`。

新增场景必须复用本节公共方法，不得在 Service 中重复实现 JWT 签发/解析样板；并在交付说明中列出 token payload 字段、有效期、公开路由和业务校验依据。
