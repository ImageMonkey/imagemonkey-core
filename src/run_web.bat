@ECHO off
set np=..\libs\ext\windows\libstd\;..\libs\ext\windows\libvips\;..\libs\ext\windows\opencv\
echo %path%|find /i "%np%">nul  || set path=%np%;%path%

go build -o web.exe web.go web_secrets.go auth.go && web.exe