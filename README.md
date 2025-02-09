# prometheus-metrics-server

A simple golang server that emits custom metrics to remote hosted Prometheus with locally hosted grafana dashboard. 

# Setup

```bash
docker compose up --build
```

This project is using a remote Prometheus instance hosted on Alibaba Cloud, so in order to set up Grafana data source, follow this [guide](https://www.alibabacloud.com/help/en/prometheus/user-guide/data-query-and-grafana-data-source-settings?spm=a2c63.p38356.help-menu-122122.d_2_3_0.4b2e781bqCGoBu).

# Next Steps

1. hosted Grafana service (if needed and budget allows)
2. use some local memory storage to aggregate the metric points, then send to grafana to save storage space to further reduce cost
3. Customed alerts
