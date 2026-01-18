# DevOps Practice - HTTP Service with Observability

一个最小的 Go HTTP 服务，用于演示现代 DevOps 实践：CI/CD、容器化、可观测性和云原生部署。

## 功能

- **HTTP 服务**：提供健康检查和就绪探针端点
- **Prometheus 指标**：内置 `/metrics` 端点，暴露请求计数和进行中请求数
- **容器化**：多阶段 Docker 构建，镜像体积小
- **CI/CD**：GitHub Actions 自动测试、构建镜像、安全扫描
- **Kubernetes**：Helm Chart 支持，开箱即用

## 端点

| 端点 | 方法 | 描述 |
|------|------|------|
| `/healthz` | GET | 健康检查，返回 "ok" |
| `/readyz` | GET | 就绪探针，返回 "ready" |
| `/metrics` | GET | Prometheus 格式指标 |

## 本地运行

### 前置要求

- Go 1.21+

### 运行

```bash
# 克隆仓库
git clone https://github.com/aLittlecrocodile/devops-practice.git
cd devops-practice

# 运行
go run ./cmd/app
```

服务默认运行在 `http://localhost:8080`

### 验证

```bash
# 健康检查
curl http://localhost:8080/healthz
# 输出: ok

# 就绪检查
curl http://localhost:8080/readyz
# 输出: ready

# Prometheus 指标
curl http://localhost:8080/metrics
# 输出: Prometheus 文本格式指标
```

### 测试

```bash
go test ./...
go vet ./...
```

## Docker

### 构建

```bash
docker build -t devops-practice:latest .
```

### 运行

```bash
docker run -p 8080:8080 devops-practice:latest
```

## Docker Compose

```bash
# 启动
docker compose up -d

# 查看日志
docker compose logs -f

# 停止
docker compose down
```

## Kubernetes (Helm)

### 前置要求

- kubectl
- Helm 3+
- Kubernetes 集群（kind / minikube / GKE / EKS 等）

### 安装

```bash
# 使用本地集群（kind）
kind create cluster --name devops-practice

# 安装 Helm Chart
helm install devops-practice ./helm/devops-practice

# 端口转发
kubectl port-forward svc/devops-practice 8080:8080

# 验证
curl http://localhost:8080/healthz
```

### 升级

```bash
# 更新镜像 tag
helm upgrade devops-practice ./helm/devops-practice --set image.tag=sha-abc123
```

### 卸载

```bash
helm uninstall devops-practice
kind delete cluster --name devops-practice
```

## GitHub Actions

### CI Pipeline (`.github/workflows/ci.yml`)

- 触发：所有 push 和 PR
- 步骤：
  - Go fmt 检查
  - go vet 静态分析
  - go test -race 测试

### Docker Pipeline (`.github/workflows/docker.yml`)

- 触发：push 到 main 分支
- 步骤：
  - 多架构构建 (amd64/arm64)
  - 推送到 GHCR (`ghcr.io/aLittlecrocodile/devops-practice`)

### Security Pipeline (`.github/workflows/security.yml`)

- 触发：PR、push、每周定时
- 步骤：
  - Trivy 镜像漏洞扫描
  - 结果上传到 GitHub Security

## 常见问题

### 端口被占用

```bash
# 更换端口
PORT=9090 go run ./cmd/app
```

### GHCR 镜像拉取失败

确保已登录 GitHub Container Registry：
```bash
echo $GITHUB_TOKEN | docker login ghcr.io -u USERNAME --password-stdin
```

### kind 集群创建失败

确保 Docker daemon 正在运行：
```bash
docker info
```

## 项目结构

```
.
├── cmd/app/                    # 应用入口
│   ├── main.go
│   └── main_test.go
├── internal/
│   ├── metrics/                # Prometheus 指标
│   └── server/                 # HTTP 服务器
├── helm/devops-practice/        # Helm Chart
├── .github/workflows/          # GitHub Actions
├── Dockerfile
├── docker-compose.yml
└── go.mod
```

## License

MIT
