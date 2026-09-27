{
  description = "A tool for compressing music";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";

  outputs =
    {
      self,
      nixpkgs,
      ...
    }:
    let
      inherit (nixpkgs) lib;

      systems = nixpkgs.lib.systems.flakeExposed;

      forAllSystems =
        f:
        lib.genAttrs systems (
          system:
          f {
            inherit system;
            pkgs = nixpkgs.legacyPackages.${system};
          }
        );
    in
    {
      devShells = forAllSystems (
        { pkgs, ... }:
        {
          default = pkgs.mkShell {
            env.CGO_ENABLED = 0;

            packages = [
              pkgs.go
              pkgs.gotools

              # Formatters
              pkgs.treefmt
              pkgs.nixfmt
              pkgs.prettier
              pkgs.taplo
            ];
          };
        }
      );

      formatter = forAllSystems ({ pkgs, ... }: pkgs.treefmt);

      packages = forAllSystems (
        { pkgs, system, ... }:
        {
          default = pkgs.callPackage ./wrapper.nix {
            mushrink-unwrapped = self.packages.${system}.mushrink-unwrapped;
          };
          mushrink-unwrapped = pkgs.callPackage ./package.nix { };
        }
      );
    };
}
