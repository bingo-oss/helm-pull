# Helm pull plugin
------
Pull charts into local folder. Will retry failed GET requests and only pull charts that do not already exist.

Examples:
```
$ helm pull ./ repo													# Only pull
$ helm pull ./ repo "registry.bingosoft.net" "hub.bingosoft.net"	# Pull and replace
```

Usage:
```
helm pull [local_folder] [repo_name] [replace_old] [replace_new]
```

Flags:

```
-h, --help                  help for helm
    --insecure  bool        Connect to server with an insecure way by skipping certificate verification [$HELM_REPO_INSECURE]
```