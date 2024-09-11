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
    {
      nixosModules.default = import ./module.nix;
    }
    // flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
      in
      {
        devShell = import ./shell.nix { inherit pkgs; };

        overlays.default = final: prev: { pductl = final.callPackage ./default.nix { }; };
        packages.default = pkgs.callPackage ./default.nix { };

        formatter = pkgs.nixfmt-rfc-style;
      }
    );
}
