GitHub 数据应用
根据 GitHub 的开源项目数据，开发一款开发者评估应用。

## 功能特性

### 基础功能

1. **TalentRank**: 对开发者的技术能力进行评价/评级。评价依据包括：
   - 项目的重要程度
   - 开发者在项目中的贡献度

2. **Nation 预测**: 当开发者 Profile 未明确说明所属国家/地区时,通过关系网络猜测其 Nation。

3. **开发者搜索**: 
   - 根据领域搜索匹配
   - 按 TalentRank 排序
   - 支持 Nation 筛选

### 高级功能 (可选)

1. **置信度**: 为所有猜测的数据提供置信度。低置信度数据在界面显示为 N/A。

2. **自动整理评估信息**: 利用类 ChatGPT 的应用,整理开发者的 GitHub 博客、网站和账号介绍等信息。

## 技术栈

- 后端: Go 1.20+
- Web框架: Gin
- 数据库: MongoDB
- 前端: React + TypeScript
- 缓存: Redis
- 消息队列: RabbitMQ (用于异步任务处理)

## 主要依赖

- github.com/gin-gonic/gin
- go.mongodb.org/mongo-driver
- github.com/go-redis/redis/v8
- github.com/google/go-github/v45
- github.com/golang-jwt/jwt

## 项目结构
├── cmd/                    # 主程序入口
│   └── server/            # API服务器
├── internal/              # 内部包
│   ├── crawler/           # 数据爬虫
│   │   ├── github.go      # GitHub API调用
│   │   └── parser.go      # 数据解析
│   ├── models/            # 模型定义
│   │   ├── developer.go   # 开发者模型
│   │   ├── talent.go      # TalentRank算法
│   │   └── nation.go      # 国家预测
│   ├── api/               # API处理
│   │   ├── handlers/      # 请求处理器
│   │   ├── middleware/    # 中间件
│   │   └── routes.go      # 路由定义
│   └── pkg/               # 公共包
│       ├── database/      # 数据库操作
│       └── utils/         # 工具函数
├── configs/               # 配置文件
├── web/                   # 前端代码
│   ├── src/
│   └── public/
├── scripts/               # 部署脚本
├── go.mod                 # Go模块文件
├── go.sum                 # 依赖版本锁定
└── README.md             # 项目文档
