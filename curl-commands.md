```
curl -X POST http://localhost:3000/api/users \
-H "Content-Type: application/json" \
-d '{"name": "John Doe"}'
```

```
curl -X POST http://localhost:3000/api/users \
-H "Content-Type: application/json" \
-d '{"name": "Rakib"}'
```

```
curl -X POST http://localhost:3000/api/user/project \
-H "Content-Type: application/json" \
-d '{"name": "p1", "user_id": 2}'
```

```
curl -X GET http://localhost:3000/api/users \
-H "Content-Type: application/json" 
```

```
curl -X POST http://localhost:3000/api/apps \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Kubernetes App",
    "description": "This is a Kubernetes-based application",
    "publisher_id": 1,
    "hourly_rate": 1.1,
    "deployment" :{
      "type": "k8s",
      "repoURL": "https://charts.bitnami.com/bitnami",
      "chartName": "nginx",
      "image": "",
      "cpu": "",
      "memory": ""
    }
  }'
```

```
curl -X POST http://localhost:3000/api/deployments \
  -H "Content-Type: application/json" \
  -d '{
    "consumer_id": 2,
    "application_id": 1,
    "project_id": 1
  }'
```

```
curl -X DELETE http://localhost:3000/api/apps/1 \
  -H "Content-Type: application/json"
```

```
curl -X GET http://localhost:3000/api/apps/1 \
  -H "Content-Type: application/json"
```

```
curl -X GET http://localhost:3000/api/apps \
  -H "Content-Type: application/json"
```

```
curl -X DELETE http://localhost:3000/api/deployments/1 \
  -H "Content-Type: application/json"
```