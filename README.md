# domena-check

### A tiny http service for domain ownership verification (<200 lines)

**Let's check out the usecase:** User provides the domain name which he wants to connect with "our" app. Before we allow the connection to go through, we first need to make sure that he really owns that domain. To verify the ownership, we instruct the user to configure the DNS with appropriate TXT and CNAME record. After that, we check if the records match. This service helps with that process.

I've made this mainly for myself thus the project will most likely change in the future to fit my needs. I'm sharing it anyway in case somebody finds it useful

### How it works?

The idea is simple. Spawn a background worker with an http call that periodically checks if the DNS records (CNAME and TXT) match the expected ones. If they match, the domain is considered verified. The worker than notifies another arbitrary service via webhook that the given domain has been successfully verified.

API exposes one **POST** route:`/api/verify`

```json
// Example of the received POST body
{
	"domain": "www.example.com",
	"expected_TXT": "wgyf8z8cgvm2qmxpnbnldrcltvk4xqfn",
	"expected_CNAME": "www.example.com."
}
```

After both DNS records are verified the worker will send a `POST` request to the specified webhook url.

```json
// Example of the sent POST body
{
	"domain": "www.example.com",
	"acquired_txt": "wgyf8z8cgvm2qmxpnbnldrcltvk4xqfn",
	"acquired_cname": "www.example.com."
}
```

### How to start a service

This is a dockerized service so in order to run the service you need to have Docker and docker-compose installed.

1. Create .env file populated with the required env variables. See the example below:

```sh
# address of the service
ADDR=:8080
# webhook url that will be called after a successful verification
WEBHOOK_ROUTE=localhost:9999/api/hooker
# how often should a worker check the DNS records (in seconds)
RETRY_INTERVAL=2
# how many retries before stoping the worker
MAX_RETRIES=10
```

2. Then run:

```bash
docker-compose up
```
