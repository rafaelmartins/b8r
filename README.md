# b8r
I don't know what I'm doing


## Bash completion

```
complete -C /path/to/b8r -o default b8r
```


## MPV plugin (Linux/Mac only)

```
$ mkdir -p ~/.config/mpv/scripts/
$ ln -s /path/to/b8r ~/.config/mpv/scripts/b8r.run
```

### Run from development build

```
$ ln -s b8r b8r.run
$ mpv --load-scripts=no --script=./b8r.run ...
```
