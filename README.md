Tailscale Redirect
==================

A simple tailscale service which simply redirects all requests to another URL.

This was designed for when we need to migrate internal services off of tailscale so existing
users get redirected and we don't end up having to remind people of the new URLs all the time

Usage
=====

tailscale-redirect --hostname my-tailscale-service --target https://my-service.internal.mycorp.com
