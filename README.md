# Short Movie Generator With Background Image

generate short movie for YouTube and other movie platforms by merging background image and mp4 movie.

## How to execute

```
./execute_convert.sh (background_filename) (start_time)
```

for example:

```
./execute_convert.sh ./background1.png 00:00:18.6
```

## How to build

```
go build -ldflags "-w" compose_image_files_for_short_movie.go
```

## How to get image files from mp4 movie

ffmpeg -i hoge.mp4 decomposed_ffmpeg_frames/frame_%d.png

## How to get audio files from mp4 movie

ffmpeg -i hoge.mp4 -vn -acodec copy mymusic.aac

## How to trim audio files

```
ffmpeg -i mymusic_raw.aac -ss 00:00:12 -c copy trim.aac
ffmpeg -i mymusic_raw.aac -ss 00:00:12.5 -c copy trim.aac
```

## How to get movie or audio file length(time)

```
ffprobe hoge.mp4
ffprobe hoge.aac
```
