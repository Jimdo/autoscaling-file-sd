[![Download on GoBuilder](http://badge.luzifer.io/v1/badge?title=Download%20on&text=GoBuilder)](https://gobuilder.me/github.com/Jimdo/autoscaling-file-sd)
[![License: Apache v2.0](https://badge.luzifer.io/v1/badge?color=5d79b5&title=license&text=Apache+v2.0)](http://www.apache.org/licenses/LICENSE-2.0)
[![Go Report Card](https://goreportcard.com/badge/github.com/Jimdo/autoscaling-file-sd)](https://goreportcard.com/report/github.com/Jimdo/autoscaling-file-sd)

# Jimdo / autoscaling-file-sd

This repository contains a small daemon periodically pulling members of an autoscaling-group in AWS and writing a [file-sd configuration for Prometheus](https://prometheus.io/docs/operating/configuration/#file_sd_config).

## Usage

```bash
# autoscaling-file-sd --help
Usage of autoscaling-file-sd:
  -a, --autoscaling-group-name="": Name of the AutoScalingGroup to fetch instances from
  -i, --interval=30s: Interval to poll for changes in the ASG
      --port=0: Port to register in SRV record
  -p, --publish-to="file://discover.json": Where to write the discovery file
      --version[=false]: Print version and exit
```

Currently two write targets for the discovery file are supported:

- `file://` for local files (examples: `file://discovery.json` `file:///home/myuser/discovery.json`)
- `s3://` for Amazon S3 (example: `s3://mybucket/path/discovery.json`)

## Example output file

```bash
# cat discover.json | jq
[
  {
    "targets": [
      "10.8.3.39:9102",
      "10.8.3.85:9102",
      "10.8.4.131:9102"
    ]
  }
]
```
