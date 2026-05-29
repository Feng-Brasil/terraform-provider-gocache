terraform {
  required_providers {
    gocache = {
      source  = "local/gocache"
      version = "1.0.0"
    }
  }
}

provider "gocache" {}

resource "gocache_dns_record" "teste" {
  domain  = "estadiomaracana.com.br"
  type    = "A"
  name    = "terraform-teste"
  content = "1.1.1.1"
  ttl     = 300
}
resource "gocache_dns_record" "api" {
  domain  = "estadiomaracana.com.br"
  type    = "A"
  name    = "terraform-challenge"
  content = "8.8.8.8"
  ttl     = 120
}

data "gocache_dns_records" "domain" {
  domain = "estadiomaracana.com.br"
}

output "dns_records" {
  value = data.gocache_dns_records.domain.records
}
