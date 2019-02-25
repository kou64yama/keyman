# データ構造

[English](data-structure.md) |
日本語

## バケット

### `refs`

有効なリビジョンへの参照を管理するバケット。
パスワード名に対し、有効なリビジョン番号を格納する。

| key     | value  |
| ------- | ------ |
| mysql   | 1      |
| google  | 3      |
| twitter | 7      |

### `blob`

全てのパスワードを格納するバケット。
パスワード名とリビジョン番号の組み合わせ `{name}:{revision}` に対し、
パスワードのバイト列を格納する。

| key       | value            |
| --------- | ---------------- |
| mysql:1   | \*\*\*\*\*\*\*\* |
| google:1  | \*\*\*\*\*\*\*\* |
| google:2  | \*\*\*\*\*\*\*\* |
| google:3  | \*\*\*\*\*\*\*\* |
| twitter:1 | \*\*\*\*\*\*\*\* |
| twitter:2 | \*\*\*\*\*\*\*\* |

### `history:{name}`

メタデータの履歴を管理するバケット。パスワード名ごとに存在する。
リビジョン番号に対して、メタデータを格納する。

| key | value      |
| ---:| ---------- |
|   1 | {Metadata} |
|   2 | {Metadata} |
|   3 | {Metadata} |