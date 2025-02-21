# bytedance
订单服务
# 订单管理系统

## 项目简介
这是一个基于 Hertz 框架开发的订单管理系统，提供订单的创建、更新、查询等功能。使用 MySQL 作为数据存储，GORM 作为 ORM 框架。系统支持订单的自动过期处理，并使用 JSON 格式存储地址和订单项信息。

## 技术栈
- 框架：Hertz (CloudWeGo)
- 数据库：MySQL
- ORM：GORM v1.25.12
- 其他：
  - github.com/google/uuid v1.6.0
  - gorm.io/driver/mysql v1.5.7

## 主要功能
- 订单管理
  - 创建订单（支持地址和订单项的JSON存储）
  - 更新订单信息
  - 查询订单详情
- 自动化处理
  - 订单过期自动取消（1分钟后过期）
  - 定期检查过期订单（每分钟检查一次）

## API接口

### 创建订单
- **POST** `/orders`
```json
{
    "user_id": "user123",
    "user_currency": "CNY",
    "address": {
        "street": "示例街道",
        "city": "示例城市"
    },
    "email": "example@email.com",
    "order_items": [
        {
            "product_id": "prod123",
            "quantity": 1
        }
    ]
}
```

### 更新订单
- **PUT** `/orders/:orderID`
```json
{
    "status": "completed",
    "email": "newemail@example.com"
}
```

### 查询订单
- **GET** `/orders/:orderID`

## 数据库设计
订单表 (`orders`):
- order_id: VARCHAR(36) - 订单ID (主键)
- user_id: VARCHAR(36) - 用户ID
- address: TEXT - 地址信息 (JSON格式)
- email: VARCHAR(255) - 邮箱
- order_items: TEXT - 订单项 (JSON格式)
- currency: VARCHAR(10) - 货币类型
- status: VARCHAR(20) - 订单状态
- created_at: TIMESTAMP - 创建时间
- updated_at: TIMESTAMP - 更新时间
- expire_at: TIMESTAMP - 过期时间

## 快速开始

### 环境要求
- Go 1.23.5+
- MySQL 5.7+

### 安装步骤
1. 克隆项目
2. 执行数据库初始化
```sql
# 执行 dingdan.sql 文件中的SQL语句
```

3. 配置数据库连接
```go
dsn := "root:root@tcp(127.0.0.1:3306)/dingdan?charset=utf8mb4&parseTime=True&loc=Local"
```

4. 安装依赖
```bash
go mod tidy
```

5. 运行项目
```bash
go run main.go
```

## 项目结构
```
.
├── config/
│   └── database.go    # 数据库配置
├── model/
│   └── order.go       # 订单模型
├── service/
│   └── order.go       # 订单服务逻辑
├── main.go            # 主程序入口
├── go.mod             # Go模块文件
├── dingdan.sql        # 数据库初始化SQL
└── README.md          # 项目说明文档
```

## 注意事项
1. 数据库配置
   - 默认使用utf8mb4编码
   - 连接池配置：最大空闲连接10个，最大连接数100个
   - 连接生命周期：1小时

2. 订单处理
   - 订单创建后1分钟自动过期
   - 系统每分钟自动检查过期订单
   - 过期订单状态自动更新为"cancelled"

3. 生产环境建议
   - 调整数据库连接池参数
   - 根据业务需求修改订单过期时间
   - 添加适当的日志记录
   - 实现更完善的错误处理机制
