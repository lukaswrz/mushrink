self:
{
  lib,
  pkgs,
  utils,
  config,
  ...
}:
let
  cfg = config.services.mushrink;
  inherit (lib) types;
  inherit (utils.systemdUtils.unitOptions) unitOption;
in
{
  options.services.mushrink = {
    enable = lib.mkEnableOption "mushrink";

    package = lib.mkPackageOption self.packages.${pkgs.stdenv.hostPlatform.system} "default" { };

    jobs = lib.mkOption {
      description = ''
        Compression jobs to run with mushrink.
      '';
      default = { };
      type = types.attrsOf (
        types.submodule {
          options = {
            input = lib.mkOption {
              type = types.str;
              description = ''
                Source directory.
              '';
              example = "/srv/music";
            };

            output = lib.mkOption {
              type = types.str;
              description = ''
                Destination directory for compressed music.
              '';
              example = "/srv/compressed-music";
            };

            workers = lib.mkOption {
              type = types.nullOr types.ints.positive;
              default = null;
              description = ''
                Number of workers.
              '';
            };

            timerConfig = lib.mkOption {
              type = lib.types.nullOr (lib.types.attrsOf unitOption);
              default = {
                OnCalendar = "daily";
                Persistent = true;
              };
              description = ''
                When to run the job.
              '';
            };

            inhibit = lib.mkOption {
              default = [ ];
              type = types.listOf (types.strMatching "^[^:]+$");
              example = [
                "sleep"
              ];
              description = ''
                Run the mushrink process with an inhibition lock taken;
                see {manpage}`systemd-inhibit(1)` for a list of possible operations.
              '';
            };
          };
        }
      );
    };
  };

  config = {
    systemd = lib.mkMerge (
      lib.mapAttrsToList (
        jobName: job:
        let
          unitName = "mushrink-job-${jobName}";
          description = "mushrink job ${jobName}";
        in
        {
          timers.${unitName} = {
            wantedBy = [ "timers.target" ];
            inherit description;
            inherit (job) timerConfig;
          };

          services.${unitName} = {
            inherit description;

            serviceConfig = {
              Type = "oneshot";
              User = "root";
              Group = "root";

              ExecStart =
                let
                  inhibitArgs = [
                    (lib.getExe' config.systemd.package "systemd-inhibit")
                    "--mode"
                    "block"
                    "--who"
                    description
                    "--what"
                    (lib.concatStringsSep ":" job.inhibit)
                    "--why"
                    "Scheduled mushrink job ${jobName}"
                    "--"
                  ];

                  args =
                    (lib.optionals (job.inhibit != [ ]) inhibitArgs)
                    ++ [ (lib.getExe cfg.package) ]
                    ++ lib.optionals (job.workers != null) [
                      "--workers"
                      job.workers
                    ]
                    ++ [
                      "--verbose"
                      "--"
                      job.input
                      job.output
                    ];
                in
                utils.escapeSystemdExecArgs args;
            };
          };
        }
      ) cfg.jobs
    );
  };
}
