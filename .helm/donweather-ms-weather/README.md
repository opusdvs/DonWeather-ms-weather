# DonWeather MS Weather Helm Chart

Минимальный Helm chart для развертывания DonWeather microservice в Kubernetes через ArgoCD.

## Установка

### Локальная установка

```bash
helm install donweather-ms-weather ./helm/donweather-ms-weather \
  --set weatherApiKey=your-api-key \
  --set database.host=postgres-service \
  --set database.password=your-password
```

### Установка через ArgoCD

Создайте Application в ArgoCD:

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: donweather-ms-weather
  namespace: argocd
spec:
  project: default
  source:
    repoURL: https://github.com/your-org/DonWeather-ms-weather
    targetRevision: main
    path: helm/donweather-ms-weather
    helm:
      valueFiles:
        - values.yaml
      parameters:
        - name: weatherApiKey
          value: your-api-key
        - name: database.host
          value: postgres-service
        - name: database.password
          value: your-password
  destination:
    server: https://kubernetes.default.svc
    namespace: default
  syncPolicy:
    automated:
      prune: true
      selfHeal: true
    syncOptions:
      - CreateNamespace=true
```

## Конфигурация

Основные параметры в `values.yaml`:

- `replicaCount`: Количество реплик (по умолчанию: 2)
- `image.repository`: Репозиторий Docker образа
- `image.tag`: Тег образа
- `service.type`: Тип Service (ClusterIP, NodePort, LoadBalancer)
- `database.*`: Параметры подключения к PostgreSQL
- `weatherApiKey`: API ключ для Weather API
- `env.corsOrigin`: CORS origin (по умолчанию: "*")
- `env.env`: Окружение (production/development)

## Секреты

Секреты хранятся в Kubernetes Secret и создаются автоматически из значений в `values.yaml`.

**Важно:** В production используйте внешние секреты (например, Sealed Secrets, External Secrets Operator) вместо хранения паролей в values.yaml.

## Health Checks

- Liveness probe: `/health/live`
- Readiness probe: `/health/ready`
- Metrics: `/metrics`

## Мониторинг

Для включения ServiceMonitor для Prometheus:

```yaml
serviceMonitor:
  enabled: true
  interval: 30s
```

## Автомасштабирование

Для включения HPA:

```yaml
autoscaling:
  enabled: true
  minReplicas: 2
  maxReplicas: 10
  targetCPUUtilizationPercentage: 80
```
