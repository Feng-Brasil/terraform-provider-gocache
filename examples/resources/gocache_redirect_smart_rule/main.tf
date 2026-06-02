resource "gocache_redirect_smart_rule" "blog" {
  domain = "gocache.com.br"
  name   = "Redirect blog"
  status = true

  match = {
    request        = "http://www.gocache.com.br/blog/*"
    request_method = ["GET", "POST"]
  }

  action = {
    redirect_type = 301
    redirect_to   = "https://blog.gocache.com.br/$1"
  }
}
