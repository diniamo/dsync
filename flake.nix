{
  description = "Dead-simple P2P file synchronization tool using the SSH protocol";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    systems.url = "github:nix-systems/default";
  };

  outputs = {nixpkgs, systems, self, ...}: let
    eachSystem = callback: nixpkgs.lib.genAttrs (import systems) (system: callback nixpkgs.legacyPackages.${system});
  in {
    devShells = eachSystem (pkgs: {
      default = with pkgs; mkShellNoCC {
        packages = [
          go
          gopls
        ];
      };
    });

    packages = eachSystem (pkgs: let
      package = pkgs.callPackage ./package.nix {
        commit = self.shortRev or "dirty";
      };
    in {
      default = package;
      dsync = package;
    });
  };
}
