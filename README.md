# GoTeleghraphUploader
### Golang tool for creating telegraph articles as image galeries

---

## How to use
- [download program](https://github.com/bohdanch-w/go-tgupload/releases)
- configure the program. See details in [configuration](#configuration)
- run program with preferred parameters

---

## Configuration
Configuration is performed by subcommands `config` and `account`. Second is meant specifically for telegra.ph account configuration.

### 1.a If you don't have a telegra.ph account:
Run this command and follow instructions to configure the basic information
```
> gotg account setup
```
Followed by:
```
> gotg account login
```
This will create a new telegra.ph account, which will be used for future posts. To use the same account via browser, additionally run `gotg account web-login`.

To validate which account is used and get basic information, use this command:
```
gotg account validate
```

You can additionally check your token if needed by executing `gotg account token`

### 1.b If you already have a telegra.ph account:
- Open telegra.ph.
- Make sure you are logged in.
- Open developer tools in your browser (`F12`)
- Go to network tab and reload the page
- In the list of requests find the one with name `check`
- In headers section under `Request Headers` find Cookie. To the right of it is its value in format like this `tph_uuid=xxx; tph_token=yyy`
- `tph_token` is what is needed. Copy whatever value is after the equal sign (usually it is a 60 character long string)
- set the token with command `gotg config set tg-access-token <your-token>`.
- execute command `gotg account sync`. Be warned, that this command, if successful, replaces all other account configuration if was previosly set. (This is opposite to using `gotg account login`)

### 2. Next step is to configure CDN.
Telegra.ph no longer allows storing images on their servers, so this should be done via external servers. Currently the two supported options are `postimages.org` and `S3` compatible storage. For most users the first option is most suitable.

First configure preferred cdn via command:
```
gotg config set preferred-cdn post-image
```
OR
```
gotg config set preferred-cdn s3
```

#### 2.a PostImages.org
To configure this CDN you only need to have an account on the [postimages.org](https://postimages.org/) website and copy your personal API token from [this page](https://postimages.org/login/api).
Then configure the program via following command:
```
gotg config set postimg-api-key <your-token>
```
Or you can choose to set it every time with `--post-img-key` flag or `POST_IMAGE_API_KEY` env value.

#### 2.b S3 compatible storage
If you chose this option, you should already know how to configure your S3 service.
Configuration options are the following keys in config (set via `gotg config set <key> <value>`)
   - aws-key-id
   - aws-secret-access-key
   - aws-region
   - aws-endpoint
   - aws-s3-bucket
   - aws-s3-location: directory where files should be stored. The service will additionally create subfolders with current timestamp to avoid name collisions.
   - aws-s3-public-url: Resulting url will be formed in format: [this public url]/[location]/[timestamp]/[filename].
The same configuration may be achieved via env values:
```
AWS_KEY_ID
AWS_SECRET_ACCESS_KEY
AWS_REGION
AWS_ENDPOINT
AWS_S3_BUCKET
AWS_S3_LOCATION
AWS_S3_PUBLIC_URL
```
Or by providing command line arguments (see help)

## Posting

When everything is configured, you can now use the program to post the articles.
Just run the command
```
gotg post <path-to-folder>
```
Where path to folder should be absolute or relative path to the directory with images you want to post. If no path is specified, you will be promted to choose a directory, unless dialog windows are disabled.
Images will be sorted in natural order, meaning that `2.png` is ordered before `10.png` unlike stardart file explorers, without the need to pad names with zeroes.

#### Full list of configuration options:
```
--loglevel value                               level of logging for application (default: "INFO")
--cache value                                  path to saved cache. If specified will use caching for CDN uploads
--no-dialog, -s                                don't prompt window for user input (default: false)
--parallel value, -p value                     set number of parallel file upload (default: 8)
--cdn value                                    type of cdn to upload images to. Supported values are ['post-image', 's3']
--browser, -a                                  auto open uploaded article in the browser (default: false)
--title value, -t value                        specify the title of the article. If empty, then you will be prompted later. (default: false)
--post-img-key value                           API key for post-image CDN [$POST_IMAGE_API_KEY]
--aws-s3-bucket value, --bucket value          name of the bucket for S3 CDN [$AWS_S3_BUCKET]
--aws-s3-location value, --location value      location in the bucket for S3 CDN [$AWS_S3_LOCATION]
--aws-s3-public-url value, --public-url value  prefix for formed URL for S3 CDN [$AWS_S3_PUBLIC_URL]
```

---

## Starting from sourse

### Golang version 1.24 or higher

Download sourse folder
Build with
```
go build cmd/
```
OR
```
make build
```

## Since part of the programm is just uploading files to CDN, it was decided to allow it's usage as separate command
```
gotg upload [files...]
```
Pathes could be as list of individual files and whole directories.
