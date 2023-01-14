# GoTeleghraphUploader
### Golang tool for creating telegraph articles from image folders

---

## TODO
- header and footer text

## How to use
- [download program](https://github.com/ZUMORl/GoTeleghraphUploader/releases) 
- create or copy [config.yaml](https://github.com/bohdanch-w/go-tgupload/blob/master/config.yaml) file. Place them in same directory
- fill config.yaml with [correct parameters](#configuration)
- start program

---

## Post Configuration
program is configured via config.yaml which has such structure:
```
title: 'Article title here'
img_folder: 'path/to/img/folder'
title_img_path: 
  - 'path/to/title.png'
caption_img_path: 
  - 'path/to/caption.png'
auth_token: 'abcdefghijklmnopqrstuvwxyz123'
output: 'chapter_link.txt'
auto_open: true

author_name: 'Author Name Full'
author_short_name: 'Short Author Name'
author_url: 'https://t.me/ZUMORl'
```

### Minimal required config
```
title: 'Article title here'
img_folder: 'path/to/img/folder'
auto_open: true

author_name: 'Author Name Full'
author_url: 'https://t.me/ZUMORl'
```

| option                  | description                                                                                                                                                                     | type   | required |
| ----------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------ | -------- |
| title                   | Title of the acticle                                                                                                                                                            | string | true     |
| img_folder              | Path to folder with images                                                                                                                                                      | path   | true     |
| title_img_path          | Path to title image (placed as first)                                                                                                                                           | path   | false    |
| caption_img_path        | Path to caption image (placed in the end)                                                                                                                                       | path   | false    |
| auth_token              | Telegraph identification token. Instruction how to find it - [here](#accessing-telegraph-access-token). You won't be able to edit generated article if you don't set this field | string | false    |
| author_name             | Full author name                                                                                                                                                                | string | true     |
| author_short_name       | Short Author Name                                                                                                                                                               | string | false    |
| author_url              | Link to follow on author click                                                                                                                                                  | url    | false    |
| output                  | Path to file with resulting article url if needed                                                                                                                               | path   | false    |
| auto_open               | Set true if you want to automatically open the article in browser                                                                                                               | bool   | false    |
| intermid_data_enabled   | Allows you to save uploaded images if some of them failed to load correctly (Don't use if you don't understand it's purpose)                                                    | bool   | false    |
| intermid_data_save_path | Path to save itermidiate images data                                                                                                                                            | path   | false    |
| intermid_data_load_path | Path to load itermidiate images data if previous atempt failed                                                                                                                  | path   | false    |

---

## Starting from sourse

### Golang version 1.19 or higher

Download sourse folder
Build with
```
go build .
```
Has optional parameter ```config``` which sets path to configuration file
```
GoTeleghraphUploader.exe -config path/to/config.yaml
```

Or just with 
```
go run main.go -config path/to/config.yaml
```

### Usage
```
tg-upload run config.yaml
```

Global commands:
```
NAME:
   tg-upload - cli tool to automate uploading to telegra.ph

USAGE:
   tg-upload [global options] command [command options] [arguments...]

COMMANDS:
   version  
   run      post telegraph article according to specified config
   upload   upload file to telegraph CDN
   help, h  Shows a list of commands or help for one command
```

Run usage (post article):
```
USAGE:
   tg-upload run [command options] <path-to-config>

OPTIONS:
   --loglevel value  level of logging for application
   --cache value     path to saved cache. If specified will use caching for CDN uploads
   --no-dialog, -s   don't prompt window for user input (default: false)
   --help, -h        show help (default: false)
```

Upload usage (post article):
```
USAGE:
   tg-upload upload [command options] <file-path> [more-files...]

OPTIONS:
   --loglevel value  level of logging for application
   --cache value     path to saved cache. If specified will use caching for CDN uploads
   --output value    path to saved output file
   --plain           use plain output format (default: false)
   --help, -h        show help (default: false)
```

---

## Accessing telegraph access token
- Go to https://telegra.ph/
- Open devtools (Right click on path -> Inspect OR F12)
- Open ```Network``` tab
- Reload page (Ctrl+R)
- Filter requests by ```Fetch/XHR```
- In list of requests choose ```check```
- In opened bar choose ```cookies```
- In ```Request cookies``` find cookies with name ```tph_token``` 
- Now you can copy its value
