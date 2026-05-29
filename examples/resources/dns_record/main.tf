resource "gocache_dns_record" "test" {
  domain  = "gocache.com.br"
  type    = "A"
  name    = "@"
  content = "1.1.1.1"
  ttl     = 300
}
