# ffl
Windows app launcher on command prompt with fuzzy-finder

## Requirments
[peco](https://github.com/peco/peco)

Put peco executable in a folder which is in `%PATH%`

## Launch ffl itself
It's convenient to use Windows `Run` to launch `ffl`.

Press <kbd>Win</kbd>–<kbd>r</kbd>, then type `ffl` 

Please make sure that `ffl.exe` is also in a folder which is in `%PATH%`

## How it works?
1. `ffl` searches Windows Recent folder `%USERPROFILE%\AppData\Roaming\Microsoft\Windows\Recent` and lists shortcut under the folder
1. Get target path and args for each shortcut
1. Run `peco`
1. Run selected file with default app or open `Explorer` if it's a folder

## Configuration
To add folders to search target, create `fflconf.json` and put it in the same folder with `ffl.exe`.

```json
{
    "Folders": [
        "C:/Users/user/shortcut/app",
        "C:/Users/user/shortcut/folder"
    ]
}
```
