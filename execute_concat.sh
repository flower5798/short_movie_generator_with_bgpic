#!/bin/zsh

output_filename=$1
should_last_frame_duplicate=$2

temp_audio_file_list=./temp_audio_file_list.txt
touch $temp_audio_file_list
# 生成した一時ファイルを削除する
function rm_tmpfile {
  [[ -f "$temp_audio_file_list" ]] && rm -f "$temp_audio_file_list"
}

temp_audio_filename="temp_audio.aac"
trim_audio_filename="trim_audio.aac"

# clean temporary files
rm decomposed_ffmpeg_frames/*.png
rm edited_movie_frames/*.png

rm $temp_audio_filename 2> /dev/null
rm $trim_audio_filename 2> /dev/null

shift
shift

all_frame_count=0

for i in `seq 1 ${#}`
do
  current_movie_filename=${1}    #第一引数を表示
  echo "decomposing $current_movie_filename..."

  # 引数で指定されたファイルからframe画像と音声抽出
  current_movie_temp_directory=temp_movie_$i
  current_movie_temp_audio_filename=temp_audio_$i.aac
  mkdir $current_movie_temp_directory 2> /dev/null
  ffmpeg -y -i $current_movie_filename $current_movie_temp_directory/frame_%d.png
  ffmpeg -y -i $current_movie_filename -vn -acodec copy $current_movie_temp_audio_filename
  # 結合するaudioファイルのリストに追加
  echo "file $current_movie_temp_audio_filename" >> $temp_audio_file_list

  # frameの画像数をカウント
  current_movie_freame_file_count=$(ls -U1 $current_movie_temp_directory/*.png | wc -l)

  # リネームしてframe画像を移動
  for j in `seq 1 $current_movie_freame_file_count`
  do
    old_filename=$current_movie_temp_directory/frame_$j.png
    new_cnt=$(printf "%04d" $((j + all_frame_count)))
    new_filename=edited_movie_frames/frame_$new_cnt.png
    cp $old_filename $new_filename
  done

  # 次の動画ファイルで使うため、ファイル移動した後にここで総frame数のカウントを更新
  all_frame_count=$((all_frame_count + current_movie_freame_file_count))

  # 一時ファイル削除
  rm -rf $current_movie_temp_directory

  shift
done

# 最後のフレームが黒画面になってしまう対策
# 元々の動画の最後は黒画面になっていなくても、小数点以下の足し算で新しいフレームが必要になってしまう模様
if [ $should_last_frame_duplicate = 1 ]; then
  echo "copying last frame..."
  last_frame_filename=edited_movie_frames/frame_$(printf "%04d" $all_frame_count).png
  new_frame_filename=edited_movie_frames/frame_$(printf "%04d" $((all_frame_count + 1))).png
  echo "last frame filename: $last_frame_filename"
  echo "new frame filename: $new_frame_filename"

  cp $last_frame_filename $new_frame_filename

  echo "copy finished."
fi

# concat aac audio files
ffmpeg -f concat -i $temp_audio_file_list -c copy $temp_audio_filename

# convert images and audio files to movie
ffmpeg -y -framerate 30 -i edited_movie_frames/frame_%04d.png -i $temp_audio_filename -c:v libx264 -r 30 -pix_fmt yuv420p -c:a aac -strict experimental $output_filename

# TODO 音声の一時ファイル削除

# 正常終了したとき
trap rm_tmpfile EXIT
# 異常終了したとき
trap 'trap - EXIT; rm_tmpfile; exit -1' INT PIPE TERM
