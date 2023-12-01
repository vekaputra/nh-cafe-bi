# Guide

## 1. MJML
For easier edit, please copy content of otp.mjml to https://mjml.io/try-it-live, then click view html to get the html, there is minify html flag if you need it minified. Current otp.mjml can be found in https://mjml.io/try-it-live/BC8xvw_4LO

## 2. HTML
I provide raw html in otp.html file or the minified one in otp.min.html, you can search by content example: "Sarah****" and replace it

## 3. HTML Template
I also provide template html in otp.tmpl.html; you only need to do string replace from your backend, the list of value that can be modified is:
- {{naik_logo_url}}
- {{nh_logo_url}}
- {{customer_name}}
- {{otp}}
- {{app_store_url}}
- {{google_play_url}}
- {{web_logo_url}}
- {{ig_logo_url}}
- {{yt_logo_url}}

## Warning
Currently i use my own server to host the image, please change the image URL ASAP, because when i delete the image from my server, it will not show up on the email sent. To check, please search the image src with url starts with `https://www.ashiwawa.com/*`