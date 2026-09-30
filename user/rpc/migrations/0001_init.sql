-- 初始化 schema：与 internal/dal/model/*.gen.go 的生成结果保持一致。
-- docker-compose 会把本目录挂到 /docker-entrypoint-initdb.d/，仅在数据卷为空时执行一次。

CREATE TABLE IF NOT EXISTS users (
    id         BIGSERIAL     PRIMARY KEY,
    username   VARCHAR(64)   NOT NULL,
    email      VARCHAR(128)  NOT NULL,
    password   VARCHAR(256)  NOT NULL DEFAULT '',
    role       VARCHAR(32)   NOT NULL DEFAULT 'user',
    created_at TIMESTAMPTZ   NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ   NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ
);

-- 唯一索引带 deleted_at IS NULL 谓词：软删后的用户名/邮箱可以重新使用
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_username ON users (username) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email    ON users (email)    WHERE deleted_at IS NULL;
CREATE INDEX        IF NOT EXISTS idx_users_deleted_at ON users (deleted_at);

CREATE TABLE IF NOT EXISTS todos (
    id         BIGSERIAL     PRIMARY KEY,
    user_id    BIGINT        NOT NULL,
    title      VARCHAR(256)  NOT NULL,
    content    TEXT          NOT NULL DEFAULT '',
    status     SMALLINT      NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ   NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ   NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_todos_user_id    ON todos (user_id);
CREATE INDEX IF NOT EXISTS idx_todos_status     ON todos (status);
CREATE INDEX IF NOT EXISTS idx_todos_deleted_at ON todos (deleted_at);

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_todos_user') THEN
        ALTER TABLE todos
            ADD CONSTRAINT fk_todos_user FOREIGN KEY (user_id) REFERENCES users (id);
    END IF;
END
$$;
