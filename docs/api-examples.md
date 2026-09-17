# API curl 示例

假设服务在 `http://localhost:8080`。

## 1. Health

```bash
curl -s http://localhost:8080/health | jq
# {"code":0,"message":"ok","data":{"status":"up"}}
```

## 2. Ready

```bash
curl -s http://localhost:8080/ready | jq
```

## 3. Login

```bash
curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"admin","password":"Admin@123"}' | jq

TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"admin","password":"Admin@123"}' | jq -r '.data.token')
```

## 4. Create site

```bash
curl -s -X POST http://localhost:8080/api/v1/sites \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{
    "name": "政务门户",
    "domain": "portal.gov.example",
    "upstream_host": "10.0.0.10",
    "upstream_port": 80,
    "protocol": "http",
    "protection_mode": "observe"
  }' | jq
```

## 5. List sites

```bash
curl -s 'http://localhost:8080/api/v1/sites?page=1&page_size=20' \
  -H "Authorization: Bearer $TOKEN" | jq
```

## 6. Create policy / node / rule

```bash
curl -s -X POST http://localhost:8080/api/v1/policies \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{"name":"默认策略","mode":"protect","protection_level":"balanced","paranoia_level":1}' | jq

curl -s -X POST http://localhost:8080/api/v1/nodes \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{"name":"edge-1","ip":"10.0.1.5","region":"cn-east","status":"unknown"}' | jq

curl -s -X POST http://localhost:8080/api/v1/rules \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{"rule_id":"942100","name":"SQL Injection Attack Detected","category":"sqli","source":"crs","severity":"critical","scope":"global"}' | jq
```

## 7. Audit logs

```bash
curl -s 'http://localhost:8080/api/v1/audit-logs?page=1' \
  -H "Authorization: Bearer $TOKEN" | jq
```
