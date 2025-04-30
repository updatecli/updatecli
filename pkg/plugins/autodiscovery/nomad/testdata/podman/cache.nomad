job "redis" {
  group "cache" {
    network {
      port "redis" { to = 6379 }
    }

    task "redis" {
      driver = "podman"
      config {
        image = "docker.io/library/redis:7"
        ports = ["redis"]
      }
    }
  }
}
