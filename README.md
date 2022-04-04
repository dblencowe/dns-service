# DNS Service

Accept DNS queries and perform encrpyted lookups via Cloudflares HTTPS DNS api

## Standalone DNS Server
Bind to port 53 and translate local dns queries

### Running a pre-compiled binary
To run the DNS server locally run the compiled binary. You may need admin priveleges as the program binds to port 53 to accept queries

```bash
sudo ./dns-service
> 2022-04-04T22:53:28+01:00: [INFO] listening on port [53]
```

Once running you can test the server with dig
```bash
dig @localhost google.com
```

To see more verbose information
```base
sudo OUTPUT_LEVEL=0 ./dns-service
```

### Filtering requests by domain
The server has the ability to redirect domains to a custom IP, often used for ad-blocking or internal hosts.
To setup a filter list create a text file like below, with each record on a new line
```text
127.0.0.1 example1.com
127.0.0.1 example2.com
```

Then start the server with the following environment variable set
```bash
sudo FILTER_FILE=myfilters.txt ./dns-service
```

## DNS Library

More to come