{
  pkgs,
  lib,
  config,
  ...
}:
let
  cfg = config.services.pdud;
  settingsFormat = pkgs.formats.yaml { };
in
{
  options = {
    services.pdud = {
      enable = lib.mkEnableOption "pdud";
      port = lib.mkOption {
        type = lib.types.port;
        default = 4141;
      };
      settings = lib.mkOption {
        type = settingsFormat.type;
        default = { };
      };
    };
  };

  config = lib.mkIf cfg.enable {
    environment.etc."pdud/config.yaml" = {
      source = settingsFormat.generate "pdud-config.yaml" cfg.settings;
    };

    systemd = {
      services.pdud = {
        description = "PDU control daemon";
        wantedBy = [ "multi-user.target" ];
        after = [ "network.target" ];
        requires = [ "network.target" ];
        restartTriggers = [ "/etc/pdud/config.yaml" ];

        serviceConfig = {
          Type = "notify";
          ExecStart = "${pkgs.pductl}/bin/pdud --config /etc/pdud/config.yaml";
        };
      };

      sockets.pdud = {
        description = "PDU control daemon";
        wantedBy = [ "sockets.target" ];
        listenStreams = [ (toString cfg.port) ];

        socketConfig = {
          Accept = true;
          BindIPv6Only = "both";
        };
      };
    };
  };
}
