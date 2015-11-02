# envmap

[![Circle CI](https://circleci.com/gh/higanworks/envmap.svg?style=svg)](https://circleci.com/gh/higanworks/envmap)

Create map object of environment variables.


## Installation

`go get github.com/higanworks/go-envmap`

## Usage

`import "github.com/higanworks/go-envmap"`

## Functions

```
func All() map[string]string
    Returns all environment variables as key-value map[string]string.

func ListKeys() []string
    Returns Keys of all environment variables as []string.

func Matched(rule string) map[string]string
    Returns filtered by matched keys of environment variables as key-value
    map[string]string.
```


## Example

```
package main

import (
  "fmt"
  envmap "github.com/higanworks/go-envmap"
)

func main() {
  envs := envmap.All()
  keys := envmap.ListKeys()

  fmt.Println(envs["HOME"])
  for _, key := range keys {
    fmt.Println(key)
  }

}
```

Outputs..

```
/Users/sawanoboly
rvm_bin_path
TERM_PROGRAM
GEM_HOME
SHELL
TERM
MYVIMRC
IRBRC
...
```

