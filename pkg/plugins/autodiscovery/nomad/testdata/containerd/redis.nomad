job "redis" {
  datacenters = ["dc1"]

  group "redis-group" {
    task "redis-task" {
      driver = "containerd-driver"

      config {
        image = "docker.io/library/redis:7"
      }

      resources {
        cpu    = 500
        memory = 256
      }
    }
  }
}