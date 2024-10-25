# GitHub 开发者评估系统

基于 GitHub 的开源项目数据，开发的开发者评估应用。通过分析开发者的项目贡献、技术栈和影响力，提供全面的开发者画像。

## 功能特性

### 基础功能

1. **TalentRank**: 开发者技术能力评价系统
   - 项目重要度评估（Star、Fork、代码量）
   - 贡献度分析（Commit 数量、质量）
   - 技术栈深度评估
   - 活跃度和持续性分析

2. **Nation 预测**: 开发者地理位置智能推断
   - 基于 Profile 信息分析
   - 代码提交时间分布分析
   - 项目描述语言特征分析
   - 多维度交叉验证

3. **开发者搜索**: 
   - 技术栈精准匹配
   - TalentRank 排序
   - Nation 筛选
   - 活跃度过滤

### 高级功能

1. **数据置信度评估**
   - 数据完整性验证
   - 预测结果可信度计算
   - 低置信数据标记（N/A）

2. **智能信息整合**
   - GitHub 博客解析
   - 个人网站信息提取
   - 项目描述分析
   - 技术栈自动归类

## 命令行工具

### 1. 环境配置

设置环境变量：

```bash
# 设置 GitHub API Token（必需）
export GITHUB_TOKEN="your_github_token"

# 设置 MongoDB 连接字符串（可选，默认 mongodb://localhost:27017）
export MONGO_URI="mongodb://username:password@localhost:27017/dbname"

# 设置 Redis 连接字符串（可选，用于缓存）
export REDIS_URL="redis://localhost:6379"
```

### 2. 爬虫命令

#### 单个用户分析

分析单个 GitHub 用户的信息：

```bash
go run cmd/crawler/main.go -users torvalds
```

#### 批量用户分析

从文件中读取用户名列表并批量分析：

```bash
go run cmd/crawler/main.go -users "torvalds,antirez,marmotedu" -concurrency 3
```

#### 命令行参数说明

```bash
# 查看帮助
go run cmd/crawler/main.go -h
```

选项说明：

| 选项                | 类型  | 默认值 | 说明                           |
|---------------------|-------|--------|--------------------------------|
| `-users`            | string|        | 指定单个或多个用户名（逗号分隔） |
| `-concurrency`      | int   | 5      | 并发数量（默认 5）               |

## API 接口文档

### 1. 健康检查

```http
GET /health
```

用于检查服务器是否正常运行。

#### 响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": "OK"
}
```

### 2. 获取所有开发者

```http
GET /api/developers
```

获取所有开发者的信息。

#### 查询参数

| 参数     | 类型 | 必需 | 描述         |
|----------|------|------|--------------|
| `page`   | int  | 否   | 页码，默认1    |
| `per_page`| int | 否   | 每页数量，默认20 |

#### 响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "username": "torvalds",
      "name": "Linus Torvalds",
      "avatar": "https://avatars.githubusercontent.com/u/1024025",
      "location": "Portland, OR",
      "nation": "US",
      "nation_confidence": 95.5,
      "skills": ["C", "Shell", "Perl"],
      "metrics": {
        "star_count": 145200,
        "fork_count": 42300,
        "commit_count": 8750
      },
      "repositories": ["linux", "subsurface", "uemacs"],
      "talent_rank": 98.7,
      "confidence": 99.9,
      "updated_at": "2024-01-20T10:30:00Z"
    },
    // ... 其他开发者信息
  ]
}
```

### 3. 获取单个开发者信息

```http
GET /api/developers/{id}
```

获取指定 ID 的开发者信息。

#### 路径参数

| 参数   | 类型 | 必需 | 描述           |
|--------|------|------|----------------|
| `id`   | string | 是   | 开发者的唯一 ID |

#### 响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "username": "torvalds",
    "name": "Linus Torvalds",
    "avatar": "https://avatars.githubusercontent.com/u/1024025",
    "location": "Portland, OR",
    "nation": "US",
    "nation_confidence": 95.5,
    "skills": ["C", "Shell", "Perl"],
    "metrics": {
      "star_count": 145200,
      "fork_count": 42300,
      "commit_count": 8750
    },
    "repositories": ["linux", "subsurface", "uemacs"],
    "talent_rank": 98.7,
    "confidence": 99.9,
    "updated_at": "2024-01-20T10:30:00Z"
  }
}
```

### 4. 搜索开发者

```http
GET /api/search
```

根据条件搜索开发者。

#### 查询参数

| 参数             | 类型   | 必需 | 描述                       | 示例           |
|------------------|--------|------|----------------------------|----------------|
| `q`              | string | 否   | 搜索关键词                  | golang         |
| `skills`         | string | 否   | 技能过滤（逗号分隔）         | go,python      |
| `nation`         | string | 否   | 国家/地区代码               | CN             |
| `min_stars`      | int    | 否   | 最小 star 数                | 1000           |
| `min_rank`       | float  | 否   | 最小 TalentRank 分数         | 80             |
| `page`           | int    | 否   | 页码（从1开始）             | 1              |
| `per_page`       | int    | 否   | 每页数量                    | 20             |
| `sort`           | string | 否   | 排序字段（如 talent_rank） | talent_rank    |
| `order`          | string | 否   | 排序方向（asc 或 desc）    | desc           |

#### 响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "total": 100,
    "page": 1,
    "per_page": 20,
    "developers": [
      {
        "username": "example",
        "name": "Example User",
        "avatar": "https://avatars.githubusercontent.com/u/...",
        "location": "Beijing, China",
        "nation": "CN",
        "nation_confidence": 95.5,
        "skills": ["Go", "Python", "JavaScript"],
        "metrics": {
          "star_count": 1200,
          "fork_count": 300,
          "commit_count": 5000
        },
        "repositories": ["repo1", "repo2"],
        "talent_rank": 85.6,
        "confidence": 92.3,
        "updated_at": "2024-01-20T10:30:00Z"
      },
      // ... 其他开发者
    ]
  }
}
```

### 5. 创建开发者

```http
POST /api/developers
```

创建新的开发者信息（需要认证）。

#### 请求体

```json
{
  "username": "example",
  "name": "Example User",
  "avatar": "https://avatars.githubusercontent.com/u/...",
  "location": "Beijing, China",
  "nation": "CN",
  "skills": ["Go", "Python", "JavaScript"],
  "metrics": {
    "star_count": 1200,
    "fork_count": 300,
    "commit_count": 5000
  },
  "repositories": ["repo1", "repo2"],
  "talent_rank": 85.6,
  "confidence": 92.3
}
```

#### 响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": "12345",
    "username": "example",
    "name": "Example User",
    "avatar": "https://avatars.githubusercontent.com/u/...",
    "location": "Beijing, China",
    "nation": "CN",
    "nation_confidence": 95.5,
    "skills": ["Go", "Python", "JavaScript"],
    "metrics": {
      "star_count": 1200,
      "fork_count": 300,
      "commit_count": 5000
    },
    "repositories": ["repo1", "repo2"],
    "talent_rank": 85.6,
    "confidence": 92.3,
    "updated_at": "2024-01-20T10:30:00Z"
  }
}
```

### 6. 更新开发者信息

```http
PUT /api/developers/{id}
```

更新指定 ID 的开发者信息（需要认证）。

#### 路径参数

| 参数 | 类型   | 必需 | 描述           |
|------|--------|------|----------------|
| `id` | string | 是   | 开发者的唯一 ID |

#### 请求体

```json
{
  "name": "New Name",
  "avatar": "https://avatars.githubusercontent.com/u/...",
  "location": "Shanghai, China",
  "nation": "CN",
  "skills": ["Go", "Rust"],
  "metrics": {
    "star_count": 1300,
    "fork_count": 350,
    "commit_count": 5500
  },
  "repositories": ["repo1", "repo3"],
  "talent_rank": 88.2,
  "confidence": 93.5
}
```

#### 响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": "12345",
    "username": "example",
    "name": "New Name",
    "avatar": "https://avatars.githubusercontent.com/u/...",
    "location": "Shanghai, China",
    "nation": "CN",
    "nation_confidence": 95.5,
    "skills": ["Go", "Rust"],
    "metrics": {
      "star_count": 1300,
      "fork_count": 350,
      "commit_count": 5500
    },
    "repositories": ["repo1", "repo3"],
    "talent_rank": 88.2,
    "confidence": 93.5,
    "updated_at": "2024-01-21T12:45:00Z"
  }
}
```

### 7. 删除开发者

```http
DELETE /api/developers/{id}
```

删除指定 ID 的开发者信息（需要认证）。

#### 路径参数

| 参数 | 类型   | 必需 | 描述           |
|------|--------|------|----------------|
| `id` | string | 是   | 开发者的唯一 ID |

#### 响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": null
}
```

## 技术栈

- **后端**: Go 1.20+
- **Web框架**: Gin
- **数据库**: MongoDB
- **前端**: React + TypeScript
- **缓存**: Redis
- **消息队列**: RabbitMQ (用于异步任务处理)

## 主要依赖

- [github.com/gin-gonic/gin](https://github.com/gin-gonic/gin)
- [go.mongodb.org/mongo-driver](https://go.mongodb.org/mongo-driver)
- [github.com/go-redis/redis/v8](https://github.com/go-redis/redis)
- [github.com/google/go-github/v45](https://github.com/google/go-github)
- [github.com/golang-jwt/jwt](https://github.com/golang-jwt/jwt)

## 项目结构

```
├── cmd/                    # 主程序入口
│   ├── crawler/            # 爬虫命令行工具
│   │   └── main.go          # 爬虫主程序
│   └── server/             # API服务器
│       └── main.go          # API服务器主程序
├── internal/               # 内部包
│   ├── crawler/            # 数据爬虫
│   │   ├── github.go       # GitHub API调用
│   │   └── parser.go       # 数据解析
│   ├── models/             # 数据模型
│   │   ├── developer.go    # 开发者模型
│   │   ├── talent.go       # TalentRank算法
│   │   ├── nation.go       # 国家预测
│   │   ├── domain.go       # 领域模型
│   │   └── common.go       # 通用模型
│   ├── api/                # API处理
│   │   ├── handlers/       # 请求处理器
│   │   ├── middleware/     # 中间件
│   │   └── routes.go       # 路由定义
│   └── pkg/                # 公共工具
│       ├── database/       # 数据库操作
│       └── utils/          # 工具函数
├── configs/                # 配置文件
├── web/                    # 前端代码
│   ├── src/
│   └── public/
├── scripts/                # 部署脚本
├── go.mod                  # Go模块文件
├── go.sum                  # 依赖版本锁定
└── README.md               # 项目文档
```

## 使用限制

### 1. API 限流

- **未认证用户**: 60 次/小时
- **已认证用户**: 5000 次/小时
- 建议使用 Token 认证并合理控制请求频率

### 2. 数据更新频率

- **活跃用户（1000+ commits）**: 每天更新
- **普通用户**: 每周更新
- 支持强制更新

### 3. 并发控制

- 建议并发数: 3-5
- 批量上限: 100 用户/次
- 自动限流和重试

### 4. 数据缓存

- **Redis 缓存**: 24 小时
- 支持手动刷新
- 缓存预热功能

## 开发计划

- [x] 基础爬虫功能
- [x] TalentRank 算法
- [x] Nation 预测
- [x] 数据缓存
- [ ] API 服务
- [ ] 前端界面
- [ ] 性能优化
- [ ] 数据可视化

## 使用示例

### cURL 示例

#### 获取单个开发者信息

```bash
curl -X GET "http://api.example.com/api/developers/torvalds" \
  -H "Authorization: Bearer your_token"
```

#### 批量获取开发者信息

```bash
curl -X POST "http://api.example.com/api/developers/batch" \
  -H "Authorization: Bearer your_token" \
  -H "Content-Type: application/json" \
  -d '{
    "usernames": ["torvalds", "vbuterin"],
    "concurrency": 2
  }'
```

#### 搜索开发者

```bash
curl -X GET "http://api.example.com/api/search?q=golang&nation=CN&min_stars=1000" \
  -H "Authorization: Bearer your_token"
```

### Python 示例

```python
import requests

API_BASE = "http://api.example.com/api"
TOKEN = "your_token"

# 获取开发者信息
def get_developer(username):
    response = requests.get(
        f"{API_BASE}/developers/{username}",
        headers={"Authorization": f"Bearer {TOKEN}"}
    )
    return response.json()

# 批量获取开发者信息
def get_developers_batch(usernames, concurrency=5):
    response = requests.post(
        f"{API_BASE}/developers/batch",
        headers={"Authorization": f"Bearer {TOKEN}", "Content-Type": "application/json"},
        json={"usernames": usernames, "concurrency": concurrency}
    )
    return response.json()

# 搜索开发者
def search_developers(query, skills=None, nation=None, min_stars=None, min_rank=None, page=1, per_page=20):
    params = {
        "q": query,
        "skills": skills,
        "nation": nation,
        "min_stars": min_stars,
        "min_rank": min_rank,
        "page": page,
        "per_page": per_page,
        "sort": "talent_rank",
        "order": "desc"
    }
    response = requests.get(
        f"{API_BASE}/search",
        headers={"Authorization": f"Bearer {TOKEN}"},
        params=params
    )
    return response.json()

# 使用示例
developer = get_developer("torvalds")
print(developer)

batch_developers = get_developers_batch(["torvalds", "vbuterin"])
print(batch_developers)

search_results = search_developers(q="golang", nation="CN", min_stars=1000)
print(search_results)
```


