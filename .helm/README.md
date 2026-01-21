# Helm Chart для DonWeather MS Weather

Минимальный Helm chart для развертывания DonWeather microservice в Kubernetes через ArgoCD.

## Структура

```
helm/donweather-ms-weather/
├── Chart.yaml                    # Метаданные chart
├── values.yaml                   # Значения по умолчанию
├── values-production.yaml        # Пример для production
├── argocd-application.yaml      # Пример Application для ArgoCD
├── .helmignore                  # Игнорируемые файлы
├── README.md                    # Документация chart
└── templates/
    ├── _helpers.tpl             # Вспомогательные шаблоны
    ├── deployment.yaml          # Deployment
    ├── service.yaml             # Service
    ├── configmap.yaml           # ConfigMap для env переменных
    ├── secret.yaml              # Secret для секретов
    ├── serviceaccount.yaml      # ServiceAccount
    ├── ingress.yaml             # Ingress (опционально)
    ├── hpa.yaml                 # HorizontalPodAutoscaler (опционально)
    └── servicemonitor.yaml      # ServiceMonitor для Prometheus (опционально)
```

## Быстрый старт

### 1. Локальная установка через Helm

```bash
# Установка с базовыми значениями
helm install donweather-ms-weather ./helm/donweather-ms-weather \
  --set weatherApiKey=your-api-key \
  --set database.host=postgres-service \
  --set database.password=your-password \
  --set image.repository=your-registry/donweather-ms-weather \
  --set image.tag=latest

# Установка с production values
helm install donweather-ms-weather ./helm/donweather-ms-weather \
  -f helm/donweather-ms-weather/values-production.yaml \
  --set weatherApiKey=your-api-key \
  --set database.password=your-password
```

### 2. Развертывание через ArgoCD

#### Вариант 1: Через UI ArgoCD

1. Откройте ArgoCD UI
2. Создайте новое Application:
   - **Application Name**: `donweather-ms-weather`
   - **Project Name**: `default`
   - **Sync Policy**: `Automatic`
   - **Repository URL**: URL вашего Git репозитория
   - **Revision**: `main` (или ваша ветка)
   - **Path**: `helm/donweather-ms-weather`
   - **Cluster URL**: `https://kubernetes.default.svc`
   - **Namespace**: `default` (или ваш namespace)

3. В разделе **Helm** добавьте параметры:
   ```
   weatherApiKey=your-api-key
   database.host=postgres-service
   database.password=your-password
   image.repository=your-registry/donweather-ms-weather
   image.tag=latest
   ```

#### Вариант 2: Через манифест Application

Используйте файл `argocd-application.yaml` как пример:

```bash
# Отредактируйте argocd-application.yaml с вашими значениями
kubectl apply -f helm/donweather-ms-weather/argocd-application.yaml
```

**Важно:** Обновите в файле:
- `repoURL`: URL вашего Git репозитория
- `targetRevision`: ваша ветка (main, master, etc.)
- Параметры в `helm.parameters`

### 3. Использование External Secrets (рекомендуется для production)

Вместо хранения секретов в values.yaml, используйте External Secrets Operator:

1. Создайте Secret в Kubernetes:
```bash
kubectl create secret generic donweather-secrets \
  --from-literal=weather-api-key=your-api-key \
  --from-literal=db-password=your-password \
  -n default
```

2. Обновите `values.yaml` или создайте ExternalSecret манифест

3. В ArgoCD используйте параметры для ссылки на секреты

## Основные параметры

| Параметр | Описание | По умолчанию |
|----------|----------|--------------|
| `replicaCount` | Количество реплик | `2` |
| `image.repository` | Docker образ | `donweather-ms-weather` |
| `image.tag` | Тег образа | `latest` |
| `service.type` | Тип Service | `ClusterIP` |
| `database.host` | Хост PostgreSQL | `postgres` |
| `database.port` | Порт PostgreSQL | `5432` |
| `database.name` | Имя БД | `weather` |
| `database.user` | Пользователь БД | `myuser` |
| `database.password` | Пароль БД | `mypassword` |
| `weatherApiKey` | API ключ Weather API | `""` (обязательно) |
| `env.corsOrigin` | CORS origin | `"*"` |
| `env.env` | Окружение | `production` |

## Health Checks

- **Liveness**: `/health/live` (проверка каждые 10 секунд)
- **Readiness**: `/health/ready` (проверка каждые 5 секунд)
- **Metrics**: `/metrics` (Prometheus метрики)

## Мониторинг

Для включения ServiceMonitor (требует Prometheus Operator):

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

## Ingress

Для включения Ingress:

```yaml
ingress:
  enabled: true
  className: "nginx"  # или ваш ingress controller
  hosts:
    - host: donweather.example.com
      paths:
        - path: /
          pathType: Prefix
```

## Проверка установки

```bash
# Проверить статус deployment
kubectl get deployment donweather-ms-weather

# Проверить pods
kubectl get pods -l app.kubernetes.io/name=donweather-ms-weather

# Проверить логи
kubectl logs -l app.kubernetes.io/name=donweather-ms-weather

# Проверить service
kubectl get svc donweather-ms-weather

# Проверить health check
kubectl port-forward svc/donweather-ms-weather 8080:8080
curl http://localhost:8080/health/live
curl http://localhost:8080/health/ready
```

## Обновление

### Через Helm
```bash
helm upgrade donweather-ms-weather ./helm/donweather-ms-weather \
  --set image.tag=new-tag
```

### Через ArgoCD
ArgoCD автоматически синхронизирует изменения из Git репозитория (если включен automated sync).

## Удаление

```bash
# Через Helm
helm uninstall donweather-ms-weather

# Через ArgoCD
kubectl delete application donweather-ms-weather -n argocd
```

## Troubleshooting

### Pod не запускается
1. Проверьте логи: `kubectl logs <pod-name>`
2. Проверьте секреты: `kubectl get secret donweather-ms-weather-secret`
3. Проверьте ConfigMap: `kubectl get configmap donweather-ms-weather-config`

### Health check fails
1. Проверьте, что приложение слушает на порту 8080
2. Проверьте, что endpoints `/health/live` и `/health/ready` доступны
3. Увеличьте `initialDelaySeconds` в values.yaml

### База данных недоступна
1. Проверьте, что PostgreSQL сервис доступен
2. Проверьте DSN в секрете: `kubectl get secret donweather-ms-weather-secret -o jsonpath='{.data.dsn}' | base64 -d`
3. Проверьте сетевые политики (NetworkPolicies)

## Дополнительные ресурсы

- [Helm Documentation](https://helm.sh/docs/)
- [ArgoCD Documentation](https://argo-cd.readthedocs.io/)
- [Kubernetes Documentation](https://kubernetes.io/docs/)
