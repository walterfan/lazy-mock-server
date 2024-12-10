# Lazy Mock Server

## Usage
* set mock response

```bash
vi app/mock_response.yaml
```

* startup

```bash
./startup.sh
```

* test

```bash

$ curl http://127.0.0.1:9000/api/get_example
{
  "message": "This is a mock GET response."
}

$ curl -d '{"key1":"value1", "key2":"value2"}' -H "Content-Type: application/json" -X POST http://127.0.0.1:9000/api/post_example
{
  "data_received": "{'key1': 'value1', 'key2': 'value2'}",
  "message": "This is a mock POST response."
}


```

## TODO
* Setup Expectations by REST API
* Add a web UI for setup expectation and checking mock history
* Add sqlite DB support for authentication and history
