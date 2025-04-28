job "multi-docker-example" {
  datacenters = ["dc1"]

  group "web-group" {
    network {
      port "web" {
        static = 8080
      }
      port "api" {
        static = 9090
      }
    }

    task "frontend" {
      driver = "docker"

      config {
        image = "nginx:latest"
        ports = ["web"]
      }

      resources {
        cpu    = 500
        memory = 256
      }
    }

    task "backend" {
      driver = "docker"

      config {
        image = "hashicorp/http-echo:latest"
        args  = [
          "-listen", ":9090",
          "-text", "Hello from Backend"
        ]
        ports = ["api"]
      }

      resources {
        cpu    = 500
        memory = 256
      }
    }
  }
}