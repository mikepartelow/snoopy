# snoopy

`snoopy` is a configurable multi-listener snooping HTTP proxy. You send requests to `snoopy`, `snoopy` logs the URLs you're fetching, optionally modifies cookies and headers, then sends your request to the configured upstream. `snoopy` will also run simple text search/replace on the response, if configured, and `snoopy` can listen on multiple ports, proxying each one to a different upstream.

## Build and Run

```bash
% make
% LOG_LEVEL=debug ./snoopy
```

## Example

Imagine that we want to log all the HTTP GET requests involved in an interaction with [PyPI](https://pypi.org/). PyPI is hosted at `pypi.org`, but any packages are downloaded from a different address, `files.pythonhosted.org`. So, to log all HTTP GET requests, we need to proxy both hosts, `pypi.org` and `files.pythonhosted.org`.

Consider this Snoopy `config.yaml`:

```yaml
- upstream: https://pypi.org
  local: 127.0.0.1:9999
  logfile: pypi-index.urls
  headers:
    - name: User-Agent
      value: snoopy/v1
  response-rewrites:
    - old: https://files.pythonhosted.org
      new: http://localhost:9998
      must-rewrite: true
- upstream: https://files.pythonhosted.org
  local: 127.0.0.1:9998
  logfile: pypi-files.urls
```

When run with this configuration, `snoopy` will:
- listen for requests on `127.0.0.1:9999`, proxied to `https://pypi.org` 
- listen for requests on `127.0.0.1:9998`, proxied to `https://files.pythonhosted.org`
- append upstream URLs to the configured per-upstream `logfile`
- modify the response body from `https://pypi.org`, substituting `http://localhost:9998` wherever `https://files.pythonhosted.org` was originally returned
- panic if no substitutions were made in the response body from `https://pypi.org`, because `must-rewrite` is `true`

```bash
% make && LOG_LEVEL=debug ./snoopy
% python3 -m venv venv.foo
% venv.foo/bin/pip install --no-cache-dir --index http://localhost:9999/simple numpy
% cat pypi-index.urls
https://pypi.org/simple/numpy/
https://pypi.org/simple/pip/
% cat pypi-files.urls
https://files.pythonhosted.org/packages/55/78/f85aab3bda3ddffe6ce8c590190b5f0d2e61dfd2fb7a8f446dcb4f8c12c7/numpy-1.26.3-cp311-cp311-macosx_11_0_arm64.whl.metadata
https://files.pythonhosted.org/packages/55/78/f85aab3bda3ddffe6ce8c590190b5f0d2e61dfd2fb7a8f446dcb4f8c12c7/numpy-1.26.3-cp311-cp311-macosx_11_0_arm64.whl
```
