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

Welcome to the ImageMonkey API! ImageMonkey is a public, open sourced image dataset project where users can contribute photos to and tag them accordingly.

# Donations

## Donate a picture


```shell
curl \
  -F "label=banana" \
  -F "image=@/home/user/Desktop/banana.jpg" \
  https://api.imagemonkey.com/v1/donate
```

> The above command returns JSON structured like this:

```json
[
  {
    "id": 1,
    "name": "Fluffums",
    "breed": "calico",
    "fluffiness": 6,
    "cuteness": 7
  },
  {
    "id": 2,
    "name": "Max",
    "breed": "unknown",
    "fluffiness": 5,
    "cuteness": 10
  }
]
```

Upload a picture and tag it with a specific label.

### HTTP Request

`POST https://api.imagemonkey.com/v1/donate

### Parameters

Parameter | Description
--------- | ------- | -----------
label  | a single string which labels the image
image | the image you want to donate

<aside class="warning">
Donating a picture only works if you specify a valid (i.e existing) label. If the label doesn't exist, please create a pull request, so that we can add it.
</aside>





# Validation
## Random Image for Validation

```shell
curl "https://api.imagemonkey.com/v1/validate"
```

> The above command returns JSON structured like this:

```json
{
    label: 'dog'
    provider: 'donation'
    url: '/donations/48750a63-df1f-48d1-99ee-6c60e535a271'
    uuid: '48750a63-df1f-48d1-99ee-6c60e535a271'
}
```

This endpoint returns the ID of an randomly chosen image together with some usuful metadata.

### HTTP Request

`GET https://api.imagemonkey.com/v1/validate/`





## Specific Image for Validation

```shell
curl "https://api.imagemonkey.com/v1/validate/48750a63-df1f-48d1-99ee-6c60e535a271"
```

> The above command returns JSON structured like this:

```json
{
    label: 'dog'
    provider: 'donation'
    url: '/donations/48750a63-df1f-48d1-99ee-6c60e535a271'
    uuid: '48750a63-df1f-48d1-99ee-6c60e535a271'
}
```


This endpoint returns some useful metadata for the specified ID.

### HTTP Request

`GET https://api.imagemonkey.com/v1/validate/<ID>`




## Validate Image

```shell
curl "https://api.imagemonkey.com/v1/validate/48750a63-df1f-48d1-99ee-6c60e535a271/yes"
```

This endpoint makes it possible to validate a specific image.

### HTTP Request

`POST https://api.imagemonkey.com/v1/validate/<ID>/<action>`

Parameter | Description
--------- | ------- | -----------
ID  | image id
action | either 'yes' or 'no'






# Export
## Export datasets

```shell
curl "https://api.imagemonkey.com/v1/export/tags=dog"
```

> The above command returns JSON structured like this:

```json
{
    label: 'dog'
    provider: 'donation'
    url: '/donations/48750a63-df1f-48d1-99ee-6c60e535a271'
    uuid: '48750a63-df1f-48d1-99ee-6c60e535a271'
    probability: 0.9230769
    num_yes: 12
    num_no: 1
}
```

This endpoint returns a collection of images that belong to one or more specific tags.

### HTTP Request

`GET https://api.imagemonkey.com/v1/export`



Parameter | Description
--------- | ------- | -----------
tags  | Comma separated list of tags you are interested in. 






# Report
## Mark an image as abusive

```shell
curl -H "Content-Type: application/json" 
     -X POST -d '{"reason":"picture contains nudity"}' 
     https://api.imagemonkey.com/v1/report/48750a63-df1f-48d1-99ee-6c60e535a271
```

### HTTP Request

`POST https://api.imagemonkey.com/v1/report/<ID>`



Parameter | Description
--------- | ------- | -----------
reason  | Brief explanation why this picture is inappropriate.
