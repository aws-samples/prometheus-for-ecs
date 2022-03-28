## Prometheus metrics collection from Amazon ECS

This Git repository contains software artifacts related to two different approaches to collect Prometheus metrics from applications deployed to an Amazon ECS cluster. Both approached use dynamic service discovery in conjunction with AWS Cloud Map.

The first approach employs a single instance of [Prometheus server deployed to an ECS cluster](https://github.com/aws-samples/prometheus-for-ecs/blob/main/deploy-prometheus/README.md).

The second approach employs a single instance of AWS Distro for OpenTelemetry (ADOT) Collector deployed to an ECS cluster. 


## Security

See [CONTRIBUTING](CONTRIBUTING.md#security-issue-notifications) for more information.

## License

This library is licensed under the MIT-0 License. See the LICENSE file.

