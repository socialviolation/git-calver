# git-calver

`git calver` is intended as a git subcommand utility, to manipulate git tags using [CalVer](https://calver.org/).

## Installation

```bash
go install github.com/socialviolation/git-calver
```

## Set up

`calver` requires a CalVer format to be provided.
```bash
# .git/config setting - Lowest Priority
$ git calver format
YYYY.0M
# It can be set with calver
$ git calver format set --format="YYYY.0M"

# OR Environment Var - medium priority
export CALVER="YYYY.0M.0D"

# OR FLAG - highest priority
$ git calver tag --format="YY.0M.DD"

#specify -A for auto-incrementing
# OR FLAG - highest priority
$ git calver tag --format="YY.0M-A"
```

## Usage
```bash
$ git calver help
CalVer is a git subcommand for managing a calendar versioning tag scheme.

Usage:
  git-calver [flags]
  git-calver [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  format      Get format from .gitconfig
  help        Help about any command
  latest      Get latest tag matching the provided format
  list        Will list all CalVer tags matching the provided format
  next        Output what the next calver tag will be
  retag       retag
  tag         tag
  untag       untag

Flags:
  -d, --dry-run           Dry run
  -f, --format string     format of calver (YYYY.0M.0D)
  -h, --help              help for git-calver
      --micro uint        Micro Version
      --minor uint        Minor Version
      --modifier string   Modifer (eg. DEV, RC, etc)

Use "git calver [command] --help" for more information about a command.
```


## Supported Formats

Review [calver.go](./ver/calver.go) for the calver format spec

Supported values are as follows:
```text
// FullYear notation - 2006, 2016, 2106
FullYear = "YYYY"
// ShortYear notation - 6, 16, 106
ShortYear = "YY"
// PaddedYear notation - 06, 16, 106
PaddedYear = "0Y"
// ShortMonth notation - 1, 2 ... 11, 12
ShortMonth = "MM"
// PaddedMonth notation - 01, 02 ... 11, 12
PaddedMonth = "0M"
// ShortWeek notation - 1, 2, 33, 52
ShortWeek = "WW"
// PaddedWeek notation - 01, 02, 33, 52
PaddedWeek = "0W"
// ShortDay notation - 1, 2 ... 30, 31
ShortDay = "DD"
// PaddedDay notation - 01, 02 ... 30, 31
PaddedDay = "0D"
// Auto Increment notation - `-A` 
Auto = "-A"

Minor = "MINOR"
Micro = "MICRO"
```
