# kubescapecontroller
The primary goal of the project is to develop an admission controller for the Kubescape
application. The admission controller will scan individual workloads (e.g. YAML files) before they
are submitted to an API server in the cluster and it should operate inside a cluster. The task also
includes that the controller should be well documented , safe to install, and instrumented with
logging and telemetry data to diagnose problems.
