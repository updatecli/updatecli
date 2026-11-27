job "mirrored" {
  datacenters = ["dc1"]

  group "updatecli" {
    task "updatecli" {
      driver = "docker"

      config {
        image = "my-registry.com/ghcr-io/updatecli/updatecli:v0.67.0"
      }
    }
  }
}
