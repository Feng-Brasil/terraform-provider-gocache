resource "gocache_general_smart_rule" "cache_static" {
  domain = "gocache.com.br"
  name   = "Cache static assets"
  status = true
  notes  = "Full cache for the static assets directory"

  match = {
    host        = "www.gocache.com.br"
    scheme      = "https"
    request_uri = "/static/*"
  }

  action = {
    cache_mode  = "full"
    cache_ttl   = 86400
    expires_ttl = 14400
    gzip_status = true
  }
}
