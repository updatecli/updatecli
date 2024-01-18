resource "person" "john" {
  first_name  = "John"
  middle_name = ""
  surname     = "Doe"
  age         = 30

  nested {
    attr1 = "val1"
  }
}
