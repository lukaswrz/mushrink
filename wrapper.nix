{
  mushrink-unwrapped,
  ffmpeg,
  symlinkJoin,
  makeWrapper,
  lib,
}:
symlinkJoin {
  inherit (mushrink-unwrapped) pname version meta;

  paths = [ mushrink-unwrapped ];

  nativeBuildInputs = [ makeWrapper ];
  postBuild = ''
    wrapProgram $out/bin/mushrink \
      --suffix PATH : ${lib.makeBinPath [ (ffmpeg.override { withMp3lame = true; }) ]}
  '';
}
