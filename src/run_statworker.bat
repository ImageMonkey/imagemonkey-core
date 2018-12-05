@ECHO off
set np=..\libs\ext\windows\libstd\;..\libs\ext\windows\libvips\;..\libs\ext\windows\opencv\
echo %path%|find /i "%np%">nul  || set path=%np%;%path%

go build -o statworker.exe statworker.go api_secrets.go && statworker.exe