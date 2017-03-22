# Aggregation Specifications

The aggregation also expects a file called ```aggregation-specifications.yaml``` in its local directory. This
file configures the aggregation operations to execute.

## Name
This will be used to reference the rule and will not appear in any metrics
Example:
```
name: CPU_total_sec
```

## Aggregated Metric Name
This will be the name of the metric(s) output from this rule. It is required (Monasca metrics must have a name)
Example:
```
aggregatedMetricName: kubernetes.cpu.total_sec
```

## Filtered Metric Name
This is the name to search for in the input stream from kafka. It will not appear in any output metrics.
Example:
```
filteredMetricName: pod.cpu.total_sec
```

## Filtered Dimensions
This is a map of key-value pairs to search for in the input stream. These will be carried over to the output metrics.
Example:
```
filteredDimensions:
  hostname: test-host-01
  service: monitoring
```

## Grouped Dimensions
This is a list of dimension keys to be combined into groups. Each new value found for a particular key will create
a new group. If multiple keys are listed, each group will be a specific combination of values. Each group will
be output as a separate metric, and will include the specific groups dimension values.
Example:
```
groupedDimensions:
  - hostname
  - service
```

## Rollup
This section allows aggregated groups defined above to be combined back into a single group via another aggregation
function. For example, to get the total memory used on all systems, you may want to average each system then report
the sum of the averages. Rollup supports a separate function and groups from the rest of the aggregation rule.
Example:
```
rollup:
  function: sum
  groupedDimensions:
    - hostname
    - service
```

# Example
A complete example (including all components) might look like:
```
- name: CPU_total_sec
  aggregatedMetricName: kubernetes.cpu.total_sec
  filteredDimensions:
    cluster: test-cluster-1
  groupedDimensions:
    - hostname
    - service
  rollup:
    function: sum
    groupedDimensions:
      - service
```
which would report the cpu use (totaled across all hosts) of each service.