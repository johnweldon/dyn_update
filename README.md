# dyn_update

Looks up apparent public IP address using OpenDNS and updates dynamic-dns records on Google Domains or Cloudflare depending on environment variables

# Usage

Set environment for Google Domains or CloudFlare.
If both are set then each will be tried until they all succeed or one fails.

## Google Domains

Set these environment variables:

```sh
GD_USERNAME=<google domains username>
GD_PASSWORD=<google domains password>
GD_HOSTNAME=<dns record to update>
```

## CloudFlare

Set these environment variables:

```sh
CF_TOKEN=<cloudflare token>
CF_ZONE_ID=<cloudflare zone id>
CF_HOSTNAME=<dns record to update>
```
