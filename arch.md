System Architecture Overview
1. User Initiates Deployment

   A consumer selects an application and submits a deployment request.
   The request includes Helm chart details, deployment type (K8s/VM), and resource requirements (CPU, memory).

2. Task Queuing

   The request is added to a Redis queue for asynchronous processing.
   A worker picks up the task and starts provisioning the environment.

3. Kubernetes Cluster Setup

   A KIND (Kubernetes-in-Docker) cluster is dynamically created with a unique timestamp-based name.
   The system switches kubeconfig context to the new cluster to ensure proper deployment.

4. Helm Chart Deployment

   The worker adds the Helm repository and deploys the application using the given chart.
   If the repo already exists, it skips adding it.
   The system validates deployment success and logs failures if any.

5. Billing Integration

   After a successful deployment, a billing record is created.
   The system calculates usage costs based on CPU, memory, and uptime.

6. Completion & User Access

   Once deployed, the consumer receives access details for the application.

