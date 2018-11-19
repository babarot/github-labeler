github-labeler
==============

CLI that sets GitHub labels exactly as written in YAML file

## Usage

```console
$ go build
$ ./github-labeler
```

## What this app does

- Create a label (e.g. when no label described in YAML)
- Edit a label (e.g. when its color was changed)
- Delete a label (e.g. when the label not described in YAML exists on GitHub)

## YAML example

```yaml
labels:
  - name: kind/proactive
    description: Categorizes issue or PR as related to proactive tasks.
    color: 9CCC65
  - name: kind/reactive
    description: Categorizes issue or PR as related to reactive tasks.
    color: FFA000

repos:
  - name: org/repo
    labels:
      - kind/proactive
      - kind/reactive
```

## Author

@b4b4r07

## License

MIT
