# GitOps with ArgoCD

本项目演示如何使用 ArgoCD 实现 GitOps 工作流。

## 架构

```
┌─────────────────┐     ┌─────────┐     ┌──────────────┐
│   GitHub Repo   │────▶│ ArgoCD  │────▶│ Kubernetes   │
│  (Single Source)│     │ (Sync)  │     │  (kind)      │
└─────────────────┘     └─────────┘     └──────────────┘
                              │
                              ▼
                       ┌──────────────┐
                       │ Auto Sync +  │
                       │ Self-Heal    │
                       └──────────────┘
```

## 前置要求

- [kind](https://kind.sigs.k8s.io/) - 本地 Kubernetes 集群
- [kubectl](https://kubernetes.io/docs/tasks/tools/) - K8s 命令行工具
- [ArgoCD CLI](https://argo-cd.readthedocs.io/en/stable/cli_installation/) (可选)

## 快速开始

### 1. 创建 kind 集群

```bash
kind create cluster --config gitops/argocd/cluster.yaml
```

### 2. 安装 ArgoCD

```bash
kubectl create namespace argocd
kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml
```

等待 ArgoCD 就绪：

```bash
kubectl wait --for=condition=available --timeout=300s \
  deployment/argocd-server -n argocd
```

### 3. 访问 ArgoCD UI

```bash
kubectl port-forward svc/argocd-server -n argocd 8081:443
```

访问：https://localhost:8081

获取初始密码：

```bash
kubectl -n argocd get secret argocd-initial-admin-secret \
  -o jsonpath="{.data.password}" | base64 -d
```

登录：
- 用户名：`admin`
- 密码：上面的输出

### 4. 部署应用

```bash
kubectl apply -f gitops/apps/devops-practice.yaml
```

### 5. 验证部署

```bash
# 查看 Application 状态
argocd app get devops-practice

# 查看 Pod
kubectl get pods -n devops-practice

# 端口转发访问应用
kubectl port-forward svc/devops-practice 8080:8080 -n devops-practice

# 测试端点
curl http://localhost:8080/healthz
```

## GitOps 自动同步演示

### 修改副本数

编辑 `helm/devops-practice/values.yaml`:

```yaml
replicaCount: 3  # 从 1 改为 3
```

提交并推送：

```bash
git add helm/devops-practice/values.yaml
git commit -m "feat: scale to 3 replicas"
git push
```

**观察 ArgoCD：**
1. UI 中会自动检测到变化
2. 几秒内自动同步到集群
3. Pod 数量从 1 变为 3

### 手动触发同步（如果自动同步未开启）

```bash
argocd app sync devops-practice
```

## ArgoCD Application 配置说明

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: devops-practice
  namespace: argocd
spec:
  project: default
  source:
    repoURL: https://github.com/aLittlecrocodile/coding-agent.git
    targetRevision: feature/ai-agent-framework  # 分支/tag
    path: helm/devops-practice                  # Helm Chart 路径
    helm:
      valueFiles:
        - values.yaml
  destination:
    server: https://kubernetes.default.svc
    namespace: devops-practice
  syncPolicy:
    automated:
      prune: true      # 自动删除 Git 中不存在的资源
      selfHeal: true   # 自动修复配置漂移
    syncOptions:
      - CreateNamespace=true  # 自动创建 namespace
```

## 常用命令

```bash
# 列出所有应用
argocd app list

# 查看应用详情
argocd app get devops-practice

# 手动同步
argocd app sync devops-practice

# 查看应用资源
argocd app resources devops-practice

# 查看应用事件
argocd app events devops-practice

# 删除应用
argocd app delete devops-practice
```

## 清理

```bash
# 删除 Application
kubectl delete -f gitops/apps/devops-practice.yaml

# 删除 ArgoCD
kubectl delete -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml

# 删除集群
kind delete cluster --name gitops-demo
```

## 学习要点

| 概念 | 说明 |
|------|------|
| **Application** | ArgoCD 中定义应用部署的 CRD |
| **Sync Policy** | 控制自动同步行为 |
| **Self-Heal** | 自动修复配置漂移（有人手动修改了 K8s 资源） |
| **Prune** | 自动删除 Git 中不存在的资源 |
| **Sync Waves** | 控制资源同步顺序 |

## 故障排查

### 应用无法同步

```bash
# 查看应用状态
argocd app get devops-practice --refresh

# 查看同步历史
argocd app sync devops-practice --dry-run
```

### Pod 无法启动

```bash
# 查看 Pod 日志
kubectl logs -n devops-practice deployment/devops-practice

# 查看 Pod 事件
kubectl describe pod -n devops-practice
```

### ArgoCD 无法访问 Git 仓库

检查 Application 中的 `repoURL` 和 `targetRevision` 是否正确。

## 下一步

- [ ] 配置 ArgoCD Notifications（Slack/Email 通知）
- [ ] 使用 App-of-Apps 模式管理多个应用
- [ ] 配置 Progressive Rollouts（蓝绿部署、金丝雀）
- [ ] 集成 Secret 管理工具（如 SOPS、External Secrets Operator）
