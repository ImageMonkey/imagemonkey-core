@ECHO off
set np=..\libs\ext\windows\libstd\;..\libs\ext\windows\libvips\;..\libs\ext\windows\opencv\
echo %path%|find /i "%np%">nul  || set path=%np%;%path%

go build -o api.exe api.go auth.go label_graph.go && api.exe
