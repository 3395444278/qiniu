

# 整体架构图


![mmexport1730992177316](https://github.com/user-attachments/assets/c4432388-6d20-48b6-8d6d-b205d34c63a8)







# 演示视频地址
http://dianpiao-test.oss-cn-shenzhen.aliyuncs.com/temp/2023-04-19/mmexport1730988912590(1).mp4

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

## 快速启动
启动 cmd下面三个包的main.go
（注意填入相应redis mongodb deepseek Api   github api的配置)

## 命令行工具

### 1. 环境配置

设置环境变量（可以直接在配置文件配置）：

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

### 搜索开发者
GET /api/search

支持多种搜索条件组合查询，所有参数都是可选的。

#### 查询参数：

| 参数            | 类型 | 说明                                          | 示例                                   |
|---------------|------|---------------------------------------------|--------------------------------------|
| name          | string | 模糊查询名字                                      | `name=zero`              |
| keyword       | string | 关键词搜索(匹配用户名/姓名/邮箱/位置)                       | `keyword=john`                       |
| domain        | string | 按领域搜索(backend/frontend/mobile/ai等)          | `domain=backend`                     |
| nations       | array | 按国家筛选(支持多个)                                 | `nations=CN,JP`                      |
| skills        | array | 按技能筛选(支持多个)                                 | `skills=Go,Python`                   |
| min_activity  | int | 最近活跃天数                                      | `min_activity=30`                    |
| min_commits   | int | 最少提交数                                       | `min_commits=1000`                   |
| min_stars     | int | 最少 star 数                                   | `min_stars=100`                      |
| min_rank      | float | 最低 TalentRank                               | `min_rank=80`                        |
| updated_after | string | 更新时间起点(RFC3339格式)                           | `updated_after=2024-01-01T00:00:00Z` |
| sort_by       | string | 排序字段(talent_rank/star_count/commit_count)   | `sort_by=talent_rank`                |
| sort_asc      | bool | 是否升序(默认降序)                                  | `sort_asc=true`                      |
| page          | int | 页码(默认1)                                     | `page=1`                             |
| page_size     | int | 每页数量(默认10)                                  | `page_size=20`                       |

#### 领域分类：
- **backend**: 后端开发
- **frontend**: 前端开发
- **mobile**: 移动开发
- **ai**: 人工智能
- **devops**: 运维开发
- **database**: 数据库
- **security**: 安全
- **blockchain**: 区块链
- **gamedev**: 游戏开发
- **embedded**: 嵌入式
- **systems**: 系统开发

#### 示例请求：

1. 搜索中国的 Go 开发者：
GET /api/search?nations=CN&skills=Go
2. 搜索后端领域的高影响力开发者：
GET /api/search?domain=backend&min_stars=1000
3. 搜索最近活跃的全栈开发者：
GET /api/search?skills=JavaScript,Python&min_activity=30
4. 按 TalentRank 排序并筛选：
GET /api/search?min_rank=80&sort_by=talent_rank&sort_asc=false
5. 关键词搜索特定地区的开发者：
GET /api/search?keyword=zhang&nations=CN
6. 组合多个查询条件：
GET /api/search?keyword=john&skills=Go,Python&min_stars=1000&sort_by=star_count

## AI 评估功能

系统现在支持自动评估开发者的技术能力：

- 分析 GitHub 个人简介
- 分析博客内容
- 分析仓库贡献
- 生成技术专长评估
- 生成经验水平评估
- 生成整体技术评价

### 启动服务

1. 确保 Redis 已启动：

```bash
redis


### 后端整体架构说明
3. # 后端项目架构设计文档

   ## 1. 整体架构

   ### 1.1 系统架构图

   graph TD
       A --> C[开发者服务]
       A --> D[搜索服务]
       A --> E[AI评估服务]
       

       B --> F[(Redis)]
       C --> G[(MongoDB)]
       H[GitHub API] --> I[爬虫服务]
       I --> G

###     1.2 核心模块

####      internal/crawler

```go
type GithubCrawler struct {
    client    *github.Client
    rateLimit *redis.Client
    storage   *mongo.Collection
}

func (c *GithubCrawler) FetchUserData(username string) (*models.Developer, error) {
    // 实现 GitHub 数据抓取
    // 包含速率限制处理
    // 错误重试机制
}
```



 internal/models

```go
type Developer struct {
    ID              primitive.ObjectID `bson:"_id,omitempty"`
    Username        string            `bson:"username"`
    Name            string            `bson:"name"`
    Location        string            `bson:"location"`
    Nation          string            `bson:"nation"`
    NationConfidence float64          `bson:"nation_confidence"`
    Skills          []string          `bson:"skills"`
    TalentRank      float64          `bson:"talent_rank"`
    Metrics         DeveloperMetrics  `bson:"metrics"`
    UpdatedAt       time.Time         `bson:"updated_at"`
}
```

#### internal/api

```go
func SetupRoutes(r *gin.Engine) {
    // 健康检查
    r.GET("/health", handlers.HealthCheck)
```

​    

```go
// API 路由组
api := r.Group("/api")
{
    // 开发者相关接口
    developers := api.Group("/developers")
    {
        developers.GET("", handlers.ListDevelopers)
        developers.POST("", middleware.Auth(), handlers.CreateDeveloper)
        developers.GET("/:id", handlers.GetDeveloper)
        developers.PUT("/:id", middleware.Auth(), handlers.UpdateDeveloper)
        developers.DELETE("/:id", middleware.Auth(), handlers.DeleteDeveloper)
    }
    
    // 搜索接口
    api.GET("/search", handlers.SearchDevelopers)
}
```

## 2. 数据库设计

### 2.1 MongoDB 集合设计

#### developers 集合

```go
// 创建索引
db.developers.createIndex({ "username": 1 }, { unique: true })
db.developers.createIndex({ "nation": 1 })
db.developers.createIndex({ "skills": 1 })
db.developers.createIndex({ "talent_rank": -1 })
db.developers.createIndex({ 
    "username": "text", 
    "name": "text", 
    "location": "text" 
})
```

### 2.2 Redis 缓存设计

```
const (
    // 缓存键模式
    DeveloperCacheKey = "dev:%s"      // 开发者信息缓存
    SearchCacheKey   = "search:%s"    // 搜索结果缓存
    RateLimitKey    = "ratelimit:%s"  // API 限流
    
    // 缓存过期时间
    DeveloperCacheTTL = 24 * time.Hour
    SearchCacheTTL   = 1 * time.Hour
)
```

## 3. 项目检查清单

### 3.1 代码规范 ✅

遵循 Go 标准项目布局

使用 golangci-lint 进行代码检查

统一的错误处理机制

完善的日志记录

### 4.3 长期改进

服务拆分

引入消息队列

实现分布式追踪

完善测试覆盖率

## 5. 总结

### 存在的问题

安全性措施不足

缓存策略不完善

监控体系缺失

性能优化空间大

### 优先级排序

紧急：安全性问题修复

重要：缓存机制实现

常规：监控系统搭建

长期：架构优化升级

## 4.人员分工

**前端**

李仕骥

**后端**

 徐瑞

