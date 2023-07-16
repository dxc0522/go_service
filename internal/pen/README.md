# Pen

pen is used to draw lines, 
it currently contains three functionality.

* server side codegen from openapi yaml
* client side codegen from openapi yaml
* generate migration files from db

## Use

### Method1(suggested): 

```
go env -w GOPROXY=https://arf.tesla.cn/artifactory/api/go/gfsh-sdk-go-virtual,https://goproxy.cn,direct
go env -w GONOSUMDB=*.teslamotors.com,*.tesla.cn

go get -u -v github.com/go_service/internal/pen
```

### Method2(cutting edge version): 

following this [instruction](https://golang.org/doc/faq#git_https)

* `go env -w GOPRIVATE="*.tesla.*"`
* `machine github.tesla.cn login {UserName} password {PersonalAccessToken}` INTO `~/.netrc`

Note: the PersonalAccessToken can be created [here](https://github.tesla.cn/settings/tokens)

# Install
for version 1.18 and later

```
cd $PENPATH (the path where PEN codes are saved)
go install .
```

for version 1.3

```
cd $GOPATH
go get -u -v github.com/go_service/internal/pen
```

# Usage

## New Module

```sh
pen module testmodule
```

will regerate three file

```
testmodule.yaml
Makefile					// quick cammand collection
pen.yaml					// as pen's optional config. you can write at here only once in place of pass in every command
```

## Gen SDK Client

```
 pen client testmodule.yaml
```

Or if use with `pen.yaml` ,can simply write as

```
pen client
```

Detail Usage see `pen client -h`

## Gen Structure

```
pen structure -app-package="github.tesla.cn/itapp/benjamin/internal/$(ModuleName)"  $(ModuleName).yaml
```

Or if use with `pen.yaml` ,can simply write as

```
pen structure
```

Detail Usage see `pen structure -h`

## Gen Migration

```
pen migration -data-source 'root:root@tcp(127.0.0.1)/bjm?collation=utf8_unicode_ci&parseTime=true&loc=UTC' -table-prefix $(ModuleName)_
```

Or if use with `pen.yaml` ,can simply write as

```
pen migration
```

Detail Usage see `pen migration -h`

## Gen Dbmodel

```
pen dbmodel \
-sqltype mysql \
-connstr 'root:root@tcp(127.0.0.1)/bjm?collation=utf8_unicode_ci&parseTime=true&loc=UTC'
-d bjm \
-t $(ModuleName)_sometable,$(ModuleName)_anothertable \
-model=dbmodel \
-gorm \
-guregu \
-overwrite 、
-out ./
```

Or if use with `pen.yaml` ,can simply write as

```
pen dbmodel
```

Detail Usage see `pen dbmodel -h`

## .pen suffix

Files have `.pen` suffix in the name is a type of file that should not be changed by user.

It will be override automatically by `pen` command,

If you really wants change it, please remove the `.pen` part,

then `pen` will know this file is out of auto generation.

It will not try to the corresponding `.pen` file.

# Example

* [petstore](../../example/petstore/Makefile)

