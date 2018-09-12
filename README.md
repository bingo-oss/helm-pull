# Helm pull plugin
------
Pull charts into local folder. Will retry failed GET requests and only pull charts that do not already exist.

Examples:
```
  $ helm pull ./ repo --source-image-repo=registry.bingosoft.net --target-image-repo=hub.bingosoft.net --insecure=true
```

Usage:
```
helm pull [local_folder] [repo_name] [flags]
```

Flags:

```
-h, --help                  help for helm
--insecure                  Connect to server with an insecure way by skipping certificate verification [$HELM_REPO_INSECURE]
--source-image-repo string  Current image repository [$HELM_SOURCE_IMAGE_REPO]
--target-image-repo string  Change to the new image repository [$HELM_TARGET_IMAGE_REPO]
```