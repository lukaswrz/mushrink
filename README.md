# Mushrink

A simple music compression tool.

If you're anything like me, you obtain music in FLAC format for archival and compress it down to MP3 files for actual listening.
Mushrink exists to automate the continuous compression work that would otherwise have to be done manually.

It:

- scans the input and output directories,
- makes sure the content is properly synchronized (i.e., there are no tracks in the output directory that don't exist in the input directory), and
- compresses the music from FLAC to MP3 in parallel.

## Dependencies

Mushrink requires `ffmpeg` with LAME support in `$PATH` in order to function, which is included in the Nix package.

## NixOS

A NixOS module is provided within this flake.
Here's an example of how to use it:

```nix
{
  services.mushrink.jobs.main = {
    input = "/srv/music";
    output = "/srv/compressed-music";
    timerConfig = {
      OnCalendar = "daily";
      Persistent = true;
    };
    inhibit = ["sleep"];
  };
}
```
