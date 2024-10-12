#!/bin/zsh

input_movie_filename=$1
background_file_image_name=$2
start_time=$3
end_time=$4
output_filename=$5

temp_audio_filename="temp_audio.aac"
trim_audio_filename="trim_audio.aac"

# clean temporary files
rm decomposed_ffmpeg_frames/*.png
rm edited_movie_frames/*.png

rm $temp_audio_filename 2> /dev/null
rm $trim_audio_filename 2> /dev/null

# decompose movie frames
ffmpeg -y -i $input_movie_filename decomposed_ffmpeg_frames/frame_%d.png

# get audio data from movie
ffmpeg -y -i $input_movie_filename -vn -acodec copy $temp_audio_filename

# trim audio data
ffmpeg -y -i $temp_audio_filename -ss $start_time -to $end_time -c copy $trim_audio_filename

# edit frame files
./compose_image_files_for_short_movie -bg_image $background_file_image_name -start_time $start_time -end_time $end_time -pic_zoom_coeff 1.3

# convert images and audio files to movie
ffmpeg -y -framerate 30 -i edited_movie_frames/frame_%04d.png -i $trim_audio_filename -c:v libx264 -r 30 -pix_fmt yuv420p -c:a aac -strict experimental $output_filename
