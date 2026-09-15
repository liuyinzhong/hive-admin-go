-- 登出与令牌撤销：已撤销凭证黑名单表。
-- 登出时写入凭证唯一标识 jti 及其原过期时刻，认证中间件按 jti 拦截；
-- 过期行由审计日志清理任务（CleanupExpiredLogs）每日顺带删除，无独立清理任务。
-- 执行本脚本前签发的旧 token 不携带 jti（claims 中为空），不参与黑名单比对，
-- 且更换签名密钥后旧 token 均已失效，无存量兼容问题。
CREATE TABLE `sys_token_blacklist` (
    `jti`         char(36)     NOT NULL COMMENT '已撤销凭证的唯一标识（JWT jti）',
    `expires_at`  datetime     NOT NULL COMMENT '凭证原过期时刻，过期后该行由清理任务删除',
    `create_date` datetime     NOT NULL COMMENT '撤销写入时间',
    PRIMARY KEY (`jti`),
    KEY `idx_expires_at` (`expires_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='已撤销登录凭证黑名单';
