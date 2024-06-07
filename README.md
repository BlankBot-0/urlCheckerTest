## Overview
The service concurrently checks and logs status and response time of URLs
## How to use
Provide target URLs in `urls` field of `config.yaml` file.
```sh
$ docker build -t url-checker .
$ docker run --rm url-checker
```