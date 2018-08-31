# Keyman

Keyman manages secret values.

## Usage

`keyman set`:

```shell
$ keyman set mysql.root
Password:
```

`keyman exec`:

```shell
$ keyman exec mysql -uroot -p%mysql.root%
mysql>
```

```shell
$ keyman list -v
mysql.root
    Revision: 1
    Created At: 2018-09-08 16:32:50.6490388 +0900 JST

twitter.secret
    Revision: 3
    Created At: 2018-09-08 16:32:50.6490388 +0900 JST

github.token
    Revision: 2
    Created At: 2018-09-08 16:32:50.6490388 +0900 JST

```

## Installation

```bash
go get -u github.com/kou64yama/keyman/cmd/keyman
```

## More examples

```bash
openssl genrsa | keyman in private_key
```

```bash
openssl rsa -in <(keyman out private_key) -pubout
```
