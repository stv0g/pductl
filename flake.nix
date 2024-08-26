# SPDX-FileCopyrightText: 2024 Steffen Vogel <post@steffenvogel.de>
# SPDX-License-Identifier: Apache-2.0
{
  inputs = {
    flake-utils.url = "github:numtide/flake-utils";
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  };
  outputs =
    {
      self,
      nixpkgs,
      flake-utils,
    }:
    flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
      in
      {
        devShell = pkgs.mkShell {
          buildInputs = with pkgs; [
            go
            golangci-lint
            reuse
          ];
        };

        packages = rec {
          pductl = pkgs.buildGoModule {
            pname = "pductl";
            version = "0.1.0";
            src = ./.;
            vendorHash = "sha256-bFNknWOwZXt+nS2+kqVWCMp+AvoL/i0/oguphFUHSw0=";
          };

          default = pductl;
        };

        formatter = nixpkgs.nixfmt-rfc-style;
      }
    );
}
