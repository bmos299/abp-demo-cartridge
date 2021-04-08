To use these, copy them into `.git/hooks`


### `commit-msg`
This will run the `gitlint` linter on your commit messages while you write them, so you can check them locally before pushing them for Travis to check.

Commit messages should take the form:

```
feat: <subject>

<body>

Contributes to: automation-base-pak/abp-planning#<issue number>

Signed-off-by: Your Name <email@ibm.com>
```

#### Pre-reqs:
- Python
- `pip install gitlint`

#### Install:
- `cp gitconfig/githooks/commit-msg .git/hooks/.`

#### Uninstall:
- `rm .git/hooks/commit-msg`
