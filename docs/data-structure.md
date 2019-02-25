# Data Structure

English |
[日本語](data-structure.ja.md)

## Bucket

### `refs`

A bucket that manages ererences to valid revisions.  For a password
name, it store a valid revision number.

| key     | value  |
| ------- | ------ |
| mysql   | 1      |
| google  | 3      |
| twitter | 7      |

### `blob`

A bucket that store all passwords.  It store a byte array of the
password for the set of password name and revision number
`{name}:{revision}`.

| key       | value            |
| --------- | ---------------- |
| mysql:1   | \*\*\*\*\*\*\*\* |
| google:1  | \*\*\*\*\*\*\*\* |
| google:2  | \*\*\*\*\*\*\*\* |
| google:3  | \*\*\*\*\*\*\*\* |
| twitter:1 | \*\*\*\*\*\*\*\* |
| twitter:2 | \*\*\*\*\*\*\*\* |

### `history:{name}`

A bucket that manage history of metadata.  It exists for password
name.  It store metadata for revision number.

| key | value      |
| ---:| ---------- |
|   1 | {Metadata} |
|   2 | {Metadata} |
|   3 | {Metadata} |
