

# GitHub 开发者评估系统

基于 GitHub 的开源项目数据，开发的开发者评估应用。通过分析开发者的项目贡献、技术栈和影响力，提供全面的开发者画像。

## 整体架构图


![mmexport1730992177316](https://github.com/user-attachments/assets/c4432388-6d20-48b6-8d6d-b205d34c63a8)







# 演示视频地址
下载：http://dianpiao-test.oss-cn-shenzhen.aliyuncs.com/temp/2023-04-19/mmexport1730988912590(1).mp4


b站：https://www.bilibili.com/video/BV1NhDDYpEM5?buvid=XXBA79DB2D09A6B89A452BC313873CA633BC5&from_spmid=dt.dt.video.0&is_story_h5=false&mid=blCUROVXIpcn%2Bqa3P2EXdw%3D%3D&plat_id=114&share_from=ugc&share_medium=android&share_plat=android&share_session_id=cba51599-403c-4ccb-8b83-ac0abef6d87a&share_source=WEIXIN&share_tag=s_i&spmid=united.player-video-detail.0.0&timestamp=1730996014&unique_k=LDgSpm9&up_id=173913923&vd_source=7d86eceb26807ff0475aec9e336148c3




# 人员分工

**前端**
李仕骥

#### 职责：

负责与项目团队沟通，理解业务需求，确定系统设计的方向和要求。

负责开发用户界面，确保用户可以直观地查看开发者评估结果。

##### 任务：

1.设计和实现前端页面，包括首页、搜索页、结果展示页等。

2.实现前端的交互逻辑，如模糊搜索、领域分类、地区分类等。

3.确保前端界面的响应式设计，适配不同设备和屏幕尺寸。

4.与后端服务进行接口对接，确保数据的准确传输和展示。

5.参与系统架构设计，确定系统的模块划分和通信方式。

6.参与设计数据库模型和API接口，确保数据的持久化和一致性。



**后端**

徐瑞

#### 职责：

负责后端业务逻辑的实现，包括API的开发和内部服务的逻辑处理

负责设计数据库表结构，确保数据的组织方式能够高效支持业务需求。

##### 任务：
1. 设计表结构，包括字段定义、数据类型、索引等。
2. 确保表结构的规范化，减少数据冗余，提高数据一致性。
3. 参与系统架构设计，确定系统的模块划分和通信方式。 
4. 参与爬虫，数据处理，接口模块的开发






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

### http爬取用户
# GitHub 用户爬取 API 文档

## 1. 运行爬虫接口

### 接口描述
启动爬虫任务，爬取指定 GitHub 用户的信息。

### 请求信息
- 请求路径：`/api/run-crawler`
- 请求方法：POST
- Content-Type：application/json

### 请求参数
| 参数名 | 类型 | 必填 | 说明 | 示例值 |
|--------|------|------|------|---------|
| usernames | string | 是 | GitHub 用户名，多个用户用逗号分隔 | "torvalds,antirez" |
| concurrency | int | 是 | 并发爬取数量，取值范围 1-6 | 3 |

### 请求示例
json
{
"usernames": "torvalds,antirez",
"concurrency": 3
}

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





## 技术栈

- **后端**: Go 1.20+
- **Web框架**: Gin
- **数据库**: MongoDB
- **前端**: Vue3 + JavaScript
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


### 后端模块说明

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

### 3.2 长期改进

服务拆分

引入消息队列

实现分布式追踪

完善测试覆盖率

## 4. 总结

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

1.GitHub开发者排名评级系统的工作流程
这个系统主要分为五个核心步骤：

第一步：开发者信息采集
我们会收集开发者的GitHub个人信息，包括代码仓库、提交记录、Star数量、Fork数量等基础数据。这些原始数据是我们评估的基础。

第二步：数据清洗
对采集到的数据进行预处理和清洗，去除无效数据，统一数据格式，确保数据的准确性和可用性。比如剔除fork的项目，只保留原创内容。

第三步：提取特征指标
我们会从清洗后的数据中提取关键特征指标，主要包括：

- 贡献度指标：代码提交数量和质量
- 项目质量指标：Star数和Fork数,项目大小
- 影响力指标：关注者数量和高质量项目加成
- 活跃度指标：更新频率和持续性和增长趋势
- 专业度指标：技术栈的广度和深度

第四步：权重计算
对各个指标进行加权计算：

- 贡献度权重30%
- 项目质量权重25%
- 影响力权重20%
- 活跃度权重15%
- 专业度权重10%

第五步：返回结果

最终将计算结果归一化到0-100分，并根据分数划分等级，给出开发者的最终评级。

这套系统的优势在于：

1. 多维度评估，全面客观

2. 权重可调整，灵活可控

3. 数据实时更新，动态反映开发者能力
![image-20241123091922990](https://github.com/user-attachments/assets/9e5ab460-0230-490c-990c-3a57ee4f0441)


2.GitHub开发者地理位置预测系统的工作流程

   这个系统主要分为五个核心步骤：

第一步：获取开发者信息
我们通过GitHub API获取开发者的基础信息，包括：

- 个人资料信息
- 仓库信息
- 代码提交记录
- 活动时间线等数据

第二步：国家地区是否确定
系统首先判断用户是否已经明确标注了地理位置信息。如果有明确的地理标识，系统会直接使用该信息；如果没有，则进入下一步分析。

第三步：分析用户数据
系统会从三个维度收集和分析数据：

1. 代码提交相关：
   - commits的时间分布
   - pull requests的提交时间
   - 代码注释语言  ，reandme信息等

2. 个人信息相关：
   - 用户名特征
   - 邮箱域名后缀，公司信息
   - 个人主页信息

3. 开发者社交网络：
   - 关注的开发者
   - 参与的组织
   - 互动的项目区域分布

第四步：AI模型预测

使用deepseek的AI模型，综合分析：

- 用户名和显示名称中的地理特征
- 代码仓库描述的语言特征
- 活动时间规律
- 技术社区互动特征

第五步：结果完成
输出预测结果，包括：

- 预测的国家/地区
- 预测的置信度
- 支持该预测的关键特征

这套系统的优势在于：

1. 多维度数据分析，提高预测准确性
2. AI模型持续学习，预测能力不断提升
3. 实时动态更新，保持数据时效性
![image-20241123092918263](https://github.com/user-attachments/assets/9809412c-ad8c-4502-84f7-5fc8949b7e9b)

GitHub开发者置信度预测系统
这个系统主要分为五个核心步骤：

第一步：开发者信息采集
我们会收集开发者的基础信息，包括：

- 个人资料信息
- 代码仓库数据
- 活动记录
- 社交网络数据

第二步：数据清洗
对采集到的原始数据进行预处理：

- 去除无效或异常数据
- 统一数据格式
- 补充缺失字段
- 标准化处理

第三步：提取特征指标
从清洗后的数据中提取关键特征：

账号基础置信度
贡献调整置信度

根据star数

根据关注者
位置信息
![image-20241123093026452](https://github.com/user-attachments/assets/7ba7317d-2677-4a37-b004-91dfbe4e4fc3)

GitHub开发者评估系统的工作流程。

这个系统主要分为六个核心步骤：

第一步：开发者信息采集
我们通过GitHub API获取开发者的全面信息，包括：

- 个人基本信息
- 代码仓库数据
- 贡献记录
- 技术栈信息

第二步：数据清洗
对采集到的原始数据进行预处理：

- 去除无效数据
- 统一数据格式
- 补充缺失字段
- 数据标准化

第三步：提取特征指标
从清洗后的数据中提取关键特征：

第四步：生成式模型调用
利用AI模型对开发者进行多维度分析：

- 技术能力评估
- 项目质量评估
- 贡献度分析
- 活跃度计算

第五步：模型引导
根据评估结果进行分类和排名：

- 技术等级划分
- 专业领域识别
- 发展潜力预测
- 综合能力评分

第六步：返回结果
生成最终的评估报告，包括：

- 技术能力评分
- 专业领域定位
- 发展建议
- 详细数据支撑

这套系统的优势在于：

1. 多维度分析，评估全面
2. AI模型支持，智能化程度高
3. 实时数据更新，保证时效性
4. 可扩展性强，支持定制化评估

