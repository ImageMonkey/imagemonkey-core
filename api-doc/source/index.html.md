---
title: API Reference

language_tabs: # must be one of https://git.io/vQNgJ
  - shell

toc_footers:
  - <a href='#'>Fork us on Github</a>

includes:
  - errors

search: true
---

# Introduction

Welcome to the ImageMonkey API! 

ImageMonkey is a public, open sourced image dataset project where users can contribute photos to and tag them accordingly.

# Donations

## Donate a picture


```shell
curl \
  -F "label=banana" \
  -F "image=@/home/user/Desktop/banana.jpg" \
  https://api.imagemonkey.io/v1/donate
```

Upload a picture and tag it with a specific label.

### HTTP Request

`POST https://api.imagemonkey.io/v1/donate`

### Parameters

Parameter | Description
--------- | ------- | -----------
label  | image description
image | the image you want to donate

<aside class="warning">
Donating a picture only works if you specify a valid (i.e existing) label. If the label doesn't exist, please create a pull request, so that we can add it. You can see the list of available labels <a href="https://github.com/bbernhard/imagemonkey-core/blob/master/wordlists/en/misc.txt">here</a>.
</aside>



## Get a specific donation

```shell
curl https://api.imagemonkey.io/v1/donation/48750a63-df1f-48d1-99ee-6c60e535a271
```
The above command returns the donated image.

### HTTP Request

`GET https://api.imagemonkey.io/v1/donation/<uuid>`

Parameter | Description
--------- | ------- | -----------
uuid  | the uuid of the image you want to fetch





# Validation
## Random Image for Validation

```shell
curl "https://api.imagemonkey.io/v1/validate"
```

> The above command returns JSON structured like this:

```json
{
    label: 'dog'
    provider: 'donation'    
    uuid: '48750a63-df1f-48d1-99ee-6c60e535a271'
}
```

This endpoint returns the ID of an randomly chosen image together with some usuful metadata.

### HTTP Request

`GET https://api.imagemonkey.io/v1/validate/`





## Specific Image for Validation

```shell
curl "https://api.imagemonkey.io/v1/validate/48750a63-df1f-48d1-99ee-6c60e535a271"
```

> The above command returns JSON structured like this:

```json
{
    label: 'dog'
    provider: 'donation'
    uuid: '48750a63-df1f-48d1-99ee-6c60e535a271'
}
```


This endpoint returns some useful metadata for the specified uuid.

### HTTP Request

`GET https://api.imagemonkey.io/v1/validate/<uuid>`

Parameter | Description
--------- | ------- | -----------
uuid  | the uuid of the image you want to validate



## Validate Image

```shell
curl "https://api.imagemonkey.io/v1/validate/48750a63-df1f-48d1-99ee-6c60e535a271/yes"
```

This endpoint makes it possible to validate a specific image.

### HTTP Request

`POST https://api.imagemonkey.io/v1/validate/<uuid>/<action>`

Parameter | Description
--------- | ------- | -----------
uuid  | image id
action | either 'yes' or 'no'






# Export
## Export datasets

```shell
curl "https://api.imagemonkey.io/v1/export/tags=dog"
```

> The above command returns JSON structured like this:

```json
{
    label: 'dog'
    provider: 'donation'
    uuid: '48750a63-df1f-48d1-99ee-6c60e535a271'
    probability: 0.9230769
    num_yes: 12
    num_no: 1
}
```

This endpoint returns a collection of images that belong to one or more specific tags.

### HTTP Request

`GET https://api.imagemonkey.io/v1/export`



Parameter | Description
--------- | ------- | -----------
tags  | Comma separated list of tags you are interested in. 






# Report
## Mark an image as abusive

```shell
curl -H "Content-Type: application/json" 
     -X POST -d '{"reason":"picture contains nudity"}' 
     https://api.imagemonkey.io/v1/report/48750a63-df1f-48d1-99ee-6c60e535a271
```

### HTTP Request

`POST https://api.imagemonkey.io/v1/report/<uuid>`



Parameter | Description
--------- | ------- | -----------
uuid  | The uuid of the image which is abusive
reason  | Brief explanation why this picture is inappropriate.
