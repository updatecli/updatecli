# CONTRIBUTING

Thanks for your interest in this project, feel free to ask for additional information if you can't find what you are looking for here

## REQUIREMENTS

To build the project, just ensure to have access to the correct golang version then just run `make build` which should return something like:

```
echo v0.0.15-6-g1e24f1d
v0.0.15-6-g1e24f1d
go build \
	-ldflags "-w -s \
        -X \"github.com/olblak/updateCli/pkg/version.BuildTime=`date -R`\" \
        -X \"github.com/olblak/updateCli/pkg/version.GoVersion=go version go1.14.2 linux/amd64\" \
        -X \"github.com/olblak/updateCli/pkg/version.Version=v0.0.15-6-g1e24f1d\""\
        -o bin/updatecli
```

or using docker

```
docker run --rm -v "$PWD":/usr/src/updatecli -w /usr/src/updatecli -e GOOS=windows -e GOARCH=386 golang:1.14 go build -v
```

## CONTRIBUTE
They are multiple ways to contribute which don't necessarily involve coding like improving documentation, processes, or just providing feedback.

### CODE
#### SOURCE

You can easily add a source not supported by:

* Creating a new directory using your packageName under the directory `pkg` like

```
pkg
├── packageName
│   ├── main.go
│   └── main_test.go

```

* In the main.go, you need to define a struct where each field needs to be capitalized so they can be filled when unmarshalling from the future configuration.yaml

```
type Capitalized_package_name struct {
	Field1        string
	Field2        string
	Field3        string
	Field4        string
}
```

* Your packageName must respect the source interface contract by defining a function named 'Source()' using no parameters and returning two parameters.

```
func (c *CapitalizedPackageName) Source() (string, error) {
  //
return 
```

* In the function Execute, inside the file pkg/engine/source/main.go, you need to add a case statement that binds a kind value and you packageName

```
	case "packageName":
		p := packageName.PackageName{}
		err := mapstructure.Decode(s.Spec, &p)

		if err != nil {
			return err
		}

		spec = &p
```

Now something like this, should be working:

config.value
```
# updatecli diff --config config.value

source:
  kind: packageName
  spec:
    field1: "value"
    field3: "value"
targets:
  idName:
    name: "updatecli"
    kind: "yaml"
    prefix: "olblak/polls@256:"
    spec:
      file: "..."
      key:  "..."
```

#### CONDITION
Adding a condition follows the same idea then for source, just replace source by condition.


### DOCUMENTATION

You'll more than probably spot phrasing issues or just a lack of documentation, feel free to open an issue and or a pull request with your contribution.
