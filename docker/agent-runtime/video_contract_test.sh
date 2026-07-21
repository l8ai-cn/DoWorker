#!/usr/bin/env bash
set -euo pipefail

work_dir="$(mktemp -d)"
trap 'rm -rf "$work_dir"' EXIT

ffmpeg -hide_banner -filters 2>&1 | grep -E ' (ass|subtitles) ' >/dev/null
fc-match 'Noto Sans CJK SC' | grep -q 'NotoSansCJK'
chromium --version
python3 --version
codex --version
video-studio-codex --version
remotion versions

ffmpeg -hide_banner -loglevel error \
  -f lavfi -i color=c=0x111827:s=1080x1920:d=1:r=30 \
  -c:v libx264 -pix_fmt yuv420p "$work_dir/source.mp4"
cat > "$work_dir/subtitles.ass" <<'ASS'
[Script Info]
ScriptType: v4.00+
PlayResX: 1080
PlayResY: 1920

[V4+ Styles]
Format: Name, Fontname, Fontsize, PrimaryColour, SecondaryColour, OutlineColour, BackColour, Bold, Italic, Underline, StrikeOut, ScaleX, ScaleY, Spacing, Angle, BorderStyle, Outline, Shadow, Alignment, MarginL, MarginR, MarginV, Encoding
Style: Default,Noto Sans CJK SC,72,&H00FFFFFF,&H000000FF,&H00111111,&H80000000,-1,0,0,0,100,100,0,0,1,4,0,2,80,80,180,1

[Events]
Format: Layer, Start, End, Style, Name, MarginL, MarginR, MarginV, Effect, Text
Dialogue: 0,0:00:00.00,0:00:00.90,Default,,0,0,0,,Agent Cloud 视频专家
ASS
ffmpeg -hide_banner -loglevel error \
  -i "$work_dir/source.mp4" \
  -vf "ass=$work_dir/subtitles.ass" \
  -c:v libx264 -pix_fmt yuv420p -an "$work_dir/subtitled.mp4"

mkdir -p "$work_dir/remotion"
ln -s /opt/video-studio/node_modules "$work_dir/remotion/node_modules"
cat > "$work_dir/remotion/index.jsx" <<'EOF'
import React from "react";
import {AbsoluteFill, Composition, registerRoot} from "remotion";

const VerticalVideo = () => (
  <AbsoluteFill style={{
    alignItems: "center",
    backgroundColor: "#111827",
    color: "white",
    display: "flex",
    fontFamily: "Noto Sans CJK SC",
    fontSize: 84,
    justifyContent: "center",
  }}>
    Agent Cloud Video Studio
  </AbsoluteFill>
);

const Root = () => (
  <Composition
    id="VerticalVideo"
    component={VerticalVideo}
    durationInFrames={30}
    fps={30}
    width={1080}
    height={1920}
  />
);

registerRoot(Root);
EOF

(
  cd "$work_dir/remotion"
  remotion render index.jsx VerticalVideo "$work_dir/remotion.mp4" \
    --browser-executable="${CHROME_BIN:-/usr/bin/chromium}" \
    --codec=h264 \
    --log=error
)

for video in "$work_dir/subtitled.mp4" "$work_dir/remotion.mp4"; do
  width="$(ffprobe -v error -select_streams v:0 \
    -show_entries stream=width -of default=noprint_wrappers=1:nokey=1 "$video")"
  height="$(ffprobe -v error -select_streams v:0 \
    -show_entries stream=height -of default=noprint_wrappers=1:nokey=1 "$video")"
  test "$width" = "1080"
  test "$height" = "1920"
  test -s "$video"
done

echo "video-studio runtime contract passed"
