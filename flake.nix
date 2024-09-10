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
        overlay = final: prev: {
          pductl = final.buildGoModule {
            pname = "pductl";
            version = "0.1.0";
            src = ./.;
            vendorHash = "sha256-RnwlwVwU1O2yYMPtUbyz8Xqv1gIZdKNapNhd9QnLjHk=";
          };
        };

        pkgs = import nixpkgs {
          inherit system;
          overlays = [ overlay ];
        };
      in
      {
        devShell = pkgs.mkShell {
          buildInputs = with pkgs; [
            go
            golangci-lint
            reuse
            oapi-codegen
          ];
        };

        overlays = {
          default = overlay;
        };

        packages = {
          default = pkgs.pductl;
        };

        formatter = nixpkgs.nixfmt-rfc-style;
      }
    );
}
